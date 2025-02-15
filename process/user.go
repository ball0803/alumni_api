package process

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

func CreateUser(ctx context.Context, driver neo4j.DriverWithContext, user models.CreateUserRequest, logger *zap.Logger) (map[string]interface{}, error) {
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

func FetchUserByID(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
      MATCH (u:UserProfile {user_id: $id})
      OPTIONAL MATCH (u)-[r:HAS_WORK_WITH]->(c:Company)
      OPTIONAL MATCH (u)-->(st:StudentType)<--(fld:Field)<--(d:Department)<--(f:Faculty)
      RETURN
          u.user_id AS user_id,
          u.username AS username,
          u.gender AS gender,
          toString(u.dob) AS dob,
          u.first_name + " " + u.last_name AS name,
          u.first_name_eng + " " + u.last_name_eng AS name_eng,
          u.profile_picture AS profile_picture,
          u.role AS role,
          {
              faculty: f.name,
              department: d.name,
              field: fld.name,
              student_type: st.name,
              education_level: u.education_level,
              admit_year: u.admit_year,
              graduate_year: u.graduate_year,
              gpax: u.gpax
          } AS student_info,
          collect({
              company: c.name,
              address: c.address,
              position: r.position
          }) AS companies,
          {
              email: u.email,
              github: u.github,
              linkedin: u.linkdin,
              facebook: u.facebook,
              phone: u.phone
          } AS contact_info
	`

	params := map[string]interface{}{
		"id": id,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	ret := utils.CleanNullValues(record.AsMap()).(map[string]interface{})

	dateFields := []string{"dob"}

	if err := utils.ConvertMapDateFields(ret, dateFields, "2006-01-02"); err != nil {
		logger.Error("Error converting date fields", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to parse date fields")
	}

	return ret, nil
}

func FetchUserByFilter(ctx context.Context, driver neo4j.DriverWithContext, filter models.UserRequestFilter, logger *zap.Logger) ([]map[string]interface{}, error) {
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

func UpdateUserByID(
	ctx context.Context,
	driver neo4j.DriverWithContext,
	id string,
	updatedData models.UpdateUserProfileRequest,
	logger *zap.Logger,
) (map[string]interface{}, error) {
	// Open a new session
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Convert the struct to a map
	properties, err := utils.StructToMap(updatedData)
	if err != nil {
		logger.Error("Failed to convert struct to map", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Internal server error")
	}

	// Execute the transaction with a write operation
	updatedUser, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:UserProfile {user_id: $id})
			SET u += $properties
			RETURN u
		`

		// Run the query
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":         id,
			"properties": properties,
		})
		if err != nil {
			logger.Error("Failed to execute query to update user profile", zap.Error(err))
			return nil, fiber.NewError(http.StatusInternalServerError, "Failed to update user profile")
		}

		// Get the single result
		record, err := result.Single(ctx)
		if err != nil {
			logger.Error("Error retrieving record from query result", zap.Error(err))
			return nil, fiber.NewError(http.StatusInternalServerError, "Error retrieving user data")
		}

		// Extract properties from the node
		userNode, found := record.Get("u")
		if !found {
			logger.Error("Failed to retrieve user node from query result")
			return nil, fiber.NewError(http.StatusInternalServerError, "Failed to retrieve user data")
		}

		node, ok := userNode.(neo4j.Node)
		if !ok {
			logger.Error("Query result is not a neo4j.Node", zap.Any("result", userNode))
			return nil, fiber.NewError(http.StatusInternalServerError, "Invalid data format")
		}

		return node.Props, nil
	})

	if err != nil {
		return nil, err
	}

	// Type assertion to map[string]interface{}
	props, ok := updatedUser.(map[string]interface{})
	if !ok {
		logger.Error("Unexpected data type returned from transaction", zap.Any("data", updatedUser))
		return nil, fiber.NewError(http.StatusInternalServerError, "Invalid user data returned")
	}

	return props, nil
}

func DeleteUserByID(ctx context.Context, driver neo4j.DriverWithContext, userID string, logger *zap.Logger) error {
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

func AddStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, college_info models.CollegeInfo, logger *zap.Logger) error {
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
		"faculty":     college_info.Faculty,
		"department":  college_info.Department,
		"field":       college_info.Field,
		"studentType": college_info.StudentType,
	}

	_, err2 := session.Run(ctx, query, params)
	if err2 != nil {
		logger.Error("Failed to create or connect student info", zap.Error(err2))
		return fiber.NewError(http.StatusInternalServerError, "Failed to create or connect student info")
	}

	return nil

}

func UpdateStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, college_info models.CollegeInfo, logger *zap.Logger) error {
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
		"faculty":     college_info.Faculty,
		"department":  college_info.Department,
		"field":       college_info.Field,
		"studentType": college_info.StudentType,
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

func DeleteStudentInfo(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) error {
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

func UserExists(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) (bool, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
        MATCH (u:UserProfile {user_id: $id})
        RETURN COUNT(u) > 0 AS userExists
    `

	result, err := session.Run(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		logger.Error("Error running query", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error checking user existence")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	userExists, ok := record.Get("userExists")
	if !ok {
		logger.Warn("User not found")
		return false, nil
	}

	return userExists.(bool), nil
}
