package dto

import "time"

// CheckInResponse represents a check-in API response
type CheckInResponse struct {
	SessionID    string    `json:"session_id"`
	CheckInTime  time.Time `json:"check_in_time"`
	CheckOutTime time.Time `json:"expected_check_out_time"`
	WorkHours    int       `json:"work_hours"` // in minutes
}

// CheckOutResponse represents a check-out API response
type CheckOutResponse struct {
	SessionID       string    `json:"session_id"`
	CheckInTime     time.Time `json:"check_in_time"`
	CheckOutTime    time.Time `json:"check_out_time"`
	OvertimeMinutes int       `json:"overtime_minutes"`
}

// ConfigResponse represents a configuration response
type ConfigResponse struct {
	WorkHours          int    `json:"work_hours"` // in minutes
	CheckInAPIURL      string `json:"check_in_api_url"`
	AutoFetchEnabled   bool   `json:"auto_fetch_enabled"`
	PAuth              string `json:"p_auth"`
	PRToken            string `json:"p_rtoken"`
	CheckInWebhookURL  string `json:"check_in_webhook_url"`
	CheckOutWebhookURL string `json:"check_out_webhook_url"`
}

// StatusResponse represents the current work status
type StatusResponse struct {
	HasCheckedIn     bool       `json:"has_checked_in"`
	CheckInTime      *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime     *time.Time `json:"check_out_time,omitempty"`
	ExpectedCheckOut *time.Time `json:"expected_check_out_time,omitempty"`
	CurrentTime      time.Time  `json:"current_time"`
	WorkHours        int        `json:"work_hours"` // in minutes
	IsCheckOutTime   bool       `json:"is_check_out_time"`
	OvertimeMinutes  int        `json:"overtime_minutes"`
}

// TodayCheckInResponse represents a response for today's check-in status
type TodayCheckInResponse struct {
	HasCheckedIn     bool       `json:"has_checked_in"`
	CheckInTime      *time.Time `json:"check_in_time,omitempty"`
	CanAutoFetch     bool       `json:"can_auto_fetch"`
	AutoFetchEnabled bool       `json:"auto_fetch_enabled"`
	APIError         string     `json:"api_error,omitempty"`
}

// MonthlyStatsResponse represents monthly overtime statistics
type MonthlyStatsResponse struct {
	CurrentMonth MonthStats `json:"current_month"`
	LastMonth    MonthStats `json:"last_month"`
}

// MonthStats represents statistics for a single month
type MonthStats struct {
	YearMonth       string `json:"year_month"`
	TotalDays       int    `json:"total_days"`
	CheckedOutDays  int    `json:"checked_out_days"`
	OvertimeMinutes int    `json:"overtime_minutes"`
}
