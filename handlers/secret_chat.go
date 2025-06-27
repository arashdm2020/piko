package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/piko/piko/crypto"
	"github.com/piko/piko/models"
	ws "github.com/piko/piko/websocket"
)

// SecretChatPool is a separate WebSocket pool for secret chats
var SecretChatPool = ws.NewPool()

func init() {
	// Start the secret chat WebSocket pool
	go SecretChatPool.Start()
}

// CreateSecretChatRequest represents a request to create a secret chat
type CreateSecretChatRequest struct {
	// No parameters needed for creation
}

// CreateSecretChatResponse represents a response to create a secret chat
type CreateSecretChatResponse struct {
	ChannelID string    `json:"channel_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// JoinSecretChatRequest represents a request to join a secret chat
type JoinSecretChatRequest struct {
	ChannelID   string `json:"channel_id"`
	DisplayName string `json:"display_name"`
}

// JoinSecretChatResponse represents a response to join a secret chat
type JoinSecretChatResponse struct {
	SessionID    string    `json:"session_id"`
	ChannelID    string    `json:"channel_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	WebSocketURL string    `json:"websocket_url"`
}

// SecretChatMessageRequest represents a request to send a message in a secret chat
type SecretChatMessageRequest struct {
	SessionID        string `json:"session_id"`
	EncryptedContent string `json:"encrypted_content"`
}

// SecretChatMessageResponse represents a message in a secret chat
type SecretChatMessageResponse struct {
	ID               string    `json:"id"`
	ChannelID        string    `json:"channel_id"`
	DisplayName      string    `json:"display_name"`
	EncryptedContent string    `json:"encrypted_content"`
	Timestamp        time.Time `json:"timestamp"`
}

// CreateSecretChat handles creating a new secret chat
func CreateSecretChat() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a new secret chat
		chat, err := models.CreateSecretChat()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create secret chat",
			})
		}

		// Return the channel ID
		return c.Status(fiber.StatusCreated).JSON(CreateSecretChatResponse{
			ChannelID: chat.ChannelID,
			ExpiresAt: chat.ExpiresAt,
		})
	}
}

// JoinSecretChat handles joining a secret chat
func JoinSecretChat() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		req := new(JoinSecretChatRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.ChannelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}
		if req.DisplayName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Display name is required",
			})
		}

		// Get chat info
		chat, err := models.GetSecretChat(req.ChannelID)
		if err != nil {
			if errors.Is(err, models.ErrSecretChatNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Secret chat not found",
				})
			}
			if errors.Is(err, models.ErrSecretChatExpired) {
				return c.Status(fiber.StatusGone).JSON(fiber.Map{
					"error": "Secret chat has expired",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get secret chat",
			})
		}

		// Join the chat
		participant, err := models.JoinSecretChat(req.ChannelID, req.DisplayName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to join secret chat",
			})
		}

		// Generate WebSocket URL
		scheme := "ws"
		if c.Protocol() == "https" {
			scheme = "wss"
		}
		wsURL := scheme + "://" + c.Hostname() + "/ws/secret/" + participant.SessionID

		// Return session info
		return c.Status(fiber.StatusOK).JSON(JoinSecretChatResponse{
			SessionID:    participant.SessionID,
			ChannelID:    participant.ChannelID,
			ExpiresAt:    chat.ExpiresAt,
			WebSocketURL: wsURL,
		})
	}
}

// SendSecretChatMessage handles sending a message in a secret chat
func SendSecretChatMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		req := new(SecretChatMessageRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.SessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Session ID is required",
			})
		}
		if req.EncryptedContent == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Encrypted content is required",
			})
		}

		// Get participant info
		participant, err := models.GetParticipant(req.SessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		// Check if chat exists and is not expired
		_, err = models.GetSecretChat(participant.ChannelID)
		if err != nil {
			if errors.Is(err, models.ErrSecretChatNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Secret chat not found",
				})
			}
			if errors.Is(err, models.ErrSecretChatExpired) {
				return c.Status(fiber.StatusGone).JSON(fiber.Map{
					"error": "Secret chat has expired",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get secret chat",
			})
		}

		// Decode encrypted content
		encryptedContent, err := crypto.DecodeBase64(req.EncryptedContent)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid encrypted content",
			})
		}

		// Generate message ID
		idBytes := make([]byte, 32)
		if _, err := rand.Read(idBytes); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate message ID",
			})
		}
		messageID := hex.EncodeToString(idBytes)

		// Create message
		message := &models.SecretChatMessage{
			ID:               messageID,
			ChannelID:        participant.ChannelID,
			SessionID:        participant.SessionID,
			DisplayName:      participant.DisplayName,
			EncryptedContent: encryptedContent,
			Timestamp:        time.Now(),
		}
		if err := models.CreateSecretChatMessage(message); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create message",
			})
		}

		// Broadcast message to all participants in the channel
		SecretChatPool.Broadcast <- ws.Message{
			Type: "secret_chat_message",
			Payload: map[string]interface{}{
				"id":                message.ID,
				"channel_id":        message.ChannelID,
				"display_name":      message.DisplayName,
				"encrypted_content": crypto.EncodeBase64(message.EncryptedContent),
				"timestamp":         message.Timestamp,
			},
			To: participant.ChannelID, // This will be used to filter recipients by channel
		}

		// Return message ID
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": messageID,
		})
	}
}

// GetSecretChatMessages handles retrieving messages from a secret chat
func GetSecretChatMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get channel ID from URL parameter
		channelID := c.Params("channel_id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Get session ID from query parameter
		sessionID := c.Query("session_id")
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Session ID is required",
			})
		}

		// Get participant info
		participant, err := models.GetParticipant(sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		// Check if participant is in the requested channel
		if participant.ChannelID != channelID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Update participant's last active timestamp
		if err := models.UpdateParticipantActivity(sessionID); err != nil {
			// Log error but continue
			// TODO: Add proper logging
		}

		// Parse pagination parameters
		limit := 50
		offset := 0
		if c.Query("limit") != "" {
			limit = c.QueryInt("limit", 50)
		}
		if c.Query("offset") != "" {
			offset = c.QueryInt("offset", 0)
		}

		// Get messages
		messages, err := models.GetSecretChatMessages(channelID, limit, offset)
		if err != nil {
			if errors.Is(err, models.ErrSecretChatNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Secret chat not found",
				})
			}
			if errors.Is(err, models.ErrSecretChatExpired) {
				return c.Status(fiber.StatusGone).JSON(fiber.Map{
					"error": "Secret chat has expired",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get messages",
			})
		}

		// Convert messages to response format
		response := make([]SecretChatMessageResponse, len(messages))
		for i, message := range messages {
			response[i] = SecretChatMessageResponse{
				ID:               message.ID,
				ChannelID:        message.ChannelID,
				DisplayName:      message.DisplayName,
				EncryptedContent: crypto.EncodeBase64(message.EncryptedContent),
				Timestamp:        message.Timestamp,
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// DeleteSecretChat handles deleting a secret chat
func DeleteSecretChat() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get channel ID from URL parameter
		channelID := c.Params("channel_id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Get session ID from query parameter
		sessionID := c.Query("session_id")
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Session ID is required",
			})
		}

		// Get participant info
		participant, err := models.GetParticipant(sessionID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session",
			})
		}

		// Check if participant is in the requested channel
		if participant.ChannelID != channelID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Delete the chat
		if err := models.DeleteSecretChat(channelID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete secret chat",
			})
		}

		// Notify all participants that the chat has been deleted
		SecretChatPool.Broadcast <- ws.Message{
			Type: "secret_chat_deleted",
			Payload: map[string]interface{}{
				"channel_id": channelID,
			},
			To: channelID,
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}

// SecretChatWebSocketHandler handles WebSocket connections for secret chats
func SecretChatWebSocketHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Get session ID from URL parameter
		sessionID := c.Params("session_id")
		if sessionID == "" {
			c.Close()
			return
		}

		// Get participant info
		participant, err := models.GetParticipant(sessionID)
		if err != nil {
			c.Close()
			return
		}

		// Create a new client
		client := &ws.Client{
			ID:      sessionID,
			Address: participant.ChannelID, // Use channel ID as the address for filtering
			Conn:    c,
			Pool:    SecretChatPool,
		}

		// Register client
		SecretChatPool.Register <- client

		// Update participant's last active timestamp
		models.UpdateParticipantActivity(sessionID)

		// Start reading messages
		client.Read()
	})
}

// CleanupExpiredSecretChats is a background task to clean up expired secret chats
func CleanupExpiredSecretChats() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		count, err := models.CleanupExpiredSecretChats()
		if err != nil {
			// TODO: Add proper logging
			continue
		}

		if count > 0 {
			// TODO: Add proper logging
		}
	}
}
