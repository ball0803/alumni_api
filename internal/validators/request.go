package validators

import (
	"alumni_api/internal/utils"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"reflect"
)

func Request(c *fiber.Ctx, req interface{}) error {
	// Parse the request body into the struct or array
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request payload")
	}

	utils.SetEmptyMapsToNil(req)

	// Check if req is a slice (array) and validate each element
	reqValue := reflect.ValueOf(req)

	// If the request is a slice, validate each item in the slice
	if reqValue.Kind() == reflect.Slice {
		for i := 0; i < reqValue.Len(); i++ {
			item := reqValue.Index(i).Interface()
			// Validate each item in the slice
			if err := validateItem(item); err != nil {
				return err
			}
		}
	} else {
		// Validate the struct itself
		if err := validateItem(req); err != nil {
			return err
		}
	}

	return nil
}

// validateItem validates a single struct or item
func validateItem(item interface{}) error {
	// Use reflection to get the underlying value of the item
	reqValue := reflect.ValueOf(item)

	// If item is a pointer, get the value it points to
	if reqValue.Kind() == reflect.Ptr {
		reqValue = reqValue.Elem()
	}

	// Ensure that reqValue is a struct
	if reqValue.Kind() != reflect.Struct {
		return fmt.Errorf("Expected struct type for validation, got %s", reqValue.Kind())
	}

	// Iterate through struct fields to check for required fields
	for i := 0; i < reqValue.NumField(); i++ {
		field := reqValue.Type().Field(i)
		value := reqValue.Field(i)
		// Check for "required" validation tag
		if field.Tag.Get("validate") == "required" && value.IsZero() {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Field '%s' is required", field.Name))
		}
	}

	// Validate the struct using the validator
	if err := validate.Struct(item); err != nil {
		// Collect validation errors
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make([]string, 0)
		for _, e := range validationErrors {
			errorMessages = append(errorMessages, e.Error())
		}
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Validation errors: %v", errorMessages))
	}

	return nil
}
