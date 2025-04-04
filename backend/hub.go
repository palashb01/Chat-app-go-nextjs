// it maintains everything the core logic of the chat messaging app system
// channels is like map<int>map<*Client> which is like 1->{client1: client1, client2: client2}
// 1 is channel id
package main

import (
    "log"
    "sync"
)

// Subscription is used for subscribe/unsubscribe events.
type Subscription struct {
    ChannelID int
    Client    *Client
}

// BroadcastMessage is a message delivered to all clients in a channel.
type BroadcastMessage struct {
    ChannelID int
    Data      []byte
}

// Hub manages multiple channels with an in-memory map: channelID -> set of clients
type Hub struct {
    channels    map[int]map[*Client]bool
    subscribe   chan Subscription
    unsubscribe chan Subscription
    broadcast   chan BroadcastMessage

    mu sync.RWMutex
}

// NewHub creates and returns a new Hub instance.
func NewHub() *Hub {
    return &Hub{
        channels:    make(map[int]map[*Client]bool),
        subscribe:   make(chan Subscription),
        unsubscribe: make(chan Subscription),
        broadcast:   make(chan BroadcastMessage),
    }
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
    for {
        select {
        case sub := <-h.subscribe:
            h.handleSubscribe(sub)
        case unsub := <-h.unsubscribe:
            h.handleUnsubscribe(unsub)
        case msg := <-h.broadcast:
            h.handleBroadcast(msg)
        }
    }
}

func (h *Hub) handleSubscribe(sub Subscription) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if h.channels[sub.ChannelID] == nil {
        h.channels[sub.ChannelID] = make(map[*Client]bool)
    }
    h.channels[sub.ChannelID][sub.Client] = true
    log.Printf("Client subscribed to channel %d", sub.ChannelID)
}

func (h *Hub) handleUnsubscribe(unsub Subscription) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if clients, ok := h.channels[unsub.ChannelID]; ok {
        delete(clients, unsub.Client)
        log.Printf("Client unsubscribed from channel %d", unsub.ChannelID)
        if len(clients) == 0 {
            delete(h.channels, unsub.ChannelID)
        }
    }
}

func (h *Hub) handleBroadcast(msg BroadcastMessage) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    if clients, ok := h.channels[msg.ChannelID]; ok {
        for client := range clients {
            err := client.conn.WriteMessage(client.messageType, msg.Data)
            if err != nil {
                log.Println("Broadcast error:", err)
                client.conn.Close()
                delete(clients, client)
            }
        }
    }
}
