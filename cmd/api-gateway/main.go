package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"

	"github.com/hampusjohansson-data/realtime-pricing-platform/internal/postgres"
	redisclient "github.com/hampusjohansson-data/realtime-pricing-platform/internal/redis"
)

type Server struct {
	db  *sql.DB
	rdb *redis.Client
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// init Postgres & Redis via dina helpers
	db := postgres.New()
	defer db.Close()

	rdb := redisclient.New()
	defer rdb.Close()

	srv := &Server{
		db:  db,
		rdb: rdb,
	}

	// Ändrat default till :8088 (prod/dev-port)
	addr := envOrDefault("HTTP_ADDR", ":8088")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS för React-devservern (Vite)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174", // 👈 Vite kör här hos dig
		},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Routes
	// Befintlig health-endpoint
	r.Get("/health", srv.handleHealth)
	// NY: /healthz specifikt för K8s probes
	r.Get("/healthz", srv.handleHealth)

	r.Get("/prices/{symbol}", srv.handleGetLatestPrice)
	r.Get("/prices/{symbol}/history", srv.handleGetPriceHistory)

	log.Printf("Starting API gateway on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

// ===== Handlers =====

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	// Kolla Postgres
	if err := s.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"db":     err.Error(),
		})
		return
	}

	// Kolla Redis
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"redis":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (s *Server) handleGetLatestPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	key := "price:" + symbol

	data, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "redis error: "+err.Error())
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusNotFound, "symbol not found")
		return
	}

	price, err := strconv.ParseFloat(data["price"], 64)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid price format in redis")
		return
	}

	volume, err := strconv.ParseFloat(data["volume"], 64)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid volume format in redis")
		return
	}

	ts, err := time.Parse(time.RFC3339Nano, data["ts"])
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid timestamp format in redis")
		return
	}

	isAnomaly := false
	if v, ok := data["is_anomaly"]; ok && v != "" {
		isAnomaly = v == "1"
	}

	resp := struct {
		Symbol    string    `json:"symbol"`
		Price     float64   `json:"price"`
		Volume    float64   `json:"volume"`
		Timestamp time.Time `json:"timestamp"`
		IsAnomaly bool      `json:"is_anomaly"`
	}{
		Symbol:    symbol,
		Price:     price,
		Volume:    volume,
		Timestamp: ts,
		IsAnomaly: isAnomaly,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetPriceHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	symbol := chi.URLParam(r, "symbol")

	if symbol == "" {
		writeError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	const q = `
		SELECT symbol, price, volume, ts, is_anomaly
		FROM price_ticks_enriched
		WHERE symbol = $1
		ORDER BY ts DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, q, symbol, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db query error: "+err.Error())
		return
	}
	defer rows.Close()

	type point struct {
		Symbol    string    `json:"symbol"`
		Price     float64   `json:"price"`
		Volume    float64   `json:"volume"`
		Timestamp time.Time `json:"timestamp"`
		IsAnomaly bool      `json:"is_anomaly"`
	}

	var result []point

	for rows.Next() {
		var p point
		if err := rows.Scan(&p.Symbol, &p.Price, &p.Volume, &p.Timestamp, &p.IsAnomaly); err != nil {
			writeError(w, http.StatusInternalServerError, "db scan error: "+err.Error())
			return
		}
		result = append(result, p)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "db rows error: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"symbol":  symbol,
		"limit":   limit,
		"count":   len(result),
		"history": result,
	})
}

// ===== helpers =====

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{
		"error": msg,
	})
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
