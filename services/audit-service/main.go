package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	dbDSN := getEnv("DB_DSN", "postgres://desa_user:desa_password@postgres:5432/order_db?sslmode=disable")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://admin:adminpassword@rabbitmq:5672/")

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatalf("DB Open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB Ping error: %v", err)
	}
	log.Println("[Audit Service] Connected to PostgreSQL!")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("RabbitMQ Dial error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ Channel error: %v", err)
	}
	defer ch.Close()

	_ = ch.ExchangeDeclare("desa.events", "topic", true, false, false, false, nil)
	q, _ := ch.QueueDeclare("q.audit.logs", true, false, false, false, nil)
	_ = ch.QueueBind(q.Name, "audit.*", "desa.events", false, nil)

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Consume error: %v", err)
	}

	log.Println("[Audit Service Worker] Listening on queue 'q.audit.logs'...")

	for d := range msgs {
		var event struct {
			EntityType      string          `json:"entity_type"`
			EntityID        string          `json:"entity_id"`
			Action          string          `json:"action"`
			PerformedByID   int64           `json:"performed_by_id"`
			PerformedByRole string          `json:"performed_by_role"`
			IPAddress       string          `json:"ip_address"`
			OldValues       json.RawMessage `json:"old_values"`
			NewValues       json.RawMessage `json:"new_values"`
		}

		if err := json.Unmarshal(d.Body, &event); err == nil {
			query := `
				INSERT INTO audit_logs (entity_type, entity_id, action, performed_by_id, performed_by_role, ip_address, old_values, new_values)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

			_, err = db.Exec(query,
				event.EntityType, event.EntityID, event.Action,
				event.PerformedByID, event.PerformedByRole, event.IPAddress,
				event.OldValues, event.NewValues,
			)

			if err == nil {
				d.Ack(false)
				log.Printf("[Audit Logged] Action: %s on %s ID: %s", event.Action, event.EntityType, event.EntityID)
				continue
			}
		}

		d.Nack(false, true)
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
