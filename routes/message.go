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
	group.Use("/ws", middlewares.WebSocketMiddleware(logger))
	group.Get("/ws", websocket.New(websockets.WebsocketHandler))

	msg := group.Group("/user/:user_id/message")
	msg.Use(middlewares.JWTMiddleware(logger))

	// Message endpoints
	msg.Post("/send", handlers.SendMessage(driver, logger))
	msg.Post("/reply", handlers.ReplyMessage(driver, logger))
	msg.Put("/:message_id", handlers.EditMessage(driver, logger))
	msg.Delete("/:message_id", handlers.DeleteMessage(driver, logger))
}
