package repositories

import (
	"alumni_api/internal/models"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func FetchReport(ctx context.Context, driver neo4j.DriverWithContext, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (author:UserProfile)-[:HAS_POST]->(p:Post)-[:BEEN_REPORT]->(r:Report)<-[:REPORT]-(u:UserProfile)
    RETURN
      p.post_id AS post_id,
      p.title AS title,
      p.content AS content,
      p.post_type AS post_type,
      p.media_urls AS media_urls,

      author.first_name + " " + author.last_name AS author_name,
      author.username AS author_username,
      author.user_id AS author_user_id,
      author.profile_picture AS author_profile_picture,

      u.first_name + " " + author.last_name AS reporter_name,
      u.username AS reporter_username,
      u.user_id AS reporter_user_id,
      u.profile_picture AS reporter_profile_picture,

      r.report_id AS report_id,
      r.additional AS additional,
      r.category AS category,
      r.status AS status,
      r.type AS type,
      r.created_timestamp AS created_timestamp
    LIMIT 100
  `

	var params = map[string]interface{}{}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to create post")
	}

	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error(models.ErrRetrievalFailed, zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, models.ErrRetrievalFailed)
	}

	var reports []map[string]interface{}

	for _, record := range records {
		reportMap := record.AsMap()
		reports = append(reports, reportMap)
	}

	return reports, nil
}

func Report(ctx context.Context, driver neo4j.DriverWithContext, report models.Report, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	reportID := uuid.New().String()
	var query string

	switch report.Type {
	case "user":
		query = `
      MATCH (s:UserProfile {user_id: $id}), (u:UserProfile {user_id: $user_id})
      `
	case "post":
		query = `
      MATCH (s:Post {post_id: $id}), (u:UserProfile {user_id: $user_id})
      `
	case "comment":
		query = `
      MATCH (s:Comment {comment_id: $id}), (u:UserProfile {user_id: $user_id})
      `
	default:
		query = `
      MATCH (s:UserProfile {user_id: $id}), (u:UserProfile {user_id: $user_id})
      `
	}

	var params = map[string]interface{}{
		"report_id":  reportID,
		"id":         report.ID,
		"user_id":    report.UserID,
		"type":       report.Type,
		"status":     "pending",
		"category":   report.Category,
		"additional": report.Additional,
	}

	query += `
    CREATE (s)-[:BEEN_REPORT]->(r:Report {
      report_id: $report_id,
      type: $type,
      status: $status,
      category: $category,
      additional: $additional,
      created_timestamp: timestamp()
  })<-[:REPORT]-(u)
  `

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to create post")
	}

	return nil
}
