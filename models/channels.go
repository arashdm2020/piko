package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrChannelNotFound is returned when a channel is not found
	ErrChannelNotFound = errors.New("channel not found")
	// ErrChannelAlreadyExists is returned when a channel with the same ID already exists
	ErrChannelAlreadyExists = errors.New("channel already exists")
	// ErrUserNotInChannel is returned when a user is not in a channel
	ErrUserNotInChannel = errors.New("user not in channel")
	// ErrUserAlreadyInChannel is returned when a user is already in a channel
	ErrUserAlreadyInChannel = errors.New("user already in channel")
	// ErrNotChannelAdmin is returned when a user is not an admin of a channel
	ErrNotChannelAdmin = errors.New("not channel admin")
)

// Channel represents a channel in the system
type Channel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	AdminAddress string    `json:"admin_address"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChannelMember represents a member of a channel
type ChannelMember struct {
	ChannelID   string    `json:"channel_id"`
	UserAddress string    `json:"user_address"`
	JoinedAt    time.Time `json:"joined_at"`
}

// ChannelMessage represents a message in a channel
type ChannelMessage struct {
	ID              string    `json:"id"`
	ChannelID       string    `json:"channel_id"`
	SenderAddress   string    `json:"sender_address"`
	EncryptedContent []byte    `json:"encrypted_content"`
	Timestamp       time.Time `json:"timestamp"`
	BlockID         *string   `json:"block_id,omitempty"`
}

// CreateChannel creates a new channel in the database
func CreateChannel(channel *Channel) error {
	// Check if channel with same ID exists
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM channels WHERE id = ?", channel.ID).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrChannelAlreadyExists
	}

	// Insert channel into database
	_, err = database.DB.Exec(
		"INSERT INTO channels (id, name, admin_address) VALUES (?, ?, ?)",
		channel.ID, channel.Name, channel.AdminAddress,
	)
	if err != nil {
		return err
	}

	// Add admin as a member
	_, err = database.DB.Exec(
		"INSERT INTO channel_members (channel_id, user_address) VALUES (?, ?)",
		channel.ID, channel.AdminAddress,
	)
	return err
}

// GetChannelByID retrieves a channel by its ID
func GetChannelByID(id string) (*Channel, error) {
	channel := &Channel{}
	err := database.DB.QueryRow(
		"SELECT id, name, admin_address, created_at FROM channels WHERE id = ?",
		id,
	).Scan(
		&channel.ID, &channel.Name, &channel.AdminAddress, &channel.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return channel, nil
}

// GetChannelsByUser retrieves all channels for a user
func GetChannelsByUser(userAddress string) ([]*Channel, error) {
	rows, err := database.DB.Query(`
		SELECT c.id, c.name, c.admin_address, c.created_at 
		FROM channels c 
		JOIN channel_members cm ON c.id = cm.channel_id 
		WHERE cm.user_address = ? 
		ORDER BY c.created_at DESC`,
		userAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	channels := []*Channel{}
	for rows.Next() {
		channel := &Channel{}
		err := rows.Scan(
			&channel.ID, &channel.Name, &channel.AdminAddress, &channel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return channels, nil
}

// UpdateChannel updates a channel's information
func UpdateChannel(channel *Channel) error {
	// Check if user is admin
	var adminAddress string
	err := database.DB.QueryRow("SELECT admin_address FROM channels WHERE id = ?", channel.ID).Scan(&adminAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrChannelNotFound
		}
		return err
	}
	if adminAddress != channel.AdminAddress {
		return ErrNotChannelAdmin
	}

	// Update channel
	_, err = database.DB.Exec(
		"UPDATE channels SET name = ? WHERE id = ?",
		channel.Name, channel.ID,
	)
	return err
}

// DeleteChannel deletes a channel by its ID
func DeleteChannel(id string, userAddress string) error {
	// Check if user is admin
	var adminAddress string
	err := database.DB.QueryRow("SELECT admin_address FROM channels WHERE id = ?", id).Scan(&adminAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrChannelNotFound
		}
		return err
	}
	if adminAddress != userAddress {
		return ErrNotChannelAdmin
	}

	// Delete channel
	_, err = database.DB.Exec("DELETE FROM channels WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Delete channel members
	_, err = database.DB.Exec("DELETE FROM channel_members WHERE channel_id = ?", id)
	if err != nil {
		return err
	}

	// Delete channel messages
	_, err = database.DB.Exec("DELETE FROM channel_messages WHERE channel_id = ?", id)
	return err
}

// AddChannelMember adds a member to a channel
func AddChannelMember(channelID string, userAddress string, adminAddress string) error {
	// Check if channel exists
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM channels WHERE id = ?", channelID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrChannelNotFound
	}

	// Check if user is admin
	var channelAdminAddress string
	err = database.DB.QueryRow("SELECT admin_address FROM channels WHERE id = ?", channelID).Scan(&channelAdminAddress)
	if err != nil {
		return err
	}
	if channelAdminAddress != adminAddress {
		return ErrNotChannelAdmin
	}

	// Check if user is already in channel
	err = database.DB.QueryRow("SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND user_address = ?", channelID, userAddress).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrUserAlreadyInChannel
	}

	// Add member
	_, err = database.DB.Exec(
		"INSERT INTO channel_members (channel_id, user_address) VALUES (?, ?)",
		channelID, userAddress,
	)
	return err
}

// RemoveChannelMember removes a member from a channel
func RemoveChannelMember(channelID string, userAddress string, adminAddress string) error {
	// Check if channel exists
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM channels WHERE id = ?", channelID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrChannelNotFound
	}

	// Check if user is admin
	var channelAdminAddress string
	err = database.DB.QueryRow("SELECT admin_address FROM channels WHERE id = ?", channelID).Scan(&channelAdminAddress)
	if err != nil {
		return err
	}
	if channelAdminAddress != adminAddress {
		return ErrNotChannelAdmin
	}

	// Check if user is in channel
	err = database.DB.QueryRow("SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND user_address = ?", channelID, userAddress).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrUserNotInChannel
	}

	// Remove member
	_, err = database.DB.Exec(
		"DELETE FROM channel_members WHERE channel_id = ? AND user_address = ?",
		channelID, userAddress,
	)
	return err
}

// IsUserInChannel checks if a user is in a channel
func IsUserInChannel(channelID string, userAddress string) (bool, error) {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND user_address = ?", channelID, userAddress).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetChannelMembers retrieves all members of a channel
func GetChannelMembers(channelID string) ([]*ChannelMember, error) {
	rows, err := database.DB.Query(
		"SELECT channel_id, user_address, joined_at FROM channel_members WHERE channel_id = ? ORDER BY joined_at",
		channelID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []*ChannelMember{}
	for rows.Next() {
		member := &ChannelMember{}
		err := rows.Scan(
			&member.ChannelID, &member.UserAddress, &member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// CreateChannelMessage creates a new channel message in the database
func CreateChannelMessage(message *ChannelMessage) error {
	// Check if user is in channel
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND user_address = ?", message.ChannelID, message.SenderAddress).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrUserNotInChannel
	}

	// Insert message
	_, err = database.DB.Exec(
		"INSERT INTO channel_messages (id, channel_id, sender_address, encrypted_content) VALUES (?, ?, ?, ?)",
		message.ID, message.ChannelID, message.SenderAddress, message.EncryptedContent,
	)
	return err
}

// GetChannelMessageByID retrieves a channel message by its ID
func GetChannelMessageByID(id string) (*ChannelMessage, error) {
	message := &ChannelMessage{}
	err := database.DB.QueryRow(
		"SELECT id, channel_id, sender_address, encrypted_content, timestamp, block_id FROM channel_messages WHERE id = ?",
		id,
	).Scan(
		&message.ID, &message.ChannelID, &message.SenderAddress, &message.EncryptedContent, &message.Timestamp, &message.BlockID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return message, nil
}

// GetChannelMessages retrieves all messages in a channel
func GetChannelMessages(channelID string, limit int, offset int) ([]*ChannelMessage, error) {
	rows, err := database.DB.Query(
		"SELECT id, channel_id, sender_address, encrypted_content, timestamp, block_id FROM channel_messages WHERE channel_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?",
		channelID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*ChannelMessage{}
	for rows.Next() {
		message := &ChannelMessage{}
		err := rows.Scan(
			&message.ID, &message.ChannelID, &message.SenderAddress, &message.EncryptedContent, &message.Timestamp, &message.BlockID,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

// UpdateChannelMessageBlockID updates the block ID of a channel message
func UpdateChannelMessageBlockID(id string, blockID string) error {
	_, err := database.DB.Exec(
		"UPDATE channel_messages SET block_id = ? WHERE id = ?",
		blockID, id,
	)
	return err
}

// DeleteChannelMessage deletes a channel message by its ID
func DeleteChannelMessage(id string, userAddress string) error {
	// Check if user is the sender or channel admin
	var senderAddress, channelID string
	err := database.DB.QueryRow("SELECT sender_address, channel_id FROM channel_messages WHERE id = ?", id).Scan(&senderAddress, &channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrMessageNotFound
		}
		return err
	}

	if senderAddress != userAddress {
		// Check if user is channel admin
		var adminAddress string
		err := database.DB.QueryRow("SELECT admin_address FROM channels WHERE id = ?", channelID).Scan(&adminAddress)
		if err != nil {
			return err
		}
		if adminAddress != userAddress {
			return ErrNotChannelAdmin
		}
	}

	// Delete message
	_, err = database.DB.Exec("DELETE FROM channel_messages WHERE id = ?", id)
	return err
} 