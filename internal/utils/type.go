package utils

import (
	"reflect"
	"time"
)

func IsSliceOfByte(val reflect.Value) bool {
	// Check if the value is a slice
	if val.Kind() != reflect.Slice {
		return false
	}

	// Check if the element type of the slice is byte (uint8)
	return val.Type().Elem().Kind() == reflect.Uint8
}

// Helper function to check if a value is zero-valued, nil, or an empty map
func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	if !v.IsValid() || v.IsZero() {
		return true
	}
	if v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len() == 0
	}

	if v.Kind() == reflect.Struct {
		// Check if all fields are zero values
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsZero() {
				return false
			}
		}
		return true
	}

	if v.Kind() == reflect.Struct && v.Type() == reflect.TypeOf(time.Time{}) {
		return v.Interface().(time.Time).IsZero()
	}

	return false
}
