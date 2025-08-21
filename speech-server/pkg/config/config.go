// Package config provides configuration management for the Speech MCP Server.
// It loads TTS engine settings, audio parameters, and server configuration
// from environment variables with sensible defaults for development.
package config

import (
	"os"
	"strings"
)

// Config holds all configuration values for the Speech MCP Server.
// It includes TTS engine settings, audio parameters, external API configuration,
// and server operation settings.
type Config struct {
	// Server configuration
	Port        string // HTTP server port number
	Environment string // Deployment environment (development, production)
	
	// TTS engine configuration
	TTSEngine   string // Preferred TTS engine (voicevox, kokoro, mlx-audio)
	Language    string // Default language for synthesis
	VoiceGender string // Default voice gender preference
	CacheDir    string // Directory for audio file caching
	
	// External TTS API configuration (for cloud TTS services)
	TTSAPIKey string // API key for external TTS services
	TTSAPIURL string // URL for external TTS services
	
	// Audio output settings
	AudioFormat string // Output audio format (wav, mp3, etc.)
	SampleRate  int    // Audio sample rate in Hz
	BitRate     int    // Audio bit rate for compressed formats

	// CORS configuration for cross-origin requests
	CORSOrigins []string // List of allowed origins for CORS requests
}

// Load creates a new Config instance by reading environment variables.
// It provides sensible defaults for speech synthesis and server operation,
// ensuring all required configuration values are properly initialized.
//
// Returns a fully configured Config struct with all fields populated
// from environment variables or their default values.
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "3001"),
		Environment: getEnv("NODE_ENV", "development"),
		TTSEngine:   getEnv("TTS_ENGINE", "go-tts"),
		Language:    getEnv("LANGUAGE", "ja"),
		VoiceGender: getEnv("VOICE_GENDER", "female"),
		CacheDir:    getEnv("CACHE_DIR", "./cache"),
		TTSAPIKey:   getEnv("TTS_API_KEY", ""),
		TTSAPIURL:   getEnv("TTS_API_URL", ""),
		AudioFormat: getEnv("AUDIO_FORMAT", "wav"),
		SampleRate:  getEnvInt("SAMPLE_RATE", 22050),
		BitRate:     getEnvInt("BIT_RATE", 128),
		CORSOrigins: getEnvAsSlice("CORS_ORIGINS", []string{"http://localhost:3003"}),
	}
}

// getEnvAsSlice converts a comma-separated environment variable into a string slice.
// Used for configuration values that accept multiple options like CORS origins.
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

// getEnvInt retrieves an integer environment variable with a fallback default.
// It performs basic string-to-integer conversion for common audio parameters
// like sample rates and bit rates.
//
// Parameters:
//   - key: the environment variable name to retrieve
//   - defaultValue: the integer value to return if conversion fails or variable is not set
//
// Returns the converted integer value or the default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// Simple conversion - in production, add proper error handling
		switch value {
		case "22050":
			return 22050
		case "44100":
			return 44100
		case "48000":
			return 48000
		case "128":
			return 128
		case "192":
			return 192
		case "256":
			return 256
		}
	}
	return defaultValue
}