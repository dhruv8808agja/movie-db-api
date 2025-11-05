package testutil

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to create test database:", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&models.Movie{}); err != nil {
		log.Fatal("failed to migrate test database:", err)
	}

	return db
}

// SetupTestRedis creates a test Redis client (using a test DB number)
func SetupTestRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       15, // Use DB 15 for testing
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis not available for tests: %v", err)
		return nil
	}

	return client
}

// CleanupTestRedis flushes the test Redis database
func CleanupTestRedis(client *redis.Client) {
	if client != nil {
		ctx := context.Background()
		client.FlushDB(ctx)
	}
}

// SetTestEnv sets environment variables for testing
func SetTestEnv() {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("RATE_LIMIT_REQUESTS", "10")
	os.Setenv("RATE_LIMIT_WINDOW_SECONDS", "60")

	// Initialize logger for tests (if not already initialized)
	if logger.Log == nil {
		logger.InitLogger()
	}
}

// ClearTestEnv clears test environment variables
func ClearTestEnv() {
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("RATE_LIMIT_REQUESTS")
	os.Unsetenv("RATE_LIMIT_WINDOW_SECONDS")
}

// InitTestStorage initializes test storage (DB and Redis)
func InitTestStorage() {
	storage.DB = SetupTestDB()
	storage.RedisClient = SetupTestRedis()
	storage.Ctx = context.Background()
}

// CreateTestMovie creates a test movie with default values
func CreateTestMovie() *models.Movie {
	return &models.Movie{
		Title:       "Test Movie",
		Description: "A test movie description",
		Director:    "Test Director",
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Genres:      models.Genres{"Action", "Drama"},
		Rating:      8.5,
	}
}

// CreateTestMovies creates multiple test movies
func CreateTestMovies(count int) []models.Movie {
	movies := make([]models.Movie, count)
	for i := 0; i < count; i++ {
		movies[i] = models.Movie{
			Title:       "Test Movie " + string(rune(i+1)),
			Description: "Description for test movie",
			Director:    "Director " + string(rune(i+1)),
			ReleaseDate: time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Action"},
			Rating:      float32(i) + 5.0,
		}
	}
	return movies
}
