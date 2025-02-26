package validators

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

func nameValidation(fl validator.FieldLevel) bool {
	firstName := fl.Parent().FieldByName("FirstName").String()
	firstNameEng := fl.Parent().FieldByName("FirstNameEng").String()
	lastName := fl.Parent().FieldByName("LastName").String()
	lastNameEng := fl.Parent().FieldByName("LastNameEng").String()

	// If FirstName is empty, FirstNameEng must be present
	if firstName == "" && firstNameEng == "" {
		return false
	}
	// If LastName is empty, LastNameEng must be present
	if lastName == "" && lastNameEng == "" {
		return false
	}
	return true
}

func validateEncryptedGeneric(fl validator.FieldLevel) bool {
	// Retrieve the custom encrypted field
	encrypted := fl.Field().Interface()
	// Use reflection to check and access the inner value
	value := reflect.ValueOf(encrypted).FieldByName("Value")

	if !value.IsValid() {
		return false // Invalid field
	}

	// Handle empty values (omitempty rule)
	if value.IsZero() {
		return true
	}

	// Retrieve validation tag (e.g., `min=2,max=100`)
	tag := fl.Param()

	// Validate based on tag rules
	validator := validator.New()
	if err := validator.Var(value.Interface(), tag); err != nil {
		return false
	}

	return true
}

func validateEncrypted(fl validator.FieldLevel) bool {
	// Get the field value
	field := fl.Field()

	// Check if the field is a struct and of type Encrypted[T]
	if field.Kind() == reflect.Struct && strings.HasPrefix(field.Type().Name(), "Encrypted[") {
		// The type is Encrypted[T], we can now access the fields
		// Accessing the "Raw" field (should be []byte)
		rawField := field.FieldByName("Raw")
		valueField := field.FieldByName("Value")

		if !rawField.IsValid() || !valueField.IsValid() {
			return false
		}

		tag := fl.Param()

		// Validate based on tag rules
		validator := validator.New()
		if err := validator.Var(valueField, tag); err != nil {
			return false
		}
		// Return true as the struct was valid
		return true
	}

	return false
}
