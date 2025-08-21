// Package mcp provides Model Context Protocol (MCP) client implementation
// for communicating with MCP servers. This package enables the intelligent
// presenter to interact with external data sources and services through
// the standardized MCP protocol.
//
// The MCP client supports JSON-RPC 2.0 communication with proper session
// management, tool invocation, resource access, and prompt handling.
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MCPClient represents an MCP client for communicating with MCP servers.
// It manages the HTTP connection, session state, and JSON-RPC protocol
// communication with remote MCP servers.
type MCPClient struct {
	serverURL string       // Base URL of the MCP server
	client    *http.Client // HTTP client for network requests
	sessionID string       // Session identifier for stateful connections
}

// MCPRequest represents an MCP JSON-RPC request structure.
// It follows the JSON-RPC 2.0 specification with MCP-specific extensions
// for method calls and parameter passing.
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`        // JSON-RPC version (always "2.0")
	ID      interface{} `json:"id"`             // Request identifier for response matching
	Method  string      `json:"method"`         // MCP method name to invoke
	Params  interface{} `json:"params,omitempty"` // Method parameters (optional)
}

// MCPResponse represents an MCP JSON-RPC response structure.
// It contains either a successful result or error information
// according to the JSON-RPC 2.0 specification.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`          // JSON-RPC version (always "2.0")
	ID      interface{}     `json:"id"`               // Request identifier matching the request
	Result  json.RawMessage `json:"result,omitempty"` // Successful result data (optional)
	Error   *MCPError       `json:"error,omitempty"`  // Error information (optional)
}

// MCPError represents an MCP error response.
// It provides structured error information including error codes,
// human-readable messages, and optional additional data.
type MCPError struct {
	Code    int         `json:"code"`             // Error code (following JSON-RPC error codes)
	Message string      `json:"message"`          // Human-readable error message
	Data    interface{} `json:"data,omitempty"`   // Additional error data (optional)
}

// NewMCPClient creates a new MCP client instance for the specified server.
// It initializes an HTTP client with appropriate timeout settings
// for reliable communication with the MCP server.
//
// Parameters:
//   - serverURL: The base URL of the MCP server to connect to
//
// Returns a configured MCPClient ready for use.
func NewMCPClient(serverURL string) *MCPClient {
	return &MCPClient{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Initialize establishes connection with the MCP server using the MCP handshake protocol.
// This method sends the initialization request with client capabilities and protocol version,
// then follows up with an initialized notification to complete the handshake.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - clientInfo: Client identification information (name, version, etc.)
//
// Returns an error if the initialization handshake fails at any step.
func (c *MCPClient) Initialize(ctx context.Context, clientInfo map[string]interface{}) error {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo":      clientInfo,
		},
	}

	response, err := c.sendRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP client: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("MCP initialization error: %s", response.Error.Message)
	}

	// Send initialized notification
	notification := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	_, err = c.sendRequest(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	return nil
}

// CallTool invokes a tool on the MCP server with the specified arguments.
// Tools are server-side functions that can perform operations like data retrieval,
// computation, or external API calls.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//   - name: The name of the tool to invoke
//   - arguments: Key-value pairs of arguments to pass to the tool
//
// Returns:
//   - *MCPResponse: The tool execution result or error
//   - error: Any communication or protocol error that occurred
func (c *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.generateID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendRequest(ctx, request)
}

// ListTools retrieves the list of available tools from the MCP server.
// This method queries the server for all tools that can be invoked,
// including their names, descriptions, and parameter schemas.
//
// Parameters:
//   - ctx: Context for request timeout and cancellation
//
// Returns:
//   - *MCPResponse: List of available tools with their metadata
//   - error: Any communication or protocol error that occurred
func (c *MCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.generateID(),
		Method:  "tools/list",
	}

	return c.sendRequest(ctx, request)
}

// ReadResource reads a resource from the MCP server
func (c *MCPClient) ReadResource(ctx context.Context, uri string) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.generateID(),
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	return c.sendRequest(ctx, request)
}

// ListResources lists available resources from the MCP server
func (c *MCPClient) ListResources(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.generateID(),
		Method:  "resources/list",
	}

	return c.sendRequest(ctx, request)
}

// GetPrompt gets a prompt from the MCP server
func (c *MCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.generateID(),
		Method:  "prompts/get",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": arguments,
		},
	}

	return c.sendRequest(ctx, request)
}

// ListPrompts lists available prompts from the MCP server
func (c *MCPClient) ListPrompts(ctx context.Context) (*MCPResponse, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.generateID(),
		Method:  "prompts/list",
	}

	return c.sendRequest(ctx, request)
}

// sendRequest sends an MCP request to the server
func (c *MCPClient) sendRequest(ctx context.Context, request MCPRequest) (*MCPResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Extract session ID from response headers
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		c.sessionID = sessionID
	}

	var mcpResponse MCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &mcpResponse, nil
}

// generateID generates a unique ID for requests
func (c *MCPClient) generateID() int64 {
	return time.Now().UnixNano()
}

// Close closes the MCP client connection
func (c *MCPClient) Close(ctx context.Context) error {
	if c.sessionID == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create close request: %w", err)
	}

	req.Header.Set("Mcp-Session-Id", c.sessionID)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send close request: %w", err)
	}
	defer resp.Body.Close()

	c.sessionID = ""
	return nil
}