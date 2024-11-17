package handlers

import (
	"alumni_api/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

const (
	errIDRequired      = "ID is required"
	errInvalidIDFormat = "Invalid ID format"
	errUserNotFound    = "No user found with that ID"
	errRetrievalFailed = "Failed to retrieve user from Neo4j"
)

// GetUserByID handles the request to get a user by ID from the Neo4j database.
func GetUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		user, err := fetchUserByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := "User retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

func FindUserByFilter(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestFilter

		if err := ValidateQuery(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		users, err := fetchUserByFilter(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to fetch users", logger, nil)
		}

		successMessage := "User(s) retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, users, logger)
	}
}

// UpdateUserProfile handles updating a user's profile in the Neo4j database.
func UpdateUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserProfile

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation failed", logger, err)
		}

		id := c.Params("id")
		if err := validateUserID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation ID failed", logger, err)
		}

		user, err := updateUserByID(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Update users", logger, err)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

func DeleteUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation ID failed", logger, err)
		}

		err := deleteUserByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, "Failed to Delete users", logger, err)
		}

		successMessage := "User removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func GetUserFriendByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUserID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, "Validation ID failed", logger, err)
		}

		user, err := getUserFriendByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		if len(user) == 0 {
			c.Locals("message", errUserNotFound)
			return HandleError(c, fiber.StatusNotFound, errUserNotFound, logger, nil)
		}

		successMessage := "User retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func CreateUser(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CreateUserRequest

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		data, err := createUser(c.Context(), driver, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := "User created successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}

func AddFriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validateUserID(userID1); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		c.BodyParser(req)
		userID2 := req.UserID
		err := addFriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := fmt.Sprintf("Successfully add user %s to user %s", userID1, userID2)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func Unfriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validateUserID(userID1); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		userID2 := req.UserID
		err := unfriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := fmt.Sprintf("Successfully remove user %s from user %s", userID1, userID2)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func AddStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.StudentInfoRequest

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		err := addStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UpdateStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.StudentInfoRequest

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		err := updateStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusOK, err.Error(), logger, nil)
		}

		successMessage := "User student info updated successfully"
		return HandleSuccess(c, fiber.StatusInternalServerError, successMessage, nil, logger)
	}
}

func DeleteStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		err := deleteStudentInfo(c.Context(), driver, id, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := "User student info removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func AddUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestCompany

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		err := addUserCompany(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UpdateUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserCompanyUpdateRequest

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validateUUIDs(userID, companyID); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		err := updateUserCompany(c.Context(), driver, userID, companyID, req, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}

		successMessage := "User company updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func DeleteUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validateUUIDs(userID, companyID); err != nil {
			return HandleFail(c, fiber.StatusBadRequest, err.Error(), logger, nil)
		}

		err := deleteUserCompany(c.Context(), driver, userID, companyID, logger)
		if err != nil {
			return HandleError(c, fiber.StatusInternalServerError, err.Error(), logger, nil)
		}
		successMessage := "User company removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}
