package websockets

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// Map to store active connections
var Clients = make(map[string]*websocket.Conn)
var Mutex = sync.Mutex{}

func WebsocketHandler(c *websocket.Conn) {
	defer c.Close()

	userID := c.Query("user_id") // Get user ID from query param
	Mutex.Lock()
	Clients[userID] = c
	Mutex.Unlock()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("WebSocket error:", err)
			break
		}
		log.Printf("Received from %s: %s", userID, msg)
	}

	Mutex.Lock()
	delete(Clients, userID)
	Mutex.Unlock()
}
