package handlers

import (
	"alumni_api/models"
	"alumni_api/validators"

	"alumni_api/encrypt"
	"alumni_api/process"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func SendMessage(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.Message
		id := c.Params("id")

		if err := validators.UUID(id); err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		exists, err := process.UserExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		if err := validators.SameUser(c, id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		req.SenderID = id

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err = process.UserExists(c.Context(), driver, req.ReceiverID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("Receive User: %s not found", id), logger, nil)
		}

		if err := encrypt.EncryptStruct(&req, models.MessageEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.SendMessage(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Send Message", logger, err)
		}

		successMessage := "Send Message Successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)

	}
}
