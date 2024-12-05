package utils

import (
	"fmt"
	"reflect"
	"strings"

	"bytes"
	"encoding/base64"
)

// Helper function to check if a string exists in a slice
func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// PKCS#7 padding function
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// GetNestedFieldValues takes a pointer to a struct and a dot-separated path.
// It returns a slice of reflect.Value representing all the matching field values at the end of the path.
func GetNestedFieldValues(ptrToStruct interface{}, fieldPath string) []reflect.Value {
	// Ensure the input is a pointer to a struct
	if reflect.TypeOf(ptrToStruct).Kind() != reflect.Ptr || reflect.ValueOf(ptrToStruct).Elem().Kind() != reflect.Struct {
		return nil // Invalid input
	}

	// Get the reflect.Value of the struct
	structValue := reflect.ValueOf(ptrToStruct).Elem()

	fields := strings.Split(fieldPath, ".")
	var result []reflect.Value

	for _, fieldName := range fields {
		if structValue.Kind() == reflect.Ptr {
			if structValue.IsNil() {
				return nil
			}
			structValue = structValue.Elem()
		}

		// Access the field by name
		structValue = structValue.FieldByName(fieldName)
		if !structValue.IsValid() {
			return nil
		}

		// Handle case where the field is an array or slice
		if structValue.Kind() == reflect.Slice || structValue.Kind() == reflect.Array {
			// Iterate over each element in the slice/array and collect matching fields
			for i := 0; i < structValue.Len(); i++ {
				elem := structValue.Index(i)
				if elem.Kind() == reflect.Ptr {
					elem = elem.Elem()
				}
				// After handling array/slice, check for next path part (if any)
				if len(fields) > 1 {
					// Call the function recursively for nested fields
					result = append(result, GetNestedFieldValues(elem.Addr().Interface(), strings.Join(fields[1:], "."))...)
				} else {
					// If there's no further path, just add the value
					result = append(result, elem.FieldByName(fieldName))
				}
			}
			return result
		}
	}

	// If we reach here, we are at the final field
	result = append(result, structValue)
	return result
}

// ModifyFieldValues takes an array of reflect.Value (fields to be modified) and a modifyFunc.
// It applies the modifyFunc to each reflect.Value in the array and modifies the field in place.
func ModifyFieldValues(fields []reflect.Value, modifyFunc func(reflect.Value) (reflect.Value, error)) error {
	for _, field := range fields {
		if field.IsValid() && field.CanSet() {
			newValue, err := modifyFunc(field)
			if newValue.IsValid() && err == nil {
				field.Set(newValue)
			} else {
				return fmt.Errorf("modification function returned invalid value for field")
			}
		} else {
			return fmt.Errorf("field is invalid or cannot be set")
		}
	}
	return nil
}
