package handlers

import (
	"alumni_api/models"
	"alumni_api/utils"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
		if err := validateUserID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		user, err := fetchUserByID(c.Context(), driver, id, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User retrieved successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    user,
		})
	}
}

func FindUserByFilter(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestFilter

		// Validate the request data
		if err := ValidateQuery(c, &req); err != nil {
			logger.Warn("Validation failed", zap.Error(err))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		// Fetch users with the given filter
		users, err := fetchUserByFilter(c.Context(), driver, req, logger)
		if err != nil {
			logger.Error("Failed to fetch users", zap.Error(err))
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve users",
				"data":    nil,
			})
		}

		// Successful response
		successMessage := "User(s) retrieved successfully"
		logger.Info(successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    users,
		})
	}
}

// UpdateUserProfile handles updating a user's profile in the Neo4j database.
func UpdateUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserProfile

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err,
				"data":    nil,
			})
		}

		id := c.Params("id")

		if err := validateUserID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		// Perform the update in the Neo4j database
		user, err := updateUserByID(c.Context(), driver, id, req, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		// Response with success
		successMessage := "User profile updated successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    user,
		})
	}
}

func DeleteUserByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := deleteUserByID(c.Context(), driver, id, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User removed successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func GetUserFriendByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUserID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		user, err := getUserFriendByID(c.Context(), driver, id, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		if len(user) == 0 {
			c.Locals("message", errUserNotFound)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "fail",
				"message": errUserNotFound,
				"data":    nil,
			})
		}

		successMessage := "User retrieved successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    user,
		})
	}
}

func CreateUser(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.CreateUserRequest
		req.UserID = uuid.New().String()

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		data, err := createUser(c.Context(), driver, req, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User created successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    data,
		})
	}
}

func AddFriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validateUserID(userID1); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		c.BodyParser(req)
		userID2 := req.UserID
		err := addFriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := fmt.Sprintf("Successfully add user %s to user %s", userID1, userID2)
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func Unfriend(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserFriendRequest
		userID1 := c.Params("id")

		if err := validateUserID(userID1); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		c.BodyParser(req)
		userID2 := req.UserID
		err := unfriend(c.Context(), driver, userID1, userID2, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := fmt.Sprintf("Successfully remove user %s from user %s", userID1, userID2)
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func AddStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.StudentInfoRequest

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err,
				"data":    nil,
			})
		}

		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := addStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User profile updated successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func UpdateStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.StudentInfoRequest

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err,
				"data":    nil,
			})
		}

		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := updateStudentInfo(c.Context(), driver, id, req, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User student info updated successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func DeleteStudentInfo(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := deleteStudentInfo(c.Context(), driver, id, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User student info removed successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func AddUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserRequestCompany

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		id := c.Params("id")
		if err := validateUUID(id); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := addUserCompany(c.Context(), driver, id, req, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User profile updated successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func UpdateUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.UserCompanyUpdateRequest

		if err := ValidateRequest(c, &req); err != nil {
			c.Locals("message", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validateUUIDs(userID, companyID); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := updateUserCompany(c.Context(), driver, userID, companyID, req, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		successMessage := "User company updated successfully"
		c.Locals("message", successMessage)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

func DeleteUserCompany(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Params("user_id")
		companyID := c.Params("company_id")
		if err := validateUUIDs(userID, companyID); err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err := deleteUserCompany(c.Context(), driver, userID, companyID, logger)
		if err != nil {
			c.Locals("message", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}
		successMessage := "User company removed successfully"
		c.Locals("message", successMessage)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": successMessage,
			"data":    nil,
		})
	}
}

// ------------------------------- main -----------------------------------------------

func createUser(ctx context.Context, driver neo4j.DriverWithContext, user models.CreateUserRequest, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Check if username already exists
	checkQuery := "MATCH (u:UserProfile {username: $username}) RETURN u LIMIT 1"
	checkParams := map[string]interface{}{"username": user.Username}
	checkResult, err := session.Run(ctx, checkQuery, checkParams)
	if err != nil {
		logger.Error("Failed to check username uniqueness", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error checking username uniqueness")
	}
	if checkResult.Next(ctx) { // If a record exists
		return nil, fiber.NewError(http.StatusConflict, "Username already exists")
	}

	// Proceed to create the user if username is unique
	query := "CREATE (u:UserProfile {"
	params := map[string]interface{}{}

	// Use reflection to iterate over struct fields
	v := reflect.ValueOf(user)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		fieldName := field.Tag.Get("mapstructure")
		if fieldName == "" {
			fieldName = field.Name
		}

		if fieldValue.IsZero() {
			continue
		}

		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			formattedDate := utils.FormatDate(fieldValue.Interface().(time.Time))
			if formattedDate != nil {
				query += fmt.Sprintf("%s: $%s, ", fieldName, fieldName)
				params[fieldName] = formattedDate
			}
		} else if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Map {
			if !utils.IsEmpty(fieldValue.Interface()) {
				query += fmt.Sprintf("%s: $%s, ", fieldName, fieldName)
				params[fieldName] = fieldValue.Interface()
			}
		} else {
			query += fmt.Sprintf("%s: $%s, ", fieldName, fieldName)
			params[fieldName] = fieldValue.Interface()
		}
	}

	query = strings.TrimSuffix(query, ", ") + "}) RETURN u.user_id AS user_id"

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error creating user")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Failed to retrieve created user ID", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error retrieving created user")
	}

	userID, ok := record.Get("user_id")
	if !ok || userID == nil {
		logger.Error("user_id not found in result")
		return nil, fiber.NewError(http.StatusInternalServerError, "User ID not returned after creation")
	}
	ret := map[string]interface{}{
		"user_id": userID,
	}

	return ret, nil
}

func getUserFriendByID(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (:UserProfile {user_id: $id})-[:FRIEND]->(f:UserProfile)
    RETURN collect({
        user_id: f.user_id,
        username: f.username,
        first_name: f.first_name,
        last_name: f.last_name,
        first_name_eng: f.first_name_eng,
        last_name_eng: f.last_name_eng,
        profile_picture: f.profile_picture
    }) AS friends
  `

	result, err := session.Run(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	// Collect the result into a single record
	records, err := result.Single(ctx)
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	// Get the "friends" field from the record
	var friends []map[string]interface{}
	friendRecords, ok := records.Get("friends")
	if !ok {
		logger.Warn("No friends found for user")
		return nil, fiber.NewError(fiber.StatusNotFound, "No friends found for this user")
	}

	// Ensure the friendRecords is a slice of interfaces
	friendList, ok := friendRecords.([]interface{})
	if !ok {
		logger.Error("Failed to cast friends data to []interface{}")
		return nil, fiber.NewError(http.StatusInternalServerError, "Error processing friends data")
	}

	for _, friendData := range friendList {
		friendMap := friendData.(map[string]interface{})

		// Remove empty maps or zero-value fields
		for key, value := range friendMap {
			if utils.IsEmpty(value) {
				delete(friendMap, key)
			}
		}

		friends = append(friends, friendMap)
	}

	return friends, nil
}

func fetchUserByID(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) (models.UserProfile, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u:UserProfile {user_id: $id})-[r:HAS_WORK_WITH]->(c:Company)
		RETURN u, collect({company: c, job: r.job}) AS companies
	`

	result, err := session.Run(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	userNode, ok := record.Get("u")
	if !ok {
		logger.Warn(errUserNotFound)
		return models.UserProfile{}, fiber.NewError(fiber.StatusNotFound, errUserNotFound)
	}
	props := userNode.(neo4j.Node).Props
	var user models.UserProfile
	if err := utils.MapToStruct(props, &user); err != nil {
		logger.Error("Error decoding user properties", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, errUserNotFound)
	}

	companyRecords, _ := record.Get("companies")
	companyList := companyRecords.([]interface{})
	user.Companies = make([]models.Company, len(companyList))

	for i, companyData := range companyList {
		compMap := companyData.(map[string]interface{})
		companyNode := compMap["company"].(neo4j.Node).Props
		jobTitle := compMap["job"].(string)

		// Map to Company struct and add the job field
		var company models.Company
		if err := utils.MapToStruct(companyNode, &company); err != nil {
			logger.Error("Error mapping company properties", zap.Error(err))
			return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, "Failed to map company")
		}
		company.Position = jobTitle

		user.Companies[i] = company
	}

	return user, nil
}

func fetchUserByFilter(ctx context.Context, driver neo4j.DriverWithContext, filter models.UserRequestFilter, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := "MATCH "
	params := make(map[string]interface{})

	if filter.Field != "" && filter.StudentType != "" {
		query += `
      (:StudentType {name: $studentTypeName})
      <-[:BELONGS_TO_STUDENT_TYPE]-(u:UserProfile)-[:BELONGS_TO_FIELD]->
      (:Field {name: $fieldName})
    `
		params["fieldName"] = filter.Field
		params["studentTypeName"] = filter.StudentType
	} else if filter.Field != "" {
		query += "(u:UserProfile)-[:BELONG_TO_FIELD]->(:Field {name: $fieldName})"
		params["fieldName"] = filter.Field
	} else if filter.StudentType != "" {
		query += "(u:UserProfile)-[:BELONG_TO_STUDENT_TYPE]->(:StudentType {name: $studentTypeName})"
		params["studentTypeName"] = filter.StudentType
	}

	query += " RETURN u"

	// Execute the query
	result, err := session.Run(ctx, query, params)

	if err != nil {
		logger.Error("Failed to run query", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error retrieving data")
	}

	// Collect query results
	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect query results", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error retrieving data")
	}

	var users []map[string]interface{}

	// Process each record
	for _, record := range records {
		userNode, ok := record.Get("u")
		if !ok {
			logger.Warn("User node not found in record")
			continue
		}

		userMap := userNode.(neo4j.Node).Props

		for key, value := range userMap {
			if value == nil || value == "" {
				delete(userMap, key)
			}
		}

		users = append(users, userMap)
	}

	return users, nil
}

func updateUserByID(ctx context.Context, driver neo4j.DriverWithContext, id string, updatedData models.UserProfile, logger *zap.Logger) (models.UserProfile, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MATCH (u:UserProfile {user_id: $id})
		SET u += $properties
		RETURN u
  `

	properties, err := utils.StructToMap(updatedData)
	if err != nil {
		logger.Error("Failed to convert struct to map", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, "Internal server error")
	}
	fmt.Println(properties)

	result, err := session.Run(ctx, query, map[string]interface{}{
		"id":         id,
		"properties": properties,
	})

	if err != nil {
		logger.Error("Failed to update user profile", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, "Failed to update user profile")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("No user found with that ID")
		return models.UserProfile{}, fiber.NewError(fiber.StatusNotFound, "No user found with that ID")
	}

	userNode, _ := record.Get("u")
	props := userNode.(neo4j.Node).Props
	var updatedUser models.UserProfile
	if err := utils.MapToStruct(props, &updatedUser); err != nil {
		logger.Error("Error decoding updated user properties", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, "Failed to decode updated user profile")
	}

	return updatedUser, nil
}

func deleteUserByID(ctx context.Context, driver neo4j.DriverWithContext, userID string, logger *zap.Logger) error {
	// Start a write transaction
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	deleteQuery := `
		MATCH (u:UserProfile {user_id: $userID})
		DETACH DELETE u
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, deleteQuery, map[string]interface{}{"userID": userID})
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		logger.Error("Failed to delete user", zap.String("userID", userID), zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to delete user with user_id %s", userID))
	}

	logger.Info("User deleted successfully", zap.String("userID", userID))
	return nil
}

func addFriend(ctx context.Context, driver neo4j.DriverWithContext, userID1 string, userID2 string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u1:UserProfile {user_id: $userID1}), (u2:UserProfile {user_id: $userID2})
    MERGE (u1)-[r:FRIEND]->(u2)
    ON CREATE SET r.created_timestamp = timestamp()
    MERGE (u2)-[r2:FRIEND]->(u1)
    ON CREATE SET r2.created_timestamp = timestamp()
    RETURN u1, u2
  `

	result, err := session.Run(ctx, query, map[string]interface{}{"userID1": userID1, "userID2": userID2})
	if err != nil {
		logger.Error("Failed to add friend", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
	}

	_, err = result.Single(ctx)
	if err != nil {
		logger.Error("Failed to retrieve result", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve result after creating relationship")
	}

	return nil
}

func unfriend(ctx context.Context, driver neo4j.DriverWithContext, userID1 string, userID2 string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	if userID1 == userID2 {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot unfriend oneself")
	}

	query := `
    MATCH (u1:UserProfile {user_id: $userID1})-[r1:FRIEND]->(u2:UserProfile {user_id: $userID2})
    DELETE r1
    WITH u1, u2
    MATCH (u2)-[r2:FRIEND]->(u1)
    DELETE r2
    RETURN u1, u2
  `

	result, err := session.Run(ctx, query, map[string]interface{}{"userID1": userID1, "userID2": userID2})
	if err != nil {
		logger.Error("Failed to add friend", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
	}

	_, err = result.Single(ctx)
	if err != nil {
		logger.Error("Failed to retrieve result", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve result after creating relationship")
	}

	return nil
}

func addStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, student_info models.StudentInfoRequest, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err1 := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err1 != nil {
		logger.Error("Failed to check user existence", zap.Error(err1))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	query := `
    MERGE (faculty:Faculty {name: $faculty})
    MERGE (faculty)-[:HAS_DEPARTMENT]->(department:Department {name: $department})
    MERGE (department)-[:HAS_FIELD]->(field:Field {name: $field})
    MERGE (field)-[:HAS_STUDENT_TYPE]->(studentType:StudentType {name: $studentType})

    WITH field, studentType
    MATCH (u:UserProfile {user_id: $userID})
    MERGE (u)-[:BELONGS_TO_FIELD]->(field)
    MERGE (u)-[:BELONGS_TO_STUDENT_TYPE]->(studentType)
  `

	params := map[string]interface{}{
		"userID":      id,
		"faculty":     student_info.Faculty,
		"department":  student_info.Department,
		"field":       student_info.Field,
		"studentType": student_info.StudentType,
	}

	_, err2 := session.Run(ctx, query, params)
	if err2 != nil {
		logger.Error("Failed to create or connect student info", zap.Error(err2))
		return fiber.NewError(http.StatusInternalServerError, "Failed to create or connect student info")
	}

	return nil

}

func updateStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, student_info models.StudentInfoRequest, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err1 := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err1 != nil {
		logger.Error("Failed to check user existence", zap.Error(err1))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to begin transaction")
	}

	query := `
    MATCH (u:UserProfile {user_id: $userID})-[r: BELONGS_TO_FIELD|BELONGS_TO_STUDENT_TYPE]->()
    DELETE r

    MERGE (f:Faculty {name: $faculty})
    MERGE (f)-[:HAS_DEPARTMENT]->(d:Department {name: $department})
    MERGE (d)-[:HAS_FIELD]->(fld:Field {name: $field})
    MERGE (fld)-[:HAS_STUDENT_TYPE]->(st:StudentType {name: $studentType})

    MERGE (u)-[:BELONGS_TO_FIELD]->(fld)
    MERGE (u)-[:BELONGS_TO_STUDENT_TYPE]->(st)
  `

	params := map[string]interface{}{
		"userID":      id,
		"faculty":     student_info.Faculty,
		"department":  student_info.Department,
		"field":       student_info.Field,
		"studentType": student_info.StudentType,
	}

	_, err = tx.Run(ctx, query, params)
	if err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to create or connect student info", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect student info")
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return nil
}

func deleteStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err1 := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err1 != nil {
		logger.Error("Failed to check user existence", zap.Error(err1))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to begin transaction")
	}

	query := `
    MATCH (u:UserProfile {user_id: $userID})-[r: BELONGS_TO_FIELD|BELONGS_TO_STUDENT_TYPE]->()
    DELETE r
  `

	params := map[string]interface{}{
		"userID": id,
	}

	_, err = tx.Run(ctx, query, params)
	if err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to remove student info", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect student info")
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return nil
}

func addUserCompany(ctx context.Context, driver neo4j.DriverWithContext, id string, companies models.UserRequestCompany, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Check if the user exists before starting the transaction
	checkUserQuery := `
    MATCH (u:UserProfile {user_id: $userID})
    RETURN u LIMIT 1
  `
	userResult, err := session.Run(ctx, checkUserQuery, map[string]interface{}{
		"userID": id,
	})
	if err != nil {
		logger.Error("Failed to check user existence", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user existence")
	}
	if !userResult.Next(ctx) {
		logger.Warn("UserProfile not found", zap.String("userID", id))
		return fiber.NewError(fiber.StatusNotFound, "UserProfile not found")
	}

	// Begin the transaction for adding or connecting companies
	tx, err := session.BeginTransaction(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to begin transaction")
	}

	// Query to create or connect the company
	query := `
    MERGE (a:Company {name: $name, address: $address})
    ON CREATE SET a.company_id = $companyID
    WITH a
    MATCH (u:UserProfile {user_id: $userID})
    MERGE (u)-[r:HAS_WORK_WITH]->(a)
    SET r.position = $position,
        r.created_timestamp = timestamp()
  `

	for _, company := range companies.Companies {
		// Generate a new companyID only if the company is being created
		params := map[string]interface{}{
			"companyID": uuid.New().String(),
			"userID":    id,
			"name":      company.Company,
			"address":   company.Address,
			"position":  company.Position,
		}
		_, err = tx.Run(ctx, query, params)
		if err != nil {
			logger.Error("Failed to create or connect user company info", zap.Error(err))
			_ = tx.Rollback(ctx)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create or connect user company info")
		}
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		logger.Error("Failed to commit transaction", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return nil
}

func updateUserCompany(ctx context.Context, driver neo4j.DriverWithContext, userID, companyID string, company models.UserCompanyUpdateRequest, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Check if both User and Company exist
	checkExistenceQuery := `
		MATCH (u:UserProfile {user_id: $userID})
		MATCH (c:Company {company_id: $companyID})
		RETURN u, c LIMIT 1
	`
	existsResult, err := session.Run(ctx, checkExistenceQuery, map[string]interface{}{
		"userID":    userID,
		"companyID": companyID,
	})
	if err != nil {
		logger.Error("Failed to check user or company existence", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user or company existence")
	}
	if !existsResult.Next(ctx) {
		logger.Warn("User or Company not found", zap.String("userID", userID), zap.String("companyID", companyID))
		return fiber.NewError(fiber.StatusNotFound, "User or Company not found")
	}

	// Update or create the relationship with position
	_, err = session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			updateQuery := `
			MATCH (a:Company {company_id: $companyID})<-[r:HAS_WORK_WITH]-(u:UserProfile {user_id: $userID})
			SET r.position = $position,
				r.updated_timestamp = timestamp()
		`
			_, err := tx.Run(ctx, updateQuery, map[string]interface{}{
				"companyID": companyID,
				"userID":    userID,
				"position":  company.Position,
			})
			if err != nil {
				logger.Error("Failed to update user company info", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update user company info")
			}
			return nil, nil
		})

	if err != nil {
		return err
	}

	return nil
}

func deleteUserCompany(ctx context.Context, driver neo4j.DriverWithContext, userID, companyID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Check if both User and Company exist
	checkExistenceQuery := `
		MATCH (u:UserProfile {user_id: $userID})
		MATCH (c:Company {company_id: $companyID})
		RETURN u, c LIMIT 1
	`
	existsResult, err := session.Run(ctx, checkExistenceQuery, map[string]interface{}{
		"userID":    userID,
		"companyID": companyID,
	})
	if err != nil {
		logger.Error("Failed to check user or company existence", zap.Error(err))
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to check user or company existence")
	}
	if !existsResult.Next(ctx) {
		logger.Warn("User or Company not found", zap.String("userID", userID), zap.String("companyID", companyID))
		return fiber.NewError(fiber.StatusNotFound, "User or Company not found")
	}

	// Update or create the relationship with position
	_, err = session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			updateQuery := `
			MATCH (a:Company {company_id: $companyID})<-[r:HAS_WORK_WITH]-(u:UserProfile {user_id: $userID})
      DELETE r
		`
			_, err := tx.Run(ctx, updateQuery, map[string]interface{}{
				"companyID": companyID,
				"userID":    userID,
			})
			if err != nil {
				logger.Error("Failed to removed user company info", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to removed user company info")
			}
			return nil, nil
		})

	if err != nil {
		return err
	}

	return nil
}
