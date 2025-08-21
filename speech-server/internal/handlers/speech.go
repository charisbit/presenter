package handlers

import (
	"encoding/json"
	"net/http"

	"speech-mcp-server/internal/models"
	"speech-mcp-server/internal/services"
	"speech-mcp-server/pkg/config"

	"github.com/gin-gonic/gin"
)

type SpeechHandler struct {
	config     *config.Config
	ttsService *services.TTSService
}

func NewSpeechHandler(cfg *config.Config) *SpeechHandler {
	return &SpeechHandler{
		config:     cfg,
		ttsService: services.NewTTSService(cfg),
	}
}

func (h *SpeechHandler) SynthesizeSpeech(c *gin.Context) {
	var req models.SpeechRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	resp, err := h.ttsService.SynthesizeSpeech(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *SpeechHandler) ServeAudioFile(c *gin.Context) {
	filename := c.Param("filename")
	c.File(h.config.CacheDir + "/" + filename)
}

func (h *SpeechHandler) ListVoices(c *gin.Context) {
	c.JSON(http.StatusOK, h.ttsService.GetAvailableVoices())
}

func (h *SpeechHandler) ListLanguages(c *gin.Context) {
	c.JSON(http.StatusOK, h.ttsService.GetSupportedLanguages())
}

func (h *SpeechHandler) HandleMCPRequest(c *gin.Context) {
	var req models.MCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid MCP request"})
		return
	}

	// For now, only support synthesize
	if req.Method != "synthesize" {
		c.JSON(http.StatusNotImplemented, models.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &models.MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		})
		return
	}

	// We need to parse the params into a SpeechRequest
	var params models.SpeechRequest
	// This is a bit of a hack, but it works for now
	// A better solution would be to use a proper JSON-RPC library
	data, _ := json.Marshal(req.Params)
	json.Unmarshal(data, &params)

	resp, err := h.ttsService.SynthesizeSpeech(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &models.MCPError{
				Code:    -32000,
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resp,
	})
}

func (h *SpeechHandler) GetCapabilities(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"capabilities": []string{"synthesize", "list_voices", "list_languages"},
	})
}