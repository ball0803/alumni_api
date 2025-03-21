package services

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AddView(ctx context.Context, driver neo4j.DriverWithContext, user_id, post_id string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile {user_id: $user_id}), (p:Post {post_id: $post_id})
    MERGE (u)-[:HAS_VIEWED]->(p)
  `

	params := map[string]interface{}{
		"user_id": user_id,
		"post_id": post_id,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Error running query", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Error checking user existence")
	}

	return nil
}
