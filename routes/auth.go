package routes

import (
	"alumni_api/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func AuthRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	auth := group.Group("/auth")

	auth.Post("/registry", handlers.Registry(driver, logger))

	auth.Post("/login", handlers.Login(driver, logger))
}
