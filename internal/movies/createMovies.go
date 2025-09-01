package movies

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func CreateMovie(c *gin.Context) {
	var newMovie models.Movie

	if err := c.BindJSON(&newMovie); err != nil {
		logger.Log.Error("failed to bind movie JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save to DB
	if err := storage.DB.Create(&newMovie).Error; err != nil {
		logger.Log.Error("failed to create movie in DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Info("movie created successfully", zap.Int("movie_id", int(newMovie.ID)))
	c.JSON(http.StatusCreated, newMovie)
}

func CreateMovies(c *gin.Context) {
	var newMovies []models.Movie

	if err := c.BindJSON(&newMovies); err != nil {
		logger.Log.Error("failed to bind movies JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save to DB
	if err := storage.DB.Create(&newMovies).Error; err != nil {
		logger.Log.Error("failed to create movies in DB", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Info("movies created successfully", zap.Int("count", len(newMovies)))
	c.JSON(http.StatusCreated, newMovies)
}
