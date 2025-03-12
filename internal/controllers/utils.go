package controllers

import (
	"alumni_api/internal/models"
	"alumni_api/internal/repositories"
	"alumni_api/internal/validators"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func Report(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		var req models.Report
		req.ID = claim.UserID

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err := repositories.Report(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Report %s Succesfully", req.Type)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}
