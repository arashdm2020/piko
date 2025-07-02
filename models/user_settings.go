package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrSettingsNotFound is returned when settings are not found
	ErrSettingsNotFound = errors.New("settings not found")
)

// ThemeType represents the theme setting
type ThemeType string

// PrivacyType represents privacy settings
type PrivacyType string

const (
	// ThemeLight represents light theme
	ThemeLight ThemeType = "light"
	// ThemeDark represents dark theme
	ThemeDark ThemeType = "dark"
	// ThemeSystem represents system theme
	ThemeSystem ThemeType = "system"

	// PrivacyEveryone allows everyone to see the information
	PrivacyEveryone PrivacyType = "everyone"
	// PrivacyContacts allows only contacts to see the information
	PrivacyContacts PrivacyType = "contacts"
	// PrivacyNobody allows nobody to see the information
	PrivacyNobody PrivacyType = "nobody"
)

// UserSettings represents user settings
type UserSettings struct {
	UserID              int         `json:"user_id"`
	Nickname            string      `json:"nickname"`
	Theme               ThemeType   `json:"theme"`
	NotificationEnabled bool        `json:"notification_enabled"`
	SoundEnabled        bool        `json:"sound_enabled"`
	Language            string      `json:"language"`
	AutoDownloadMedia   bool        `json:"auto_download_media"`
	PrivacyLastSeen     PrivacyType `json:"privacy_last_seen"`
	PrivacyProfilePhoto PrivacyType `json:"privacy_profile_photo"`
	PrivacyStatus       PrivacyType `json:"privacy_status"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}

// GetUserSettings retrieves settings for a user
func GetUserSettings(userID int) (*UserSettings, error) {
	settings := &UserSettings{}
	err := database.DB.QueryRow(`
		SELECT user_id, nickname, theme, notification_enabled, sound_enabled, 
		       language, auto_download_media, privacy_last_seen, 
		       privacy_profile_photo, privacy_status, created_at, updated_at 
		FROM user_settings 
		WHERE user_id = ?
	`, userID).Scan(
		&settings.UserID, &settings.Nickname, &settings.Theme, &settings.NotificationEnabled,
		&settings.SoundEnabled, &settings.Language, &settings.AutoDownloadMedia,
		&settings.PrivacyLastSeen, &settings.PrivacyProfilePhoto, &settings.PrivacyStatus,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSettingsNotFound
		}
		return nil, err
	}

	return settings, nil
}

// CreateDefaultSettings creates default settings for a user
func CreateDefaultSettings(userID int) (*UserSettings, error) {
	// Check if settings already exist
	_, err := GetUserSettings(userID)
	if err == nil {
		return nil, errors.New("settings already exist for this user")
	}
	if err != ErrSettingsNotFound {
		return nil, err
	}

	// Create default settings
	settings := &UserSettings{
		UserID:              userID,
		Theme:               ThemeSystem,
		NotificationEnabled: true,
		SoundEnabled:        true,
		Language:            "en",
		AutoDownloadMedia:   true,
		PrivacyLastSeen:     PrivacyEveryone,
		PrivacyProfilePhoto: PrivacyEveryone,
		PrivacyStatus:       PrivacyEveryone,
	}

	_, err = database.DB.Exec(`
		INSERT INTO user_settings (
			user_id, nickname, theme, notification_enabled, sound_enabled,
			language, auto_download_media, privacy_last_seen,
			privacy_profile_photo, privacy_status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		settings.UserID, settings.Nickname, settings.Theme, settings.NotificationEnabled,
		settings.SoundEnabled, settings.Language, settings.AutoDownloadMedia,
		settings.PrivacyLastSeen, settings.PrivacyProfilePhoto, settings.PrivacyStatus,
	)

	if err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateUserSettings updates settings for a user
func UpdateUserSettings(settings *UserSettings) error {
	_, err := database.DB.Exec(`
		UPDATE user_settings SET
			nickname = ?,
			theme = ?,
			notification_enabled = ?,
			sound_enabled = ?,
			language = ?,
			auto_download_media = ?,
			privacy_last_seen = ?,
			privacy_profile_photo = ?,
			privacy_status = ?
		WHERE user_id = ?
	`,
		settings.Nickname, settings.Theme, settings.NotificationEnabled,
		settings.SoundEnabled, settings.Language, settings.AutoDownloadMedia,
		settings.PrivacyLastSeen, settings.PrivacyProfilePhoto, settings.PrivacyStatus,
		settings.UserID,
	)

	return err
}

// UpdateNickname updates only the nickname for a user
func UpdateNickname(userID int, nickname string) error {
	_, err := database.DB.Exec(
		"UPDATE user_settings SET nickname = ? WHERE user_id = ?",
		nickname, userID,
	)
	return err
}
