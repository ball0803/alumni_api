package routes

import (
	"alumni_api/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func SetupRoutes(app *fiber.App, driver neo4j.DriverWithContext, logger *zap.Logger) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/users/:id", handlers.GetUserByID(driver, logger))
}
