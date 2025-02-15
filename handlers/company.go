package handlers

import (
	"alumni_api/encrypt"
	"alumni_api/models"
	"alumni_api/process"
	"alumni_api/validators"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AddUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestCompany

		id := c.Params("id")

		if err := validators.UUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
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

		if err := encrypt.EncryptStruct(req, models.UserEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.AddUserCompany(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UpdateUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserCompanyUpdateRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validators.MultipleUUID(userID, companyID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := process.UserExists(c.Context(), driver, userID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID), logger, nil)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := encrypt.EncryptStruct(req, models.UserEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.UpdateUserCompany(c.Context(), driver, userID, companyID, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User company updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func DeleteUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validators.MultipleUUID(userID, companyID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := process.UserExists(c.Context(), driver, userID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID), logger, nil)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.DeleteUserCompany(c.Context(), driver, userID, companyID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}
		successMessage := "User company removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}
