package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"password"`
}

// LoginResponse represents the JWT token response
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
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
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hardcoded credentials (replace with DB lookup in real app)
	if creds.Username != "admin" || creds.Password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := GenerateToken(creds.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
