// Package services provides business logic services for the intelligent presenter.
// This file specifically handles Backlog project management integration through
// the Model Context Protocol (MCP) for retrieving project data and analytics.
package services

import (
	"context"
	"encoding/json"
	"fmt"

	"intelligent-presenter-backend/internal/mcp"
	"intelligent-presenter-backend/pkg/config"
)

// BacklogService handles Backlog-specific operations using MCP (Model Context Protocol).
// It provides an abstraction layer for accessing Backlog project management data
// including projects, issues, users, activities, and repository information.
// All operations are performed through MCP client calls to the Backlog MCP server.
type BacklogService struct {
	mcpClient *mcp.MCPClient  // MCP client for communicating with Backlog MCP server
	config    *config.Config  // Application configuration including MCP server URLs
}

// NewBacklogService creates a new Backlog service instance with MCP client initialization.
// It establishes a connection to the Backlog MCP server using the configured URL.
//
// Parameters:
//   - cfg: Application configuration containing Backlog MCP server URL
//
// Returns a configured BacklogService ready for use.
func NewBacklogService(cfg *config.Config) *BacklogService {
	mcpClient := mcp.NewMCPClient(cfg.MCPBacklogURL)
	
	return &BacklogService{
		mcpClient: mcpClient,
		config:    cfg,
	}
}

// Initialize initializes the Backlog MCP client with proper handshake.
// This method establishes the MCP protocol connection with the Backlog server
// and sends the required initialization sequence.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//
// Returns an error if the initialization handshake fails.
func (s *BacklogService) Initialize(ctx context.Context) error {
	clientInfo := map[string]interface{}{
		"name":    "intelligent-presenter-backend",
		"version": "1.0.0",
	}

	return s.mcpClient.Initialize(ctx, clientInfo)
}

// GetProjects retrieves all accessible projects from Backlog.
// This method calls the Backlog MCP server to fetch the complete list
// of projects that the authenticated user has access to.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//
// Returns:
//   - []interface{}: List of project objects containing project details
//   - error: Any error that occurred during the MCP call or data parsing
func (s *BacklogService) GetProjects(ctx context.Context) ([]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getProjectList", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Parse the text content as JSON
	if len(result.Content) > 0 {
		var projects []interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &projects); err != nil {
			return nil, fmt.Errorf("failed to parse projects: %w", err)
		}
		return projects, nil
	}

	return []interface{}{}, nil
}

// GetProject retrieves detailed information for a specific project.
// This method fetches comprehensive project data including metadata,
// settings, and configuration details from Backlog.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - projectKey: The project key or ID to retrieve information for
//
// Returns:
//   - map[string]interface{}: Project data including name, description, settings
//   - error: Any error that occurred during the MCP call or data parsing
func (s *BacklogService) GetProject(ctx context.Context, projectKey string) (map[string]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getProject", map[string]interface{}{
		"projectKey": projectKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}

// GetIssues retrieves issues for a specific project from Backlog.
// This method fetches project issues with their current status, assignees,
// priority levels, and other issue metadata for analysis and reporting.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - projectID: The project identifier to get issues for
//   - count: Maximum number of issues to retrieve (pagination limit)
//
// Returns:
//   - []interface{}: List of issue objects with detailed information
//   - error: Any error that occurred during the MCP call or data parsing
func (s *BacklogService) GetIssues(ctx context.Context, projectID string, count int) ([]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getIssues", map[string]interface{}{
		"projectId": []string{projectID},
		"count":     count,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Parse the text content as JSON
	if len(result.Content) > 0 {
		var issues []interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &issues); err != nil {
			return nil, fmt.Errorf("failed to parse issues: %w", err)
		}
		return issues, nil
	}

	return []interface{}{}, nil
}

// GetProjectUsers retrieves all users associated with a project.
// This method fetches user information including roles, permissions,
// and activity status for team collaboration analysis.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - projectKey: The project key to get user information for
//
// Returns:
//   - []interface{}: List of user objects with roles and details
//   - error: Any error that occurred during the MCP call or data parsing
func (s *BacklogService) GetProjectUsers(ctx context.Context, projectKey string) ([]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getUsers", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Parse the text content as JSON
	if len(result.Content) > 0 {
		var users []interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &users); err != nil {
			return nil, fmt.Errorf("failed to parse users: %w", err)
		}
		return users, nil
	}

	return []interface{}{}, nil
}

// GetProjectActivities retrieves recent project activities and events.
// This method attempts to fetch project activity data, though the functionality
// may be limited depending on the MCP server implementation.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - projectKey: The project key to get activities for
//   - count: Maximum number of activities to retrieve
//
// Returns:
//   - []interface{}: List of activity objects (may be empty if not supported)
//   - error: Any error that occurred during the MCP call
func (s *BacklogService) GetProjectActivities(ctx context.Context, projectKey string, count int) ([]interface{}, error) {
	// Note: This might not be directly available in backlog-mcp-server
	// We'll need to check the available tools first
	_, err := s.mcpClient.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// For now, return empty array as activities might not be available
	return []interface{}{}, nil
}

// GetPullRequests retrieves pull requests for a specific Git repository.
// This method fetches pull request data including status, reviewers,
// and merge information for code collaboration analysis.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - projectKey: The project key containing the repository
//   - repoName: The name of the Git repository to get pull requests for
//
// Returns:
//   - []interface{}: List of pull request objects with detailed information
//   - error: Any error that occurred during the MCP call or data parsing
func (s *BacklogService) GetPullRequests(ctx context.Context, projectKey string, repoName string) ([]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getPullRequests", map[string]interface{}{
		"projectKey": projectKey,
		"repoName":   repoName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Parse the text content as JSON
	if len(result.Content) > 0 {
		var pullRequests []interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &pullRequests); err != nil {
			return nil, fmt.Errorf("failed to parse pull requests: %w", err)
		}
		return pullRequests, nil
	}

	return []interface{}{}, nil
}

// GetGitRepositories gets project Git repositories
func (s *BacklogService) GetGitRepositories(ctx context.Context, projectKey string) ([]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getGitRepositories", map[string]interface{}{
		"projectKey": projectKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get git repositories: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Parse the text content as JSON
	if len(result.Content) > 0 {
		var repositories []interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &repositories); err != nil {
			return nil, fmt.Errorf("failed to parse repositories: %w", err)
		}
		return repositories, nil
	}

	return []interface{}{}, nil
}

// GetWikiPages gets project wiki pages
func (s *BacklogService) GetWikiPages(ctx context.Context, projectKey string) ([]interface{}, error) {
	response, err := s.mcpClient.CallTool(ctx, "getWikiPages", map[string]interface{}{
		"projectKey": projectKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get wiki pages: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Parse the text content as JSON
	if len(result.Content) > 0 {
		var wikiPages []interface{}
		if err := json.Unmarshal([]byte(result.Content[0].Text), &wikiPages); err != nil {
			return nil, fmt.Errorf("failed to parse wiki pages: %w", err)
		}
		return wikiPages, nil
	}

	return []interface{}{}, nil
}

// Close gracefully closes the Backlog service and its MCP client connection.
// This method should be called when the service is no longer needed
// to properly clean up resources and close the MCP protocol connection.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//
// Returns an error if the connection cleanup fails.
func (s *BacklogService) Close(ctx context.Context) error {
	return s.mcpClient.Close(ctx)
}