package repositories

import (
	"alumni_api/internal/models"
	"alumni_api/internal/utils"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func CreateProfile(ctx context.Context, driver neo4j.DriverWithContext, user models.CreateProfileRequest, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	userID := uuid.New().String()
	user.UserID = userID

	// Prepare the query and parameters
	query := "CREATE (u:UserProfile {"
	params, err := utils.StructToMap(user)
	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Error creating user")
	}
	var queryBuilder strings.Builder
	queryBuilder.WriteString(query)

	fieldCount := 0 // Track field additions

	for fieldName, fieldValue := range params {
		if fieldValue == nil {
			continue
		}

		if fieldCount > 0 {
			queryBuilder.WriteString(", ")
		}
		queryBuilder.WriteString(fmt.Sprintf("%s: $%s", fieldName, fieldName))
		params[fieldName] = fieldValue // Store the field value in params
		fieldCount++
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

func FullTextSeach(ctx context.Context, driver neo4j.DriverWithContext, query_term models.UserFulltextSearch, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    CALL db.index.fulltext.queryNodes("name", $name) YIELD node, score
    RETURN 
        node.user_id as user_id,
        node.first_name + ' ' + node.last_name as fullname,
        node.first_name_eng + ' ' + node.last_name_eng as fullname_eng,
        score
    ORDER BY score DESC
    LIMIT 10
  `
	params := make(map[string]interface{})

	switch query_term.Mode {
	case "contain":
		// For "contain" mode, use wildcard search
		params["name"] = "*" + query_term.Name + "*"
	case "fuzzy":
		// For "fuzzy" mode, use fuzzy search
		params["name"] = "~" + query_term.Name + "~"
	case "exact":
		// For "exact" mode, use exact match (quotes around the name)
		params["name"] = `"` + query_term.Name + `"`
	default:
		// Default to "contain" mode if Mode is unspecified
		params["name"] = "*" + query_term.Name + "*"
	}

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
		users = append(users, record.AsMap())
	}

	return users, nil
}
