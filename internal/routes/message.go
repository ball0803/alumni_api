package routes

import (
	"alumni_api/internal/controllers"
	"alumni_api/internal/middlewares"
	"alumni_api/internal/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func MessageRoutes(group fiber.Router, driver neo4j.DriverWithContext, logger *zap.Logger) {
	group.Use("/ws", middlewares.JWTMiddleware(logger))
	group.Use("/ws", middlewares.WebSocketMiddleware(logger))
	group.Get("/ws", websocket.New(websockets.WebsocketHandler))

	msg := group.Group("/user/:user_id/message")
	msg.Use(middlewares.JWTMiddleware(logger))

	chatMsg := group.Group("/user/:user_id/chat_message")
	chatMsg.Use(middlewares.JWTMiddleware(logger))

	// Message endpoints
	msg.Post("/send", controllers.SendMessage(driver, logger))
	msg.Post("/reply", controllers.ReplyMessage(driver, logger))
	msg.Put("/:message_id", controllers.EditMessage(driver, logger))
	msg.Delete("/:message_id", controllers.DeleteMessage(driver, logger))

	chatMsg.Get("/:other_user_id", controllers.GetChatMessage(driver, logger))
}
