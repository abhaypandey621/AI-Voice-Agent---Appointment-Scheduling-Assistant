package reminder

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/voice-agent/backend/internal/config"
	"github.com/voice-agent/backend/internal/database"
	"github.com/voice-agent/backend/internal/models"
)

// ReminderType represents the type of reminder
type ReminderType string

const (
	ReminderType24Hour ReminderType = "24_hours"
	ReminderType1Hour  ReminderType = "1_hour"
	ReminderTypeOnDay  ReminderType = "on_day"
)

// ReminderService manages appointment reminders
type ReminderService struct {
	config    *config.Config
	ticker    *time.Ticker
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	reminders map[string]*ReminderRecord // key: appointmentID
	callbacks map[ReminderType]func(*models.Appointment, ReminderType)
}

// ReminderRecord tracks reminder state for an appointment
type ReminderRecord struct {
	AppointmentID string
	RemindersSent map[ReminderType]bool
	NextCheck     time.Time
}

// NewReminderService creates a new reminder service
func NewReminderService(cfg *config.Config) *ReminderService {
	ctx, cancel := context.WithCancel(context.Background())

	rs := &ReminderService{
		config:    cfg,
		ticker:    time.NewTicker(1 * time.Minute), // Check every minute
		ctx:       ctx,
		cancel:    cancel,
		reminders: make(map[string]*ReminderRecord),
		callbacks: make(map[ReminderType]func(*models.Appointment, ReminderType)),
	}

	// Start the reminder loop
	go rs.reminderLoop()

	return rs
}

// RegisterCallback registers a callback for a reminder type
func (rs *ReminderService) RegisterCallback(reminderType ReminderType, callback func(*models.Appointment, ReminderType)) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.callbacks[reminderType] = callback
}

// AddAppointment adds an appointment to be tracked for reminders
func (rs *ReminderService) AddAppointment(appointment *models.Appointment) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.reminders[appointment.ID] = &ReminderRecord{
		AppointmentID: appointment.ID,
		RemindersSent: make(map[ReminderType]bool),
		NextCheck:     time.Now(),
	}
}

// RemoveAppointment stops tracking reminders for an appointment
func (rs *ReminderService) RemoveAppointment(appointmentID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.reminders, appointmentID)
}

// reminderLoop checks for appointments needing reminders
func (rs *ReminderService) reminderLoop() {
	for {
		select {
		case <-rs.ctx.Done():
			return
		case <-rs.ticker.C:
			rs.checkReminders()
		}
	}
}

// checkReminders checks all tracked appointments for pending reminders
func (rs *ReminderService) checkReminders() {
	rs.mu.RLock()
	remindersCopy := make(map[string]*ReminderRecord)
	for k, v := range rs.reminders {
		remindersCopy[k] = v
	}
	rs.mu.RUnlock()

	now := time.Now()

	for appointmentID, record := range remindersCopy {
		// Fetch appointment details
		appointment, err := database.DB.GetAppointmentByID(appointmentID)
		if err != nil {
			log.Printf("Failed to fetch appointment %s: %v", appointmentID, err)
			continue
		}

		if appointment == nil || appointment.Status != models.StatusBooked {
			rs.RemoveAppointment(appointmentID)
			continue
		}

		timeUntilAppointment := time.Until(appointment.DateTime)

		// Check for 24-hour reminder
		if !record.RemindersSent[ReminderType24Hour] && timeUntilAppointment > 0 && timeUntilAppointment <= 24*time.Hour+1*time.Minute {
			rs.sendReminder(appointment, ReminderType24Hour)
			rs.markReminderSent(appointmentID, ReminderType24Hour)
		}

		// Check for 1-hour reminder
		if !record.RemindersSent[ReminderType1Hour] && timeUntilAppointment > 0 && timeUntilAppointment <= 1*time.Hour+1*time.Minute {
			rs.sendReminder(appointment, ReminderType1Hour)
			rs.markReminderSent(appointmentID, ReminderType1Hour)
		}

		// Check for on-day reminder
		if !record.RemindersSent[ReminderTypeOnDay] && timeUntilAppointment > 0 && timeUntilAppointment <= 24*time.Hour && isNextDay(now, appointment.DateTime) {
			rs.sendReminder(appointment, ReminderTypeOnDay)
			rs.markReminderSent(appointmentID, ReminderTypeOnDay)
		}

		// Remove if appointment has passed
		if timeUntilAppointment < 0 {
			rs.RemoveAppointment(appointmentID)
		}
	}
}

// sendReminder sends a reminder to the user
func (rs *ReminderService) sendReminder(appointment *models.Appointment, reminderType ReminderType) {
	rs.mu.RLock()
	callback, exists := rs.callbacks[reminderType]
	rs.mu.RUnlock()

	if exists && callback != nil {
		callback(appointment, reminderType)
	}

	// Log reminder
	log.Printf("Reminder sent: %s for appointment %s (user: %s)", reminderType, appointment.ID, appointment.UserPhone)
}

// markReminderSent marks a reminder as sent
func (rs *ReminderService) markReminderSent(appointmentID string, reminderType ReminderType) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if record, exists := rs.reminders[appointmentID]; exists {
		record.RemindersSent[reminderType] = true
	}
}

// isNextDay checks if two times are on different calendar days
func isNextDay(now, appointmentTime time.Time) bool {
	return now.YearDay() != appointmentTime.YearDay() || now.Year() != appointmentTime.Year()
}

// Stop stops the reminder service
func (rs *ReminderService) Stop() {
	rs.cancel()
	rs.ticker.Stop()
}

// GetReminderStatus returns the status of reminders for an appointment
func (rs *ReminderService) GetReminderStatus(appointmentID string) map[string]interface{} {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	if record, exists := rs.reminders[appointmentID]; exists {
		return map[string]interface{}{
			"appointment_id": appointmentID,
			"reminders_sent": record.RemindersSent,
			"next_check":     record.NextCheck,
		}
	}

	return nil
}

// LoadPendingAppointments loads future appointments for reminders
func (rs *ReminderService) LoadPendingAppointments() error {
	now := time.Now()
	futureDate := now.Add(30 * 24 * time.Hour)
	appointments, err := database.DB.GetUpcomingAppointmentsInWindow(now, futureDate)
	if err != nil {
		return fmt.Errorf("failed to load pending appointments: %w", err)
	}

	for _, apt := range appointments {
		rs.AddAppointment(&apt)
	}

	log.Printf("Loaded %d pending appointments for reminders", len(appointments))
	return nil
}
