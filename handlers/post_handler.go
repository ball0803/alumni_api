package handlers

import (
	"alumni_api/models"
	"alumni_api/utils"
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func getAllPosts(ctx context.Context, driver neo4j.DriverWithContext, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (p:Post)<-[:HAS_POST]-(author:UserProfile)
    MATCH (p)<-[l:LIKES]-(:UserProfile)
    MATCH (p)<-[c:COMMENTED_ON]-(:Comment)
    RETURN 
      p.post_id AS post_id,
      p.title AS title,
      p.post_type AS post_type,
      p.media_urls AS media_urls,
      author.first_name + " " + author.last_name AS name,
      author.user_id AS user_id,
      author.profile_picture AS profile_picture,
      COUNT(l) AS likes_count,
      COUNT(c) AS comments_count
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

func getPostByID(ctx context.Context, driver neo4j.DriverWithContext, postID string, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (p:Post {post_id: $post_id})<-[:HAS_POST]-(author:UserProfile)
    OPTIONAL MATCH (p)<-[l:LIKES]-(:UserProfile)
    OPTIONAL MATCH (p)<-[c:COMMENTED_ON]-(comment:Comment)-[:COMMENTED_BY]->(commenter:UserProfile)
    OPTIONAL MATCH (comment)<-[:REPLIES_TO]-(reply:Comment)-[:COMMENTED_BY]->(replier:UserProfile)
    RETURN 
      p.post_id AS post_id,
      p.title AS title,
      p.content AS content,
      p.post_type AS post_type,
      p.media_urls AS media_urls,
      p.created_timestamp AS created_timestamp,
      author.first_name + " " + author.last_name AS author_name,
      author.user_id AS author_user_id,
      author.profile_picture AS author_profile_picture,
      COUNT(l) AS likes_count,
          COLLECT({
            comment_id: comment.comment_id,
            content: comment.content,
            created_timestamp: comment.created_timestamp,
            commenter_name: commenter.first_name + " " + commenter.last_name,
            commenter_user_id: commenter.user_id,
            commenter_profile_picture: commenter.profile_picture,
            reply_id: reply.comment_id,
            reply_content: reply.content,
            reply_timestamp: reply.created_timestamp,
            replier_name: replier.first_name + " " +  replier.last_name,
            replier_user_id: replier.user_id,
            replier_profile_picture: replier.profile_picture
          }) AS comments
  `

	params := map[string]interface{}{
		"post_id": postID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to retrieve posts", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to retrieve posts")
	}

	// records, err := result.Single(ctx)
	// if err != nil {
	// 	logger.Error("Failed to collect results", zap.Error(err))
	// 	return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect results")
	// }

	return nil, nil
}

func createPost(ctx context.Context, driver neo4j.DriverWithContext, userID string, post models.Post, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	postID := uuid.New().String()

	query := `
    MATCH (u:UserProfile {user_id: $user_id})
    CREATE (u)-[:HAS_POST]->(p:Post {
      post_id: $post_id,
      title: $title,
      content: $content,
      post_type: $post_type,
      visibility: $visibility,
      created_timestamp: timestamp()
    `

	var params = map[string]interface{}{
		"user_id":    userID,
		"post_id":    postID,
		"title":      post.Title,
		"content":    post.Content,
		"post_type":  post.PostType,
		"visibility": post.Visibility,
	}

	if len(post.MediaURL) > 0 {
		query += `, media_urls: $media_urls`
		params["media_urls"] = post.MediaURL
	}

	query += `
    })
  `

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to create post")
	}

	ret := map[string]interface{}{
		"post_id": postID,
	}

	return ret, nil
}

func updatePostByID(ctx context.Context, driver neo4j.DriverWithContext, postID string, updatedData models.UpdatePostRequest, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Convert the updatedData struct to a map for partial update
	properties, err := utils.StructToMap(updatedData)
	if err != nil {
		logger.Error("Failed to convert struct to map", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Internal server error")
	}

	properties["updated_timestamp"] = time.Now().Unix()
	// Directly run the update query
	query := `
		MATCH (p:Post {post_id: $post_id})
		SET p += $properties
	`
	_, err = session.Run(ctx, query, map[string]interface{}{
		"post_id":    postID,
		"properties": properties,
	})

	if err != nil {
		logger.Error("Failed to update post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to update post")
	}

	return nil
}

func deletePostByID(ctx context.Context, driver neo4j.DriverWithContext, postID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MATCH (p:Post {post_id: $post_id})
		DETACH DELETE p
		RETURN count(p) AS deleted
	`
	result, err := session.Run(ctx, query, map[string]interface{}{
		"post_id": postID,
	})

	if err != nil {
		logger.Error("Failed to delete post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to delete post")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Error retrieving result", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Error retrieving result")
	}

	deletedCount, _ := record.Get("deleted")
	if deletedCount == 0 {
		logger.Warn("No post found with given post_id")
		return fiber.NewError(http.StatusNotFound, "Post not found")
	}

	return nil
}

func getPostUserID(ctx context.Context, driver neo4j.DriverWithContext, postID string, logger *zap.Logger) (string, error) {
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
