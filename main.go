package main

import (
	"alumni_api/config"
	"alumni_api/db"
	"alumni_api/logger"
	"alumni_api/routes"
	"context"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	logger := logger.NewLogger()
	defer logger.Sync()

	ctx := context.Background()
	cfg := config.LoadConfig()

	driver, err := db.ConnectToDB(ctx, cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword, logger)
	if err != nil {
		logger.Fatal("Could not connect to Neo4j", zap.Error(err))
	}
	defer driver.Close(ctx)

	// Set up Fiber app
	app := fiber.New()
	routes.SetupRoutes(app, driver, logger)

	// Start the server
	if err := app.Listen(":3000"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
