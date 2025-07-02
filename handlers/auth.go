package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/config"
	"github.com/piko/piko/crypto"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
	"github.com/piko/piko/utils"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Phone string `json:"phone"`
}

// VerifyOTPRequest represents an OTP verification request
type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Phone string `json:"phone"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token   string `json:"token"`
	Address string `json:"address"`
}

// Register handles user registration - Step 1: Send OTP
func Register(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("Register handler called")

		// Parse request body
		req := new(RegisterRequest)
		if err := c.BodyParser(req); err != nil {
			fmt.Printf("Error parsing request body: %v\n", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		fmt.Printf("Register request: phone=%s\n", req.Phone)

		// Validate request
		if req.Phone == "" {
			fmt.Println("Phone number is required")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone number is required",
			})
		}

		// Validate phone number format
		if !utils.IsValidPhone(req.Phone) {
			fmt.Printf("Invalid phone number format: %s\n", req.Phone)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid phone number format",
			})
		}

		// Check if phone number already exists
		_, err := models.GetUserByPhone(req.Phone)
		if err == nil {
			// User already exists, we'll let them log in instead
			fmt.Printf("Phone number already registered: %s\n", req.Phone)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":  "Phone number already registered",
				"action": "login",
			})
		} else if !errors.Is(err, models.ErrUserNotFound) {
			// Database error
			fmt.Printf("Database error checking phone: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check phone number",
			})
		}

		// Generate OTP
		fmt.Printf("Generating OTP for phone: %s\n", req.Phone)
		otp, err := models.GenerateOTP(req.Phone, cfg.Auth.OTPExpiryMinutes)
		if err != nil {
			fmt.Printf("Failed to generate OTP: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate OTP",
			})
		}

		// Send OTP via SMS
		fmt.Printf("Sending OTP to phone: %s, code: %s\n", req.Phone, otp.Code)
		err = utils.SendOTP(utils.FromConfigSMS(&cfg.SMS), req.Phone, otp.Code)
		if err != nil {
			fmt.Printf("Failed to send OTP: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send OTP",
			})
		}

		fmt.Printf("OTP sent successfully to: %s\n", req.Phone)
		// Return success
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":    "OTP sent to your phone",
			"expires_in": cfg.Auth.OTPExpiryMinutes,
		})
	}
}

// VerifyRegister handles user registration - Step 2: Verify OTP and create user
func VerifyRegister(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		req := new(VerifyOTPRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Phone == "" || req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone number and verification code are required",
			})
		}

		// Verify OTP
		verified, err := models.VerifyOTP(req.Phone, req.Code)
		if err != nil {
			if errors.Is(err, models.ErrOTPNotFound) || errors.Is(err, models.ErrOTPExpired) || errors.Is(err, models.ErrOTPInvalid) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			if errors.Is(err, models.ErrOTPMaxAttempts) {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Maximum verification attempts reached. Please request a new OTP.",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify OTP",
			})
		}

		// If OTP verification failed, return an error and don't create the user
		if !verified {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid verification code",
			})
		}

		// OTP verification successful, now check if user already exists
		existingUser, err := models.GetUserByPhone(req.Phone)
		if err == nil {
			// User already exists, generate token and return
			token, err := middleware.GenerateJWT(existingUser, cfg.Auth.JWTSecret, cfg.Auth.JWTExpirationTime)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to generate token",
				})
			}

			return c.Status(fiber.StatusOK).JSON(AuthResponse{
				Token:   token,
				Address: existingUser.Address,
			})
		} else if !errors.Is(err, models.ErrUserNotFound) {
			// Database error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check user",
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

		// Create a random password hash (not used for authentication, but needed for DB schema)
		randomBytes, err := crypto.GenerateRandomBytes(16)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate random bytes",
			})
		}
		passwordHash := base64.StdEncoding.EncodeToString(randomBytes)

		// Create user
		user := &models.User{
			Phone:        req.Phone,
			PasswordHash: passwordHash,
			PublicKey:    keyPair.PublicKey,
			Address:      address,
		}
		err = models.CreateUser(user)
		if err != nil {
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

// Login handles user login - Step 1: Send OTP
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
		if req.Phone == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone number is required",
			})
		}

		// Check if user exists
		_, err := models.GetUserByPhone(req.Phone)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":  "User not found",
					"action": "register",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to find user",
			})
		}

		// Generate OTP
		otp, err := models.GenerateOTP(req.Phone, cfg.Auth.OTPExpiryMinutes)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate OTP",
			})
		}

		// Send OTP via SMS
		err = utils.SendOTP(utils.FromConfigSMS(&cfg.SMS), req.Phone, otp.Code)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send OTP",
			})
		}

		// Return success
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":    "OTP sent to your phone",
			"expires_in": cfg.Auth.OTPExpiryMinutes,
		})
	}
}

// VerifyLogin handles user login - Step 2: Verify OTP
func VerifyLogin(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		req := new(VerifyOTPRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Phone == "" || req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone number and verification code are required",
			})
		}

		// Verify OTP
		verified, err := models.VerifyOTP(req.Phone, req.Code)
		if err != nil {
			if errors.Is(err, models.ErrOTPNotFound) || errors.Is(err, models.ErrOTPExpired) || errors.Is(err, models.ErrOTPInvalid) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			if errors.Is(err, models.ErrOTPMaxAttempts) {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Maximum verification attempts reached. Please request a new OTP.",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify OTP",
			})
		}

		if !verified {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid verification code",
			})
		}

		// Find user by phone
		user, err := models.GetUserByPhone(req.Phone)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to find user",
			})
		}

		// Generate JWT token with extended expiration for persistent login
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
		return c.Status(fiber.StatusOK).JSON(user)
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
		var updateReq struct {
			Phone string `json:"phone,omitempty"`
		}
		if err := c.BodyParser(&updateReq); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Update user fields
		if updateReq.Phone != "" {
			user.Phone = updateReq.Phone
		}

		// Save changes
		if err := models.UpdateUser(user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update user",
			})
		}

		// Return updated user
		return c.Status(fiber.StatusOK).JSON(user)
	}
}

// RequestOTP handles requests for OTP verification
func RequestOTP() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		var req struct {
			Phone string `json:"phone"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate phone number
		if req.Phone == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone number is required",
			})
		}

		// Generate OTP code
		code, err := utils.GenerateOTP(6)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate OTP",
			})
		}

		// Set expiration time (5 minutes)
		expiresAt := time.Now().Add(5 * time.Minute)

		// Save OTP to database
		otp := &models.OTP{
			Phone:     req.Phone,
			Code:      code,
			ExpiresAt: expiresAt,
		}
		if err := models.SaveOTP(otp); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save OTP",
			})
		}

		// Get SMS config
		smsConfig := utils.DefaultSMSConfig()

		// Send OTP via SMS
		if err := utils.SendOTP(smsConfig, req.Phone, code); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send OTP",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "OTP sent successfully",
		})
	}
}

// VerifyOTP handles OTP verification
func VerifyOTP() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		var req struct {
			Phone string `json:"phone"`
			Code  string `json:"verfication-code"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate request
		if req.Phone == "" || req.Code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Phone number and verification code are required",
			})
		}

		// Verify OTP
		verified, err := models.VerifyOTP(req.Phone, req.Code)
		if err != nil {
			if errors.Is(err, models.ErrOTPNotFound) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "OTP not found",
				})
			}
			if errors.Is(err, models.ErrOTPExpired) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "OTP expired",
				})
			}
			if errors.Is(err, models.ErrOTPInvalid) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid OTP",
				})
			}
			if errors.Is(err, models.ErrOTPMaxAttempts) {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Maximum verification attempts reached. Please request a new OTP.",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify OTP",
			})
		}

		if !verified {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired verification code",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Phone number verified successfully",
		})
	}
}
