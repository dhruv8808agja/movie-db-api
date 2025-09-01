package crudMovies

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/pkg/models/movie"

	"github.com/gin-gonic/gin"
)

var movies = make(map[string]movie.Movie) // in-memory store

func createMovie(c *gin.Context) {
	var newMovie movie.Movie

	if err := c.BindJSON(&newMovie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movies[newMovie.ID] = newMovie
	c.JSON(http.StatusCreated, newMovie)
}
