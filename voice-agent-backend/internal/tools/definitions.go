package tools

import (
	"github.com/sashabaranov/go-openai"
)

// GetToolDefinitions returns all available tool definitions for the LLM
func GetToolDefinitions() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "identify_user",
				Description: "Identify the user by their phone number, name, and email. Use this when you need to know who you're speaking with or before booking/retrieving appointments. All three fields are required.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"phone_number": map[string]interface{}{
							"type":        "string",
							"description": "The user's phone number in format like +1234567890 or 1234567890",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The user's full name (cannot be empty or 'null')",
						},
						"email": map[string]interface{}{
							"type":        "string",
							"description": "The user's email address in format user@domain.com (cannot be empty or 'null')",
						},
					},
					"required": []string{"phone_number", "name", "email"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "fetch_slots",
				Description: "Fetch available appointment time slots for a given date. Returns list of available times.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"date": map[string]interface{}{
							"type":        "string",
							"description": "The date to check availability for in YYYY-MM-DD format",
						},
					},
					"required": []string{"date"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "book_appointment",
				Description: "Book an appointment for the user. Requires user to be identified first.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"date_time": map[string]interface{}{
							"type":        "string",
							"description": "The appointment date and time in ISO 8601 format (e.g., 2024-01-15T10:00:00Z)",
						},
						"duration": map[string]interface{}{
							"type":        "integer",
							"description": "Duration of the appointment in minutes (default 30)",
						},
						"purpose": map[string]interface{}{
							"type":        "string",
							"description": "The purpose or reason for the appointment",
						},
						"notes": map[string]interface{}{
							"type":        "string",
							"description": "Any additional notes for the appointment",
						},
					},
					"required": []string{"date_time"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "retrieve_appointments",
				Description: "Retrieve the user's appointments. Can fetch upcoming appointments or all past appointments.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"upcoming", "all"},
							"description": "Type of appointments to retrieve: 'upcoming' for future appointments, 'all' for all appointments",
						},
					},
					"required": []string{"type"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "cancel_appointment",
				Description: "Cancel an existing appointment by its ID.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"appointment_id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the appointment to cancel",
						},
						"reason": map[string]interface{}{
							"type":        "string",
							"description": "Optional reason for cancellation",
						},
					},
					"required": []string{"appointment_id"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "modify_appointment",
				Description: "Modify an existing appointment's date, time, or details.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"appointment_id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the appointment to modify",
						},
						"new_date_time": map[string]interface{}{
							"type":        "string",
							"description": "New date and time in ISO 8601 format (optional)",
						},
						"new_duration": map[string]interface{}{
							"type":        "integer",
							"description": "New duration in minutes (optional)",
						},
						"new_purpose": map[string]interface{}{
							"type":        "string",
							"description": "New purpose/reason (optional)",
						},
						"new_notes": map[string]interface{}{
							"type":        "string",
							"description": "New notes (optional)",
						},
					},
					"required": []string{"appointment_id"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "end_conversation",
				Description: "End the current conversation. Use this when the user says goodbye, wants to end the call, or the conversation has naturally concluded.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"reason": map[string]interface{}{
							"type":        "string",
							"description": "Reason for ending the conversation",
						},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "process_payment",
				Description: "Process payment for an appointment booking. Returns payment details and confirmation.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"appointment_id": map[string]interface{}{
							"type":        "string",
							"description": "The appointment ID to pay for",
						},
						"amount_cents": map[string]interface{}{
							"type":        "integer",
							"description": "The amount in cents (e.g., 1500 for $15.00)",
						},
						"payment_method": map[string]interface{}{
							"type":        "string",
							"description": "Payment method (card, stripe_token, etc.)",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Payment description for the transaction",
						},
					},
					"required": []string{"appointment_id", "amount_cents", "payment_method"},
				},
			},
		},
	}
}

// ToolNames for easy reference
const (
	ToolIdentifyUser         = "identify_user"
	ToolFetchSlots           = "fetch_slots"
	ToolBookAppointment      = "book_appointment"
	ToolRetrieveAppointments = "retrieve_appointments"
	ToolCancelAppointment    = "cancel_appointment"
	ToolModifyAppointment    = "modify_appointment"
	ToolEndConversation      = "end_conversation"
	ToolProcessPayment       = "process_payment"
)
