package routes

import (
	"alumni_api/internal/controllers"
	"alumni_api/internal/middlewares"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func UtilsRoute(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	utils := group.Group("/utils")
	utils.Use(middlewares.JWTMiddleware(logger))

	utils.Post("/report", controllers.Report(driver, logger))
}
