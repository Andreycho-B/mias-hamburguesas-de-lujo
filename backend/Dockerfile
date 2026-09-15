# ─────────────────────────────────────────────────────────────
# Mías Hamburguesas de Lujo — Backend Go Multi-Stage Build
# Imagen final: ~25MB, usuario no-root, binario estático
# ─────────────────────────────────────────────────────────────

# ── Etapa 1: Compilación ────────────────────────────────────
FROM golang:1.27-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

# Cacheo de dependencias (solo se invalida si cambian go.mod/go.sum)
COPY go.mod go.sum ./
RUN go mod download

# Compilar binario estático con CGO deshabilitado
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /mias-api ./cmd/api

# ── Etapa 2: Imagen de Producción ───────────────────────────
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S mias && adduser -S mias -G mias

WORKDIR /app

COPY --from=builder /mias-api .
COPY --from=builder /build/migrations ./migrations

# Ejecutar como usuario no-root
USER mias

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["./mias-api"]
