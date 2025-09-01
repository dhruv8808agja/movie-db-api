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

	r.POST("/movies", movies.CreateMovie)
	r.GET("/movies", movies.ListMovies)
	r.GET("/movies/:id", movies.GetMovie)
	r.PUT("/movies/:id", movies.UpdateMovie)
	r.DELETE("/movies/:id", movies.DeleteMovie)

	r.Run(":8080")
}
