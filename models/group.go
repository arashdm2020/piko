package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/piko/piko/database"
)

var (
	// ErrGroupNotFound is returned when a group is not found
	ErrGroupNotFound = errors.New("group not found")
	// ErrGroupMemberNotFound is returned when a group member is not found
	ErrGroupMemberNotFound = errors.New("group member not found")
	// ErrNotGroupAdmin is returned when a user is not a group admin
	ErrNotGroupAdmin = errors.New("user is not a group admin")
	// ErrAlreadyGroupMember is returned when a user is already a group member
	ErrAlreadyGroupMember = errors.New("user is already a group member")
)

// GroupRole defines the role of a user in a group
type GroupRole string

const (
	// GroupRoleAdmin is the role for group admins
	GroupRoleAdmin GroupRole = "admin"
	// GroupRoleMember is the role for regular group members
	GroupRoleMember GroupRole = "member"
)

// Group represents a group chat
type Group struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CreatorAddress string    `json:"creator_address"`
	PhotoURL       string    `json:"photo_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	MemberCount    int       `json:"member_count"`
}

// GroupMember represents a member of a group
type GroupMember struct {
	GroupID     string    `json:"group_id"`
	UserAddress string    `json:"user_address"`
	Role        GroupRole `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// GroupMessage represents a message in a group
type GroupMessage struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"group_id"`
	SenderAddress string    `json:"sender_address"`
	Content       []byte    `json:"content"`
	Timestamp     time.Time `json:"timestamp"`
	BlockID       *string   `json:"block_id,omitempty"`
}

// CreateGroup creates a new group
func CreateGroup(group *Group, creatorAddress string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert group
	_, err = tx.Exec(
		"INSERT INTO groups (id, name, description, creator_address, photo_url) VALUES (?, ?, ?, ?, ?)",
		group.ID, group.Name, group.Description, creatorAddress, group.PhotoURL,
	)
	if err != nil {
		return err
	}

	// Add creator as admin
	_, err = tx.Exec(
		"INSERT INTO group_members (group_id, user_address, role) VALUES (?, ?, ?)",
		group.ID, creatorAddress, GroupRoleAdmin,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetGroupByID retrieves a group by its ID
func GetGroupByID(id string) (*Group, error) {
	group := &Group{}
	err := database.DB.QueryRow(
		`SELECT g.id, g.name, g.description, g.creator_address, g.photo_url, g.created_at, g.updated_at, 
		(SELECT COUNT(*) FROM group_members WHERE group_id = g.id) as member_count 
		FROM groups g WHERE g.id = ?`,
		id,
	).Scan(
		&group.ID, &group.Name, &group.Description, &group.CreatorAddress, &group.PhotoURL,
		&group.CreatedAt, &group.UpdatedAt, &group.MemberCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return group, nil
}

// GetUserGroups retrieves all groups a user is a member of
func GetUserGroups(userAddress string) ([]*Group, error) {
	rows, err := database.DB.Query(
		`SELECT g.id, g.name, g.description, g.creator_address, g.photo_url, g.created_at, g.updated_at, 
		(SELECT COUNT(*) FROM group_members WHERE group_id = g.id) as member_count 
		FROM groups g 
		JOIN group_members gm ON g.id = gm.group_id 
		WHERE gm.user_address = ? 
		ORDER BY g.updated_at DESC`,
		userAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []*Group{}
	for rows.Next() {
		group := &Group{}
		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.CreatorAddress, &group.PhotoURL,
			&group.CreatedAt, &group.UpdatedAt, &group.MemberCount,
		)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

// UpdateGroup updates a group's information
func UpdateGroup(group *Group) error {
	_, err := database.DB.Exec(
		"UPDATE groups SET name = ?, description = ?, photo_url = ?, updated_at = NOW() WHERE id = ?",
		group.Name, group.Description, group.PhotoURL, group.ID,
	)
	return err
}

// DeleteGroup deletes a group
func DeleteGroup(id string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete group members
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = ?", id)
	if err != nil {
		return err
	}

	// Delete group messages
	_, err = tx.Exec("DELETE FROM group_messages WHERE group_id = ?", id)
	if err != nil {
		return err
	}

	// Delete group
	_, err = tx.Exec("DELETE FROM groups WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// AddGroupMember adds a member to a group
func AddGroupMember(groupID, userAddress string, role GroupRole) error {
	// Check if user is already a member
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_address = ?",
		groupID, userAddress).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAlreadyGroupMember
	}

	// Add member
	_, err = database.DB.Exec(
		"INSERT INTO group_members (group_id, user_address, role) VALUES (?, ?, ?)",
		groupID, userAddress, role,
	)
	return err
}

// RemoveGroupMember removes a member from a group
func RemoveGroupMember(groupID, userAddress string) error {
	result, err := database.DB.Exec(
		"DELETE FROM group_members WHERE group_id = ? AND user_address = ?",
		groupID, userAddress,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrGroupMemberNotFound
	}

	return nil
}

// GetGroupMembers retrieves all members of a group
func GetGroupMembers(groupID string) ([]*GroupMember, error) {
	rows, err := database.DB.Query(
		"SELECT group_id, user_address, role, joined_at FROM group_members WHERE group_id = ?",
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []*GroupMember{}
	for rows.Next() {
		member := &GroupMember{}
		var role string
		err := rows.Scan(&member.GroupID, &member.UserAddress, &role, &member.JoinedAt)
		if err != nil {
			return nil, err
		}
		member.Role = GroupRole(role)
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// IsGroupAdmin checks if a user is an admin of a group
func IsGroupAdmin(groupID, userAddress string) (bool, error) {
	var role string
	err := database.DB.QueryRow(
		"SELECT role FROM group_members WHERE group_id = ? AND user_address = ?",
		groupID, userAddress,
	).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrGroupMemberNotFound
		}
		return false, err
	}

	return GroupRole(role) == GroupRoleAdmin, nil
}

// UpdateMemberRole updates a member's role in a group
func UpdateMemberRole(groupID, userAddress string, role GroupRole) error {
	result, err := database.DB.Exec(
		"UPDATE group_members SET role = ? WHERE group_id = ? AND user_address = ?",
		role, groupID, userAddress,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrGroupMemberNotFound
	}

	return nil
}

// CreateGroupMessage creates a new message in a group
func CreateGroupMessage(message *GroupMessage) error {
	_, err := database.DB.Exec(
		"INSERT INTO group_messages (id, group_id, sender_address, content) VALUES (?, ?, ?, ?)",
		message.ID, message.GroupID, message.SenderAddress, message.Content,
	)
	return err
}

// GetGroupMessages retrieves messages from a group
func GetGroupMessages(groupID string, limit, offset int) ([]*GroupMessage, error) {
	rows, err := database.DB.Query(
		"SELECT id, group_id, sender_address, content, timestamp, block_id FROM group_messages WHERE group_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?",
		groupID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []*GroupMessage{}
	for rows.Next() {
		message := &GroupMessage{}
		err := rows.Scan(
			&message.ID, &message.GroupID, &message.SenderAddress, &message.Content,
			&message.Timestamp, &message.BlockID,
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

// DeleteGroupMessage deletes a message from a group
func DeleteGroupMessage(id string) error {
	_, err := database.DB.Exec("DELETE FROM group_messages WHERE id = ?", id)
	return err
}
