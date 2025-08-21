// Package models defines the data structures and types used throughout
// the intelligent presenter application. This includes slide content models,
// request/response structures, and WebSocket message types.
package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// SlideTheme represents different types of slides that can be generated
// based on various aspects of project management and analysis.
// Each theme focuses on specific data types and presentation styles.
type SlideTheme string

const (
	// ThemeProjectOverview generates slides with basic project information,
	// including project name, objectives, timeline, and team structure
	ThemeProjectOverview SlideTheme = "project_overview"
	
	// ThemeProjectProgress creates slides showing completion rates,
	// milestone achievements, and timeline progress
	ThemeProjectProgress SlideTheme = "project_progress"
	
	// ThemeIssueManagement focuses on issue tracking, resolution rates,
	// and priority distribution across the project
	ThemeIssueManagement SlideTheme = "issue_management"
	
	// ThemeRiskAnalysis presents identified risks, their impact levels,
	// and mitigation strategies
	ThemeRiskAnalysis SlideTheme = "risk_analysis"
	
	// ThemeTeamCollaboration showcases team member activities,
	// collaboration metrics, and communication patterns
	ThemeTeamCollaboration SlideTheme = "team_collaboration"
	
	// ThemeDocumentManagement covers documentation status,
	// knowledge sharing, and information accessibility
	ThemeDocumentManagement SlideTheme = "document_management"
	
	// ThemeCodebaseActivity displays development metrics,
	// commit patterns, and code quality indicators
	ThemeCodebaseActivity SlideTheme = "codebase_activity"
	
	// ThemeNotifications presents communication efficiency,
	// notification handling, and information flow
	ThemeNotifications SlideTheme = "notifications"
	
	// ThemePredictiveAnalysis shows forecasts, trend analysis,
	// and predictive insights based on historical data
	ThemePredictiveAnalysis SlideTheme = "predictive_analysis"
	
	// ThemeSummaryPlan provides project summaries, key achievements,
	// and future planning recommendations
	ThemeSummaryPlan SlideTheme = "summary_plan"
)

// ProjectID is a custom type that can handle both string and number types from JSON.
// Backlog APIs may return project IDs as either strings or numbers, so this type
// provides flexible unmarshaling to ensure compatibility with different API responses.
type ProjectID string

// UnmarshalJSON implements custom JSON unmarshaling for ProjectID.
// It accepts both string and numeric project ID formats from JSON input.
//
// Supported formats:
//   - String: "123" or "PROJECT_KEY"
//   - Number: 123
//
// Returns an error if the input cannot be converted to a valid project ID.
func (p *ProjectID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*p = ProjectID(s)
		return nil
	}
	
	// If that fails, try as number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*p = ProjectID(n.String())
		return nil
	}
	
	return fmt.Errorf("projectId must be a string or number")
}

// String returns the string representation of the ProjectID.
func (p ProjectID) String() string {
	return string(p)
}

// SlideGenerationRequest represents a client request to generate presentation slides.
// It specifies which project to analyze, what themes to include, and the target language.
type SlideGenerationRequest struct {
	ProjectID ProjectID    `json:"projectId" binding:"required"` // Backlog project identifier
	Themes    []SlideTheme `json:"themes" binding:"required"`    // List of slide themes to generate
	Language  string       `json:"language" binding:"required"`  // Target language ("ja" or "en")
}

// SlideGenerationResponse represents the server response to a slide generation request.
// It provides the session ID and WebSocket URL for real-time generation updates.
type SlideGenerationResponse struct {
	SlideID      string `json:"slideId"`      // Unique identifier for this generation session
	Status       string `json:"status"`       // Current generation status
	WebSocketURL string `json:"websocketUrl"` // WebSocket endpoint for real-time updates
}

// SlideContent represents a complete slide with both markdown source and rendered HTML.
// This structure contains all the information needed to display and manage a single slide.
type SlideContent struct {
	Index       int        `json:"index"`       // Slide position in the presentation (1-based)
	Theme       SlideTheme `json:"theme"`       // Theme that generated this slide
	Title       string     `json:"title"`       // Slide title for navigation and display
	Markdown    string     `json:"markdown"`    // Source markdown content
	HTML        string     `json:"html"`        // Rendered HTML content (LLM-generated)
	GeneratedAt time.Time  `json:"generatedAt"` // Timestamp when slide was created
}

// SlideNarration represents narration text for a slide
type SlideNarration struct {
	SlideIndex int    `json:"slideIndex"`
	Text       string `json:"text"`
	Language   string `json:"language"`
}

// SlideAudio represents audio information for a slide
type SlideAudio struct {
	SlideIndex int    `json:"slideIndex"`
	AudioURL   string `json:"audioUrl"`
	Duration   int    `json:"duration"` // in seconds
}

// SlideGenerationStarted represents the start of slide generation
type SlideGenerationStarted struct {
	SlideIndex int        `json:"slideIndex"`
	Theme      SlideTheme `json:"theme"`
}

// PresentationComplete represents completion of slide generation
type PresentationComplete struct {
	TotalSlides int    `json:"totalSlides"`
	Duration    string `json:"duration"`
}

// WebSocketMessage represents messages sent through WebSocket
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WebSocket message types
const (
	MessageTypeSlideGenerationStarted = "slide_generation_started"
	MessageTypeSlideContent           = "slide_content"
	MessageTypeSlideNarration        = "slide_narration"
	MessageTypeSlideAudio            = "slide_audio"
	MessageTypePresentationComplete   = "presentation_complete"
	MessageTypeError                 = "error"
)

// ErrorMessage represents error information
type ErrorMessage struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}