# Mías Hamburguesas de Lujo
> *«Lo Íntimo Hecho para Compartir»* — Plataforma gastronómica de alta gama, comandas y fidelización.

---

## 1. Visión y Fundamentos de Ingeniería

**Mías Hamburguesas de Lujo** es una experiencia gastronómica boutique centrada en hamburguesas artesanales al carbón preparadas en horno de autor, con carne de res premium de 160g sobre pan pretzel artesanal y perros calientes en pan brioche.

La plataforma combina un backend tipado en Go y una interfaz reactiva orientada a dispositivos móviles con retícula editorial suiza.

```
                                  [ CLIENTE / PERSONAL ]
                                             │
                   ┌─────────────────────────┴─────────────────────────┐
                   ▼                                                   ▼
         [ Svelte 5 + Runes ]                                 [ PWA Service Worker ]
         - Retícula Editorial & Fotos Reales                  - Anillo Cero (Offline)
         - Maridaje Cuadrado & Pedidos Móviles                - Despacho WhatsApp
                   │
                   ▼
           [ Chi Router v5 ] ─── (CORS, RequestID, Structured JSON slog, Recoverer)
                   │
         ┌─────────┴───────────────────────────────────────────┐
         ▼                                                     ▼
  [ Argon2id & JWT ]                                      [ SSE Hub ]
  (RFC 9106, Dual Token Memory/Cookie)                    (Pulsos Heartbeat 15s)
         │                                                     │
         └───────────────────────────┬─────────────────────────┘
                                     ▼
                      [ SQLC v1.31 + pgx/v5 Pool ]
                      (Consultas SQL puras 100% tipadas en Go)
                                     │
                                     ▼
                       [ PostgreSQL 16 Alpine ]
                       (Libro Mayor Inmutable & Antifraude)
```

---

## 2. Pila Tecnológica

### Backend
- **Lenguaje:** Go 1.27
- **Router HTTP:** `github.com/go-chi/chi/v5`
- **Capa de Datos:** `github.com/jackc/pgx/v5` con pool de conexiones (`pgxpool`)
- **Compilador SQL Tipado:** SQLC v1.31.1
- **Seguridad & Hash:** Argon2id (`golang.org/x/crypto/argon2`) y JWT Dual (`github.com/golang-jwt/jwt/v5`)
- **Eventos en Vivo:** Server-Sent Events (SSE) nativo con canales multiplexados

### Frontend
- **Framework:** Svelte 5 con Runes nativos (`$state`, `$derived`, `$props`, `$effect`)
- **Bundler:** Vite
- **Estilos:** Tailwind CSS con tokens gastronómicos de autor (`cream`, `bone`, `char`, `ember`, `gold`)
- **PWA & Offline:** Service Worker nativo (Anillo 0) con catálogo embebido en cliente y soporte de caché para activos compilados
- **Audio:** Chime acústico armónico sintetizado con Web Audio API reutilizando un contexto singleton

### Infraestructura
- **Base de Datos:** PostgreSQL 16 Alpine en contenedor Docker (`mias_postgres_16`)
- **Migrador Nativo:** CLI en Go (`backend/cmd/migrate`) con soporte de transacciones atómicas y tabla `schema_migrations`

---

## 3. Reglas de Negocio Estrictas

1. **Pan Pretzel Exclusivo:** Todas las hamburguesas utilizan pan pretzel de 160g. El pan brioche artesanal está reservado exclusivamente para los perros calientes.
2. **Horarios de Atención:** Sábado a Jueves de 3:00 PM a 9:30 PM (Viernes cerrado).
3. **Fidelización por Factura Física:** El cliente digita el número consecutivo de su factura física (ej. `#00124`). PostgreSQL previene duplicados mediante restricción `UNIQUE(receipt_number)` y libro mayor de doble partida.
4. **Redención Presencial:** 10 sellos = 1 Hamburguesa Original gratis ($17.000 COP). Estado de aprobación en vivo validado presencialmente por caja.
5. **Modalidades de Comanda:**
   - **Consumo en Salón:** Código `ORD-XXXX` vinculado a mesa o retiro en barra.
   - **Para Llevar:** Código `TKA-XXXX` sin recargo de envío, listo para recoger en punto.
   - **A Domicilio:** Código `DOM-XXXX`, wizard paso a paso con cálculo de vueltas, tarifa base oficial de $6.000 COP en Duitama y despacho formateado a WhatsApp (`+57 321 252 0881`).

---

## 4. Puesta en Marcha Local

### Prerrequisitos
- Docker & Docker Compose
- Go 1.27+
- Node.js 20+ y pnpm

### 1. Iniciar Base de Datos (PostgreSQL 16)
```bash
docker compose up -d
```

### 2. Ejecutar Migraciones y Seeds
```bash
cd backend
go run ./cmd/migrate
```

### 3. Iniciar Backend Go API
```bash
cd backend
go run ./cmd/api
# Escuchando en http://localhost:8080
```

### 4. Iniciar Frontend Svelte 5
```bash
cd frontend
pnpm install
pnpm run dev
# Disponible en http://localhost:5173
```

---

## 5. Credenciales de Personal y Auditoría

La base de datos incluye cuentas de personal preconfiguradas con hash Argon2id para pruebas operativas:

| Rol | Correo Electrónico | Contraseña | Capacidades |
| :--- | :--- | :--- | :--- |
| **Caja Central** | `cajero@mias.com` | `CajeroMias2026!` | Cobro de órdenes, aprobación de facturas de sellos, redención de gratis e impresión térmica 80mm. |
| **Meseros** | `mesero@mias.com` | `MeseroMias2026!` | Consulta de comandas de salón en tiempo real y verificación de sellos. |
| **Administrador** | `admin@mias.com` | `AdminMias2026!` | Tablero de inteligencia de negocio (BI), control de stock de productos y auditoría. |

---

## 6. Documentación Normativa (`docs/`)

- [`docs/00_PROJECT_CONTEXT.md`](docs/00_PROJECT_CONTEXT.md): Contexto e historia de la marca.
- [`docs/01_PRD.md`](docs/01_PRD.md): Requisitos de producto y casos de uso.
- [`docs/02_TRD.md`](docs/02_TRD.md): Especificación técnica, esquemas y endpoints.
- [`docs/03_APP_FLOW.md`](docs/03_APP_FLOW.md): Flujos de navegación y estados de comanda.
- [`docs/05_BRAND_SYSTEM.md`](docs/05_BRAND_SYSTEM.md): Sistema de diseño, tokens y tipografía.
- [`docs/13_DECISION_LOG.md`](docs/13_DECISION_LOG.md): Registro inmutable de decisiones arquitectónicas (ADRs).
- [`docs/17_PROJECT_GLOSSARY.md`](docs/17_PROJECT_GLOSSARY.md): Glosario de términos del negocio.
