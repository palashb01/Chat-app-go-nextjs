package models

import "time"

// User represents a row in the "users" table.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// Channel represents a row in the "channels" table.
type Channel struct {
	ID          int       `json:"id"`
	ChannelName string    `json:"channel_name"`
	ChannelType string    `json:"channel_type"` // DIRECT or GROUP
	CreatedAt   time.Time `json:"created_at"`
}

// Message represents a row in the "messages" table.
type Message struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	SenderID  int       `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// For WebSocket incoming JSON
type WSIncoming struct {
	Type      string `json:"type"`      // "subscribe", "unsubscribe", "message"
	ChannelID int    `json:"channelID"` // which channel
	Text      string `json:"text"`      // the message content
}

// For broadcasting out via WebSocket
type WSOutgoing struct {
	Type      string    `json:"type"` // "message"
	ChannelID int       `json:"channelID"`
	SenderID  int       `json:"senderID"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
