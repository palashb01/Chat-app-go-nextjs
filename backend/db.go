package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"chat-app/backend/models"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ConnectDB opens a connection to Neon (Postgres).
func ConnectDB() (*sql.DB, error) {
	// It's best practice to load the connection string from an environment variable.
	godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Fallback or example connection string
		connStr = "postgres://<USER>:<PASS>@<NEON_SUBDOMAIN>.neon.tech/<DBNAME>?sslmode=require"
		log.Println("Warning: DATABASE_URL not set. Using fallback connection string.")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB: %w", err)
	}
	// Verify connectivity
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping DB: %w", err)
	}

	log.Println("Connected to Neon (Postgres) successfully!")
	return db, nil
}

// CreateUser inserts a new user. For simplicity, no password or auth in this demo.
func CreateUser(db *sql.DB, username string) (int, error) {
	var id int
	query := `INSERT INTO users (username) VALUES ($1) RETURNING id`
	err := db.QueryRow(query, username).Scan(&id)
	return id, err
}

// CreateChannel creates a new channel row. For DIRECT, you typically won't store a channel_name.
func CreateChannel(db *sql.DB, channelName string, channelType string) (int, error) {
	var channelID int
	query := `INSERT INTO channels (channel_name, channel_type) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(query, channelName, channelType).Scan(&channelID)
	return channelID, err
}

// AddChannelMembers associates users with a channel.
func AddChannelMembers(db *sql.DB, channelID int, userIDs []int) error {
	for _, uid := range userIDs {
		_, err := db.Exec(`INSERT INTO channel_members (channel_id, user_id) VALUES ($1, $2)`, channelID, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertMessage inserts a new message into the messages table.
func InsertMessage(db *sql.DB, channelID, senderID int, content string) error {
	_, err := db.Exec(`
        INSERT INTO messages (channel_id, sender_id, content) VALUES ($1, $2, $3)
    `, channelID, senderID, content)
	return err
}

// FetchChannelMessages retrieves all messages for a channel in chronological order.
func FetchChannelMessages(db *sql.DB, channelID int) ([]models.Message, error) {
	rows, err := db.Query(`
        SELECT id, channel_id, sender_id, content, created_at
        FROM messages
        WHERE channel_id = $1
        ORDER BY created_at ASC
    `, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		err := rows.Scan(&m.ID, &m.ChannelID, &m.SenderID, &m.Content, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// CheckChannelMembership returns true if the user is a member of the given channel.
func CheckChannelMembership(db *sql.DB, channelID, userID int) (bool, error) {
	var count int
	err := db.QueryRow(`
        SELECT COUNT(*) FROM channel_members 
        WHERE channel_id = $1 AND user_id = $2
    `, channelID, userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func FetchUserChannels(db *sql.DB, userID int) ([]models.Channel, error) {
	rows, err := db.Query(`
        SELECT c.id, c.channel_name, c.channel_type, c.created_at
        FROM channel_members cm
        JOIN channels c ON cm.channel_id = c.id
        WHERE cm.user_id = $1
        ORDER BY c.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.Channel
	for rows.Next() {
		var ch models.Channel
		if err := rows.Scan(&ch.ID, &ch.ChannelName, &ch.ChannelType, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}
