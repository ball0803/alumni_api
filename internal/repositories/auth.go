package repositories

import (
	"alumni_api/internal/auth"
	"alumni_api/internal/models"
	"alumni_api/internal/utils"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func Login(ctx context.Context, driver neo4j.DriverWithContext, username string, logger *zap.Logger) (models.LoginResponse, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {username: $username})
    RETURN u.user_id AS user_id, u.user_password AS user_password, u.role AS role,
      u.admit_year AS admit_year
  `
	params := map[string]interface{}{
		"username": username,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to query user", zap.Error(err))
		return models.LoginResponse{}, fiber.NewError(fiber.StatusInternalServerError, "Error querying user")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Warn("User not found", zap.String("username", username))
		return models.LoginResponse{}, fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}

	var res models.LoginResponse
	if err := utils.MapToStruct(record.AsMap(), &res); err != nil {
		logger.Error("Error decoding user properties", zap.Error(err))
		return models.LoginResponse{}, fiber.NewError(fiber.StatusInternalServerError, "Error decoding user properties")
	}

	return res, nil
}

func Registry(ctx context.Context, driver neo4j.DriverWithContext, user models.ReqistryRequest, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	userID := uuid.New().String()

	checkQuery := "MATCH (u:UserProfile {username: $username}) RETURN u LIMIT 1"
	checkParams := map[string]interface{}{"username": user.Username}
	checkResult, err := session.Run(ctx, checkQuery, checkParams)
	if err != nil {
		logger.Error("Failed to check username uniqueness", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error checking username uniqueness")
	}
	if checkResult.Next(ctx) {
		return nil, fiber.NewError(fiber.StatusConflict, "Username already exists")
	}

	hashedPass, err := auth.HashPassword(user.Password)
	if err != nil {
		logger.Error("Failed to hash a password", zap.Error(err))
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error hash password")
	}

	ret, err := session.ExecuteWrite(ctx,
		func(tx neo4j.ManagedTransaction) (interface{}, error) {

			query := `
      CREATE (u:UserProfile {
        user_id: $userID,
        username: $username,
        user_password: $password,
        role: $role
      })
    `
			params := map[string]interface{}{
				"userID":   userID,
				"username": user.Username,
				"password": hashedPass,
				"role":     "alumnus",
			}

			_, err = tx.Run(ctx, query, params)
			if err != nil {
				logger.Error("Failed to create user", zap.Error(err))
				return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to add friend")
			}

			ret := map[string]interface{}{
				"user_id": userID,
			}
			return ret, nil
		})

	if err != nil {
		return nil, err
	}

	return ret.(map[string]interface{}), nil
}
