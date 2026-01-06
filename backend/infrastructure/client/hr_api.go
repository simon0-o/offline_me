package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/simon0-o/offline_me/backend/domain"
)

// HRAttendanceInfo represents the HR API response structure
type HRAttendanceInfo struct {
	Code          string             `json:"code"`
	Message       string             `json:"message"`
	DetailMessage *string            `json:"detailMessage"`
	Success       bool               `json:"success"`
	Data          []AttendanceRecord `json:"data"`
}

// AttendanceRecord represents a single attendance record
type AttendanceRecord struct {
	AttendanceDate     string   `json:"attendanceDate"`   // "2025-10-13"
	FirstClockInTime   *string  `json:"firstClockInTime"` // "09:34" or null
	LastClockOutTime   *string  `json:"lastClockOutTime"` // "18:00" or null
	AnnotationType     *int     `json:"annotationType"`
	PointFlag          int      `json:"pointFlag"`
	LeaveMainIdList    *string  `json:"leaveMainIdList"`
	AbnormalMainIdList []string `json:"abnormalMainIdList"`
	ModifiTime         *string  `json:"modifiTime"`
}

// HRAPIClient handles communication with the HR attendance API
type HRAPIClient struct {
	httpClient *http.Client
}

// NewHRAPIClient creates a new HR API client and returns it as AttendanceProvider interface
func NewHRAPIClient() domain.AttendanceProvider {
	return &HRAPIClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchAttendanceStatus fetches attendance records for a specific date
func (c *HRAPIClient) FetchAttendanceStatus(config *domain.WorkConfig, date string) (checkedIn, checkedOut *time.Time, err error) {
	if !config.HasAPIConfig() {
		return nil, nil, fmt.Errorf("HR API not properly configured")
	}

	apiURL := c.buildAPIURL(config.CheckInAPIURL, date)
	req, err := c.createRequest(apiURL, config)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	var hrResponse HRAttendanceInfo
	if err := json.Unmarshal(bodyBytes, &hrResponse); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if hrResponse.Code != "200" || !hrResponse.Success {
		return nil, nil, fmt.Errorf("API error: %s", hrResponse.Message)
	}

	return c.extractCheckTime(hrResponse.Data, date)
}

// buildAPIURL constructs the API URL with the monthly parameter
func (c *HRAPIClient) buildAPIURL(baseURL, date string) string {
	if strings.Contains(baseURL, "monthly=") {
		return baseURL
	}

	yearMonth := date[:7] // Extract YYYY-MM from YYYY-MM-DD
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%smonthly=%s", baseURL, separator, yearMonth)
}

// createRequest creates an HTTP request with all required headers
func (c *HRAPIClient) createRequest(url string, config *domain.WorkConfig) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add required headers to match the working curl command
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-CN,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Lang-Code", "en")
	req.Header.Set("P-Auth", config.PAuth)
	req.Header.Set("P-Rtoken", config.PRToken)
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="140", "Not=A?Brand";v="24", "Google Chrome";v="140"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	return req, nil
}

// extractCheckTime extracts and parses the check-in and check-out time from attendance records
func (c *HRAPIClient) extractCheckTime(records []AttendanceRecord, date string) (checkInTime *time.Time, checkOutTime *time.Time, err error) {
	for _, record := range records {
		if record.AttendanceDate == date {
			if record.FirstClockInTime != nil && *record.FirstClockInTime != "" {
				checkInStr := fmt.Sprintf("%s %s:00", date, *record.FirstClockInTime)
				slog.Info("[HR API] Parsing check-in time", "time_str", checkInStr)

				ct, err := time.Parse("2006-01-02 15:04:05", checkInStr)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse time %s: %w", checkInStr, err)
				}

				// Adjust timezone (subtract 8 hours to convert to local time)
				ct = ct.In(time.Local).Add(-time.Hour * 8)
				checkInTime = &ct
				slog.Info("[HR API] Successfully parsed check-in time", "time", checkInTime)
			}
			if record.LastClockOutTime != nil && *record.LastClockOutTime != "" {
				checkOutStr := fmt.Sprintf("%s %s:00", date, *record.LastClockOutTime)
				slog.Info("[HR API] Parsing check-out time", "time_str", checkOutStr)

				ct, err := time.Parse("2006-01-02 15:04:05", checkOutStr)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse time %s: %w", checkOutStr, err)
				}

				// Adjust timezone (subtract 8 hours to convert to local time)
				ct = ct.In(time.Local).Add(-time.Hour * 8)
				checkOutTime = &ct
				slog.Info("[HR API] Successfully parsed check-out time", "time", checkOutTime)
			}
			return
		}
	}

	return nil, nil, fmt.Errorf("no attendance record found for date %s", date)
}
