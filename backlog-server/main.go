// Package main provides a Backlog MCP (Model Context Protocol) server implementation.
// This server bridges the Backlog API with MCP protocol, enabling intelligent presenter
// applications to access Backlog project management data through standardized MCP tools.
//
// The server supports two operational modes:
//   1. MCP Server Mode: Direct stdin/stdout JSON-RPC communication for MCP clients
//   2. HTTP Bridge Mode: RESTful HTTP API that translates HTTP requests to MCP calls
//
// Authentication methods supported:
//   - API Key authentication for direct API access
//   - OAuth2 access token authentication for user-specific access
//   - Dynamic token provision via HTTP bridge requests
//
// The server provides comprehensive Backlog API coverage including:
//   - Project management (CRUD operations)
//   - Issue tracking and management
//   - Wiki page management
//   - Git repository and pull request access
//   - User and team management
//   - Notification handling
//   - Document management
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

// ==========================================
// MCP Protocol Types
// ==========================================

// MCPRequest represents a Model Context Protocol JSON-RPC request.
// It follows the JSON-RPC 2.0 specification with MCP-specific extensions
// for method calls and parameter passing to Backlog API tools.
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`        // JSON-RPC version (always "2.0")
	ID      *int64      `json:"id,omitempty"`   // Request identifier for response matching
	Method  string      `json:"method"`         // MCP method name to invoke
	Params  interface{} `json:"params,omitempty"` // Method parameters (tool-specific)
}

// MCPResponse represents a Model Context Protocol JSON-RPC response.
// It contains either successful tool execution results or error information
// according to the JSON-RPC 2.0 specification.
type MCPResponse struct {
	JSONRPC string           `json:"jsonrpc"`          // JSON-RPC version (always "2.0")
	ID      *int64           `json:"id,omitempty"`     // Request identifier matching the request
	Result  *json.RawMessage `json:"result,omitempty"` // Successful result data from tool execution
	Error   *MCPError        `json:"error,omitempty"`  // Error information if tool execution failed
}

// MCPError represents an MCP protocol error response.
// It provides structured error information including standard JSON-RPC error codes
// and detailed error messages for debugging and client handling.
type MCPError struct {
	Code    int         `json:"code"`             // Error code (following JSON-RPC error codes)
	Message string      `json:"message"`          // Human-readable error message
	Data    interface{} `json:"data,omitempty"`   // Additional error data (optional)
}

// InitializeResult represents the MCP server initialization response.
// It contains protocol version information, server capabilities,
// and metadata about the Backlog MCP server implementation.
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"` // MCP protocol version supported
	Capabilities    map[string]interface{} `json:"capabilities"`    // Server capabilities (tools, resources, etc.)
	ServerInfo      ServerInfo             `json:"serverInfo"`      // Server identification information
}

// ServerInfo contains identification metadata for the MCP server.
// This information is used by clients to identify the server implementation
// and version for compatibility and debugging purposes.
type ServerInfo struct {
	Name    string `json:"name"`    // Server implementation name
	Version string `json:"version"` // Server version string
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]Property    `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Items      *Property              `json:"items,omitempty"`
	Enum       []string               `json:"enum,omitempty"`
}

type Property struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Items       *Property              `json:"items,omitempty"`
	Properties  map[string]Property    `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Maximum     *float64               `json:"maximum,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ==========================================
// Backlog API Client
// ==========================================

// BacklogClient provides HTTP client functionality for accessing Backlog API.
// It handles authentication (OAuth2 access tokens or API keys), request formatting,
// parameter serialization, and response processing for all Backlog API endpoints.
// The client supports both read and write operations across all Backlog features.
type BacklogClient struct {
	client      *resty.Client // HTTP client for API requests
	baseURL     string        // Backlog API base URL (e.g., https://example.backlog.jp/api/v2)
	accessToken string        // OAuth2 access token for user authentication
	apiKey      string        // API key for service authentication
}

// NewBacklogClient creates a new Backlog API client with authentication.
// It initializes the HTTP client, constructs the API base URL, and sets up
// authentication headers based on the provided credentials.
//
// Parameters:
//   - domain: Backlog space domain (e.g., "yourspace.backlog.jp")
//   - accessToken: OAuth2 access token for user authentication (optional)
//   - apiKey: API key for service authentication (optional)
//
// Returns:
//   - *BacklogClient: Configured client ready for API calls
//   - error: Error if domain validation fails
//
// At least one authentication method (accessToken or apiKey) should be provided.
func NewBacklogClient(domain, accessToken, apiKey string) (*BacklogClient, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	client := resty.New()
	baseURL := fmt.Sprintf("https://%s/api/v2", domain)

	bc := &BacklogClient{
		client:      client,
		baseURL:     baseURL,
		accessToken: accessToken,
		apiKey:      apiKey,
	}

	bc.setupAuth()
	return bc, nil
}

// setupAuth configures authentication headers and parameters for the HTTP client.
// It sets up either OAuth2 Bearer token authentication or API key query parameter
// authentication based on the available credentials.
//
// Authentication priority:
//   1. OAuth2 access token (Bearer header) - preferred for user-specific access
//   2. API key (query parameter) - fallback for service access
func (bc *BacklogClient) setupAuth() {
	if bc.accessToken != "" {
		bc.client.SetHeader("Authorization", "Bearer "+bc.accessToken)
		bc.client.SetHeader("Content-Type", "application/json")
	} else if bc.apiKey != "" {
		bc.client.SetQueryParam("apiKey", bc.apiKey)
		bc.client.SetHeader("Content-Type", "application/json")
	}
}

func (bc *BacklogClient) makeRequest(method, endpoint string, params map[string]interface{}, body interface{}) (interface{}, error) {
	var result interface{}
	req := bc.client.R().SetResult(&result)

	// Add query parameters for GET requests
	if method == "GET" && params != nil {
		for key, value := range params {
			if key == "projectId" || key == "issueTypeId" || key == "statusId" || key == "priorityId" || key == "assigneeId" || key == "createdUserId" || key == "issueId" || key == "categoryId" || key == "versionId" || key == "milestoneId" || key == "notifiedUserId" || key == "attachmentId" || key == "repoId" || key == "pullRequestId" {
				if ids, ok := value.([]interface{}); ok {
					for _, id := range ids {
						req = req.SetQueryParam(key+"[]", fmt.Sprintf("%v", id))
					}
				} else {
					req = req.SetQueryParam(key, fmt.Sprintf("%v", value))
				}
			} else {
				req = req.SetQueryParam(key, fmt.Sprintf("%v", value))
			}
		}
	}

	// Add form data for POST/PUT requests with body
	if (method == "POST" || method == "PUT") && body != nil {
		if bodyMap, ok := body.(map[string]interface{}); ok {
			formData := make(map[string]string)
			for key, value := range bodyMap {
				if key == "categoryId" || key == "versionId" || key == "milestoneId" || key == "notifiedUserId" || key == "attachmentId" {
					if ids, ok := value.([]interface{}); ok {
						for i, id := range ids {
							formData[key+"["+fmt.Sprintf("%d", i)+"]"] = fmt.Sprintf("%v", id)
						}
					} else {
						formData[key] = fmt.Sprintf("%v", value)
					}
				} else {
					formData[key] = fmt.Sprintf("%v", value)
				}
			}
			req = req.SetFormData(formData)
		}
	}

	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = req.Get(bc.baseURL + endpoint)
	case "POST":
		resp, err = req.Post(bc.baseURL + endpoint)
	case "PUT":
		resp, err = req.Put(bc.baseURL + endpoint)
	case "DELETE":
		resp, err = req.Delete(bc.baseURL + endpoint)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		log.Printf("HTTP request failed for %s %s: %v", method, endpoint, err)
		return nil, fmt.Errorf("failed to make request to %s: %w", endpoint, err)
	}

	log.Printf("HTTP response for %s %s: status=%d, body_length=%d", method, endpoint, resp.StatusCode(), len(resp.Body()))

	if resp.IsError() {
		log.Printf("API error for %s %s: status=%d, response=%s", method, endpoint, resp.StatusCode(), resp.String())
		return nil, fmt.Errorf("API error: %s", resp.String())
	}

	return result, nil
}

// ==========================================
// MCP Server
// ==========================================

// MCPServer implements the Model Context Protocol server for Backlog API access.
// It manages tool definitions, handles MCP protocol requests, and executes
// Backlog API operations through the configured BacklogClient.
type MCPServer struct {
	backlogClient *BacklogClient // Backlog API client for executing operations
	tools         []Tool         // Available MCP tools for Backlog operations
}

// NewMCPServer creates a new MCP server instance with Backlog integration.
// It initializes the server with the provided Backlog client and sets up
// all available tools for Backlog API operations.
//
// Parameters:
//   - backlogClient: Configured Backlog API client (can be nil for OAuth-only mode)
//
// Returns a fully configured MCP server ready to handle protocol requests.
func NewMCPServer(backlogClient *BacklogClient) *MCPServer {
	s := &MCPServer{
		backlogClient: backlogClient,
	}
	s.initializeTools()
	return s
}

func (s *MCPServer) initializeTools() {
	s.tools = []Tool{
		// Space tools
		{Name: "get_space", Description: "Get information about the Backlog space", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}}},
		{Name: "get_users", Description: "Get list of users in the space", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}}},
		{Name: "get_myself", Description: "Get information about the current user", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}}},

		// Project tools
		{
			Name:        "get_project_list",
			Description: "Get list of projects",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"archived": {Type: "boolean", Description: "Filter by archived status"},
					"all":      {Type: "boolean", Description: "Get all projects (admin only)"},
				},
			},
		},
		{
			Name:        "get_project",
			Description: "Get project details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":     {Type: "number", Description: "Project ID"},
					"projectKey":    {Type: "string", Description: "Project key"},
					"projectIdOrKey": {Type: "string", Description: "Project ID or key"},
				},
			},
		},
		{
			Name:        "add_project",
			Description: "Create a new project",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":             {Type: "string", Description: "Project name"},
					"key":              {Type: "string", Description: "Project key"},
					"chartEnabled":     {Type: "boolean", Description: "Enable charts"},
					"subtaskingEnabled": {Type: "boolean", Description: "Enable subtasking"},
					"projectLeaderCanEditProjectLeader": {Type: "boolean", Description: "Allow project leader to edit project leader"},
					"useWikiTreeView":                   {Type: "boolean", Description: "Use wiki tree view"},
					"textFormattingRule":                {Type: "string", Description: "Text formatting rule"},
				},
				Required: []string{"name", "key"},
			},
		},
		{
			Name:        "update_project",
			Description: "Update project settings",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
					"name":       {Type: "string", Description: "Project name"},
					"archived":   {Type: "boolean", Description: "Archive status"},
				},
			},
		},
		{
			Name:        "delete_project",
			Description: "Delete a project",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
				},
			},
		},

		// Issue tools (existing + new)
		{
			Name:        "get_issues",
			Description: "Get list of issues",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":    {Type: "array", Items: &Property{Type: "number"}, Description: "Project IDs"},
					"issueTypeId":  {Type: "array", Items: &Property{Type: "number"}, Description: "Issue type IDs"},
					"statusId":     {Type: "array", Items: &Property{Type: "number"}, Description: "Status IDs"},
					"priorityId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Priority IDs"},
					"assigneeId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Assignee user IDs"},
					"createdUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Created user IDs"},
					"resolutionId": {Type: "array", Items: &Property{Type: "number"}, Description: "Resolution IDs"},
					"parentIssueId": {Type: "array", Items: &Property{Type: "number"}, Description: "Parent issue IDs"},
					"keyword":      {Type: "string", Description: "Search keyword"},
					"sort":         {Type: "string", Description: "Sort field"},
					"order":        {Type: "string", Enum: []string{"asc", "desc"}, Description: "Sort order"},
					"offset":       {Type: "number", Description: "Offset for pagination"},
					"count":        {Type: "number", Description: "Number of items to return"},
					"createdSince": {Type: "string", Description: "Created since (yyyy-MM-dd)"},
					"createdUntil": {Type: "string", Description: "Created until (yyyy-MM-dd)"},
					"updatedSince": {Type: "string", Description: "Updated since (yyyy-MM-dd)"},
					"updatedUntil": {Type: "string", Description: "Updated until (yyyy-MM-dd)"},
					"startDateSince": {Type: "string", Description: "Start date since (yyyy-MM-dd)"},
					"startDateUntil": {Type: "string", Description: "Start date until (yyyy-MM-dd)"},
					"dueDateSince":   {Type: "string", Description: "Due date since (yyyy-MM-dd)"},
					"dueDateUntil":   {Type: "string", Description: "Due date until (yyyy-MM-dd)"},
				},
			},
		},
		{
			Name:        "get_issue",
			Description: "Get specific issue details",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{"issueIdOrKey": {Type: "string", Description: "Issue ID or key"}},
				Required:   []string{"issueIdOrKey"},
			},
		},
		{
			Name:        "add_issue",
			Description: "Create a new issue",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":      {Type: "number", Description: "Project ID"},
					"summary":        {Type: "string", Description: "Issue summary"},
					"issueTypeId":    {Type: "number", Description: "Issue type ID"},
					"priorityId":     {Type: "number", Description: "Priority ID"},
					"description":    {Type: "string", Description: "Issue description"},
					"startDate":      {Type: "string", Description: "Start date (yyyy-MM-dd)"},
					"dueDate":        {Type: "string", Description: "Due date (yyyy-MM-dd)"},
					"estimatedHours": {Type: "number", Description: "Estimated hours"},
					"actualHours":    {Type: "number", Description: "Actual hours"},
					"assigneeId":     {Type: "number", Description: "Assignee user ID"},
					"parentIssueId":  {Type: "number", Description: "Parent issue ID"},
					"categoryId":     {Type: "array", Items: &Property{Type: "number"}, Description: "Category IDs"},
					"versionId":      {Type: "array", Items: &Property{Type: "number"}, Description: "Version IDs"},
					"milestoneId":    {Type: "array", Items: &Property{Type: "number"}, Description: "Milestone IDs"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
					"attachmentId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Attachment IDs"},
				},
				Required: []string{"projectId", "summary", "issueTypeId", "priorityId"},
			},
		},
		{
			Name:        "update_issue",
			Description: "Update an existing issue",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"issueIdOrKey":   {Type: "string", Description: "Issue ID or key"},
					"summary":        {Type: "string", Description: "Issue summary"},
					"description":    {Type: "string", Description: "Issue description"},
					"statusId":       {Type: "number", Description: "Status ID"},
					"priorityId":     {Type: "number", Description: "Priority ID"},
					"assigneeId":     {Type: "number", Description: "Assignee user ID"},
					"resolutionId":   {Type: "number", Description: "Resolution ID"},
					"startDate":      {Type: "string", Description: "Start date (yyyy-MM-dd)"},
					"dueDate":        {Type: "string", Description: "Due date (yyyy-MM-dd)"},
					"estimatedHours": {Type: "number", Description: "Estimated hours"},
					"actualHours":    {Type: "number", Description: "Actual hours"},
					"parentIssueId":  {Type: "number", Description: "Parent issue ID"},
					"categoryId":     {Type: "array", Items: &Property{Type: "number"}, Description: "Category IDs"},
					"versionId":      {Type: "array", Items: &Property{Type: "number"}, Description: "Version IDs"},
					"milestoneId":    {Type: "array", Items: &Property{Type: "number"}, Description: "Milestone IDs"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
					"comment":        {Type: "string", Description: "Update comment"},
				},
				Required: []string{"issueIdOrKey"},
			},
		},
		{
			Name:        "delete_issue",
			Description: "Delete an issue",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{"issueIdOrKey": {Type: "string", Description: "Issue ID or key"}},
				Required:   []string{"issueIdOrKey"},
			},
		},
		{
			Name:        "get_issue_comments",
			Description: "Get comments for an issue",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"issueIdOrKey": {Type: "string", Description: "Issue ID or key"},
					"minId":        {Type: "number", Description: "Minimum comment ID"},
					"maxId":        {Type: "number", Description: "Maximum comment ID"},
					"count":        {Type: "number", Description: "Number of comments to return"},
					"order":        {Type: "string", Enum: []string{"asc", "desc"}, Description: "Sort order"},
				},
				Required: []string{"issueIdOrKey"},
			},
		},
		{
			Name:        "add_issue_comment",
			Description: "Add comment to an issue",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"issueIdOrKey":   {Type: "string", Description: "Issue ID or key"},
					"content":        {Type: "string", Description: "Comment content"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
					"attachmentId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Attachment IDs"},
				},
				Required: []string{"issueIdOrKey", "content"},
			},
		},
		{
			Name:        "count_issues",
			Description: "Count issues matching criteria",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId": {Type: "array", Items: &Property{Type: "number"}, Description: "Project IDs"},
					"statusId":  {Type: "array", Items: &Property{Type: "number"}, Description: "Status IDs"},
				},
			},
		},
		{
			Name:        "get_custom_fields",
			Description: "Get custom fields for a project",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{"projectIdOrKey": {Type: "string", Description: "Project ID or key"}},
				Required:   []string{"projectIdOrKey"},
			},
		},
		{
			Name:        "get_watching_list_items",
			Description: "Get watching list items",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"userId": {Type: "number", Description: "User ID"},
					"order":  {Type: "string", Enum: []string{"asc", "desc"}, Description: "Sort order"},
					"sort":   {Type: "string", Description: "Sort field"},
					"offset": {Type: "number", Description: "Offset for pagination"},
					"count":  {Type: "number", Description: "Number of items to return"},
				},
			},
		},
		{
			Name:        "get_watching_list_count",
			Description: "Get count of watching list items",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"userId": {Type: "number", Description: "User ID"},
				},
			},
		},

		// Issue metadata tools
		{
			Name:        "get_issue_types",
			Description: "Get issue types for a project",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{"projectIdOrKey": {Type: "string", Description: "Project ID or key"}},
				Required:   []string{"projectIdOrKey"},
			},
		},
		{Name: "get_priorities", Description: "Get issue priorities", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}}},
		{Name: "get_resolutions", Description: "Get issue resolutions", InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}}},
		{
			Name:        "get_categories",
			Description: "Get categories for a project",
			InputSchema: InputSchema{
				Type:       "object",
				Properties: map[string]Property{"projectIdOrKey": {Type: "string", Description: "Project ID or key"}},
				Required:   []string{"projectIdOrKey"},
			},
		},

		// Wiki tools
		{
			Name:        "get_wiki_pages",
			Description: "Get list of wiki pages",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
					"keyword":    {Type: "string", Description: "Search keyword"},
				},
			},
		},
		{
			Name:        "get_wikis_count",
			Description: "Get count of wiki pages",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
				},
			},
		},
		{
			Name:        "get_wiki",
			Description: "Get wiki page details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"wikiId": {Type: "number", Description: "Wiki page ID"},
				},
				Required: []string{"wikiId"},
			},
		},
		{
			Name:        "add_wiki",
			Description: "Create a new wiki page",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":      {Type: "number", Description: "Project ID"},
					"name":           {Type: "string", Description: "Wiki page name"},
					"content":        {Type: "string", Description: "Wiki page content"},
					"mailNotify":     {Type: "boolean", Description: "Send email notification"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
				},
				Required: []string{"projectId", "name", "content"},
			},
		},

		// Git & Pull Request tools
		{
			Name:        "get_git_repositories",
			Description: "Get git repositories for a project",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
				},
			},
		},
		{
			Name:        "get_git_repository",
			Description: "Get git repository details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
					"repoId":     {Type: "number", Description: "Repository ID"},
					"repoName":   {Type: "string", Description: "Repository name"},
				},
			},
		},
		{
			Name:        "get_pull_requests",
			Description: "Get pull requests for a repository",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":     {Type: "number", Description: "Project ID"},
					"projectKey":    {Type: "string", Description: "Project key"},
					"repoId":        {Type: "number", Description: "Repository ID"},
					"repoName":      {Type: "string", Description: "Repository name"},
					"statusId":      {Type: "array", Items: &Property{Type: "number"}, Description: "Status IDs"},
					"assigneeId":    {Type: "array", Items: &Property{Type: "number"}, Description: "Assignee user IDs"},
					"issueId":       {Type: "array", Items: &Property{Type: "number"}, Description: "Issue IDs"},
					"createdUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Created user IDs"},
					"offset":        {Type: "number", Description: "Offset for pagination"},
					"count":         {Type: "number", Description: "Number of items to return"},
				},
			},
		},
		{
			Name:        "get_pull_requests_count",
			Description: "Get count of pull requests",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
					"repoId":     {Type: "number", Description: "Repository ID"},
					"repoName":   {Type: "string", Description: "Repository name"},
					"statusId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Status IDs"},
				},
			},
		},
		{
			Name:        "get_pull_request",
			Description: "Get pull request details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":    {Type: "number", Description: "Project ID"},
					"projectKey":   {Type: "string", Description: "Project key"},
					"repoId":       {Type: "number", Description: "Repository ID"},
					"repoName":     {Type: "string", Description: "Repository name"},
					"pullRequestId": {Type: "number", Description: "Pull request ID"},
				},
				Required: []string{"pullRequestId"},
			},
		},
		{
			Name:        "add_pull_request",
			Description: "Create a new pull request",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":      {Type: "number", Description: "Project ID"},
					"projectKey":     {Type: "string", Description: "Project key"},
					"repoId":         {Type: "number", Description: "Repository ID"},
					"repoName":       {Type: "string", Description: "Repository name"},
					"summary":        {Type: "string", Description: "Pull request summary"},
					"description":    {Type: "string", Description: "Pull request description"},
					"base":           {Type: "string", Description: "Base branch"},
					"branch":         {Type: "string", Description: "Feature branch"},
					"assigneeId":     {Type: "number", Description: "Assignee user ID"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
					"attachmentId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Attachment IDs"},
				},
				Required: []string{"summary", "base", "branch"},
			},
		},
		{
			Name:        "update_pull_request",
			Description: "Update a pull request",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":      {Type: "number", Description: "Project ID"},
					"projectKey":     {Type: "string", Description: "Project key"},
					"repoId":         {Type: "number", Description: "Repository ID"},
					"repoName":       {Type: "string", Description: "Repository name"},
					"pullRequestId":  {Type: "number", Description: "Pull request ID"},
					"summary":        {Type: "string", Description: "Pull request summary"},
					"description":    {Type: "string", Description: "Pull request description"},
					"assigneeId":     {Type: "number", Description: "Assignee user ID"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
					"comment":        {Type: "string", Description: "Update comment"},
				},
				Required: []string{"pullRequestId"},
			},
		},
		{
			Name:        "get_pull_request_comments",
			Description: "Get comments for a pull request",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":     {Type: "number", Description: "Project ID"},
					"projectKey":    {Type: "string", Description: "Project key"},
					"repoId":        {Type: "number", Description: "Repository ID"},
					"repoName":      {Type: "string", Description: "Repository name"},
					"pullRequestId": {Type: "number", Description: "Pull request ID"},
					"minId":         {Type: "number", Description: "Minimum comment ID"},
					"maxId":         {Type: "number", Description: "Maximum comment ID"},
					"count":         {Type: "number", Description: "Number of comments to return"},
					"order":         {Type: "string", Enum: []string{"asc", "desc"}, Description: "Sort order"},
				},
				Required: []string{"pullRequestId"},
			},
		},
		{
			Name:        "add_pull_request_comment",
			Description: "Add comment to a pull request",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":      {Type: "number", Description: "Project ID"},
					"projectKey":     {Type: "string", Description: "Project key"},
					"repoId":         {Type: "number", Description: "Repository ID"},
					"repoName":       {Type: "string", Description: "Repository name"},
					"pullRequestId":  {Type: "number", Description: "Pull request ID"},
					"content":        {Type: "string", Description: "Comment content"},
					"notifiedUserId": {Type: "array", Items: &Property{Type: "number"}, Description: "Notified user IDs"},
					"attachmentId":   {Type: "array", Items: &Property{Type: "number"}, Description: "Attachment IDs"},
				},
				Required: []string{"pullRequestId", "content"},
			},
		},
		{
			Name:        "update_pull_request_comment",
			Description: "Update a pull request comment",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":     {Type: "number", Description: "Project ID"},
					"projectKey":    {Type: "string", Description: "Project key"},
					"repoId":        {Type: "number", Description: "Repository ID"},
					"repoName":      {Type: "string", Description: "Repository name"},
					"pullRequestId": {Type: "number", Description: "Pull request ID"},
					"commentId":     {Type: "number", Description: "Comment ID"},
					"content":       {Type: "string", Description: "Updated comment content"},
				},
				Required: []string{"pullRequestId", "commentId", "content"},
			},
		},

		// Document tools
		{
			Name:        "get_documents",
			Description: "Get documents for a project",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
					"path":       {Type: "string", Description: "Document path"},
				},
			},
		},
		{
			Name:        "get_document_tree",
			Description: "Get document tree structure",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"projectId":  {Type: "number", Description: "Project ID"},
					"projectKey": {Type: "string", Description: "Project key"},
				},
			},
		},
		{
			Name:        "get_document",
			Description: "Get document details",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"documentId": {Type: "number", Description: "Document ID"},
				},
				Required: []string{"documentId"},
			},
		},

		// Notifications tools
		{
			Name:        "get_notifications",
			Description: "Get notifications for current user",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"minId":   {Type: "number", Description: "Minimum notification ID"},
					"maxId":   {Type: "number", Description: "Maximum notification ID"},
					"count":   {Type: "number", Description: "Number of notifications to return"},
					"order":   {Type: "string", Enum: []string{"asc", "desc"}, Description: "Sort order"},
				},
			},
		},
		{
			Name:        "get_notifications_count",
			Description: "Get count of notifications",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"alreadyRead": {Type: "boolean", Description: "Filter by read status"},
				},
			},
		},
		{
			Name:        "reset_unread_notification_count",
			Description: "Reset unread notification count",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "mark_notification_as_read",
			Description: "Mark notification as read",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "number", Description: "Notification ID"},
				},
				Required: []string{"id"},
			},
		},

		
	}
}

func (s *MCPServer) HandleRequest(request MCPRequest) MCPResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "notifications/initialized":
		return MCPResponse{JSONRPC: "2.0", ID: request.ID}
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(request)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error:   &MCPError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", request.Method)},
		}
	}
}

func (s *MCPServer) handleInitialize(request MCPRequest) MCPResponse {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]interface{}{"tools": map[string]interface{}{}},
		ServerInfo:      ServerInfo{Name: "backlog-mcp-go", Version: "1.0.0"},
	}

	resultBytes, _ := json.Marshal(result)
	resultRaw := json.RawMessage(resultBytes)

	return MCPResponse{JSONRPC: "2.0", ID: request.ID, Result: &resultRaw}
}

func (s *MCPServer) handleToolsList(request MCPRequest) MCPResponse {
	result := ToolsListResult{Tools: s.tools}
	resultBytes, _ := json.Marshal(result)
	resultRaw := json.RawMessage(resultBytes)

	return MCPResponse{JSONRPC: "2.0", ID: request.ID, Result: &resultRaw}
}

func (s *MCPServer) handleToolsCall(request MCPRequest) MCPResponse {
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return MCPResponse{JSONRPC: "2.0", ID: request.ID, Error: &MCPError{Code: -32602, Message: "Invalid params"}}
	}

	var params CallToolParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return MCPResponse{JSONRPC: "2.0", ID: request.ID, Error: &MCPError{Code: -32602, Message: "Invalid params"}}
	}

	result, err := s.executeTool(params.Name, params.Arguments)
	if err != nil {
		return MCPResponse{JSONRPC: "2.0", ID: request.ID, Error: &MCPError{Code: -32603, Message: err.Error()}}
	}

	resultBytes, _ := json.Marshal(result)
	resultRaw := json.RawMessage(resultBytes)

	return MCPResponse{JSONRPC: "2.0", ID: request.ID, Result: &resultRaw}
}

func (s *MCPServer) executeTool(toolName string, args map[string]interface{}) (*CallToolResult, error) {
	var data interface{}
	var err error

	log.Printf("Executing tool: %s with args: %+v", toolName, args)

	switch toolName {
	// Space tools
	case "get_space":
		log.Printf("Making request to /space")
		data, err = s.backlogClient.makeRequest("GET", "/space", nil, nil)
	case "get_users":
		log.Printf("Making request to /users")
		data, err = s.backlogClient.makeRequest("GET", "/users", nil, nil)
		if err != nil {
			log.Printf("get_users failed with error: %v", err)
		} else {
			log.Printf("get_users succeeded, data type: %T", data)
		}
	case "get_myself":
		log.Printf("Making request to /users/myself")
		data, err = s.backlogClient.makeRequest("GET", "/users/myself", nil, nil)

	// Project tools
	case "get_project_list":
		params := make(map[string]interface{})
		if archived, ok := args["archived"]; ok {
			params["archived"] = archived
		}
		if all, ok := args["all"]; ok {
			params["all"] = all
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects", params, nil)

	case "get_project":
		var projectIdOrKey string
		if projectIdOrKeyParam, ok := args["projectIdOrKey"].(string); ok {
			projectIdOrKey = projectIdOrKeyParam
		} else if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId, projectKey, or projectIdOrKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey, nil, nil)

	case "add_project":
		if name, ok := args["name"].(string); !ok || name == "" {
			return nil, fmt.Errorf("name is required")
		}
		if key, ok := args["key"].(string); !ok || key == "" {
			return nil, fmt.Errorf("key is required")
		}
		data, err = s.backlogClient.makeRequest("POST", "/projects", nil, args)

	case "update_project":
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		delete(args, "projectId")
		delete(args, "projectKey")
		data, err = s.backlogClient.makeRequest("PUT", "/projects/"+projectIdOrKey, nil, args)

	case "delete_project":
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		data, err = s.backlogClient.makeRequest("DELETE", "/projects/"+projectIdOrKey, nil, nil)

	// Issue tools
	case "get_issues":
		params := make(map[string]interface{})
		for key, value := range args {
			params[key] = value
		}
		data, err = s.backlogClient.makeRequest("GET", "/issues", params, nil)

	

	case "get_issue":
		issueIdOrKey, ok := args["issueIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("issueIdOrKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/issues/"+issueIdOrKey, nil, nil)

	case "add_issue":
		requiredFields := []string{"projectId", "summary", "issueTypeId", "priorityId"}
		for _, field := range requiredFields {
			if _, ok := args[field]; !ok {
				return nil, fmt.Errorf("%s is required", field)
			}
		}
		data, err = s.backlogClient.makeRequest("POST", "/issues", nil, args)

	case "update_issue":
		issueIdOrKey, ok := args["issueIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("issueIdOrKey is required")
		}
		delete(args, "issueIdOrKey")
		data, err = s.backlogClient.makeRequest("PUT", "/issues/"+issueIdOrKey, nil, args)

	case "delete_issue":
		issueIdOrKey, ok := args["issueIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("issueIdOrKey is required")
		}
		data, err = s.backlogClient.makeRequest("DELETE", "/issues/"+issueIdOrKey, nil, nil)

	case "get_issue_comments":
		issueIdOrKey, ok := args["issueIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("issueIdOrKey is required")
		}
		params := make(map[string]interface{})
		for key, value := range args {
			if key != "issueIdOrKey" {
				params[key] = value
			}
		}
		data, err = s.backlogClient.makeRequest("GET", "/issues/"+issueIdOrKey+"/comments", params, nil)

	case "add_issue_comment":
		issueIdOrKey, ok := args["issueIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("issueIdOrKey is required")
		}
		if _, ok := args["content"]; !ok {
			return nil, fmt.Errorf("content is required")
		}
		delete(args, "issueIdOrKey")
		data, err = s.backlogClient.makeRequest("POST", "/issues/"+issueIdOrKey+"/comments", nil, args)

	case "count_issues":
		params := make(map[string]interface{})
		if projectId, ok := args["projectId"]; ok {
			params["projectId"] = projectId
		}
		if statusId, ok := args["statusId"]; ok {
			params["statusId"] = statusId
		}
		data, err = s.backlogClient.makeRequest("GET", "/issues/count", params, nil)

	case "get_custom_fields":
		projectIdOrKey, ok := args["projectIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("projectIdOrKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/customFields", nil, nil)

	case "get_watching_list_items":
		params := make(map[string]interface{})
		for key, value := range args {
			params[key] = value
		}
		data, err = s.backlogClient.makeRequest("GET", "/users/myself/watchings", params, nil)

	case "get_watching_list_count":
		params := make(map[string]interface{})
		if userId, ok := args["userId"]; ok {
			params["userId"] = userId
		}
		data, err = s.backlogClient.makeRequest("GET", "/users/myself/watchings/count", params, nil)

	// Issue metadata tools
	case "get_issue_types":
		projectIdOrKey, ok := args["projectIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("projectIdOrKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/issueTypes", nil, nil)

	case "get_priorities":
		data, err = s.backlogClient.makeRequest("GET", "/priorities", nil, nil)

	case "get_resolutions":
		data, err = s.backlogClient.makeRequest("GET", "/resolutions", nil, nil)

	case "get_categories":
		projectIdOrKey, ok := args["projectIdOrKey"].(string)
		if !ok {
			return nil, fmt.Errorf("projectIdOrKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/categories", nil, nil)

	// Wiki tools
	case "get_wiki_pages":
		params := make(map[string]interface{})
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if keyword, ok := args["keyword"]; ok {
			params["keyword"] = keyword
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/wikis", params, nil)

	case "get_wikis_count":
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/wikis/count", nil, nil)

	case "get_wiki":
		wikiId, ok := args["wikiId"].(float64)
		if !ok {
			return nil, fmt.Errorf("wikiId is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/wikis/"+fmt.Sprintf("%.0f", wikiId), nil, nil)

	case "add_wiki":
		requiredFields := []string{"projectId", "name", "content"}
		for _, field := range requiredFields {
			if _, ok := args[field]; !ok {
				return nil, fmt.Errorf("%s is required", field)
			}
		}
		projectId := args["projectId"].(float64)
		delete(args, "projectId")
		data, err = s.backlogClient.makeRequest("POST", "/projects/"+fmt.Sprintf("%.0f", projectId)+"/wikis", nil, args)

	// Git & Pull Request tools
	case "get_git_repositories":
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/git/repositories", nil, nil)

	case "get_git_repository":
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName, nil, nil)

	case "get_pull_requests":
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		params := make(map[string]interface{})
		for key, value := range args {
			if key != "projectId" && key != "projectKey" && key != "repoId" && key != "repoName" {
				params[key] = value
			}
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests", params, nil)

	case "get_pull_requests_count":
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		params := make(map[string]interface{})
		for key, value := range args {
			if key != "projectId" && key != "projectKey" && key != "repoId" && key != "repoName" {
				params[key] = value
			}
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests/count", params, nil)

	case "get_pull_request":
		pullRequestId, ok := args["pullRequestId"].(float64)
		if !ok {
			return nil, fmt.Errorf("pullRequestId is required")
		}
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		}
		if projectIdOrKey != "" && repoIdOrName != "" {
			data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests/"+fmt.Sprintf("%.0f", pullRequestId), nil, nil)
		} else {
			data, err = s.backlogClient.makeRequest("GET", "/pullRequests/"+fmt.Sprintf("%.0f", pullRequestId), nil, nil)
		}

	case "add_pull_request":
		requiredFields := []string{"summary", "base", "branch"}
		for _, field := range requiredFields {
			if _, ok := args[field]; !ok {
				return nil, fmt.Errorf("%s is required", field)
			}
		}
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		delete(args, "projectId")
		delete(args, "projectKey")
		delete(args, "repoId")
		delete(args, "repoName")
		data, err = s.backlogClient.makeRequest("POST", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests", nil, args)

	case "update_pull_request":
		pullRequestId, ok := args["pullRequestId"].(float64)
		if !ok {
			return nil, fmt.Errorf("pullRequestId is required")
		}
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		delete(args, "projectId")
		delete(args, "projectKey")
		delete(args, "repoId")
		delete(args, "repoName")
		delete(args, "pullRequestId")
		data, err = s.backlogClient.makeRequest("PUT", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests/"+fmt.Sprintf("%.0f", pullRequestId), nil, args)

	case "get_pull_request_comments":
		pullRequestId, ok := args["pullRequestId"].(float64)
		if !ok {
			return nil, fmt.Errorf("pullRequestId is required")
		}
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		params := make(map[string]interface{})
		for key, value := range args {
			if key != "projectId" && key != "projectKey" && key != "repoId" && key != "repoName" && key != "pullRequestId" {
				params[key] = value
			}
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests/"+fmt.Sprintf("%.0f", pullRequestId)+"/comments", params, nil)

	case "add_pull_request_comment":
		pullRequestId, ok := args["pullRequestId"].(float64)
		if !ok {
			return nil, fmt.Errorf("pullRequestId is required")
		}
		if _, ok := args["content"]; !ok {
			return nil, fmt.Errorf("content is required")
		}
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		delete(args, "projectId")
		delete(args, "projectKey")
		delete(args, "repoId")
		delete(args, "repoName")
		delete(args, "pullRequestId")
		data, err = s.backlogClient.makeRequest("POST", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests/"+fmt.Sprintf("%.0f", pullRequestId)+"/comments", nil, args)

	case "update_pull_request_comment":
		pullRequestId, ok := args["pullRequestId"].(float64)
		if !ok {
			return nil, fmt.Errorf("pullRequestId is required")
		}
		commentId, ok := args["commentId"].(float64)
		if !ok {
			return nil, fmt.Errorf("commentId is required")
		}
		if _, ok := args["content"]; !ok {
			return nil, fmt.Errorf("content is required")
		}
		var projectIdOrKey, repoIdOrName string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		if repoId, ok := args["repoId"].(float64); ok {
			repoIdOrName = fmt.Sprintf("%.0f", repoId)
		} else if repoName, ok := args["repoName"].(string); ok {
			repoIdOrName = repoName
		} else {
			return nil, fmt.Errorf("either repoId or repoName is required")
		}
		delete(args, "projectId")
		delete(args, "projectKey")
		delete(args, "repoId")
		delete(args, "repoName")
		delete(args, "pullRequestId")
		delete(args, "commentId")
		data, err = s.backlogClient.makeRequest("PUT", "/projects/"+projectIdOrKey+"/git/repositories/"+repoIdOrName+"/pullRequests/"+fmt.Sprintf("%.0f", pullRequestId)+"/comments/"+fmt.Sprintf("%.0f", commentId), nil, args)

	// Document tools
	case "get_documents":
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		params := make(map[string]interface{})
		if path, ok := args["path"]; ok {
			params["path"] = path
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/files/metadata", params, nil)

	case "get_document_tree":
		var projectIdOrKey string
		if projectId, ok := args["projectId"].(float64); ok {
			projectIdOrKey = fmt.Sprintf("%.0f", projectId)
		} else if projectKey, ok := args["projectKey"].(string); ok {
			projectIdOrKey = projectKey
		} else {
			return nil, fmt.Errorf("either projectId or projectKey is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/projects/"+projectIdOrKey+"/files/metadata", nil, nil)

	case "get_document":
		documentId, ok := args["documentId"].(float64)
		if !ok {
			return nil, fmt.Errorf("documentId is required")
		}
		data, err = s.backlogClient.makeRequest("GET", "/files/"+fmt.Sprintf("%.0f", documentId), nil, nil)

	// Notifications tools
	case "get_notifications":
		params := make(map[string]interface{})
		for key, value := range args {
			params[key] = value
		}
		data, err = s.backlogClient.makeRequest("GET", "/notifications", params, nil)

	case "get_notifications_count":
		params := make(map[string]interface{})
		if alreadyRead, ok := args["alreadyRead"]; ok {
			params["alreadyRead"] = alreadyRead
		}
		data, err = s.backlogClient.makeRequest("GET", "/notifications/count", params, nil)

	case "reset_unread_notification_count":
		data, err = s.backlogClient.makeRequest("PUT", "/notifications/markAsRead", nil, nil)

	case "mark_notification_as_read":
		id, ok := args["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("id is required")
		}
		data, err = s.backlogClient.makeRequest("PUT", "/notifications/"+fmt.Sprintf("%.0f", id)+"/markAsRead", nil, nil)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	if err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		jsonData = []byte("{}")
	}

	return &CallToolResult{
		Content: []Content{{Type: "text", Text: string(jsonData)}},
	}, nil
}

// ==========================================
// HTTP Bridge
// ==========================================

type HTTPBridge struct {
	mcpServer *MCPServer
}

func NewHTTPBridge(mcpServer *MCPServer) *HTTPBridge {
	return &HTTPBridge{mcpServer: mcpServer}
}

func (h *HTTPBridge) handleMCPCall(c *gin.Context) {
	var req struct {
		Tool        string                 `json:"tool" binding:"required"`
		Args        map[string]interface{} `json:"args"`
		AccessToken string                 `json:"accessToken,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create MCP request
	mcpReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      func() *int64 { i := int64(1); return &i }(),
		Method:  "tools/call",
		Params: CallToolParams{
			Name:      req.Tool,
			Arguments: req.Args,
		},
	}

	// If AccessToken is provided, create temporary client
	if req.AccessToken != "" {
		domain := os.Getenv("BACKLOG_DOMAIN")
		tempClient, err := NewBacklogClient(domain, req.AccessToken, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tempServer := NewMCPServer(tempClient)
		resp := tempServer.HandleRequest(mcpReq)
		
		if resp.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error.Message, "code": resp.Error.Code})
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": resp.Result})
		return
	}

	// Use default server if it has a client, otherwise return error
	if h.mcpServer.backlogClient == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No credentials configured. Please provide accessToken in request or configure environment variables."})
		return
	}
	
	resp := h.mcpServer.HandleRequest(mcpReq)
	if resp.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error.Message, "code": resp.Error.Code})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": resp.Result})
}

// ==========================================
// Main Application
// ==========================================

func main() {
	// Get environment variables
	domain := os.Getenv("BACKLOG_DOMAIN")
	accessToken := os.Getenv("BACKLOG_ACCESS_TOKEN")
	apiKey := os.Getenv("BACKLOG_API_KEY")

	if domain == "" {
		log.Fatal("BACKLOG_DOMAIN environment variable is required")
	}

	// Allow startup without credentials when using OAuth mode
	// OAuth tokens will be provided dynamically via HTTP bridge

	// Check if running as MCP server (stdin mode) or HTTP bridge
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Running as MCP server via stdin/stdout
		runMCPServer(domain, accessToken, apiKey)
	} else {
		// Running as HTTP bridge
		runHTTPBridge(domain, accessToken, apiKey)
	}
}

func runMCPServer(domain, accessToken, apiKey string) {
	// Create Backlog client (may be nil for OAuth-only mode)
	var backlogClient *BacklogClient
	var err error
	
	if accessToken != "" || apiKey != "" {
		backlogClient, err = NewBacklogClient(domain, accessToken, apiKey)
		if err != nil {
			log.Fatal("Failed to create Backlog client:", err)
		}
	}

	// Create MCP server (handles nil client for OAuth-only mode)
	mcpServer := NewMCPServer(backlogClient)

	// Setup stdio transport
	scanner := bufio.NewScanner(os.Stdin)
	writer := os.Stdout

	log.Println("Backlog MCP Server (Golang) started")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var request MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		response := mcpServer.HandleRequest(request)

		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		fmt.Fprintf(writer, "%s\n", responseBytes)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading from stdin:", err)
	}
}

func runHTTPBridge(domain, accessToken, apiKey string) {
	// Create Backlog client (may be nil for OAuth-only mode)
	var backlogClient *BacklogClient
	var err error
	
	if accessToken != "" || apiKey != "" {
		backlogClient, err = NewBacklogClient(domain, accessToken, apiKey)
		if err != nil {
			log.Fatal("Failed to create Backlog client:", err)
		}
	}

	// Create MCP server and HTTP bridge (handles nil client for OAuth-only mode)
	mcpServer := NewMCPServer(backlogClient)
	bridge := NewHTTPBridge(mcpServer)

	// Setup Gin router
	r := gin.Default()
	r.POST("/mcp/call", bridge.handleMCPCall)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("Backlog MCP Server (Golang HTTP Bridge) starting on :3001")
	log.Fatal(http.ListenAndServe(":3001", r))
}