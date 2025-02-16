package validators

import (
	"alumni_api/pkg/customtypes"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Init() {
	// Register the custom validation tag "encrypted"
	validate.RegisterCustomTypeFunc(ValidateValuer, customtypes.Encrypted[float32]{}, customtypes.Encrypted[int16]{}, customtypes.Encrypted[int32]{}, customtypes.Encrypted[string]{})
	validate.RegisterValidation("encrypted", validateEncrypted)
	validate.RegisterValidation("encrypted_dynamic", validateEncryptedGeneric)
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

func ValidateValuer(field reflect.Value) interface{} {

	if field.Kind() == reflect.Struct && strings.HasPrefix(field.Type().Name(), "Encrypted[") {
		valueField := field.FieldByName("Value")

		if !valueField.IsValid() {
			return nil
		}

		return valueField.Interface()
	}

	return nil
}

// func validateEncrypted(fl validator.FieldLevel) bool {
//
// 	field := fl.Field()
// 	fieldType := field.Type()
// 	encrypted := fl.Field().Interface()
//
// 	fmt.Println(field, fieldType)
// 	// encrypted, ok := fl.Field().Interface().(customtypes.Encrypted[any])
// 	// fmt.Println(encrypted, ok)
// 	// if !ok {
// 	// 	return false
// 	// }
// 	//
// 	// // Validate the Value or other criteria
// 	return false
// 	// return encrypted.Value != nil
// }
