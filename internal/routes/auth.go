package routes

import (
	"alumni_api/internal/controllers"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AuthRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	auth := group.Group("/auth")

	auth.Post("/registry", controllers.Registry(driver, logger))

	auth.Post("/login", controllers.Login(driver, logger))
}
