package main

import (
	"github.com/gin-gonic/gin"

	"github.com/dhruv8808agja/movie-db-api/internal/movies"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
)

func main() {
	// Initialize database (if needed)
	storage.InitDB()

	// Setup Gin router
	r := gin.Default()

	// Movie routes
	// Create
	r.POST("/movies", movies.CreateMovie)
	r.POST("/movies/bulk", movies.CreateMovies)
	// Read
	r.GET("/movies", movies.ListMovies)
	r.GET("/movies/:id", movies.GetMovie)
	// Update
	r.PUT("/movies/:id", movies.UpdateMovie)
	// Delete
	r.DELETE("/movies/:id", movies.DeleteMovie)
	r.DELETE("/movies", movies.DeleteMovies)

	r.Run(":8080")
}
