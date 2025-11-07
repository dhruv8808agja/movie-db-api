package auth

import (
	"net/http"
	"strings"

	"github.com/dhruv8808agja/movie-db-api/internal/storage"
	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"password"`
}

// LoginResponse represents the JWT token response
type LoginResponse struct {
	Token string             `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  models.UserProfile `json:"user"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}

// Login authenticates a user and returns a JWT token
// @Summary      User login
// @Description  Authenticate with username and password to receive a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest   true  "Login credentials"
// @Success      200          {object}  LoginResponse  "Successfully authenticated"
// @Failure      400          {object}  ErrorResponse  "Invalid request body"
// @Failure      401          {object}  ErrorResponse  "Invalid credentials"
// @Failure      500          {object}  ErrorResponse  "Failed to generate token"
// @Router       /login [post]
func Login(c *gin.Context) {
	var creds LoginRequest

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize username
	username := strings.TrimSpace(strings.ToLower(creds.Username))

	// Find user by username
	var user models.User
	if err := storage.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account is inactive"})
		return
	}

	// Verify password
	if !CheckPassword(creds.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate JWT token with user claims
	token, err := GenerateTokenWithClaims(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user.ToProfile(),
	})
}
