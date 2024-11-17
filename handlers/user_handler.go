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

func createUser(ctx context.Context, driver neo4j.DriverWithContext, user models.CreateUserRequest, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	userID := uuid.New().String()

	// Check for username uniqueness
	checkQuery := "MATCH (u:UserProfile {username: $username}) RETURN u LIMIT 1"
	checkParams := map[string]interface{}{"username": user.Username}
	checkResult, err := session.Run(ctx, checkQuery, checkParams)
	if err != nil {
		logger.Error("Failed to check username uniqueness", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error checking username uniqueness")
	}
	if checkResult.Next(ctx) {
		return nil, fiber.NewError(http.StatusConflict, "Username already exists")
	}

	user.UserID = userID

	// Prepare the query and parameters
	query := "CREATE (u:UserProfile {"
	var params = map[string]interface{}{}
	var queryBuilder strings.Builder
	queryBuilder.WriteString(query)

	// Iterate over struct fields using reflection
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

		// Handling time fields with special formatting
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			formattedDate := utils.FormatDate(fieldValue.Interface().(time.Time))
			if formattedDate != nil {
				queryBuilder.WriteString(fmt.Sprintf("%s: $%s, ", fieldName, fieldName))
				params[fieldName] = formattedDate
			}
		} else if fieldValue.Kind() == reflect.Struct || fieldValue.Kind() == reflect.Map {
			if !utils.IsEmpty(fieldValue.Interface()) {
				queryBuilder.WriteString(fmt.Sprintf("%s: $%s, ", fieldName, fieldName))
				params[fieldName] = fieldValue.Interface()
			}
		} else {
			queryBuilder.WriteString(fmt.Sprintf("%s: $%s, ", fieldName, fieldName))
			params[fieldName] = fieldValue.Interface()
		}
	}

	queryBuilder.WriteString("}) RETURN u.user_id AS user_id")
	query = queryBuilder.String()

	// Run the query
	_, err = session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error creating user")
	}

	// Prepare the result map
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

	records, err := result.Single(ctx)
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	var friends []map[string]interface{}
	friendRecords, ok := records.Get("friends")
	if !ok {
		logger.Warn("No friends found for user")
		return nil, fiber.NewError(fiber.StatusNotFound, "No friends found for this user")
	}
	friendList, ok := friendRecords.([]interface{})
	if !ok {
		logger.Error("Failed to cast friends data to []interface{}")
		return nil, fiber.NewError(http.StatusInternalServerError, "Error processing friends data")
	}

	for _, friendData := range friendList {
		friendMap := friendData.(map[string]interface{})
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
		RETURN u, collect({company: c, position: r.position}) AS companies
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

	dateFields := []string{"dob"}
	if err := utils.ConvertMapDateFields(props, dateFields, "2006-01-02"); err != nil {
		logger.Error("Error converting date fields", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, "Failed to parse date fields")
	}

	if err := utils.MapToStruct(props, &user); err != nil {
		logger.Error("Error decoding user properties", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	companyRecords, _ := record.Get("companies")
	companyList := companyRecords.([]interface{})
	user.Companies = make([]models.Company, len(companyList))

	for i, companyData := range companyList {
		compMap := companyData.(map[string]interface{})
		companyNode := compMap["company"].(neo4j.Node).Props
		jobTitle := compMap["position"].(string)

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

	properties, err := utils.StructToMap(updatedData)
	if err != nil {
		logger.Error("Failed to convert struct to map", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, "Internal server error")
	}

	// Execute the transaction with the write operation
	updatedUser, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:UserProfile {user_id: $id})
			SET u += $properties
			RETURN u
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":         id,
			"properties": properties,
		})

		if err != nil {
			logger.Error("Failed to update user profile", zap.Error(err))
			return nil, fiber.NewError(http.StatusInternalServerError, "Failed to update user profile")
		}

		record, err := result.Single(ctx)
		if err != nil {
			logger.Warn("No user found with that ID")
			return nil, fiber.NewError(fiber.StatusNotFound, "No user found with that ID")
		}

		userNode, _ := record.Get("u")
		props := userNode.(neo4j.Node).Props
		var updatedUser models.UserProfile
		if err := utils.MapToStruct(props, &updatedUser); err != nil {
			logger.Error("Error decoding updated user properties", zap.Error(err))
			return nil, fiber.NewError(http.StatusInternalServerError, "Failed to decode updated user profile")
		}

		return updatedUser, nil
	})

	if err != nil {
		return models.UserProfile{}, err
	}

	return updatedUser.(models.UserProfile), nil
}

func deleteUserByID(ctx context.Context, driver neo4j.DriverWithContext, userID string, logger *zap.Logger) error {
	// Start a write transaction
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		deleteQuery := `
      MATCH (u:UserProfile {user_id: $userID})
      DETACH DELETE u
    `
		_, err := tx.Run(ctx, deleteQuery, map[string]interface{}{
			"userID": userID,
		})
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

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			query := `
        MATCH (u1:UserProfile {user_id: $userID1}), (u2:UserProfile {user_id: $userID2})
        MERGE (u1)-[r:FRIEND]->(u2)
        ON CREATE SET r.created_timestamp = timestamp()
        MERGE (u2)-[r2:FRIEND]->(u1)
        ON CREATE SET r2.created_timestamp = timestamp()
        RETURN u1, u2
      `

			result, err := tx.Run(ctx, query, map[string]interface{}{
				"userID1": userID1,
				"userID2": userID2,
			})

			if err != nil {
				logger.Error("Failed to add friend", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
			}

			_, err = result.Single(ctx)
			if err != nil {
				logger.Error("Failed to retrieve result", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve result after creating relationship")
			}
			return nil, nil
		})

	if err != nil {
		return err
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

	_, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {
			query := `
      MATCH (u1:UserProfile {user_id: $userID1})-[r1:FRIEND]->(u2:UserProfile {user_id: $userID2})
      DELETE r1
      WITH u1, u2
      MATCH (u2)-[r2:FRIEND]->(u1)
      DELETE r2
      RETURN u1, u2
    `
			_, err := tx.Run(ctx, query, map[string]interface{}{
				"userID1": userID1,
				"userID2": userID2,
			})
			if err != nil {
				logger.Error("Failed to add friend", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
			}
			return nil, nil
		})

	if err != nil {
		return err
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
