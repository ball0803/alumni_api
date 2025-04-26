package services

import (
	"alumni_api/internal/models"
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

func GetAuthorUserID(ctx context.Context, driver neo4j.DriverWithContext, postID string, logger *zap.Logger) (string, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (u:UserProfile)-[:HAS_POST]->(p:Post {post_id: $post_id})
    RETURN u.user_id AS user_id LIMIT 1
    `

	var params = map[string]interface{}{
		"post_id": postID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Post not found", zap.Error(err))
		return "", fiber.NewError(http.StatusInternalServerError, "Post not found")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return "", fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	// Ensure correct key and type assertion
	userID, ok := record.Get("user_id")
	if !ok {
		logger.Warn("User not found")
		return "", nil
	}

	// Safely assert the value to a string
	userIDStr, ok := userID.(string)
	if !ok {
		logger.Error("Error asserting user_id to string")
		return "", fiber.NewError(http.StatusInternalServerError, "Error retrieving user ID")
	}

	return userIDStr, nil
}

func GetCommentUserID(ctx context.Context, driver neo4j.DriverWithContext, commentID string, logger *zap.Logger) (string, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
  MATCH (u:UserProfile)<-[:COMMENTED_BY]-(c:Comment {comment_id: $comment_id})
    RETURN u.user_id AS user_id LIMIT 1
    `

	var params = map[string]interface{}{
		"comment_id": commentID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Comment not found", zap.Error(err))
		return "", fiber.NewError(http.StatusInternalServerError, "Comment not found")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return "", fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	// Ensure correct key and type assertion
	userID, ok := record.Get("user_id")
	if !ok {
		logger.Warn("User not found")
		return "", nil
	}

	// Safely assert the value to a string
	userIDStr, ok := userID.(string)
	if !ok {
		logger.Error("Error asserting user_id to string")
		return "", fiber.NewError(http.StatusInternalServerError, "Error retrieving user ID")
	}

	return userIDStr, nil
}

func BuildCommentTree(flat []models.Comment) []models.Comment {
	idToComment := make(map[string]*models.Comment)
	var roots []*models.Comment

	for i := range flat {
		idToComment[flat[i].CommentID] = &flat[i]
	}

	for i := range flat {
		c := &flat[i]
		if c.ParentCommentID == nil {
			roots = append(roots, c)
		} else if parent, ok := idToComment[*c.ParentCommentID]; ok {
			c.ParentCommentID = nil
			parent.Replies = append(parent.Replies, *c)
		}
	}

	var result []models.Comment
	for _, r := range roots {
		result = append(result, *r)
	}
	return result
}
