package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/hampusjohansson-data/realtime-pricing-platform/internal/postgres"
	redisclient "github.com/hampusjohansson-data/realtime-pricing-platform/internal/redis"
)

type Processor struct {
	db  *sql.DB
	rdb *redis.Client
}

type Tick struct {
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Volume float64   `json:"volume"`
	Time   time.Time `json:"ts"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// --- Config från env ---
	kafkaBroker := envOrDefault("KAFKA_BROKER", "localhost:29092")
	kafkaTopic := envOrDefault("KAFKA_TOPIC_TICKS", "price_ticks")
	kafkaGroup := envOrDefault("KAFKA_GROUP_ID", "pipeline-processor")

	log.Printf("Starting pipeline-processor | broker=%s | topic=%s | group=%s",
		kafkaBroker, kafkaTopic, kafkaGroup,
	)

	// --- Postgres & Redis ---
	db := postgres.New()
	defer db.Close()

	rdb := redisclient.New()
	defer rdb.Close()

	p := &Processor{
		db:  db,
		rdb: rdb,
	}

	// --- Kafka Reader ---
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   kafkaTopic,
		GroupID: kafkaGroup,
	})
	defer reader.Close()

	log.Println("pipeline-processor: connected to Kafka, starting consume loop...")

	ctx := context.Background()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("failed to read message from kafka: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var tick Tick
		if err := json.Unmarshal(msg.Value, &tick); err != nil {
			log.Printf("failed to unmarshal tick: %v", err)
			continue
		}

		// TODO: riktig anomali-detektion
		isAnomaly := false

		if err := p.storeTick(ctx, tick, isAnomaly); err != nil {
			log.Printf("failed to store tick: %v", err)
		} else {
			log.Printf("processed tick | symbol=%s price=%.2f volume=%.4f anomaly=%v",
				tick.Symbol, tick.Price, tick.Volume, isAnomaly)
		}
	}
}

func (p *Processor) storeTick(ctx context.Context, t Tick, isAnomaly bool) error {
	// 1) Skriv till Postgres
	const q = `
		INSERT INTO price_ticks_enriched
		    (symbol, price, volume, ts, ma_1m, ma_5m, vol_1m, is_anomaly)
		VALUES ($1,     $2,    $3,    $4, NULL,  NULL,  NULL,   $5)
	`
	if _, err := p.db.ExecContext(ctx, q,
		t.Symbol,
		t.Price,
		t.Volume,
		t.Time,
		isAnomaly,
	); err != nil {
		return fmt.Errorf("postgres insert: %w", err)
	}

	// 2) Uppdatera senaste priset i Redis så API:t kan läsa det
	key := "price:" + t.Symbol

	isAnomalyStr := "0"
	if isAnomaly {
		isAnomalyStr = "1"
	}

	fields := map[string]interface{}{
		"symbol":     t.Symbol,
		"price":      fmt.Sprintf("%f", t.Price),
		"volume":     fmt.Sprintf("%f", t.Volume),
		"ts":         t.Time.Format(time.RFC3339Nano),
		"is_anomaly": isAnomalyStr,
	}

	if err := p.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("redis HSET: %w", err)
	}

	return nil
}

// samma helper som i api-gateway
func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
