package encrypt

import (
	"alumni_api/utils"
	"fmt"
	"reflect"
)

func EncryptMaps(inputMaps interface{}, fieldToEncrypt []string) error {
	v := utils.FindMapFields(inputMaps, fieldToEncrypt)

	for _, f := range v {
		value := f.Parent.MapIndex(reflect.ValueOf(f.Field))
		if !value.IsValid() {
			return fmt.Errorf("Field '%s' not found in the map", f.Field)
		}

		// Dereference pointers or interfaces
		for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
			value = value.Elem()
		}

		encryptValue, err := AESEncryptWithHeader(value)
		if err != nil {
			return err
		}

		f.Parent.SetMapIndex(reflect.ValueOf(f.Field), encryptValue)
	}

	return nil
}

func DecryptMaps(inputMaps interface{}, fieldToDecrypt []string) error {
	v := utils.FindMapFields(inputMaps, fieldToDecrypt)

	for _, f := range v {
		value := f.Parent.MapIndex(reflect.ValueOf(f.Field))
		if !value.IsValid() {
			return fmt.Errorf("Field '%s' not found in the map", f.Field)
		}

		// Dereference pointers or interfaces
		for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
			value = value.Elem()
		}

		encryptValue, err := AESDecryptWithHeader(value)
		if err != nil {
			return err
		}

		f.Parent.SetMapIndex(reflect.ValueOf(f.Field), encryptValue)
	}

	return nil
}
