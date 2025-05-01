package repositories

import (
	"alumni_api/internal/models"
	"alumni_api/internal/services"
	"alumni_api/internal/utils"
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func GetAllPosts(ctx context.Context, driver neo4j.DriverWithContext, logger *zap.Logger) ([]map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (p:Post)<-[:HAS_POST]-(author:UserProfile)
    OPTIONAL MATCH (p)<-[l:LIKES]-(:UserProfile)
    OPTIONAL MATCH (p)<-[v:HAS_VIEWED]-(:UserProfile)
    OPTIONAL MATCH (p)<-[c:COMMENTED_ON]-(:Comment)
    RETURN 
      p.post_id AS post_id,
      p.title AS title,
      p.post_type AS post_type,
      p.media_urls AS media_urls,
      p.start_date AS start_date,
      p.end_date AS end_date,
      author.first_name + " " + author.last_name AS name,
      author.user_id AS user_id,
      author.profile_picture AS profile_picture,
      COUNT(l) AS likes_count,
      COUNT(v) AS views_count,
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

func GetPostByID(ctx context.Context, driver neo4j.DriverWithContext, postID, userID string, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (p:Post {post_id: $post_id})<-[:HAS_POST]-(author:UserProfile)
    OPTIONAL MATCH (p)<-[l:LIKES]-(:UserProfile)
    OPTIONAL MATCH (p)<-[v:HAS_VIEWED]-(viewer:UserProfile)
    OPTIONAL MATCH (p)<-[userLike:LIKES]-(:UserProfile {user_id: $user_id})
    RETURN
      p.post_id AS post_id,
      p.title AS title,
      p.content AS content,
      p.post_type AS post_type,
      p.media_urls AS media_urls,
      p.start_date AS start_date,
      p.end_date AS end_date,
      p.created_timestamp AS created_timestamp,
      author.first_name + " " + author.last_name AS author_name,
      author.user_id AS author_user_id,
      author.profile_picture AS author_profile_picture,
      SIZE(COLLECT(DISTINCT l)) AS likes_count,
      SIZE(COLLECT(DISTINCT v)) AS views_count,
      CASE WHEN userLike IS NULL THEN false ELSE true END AS has_liked
  `

	params := map[string]interface{}{
		"post_id": postID,
		"user_id": userID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to retrieve posts", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to retrieve posts")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Failed to collect results", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect results")
	}

	return record.AsMap(), nil
}

func CreatePost(ctx context.Context, driver neo4j.DriverWithContext, userID string, post models.Post, logger *zap.Logger) (map[string]interface{}, error) {
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

	params := map[string]interface{}{
		"user_id":    userID,
		"post_id":    postID,
		"title":      post.Title,
		"content":    post.Content,
		"post_type":  post.PostType,
		"visibility": post.Visibility,
	}

	if slices.Contains(models.AllowRangeType, post.PostType) {
		query += `,
      start_date: $start_date,
      end_date: $end_date
    `

		params["start_date"] = post.StartDate
		params["end_date"] = post.EndDate
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

func UpdatePostByID(ctx context.Context, driver neo4j.DriverWithContext, postID string, updatedData models.UpdatePostRequest, logger *zap.Logger) error {
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

	if !slices.Contains(models.AllowRangeType, updatedData.PostType) {
		delete(properties, "start_date")
		delete(properties, "end_date")
	}

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

func DeletePostByID(ctx context.Context, driver neo4j.DriverWithContext, postID string, logger *zap.Logger) error {
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

func LikePost(ctx context.Context, driver neo4j.DriverWithContext, userID, postID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
  MATCH (u:UserProfile {user_id: $user_id})
  MATCH (p:Post {post_id: $post_id})
  MERGE (p)<-[:LIKES]-(u)
  `

	params := map[string]interface{}{
		"user_id": userID,
		"post_id": postID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to like post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to like post")
	}

	return nil
}

func UnlikePost(ctx context.Context, driver neo4j.DriverWithContext, userID, postID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
  MATCH (u:UserProfile {user_id: $user_id})-[l:LIKES]->(p:Post {post_id: $post_id})
  DELETE l
  `

	params := map[string]interface{}{
		"user_id": userID,
		"post_id": postID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to unlike post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to unlike post")
	}

	return nil
}

func GetCommentByPostID(ctx context.Context, driver neo4j.DriverWithContext, postID, userID string, logger *zap.Logger) ([]models.Comment, error) {
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
    OPTIONAL MATCH (comment)<-[l:LIKES]-(:UserProfile)
    OPTIONAL MATCH (comment)<-[userLike:LIKES]-(:UserProfile {user_id: $user_id})
    RETURN
        comment.comment_id AS comment_id,
        comment.comment AS content,
        comment.created_timestamp AS created_timestamp,
        user.user_id AS user_id,
        user.username AS username,
        user.first_name + " " + user.last_name AS fullname,
        user.first_name_eng + " " + user.last_name_eng AS fullname_eng,
        user.profile_picture AS profile_picture,
        CASE
            WHEN target:Post THEN null
            ELSE target.comment_id
        END AS parent_comment_id,
        count(l) AS like_count,
        CASE WHEN userLike IS NULL THEN false ELSE true END AS has_liked
    ORDER BY created_timestamp
    `

	params := map[string]any{
		"post_id": postID,
		"user_id": userID,
	}

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

	var comments []models.Comment

	for _, record := range records {
		comment := models.Comment{
			CommentID:       record.Values[0].(string),
			Content:         record.Values[1].(string),
			CreatedAt:       int64(record.Values[2].(int64)),
			UserID:          utils.SafeString(record.Values[3]),
			Username:        utils.SafeString(record.Values[4]),
			Name:            utils.SafeString(record.Values[5]),
			ProfilePicture:  utils.SafeString(record.Values[6]),
			ParentCommentID: utils.OptionalString(record.Values[7]),
			LikeCounts:      int64(record.Values[8].(int64)),
			HasLike:         record.Values[9].(bool),
		}
		comments = append(comments, comment)
	}
	nested := services.BuildCommentTree(comments)
	return nested, nil
}

func CommentPost(ctx context.Context, driver neo4j.DriverWithContext, userID, postID, comment string, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	commentID := uuid.New().String()

	query := `
    MATCH (u:UserProfile {user_id: $user_id})
    MATCH (p:Post {post_id: $post_id})
    CREATE (u)<-[:COMMENTED_BY]-(c:Comment {
      comment_id: $comment_id,
      comment: $comment,
      created_timestamp: timestamp()
    })-[:COMMENTED_ON]->(p)
		RETURN
			u.user_id AS user_id,
			u.username AS username,
			c.comment AS content,
			c.comment_id AS comment_id,
			c.created_timestamp AS created_timestamp
  `

	params := map[string]interface{}{
		"user_id":    userID,
		"post_id":    postID,
		"comment_id": commentID,
		"comment":    comment,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to comment on post", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to return comment")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Failed to collect results", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect results")
	}

	return record.AsMap(), nil
}

func ReplyComment(ctx context.Context, driver neo4j.DriverWithContext, userID, commentID, comment string, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	replyID := uuid.New().String()

	query := `
    MATCH (u:UserProfile {user_id: $user_id})
    MATCH (c:Comment {comment_id: $comment_id})
    CREATE (u)<-[:COMMENTED_BY]-(r:Comment {
      comment_id: $reply_id,
      comment: $comment,
      created_timestamp: timestamp()
    })-[:COMMENTED_ON]->(c)
		RETURN
			u.user_id AS user_id,
			u.username AS username,
			r.comment AS content,
			r.comment_id AS comment_id,
			r.created_timestamp AS created_timestamp
  `

	params := map[string]interface{}{
		"user_id":    userID,
		"comment_id": commentID,
		"reply_id":   replyID,
		"comment":    comment,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to comment on post", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to return comment")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Failed to collect results", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect results")
	}

	return record.AsMap(), nil
}

func UpdateCommentPost(ctx context.Context, driver neo4j.DriverWithContext, commentID, comment string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (c:Comment {comment_id: $comment_id})
    SET c.comment = $comment,
        c.updated_timestamp = timestamp()
  `

	params := map[string]interface{}{
		"comment_id": commentID,
		"comment":    comment,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to update comment on post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to update comment on post")
	}

	return nil
}

func DeleteCommentPost(ctx context.Context, driver neo4j.DriverWithContext, commentID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
    MATCH (c:Comment {comment_id: $comment_id})
    DETACH DELETE c
  `

	params := map[string]interface{}{
		"comment_id": commentID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to delete comment on post", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to delete comment on post")
	}

	return nil
}

func LikeComment(ctx context.Context, driver neo4j.DriverWithContext, userID, commentID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
  MATCH (u:UserProfile {user_id: $user_id})
  MATCH (c:Comment {comment_id: $comment_id})
  MERGE (c)<-[:LIKES]-(u)
  `

	params := map[string]interface{}{
		"user_id":    userID,
		"comment_id": commentID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to like comment", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to like comment")
	}

	return nil
}

func UnlikeComment(ctx context.Context, driver neo4j.DriverWithContext, userID, commentID string, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
  MATCH (u:UserProfile {user_id: $user_id})-[l:LIKES]->(c:Comment {comment_id: $comment_id})
  DELETE l
  `

	params := map[string]interface{}{
		"user_id":    userID,
		"comment_id": commentID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		logger.Error("Failed to unlike comment", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to unlike comment")
	}

	return nil
}
