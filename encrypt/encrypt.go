package encrypt

import (
	"alumni_api/config"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var encryptionKey = []byte(config.GetEnv("ENCRYPTION_KEY", "default_32_byte_key")) // Must be 32 bytes for AES-256

// Encrypt encrypts plaintext using AES-256 GCM
func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptMapFields encrypts specified fields in a map and reassigns them.
func EncryptMapFields(data map[string]interface{}, fields []string) error {
	// Iterate over the top-level fields and encrypt them if they match.
	for key, value := range data {
		// If the field is in the fields list, encrypt it
		if contains(fields, key) {
			// Encrypt if it's a string
			if strValue, ok := value.(string); ok {
				if err := encryptStringField(data, key, strValue); err != nil {
					return err
				}
			}
		}

		// If it's a nested map, recurse through it
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recurse into the nested map
			if err := EncryptMapFields(nestedMap, fields); err != nil {
				return err
			}
		}

		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Recurse into the map inside the array
					if err := EncryptMapFields(itemMap, fields); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// Encrypt a single field
func encryptStringField(data map[string]interface{}, field string, plaintext string) error {
	// AES encryption logic
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the field
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	// Store the encrypted field as a base64 string
	data[field] = base64.StdEncoding.EncodeToString(ciphertext)

	return nil
}

func encryptStringStructField(data reflect.Value, fieldName string, plaintext string) error {
	// AES encryption logic
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the field
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	// Store the encrypted field as a base64 string
	data.FieldByName(fieldName).SetString(base64.StdEncoding.EncodeToString(ciphertext))

	return nil
}

// EncryptStructFields encrypts specified fields in a struct, handling nested fields, arrays, and maps
func EncryptStructFields(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("data must be a pointer to a struct")
	}

	v = v.Elem()

	return nil
}

// Decrypt decrypts ciphertext using AES-256 GCM
func Decrypt(ciphertext string) (string, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(decodedCiphertext) < nonceSize {
		return "", errors.New("invalid ciphertext")
	}

	nonce, encryptedData := decodedCiphertext[:nonceSize], decodedCiphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
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

// DecryptMapFields decrypts specified fields in a map and reassigns them.
func DecryptMapFields(data map[string]interface{}, fields []string, parent string) error {
	// Iterate over the top-level fields and decrypt them if they match.
	for key, value := range data {
		// If the field is in the fields list, decrypt it
		select_key := key
		if parent != "" {
			select_key = strings.Join([]string{parent, key}, ".")
		}

		if contains(fields, select_key) {
			if strValue, ok := value.(string); ok {
				if isBase64(strValue) {
					if err := decryptStringField(data, key, strValue); err != nil {
						return err
					}
				}
			}
		}

		// If it's a nested map, recurse through it
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recurse into the nested map
			if err := DecryptMapFields(nestedMap, fields, select_key); err != nil {
				return err
			}
		}

		// If it's an array, handle each item
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				// Check if the item is a map (the companies item structure)
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Recursively decrypt the fields for each item in the array
					if err := DecryptMapFields(itemMap, fields, select_key); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Decrypt a single field
func decryptStringField(data map[string]interface{}, field string, encodedCiphertext string) error {
	// AES decryption logic
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decode the base64 encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return fmt.Errorf("failed to decode base64 ciphertext for field %q: %w", field, err)
	}

	// Ensure the ciphertext includes a nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext for field %q is too short", field)
	}

	// Separate nonce and actual ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the field
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt field %q: %w", field, err)
	}

	// Reassign the decrypted value to the map
	data[field] = string(plaintext)

	return nil
}

// Helper function to check if a string exists in a slice
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
