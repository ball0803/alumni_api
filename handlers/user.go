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

func CreateUser(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CreateUserRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		data, err := process.CreateUser(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User created successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}

// GetUserByID handles the request to get a user by ID from the Neo4j database.
func GetUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		// if err := validators.UUID(id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		exists, err := process.UserExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		// if err := validators.SameUser(c, id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		user, err := process.FetchUserByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := encrypt.DecryptMaps(user, models.UserDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		successMessage := "User retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

// UpdateUserProfile handles updating a user's profile in the Neo4j database.
func UpdateUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UpdateUserProfileRequest

		id := c.Params("id")

		// if err := validators.UUID(id); err != nil {
		// 	return HandleErrorWithStatus(c, err, logger)
		// }

		exists, err := process.UserExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		// if err := validators.SameUser(c, id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := encrypt.EncryptStruct(&req, models.UserEncryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		user, err := process.UpdateUserByID(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Update users", logger, err)
		}

		if err := encrypt.DecryptMaps(user, models.StudentInfoDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

func DeleteUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		if err := validators.UUID(id); err != nil {
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

		err = process.DeleteUserByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func FindUserByFilter(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestFilter

		if err := validators.Query(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		users, err := process.FetchUserByFilter(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := encrypt.DecryptMaps(users, models.UserDecryptField); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		successMessage := "User(s) retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, users, logger)
	}
}

func NameFullTextSearch(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFulltextSearch

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		users, err := process.FullTextSeach(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User(s) retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, users, logger)
	}
}
