// Package config provides configuration management for the intelligent presenter backend.
// It loads configuration values from environment variables with sensible defaults
// for development and production deployments.
package config

import (
	"os"
	"strings"
)

// Config holds all configuration values for the intelligent presenter backend.
// It includes settings for server operation, external service integrations,
// authentication, and security configurations.
type Config struct {
	// Port specifies the HTTP server port number
	Port string
	// Environment indicates the deployment environment (debug, release, production)
	Environment string
	
	// Backlog OAuth configuration for integrating with Backlog project management
	BacklogDomain       string // Backlog space domain (e.g., "yourspace.backlog.jp")
	BacklogClientID     string // OAuth2 client ID for Backlog API
	BacklogClientSecret string // OAuth2 client secret for Backlog API
	OAuthRedirectURL    string // OAuth2 callback URL for authentication flow
	
	// AI Provider configuration for slide content generation
	AIProvider   string // AI service to use: "openai" or "bedrock"
	OpenAIAPIKey string // API key for OpenAI services
	
	// AWS Bedrock configuration for AI content generation
	AWSRegion          string // AWS region for Bedrock service
	AWSAccessKeyID     string // AWS access key for authentication
	AWSSecretAccessKey string // AWS secret key for authentication
	BedrockModelID     string // Bedrock model identifier for content generation
	
	// MCP Server URLs for Model Context Protocol integration
	MCPBacklogURL string // URL of the Backlog MCP server
	MCPSpeechURL  string // URL of the Speech MCP server
	
	// JWT configuration for session management
	JWTSecret string // Secret key for JWT token signing and verification

    // Frontend base URL for OAuth redirects and CORS
    FrontendBaseURL string // Base URL of the frontend application

    // CORS configuration for cross-origin request handling
    CORSOrigins []string // List of allowed origins for CORS requests
}

// Load creates a new Config instance by reading environment variables.
// It provides sensible defaults for development environments and ensures
// all required configuration values are properly initialized.
//
// Returns a fully configured Config struct with all fields populated
// from environment variables or their default values.
func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		Environment:         getEnv("GIN_MODE", "debug"),
		BacklogDomain:       getEnv("BACKLOG_DOMAIN", ""),
		BacklogClientID:     getEnv("BACKLOG_CLIENT_ID", ""),
		BacklogClientSecret: getEnv("BACKLOG_CLIENT_SECRET", ""),
        OAuthRedirectURL:    getEnv("OAUTH_REDIRECT_URL", "http://localhost:8081/api/v1/auth/callback"),
		AIProvider:          getEnv("AI_PROVIDER", "openai"),
		OpenAIAPIKey:        getEnv("OPENAI_API_KEY", ""),
		AWSRegion:           getEnv("AWS_REGION", "ap-northeast-1"),
		AWSAccessKeyID:      getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),
		BedrockModelID:      getEnv("BEDROCK_MODEL_ID", "anthropic.claude-3-haiku-20240307-v1:0"),
        MCPBacklogURL:       getEnv("MCP_BACKLOG_URL", "http://localhost:3001"),
		MCPSpeechURL:        getEnv("MCP_SPEECH_URL", "http://localhost:3002"),
		JWTSecret:           getEnv("JWT_SECRET", "intelligent-presenter-secret-key"),
        FrontendBaseURL:     getEnv("FRONTEND_BASE_URL", "http://localhost:3003"),
		CORSOrigins:         getEnvAsSlice("CORS_ORIGINS", []string{"http://localhost:3003"}),
	}
}

// getEnvAsSlice converts a comma-separated environment variable into a string slice.
// If the environment variable is empty or not set, it returns the provided default slice.
//
// Parameters:
//   - name: the environment variable name to read
//   - defaultVal: the default slice to return if the environment variable is not set
//
// Returns a slice of strings split by commas, or the default value if not found.
func getEnvAsSlice(name string, defaultVal []string) []string {
    valStr := getEnv(name, "")
    if valStr == "" {
        return defaultVal
    }
    return strings.Split(valStr, ",")
}

// getEnv retrieves an environment variable value with a fallback default.
// This is a utility function used throughout the configuration loading process.
//
// Parameters:
//   - key: the environment variable name to retrieve
//   - defaultValue: the value to return if the environment variable is not set
//
// Returns the environment variable value if set, otherwise returns the default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}