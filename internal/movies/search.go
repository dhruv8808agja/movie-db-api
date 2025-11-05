package movies

import (
	"net/http"
	"strings"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// ListMoviesWithFilter retrieves a list of movies with optional filters and pagination
// @Summary      List movies with filters
// @Description  Get a paginated list of movies with optional search filters. Public endpoint (no authentication required).
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number"          default(1)      minimum(1)
// @Param        limit     query     int     false  "Items per page"       default(10)     minimum(1)      maximum(100)
// @Param        title     query     string  false  "Filter by title (case-insensitive partial match)"
// @Param        director  query     string  false  "Filter by director (case-insensitive partial match)"
// @Param        genre     query     string  false  "Filter by genre"
// @Success      200       {object}  paginatedMoviesResponse  "List of movies with pagination info"
// @Failure      500       {object}  map[string]string        "Internal server error"
// @Router       /movies [get]
func ListMoviesWithFilter(c *gin.Context) {
	page, pageSize := GetPagination(c)

	// query params
	title := c.Query("title")
	director := c.Query("director")
	genre := c.Query("genre")

	var (
		items []models.Movie
		total int64
	)

	q := storage.DB.Model(&models.Movie{})

	// Apply filters if provided
	if title != "" {
		q = q.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}
	if director != "" {
		q = q.Where("LOWER(director) LIKE ?", "%"+strings.ToLower(director)+"%")
	}
	if genre != "" {
		// For JSON stored genres, use SQLite JSON functions or LIKE for simplicity
		q = q.Where("genres LIKE ?", "%"+genre+"%")
	}

	// Count total after filtering
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count movies"})
		return
	}

	// Apply pagination
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	c.JSON(http.StatusOK, paginatedMoviesResponse{
		Data: items,
		Pagination: paginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1 && totalPages > 0,
		},
	})
}
