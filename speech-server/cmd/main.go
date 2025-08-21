// Package main provides the entry point for the Speech MCP Server.
// This server provides text-to-speech (TTS) capabilities through multiple engines
// including VOICEVOX, Kokoro TTS, and MLX-Audio, supporting both Japanese and
// multilingual speech synthesis with high-quality neural voice models.
//
// The server supports:
//   - Multiple TTS engines with automatic fallback
//   - Japanese high-quality synthesis via VOICEVOX and MLX-Audio
//   - Multilingual synthesis via Kokoro TTS (82M parameter model)
//   - Audio caching for improved performance
//   - MCP protocol integration for intelligent presenter
//   - RESTful API for direct TTS access
//   - Configurable voice selection and audio formats
//
// The server operates in HTTP mode providing both MCP protocol endpoints
// and direct API access for speech synthesis operations.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"speech-mcp-server/internal/handlers"
	"speech-mcp-server/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// main initializes and starts the Speech MCP Server.
// It handles environment configuration, server setup, graceful shutdown,
// and provides both MCP protocol and REST API endpoints for TTS operations.
//
// The startup process includes:
//   1. Loading environment variables and configuration
//   2. Setting up Gin web framework and CORS middleware
//   3. Registering API routes and MCP protocol handlers
//   4. Starting the HTTP server with graceful shutdown support
//
// The server listens for SIGINT and SIGTERM signals for clean shutdown.
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// CORS middleware
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.CORSOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "speech-mcp-server",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// Initialize handlers
	speechHandler := handlers.NewSpeechHandler(cfg)

	// Setup routes
	setupRoutes(router, speechHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Speech MCP Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Speech MCP Server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Speech MCP Server exited")
}

// setupRoutes configures all HTTP routes and endpoints for the Speech MCP Server.
// It organizes routes into logical groups for API versioning and MCP protocol support.
//
// Route organization:
//   - /api/v1/* - RESTful API endpoints for direct TTS access
//   - /mcp/* - MCP protocol endpoints for intelligent presenter integration
//   - /cache/* - Static file serving for cached audio files
//
// Parameters:
//   - router: the Gin engine instance to configure
//   - speechHandler: initialized speech handler with TTS capabilities
func setupRoutes(router *gin.Engine, speechHandler *handlers.SpeechHandler) {
	// MCP routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/synthesize", speechHandler.SynthesizeSpeech)
		v1.GET("/audio/:filename", speechHandler.ServeAudioFile)
		v1.GET("/voices", speechHandler.ListVoices)
		v1.GET("/languages", speechHandler.ListLanguages)
	}

	// MCP Protocol endpoints
	mcp := router.Group("/mcp")
	{
		mcp.POST("/", speechHandler.HandleMCPRequest)
		mcp.GET("/capabilities", speechHandler.GetCapabilities)
	}

	// Static file serving for audio cache
	router.Static("/cache", "./cache")
}