package main

import (
	"github.com/gin-gonic/gin"

	"github.com/dhruv8808agja/movie-db-api/internal/auth"
	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/monitor"
	"github.com/dhruv8808agja/movie-db-api/internal/movies"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
)

func main() {
	// Initialize logger
	logger.InitLogger()

	// Initialize DB and Redis
	storage.InitDB()
	storage.InitRedis()

	// Setup Gin router
	r := gin.New()
	r.Use(logger.GinLogger(), gin.Recovery()) // logging + recovery

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
	secured.GET("/movies/:id", movies.GetMovieByID)

	// Update
	secured.PUT("/movies/:id", movies.UpdateMovie)
	// Delete
	secured.DELETE("/movies/:id", movies.DeleteMovie)
	secured.DELETE("/movies", movies.DeleteMovies)

	// Prometheus metrics
	monitor.RegisterMetrics(r)

	r.Run(":8080")
}
