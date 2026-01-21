package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/models"
	"github.com/voice-agent/backend/internal/tools"
)

// filterToolCallAnnouncements removes tool call announcements from LLM responses
// These are patterns like "[Calling the X tool]" or "[Using X tool for...]"
func filterToolCallAnnouncements(content string) string {
	// Pattern to match tool call announcements in square brackets
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\[Calling the \w+ tool[^\]]*\]`),
		regexp.MustCompile(`\[Using \w+ tool[^\]]*\]`),
		regexp.MustCompile(`\[Call \w+ tool[^\]]*\]`),
		regexp.MustCompile(`\[Calling \w+[^\]]*\]`),
	}

	result := content
	for _, pattern := range patterns {
		result = pattern.ReplaceAllString(result, "")
	}

	// Clean up extra whitespace
	result = strings.TrimSpace(result)
	// Replace multiple spaces with single space
	spacePattern := regexp.MustCompile(`\s+`)
	result = spacePattern.ReplaceAllString(result, " ")

	return result
}

// getSystemPrompt returns the system prompt with current date
func getSystemPrompt() string {
	currentDate := time.Now().Format("January 2, 2006")
	currentYear := time.Now().Year()

	return fmt.Sprintf(`You are a friendly and professional AI voice assistant for an appointment scheduling service. Your name is "Ava".

IMPORTANT: Today's date is %s. The current year is %d. When users say "tomorrow", "next week", etc., calculate dates relative to TODAY.

Your capabilities:
1. Help users identify themselves intelligently (ask phone first, then name/email only if they're new)
2. Check available appointment time slots
3. Book new appointments
4. Retrieve existing appointments
5. Cancel appointments
6. Modify appointment details
7. End conversations politely

CRITICAL - Smart User Identification:
The identify_user tool is intelligent. It checks the database automatically:

STEP 1: Always ask for phone number first
STEP 2: Call identify_user with just the phone_number (empty name and email)
STEP 3: Check the response:
  - If response shows "Welcome back" → User already exists! Use their data and proceed
  - If response shows "name is required for new registration" → User is NEW, ask for name
STEP 4: For NEW users only:
  - Ask for full name
  - Ask for email address
  - Call identify_user again with phone_number, name, and email

Example flow - EXISTING USER (quicker!):
  User: "I want to check my appointments"
  You: "I'd be happy to help! Could you please provide your phone number?"
  User: "+1-555-1234"
  You: [Call identify_user with phone_number: "+1-555-1234", name: "", email: ""]
  System: Returns "Welcome back, John!" with their stored name and email
  You: "Perfect John! Let me retrieve your appointments..."

Example flow - NEW USER:
  User: "I want to book an appointment"
  You: "I'd be happy to help! Could you please provide your phone number?"
  User: "+1-555-1234"
  You: [Call identify_user with phone_number: "+1-555-1234", name: "", email: ""]
  System: Returns error "name is required for new registration"
  You: "I see this is your first time. May I have your full name?"
  User: "John Smith"
  You: "Thank you! And your email address?"
  User: "john@example.com"
  You: [Call identify_user with phone_number: "+1-555-1234", name: "John Smith", email: "john@example.com"]
  System: Returns success with user created
  You: "Welcome John! Now let's book your appointment..."

Guidelines:
- Always be polite, professional, and helpful
- Speak naturally as if having a phone conversation
- Keep responses concise since this is a voice interface (1-3 sentences typically)
- Always confirm appointment details before booking
- If a slot is unavailable, suggest alternatives
- When ending a call, summarize any actions taken
- Use natural language for dates and times (e.g., "tomorrow at 2 PM" instead of ISO format)
- If user seems confused, offer to help guide them
- When using fetch_slots tool, always use dates in YYYY-MM-DD format

Important:
- You MUST use tools to perform actions - don't just say you'll do something, actually call the tool
- After identifying a user, greet them by name
- Double-check details before making bookings
- Be proactive in offering help but don't be pushy
- ALWAYS use the current year %d for any dates
- For identify_user: pass phone_number always, name and email only when available
- Listen to the tool's error messages - they guide you on what's needed`, currentDate, currentYear, currentYear)
}

// Service handles LLM interactions
type Service struct {
	client     *openai.Client
	model      string
	tokenCount int
	toolDefs   []openai.Tool
}

// NewService creates a new LLM service
func NewService(cfg *config.Config) *Service {
	clientConfig := openai.DefaultConfig(cfg.LLMAPIKey)
	if cfg.LLMBaseURL != "" && cfg.LLMBaseURL != "https://api.openai.com/v1" {
		clientConfig.BaseURL = cfg.LLMBaseURL
	}

	return &Service{
		client:   openai.NewClientWithConfig(clientConfig),
		model:    cfg.LLMModel,
		toolDefs: tools.GetToolDefinitions(),
	}
}

// Message represents a conversation message
type Message struct {
	Role       string            `json:"role"`
	Content    string            `json:"content"`
	ToolCalls  []openai.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
}

// Response represents an LLM response
type Response struct {
	Content    string
	ToolCalls  []ToolCall
	TokensUsed int
	ShouldEnd  bool
}

// ToolCall represents a tool call from the LLM
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// Chat sends a message and gets a response with tool support
func (s *Service) Chat(ctx context.Context, messages []models.ConversationMsg, toolExecutor *tools.ToolExecutor) (*Response, error) {
	// Convert to OpenAI messages
	openAIMessages := s.convertMessages(messages)

	// Add system prompt at the beginning (with current date)
	openAIMessages = append([]openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: getSystemPrompt(),
		},
	}, openAIMessages...)

	for {
		// Make the API call
		resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       s.model,
			Messages:    openAIMessages,
			Tools:       s.toolDefs,
			Temperature: 0.7,
			MaxTokens:   500,
		})
		if err != nil {
			return nil, fmt.Errorf("chat completion failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no choices in response")
		}

		choice := resp.Choices[0]
		s.tokenCount += resp.Usage.TotalTokens

		// Check if there are tool calls
		if len(choice.Message.ToolCalls) > 0 {
			// Add assistant message with tool calls
			openAIMessages = append(openAIMessages, choice.Message)

			// Execute each tool call
			shouldEnd := false
			for _, tc := range choice.Message.ToolCalls {
				result, err := toolExecutor.ExecuteTool(tc.Function.Name, json.RawMessage(tc.Function.Arguments))

				var resultStr string
				if err != nil {
					resultStr = fmt.Sprintf(`{"error": "%s"}`, err.Error())
				} else {
					resultBytes, _ := json.Marshal(result)
					resultStr = string(resultBytes)

					// Check if this is an end conversation call
					if tc.Function.Name == tools.ToolEndConversation {
						if resultMap, ok := result.(map[string]interface{}); ok {
							if end, ok := resultMap["should_end"].(bool); ok && end {
								shouldEnd = true
							}
						}
					}
				}

				// Add tool result message
				openAIMessages = append(openAIMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    resultStr,
					ToolCallID: tc.ID,
				})
			}

			// If should end, return immediately with appropriate message
			if shouldEnd {
				// Get final response
				finalResp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
					Model:       s.model,
					Messages:    openAIMessages,
					Temperature: 0.7,
					MaxTokens:   200,
				})
				if err != nil {
					return &Response{
						Content:    "Thank you for calling. Goodbye!",
						ShouldEnd:  true,
						TokensUsed: s.tokenCount,
					}, nil
				}

				s.tokenCount += finalResp.Usage.TotalTokens
				content := ""
				if len(finalResp.Choices) > 0 {
					content = filterToolCallAnnouncements(finalResp.Choices[0].Message.Content)
				}

				return &Response{
					Content:    content,
					ShouldEnd:  true,
					TokensUsed: s.tokenCount,
				}, nil
			}

			// Continue the loop to get the next response
			continue
		}

		// No tool calls, return the content (filtered)
		return &Response{
			Content:    filterToolCallAnnouncements(choice.Message.Content),
			TokensUsed: s.tokenCount,
			ShouldEnd:  false,
		}, nil
	}
}

// GenerateSummary creates a call summary
func (s *Service) GenerateSummary(ctx context.Context, messages []models.ConversationMsg, appointments []models.Appointment) (*models.CallSummary, error) {
	summaryPrompt := `You are analyzing a call between a user and an AI appointment assistant. Generate a comprehensive call summary.

Respond ONLY with valid JSON in this exact format (no markdown, no code blocks):
{
  "summary": "A 2-3 sentence summary of what happened in the call",
  "user_preferences": ["preference 1", "preference 2"],
  "key_topics": ["topic 1", "topic 2"]
}

Guidelines:
- "summary": Describe what the user wanted and what actions were taken
- "user_preferences": List any stated preferences (times, days, contact methods, etc.)
- "key_topics": List the main topics discussed (booking, cancellation, inquiry, etc.)`

	// Build conversation text
	convText := "Conversation History:\n"
	for _, msg := range messages {
		role := msg.Role
		if role == "assistant" {
			role = "Agent"
		} else if role == "user" {
			role = "User"
		}
		convText += fmt.Sprintf("%s: %s\n", role, msg.Content)
	}

	// Add appointment info if any
	if len(appointments) > 0 {
		convText += "\n\nCurrent User Appointments:\n"
		for _, apt := range appointments {
			convText += fmt.Sprintf("- %s: %s (%d min) - Status: %s\n",
				apt.DateTime.Format("Monday, January 2, 2006 at 3:04 PM"),
				apt.Purpose,
				apt.Duration,
				apt.Status,
			)
		}
	}

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: summaryPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: convText,
			},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	s.tokenCount += resp.Usage.TotalTokens

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response for summary")
	}

	responseContent := resp.Choices[0].Message.Content

	// Parse the JSON response
	var summaryData struct {
		Summary         string   `json:"summary"`
		UserPreferences []string `json:"user_preferences"`
		KeyTopics       []string `json:"key_topics"`
	}

	// Try to extract JSON from the response (in case LLM wraps it in markdown)
	jsonContent := responseContent
	if idx := findJSONStart(responseContent); idx >= 0 {
		jsonContent = responseContent[idx:]
		if endIdx := findJSONEnd(jsonContent); endIdx > 0 {
			jsonContent = jsonContent[:endIdx+1]
		}
	}

	if err := json.Unmarshal([]byte(jsonContent), &summaryData); err != nil {
		// If JSON parsing fails, use the raw content as summary
		summaryData.Summary = responseContent
		summaryData.UserPreferences = []string{}
		summaryData.KeyTopics = []string{"appointment scheduling"}
	}

	// Ensure we have at least some default values
	if summaryData.Summary == "" {
		summaryData.Summary = "Call completed with the appointment assistant."
	}
	if summaryData.KeyTopics == nil {
		summaryData.KeyTopics = []string{}
	}
	if summaryData.UserPreferences == nil {
		summaryData.UserPreferences = []string{}
	}

	return &models.CallSummary{
		Summary:            summaryData.Summary,
		AppointmentsBooked: appointments,
		UserPreferences:    summaryData.UserPreferences,
		KeyTopics:          summaryData.KeyTopics,
		CreatedAt:          time.Now(),
	}, nil
}

// findJSONStart finds the start of a JSON object in a string
func findJSONStart(s string) int {
	for i, c := range s {
		if c == '{' {
			return i
		}
	}
	return -1
}

// findJSONEnd finds the end of a JSON object in a string
func findJSONEnd(s string) int {
	depth := 0
	for i, c := range s {
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// GetTokenCount returns total tokens used
func (s *Service) GetTokenCount() int {
	return s.tokenCount
}

// ResetTokenCount resets the token counter
func (s *Service) ResetTokenCount() {
	s.tokenCount = 0
}

func (s *Service) convertMessages(messages []models.ConversationMsg) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		role := openai.ChatMessageRoleUser
		switch msg.Role {
		case "assistant":
			role = openai.ChatMessageRoleAssistant
		case "system":
			role = openai.ChatMessageRoleSystem
		}
		result = append(result, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}
	return result
}
