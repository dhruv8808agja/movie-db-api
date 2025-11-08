package recommendations

import (
	"math"
	"sort"

	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"go.uber.org/zap"
)

// Engine provides recommendation functionality
type Engine struct{}

// NewEngine creates a new recommendation engine
func NewEngine() *Engine {
	return &Engine{}
}

// GetRecommendations returns personalized movie recommendations for a user
func (e *Engine) GetRecommendations(userID uint, limit int) ([]models.MovieRecommendation, error) {
	if limit <= 0 {
		limit = 10
	}

	recommendations := make([]models.MovieRecommendation, 0)

	// Strategy 1: Content-based filtering (genre and director similarity)
	contentBased, err := e.getContentBasedRecommendations(userID, limit)
	if err != nil {
		logger.Log.Error("failed to get content-based recommendations", zap.Error(err))
	} else {
		recommendations = append(recommendations, contentBased...)
	}

	// Strategy 2: Collaborative filtering (similar users)
	collaborative, err := e.getCollaborativeRecommendations(userID, limit)
	if err != nil {
		logger.Log.Error("failed to get collaborative recommendations", zap.Error(err))
	} else {
		recommendations = append(recommendations, collaborative...)
	}

	// Strategy 3: Trending movies (popular among all users)
	if len(recommendations) < limit {
		trending, err := e.getTrendingMovies(limit)
		if err != nil {
			logger.Log.Error("failed to get trending movies", zap.Error(err))
		} else {
			recommendations = append(recommendations, trending...)
		}
	}

	// Remove duplicates and sort by score
	recommendations = e.deduplicateAndSort(recommendations)

	// Limit results
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	logger.Log.Info("generated recommendations",
		zap.Uint("user_id", userID),
		zap.Int("count", len(recommendations)))

	return recommendations, nil
}

// getContentBasedRecommendations finds movies similar to what user has liked
func (e *Engine) getContentBasedRecommendations(userID uint, limit int) ([]models.MovieRecommendation, error) {
	// Get user's highly rated movies
	var ratings []models.UserMovieRating
	err := storage.DB.Where("user_id = ? AND rating >= ?", userID, 7.0).
		Order("rating DESC").
		Limit(10).
		Find(&ratings).Error
	if err != nil {
		return nil, err
	}

	if len(ratings) == 0 {
		return []models.MovieRecommendation{}, nil
	}

	// Get the movies user liked
	var likedMovieIDs []uint
	for _, r := range ratings {
		likedMovieIDs = append(likedMovieIDs, r.MovieID)
	}

	var likedMovies []models.Movie
	if err := storage.DB.Where("id IN ?", likedMovieIDs).Find(&likedMovies).Error; err != nil {
		return nil, err
	}

	// Extract user's favorite genres and directors
	genreCount := make(map[string]int)
	directorCount := make(map[string]int)

	for _, movie := range likedMovies {
		for _, genre := range movie.Genres {
			genreCount[genre]++
		}
		directorCount[movie.Director]++
	}

	// Find similar movies (same genre or director, not already rated)
	var similarMovies []models.Movie
	var ratedMovieIDs []uint

	// Get all movies user has rated
	var allRatings []models.UserMovieRating
	storage.DB.Where("user_id = ?", userID).Find(&allRatings)
	for _, r := range allRatings {
		ratedMovieIDs = append(ratedMovieIDs, r.MovieID)
	}

	query := storage.DB.Limit(limit * 2)
	if len(ratedMovieIDs) > 0 {
		query = query.Where("id NOT IN ?", ratedMovieIDs)
	}

	query.Find(&similarMovies)

	// Score movies based on similarity
	var recommendations []models.MovieRecommendation
	for _, movie := range similarMovies {
		score := 0.0
		var reasons []string

		// Genre similarity
		for _, genre := range movie.Genres {
			if count, exists := genreCount[genre]; exists {
				score += float64(count) * 0.3
				reasons = append(reasons, genre)
			}
		}

		// Director similarity
		if count, exists := directorCount[movie.Director]; exists {
			score += float64(count) * 0.5
			reasons = append(reasons, movie.Director)
		}

		// Rating boost
		score += float64(movie.Rating) * 0.2

		if score > 0 {
			reason := "Similar to movies you liked"
			if len(reasons) > 0 {
				reason = "Based on your interest in " + reasons[0]
			}

			recommendations = append(recommendations, models.MovieRecommendation{
				Movie:      movie,
				Score:      score,
				ReasonType: "content-based",
				Reason:     reason,
			})
		}
	}

	return recommendations, nil
}

// getCollaborativeRecommendations finds movies liked by similar users
func (e *Engine) getCollaborativeRecommendations(userID uint, limit int) ([]models.MovieRecommendation, error) {
	// Get current user's ratings
	var userRatings []models.UserMovieRating
	err := storage.DB.Where("user_id = ?", userID).Find(&userRatings).Error
	if err != nil {
		return nil, err
	}

	if len(userRatings) < 3 {
		// Not enough data for collaborative filtering
		return []models.MovieRecommendation{}, nil
	}

	// Create user rating map
	userRatingMap := make(map[uint]float32)
	for _, r := range userRatings {
		userRatingMap[r.MovieID] = r.Rating
	}

	// Find similar users (users who rated same movies similarly)
	var allRatings []models.UserMovieRating
	var movieIDs []uint
	for movieID := range userRatingMap {
		movieIDs = append(movieIDs, movieID)
	}

	storage.DB.Where("movie_id IN ? AND user_id != ?", movieIDs, userID).
		Find(&allRatings)

	// Calculate similarity with other users (cosine similarity)
	otherUserRatings := make(map[uint]map[uint]float32)
	for _, r := range allRatings {
		if _, exists := otherUserRatings[r.UserID]; !exists {
			otherUserRatings[r.UserID] = make(map[uint]float32)
		}
		otherUserRatings[r.UserID][r.MovieID] = r.Rating
	}

	// Find most similar users
	type userSimilarity struct {
		userID     uint
		similarity float64
	}
	var similarities []userSimilarity

	for otherUserID, otherRatings := range otherUserRatings {
		similarity := calculateCosineSimilarity(userRatingMap, otherRatings)
		if similarity > 0.3 { // Threshold for similarity
			similarities = append(similarities, userSimilarity{otherUserID, similarity})
		}
	}

	// Sort by similarity
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].similarity > similarities[j].similarity
	})

	// Get top 5 similar users
	if len(similarities) > 5 {
		similarities = similarities[:5]
	}

	if len(similarities) == 0 {
		return []models.MovieRecommendation{}, nil
	}

	// Get movies highly rated by similar users that current user hasn't rated
	var similarUserIDs []uint
	for _, s := range similarities {
		similarUserIDs = append(similarUserIDs, s.userID)
	}

	var recommendedRatings []models.UserMovieRating
	var ratedMovieIDs []uint
	for movieID := range userRatingMap {
		ratedMovieIDs = append(ratedMovieIDs, movieID)
	}

	query := storage.DB.Where("user_id IN ? AND rating >= ?", similarUserIDs, 7.0).
		Limit(limit * 2)

	if len(ratedMovieIDs) > 0 {
		query = query.Where("movie_id NOT IN ?", ratedMovieIDs)
	}

	query.Find(&recommendedRatings)

	// Aggregate scores for recommended movies
	movieScores := make(map[uint]float64)
	for _, r := range recommendedRatings {
		// Weight by user similarity
		for _, s := range similarities {
			if s.userID == r.UserID {
				movieScores[r.MovieID] += float64(r.Rating) * s.similarity
				break
			}
		}
	}

	// Convert to recommendations
	var recommendations []models.MovieRecommendation
	for movieID, score := range movieScores {
		var movie models.Movie
		if err := storage.DB.First(&movie, movieID).Error; err != nil {
			continue
		}

		recommendations = append(recommendations, models.MovieRecommendation{
			Movie:      movie,
			Score:      score,
			ReasonType: "collaborative",
			Reason:     "Users with similar taste enjoyed this",
		})
	}

	return recommendations, nil
}

// getTrendingMovies returns popular movies based on recent interactions
func (e *Engine) getTrendingMovies(limit int) ([]models.MovieRecommendation, error) {
	// Get most interacted movies in last 30 days
	var interactions []struct {
		MovieID uint
		Count   int64
	}

	err := storage.DB.Model(&models.UserInteraction{}).
		Select("movie_id, COUNT(*) as count").
		Where("created_at > datetime('now', '-30 days')").
		Group("movie_id").
		Order("count DESC").
		Limit(limit).
		Scan(&interactions).Error

	if err != nil {
		return nil, err
	}

	var recommendations []models.MovieRecommendation
	for _, interaction := range interactions {
		var movie models.Movie
		if err := storage.DB.First(&movie, interaction.MovieID).Error; err != nil {
			continue
		}

		recommendations = append(recommendations, models.MovieRecommendation{
			Movie:      movie,
			Score:      float64(interaction.Count),
			ReasonType: "trending",
			Reason:     "Popular this month",
		})
	}

	// If no interactions, return highest rated movies
	if len(recommendations) == 0 {
		var movies []models.Movie
		storage.DB.Order("rating DESC").Limit(limit).Find(&movies)

		for _, movie := range movies {
			recommendations = append(recommendations, models.MovieRecommendation{
				Movie:      movie,
				Score:      float64(movie.Rating),
				ReasonType: "trending",
				Reason:     "Highly rated",
			})
		}
	}

	return recommendations, nil
}

// deduplicateAndSort removes duplicate movies and sorts by score
func (e *Engine) deduplicateAndSort(recommendations []models.MovieRecommendation) []models.MovieRecommendation {
	seen := make(map[uint]bool)
	var unique []models.MovieRecommendation

	for _, rec := range recommendations {
		if !seen[rec.Movie.ID] {
			seen[rec.Movie.ID] = true
			unique = append(unique, rec)
		}
	}

	// Sort by score descending
	sort.Slice(unique, func(i, j int) bool {
		return unique[i].Score > unique[j].Score
	})

	return unique
}

// calculateCosineSimilarity calculates cosine similarity between two rating vectors
func calculateCosineSimilarity(ratings1, ratings2 map[uint]float32) float64 {
	var dotProduct, magnitude1, magnitude2 float64

	for movieID, rating1 := range ratings1 {
		if rating2, exists := ratings2[movieID]; exists {
			dotProduct += float64(rating1 * rating2)
		}
		magnitude1 += float64(rating1 * rating1)
	}

	for _, rating2 := range ratings2 {
		magnitude2 += float64(rating2 * rating2)
	}

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(magnitude1) * math.Sqrt(magnitude2))
}
