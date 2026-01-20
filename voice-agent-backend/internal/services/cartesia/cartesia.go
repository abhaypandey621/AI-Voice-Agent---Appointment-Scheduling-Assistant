package cartesia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/voice-agent/backend/internal/config"
)

const (
	cartesiaAPIURL = "https://api.cartesia.ai/tts/bytes"
	cartesiaWSURL  = "wss://api.cartesia.ai/tts/websocket"
)

// Service handles Cartesia TTS operations
type Service struct {
	apiKey          string
	voiceID         string
	totalCharacters int
	mu              sync.Mutex
}

// StreamingClient handles real-time TTS
type StreamingClient struct {
	conn        *websocket.Conn
	onAudio     func([]byte)
	onComplete  func()
	onError     func(error)
	done        chan struct{}
	service     *Service
}

// NewService creates a new Cartesia service
func NewService(cfg *config.Config) *Service {
	return &Service{
		apiKey:  cfg.CartesiaAPIKey,
		voiceID: cfg.CartesiaVoiceID,
	}
}

// SynthesizeSpeech converts text to speech (REST API)
func (s *Service) SynthesizeSpeech(text string) ([]byte, error) {
	s.mu.Lock()
	s.totalCharacters += len(text)
	s.mu.Unlock()

	reqBody := map[string]interface{}{
		"transcript": text,
		"model_id":   "sonic-english",
		"voice": map[string]interface{}{
			"mode": "id",
			"id":   s.voiceID,
		},
		"output_format": map[string]interface{}{
			"container": "raw",
			"encoding":  "pcm_s16le",
			"sample_rate": 24000,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", cartesiaAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.apiKey)
	req.Header.Set("Cartesia-Version", "2024-06-10")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cartesia error (status %d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// NewStreamingClient creates a real-time TTS client
func (s *Service) NewStreamingClient(onAudio func([]byte), onComplete func(), onError func(error)) (*StreamingClient, error) {
	header := http.Header{}
	header.Set("X-API-Key", s.apiKey)
	header.Set("Cartesia-Version", "2024-06-10")

	conn, _, err := websocket.DefaultDialer.Dial(cartesiaWSURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cartesia: %w", err)
	}

	client := &StreamingClient{
		conn:       conn,
		onAudio:    onAudio,
		onComplete: onComplete,
		onError:    onError,
		done:       make(chan struct{}),
		service:    s,
	}

	go client.readMessages()

	return client, nil
}

// Speak sends text to be converted to speech
func (c *StreamingClient) Speak(text string, contextID string) error {
	c.service.mu.Lock()
	c.service.totalCharacters += len(text)
	c.service.mu.Unlock()

	msg := map[string]interface{}{
		"transcript": text,
		"model_id":   "sonic-english",
		"voice": map[string]interface{}{
			"mode": "id",
			"id":   c.service.voiceID,
		},
		"output_format": map[string]interface{}{
			"container": "raw",
			"encoding":  "pcm_s16le",
			"sample_rate": 24000,
		},
		"context_id": contextID,
		"continue":   false,
	}

	return c.conn.WriteJSON(msg)
}

// SpeakStreaming sends text for streaming TTS (allows continuation)
func (c *StreamingClient) SpeakStreaming(text string, contextID string, isContinue bool) error {
	c.service.mu.Lock()
	c.service.totalCharacters += len(text)
	c.service.mu.Unlock()

	msg := map[string]interface{}{
		"transcript": text,
		"model_id":   "sonic-english",
		"voice": map[string]interface{}{
			"mode": "id",
			"id":   c.service.voiceID,
		},
		"output_format": map[string]interface{}{
			"container": "raw",
			"encoding":  "pcm_s16le",
			"sample_rate": 24000,
		},
		"context_id": contextID,
		"continue":   isContinue,
	}

	return c.conn.WriteJSON(msg)
}

// Close closes the streaming client
func (c *StreamingClient) Close() error {
	close(c.done)
	return c.conn.Close()
}

func (c *StreamingClient) readMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			messageType, message, err := c.conn.ReadMessage()
			if err != nil {
				if c.onError != nil {
					c.onError(fmt.Errorf("websocket read error: %w", err))
				}
				return
			}

			if messageType == websocket.BinaryMessage {
				// Audio data
				if c.onAudio != nil {
					c.onAudio(message)
				}
			} else if messageType == websocket.TextMessage {
				// Control message
				var resp cartesiaResponse
				if err := json.Unmarshal(message, &resp); err != nil {
					continue
				}

				if resp.Type == "done" && c.onComplete != nil {
					c.onComplete()
				} else if resp.Type == "error" && c.onError != nil {
					c.onError(fmt.Errorf("cartesia error: %s", resp.Error))
				}
			}
		}
	}
}

// GetTotalCharacters returns total characters synthesized
func (s *Service) GetTotalCharacters() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalCharacters
}

// ResetCharacters resets the character counter
func (s *Service) ResetCharacters() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalCharacters = 0
}

type cartesiaResponse struct {
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}
