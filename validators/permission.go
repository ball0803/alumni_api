package validators

import (
	"alumni_api/models"
	"github.com/gofiber/fiber/v2"
)

func SameUser(c *fiber.Ctx, user_id string) error {

	claims, ok := c.Locals("claims").(*models.Claims)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized")
	}

	if claims.UserID != user_id {
		return fiber.NewError(fiber.StatusForbidden, "You do not have permission to this profile")
	}

	return nil
}
