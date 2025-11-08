package recommendations

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/internal/testutil"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetRecommendationsForUser_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	// Create test movies
	movie := &models.Movie{
		Title:       "Test Movie",
		Director:    "Director A",
		Genres:      models.Genres{"Action"},
		Rating:      8.0,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to set user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.GET("/recommendations", GetRecommendationsForUser)

	req := httptest.NewRequest("GET", "/recommendations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["recommendations"])
	assert.NotNil(t, response["count"])
}

func TestGetRecommendationsForUser_WithLimit(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.GET("/recommendations", GetRecommendationsForUser)

	req := httptest.NewRequest("GET", "/recommendations?limit=5", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetRecommendationsForUser_NoAuth(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/recommendations", GetRecommendationsForUser)

	req := httptest.NewRequest("GET", "/recommendations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRateMovie_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	movie := &models.Movie{
		Title:       "Test Movie",
		Director:    "Director A",
		Genres:      models.Genres{"Action"},
		Rating:      8.0,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.POST("/movies/:movieId/rate", RateMovie)

	reqBody := RateMovieRequest{Rating: 8.5}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/movies/1/rate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify rating was saved
	var rating models.UserMovieRating
	storage.DB.Where("user_id = ? AND movie_id = ?", user.ID, movie.ID).First(&rating)
	assert.Equal(t, float32(8.5), rating.Rating)
}

func TestRateMovie_UpdateExistingRating(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	movie := &models.Movie{
		Title:       "Test Movie",
		Director:    "Director A",
		Genres:      models.Genres{"Action"},
		Rating:      8.0,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie)

	// Create initial rating
	initialRating := &models.UserMovieRating{
		UserID:  user.ID,
		MovieID: movie.ID,
		Rating:  7.0,
	}
	storage.DB.Create(initialRating)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.POST("/movies/:movieId/rate", RateMovie)

	reqBody := RateMovieRequest{Rating: 9.0}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/movies/1/rate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify rating was updated
	var rating models.UserMovieRating
	storage.DB.Where("user_id = ? AND movie_id = ?", user.ID, movie.ID).First(&rating)
	assert.Equal(t, float32(9.0), rating.Rating)
}

func TestRateMovie_InvalidRating(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.POST("/movies/:movieId/rate", RateMovie)

	testCases := []struct {
		name   string
		rating float32
	}{
		{"Negative rating", -1.0},
		{"Too high rating", 11.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := RateMovieRequest{Rating: tc.rating}
			jsonData, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/movies/1/rate", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestRateMovie_MovieNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.POST("/movies/:movieId/rate", RateMovie)

	reqBody := RateMovieRequest{Rating: 8.5}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/movies/999/rate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserRatings_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	// Create movies and ratings
	movie1 := &models.Movie{
		Title:       "Movie 1",
		Director:    "Director A",
		Genres:      models.Genres{"Action"},
		Rating:      8.0,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie1)

	rating1 := &models.UserMovieRating{
		UserID:  user.ID,
		MovieID: movie1.ID,
		Rating:  8.5,
	}
	storage.DB.Create(rating1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.GET("/ratings", GetUserRatings)

	req := httptest.NewRequest("GET", "/ratings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["ratings"])
}

func TestTrackInteraction_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	movie := &models.Movie{
		Title:       "Test Movie",
		Director:    "Director A",
		Genres:      models.Genres{"Action"},
		Rating:      8.0,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.POST("/interactions", TrackInteraction)

	watchDuration := 3600
	reqBody := TrackInteractionRequest{
		MovieID:         movie.ID,
		InteractionType: "watch",
		WatchDuration:   &watchDuration,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/interactions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify interaction was saved
	var interaction models.UserInteraction
	storage.DB.Where("user_id = ? AND movie_id = ?", user.ID, movie.ID).First(&interaction)
	assert.Equal(t, "watch", interaction.InteractionType)
	assert.Equal(t, 3600, *interaction.WatchDuration)
}

func TestTrackInteraction_InvalidType(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	user := testutil.CreateTestUser("testuser", "password", models.RoleUser, true)
	storage.DB.Create(user)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})

	router.POST("/interactions", TrackInteraction)

	reqBody := TrackInteractionRequest{
		MovieID:         1,
		InteractionType: "invalid_type",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/interactions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSimilarMovies_Success(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	// Create movies
	movie1 := &models.Movie{
		Title:       "Action Movie",
		Director:    "Director A",
		Genres:      models.Genres{"Action", "Thriller"},
		Rating:      8.5,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie1)

	movie2 := &models.Movie{
		Title:       "Similar Action Movie",
		Director:    "Director A",
		Genres:      models.Genres{"Action"},
		Rating:      8.0,
		ReleaseDate: time.Now(),
	}
	storage.DB.Create(movie2)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/movies/:movieId/similar", GetSimilarMovies)

	req := httptest.NewRequest("GET", "/movies/1/similar", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["recommendations"])
}

func TestGetSimilarMovies_MovieNotFound(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	testutil.InitTestStorage()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/movies/:movieId/similar", GetSimilarMovies)

	req := httptest.NewRequest("GET", "/movies/999/similar", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSimilarMovies_InvalidMovieID(t *testing.T) {
	testutil.SetTestEnv()
	defer testutil.ClearTestEnv()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/movies/:movieId/similar", GetSimilarMovies)

	req := httptest.NewRequest("GET", "/movies/invalid/similar", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
