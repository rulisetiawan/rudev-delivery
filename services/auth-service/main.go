package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}

type AuthRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"` // 'admin', 'courier', 'customer'
	VehiclePlate string `json:"vehicle_plate,omitempty"`
}

var db *sql.DB
var jwtSecret []byte

func main() {
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "postgres://desa_user:desa_password@postgres:5432/order_db?sslmode=disable"
	}
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	var err error
	db, err = sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatalf("DB Open error: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB Ping error: %v", err)
	}
	log.Println("[Auth Service] Connected to PostgreSQL!")

	http.HandleFunc("/api/v1/auth/register", handleRegister)
	http.HandleFunc("/api/v1/auth/login", handleLogin)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("[Auth Service] Running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "customer"
	}

	var userID int64
	err := db.QueryRow("INSERT INTO users (name, phone_number, role) VALUES ($1, $2, $3) RETURNING id",
		req.Name, req.PhoneNumber, req.Role).Scan(&userID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal mendaftarkan user: %v", err), http.StatusInternalServerError)
		return
	}

	if req.Role == "courier" {
		plate := req.VehiclePlate
		if plate == "" {
			plate = "Z 1234 XX"
		}
		_, _ = db.Exec("INSERT INTO courier_profiles (user_id, vehicle_plate) VALUES ($1, $2)", userID, plate)
	}

	token, _ := generateJWT(userID, req.Role)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"name":    req.Name,
		"role":    req.Role,
		"token":   token,
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow("SELECT id, name, phone_number, role FROM users WHERE phone_number = $1", req.PhoneNumber).
		Scan(&user.ID, &user.Name, &user.PhoneNumber, &user.Role)

	if err != nil {
		http.Error(w, "User tidak ditemukan. Silakan registrasi dahulu.", http.StatusNotFound)
		return
	}

	token, _ := generateJWT(user.ID, user.Role)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func generateJWT(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
