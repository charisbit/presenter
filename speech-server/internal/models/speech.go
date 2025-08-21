// Package models defines data structures for the Speech MCP Server.
// It includes request/response models for TTS operations, MCP protocol types,
// and voice/language information used throughout the speech synthesis system.
package models

import "time"

// SpeechRequest represents a text-to-speech synthesis request.
// It contains all parameters needed to generate speech audio from text
// using the configured TTS engines.
type SpeechRequest struct {
	Text     string  `json:"text" binding:"required"`     // Text content to synthesize into speech
	Language string  `json:"language" binding:"required"` // Target language code (ja, en, es, etc.)
	Voice    string  `json:"voice"`                       // Voice identifier or preference
	Speed    float32 `json:"speed"`                       // Speech speed multiplier (1.0 = normal)
}

// SpeechResponse represents the result of a text-to-speech synthesis operation.
// It provides the generated audio file information, metadata, and performance details.
type SpeechResponse struct {
	AudioURL  string        `json:"audioUrl"`  // URL path to the generated audio file
	Duration  time.Duration `json:"duration"`  // Estimated duration of the audio
	Language  string        `json:"language"`  // Language used for synthesis
	Voice     string        `json:"voice"`     // Voice used for synthesis
	CacheHit  bool          `json:"cacheHit"`  // Whether audio was served from cache
	RequestID string        `json:"requestId"` // Unique identifier for this request
}

// MCPRequest represents an MCP JSON-RPC request for speech operations.
// It follows the JSON-RPC 2.0 specification with MCP-specific extensions
// for speech synthesis tool calls and protocol methods.
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`        // JSON-RPC version (always "2.0")
	ID      interface{} `json:"id"`             // Request identifier for response matching
	Method  string      `json:"method"`         // MCP method name (tools/call, etc.)
	Params  interface{} `json:"params,omitempty"` // Method parameters (speech-specific)
}

// MCPResponse represents an MCP JSON-RPC response for speech operations.
// It contains either successful speech synthesis results or error information
// according to the JSON-RPC 2.0 specification.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`          // JSON-RPC version (always "2.0")
	ID      interface{} `json:"id"`               // Request identifier matching the request
	Result  interface{} `json:"result,omitempty"` // Successful speech operation result
	Error   *MCPError   `json:"error,omitempty"`  // Error information if operation failed
}

// MCPError represents an MCP protocol error for speech operations.
// It provides structured error information including standard JSON-RPC error codes
// and speech-specific error details for debugging.
type MCPError struct {
	Code    int         `json:"code"`             // Error code (following JSON-RPC error codes)
	Message string      `json:"message"`          // Human-readable error message
	Data    interface{} `json:"data,omitempty"`   // Additional error data (speech-specific)
}

// MCPTool represents an MCP tool definition
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// MCPToolResult represents the result of an MCP tool call
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in MCP responses
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}

// VoiceInfo represents available voice information from TTS engines.
// It provides metadata about voice characteristics, supported languages,
// and available synthesis styles for client voice selection.
type VoiceInfo struct {
	ID       string   `json:"id"`                // Unique voice identifier
	Name     string   `json:"name"`              // Human-readable voice name
	Language string   `json:"language"`          // Language code supported by this voice
	Gender   string   `json:"gender"`            // Voice gender (male, female, neutral)
	Styles   []string `json:"styles,omitempty"`  // Available speaking styles for this voice
}

// LanguageInfo represents available language information
type LanguageInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	NativeName  string `json:"nativeName"`
	Voices      int    `json:"voices"`
	Supported   bool   `json:"supported"`
}