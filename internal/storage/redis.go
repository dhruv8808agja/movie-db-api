package storage

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379" // default value
	}

	password := os.Getenv("REDIS_PASSWORD")
	// Empty password is acceptable, so no default needed

	dbStr := os.Getenv("REDIS_DB")
	db := 0 // default value
	if dbStr != "" {
		var err error
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			log.Printf("Invalid REDIS_DB value '%s', using default 0", dbStr)
			db = 0
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
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
