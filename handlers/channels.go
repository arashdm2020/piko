package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/crypto"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// CreateChannelRequest represents a request to create a channel
type CreateChannelRequest struct {
	Name string `json:"name"`
}

// ChannelResponse represents a channel response
type ChannelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AdminAddress string `json:"admin_address"`
	CreatedAt   string `json:"created_at"`
}

// ChannelMessageRequest represents a request to send a message to a channel
type ChannelMessageRequest struct {
	EncryptedContent string `json:"encrypted_content"`
}

// ChannelMessageResponse represents a channel message response
type ChannelMessageResponse struct {
	ID              string `json:"id"`
	ChannelID       string `json:"channel_id"`
	SenderAddress   string `json:"sender_address"`
	EncryptedContent string `json:"encrypted_content"`
	Timestamp       string `json:"timestamp"`
	BlockID         string `json:"block_id,omitempty"`
}

// CreateChannel handles creating a new channel
func CreateChannel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		adminAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Parse request body
		req := new(CreateChannelRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel name is required",
			})
		}

		// Generate channel ID
		idBytes := make([]byte, 32)
		if _, err := rand.Read(idBytes); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate channel ID",
			})
		}
		channelID := hex.EncodeToString(idBytes)

		// Create channel
		channel := &models.Channel{
			ID:          channelID,
			Name:        req.Name,
			AdminAddress: adminAddress,
		}
		if err := models.CreateChannel(channel); err != nil {
			if errors.Is(err, models.ErrChannelAlreadyExists) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Channel already exists",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create channel",
			})
		}

		// Return channel ID
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": channelID,
		})
	}
}

// GetChannels handles retrieving all channels for a user
func GetChannels() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channels from database
		channels, err := models.GetChannelsByUser(userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get channels",
			})
		}

		// Convert channels to response format
		response := make([]ChannelResponse, len(channels))
		for i, channel := range channels {
			response[i] = ChannelResponse{
				ID:          channel.ID,
				Name:        channel.Name,
				AdminAddress: channel.AdminAddress,
				CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// GetChannel handles retrieving a specific channel
func GetChannel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Get channel from database
		channel, err := models.GetChannelByID(channelID)
		if err != nil {
			if errors.Is(err, models.ErrChannelNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Channel not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get channel",
			})
		}

		// Check if user is a member of the channel
		isMember, err := models.IsUserInChannel(channelID, userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check channel membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Return channel
		return c.Status(fiber.StatusOK).JSON(ChannelResponse{
			ID:          channel.ID,
			Name:        channel.Name,
			AdminAddress: channel.AdminAddress,
			CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
		})
	}
}

// UpdateChannel handles updating a channel
func UpdateChannel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Parse request body
		req := new(CreateChannelRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel name is required",
			})
		}

		// Get channel from database
		channel, err := models.GetChannelByID(channelID)
		if err != nil {
			if errors.Is(err, models.ErrChannelNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Channel not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get channel",
			})
		}

		// Check if user is the admin
		if channel.AdminAddress != userAddress {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Only the channel admin can update the channel",
			})
		}

		// Update channel
		channel.Name = req.Name
		if err := models.UpdateChannel(channel); err != nil {
			if errors.Is(err, models.ErrNotChannelAdmin) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Only the channel admin can update the channel",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update channel",
			})
		}

		// Return updated channel
		return c.Status(fiber.StatusOK).JSON(ChannelResponse{
			ID:          channel.ID,
			Name:        channel.Name,
			AdminAddress: channel.AdminAddress,
			CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
		})
	}
}

// DeleteChannel handles deleting a channel
func DeleteChannel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Delete channel
		if err := models.DeleteChannel(channelID, userAddress); err != nil {
			if errors.Is(err, models.ErrChannelNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Channel not found",
				})
			}
			if errors.Is(err, models.ErrNotChannelAdmin) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Only the channel admin can delete the channel",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete channel",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Channel deleted",
		})
	}
}

// AddChannelMember handles adding a member to a channel
func AddChannelMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		adminAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Parse request body
		type AddMemberRequest struct {
			UserAddress string `json:"user_address"`
		}
		req := new(AddMemberRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.UserAddress == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User address is required",
			})
		}

		// Verify user exists
		_, err := models.GetUserByAddress(req.UserAddress)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify user",
			})
		}

		// Add member to channel
		err = models.AddChannelMember(channelID, req.UserAddress, adminAddress)
		if err != nil {
			if errors.Is(err, models.ErrChannelNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Channel not found",
				})
			}
			if errors.Is(err, models.ErrNotChannelAdmin) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Only the channel admin can add members",
				})
			}
			if errors.Is(err, models.ErrUserAlreadyInChannel) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "User is already a member of the channel",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to add member to channel",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Member added to channel",
		})
	}
}

// RemoveChannelMember handles removing a member from a channel
func RemoveChannelMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		adminAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Get user address from URL parameter
		userAddress := c.Params("address")
		if userAddress == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User address is required",
			})
		}

		// Remove member from channel
		err := models.RemoveChannelMember(channelID, userAddress, adminAddress)
		if err != nil {
			if errors.Is(err, models.ErrChannelNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Channel not found",
				})
			}
			if errors.Is(err, models.ErrNotChannelAdmin) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Only the channel admin can remove members",
				})
			}
			if errors.Is(err, models.ErrUserNotInChannel) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User is not a member of the channel",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to remove member from channel",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Member removed from channel",
		})
	}
}

// GetChannelMembers handles retrieving all members of a channel
func GetChannelMembers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Check if user is a member of the channel
		isMember, err := models.IsUserInChannel(channelID, userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check channel membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Get channel members
		members, err := models.GetChannelMembers(channelID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get channel members",
			})
		}

		// Convert members to response format
		type MemberResponse struct {
			UserAddress string `json:"user_address"`
			JoinedAt    string `json:"joined_at"`
		}
		response := make([]MemberResponse, len(members))
		for i, member := range members {
			response[i] = MemberResponse{
				UserAddress: member.UserAddress,
				JoinedAt:    member.JoinedAt.Format(time.RFC3339),
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// SendChannelMessage handles sending a message to a channel
func SendChannelMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		senderAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Parse request body
		req := new(ChannelMessageRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.EncryptedContent == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Encrypted content is required",
			})
		}

		// Check if user is a member of the channel
		isMember, err := models.IsUserInChannel(channelID, senderAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check channel membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
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

		// Create channel message
		message := &models.ChannelMessage{
			ID:              messageID,
			ChannelID:       channelID,
			SenderAddress:   senderAddress,
			EncryptedContent: encryptedContent,
		}
		if err := models.CreateChannelMessage(message); err != nil {
			if errors.Is(err, models.ErrUserNotInChannel) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "User is not a member of the channel",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create channel message",
			})
		}

		// Return message ID
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": messageID,
		})
	}
}

// GetChannelMessages handles retrieving messages from a channel
func GetChannelMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Check if user is a member of the channel
		isMember, err := models.IsUserInChannel(channelID, userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check channel membership",
			})
		}
		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Get pagination parameters
		limit := 50
		offset := 0
		if c.Query("limit") != "" {
			limit, err = strconv.Atoi(c.Query("limit"))
			if err != nil || limit <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid limit parameter",
				})
			}
		}
		if c.Query("offset") != "" {
			offset, err = strconv.Atoi(c.Query("offset"))
			if err != nil || offset < 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid offset parameter",
				})
			}
		}

		// Get channel messages
		messages, err := models.GetChannelMessages(channelID, limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get channel messages",
			})
		}

		// Convert messages to response format
		response := make([]ChannelMessageResponse, len(messages))
		for i, message := range messages {
			response[i] = ChannelMessageResponse{
				ID:              message.ID,
				ChannelID:       message.ChannelID,
				SenderAddress:   message.SenderAddress,
				EncryptedContent: crypto.EncodeBase64(message.EncryptedContent),
				Timestamp:       message.Timestamp.Format(time.RFC3339),
			}
			if message.BlockID != nil {
				response[i].BlockID = *message.BlockID
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// DeleteChannelMessage handles deleting a channel message
func DeleteChannelMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get channel ID from URL parameter
		channelID := c.Params("channel_id")
		if channelID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Channel ID is required",
			})
		}

		// Get message ID from URL parameter
		messageID := c.Params("message_id")
		if messageID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Message ID is required",
			})
		}

		// Delete channel message
		if err := models.DeleteChannelMessage(messageID, userAddress); err != nil {
			if errors.Is(err, models.ErrMessageNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Message not found",
				})
			}
			if errors.Is(err, models.ErrNotChannelAdmin) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Only the channel admin or message sender can delete the message",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete message",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Message deleted",
		})
	}
} 