package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

// FetchCheckInTime fetches the check-in time for a specific date from HR API
func (c *HRAPIClient) FetchCheckInTime(config *domain.WorkConfig, date string) (*time.Time, error) {
	if !config.ShouldAutoFetch() {
		return nil, fmt.Errorf("HR API not properly configured or auto-fetch disabled")
	}

	log.Printf("[HR API] Fetching check-in time for date: %s", date)

	// Build API URL with monthly parameter
	apiURL := c.buildAPIURL(config.CheckInAPIURL, date)
	log.Printf("[HR API] Request URL: %s", apiURL)

	// Create and configure request
	req, err := c.createRequest(apiURL, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[HR API] Request failed: %v", err)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[HR API] Response status: %d", resp.StatusCode)

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("[HR API] Raw response: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var hrResponse HRAttendanceInfo
	if err := json.Unmarshal(bodyBytes, &hrResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[HR API] Response - Code: %s, Success: %v, Message: %s, Records: %d",
		hrResponse.Code, hrResponse.Success, hrResponse.Message, len(hrResponse.Data))

	if hrResponse.Code != "200" || !hrResponse.Success {
		return nil, fmt.Errorf("API error (code %s): %s", hrResponse.Code, hrResponse.Message)
	}

	// Find check-in record for the specified date
	return c.extractCheckInTime(hrResponse.Data, date)
}

// FetchAttendanceStatus fetches attendance records for a specific date
func (c *HRAPIClient) FetchAttendanceStatus(config *domain.WorkConfig, date string) (hasCheckedIn, hasCheckedOut bool, err error) {
	if !config.HasAPIConfig() {
		return false, false, fmt.Errorf("HR API not properly configured")
	}

	apiURL := c.buildAPIURL(config.CheckInAPIURL, date)
	req, err := c.createRequest(apiURL, config)
	if err != nil {
		return false, false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, false, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, false, fmt.Errorf("failed to read response: %w", err)
	}

	var hrResponse HRAttendanceInfo
	if err := json.Unmarshal(bodyBytes, &hrResponse); err != nil {
		return false, false, fmt.Errorf("failed to decode response: %w", err)
	}

	if hrResponse.Code != "200" || !hrResponse.Success {
		return false, false, fmt.Errorf("API error: %s", hrResponse.Message)
	}

	// Check attendance status for the date
	for _, record := range hrResponse.Data {
		if record.AttendanceDate == date {
			hasCheckedIn = record.FirstClockInTime != nil && *record.FirstClockInTime != ""
			hasCheckedOut = record.LastClockOutTime != nil && *record.LastClockOutTime != ""
			log.Printf("[HR API] Attendance status for %s - CheckedIn: %v, CheckedOut: %v",
				date, hasCheckedIn, hasCheckedOut)
			return hasCheckedIn, hasCheckedOut, nil
		}
	}

	log.Printf("[HR API] No attendance record found for date: %s", date)
	return false, false, nil
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

// extractCheckInTime extracts and parses the check-in time from attendance records
func (c *HRAPIClient) extractCheckInTime(records []AttendanceRecord, date string) (*time.Time, error) {
	for _, record := range records {
		if record.AttendanceDate == date && record.FirstClockInTime != nil && *record.FirstClockInTime != "" {
			checkInStr := fmt.Sprintf("%s %s:00", date, *record.FirstClockInTime)
			log.Printf("[HR API] Parsing check-in time: %s", checkInStr)

			checkInTime, err := time.Parse("2006-01-02 15:04:05", checkInStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse time %s: %w", checkInStr, err)
			}

			// Adjust timezone (subtract 8 hours to convert to local time)
			checkInTime = checkInTime.In(time.Local).Add(-time.Hour * 8)
			log.Printf("[HR API] Successfully parsed check-in time: %v", checkInTime)
			return &checkInTime, nil
		}
	}

	return nil, fmt.Errorf("no check-in found for date %s", date)
}
