# DOKUMEN RENCANA SISTEM DALAM SEMUA TAHAP (SYSTEM DEVELOPMENT ROADMAP)

## 📌 Executive Summary
**Pesan Antar Desa** adalah platform micro-logistics dan e-commerce mikro berbasis komunitas yang menghubungkan warga desa, warung lokal, dan kurir pemuda desa. Sistem memanfaatkan **WhatsApp HTTP API (WAHA)** sebagai gateway komunikasi utama tanpa memaksa warga mengunduh aplikasi baru yang memberatkan kuota.

Pembangunan sistem dibagi menjadi **4 Tahapan Utama (Phased Rollout)** untuk menjamin kelancaran adopsi di lapangan (*zero friction*), efisiensi biaya, dan validasi model bisnis.

---

## 🗺️ Peta Jalan Pembangunan Sistem (System Development Roadmap)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ TAHAP 1: MVP LAUNCH (DRIVER-CENTRIC & JASA TITIP DESA) - [SEKARANG / SELESAI]  │
├─────────────────────────────────────────────────────────────────────────────────┤
│ • Model Jasa Titip (Toko tidak perlu mendaftar/menggunakan app).               │
│ • Admin BUMDes yang menginput data toko dan katalog produk.                    │
│ • PWA Web Katalog untuk Pembeli + WAHA Integration (Auto Share Location WA).   │
│ • React Native Mobile App untuk Driver (Foto nota kertas warung ke MinIO).       │
│ • Skema Ongkir Fleksibel: Rp 10.000 (Dasar) + Rp 2.000 per toko tambahan.       │
│ • Pembayaran COD Tunai + Pinjaman Modal Talangan Kas Admin Pagi.               │
│ • Message Broker RabbitMQ + PostgreSQL JSONB Audit Trail Log.                  │
│ • Link Viral Status WA & Social Media Preview Cards (Open Graph).              │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ TAHAP 2: INTEGRASI TOKO MITRA & PAYMENT GATEWAY (PARTNER STORES & QRIS)        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ • Aktifkan Flag Toko Mitra (`is_partner = true`).                               │
│ • Web Dashboard Pemilik Toko (Merchant Portal) untuk kelola menu & stok mandiri.│
│ • Integrasi Payment Gateway Nontunai (Dynamic QRIS BCA, Mandiri, GoPay, DANA).  │
│ • Modul Pencairan Omzet Toko Harian (Store Payouts / Transfer BUMDes).          │
│ • In-App Real-time Chat (WebSockets Go + Redis Pub/Sub) antara Pembeli & Toko. │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ TAHAP 3: REAL-TIME ANALYTICS & SKALASI MULTI-DESA (MULTI-VILLAGE HUB)          │
├─────────────────────────────────────────────────────────────────────────────────┤
│ • Ekspansi Layanan ke 5-10 Desa Sekitar dalam 1 Kecamatan (Cluster Hub).        │
│ • Advanced BI Analytics Engine (ClickHouse OLAP / Metabase).                    │
│ • Algoritma Auto-Dispatch Kurir Berbasis Jarak GPS & Beban Kerja (Load Balancing).│
│ • Sistem Rating & Program Loyalitas Pelanggan / Kurir (Courier Performance Bonus).│
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ TAHAP 4: EKOSISTEM DIGITAL BUMDES & SMART LOGISTICS                             │
├─────────────────────────────────────────────────────────────────────────────────┤
│ • Integrasi dengan Unit Usaha BUMDes (Air Galon, Gas LPG, Sembako Pasar Desa).  │
│ • Fitur Langganan Rutin (Scheduled Weekly Subscription Delivery).              │
│ • Integrasi Payment Gateway untuk Pembelian Pulsa, PLN, dan Tagihan Desa.       │
│ • AI Demand Forecasting untuk Prediksi Stok Bahan Pokok Desa.                   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔍 Detail Spesifikasi Per Tahap

### TAHAP 1: MVP Launch (Model Jasa Titip & Driver-Centric)
* **Tujuan Utama:** Validasi adopsi pasar tanpa hambatan edukasi ke warung desa.
* **Fitur Utama:**
  * Catalog Management oleh Admin.
  * Formulasi Ongkir: `10.000 + (N-1) * 2.000`.
  * Pembayaran Tunai (COD) dengan Pinjaman Modal Talangan Pagi (`courier_loans`).
  * Capture Share Location via Webhook WAHA.
  * Upload Foto Nota Fisik Warung oleh Driver via Kamera HP ke MinIO Storage.
  * Asynchronous Audit Logging & Event Messaging via RabbitMQ.
* **Target Waktu:** 4 Minggu.

### TAHAP 2: Integrasi Toko Mitra & Payment Gateway
* **Tujuan Utama:** Penghematan waktu belanja driver dan digitalisasi transaksi desa.
* **Fitur Utama:**
  * Merchant Dashboard Web App (Store Owner login).
  * Dynamic QRIS Payment via Midtrans/Xendit.
  * Modul Rekonsiliasi & Transfer Saldo Toko (`store_payouts`).
  * In-App WebSockets Real-time Chat.
* **Target Waktu:** 6 Minggu setelah Tahap 1 Stabil.

### TAHAP 3: Skalasi Multi-Desa & Analytics
* **Tujuan Utama:** Memperluas jangkauan pasar dan efisiensi pengantaran antar-desa.
* **Fitur Utama:**
  * Multi-Tenancy / Cluster Desa support.
  * Integration dengan ClickHouse OLAP untuk analisis heatmap rute instan.
  * Dispatch Algorithm berbasis titik lokasi GPS Driver.
* **Target Waktu:** 3 Bulan setelah Tahap 2 Stabil.

### TAHAP 4: Ekosistem Digital BUMDes
* **Tujuan Utama:** Menjadikan aplikasi sebagai pusat ekonomi digital dan logistik desa.
* **Fitur Utama:**
  * Pengantaran terjadwal (Gas LPG, Galon Air, Sembako).
  * Prediksi kebutuhan stok desa menggunakan Machine Learning sederhana.
* **Target Waktu:** 6 Bulan setelah Tahap 3 Stabil.
