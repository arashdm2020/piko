package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/crypto"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
	"github.com/piko/piko/websocket"
)

// CreateGroupRequest represents a request to create a group
type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoURL    string `json:"photo_url,omitempty"`
}

// GroupResponse represents a group response
type GroupResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoURL    string `json:"photo_url,omitempty"`
	CreatedBy   string `json:"created_by"`
	MemberCount int    `json:"member_count"`
}

// GroupMemberResponse represents a group member response
type GroupMemberResponse struct {
	UserAddress string `json:"user_address"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}

// SendGroupMessageRequest represents a request to send a message to a group
type SendGroupMessageRequest struct {
	Content string `json:"content"`
}

// GroupMessageResponse represents a group message response
type GroupMessageResponse struct {
	ID            string `json:"id"`
	GroupID       string `json:"group_id"`
	SenderAddress string `json:"sender_address"`
	Content       string `json:"content"`
	Timestamp     string `json:"timestamp"`
}

// CreateGroup handles creating a new group
func CreateGroup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Parse request body
		req := new(CreateGroupRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group name is required",
			})
		}

		// Generate group ID
		idBytes := make([]byte, 32)
		if _, err := rand.Read(idBytes); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate group ID",
			})
		}
		groupID := hex.EncodeToString(idBytes)

		// Create group
		group := &models.Group{
			ID:             groupID,
			Name:           req.Name,
			Description:    req.Description,
			CreatorAddress: userAddress,
			PhotoURL:       req.PhotoURL,
		}
		if err := models.CreateGroup(group, userAddress); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create group",
			})
		}

		// Return group ID
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": groupID,
		})
	}
}

// GetGroups handles retrieving all groups a user is a member of
func GetGroups() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get groups from database
		groups, err := models.GetUserGroups(userAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get groups",
			})
		}

		// Convert groups to response format
		response := make([]GroupResponse, len(groups))
		for i, group := range groups {
			response[i] = GroupResponse{
				ID:          group.ID,
				Name:        group.Name,
				Description: group.Description,
				PhotoURL:    group.PhotoURL,
				CreatedBy:   group.CreatorAddress,
				MemberCount: group.MemberCount,
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// GetGroup handles retrieving a specific group
func GetGroup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Check if user is a member of the group
		members, err := models.GetGroupMembers(groupID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group members",
			})
		}

		isMember := false
		for _, member := range members {
			if member.UserAddress == userAddress {
				isMember = true
				break
			}
		}

		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not a member of this group",
			})
		}

		// Get group from database
		group, err := models.GetGroupByID(groupID)
		if err != nil {
			if errors.Is(err, models.ErrGroupNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Group not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group",
			})
		}

		// Return group
		return c.Status(fiber.StatusOK).JSON(GroupResponse{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			PhotoURL:    group.PhotoURL,
			CreatedBy:   group.CreatorAddress,
			MemberCount: group.MemberCount,
		})
	}
}

// UpdateGroup handles updating a group's information
func UpdateGroup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Check if user is an admin of the group
		isAdmin, err := models.IsGroupAdmin(groupID, userAddress)
		if err != nil {
			if errors.Is(err, models.ErrGroupMemberNotFound) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You are not a member of this group",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check admin status",
			})
		}

		if !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not an admin of this group",
			})
		}

		// Get group from database
		group, err := models.GetGroupByID(groupID)
		if err != nil {
			if errors.Is(err, models.ErrGroupNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Group not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group",
			})
		}

		// Parse request body
		req := new(CreateGroupRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Update group fields
		if req.Name != "" {
			group.Name = req.Name
		}
		group.Description = req.Description
		group.PhotoURL = req.PhotoURL

		// Save changes
		if err := models.UpdateGroup(group); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update group",
			})
		}

		// Return updated group
		return c.Status(fiber.StatusOK).JSON(GroupResponse{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			PhotoURL:    group.PhotoURL,
			CreatedBy:   group.CreatorAddress,
			MemberCount: group.MemberCount,
		})
	}
}

// DeleteGroup handles deleting a group
func DeleteGroup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Get group from database
		group, err := models.GetGroupByID(groupID)
		if err != nil {
			if errors.Is(err, models.ErrGroupNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Group not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group",
			})
		}

		// Check if user is the creator of the group
		if group.CreatorAddress != userAddress {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Only the creator can delete the group",
			})
		}

		// Delete group
		if err := models.DeleteGroup(groupID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete group",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Group deleted successfully",
		})
	}
}

// AddGroupMember handles adding a member to a group
func AddGroupMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Check if user is an admin of the group
		isAdmin, err := models.IsGroupAdmin(groupID, userAddress)
		if err != nil {
			if errors.Is(err, models.ErrGroupMemberNotFound) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You are not a member of this group",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check admin status",
			})
		}

		if !isAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not an admin of this group",
			})
		}

		// Parse request body
		var req struct {
			UserAddress string `json:"user_address"`
			IsAdmin     bool   `json:"is_admin"`
		}
		if err := c.BodyParser(&req); err != nil {
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

		// Check if user exists
		_, err = models.GetUserByAddress(req.UserAddress)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check user",
			})
		}

		// Determine role
		role := models.GroupRoleMember
		if req.IsAdmin {
			role = models.GroupRoleAdmin
		}

		// Add member to group
		err = models.AddGroupMember(groupID, req.UserAddress, role)
		if err != nil {
			if errors.Is(err, models.ErrAlreadyGroupMember) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "User is already a member of this group",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to add member",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Member added successfully",
		})
	}
}

// RemoveGroupMember handles removing a member from a group
func RemoveGroupMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID and member address from URL parameters
		groupID := c.Params("id")
		memberAddress := c.Params("address")
		if groupID == "" || memberAddress == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID and member address are required",
			})
		}

		// Check if user is an admin of the group or is removing themselves
		if userAddress != memberAddress {
			isAdmin, err := models.IsGroupAdmin(groupID, userAddress)
			if err != nil {
				if errors.Is(err, models.ErrGroupMemberNotFound) {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "You are not a member of this group",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to check admin status",
				})
			}

			if !isAdmin {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "You are not an admin of this group",
				})
			}
		}

		// Remove member from group
		err := models.RemoveGroupMember(groupID, memberAddress)
		if err != nil {
			if errors.Is(err, models.ErrGroupMemberNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Member not found in group",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to remove member",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Member removed successfully",
		})
	}
}

// GetGroupMembers handles retrieving all members of a group
func GetGroupMembers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Check if user is a member of the group
		members, err := models.GetGroupMembers(groupID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group members",
			})
		}

		isMember := false
		for _, member := range members {
			if member.UserAddress == userAddress {
				isMember = true
				break
			}
		}

		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not a member of this group",
			})
		}

		// Convert members to response format
		response := make([]GroupMemberResponse, len(members))
		for i, member := range members {
			response[i] = GroupMemberResponse{
				UserAddress: member.UserAddress,
				Role:        string(member.Role),
				JoinedAt:    member.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// SendGroupMessage handles sending a message to a group
func SendGroupMessage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Check if user is a member of the group
		members, err := models.GetGroupMembers(groupID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group members",
			})
		}

		isMember := false
		for _, member := range members {
			if member.UserAddress == userAddress {
				isMember = true
				break
			}
		}

		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not a member of this group",
			})
		}

		// Parse request body
		req := new(SendGroupMessageRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Content == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Content is required",
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
		content, err := crypto.DecodeBase64(req.Content)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid content encoding",
			})
		}

		message := &models.GroupMessage{
			ID:            messageID,
			GroupID:       groupID,
			SenderAddress: userAddress,
			Content:       content,
		}

		// Save message to database
		if err := models.CreateGroupMessage(message); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create message",
			})
		}

		// Notify group members via WebSocket
		go notifyGroupMessage(groupID, message)

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": messageID,
		})
	}
}

// GetGroupMessages handles retrieving messages from a group
func GetGroupMessages() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user address from context
		userAddress, ok := middleware.GetUserAddress(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get group ID from URL parameter
		groupID := c.Params("id")
		if groupID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Group ID is required",
			})
		}

		// Check if user is a member of the group
		members, err := models.GetGroupMembers(groupID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get group members",
			})
		}

		isMember := false
		for _, member := range members {
			if member.UserAddress == userAddress {
				isMember = true
				break
			}
		}

		if !isMember {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not a member of this group",
			})
		}

		// Get pagination parameters
		limit := 20
		offset := 0
		if c.Query("limit") != "" {
			l, err := strconv.Atoi(c.Query("limit"))
			if err == nil && l > 0 {
				limit = l
			}
		}
		if c.Query("offset") != "" {
			o, err := strconv.Atoi(c.Query("offset"))
			if err == nil && o >= 0 {
				offset = o
			}
		}

		// Get messages from database
		messages, err := models.GetGroupMessages(groupID, limit, offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get messages",
			})
		}

		// Convert messages to response format
		response := make([]GroupMessageResponse, len(messages))
		for i, message := range messages {
			response[i] = GroupMessageResponse{
				ID:            message.ID,
				GroupID:       message.GroupID,
				SenderAddress: message.SenderAddress,
				Content:       crypto.EncodeBase64(message.Content),
				Timestamp:     message.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// notifyGroupMessage notifies all group members about a new message
func notifyGroupMessage(groupID string, message *models.GroupMessage) {
	// Get group members
	members, err := models.GetGroupMembers(groupID)
	if err != nil {
		return
	}

	// Notify all online members except the sender
	for _, member := range members {
		if member.UserAddress == message.SenderAddress {
			continue
		}

		WebSocketPool.Broadcast <- websocket.Message{
			Type: "new_group_message",
			Payload: map[string]interface{}{
				"id":             message.ID,
				"group_id":       message.GroupID,
				"sender_address": message.SenderAddress,
			},
			To: member.UserAddress,
		}
	}
}
