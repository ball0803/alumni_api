package process

import (
	"alumni_api/models"
	"alumni_api/websockets"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
	"net/http"
)

func SendMessage(ctx context.Context, driver neo4j.DriverWithContext, msg models.Message, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	msg.MessageID = uuid.New().String()

	query := `
        MATCH (s:UserProfile {user_id: $sender}), (r:UserProfile {user_id: $receiver})
        CREATE (s)-[:SENT]->(m:Message {
          message_id: $message_id,
          content: $content,
          created_timestamp: timestamp()
        })<-[:RECEIVED]-(r)
        RETURN 
        m, s.username AS sender_username,
        s.first_name + " " + s.last_name AS sender_fullname,
        s.profile_picture AS sender_picture,
        m.created_timestamp AS timestamp
    `

	params := map[string]interface{}{
		"message_id": msg.MessageID,
		"sender":     msg.SenderID,
		"receiver":   msg.ReceiverID,
		"content":    msg.Content.Raw,
	}

	result, err := session.Run(ctx, query, params)

	if err != nil {
		logger.Error("Failed to send message", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to send message")
	}

	messageData := map[string]interface{}{
		"message_id": msg.MessageID,
		"content":    msg.Content.Value,
		"sender_id":  msg.SenderID,
	}

	if result.Next(ctx) {
		record := result.Record()
		if senderUsername, ok := record.Get("sender_username"); ok {
			messageData["sender_name"] = senderUsername
		}

		if senderFullname, ok := record.Get("sender_fullname"); ok && senderFullname != nil {
			messageData["sender_picture"] = senderFullname
		}

		if senderPicture, ok := record.Get("sender_picture"); ok && senderPicture != nil {
			messageData["sender_picture"] = senderPicture
		}

		if timestamp, ok := record.Get("timestamp"); ok {
			messageData["timestamp"] = timestamp
		}
	}

	jsonMessage, _ := json.Marshal(messageData)

	// Notify receiver via WebSocket if online
	websockets.Mutex.Lock()
	conn, online := websockets.Clients[msg.ReceiverID]
	websockets.Mutex.Unlock()

	if online {
		_ = conn.WriteMessage(websocket.TextMessage, jsonMessage)
	}

	return nil
}
