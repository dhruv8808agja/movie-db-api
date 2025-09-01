package main

import (
	"github.com/gin-gonic/gin"

	"github.com/dhruv8808agja/movie-db-api/internal/auth"
	"github.com/dhruv8808agja/movie-db-api/internal/movies"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
)

func main() {
	// Initialize database (if needed)
	storage.InitDB()

	// Initialize Redis
	storage.InitRedis()

	// Setup Gin router
	r := gin.Default()

	// Public routes
	// Authentication
	r.POST("/login", auth.Login)

	// Read
	r.GET("/movies", movies.ListMoviesWithFilter)

	// Secured routes
	secured := r.Group("/")
	secured.Use(auth.JWTMiddleware())
	// Movie routes
	// Create
	secured.POST("/movies", movies.CreateMovie)
	secured.POST("/movies/bulk", movies.CreateMovies)

	// Read
	r.GET("/movies/:id", movies.GetMovieByID)

	// Update
	secured.PUT("/movies/:id", movies.UpdateMovie)
	// Delete
	secured.DELETE("/movies/:id", movies.DeleteMovie)
	secured.DELETE("/movies", movies.DeleteMovies)

	r.Run(":8080")
}
