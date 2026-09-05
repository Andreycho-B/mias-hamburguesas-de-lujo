-- +goose Up
-- SQL Schema para Mías.8 Hamburguesas de Lujo (PostgreSQL 16)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Usuarios y Personal (Zero-Trust)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    avatar_url TEXT,
    phone VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'waiter', 'cashier', 'admin')),
    password_hash TEXT, -- Argon2id para staff (admin, mesero, cajero), NULL para clientes Google OAuth
    last_order_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Categorías del Menú
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- 3. Productos y Catálogo Gastronómico
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NOT NULL,
    price INT NOT NULL CHECK (price >= 0),
    bread_type VARCHAR(50), -- 'pretzel' para hamburguesas, 'brioche' para perros calientes
    image_url TEXT,
    is_bestseller BOOLEAN NOT NULL DEFAULT FALSE,
    is_chef_choice BOOLEAN NOT NULL DEFAULT FALSE,
    is_season BOOLEAN NOT NULL DEFAULT FALSE,
    is_in_stock BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Pedidos y Comandas
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL, -- 'ORD-A101', 'DOM-B201'
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    order_type VARCHAR(20) NOT NULL CHECK (order_type IN ('in_store', 'delivery')),
    status VARCHAR(30) NOT NULL CHECK (status IN ('PENDING_PAYMENT', 'PAID', 'IN_PREPARATION', 'ON_THE_WAY', 'DELIVERED', 'CANCELLED')),
    subtotal INT NOT NULL CHECK (subtotal >= 0),
    delivery_fee INT NOT NULL DEFAULT 0 CHECK (delivery_fee >= 0),
    discount_amount INT NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    total_amount INT NOT NULL CHECK (total_amount >= 0),
    delivery_name VARCHAR(150),
    delivery_phone VARCHAR(20),
    delivery_address TEXT,
    delivery_notes TEXT,
    payment_method VARCHAR(50),
    cash_change_amount INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Detalle de Items por Pedido
CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price INT NOT NULL CHECK (unit_price >= 0),
    customizations JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- 5. Calificaciones de Domicilio (1 a 5 estrellas)
CREATE TABLE delivery_reviews (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Fidelización: Libro Mayor Inmutable (Double-Entry Ledger)
CREATE TABLE stamp_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    amount INT NOT NULL, -- +N por hamburguesas compradas, -10 por redención
    reason VARCHAR(100) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, -- 6 meses desde emisión
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Solicitudes de Sellos en Caja y Mesero
CREATE TABLE stamp_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receipt_number VARCHAR(50) NOT NULL,
    burgers_count INT NOT NULL CHECK (burgers_count > 0),
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'EXPIRED')),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_receipt_number UNIQUE (receipt_number) -- Blindaje contra reutilización
);

-- Solicitudes de Redención de Hamburguesa Gratis
CREATE TABLE redemption_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 7. Suscripciones Web Push VAPID
CREATE TABLE push_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 8. Auditoría Forense Administrativa (Inmutable)
CREATE TABLE admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address VARCHAR(45) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices Estratégicos de Alto Rendimiento
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_orders_code ON orders(code);
CREATE INDEX idx_stamp_requests_status_created ON stamp_requests(status, created_at);
CREATE INDEX idx_stamp_ledger_user ON stamp_ledger(user_id);
CREATE INDEX idx_stamp_ledger_expires ON stamp_ledger(expires_at);
CREATE INDEX idx_products_category ON products(category_id, is_active);

-- +goose Down
DROP TABLE IF EXISTS admin_audit_logs;
DROP TABLE IF EXISTS push_subscriptions;
DROP TABLE IF EXISTS redemption_requests;
DROP TABLE IF EXISTS stamp_requests;
DROP TABLE IF EXISTS stamp_ledger;
DROP TABLE IF EXISTS delivery_reviews;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;
