// Package main provides the entry point for the intelligent presenter backend server.
// This server provides APIs for slide generation, authentication, and integration
// with Backlog project management and various AI services for content creation.
//
// The server supports:
//   - OAuth authentication with Backlog
//   - Slide generation using AI providers (OpenAI or AWS Bedrock)
//   - Real-time communication via WebSockets
//   - MCP (Model Context Protocol) integration for external services
//   - Text-to-speech functionality for slide narration
//
// Environment variables are used for configuration, with .env file support
// for development environments.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"intelligent-presenter-backend/internal/api"
	"intelligent-presenter-backend/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// main initializes and starts the intelligent presenter backend server.
// It handles environment configuration, server setup, graceful shutdown,
// and provides the main HTTP API endpoints for the application.
//
// The startup process includes:
//   1. Loading environment variables from .env file or system environment
//   2. Configuring the Gin web framework and middleware
//   3. Setting up CORS for cross-origin requests
//   4. Registering API routes and handlers
//   5. Starting the HTTP server with graceful shutdown support
//
// The server listens for SIGINT and SIGTERM signals for clean shutdown.
func main() {
	// Load environment variables from .env file if available
	// Falls back to system environment variables if .env is not found
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load application configuration from environment variables
	cfg := config.Load()

	// Note: In Docker mode, MCP servers run in separate containers
	// The MCP service will be initialized when needed by handlers

	// Configure Gin framework mode based on environment
	// Production mode disables debug logging and improves performance
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize the Gin router with default middleware
	router := gin.Default()

	// Configure Cross-Origin Resource Sharing (CORS) middleware
	// Allows frontend applications to access the API from different origins
	corsConfig := cors.DefaultConfig()
    corsConfig.AllowOrigins = cfg.CORSOrigins
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Register health check endpoint for monitoring and load balancer health checks
	// Returns server status, timestamp, and version information
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// Initialize and register all API routes with their respective handlers
	api.SetupRoutes(router, cfg)

	// Create HTTP server instance with configured router
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start HTTP server in a separate goroutine to allow for graceful shutdown
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	// Listens for interrupt signals (Ctrl+C) and termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Perform graceful shutdown with a 30-second timeout
	// Allows ongoing requests to complete before forcing shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}