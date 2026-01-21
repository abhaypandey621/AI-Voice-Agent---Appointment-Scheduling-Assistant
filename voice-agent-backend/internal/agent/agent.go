package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/database"
	"github.com/voice-agent/backend/internal/models"
	"github.com/voice-agent/backend/internal/services/cartesia"
	"github.com/voice-agent/backend/internal/services/deepgram"
	"github.com/voice-agent/backend/internal/services/llm"
	"github.com/voice-agent/backend/internal/tools"
)

// VoiceAgent manages a voice conversation session
type VoiceAgent struct {
	ID               string
	RoomName         string
	session          *models.CallSession
	llmService       *llm.Service
	deepgramService  *deepgram.Service
	cartesiaService  *cartesia.Service
	toolExecutor     *tools.ToolExecutor
	config           *config.Config

	// Streaming clients
	sttClient        *deepgram.StreamingClient
	ttsClient        *cartesia.StreamingClient

	// Callbacks
	onTranscript     func(text string, isFinal bool)
	onAgentResponse  func(text string)
	onToolCall       func(payload models.ToolCallPayload)
	onToolResult     func(payload models.ToolResultPayload)
	onAudioOutput    func(audio []byte)
	onCallEnd        func(summary *models.CallSummary, cost *models.CostBreakdown)
	onError          func(err error)

	// State
	messages         []models.ConversationMsg
	toolCalls        []models.ToolCallRecord
	isProcessing     bool
	shouldEnd        bool
	startTime        time.Time
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

// AgentConfig holds agent configuration
type AgentConfig struct {
	OnTranscript    func(text string, isFinal bool)
	OnAgentResponse func(text string)
	OnToolCall      func(payload models.ToolCallPayload)
	OnToolResult    func(payload models.ToolResultPayload)
	OnAudioOutput   func(audio []byte)
	OnCallEnd       func(summary *models.CallSummary, cost *models.CostBreakdown)
	OnError         func(err error)
}

// NewVoiceAgent creates a new voice agent
func NewVoiceAgent(cfg *config.Config, roomName string, agentCfg *AgentConfig) (*VoiceAgent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	agentID := uuid.New().String()

	agent := &VoiceAgent{
		ID:              agentID,
		RoomName:        roomName,
		config:          cfg,
		llmService:      llm.NewService(cfg),
		deepgramService: deepgram.NewService(cfg),
		cartesiaService: cartesia.NewService(cfg),
		messages:        make([]models.ConversationMsg, 0),
		toolCalls:       make([]models.ToolCallRecord, 0),
		startTime:       time.Now(),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Set callbacks
	if agentCfg != nil {
		agent.onTranscript = agentCfg.OnTranscript
		agent.onAgentResponse = agentCfg.OnAgentResponse
		agent.onToolCall = agentCfg.OnToolCall
		agent.onToolResult = agentCfg.OnToolResult
		agent.onAudioOutput = agentCfg.OnAudioOutput
		agent.onCallEnd = agentCfg.OnCallEnd
		agent.onError = agentCfg.OnError
	}

	// Create tool executor
	agent.toolExecutor = tools.NewToolExecutor(
		agentID,
		func(payload models.ToolCallPayload) {
			agent.mu.Lock()
			agent.toolCalls = append(agent.toolCalls, models.ToolCallRecord{
				ID:        payload.ID,
				Name:      payload.Name,
				Arguments: payload.Arguments,
				Timestamp: time.Now(),
			})
			agent.mu.Unlock()

			if agent.onToolCall != nil {
				agent.onToolCall(payload)
			}
		},
		func(payload models.ToolResultPayload) {
			agent.mu.Lock()
			for i := range agent.toolCalls {
				if agent.toolCalls[i].ID == payload.ID {
					agent.toolCalls[i].Result = payload.Result
					break
				}
			}
			agent.mu.Unlock()

			if agent.onToolResult != nil {
				agent.onToolResult(payload)
			}
		},
	)

	// Initialize session
	agent.session = &models.CallSession{
		ID:        agentID,
		RoomName:  roomName,
		StartedAt: agent.startTime,
		Messages:  agent.messages,
		ToolCalls: agent.toolCalls,
	}

	return agent, nil
}

// Start starts the voice agent
func (a *VoiceAgent) Start() error {
	// Note: STT streaming is initialized lazily when first audio arrives
	// This prevents Deepgram timeout when user hasn't started speaking yet

	// Initialize TTS streaming (optional)
	ttsClient, err := a.cartesiaService.NewStreamingClient(
		func(audio []byte) {
			if a.onAudioOutput != nil {
				a.onAudioOutput(audio)
			}
		},
		func() {
			// TTS complete
		},
		func(err error) {
			if a.onError != nil {
				a.onError(fmt.Errorf("TTS error: %w", err))
			}
		},
	)
	if err != nil {
		// TTS streaming is optional, continue without it
		a.ttsClient = nil
	} else {
		a.ttsClient = ttsClient
	}

	// Send initial greeting
	go a.sendGreeting()

	return nil
}

// initSTT initializes the STT streaming client (called on first audio)
func (a *VoiceAgent) initSTT() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sttClient != nil {
		return nil // Already initialized
	}

	sttClient, err := a.deepgramService.NewStreamingClient(
		func(result deepgram.TranscriptResult) {
			if a.onTranscript != nil {
				a.onTranscript(result.Transcript, result.IsFinal)
			}

			// Process final transcripts
			if result.IsFinal && result.Transcript != "" {
				go a.ProcessUserInput(result.Transcript)
			}
		},
		func(err error) {
			if a.onError != nil {
				a.onError(fmt.Errorf("STT error: %w", err))
			}
			// Reset client so it can be reinitialized
			a.mu.Lock()
			a.sttClient = nil
			a.mu.Unlock()
		},
	)
	if err != nil {
		return fmt.Errorf("failed to start STT: %w", err)
	}
	a.sttClient = sttClient
	return nil
}

// Stop stops the voice agent
func (a *VoiceAgent) Stop() {
	a.cancel()

	if a.sttClient != nil {
		a.sttClient.Close()
	}

	if a.ttsClient != nil {
		a.ttsClient.Close()
	}
}

// SendAudio sends audio data for transcription
func (a *VoiceAgent) SendAudio(audioData []byte) error {
	// Lazy initialize STT on first audio data
	if a.sttClient == nil {
		if err := a.initSTT(); err != nil {
			return err
		}
	}
	return a.sttClient.SendAudio(audioData)
}

// ProcessUserInput processes user speech input
func (a *VoiceAgent) ProcessUserInput(text string) {
	log.Printf("ProcessUserInput called with: %s", text)

	a.mu.Lock()
	if a.isProcessing {
		log.Printf("Already processing, skipping input")
		a.mu.Unlock()
		return
	}
	a.isProcessing = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.isProcessing = false
		a.mu.Unlock()
	}()

	// Add user message
	a.mu.Lock()
	a.messages = append(a.messages, models.ConversationMsg{
		Role:      "user",
		Content:   text,
		Timestamp: time.Now(),
	})
	a.mu.Unlock()

	// Get LLM response
	log.Printf("Calling LLM with %d messages", len(a.messages))
	response, err := a.llmService.Chat(a.ctx, a.messages, a.toolExecutor)
	if err != nil {
		log.Printf("LLM error: %v", err)
		if a.onError != nil {
			a.onError(fmt.Errorf("LLM error: %w", err))
		}
		return
	}
	log.Printf("LLM response: %s", response.Content)

	// Add assistant message
	a.mu.Lock()
	a.messages = append(a.messages, models.ConversationMsg{
		Role:      "assistant",
		Content:   response.Content,
		Timestamp: time.Now(),
	})
	// Update user info in session
	a.session.UserPhone = a.toolExecutor.GetUserPhone()
	a.session.UserName = a.toolExecutor.GetUserName()
	a.mu.Unlock()

	// Notify response
	if a.onAgentResponse != nil {
		a.onAgentResponse(response.Content)
	}

	// Synthesize speech
	a.synthesizeSpeech(response.Content)

	// Check if should end
	if response.ShouldEnd {
		a.mu.Lock()
		a.shouldEnd = true
		a.mu.Unlock()
		go a.endConversation()
	}
}

// ProcessTextInput processes direct text input (for testing)
func (a *VoiceAgent) ProcessTextInput(text string) {
	log.Printf("Agent processing text input: %s", text)
	if a.onTranscript != nil {
		a.onTranscript(text, true)
	}
	a.ProcessUserInput(text)
}

func (a *VoiceAgent) sendGreeting() {
	greeting := "Hello! I'm Ava, your appointment scheduling assistant. How can I help you today? You can book, check, or manage your appointments."

	a.mu.Lock()
	a.messages = append(a.messages, models.ConversationMsg{
		Role:      "assistant",
		Content:   greeting,
		Timestamp: time.Now(),
	})
	a.mu.Unlock()

	if a.onAgentResponse != nil {
		a.onAgentResponse(greeting)
	}

	a.synthesizeSpeech(greeting)
}

func (a *VoiceAgent) synthesizeSpeech(text string) {
	if text == "" {
		return
	}

	// Use streaming TTS if available
	if a.ttsClient != nil {
		contextID := uuid.New().String()
		if err := a.ttsClient.Speak(text, contextID); err != nil {
			// Fall back to REST API
			a.synthesizeSpeechREST(text)
		}
		return
	}

	// Use REST API
	a.synthesizeSpeechREST(text)
}

func (a *VoiceAgent) synthesizeSpeechREST(text string) {
	audio, err := a.cartesiaService.SynthesizeSpeech(text)
	if err != nil {
		if a.onError != nil {
			a.onError(fmt.Errorf("TTS synthesis error: %w", err))
		}
		return
	}

	if a.onAudioOutput != nil {
		a.onAudioOutput(audio)
	}
}

func (a *VoiceAgent) endConversation() {
	log.Printf("[endConversation] Starting summary generation for session %s", a.ID)

	a.mu.RLock()
	messages := make([]models.ConversationMsg, len(a.messages))
	copy(messages, a.messages)
	a.mu.RUnlock()

	log.Printf("[endConversation] Copied %d messages for summary", len(messages))

	// Get user's appointments for summary
	var appointments []models.Appointment
	userPhone := a.toolExecutor.GetUserPhone()
	if userPhone != "" {
		log.Printf("[endConversation] Fetching appointments for user: %s", userPhone)
		apts, err := database.DB.GetUpcomingAppointments(userPhone)
		if err == nil {
			appointments = apts
			log.Printf("[endConversation] Found %d appointments", len(appointments))
		} else {
			log.Printf("[endConversation] Error fetching appointments: %v", err)
		}
	} else {
		log.Printf("[endConversation] No user phone set, skipping appointment fetch")
	}

	// Generate summary
	log.Printf("[endConversation] Generating LLM summary...")
	summary, err := a.llmService.GenerateSummary(a.ctx, messages, appointments)
	if err != nil {
		log.Printf("[endConversation] ERROR generating summary: %v", err)
		if a.onError != nil {
			a.onError(fmt.Errorf("summary generation error: %w", err))
		}
		summary = &models.CallSummary{
			Summary:            "Call completed with the appointment assistant.",
			AppointmentsBooked: appointments,
			UserPreferences:    []string{},
			KeyTopics:          []string{"appointment scheduling"},
			CreatedAt:          time.Now(),
		}
	} else {
		log.Printf("[endConversation] Summary generated successfully: %s", summary.Summary)
	}

	// Set session info
	summary.ID = uuid.New().String()
	summary.SessionID = a.ID
	summary.UserPhone = userPhone
	summary.Duration = int(time.Since(a.startTime).Seconds())

	log.Printf("[endConversation] Call duration: %d seconds", summary.Duration)

	// Calculate costs
	cost := a.calculateCosts()
	log.Printf("[endConversation] Costs calculated - Total: $%.4f", cost.TotalCost)

	// Save summary to database
	if database.DB != nil {
		if err := database.DB.SaveCallSummary(summary); err != nil {
			log.Printf("[endConversation] ERROR saving summary to database: %v", err)
		} else {
			log.Printf("[endConversation] Summary saved to database")
		}
	}

	// Notify call end
	if a.onCallEnd != nil {
		log.Printf("[endConversation] Sending call summary to client")
		a.onCallEnd(summary, cost)
	} else {
		log.Printf("[endConversation] WARNING: onCallEnd callback is nil")
	}

	log.Printf("[endConversation] Completed")
}

func (a *VoiceAgent) calculateCosts() *models.CostBreakdown {
	sttMinutes := a.deepgramService.GetTotalMinutes()
	ttsCharacters := a.cartesiaService.GetTotalCharacters()
	llmTokens := a.llmService.GetTokenCount()

	sttCost := sttMinutes * a.config.DeepgramPricePerMin
	ttsCost := float64(ttsCharacters) * a.config.CartesiaPricePerChar
	llmCost := float64(llmTokens) * a.config.LLMPricePerToken

	return &models.CostBreakdown{
		STTCost:       sttCost,
		TTSCost:       ttsCost,
		LLMCost:       llmCost,
		TotalCost:     sttCost + ttsCost + llmCost,
		STTMinutes:    sttMinutes,
		TTSCharacters: ttsCharacters,
		LLMTokens:     llmTokens,
	}
}

// GetSession returns the current session state
func (a *VoiceAgent) GetSession() *models.CallSession {
	a.mu.RLock()
	defer a.mu.RUnlock()

	a.session.Messages = a.messages
	a.session.ToolCalls = a.toolCalls
	a.session.CostBreakdown = a.calculateCosts()

	return a.session
}

// GetMessages returns conversation messages
func (a *VoiceAgent) GetMessages() []models.ConversationMsg {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]models.ConversationMsg, len(a.messages))
	copy(result, a.messages)
	return result
}

// GetToolCalls returns tool call history
func (a *VoiceAgent) GetToolCalls() []models.ToolCallRecord {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]models.ToolCallRecord, len(a.toolCalls))
	copy(result, a.toolCalls)
	return result
}

// EndCall manually ends the call
func (a *VoiceAgent) EndCall() {
	a.mu.Lock()
	if a.shouldEnd {
		a.mu.Unlock()
		return
	}
	a.shouldEnd = true
	a.mu.Unlock()

	a.endConversation()
}

// ToJSON serializes agent state
func (a *VoiceAgent) ToJSON() ([]byte, error) {
	return json.Marshal(a.GetSession())
}
