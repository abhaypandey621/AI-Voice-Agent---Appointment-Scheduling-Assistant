package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/voice-agent/backend/internal/agent"
	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a WebSocket client connection
type Client struct {
	conn  *websocket.Conn
	agent *agent.VoiceAgent
	send  chan []byte
	done  chan struct{}
	mu    sync.Mutex
}

// Manager manages WebSocket connections
type Manager struct {
	clients map[string]*Client
	config  *config.Config
	mu      sync.RWMutex
}

// NewManager creates a new WebSocket manager
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		clients: make(map[string]*Client),
		config:  cfg,
	}
}

// HandleConnection handles a new WebSocket connection
func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		roomName = fmt.Sprintf("room-%d", time.Now().UnixNano())
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	// Create agent with callbacks
	voiceAgent, err := agent.NewVoiceAgent(m.config, roomName, &agent.AgentConfig{
		OnTranscript: func(text string, isFinal bool) {
			client.sendMessage(models.WSMessage{
				Type: models.WSTypeTranscript,
				Payload: map[string]interface{}{
					"text":     text,
					"is_final": isFinal,
				},
			})
		},
		OnAgentResponse: func(text string) {
			client.sendMessage(models.WSMessage{
				Type:    models.WSTypeAgentResponse,
				Payload: text,
			})
		},
		OnToolCall: func(payload models.ToolCallPayload) {
			client.sendMessage(models.WSMessage{
				Type:    models.WSTypeToolCall,
				Payload: payload,
			})
		},
		OnToolResult: func(payload models.ToolResultPayload) {
			client.sendMessage(models.WSMessage{
				Type:    models.WSTypeToolResult,
				Payload: payload,
			})
		},
		OnAudioOutput: func(audio []byte) {
			// Send audio as binary message
			client.mu.Lock()
			defer client.mu.Unlock()
			if client.conn != nil {
				_ = client.conn.WriteMessage(websocket.BinaryMessage, audio)
			}
		},
		OnCallEnd: func(summary *models.CallSummary, cost *models.CostBreakdown) {
			log.Printf("[WebSocket] OnCallEnd callback triggered")
			log.Printf("[WebSocket] Summary: %+v", summary)
			log.Printf("[WebSocket] Cost: %+v", cost)

			summaryMsg := models.WSMessage{
				Type: models.WSTypeCallSummary,
				Payload: map[string]interface{}{
					"summary": summary,
					"cost":    cost,
				},
			}
			log.Printf("[WebSocket] Sending call_summary message")
			client.sendMessage(summaryMsg)

			endMsg := models.WSMessage{
				Type:    models.WSTypeCallEnd,
				Payload: "Call ended",
			}
			log.Printf("[WebSocket] Sending call_end message")
			client.sendMessage(endMsg)
			log.Printf("[WebSocket] OnCallEnd completed")
		},
		OnError: func(err error) {
			client.sendMessage(models.WSMessage{
				Type:    models.WSTypeError,
				Payload: err.Error(),
			})
		},
	})

	if err != nil {
		conn.WriteJSON(models.WSMessage{
			Type:    models.WSTypeError,
			Payload: fmt.Sprintf("Failed to create agent: %v", err),
		})
		conn.Close()
		return
	}

	client.agent = voiceAgent

	// Register client
	m.mu.Lock()
	m.clients[voiceAgent.ID] = client
	m.mu.Unlock()

	// Start agent
	if err := voiceAgent.Start(); err != nil {
		conn.WriteJSON(models.WSMessage{
			Type:    models.WSTypeError,
			Payload: fmt.Sprintf("Failed to start agent: %v", err),
		})
		conn.Close()
		return
	}

	// Send connection success
	client.sendMessage(models.WSMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"agent_id":  voiceAgent.ID,
			"room_name": roomName,
		},
	})

	// Start goroutines
	go client.writePump()
	go client.readPump(m)
}

func (c *Client) readPump(m *Manager) {
	defer func() {
		c.cleanup(m)
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		switch messageType {
		case websocket.BinaryMessage:
			// Audio data for STT
			if c.agent != nil {
				if err := c.agent.SendAudio(message); err != nil {
					c.sendMessage(models.WSMessage{
						Type:    models.WSTypeError,
						Payload: fmt.Sprintf("Audio processing error: %v", err),
					})
				}
			}

		case websocket.TextMessage:
			// Control messages
			var msg clientMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "text_input":
				// Direct text input (for testing without audio)
				if text, ok := msg.Payload.(string); ok && c.agent != nil {
					// Validate text input - reject null/empty values
					cleanText := strings.TrimSpace(text)
					if cleanText == "" || cleanText == "null" || cleanText == "undefined" {
						c.sendMessage(models.WSMessage{
							Type:    models.WSTypeError,
							Payload: "Invalid input: Please provide valid text",
						})
						continue
					}
					log.Printf("Received text input: %s", cleanText)
					c.agent.ProcessTextInput(cleanText)
				} else {
					log.Printf("Invalid text input payload: %v", msg.Payload)
					c.sendMessage(models.WSMessage{
						Type:    models.WSTypeError,
						Payload: "Invalid text input format",
					})
				}

			case "end_call":
				log.Printf("[WebSocket] Received end_call request")
				if c.agent != nil {
					log.Printf("[WebSocket] Calling agent.EndCall()")
					c.agent.EndCall()
					log.Printf("[WebSocket] agent.EndCall() completed")
				} else {
					log.Printf("[WebSocket] WARNING: Agent is nil, cannot end call")
				}

			case "get_session":
				if c.agent != nil {
					session := c.agent.GetSession()
					c.sendMessage(models.WSMessage{
						Type:    "session",
						Payload: session,
					})
				}

			case "ping":
				c.sendMessage(models.WSMessage{
					Type:    "pong",
					Payload: time.Now().UnixMilli(),
				})
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.mu.Unlock()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			if conn == nil {
				return
			}
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			if conn == nil {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *Client) sendMessage(msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
		// Channel full, drop message
	}
}

func (c *Client) cleanup(m *Manager) {
	close(c.done)

	if c.agent != nil {
		// Remove from manager
		m.mu.Lock()
		delete(m.clients, c.agent.ID)
		m.mu.Unlock()

		// Stop agent
		c.agent.Stop()
	}

	c.mu.Lock()
	c.conn.Close()
	c.conn = nil
	c.mu.Unlock()
}

// GetClient returns a client by agent ID
func (m *Manager) GetClient(agentID string) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[agentID]
}

// GetActiveConnections returns count of active connections
func (m *Manager) GetActiveConnections() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

type clientMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
