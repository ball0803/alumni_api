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

    OPTIONAL MATCH (p)<-[v:HAS_VIEWED]-(view_user:UserProfile)
    WITH p, author, collect(view_user.generation) AS view_gens, collect(DISTINCT view_user.generation) AS view_gen_unique
    WITH p, author, view_gens, view_gen_unique

    OPTIONAL MATCH (p)<-[l:LIKES]-(like_user:UserProfile)
    WITH p, author, view_gens, view_gen_unique, collect(like_user.generation) AS like_gens, collect(DISTINCT like_user.generation) AS like_gen_unique

    OPTIONAL MATCH (p)<-[:COMMENTED_ON]-(:Comment)-[:COMMENTED_BY]->(comment_user:UserProfile)
    WITH p, author, view_gens, view_gen_unique, like_gens, like_gen_unique, collect(comment_user.generation) AS comment_gens, collect(DISTINCT comment_user.generation) AS comment_gen_unique

    RETURN
      p.post_id,
      p.title AS title,
      p.post_type AS post_type,
      p.media_urls AS media_urls,
      author.first_name + " " + author.last_name AS name,
      author.user_id AS user_id,
      author.profile_picture AS profile_picture,
      size(view_gens) AS view_count,
      size(like_gens) AS like_count,
      size(comment_gens) AS comment_count,
      {
          key: view_gen_unique, 
          value: [gen IN view_gen_unique | size([x IN view_gens WHERE x = gen])]
      } AS view_user_gen,
      {
          key: like_gen_unique, 
          value: [gen IN like_gen_unique | size([x IN like_gens WHERE x = gen])]
      } AS like_user_gen,
      {
          key: comment_gen_unique, 
          value: [gen IN comment_gen_unique | size([x IN comment_gens WHERE x = gen])]
      } AS comment_user_gen
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
