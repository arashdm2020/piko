package models

import (
	"database/sql"
	"errors"
	"regexp"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrPhoneAlreadyExists is returned when a user with the same phone already exists
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	// ErrAddressAlreadyExists is returned when a user with the same address already exists
	ErrAddressAlreadyExists = errors.New("address already exists")
	// ErrUsernameAlreadyExists is returned when a user with the same username already exists
	ErrUsernameAlreadyExists = errors.New("username already exists")
	// ErrInvalidUsername is returned when the username format is invalid
	ErrInvalidUsername = errors.New("invalid username format")
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Phone        string    `json:"phone"`
	Username     string    `json:"username,omitempty"`
	PasswordHash string    `json:"-"`
	PublicKey    []byte    `json:"public_key"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUser creates a new user in the database
func CreateUser(user *User) error {
	// Check if user with same phone exists
	if user.Phone != "" {
		var count int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE phone = ?", user.Phone).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrPhoneAlreadyExists
		}
	}

	// Check if user with same address exists
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE address = ?", user.Address).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAddressAlreadyExists
	}

	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert user into database - username is not set during registration
	result, err := tx.Exec(
		"INSERT INTO users (phone, password_hash, public_key, address) VALUES (?, ?, ?, ?)",
		user.Phone, user.PasswordHash, user.PublicKey, user.Address,
	)
	if err != nil {
		return err
	}

	// Get the ID of the inserted user
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = int(id)

	// Create default settings for the user
	_, err = tx.Exec(`
		INSERT INTO user_settings (
			user_id, theme, notification_enabled, sound_enabled,
			language, auto_download_media, privacy_last_seen,
			privacy_profile_photo, privacy_status
		) VALUES (?, 'system', TRUE, TRUE, 'en', TRUE, 'everyone', 'everyone', 'everyone')
	`, user.ID)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id int) (*User, error) {
	user := &User{}
	err := database.DB.QueryRow(
		"SELECT id, phone, username, password_hash, public_key, address, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(
		&user.ID, &user.Phone, &user.Username, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByPhone retrieves a user by their phone number
func GetUserByPhone(phone string) (*User, error) {
	user := &User{}
	err := database.DB.QueryRow(
		"SELECT id, phone, username, password_hash, public_key, address, created_at, updated_at FROM users WHERE phone = ?",
		phone,
	).Scan(
		&user.ID, &user.Phone, &user.Username, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByAddress retrieves a user by their address
func GetUserByAddress(address string) (*User, error) {
	user := &User{}
	err := database.DB.QueryRow(
		"SELECT id, phone, username, password_hash, public_key, address, created_at, updated_at FROM users WHERE address = ?",
		address,
	).Scan(
		&user.ID, &user.Phone, &user.Username, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := database.DB.QueryRow(
		"SELECT id, phone, username, password_hash, public_key, address, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(
		&user.ID, &user.Phone, &user.Username, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// SearchUsers searches for users by username, phone, or address
func SearchUsers(query string) ([]*User, error) {
	rows, err := database.DB.Query(
		"SELECT id, phone, username, password_hash, public_key, address, created_at, updated_at FROM users WHERE username LIKE ? OR phone LIKE ? OR address LIKE ? LIMIT 20",
		"%"+query+"%", "%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Phone, &user.Username, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates a user's information
func UpdateUser(user *User) error {
	_, err := database.DB.Exec(
		"UPDATE users SET phone = ?, username = ?, password_hash = ?, public_key = ? WHERE id = ?",
		user.Phone, user.Username, user.PasswordHash, user.PublicKey, user.ID,
	)
	return err
}

// SetUsername sets or updates a user's username
func SetUsername(userID int, username string) error {
	// Validate username format
	if !IsValidUsername(username) {
		return ErrInvalidUsername
	}

	// Check if username is already taken
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", username, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrUsernameAlreadyExists
	}

	// Update username
	_, err = database.DB.Exec("UPDATE users SET username = ? WHERE id = ?", username, userID)
	return err
}

// IsValidUsername checks if a username is valid
func IsValidUsername(username string) bool {
	// Username must be 3-30 characters long and contain only alphanumeric characters and underscores
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Check if username contains only alphanumeric characters and underscores
	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	return match
}

// DeleteUser deletes a user by ID
func DeleteUser(id int) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
