// Package domain contains the core business entities and repository interfaces.
package domain

import "time"

// Constants for work time calculations
const (
	MinutesPerHour           = 60
	StandardWorkHours        = 8
	StandardWorkMinutes      = StandardWorkHours * MinutesPerHour // 480 minutes
	OvertimeThresholdHours   = 10
	OvertimeThresholdMinutes = OvertimeThresholdHours * MinutesPerHour // 600 minutes
	MaxWorkHoursPerDay       = 24
	MaxWorkMinutesPerDay     = MaxWorkHoursPerDay * MinutesPerHour // 1440 minutes
)

// WorkSession represents a single work session for a specific date
type WorkSession struct {
	ID        string
	Date      string // YYYY-MM-DD format
	CheckIn   time.Time
	CheckOut  *time.Time
	WorkHours int // Expected work hours in minutes
}

// HasCheckedOut returns true if the session has a check-out time
func (s *WorkSession) HasCheckedOut() bool {
	return s.CheckOut != nil
}

// CalculateExpectedCheckOut calculates when the user should check out
func (s *WorkSession) CalculateExpectedCheckOut() time.Time {
	return s.CheckIn.Add(time.Duration(s.WorkHours) * time.Minute)
}

// CalculateActualWorkMinutes calculates how many minutes were actually worked
func (s *WorkSession) CalculateActualWorkMinutes() int {
	if !s.HasCheckedOut() {
		return 0
	}
	return int(s.CheckOut.Sub(s.CheckIn).Minutes())
}

// CalculateOvertime calculates overtime relative to 10-hour threshold
// Returns positive for overtime (worked > 10h), negative for under-time
func (s *WorkSession) CalculateOvertime() int {
	if !s.HasCheckedOut() {
		return 0
	}
	actualMinutes := s.CalculateActualWorkMinutes()
	return actualMinutes - OvertimeThresholdMinutes
}

// WorkConfig represents the global work configuration
type WorkConfig struct {
	ID                 string `json:"id"`
	DefaultWorkHours   int    `json:"default_work_hours"`    // in minutes, default 480 (8 hours)
	CheckInAPIURL      string `json:"check_in_api_url"`      // HR API endpoint to fetch attendance
	AutoFetchEnabled   bool   `json:"auto_fetch_enabled"`    // whether to auto-fetch from HR API
	PAuth              string `json:"p_auth"`                // P-Auth header for HR API
	PRToken            string `json:"p_rtoken"`              // P-Rtoken header for HR API
	CheckInWebhookURL  string `json:"check_in_webhook_url"`  // Webhook for check-in reminders
	CheckOutWebhookURL string `json:"check_out_webhook_url"` // Webhook for check-out reminders
}

// HasAPIConfig returns true if HR API is configured
func (c *WorkConfig) HasAPIConfig() bool {
	return c.CheckInAPIURL != "" &&
		c.PAuth != "" &&
		c.PRToken != ""
}

// ShouldAutoFetch returns true if auto-fetch is enabled and configured
func (c *WorkConfig) ShouldAutoFetch() bool {
	return c.AutoFetchEnabled && c.HasAPIConfig()
}

// MonthlyStats represents aggregated statistics for a month
type MonthlyStats struct {
	YearMonth       string // YYYY-MM format
	TotalDays       int
	CheckedOutDays  int
	OvertimeMinutes int
}

// CalculateStats aggregates statistics from multiple sessions
func CalculateStats(sessions []*WorkSession, yearMonth string) *MonthlyStats {
	stats := &MonthlyStats{
		YearMonth:       yearMonth,
		TotalDays:       len(sessions),
		CheckedOutDays:  0,
		OvertimeMinutes: 0,
	}

	for _, session := range sessions {
		if session.HasCheckedOut() {
			stats.CheckedOutDays++
			overtime := session.CalculateOvertime()
			if overtime > 0 {
				stats.OvertimeMinutes += overtime
			}
		}
	}

	return stats
}
