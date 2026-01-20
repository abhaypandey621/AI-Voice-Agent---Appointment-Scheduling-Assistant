package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/database"
	"github.com/voice-agent/backend/internal/handlers"
	"github.com/voice-agent/backend/internal/middleware"
	"github.com/voice-agent/backend/internal/services/avatar"
	"github.com/voice-agent/backend/internal/services/livekit"
	"github.com/voice-agent/backend/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	if err := database.Initialize(cfg); err != nil {
		log.Printf("Warning: Failed to initialize database: %v", err)
	}

	// Initialize services
	livekitService := livekit.NewService(cfg)
	avatarService := avatar.NewService(cfg)
	wsManager := websocket.NewManager(cfg)

	// Initialize handlers
	h := handlers.NewHandler(cfg, livekitService, avatarService, wsManager)

	// Setup router
	router := setupRouter(h)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func setupRouter(h *handlers.Handler) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(gin.Logger())

	// Health check
	router.GET("/health", h.HealthCheck)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    "Voice Agent API",
			"version": "1.0.0",
			"docs":    "/api/docs",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Room management
		api.POST("/rooms", h.CreateRoom)
		api.GET("/token", h.GetToken)

		// Avatar sessions
		api.POST("/avatar/session", h.CreateAvatarSession)
		api.POST("/avatar/session/:id/end", h.EndAvatarSession)
		api.GET("/avatar/replicas", h.GetAvatarReplicas)

		// Appointments
		api.GET("/appointments", h.GetAppointments)
		api.GET("/slots", h.GetAvailableSlots)

		// Call summaries
		api.GET("/summaries", h.GetCallSummaries)

		// Stats
		api.GET("/stats", h.GetStats)
	}

	// WebSocket endpoint
	router.GET("/ws", h.WebSocketHandler)

	// API documentation
	router.GET("/api/docs", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"endpoints": []gin.H{
				{"method": "GET", "path": "/health", "description": "Health check"},
				{"method": "POST", "path": "/api/rooms", "description": "Create a new room"},
				{"method": "GET", "path": "/api/token", "description": "Get access token for a room"},
				{"method": "POST", "path": "/api/avatar/session", "description": "Create avatar session"},
				{"method": "POST", "path": "/api/avatar/session/:id/end", "description": "End avatar session"},
				{"method": "GET", "path": "/api/avatar/replicas", "description": "List available avatar replicas"},
				{"method": "GET", "path": "/api/appointments", "description": "Get appointments by phone"},
				{"method": "GET", "path": "/api/slots", "description": "Get available slots for a date"},
				{"method": "GET", "path": "/api/summaries", "description": "Get call summaries by phone"},
				{"method": "GET", "path": "/api/stats", "description": "Get server statistics"},
				{"method": "GET", "path": "/ws", "description": "WebSocket endpoint for voice agent"},
			},
			"websocket": gin.H{
				"url": "/ws",
				"messages": gin.H{
					"incoming": []string{
						"binary: Audio data for STT",
						"text_input: Direct text input for testing",
						"end_call: End the current call",
						"get_session: Get current session state",
						"ping: Health check",
					},
					"outgoing": []string{
						"connected: Connection established",
						"transcript: STT transcription result",
						"agent_response: Agent text response",
						"tool_call: Tool being executed",
						"tool_result: Tool execution result",
						"call_summary: Call summary at end",
						"call_end: Call ended notification",
						"error: Error message",
						"binary: TTS audio output",
					},
				},
			},
		})
	})

	return router
}
