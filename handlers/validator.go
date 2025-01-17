package handlers

import (
	"alumni_api/models"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"reflect"
)

var validate = validator.New()

// validateUserID validates the user ID from the request parameters.
func validateUserID(id string) error {
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, errIDRequired)
	}
	if err := validate.Var(id, "len=6,numeric"); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errInvalidIDFormat)
	}
	return nil
}

func validateUUID(id string) error {
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, errIDRequired)
	}
	if err := validate.Var(id, "uuid4"); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errInvalidIDFormat)
	}
	return nil
}

func validateUUIDs(ids ...string) error {
	for _, id := range ids {
		if err := validateUUID(id); err != nil {
			return err
		}
	}
	return nil
}

func ValidateRequest(c *fiber.Ctx, req interface{}) error {
	// Parse the request body into the struct or array
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request payload")
	}

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

func ValidateQuery(c *fiber.Ctx, req interface{}) error {
	// Parse the query parameters into the struct
	if err := c.QueryParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid query parameters")
	}

	// Reflect on the fields to check for required tags
	reqValue := reflect.ValueOf(req).Elem()
	for i := 0; i < reqValue.NumField(); i++ {
		field := reqValue.Type().Field(i)
		value := reqValue.Field(i)
		if field.Tag.Get("validate") == "required" && value.IsZero() {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Query parameter '%s' is required", field.Name))
		}
	}

	// Validate the struct using validator
	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make([]string, 0)
		for _, e := range validationErrors {
			errorMessages = append(errorMessages, e.Error())
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errorMessages})
	}

	return nil
}

func ValidateSameUser(c *fiber.Ctx, user_id string) error {

	claims, ok := c.Locals("claims").(*models.Claims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	if claims.UserID != user_id {
		return fiber.NewError(fiber.StatusForbidden, "You do not have permission to this profile")
	}

	return nil
}
