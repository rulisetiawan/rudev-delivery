package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WahaWebhookPayload struct {
	Event   string `json:"event"`
	Session string `json:"session"`
	Payload struct {
		From string `json:"from"`
		Type string `json:"type"`
		Data struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"_data"`
	} `json:"payload"`
}

type SendTextMessagePayload struct {
	Session string `json:"session"`
	ChatID  string `json:"chatId"`
	Text    string `json:"text"`
}

func main() {
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://admin:adminpassword@rabbitmq:5672/")

	// HTTP Webhook Server for WAHA
	http.HandleFunc("/api/v1/waha/webhook", handleWahaWebhook)

	// Start RabbitMQ Consumer Worker
	go startRabbitConsumer(rabbitURL)

	port := getEnv("PORT", "8083")
	log.Printf("[Notification Service] Running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleWahaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var webhook WahaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// 1. Capture Share Location from Customer
	if webhook.Event == "message" && webhook.Payload.Type == "location" {
		sender := webhook.Payload.From
		lat := webhook.Payload.Data.Lat
		lng := webhook.Payload.Data.Lng

		log.Printf("[WAHA WEBHOOK] Share Loc received from %s: Lat=%.6f, Lng=%.6f", sender, lat, lng)

		// Reply auto acknowledgement
		replyMsg := fmt.Sprintf("✅ Lokasi Anda berhasil disimpan! (Lat: %.4f, Lng: %.4f)\nPesanan Anda segera diproses & dicarikan kurir.", lat, lng)
		go sendWAHAMessage(sender, replyMsg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}

func startRabbitConsumer(rabbitURL string) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Printf("[RabbitMQ Consumer] Error connecting: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	_ = ch.ExchangeDeclare("desa.events", "topic", true, false, false, false, nil)
	q, _ := ch.QueueDeclare("q.notification.wa", true, false, false, false, nil)

	_ = ch.QueueBind(q.Name, "order.created", "desa.events", false, nil)
	_ = ch.QueueBind(q.Name, "order.cancelled", "desa.events", false, nil)
	_ = ch.QueueBind(q.Name, "order.on_delivery", "desa.events", false, nil)

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Gagal consume queue: %v", err)
	}

	log.Println("[Notification Service Worker] Consumer running on queue 'q.notification.wa'...")

	for d := range msgs {
		log.Printf("[Worker Received Event] RoutingKey: %s", d.RoutingKey)

		var payload map[string]interface{}
		_ = json.Unmarshal(d.Body, &payload)

		switch d.RoutingKey {
		case "order.created":
			custPhone, _ := payload["customer_phone"].(string)
			orderCode, _ := payload["order_code"].(string)
			if custPhone != "" {
				msg := fmt.Sprintf("🛍️ *PESANAN DITERIMA (#%s)*\n\nTerima kasih sudah memesan! Silakan *Kirim Lokasi (Share Location)* Anda di chat ini agar kurir tahu alamat rumah Anda.", orderCode)
				sendWAHAMessage(custPhone+"@c.us", msg)
			}
		case "order.on_delivery":
			orderID := payload["order_id"]
			totalAmt := payload["total_amount"]
			msg := fmt.Sprintf("🚚 *PESANAN SEDANG DIANTAR (#ORD-%v)*\n\nDriver telah selesai membeli pesanan Anda di warung.\n💰 Total Bayar COD: Rp %.0f", orderID, totalAmt)
			log.Printf("[WAHA Notification] Broadcast to Customer: %s", msg)
		case "order.cancelled":
			orderID := payload["order_id"]
			reason := payload["reason"]
			msg := fmt.Sprintf("❌ *PESANAN DIBATALKAN (#ORD-%v)*\n\nAlasan: %v", orderID, reason)
			log.Printf("[WAHA Notification] Broadcast cancellation: %s", msg)
		}

		d.Ack(false)
	}
}

func sendWAHAMessage(chatID string, text string) {
	wahaURL := getEnv("WAHA_URL", "http://waha:3000") + "/api/sendText"

	payload := SendTextMessagePayload{
		Session: "default",
		ChatID:  chatID,
		Text:    text,
	}

	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(wahaURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Error sending message to WAHA: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("WAHA Message sent to %s, Status: %s", chatID, resp.Status)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
