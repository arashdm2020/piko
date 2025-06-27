package models

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrSecretChatNotFound is returned when a secret chat is not found
	ErrSecretChatNotFound = errors.New("secret chat not found")
	// ErrSecretChatExpired is returned when a secret chat has expired
	ErrSecretChatExpired = errors.New("secret chat expired")
)

// SecretChat represents a temporary anonymous chat room
type SecretChat struct {
	ChannelID    string    `json:"channel_id"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	MessageCount int       `json:"message_count"`
}

// SecretChatParticipant represents a participant in a secret chat
type SecretChatParticipant struct {
	SessionID    string    `json:"session_id"`
	ChannelID    string    `json:"channel_id"`
	DisplayName  string    `json:"display_name"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// SecretChatMessage represents a message in a secret chat
type SecretChatMessage struct {
	ID               string    `json:"id"`
	ChannelID        string    `json:"channel_id"`
	SessionID        string    `json:"session_id"`
	DisplayName      string    `json:"display_name"`
	EncryptedContent []byte    `json:"encrypted_content"`
	Timestamp        time.Time `json:"timestamp"`
}

// GenerateSecretChatID generates a unique ID for a secret chat
func GenerateSecretChatID() (string, error) {
	// Generate 6 random bytes
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Format as "cXX-XXXXXX" where X is a hex digit
	return fmt.Sprintf("c%02x-%02x%02x%02x%02x",
		randomBytes[0]&0xFF,
		randomBytes[1]&0xFF,
		randomBytes[2]&0xFF,
		randomBytes[3]&0xFF,
		randomBytes[4]&0xFF), nil
}

// CreateSecretChat creates a new secret chat
func CreateSecretChat() (*SecretChat, error) {
	// Generate channel ID
	channelID, err := GenerateSecretChatID()
	if err != nil {
		return nil, err
	}

	// Set expiration time to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create secret chat in database
	_, err = database.DB.Exec(
		"INSERT INTO secret_chats (channel_id, expires_at) VALUES (?, ?)",
		channelID, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	// Return the created secret chat
	return &SecretChat{
		ChannelID:    channelID,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		MessageCount: 0,
	}, nil
}

// GetSecretChat retrieves a secret chat by its ID
func GetSecretChat(channelID string) (*SecretChat, error) {
	chat := &SecretChat{}
	err := database.DB.QueryRow(
		"SELECT channel_id, created_at, expires_at, (SELECT COUNT(*) FROM secret_chat_messages WHERE channel_id = ?) AS message_count FROM secret_chats WHERE channel_id = ?",
		channelID, channelID,
	).Scan(&chat.ChannelID, &chat.CreatedAt, &chat.ExpiresAt, &chat.MessageCount)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSecretChatNotFound
		}
		return nil, err
	}

	// Check if chat has expired
	if time.Now().After(chat.ExpiresAt) {
		return nil, ErrSecretChatExpired
	}

	return chat, nil
}

// JoinSecretChat adds a participant to a secret chat
func JoinSecretChat(channelID string, displayName string) (*SecretChatParticipant, error) {
	// Check if chat exists and is not expired
	_, err := GetSecretChat(channelID)
	if err != nil {
		return nil, err
	}

	// Generate session ID
	sessionID := GenerateSessionID()

	// Create participant in database
	now := time.Now()
	_, err = database.DB.Exec(
		"INSERT INTO secret_chat_participants (session_id, channel_id, display_name, joined_at, last_active_at) VALUES (?, ?, ?, ?, ?)",
		sessionID, channelID, displayName, now, now,
	)
	if err != nil {
		return nil, err
	}

	// Return the created participant
	return &SecretChatParticipant{
		SessionID:    sessionID,
		ChannelID:    channelID,
		DisplayName:  displayName,
		JoinedAt:     now,
		LastActiveAt: now,
	}, nil
}

// GenerateSessionID generates a unique session ID
func GenerateSessionID() string {
	// Generate 16 random bytes
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	// Format as hex string
	return fmt.Sprintf("%x", randomBytes)
}

// UpdateParticipantActivity updates the last active timestamp for a participant
func UpdateParticipantActivity(sessionID string) error {
	_, err := database.DB.Exec(
		"UPDATE secret_chat_participants SET last_active_at = ? WHERE session_id = ?",
		time.Now(), sessionID,
	)
	return err
}

// GetParticipant retrieves a participant by session ID
func GetParticipant(sessionID string) (*SecretChatParticipant, error) {
	participant := &SecretChatParticipant{}
	err := database.DB.QueryRow(
		"SELECT session_id, channel_id, display_name, joined_at, last_active_at FROM secret_chat_participants WHERE session_id = ?",
		sessionID,
	).Scan(&participant.SessionID, &participant.ChannelID, &participant.DisplayName, &participant.JoinedAt, &participant.LastActiveAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("participant not found")
		}
		return nil, err
	}

	return participant, nil
}

// GetParticipantsByChannel retrieves all participants in a channel
func GetParticipantsByChannel(channelID string) ([]*SecretChatParticipant, error) {
	rows, err := database.DB.Query(
		"SELECT session_id, channel_id, display_name, joined_at, last_active_at FROM secret_chat_participants WHERE channel_id = ?",
		channelID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := []*SecretChatParticipant{}
	for rows.Next() {
		participant := &SecretChatParticipant{}
		err := rows.Scan(&participant.SessionID, &participant.ChannelID, &participant.DisplayName, &participant.JoinedAt, &participant.LastActiveAt)
		if err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	return participants, nil
}

// CreateSecretChatMessage creates a new message in a secret chat
func CreateSecretChatMessage(message *SecretChatMessage) error {
	// Check if chat exists and is not expired
	_, err := GetSecretChat(message.ChannelID)
	if err != nil {
		return err
	}

	// Get participant info
	participant, err := GetParticipant(message.SessionID)
	if err != nil {
		return err
	}

	// Update participant's last active timestamp
	if err := UpdateParticipantActivity(message.SessionID); err != nil {
		return err
	}

	// Insert message into database
	_, err = database.DB.Exec(
		"INSERT INTO secret_chat_messages (id, channel_id, session_id, display_name, encrypted_content) VALUES (?, ?, ?, ?, ?)",
		message.ID, message.ChannelID, message.SessionID, participant.DisplayName, message.EncryptedContent,
	)
	return err
}

// GetSecretChatMessages retrieves messages from a secret chat
func GetSecretChatMessages(channelID string, limit int, offset int) ([]*SecretChatMessage, error) {
	// Check if chat exists and is not expired
	_, err := GetSecretChat(channelID)
	if err != nil {
		return nil, err
	}

	// Query messages
	rows, err := database.DB.Query(
		"SELECT id, channel_id, session_id, display_name, encrypted_content, timestamp FROM secret_chat_messages WHERE channel_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?",
		channelID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*SecretChatMessage{}
	for rows.Next() {
		message := &SecretChatMessage{}
		err := rows.Scan(&message.ID, &message.ChannelID, &message.SessionID, &message.DisplayName, &message.EncryptedContent, &message.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// DeleteSecretChat deletes a secret chat and all its messages
func DeleteSecretChat(channelID string) error {
	// Start a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete messages
	_, err = tx.Exec("DELETE FROM secret_chat_messages WHERE channel_id = ?", channelID)
	if err != nil {
		return err
	}

	// Delete participants
	_, err = tx.Exec("DELETE FROM secret_chat_participants WHERE channel_id = ?", channelID)
	if err != nil {
		return err
	}

	// Delete chat
	_, err = tx.Exec("DELETE FROM secret_chats WHERE channel_id = ?", channelID)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// CleanupExpiredSecretChats deletes all expired secret chats
func CleanupExpiredSecretChats() (int, error) {
	// Get expired chat IDs
	rows, err := database.DB.Query("SELECT channel_id FROM secret_chats WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var expiredChats []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return 0, err
		}
		expiredChats = append(expiredChats, channelID)
	}

	// Delete each expired chat
	for _, channelID := range expiredChats {
		if err := DeleteSecretChat(channelID); err != nil {
			return 0, err
		}
	}

	return len(expiredChats), nil
}
