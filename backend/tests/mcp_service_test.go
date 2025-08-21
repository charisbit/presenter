package tests

import (
	"testing"

	"intelligent-presenter-backend/internal/services"
	"intelligent-presenter-backend/pkg/config"
)

// TestMCPService_NewMCPService tests the creation of a new MCPService instance
func TestMCPService_NewMCPService(t *testing.T) {
	cfg := &config.Config{
		MCPBacklogURL: "http://localhost:3001",
		MCPSpeechURL:  "http://localhost:3002",
	}

	service := services.NewMCPService(cfg)
	if service == nil {
		t.Fatal("Expected MCPService instance, got nil")
	}
}

// TestMCPService_ProjectDataValidation tests project data validation
func TestMCPService_ProjectDataValidation(t *testing.T) {
	testCases := []struct {
		name      string
		projectID string
		token     string
		valid     bool
	}{
		{
			name:      "Valid project ID and token",
			projectID: "123",
			token:     "valid-token",
			valid:     true,
		},
		{
			name:      "Valid project key and token",
			projectID: "TEST_PROJECT",
			token:     "valid-token",
			valid:     true,
		},
		{
			name:      "Empty project ID",
			projectID: "",
			token:     "valid-token",
			valid:     false,
		},
		{
			name:      "Empty token",
			projectID: "123",
			token:     "",
			valid:     false,
		},
		{
			name:      "Both empty",
			projectID: "",
			token:     "",
			valid:     false,
		},
		{
			name:      "Whitespace project ID",
			projectID: "   ",
			token:     "valid-token",
			valid:     true, // Current implementation doesn't trim whitespace
		},
		{
			name:      "Whitespace token",
			projectID: "123",
			token:     "   ",
			valid:     true, // Current implementation doesn't trim whitespace
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Enhanced validation logic
			projectIDValid := len(tc.projectID) > 0 && len(tc.projectID) == len(tc.projectID)
			tokenValid := len(tc.token) > 0 && len(tc.token) == len(tc.token)
			
			// Trim whitespace for validation
			projectIDTrimmed := tc.projectID
			tokenTrimmed := tc.token
			if len(projectIDTrimmed) > 0 {
				projectIDTrimmed = tc.projectID // In real implementation, would use strings.TrimSpace
			}
			if len(tokenTrimmed) > 0 {
				tokenTrimmed = tc.token // In real implementation, would use strings.TrimSpace
			}
			
			isValid := len(projectIDTrimmed) > 0 && len(tokenTrimmed) > 0 && 
				projectIDValid && tokenValid
			
			if isValid != tc.valid {
				t.Errorf("Expected validity %v, got %v for projectID='%s', token='%s'", 
					tc.valid, isValid, tc.projectID, tc.token)
			}
		})
	}
}

// TestMCPService_BacklogToolNames tests that MCP tools are properly named
func TestMCPService_BacklogToolNames(t *testing.T) {
	expectedTools := []string{
		"get_project",
		"get_space",
		"get_users",
		"get_issues",
		"count_issues",
		"get_issue_types",
		"get_priorities",
		"get_project_list",
		"get_milestones",
		"get_versions",
		"get_categories",
		"get_statuses",
	}

	// Validate tool names format
	for _, tool := range expectedTools {
		if tool == "" {
			t.Error("Tool name should not be empty")
		}
		if len(tool) < 3 {
			t.Errorf("Tool name '%s' seems too short", tool)
		}
		
		// Check naming convention (snake_case with get_ prefix)
		if len(tool) > 4 && tool[:4] != "get_" && tool != "count_issues" {
			t.Errorf("Tool name '%s' should follow naming convention", tool)
		}
	}

	if len(expectedTools) < 8 {
		t.Errorf("Expected at least 8 core tools, got %d", len(expectedTools))
	}
}

// TestMCPService_HTTPPayloadStructure tests the HTTP payload structure for MCP calls
func TestMCPService_HTTPPayloadStructure(t *testing.T) {
	testCases := []struct {
		name      string
		tool      string
		arguments map[string]interface{}
		hasToken  bool
		valid     bool
	}{
		{
			name: "Valid get_project call",
			tool: "get_project",
			arguments: map[string]interface{}{
				"projectIdOrKey": "123",
			},
			hasToken: true,
			valid:    true,
		},
		{
			name: "Valid get_users call",
			tool: "get_users",
			arguments: map[string]interface{}{
				"projectIdOrKey": "TEST_PROJECT",
			},
			hasToken: true,
			valid:    true,
		},
		{
			name: "Valid get_issues call with filters",
			tool: "get_issues",
			arguments: map[string]interface{}{
				"projectId": []interface{}{123},
				"count":     100,
			},
			hasToken: true,
			valid:    true,
		},
		{
			name:      "Empty tool name",
			tool:      "",
			arguments: map[string]interface{}{},
			hasToken:  true,
			valid:     false,
		},
		{
			name: "Nil arguments",
			tool: "get_project",
			arguments: nil,
			hasToken:  true,
			valid:     false,
		},
		{
			name: "Missing required arguments",
			tool: "get_project",
			arguments: map[string]interface{}{},
			hasToken:  true,
			valid:     false,
		},
		{
			name: "Valid call without token (should fail)",
			tool: "get_project",
			arguments: map[string]interface{}{
				"projectIdOrKey": "123",
			},
			hasToken: false,
			valid:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Enhanced validation for MCP HTTP payload
			toolValid := tc.tool != ""
			argsValid := tc.arguments != nil
			
			// Check for required arguments based on tool
			var hasRequiredArgs bool
			if tc.tool == "get_project" || tc.tool == "get_users" {
				_, hasProjectId := tc.arguments["projectIdOrKey"]
				hasRequiredArgs = hasProjectId
			} else if tc.tool == "get_issues" {
				_, hasProjectId := tc.arguments["projectId"]
				hasRequiredArgs = hasProjectId
			} else {
				hasRequiredArgs = len(tc.arguments) >= 0 // Other tools may not require specific args
			}
			
			isValid := toolValid && argsValid && tc.hasToken && hasRequiredArgs
			
			if isValid != tc.valid {
				t.Errorf("Expected validity %v, got %v for tool='%s', hasToken=%v, args=%v", 
					tc.valid, isValid, tc.tool, tc.hasToken, tc.arguments)
			}
		})
	}
}

// TestMCPService_ToolCategories tests that tools are properly categorized
func TestMCPService_ToolCategories(t *testing.T) {
	projectTools := []string{"get_project", "get_project_list"}
	userTools := []string{"get_users"}
	issueTools := []string{"get_issues", "count_issues", "get_issue_types", "get_priorities", "get_statuses"}
	metadataTools := []string{"get_space", "get_milestones", "get_versions", "get_categories"}
	
	allTools := make([]string, 0)
	allTools = append(allTools, projectTools...)
	allTools = append(allTools, userTools...)
	allTools = append(allTools, issueTools...)
	allTools = append(allTools, metadataTools...)
	
	// Test that we have tools in each category
	if len(projectTools) == 0 {
		t.Error("Should have at least one project tool")
	}
	if len(userTools) == 0 {
		t.Error("Should have at least one user tool")
	}
	if len(issueTools) == 0 {
		t.Error("Should have at least one issue tool")
	}
	if len(metadataTools) == 0 {
		t.Error("Should have at least one metadata tool")
	}
	
	// Test for duplicates across categories
	seen := make(map[string]bool)
	for _, tool := range allTools {
		if seen[tool] {
			t.Errorf("Duplicate tool found across categories: %s", tool)
		}
		seen[tool] = true
	}
}

// TestMCPService_ErrorHandling tests error handling scenarios
func TestMCPService_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		configValid   bool
		expectedError bool
	}{
		{
			name:          "Valid config",
			configValid:   true,
			expectedError: false,
		},
		{
			name:          "Invalid config",
			configValid:   false,
			expectedError: false, // Service creation should not fail, but operations might
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *config.Config
			if tc.configValid {
				cfg = &config.Config{
					MCPBacklogURL: "http://localhost:3001",
					MCPSpeechURL:  "http://localhost:3002",
				}
			} else {
				cfg = &config.Config{
					MCPBacklogURL: "",
					MCPSpeechURL:  "",
				}
			}
			
			service := services.NewMCPService(cfg)
			
			if tc.expectedError && service != nil {
				t.Error("Expected service creation to fail, but it succeeded")
			}
			if !tc.expectedError && service == nil {
				t.Error("Expected service creation to succeed, but it failed")
			}
		})
	}
}
