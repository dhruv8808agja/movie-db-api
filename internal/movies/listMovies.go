package movies

import (
	"net/http"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

type paginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type paginatedMoviesResponse struct {
	Data       []models.Movie `json:"data"`
	Pagination paginationMeta `json:"pagination"`
}

func ListMovies(c *gin.Context) {
	page, pageSize := GetPagination(c)

	var (
		items []models.Movie
		total int64
	)

	q := storage.DB.Model(&models.Movie{})

	// Total count for pagination
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count movies"})
		return
	}

	// Page slice
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize)) // ceil
	}

	resp := paginatedMoviesResponse{
		Data: items,
		Pagination: paginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1 && totalPages > 0,
		},
	}

	c.JSON(http.StatusOK, resp)
}
