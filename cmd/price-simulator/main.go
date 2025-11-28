package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/hampusjohansson-data/realtime-pricing-platform/internal/pricing"
)

func main() {
	ctx := context.Background()

	broker := envOrDefault("KAFKA_BROKER", "localhost:29092")
	topic := envOrDefault("KAFKA_TOPIC_TICKS", "price_ticks")
	symbolsEnv := envOrDefault("SYMBOLS", "BTC-USD,ETH-USD,SOL-USD")

	symbols := strings.Split(symbolsEnv, ",")

	log.Printf("Starting price simulator | broker=%s | topic=%s | symbols=%v",
		broker, topic, symbols,
	)

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	// lite slumpbaserad startpris per symbol
	base := map[string]float64{}
	for _, s := range symbols {
		switch s {
		case "BTC-USD":
			base[s] = 50000
		case "ETH-USD":
			base[s] = 3000
		default:
			base[s] = 100
		}
	}

	rand.Seed(time.Now().UnixNano())

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for t := range ticker.C {
		for _, sym := range symbols {
			// slumpa lite kring baspris
			delta := (rand.Float64() - 0.5) * 0.02 // ±1%
			price := base[sym] * (1 + delta)
			volume := 0.1 + rand.Float64()*5

			tick := pricing.PriceTick{
				Symbol: sym,
				Price:  price,
				Volume: volume,
				Ts:     t.UTC(),
			}

			payload, err := json.Marshal(tick)
			if err != nil {
				log.Printf("failed to marshal tick: %v", err)
				continue
			}

			err = writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte(sym),
				Value: payload,
			})
			if err != nil {
				log.Printf("failed to write message to kafka: %v", err)
				continue
			}

			log.Printf("produced tick | symbol=%s price=%.2f volume=%.4f", sym, price, volume)
		}
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
