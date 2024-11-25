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

	post.Get("", handlers.GetAllPost(driver, logger))

	post.Post("", handlers.CreatePost(driver, logger))

	post.Put("/:post_id", handlers.UpdatePostByID(driver, logger))

	post.Delete("/:post_id", handlers.DeletePostByID(driver, logger))
}
