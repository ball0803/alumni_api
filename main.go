package main

import (
	"alumni_api/config"
	"alumni_api/internal/db"
	"alumni_api/internal/logger"
	"alumni_api/internal/middlewares"
	"alumni_api/internal/routes"
	"alumni_api/internal/validators"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

func main() {
	logger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx := context.Background()
	cfg := config.LoadConfig()
	validators.Init()

	driver, err := db.ConnectToDB(ctx, cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword, logger)
	if err != nil {
		logger.Fatal("Could not connect to Neo4j", zap.Error(err))
	}
	defer driver.Close(ctx)

	// Set up Fiber app
	app := fiber.New()
	api := app.Group("/v1")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Content-Type,Authorization",
		AllowCredentials: true,
	}))
	api.Static("/uploads", "/app/uploads")

	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Use(middlewares.RequestLogger(logger))

	routes.UserRoutes(api, driver, logger)

	routes.AuthRoutes(api, driver, logger)

	routes.PostRoutes(api, driver, logger)

	routes.UploadRoutes(api, driver, logger)

	routes.MessageRoutes(api, driver, logger)

	routes.StatRoutes(api, driver, logger)

	// Start the server
	if err := app.Listen(cfg.ServerPort); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
