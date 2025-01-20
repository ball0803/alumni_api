package encrypt

import (
	"alumni_api/utils"
	"errors"
	"reflect"
)

// EncryptStructFields encrypts specified fields in a struct, handling nested fields, arrays, and maps
func EncryptStruct(inputMaps interface{}, fieldToEncrypt []string) error {
	v := utils.FindStructFields(inputMaps, fieldToEncrypt)

	for _, f := range v {
		if f.Kind() != reflect.Ptr && f.CanAddr() {
			f = f.Addr()
		}

		encryptMethod := f.MethodByName("Encrypt")
		if !encryptMethod.IsValid() {
			return errors.New("Method 'Encrypt' not found")
		}
		encryptMethod.Call([]reflect.Value{})
	}
	return nil
}

// DecryptStructFields decrypts specified fields in a struct and reassigns them.
func DecryptStruct(inputMaps interface{}, fieldToDecrypt []string) error {
	v := utils.FindStructFields(inputMaps, fieldToDecrypt)

	for _, f := range v {
		if f.Kind() != reflect.Ptr && f.CanAddr() {
			f = f.Addr()
		}

		encryptMethod := f.MethodByName("Decrypt")
		if !encryptMethod.IsValid() {
			return errors.New("Method 'Decrypt' not found")
		}
		encryptMethod.Call([]reflect.Value{})
	}
	return nil
}
