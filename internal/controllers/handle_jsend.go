package controllers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func HandleSuccess(c *fiber.Ctx, statusCode int, message string, data interface{}, logger *zap.Logger) error {
	logger.Info(message)

	c.Locals("message", message)
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

func HandleError(c *fiber.Ctx, statusCode int, message string, logger *zap.Logger, err error) error {
	if err != nil {
		logger.Error(message, zap.Error(err))
	} else {
		logger.Warn(message)
	}

	c.Locals("message", message)
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "error",
		"message": message,
		"data":    nil,
	})
}

func HandleFail(c *fiber.Ctx, statusCode int, message string, logger *zap.Logger, err error) error {
	if err != nil {
		logger.Error(message, zap.Error(err))
	} else {
		logger.Warn(message)
	}

	c.Locals("message", message)
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "Fail",
		"message": message,
		"data":    nil,
	})
}

func HandleFailWithStatus(c *fiber.Ctx, err error, logger *zap.Logger) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return HandleFail(c, fiberErr.Code, fiberErr.Message, logger, nil)
	}

	return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
}

func HandleErrorWithStatus(c *fiber.Ctx, err error, logger *zap.Logger) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return HandleError(c, fiberErr.Code, fiberErr.Message, logger, nil)
	}

	return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
}
