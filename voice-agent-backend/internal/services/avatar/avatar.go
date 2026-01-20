package avatar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/voice-agent/backend/internal/config"
)

const (
	tavusAPIURL = "https://tavusapi.com/v2"
)

// Service handles avatar operations
type Service struct {
	provider string
	apiKey   string
	avatarID string
	client   *http.Client
}

// ConversationSession represents an active avatar conversation
type ConversationSession struct {
	ConversationID   string `json:"conversation_id"`
	ConversationURL  string `json:"conversation_url"`
	Status           string `json:"status"`
}

// NewService creates a new avatar service
func NewService(cfg *config.Config) *Service {
	return &Service{
		provider: cfg.AvatarProvider,
		apiKey:   cfg.AvatarAPIKey,
		avatarID: cfg.AvatarAvatarID,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateConversation creates a new avatar conversation session
func (s *Service) CreateConversation(replicaID string, callbackURL string) (*ConversationSession, error) {
	if s.provider == "tavus" {
		return s.createTavusConversation(replicaID, callbackURL)
	}
	return nil, fmt.Errorf("unsupported avatar provider: %s", s.provider)
}

// EndConversation ends an avatar conversation
func (s *Service) EndConversation(conversationID string) error {
	if s.provider == "tavus" {
		return s.endTavusConversation(conversationID)
	}
	return fmt.Errorf("unsupported avatar provider: %s", s.provider)
}

// GetConversation gets conversation details
func (s *Service) GetConversation(conversationID string) (*ConversationSession, error) {
	if s.provider == "tavus" {
		return s.getTavusConversation(conversationID)
	}
	return nil, fmt.Errorf("unsupported avatar provider: %s", s.provider)
}

// Tavus-specific implementations

func (s *Service) createTavusConversation(replicaID string, callbackURL string) (*ConversationSession, error) {
	if replicaID == "" {
		replicaID = s.avatarID
	}

	reqBody := map[string]interface{}{
		"replica_id": replicaID,
	}

	if callbackURL != "" {
		reqBody["callback_url"] = callbackURL
	}

	// Configure conversation settings
	reqBody["conversation_name"] = fmt.Sprintf("voice-agent-%d", time.Now().Unix())
	reqBody["conversational_context"] = "You are a helpful AI assistant named Ava. Help users with appointment scheduling."
	reqBody["custom_greeting"] = "Hello! I'm Ava, your appointment scheduling assistant. How can I help you today?"
	reqBody["properties"] = map[string]interface{}{
		"max_call_duration":    1800, // 30 minutes max
		"participant_left_timeout": 60,
		"enable_recording":     false,
		"language":             "english",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", tavusAPIURL+"/conversations", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("tavus error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ConversationID  string `json:"conversation_id"`
		ConversationURL string `json:"conversation_url"`
		Status          string `json:"status"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ConversationSession{
		ConversationID:  result.ConversationID,
		ConversationURL: result.ConversationURL,
		Status:          result.Status,
	}, nil
}

func (s *Service) endTavusConversation(conversationID string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/conversations/%s/end", tavusAPIURL, conversationID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tavus error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *Service) getTavusConversation(conversationID string) (*ConversationSession, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/conversations/%s", tavusAPIURL, conversationID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tavus error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		ConversationID  string `json:"conversation_id"`
		ConversationURL string `json:"conversation_url"`
		Status          string `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ConversationSession{
		ConversationID:  result.ConversationID,
		ConversationURL: result.ConversationURL,
		Status:          result.Status,
	}, nil
}

// ListReplicas lists available avatar replicas
func (s *Service) ListReplicas() ([]map[string]interface{}, error) {
	if s.provider != "tavus" {
		return nil, fmt.Errorf("unsupported avatar provider: %s", s.provider)
	}

	req, err := http.NewRequest("GET", tavusAPIURL+"/replicas", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tavus error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}
