package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrAvatarNotFound is returned when an avatar is not found
	ErrAvatarNotFound = errors.New("avatar not found")
)

// UserAvatar represents a user avatar
type UserAvatar struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	FilePath  string    `json:"file_path"`
	FileName  string    `json:"file_name"`
	FileSize  int       `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAvatar creates a new avatar for a user
func CreateAvatar(avatar *UserAvatar) error {
	// If this is set as active, deactivate all other avatars for this user
	if avatar.IsActive {
		_, err := database.DB.Exec("UPDATE user_avatars SET is_active = FALSE WHERE user_id = ?", avatar.UserID)
		if err != nil {
			return err
		}
	}

	result, err := database.DB.Exec(`
		INSERT INTO user_avatars (
			user_id, file_path, file_name, file_size, 
			mime_type, width, height, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		avatar.UserID, avatar.FilePath, avatar.FileName, avatar.FileSize,
		avatar.MimeType, avatar.Width, avatar.Height, avatar.IsActive,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	avatar.ID = int(id)
	return nil
}

// GetAvatarByID retrieves an avatar by ID
func GetAvatarByID(id int) (*UserAvatar, error) {
	avatar := &UserAvatar{}
	err := database.DB.QueryRow(`
		SELECT id, user_id, file_path, file_name, file_size, 
		       mime_type, width, height, is_active, created_at
		FROM user_avatars 
		WHERE id = ?
	`, id).Scan(
		&avatar.ID, &avatar.UserID, &avatar.FilePath, &avatar.FileName,
		&avatar.FileSize, &avatar.MimeType, &avatar.Width, &avatar.Height,
		&avatar.IsActive, &avatar.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAvatarNotFound
		}
		return nil, err
	}

	return avatar, nil
}

// GetActiveAvatarForUser retrieves the active avatar for a user
func GetActiveAvatarForUser(userID int) (*UserAvatar, error) {
	avatar := &UserAvatar{}
	err := database.DB.QueryRow(`
		SELECT id, user_id, file_path, file_name, file_size, 
		       mime_type, width, height, is_active, created_at
		FROM user_avatars 
		WHERE user_id = ? AND is_active = TRUE
		LIMIT 1
	`, userID).Scan(
		&avatar.ID, &avatar.UserID, &avatar.FilePath, &avatar.FileName,
		&avatar.FileSize, &avatar.MimeType, &avatar.Width, &avatar.Height,
		&avatar.IsActive, &avatar.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrAvatarNotFound
		}
		return nil, err
	}

	return avatar, nil
}

// GetAllAvatarsForUser retrieves all avatars for a user
func GetAllAvatarsForUser(userID int) ([]*UserAvatar, error) {
	rows, err := database.DB.Query(`
		SELECT id, user_id, file_path, file_name, file_size, 
		       mime_type, width, height, is_active, created_at
		FROM user_avatars 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var avatars []*UserAvatar
	for rows.Next() {
		avatar := &UserAvatar{}
		err := rows.Scan(
			&avatar.ID, &avatar.UserID, &avatar.FilePath, &avatar.FileName,
			&avatar.FileSize, &avatar.MimeType, &avatar.Width, &avatar.Height,
			&avatar.IsActive, &avatar.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		avatars = append(avatars, avatar)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return avatars, nil
}

// SetActiveAvatar sets an avatar as active and deactivates all others
func SetActiveAvatar(avatarID int, userID int) error {
	// First verify that the avatar belongs to the user
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM user_avatars WHERE id = ? AND user_id = ?", avatarID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrAvatarNotFound
	}

	// Deactivate all avatars for this user
	_, err = database.DB.Exec("UPDATE user_avatars SET is_active = FALSE WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Set the specified avatar as active
	_, err = database.DB.Exec("UPDATE user_avatars SET is_active = TRUE WHERE id = ?", avatarID)
	return err
}

// DeleteAvatar deletes an avatar
func DeleteAvatar(avatarID int, userID int) error {
	// First verify that the avatar belongs to the user
	var isActive bool
	err := database.DB.QueryRow("SELECT is_active FROM user_avatars WHERE id = ? AND user_id = ?", avatarID, userID).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrAvatarNotFound
		}
		return err
	}

	// Delete the avatar
	_, err = database.DB.Exec("DELETE FROM user_avatars WHERE id = ?", avatarID)
	if err != nil {
		return err
	}

	// If this was the active avatar, set the most recent one as active
	if isActive {
		_, err = database.DB.Exec(`
			UPDATE user_avatars 
			SET is_active = TRUE 
			WHERE user_id = ? 
			ORDER BY created_at DESC 
			LIMIT 1
		`, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
