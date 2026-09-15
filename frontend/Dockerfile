# ─────────────────────────────────────────────────────────────
# Mías Hamburguesas de Lujo — Frontend Svelte 5 Multi-Stage
# Etapa 1: Node 20 + pnpm build → Etapa 2: Nginx Alpine
# ─────────────────────────────────────────────────────────────

# ── Etapa 1: Compilación de Assets ──────────────────────────
FROM node:20-alpine AS builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /build

# Cacheo de dependencias
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copiar fuentes y compilar
COPY . .

# Variables de entorno de build (inyectadas en docker-compose.prod.yml)
ARG VITE_API_URL
ARG VITE_GOOGLE_CLIENT_ID
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_GOOGLE_CLIENT_ID=${VITE_GOOGLE_CLIENT_ID}

RUN pnpm run build

# ── Etapa 2: Servidor Estático Nginx ────────────────────────
FROM nginx:alpine

# Eliminar la configuración por defecto de Nginx
RUN rm /etc/nginx/conf.d/default.conf

# Copiar configuración personalizada para SPA
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copiar assets compilados desde la etapa de build
COPY --from=builder /build/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
