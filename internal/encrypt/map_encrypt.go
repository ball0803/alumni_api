package encrypt

import (
	"alumni_api/internal/utils"
	"fmt"
	"reflect"
)

func EncryptMaps(inputMaps interface{}, fieldsGroups ...[]string) error {
	// Iterate over each group of fields
	for _, fieldsToEncrypt := range fieldsGroups {
		// Iterate over each field in the current group
		for _, field := range fieldsToEncrypt {
			// Find the field in the map
			fieldMappings := utils.FindMapFields(inputMaps, []string{field})

			if len(fieldMappings) == 0 {
				return fmt.Errorf("Field '%s' not found in the map", field)
			}

			// Encrypt all occurrences of the field
			for _, f := range fieldMappings {
				// Get the field value
				value := f.Parent.MapIndex(reflect.ValueOf(f.Field))
				if !value.IsValid() {
					return fmt.Errorf("Field '%s' not found in the map", f.Field)
				}

				// Dereference pointers or interfaces
				for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
					value = value.Elem()
				}

				// Encrypt the value
				encryptValue, err := AESEncryptWithHeader(value)
				if err != nil {
					return err
				}

				// Update the map with the encrypted value
				f.Parent.SetMapIndex(reflect.ValueOf(f.Field), encryptValue)
			}
		}
	}

	return nil
}

func DecryptMaps(inputMaps interface{}, fieldsGroups ...[]string) error {
	// Iterate over each group of fields
	for _, fieldsToDecrypt := range fieldsGroups {
		// Iterate over each field in the current group
		for _, field := range fieldsToDecrypt {
			// Find the field in the map
			fieldMappings := utils.FindMapFields(inputMaps, []string{field})

			if len(fieldMappings) == 0 {
				return fmt.Errorf("Field '%s' not found in the map", field)
			}

			// Decrypt all occurrences of the field
			for _, f := range fieldMappings {
				// Get the field value
				value := f.Parent.MapIndex(reflect.ValueOf(f.Field))
				if !value.IsValid() {
					return fmt.Errorf("Field '%s' not found in the map", f.Field)
				}

				// Dereference pointers or interfaces
				for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
					value = value.Elem()
				}

				// Decrypt the value
				decryptedValue, err := AESDecryptWithHeader(value)
				if err != nil {
					return err
				}

				// Update the map with the decrypted value
				f.Parent.SetMapIndex(reflect.ValueOf(f.Field), decryptedValue)
			}
		}
	}

	return nil
}
