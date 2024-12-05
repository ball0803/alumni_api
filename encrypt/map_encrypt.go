package encrypt

import (
	"alumni_api/utils"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"
)

// Encrypt a single field
func encryptStringField(data map[string]interface{}, field string, plaintext string) error {
	encryptText, err := Encrypt(plaintext)
	if err != nil {
		fmt.Errorf(err.Error())
		return err
	}
	data[field] = encryptText

	return nil
}

// EncryptMapFields encrypts specified fields in a map and reassigns them.
func EncryptMapFields(data map[string]interface{}, fields []string) error {
	// Iterate over the top-level fields and encrypt them if they match.
	for key, value := range data {
		// If the field is in the fields list, encrypt it
		if utils.Contains(fields, key) {
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

// DecryptMapFields decrypts specified fields in a map and reassigns them.
func DecryptMapFields(data map[string]interface{}, fields []string, parent string) error {
	// Iterate over the top-level fields and decrypt them if they match.
	for key, value := range data {
		// If the field is in the fields list, decrypt it
		select_key := key
		if parent != "" {
			select_key = strings.Join([]string{parent, key}, ".")
		}

		if utils.Contains(fields, select_key) {
			if strValue, ok := value.(string); ok {
				if utils.IsBase64(strValue) {
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
