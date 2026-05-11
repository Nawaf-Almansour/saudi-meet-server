package models

import "time"

type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type WorkspaceMember struct {
	WorkspaceID string    `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type Channel struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChannelMember struct {
	ChannelID  string    `json:"channel_id"`
	UserID     string    `json:"user_id"`
	Role       string    `json:"role"`
	JoinedAt   time.Time `json:"joined_at"`
	LastReadAt time.Time `json:"last_read_at"`
}

type Message struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channel_id"`
	SenderID    string    `json:"sender_id"`
	SenderName  string    `json:"sender_name,omitempty"`
	ParentID    *string   `json:"parent_id,omitempty"`
	Body        string    `json:"body"`
	MessageType string    `json:"message_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
