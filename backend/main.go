package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// 1) Connect to Neon DB
	db, err := ConnectDB()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// 2) Create our Hub and start its goroutine
	hub := NewHub()
	go hub.Run()

	// 3) Set up a gorilla/mux Router
	r := mux.NewRouter()
	r.Use(RateLimitMiddleware)
	// Health check
	r.HandleFunc("/", HealthCheckHandler).Methods("GET")

	// WebSocket
	r.HandleFunc("/ws", ServeWS(hub, db)).Methods("GET")

	// Channels
	r.HandleFunc("/create_channel", HandleCreateChannel(db)).Methods("POST")
	r.HandleFunc("/fetch_messages", HandleFetchMessages(db)).Methods("GET")
	r.HandleFunc("/channels/{channel_id}/members", HandleAddMemberToChannel(db)).Methods("POST")

	// Users
	r.HandleFunc("/users", HandleCreateUser(db)).Methods("POST")
	r.HandleFunc("/my_channels", HandleGetMyChannels(db)).Methods("GET")
	r.HandleFunc("/check_user", HandleCheckIfUserExists(db)).Methods("GET")

	// 4) Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// 5) Start the server with CORS middleware
	addr := ":8080"
	log.Println("Server running on", addr)
	log.Fatal(http.ListenAndServe(addr, c.Handler(r)))
}
