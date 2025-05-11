package services

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func IsRequestApproved(ctx context.Context, driver neo4j.DriverWithContext, userID string, logger *zap.Logger) (bool, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {user_id: $user_id})--(r:Request)
		RETURN r.status = "approve" AS isApproved
	`

	params := map[string]interface{}{
		"user_id": userID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Error running query", zap.Error(err))
		return false, fiber.NewError(fiber.StatusInternalServerError, "Error checking request approval")
	}

	if result.Next(ctx) {
		isApproved, ok := result.Record().Get("isApproved")
		if ok {
			return isApproved.(bool), nil
		}
	}

	return false, nil
}
