// Package api provides HTTP route configuration for the intelligent presenter backend.
// It sets up all API endpoints, middleware, and handler mappings for the application,
// including authentication routes, project data routes, slide generation, and WebSocket endpoints.
package api

import (
	"intelligent-presenter-backend/internal/api/handlers"
	"intelligent-presenter-backend/internal/auth"
	"intelligent-presenter-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all HTTP routes and WebSocket endpoints for the application.
// It initializes handlers, sets up middleware, and organizes routes into logical groups
// with appropriate authentication requirements.
//
// Route organization:
//   - /api/v1/auth/* - Authentication and OAuth flow
//   - /api/v1/projects/* - Project data from Backlog (authenticated)
//   - /api/v1/slides/* - Slide generation endpoints (authenticated)
//   - /api/v1/speech/* - Speech synthesis endpoints (authenticated)
//   - /ws/slides/* - WebSocket endpoint for real-time slide delivery
//   - /cache/* - Static audio file serving
//
// Parameters:
//   - router: the Gin engine instance to configure
//   - cfg: application configuration containing service URLs and credentials
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	slideHandler := handlers.NewSlideHandler(cfg)
	mcpHandler := handlers.NewMCPHandler(cfg)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		authGroup := v1.Group("/auth")
		{
			authGroup.GET("/login", authHandler.InitiateOAuth)
			authGroup.GET("/callback", authHandler.HandleCallback)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.GET("/me", auth.RequireAuth(cfg), authHandler.GetUserInfo)
			authGroup.POST("/logout", authHandler.Logout)
		}

		// Project data routes (requires authentication)
		projectGroup := v1.Group("/projects", auth.RequireAuth(cfg))
		{
			projectGroup.GET("", mcpHandler.GetProjects)
			projectGroup.GET("/:projectId/overview", mcpHandler.GetProjectOverview)
			projectGroup.GET("/:projectId/progress", mcpHandler.GetProjectProgress)
			projectGroup.GET("/:projectId/issues", mcpHandler.GetProjectIssues)
			projectGroup.GET("/:projectId/team", mcpHandler.GetProjectTeam)
			projectGroup.GET("/:projectId/risks", mcpHandler.GetProjectRisks)
		}

		// Slide generation routes (requires authentication)
		slideGroup := v1.Group("/slides", auth.RequireAuth(cfg))
		{
			slideGroup.POST("/generate", slideHandler.GenerateSlides)
			slideGroup.GET("/:slideId/status", slideHandler.GetSlideStatus)
		}

		// Speech synthesis routes (requires authentication)
		speechGroup := v1.Group("/speech", auth.RequireAuth(cfg))
		{
			speechGroup.POST("/synthesize", mcpHandler.SynthesizeSpeech)
			speechGroup.GET("/audio/:filename", mcpHandler.GetAudioFile)
		}
	}

	// Audio cache routes (no authentication required for cached audio files)
	router.GET("/cache/:filename", mcpHandler.GetAudioFile)

	// WebSocket endpoint for real-time slide delivery
	router.GET("/ws/slides/:slideId", auth.RequireAuthWS(cfg), slideHandler.HandleWebSocket)
}