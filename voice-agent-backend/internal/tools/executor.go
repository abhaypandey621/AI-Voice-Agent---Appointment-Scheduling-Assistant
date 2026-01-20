package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/voice-agent/backend/internal/database"
	"github.com/voice-agent/backend/internal/models"
)

// ToolExecutor handles the execution of tool calls
type ToolExecutor struct {
	sessionID   string
	userPhone   string
	userName    string
	onToolCall  func(payload models.ToolCallPayload)
	onToolResult func(payload models.ToolResultPayload)
}

// NewToolExecutor creates a new tool executor for a session
func NewToolExecutor(sessionID string, onToolCall func(models.ToolCallPayload), onToolResult func(models.ToolResultPayload)) *ToolExecutor {
	return &ToolExecutor{
		sessionID:   sessionID,
		onToolCall:  onToolCall,
		onToolResult: onToolResult,
	}
}

// SetUserIdentity sets the identified user for the session
func (e *ToolExecutor) SetUserIdentity(phone, name string) {
	e.userPhone = phone
	e.userName = name
}

// GetUserPhone returns the current user's phone
func (e *ToolExecutor) GetUserPhone() string {
	return e.userPhone
}

// GetUserName returns the current user's name
func (e *ToolExecutor) GetUserName() string {
	return e.userName
}

// ExecuteTool executes a tool call and returns the result
func (e *ToolExecutor) ExecuteTool(toolName string, arguments json.RawMessage) (interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	toolCallID := uuid.New().String()

	// Notify tool call started
	if e.onToolCall != nil {
		e.onToolCall(models.ToolCallPayload{
			ID:        toolCallID,
			Name:      toolName,
			Arguments: args,
			Status:    "executing",
		})
	}

	var result interface{}
	var err error

	switch toolName {
	case ToolIdentifyUser:
		result, err = e.identifyUser(args)
	case ToolFetchSlots:
		result, err = e.fetchSlots(args)
	case ToolBookAppointment:
		result, err = e.bookAppointment(args)
	case ToolRetrieveAppointments:
		result, err = e.retrieveAppointments(args)
	case ToolCancelAppointment:
		result, err = e.cancelAppointment(args)
	case ToolModifyAppointment:
		result, err = e.modifyAppointment(args)
	case ToolEndConversation:
		result, err = e.endConversation(args)
	default:
		err = fmt.Errorf("unknown tool: %s", toolName)
	}

	// Notify tool result
	if e.onToolResult != nil {
		payload := models.ToolResultPayload{
			ID:     toolCallID,
			Name:   toolName,
			Result: result,
		}
		if err != nil {
			payload.Error = err.Error()
		}
		e.onToolResult(payload)
	}

	return result, err
}

func (e *ToolExecutor) identifyUser(args map[string]interface{}) (interface{}, error) {
	phone, ok := args["phone_number"].(string)
	if !ok || phone == "" {
		return nil, fmt.Errorf("phone_number is required")
	}

	name, _ := args["name"].(string)

	// Normalize phone number (basic)
	phone = normalizePhoneNumber(phone)

	// Check if user exists
	user, err := database.DB.GetUserByPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check user: %w", err)
	}

	if user == nil {
		// Create new user
		user = &models.User{
			ID:          uuid.New().String(),
			PhoneNumber: phone,
			Name:        name,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := database.DB.CreateUser(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else if name != "" && user.Name == "" {
		// Update name if not set
		user.Name = name
		user.UpdatedAt = time.Now()
		_ = database.DB.UpdateUser(user)
	}

	e.SetUserIdentity(phone, user.Name)

	return map[string]interface{}{
		"success":      true,
		"user_id":      user.ID,
		"phone_number": user.PhoneNumber,
		"name":         user.Name,
		"is_new_user":  user.Name == name && name != "",
		"message":      fmt.Sprintf("User identified: %s (%s)", user.Name, user.PhoneNumber),
	}, nil
}

func (e *ToolExecutor) fetchSlots(args map[string]interface{}) (interface{}, error) {
	dateStr, ok := args["date"].(string)
	if !ok || dateStr == "" {
		return nil, fmt.Errorf("date is required")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}

	// Generate hardcoded available slots (9 AM to 5 PM, every 30 minutes)
	slots := []map[string]interface{}{}
	loc := time.Local

	for hour := 9; hour < 17; hour++ {
		for _, minute := range []int{0, 30} {
			slotTime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, loc)

			// Skip past slots
			if slotTime.Before(time.Now()) {
				continue
			}

			// Check availability in database
			available, err := database.DB.CheckSlotAvailability(slotTime, 30)
			if err != nil {
				available = true // Default to available on error
			}

			slots = append(slots, map[string]interface{}{
				"date_time":  slotTime.Format(time.RFC3339),
				"time":       slotTime.Format("3:04 PM"),
				"available":  available,
				"duration":   30,
			})
		}
	}

	availableCount := 0
	for _, slot := range slots {
		if slot["available"].(bool) {
			availableCount++
		}
	}

	return map[string]interface{}{
		"date":            dateStr,
		"slots":           slots,
		"total_slots":     len(slots),
		"available_slots": availableCount,
		"message":         fmt.Sprintf("Found %d available slots out of %d total for %s", availableCount, len(slots), dateStr),
	}, nil
}

func (e *ToolExecutor) bookAppointment(args map[string]interface{}) (interface{}, error) {
	if e.userPhone == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "User not identified. Please identify the user first by asking for their phone number.",
		}, nil
	}

	dateTimeStr, ok := args["date_time"].(string)
	if !ok || dateTimeStr == "" {
		return nil, fmt.Errorf("date_time is required")
	}

	dateTime, err := time.Parse(time.RFC3339, dateTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date_time format, use ISO 8601 (e.g., 2024-01-15T10:00:00Z)")
	}

	// Check if slot is in the past
	if dateTime.Before(time.Now()) {
		return map[string]interface{}{
			"success": false,
			"error":   "Cannot book appointments in the past",
		}, nil
	}

	duration := 30
	if d, ok := args["duration"].(float64); ok {
		duration = int(d)
	}

	// Check slot availability
	available, err := database.DB.CheckSlotAvailability(dateTime, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}

	if !available {
		return map[string]interface{}{
			"success": false,
			"error":   "This time slot is already booked. Please choose another time.",
		}, nil
	}

	purpose, _ := args["purpose"].(string)
	notes, _ := args["notes"].(string)

	appointment := &models.Appointment{
		ID:        uuid.New().String(),
		UserPhone: e.userPhone,
		UserName:  e.userName,
		DateTime:  dateTime,
		Duration:  duration,
		Purpose:   purpose,
		Notes:     notes,
		Status:    models.StatusBooked,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := database.DB.CreateAppointment(appointment); err != nil {
		return nil, fmt.Errorf("failed to book appointment: %w", err)
	}

	return map[string]interface{}{
		"success":        true,
		"appointment_id": appointment.ID,
		"date_time":      appointment.DateTime.Format("Monday, January 2, 2006 at 3:04 PM"),
		"duration":       appointment.Duration,
		"purpose":        appointment.Purpose,
		"message":        fmt.Sprintf("Appointment successfully booked for %s", appointment.DateTime.Format("Monday, January 2, 2006 at 3:04 PM")),
	}, nil
}

func (e *ToolExecutor) retrieveAppointments(args map[string]interface{}) (interface{}, error) {
	if e.userPhone == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "User not identified. Please identify the user first by asking for their phone number.",
		}, nil
	}

	retrieveType, _ := args["type"].(string)
	if retrieveType == "" {
		retrieveType = "upcoming"
	}

	var appointments []models.Appointment
	var err error

	if retrieveType == "upcoming" {
		appointments, err = database.DB.GetUpcomingAppointments(e.userPhone)
	} else {
		appointments, err = database.DB.GetAppointmentsByPhone(e.userPhone)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve appointments: %w", err)
	}

	formattedAppointments := make([]map[string]interface{}, len(appointments))
	for i, apt := range appointments {
		formattedAppointments[i] = map[string]interface{}{
			"id":        apt.ID,
			"date_time": apt.DateTime.Format("Monday, January 2, 2006 at 3:04 PM"),
			"duration":  apt.Duration,
			"purpose":   apt.Purpose,
			"status":    apt.Status,
			"notes":     apt.Notes,
		}
	}

	message := fmt.Sprintf("Found %d %s appointment(s)", len(appointments), retrieveType)
	if len(appointments) == 0 {
		message = fmt.Sprintf("No %s appointments found", retrieveType)
	}

	return map[string]interface{}{
		"success":      true,
		"appointments": formattedAppointments,
		"count":        len(appointments),
		"type":         retrieveType,
		"message":      message,
	}, nil
}

func (e *ToolExecutor) cancelAppointment(args map[string]interface{}) (interface{}, error) {
	if e.userPhone == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "User not identified. Please identify the user first.",
		}, nil
	}

	appointmentID, ok := args["appointment_id"].(string)
	if !ok || appointmentID == "" {
		return nil, fmt.Errorf("appointment_id is required")
	}

	appointment, err := database.DB.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	if appointment == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Appointment not found",
		}, nil
	}

	// Verify ownership
	if appointment.UserPhone != e.userPhone {
		return map[string]interface{}{
			"success": false,
			"error":   "You can only cancel your own appointments",
		}, nil
	}

	if appointment.Status == models.StatusCancelled {
		return map[string]interface{}{
			"success": false,
			"error":   "Appointment is already cancelled",
		}, nil
	}

	reason, _ := args["reason"].(string)

	appointment.Status = models.StatusCancelled
	if reason != "" {
		appointment.Notes = fmt.Sprintf("%s\nCancellation reason: %s", appointment.Notes, reason)
	}

	if err := database.DB.UpdateAppointment(appointment); err != nil {
		return nil, fmt.Errorf("failed to cancel appointment: %w", err)
	}

	return map[string]interface{}{
		"success":        true,
		"appointment_id": appointmentID,
		"date_time":      appointment.DateTime.Format("Monday, January 2, 2006 at 3:04 PM"),
		"message":        fmt.Sprintf("Appointment on %s has been cancelled", appointment.DateTime.Format("Monday, January 2, 2006 at 3:04 PM")),
	}, nil
}

func (e *ToolExecutor) modifyAppointment(args map[string]interface{}) (interface{}, error) {
	if e.userPhone == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "User not identified. Please identify the user first.",
		}, nil
	}

	appointmentID, ok := args["appointment_id"].(string)
	if !ok || appointmentID == "" {
		return nil, fmt.Errorf("appointment_id is required")
	}

	appointment, err := database.DB.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	if appointment == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "Appointment not found",
		}, nil
	}

	// Verify ownership
	if appointment.UserPhone != e.userPhone {
		return map[string]interface{}{
			"success": false,
			"error":   "You can only modify your own appointments",
		}, nil
	}

	if appointment.Status == models.StatusCancelled {
		return map[string]interface{}{
			"success": false,
			"error":   "Cannot modify a cancelled appointment",
		}, nil
	}

	modified := false
	changes := []string{}

	// Handle new date_time
	if newDateTimeStr, ok := args["new_date_time"].(string); ok && newDateTimeStr != "" {
		newDateTime, err := time.Parse(time.RFC3339, newDateTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid new_date_time format")
		}

		if newDateTime.Before(time.Now()) {
			return map[string]interface{}{
				"success": false,
				"error":   "Cannot reschedule to a past time",
			}, nil
		}

		// Check availability for new time
		duration := appointment.Duration
		if newDur, ok := args["new_duration"].(float64); ok {
			duration = int(newDur)
		}

		available, err := database.DB.CheckSlotAvailability(newDateTime, duration)
		if err != nil {
			return nil, fmt.Errorf("failed to check availability: %w", err)
		}

		if !available {
			return map[string]interface{}{
				"success": false,
				"error":   "The new time slot is not available",
			}, nil
		}

		appointment.DateTime = newDateTime
		modified = true
		changes = append(changes, fmt.Sprintf("rescheduled to %s", newDateTime.Format("Monday, January 2, 2006 at 3:04 PM")))
	}

	// Handle new duration
	if newDur, ok := args["new_duration"].(float64); ok && int(newDur) != appointment.Duration {
		appointment.Duration = int(newDur)
		modified = true
		changes = append(changes, fmt.Sprintf("duration changed to %d minutes", appointment.Duration))
	}

	// Handle new purpose
	if newPurpose, ok := args["new_purpose"].(string); ok && newPurpose != "" {
		appointment.Purpose = newPurpose
		modified = true
		changes = append(changes, "purpose updated")
	}

	// Handle new notes
	if newNotes, ok := args["new_notes"].(string); ok && newNotes != "" {
		appointment.Notes = newNotes
		modified = true
		changes = append(changes, "notes updated")
	}

	if !modified {
		return map[string]interface{}{
			"success": false,
			"error":   "No changes specified",
		}, nil
	}

	if err := database.DB.UpdateAppointment(appointment); err != nil {
		return nil, fmt.Errorf("failed to modify appointment: %w", err)
	}

	return map[string]interface{}{
		"success":        true,
		"appointment_id": appointmentID,
		"changes":        changes,
		"new_date_time":  appointment.DateTime.Format("Monday, January 2, 2006 at 3:04 PM"),
		"new_duration":   appointment.Duration,
		"message":        fmt.Sprintf("Appointment modified: %v", changes),
	}, nil
}

func (e *ToolExecutor) endConversation(args map[string]interface{}) (interface{}, error) {
	reason, _ := args["reason"].(string)

	return map[string]interface{}{
		"success":     true,
		"action":      "end_conversation",
		"reason":      reason,
		"message":     "Conversation ended",
		"should_end":  true,
	}, nil
}

// Helper function to normalize phone numbers
func normalizePhoneNumber(phone string) string {
	// Remove all non-digit characters except leading +
	var result []rune
	for i, r := range phone {
		if r == '+' && i == 0 {
			result = append(result, r)
		} else if r >= '0' && r <= '9' {
			result = append(result, r)
		}
	}
	return string(result)
}
