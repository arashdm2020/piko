package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/config"
	"github.com/piko/piko/handlers"
	"github.com/piko/piko/middleware"
)

// ErrorHandler handles API errors
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default error
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Return JSON response
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}

// RegisterRoutes registers all API routes
func RegisterRoutes(app *fiber.App) {
	// Load configuration
	cfg := config.DefaultConfig()

	// Public routes
	app.Post("/api/auth/register", handlers.Register(cfg))
	app.Post("/api/auth/login", handlers.Login(cfg))

	// Auth middleware for protected routes
	authMiddleware := middleware.AuthRequired(cfg)

	// User routes
	app.Get("/api/profile", authMiddleware, handlers.GetProfile())
	app.Put("/api/profile", authMiddleware, handlers.UpdateProfile())

	// Message routes
	app.Post("/api/messages", authMiddleware, handlers.SendMessage())
	app.Get("/api/messages/inbox", authMiddleware, handlers.GetInbox())
	app.Get("/api/messages/sent", authMiddleware, handlers.GetSentMessages())
	app.Get("/api/messages/:id", authMiddleware, handlers.GetMessage())
	app.Delete("/api/messages/:id", authMiddleware, handlers.DeleteMessage())

	// Channel routes
	app.Post("/api/channels", authMiddleware, handlers.CreateChannel())
	app.Get("/api/channels", authMiddleware, handlers.GetChannels())
	app.Get("/api/channels/:id", authMiddleware, handlers.GetChannel())
	app.Put("/api/channels/:id", authMiddleware, handlers.UpdateChannel())
	app.Delete("/api/channels/:id", authMiddleware, handlers.DeleteChannel())
	app.Post("/api/channels/:id/members", authMiddleware, handlers.AddChannelMember())
	app.Get("/api/channels/:id/members", authMiddleware, handlers.GetChannelMembers())
	app.Delete("/api/channels/:id/members/:address", authMiddleware, handlers.RemoveChannelMember())
	app.Post("/api/channels/:id/messages", authMiddleware, handlers.SendChannelMessage())
	app.Get("/api/channels/:id/messages", authMiddleware, handlers.GetChannelMessages())
	app.Delete("/api/channels/:channel_id/messages/:message_id", authMiddleware, handlers.DeleteChannelMessage())

	// Blockchain routes
	app.Get("/api/blocks/:id", authMiddleware, handlers.GetBlock())
	app.Get("/api/blocks/height/:height", authMiddleware, handlers.GetBlockByHeight())
	app.Get("/api/transactions/:hash", authMiddleware, handlers.GetTransaction())
	app.Get("/api/explore/:address", authMiddleware, handlers.ExploreAddress())
	app.Get("/api/proof/:message_id", authMiddleware, handlers.GetProof())
	app.Get("/api/blockchain/stats", authMiddleware, handlers.GetBlockchainStats())

	// Secret Chat routes (no authentication required)
	app.Post("/api/secret-chat/create", handlers.CreateSecretChat())
	app.Post("/api/secret-chat/join", handlers.JoinSecretChat())
	app.Post("/api/secret-chat/send", handlers.SendSecretChatMessage())
	app.Get("/api/secret-chat/messages/:channel_id", handlers.GetSecretChatMessages())
	app.Delete("/api/secret-chat/:channel_id", handlers.DeleteSecretChat())

	// Secret Chat WebSocket route
	app.Get("/ws/secret/:session_id", handlers.SecretChatWebSocketHandler())

	// Regular WebSocket route
	app.Get("/ws", handlers.WebSocketHandler())
}
