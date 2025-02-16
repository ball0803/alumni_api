package websockets

import (
	"alumni_api/internal/models"
	"encoding/json"
	"github.com/gofiber/websocket/v2"
	"log"
	"sync"
)

// Map to store active connections
var Clients = make(map[string]*websocket.Conn)
var Mutex = sync.Mutex{}

func WebsocketHandler(c *websocket.Conn) {
	defer c.Close()

	claims, ok := c.Locals("claims").(*models.Claims)
	if !ok {
		return
	}

	Mutex.Lock()
	Clients[claims.UserID] = c
	Mutex.Unlock()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("WebSocket error:", err)
			break
		}
		log.Printf("Received from %s: %s", claims.UserID, msg)
	}

	Mutex.Lock()
	delete(Clients, claims.UserID)
	Mutex.Unlock()
}

// SendWebSocketNotification sends a JSON message to a connected user
func SendNotification(userID string, message interface{}) {
	Mutex.Lock()
	conn, online := Clients[userID]
	Mutex.Unlock()

	if online {
		jsonMessage, _ := json.Marshal(message)
		_ = conn.WriteMessage(websocket.TextMessage, jsonMessage)
	}
}
