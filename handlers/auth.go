package handlers

import (
	"encoding/base64"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/config"
	"github.com/piko/piko/crypto"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token   string `json:"token"`
	Address string `json:"address"`
}

// Register handles user registration
func Register(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		req := new(RegisterRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Password is required",
			})
		}
		if req.Email == "" && req.Phone == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email or phone number is required",
			})
		}

		// Generate key pair
		keyPair, err := crypto.GenerateKeyPair()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate key pair",
			})
		}

		// Generate user address
		address, err := crypto.GenerateAddress(keyPair.PublicKey, cfg.Crypto.AddressLength)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate address",
			})
		}

		// Hash password
		passwordHash, err := crypto.HashPassword(
			req.Password,
			nil,
			cfg.Auth.Argon2Time,
			cfg.Auth.Argon2Memory,
			cfg.Auth.Argon2Threads,
			cfg.Auth.Argon2KeyLength,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash password",
			})
		}

		// Create user
		user := &models.User{
			Email:        req.Email,
			Phone:        req.Phone,
			PasswordHash: base64.StdEncoding.EncodeToString(passwordHash),
			PublicKey:    keyPair.PublicKey,
			Address:      address,
		}
		err = models.CreateUser(user)
		if err != nil {
			if errors.Is(err, models.ErrEmailAlreadyExists) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Email already exists",
				})
			}
			if errors.Is(err, models.ErrPhoneAlreadyExists) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Phone number already exists",
				})
			}
			if errors.Is(err, models.ErrAddressAlreadyExists) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Address already exists",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		// Generate JWT token
		token, err := middleware.GenerateJWT(user, cfg.Auth.JWTSecret, cfg.Auth.JWTExpirationTime)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		// Return private key and token
		// IMPORTANT: Private key is only returned once during registration
		// Client must store it securely
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"token":       token,
			"address":     address,
			"private_key": base64.StdEncoding.EncodeToString(keyPair.PrivateKey),
		})
	}
}

// Login handles user login
func Login(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		req := new(LoginRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Password is required",
			})
		}
		if req.Email == "" && req.Phone == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Email or phone number is required",
			})
		}

		// Find user by email or phone
		var user *models.User
		var err error
		if req.Email != "" {
			user, err = models.GetUserByEmail(req.Email)
		} else {
			user, err = models.GetUserByPhone(req.Phone)
		}
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid credentials",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to find user",
			})
		}

		// Verify password
		valid, err := crypto.VerifyPassword(
			req.Password,
			user.PasswordHash,
			cfg.Auth.Argon2Time,
			cfg.Auth.Argon2Memory,
			cfg.Auth.Argon2Threads,
			cfg.Auth.Argon2KeyLength,
		)
		if err != nil || !valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Generate JWT token
		token, err := middleware.GenerateJWT(user, cfg.Auth.JWTSecret, cfg.Auth.JWTExpirationTime)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		// Return token and address
		return c.Status(fiber.StatusOK).JSON(AuthResponse{
			Token:   token,
			Address: user.Address,
		})
	}
}

// GetProfile handles getting the user's profile
func GetProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get user from database
		user, err := models.GetUserByID(userID)
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

		// Return user profile
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"phone":      user.Phone,
			"address":    user.Address,
			"public_key": base64.StdEncoding.EncodeToString(user.PublicKey),
			"created_at": user.CreatedAt,
		})
	}
}

// UpdateProfile handles updating the user's profile
func UpdateProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get user from database
		user, err := models.GetUserByID(userID)
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

		// Parse request body
		type UpdateRequest struct {
			Email    *string `json:"email,omitempty"`
			Phone    *string `json:"phone,omitempty"`
			Password *string `json:"password,omitempty"`
		}
		req := new(UpdateRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Update user fields
		if req.Email != nil {
			user.Email = *req.Email
		}
		if req.Phone != nil {
			user.Phone = *req.Phone
		}

		// Update user in database
		if err := models.UpdateUser(user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user",
			})
		}

		// Return updated user profile
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"phone":      user.Phone,
			"address":    user.Address,
			"public_key": base64.StdEncoding.EncodeToString(user.PublicKey),
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}
} 