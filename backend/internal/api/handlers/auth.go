package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"intelligent-presenter-backend/internal/auth"
	"intelligent-presenter-backend/internal/models"
	"intelligent-presenter-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type StateStore struct {
	states map[string]time.Time
	mutex  sync.RWMutex
}

func NewStateStore() *StateStore {
	store := &StateStore{
		states: make(map[string]time.Time),
	}
	
	// Cleanup expired states in background goroutine
	go store.cleanup()
	
	return store
}

func (s *StateStore) Set(state string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.states[state] = time.Now().Add(10 * time.Minute) // 10 minutes expiration
}

func (s *StateStore) Validate(state string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	expiry, exists := s.states[state]
	return exists && time.Now().Before(expiry)
}

func (s *StateStore) Delete(state string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.states, state)
}

func (s *StateStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.mutex.Lock()
			now := time.Now()
			for state, expiry := range s.states {
				if now.After(expiry) {
					delete(s.states, state)
				}
			}
			s.mutex.Unlock()
		}
	}
}

type AuthHandler struct {
	config      *config.Config
	oauthConfig *oauth2.Config
	stateStore  *StateStore
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		config: cfg,
		stateStore: NewStateStore(),
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.BacklogClientID,
			ClientSecret: cfg.BacklogClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://%s/OAuth2AccessRequest.action", cfg.BacklogDomain),
				TokenURL: fmt.Sprintf("https://%s/api/v2/oauth2/token", cfg.BacklogDomain),
			},
			RedirectURL: cfg.OAuthRedirectURL,
			Scopes:      []string{},
		},
	}
}

func (h *AuthHandler) InitiateOAuth(c *gin.Context) {
	fmt.Printf("=== InitiateOAuth called ===\n")
	state := h.generateJWTState()
	
	// Debug logging
	fmt.Printf("Generated JWT state: %s\n", state)
	
	authURL := h.oauthConfig.AuthCodeURL(state)
	
	c.JSON(http.StatusOK, gin.H{
		"authUrl": authURL,
		"state":   state,
	})
}

func (h *AuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	
	// Debug logging
	fmt.Printf("Received callback - code: %s, state: %s\n", code, state)
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Authorization code not provided",
		})
		return
	}
	
	// Validate state parameter using stateless JWT-based approach
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "State parameter is required",
		})
		return
	}
	
	// Validate JWT state token
	if !h.validateJWTState(state) {
		fmt.Printf("JWT state validation failed for state: %s\n", state)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid or expired state parameter",
		})
		return
	}
	
	fmt.Printf("JWT state validation successful for state: %s\n", state)
	
	// Exchange code for token
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to exchange code for token",
		})
		return
	}
	
	// Get user information
	userInfo, err := h.getUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}
	
	// Generate JWT token
	jwtToken, err := auth.GenerateToken(userInfo.ID, token.AccessToken, h.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate JWT token",
		})
		return
	}
	
    // Redirect to frontend callback page with authentication info
    frontendBase := h.config.FrontendBaseURL
    if frontendBase == "" {
        frontendBase = "http://localhost:3003"
    }
    frontendCallbackURL := fmt.Sprintf("%s/auth/callback?token=%s&success=true", strings.TrimRight(frontendBase, "/"), jwtToken)
    c.Redirect(http.StatusFound, frontendCallbackURL)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Implementation for token refresh
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Token refresh not implemented yet",
	})
}

func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	
	backlogToken, exists := c.Get("backlogToken")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Backlog token not found",
		})
		return
	}
	
	userInfo, err := h.getUserInfo(backlogToken.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}
	
	c.JSON(http.StatusOK, userInfo)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// In a real application, you might want to blacklist the JWT token
	// or revoke refresh tokens
	
	// Clear any possible session cookies
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

func (h *AuthHandler) getUserInfo(accessToken string) (*models.UserInfo, error) {
	url := fmt.Sprintf("https://%s/api/v2/users/myself", h.config.BacklogDomain)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}
	
	var userInfo models.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	
	return &userInfo, nil
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// JWT-based state generation (stateless, survives container restarts)
func (h *AuthHandler) generateJWTState() string {
	fmt.Printf("JWT Secret length: %d\n", len(h.config.JWTSecret))
	
	// Create claims for the state token
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(10 * time.Minute).Unix(), // 10 minutes expiration
		"iss": "intelligent-presenter",
		"purpose": "oauth-state",
	}
	
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign token with JWT secret
	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		fmt.Printf("JWT signing failed: %v\n", err)
		// Fallback to random state if JWT fails
		return generateRandomState()
	}
	
	fmt.Printf("Generated JWT token length: %d\n", len(tokenString))
	return tokenString
}

// Validate JWT-based state token
func (h *AuthHandler) validateJWTState(stateToken string) bool {
	// Parse and validate the JWT
	token, err := jwt.Parse(stateToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.config.JWTSecret), nil
	})
	
	if err != nil {
		return false
	}
	
	// Check if token is valid and contains expected claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verify purpose claim
		if purpose, ok := claims["purpose"].(string); ok && purpose == "oauth-state" {
			// Verify issuer
			if iss, ok := claims["iss"].(string); ok && iss == "intelligent-presenter" {
				return true
			}
		}
	}
	
	return false
}