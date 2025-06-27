package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrMessageNotFound is returned when a message is not found
	ErrMessageNotFound = errors.New("message not found")
)

// MessageStatus represents the status of a message
type MessageStatus string

const (
	// MessageStatusPending indicates the message is pending delivery
	MessageStatusPending MessageStatus = "pending"
	// MessageStatusDelivered indicates the message has been delivered
	MessageStatusDelivered MessageStatus = "delivered"
	// MessageStatusRead indicates the message has been read
	MessageStatusRead MessageStatus = "read"
)

// Message represents a message in the system
type Message struct {
	ID              string       `json:"id"`
	SenderAddress   string       `json:"sender_address"`
	RecipientAddress string      `json:"recipient_address"`
	EncryptedContent []byte      `json:"encrypted_content"`
	Timestamp       time.Time    `json:"timestamp"`
	Status          MessageStatus `json:"status"`
	ExpirationTime  *time.Time   `json:"expiration_time,omitempty"`
	BlockID         *string      `json:"block_id,omitempty"`
}

// CreateMessage creates a new message in the database
func CreateMessage(message *Message) error {
	_, err := database.DB.Exec(
		"INSERT INTO messages (id, sender_address, recipient_address, encrypted_content, status, expiration_time) VALUES (?, ?, ?, ?, ?, ?)",
		message.ID, message.SenderAddress, message.RecipientAddress, message.EncryptedContent, message.Status, message.ExpirationTime,
	)
	return err
}

// GetMessageByID retrieves a message by its ID
func GetMessageByID(id string) (*Message, error) {
	message := &Message{}
	var status string
	err := database.DB.QueryRow(
		"SELECT id, sender_address, recipient_address, encrypted_content, timestamp, status, expiration_time, block_id FROM messages WHERE id = ?",
		id,
	).Scan(
		&message.ID, &message.SenderAddress, &message.RecipientAddress, &message.EncryptedContent, &message.Timestamp, &status, &message.ExpirationTime, &message.BlockID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	message.Status = MessageStatus(status)
	return message, nil
}

// GetMessagesByRecipient retrieves all messages for a recipient
func GetMessagesByRecipient(recipientAddress string) ([]*Message, error) {
	rows, err := database.DB.Query(
		"SELECT id, sender_address, recipient_address, encrypted_content, timestamp, status, expiration_time, block_id FROM messages WHERE recipient_address = ? ORDER BY timestamp DESC",
		recipientAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		message := &Message{}
		var status string
		err := rows.Scan(
			&message.ID, &message.SenderAddress, &message.RecipientAddress, &message.EncryptedContent, &message.Timestamp, &status, &message.ExpirationTime, &message.BlockID,
		)
		if err != nil {
			return nil, err
		}
		message.Status = MessageStatus(status)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

// GetMessagesBySender retrieves all messages sent by a sender
func GetMessagesBySender(senderAddress string) ([]*Message, error) {
	rows, err := database.DB.Query(
		"SELECT id, sender_address, recipient_address, encrypted_content, timestamp, status, expiration_time, block_id FROM messages WHERE sender_address = ? ORDER BY timestamp DESC",
		senderAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*Message{}
	for rows.Next() {
		message := &Message{}
		var status string
		err := rows.Scan(
			&message.ID, &message.SenderAddress, &message.RecipientAddress, &message.EncryptedContent, &message.Timestamp, &status, &message.ExpirationTime, &message.BlockID,
		)
		if err != nil {
			return nil, err
		}
		message.Status = MessageStatus(status)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

// UpdateMessageStatus updates the status of a message
func UpdateMessageStatus(id string, status MessageStatus) error {
	_, err := database.DB.Exec(
		"UPDATE messages SET status = ? WHERE id = ?",
		status, id,
	)
	return err
}

// UpdateMessageBlockID updates the block ID of a message
func UpdateMessageBlockID(id string, blockID string) error {
	_, err := database.DB.Exec(
		"UPDATE messages SET block_id = ? WHERE id = ?",
		blockID, id,
	)
	return err
}

// DeleteMessage deletes a message by its ID
func DeleteMessage(id string) error {
	_, err := database.DB.Exec("DELETE FROM messages WHERE id = ?", id)
	return err
}

// DeleteExpiredMessages deletes all expired messages
func DeleteExpiredMessages() error {
	_, err := database.DB.Exec("DELETE FROM messages WHERE expiration_time IS NOT NULL AND expiration_time < NOW()")
	return err
} 