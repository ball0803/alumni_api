package handlers

import (
	"alumni_api/models"
	"alumni_api/process"
	"alumni_api/validators"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func GetAllPost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {

		posts, err := process.GetAllPosts(c.Context(), driver, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Get Post Sucessfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, posts, logger)
	}
}

func GetPostByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID := c.Params("post_id")

		// if err := validators.UUID(postID); err != nil {
		// 	return HandleFailWithStatus(c, err, logger)
		// }

		posts, err := process.GetPostByID(c.Context(), driver, postID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Get Post Sucessfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, posts, logger)
	}
}

func CreatePost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		var req models.Post

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		data, err := process.CreatePost(c.Context(), driver, claim.UserID, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Create post Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, data, logger)
	}
}

func UpdatePostByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID := c.Params("post_id")

		if err := validators.UUID(postID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID, err := process.GetPostUserID(c.Context(), driver, postID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		var req models.UpdatePostRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.UpdatePostByID(c.Context(), driver, postID, req, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Update post Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func DeletePostByID(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID := c.Params("post_id")

		if err := validators.UUID(postID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID, err := process.GetPostUserID(c.Context(), driver, postID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.DeletePostByID(c.Context(), driver, postID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Deleted post Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func LikePost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID := c.Params("post_id")

		if err := validators.UUID(postID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		err := process.LikePost(c.Context(), driver, claim.UserID, postID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Create like Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UnlikePost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID := c.Params("post_id")

		if err := validators.UUID(postID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		err := process.UnlikePost(c.Context(), driver, claim.UserID, postID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Remove like Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func CommentPost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID := c.Params("post_id")

		if err := validators.UUID(postID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		var req models.CommentRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err := process.CommentPost(c.Context(), driver, claim.UserID, postID, req.Comment, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Create comment Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func ReplyComment(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentID := c.Params("comment_id")

		if err := validators.UUID(commentID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		var req models.CommentRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err := process.ReplyComment(c.Context(), driver, claim.UserID, commentID, req.Comment, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Create comment Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UpdateCommentPost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentID := c.Params("comment_id")

		if err := validators.UUID(commentID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID, err := process.GetCommentUserID(c.Context(), driver, commentID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		var req models.CommentRequest

		if err := validators.Request(c, &req); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.UpdateCommentPost(c.Context(), driver, commentID, req.Comment, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Update comment %s Succesfully", commentID)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func DeleteCommentPost(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentID := c.Params("comment_id")

		if err := validators.UUID(commentID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		userID, err := process.GetCommentUserID(c.Context(), driver, commentID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		if err := validators.SameUser(c, userID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		err = process.DeleteCommentPost(c.Context(), driver, commentID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := fmt.Sprintf("Delete comment %s Succesfully", commentID)
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func LikeComment(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentID := c.Params("comment_id")

		if err := validators.UUID(commentID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		err := process.LikeComment(c.Context(), driver, claim.UserID, commentID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Create like Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}

func UnlikeComment(driver neo4j.DriverWithContext, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentID := c.Params("comment_id")

		if err := validators.UUID(commentID); err != nil {
			return HandleFailWithStatus(c, err, logger)
		}

		claim, ok := c.Locals("claims").(*models.Claims)
		if !ok {
			return HandleFail(c, fiber.StatusUnauthorized, "Unauthorized claim", logger, nil)
		}

		err := process.UnlikeComment(c.Context(), driver, claim.UserID, commentID, logger)
		if err != nil {
			return HandleErrorWithStatus(c, err, logger)
		}

		successMessage := "Remove like Succesfully"
		return HandleSuccess(c, fiber.StatusOK, successMessage, nil, logger)
	}
}
