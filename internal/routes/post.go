package routes

import (
	"alumni_api/internal/controllers"
	"alumni_api/internal/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func PostRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	post := group.Group("/post")

	post.Use(middlewares.JWTMiddleware(logger))

	post.Get("/all", controllers.GetAllPost(driver, logger))

	post.Get("/:post_id", controllers.GetPostByID(driver, logger))

	post.Post("", controllers.CreatePost(driver, logger))

	post.Put("/:post_id", controllers.UpdatePostByID(driver, logger))

	post.Delete("/:post_id", controllers.DeletePostByID(driver, logger))

	post.Post("/:post_id/like", controllers.LikePost(driver, logger))

	post.Delete("/:post_id/like", controllers.UnlikePost(driver, logger))

	post.Post("/:post_id/comment", controllers.CommentPost(driver, logger))

	post.Post("/:post_id/comment/:comment_id", controllers.ReplyComment(driver, logger))

	post.Put("/:post_id/comment/:comment_id", controllers.UpdateCommentPost(driver, logger))

	post.Delete("/:post_id/comment/:comment_id", controllers.DeleteCommentPost(driver, logger))

	post.Post("/comment/:comment_id/like", controllers.LikeComment(driver, logger))

	post.Delete("/comment/:comment_id/like", controllers.UnlikeComment(driver, logger))
}
