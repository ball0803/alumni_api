package controllers

import (
	"alumni_api/internal/encrypt"
	"alumni_api/internal/models"
	"alumni_api/internal/repositories"
	"alumni_api/internal/services"
	"alumni_api/internal/validators"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func CompanyFullTextSearch(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFulltextSearch

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		users, err := repositories.CompanyFullTextSearch(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Company retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, users, logger)
	}
}

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

		exists, err := services.UserExist(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		err1 := validators.SameUser(c, id)
		err2 := validators.UserAdmin(c)
		if err1 != nil && err2 != nil {
			return HandleFailWithStatus(c, err1, logger)
		}

		if err := encrypt.EncryptStruct(&req, models.CompanyEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		companies, err := repositories.AddUserCompany(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := encrypt.DecryptMaps(companies, models.CompanyDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, companies, logger)
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

		exists, err := services.UserExist(c.Context(), driver, userID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID), logger, nil)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := encrypt.EncryptStruct(&req, models.CompanyEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = repositories.UpdateUserCompany(c.Context(), driver, userID, companyID, req, logger)
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

		exists, err := services.UserExist(c.Context(), driver, userID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID), logger, nil)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = repositories.DeleteUserCompany(c.Context(), driver, userID, companyID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}
		successMessage := "User company removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func FindCompanyAssociate(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.Company

		if err := validators.Query(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := encrypt.EncryptStruct(&req, models.CompanyEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		users, err := repositories.FindCompanyAssociate(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Find Associate Users successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, users, logger)
	}
}
