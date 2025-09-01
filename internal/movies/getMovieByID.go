package movies

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

func GetMovieByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	cacheKey := "movie:" + idStr
	cached, err := storage.GetCache(cacheKey)
	if err == nil {
		// Cache hit
		var m models.Movie
		if err := json.Unmarshal([]byte(cached), &m); err == nil {
			c.JSON(http.StatusOK, m)
			return
		}
	}

	// Cache miss → fetch from DB
	var m models.Movie
	if err := storage.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	// Save to Redis for next time
	data, _ := json.Marshal(m)
	_ = storage.SetCache(cacheKey, data, 5*time.Minute)

	c.JSON(http.StatusOK, m)
}
