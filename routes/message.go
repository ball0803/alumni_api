package routes

import (
	"alumni_api/handlers"
	"alumni_api/middlewares"
	"alumni_api/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func MessageRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	group.Get("/ws", websocket.New(websockets.WebsocketHandler))

	msg := group.Group("/message")
	msg.Use(middlewares.JWTMiddleware(logger))

	// Message endpoints
	msg.Post("/:id/send", handlers.SendMessage(driver, logger))
}
