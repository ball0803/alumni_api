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
	// accessible to public
	post.Get("/all", controllers.GetAllPost(driver, logger))
	post.Get("/:post_id", controllers.GetPostByID(driver, logger))

	postWithAuth := group.Group("/post")
	postWithAuth.Use(middlewares.JWTMiddleware(logger))

	// post
	postWithAuth.Post("", controllers.CreatePost(driver, logger))
	postWithAuth.Put("/:post_id", controllers.UpdatePostByID(driver, logger))
	postWithAuth.Delete("/:post_id", controllers.DeletePostByID(driver, logger))

	// like post
	postWithAuth.Post("/:post_id/like", controllers.LikePost(driver, logger))
	postWithAuth.Delete("/:post_id/like", controllers.UnlikePost(driver, logger))

	// comment post
	postWithAuth.Post("/:post_id/comment", controllers.CommentPost(driver, logger))
	postWithAuth.Post("/:post_id/comment/:comment_id", controllers.ReplyComment(driver, logger))
	postWithAuth.Put("/:post_id/comment/:comment_id", controllers.UpdateCommentPost(driver, logger))
	postWithAuth.Delete("/:post_id/comment/:comment_id", controllers.DeleteCommentPost(driver, logger))

	// like comment
	postWithAuth.Post("/comment/:comment_id/like", controllers.LikeComment(driver, logger))
	postWithAuth.Delete("/comment/:comment_id/like", controllers.UnlikeComment(driver, logger))
}
