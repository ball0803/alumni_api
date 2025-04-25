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
type Comment struct {
    CommentID           string     `json:"comment_id"`
    Content             string     `json:"content"`
    CreatedAt           int64      `json:"created_timestamp"`
    UserID              string     `json:"user_id"`
    Username            string     `json:"username"`
    ParentCommentID     *string    `json:"parent_comment_id"`
    Replies             []Comment  `json:"replies,omitempty"`
}

func buildCommentTree(flat []Comment) []Comment {
    idToComment := make(map[string]*Comment)
    var roots []Comment

    for i := range flat {
        idToComment[flat[i].CommentID] = &flat[i]
    }

    for i := range flat {
        c := &flat[i]
        if c.ParentCommentID == nil {
            roots = append(roots, *c)
        } else if parent, ok := idToComment[*c.ParentCommentID]; ok {
            parent.Replies = append(parent.Replies, *c)
        }
    }

    return roots
}

func GetCommentByPostID(ctx context.Context, driver neo4j.DriverWithContext, postID string, logger *zap.Logger) ([]Comment, error) {
    session := driver.NewSession(ctx, neo4j.SessionConfig{
        DatabaseName: "neo4j",
        AccessMode:   neo4j.AccessModeRead,
    })
    defer session.Close(ctx)

    query := `
        MATCH (p:Post {post_id: $post_id})<-[:HAS_POST]-(:UserProfile)
        MATCH (comment:Comment)-[:COMMENTED_ON]->(target)
        WHERE target = p OR target:Comment
        OPTIONAL MATCH (comment)-[:COMMENTED_BY]->(user:UserProfile)
        RETURN
            comment.comment_id AS comment_id,
            comment.comment AS content,
            comment.created_timestamp AS created_timestamp,
            user.user_id AS user_id,
            user.username AS username,
            CASE
                WHEN target:Post THEN null
                ELSE target.comment_id
            END AS parent_comment_id
        ORDER BY created_timestamp
    `

    params := map[string]any{"post_id": postID}
    result, err := session.Run(ctx, query, params)
    if err != nil {
        logger.Error("Comment query failed", zap.Error(err))
        return nil, fiber.NewError(http.StatusInternalServerError, "Failed to fetch comments")
    }

    records, err := result.Collect(ctx)
    if err != nil {
        logger.Error("Failed to collect comments", zap.Error(err))
        return nil, fiber.NewError(http.StatusInternalServerError, "Failed to process comments")
    }

    var comments []Comment

    for _, record := range records {
        comment := Comment{
            CommentID:       record.Values[0].(string),
            Content:         record.Values[1].(string),
            CreatedAt:       int64(record.Values[2].(int64)),
            UserID:          safeString(record.Values[3]),
            Username:        safeString(record.Values[4]),
            ParentCommentID: optionalString(record.Values[5]),
        }
        comments = append(comments, comment)
    }

    nested := buildCommentTree(comments)
    return nested, nil
}

// Helpers to safely convert Neo4j values
func safeString(v any) string {
    if v == nil {
        return ""
    }
    return v.(string)
}

func optionalString(v any) *string {
    if v == nil {
        return nil
    }
    s := v.(string)
    return &s
}
