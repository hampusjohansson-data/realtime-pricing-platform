#!/usr/bin/env bash
set -euo pipefail

# --- ENV variabler som alla Go-tjänster använder ---
export KAFKA_BROKER=localhost:29092
export KAFKA_TOPIC_TICKS=price_ticks
export KAFKA_GROUP_ID=pipeline-processor
export REDIS_ADDR=localhost:6379
export POSTGRES_DSN=postgres://pricing:pricing@localhost:5432/pricing?sslmode=disable
export HTTP_ADDR=":8088"

echo "Starting dev stack (simulator + pipeline + api + web-ui)..."

# När du trycker Ctrl+C -> döda alla child-processer
trap 'echo "Stopping dev stack..."; kill 0' EXIT

# --- Starta price-simulator ---
go run ./cmd/price-simulator &
SIM_PID=$!
echo "price-simulator PID=$SIM_PID"

# --- Starta pipeline-processor ---
go run ./cmd/pipeline-processor &
PIPE_PID=$!
echo "pipeline-processor PID=$PIPE_PID"

# --- Starta API-gateway ---
go run ./cmd/api-gateway &
API_PID=$!
echo "api-gateway PID=$API_PID"

# --- Starta web-ui (Vite) ---
cd web-ui
npm run dev &
WEB_PID=$!
echo "web-ui (Vite) PID=$WEB_PID"

echo ""
echo "All services started."
echo "→ API:     http://localhost:8088"
echo "→ Web-UI:  http://localhost:5173 (eller 5174 om 5173 är upptagen)"
echo ""
echo "Tryck Ctrl+C för att stoppa allt."

# Vänta på alla barn-processer
wait

