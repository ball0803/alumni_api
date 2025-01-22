package validators

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"reflect"
)

func Query(c *fiber.Ctx, req interface{}) error {
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
