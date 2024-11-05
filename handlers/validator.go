package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func ValidateRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request payload")
	}

	// Validate the struct
	if err := validate.Struct(req); err != nil {
		// Collect validation errors
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make([]string, 0)
		for _, e := range validationErrors {
			errorMessages = append(errorMessages, e.Error())
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errorMessages})
	}

	return nil
}
