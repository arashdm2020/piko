package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// SearchUsersRequest represents a request to search for users
type SearchUsersRequest struct {
	Query string `json:"query"`
}

// UserResponse represents a user response with minimal information
type UserResponse struct {
	Address  string `json:"address"`
	Username string `json:"username,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// SetUsernameRequest represents a request to set or update a username
type SetUsernameRequest struct {
	Username string `json:"username"`
}

// SearchUsers handles searching for users by address or other identifiers
func SearchUsers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context to ensure the requester is authenticated
		_, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get query from request
		query := c.Query("query")
		if query == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Search query is required",
			})
		}

		// Search for users by address, phone, or username
		users, err := models.SearchUsers(query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to search users",
			})
		}

		// Convert users to response format
		response := make([]UserResponse, len(users))
		for i, user := range users {
			response[i] = UserResponse{
				Address:  user.Address,
				Username: user.Username,
				Phone:    maskPhone(user.Phone),
			}
		}

		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// GetUser handles retrieving a user by their address
func GetUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context to ensure the requester is authenticated
		_, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get address from URL parameter
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Address is required",
			})
		}

		// Get user by address
		user, err := models.GetUserByAddress(address)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user",
			})
		}

		// Return user with masked sensitive information
		return c.Status(fiber.StatusOK).JSON(UserResponse{
			Address:  user.Address,
			Username: user.Username,
			Phone:    maskPhone(user.Phone),
		})
	}
}

// SetUsername handles setting or updating a user's username
func SetUsername() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context to ensure the requester is authenticated
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Parse request body
		req := new(SetUsernameRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate username
		if req.Username == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username is required",
			})
		}

		// Set username
		err := models.SetUsername(userID, req.Username)
		if err != nil {
			if errors.Is(err, models.ErrInvalidUsername) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid username format. Username must be 3-30 characters long and contain only alphanumeric characters and underscores.",
				})
			}
			if errors.Is(err, models.ErrUsernameAlreadyExists) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Username already exists",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to set username",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":  "Username set successfully",
			"username": req.Username,
		})
	}
}

// Helper functions to mask sensitive information
func maskPhone(phone string) string {
	if phone == "" {
		return ""
	}
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
