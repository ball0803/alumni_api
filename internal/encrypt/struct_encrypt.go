package encrypt

import (
	"alumni_api/internal/utils"
	"fmt"
	"reflect"
)

// EncryptStruct encrypts specified fields in a struct, handling nested fields, arrays, and maps.
func EncryptStruct(inputStruct interface{}, fieldsGroups ...[]string) error {
	for _, fieldsToEncrypt := range fieldsGroups {
		// Find the fields to encrypt in the struct
		for _, field := range fieldsToEncrypt {
			fieldValue := utils.FindStructFields(inputStruct, []string{field})
			fmt.Println(field, fieldValue)

			if len(fieldValue) == 0 {
				continue
			}

			for _, f := range fieldValue {
				// Ensure the field is addressable (pointers or values)
				if f.Kind() != reflect.Ptr && f.CanAddr() {
					f = f.Addr()
				}

				// Call the 'Encrypt' method
				encryptMethod := f.MethodByName("Encrypt")
				if !encryptMethod.IsValid() {
					return fmt.Errorf("Method 'Encrypt' not found for field '%s'", field)
				}

				// Call Encrypt method
				encryptMethod.Call([]reflect.Value{})
			}
		}
	}
	return nil
}

// DecryptStruct decrypts specified fields in a struct and reassigns them.
func DecryptStruct(inputStruct interface{}, fieldsGroups ...[]string) error {
	for _, fieldsToDecrypt := range fieldsGroups {
		// Find the fields to decrypt in the struct
		for _, field := range fieldsToDecrypt {
			fieldValue := utils.FindStructFields(inputStruct, []string{field})

			if len(fieldValue) == 0 {
				continue
			}

			for _, f := range fieldValue {
				// Ensure the field is addressable (pointers or values)
				if f.Kind() != reflect.Ptr && f.CanAddr() {
					f = f.Addr()
				}

				// Call the 'Decrypt' method
				decryptMethod := f.MethodByName("Decrypt")
				if !decryptMethod.IsValid() {
					return fmt.Errorf("Method 'Decrypt' not found for field '%s'", field)
				}

				// Call Decrypt method
				decryptMethod.Call([]reflect.Value{})
			}
		}
	}
	return nil
}
