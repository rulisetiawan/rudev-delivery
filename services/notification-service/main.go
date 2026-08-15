package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var db *sql.DB

type WahaWebhookPayload struct {
	Event   string      `json:"event"` // 'message', 'session.status'
	Session string      `json:"session"`
	Payload WahaMessage `json:"payload"`
}

type WahaMessage struct {
	ID       string        `json:"id"`
	From     string        `json:"from"`
	Body     string        `json:"body"`
	HasMedia bool          `json:"hasMedia"`
	Location *WahaLocation `json:"location,omitempty"`
}

type WahaLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Description string `json:"description,omitempty"`
}

func main() {
	dbDSN := getEnv("DB_DSN", "postgres://desa_user:desa_password@postgres:5432/order_db?sslmode=disable")
	wahaURL := getEnv("WAHA_URL", "http://waha:3000")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://admin:adminpassword@rabbitmq:5672/")

	var err error
	db, err = sql.Open("postgres", dbDSN)
	if err != nil {
		log.Printf("[Notification Service] DB Warning: %v", err)
	}

	// Connect RabbitMQ Consumer
	go startRabbitMQConsumer(rabbitURL, wahaURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/waha/webhook", func(w http.ResponseWriter, r *http.Request) {
		handleWahaWebhookWithBot(w, r, wahaURL)
	})

	port := getEnv("PORT", "8083")
	log.Printf("[Notification Service] Running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func parseWahaBotCommand(bodyStr string) string {
	cmd := strings.ToUpper(strings.TrimSpace(bodyStr))
	if strings.Contains(cmd, "CEK ORDER") || strings.Contains(cmd, "STATUS") {
		return "🤖 *BOT DESA*: Pesanan Anda sedang diproses oleh kurir terdekat! Silakan pantau status via PWA."
	}
	if strings.Contains(cmd, "WARUNG") || strings.Contains(cmd, "MENU") {
		return "🤖 *BOT DESA*: Warung mitra buka hari ini: Warung Bu Ani, Toko Kelontong Pak Eko. Akses katalog lengkap di http://localhost/"
	}
	if strings.Contains(cmd, "TARIF") || strings.Contains(cmd, "ONGKIR") {
		return "🤖 *BOT DESA*: Tarif ongkir desa flat Rp 10.000 + Rp 2.000 per toko tambahan."
	}
	if strings.Contains(cmd, "HALO") || strings.Contains(cmd, "HELP") || strings.Contains(cmd, "BANTUAN") {
		return "🤖 *BOT DESA*: Selamat datang di Pesan Antar Desa!\nKetik:\n- *CEK ORDER* (Status pesanan)\n- *WARUNG* (Daftar toko)\n- *TARIF* (Informasi ongkir)"
	}
	return ""
}

func handleWahaWebhookWithBot(w http.ResponseWriter, r *http.Request, wahaURL string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WahaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"Payload tidak valid"}`, http.StatusBadRequest)
		return
	}

	if payload.Event == "message" && payload.Payload.From != "" {
		phone := strings.Replace(payload.Payload.From, "@c.us", "", 1)

		// 1. Handle Share Location
		if payload.Payload.Location != nil {
			log.Printf("[WAHA WEBHOOK] Share Loc from %s: Lat=%f, Lng=%f",
				phone, payload.Payload.Location.Latitude, payload.Payload.Location.Longitude)

			if db != nil {
				_, _ = db.Exec("UPDATE orders SET latitude = $1, longitude = $2, status = 'SEARCHING_COURIER' WHERE customer_id IN (SELECT id FROM users WHERE phone_number = $3) AND status = 'WAITING_LOCATION'",
					payload.Payload.Location.Latitude, payload.Payload.Location.Longitude, phone)
			}
			sendWahaText(wahaURL, phone, "📍 Lokasi pengantaran Anda berhasil diterima! Kurir terdekat sedang disiapkan.")
		}

		// 2. Handle Auto-Bot Interactive CS Commands (Phase 4)
		botReply := parseWahaBotCommand(payload.Payload.Body)
		if botReply != "" {
			sendWahaText(wahaURL, phone, botReply)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Webhook WAHA diproses!"})
}

func sendWahaText(wahaURL, phone, text string) {
	reqBody, _ := json.Marshal(map[string]string{
		"chatId":  phone + "@c.us",
		"text":    text,
		"session": "default",
	})
	resp, err := http.Post(wahaURL+"/api/sendText", "application/json", bytes.NewBuffer(reqBody))
	if err == nil {
		defer resp.Body.Close()
	}
}

func startRabbitMQConsumer(rabbitURL, wahaURL string) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Printf("[RabbitMQ Consumer Warning]: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	q, _ := ch.QueueDeclare("waha_notifications", true, false, false, false, nil)
	_ = ch.QueueBind(q.Name, "order.*", "desa.events", false, nil)

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return
	}

	log.Println("[Notification Service] RabbitMQ Notification Consumer Listening...")
	for d := range msgs {
		var event map[string]interface{}
		if err := json.Unmarshal(d.Body, &event); err == nil {
			orderIDFloat, _ := event["order_id"].(float64)
			orderID := int64(orderIDFloat)

			if d.RoutingKey == "order.assigned" {
				courierIDFloat, _ := event["courier_id"].(float64)
				sendWahaText(wahaURL, "6281234567890", fmt.Sprintf("🛵 *PESANAN DESA #%d*: Kurir #%d telah ditugaskan secara otomatis!", orderID, int64(courierIDFloat)))
			} else if d.RoutingKey == "order.paid" {
				sendWahaText(wahaURL, "6281234567890", fmt.Sprintf("💳 *PEMBAYARAN LUNAS #%d*: Pembayaran Dynamic QRIS terverifikasi!", orderID))
			}
		}
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
