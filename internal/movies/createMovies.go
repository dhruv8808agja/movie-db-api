package movies

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"

	"github.com/gin-gonic/gin"
)

func CreateMovie(c *gin.Context) {
	var newMovie models.Movie

	if err := c.BindJSON(&newMovie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save to DB
	if err := storage.DB.Create(&newMovie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newMovie)
}

func CreateMovies(c *gin.Context) {
	var newMovies []models.Movie

	if err := c.BindJSON(&newMovies); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save to DB
	if err := storage.DB.Create(&newMovies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newMovies)
}
