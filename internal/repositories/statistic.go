package repositories

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func GetPostStat(ctx context.Context, driver neo4j.DriverWithContext, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (p:Post)<-[:HAS_POST]-(author:UserProfile)
    OPTIONAL MATCH (p)<-[l:LIKES]-(like_user:UserProfile)
    OPTIONAL MATCH (p)<-[v:HAS_VIEWED]-(view_user:UserProfile)
    OPTIONAL MATCH (p)<-[c:COMMENTED_ON]-(comment_user:Comment)

    WITH
      p,
      author
      like_user.generation AS like_user_gen,
      view_user.generation AS view_user_gen,
      comment_user.generation AS comment_user_gen

    WITH
      p,
      author
      collect(like_user_gen) AS like_user_gens,
      collect(DISTINCT like_user_gen) AS like_user_gen_unique,
      collect(view_user_gen) AS view_user_gens,
      collect(DISTINCT view_user_gen) AS view_user_gen_unique,
      collect(comment_user_gen) AS comment_user_gens,
      collect(DISTINCT comment_user_gen) AS comment_user_gen_unique,

    RETURN
      p.post_id AS post_id,
      p.title AS title,
      p.post_type AS post_type,
      p.media_urls AS media_urls,
      author.first_name + " " + author.last_name AS name,
      author.user_id AS user_id,
      author.profile_picture AS profile_picture,
      size(like_user) AS like_count,
      size(comment_user) AS comment_count,
      size(view_user) AS view_count,
      {
          key: like_user_gen_unique, 
          value: [gen IN like_user_gen_unique | size([x IN like_user_gens WHERE x = gen])]
      } AS like_user_gen,
      {
          key: view_user_gen_unique, 
          value: [gen IN view_user_gen_unique | size([x IN view_user_gens WHERE x = gen])]
      } AS view_user_gen,
      {
          key: comment_user_gen_unique, 
          value: [gen IN comment_user_gen_unique | size([x IN comment_user_gens WHERE x = gen])]
      } AS comment_user_gen,
  `

	// Run the query
	result, err := session.Run(ctx, query, nil)
	if err != nil {
		logger.Error("Failed to retrieve posts", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to retrieve posts")
	}

	// Collect the results
	records, err := result.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect results", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect results")
	}

	var posts []map[string]interface{}

	// Iterate over the records and prepare the results
	for _, record := range records {
		post := record.AsMap()

		// Clean up nil or empty values
		for key, value := range post {
			if value == nil || value == "" {
				delete(post, key)
			}
		}

		posts = append(posts, post)
	}

	return posts, nil
}
