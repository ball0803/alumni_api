package services

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func UserExist(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) (bool, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
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

func UsernameVerify(ctx context.Context, driver neo4j.DriverWithContext, username string, logger *zap.Logger) (bool, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {username: $username})
    RETURN COALESCE(u.is_verify, false) AS is_verify
  `

	params := map[string]interface{}{
		"username": username,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Error running query", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error checking user existence")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	isVerify, ok := record.Get("is_verify")
	if !ok {
		logger.Warn("User not found")
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	return isVerify.(bool), nil
}

func UserVerify(ctx context.Context, driver neo4j.DriverWithContext, id string, logger *zap.Logger) (bool, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {user_id: $id})
    RETURN COALESCE(u.is_verify, false) AS is_verify
  `

	params := map[string]interface{}{
		"id": id,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Error running query", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error checking user existence")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	isVerify, ok := record.Get("is_verify")
	if !ok {
		logger.Warn("User not found")
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	return isVerify.(bool), nil
}

func EmailExist(ctx context.Context, driver neo4j.DriverWithContext, email string, logger *zap.Logger) (bool, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {email: $email})
		RETURN COUNT(u) > 0 AS isEmailExist
  `

	params := map[string]interface{}{
		"email": email,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Error running query", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error checking user existence")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	isVerify, ok := record.Get("isEmailExist")
	if !ok {
		logger.Warn("User not found")
		return false, fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	return isVerify.(bool), nil
}
