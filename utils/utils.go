package utils

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"reflect"
	// "strings"
	"time"
)

// Custom decode hook for dbtype.Date to time.Time conversion
func dateToTimeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from == reflect.TypeOf(dbtype.Date{}) && to == reflect.TypeOf(time.Time{}) {
		return data.(dbtype.Date).Time(), nil
	}
	return data, nil
}

// MapToStruct decodes a map into a struct, using custom hooks and tags
func MapToStruct(input map[string]interface{}, output interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook:       dateToTimeHook,
		Result:           output,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// StructToMap converts a struct into a map with field names as keys
func StructToMap(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Use mapstructure to decode struct into a map with tags support
	config := &mapstructure.DecoderConfig{
		TagName:    "mapstructure",
		Result:     &result,
		ZeroFields: false, // Avoid filling zero-valued fields in the map
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(data); err != nil {
		return nil, fmt.Errorf("failed to decode struct: %w", err)
	}

	// Now recursively convert fields that are slices of structs
	for key, value := range result {
		// If the value is a slice, we need to handle each element in the slice
		if reflect.ValueOf(value).Kind() == reflect.Slice {
			// Convert each element in the slice if it's a struct
			sliceValue := reflect.ValueOf(value)
			var updatedSlice []interface{}
			for i := 0; i < sliceValue.Len(); i++ {
				element := sliceValue.Index(i).Interface()
				if reflect.TypeOf(element).Kind() == reflect.Struct {
					// Recursively convert the struct
					mapElement, err := StructToMap(element)
					if err != nil {
						return nil, fmt.Errorf("failed to convert element in slice: %w", err)
					}
					updatedSlice = append(updatedSlice, mapElement)
				} else {
					// If it's not a struct, just append as is
					updatedSlice = append(updatedSlice, element)
				}
			}
			result[key] = updatedSlice
		}
	}

	ReplaceMapsWithRaw(result)

	// Remove nil or empty maps from `result`
	for key, value := range result {
		if IsEmpty(value) {
			delete(result, key)
		}
	}

	return result, nil
}

// SetEmptyMapsToNil will set empty maps in the struct to nil
func SetEmptyMapsToNil(v interface{}) {
	val := reflect.ValueOf(v)
	// Only proceed if the input is a pointer to a struct
	if val.Kind() != reflect.Ptr || val.IsNil() || val.Elem().Kind() != reflect.Struct {
		return
	}

	// Dereference the pointer to get the actual struct value
	val = val.Elem()

	// Iterate over all fields in the struct
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Check if the field is a map and if it's empty
		if field.Kind() == reflect.Map && field.IsNil() {
			continue
		}
		if field.Kind() == reflect.Map && field.Len() == 0 {
			// Set the empty map to nil
			field.Set(reflect.Zero(field.Type()))
		} else if field.Kind() == reflect.Struct {
			// Recursively handle nested structs
			SetEmptyMapsToNil(field.Addr().Interface())
		}
	}
}

// CleanNullValues removes null fields, empty arrays, and empty maps from a map
func CleanNullValues(input interface{}) interface{} {
	// Check the type of input
	switch v := input.(type) {
	case map[string]interface{}:
		// Iterate over the map
		for key, value := range v {
			// Clean nested maps and arrays
			v[key] = CleanNullValues(value)
			// If the value is null or empty, remove it from the map
			if IsEmpty(v[key]) {
				delete(v, key)
			}
		}
		return v
	case []interface{}:
		// Clean each element of the array
		var cleaned []interface{}
		for _, item := range v {
			cleanedItem := CleanNullValues(item)
			if !IsEmpty(cleanedItem) {
				cleaned = append(cleaned, cleanedItem)
			}
		}
		return cleaned
	default:
		// For non-map, non-array types, just return the value as it is
		return v
	}
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

func FormatDate(value time.Time) interface{} {
	if value.IsZero() {
		return nil
	}
	return value.Format("2006-01-02")
}

func ConvertMapDateFields(inputMap map[string]interface{}, dateFields []string, dateFormat string) error {
	for _, field := range dateFields {
		if strVal, ok := inputMap[field].(string); ok {
			parsedTime, err := time.Parse(dateFormat, strVal)
			if err != nil {
				return errors.New("failed to parse date for field " + field + ": " + err.Error())
			}
			inputMap[field] = parsedTime
		}
	}
	return nil
}
