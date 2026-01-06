package client

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	HolidayAPIURL     = "http://api.haoshenqi.top/holiday/today"
	HolidayStatusWork = "工作"
	HolidayStatusRest = "休息"
)

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
	slog.Info("[Holiday API] Checking holiday status", "url", c.apiURL)

	resp, err := c.httpClient.Get(c.apiURL)
	if err != nil {
		return false, fmt.Errorf("holiday API request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	slog.Info("[Holiday API] Response", "body", string(bodyBytes))

	status := string(bodyBytes)

	isHoliday := status == HolidayStatusRest
	slog.Info("[Holiday API] Status", "status", status, "is_holiday", isHoliday)

	return isHoliday, nil
}
