package cronjob

import (
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/simon0-o/offline_me/backend/domain"
	"github.com/simon0-o/offline_me/backend/infrastructure/client"
)

// Scheduler handles scheduled tasks like reminders
type Scheduler struct {
	cron               *cron.Cron
	store              domain.Repository
	attendanceProvider domain.AttendanceProvider
	holidayClient      *client.HolidayAPIClient
	webhookClient      *client.WebhookClient
}

// NewScheduler creates a new scheduler instance
func NewScheduler(store domain.Repository) *Scheduler {
	// Use Asia/Shanghai timezone for cron jobs
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		slog.Info("[Scheduler] Failed to load Asia/Shanghai timezone, using Local", "error", err)
		location = time.Local
	}

	return &Scheduler{
		cron:               cron.New(cron.WithLocation(location)),
		store:              store,
		attendanceProvider: client.NewHRAPIClient(),
		holidayClient:      client.NewHolidayAPIClient(),
		webhookClient:      client.NewWebhookClient(),
	}
}

// Start starts the cron scheduler
func (s *Scheduler) Start() {
	slog.Info("[Scheduler] Starting cronjob scheduler...")

	// Task 1: Check-in reminder at 9:55 AM (China time)
	if _, err := s.cron.AddFunc("55 9 * * *", s.checkInReminder); err != nil {
		slog.Info("[Scheduler] Failed to add check-in reminder job", "error", err)
	} else {
		slog.Info("[Scheduler] Added check-in reminder: 9:55 AM daily")
	}

	// Task 2: Check-out reminder at 8:30 PM (China time)
	if _, err := s.cron.AddFunc("30 20 * * *", s.checkOutReminder); err != nil {
		slog.Info("[Scheduler] Failed to add check-out reminder (8:30 PM)", "error", err)
	} else {
		slog.Info("[Scheduler] Added check-out reminder: 8:30 PM daily")
	}

	// Task 3: Check-out reminder at 9:30 PM (China time)
	if _, err := s.cron.AddFunc("30 21 * * *", s.checkOutReminder); err != nil {
		slog.Info("[Scheduler] Failed to add check-out reminder (9:30 PM)", "error", err)
	} else {
		slog.Info("[Scheduler] Added check-out reminder: 9:30 PM daily")
	}

	s.cron.Start()
	slog.Info("[Scheduler] Cronjob scheduler started successfully")
}

// Stop stops the cron scheduler
func (s *Scheduler) Stop() {
	slog.Info("[Scheduler] Stopping cronjob scheduler...")
	s.cron.Stop()
	slog.Info("[Scheduler] Cronjob scheduler stopped")
}

// checkInReminder sends a reminder to check in if not already done
func (s *Scheduler) checkInReminder() {
	slog.Info("[CheckInReminder] Running task...")

	config, err := s.store.GetConfig()
	if err != nil {
		slog.Info("[CheckInReminder] Failed to get config", "error", err)
		return
	}
	if config.CheckInWebhookURL == "" {
		slog.Info("[CheckInReminder] Webhook URL not configured, skipping")
		return
	}

	// Step 1: Check if today is a holiday
	if s.isHolidayToday() {
		slog.Info("[CheckInReminder] Today is a holiday, skipping")
		return
	}

	// Step 2: Check if already checked in via HR API
	today := time.Now().Format("2006-01-02")
	if s.hasCheckedIn(config, today) {
		slog.Info("[CheckInReminder] Already checked in, skipping")
		return
	}

	// Step 3: Send ntfy notification
	slog.Info("[CheckInReminder] Sending notification", "url", config.CheckInWebhookURL)
	if err := s.webhookClient.Alarm(config.CheckInWebhookURL, "⏰ Time to check in! Don't forget to clock in for work."); err != nil {
		slog.Info("[CheckInReminder] Failed to send notification", "error", err)
	} else {
		slog.Info("[CheckInReminder] Notification sent successfully")
	}
}

// checkOutReminder sends a reminder to check out if not already done
func (s *Scheduler) checkOutReminder() {
	slog.Info("[CheckOutReminder] Running task...")

	config, err := s.store.GetConfig()
	if err != nil {
		slog.Info("[CheckOutReminder] Failed to get config", "error", err)
		return
	}
	if config.CheckOutWebhookURL == "" {
		slog.Info("[CheckOutReminder] Webhook URL not configured, skipping")
		return
	}

	// Step 1: Check if today is a holiday
	if s.isHolidayToday() {
		slog.Info("[CheckOutReminder] Today is a holiday, skipping")
		return
	}

	// Step 2: Check if already checked out via HR API
	today := time.Now().Format("2006-01-02")
	if s.hasCheckedOut(config, today) {
		slog.Info("[CheckOutReminder] Already checked out, skipping")
		return
	}

	// Step 3: Send ntfy notification
	slog.Info("[CheckOutReminder] Sending notification", "url", config.CheckOutWebhookURL)
	if err := s.webhookClient.Alarm(config.CheckOutWebhookURL, "✅ Time to check out! Remember to clock out from work."); err != nil {
		slog.Info("[CheckOutReminder] Failed to send notification", "error", err)
	} else {
		slog.Info("[CheckOutReminder] Notification sent successfully")
	}
}

// isHolidayToday checks if today is a holiday
func (s *Scheduler) isHolidayToday() bool {
	isHoliday, err := s.holidayClient.IsHoliday()
	if err != nil {
		slog.Error("[Scheduler] Failed to check holiday status", "error", err)
		return false // Assume not a holiday on error
	}
	return isHoliday
}

// hasCheckedIn checks if already checked in today via HR API
func (s *Scheduler) hasCheckedIn(config *domain.WorkConfig, date string) bool {
	if !config.HasAPIConfig() {
		return false
	}

	checkedIn, _, err := s.attendanceProvider.FetchAttendanceStatus(config, date)
	if err != nil {
		slog.Info("[Scheduler] Failed to check HR check-in status", "error", err)
		return false // Assume not checked in on error
	}

	return checkedIn != nil
}

// hasCheckedOut checks if already checked out today via HR API
func (s *Scheduler) hasCheckedOut(config *domain.WorkConfig, date string) bool {
	if !config.HasAPIConfig() {
		return false
	}

	checkedIn, checkedOut, err := s.attendanceProvider.FetchAttendanceStatus(config, date)
	if err != nil {
		slog.Info("[Scheduler] Failed to check HR check-out status", "error", err)
		return false // Assume not checked out on error
	}
	if checkedIn == nil || checkedOut == nil {
		return false
	}
	// update the work session with check-out time
	session := s.store.GetTodaySession(date)
	if session != nil && session.CheckOut == nil {
		session.CheckOut = checkedOut
		if err := s.store.SaveSession(session); err != nil {
			slog.Info("[Scheduler] Failed to save session", "error", err)
		}
	}

	expectedCheckOut := config.CalculateExpectedCheckOut(*checkedIn)
	return checkedOut.After(expectedCheckOut)
}
