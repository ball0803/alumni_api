package encrypt

import (
	"alumni_api/utils"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
)

// AES encrypt function for any type of reflect.Value (string, []byte, int, float)
func AESEncrypt(input reflect.Value) (reflect.Value, error) {
	var data []byte
	switch input.Kind() {
	case reflect.String:
		// Convert string to byte slice
		data = []byte(input.String())
	case reflect.Slice:
		if input.Type().Elem().Kind() == reflect.Uint8 {
			// Already a byte slice, use as is
			data = input.Bytes()
		} else {
			return reflect.Value{}, errors.New("only byte slices are supported for AES encryption")
		}
	case reflect.Int:
		// Convert int to byte slice
		data = make([]byte, 8)
		binary.BigEndian.PutUint64(data, uint64(input.Int()))
	case reflect.Float64:
		// Convert float64 to byte slice
		data = make([]byte, 8)
		binary.BigEndian.PutUint64(data, math.Float64bits(input.Float()))
	case reflect.Float32:
		data = make([]byte, 4)
		binary.BigEndian.PutUint32(data, math.Float32bits(float32(input.Float())))
	default:
		return reflect.Value{}, errors.New("unsupported input type, only string, []byte, int, and float64 are supported")
	}

	// Encrypt the data using AES encryption
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return reflect.Value{}, err
	}

	// Create a random initialization vector (IV)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return reflect.Value{}, err
	}

	// Apply PKCS#7 padding to the data
	data = utils.PKCS7Pad(data, aes.BlockSize)

	// Create a new cipher block mode for encryption (CBC mode)
	mode := cipher.NewCBCEncrypter(block, iv)

	// Allocate space for the ciphertext
	ciphertext := make([]byte, len(data))
	mode.CryptBlocks(ciphertext, data)

	// Prepend the IV to the ciphertext (required for CBC mode decryption)
	ciphertext = append(iv, ciphertext...)

	// Return the encrypted data wrapped in a reflect.Value
	return reflect.ValueOf(ciphertext), nil
}

// EncryptStructFields encrypts specified fields in a struct, handling nested fields, arrays, and maps
func EncryptStructFields(data interface{}, fields []string) error {
	for _, field := range fields {
		v := utils.GetNestedFieldValues(data, field)
		if err := utils.ModifyFieldValues(v, AESEncrypt); err != nil {
			return err
		}
	}
	return nil
}

// DecryptStructFields decrypts specified fields in a struct and reassigns them.
func DecryptStructFields(data any, fields []string) error {
	// Validate that data is a pointer to a struct
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("data must be a pointer to a struct")
	}

	// Dereference the pointer to access the struct
	v = v.Elem()

	// Prepare AES-GCM cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	for _, field := range fields {
		// Access the field by name
		f := v.FieldByName(field)
		if !f.IsValid() {
			return fmt.Errorf("field %q not found in struct", field)
		}

		// Ensure the field is settable
		if !f.CanSet() {
			return fmt.Errorf("field %q is not settable", field)
		}

		// Ensure the field is of type string
		if f.Kind() != reflect.String {
			return fmt.Errorf("field %q must be of type string", field)
		}

		encodedCiphertext := f.String()
		ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
		if err != nil {
			return fmt.Errorf("failed to decode base64 ciphertext for field %q: %w", field, err)
		}

		nonceSize := aesGCM.NonceSize()
		if len(ciphertext) < nonceSize {
			return fmt.Errorf("ciphertext for field %q is too short", field)
		}

		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return fmt.Errorf("failed to decrypt field %q: %w", field, err)
		}

		f.SetString(string(plaintext))
	}

	return nil
}
