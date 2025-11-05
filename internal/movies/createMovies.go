package movies

import (
	"fmt"
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// CreateMovie creates a new movie in the database
// @Summary      Create a new movie
// @Description  Create a new movie with the provided details. Requires authentication.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        movie  body      models.Movie  true  "Movie object to create"
// @Success      201    {object}  models.Movie  "Successfully created movie"
// @Failure      400    {object}  map[string]string  "Invalid request or validation error"
// @Failure      500    {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /movies [post]
func CreateMovie(c *gin.Context) {
	var newMovie models.Movie

	if err := c.BindJSON(&newMovie); err != nil {
		logger.Log.Error("failed to bind movie JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate movie data
	if err := ValidateMovie(&newMovie); err != nil {
		logger.Log.Warn("movie validation failed", zap.Error(err))
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

// CreateMovies creates multiple movies in bulk
// @Summary      Create multiple movies
// @Description  Create multiple movies in a single request. Requires authentication.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        movies  body      []models.Movie  true  "Array of movie objects to create"
// @Success      201     {object}  []models.Movie  "Successfully created movies"
// @Failure      400     {object}  map[string]string  "Invalid request or validation error"
// @Failure      500     {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /movies/bulk [post]
func CreateMovies(c *gin.Context) {
	var newMovies []models.Movie

	if err := c.BindJSON(&newMovies); err != nil {
		logger.Log.Error("failed to bind movies JSON", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate each movie
	for i, movie := range newMovies {
		if err := ValidateMovie(&movie); err != nil {
			logger.Log.Warn("movie validation failed in bulk create",
				zap.Int("index", i),
				zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "validation failed for movie at index " + fmt.Sprint(i) + ": " + err.Error(),
			})
			return
		}
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
