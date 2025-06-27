package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyExists is returned when a user with the same email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")
	// ErrPhoneAlreadyExists is returned when a user with the same phone already exists
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	// ErrAddressAlreadyExists is returned when a user with the same address already exists
	ErrAddressAlreadyExists = errors.New("address already exists")
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	PasswordHash string    `json:"-"`
	PublicKey    []byte    `json:"public_key"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUser creates a new user in the database
func CreateUser(user *User) error {
	// Check if user with same email exists
	if user.Email != "" {
		var count int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrEmailAlreadyExists
		}
	}

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

	// Insert user into database
	result, err := database.DB.Exec(
		"INSERT INTO users (email, phone, password_hash, public_key, address) VALUES (?, ?, ?, ?, ?)",
		user.Email, user.Phone, user.PasswordHash, user.PublicKey, user.Address,
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

	return nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(id int) (*User, error) {
	user := &User{}
	err := database.DB.QueryRow(
		"SELECT id, email, phone, password_hash, public_key, address, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := database.DB.QueryRow(
		"SELECT id, email, phone, password_hash, public_key, address, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
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
		"SELECT id, email, phone, password_hash, public_key, address, created_at, updated_at FROM users WHERE phone = ?",
		phone,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
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
		"SELECT id, email, phone, password_hash, public_key, address, created_at, updated_at FROM users WHERE address = ?",
		address,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.PublicKey, &user.Address, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user's information
func UpdateUser(user *User) error {
	_, err := database.DB.Exec(
		"UPDATE users SET email = ?, phone = ?, password_hash = ?, public_key = ? WHERE id = ?",
		user.Email, user.Phone, user.PasswordHash, user.PublicKey, user.ID,
	)
	return err
}

// DeleteUser deletes a user by their ID
func DeleteUser(id int) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
} 