#!/bin/sh
set -e

echo "[MÍAS] Inicializando contenedor Backend..."

# Ejecutar migraciones automáticas si DATABASE_URL está configurada
if [ -n "$DATABASE_URL" ]; then
    echo "[MÍAS] Verificando y aplicando migraciones de base de datos..."
    ./mias-migrate || echo "[MÍAS] Advertencia: Migración omitida o con error no crítico, continuando inicio..."
fi

echo "[MÍAS] Iniciando servidor API en puerto ${PORT:-8080}..."
exec ./mias-api
