package handlers

import (
	"alumni_api/encrypt"
	"alumni_api/models"
	"alumni_api/process"
	// "alumni_api/utils"
	"alumni_api/validators"
	// "encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

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

func GetUserFriendByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
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

		user, err := process.GetUserFriendByID(c.Context(), driver, id, logger)
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

func AddFriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validators.UUID(userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := process.UserExists(c.Context(), driver, userID1, logger)
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

		exists, err = process.UserExists(c.Context(), driver, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID2), logger, nil)
		}

		err = process.AddFriend(c.Context(), driver, userID1, userID2, logger)
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

		exists, err := process.UserExists(c.Context(), driver, userID1, logger)
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

		exists, err = process.UserExists(c.Context(), driver, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID2), logger, nil)
		}

		err = process.Unfriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Successfully remove user %s from user %s", userID1, userID2)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func AddStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CollegeInfo

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

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.AddStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UpdateStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CollegeInfo

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

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.UpdateStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User student info updated successfully"
		return HandleSuccess(c, fiber.StatusInternalServerError, successMessage, nil, logger)
	}
}

func DeleteStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
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

		err = process.DeleteStudentInfo(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User student info removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
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
