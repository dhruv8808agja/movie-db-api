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
	"golang.org/x/crypto/bcrypt"
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
	if err := db.AutoMigrate(
		&models.Movie{},
		&models.User{},
		&models.Video{},
		&models.UploadSession{},
		&models.TranscodedVideo{},
		&models.TranscodingJob{},
		&models.UserInteraction{},
		&models.UserMovieRating{},
		&models.UserPreferences{},
		&models.WatchHistory{},
	); err != nil {
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

// CreateTestUser creates a test user with a hashed password
// Note: The password is hashed using bcrypt
func CreateTestUser(username, password string, role models.UserRole, isActive bool) *models.User {
	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("failed to hash password:", err)
	}

	return &models.User{
		Username: username,
		Email:    username + "@test.com",
		Password: string(hashedPassword),
		Role:     role,
		IsActive: isActive,
	}
}

// SeedTestUsers seeds the database with default test users
func SeedTestUsers(db *gorm.DB) {
	users := []*models.User{
		CreateTestUser("admin", "password", models.RoleAdmin, true),
		CreateTestUser("user", "password", models.RoleUser, true),
		CreateTestUser("moderator", "password", models.RoleModerator, true),
		CreateTestUser("inactive", "password", models.RoleUser, false),
	}

	for _, user := range users {
		db.Create(user)
	}
}

// SeedTestMovies seeds the database with default test movies
func SeedTestMovies(db *gorm.DB) {
	movies := []models.Movie{
		{
			Title:       "The Matrix",
			Description: "A computer hacker learns about the true nature of reality",
			Director:    "The Wachowskis",
			ReleaseDate: time.Date(1999, 3, 31, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Action", "Sci-Fi"},
			Rating:      8.7,
		},
		{
			Title:       "Inception",
			Description: "A thief who steals corporate secrets through dream-sharing technology",
			Director:    "Christopher Nolan",
			ReleaseDate: time.Date(2010, 7, 16, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Action", "Sci-Fi", "Thriller"},
			Rating:      8.8,
		},
		{
			Title:       "The Shawshank Redemption",
			Description: "Two imprisoned men bond over a number of years",
			Director:    "Frank Darabont",
			ReleaseDate: time.Date(1994, 9, 23, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Drama"},
			Rating:      9.3,
		},
		{
			Title:       "Pulp Fiction",
			Description: "The lives of two mob hitmen, a boxer, and a pair of diner bandits intertwine",
			Director:    "Quentin Tarantino",
			ReleaseDate: time.Date(1994, 10, 14, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Crime", "Drama"},
			Rating:      8.9,
		},
		{
			Title:       "The Dark Knight",
			Description: "Batman faces the Joker, a criminal mastermind",
			Director:    "Christopher Nolan",
			ReleaseDate: time.Date(2008, 7, 18, 0, 0, 0, 0, time.UTC),
			Genres:      models.Genres{"Action", "Crime", "Drama"},
			Rating:      9.0,
		},
	}

	for _, movie := range movies {
		db.Create(&movie)
	}
}
