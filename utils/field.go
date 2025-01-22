package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// GetNestedFieldValues takes a pointer to a struct and an array of dot-separated paths.
// It returns a slice of reflect.Value representing all the matching field values at the end of each path.
func FindStructFields(ptrToStruct interface{}, fieldPaths []string) []reflect.Value {
	// Ensure the input is a pointer to a struct
	if reflect.TypeOf(ptrToStruct).Kind() != reflect.Ptr || reflect.ValueOf(ptrToStruct).Elem().Kind() != reflect.Struct {
		return nil // Invalid input
	}

	// Get the reflect.Value of the struct
	structValue := reflect.ValueOf(ptrToStruct).Elem()
	var result []reflect.Value

	// Iterate over each field path and collect matching values
	for _, fieldPath := range fieldPaths {
		// Split the fieldPath into individual field names
		fields := strings.Split(fieldPath, ".")
		result = append(result, traverseStruct(structValue, fields)...)
	}

	return result
}

// Recursive helper function to handle nested field access
func traverseStruct(structValue reflect.Value, fields []string) []reflect.Value {
	var result []reflect.Value

	// If the fieldPath is empty, return an empty slice
	if len(fields) == 0 {
		return result
	}

	// traverseStruct the fields
	for structValue.Kind() == reflect.Ptr || structValue.Kind() == reflect.Interface {
		if structValue.IsNil() {
			return nil
		}
		structValue = structValue.Elem()
	}

	// Iterate through the field names in the path
	for i, fieldName := range fields {
		// If this is the last field in the path
		if i == len(fields)-1 {
			// Add the final field's value to the result
			field := structValue.FieldByName(fieldName)

			if field.IsValid() && !field.IsZero() {
				result = append(result, field)
			}
			return result
		}

		// Access nested fields (e.g., handle struct fields)
		field := structValue.FieldByName(fieldName)
		if !field.IsValid() {
			return nil // If a field doesn't exist, return nil
		}

		// If the field is an array or slice, recurse for each element
		if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			for i := 0; i < field.Len(); i++ {
				elem := field.Index(i)
				result = append(result, traverseStruct(elem, fields[i+1:])...)
			}
			return result
		}

		// Move to the next field (deepest level)
		structValue = field
	}

	return result
}

// Function to find all parent reflect.Values for the given paths in the map
func FindMapFields(input interface{}, paths []string) []struct {
	Parent reflect.Value
	Field  string
} {
	var result []struct {
		Parent reflect.Value
		Field  string
	}

	// Traverse each path
	for _, path := range paths {
		fields := strings.Split(path, ".")
		traverseMap(reflect.ValueOf(input), fields, &result)
	}

	return result
}

// Recursive function to traverse the map/slice based on fields and ensure the last field exists
func traverseMap(value reflect.Value, fields []string, result *[]struct {
	Parent reflect.Value
	Field  string
}) {
	// Dereference pointers and handle interface{} values
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if value.IsValid() {
			value = value.Elem()
		} else {
			return // Avoid dereferencing a nil value
		}
	}

	// Handle different types using a switch case
	switch value.Kind() {
	case reflect.Map:
		// If we've reached the parent of the last field, validate that the last key exists
		//
		if len(fields) == 1 {
			lastKey := reflect.ValueOf(fields[0])
			if value.MapIndex(lastKey).IsValid() {
				// Append the parent map to the result
				*result = append(*result, struct {
					Parent reflect.Value
					Field  string
				}{
					Parent: value,
					Field:  fields[0],
				})
			}
			return
		}

		// Continue traversing the map
		nextValue := value.MapIndex(reflect.ValueOf(fields[0]))
		if nextValue.IsValid() {
			traverseMap(nextValue, fields[1:], result)
		}

	case reflect.Slice:
		// For slices, we need to iterate through each element and continue
		for i := 0; i < value.Len(); i++ {
			traverseMap(value.Index(i), fields, result)
		}

	case reflect.Struct:
		// For structs, fetch the field by name and continue
		field := value.FieldByName(fields[0])
		if field.IsValid() {
			traverseMap(field, fields[1:], result)
		}

	default:
		// If we encounter a type we cannot handle (e.g., basic types), return early
		return
	}
}

// ReplaceMapsWithRaw traverses deeply nested maps, slices, and structs to replace
// maps with both "Raw" and "Value" fields with only the "Raw" value.
func ReplaceMapsWithRaw(input interface{}) interface{} {
	val := reflect.ValueOf(input)

	// Handle pointer and interface dereferencing
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsValid() {
			val = val.Elem()
		} else {
			return input // Return early if nil
		}
	}

	switch val.Kind() {
	case reflect.Map:
		// Iterate over all keys in the map
		// fmt.Println(val)
		for _, key := range val.MapKeys() {
			mapValue := val.MapIndex(key).Elem()
			fmt.Println(key, mapValue, mapValue.Kind())
			// Check if this is a map with "Raw" and "Value" fields
			if isRawValueMap(mapValue) {
				// Replace the map with the "Raw" field value directly
				rawValue := mapValue.MapIndex(reflect.ValueOf("Raw"))
				// Set the "Raw" value in place in the map
				val.SetMapIndex(key, rawValue)
			} else {
				// Process nested values recursively
				processedValue := ReplaceMapsWithRaw(mapValue.Interface())
				val.SetMapIndex(key, reflect.ValueOf(processedValue))
			}
		}
		return input // Return the modified map

	case reflect.Slice:
		// Iterate over the slice and recursively process elements
		for i := 0; i < val.Len(); i++ {
			processedElement := ReplaceMapsWithRaw(val.Index(i).Interface())
			val.Index(i).Set(reflect.ValueOf(processedElement))
		}
		return input // Return the modified slice

	case reflect.Struct:
		// Traverse struct fields and recursively process them
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			// fmt.Println(field, field.Kind())
			if field.CanSet() {
				processedField := ReplaceMapsWithRaw(field.Interface())
				field.Set(reflect.ValueOf(processedField))
			}
		}
		return input // Return the modified struct

	default:
		// Return the value unchanged for unsupported types
		return input
	}
}

// isRawValueMap checks if a map has both "Raw" and "Value" keys.
func isRawValueMap(value reflect.Value) bool {
	// Dereference pointers or interfaces until we reach the actual value
	for value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		if value.IsValid() {
			value = value.Elem()
		} else {
			return false // Invalid value
		}
	}

	// Ensure the value is a map before proceeding
	if value.Kind() != reflect.Map {
		return false
	} else {
		// Safely check for "Raw" and "Value" keys
		rawKey := reflect.ValueOf("Raw")
		valueKey := reflect.ValueOf("Value")

		hasRaw := value.MapIndex(rawKey).IsValid()
		hasValue := value.MapIndex(valueKey).IsValid()

		return hasRaw && hasValue
	}

}
