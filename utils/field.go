package utils

import (
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
			if field.IsValid() {
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
