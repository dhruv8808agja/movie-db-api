package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/dhruv8808agja/movie-db-api/docs" // Import generated docs
	"github.com/dhruv8808agja/movie-db-api/internal/auth"
	"github.com/dhruv8808agja/movie-db-api/internal/logger"
	"github.com/dhruv8808agja/movie-db-api/internal/middleware"
	"github.com/dhruv8808agja/movie-db-api/internal/monitor"
	"github.com/dhruv8808agja/movie-db-api/internal/movies"
	"github.com/dhruv8808agja/movie-db-api/internal/storage"
)

// @title           Movie Database API
// @version         1.0
// @description     A movie database API with authentication, rate limiting, and caching
// @description     This API allows you to manage movies with CRUD operations
//
// @contact.name    API Support
// @contact.email   support@moviedb.com
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and JWT token
//
// @tag.name         auth
// @tag.description  Authentication endpoints
//
// @tag.name         movies
// @tag.description  Movie management endpoints
func main() {
	// Initialize logger
	logger.InitLogger()

	// Initialize DB and Redis
	storage.InitDB()
	storage.InitRedis()

	// Setup Gin router
	r := gin.New()
	r.Use(logger.GinLogger(), gin.Recovery()) // logging + recovery

	// Setup rate limiting
	rateLimitConfig := middleware.GetRateLimiterConfig()
	r.Use(middleware.RateLimiter(rateLimitConfig))

	// Public routes
	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Authentication
	r.POST("/login", auth.Login)

	// Read
	r.GET("/movies", movies.ListMoviesWithFilter)

	// Secured routes
	secured := r.Group("/")
	secured.Use(auth.JWTMiddleware())
	// Movie routes
	// Create
	secured.POST("/movies", movies.CreateMovie)
	secured.POST("/movies/bulk", movies.CreateMovies)

	// Read
	secured.GET("/movies/:id", movies.GetMovieByID)

	// Update
	secured.PUT("/movies/:id", movies.UpdateMovie)
	// Delete
	secured.DELETE("/movies/:id", movies.DeleteMovie)
	secured.DELETE("/movies", movies.DeleteMovies)

	// Prometheus metrics
	monitor.RegisterMetrics(r)

	r.Run(":8080")
}
