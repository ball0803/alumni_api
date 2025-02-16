package validators

import (
	"alumni_api/internal/models"
	"github.com/gofiber/fiber/v2"
)

// validateUserID validates the user ID from the request parameters.
func UID(id string) error {
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, models.ErrIDRequired)
	}
	if err := validate.Var(id, "len=6,numeric"); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, models.ErrInvalidIDFormat)
	}
	return nil
}

func UUID(id string) error {
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, models.ErrIDRequired)
	}
	if err := validate.Var(id, "uuid4"); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, models.ErrInvalidIDFormat)
	}
	return nil
}

func MultipleUUID(ids ...string) error {
	for _, id := range ids {
		if err := UUID(id); err != nil {
			return err
		}
	}
	return nil
}
