package models

import (
	"time"
)

// User represents a user identified by phone number
type User struct {
	ID           string    `json:"id"`
	PhoneNumber  string    `json:"phone_number"`
	Name         string    `json:"name,omitempty"`
	Email        string    `json:"email,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Appointment represents a booked appointment
type Appointment struct {
	ID            string    `json:"id"`
	UserPhone     string    `json:"user_phone"`
	UserName      string    `json:"user_name,omitempty"`
	DateTime      time.Time `json:"date_time"`
	Duration      int       `json:"duration"` // in minutes
	Purpose       string    `json:"purpose,omitempty"`
	Status        string    `json:"status"` // booked, cancelled, completed
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AppointmentStatus constants
const (
	StatusBooked    = "booked"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
)

// TimeSlot represents an available time slot
type TimeSlot struct {
	DateTime  time.Time `json:"date_time"`
	Available bool      `json:"available"`
	Duration  int       `json:"duration"`
}

// CallSession represents an active voice call session
type CallSession struct {
	ID              string            `json:"id"`
	RoomName        string            `json:"room_name"`
	UserPhone       string            `json:"user_phone,omitempty"`
	UserName        string            `json:"user_name,omitempty"`
	StartedAt       time.Time         `json:"started_at"`
	EndedAt         *time.Time        `json:"ended_at,omitempty"`
	Messages        []ConversationMsg `json:"messages"`
	ToolCalls       []ToolCallRecord  `json:"tool_calls"`
	CostBreakdown   *CostBreakdown    `json:"cost_breakdown,omitempty"`
}

// ConversationMsg represents a message in the conversation
type ConversationMsg struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ToolCallRecord represents a tool call made during the conversation
type ToolCallRecord struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Result    interface{}            `json:"result"`
	Timestamp time.Time              `json:"timestamp"`
}

// CallSummary represents the summary generated at call end
type CallSummary struct {
	ID                 string       `json:"id"`
	SessionID          string       `json:"session_id"`
	UserPhone          string       `json:"user_phone,omitempty"`
	Summary            string       `json:"summary"`
	AppointmentsBooked []Appointment `json:"appointments_booked"`
	UserPreferences    []string     `json:"user_preferences"`
	KeyTopics          []string     `json:"key_topics"`
	Duration           int          `json:"duration_seconds"`
	CreatedAt          time.Time    `json:"created_at"`
}

// CostBreakdown shows the cost breakdown for a call
type CostBreakdown struct {
	STTCost       float64 `json:"stt_cost"`       // Speech to text (Deepgram)
	TTSCost       float64 `json:"tts_cost"`       // Text to speech (Cartesia)
	LLMCost       float64 `json:"llm_cost"`       // LLM tokens
	AvatarCost    float64 `json:"avatar_cost"`    // Avatar streaming
	TotalCost     float64 `json:"total_cost"`
	STTMinutes    float64 `json:"stt_minutes"`
	TTSCharacters int     `json:"tts_characters"`
	LLMTokens     int     `json:"llm_tokens"`
}

// WebSocket message types
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocket message type constants
const (
	WSTypeTranscript     = "transcript"
	WSTypeAgentResponse  = "agent_response"
	WSTypeToolCall       = "tool_call"
	WSTypeToolResult     = "tool_result"
	WSTypeCallSummary    = "call_summary"
	WSTypeCallEnd        = "call_end"
	WSTypeError          = "error"
	WSTypeAvatarState    = "avatar_state"
	WSTypeCostUpdate     = "cost_update"
)

// ToolCallPayload for WebSocket
type ToolCallPayload struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Status    string                 `json:"status"` // pending, executing, completed, failed
}

// ToolResultPayload for WebSocket
type ToolResultPayload struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

// LLM Tool definitions
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Token join response
type TokenResponse struct {
	Token    string `json:"token"`
	RoomName string `json:"room_name"`
	URL      string `json:"url"`
}
