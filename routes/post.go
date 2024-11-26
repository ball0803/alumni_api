package routes

import (
	"alumni_api/handlers"
	"alumni_api/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func PostRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	post := group.Group("/post")

	post.Use(middlewares.JWTMiddleware(logger))

	post.Get("/all", handlers.GetAllPost(driver, logger))

	post.Get("/:post_id", handlers.GetPostByID(driver, logger))

	post.Post("", handlers.CreatePost(driver, logger))

	post.Put("/:post_id", handlers.UpdatePostByID(driver, logger))

	post.Delete("/:post_id", handlers.DeletePostByID(driver, logger))

	post.Post("/:post_id/like", handlers.LikePost(driver, logger))

	post.Delete("/:post_id/like", handlers.UnlikePost(driver, logger))

	post.Post("/:post_id/comment", handlers.CommentPost(driver, logger))

	post.Post("/:post_id/comment/:comment_id", handlers.ReplyComment(driver, logger))

	post.Put("/:post_id/comment/:comment_id", handlers.UpdateCommentPost(driver, logger))

	post.Delete("/:post_id/comment/:comment_id", handlers.DeleteCommentPost(driver, logger))
}
