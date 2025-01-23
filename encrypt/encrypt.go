package encrypt

import (
	"alumni_api/config"
	"alumni_api/utils"
	"reflect"
	"strings"
)

var config_value = config.LoadConfig()
var encryptionKey = config_value.AESEncryptionKey

// AES encrypt function for any type of reflect.Value (string, []byte, int, float)
func AESEncrypt(input reflect.Value) (reflect.Value, error) {
	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	// Check if the type is customtype.Encrypted with any type parameter
	if strings.HasPrefix(input.Type().String(), "customtypes.Encrypted[") && input.Kind() == reflect.Struct {
		// Extract actual value (e.g., `Value` field from a struct)
		var err error
		input, err = extractActualValue(input, "Value")
		if err != nil {
			return reflect.Value{}, err
		}
	}

	// Convert to bytes
	data, err := convertToBytes(input)
	if err != nil {
		return reflect.Value{}, err
	}

	// Encrypt the data
	encrypted, err := encryptAES(data, encryptionKey)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(encrypted), nil
}

func AESDecrypt(input reflect.Value, originalType reflect.Kind) (reflect.Value, error) {
	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	// Check if the type is customtype.Encrypted with any type parameter
	if strings.HasPrefix(input.Type().String(), "customtypes.Encrypted[") && input.Kind() == reflect.Struct {
		// Extract actual value (e.g., `Value` field from a struct)
		var err error
		input, err = extractActualValue(input, "Raw")
		if err != nil {
			return reflect.Value{}, err
		}
	}

	// Decrypt the data
	data, err := decryptAES(input.Bytes(), encryptionKey)
	if err != nil {
		return reflect.Value{}, err
	}

	// Convert decrypted data back to its original type
	return convertFromBytes(data, originalType)
}

// Encrypt any type of reflect.Value (string, []byte, int, float) with type header
func AESEncryptWithHeader(input reflect.Value) (reflect.Value, error) {
	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	// Check if the type is customtype.Encrypted with any type parameter
	if strings.HasPrefix(input.Type().String(), "customtypes.Encrypted[") && input.Kind() == reflect.Struct {
		// Extract actual value (e.g., `Value` field from a struct)
		var err error
		input, err = extractActualValue(input, "Value")
		if err != nil {
			return reflect.Value{}, err
		}
	}

	// Convert to bytes with a type header
	data, err := convertToBytesWithHeader(input)
	if err != nil {
		return reflect.Value{}, err
	}

	// Encrypt the data
	encrypted, err := encryptAES(data, encryptionKey)
	if err != nil {
		return reflect.Value{}, err
	}

	// Return the encrypted data as a reflect.Value
	return reflect.ValueOf(encrypted), nil
}

// Decrypt data with AES and infer type from the embedded header
func AESDecryptWithHeader(input reflect.Value) (reflect.Value, error) {
	if input.Kind() == reflect.Ptr {
		input = input.Elem()
	}

	if !utils.IsSliceOfByte(input) {
		return input, nil
	}

	// Check if the type is customtype.Encrypted with any type parameter
	if strings.HasPrefix(input.Type().String(), "customtypes.Encrypted[") && input.Kind() == reflect.Struct {
		// Extract actual value (e.g., `Value` field from a struct)
		var err error
		input, err = extractActualValue(input, "Raw")
		if err != nil {
			return reflect.Value{}, err
		}
	}

	// Decrypt the data
	decryptedData, err := decryptAES(input.Bytes(), encryptionKey)
	if err != nil {
		return reflect.Value{}, err
	}

	// Convert decrypted data back to its original type using the type header
	return convertFromBytesWithHeader(decryptedData)
}
