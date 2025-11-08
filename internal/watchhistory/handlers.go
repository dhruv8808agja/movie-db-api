package watchhistory

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RecordWatchRequest represents the request to record a watch history entry
type RecordWatchRequest struct {
	MovieID       uint    `json:"movie_id" binding:"required" example:"1"`
	WatchDuration int     `json:"watch_duration" binding:"required,min=0" example:"3600"`
	LastPosition  int     `json:"last_position" binding:"required,min=0" example:"2700"`
	Progress      float32 `json:"progress" binding:"required,min=0,max=100" example:"75.5"`
	Completed     bool    `json:"completed" example:"false"`
	Quality       string  `json:"quality" example:"1080p"`
	DeviceType    string  `json:"device_type" example:"web"`
}

// RecordWatchResponse represents the response after recording watch history
type RecordWatchResponse struct {
	Message      string               `json:"message" example:"Watch history recorded successfully"`
	WatchHistory models.WatchHistory  `json:"watch_history"`
}

// RecordWatch records or updates a watch history entry for a user
// @Summary      Record watch history
// @Description  Record or update watch history for a movie. Authenticated users only.
// @Tags         watch-history
// @Accept       json
// @Produce      json
// @Param        watch  body      RecordWatchRequest   true  "Watch history data"
// @Success      200    {object}  RecordWatchResponse  "Watch history recorded successfully"
// @Failure      400    {object}  map[string]string    "Invalid request"
// @Failure      401    {object}  map[string]string    "Unauthorized"
// @Failure      404    {object}  map[string]string    "Movie not found"
// @Failure      500    {object}  map[string]string    "Internal server error"
// @Security     BearerAuth
// @Router       /watch-history [post]
func RecordWatch(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req RecordWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify movie exists
	var movie models.Movie
	if err := storage.DB.First(&movie, req.MovieID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify movie"})
		return
	}

	// Check if watch history already exists for today
	var existingHistory models.WatchHistory
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := storage.DB.Where("user_id = ? AND movie_id = ? AND watched_at >= ? AND watched_at < ?",
		userID, req.MovieID, today, tomorrow).First(&existingHistory).Error

	if err == nil {
		// Update existing entry
		existingHistory.WatchDuration += req.WatchDuration
		existingHistory.LastPosition = req.LastPosition
		existingHistory.Progress = req.Progress
		existingHistory.Completed = req.Completed
		existingHistory.Quality = req.Quality
		existingHistory.DeviceType = req.DeviceType
		existingHistory.WatchedAt = time.Now()
		existingHistory.WatchCount++

		if err := storage.DB.Save(&existingHistory).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update watch history"})
			return
		}

		c.JSON(http.StatusOK, RecordWatchResponse{
			Message:      "Watch history updated successfully",
			WatchHistory: existingHistory,
		})
		return
	}

	// Create new entry
	watchHistory := models.WatchHistory{
		UserID:        userID.(uint),
		MovieID:       req.MovieID,
		WatchedAt:     time.Now(),
		WatchDuration: req.WatchDuration,
		Progress:      req.Progress,
		Completed:     req.Completed,
		LastPosition:  req.LastPosition,
		WatchCount:    1,
		Quality:       req.Quality,
		DeviceType:    req.DeviceType,
	}

	if err := storage.DB.Create(&watchHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record watch history"})
		return
	}

	c.JSON(http.StatusOK, RecordWatchResponse{
		Message:      "Watch history recorded successfully",
		WatchHistory: watchHistory,
	})
}

// GetWatchHistory retrieves the watch history for the authenticated user
// @Summary      Get watch history
// @Description  Get watch history for the authenticated user with pagination
// @Tags         watch-history
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        page_size query     int     false  "Page size (default: 20, max: 100)"
// @Param        movie_id  query     int     false  "Filter by movie ID"
// @Param        completed query     bool    false  "Filter by completion status"
// @Success      200       {object}  map[string]interface{}  "Watch history retrieved successfully"
// @Failure      400       {object}  map[string]string       "Invalid request"
// @Failure      401       {object}  map[string]string       "Unauthorized"
// @Failure      500       {object}  map[string]string       "Internal server error"
// @Security     BearerAuth
// @Router       /watch-history [get]
func GetWatchHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Build query
	query := storage.DB.Where("user_id = ?", userID)

	// Filters
	if movieID := c.Query("movie_id"); movieID != "" {
		query = query.Where("movie_id = ?", movieID)
	}
	if completed := c.Query("completed"); completed != "" {
		completedBool, _ := strconv.ParseBool(completed)
		query = query.Where("completed = ?", completedBool)
	}

	// Get total count
	var total int64
	query.Model(&models.WatchHistory{}).Count(&total)

	// Get watch history with movie details
	var history []models.WatchHistory
	if err := query.Order("watched_at DESC").
		Limit(pageSize).
		Offset(offset).
		Preload("Movie").
		Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve watch history"})
		return
	}

	// Convert to response with movie details
	var historyWithMovies []models.WatchHistoryWithMovie
	for _, h := range history {
		historyWithMovies = append(historyWithMovies, models.WatchHistoryWithMovie{
			ID:            h.ID,
			UserID:        h.UserID,
			MovieID:       h.MovieID,
			WatchedAt:     h.WatchedAt,
			WatchDuration: h.WatchDuration,
			Progress:      h.Progress,
			Completed:     h.Completed,
			LastPosition:  h.LastPosition,
			WatchCount:    h.WatchCount,
			Quality:       h.Quality,
			DeviceType:    h.DeviceType,
			Movie:         h.Movie,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": historyWithMovies,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetContinueWatching retrieves movies that the user is currently watching
// @Summary      Get continue watching
// @Description  Get list of movies that the user is currently watching (not completed)
// @Tags         watch-history
// @Accept       json
// @Produce      json
// @Param        limit query     int     false  "Limit number of results (default: 10, max: 50)"
// @Success      200   {array}   models.ContinueWatching  "Continue watching list retrieved successfully"
// @Failure      401   {object}  map[string]string        "Unauthorized"
// @Failure      500   {object}  map[string]string        "Internal server error"
// @Security     BearerAuth
// @Router       /watch-history/continue-watching [get]
func GetContinueWatching(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Get latest watch history for each movie that is not completed
	var continueWatching []models.ContinueWatching

	// Use raw SQL for better performance with GROUP BY
	err := storage.DB.Raw(`
		SELECT
			wh.id as watch_history_id,
			wh.movie_id,
			wh.last_position,
			wh.progress,
			wh.watched_at
		FROM watch_histories wh
		INNER JOIN (
			SELECT movie_id, MAX(watched_at) as max_watched
			FROM watch_histories
			WHERE user_id = ? AND completed = false AND progress > 5
			GROUP BY movie_id
		) latest ON wh.movie_id = latest.movie_id AND wh.watched_at = latest.max_watched
		WHERE wh.user_id = ?
		ORDER BY wh.watched_at DESC
		LIMIT ?
	`, userID, userID, limit).Scan(&continueWatching).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve continue watching list"})
		return
	}

	// Load movie details
	for i := range continueWatching {
		var movie models.Movie
		if err := storage.DB.First(&movie, continueWatching[i].MovieID).Error; err == nil {
			continueWatching[i].Movie = movie
		}
	}

	c.JSON(http.StatusOK, continueWatching)
}

// GetWatchStats retrieves watch statistics for the authenticated user
// @Summary      Get watch statistics
// @Description  Get aggregated watch statistics for the authenticated user
// @Tags         watch-history
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.WatchStats     "Watch statistics retrieved successfully"
// @Failure      401  {object}  map[string]string     "Unauthorized"
// @Failure      500  {object}  map[string]string     "Internal server error"
// @Security     BearerAuth
// @Router       /watch-history/stats [get]
func GetWatchStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var stats models.WatchStats

	// Get total unique movies watched
	var totalMovies int64
	storage.DB.Model(&models.WatchHistory{}).
		Where("user_id = ?", userID).
		Distinct("movie_id").
		Count(&totalMovies)
	stats.TotalMoviesWatched = int(totalMovies)

	// Get total watch time
	var totalTime struct {
		Total int
	}
	storage.DB.Model(&models.WatchHistory{}).
		Select("COALESCE(SUM(watch_duration), 0) as total").
		Where("user_id = ?", userID).
		Scan(&totalTime)
	stats.TotalWatchTime = totalTime.Total

	// Get completed movies count
	var completedCount int64
	storage.DB.Model(&models.WatchHistory{}).
		Where("user_id = ? AND completed = true", userID).
		Distinct("movie_id").
		Count(&completedCount)
	stats.CompletedMovies = int(completedCount)

	// Get in-progress movies count
	var inProgressCount int64
	storage.DB.Model(&models.WatchHistory{}).
		Where("user_id = ? AND completed = false AND progress > 5", userID).
		Distinct("movie_id").
		Count(&inProgressCount)
	stats.InProgressMovies = int(inProgressCount)

	// Get last watched timestamp
	var lastWatch models.WatchHistory
	if err := storage.DB.Where("user_id = ?", userID).
		Order("watched_at DESC").
		First(&lastWatch).Error; err == nil {
		stats.LastWatchedAt = lastWatch.WatchedAt
	}

	// Get average progress
	var avgProgress struct {
		Avg float32
	}
	storage.DB.Model(&models.WatchHistory{}).
		Select("COALESCE(AVG(progress), 0) as avg").
		Where("user_id = ?", userID).
		Scan(&avgProgress)
	stats.AverageProgress = avgProgress.Avg

	// Get favorite genres from watched movies
	// Load movies that user has watched and extract genres
	var watchedMovies []models.Movie
	storage.DB.Where("id IN (?)",
		storage.DB.Table("watch_histories").
			Select("DISTINCT movie_id").
			Where("user_id = ?", userID),
	).Find(&watchedMovies)

	// Count genre frequency
	genreCount := make(map[string]int)
	for _, movie := range watchedMovies {
		for _, genre := range movie.Genres {
			genreCount[genre]++
		}
	}

	// Get top 5 genres
	type genrePair struct {
		genre string
		count int
	}
	var genrePairs []genrePair
	for genre, count := range genreCount {
		genrePairs = append(genrePairs, genrePair{genre, count})
	}

	// Sort by count (simple bubble sort for small dataset)
	for i := 0; i < len(genrePairs); i++ {
		for j := i + 1; j < len(genrePairs); j++ {
			if genrePairs[j].count > genrePairs[i].count {
				genrePairs[i], genrePairs[j] = genrePairs[j], genrePairs[i]
			}
		}
	}

	// Extract top 5 genre names
	var genres []string
	limit := 5
	if len(genrePairs) < limit {
		limit = len(genrePairs)
	}
	for i := 0; i < limit; i++ {
		genres = append(genres, genrePairs[i].genre)
	}
	stats.FavoriteGenres = genres

	c.JSON(http.StatusOK, stats)
}

// DeleteWatchHistory deletes a specific watch history entry
// @Summary      Delete watch history entry
// @Description  Delete a specific watch history entry by ID
// @Tags         watch-history
// @Accept       json
// @Produce      json
// @Param        id   path      int               true  "Watch History ID"
// @Success      200  {object}  map[string]string "Watch history deleted successfully"
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Forbidden - not your watch history"
// @Failure      404  {object}  map[string]string "Watch history not found"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     BearerAuth
// @Router       /watch-history/{id} [delete]
func DeleteWatchHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid watch history ID"})
		return
	}

	var watchHistory models.WatchHistory
	if err := storage.DB.First(&watchHistory, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Watch history not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve watch history"})
		return
	}

	// Verify ownership
	if watchHistory.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own watch history"})
		return
	}

	if err := storage.DB.Delete(&watchHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete watch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watch history deleted successfully"})
}

// ClearWatchHistory clears all watch history for the authenticated user
// @Summary      Clear all watch history
// @Description  Delete all watch history entries for the authenticated user
// @Tags         watch-history
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string "Watch history cleared successfully"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     BearerAuth
// @Router       /watch-history/clear [delete]
func ClearWatchHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	result := storage.DB.Where("user_id = ?", userID).Delete(&models.WatchHistory{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear watch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Watch history cleared successfully",
		"deleted_count": result.RowsAffected,
	})
}
