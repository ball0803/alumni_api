package handlers

import (
	"alumni_api/models"
	"alumni_api/validators"
	"alumni_api/websockets"

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
		id := c.Params("user_id")

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

		msg, err := process.SendMessage(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Send Message", logger, err)
		}

		websockets.SendNotification(req.ReceiverID, msg)

		successMessage := "Send Message Successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, msg, logger)

	}
}

func ReplyMessage(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.ReplyMessage
		id := c.Params("user_id")

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

		msg, err := process.ReplyMessage(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Send Message", logger, err)
		}

		if err := encrypt.DecryptMaps(msg, models.MessageDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		websockets.SendNotification(req.ReceiverID, msg)

		successMessage := "Send Reply Message Successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, msg, logger)

	}
}

func EditMessage(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.EditMessage
		id := c.Params("user_id")
		message_id := c.Params("message_id")

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

		req.MessageID = message_id

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := encrypt.EncryptStruct(&req, models.MessageEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.EditMessage(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Send Message", logger, err)
		}

		successMessage := "Edit Message Successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func DeleteMessage(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.DeleteMessage
		id := c.Params("user_id")
		message_id := c.Params("message_id")

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

		req.MessageID = message_id

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.DeleteMessage(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Send Message", logger, err)
		}

		successMessage := "Delete Message Successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func GetChatMessage(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("user_id")
		other_id := c.Params("other_user_id")

		if err := validators.MultipleUUID(id, other_id); err != nil {
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

		exists, err = process.UserExists(c.Context(), driver, other_id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("Receive User: %s not found", id), logger, nil)
		}

		chatMsg, err := process.GetMessage(c.Context(), driver, id, other_id, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Send Message", logger, err)
		}

		if err := encrypt.DecryptMaps(chatMsg, models.ChatMessageDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		successMessage := "Get Chat Message Successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, chatMsg, logger)

	}
}
