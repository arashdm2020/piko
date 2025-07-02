package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/piko/piko/middleware"
	"github.com/piko/piko/models"
)

// GetUserSettings handles retrieving user settings
func GetUserSettings() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get user settings
		settings, err := models.GetUserSettings(userID)
		if err != nil {
			if errors.Is(err, models.ErrSettingsNotFound) {
				// Create default settings if not found
				settings, err = models.CreateDefaultSettings(userID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to create default settings",
					})
				}
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to get settings",
				})
			}
		}

		return c.Status(fiber.StatusOK).JSON(settings)
	}
}

// UpdateUserSettings handles updating user settings
func UpdateUserSettings() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Get current settings
		settings, err := models.GetUserSettings(userID)
		if err != nil {
			if errors.Is(err, models.ErrSettingsNotFound) {
				// Create default settings if not found
				settings, err = models.CreateDefaultSettings(userID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to create default settings",
					})
				}
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to get settings",
				})
			}
		}

		// Parse request body
		var updateReq struct {
			Nickname            string             `json:"nickname"`
			Theme               models.ThemeType   `json:"theme"`
			NotificationEnabled *bool              `json:"notification_enabled"`
			SoundEnabled        *bool              `json:"sound_enabled"`
			Language            string             `json:"language"`
			AutoDownloadMedia   *bool              `json:"auto_download_media"`
			PrivacyLastSeen     models.PrivacyType `json:"privacy_last_seen"`
			PrivacyProfilePhoto models.PrivacyType `json:"privacy_profile_photo"`
			PrivacyStatus       models.PrivacyType `json:"privacy_status"`
		}
		if err := c.BodyParser(&updateReq); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Update only provided fields
		if updateReq.Nickname != "" {
			settings.Nickname = updateReq.Nickname
		}
		if updateReq.Theme != "" {
			settings.Theme = updateReq.Theme
		}
		if updateReq.NotificationEnabled != nil {
			settings.NotificationEnabled = *updateReq.NotificationEnabled
		}
		if updateReq.SoundEnabled != nil {
			settings.SoundEnabled = *updateReq.SoundEnabled
		}
		if updateReq.Language != "" {
			settings.Language = updateReq.Language
		}
		if updateReq.AutoDownloadMedia != nil {
			settings.AutoDownloadMedia = *updateReq.AutoDownloadMedia
		}
		if updateReq.PrivacyLastSeen != "" {
			settings.PrivacyLastSeen = updateReq.PrivacyLastSeen
		}
		if updateReq.PrivacyProfilePhoto != "" {
			settings.PrivacyProfilePhoto = updateReq.PrivacyProfilePhoto
		}
		if updateReq.PrivacyStatus != "" {
			settings.PrivacyStatus = updateReq.PrivacyStatus
		}

		// Save changes
		if err := models.UpdateUserSettings(settings); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update settings",
			})
		}

		return c.Status(fiber.StatusOK).JSON(settings)
	}
}

// UpdateNickname handles updating just the user's nickname
func UpdateNickname() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := middleware.GetUserID(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Parse request body
		var req struct {
			Nickname string `json:"nickname"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate nickname
		if req.Nickname == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Nickname is required",
			})
		}

		// Update nickname
		err := models.UpdateNickname(userID, req.Nickname)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update nickname",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":  "Nickname updated successfully",
			"nickname": req.Nickname,
		})
	}
}
