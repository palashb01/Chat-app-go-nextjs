package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// NeonDB is a simple wrapper implementing DBInterface for the client.
type NeonDB struct {
	DB *sql.DB
}

func (n *NeonDB) InsertMessage(channelID, senderID int, content string) error {
	return InsertMessage(n.DB, channelID, senderID, content)
}

func (n *NeonDB) CheckMembership(channelID, userID int) (bool, error) {
	return CheckChannelMembership(n.DB, channelID, userID)
}

// Upgrader handles HTTP -> WebSocket upgrade.
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, you'd restrict origins as needed.
		return true
	},
}

// ServeWS automatically subscribes the user to all their channels when they connect.
func ServeWS(h *Hub, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			http.Error(w, "Missing user_id query param", http.StatusBadRequest)
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}

		conn, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		// Build the client object
		client := &Client{
			hub:         h,
			conn:        conn,
			db:          &NeonDB{DB: db},
			userID:      userID,
			messageType: websocket.TextMessage,
		}

		// Fetch all channels for this user and auto-subscribe in the Hub
		channels, err := FetchUserChannels(db, userID)
		if err != nil {
			log.Println("Failed to fetch user channels:", err)
		} else {
			for _, ch := range channels {
				h.subscribe <- Subscription{
					ChannelID: ch.ID,
					Client:    client,
				}
			}
		}

		// Start reading messages in a separate goroutine
		go client.ReadPump()

		log.Printf("User %d connected. Auto-subscribed to %d channel(s)\n", userID, len(channels))
	}
}

// HandleCreateUser (POST /users) creates a new user in the database.
func HandleCreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		userID, err := CreateUser(db, body.Username)
		if err != nil {
			log.Println("CreateUser error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"user_id": userID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandleCheckIfUserExists(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		var id int
		err := db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"exists": false})
			return
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"exists": true,
			"id":     id,
		})
	}
}

// HandleGetMyChannels (GET /my_channels?user_id=123) returns channels that user belongs to.
func HandleGetMyChannels(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			http.Error(w, "Missing user_id", http.StatusBadRequest)
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		channels, err := FetchUserChannels(db, userID)
		if err != nil {
			log.Println("FetchUserChannels error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(channels)
	}
}

// HandleCreateChannel (POST /create_channel) for creating DIRECT or GROUP channels.
func HandleCreateChannel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// e.g. { "channel_type":"DIRECT", "channel_name":"", "user_ids":[1,2] }
		var req struct {
			ChannelType string `json:"channel_type"`
			ChannelName string `json:"channel_name"`
			UserIDs     []int  `json:"user_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if strings.ToUpper(req.ChannelType) == "DIRECT" && len(req.UserIDs) != 2 {
			http.Error(w, "DIRECT channel requires exactly two user IDs", http.StatusBadRequest)
			return
		}
		if strings.ToUpper(req.ChannelType) == "GROUP" && (req.ChannelName == "" || len(req.UserIDs) < 2) {
			http.Error(w, "GROUP channel requires a channel_name and at least 2 user IDs", http.StatusBadRequest)
			return
		}

		channelType := strings.ToUpper(req.ChannelType)
		channelName := req.ChannelName
		if channelType == "DIRECT" {
			channelName = "direct"
		}

		channelID, err := CreateChannel(db, channelName, channelType)
		if err != nil {
			log.Println("CreateChannel error:", err)
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		if err := AddChannelMembers(db, channelID, req.UserIDs); err != nil {
			log.Println("AddChannelMembers error:", err)
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"channel_id": channelID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HandleFetchMessages (GET /fetch_messages?channel_id=123) returns messages for a channel.
func HandleFetchMessages(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		channelIDStr := r.URL.Query().Get("channel_id")
		if channelIDStr == "" {
			http.Error(w, "Missing channel_id", http.StatusBadRequest)
			return
		}
		channelID, err := strconv.Atoi(channelIDStr)
		if err != nil {
			http.Error(w, "Invalid channel_id", http.StatusBadRequest)
			return
		}
		messages, err := FetchChannelMessages(db, channelID)
		if err != nil {
			log.Println("FetchChannelMessages error:", err)
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

// HandleAddMemberToChannel (POST /channels/{channel_id}/members) - add user to a channel.
func HandleAddMemberToChannel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		vars := mux.Vars(r)
		channelIDStr, ok := vars["channel_id"]
		if !ok {
			http.Error(w, "Missing channel_id in path", http.StatusBadRequest)
			return
		}
		channelID, err := strconv.Atoi(channelIDStr)
		if err != nil {
			http.Error(w, "Invalid channel_id", http.StatusBadRequest)
			return
		}

		var body struct {
			UserID int `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := AddChannelMembers(db, channelID, []int{body.UserID}); err != nil {
			log.Println("AddChannelMembers error:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"message":    "User added to channel",
			"channel_id": channelID,
			"added_user": body.UserID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// HealthCheckHandler just confirms the server is running.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Go + Neon + WebSocket Chat Server Running %s\n", time.Now().Format(time.RFC3339))
}
