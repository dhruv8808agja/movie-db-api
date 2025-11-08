package recommendations

import (
	"math"
	"net/http"
	"strconv"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var engine = NewEngine()

// GetRecommendationsForUser returns personalized movie recommendations
// @Summary      Get personalized recommendations
// @Description  Get movie recommendations based on user's watch history and ratings
// @Tags         recommendations
// @Accept       json
// @Produce      json
// @Param        limit  query     int  false  "Number of recommendations (default 10)"
// @Success      200    {array}   models.MovieRecommendation
// @Failure      401    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Security     BearerAuth
// @Router       /recommendations [get]
func GetRecommendationsForUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	recommendations, err := engine.GetRecommendations(userID.(uint), limit)
	if err != nil {
		logger.Log.Error("failed to get recommendations",
			zap.Uint("user_id", userID.(uint)),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// RateMovie allows a user to rate a movie
// @Summary      Rate a movie
// @Description  Submit or update a rating for a movie
// @Tags         recommendations
// @Accept       json
// @Produce      json
// @Param        movieId  path      int                   true  "Movie ID"
// @Param        rating   body      RateMovieRequest      true  "Rating information"
// @Success      200      {object}  models.UserMovieRating
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /movies/{movieId}/rate [post]
func RateMovie(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	movieIDStr := c.Param("id")
	movieID, err := strconv.ParseUint(movieIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	var req RateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate rating
	if req.Rating < 0 || req.Rating > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 0 and 10"})
		return
	}

	// Check if movie exists
	var movie models.Movie
	if err := storage.DB.First(&movie, movieID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	// Create or update rating
	rating := models.UserMovieRating{
		UserID:  userID.(uint),
		MovieID: uint(movieID),
		Rating:  req.Rating,
	}

	result := storage.DB.Where("user_id = ? AND movie_id = ?", userID.(uint), movieID).
		Assign(rating).
		FirstOrCreate(&rating)

	if result.Error != nil {
		logger.Log.Error("failed to save rating", zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save rating"})
		return
	}

	// Track interaction
	interaction := models.UserInteraction{
		UserID:          userID.(uint),
		MovieID:         uint(movieID),
		InteractionType: "rate",
		Rating:          &req.Rating,
	}
	storage.DB.Create(&interaction)

	logger.Log.Info("movie rated",
		zap.Uint("user_id", userID.(uint)),
		zap.Uint("movie_id", uint(movieID)),
		zap.Float32("rating", req.Rating))

	c.JSON(http.StatusOK, rating)
}

// GetUserRatings returns all ratings by the current user
// @Summary      Get user's ratings
// @Description  Get all movies rated by the current user
// @Tags         recommendations
// @Accept       json
// @Produce      json
// @Success      200  {array}   UserRatingResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /ratings [get]
func GetUserRatings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var ratings []models.UserMovieRating
	err := storage.DB.Where("user_id = ?", userID.(uint)).
		Order("updated_at DESC").
		Find(&ratings).Error

	if err != nil {
		logger.Log.Error("failed to get user ratings", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get ratings"})
		return
	}

	// Get movie details for each rating
	var movieIDs []uint
	for _, r := range ratings {
		movieIDs = append(movieIDs, r.MovieID)
	}

	var movies []models.Movie
	storage.DB.Where("id IN ?", movieIDs).Find(&movies)

	// Create movie map for quick lookup
	movieMap := make(map[uint]models.Movie)
	for _, m := range movies {
		movieMap[m.ID] = m
	}

	// Build response
	var response []UserRatingResponse
	for _, r := range ratings {
		if movie, exists := movieMap[r.MovieID]; exists {
			response = append(response, UserRatingResponse{
				Movie:     movie,
				Rating:    r.Rating,
				UpdatedAt: r.UpdatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ratings": response,
		"count":   len(response),
	})
}

// TrackInteraction tracks user interactions with movies
// @Summary      Track user interaction
// @Description  Track when a user views or watches a movie
// @Tags         recommendations
// @Accept       json
// @Produce      json
// @Param        interaction  body      TrackInteractionRequest  true  "Interaction details"
// @Success      201          {object}  models.UserInteraction
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Security     BearerAuth
// @Router       /interactions [post]
func TrackInteraction(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var req TrackInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate interaction type
	validTypes := map[string]bool{"view": true, "watch": true, "like": true}
	if !validTypes[req.InteractionType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid interaction type"})
		return
	}

	interaction := models.UserInteraction{
		UserID:          userID.(uint),
		MovieID:         req.MovieID,
		InteractionType: req.InteractionType,
		WatchDuration:   req.WatchDuration,
	}

	if err := storage.DB.Create(&interaction).Error; err != nil {
		logger.Log.Error("failed to track interaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track interaction"})
		return
	}

	logger.Log.Info("interaction tracked",
		zap.Uint("user_id", userID.(uint)),
		zap.Uint("movie_id", req.MovieID),
		zap.String("type", req.InteractionType))

	c.JSON(http.StatusCreated, interaction)
}

// GetSimilarMovies returns movies similar to a given movie
// @Summary      Get similar movies
// @Description  Get movies similar to a specific movie based on genre and director
// @Tags         recommendations
// @Accept       json
// @Produce      json
// @Param        movieId  path      int  true   "Movie ID"
// @Param        limit    query     int  false  "Number of similar movies (default 10)"
// @Success      200      {array}   models.MovieRecommendation
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /movies/{movieId}/similar [get]
func GetSimilarMovies(c *gin.Context) {
	movieIDStr := c.Param("id")
	movieID, err := strconv.ParseUint(movieIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get the target movie
	var movie models.Movie
	if err := storage.DB.First(&movie, movieID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	// Find similar movies
	var allMovies []models.Movie
	storage.DB.Where("id != ?", movieID).Limit(limit * 5).Find(&allMovies)

	var recommendations []models.MovieRecommendation
	for _, m := range allMovies {
		score := 0.0
		var matchedGenres []string

		// Genre matching
		for _, genre := range m.Genres {
			for _, targetGenre := range movie.Genres {
				if genre == targetGenre {
					score += 1.0
					matchedGenres = append(matchedGenres, genre)
					break
				}
			}
		}

		// Director matching
		if m.Director == movie.Director && movie.Director != "" {
			score += 2.0
		}

		// Rating similarity
		ratingDiff := math.Abs(float64(m.Rating - movie.Rating))
		if ratingDiff < 1.0 {
			score += 0.5
		}

		if score > 0 {
			reason := "Similar movie"
			if len(matchedGenres) > 0 {
				reason = "Similar genre: " + matchedGenres[0]
			}
			if m.Director == movie.Director && movie.Director != "" {
				reason = "Same director: " + m.Director
			}

			recommendations = append(recommendations, models.MovieRecommendation{
				Movie:      m,
				Score:      score,
				ReasonType: "content-based",
				Reason:     reason,
			})
		}
	}

	// Sort by score
	recommendations = engine.deduplicateAndSort(recommendations)

	// Limit results
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"similar_to":      movie.Title,
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// Request/Response types
type RateMovieRequest struct {
	Rating float32 `json:"rating" binding:"required" minimum:"0" maximum:"10" example:"8.5"`
}

type TrackInteractionRequest struct {
	MovieID         uint   `json:"movie_id" binding:"required" example:"1"`
	InteractionType string `json:"interaction_type" binding:"required" example:"view"`
	WatchDuration   *int   `json:"watch_duration,omitempty" example:"3600"`
}

type UserRatingResponse struct {
	Movie     models.Movie `json:"movie"`
	Rating    float32      `json:"rating"`
	UpdatedAt interface{}  `json:"updated_at"`
}
