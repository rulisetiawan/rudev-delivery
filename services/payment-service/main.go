package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

var db *sql.DB
var rabbitChannel *amqp.Channel

type ChargeReq struct {
	OrderID     int64   `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
	PaymentType string  `json:"payment_type"` // 'QRIS', 'EWALLET', 'VA'
}

type MidtransWebhookPayload struct {
	TransactionStatus string `json:"transaction_status"` // 'settlement', 'capture', 'expire', 'cancel'
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	GrossAmount       string `json:"gross_amount"`
}

func main() {
	dbDSN := getEnv("DB_DSN", "postgres://desa_user:desa_password@postgres:5432/order_db?sslmode=disable")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://admin:adminpassword@rabbitmq:5672/")

	var err error
	db, err = sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}
	log.Println("[Payment Service] Connected to PostgreSQL!")

	// RabbitMQ Init
	rabbitConn, err := amqp.Dial(rabbitURL)
	if err == nil {
		rabbitChannel, _ = rabbitConn.Channel()
		_ = rabbitChannel.ExchangeDeclare("desa.events", "topic", true, false, false, false, nil)
		log.Println("[Payment Service] Connected to RabbitMQ Exchange 'desa.events'!")
	}

	mux := http.NewServeMux()

	// Payment Routes
	mux.HandleFunc("/api/v1/payments/charge", handleChargeQRIS)
	mux.HandleFunc("/api/v1/payments/webhook", handleMidtransWebhook)

	port := getEnv("PORT", "8085")
	log.Printf("[Payment Service] Running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleChargeQRIS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChargeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload tidak valid"}`, http.StatusBadRequest)
		return
	}

	txID := fmt.Sprintf("TRX-%s", uuid.New().String()[:8])
	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=00020101021226680016ID.CO.QRIS.WWW01189360091100000088215204581253033605802ID5914DESA_DELIVERY6007TASIK88%s", txID)

	if db != nil {
		_, err := db.Exec(`
			INSERT INTO payments (order_id, payment_gateway, transaction_id, payment_type, gross_amount, qr_code_url, status)
			VALUES ($1, 'MIDTRANS_SANDBOX', $2, $3, $4, $5, 'PENDING')`,
			req.OrderID, txID, req.PaymentType, req.GrossAmount, qrURL)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Gagal generate QRIS: %v"}`, err), http.StatusInternalServerError)
			return
		}

		_, _ = db.Exec("UPDATE orders SET payment_method = 'QRIS', payment_status = 'PENDING', qr_code_url = $1 WHERE id = $2", qrURL, req.OrderID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id":       req.OrderID,
		"transaction_id": txID,
		"gross_amount":   req.GrossAmount,
		"qr_code_url":    qrURL,
		"payment_type":   "QRIS",
		"message":        "Dynamic QRIS berhasil dibuat! Silakan scan untuk membayar.",
	})
}

func handleMidtransWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhook MidtransWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, `{"error":"Invalid webhook payload"}`, http.StatusBadRequest)
		return
	}

	if webhook.TransactionStatus == "settlement" || webhook.TransactionStatus == "capture" {
		orderID, _ := strconv.ParseInt(webhook.OrderID, 10, 64)

		if db != nil {
			_, _ = db.Exec("UPDATE payments SET status = 'PAID', updated_at = NOW() WHERE transaction_id = $1", webhook.TransactionID)
			_, _ = db.Exec("UPDATE orders SET payment_status = 'PAID', paid_at = NOW() WHERE id = $1", orderID)
		}

		publishEvent("order.paid", map[string]interface{}{
			"order_id":       orderID,
			"transaction_id": webhook.TransactionID,
			"paid_at":        time.Now().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Webhook pembayaran diproses!"})
}

func publishEvent(routingKey string, payload interface{}) {
	if rabbitChannel == nil {
		return
	}
	body, _ := json.Marshal(payload)
	_ = rabbitChannel.PublishWithContext(context.Background(), "desa.events", routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
