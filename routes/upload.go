package routes

import (
	"alumni_api/handlers"
	"alumni_api/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func UploadRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	upload := group.Group("/upload")
	upload.Use(middlewares.JWTMiddleware(logger))
	upload.Post("", handlers.Upload(driver, logger))
}
