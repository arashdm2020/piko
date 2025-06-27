package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/piko/piko/config"
	"github.com/piko/piko/models"
)

var (
	// ErrNoAuthHeader is returned when no authorization header is provided
	ErrNoAuthHeader = errors.New("no authorization header provided")
	// ErrInvalidAuthHeader is returned when the authorization header is invalid
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
	// ErrInvalidToken is returned when the JWT token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when the JWT token is expired
	ErrTokenExpired = errors.New("token expired")
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID  int    `json:"user_id"`
	Address string `json:"address"`
	jwt.StandardClaims
}

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(user *models.User, secret string, expirationTime time.Duration) (string, error) {
	claims := JWTClaims{
		UserID:  user.ID,
		Address: user.Address,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expirationTime).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "piko",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AuthRequired is a middleware that checks if the user is authenticated
func AuthRequired(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrNoAuthHeader.Error(),
			})
		}

		// Check if the header is in the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrInvalidAuthHeader.Error(),
			})
		}

		// Parse the token
		tokenString := parts[1]
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Auth.JWTSecret), nil
		})

		// Check for parsing errors
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": ErrTokenExpired.Error(),
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrInvalidToken.Error(),
			})
		}

		// Check if the token is valid
		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrInvalidToken.Error(),
			})
		}

		// Get the claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrInvalidToken.Error(),
			})
		}

		// Store the claims in the context
		c.Locals("user_id", claims.UserID)
		c.Locals("address", claims.Address)

		// Continue to the next middleware/handler
		return c.Next()
	}
}

// GetUserID gets the user ID from the context
func GetUserID(c *fiber.Ctx) (int, bool) {
	userID, ok := c.Locals("user_id").(int)
	return userID, ok
}

// GetUserAddress gets the user address from the context
func GetUserAddress(c *fiber.Ctx) (string, bool) {
	address, ok := c.Locals("address").(string)
	return address, ok
} 