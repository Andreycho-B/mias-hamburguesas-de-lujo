# 02. Documento de Requisitos Técnicos (TRD)

**Proyecto:** Mías.8 Hamburguesas de Lujo  
**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  

---

## 1. Arquitectura del Sistema y Pila Tecnológica

El sistema se estructura bajo una **Arquitectura Hexagonal (Puertos y Adaptadores)** de alto rendimiento, dividida en dos componentes principales:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        ARQUITECTURA DEL SISTEMA                        │
├───────────────────────────────────┬────────────────────────────────────┤
│ FRONTEND (Cloudflare Pages)       │ BACKEND (Oracle Cloud / Coolify)   │
│ • Svelte 5 (Runes $state, $derived)│ • Go 1.27 + Chi Router            │
│ • Vite + Tailwind CSS (Tokens)    │ • SQLC + pgx/v5 (Pool Conexiones)  │
│ • GSAP (Flip) + Lenis Scroll      │ • PostgreSQL 16 Relacional         │
│ • Bits UI (Modales/Drawers)       │ • Argon2id (Hashing de Personal)   │
│ • Native SSE (Eventos en Vivo)    │ • JWT + Refresh Cookie HttpOnly    │
│ • PWA + Web Audio API + Web Push  │ • VAPID Web Push Engine            │
└───────────────────────────────────┴────────────────────────────────────┘
```

---

## 2. Esquema Relacional de Base de Datos (PostgreSQL 16)

Gestionado mediante migraciones versionadas con **`goose`** e interactuado estrictamente a través de código Go compilado por **`sqlc`**.

```sql
-- 1. Extensiones
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Usuarios y Personal
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    avatar_url TEXT,
    phone VARCHAR(20),
    role VARCHAR(20) NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'waiter', 'cashier', 'admin')),
    password_hash TEXT, -- Argon2id para staff, NULL para OAuth
    last_order_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Categorías y Catálogo CMS
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    description TEXT NOT NULL,
    price INT NOT NULL CHECK (price >= 0),
    bread_type VARCHAR(50), -- 'pretzel' o 'brioche'
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

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price INT NOT NULL CHECK (unit_price >= 0),
    customizations JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- 5. Calificaciones de Domicilio
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
    amount INT NOT NULL, -- +N acreditados, -10 redimidos
    reason VARCHAR(100) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, -- 6 meses desde emisión
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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
    CONSTRAINT uq_receipt_number UNIQUE (receipt_number)
);

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

-- 8. Auditoría Forense Administrativa
CREATE TABLE admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address VARCHAR(45) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índices Estratégicos
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
CREATE INDEX idx_stamp_requests_status_created ON stamp_requests(status, created_at);
CREATE INDEX idx_stamp_ledger_user ON stamp_ledger(user_id);
CREATE INDEX idx_stamp_ledger_expires ON stamp_ledger(expires_at);
```

---

## 3. Seguridad y Protocolos Criptográficos

1. **Hashing de Contraseñas (Argon2id)**:
   * Memoria: `64 MB (65536 KB)`
   * Iteraciones: `3`
   * Paralelismo: `2 hilos`
   * Salt: `16 bytes` criptográficamente seguros (`crypto/rand`).
   * Longitud de clave: `32 bytes`.
2. **Ciclo de Vida de Tokens (Zero-Trust)**:
   * **Access Token**: JWT firmado con HMAC-SHA256, expiración a los **15 minutos**, almacenado únicamente en la memoria volátil del cliente Svelte.
   * **Refresh Token**: Cadena opaca criptográfica aleatoria de 32 bytes con hash SHA-256 en base de datos. Se transporta exclusivamente en cookie:
     `HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=604800 (7 días)`.
   * **Rotación y Detección de Reuso**: Al refrescar la sesión, el token previo queda revocado. Si un token revocado es presentado nuevamente, se invalidan inmediatamente todas las sesiones asociadas a ese usuario.
3. **Claves de Idempotencia**:
   * Header `Idempotency-Key: <UUID>` en peticiones `POST /orders` y `POST /stamps/request`.
   * Bloqueo en memoria / Redis con TTL de 60 segundos para evitar doble inserción en caso de fluctuaciones de red móvil.

---

## 4. Server-Sent Events (SSE) y Concurrencia en Go

* **Canal Staff Centralizado (`/api/v1/events/staff`)**:
  * Autenticación requerida con rol `waiter`, `cashier` o `admin`.
  * Formato de evento estándar:
    ```text
    event: order_pending_payment
    id: 18492
    data: {"order_id":"...","code":"ORD-A101","total":38000,"type":"in_store"}
    ```
  * **Heartbeat**: Ping en blanco cada 15 segundos (`: ping\n\n`) para impedir desconexiones automáticas por proxies o redes móviles.
  * **Reconexión**: Soporte del header `Last-Event-ID` con buffer circular de eventos para reenviar mensajes perdidos si el dispositivo del personal se desconecta brevemente.
* **Worker Autónomo de Limpieza**:
  * Goroutine en Go con `time.Ticker` que corre cada 2 minutos:
    * Expira solicitudes de sellos pendientes con más de 60 minutos de antigüedad (`status = 'EXPIRED'`).
    * Evalúa sellos con fecha de caducidad vencida (`expires_at < NOW()`).

---

## 5. Implementaciones Frontend de Alta Gama

* **Web Audio API**: Síntesis sinusoidal de alta fidelidad con ataque suave y caída armónica que reproduce un chime elegante en los paneles de sala y caja sin descargar archivos de audio externos.
* **Impresión Térmica**: Regla `@media print` que oculta la interfaz de usuario y renderiza una comanda térmica en ancho de 80mm/58mm con fuentes monoespaciadas y cortes de línea limpios.
* **Manejo de Intro (3.5s)**: Función síncrona en el montaje inicial de la app que lee `localStorage.getItem('mias_intro_seen_at')`. Si la diferencia es menor a 3600 segundos (1 hora), pasa a `display: none` antes de que ocurra el primer pintado (cero parpadeos FOUC).
