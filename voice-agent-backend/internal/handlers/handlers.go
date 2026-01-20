package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/database"
	"github.com/voice-agent/backend/internal/models"
	"github.com/voice-agent/backend/internal/services/avatar"
	"github.com/voice-agent/backend/internal/services/livekit"
	"github.com/voice-agent/backend/internal/websocket"
)

// Handler holds all HTTP handlers
type Handler struct {
	config         *config.Config
	livekitService *livekit.Service
	avatarService  *avatar.Service
	wsManager      *websocket.Manager
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config, lkService *livekit.Service, avService *avatar.Service, wsManager *websocket.Manager) *Handler {
	return &Handler{
		config:         cfg,
		livekitService: lkService,
		avatarService:  avService,
		wsManager:      wsManager,
	}
}

// HealthCheck returns service health status
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// CreateRoom creates a new LiveKit room and returns connection details
func (h *Handler) CreateRoom(c *gin.Context) {
	var req struct {
		RoomName string `json:"room_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Generate room name if not provided
		req.RoomName = fmt.Sprintf("room-%d", time.Now().UnixNano())
	}

	if req.RoomName == "" {
		req.RoomName = fmt.Sprintf("room-%d", time.Now().UnixNano())
	}

	// Create room in LiveKit
	room, err := h.livekitService.CreateRoom(c.Request.Context(), req.RoomName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create room: %v", err),
		})
		return
	}

	// Generate participant token
	participantName := fmt.Sprintf("user-%s", uuid.New().String()[:8])
	token, err := h.livekitService.GenerateToken(req.RoomName, participantName, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to generate token: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room_name":        room.Name,
		"token":            token,
		"url":              h.livekitService.GetURL(),
		"participant_name": participantName,
	})
}

// GetToken generates a token for joining a room
func (h *Handler) GetToken(c *gin.Context) {
	roomName := c.Query("room")
	participantName := c.Query("participant")

	if roomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room parameter is required",
		})
		return
	}

	if participantName == "" {
		participantName = fmt.Sprintf("user-%s", uuid.New().String()[:8])
	}

	token, err := h.livekitService.GenerateToken(roomName, participantName, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to generate token: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.TokenResponse{
		Token:    token,
		RoomName: roomName,
		URL:      h.livekitService.GetURL(),
	})
}

// CreateAvatarSession creates a new avatar conversation session
func (h *Handler) CreateAvatarSession(c *gin.Context) {
	var req struct {
		ReplicaID   string `json:"replica_id"`
		CallbackURL string `json:"callback_url"`
	}

	_ = c.ShouldBindJSON(&req)

	session, err := h.avatarService.CreateConversation(req.ReplicaID, req.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create avatar session: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// EndAvatarSession ends an avatar conversation session
func (h *Handler) EndAvatarSession(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "conversation_id is required",
		})
		return
	}

	if err := h.avatarService.EndConversation(conversationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to end avatar session: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ended",
	})
}

// GetAvatarReplicas lists available avatar replicas
func (h *Handler) GetAvatarReplicas(c *gin.Context) {
	replicas, err := h.avatarService.ListReplicas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list replicas: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"replicas": replicas,
	})
}

// GetAppointments gets appointments for a user
func (h *Handler) GetAppointments(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone parameter is required",
		})
		return
	}

	appointments, err := database.DB.GetAppointmentsByPhone(phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get appointments: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"count":        len(appointments),
	})
}

// GetAvailableSlots gets available appointment slots for a date
func (h *Handler) GetAvailableSlots(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "date parameter is required (format: YYYY-MM-DD)",
		})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Generate slots (9 AM to 5 PM)
	slots := []gin.H{}
	loc := time.Local

	for hour := 9; hour < 17; hour++ {
		for _, minute := range []int{0, 30} {
			slotTime := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), hour, minute, 0, 0, loc)

			if slotTime.Before(time.Now()) {
				continue
			}

			available, _ := database.DB.CheckSlotAvailability(slotTime, 30)

			slots = append(slots, gin.H{
				"date_time":  slotTime.Format(time.RFC3339),
				"time":       slotTime.Format("3:04 PM"),
				"available":  available,
				"duration":   30,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  date,
		"slots": slots,
	})
}

// GetCallSummaries gets call summaries for a user
func (h *Handler) GetCallSummaries(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone parameter is required",
		})
		return
	}

	summaries, err := database.DB.GetCallSummariesByPhone(phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get call summaries: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries": summaries,
		"count":     len(summaries),
	})
}

// WebSocketHandler handles WebSocket upgrade
func (h *Handler) WebSocketHandler(c *gin.Context) {
	h.wsManager.HandleConnection(c.Writer, c.Request)
}

// GetStats returns server statistics
func (h *Handler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"active_connections": h.wsManager.GetActiveConnections(),
		"timestamp":          time.Now().Format(time.RFC3339),
	})
}
