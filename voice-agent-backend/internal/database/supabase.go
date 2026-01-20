package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/models"
)

type SupabaseClient struct {
	URL    string
	APIKey string
	client *http.Client
}

var DB *SupabaseClient

func Initialize(cfg *config.Config) error {
	DB = &SupabaseClient{
		URL:    cfg.SupabaseURL,
		APIKey: cfg.SupabaseAPIKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	return nil
}

func (s *SupabaseClient) doRequest(method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", s.URL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("supabase error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// User operations
func (s *SupabaseClient) GetUserByPhone(phone string) (*models.User, error) {
	var users []models.User
	encodedPhone := url.QueryEscape(phone)
	endpoint := fmt.Sprintf("users?phone_number=eq.%s", encodedPhone)

	if err := s.doRequest("GET", endpoint, nil, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return &users[0], nil
}

func (s *SupabaseClient) CreateUser(user *models.User) error {
	var result []models.User
	if err := s.doRequest("POST", "users", user, &result); err != nil {
		return err
	}
	if len(result) > 0 {
		*user = result[0]
	}
	return nil
}

func (s *SupabaseClient) UpdateUser(user *models.User) error {
	endpoint := fmt.Sprintf("users?id=eq.%s", user.ID)
	return s.doRequest("PATCH", endpoint, user, nil)
}

// Appointment operations
func (s *SupabaseClient) CreateAppointment(apt *models.Appointment) error {
	var result []models.Appointment
	if err := s.doRequest("POST", "appointments", apt, &result); err != nil {
		return err
	}
	if len(result) > 0 {
		*apt = result[0]
	}
	return nil
}

func (s *SupabaseClient) GetAppointmentsByPhone(phone string) ([]models.Appointment, error) {
	var appointments []models.Appointment
	endpoint := fmt.Sprintf("appointments?user_phone=eq.%s&order=date_time.desc", phone)

	if err := s.doRequest("GET", endpoint, nil, &appointments); err != nil {
		return nil, err
	}

	return appointments, nil
}

func (s *SupabaseClient) GetAppointmentByID(id string) (*models.Appointment, error) {
	var appointments []models.Appointment
	endpoint := fmt.Sprintf("appointments?id=eq.%s", id)

	if err := s.doRequest("GET", endpoint, nil, &appointments); err != nil {
		return nil, err
	}

	if len(appointments) == 0 {
		return nil, nil
	}

	return &appointments[0], nil
}

func (s *SupabaseClient) UpdateAppointment(apt *models.Appointment) error {
	endpoint := fmt.Sprintf("appointments?id=eq.%s", apt.ID)
	apt.UpdatedAt = time.Now()
	return s.doRequest("PATCH", endpoint, apt, nil)
}

func (s *SupabaseClient) CheckSlotAvailability(dateTime time.Time, duration int) (bool, error) {
	// Check if there's any overlapping appointment
	// An appointment overlaps if it starts before the requested slot ends
	// and its end time (date_time + duration) is after the requested slot starts
	requestedStart := dateTime
	requestedEnd := dateTime.Add(time.Duration(duration) * time.Minute)

	// Get appointments that start before the requested end time and are booked
	// We'll need to check overlap in code since PostgREST doesn't easily support
	// checking if (date_time + duration) overlaps with our range
	var appointments []models.Appointment
	endpoint := fmt.Sprintf(
		"appointments?status=eq.booked&date_time=lt.%s&order=date_time.desc",
		requestedEnd.Format(time.RFC3339),
	)

	if err := s.doRequest("GET", endpoint, nil, &appointments); err != nil {
		return false, err
	}

	// Check if any appointment overlaps with the requested slot
	for _, apt := range appointments {
		aptStart := apt.DateTime
		aptEnd := aptStart.Add(time.Duration(apt.Duration) * time.Minute)

		// Check for overlap: appointments overlap if one starts before the other ends
		if aptStart.Before(requestedEnd) && aptEnd.After(requestedStart) {
			return false, nil
		}
	}

	return true, nil
}

// GetUpcomingAppointments gets all upcoming appointments for a user (filters locally, no DB timestamp issues)
func (s *SupabaseClient) GetUpcomingAppointments(phone string) ([]models.Appointment, error) {
	// Get all appointments for the user
	appointments, err := s.GetAppointmentsByPhone(phone)
	if err != nil {
		return nil, err
	}

	// Filter for upcoming appointments only (done in Go, not in DB query)
	now := time.Now()
	var upcoming []models.Appointment
	for _, apt := range appointments {
		if apt.DateTime.After(now) && apt.Status == "booked" {
			upcoming = append(upcoming, apt)
		}
	}

	return upcoming, nil
}

// GetUpcomingAppointmentsInWindow gets all upcoming appointments within a time window
func (s *SupabaseClient) GetUpcomingAppointmentsInWindow(from time.Time, to time.Time) ([]models.Appointment, error) {
	var appointments []models.Appointment
	fromStr := from.Format(time.RFC3339)
	toStr := to.Format(time.RFC3339)
	endpoint := fmt.Sprintf(
		"appointments?status=eq.booked&date_time=gte.%s&date_time=lte.%s&order=date_time.asc",
		fromStr, toStr,
	)

	if err := s.doRequest("GET", endpoint, nil, &appointments); err != nil {
		return nil, err
	}

	return appointments, nil
}

// Call Summary operations
func (s *SupabaseClient) SaveCallSummary(summary *models.CallSummary) error {
	var result []models.CallSummary
	if err := s.doRequest("POST", "call_summaries", summary, &result); err != nil {
		return err
	}
	if len(result) > 0 {
		*summary = result[0]
	}
	return nil
}

func (s *SupabaseClient) GetCallSummariesByPhone(phone string) ([]models.CallSummary, error) {
	var summaries []models.CallSummary
	endpoint := fmt.Sprintf("call_summaries?user_phone=eq.%s&order=created_at.desc", phone)

	if err := s.doRequest("GET", endpoint, nil, &summaries); err != nil {
		return nil, err
	}

	return summaries, nil
}
