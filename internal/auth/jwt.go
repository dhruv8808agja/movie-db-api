package auth

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dhruv8808agja/movie-db-api/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = getJWTSecret()

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	return []byte(secret)
}

// GenerateToken creates a JWT for a given username
func GenerateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 1 day expiry
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateTokenWithClaims creates a JWT with user ID, username, and role
func GenerateTokenWithClaims(userID uint, username string, role models.UserRole) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 1 day expiry
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// JWTMiddleware checks for valid JWT token
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenMalformed
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// token is valid, save claims in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("username", claims["username"])
			// Set user_id and role if present (for enhanced tokens)
			if userID, ok := claims["user_id"].(float64); ok {
				c.Set("user_id", uint(userID))
			}
			if role, ok := claims["role"].(string); ok {
				c.Set("user_role", models.UserRole(role))
			}
		}

		c.Next()
	}
}
