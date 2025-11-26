package cronjob

import (
	"log"
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
		log.Printf("[Scheduler] Failed to load Asia/Shanghai timezone, using Local: %v", err)
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
	log.Println("[Scheduler] Starting cronjob scheduler...")

	// Task 1: Check-in reminder at 9:55 AM (China time)
	if _, err := s.cron.AddFunc("55 9 * * *", s.checkInReminder); err != nil {
		log.Printf("[Scheduler] Failed to add check-in reminder job: %v", err)
	} else {
		log.Println("[Scheduler] Added check-in reminder: 9:55 AM daily")
	}

	// Task 2: Check-out reminder at 8:30 PM (China time)
	if _, err := s.cron.AddFunc("30 20 * * *", s.checkOutReminder); err != nil {
		log.Printf("[Scheduler] Failed to add check-out reminder (8:30 PM): %v", err)
	} else {
		log.Println("[Scheduler] Added check-out reminder: 8:30 PM daily")
	}

	// Task 3: Check-out reminder at 9:30 PM (China time)
	if _, err := s.cron.AddFunc("30 21 * * *", s.checkOutReminder); err != nil {
		log.Printf("[Scheduler] Failed to add check-out reminder (9:30 PM): %v", err)
	} else {
		log.Println("[Scheduler] Added check-out reminder: 9:30 PM daily")
	}

	s.cron.Start()
	log.Println("[Scheduler] Cronjob scheduler started successfully")
}

// Stop stops the cron scheduler
func (s *Scheduler) Stop() {
	log.Println("[Scheduler] Stopping cronjob scheduler...")
	s.cron.Stop()
	log.Println("[Scheduler] Cronjob scheduler stopped")
}

// checkInReminder sends a reminder to check in if not already done
func (s *Scheduler) checkInReminder() {
	log.Println("[CheckInReminder] Running task...")

	config, err := s.store.GetConfig()
	if err != nil {
		log.Printf("[CheckInReminder] Failed to get config: %v", err)
		return
	}
	if config.CheckInWebhookURL == "" {
		log.Println("[CheckInReminder] Webhook URL not configured, skipping")
		return
	}

	// Step 1: Check if today is a holiday
	if s.isHolidayToday() {
		log.Println("[CheckInReminder] Today is a holiday, skipping")
		return
	}

	// Step 2: Check if already checked in via HR API
	today := time.Now().Format("2006-01-02")
	if s.hasCheckedIn(config, today) {
		log.Println("[CheckInReminder] Already checked in, skipping")
		return
	}

	// Step 3: Send ntfy notification
	log.Printf("[CheckInReminder] Sending notification to: %s", config.CheckInWebhookURL)
	if err := s.webhookClient.Alarm(config.CheckInWebhookURL, "⏰ Time to check in! Don't forget to clock in for work."); err != nil {
		log.Printf("[CheckInReminder] Failed to send notification: %v", err)
	} else {
		log.Println("[CheckInReminder] Notification sent successfully")
	}
}

// checkOutReminder sends a reminder to check out if not already done
func (s *Scheduler) checkOutReminder() {
	log.Println("[CheckOutReminder] Running task...")

	config, err := s.store.GetConfig()
	if err != nil {
		log.Printf("[CheckOutReminder] Failed to get config: %v", err)
		return
	}
	if config.CheckOutWebhookURL == "" {
		log.Println("[CheckOutReminder] Webhook URL not configured, skipping")
		return
	}

	// Step 1: Check if today is a holiday
	if s.isHolidayToday() {
		log.Println("[CheckOutReminder] Today is a holiday, skipping")
		return
	}

	// Step 2: Check if already checked out via HR API
	today := time.Now().Format("2006-01-02")
	if s.hasCheckedOut(config, today) {
		log.Println("[CheckOutReminder] Already checked out, skipping")
		return
	}

	// Step 3: Send ntfy notification
	log.Printf("[CheckOutReminder] Sending notification to: %s", config.CheckOutWebhookURL)
	if err := s.webhookClient.Alarm(config.CheckOutWebhookURL, "✅ Time to check out! Remember to clock out from work."); err != nil {
		log.Printf("[CheckOutReminder] Failed to send notification: %v", err)
	} else {
		log.Println("[CheckOutReminder] Notification sent successfully")
	}
}

// isHolidayToday checks if today is a holiday
func (s *Scheduler) isHolidayToday() bool {
	isHoliday, err := s.holidayClient.IsHoliday()
	if err != nil {
		log.Printf("[Scheduler] Failed to check holiday status: %v", err)
		return false // Assume not a holiday on error
	}
	return isHoliday
}

// hasCheckedIn checks if already checked in today via HR API
func (s *Scheduler) hasCheckedIn(config *domain.WorkConfig, date string) bool {
	if !config.HasAPIConfig() {
		return false
	}

	hasCheckedIn, _, err := s.attendanceProvider.FetchAttendanceStatus(config, date)
	if err != nil {
		log.Printf("[Scheduler] Failed to check HR check-in status: %v", err)
		return false // Assume not checked in on error
	}

	return hasCheckedIn
}

// hasCheckedOut checks if already checked out today via HR API
func (s *Scheduler) hasCheckedOut(config *domain.WorkConfig, date string) bool {
	if !config.HasAPIConfig() {
		return false
	}

	_, hasCheckedOut, err := s.attendanceProvider.FetchAttendanceStatus(config, date)
	if err != nil {
		log.Printf("[Scheduler] Failed to check HR check-out status: %v", err)
		return false // Assume not checked out on error
	}

	return hasCheckedOut
}
