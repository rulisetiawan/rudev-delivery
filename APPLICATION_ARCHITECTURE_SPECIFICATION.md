# DOKUMEN SPESIFIKASI ARSITEKTUR APLIKASI (APPLICATION ARCHITECTURE SPECIFICATION)

---

## 🏗️ 1. High-Level Architecture Diagram

Sistem dibangun menggunakan **Arsitektur Microservices Event-Driven** dengan **API Gateway** sebagai pintu masuk tunggal, **RabbitMQ** sebagai pengatur antrean event asynchronous, **WAHA (WhatsApp HTTP API)** sebagai gateway komunikasi, dan **PostgreSQL + MinIO** untuk penyimpanan data dan objek media.

```
                                  ┌────────────────────────────────────────────────────────┐
                                  │                CLIENT APPS / USERS                     │
                                  │  - Customer PWA (Vue.js / Web)                         │
                                  │  - Driver App (React Native Mobile)                    │
                                  │  - Social Media Links (WA Status, FB, IG Open Graph)   │
                                  └───────────────────────────┬────────────────────────────┘
                                                              │
                                                              ▼
                                  ┌────────────────────────────────────────────────────────┐
                                  │           CUSTOM API GATEWAY (Golang :8080)            │
                                  │  - Centralized CORS & Security Hardening               │
                                  │  - JWT Verification & Identity Header Offloading       │
                                  │  - Redis Token Bucket Rate Limiting (60 req/min)       │
                                  │  - Request Correlation ID Injector (X-Request-ID)      │
                                  └───────────────────────────┬────────────────────────────┘
                                                              │
         ┌────────────────────────────────────────────────────┼────────────────────────────────────────────────────┐
         │ (HTTP Proxy)                                       │ (HTTP Proxy)                                       │ (HTTP Proxy)
         ▼                                                    ▼                                                    ▼
┌──────────────────────────┐                       ┌──────────────────────────┐                       ┌──────────────────────────┐
│   AUTH & USER SERVICE    │                       │  ORDER & CATALOG SERVICE │                       │   NOTIFICATION SERVICE   │
│   (Golang Port 8081)     │                       │   (Golang Port 8082)     │                       │   (Golang Port 8083)     │
└────────────┬─────────────┘                       └────────────┬─────────────┘                       └────────────┬─────────────┘
             │                                                  │                                                  │
             ▼                                                  ▼                                                  ▼
┌──────────────────────────┐                       ┌──────────────────────────┐                       ┌──────────────────────────┐
│ PostgreSQL 1 (Auth DB)   │                       │ PostgreSQL 2 (Order DB)  │                       │  WAHA Docker Container   │
└──────────────────────────┘                       └────────────┬─────────────┘                       └──────────────────────────┘
                                                                │
                                                                ▼ (Publish AMQP Events)
                                                   ┌──────────────────────────┐
                                                   │ MESSAGE BROKER: RABBITMQ │
                                                   │ (Exchange: desa.events)  │
                                                   └────────────┬─────────────┘
                                                                │
         ┌──────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┐
         │ (Topic: audit.*)                                     │ (Topic: order.*)                                     │ (OLAP Sync)
         ▼                                                      ▼                                                      ▼
┌──────────────────────────┐                       ┌──────────────────────────┐                       ┌──────────────────────────┐
│    AUDIT LOG SERVICE     │                       │  MINIO OBJECT STORAGE    │                       │   METABASE ANALYTICS     │
│   (Golang Port 8084)     │                       │  (Bucket: desa-media)    │                       │   (BI Dashboard :3005)   │
└──────────────────────────┘                       └──────────────────────────┘                       └──────────────────────────┘
```

---

## 🛠️ 2. Detail Komponen & Microservices Stack

### A. Custom API Gateway (`services/api-gateway`)
* **Teknologi:** Golang + Redis.
* **Tugas Utama:**
  1. **Single Entrypoint:** Memetakan seluruh request dari PWA dan Mobile App ke service internal.
  2. **Security Offloading:** Memverifikasi Token JWT untuk protected routes (`/api/v1/orders/`). Mengekstrak `user_id` dan `role`, lalu menyuntikkannya ke HTTP Header internal (`X-User-ID`, `X-User-Role`).
  3. **Redis Rate Limiter:** Membatasi request maksimal 60 req/menit per User/IP menggunakan Redis Fixed Window untuk mencegah DDoS/Brute Force.
  4. **Tracing:** Menyuntikkan Correlation ID (`X-Request-ID`) ke setiap request yang lewat.

### B. Auth & User Service (`services/auth-service`)
* **Teknologi:** Golang + PostgreSQL.
* **Tugas Utama:** Mengelola autentikasi, registrasi pembeli, pendaftaran driver resmi, serta generasi Token JWT.

### C. Order & Catalog Service (`services/order-service`)
* **Teknologi:** Golang + PostgreSQL + MinIO Client + RabbitMQ Producer.
* **Tugas Utama:**
  1. **Catalog CRUD:** Mengelola data warung dan produk makanan/sembako oleh Admin.
  2. **Ongkir Calculator:** Formulasi `10.000 + (Jumlah Toko - 1) * 2.000`.
  3. **Order Lifecycle & Revisions:** Menangani pembatalan, perubahan item, dan penyesuaian harga oleh driver.
  4. **Receipt Upload:** Mengompresi gambar nota kertas warung dari kamera HP driver (Max width 800px, JPEG 75%) dan mengunggahnya ke MinIO.
  5. **Social Share OG Tags:** Menyajikan HTML Meta Tags Open Graph untuk Preview Kartu Status WA.

### D. Notification Service & WAHA Adapter (`services/notification-service`)
* **Teknologi:** Golang + WAHA HTTP API + RabbitMQ Consumer.
* **Tugas Utama:**
  1. **Webhook Parser:** Menangkap pesan Share Location (Lat/Lng) dari WhatsApp Pembeli.
  2. **Async WA Outbound:** Mengonsumsi event dari RabbitMQ (`order.created`, `order.on_delivery`, `order.cancelled`) dan menembak REST API WAHA untuk mengirimkan pesan WhatsApp otomatis.

### E. Audit Service (`services/audit-service`)
* **Teknologi:** Golang + PostgreSQL + RabbitMQ Consumer.
* **Tugas Utama:** Mengonsumsi event `audit.*` dari RabbitMQ secara asynchronous dan mencatat *snapshot* perubahan data (`old_values` & `new_values`) ke dalam tabel PostgreSQL `audit_logs` bertipe JSONB.

### F. Object Storage (MinIO Storage)
* **Teknologi:** MinIO Container (S3 Compatible).
* **Tugas Utama:** Menyimpan aset foto produk katalog dan foto bukti nota kertas warung. Dilengkapi init script `create-buckets` untuk otomatisasi bucket `desa-media`.

### G. Message Broker (RabbitMQ)
* **Teknologi:** RabbitMQ 3 Management (AMQP 0-9-1).
* **Exchange:** `desa.events` (Topic Exchange).
* **Queues Utama:** `q.notification.wa` dan `q.audit.logs`.

---

## 🗄️ 3. Skema Relasi Database (ERD Data Model Summary)

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│    USERS     │1    * │    STORES    │1    * │   PRODUCTS   │
│ (ID, Phone,  ├───────┤ (ID, Slug,   ├───────┤ (ID, Price,  │
│  Role)       │       │  is_partner) │       │  ImageURL)   │
└──────┬───────┘       └──────┬───────┘       └──────┬───────┘
       │1                     │1                     │1
       │                      │                      │
       │*                     │*                     │*
┌──────┴──────────────────────┴──────────────────────┴──────┐
│                         ORDERS                            │
│ (ID, OrderCode, CustomerID, CourierID, Status, Total)     │
└──────┬──────────────────────┬──────────────────────┬──────┘
       │1                     │1                     │1
       │                      │                      │
       │*                     │*                     │*
┌──────┴───────┐       ┌──────┴───────┐       ┌──────┴───────┐
│ ORDER_ITEMS  │       │ORDER_RECEIPTS│       │ORDER_REVISIONS│
│ (ItemName,   │       │(ReceiptImage,│       │(OldTotal,    │
│  Price, Qty) │       │ ActualPrice) │       │ NewTotal)    │
└──────────────┘       └──────────────┘       └──────────────┘
```

---

## 🔒 4. Keamanan & Penanganan Kegagalan (Security & Resilience Strategy)

1. **Anti-Ban WAHA Throughput Delay:** `notification-service` menggunakan antrean RabbitMQ dengan jeda acak 1.5 - 3 detik antar pengiriman pesan WAHA untuk menjaga nomor WhatsApp tetap aman dari blokir Meta.
2. **Idempotency Event:** Webhook parser mencatat `message_id` WAHA ke Redis agar event ganda akibat gangguan jaringan tidak menyebabkan pesanan diproses dua kali.
3. **Signed URLs pada Link Kurir:** Link aksi kurir seperti klaim job dilengkapi HMAC Token untuk mencegah manipulasi ID oleh pihak ketiga.
4. **Graceful Shutdown:** Seluruh Golang microservices mengimplementasikan *graceful shutdown* `sys.SIGTERM` untuk memastikan koneksi database dan pesan RabbitMQ tersimpan rapi saat restart.

---

## 🐳 5. Skrip Docker Compose Stack Integrasi

Seluruh stack terdefinisi di file master [`docker-compose.yml`](file:///Users/rulli/belajar/RuDev/docker-compose.yml):

```yaml
# Daftar Service Terintegrasi:
# 1. postgres (Port 5432)
# 2. redis (Port 6379)
# 3. minio (Port 9000 S3, 9001 Dashboard)
# 4. rabbitmq (Port 5672 AMQP, 15672 Dashboard)
# 5. waha (Port 3000 WA Engine)
# 6. metabase (Port 3005 BI Analytics)
# 7. api-gateway (Port 8080)
# 8. auth-service (Port 8081)
# 9. order-service (Port 8082)
# 10. notification-service (Port 8083)
# 11. audit-service (Port 8084)
# 12. frontend (Port 80 Nginx PWA)
```

Perintah Menjalankan:
```bash
cd /Users/rulli/belajar/RuDev
docker compose up -d --build
```
