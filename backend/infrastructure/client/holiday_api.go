package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	HolidayAPIURL      = "http://api.haoshenqi.top/holiday/today"
	HolidayStatusWork  = "工作"
	HolidayStatusRest  = "休息"
)

// HolidayResponse represents the holiday API response
type HolidayResponse struct {
	Code int `json:"code"`
	Data struct {
		Date   string `json:"date"`   // "2025-10-13"
		Status string `json:"status"` // "休息" or "工作"
	} `json:"data"`
}

// HolidayAPIClient handles communication with the holiday API
type HolidayAPIClient struct {
	httpClient *http.Client
	apiURL     string
}

// NewHolidayAPIClient creates a new holiday API client
func NewHolidayAPIClient() *HolidayAPIClient {
	return &HolidayAPIClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: HolidayAPIURL,
	}
}

// IsHoliday checks if today is a holiday (休息日)
func (c *HolidayAPIClient) IsHoliday() (bool, error) {
	log.Printf("[Holiday API] Checking holiday status: %s", c.apiURL)

	resp, err := c.httpClient.Get(c.apiURL)
	if err != nil {
		return false, fmt.Errorf("holiday API request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[Holiday API] Response: %s", string(bodyBytes))

	var holidayResp HolidayResponse
	if err := json.Unmarshal(bodyBytes, &holidayResp); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	isHoliday := holidayResp.Data.Status == HolidayStatusRest
	log.Printf("[Holiday API] Status: %s (is holiday: %v)", holidayResp.Data.Status, isHoliday)

	return isHoliday, nil
}
