package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/piko/piko/models"
)

// Client represents a WebSocket client
type Client struct {
	ID      string
	Address string
	Conn    *websocket.Conn
	Pool    *Pool
	mu      sync.Mutex
}

// Pool represents a pool of WebSocket clients
type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[string]*Client
	Broadcast  chan Message
	mu         sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	From    string                 `json:"from,omitempty"`
	To      string                 `json:"to,omitempty"`
}

// MessageStatus represents the status of a message
type MessageStatus string

const (
	// StatusSent indicates the message was sent to the server
	StatusSent MessageStatus = "sent"
	// StatusDelivered indicates the message was delivered to the recipient
	StatusDelivered MessageStatus = "delivered"
	// StatusRead indicates the message was read by the recipient
	StatusRead MessageStatus = "read"
)

// MessageType constants
const (
	// MessageTypeNewMessage is sent when a new direct message is received
	MessageTypeNewMessage = "new_message"

	// MessageTypeMessageStatus is sent when a message status changes
	MessageTypeMessageStatus = "message_status"

	// MessageTypeNewChannelMessage is sent when a new channel message is received
	MessageTypeNewChannelMessage = "new_channel_message"

	// MessageTypeNewGroupMessage is sent when a new group message is received
	MessageTypeNewGroupMessage = "new_group_message"
)

// NewPool creates a new WebSocket pool
func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan Message),
	}
}

// Start starts the WebSocket pool
func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.mu.Lock()
			pool.Clients[client.Address] = client
			pool.mu.Unlock()
			log.Printf("Client connected: %s", client.Address)

			// Send presence update to all clients
			pool.Broadcast <- Message{
				Type: "presence",
				Payload: map[string]interface{}{
					"address": client.Address,
					"status":  "online",
				},
			}

			// Send welcome message to client
			client.SendMessage(Message{
				Type: "welcome",
				Payload: map[string]interface{}{
					"message": "Welcome to Piko!",
				},
			})

			// Mark all pending messages as delivered
			go func() {
				// Get all pending messages for this client
				messages, err := models.GetMessagesByRecipient(client.Address)
				if err != nil {
					log.Printf("Error getting messages for %s: %v", client.Address, err)
					return
				}

				for _, msg := range messages {
					if msg.Status == models.MessageStatusPending {
						// Update status to delivered
						if err := models.UpdateMessageStatus(msg.ID, models.MessageStatusDelivered); err != nil {
							log.Printf("Error updating message status: %v", err)
							continue
						}

						// Notify sender about delivery
						pool.mu.RLock()
						senderClient, senderOnline := pool.Clients[msg.SenderAddress]
						pool.mu.RUnlock()

						if senderOnline {
							senderClient.SendMessage(Message{
								Type: "status_update",
								Payload: map[string]interface{}{
									"message_id": msg.ID,
									"status":     "delivered",
									"recipient":  client.Address,
									"timestamp":  time.Now().Format(time.RFC3339),
								},
							})
						}
					}
				}
			}()

		case client := <-pool.Unregister:
			pool.mu.Lock()
			delete(pool.Clients, client.Address)
			pool.mu.Unlock()
			log.Printf("Client disconnected: %s", client.Address)

			// Send presence update to all clients
			pool.Broadcast <- Message{
				Type: "presence",
				Payload: map[string]interface{}{
					"address": client.Address,
					"status":  "offline",
				},
			}

		case message := <-pool.Broadcast:
			// If message has a specific recipient, send only to that client
			if message.To != "" {
				pool.mu.RLock()
				client, ok := pool.Clients[message.To]
				pool.mu.RUnlock()
				if ok {
					client.SendMessage(message)
				}
			} else {
				// Otherwise, broadcast to all clients
				pool.mu.RLock()
				for _, client := range pool.Clients {
					client.SendMessage(message)
				}
				pool.mu.RUnlock()
			}
		}
	}
}

// SendMessage sends a message to a client
func (client *Client) SendMessage(message Message) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if err := client.Conn.WriteJSON(message); err != nil {
		log.Printf("Error sending message to client %s: %v", client.Address, err)
	}
}

// Read reads messages from a client
func (client *Client) Read() {
	defer func() {
		client.Pool.Unregister <- client
		client.Conn.Close()
	}()

	for {
		messageType, p, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %v", client.Address, err)
			return
		}

		// Handle different message types
		if messageType == websocket.TextMessage {
			var message Message
			if err := json.Unmarshal(p, &message); err != nil {
				log.Printf("Error unmarshaling message from client %s: %v", client.Address, err)
				continue
			}

			// Handle message based on type
			switch message.Type {
			case "ping":
				// Respond with pong
				client.SendMessage(Message{
					Type:    "pong",
					Payload: map[string]interface{}{"time": time.Now().Format(time.RFC3339)},
				})

			case "typing":
				// Handle typing indicator
				if to, ok := message.Payload["to"].(string); ok {
					// Forward typing indicator to recipient
					client.Pool.Broadcast <- Message{
						Type: "typing",
						Payload: map[string]interface{}{
							"from": client.Address,
						},
						To: to,
					}
				}

			case "read":
				// Handle message read status
				if messageID, ok := message.Payload["message_id"].(string); ok {
					// Update message status in database
					if err := models.UpdateMessageStatus(messageID, models.MessageStatusRead); err != nil {
						log.Printf("Error updating message status: %v", err)
					} else {
						// Get message to find sender
						msg, err := models.GetMessageByID(messageID)
						if err == nil {
							// Notify sender that message was read
							client.Pool.Broadcast <- Message{
								Type: "status_update",
								Payload: map[string]interface{}{
									"message_id": messageID,
									"status":     "read",
									"recipient":  client.Address,
									"timestamp":  time.Now().Format(time.RFC3339),
								},
								To: msg.SenderAddress,
							}
						}
					}
				}

			case "received":
				// Handle message received status (client acknowledges receipt)
				if messageID, ok := message.Payload["message_id"].(string); ok {
					// Update message status in database
					if err := models.UpdateMessageStatus(messageID, models.MessageStatusDelivered); err != nil {
						log.Printf("Error updating message status: %v", err)
					} else {
						// Get message to find sender
						msg, err := models.GetMessageByID(messageID)
						if err == nil {
							// Notify sender that message was delivered
							client.Pool.Broadcast <- Message{
								Type: "status_update",
								Payload: map[string]interface{}{
									"message_id": messageID,
									"status":     "delivered",
									"recipient":  client.Address,
									"timestamp":  time.Now().Format(time.RFC3339),
								},
								To: msg.SenderAddress,
							}
						}
					}
				}

			default:
				// Ignore unknown message types
				log.Printf("Unknown message type from client %s: %s", client.Address, message.Type)
			}
		}
	}
}

// NotifyNewMessage notifies a client about a new message
func NotifyNewMessage(pool *Pool, message *models.Message) {
	// Check if recipient is connected
	pool.mu.RLock()
	client, ok := pool.Clients[message.RecipientAddress]
	pool.mu.RUnlock()

	if ok {
		// Send notification to recipient
		client.SendMessage(Message{
			Type: "new_message",
			Payload: map[string]interface{}{
				"id":             message.ID,
				"sender_address": message.SenderAddress,
			},
		})

		// Update message status to delivered
		if err := models.UpdateMessageStatus(message.ID, models.MessageStatusDelivered); err != nil {
			log.Printf("Error updating message status: %v", err)
		} else {
			// Notify sender about delivery
			pool.mu.RLock()
			senderClient, senderOnline := pool.Clients[message.SenderAddress]
			pool.mu.RUnlock()

			if senderOnline {
				senderClient.SendMessage(Message{
					Type: "status_update",
					Payload: map[string]interface{}{
						"message_id": message.ID,
						"status":     "delivered",
						"recipient":  message.RecipientAddress,
						"timestamp":  time.Now().Format(time.RFC3339),
					},
				})
			}
		}
	} else {
		// Recipient is offline, message stays in pending state
		log.Printf("Recipient %s is offline, message %s is pending", message.RecipientAddress, message.ID)
	}
}

// NotifyNewChannelMessage notifies clients about a new channel message
func NotifyNewChannelMessage(pool *Pool, message *models.ChannelMessage) {
	// Get channel members
	members, err := models.GetChannelMembers(message.ChannelID)
	if err != nil {
		log.Printf("Error getting channel members: %v", err)
		return
	}

	// Notify all online members except the sender
	for _, member := range members {
		if member.UserAddress == message.SenderAddress {
			continue
		}

		pool.mu.RLock()
		client, ok := pool.Clients[member.UserAddress]
		pool.mu.RUnlock()
		if ok {
			client.SendMessage(Message{
				Type: "new_channel_message",
				Payload: map[string]interface{}{
					"id":             message.ID,
					"channel_id":     message.ChannelID,
					"sender_address": message.SenderAddress,
				},
			})
		}
	}
}

// GetOnlineUsers returns a list of online users
func GetOnlineUsers(pool *Pool) []string {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	users := make([]string, 0, len(pool.Clients))
	for address := range pool.Clients {
		users = append(users, address)
	}
	return users
}

// IsUserOnline checks if a user is online
func IsUserOnline(pool *Pool, address string) bool {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	_, ok := pool.Clients[address]
	return ok
}
