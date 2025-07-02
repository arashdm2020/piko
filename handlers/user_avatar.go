package handlers

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// UploadAvatar handles uploading a user avatar
func UploadAvatar() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get file from request
		file, err := c.FormFile("avatar")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No avatar file provided",
			})
		}

		// Validate file size (max 5MB)
		if file.Size > 5*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Avatar file too large (max 5MB)",
			})
		}

		// Validate file type
		contentType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid file type. Only images are allowed",
			})
		}

		// Create uploads directory if it doesn't exist
		uploadsDir := "./uploads/avatars"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create uploads directory",
			})
		}

		// Generate a unique filename
		filename := fmt.Sprintf("%d_%s", userID, filepath.Base(file.Filename))
		filepath := filepath.Join(uploadsDir, filename)

		// Save the file
		if err := c.SaveFile(file, filepath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save avatar",
			})
		}

		// TODO: Process image (resize, compress, etc.)
		// For now, we'll just use placeholder values for width and height
		width := 200
		height := 200

		// Create avatar record in database
		avatar := &models.UserAvatar{
			UserID:   userID,
			FilePath: filepath,
			FileName: filename,
			FileSize: int(file.Size),
			MimeType: contentType,
			Width:    width,
			Height:   height,
			IsActive: true, // Set as active by default
		}

		if err := models.CreateAvatar(avatar); err != nil {
			// Delete the file if database insertion fails
			os.Remove(filepath)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save avatar information",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(avatar)
	}
}

// GetUserAvatars handles retrieving all avatars for a user
func GetUserAvatars() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get avatars
		avatars, err := models.GetAllAvatarsForUser(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get avatars",
			})
		}

		return c.Status(fiber.StatusOK).JSON(avatars)
	}
}

// GetActiveAvatar handles retrieving the active avatar for a user
func GetActiveAvatar() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get active avatar
		avatar, err := models.GetActiveAvatarForUser(userID)
		if err != nil {
			if errors.Is(err, models.ErrAvatarNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "No active avatar found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get active avatar",
			})
		}

		return c.Status(fiber.StatusOK).JSON(avatar)
	}
}

// SetActiveAvatar handles setting an avatar as active
func SetActiveAvatar() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get avatar ID from URL parameter
		avatarID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid avatar ID",
			})
		}

		// Set avatar as active
		if err := models.SetActiveAvatar(avatarID, userID); err != nil {
			if errors.Is(err, models.ErrAvatarNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Avatar not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to set active avatar",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Avatar set as active",
		})
	}
}

// DeleteAvatar handles deleting an avatar
func DeleteAvatar() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get avatar ID from URL parameter
		avatarID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid avatar ID",
			})
		}

		// Get the avatar to get the file path
		avatar, err := models.GetAvatarByID(avatarID)
		if err != nil {
			if errors.Is(err, models.ErrAvatarNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Avatar not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get avatar",
			})
		}

		// Verify that the avatar belongs to the user
		if avatar.UserID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You don't have permission to delete this avatar",
			})
		}

		// Delete avatar from database
		if err := models.DeleteAvatar(avatarID, userID); err != nil {
			if errors.Is(err, models.ErrAvatarNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Avatar not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete avatar",
			})
		}

		// Delete the file
		if err := os.Remove(avatar.FilePath); err != nil && !os.IsNotExist(err) {
			// Log error but don't return it to the client
			// The database record is already deleted
			fmt.Printf("Error deleting avatar file: %v\n", err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Avatar deleted successfully",
		})
	}
}

// ServeAvatar handles serving an avatar file
func ServeAvatar() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get avatar ID from URL parameter
		avatarID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid avatar ID",
			})
		}

		// Get avatar from database
		avatar, err := models.GetAvatarByID(avatarID)
		if err != nil {
			if errors.Is(err, models.ErrAvatarNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Avatar not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get avatar",
			})
		}

		// Open the file
		file, err := os.Open(avatar.FilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Avatar file not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to open avatar file",
			})
		}
		defer file.Close()

		// Set content type header
		c.Set("Content-Type", avatar.MimeType)
		c.Set("Content-Length", strconv.Itoa(avatar.FileSize))

		// Stream the file to the response
		_, err = io.Copy(c, file)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send avatar file",
			})
		}

		return nil
	}
}
