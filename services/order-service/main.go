package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nfnt/resize"
	amqp "github.com/rabbitmq/amqp091-go"
)

var db *sql.DB
var minioClient *minio.Client
var rabbitChannel *amqp.Channel
var minioBucket string
var minioPublicURL string

type Store struct {
	ID        int64  `json:"id"`
	StoreName string `json:"store_name"`
	Slug      string `json:"slug"`
	Address   string `json:"address_text"`
	ImageURL  string `json:"image_url"`
	IsPartner bool   `json:"is_partner"`
	IsActive  bool   `json:"is_active"`
}

type Product struct {
	ID          int64   `json:"id"`
	StoreID     int64   `json:"store_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	IsAvailable bool    `json:"is_available"`
	ImageURL    string  `json:"image_url"`
}

type OrderItemReq struct {
	ProductID    int64   `json:"product_id"`
	ItemName     string  `json:"item_name"`
	Quantity     int     `json:"quantity"`
	PricePerItem float64 `json:"price_per_item"`
	Notes        string  `json:"notes"`
	IsCustom     bool    `json:"is_custom_item"`
}

type CheckoutReq struct {
	StoreID       int64          `json:"store_id"`
	CustomerName  string         `json:"customer_name"`
	CustomerPhone string         `json:"customer_phone"`
	AddressText   string         `json:"delivery_address_text"`
	Items         []OrderItemReq `json:"items"`
}

type TariffSettings struct {
	BaseDeliveryFee  float64 `json:"base_delivery_fee"`
	PerExtraStoreFee float64 `json:"per_extra_store_fee"`
	MinOrderAmount   float64 `json:"min_order_amount"`
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
	log.Println("[Order Service] Connected to PostgreSQL!")

	// MinIO Init
	minioEndpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	minioAccess := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	minioSecret := getEnv("MINIO_SECRET_KEY", "miniopassword")
	minioBucket = getEnv("MINIO_BUCKET_NAME", "desa-media")
	minioPublicURL = getEnv("MINIO_PUBLIC_URL", "http://localhost:9000")

	minioClient, err = minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccess, minioSecret, ""),
		Secure: false,
	})
	if err != nil {
		log.Printf("[WARNING] MinIO init error: %v", err)
	} else {
		log.Println("[Order Service] Connected to MinIO!")
	}

	// RabbitMQ Init
	rabbitConn, err := amqp.Dial(rabbitURL)
	if err == nil {
		rabbitChannel, _ = rabbitConn.Channel()
		_ = rabbitChannel.ExchangeDeclare("desa.events", "topic", true, false, false, false, nil)
		log.Println("[Order Service] Connected to RabbitMQ Exchange 'desa.events'!")
	}

	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/api/v1/public/stores", handleGetStores)
	mux.HandleFunc("/api/v1/public/products", handleGetProducts)
	mux.HandleFunc("/api/v1/public/share/store/", handleShareStoreOG)
	mux.HandleFunc("/api/v1/public/checkout", handleCheckout)
	mux.HandleFunc("/api/v1/public/settings/tariff", handleGetTariffSettings)

	// Protected Admin CRUD Routes
	mux.HandleFunc("/api/v1/orders/admin/stores", handleAdminStoreCRUD)
	mux.HandleFunc("/api/v1/orders/admin/products", handleAdminProductCRUD)
	mux.HandleFunc("/api/v1/orders/admin/products/upload-image", handleAdminProductUploadImage)
	mux.HandleFunc("/api/v1/orders/admin/couriers", handleAdminCouriersCRUD)
	mux.HandleFunc("/api/v1/orders/admin/couriers/loan", handleAdminCourierLoan)
	mux.HandleFunc("/api/v1/orders/admin/settings/tariff", handleAdminUpdateTariffSettings)

	// Protected Courier Routes
	mux.HandleFunc("/api/v1/orders/courier/receipt", handleSubmitReceipt)
	mux.HandleFunc("/api/v1/orders/cancel", handleCancelOrder)

	port := getEnv("PORT", "8082")
	log.Printf("[Order Service] Running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getDynamicTariffs() TariffSettings {
	baseFee := 10000.0
	extraFee := 2000.0
	minAmt := 0.0

	_ = db.QueryRow("SELECT CAST(value AS NUMERIC) FROM system_settings WHERE key = 'base_delivery_fee'").Scan(&baseFee)
	_ = db.QueryRow("SELECT CAST(value AS NUMERIC) FROM system_settings WHERE key = 'per_extra_store_fee'").Scan(&extraFee)
	_ = db.QueryRow("SELECT CAST(value AS NUMERIC) FROM system_settings WHERE key = 'min_order_amount'").Scan(&minAmt)

	return TariffSettings{
		BaseDeliveryFee:  baseFee,
		PerExtraStoreFee: extraFee,
		MinOrderAmount:   minAmt,
	}
}

func handleGetTariffSettings(w http.ResponseWriter, r *http.Request) {
	tariffs := getDynamicTariffs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tariffs)
}

func handleAdminUpdateTariffSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TariffSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload tidak valid"}`, http.StatusBadRequest)
		return
	}

	_, _ = db.Exec("INSERT INTO system_settings (key, value) VALUES ('base_delivery_fee', $1) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", fmt.Sprintf("%.2f", req.BaseDeliveryFee))
	_, _ = db.Exec("INSERT INTO system_settings (key, value) VALUES ('per_extra_store_fee', $1) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", fmt.Sprintf("%.2f", req.PerExtraStoreFee))
	_, _ = db.Exec("INSERT INTO system_settings (key, value) VALUES ('min_order_amount', $1) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value", fmt.Sprintf("%.2f", req.MinOrderAmount))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Pengaturan tarif pengantaran berhasil diperbarui!",
		"data":    req,
	})
}

func handleGetStores(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, store_name, slug, COALESCE(address_text,''), COALESCE(image_url,''), is_partner, is_active FROM stores WHERE is_active = true ORDER BY id DESC")
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var s Store
		_ = rows.Scan(&s.ID, &s.StoreName, &s.Slug, &s.Address, &s.ImageURL, &s.IsPartner, &s.IsActive)
		stores = append(stores, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stores)
}

func handleGetProducts(w http.ResponseWriter, r *http.Request) {
	storeIDStr := r.URL.Query().Get("store_id")
	query := "SELECT id, store_id, name, price, category, is_available, COALESCE(image_url,'') FROM products WHERE is_available = true"
	if storeIDStr != "" {
		query += " AND store_id = " + storeIDStr
	}
	query += " ORDER BY id DESC"

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		_ = rows.Scan(&p.ID, &p.StoreID, &p.Name, &p.Price, &p.Category, &p.IsAvailable, &p.ImageURL)
		products = append(products, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func handleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckoutReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var customerID int64
	err := db.QueryRow("SELECT id FROM users WHERE phone_number = $1", req.CustomerPhone).Scan(&customerID)
	if err != nil {
		_ = db.QueryRow("INSERT INTO users (name, phone_number, role) VALUES ($1, $2, 'customer') RETURNING id",
			req.CustomerName, req.CustomerPhone).Scan(&customerID)
	}

	var subtotal float64
	for _, item := range req.Items {
		subtotal += item.PricePerItem * float64(item.Quantity)
	}

	// Dynamic Delivery Tariff Calculation
	tariffs := getDynamicTariffs()
	totalStores := 1
	deliveryFee := tariffs.BaseDeliveryFee + float64(totalStores-1)*tariffs.PerExtraStoreFee
	totalAmount := subtotal + deliveryFee
	orderCode := fmt.Sprintf("ORD-%d", time.Now().UnixNano()%1000000)

	var orderID int64
	err = db.QueryRow(`
		INSERT INTO orders (order_code, customer_id, store_id, status, delivery_address_text, subtotal, base_delivery_fee, store_add_fee, delivery_fee, total_amount, payment_method)
		VALUES ($1, $2, $3, 'CREATED', $4, $5, $6, $7, $8, $9, 'COD') RETURNING id`,
		orderCode, customerID, req.StoreID, req.AddressText, subtotal, tariffs.BaseDeliveryFee, tariffs.PerExtraStoreFee, deliveryFee, totalAmount).Scan(&orderID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Gagal buat order: %v", err), http.StatusInternalServerError)
		return
	}

	for _, item := range req.Items {
		var productID *int64
		if item.ProductID > 0 {
			productID = &item.ProductID
		}
		_, _ = db.Exec(`
			INSERT INTO order_items (order_id, product_id, item_name, quantity, price_per_item, notes, is_custom_item)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			orderID, productID, item.ItemName, item.Quantity, item.PricePerItem, item.Notes, item.IsCustom)
	}

	publishEvent("order.created", map[string]interface{}{
		"order_id":       orderID,
		"order_code":     orderCode,
		"customer_name":  req.CustomerName,
		"customer_phone": req.CustomerPhone,
		"total_amount":   totalAmount,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id":       orderID,
		"order_code":     orderCode,
		"total_amount":   totalAmount,
		"delivery_fee":   deliveryFee,
		"waha_bot_number": "6281234567890",
		"message":        "Pesanan berhasil dibuat! Silakan share location di WhatsApp.",
	})
}

func handleShareStoreOG(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/public/share/store/")

	var name, imgURL string
	var err error
	if db != nil {
		err = db.QueryRow("SELECT store_name, COALESCE(image_url,'') FROM stores WHERE slug = $1", slug).Scan(&name, &imgURL)
	}
	if db == nil || err != nil {
		name = "Warung Desa"
		imgURL = "https://via.placeholder.com/600x400"
	}

	userAgent := strings.ToLower(r.UserAgent())
	if strings.Contains(userAgent, "whatsapp") || strings.Contains(userAgent, "facebookexternalhit") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <meta property="og:title" content="🛒 %s - Pesan Antar Desa" />
    <meta property="og:description" content="Pesan makanan & sembako dari %s se-Desa! Siap antar cepat." />
    <meta property="og:image" content="%s" />
</head>
<body><p>%s</p></body>
</html>`, name, name, imgURL, name)
		return
	}

	http.Redirect(w, r, "http://localhost/#/store/"+slug, http.StatusFound)
}

func handleAdminStoreCRUD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, store_name, slug, COALESCE(address_text,''), COALESCE(image_url,''), is_partner, is_active FROM stores ORDER BY id DESC")
		if err != nil {
			http.Error(w, `{"error":"DB Error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var stores []Store
		for rows.Next() {
			var s Store
			_ = rows.Scan(&s.ID, &s.StoreName, &s.Slug, &s.Address, &s.ImageURL, &s.IsPartner, &s.IsActive)
			stores = append(stores, s)
		}
		json.NewEncoder(w).Encode(stores)

	case http.MethodPost:
		var s Store
		_ = json.NewDecoder(r.Body).Decode(&s)
		if s.Slug == "" {
			s.Slug = strings.ToLower(strings.ReplaceAll(s.StoreName, " ", "-"))
		}
		var id int64
		err := db.QueryRow("INSERT INTO stores (store_name, slug, address_text, image_url, is_partner, is_active) VALUES ($1, $2, $3, $4, $5, true) RETURNING id",
			s.StoreName, s.Slug, s.Address, s.ImageURL, s.IsPartner).Scan(&id)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Gagal tambah warung: %v"}`, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "message": "Warung berhasil ditambahkan!"})

	case http.MethodPut:
		var s Store
		_ = json.NewDecoder(r.Body).Decode(&s)
		_, err := db.Exec("UPDATE stores SET store_name = $1, address_text = $2, image_url = $3, is_partner = $4, is_active = $5 WHERE id = $6",
			s.StoreName, s.Address, s.ImageURL, s.IsPartner, s.IsActive, s.ID)

		if err != nil {
			http.Error(w, `{"error":"Gagal update warung"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Data warung berhasil diperbarui!"})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		_, _ = db.Exec("UPDATE stores SET is_active = false WHERE id = $1", id)
		json.NewEncoder(w).Encode(map[string]string{"message": "Warung dinonaktifkan."})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAdminProductCRUD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		storeIDStr := r.URL.Query().Get("store_id")
		query := "SELECT id, store_id, name, price, category, is_available, COALESCE(image_url,'') FROM products"
		if storeIDStr != "" {
			query += " WHERE store_id = " + storeIDStr
		}
		query += " ORDER BY id DESC"

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, `{"error":"DB Error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var products []Product
		for rows.Next() {
			var p Product
			_ = rows.Scan(&p.ID, &p.StoreID, &p.Name, &p.Price, &p.Category, &p.IsAvailable, &p.ImageURL)
			products = append(products, p)
		}
		json.NewEncoder(w).Encode(products)

	case http.MethodPost:
		var p Product
		_ = json.NewDecoder(r.Body).Decode(&p)
		var id int64
		err := db.QueryRow("INSERT INTO products (store_id, name, price, category, image_url, is_available) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
			p.StoreID, p.Name, p.Price, p.Category, p.ImageURL, p.IsAvailable).Scan(&id)

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Gagal tambah produk: %v"}`, err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "message": "Produk berhasil ditambahkan ke katalog!"})

	case http.MethodPut:
		var p Product
		_ = json.NewDecoder(r.Body).Decode(&p)
		_, err := db.Exec("UPDATE products SET name = $1, price = $2, category = $3, image_url = $4, is_available = $5 WHERE id = $6",
			p.Name, p.Price, p.Category, p.ImageURL, p.IsAvailable, p.ID)

		if err != nil {
			http.Error(w, `{"error":"Gagal update produk"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Produk berhasil diperbarui!"})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		_, _ = db.Exec("DELETE FROM products WHERE id = $1", id)
		json.NewEncoder(w).Encode(map[string]string{"message": "Produk berhasil dihapus."})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAdminProductUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, `{"error":"File gambar wajib diunggah"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageURL := uploadToMinIO(fileHeader)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Gambar produk berhasil diunggah ke MinIO!",
		"image_url": imageURL,
	})
}

func handleAdminCouriersCRUD(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query(`
			SELECT u.id, u.name, u.phone_number, cp.vehicle_plate, cp.is_active, cp.is_online, cp.current_status, cp.total_deliveries
			FROM users u JOIN courier_profiles cp ON u.id = cp.user_id WHERE u.role = 'courier' ORDER BY u.id DESC`)
		if err != nil {
			http.Error(w, `{"error":"DB Error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var result []map[string]interface{}
		for rows.Next() {
			var id int64
			var name, phone, plate, status string
			var active, online bool
			var total int
			_ = rows.Scan(&id, &name, &phone, &plate, &active, &online, &status, &total)

			result = append(result, map[string]interface{}{
				"id": id, "name": name, "phone_number": phone, "vehicle_plate": plate,
				"is_active": active, "is_online": online, "status": status, "total_deliveries": total,
			})
		}
		json.NewEncoder(w).Encode(result)

	case http.MethodPost:
		var req struct {
			Name         string `json:"name"`
			PhoneNumber  string `json:"phone_number"`
			VehiclePlate string `json:"vehicle_plate"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		var userID int64
		err := db.QueryRow("INSERT INTO users (name, phone_number, role) VALUES ($1, $2, 'courier') RETURNING id",
			req.Name, req.PhoneNumber).Scan(&userID)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Gagal tambah driver: %v"}`, err), http.StatusInternalServerError)
			return
		}

		plate := req.VehiclePlate
		if plate == "" {
			plate = "Z 1234 XX"
		}
		_, _ = db.Exec("INSERT INTO courier_profiles (user_id, vehicle_plate) VALUES ($1, $2)", userID, plate)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":       userID,
			"name":          req.Name,
			"phone_number":   req.PhoneNumber,
			"vehicle_plate": plate,
			"message":       "Driver baru berhasil terdaftar!",
		})

	case http.MethodPut:
		var req struct {
			ID       int64 `json:"id"`
			IsActive bool  `json:"is_active"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		_, err := db.Exec("UPDATE courier_profiles SET is_active = $1 WHERE user_id = $2", req.IsActive, req.ID)
		if err != nil {
			http.Error(w, `{"error":"Gagal update status driver"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Status driver berhasil diperbarui!"})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		_, _ = db.Exec("UPDATE courier_profiles SET is_active = false WHERE user_id = $1", id)
		json.NewEncoder(w).Encode(map[string]string{"message": "Driver berhasil dinonaktifkan."})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAdminCourierLoan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CourierID int64   `json:"courier_id"`
		Amount    float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload tidak valid"}`, http.StatusBadRequest)
		return
	}

	var loanID int64
	err := db.QueryRow("INSERT INTO courier_loans (courier_id, initial_loan_amount) VALUES ($1, $2) RETURNING id",
		req.CourierID, req.Amount).Scan(&loanID)

	if err != nil {
		http.Error(w, `{"error":"Gagal memunculkan modal pinjaman admin"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"loan_id": loanID,
		"message": fmt.Sprintf("Modal kas admin Rp %.0f berhasil dipinjamkan ke Driver #%d", req.Amount, req.CourierID),
	})
}

func handleSubmitReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)
	orderIDStr := r.FormValue("order_id")
	priceStr := r.FormValue("actual_price")
	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
	actualPrice, _ := strconv.ParseFloat(priceStr, 64)

	courierIDStr := r.Header.Get("X-User-ID")
	courierID, _ := strconv.ParseInt(courierIDStr, 10, 64)

	file, fileHeader, err := r.FormFile("receipt_image")
	imageURL := ""
	if err == nil {
		defer file.Close()
		imageURL = uploadToMinIO(fileHeader)
	}

	_, _ = db.Exec("INSERT INTO order_receipts (order_id, courier_id, receipt_image_url, actual_store_price) VALUES ($1, $2, $3, $4)",
		orderID, courierID, imageURL, actualPrice)

	tariffs := getDynamicTariffs()
	totalAmount := actualPrice + tariffs.BaseDeliveryFee
	_, _ = db.Exec("UPDATE orders SET status = 'ON_DELIVERY', subtotal = $1, total_amount = $2 WHERE id = $3", actualPrice, totalAmount, orderID)

	publishEvent("order.on_delivery", map[string]interface{}{
		"order_id":     orderID,
		"total_amount": totalAmount,
		"image_url":    imageURL,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Nota warung disimpan, pesanan dalam pengantaran!",
		"total_amount": totalAmount,
		"receipt_url":  imageURL,
	})
}

func handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID int64  `json:"order_id"`
		Reason  string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	userIDStr := r.Header.Get("X-User-ID")
	userRole := r.Header.Get("X-User-Role")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	_, err := db.Exec(`
		UPDATE orders SET status = 'CANCELLED', cancelled_at = NOW(), cancelled_by_role = $1, cancelled_by_user_id = $2, cancel_reason = $3
		WHERE id = $4`, userRole, userID, req.Reason, req.OrderID)

	if err != nil {
		http.Error(w, "Gagal membatalkan pesanan", http.StatusInternalServerError)
		return
	}

	publishEvent("order.cancelled", map[string]interface{}{
		"order_id": req.OrderID,
		"reason":   req.Reason,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Pesanan berhasil dibatalkan."})
}

func uploadToMinIO(fileHeader *multipart.FileHeader) string {
	if minioClient == nil {
		return ""
	}
	file, err := fileHeader.Open()
	if err != nil {
		return ""
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return ""
	}

	var resizedImg image.Image = img
	if img.Bounds().Dx() > 800 {
		resizedImg = resize.Resize(800, 0, img, resize.Lanczos3)
	}

	buf := new(bytes.Buffer)
	_ = jpeg.Encode(buf, resizedImg, &jpeg.Options{Quality: 75})

	objectName := fmt.Sprintf("media/%s.jpg", uuid.New().String())
	_, err = minioClient.PutObject(context.Background(), minioBucket, objectName, buf, int64(buf.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s/%s/%s", minioPublicURL, minioBucket, objectName)
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
