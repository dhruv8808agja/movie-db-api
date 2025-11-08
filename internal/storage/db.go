package storage

import (
	"log"

	"github.com/dhruv8808agja/movie-db-api/internal/config"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var User = models.User{}
var Movie = models.Movie{}
var Video = models.Video{}
var UploadSession = models.UploadSession{}
var TranscodedVideo = models.TranscodedVideo{}
var TranscodingJob = models.TranscodingJob{}
var UserInteraction = models.UserInteraction{}
var UserMovieRating = models.UserMovieRating{}
var UserPreferences = models.UserPreferences{}

// InitDB initializes the database connection and performs auto-migration.
func InitDB() {
	cfg := config.LoadConfig()

	var err error
	var dialector gorm.Dialector

	switch cfg.Database.Driver {
	case "postgres":
		dsn := cfg.Database.GetDSN()
		log.Printf("Connecting to PostgreSQL database with DSN: %s\n", dsn)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dsn := cfg.Database.GetDSN()
		log.Printf("Connecting to SQLite database: %s\n", dsn)
		dialector = sqlite.Open(dsn)
	default:
		log.Fatalf("Unsupported database driver: %s", cfg.Database.Driver)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Auto-migrate schema
	err = DB.AutoMigrate(
		&User,
		&Movie,
		&Video,
		&UploadSession,
		&TranscodedVideo,
		&TranscodingJob,
		&UserInteraction,
		&UserMovieRating,
		&UserPreferences,
	)
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
	log.Printf("Database connection established (%s) and migrated\n", cfg.Database.Driver)
}
