package storage

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password by default
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("failed to connect Redis:", err)
	}

	log.Println("Redis connected")
}

// Set with TTL
func SetCache(key string, value interface{}, ttl time.Duration) error {
	return RedisClient.Set(Ctx, key, value, ttl).Err()
}

// Get
func GetCache(key string) (string, error) {
	return RedisClient.Get(Ctx, key).Result()
}
