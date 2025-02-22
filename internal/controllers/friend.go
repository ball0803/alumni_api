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

func GetUserFriendByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		if err := validators.UUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := repositories.UserExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		user, err := repositories.GetUserFriendByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if len(user) == 0 {
			c.Locals("message", models.ErrUserNotFound)
			return HandleError(c, fiber.StatusNotFound, models.ErrUserNotFound, logger, nil)
		}

		successMessage := "User retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func AddFriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validators.UUID(userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := repositories.UserExists(c.Context(), driver, userID1, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID1), logger, nil)
		}

		if err := validators.SameUser(c, userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID2 := req.UserID

		exists, err = repositories.UserExists(c.Context(), driver, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID2), logger, nil)
		}

		err = repositories.AddFriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Successfully add user %s to user %s", userID1, userID2)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func Unfriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validators.UUID(userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := repositories.UserExists(c.Context(), driver, userID1, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID1), logger, nil)
		}

		if err := validators.SameUser(c, userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID2 := req.UserID

		exists, err = repositories.UserExists(c.Context(), driver, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID2), logger, nil)
		}

		err = repositories.Unfriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Successfully remove user %s from user %s", userID1, userID2)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func GetFOAF(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFOAFRequest

		user_id := c.Params("user_id")
		other_id := c.Params("other_id")

		if err := validators.Query(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		// if err := validators.MultipleUUID(user_id, other_id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		exists, err := repositories.UserExists(c.Context(), driver, user_id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", user_id), logger, nil)
		}

		exists, err = repositories.UserExists(c.Context(), driver, other_id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", other_id), logger, nil)
		}

		// if err := validators.SameUser(c, user_id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		degree := int(req.Degree)
		if degree == 0 {
			degree = 3
		}

		foaf, err := repositories.GetFOAF(c.Context(), driver, user_id, other_id, degree, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Successfully retrieve friend of a friend")
		return HandleSuccess(c, fiber.StatusOK, successMessage, foaf, logger)
	}
}
