package dto

import "time"

// CheckInRequest represents a check-in API request
type CheckInRequest struct {
	CheckInTime time.Time `json:"check_in_time"`
}

// CheckOutRequest represents a check-out API request
type CheckOutRequest struct {
	CheckOutTime time.Time `json:"check_out_time"`
}

// ConfigRequest represents a configuration update request
type ConfigRequest struct {
	WorkHours          int    `json:"work_hours"` // in minutes
	CheckInAPIURL      string `json:"check_in_api_url"`
	AutoFetchEnabled   bool   `json:"auto_fetch_enabled"`
	PAuth              string `json:"p_auth"`
	PRToken            string `json:"p_rtoken"`
	CheckInWebhookURL  string `json:"check_in_webhook_url"`
	CheckOutWebhookURL string `json:"check_out_webhook_url"`
}

// TodayCheckInRequest represents a request to get/auto-fetch today's check-in
type TodayCheckInRequest struct {
	Date      string `json:"date"` // YYYY-MM-DD format
	ReCheckIn bool   `json:"re_check_in"`
}
