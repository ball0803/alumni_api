package process

import (
	"alumni_api/models"
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func SendMessage(ctx context.Context, driver neo4j.DriverWithContext, msg models.Message, logger *zap.Logger) (map[string]interface{}, error) {
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
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to send message")
	}

	messageData := map[string]interface{}{
		"message_id": msg.MessageID,
		"content":    msg.Content.Value,
		"sender_id":  msg.SenderID,
	}

	if result.Next(ctx) {
		record := result.Record()
		if senderUsername, ok := record.Get("sender_username"); ok {
			messageData["sender_username"] = senderUsername
		}

		if senderFullname, ok := record.Get("sender_fullname"); ok && senderFullname != nil {
			messageData["sender_fullname"] = senderFullname
		}

		if senderPicture, ok := record.Get("sender_picture"); ok && senderPicture != nil {
			messageData["sender_picture"] = senderPicture
		}

		if timestamp, ok := record.Get("timestamp"); ok {
			messageData["timestamp"] = timestamp
		}
	}

	return messageData, nil
}

func ReplyMessage(ctx context.Context, driver neo4j.DriverWithContext, msg models.ReplyMessage, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	msg.MessageID = uuid.New().String()

	query := `
    MATCH 
      (s:UserProfile {user_id: $sender}),
      (r:UserProfile {user_id: $receiver}),
      (rm:Message {message_id: $reply_id})
    CREATE
      (s)-[:SENT]->(m:Message {
        message_id: $message_id,
        content: $content,
        created_timestamp: timestamp()
      })<-[:RECEIVED]-(r),
      (m)-[:REPLIED]->(rm)
    RETURN
      m,
      rm.content AS reply_content,
      s.username AS sender_username,
      s.first_name + " " + s.last_name AS sender_fullname,
      s.profile_picture AS sender_picture,
      m.created_timestamp AS timestamp
  `

	params := map[string]interface{}{
		"message_id": msg.MessageID,
		"reply_id":   msg.ReplyID,
		"sender":     msg.SenderID,
		"receiver":   msg.ReceiverID,
		"content":    msg.Content.Raw,
	}

	result, err := session.Run(ctx, query, params)

	if err != nil {
		logger.Error("Failed to send message", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to send message")
	}

	messageData := map[string]interface{}{
		"message_id":       msg.MessageID,
		"reply_message_id": msg.ReplyID,
		"content":          msg.Content.Value,
		"sender_id":        msg.SenderID,
	}

	if result.Next(ctx) {
		record := result.Record()
		if senderUsername, ok := record.Get("sender_username"); ok {
			messageData["sender_username"] = senderUsername
		}

		if replyContent, ok := record.Get("reply_content"); ok {
			messageData["reply_content"] = replyContent
		}

		if senderFullname, ok := record.Get("sender_fullname"); ok && senderFullname != nil {
			messageData["sender_fullname"] = senderFullname
		}

		if senderPicture, ok := record.Get("sender_picture"); ok && senderPicture != nil {
			messageData["sender_picture"] = senderPicture
		}

		if timestamp, ok := record.Get("timestamp"); ok {
			messageData["timestamp"] = timestamp
		}
	}

	return messageData, nil
}

func EditMessage(ctx context.Context, driver neo4j.DriverWithContext, msg models.EditMessage, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
        MATCH (m:Message {message_id: $message_id})
        SET
          m.content = $content,
          m.updated_timestamp = timestamp()
        RETURN m
    `

	params := map[string]interface{}{
		"message_id": msg.MessageID,
		"content":    msg.Content.Raw,
	}

	_, err := session.Run(ctx, query, params)

	if err != nil {
		logger.Error("Failed to edit message", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to edit message")
	}

	return nil
}

func DeleteMessage(ctx context.Context, driver neo4j.DriverWithContext, msg models.DeleteMessage, logger *zap.Logger) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	msg.MessageID = uuid.New().String()

	query := `
        MATCH (m:Message {message_id: $message_id})
        DETACH DELETE m
    `

	params := map[string]interface{}{
		"message_id": msg.MessageID,
	}

	_, err := session.Run(ctx, query, params)

	if err != nil {
		logger.Error("Failed to delete message", zap.Error(err))
		return fiber.NewError(http.StatusInternalServerError, "Failed to delete message")
	}

	return nil
}

func GetMessage(ctx context.Context, driver neo4j.DriverWithContext, sender_id, receiver_id string, logger *zap.Logger) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: "neo4j",
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
    MATCH (s:UserProfile)-[:SENT]->(m:Message)<-[:RECEIVED]-(r:UserProfile)
    OPTIONAL MATCH (m)-[:REPLIED]->(rm:Message)  // Find the replied message (if exists)
    WHERE (s.user_id = $sender_id AND r.user_id = $receiver_id)
      OR (s.user_id = $receiver_id AND r.user_id = $sender_id)
    WITH 
      COLLECT(
        CASE
          WHEN s.user_id = $sender_id THEN {
            id: s.user_id,
            username: s.username,
            name: s.first_name + " " + s.last_name,
            picture: s.profile_picture,
            message: {
              message_id: m.message_id,
              content: m.content,
              created_timestamp: m.created_timestamp,
              update_timestamp: m.updated_timestamp,
              reply_message_id: rm.message_id,
              reply_message_content: rm.content
            }
          }
        END
      ) AS me,
      COLLECT(
        CASE
          WHEN s.user_id = $receiver_id THEN {
            id: s.user_id,
            username: s.username,
            name: s.first_name + " " + s.last_name,
            picture: s.profile_picture,
            message: {
              message_id: m.message_id,
              content: m.content,
              created_timestamp: m.created_timestamp,
              update_timestamp: m.updated_timestamp,
              reply_message_id: rm.message_id,
              reply_message_content: rm.content
            }
          }
        END
      ) AS other
    RETURN me, other
    `

	params := map[string]interface{}{
		"receiver_id": receiver_id,
		"sender_id":   sender_id,
	}

	result, err := session.Run(ctx, query, params)

	if err != nil {
		logger.Error("Failed to send message", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to send message")
	}

	record, err := result.Single(ctx)
	if err != nil {
		logger.Error("Failed to collect results", zap.Error(err))
		return nil, fiber.NewError(http.StatusInternalServerError, "Failed to collect results")
	}

	messageData := record.AsMap()

	return messageData, nil
}
