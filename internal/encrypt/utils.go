package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
)

const (
	TypeHeaderString  byte = 0x01
	TypeHeaderInt     byte = 0x02
	TypeHeaderFloat32 byte = 0x03
	TypeHeaderFloat64 byte = 0x04
	TypeHeaderBytes   byte = 0x05
)

func IsSliceOfByte(val reflect.Value) bool {
	// Check if the value is a slice
	if val.Kind() != reflect.Slice {
		return false
	}

	// Check if the element type of the slice is byte (uint8)
	return val.Type().Elem().Kind() == reflect.Uint8
}

func extractActualValue(input reflect.Value, fieldName string) (reflect.Value, error) {
	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	if input.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("input is not a struct")
	}

	field := input.FieldByName(fieldName)
	if !field.IsValid() {
		return reflect.Value{}, fmt.Errorf("field '%s' not found", fieldName)
	}
	return field, nil
}

// Convert a reflect.Value to bytes with a type header
func convertToBytesWithHeader(input reflect.Value) ([]byte, error) {
	var data []byte
	var header byte

	switch input.Kind() {
	case reflect.String:
		header = TypeHeaderString
		data = []byte(input.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		header = TypeHeaderInt
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(input.Int()))
		data = buf
	case reflect.Float32:
		header = TypeHeaderFloat32
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, math.Float32bits(float32(input.Float())))
		data = buf
	case reflect.Float64:
		header = TypeHeaderFloat64
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(input.Float()))
		data = buf
	default:
		return nil, fmt.Errorf("unsupported type: %s", input.Kind())
	}

	// Prepend the header to the data
	return append([]byte{header}, data...), nil
}

// Convert bytes to a reflect.Value using the type header
func convertFromBytesWithHeader(data []byte) (reflect.Value, error) {
	if len(data) < 1 {
		return reflect.Value{}, errors.New("data too short to contain a header")
	}

	header := data[0]
	content := data[1:]

	switch header {
	case TypeHeaderString:
		return reflect.ValueOf(string(content)), nil
	case TypeHeaderInt:
		if len(content) != 8 {
			return reflect.Value{}, errors.New("invalid data length for int")
		}
		return reflect.ValueOf(int64(binary.BigEndian.Uint64(content))), nil
	case TypeHeaderFloat32:
		if len(content) != 4 {
			return reflect.Value{}, errors.New("invalid data length for float32")
		}
		return reflect.ValueOf(math.Float32frombits(binary.BigEndian.Uint32(content))), nil
	case TypeHeaderFloat64:
		if len(content) != 8 {
			return reflect.Value{}, errors.New("invalid data length for float64")
		}
		return reflect.ValueOf(math.Float64frombits(binary.BigEndian.Uint64(content))), nil
	default:
		return reflect.Value{}, fmt.Errorf("unknown type header: %x", header)
	}
}

// Encrypt data with AES and a type header
func encryptAESWithHeader(input reflect.Value, key []byte) ([]byte, error) {
	dataWithHeader, err := convertToBytesWithHeader(input)
	if err != nil {
		return nil, err
	}
	return encryptAES(dataWithHeader, key)
}

// Decrypt data with AES and extract type using the header
func decryptAESWithHeader(data []byte, key []byte) (reflect.Value, error) {
	decrypted, err := decryptAES(data, key)
	if err != nil {
		return reflect.Value{}, err
	}
	return convertFromBytesWithHeader(decrypted)
}

func convertToBytes(input reflect.Value) ([]byte, error) {
	switch input.Kind() {
	case reflect.String:
		return []byte(input.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(input.Int()))
		return buf, nil
	case reflect.Float32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, math.Float32bits(float32(input.Float())))
		return buf, nil
	case reflect.Float64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(input.Float()))
		return buf, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", input.Kind())
	}
}

func convertFromBytes(data []byte, kind reflect.Kind) (reflect.Value, error) {
	switch kind {
	case reflect.String:
		return reflect.ValueOf(string(data)), nil
	case reflect.Int, reflect.Int64:
		return reflect.ValueOf(int64(binary.BigEndian.Uint64(data))), nil
	case reflect.Float32:
		return reflect.ValueOf(math.Float32frombits(binary.BigEndian.Uint32(data))), nil
	case reflect.Float64:
		return reflect.ValueOf(math.Float64frombits(binary.BigEndian.Uint64(data))), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type: %s", kind)
	}
}

func encryptAES(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	data = PKCS7Pad(data, aes.BlockSize)
	ciphertext := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, data)

	return append(iv, ciphertext...), nil
}

func decryptAES(data []byte, key []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv, ciphertext := data[:aes.BlockSize], data[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	return PKCS7Unpad(decrypted, aes.BlockSize)
}

func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func PKCS7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padded data")
	}

	// Get the padding size from the last byte
	padSize := int(data[len(data)-1])
	if padSize <= 0 || padSize > blockSize {
		return nil, errors.New("invalid padding size")
	}

	// Check if all the padding bytes are valid
	for i := 0; i < padSize; i++ {
		if data[len(data)-1-i] != byte(padSize) {
			return nil, errors.New("invalid padding bytes")
		}
	}

	// Remove the padding and return the original data
	return data[:len(data)-padSize], nil
}
