// Package models defines authentication-related data structures for the intelligent presenter.
// It includes OAuth state management, token handling, user information, and JWT claims
// used throughout the authentication flow with Backlog integration.
package models

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)

// OAuthState represents OAuth state information used during the authentication flow.
// It stores temporary state data to prevent CSRF attacks and manage redirect flows.
type OAuthState struct {
	State       string    `json:"state"`       // Random state value for CSRF protection
	RedirectURL string    `json:"redirectUrl"` // Frontend URL to redirect after authentication
	CreatedAt   time.Time `json:"createdAt"`   // Timestamp when state was created
}

// TokenInfo represents OAuth token information received from Backlog API.
// It contains all token-related data needed for API authentication and refresh.
type TokenInfo struct {
	AccessToken  string    `json:"accessToken"`  // OAuth2 access token for API calls
	RefreshToken string    `json:"refreshToken"` // OAuth2 refresh token for token renewal
	TokenType    string    `json:"tokenType"`    // Token type (typically "Bearer")
	ExpiresAt    time.Time `json:"expiresAt"`    // Token expiration timestamp
	Scope        string    `json:"scope"`        // OAuth2 scopes granted to the token
}

// UserInfo represents Backlog user information retrieved from the API.
// It contains user profile data and account details from Backlog.
type UserInfo struct {
	ID           int    `json:"id"`           // Backlog user ID (numeric)
	UserID       string `json:"userId"`       // Backlog user ID (string identifier)
	Name         string `json:"name"`         // User's display name
	RoleType     int    `json:"roleType"`     // User's role type in Backlog
	Lang         string `json:"lang"`         // User's preferred language setting
	MailAddress  string `json:"mailAddress"`  // User's email address
	Account struct {
		AccountID string `json:"accountId"` // Account ID
		Name      string `json:"name"`      // Account name
		UniqueID  string `json:"uniqueId"`  // Unique identifier for account
	} `json:"account"` // Nested account information
}

// AuthResponse represents the authentication response sent to the client.
// It includes the JWT token, user information, and token expiration details.
type AuthResponse struct {
	Token     string    `json:"token"`     // JWT token for authenticated requests
	UserInfo  UserInfo  `json:"userInfo"`  // User profile information
	ExpiresAt time.Time `json:"expiresAt"` // JWT token expiration timestamp
}

// JWTClaims represents JWT token claims for session management.
// It extends the standard JWT claims with application-specific data.
type JWTClaims struct {
	UserID       int    `json:"userId"`       // Backlog user ID for user identification
	BacklogToken string `json:"backlogToken"` // Backlog access token for API calls
	jwt.RegisteredClaims                      // Standard JWT claims (exp, iat, etc.)
}

// JWT Claims interface implementation methods
// These methods implement the jwt.Claims interface by delegating to RegisteredClaims

// GetExpirationTime returns the token expiration time claim.
// It delegates to the embedded RegisteredClaims for standard JWT exp claim handling.
func (c *JWTClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetExpirationTime()
}

// GetIssuedAt returns the token issued-at time claim.
// It delegates to the embedded RegisteredClaims for standard JWT iat claim handling.
func (c *JWTClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetIssuedAt()
}

// GetNotBefore returns the token not-before time claim.
// It delegates to the embedded RegisteredClaims for standard JWT nbf claim handling.
func (c *JWTClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.RegisteredClaims.GetNotBefore()
}

// GetIssuer returns the token issuer claim.
// It delegates to the embedded RegisteredClaims for standard JWT iss claim handling.
func (c *JWTClaims) GetIssuer() (string, error) {
	return c.RegisteredClaims.GetIssuer()
}

// GetSubject returns the token subject claim.
// It delegates to the embedded RegisteredClaims for standard JWT sub claim handling.
func (c *JWTClaims) GetSubject() (string, error) {
	return c.RegisteredClaims.GetSubject()
}

// GetAudience returns the token audience claim.
// It delegates to the embedded RegisteredClaims for standard JWT aud claim handling.
func (c *JWTClaims) GetAudience() (jwt.ClaimStrings, error) {
	return c.RegisteredClaims.GetAudience()
}