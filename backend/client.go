package main

import (
	"chat-app/backend/models"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket connection.
type Client struct {
    hub         *Hub
    conn        *websocket.Conn
    db          DBInterface
    userID      int
    messageType int // We'll assume we always use TextMessage
}

// DBInterface allows us to mock DB calls if needed.
type DBInterface interface {
    InsertMessage(channelID, senderID int, content string) error
    CheckMembership(channelID, userID int) (bool, error)
}

// ReadPump listens for incoming WebSocket messages from the client.
func (c *Client) ReadPump() {
    defer func() {
        // Optionally unsubscribe from all channels or handle cleanup.
        c.conn.Close()
    }()

    for {
        _, data, err := c.conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }

        var incoming models.WSIncoming
        if err := json.Unmarshal(data, &incoming); err != nil {
            log.Println("JSON parse error:", err)
            continue
        }

        switch incoming.Type {
        case "subscribe":
            // Manual subscription still possible, if you want to keep that logic
            isMember, err := c.db.CheckMembership(incoming.ChannelID, c.userID)
            if err != nil {
                log.Println("CheckMembership error:", err)
                continue
            }
            if !isMember {
                log.Printf("User %d is not a member of channel %d\n", c.userID, incoming.ChannelID)
                continue
            }
            c.hub.subscribe <- Subscription{
                ChannelID: incoming.ChannelID,
                Client:    c,
            }

        case "unsubscribe":
            c.hub.unsubscribe <- Subscription{
                ChannelID: incoming.ChannelID,
                Client:    c,
            }

        case "message":
            // Insert into DB
            err := c.db.InsertMessage(incoming.ChannelID, c.userID, incoming.Text)
            if err != nil {
                log.Println("InsertMessage error:", err)
            }
            // Broadcast to the channel
            out := models.WSOutgoing{
                Type:      "message",
                ChannelID: incoming.ChannelID,
                SenderID:  c.userID,
                Content:   incoming.Text,
                // CreatedAt can be retrieved from DB if needed, or set now
            }
            encoded, _ := json.Marshal(out)

            c.hub.broadcast <- BroadcastMessage{
                ChannelID: incoming.ChannelID,
                Data:      encoded,
            }

        default:
            log.Println("Unknown message type:", incoming.Type)
        }
    }
}
