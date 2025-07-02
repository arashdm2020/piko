package models

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrOTPNotFound is returned when an OTP is not found
	ErrOTPNotFound = errors.New("otp not found")
	// ErrOTPExpired is returned when an OTP has expired
	ErrOTPExpired = errors.New("otp expired")
	// ErrOTPInvalid is returned when an OTP is invalid
	ErrOTPInvalid = errors.New("otp invalid")
	// ErrOTPMaxAttempts is returned when maximum attempts are reached
	ErrOTPMaxAttempts = errors.New("maximum verification attempts reached")
)

// Maximum allowed failed attempts before OTP is invalidated
const MaxOTPFailedAttempts = 3

// OTP represents a one-time password for phone verification
type OTP struct {
	ID             int       `json:"id"`
	Phone          string    `json:"phone"`
	Code           string    `json:"code"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Verified       bool      `json:"verified"`
	FailedAttempts int       `json:"failed_attempts"`
}

// GenerateOTP generates a new OTP for a phone number
func GenerateOTP(phone string, expiryMinutes int) (*OTP, error) {
	// Delete any existing OTPs for this phone number
	_, err := database.DB.Exec("DELETE FROM otp WHERE phone = ?", phone)
	if err != nil {
		fmt.Printf("Error deleting existing OTPs: %v\n", err)
		return nil, err
	}

	// Generate a random 6-digit code
	code := generateRandomCode(6)
	fmt.Printf("Generated OTP code for %s: %s\n", phone, code)

	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	// Insert the OTP into the database
	result, err := database.DB.Exec(
		"INSERT INTO otp (phone, code, expires_at, failed_attempts) VALUES (?, ?, ?, 0)",
		phone, code, expiresAt,
	)
	if err != nil {
		fmt.Printf("Error inserting OTP into database: %v\n", err)
		return nil, err
	}

	// Get the ID of the inserted OTP
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("Error getting last insert ID: %v\n", err)
		return nil, err
	}

	// Create and return the OTP
	otp := &OTP{
		ID:             int(id),
		Phone:          phone,
		Code:           code,
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
		Verified:       false,
		FailedAttempts: 0,
	}

	fmt.Printf("OTP created successfully: ID=%d, Phone=%s, Code=%s, ExpiresAt=%v\n",
		otp.ID, otp.Phone, otp.Code, otp.ExpiresAt)

	return otp, nil
}

// SaveOTP saves an OTP to the database
func SaveOTP(otp *OTP) error {
	// First, invalidate any existing OTPs for this phone number
	_, err := database.DB.Exec("UPDATE otp SET verified = TRUE WHERE phone = ? AND verified = FALSE", otp.Phone)
	if err != nil {
		return err
	}

	// Insert new OTP
	_, err = database.DB.Exec(
		"INSERT INTO otp (phone, code, expires_at, failed_attempts) VALUES (?, ?, ?, 0)",
		otp.Phone, otp.Code, otp.ExpiresAt,
	)
	return err
}

// VerifyOTP checks if an OTP is valid and marks it as verified if it is
func VerifyOTP(phone, code string) (bool, error) {
	// Get the OTP from the database
	var otp OTP
	err := database.DB.QueryRow(
		"SELECT id, phone, code, created_at, expires_at, verified, failed_attempts FROM otp WHERE phone = ? ORDER BY id DESC LIMIT 1",
		phone,
	).Scan(&otp.ID, &otp.Phone, &otp.Code, &otp.CreatedAt, &otp.ExpiresAt, &otp.Verified, &otp.FailedAttempts)

	if err != nil {
		if err == sql.ErrNoRows {
			// No matching OTP found
			return false, ErrOTPNotFound
		}
		return false, err
	}

	// Check if the OTP is already verified
	if otp.Verified {
		return false, ErrOTPInvalid
	}

	// Check if the OTP has expired
	if time.Now().After(otp.ExpiresAt) {
		return false, ErrOTPExpired
	}

	// Check if max attempts reached
	if otp.FailedAttempts >= MaxOTPFailedAttempts {
		return false, ErrOTPMaxAttempts
	}

	// Check if the code matches
	if otp.Code != code {
		// Increment failed attempts
		_, err = database.DB.Exec(
			"UPDATE otp SET failed_attempts = failed_attempts + 1 WHERE id = ?",
			otp.ID,
		)
		if err != nil {
			return false, err
		}

		// Check if this attempt exceeds the max attempts
		if otp.FailedAttempts+1 >= MaxOTPFailedAttempts {
			// Invalidate the OTP
			_, err = database.DB.Exec("UPDATE otp SET verified = TRUE WHERE id = ?", otp.ID)
			if err != nil {
				return false, err
			}
			return false, ErrOTPMaxAttempts
		}

		return false, ErrOTPInvalid
	}

	// Mark the OTP as verified
	_, err = database.DB.Exec("UPDATE otp SET verified = TRUE WHERE id = ?", otp.ID)
	if err != nil {
		return false, err
	}

	// OTP is valid and now marked as verified
	return true, nil
}

// DeleteOTP deletes an OTP for a phone number
func DeleteOTP(phone string) error {
	_, err := database.DB.Exec("DELETE FROM otp WHERE phone = ?", phone)
	return err
}

// generateRandomCode generates a random numeric code of the specified length
func generateRandomCode(length int) string {
	const digits = "0123456789"

	// Create a new random source with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = digits[r.Intn(len(digits))]
	}

	return string(code)
}
