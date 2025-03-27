package controllers

import (
	"alumni_api/internal/encrypt"
	"alumni_api/internal/models"
	"alumni_api/internal/repositories"
	"alumni_api/internal/validators"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func GetPostStat(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := validators.UserAdmin(c); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		posts, err := repositories.GetPostStat(c.Context(), driver, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Get Post Statistic Sucessfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, posts, logger)
	}
}

func GetRegistryStat(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := validators.UserAdmin(c); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		posts, err := repositories.GetRegistryStat(c.Context(), driver, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Get Registry Statistic Sucessfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, posts, logger)
	}
}

func GetGenerationSTStat(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.GenerationStat

		if err := validators.UserAdmin(c); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		posts, err := repositories.GetGenerationSTStat(c.Context(), driver, req.CPE, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Get Registry Statistic Sucessfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, posts, logger)
	}
}

func GetUserJob(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := validators.UserAdmin(c); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		user, err := repositories.GetUserJob(c.Context(), driver, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := encrypt.DecryptMaps(user, models.CompanyDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		successMessage := "User retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}
