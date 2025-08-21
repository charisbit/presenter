package handlers

import (
	"fmt"
	"net/http"
	"sync"

	"intelligent-presenter-backend/internal/models"
	"intelligent-presenter-backend/internal/services"
	"intelligent-presenter-backend/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type SlideHandler struct {
	config         *config.Config
	slideService   *services.SlideService
	activeSlides   map[string]*SlideSession
	slidesMutex    sync.RWMutex
	wsUpgrader     websocket.Upgrader
}

type SlideSession struct {
	ID          string
	ProjectID   models.ProjectID
	Themes      []models.SlideTheme
	Language    string
	Status      string
	Connections map[*websocket.Conn]bool
	ConnMutex   sync.RWMutex
	// Store generated slides data
	Slides      []*models.SlideContent    `json:"slides"`
	Narrations  []*models.SlideNarration  `json:"narrations"`
	AudioFiles  []*models.SlideAudio      `json:"audioFiles"`
}

func NewSlideHandler(cfg *config.Config) *SlideHandler {
	return &SlideHandler{
		config:       cfg,
		slideService: services.NewSlideService(cfg),
		activeSlides: make(map[string]*SlideSession),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

func (h *SlideHandler) GenerateSlides(c *gin.Context) {
	var req models.SlideGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	fmt.Printf("Received request: ProjectID=%s, Language=%s, Themes=%v\n", req.ProjectID, req.Language, req.Themes)

	// Validate themes
	if len(req.Themes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one theme must be specified",
		})
		return
	}

	// Generate unique slide ID
	slideID := uuid.New().String()

	// Create slide session
	session := &SlideSession{
		ID:          slideID,
		ProjectID:   req.ProjectID,
		Themes:      req.Themes,
		Language:    req.Language,
		Status:      "generating",
		Connections: make(map[*websocket.Conn]bool),
		Slides:      make([]*models.SlideContent, 0),
		Narrations:  make([]*models.SlideNarration, 0),
		AudioFiles:  make([]*models.SlideAudio, 0),
	}

	h.slidesMutex.Lock()
	h.activeSlides[slideID] = session
	h.slidesMutex.Unlock()

	// Start slide generation in background
	go h.generateSlidesAsync(session, c.GetInt("userID"), c.GetString("backlogToken"))

	// Return response
	c.JSON(http.StatusOK, models.SlideGenerationResponse{
		SlideID:      slideID,
		Status:       "generating",
		WebSocketURL: fmt.Sprintf("ws://localhost:%s/ws/slides/%s", h.config.Port, slideID),
	})
}

func (h *SlideHandler) GetSlideStatus(c *gin.Context) {
	slideID := c.Param("slideId")

	h.slidesMutex.RLock()
	session, exists := h.activeSlides[slideID]
	h.slidesMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Slide not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slideId":    session.ID,
		"projectId":  session.ProjectID,
		"status":     session.Status,
		"themes":     session.Themes,
		"slides":     session.Slides,
		"narrations": session.Narrations,
		"audioFiles": session.AudioFiles,
	})
}

func (h *SlideHandler) HandleWebSocket(c *gin.Context) {
	slideID := c.Param("slideId")

	h.slidesMutex.RLock()
	session, exists := h.activeSlides[slideID]
	h.slidesMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Slide not found",
		})
		return
	}

	conn, err := h.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upgrade to WebSocket",
		})
		return
	}
	defer conn.Close()

	// Add connection to session
	session.ConnMutex.Lock()
	session.Connections[conn] = true
	session.ConnMutex.Unlock()

	// Remove connection when done
	defer func() {
		session.ConnMutex.Lock()
		delete(session.Connections, conn)
		session.ConnMutex.Unlock()
	}()

	// Keep connection alive and handle messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *SlideHandler) generateSlidesAsync(session *SlideSession, userID int, backlogToken string) {
	defer func() {
		session.Status = "completed"
	}()

	for i, theme := range session.Themes {
		// Broadcast slide generation started
		h.broadcastSlideGenerationStarted(session, &models.SlideGenerationStarted{
			SlideIndex: i,
			Theme:      theme,
		})

		// Generate slide content
		slideContent, err := h.slideService.GenerateSlideContent(
			session.ProjectID.String(),
			theme,
			session.Language,
			backlogToken,
		)
		if err != nil {
			h.broadcastError(session, fmt.Sprintf("Failed to generate slide %d: %v", i+1, err))
			continue
		}

		slideContent.Index = i
		// Store slide data in session
		session.Slides = append(session.Slides, slideContent)
		h.broadcastSlideContent(session, slideContent)

		// Generate narration
		narration, err := h.slideService.GenerateSlideNarration(slideContent, session.Language)
		if err != nil {
			h.broadcastError(session, fmt.Sprintf("Failed to generate narration for slide %d: %v", i+1, err))
		} else {
			// Store narration data in session
			session.Narrations = append(session.Narrations, narration)
			h.broadcastSlideNarration(session, narration)
			
			// Generate audio for the narration
			audio, err := h.slideService.GenerateSlideAudio(narration)
			if err != nil {
				h.broadcastError(session, fmt.Sprintf("Failed to generate audio for slide %d: %v", i+1, err))
			} else {
				// Store audio data in session
				session.AudioFiles = append(session.AudioFiles, audio)
				h.broadcastSlideAudio(session, audio)
			}
		}
	}

	// Send completion message
	h.broadcastPresentationComplete(session, &models.PresentationComplete{
		TotalSlides: len(session.Themes),
		Duration:    "Generated successfully",
	})
}

func (h *SlideHandler) broadcastSlideGenerationStarted(session *SlideSession, started *models.SlideGenerationStarted) {
	message := models.WebSocketMessage{
		Type: models.MessageTypeSlideGenerationStarted,
		Data: started,
	}
	h.broadcastToSession(session, message)
}

func (h *SlideHandler) broadcastSlideContent(session *SlideSession, content *models.SlideContent) {
	message := models.WebSocketMessage{
		Type: models.MessageTypeSlideContent,
		Data: content,
	}
	h.broadcastToSession(session, message)
}

func (h *SlideHandler) broadcastSlideNarration(session *SlideSession, narration *models.SlideNarration) {
	message := models.WebSocketMessage{
		Type: models.MessageTypeSlideNarration,
		Data: narration,
	}
	h.broadcastToSession(session, message)
}

func (h *SlideHandler) broadcastSlideAudio(session *SlideSession, audio *models.SlideAudio) {
	message := models.WebSocketMessage{
		Type: models.MessageTypeSlideAudio,
		Data: audio,
	}
	h.broadcastToSession(session, message)
}

func (h *SlideHandler) broadcastPresentationComplete(session *SlideSession, complete *models.PresentationComplete) {
	message := models.WebSocketMessage{
		Type: models.MessageTypePresentationComplete,
		Data: complete,
	}
	h.broadcastToSession(session, message)
}

func (h *SlideHandler) broadcastError(session *SlideSession, errMsg string) {
	message := models.WebSocketMessage{
		Type: models.MessageTypeError,
		Data: models.ErrorMessage{
			Message: errMsg,
			Code:    "GENERATION_ERROR",
		},
	}
	h.broadcastToSession(session, message)
}

func (h *SlideHandler) broadcastToSession(session *SlideSession, message models.WebSocketMessage) {
	session.ConnMutex.RLock()
	defer session.ConnMutex.RUnlock()

	for conn := range session.Connections {
		if err := conn.WriteJSON(message); err != nil {
			// Remove failed connection
			go func(c *websocket.Conn) {
				session.ConnMutex.Lock()
				delete(session.Connections, c)
				session.ConnMutex.Unlock()
				c.Close()
			}(conn)
		}
	}
}