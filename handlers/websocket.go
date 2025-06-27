package handlers

import (
	"github.com/gofiber/fiber/v2"
	wsfiber "github.com/gofiber/websocket/v2"
	"github.com/piko/piko/websocket"
)

var (
	// WebSocketPool is the global WebSocket pool
	WebSocketPool = websocket.NewPool()
)

func init() {
	// Start the WebSocket pool
	go WebSocketPool.Start()
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler() fiber.Handler {
	return wsfiber.New(func(c *wsfiber.Conn) {
		// Get user address from query parameter
		address := c.Query("address")
		if address == "" {
			c.Close()
			return
		}

		// Get token from query parameter
		token := c.Query("token")
		if token == "" {
			c.Close()
			return
		}

		// TODO: Validate token and address
		// For now, we'll trust the client

		// Create a new client
		client := &websocket.Client{
			Address: address,
			Conn:    c,
			Pool:    WebSocketPool,
		}

		// Register client
		WebSocketPool.Register <- client

		// Start reading messages
		client.Read()
	})
}
