package storage

import (
	"log"

	"github.com/dhruv8808agja/movie-db-api/pkg/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var Movie = models.Movie{}

// InitDB initializes the database connection and performs auto-migration.
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("movies.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Auto-migrate schema
	err = DB.AutoMigrate(&Movie)
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}
	log.Println("Database connection established and migrated")
}
