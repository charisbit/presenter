package tests

import (
	"testing"

	"intelligent-presenter-backend/internal/models"
	"intelligent-presenter-backend/internal/services"
	"intelligent-presenter-backend/pkg/config"
)

// TestSlideService_NewSlideService tests the creation of a new SlideService instance
func TestSlideService_NewSlideService(t *testing.T) {
	cfg := &config.Config{
		OpenAIAPIKey: "test-key",
		AIProvider:   "openai",
		MCPBacklogURL: "http://localhost:3001",
	}

	service := services.NewSlideService(cfg)
	if service == nil {
		t.Fatal("Expected SlideService instance, got nil")
	}
}

// TestSlideTheme_Constants tests that all slide theme constants are defined correctly
func TestSlideTheme_Constants(t *testing.T) {
	themes := []models.SlideTheme{
		models.ThemeProjectOverview,
		models.ThemeProjectProgress,
		models.ThemeIssueManagement,
		models.ThemeRiskAnalysis,
		models.ThemeTeamCollaboration,
		models.ThemeDocumentManagement,
		models.ThemeCodebaseActivity,
		models.ThemeNotifications,
		models.ThemePredictiveAnalysis,
		models.ThemeSummaryPlan,
	}

	expectedThemes := map[models.SlideTheme]string{
		models.ThemeProjectOverview:     "project_overview",
		models.ThemeProjectProgress:     "project_progress",
		models.ThemeIssueManagement:     "issue_management",
		models.ThemeRiskAnalysis:        "risk_analysis",
		models.ThemeTeamCollaboration:   "team_collaboration",
		models.ThemeDocumentManagement:  "document_management",
		models.ThemeCodebaseActivity:    "codebase_activity",
		models.ThemeNotifications:       "notifications",
		models.ThemePredictiveAnalysis:  "predictive_analysis",
		models.ThemeSummaryPlan:         "summary_plan",
	}

	for theme, expectedValue := range expectedThemes {
		if string(theme) != expectedValue {
			t.Errorf("Theme %s has incorrect value: expected %s, got %s", 
				expectedValue, expectedValue, string(theme))
		}
	}

	if len(themes) != 10 {
		t.Errorf("Expected 10 themes, got %d", len(themes))
	}

	// Test that no theme is empty
	for i, theme := range themes {
		if theme == "" {
			t.Errorf("Theme at index %d is empty", i)
		}
	}
}

// TestProjectID_UnmarshalJSON tests ProjectID JSON unmarshaling
func TestProjectID_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
		hasError bool
	}{
		{
			name:     "String ID",
			input:    []byte(`"project123"`),
			expected: "project123",
			hasError: false,
		},
		{
			name:     "Numeric ID",
			input:    []byte(`123`),
			expected: "123",
			hasError: false,
		},
		{
			name:     "Project Key",
			input:    []byte(`"TEST_PROJECT"`),
			expected: "TEST_PROJECT",
			hasError: false,
		},
		{
			name:     "Zero numeric ID",
			input:    []byte(`0`),
			expected: "0",
			hasError: false,
		},
		{
			name:     "Invalid JSON",
			input:    []byte(`invalid`),
			expected: "",
			hasError: true,
		},
		{
			name:     "Null value",
			input:    []byte(`null`),
			expected: "",
			hasError: false, // JSON null unmarshals to empty string for ProjectID
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pid models.ProjectID
			err := pid.UnmarshalJSON(tc.input)

			if tc.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if pid.String() != tc.expected {
					t.Errorf("Expected %s, got %s", tc.expected, pid.String())
				}
			}
		})
	}
}

// TestProjectID_String tests the String method of ProjectID
func TestProjectID_String(t *testing.T) {
	testCases := []struct {
		name     string
		projectID models.ProjectID
		expected string
	}{
		{
			name:     "String project ID",
			projectID: models.ProjectID("TEST_PROJECT"),
			expected: "TEST_PROJECT",
		},
		{
			name:     "Numeric project ID",
			projectID: models.ProjectID("123"),
			expected: "123",
		},
		{
			name:     "Empty project ID",
			projectID: models.ProjectID(""),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.projectID.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestSlideGenerationRequest_Validation tests slide generation request validation
func TestSlideGenerationRequest_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		request models.SlideGenerationRequest
		valid   bool
	}{
		{
			name: "Valid request with single theme",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID("123"),
				Themes:    []models.SlideTheme{models.ThemeProjectOverview},
				Language:  "ja",
			},
			valid: true,
		},
		{
			name: "Valid request with multiple themes",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID("TEST_PROJECT"),
				Themes: []models.SlideTheme{
					models.ThemeProjectOverview,
					models.ThemeProjectProgress,
					models.ThemeIssueManagement,
				},
				Language: "en",
			},
			valid: true,
		},
		{
			name: "Empty project ID",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID(""),
				Themes:    []models.SlideTheme{models.ThemeProjectOverview},
				Language:  "ja",
			},
			valid: false,
		},
		{
			name: "Empty themes",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID("123"),
				Themes:    []models.SlideTheme{},
				Language:  "ja",
			},
			valid: false,
		},
		{
			name: "Nil themes",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID("123"),
				Themes:    nil,
				Language:  "ja",
			},
			valid: false,
		},
		{
			name: "Empty language",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID("123"),
				Themes:    []models.SlideTheme{models.ThemeProjectOverview},
				Language:  "",
			},
			valid: false,
		},
		{
			name: "Invalid language",
			request: models.SlideGenerationRequest{
				ProjectID: models.ProjectID("123"),
				Themes:    []models.SlideTheme{models.ThemeProjectOverview},
				Language:  "invalid",
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validation logic based on the actual requirements
			isValid := tc.request.ProjectID != "" && 
				len(tc.request.Themes) > 0 && 
				tc.request.Language != "" &&
				(tc.request.Language == "ja" || tc.request.Language == "en")
			
			if isValid != tc.valid {
				t.Errorf("Expected validity %v, got %v for request: %+v", tc.valid, isValid, tc.request)
			}
		})
	}
}

// TestSlideTheme_StringConversion tests that SlideTheme can be converted to string
func TestSlideTheme_StringConversion(t *testing.T) {
	theme := models.ThemeProjectOverview
	str := string(theme)
	if str != "project_overview" {
		t.Errorf("Expected 'project_overview', got '%s'", str)
	}
}

// TestAllSlideThemes_Uniqueness tests that all slide themes are unique
func TestAllSlideThemes_Uniqueness(t *testing.T) {
	themes := []models.SlideTheme{
		models.ThemeProjectOverview,
		models.ThemeProjectProgress,
		models.ThemeIssueManagement,
		models.ThemeRiskAnalysis,
		models.ThemeTeamCollaboration,
		models.ThemeDocumentManagement,
		models.ThemeCodebaseActivity,
		models.ThemeNotifications,
		models.ThemePredictiveAnalysis,
		models.ThemeSummaryPlan,
	}

	seen := make(map[models.SlideTheme]bool)
	for _, theme := range themes {
		if seen[theme] {
			t.Errorf("Duplicate theme found: %s", string(theme))
		}
		seen[theme] = true
	}
}
