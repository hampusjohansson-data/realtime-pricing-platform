package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func New() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		// fallback om env-variabel saknas
		addr = "localhost:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0, // default-databasen
	})
}
