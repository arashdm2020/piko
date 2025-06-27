package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/crypto"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	RecipientAddress  string  `json:"recipient_address"`
	EncryptedContent  string  `json:"encrypted_content"`
	TTL               *int64  `json:"ttl,omitempty"` // Time to live in seconds
}

// MessageResponse represents a message response
type MessageResponse struct {
	ID               string     `json:"id"`
	SenderAddress    string     `json:"sender_address"`
	RecipientAddress string     `json:"recipient_address"`
	EncryptedContent string     `json:"encrypted_content"`
	Timestamp        time.Time  `json:"timestamp"`
	Status           string     `json:"status"`
	ExpirationTime   *time.Time `json:"expiration_time,omitempty"`
	BlockID          *string    `json:"block_id,omitempty"`
}

// SendMessage handles sending a message
func SendMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		senderAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Parse request body
		req := new(SendMessageRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.RecipientAddress == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Recipient address is required",
			})
		}
		if req.EncryptedContent == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Encrypted content is required",
			})
		}

		// Verify recipient address exists
		_, err := models.GetUserByAddress(req.RecipientAddress)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Recipient not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify recipient",
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

		// Calculate expiration time if TTL is provided
		var expirationTime *time.Time
		if req.TTL != nil && *req.TTL > 0 {
			expTime := time.Now().Add(time.Duration(*req.TTL) * time.Second)
			expirationTime = &expTime
		}

		// Create message
		message := &models.Message{
			ID:               messageID,
			SenderAddress:    senderAddress,
			RecipientAddress: req.RecipientAddress,
			EncryptedContent: encryptedContent,
			Status:           models.MessageStatusPending,
			ExpirationTime:   expirationTime,
		}
		if err := models.CreateMessage(message); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create message",
			})
		}

		// Return message ID
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": messageID,
		})
	}
}

// GetInbox handles retrieving a user's inbox
func GetInbox() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get messages from database
		messages, err := models.GetMessagesByRecipient(userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get messages",
			})
		}

		// Convert messages to response format
		response := make([]MessageResponse, len(messages))
		for i, message := range messages {
			response[i] = MessageResponse{
				ID:               message.ID,
				SenderAddress:    message.SenderAddress,
				RecipientAddress: message.RecipientAddress,
				EncryptedContent: crypto.EncodeBase64(message.EncryptedContent),
				Timestamp:        message.Timestamp,
				Status:           string(message.Status),
				ExpirationTime:   message.ExpirationTime,
				BlockID:          message.BlockID,
			}

			// Update message status to delivered if it's pending
			if message.Status == models.MessageStatusPending {
				if err := models.UpdateMessageStatus(message.ID, models.MessageStatusDelivered); err != nil {
					// Log error but continue
					// TODO: Add proper logging
				}
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// GetSentMessages handles retrieving a user's sent messages
func GetSentMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get messages from database
		messages, err := models.GetMessagesBySender(userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get messages",
			})
		}

		// Convert messages to response format
		response := make([]MessageResponse, len(messages))
		for i, message := range messages {
			response[i] = MessageResponse{
				ID:               message.ID,
				SenderAddress:    message.SenderAddress,
				RecipientAddress: message.RecipientAddress,
				EncryptedContent: crypto.EncodeBase64(message.EncryptedContent),
				Timestamp:        message.Timestamp,
				Status:           string(message.Status),
				ExpirationTime:   message.ExpirationTime,
				BlockID:          message.BlockID,
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// GetMessage handles retrieving a specific message
func GetMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get message ID from URL parameter
		messageID := c.Params("id")
		if messageID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Message ID is required",
			})
		}

		// Get message from database
		message, err := models.GetMessageByID(messageID)
		if err != nil {
			if errors.Is(err, models.ErrMessageNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Message not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get message",
			})
		}

		// Check if user is sender or recipient
		if message.SenderAddress != userAddress && message.RecipientAddress != userAddress {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Update message status to read if user is recipient and message is not read
		if message.RecipientAddress == userAddress && message.Status != models.MessageStatusRead {
			if err := models.UpdateMessageStatus(message.ID, models.MessageStatusRead); err != nil {
				// Log error but continue
				// TODO: Add proper logging
			}
			message.Status = models.MessageStatusRead
		}

		// Convert message to response format
		response := MessageResponse{
			ID:               message.ID,
			SenderAddress:    message.SenderAddress,
			RecipientAddress: message.RecipientAddress,
			EncryptedContent: crypto.EncodeBase64(message.EncryptedContent),
			Timestamp:        message.Timestamp,
			Status:           string(message.Status),
			ExpirationTime:   message.ExpirationTime,
			BlockID:          message.BlockID,
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// DeleteMessage handles deleting a message
func DeleteMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get message ID from URL parameter
		messageID := c.Params("id")
		if messageID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Message ID is required",
			})
		}

		// Get message from database
		message, err := models.GetMessageByID(messageID)
		if err != nil {
			if errors.Is(err, models.ErrMessageNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Message not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get message",
			})
		}

		// Check if user is sender or recipient
		if message.SenderAddress != userAddress && message.RecipientAddress != userAddress {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Delete message
		if err := models.DeleteMessage(messageID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete message",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Message deleted",
		})
	}
} 