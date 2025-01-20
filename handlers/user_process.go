package handlers

import (
	"alumni_api/encrypt"
	"alumni_api/models"
	"alumni_api/utils"
	// "encoding/json"
	"fmt"
	// "net/url"

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

		// if err := validateUUID(id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		// if err := ValidateSameUser(c, id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		user, err := fetchUserByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
	}
}

func FindUserByFilter(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestFilter

		if err := ValidateQuery(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		users, err := fetchUserByFilter(c.Context(), driver, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User(s) retrieved successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, users, logger)
	}
}

// UpdateUserProfile handles updating a user's profile in the Neo4j database.
func UpdateUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UpdateUserProfileRequest

		// Print the parsed request map
		// fmt.Println("Parsed request:", out)

		// if err != nil {
		// 	return HandleErrorWithStatus(c, err, logger)
		// }
		//
		// id := c.Params("id")

		// if err := validateUUID(id); err != nil {
		// 	return HandleErrorWithStatus(c, err, logger)
		// }

		// exists, err := userExists(c.Context(), driver, id, logger)
		// if err != nil {
		// 	return HandleErrorWithStatus(c, err, logger)
		// }
		//
		// if !exists {
		// 	return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		// }

		// if err := ValidateSameUser(c, id); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		reqMap, err := utils.StructToMap(req)
		if err != nil {
			return err
		}

		// fmt.Println(reqMap)

		FieldToEncrypt := []string{
			"gpax",
			// "student_info.admit_year",
			// "student_info.graduate_year",
			// "student_info.education_level",
			// "student_info.email",
			// "student_info.github",
			// "student_info.linkedin",
			// "student_info.facebook",
			// "student_info.phone",
			// "Companies.Company",
			// "Companies.Address",
			// "Companies.Position",
		}

		if err := encrypt.EncryptMapFields(reqMap, FieldToEncrypt); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		// fmt.Println("after encrypt", reqMap)

		// user, err := updateUserByID(c.Context(), driver, id, req, logger)
		// if err != nil {
		// 	return HandleError(c, fiber.StatusInternalServerError, "Failed to Update users", logger, err)
		// }
		//
		// successMessage := "User profile updated successfully"
		// return HandleSuccess(c, fiber.StatusOK, successMessage, user, logger)
		return nil
	}
}

func DeleteUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		if err := validateUUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		if err := ValidateSameUser(c, id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = deleteUserByID(c.Context(), driver, id, logger)
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

		if err := validateUUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		user, err := getUserFriendByID(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
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
			return HandleFailWithStatus(c, err, logger)
		}

		data, err := createUser(c.Context(), driver, req, logger)
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

		if err := validateUUID(userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, userID1, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID1), logger, nil)
		}

		if err := ValidateSameUser(c, userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID2 := req.UserID

		exists, err = userExists(c.Context(), driver, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID2), logger, nil)
		}

		err = addFriend(c.Context(), driver, userID1, userID2, logger)
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

		if err := validateUUID(userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, userID1, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID1), logger, nil)
		}

		if err := ValidateSameUser(c, userID1); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID2 := req.UserID

		exists, err = userExists(c.Context(), driver, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID2), logger, nil)
		}

		err = unfriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Successfully remove user %s from user %s", userID1, userID2)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func AddStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.StudentInfoRequest

		id := c.Params("id")

		if err := validateUUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		if err := ValidateSameUser(c, id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = addStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "User profile updated successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UpdateStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.StudentInfoRequest

		id := c.Params("id")

		if err := validateUUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		if err := ValidateSameUser(c, id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = updateStudentInfo(c.Context(), driver, id, req, logger)
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

		if err := validateUUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		if err := ValidateSameUser(c, id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = deleteStudentInfo(c.Context(), driver, id, logger)
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

		if err := validateUUID(id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, id, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", id), logger, nil)
		}

		if err := ValidateSameUser(c, id); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = addUserCompany(c.Context(), driver, id, req, logger)
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

		if err := ValidateRequest(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validateUUIDs(userID, companyID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, userID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID), logger, nil)
		}

		if err := ValidateSameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = updateUserCompany(c.Context(), driver, userID, companyID, req, logger)
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
		if err := validateUUIDs(userID, companyID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		exists, err := userExists(c.Context(), driver, userID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if !exists {
			return HandleFail(c, fiber.StatusNotFound, fmt.Sprintf("User: %s not found", userID), logger, nil)
		}

		if err := ValidateSameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = deleteUserCompany(c.Context(), driver, userID, companyID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}
		successMessage := "User company removed successfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}
