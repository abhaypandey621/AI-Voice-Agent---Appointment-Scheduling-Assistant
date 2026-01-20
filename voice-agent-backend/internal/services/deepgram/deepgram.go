package deepgram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/voice-agent/backend/internal/config"
)

const (
	deepgramAPIURL    = "https://api.deepgram.com/v1/listen"
	deepgramWSURL     = "wss://api.deepgram.com/v1/listen"
)

// Service handles Deepgram STT operations
type Service struct {
	apiKey         string
	totalMinutes   float64
	mu             sync.Mutex
}

// TranscriptResult represents a transcription result
type TranscriptResult struct {
	Transcript string  `json:"transcript"`
	Confidence float64 `json:"confidence"`
	IsFinal    bool    `json:"is_final"`
	Words      []Word  `json:"words,omitempty"`
}

// Word represents a transcribed word with timing
type Word struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence"`
}

// StreamingClient handles real-time transcription
type StreamingClient struct {
	conn       *websocket.Conn
	onResult   func(TranscriptResult)
	onError    func(error)
	done       chan struct{}
	service    *Service
	startTime  time.Time
}

// NewService creates a new Deepgram service
func NewService(cfg *config.Config) *Service {
	return &Service{
		apiKey: cfg.DeepgramAPIKey,
	}
}

// TranscribeAudio transcribes an audio buffer (REST API)
func (s *Service) TranscribeAudio(audioData []byte, mimeType string) (*TranscriptResult, error) {
	params := url.Values{}
	params.Set("model", "nova-2")
	params.Set("smart_format", "true")
	params.Set("punctuate", "true")

	reqURL := fmt.Sprintf("%s?%s", deepgramAPIURL, params.Encode())

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(audioData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+s.apiKey)
	req.Header.Set("Content-Type", mimeType)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deepgram error (status %d): %s", resp.StatusCode, string(body))
	}

	var result deepgramResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Track minutes for cost calculation
	if result.Metadata != nil {
		s.mu.Lock()
		s.totalMinutes += result.Metadata.Duration / 60.0
		s.mu.Unlock()
	}

	if len(result.Results.Channels) > 0 && len(result.Results.Channels[0].Alternatives) > 0 {
		alt := result.Results.Channels[0].Alternatives[0]
		return &TranscriptResult{
			Transcript: alt.Transcript,
			Confidence: alt.Confidence,
			IsFinal:    true,
			Words:      convertWords(alt.Words),
		}, nil
	}

	return &TranscriptResult{}, nil
}

// NewStreamingClient creates a real-time transcription client
func (s *Service) NewStreamingClient(onResult func(TranscriptResult), onError func(error)) (*StreamingClient, error) {
	params := url.Values{}
	params.Set("model", "nova-2")
	params.Set("smart_format", "true")
	params.Set("punctuate", "true")
	params.Set("interim_results", "true")
	params.Set("endpointing", "300")
	params.Set("vad_events", "true")
	params.Set("encoding", "linear16")
	params.Set("sample_rate", "16000")
	params.Set("channels", "1")

	wsURL := fmt.Sprintf("%s?%s", deepgramWSURL, params.Encode())

	header := http.Header{}
	header.Set("Authorization", "Token "+s.apiKey)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Deepgram: %w", err)
	}

	client := &StreamingClient{
		conn:      conn,
		onResult:  onResult,
		onError:   onError,
		done:      make(chan struct{}),
		service:   s,
		startTime: time.Now(),
	}

	go client.readMessages()

	return client, nil
}

// SendAudio sends audio data to Deepgram for transcription
func (c *StreamingClient) SendAudio(audioData []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, audioData)
}

// Close closes the streaming client
func (c *StreamingClient) Close() error {
	close(c.done)

	// Send close message to Deepgram
	_ = c.conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "CloseStream"}`))

	// Calculate minutes used
	duration := time.Since(c.startTime)
	c.service.mu.Lock()
	c.service.totalMinutes += duration.Minutes()
	c.service.mu.Unlock()

	return c.conn.Close()
}

func (c *StreamingClient) readMessages() {
	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if c.onError != nil {
					c.onError(fmt.Errorf("websocket read error: %w", err))
				}
				return
			}

			var resp deepgramStreamResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				continue
			}

			if resp.Type == "Results" && len(resp.Channel.Alternatives) > 0 {
				alt := resp.Channel.Alternatives[0]
				if alt.Transcript != "" {
					result := TranscriptResult{
						Transcript: alt.Transcript,
						Confidence: alt.Confidence,
						IsFinal:    resp.IsFinal,
						Words:      convertWords(alt.Words),
					}
					if c.onResult != nil {
						c.onResult(result)
					}
				}
			}
		}
	}
}

// GetTotalMinutes returns total minutes transcribed
func (s *Service) GetTotalMinutes() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalMinutes
}

// ResetMinutes resets the minute counter
func (s *Service) ResetMinutes() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalMinutes = 0
}

// Internal types for Deepgram API responses
type deepgramResponse struct {
	Metadata *struct {
		Duration float64 `json:"duration"`
	} `json:"metadata"`
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string          `json:"transcript"`
				Confidence float64         `json:"confidence"`
				Words      []deepgramWord  `json:"words"`
			} `json:"alternatives"`
		} `json:"channels"`
	} `json:"results"`
}

type deepgramStreamResponse struct {
	Type    string `json:"type"`
	IsFinal bool   `json:"is_final"`
	Channel struct {
		Alternatives []struct {
			Transcript string         `json:"transcript"`
			Confidence float64        `json:"confidence"`
			Words      []deepgramWord `json:"words"`
		} `json:"alternatives"`
	} `json:"channel"`
}

type deepgramWord struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence"`
}

func convertWords(words []deepgramWord) []Word {
	result := make([]Word, len(words))
	for i, w := range words {
		result[i] = Word{
			Word:       w.Word,
			Start:      w.Start,
			End:        w.End,
			Confidence: w.Confidence,
		}
	}
	return result
}
