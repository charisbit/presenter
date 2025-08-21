package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"intelligent-presenter-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BacklogMCPWrapper wraps the stdio Backlog MCP Server as an HTTP service
type BacklogMCPWrapper struct {
	config      *config.Config
	process     *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	scanner     *bufio.Scanner
	requestID   int64
	sessions    map[string]*MCPSession
	sessionMux  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isRunning   bool
}

type MCPSession struct {
	ID        string
	responses map[int64]chan *MCPResponse
	respMutex sync.RWMutex
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewBacklogMCPWrapper(cfg *config.Config) *BacklogMCPWrapper {
	ctx, cancel := context.WithCancel(context.Background())
	return &BacklogMCPWrapper{
		config:   cfg,
		sessions: make(map[string]*MCPSession),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (w *BacklogMCPWrapper) Start() error {
	if w.isRunning {
		return nil
	}

	// In Docker environment, we don't start the process but mark as running
	// The external backlog-mcp-server container handles the MCP communication
	w.isRunning = true

	log.Printf("Backlog MCP Wrapper marked as started (using external container)")
	return nil
}

func (w *BacklogMCPWrapper) Stop() error {
	w.cancel()
	w.isRunning = false
	
	if w.stdin != nil {
		w.stdin.Close()
	}
	if w.stdout != nil {
		w.stdout.Close()
	}
	if w.process != nil {
		w.process.Process.Kill()
		w.process.Wait()
	}
	return nil
}

func (w *BacklogMCPWrapper) initialize() error {
	// Create a temporary session for initialization
	session := &MCPSession{
		ID:        "init",
		responses: make(map[int64]chan *MCPResponse),
	}
	
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": false,
			},
		},
		"clientInfo": map[string]interface{}{
			"name":    "intelligent-presenter-backend",
			"version": "1.0.0",
		},
	}

	_, err := w.sendRequest(session, "initialize", initParams)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	// Send initialized notification
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}

	return w.sendMessage(notification)
}

func (w *BacklogMCPWrapper) handleMessages() {
	defer w.Stop()

	for w.scanner.Scan() {
		line := w.scanner.Text()
		if line == "" {
			continue
		}

		var response MCPResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			log.Printf("Failed to parse MCP response: %v, line: %s", err, line)
			continue
		}

		// Broadcast to all sessions
		w.sessionMux.RLock()
		for _, session := range w.sessions {
			session.respMutex.RLock()
			if ch, ok := session.responses[response.ID]; ok {
				select {
				case ch <- &response:
				default:
				}
				delete(session.responses, response.ID)
			}
			session.respMutex.RUnlock()
		}
		w.sessionMux.RUnlock()
	}
}

func (w *BacklogMCPWrapper) sendRequest(session *MCPSession, method string, params interface{}) (json.RawMessage, error) {
	if !w.isRunning {
		return nil, fmt.Errorf("MCP wrapper is not running")
	}

	id := atomic.AddInt64(&w.requestID, 1)
	
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	respCh := make(chan *MCPResponse, 1)
	session.respMutex.Lock()
	session.responses[id] = respCh
	session.respMutex.Unlock()

	// Send request
	if err := w.sendMessage(request); err != nil {
		session.respMutex.Lock()
		delete(session.responses, id)
		session.respMutex.Unlock()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case response := <-respCh:
		if response.Error != nil {
			return nil, fmt.Errorf("MCP error %d: %s", response.Error.Code, response.Error.Message)
		}
		return response.Result, nil
	case <-time.After(30 * time.Second):
		session.respMutex.Lock()
		delete(session.responses, id)
		session.respMutex.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

func (w *BacklogMCPWrapper) sendMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = w.stdin.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	return nil
}

// HTTP Handlers for MCP over HTTP

func (w *BacklogMCPWrapper) HandleHTTP(c *gin.Context) {
	sessionID := c.GetHeader("Mcp-Session-Id")
	if sessionID == "" {
		sessionID = uuid.New().String()
		c.Header("Mcp-Session-Id", sessionID)
	}

	// Get or create session
	w.sessionMux.Lock()
	session, exists := w.sessions[sessionID]
	if !exists {
		session = &MCPSession{
			ID:        sessionID,
			responses: make(map[int64]chan *MCPResponse),
		}
		w.sessions[sessionID] = session
	}
	w.sessionMux.Unlock()

	var request MCPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	result, err := w.sendRequest(session, request.Method, request.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	c.JSON(http.StatusOK, response)
}

func (w *BacklogMCPWrapper) HandleCloseSession(c *gin.Context) {
	sessionID := c.GetHeader("Mcp-Session-Id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session ID"})
		return
	}

	w.sessionMux.Lock()
	delete(w.sessions, sessionID)
	w.sessionMux.Unlock()

	c.JSON(http.StatusOK, gin.H{"status": "closed"})
}