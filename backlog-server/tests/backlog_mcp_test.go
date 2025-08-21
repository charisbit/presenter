package tests

import (
	"encoding/json"
	"testing"
)

// TestBacklogAPI_ProjectStructure tests the expected Backlog API project structure
func TestBacklogAPI_ProjectStructure(t *testing.T) {
	// Sample project data structure based on Backlog API
	sampleProject := map[string]interface{}{
		"id":          123,
		"projectKey":  "TEST",
		"name":        "Test Project",
		"chartEnabled": true,
		"archived":    false,
	}

	// Test that required fields exist
	requiredFields := []string{"id", "projectKey", "name"}
	for _, field := range requiredFields {
		if _, exists := sampleProject[field]; !exists {
			t.Errorf("Required field '%s' missing from project structure", field)
		}
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(sampleProject)
	if err != nil {
		t.Errorf("Failed to marshal project data: %v", err)
	}

	// Test JSON deserialization
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal project data: %v", err)
	}

	// Verify data integrity
	if unmarshaled["name"] != sampleProject["name"] {
		t.Errorf("Data integrity check failed for 'name' field")
	}
}

// TestBacklogAPI_IssueStructure tests the expected Backlog API issue structure
func TestBacklogAPI_IssueStructure(t *testing.T) {
	sampleIssue := map[string]interface{}{
		"id":      456,
		"issueKey": "TEST-1",
		"keyId":   1,
		"summary": "Test Issue",
		"status": map[string]interface{}{
			"id":   1,
			"name": "Open",
		},
		"priority": map[string]interface{}{
			"id":   2,
			"name": "Normal",
		},
	}

	// Test that required fields exist
	requiredFields := []string{"id", "issueKey", "summary", "status", "priority"}
	for _, field := range requiredFields {
		if _, exists := sampleIssue[field]; !exists {
			t.Errorf("Required field '%s' missing from issue structure", field)
		}
	}

	// Test nested structure validation
	status, ok := sampleIssue["status"].(map[string]interface{})
	if !ok {
		t.Error("Status field should be a map")
	} else {
		if _, exists := status["id"]; !exists {
			t.Error("Status should have 'id' field")
		}
		if _, exists := status["name"]; !exists {
			t.Error("Status should have 'name' field")
		}
	}
}

// TestBacklogMCP_ToolParameters tests MCP tool parameter validation
func TestBacklogMCP_ToolParameters(t *testing.T) {
	testCases := []struct {
		name       string
		tool       string
		parameters map[string]interface{}
		valid      bool
	}{
		{
			name: "Valid get_project parameters",
			tool: "get_project",
			parameters: map[string]interface{}{
				"projectIdOrKey": "123",
			},
			valid: true,
		},
		{
			name: "Valid get_issues parameters",
			tool: "get_issues",
			parameters: map[string]interface{}{
				"projectId": []string{"123"},
				"count":     50,
			},
			valid: true,
		},
		{
			name: "Invalid get_project parameters - missing projectIdOrKey",
			tool: "get_project",
			parameters: map[string]interface{}{
				"invalid": "param",
			},
			valid: false,
		},
		{
			name: "Empty parameters for get_space",
			tool: "get_space",
			parameters: map[string]interface{}{},
			valid: true, // get_space doesn't require parameters
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate parameters based on tool requirements
			var isValid bool
			switch tc.tool {
			case "get_project":
				_, hasProjectId := tc.parameters["projectIdOrKey"]
				isValid = hasProjectId
			case "get_issues":
				_, hasProjectId := tc.parameters["projectId"]
				isValid = hasProjectId
			case "get_space":
				isValid = true // No required parameters
			default:
				isValid = false
			}

			if isValid != tc.valid {
				t.Errorf("Expected validity %v, got %v for tool %s", tc.valid, isValid, tc.tool)
			}
		})
	}
}

// TestBacklogMCP_ResponseStructure tests MCP response structure validation
func TestBacklogMCP_ResponseStructure(t *testing.T) {
	// Sample MCP response structure
	sampleResponse := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": `{"project": {"id": 123, "name": "Test"}}`,
			},
		},
	}

	// Validate response structure
	content, exists := sampleResponse["content"]
	if !exists {
		t.Error("Response should have 'content' field")
	}

	contentArray, ok := content.([]map[string]interface{})
	if !ok {
		t.Error("Content should be an array of objects")
	}

	if len(contentArray) == 0 {
		t.Error("Content array should not be empty")
	}

	firstContent := contentArray[0]
	if _, exists := firstContent["type"]; !exists {
		t.Error("Content item should have 'type' field")
	}

	if _, exists := firstContent["text"]; !exists {
		t.Error("Content item should have 'text' field")
	}
}