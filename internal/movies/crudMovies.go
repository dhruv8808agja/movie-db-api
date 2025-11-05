package movies

import (
	"fmt"
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"

	"github.com/gin-gonic/gin"
)

// UpdateMovie updates an existing movie
// @Summary      Update movie
// @Description  Update an existing movie by ID. Requires authentication.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id     path      int           true  "Movie ID"
// @Param        movie  body      models.Movie  true  "Updated movie object"
// @Success      200    {object}  models.Movie  "Successfully updated movie"
// @Failure      400    {object}  map[string]string  "Invalid ID or request"
// @Failure      404    {object}  map[string]string  "Movie not found"
// @Failure      500    {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /movies/{id} [put]
func UpdateMovie(c *gin.Context) {
	idParam := c.Param("id")
	var id uint
	_, err := fmt.Sscanf(idParam, "%d", &id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var updatedMovie models.Movie
	if err := c.BindJSON(&updatedMovie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate movie data
	if err := ValidateMovie(&updatedMovie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var movie models.Movie
	if err := storage.DB.First(&movie, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	updatedMovie.ID = id
	if err := storage.DB.Save(&updatedMovie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedMovie)
}
