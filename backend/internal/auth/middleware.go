// Package auth provides authentication middleware and JWT token handling
// for the intelligent presenter backend. It includes middleware functions
// for HTTP and WebSocket authentication, token validation, and token generation.
package auth

import (
	"net/http"
	"strings"
	"time"

	"intelligent-presenter-backend/internal/models"
	"intelligent-presenter-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// RequireAuth is a middleware that requires JWT authentication for HTTP requests.
// It extracts the JWT token from the Authorization header, validates it,
// and stores user information in the Gin context for use by handlers.
//
// The middleware expects the Authorization header to be in the format:
// "Bearer <jwt_token>"
//
// If authentication fails, it returns a 401 Unauthorized response and aborts the request.
// If successful, it sets "userID" and "backlogToken" in the context.
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			c.Abort()
			return
		}

		claims, err := validateToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("userID", claims.UserID)
		c.Set("backlogToken", claims.BacklogToken)
		c.Next()
	})
}

// RequireAuthWS is a middleware for WebSocket authentication.
// Unlike HTTP authentication, WebSocket authentication receives the JWT token
// as a query parameter named "token" rather than in the Authorization header.
//
// This is necessary because WebSocket connections cannot send custom headers
// during the initial handshake in browser environments.
//
// If authentication fails, it returns a 401 Unauthorized response and aborts the request.
// If successful, it sets "userID" and "backlogToken" in the context.
func RequireAuthWS(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token required for WebSocket connection",
			})
			c.Abort()
			return
		}

		claims, err := validateToken(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set("userID", claims.UserID)
		c.Set("backlogToken", claims.BacklogToken)
		c.Next()
	})
}

// extractToken extracts JWT token from Authorization header.
// It expects the header to be in the format "Bearer <token>" and returns
// the token portion, or an empty string if the format is invalid.
//
// Parameters:
//   - c: the Gin context containing the HTTP request
//
// Returns the JWT token string, or empty string if not found or invalid format.
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// validateToken validates JWT token and returns claims.
// It parses the JWT token, verifies the signature using the provided secret,
// and returns the custom claims if the token is valid.
//
// Parameters:
//   - tokenString: the JWT token to validate
//   - secret: the secret key used to sign the token
//
// Returns the JWTClaims if valid, or an error if validation fails.
func validateToken(tokenString, secret string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// GenerateToken generates a new JWT token for authenticated users.
// It creates a JWT token containing the user ID and Backlog access token,
// with a 7-day expiration time.
//
// Parameters:
//   - userID: the Backlog user ID to include in the token
//   - backlogToken: the Backlog OAuth access token for API calls
//   - secret: the secret key used to sign the JWT token
//
// Returns the signed JWT token string, or an error if token generation fails.
func GenerateToken(userID int, backlogToken, secret string) (string, error) {
	now := time.Now()
	claims := &models.JWTClaims{
		UserID:       userID,
		BacklogToken: backlogToken,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * 7 * time.Hour)), // 7 days
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}