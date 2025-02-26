package utils

import (
	"reflect"
	"time"
)

func CheckMapWithTimeField(data interface{}) bool {
	// Use reflection to get the value of the data
	v := reflect.ValueOf(data)

	// Check if the value is a map
	if v.Kind() == reflect.Map {
		// Check if the map has a "Time" key
		for _, key := range v.MapKeys() {
			if key.String() == "Time" {
				// Check if the value of the "Time" key is of the expected type
				timeValue := v.MapIndex(key)
				if timeValue.IsValid() {
					return true
				}
			}
		}
	}

	// Return false if it's not a map with the "Time" field
	return false
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
