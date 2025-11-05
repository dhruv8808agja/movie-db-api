package movies

import (
	"fmt"
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"

	"github.com/gin-gonic/gin"
)

// DeleteMovie deletes a movie by ID
// @Summary      Delete movie
// @Description  Delete a movie by its ID. Requires authentication.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Movie ID"
// @Success      204  "Successfully deleted"
// @Failure      400  {object}  map[string]string  "Invalid ID"
// @Failure      404  {object}  map[string]string  "Movie not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /movies/{id} [delete]
func DeleteMovie(c *gin.Context) {
	idParam := c.Param("id")
	var id uint
	_, err := fmt.Sscanf(idParam, "%d", &id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var movie models.Movie
	if err := storage.DB.First(&movie, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	if err := storage.DB.Delete(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteMoviesRequest represents the request body for bulk delete
type DeleteMoviesRequest struct {
	IDs []uint `json:"ids" example:"1,2,3"`
}

// DeleteMovies deletes multiple movies by their IDs
// @Summary      Delete multiple movies
// @Description  Delete multiple movies in a single request. Requires authentication.
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        ids  body      DeleteMoviesRequest  true  "Array of movie IDs to delete"
// @Success      204  "Successfully deleted"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /movies [delete]
func DeleteMovies(c *gin.Context) {
	var ids struct {
		IDs []uint `json:"ids"`
	}

	if err := c.BindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(ids.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided"})
		return
	}

	if err := storage.DB.Where("id IN ?", ids.IDs).Delete(&models.Movie{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
