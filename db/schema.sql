-- =============================================================================
-- DATABASE DDL SCHEMA FOR PESAN ANTAR DESA VIA WHATSAPP (POSTGRESQL 16)
-- =============================================================================

CREATE TYPE user_role AS ENUM ('admin', 'store_owner', 'courier', 'customer');
CREATE TYPE order_status AS ENUM (
    'CREATED',
    'WAITING_LOCATION',
    'SEARCHING_COURIER',
    'COURIER_ASSIGNED',
    'ON_DELIVERY',
    'COMPLETED',
    'CANCELLED'
);
CREATE TYPE payment_method_type AS ENUM ('COD', 'QRIS', 'EWALLET', 'VA');
CREATE TYPE payment_status_type AS ENUM ('UNPAID', 'PENDING', 'PAID', 'EXPIRED', 'REFUNDED');

-- 1. TABEL USERS
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL UNIQUE, -- Format: 628123456789
    role user_role NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. TABEL PROFIL COURIER/DRIVER
CREATE TABLE courier_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    vehicle_plate VARCHAR(20) NOT NULL,
    vehicle_type VARCHAR(50) DEFAULT 'Motor',
    is_active BOOLEAN DEFAULT TRUE,
    is_online BOOLEAN DEFAULT FALSE,
    current_status VARCHAR(20) DEFAULT 'IDLE', -- 'IDLE', 'ON_DELIVERY', 'OFFLINE'
    total_deliveries INT DEFAULT 0,
    rating NUMERIC(3,2) DEFAULT 5.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. TABEL STORES (TOKO/WARUNG DESA)
CREATE TABLE stores (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    store_name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    address_text TEXT NULL,
    image_url TEXT NULL,
    is_partner BOOLEAN DEFAULT FALSE, -- FALSE = Jasa Titip (Phase 1), TRUE = Toko Mitra (Phase 2)
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. TABEL PRODUCTS (KATALOG MENU/BARANG)
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    category VARCHAR(50) DEFAULT 'Umum',
    is_available BOOLEAN DEFAULT TRUE,
    image_url TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. TABEL ORDERS (TRANSAKSI PESANAN)
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_code VARCHAR(20) NOT NULL UNIQUE, -- Contoh: ORD-8821
    customer_id BIGINT NOT NULL REFERENCES users(id),
    store_id BIGINT NOT NULL REFERENCES stores(id),
    courier_id BIGINT NULL REFERENCES users(id),
    status order_status NOT NULL DEFAULT 'CREATED',
    delivery_address_text TEXT NULL,
    latitude NUMERIC(10, 8) NULL,   -- Ditangkap dari WAHA Share Location
    longitude NUMERIC(11, 8) NULL,  -- Ditangkap dari WAHA Share Location
    total_stores INT NOT NULL DEFAULT 1,
    subtotal NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    base_delivery_fee NUMERIC(10,2) NOT NULL DEFAULT 10000.00, -- Rp 10.000
    store_add_fee NUMERIC(10,2) NOT NULL DEFAULT 2000.00,       -- Rp 2.000 per toko tambahan
    delivery_fee NUMERIC(10,2) NOT NULL DEFAULT 10000.00,
    total_amount NUMERIC(10,2) NOT NULL,
    payment_method payment_method_type NOT NULL DEFAULT 'COD',
    payment_status payment_status_type NOT NULL DEFAULT 'UNPAID',
    payment_gateway_tx_id VARCHAR(100) NULL,
    qr_code_url TEXT NULL,
    paid_at TIMESTAMP WITH TIME ZONE NULL,
    cancelled_at TIMESTAMP WITH TIME ZONE NULL,
    cancelled_by_role VARCHAR(20) NULL,
    cancelled_by_user_id BIGINT NULL REFERENCES users(id),
    cancel_reason TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. TABEL ORDER ITEMS (RINCIAN MIKRO BARANG)
CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT NULL REFERENCES products(id),
    item_name VARCHAR(150) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    price_per_item NUMERIC(10,2) NOT NULL,
    notes TEXT NULL,
    is_custom_item BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'ACTIVE' -- 'ACTIVE', 'CANCELLED'
);

-- 7. TABEL SYSTEM SETTINGS (PENGATURAN TARIF FLEKSIBEL DYNAMIC)
CREATE TABLE system_settings (
    key VARCHAR(50) PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    description TEXT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- SEED PENGATURAN TARIF DANA AWAL
INSERT INTO system_settings (key, value, description) VALUES
('base_delivery_fee', '10000', 'Tarif dasar pengantaran desa (Rp)'),
('per_extra_store_fee', '2000', 'Biaya per toko tambahan (Rp)'),
('min_order_amount', '0', 'Minimal transaksi pesanan (Rp)');

-- 8. TABEL ORDER REVISIONS (AUDIT REVISI ITEM/HARGA DRIVER)
CREATE TABLE order_revisions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    revised_by_user_id BIGINT NOT NULL REFERENCES users(id),
    revision_type VARCHAR(50) NOT NULL,
    reason TEXT NULL,
    old_subtotal NUMERIC(10,2) NOT NULL,
    new_subtotal NUMERIC(10,2) NOT NULL,
    old_total_amount NUMERIC(10,2) NOT NULL,
    new_total_amount NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 9. TABEL ORDER RECEIPTS (FOTO NOTA WARUNG OLEH DRIVER)
CREATE TABLE order_receipts (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    courier_id BIGINT NOT NULL REFERENCES users(id),
    receipt_image_url TEXT NOT NULL,
    actual_store_price NUMERIC(10,2) NOT NULL,
    notes TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 10. TABEL COURIER LOANS (PINJAMAN MODAL TALANGAN ADMIN PAGI)
CREATE TABLE courier_loans (
    id BIGSERIAL PRIMARY KEY,
    courier_id BIGINT NOT NULL REFERENCES users(id),
    loan_date DATE NOT NULL DEFAULT CURRENT_DATE,
    initial_loan_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    returned_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 11. TABEL CHAT MESSAGES (CHAT DALAM APLIKASI)
CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES users(id),
    sender_role VARCHAR(20) NOT NULL,
    message_text TEXT NULL,
    attachment_url TEXT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 12. TABEL AUDIT LOGS (PERMANENT AUDIT TRAIL)
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    performed_by_id BIGINT NULL,
    performed_by_role VARCHAR(20) NULL,
    ip_address VARCHAR(45) NULL,
    old_values JSONB NULL,
    new_values JSONB NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- INDEXING PERFORMA QUERY
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_courier ON orders(courier_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_products_store ON products(store_id);
CREATE INDEX idx_chat_order ON chat_messages(order_id, created_at ASC);
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at DESC);
