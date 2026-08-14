package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var db *sql.DB
var rdb *redis.Client

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for WebSockets
	},
}

type ChatMessage struct {
	ID          int64  `json:"id,omitempty"`
	OrderID     int64  `json:"order_id"`
	SenderID    int64  `json:"sender_id"`
	SenderRole  string `json:"sender_role"`
	MessageText string `json:"message_text"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type Client struct {
	conn    *websocket.Conn
	orderID int64
	send    chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

var hub = Hub{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func main() {
	dbDSN := getEnv("DB_DSN", "postgres://desa_user:desa_password@postgres:5432/order_db?sslmode=disable")
	redisHost := getEnv("REDIS_HOST", "redis:6379")

	var err error
	db, err = sql.Open("postgres", dbDSN)
	if err != nil {
		log.Printf("[Chat Service] DB Error: %v", err)
	} else {
		log.Println("[Chat Service] Connected to PostgreSQL!")
	}

	rdb = redis.NewClient(&redis.Options{Addr: redisHost})

	go hub.run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/chat", handleWebSocket)
	mux.HandleFunc("/api/v1/chat/history", handleChatHistory)

	port := getEnv("PORT", "8086")
	log.Printf("[Chat Service] Running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	orderIDStr := r.URL.Query().Get("order_id")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)

	client := &Client{conn: conn, orderID: orderID, send: make(chan []byte, 256)}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ChatMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			if db != nil {
				_, _ = db.Exec("INSERT INTO chat_messages (order_id, sender_id, sender_role, message_text) VALUES ($1, $2, $3, $4)",
					msg.OrderID, msg.SenderID, msg.SenderRole, msg.MessageText)
			}
			hub.broadcast <- message
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		_ = c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

func handleChatHistory(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("order_id")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)

	var messages []ChatMessage
	if db != nil {
		rows, err := db.Query("SELECT id, order_id, sender_id, sender_role, message_text, created_at FROM chat_messages WHERE order_id = $1 ORDER BY created_at ASC", orderID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var msg ChatMessage
				_ = rows.Scan(&msg.ID, &msg.OrderID, &msg.SenderID, &msg.SenderRole, &msg.MessageText, &msg.CreatedAt)
				messages = append(messages, msg)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
