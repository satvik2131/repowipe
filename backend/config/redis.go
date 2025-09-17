package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()
var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // e.g. "localhost:6379"
		Password: os.Getenv("REDIS_PASSWORD"), // empty if no password
		DB:       0,                           // default DB
	})

	// Test connection
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Could not connect to Redis: %v", err)
	} else {
		log.Println("✅ Connected to Redis")
	}
}
