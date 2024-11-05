package handlers

import (
	"alumni_api/models"
	"alumni_api/utils"
	"context"
	"net/http"

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
		if err := validateUserID(id); err != nil {
			return err
		}

		user, err := fetchUserByID(c.Context(), driver, id, logger)
		if err != nil {
			return err
		}

		return c.Status(http.StatusOK).JSON(user)
	}
}

// validateUserID validates the user ID from the request parameters.
func validateUserID(id string) error {
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, errIDRequired)
	}
	if err := validate.Var(id, "len=6,numeric"); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, errInvalidIDFormat)
	}
	return nil
}

// fetchUserByID queries the Neo4j database for a user by ID.
func fetchUserByID(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) (models.UserProfile, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j", AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (u:UserProfile {user_id: $id})
		RETURN u
	`

	result, err := session.Run(ctx, query, map[string]interface{}{"id": id})
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error(errRetrievalFailed, zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, errRetrievalFailed)
	}

	if len(records) == 0 {
		logger.Warn(errUserNotFound)
		return models.UserProfile{}, fiber.NewError(fiber.StatusNotFound, errUserNotFound)
	}

	userNode, ok := records[0].Get("u")
	if !ok {
		logger.Warn(errUserNotFound)
		return models.UserProfile{}, fiber.NewError(fiber.StatusNotFound, errUserNotFound)
	}

	props := userNode.(neo4j.Node).Props
	var user models.UserProfile
	if err := utils.DecodeToStruct(props, &user); err != nil {
		logger.Error("Error decoding user properties", zap.Error(err))
		return models.UserProfile{}, fiber.NewError(http.StatusInternalServerError, errUserNotFound)
	}

	return user, nil
}
