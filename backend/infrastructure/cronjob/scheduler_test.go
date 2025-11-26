package cronjob

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/simon0-o/offline_me/backend/domain"
	"github.com/simon0-o/offline_me/backend/infrastructure/client"
	"github.com/stretchr/testify/assert"
)

// MockStore is a mock implementation of domain.Repository for testing
type MockStore struct {
	config *domain.WorkConfig
}

func (m *MockStore) GetConfig() (*domain.WorkConfig, error) {
	if m.config == nil {
		return &domain.WorkConfig{
			ID:                 "default",
			DefaultWorkHours:   480,
			CheckInAPIURL:      "https://api.example.com/attendance",
			AutoFetchEnabled:   true,
			PAuth:              "test-p-auth",
			PRToken:            "test-p-rtoken",
			CheckInWebhookURL:  "https://webhook.example.com/checkin",
			CheckOutWebhookURL: "https://webhook.example.com/checkout",
		}, nil
	}
	return m.config, nil
}

func (m *MockStore) GetTodaySession(date string) *domain.WorkSession {
	return nil
}

func (m *MockStore) GetSessionsByMonth(yearMonth string) ([]*domain.WorkSession, error) {
	return nil, nil
}

func (m *MockStore) SaveSession(session *domain.WorkSession) error {
	return nil
}

func (m *MockStore) SaveConfig(config *domain.WorkConfig) error {
	return nil
}

func (m *MockStore) Close() error {
	return nil
}

func TestIsHolidayToday_WorkingDay(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock holiday API response for working day
	httpmock.RegisterResponder("GET", "http://api.haoshenqi.top/holiday/today",
		httpmock.NewJsonResponderOrPanic(200, client.HolidayResponse{
			Code: 200,
			Data: struct {
				Date   string `json:"date"`
				Status string `json:"status"`
			}{
				Date:   "2025-10-13",
				Status: client.HolidayStatusWork,
			},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	isHoliday := scheduler.isHolidayToday()

	assert.False(t, isHoliday)
	assert.Equal(t, 1, httpmock.GetTotalCallCount())
}

func TestIsHolidayToday_Holiday(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock holiday API response for holiday
	httpmock.RegisterResponder("GET", "http://api.haoshenqi.top/holiday/today",
		httpmock.NewJsonResponderOrPanic(200, client.HolidayResponse{
			Code: 200,
			Data: struct {
				Date   string `json:"date"`
				Status string `json:"status"`
			}{
				Date:   "2025-10-01",
				Status: client.HolidayStatusRest,
			},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	isHoliday := scheduler.isHolidayToday()

	assert.True(t, isHoliday)
}

func TestIsHolidayToday_APIError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock API error
	httpmock.RegisterResponder("GET", "http://api.haoshenqi.top/holiday/today",
		httpmock.NewStringResponder(500, "Internal Server Error"))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	isHoliday := scheduler.isHolidayToday()

	// Should return false on error (assume not holiday)
	assert.False(t, isHoliday)
}

func TestHasCheckedIn_Checked(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	checkInTime := "09:30"
	// Mock HR API response with check-in
	httpmock.RegisterResponder("GET", "https://api.example.com/attendance?monthly=2025-10",
		httpmock.NewJsonResponderOrPanic(200, client.HRAttendanceInfo{
			Code:    "200",
			Message: "Success",
			Success: true,
			Data: []client.AttendanceRecord{
				{
					AttendanceDate:   "2025-10-13",
					FirstClockInTime: &checkInTime,
					LastClockOutTime: nil,
				},
			},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)
	config, error := mockStore.GetConfig()
	assert.NoError(t, error)

	hasCheckedIn := scheduler.hasCheckedIn(config, "2025-10-13")

	assert.True(t, hasCheckedIn)
}

func TestHasCheckedIn_NotChecked(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock HR API response without check-in
	httpmock.RegisterResponder("GET", "https://api.example.com/attendance?monthly=2025-10",
		httpmock.NewJsonResponderOrPanic(200, client.HRAttendanceInfo{
			Code:    "200",
			Message: "Success",
			Success: true,
			Data:    []client.AttendanceRecord{},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)
	config, err := mockStore.GetConfig()
	assert.NoError(t, err)

	hasCheckedIn := scheduler.hasCheckedIn(config, "2025-10-13")

	assert.False(t, hasCheckedIn)
}

func TestHasCheckedIn_NoAPIConfig(t *testing.T) {
	mockStore := &MockStore{
		config: &domain.WorkConfig{
			CheckInAPIURL:    "",
			AutoFetchEnabled: false,
		},
	}
	scheduler := NewScheduler(mockStore)
	config, err := mockStore.GetConfig()
	assert.NoError(t, err)

	hasCheckedIn := scheduler.hasCheckedIn(config, "2025-10-13")

	// Should return false when API not configured
	assert.False(t, hasCheckedIn)
}

func TestHasCheckedOut_Checked(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	checkOutTime := "18:30"
	// Mock HR API response with check-out
	httpmock.RegisterResponder("GET", "https://api.example.com/attendance?monthly=2025-10",
		httpmock.NewJsonResponderOrPanic(200, client.HRAttendanceInfo{
			Code:    "200",
			Message: "Success",
			Success: true,
			Data: []client.AttendanceRecord{
				{
					AttendanceDate:   "2025-10-13",
					FirstClockInTime: nil,
					LastClockOutTime: &checkOutTime,
				},
			},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)
	config, err := mockStore.GetConfig()
	assert.NoError(t, err)

	hasCheckedOut := scheduler.hasCheckedOut(config, "2025-10-13")

	assert.True(t, hasCheckedOut)
}

func TestHasCheckedOut_NotChecked(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock HR API response without check-out
	httpmock.RegisterResponder("GET", "https://api.example.com/attendance?monthly=2025-10",
		httpmock.NewJsonResponderOrPanic(200, client.HRAttendanceInfo{
			Code:    "200",
			Message: "Success",
			Success: true,
			Data:    []client.AttendanceRecord{},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)
	config, err := mockStore.GetConfig()
	assert.NoError(t, err)

	hasCheckedOut := scheduler.hasCheckedOut(config, "2025-10-13")

	assert.False(t, hasCheckedOut)
}

func TestCheckInReminder_NoWebhookURL(t *testing.T) {
	mockStore := &MockStore{
		config: &domain.WorkConfig{
			CheckInWebhookURL: "",
		},
	}
	scheduler := NewScheduler(mockStore)

	// Should not panic and should skip execution
	scheduler.checkInReminder()
	// No assertions needed - just verify it doesn't panic
}

func TestCheckInReminder_Holiday(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock holiday API to return holiday
	httpmock.RegisterResponder("GET", "http://api.haoshenqi.top/holiday/today",
		httpmock.NewJsonResponderOrPanic(200, client.HolidayResponse{
			Code: 200,
			Data: struct {
				Date   string `json:"date"`
				Status string `json:"status"`
			}{
				Date:   "2025-10-13",
				Status: client.HolidayStatusRest,
			},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	// Should skip webhook call on holiday
	scheduler.checkInReminder()

	// Verify webhook was not called
	info := httpmock.GetCallCountInfo()
	webhookCalls := 0
	for url := range info {
		if url == "POST https://webhook.example.com/checkin" {
			webhookCalls = info[url]
		}
	}
	assert.Equal(t, 0, webhookCalls)
}

func TestCheckOutReminder_NoWebhookURL(t *testing.T) {
	mockStore := &MockStore{
		config: &domain.WorkConfig{
			CheckOutWebhookURL: "",
		},
	}
	scheduler := NewScheduler(mockStore)

	// Should not panic and should skip execution
	scheduler.checkOutReminder()
	// No assertions needed - just verify it doesn't panic
}

func TestCheckOutReminder_Holiday(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock holiday API to return holiday
	httpmock.RegisterResponder("GET", "http://api.haoshenqi.top/holiday/today",
		httpmock.NewJsonResponderOrPanic(200, client.HolidayResponse{
			Code: 200,
			Data: struct {
				Date   string `json:"date"`
				Status string `json:"status"`
			}{
				Date:   "2025-10-13",
				Status: client.HolidayStatusRest,
			},
		}))

	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	// Should skip webhook call on holiday
	scheduler.checkOutReminder()

	// Verify webhook was not called
	info := httpmock.GetCallCountInfo()
	webhookCalls := 0
	for url := range info {
		if url == "POST https://webhook.example.com/checkout" {
			webhookCalls = info[url]
		}
	}
	assert.Equal(t, 0, webhookCalls)
}

func TestNewScheduler(t *testing.T) {
	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	assert.NotNil(t, scheduler)
	assert.NotNil(t, scheduler.cron)
	assert.Equal(t, mockStore, scheduler.store)
	assert.NotNil(t, scheduler.attendanceProvider)
	assert.NotNil(t, scheduler.holidayClient)
	assert.NotNil(t, scheduler.webhookClient)
}

func TestScheduler_StartStop(t *testing.T) {
	mockStore := &MockStore{}
	scheduler := NewScheduler(mockStore)

	// Start scheduler
	scheduler.Start()

	// Verify cron jobs were added (3 jobs: 1 check-in, 2 check-out)
	entries := scheduler.cron.Entries()
	assert.Equal(t, 3, len(entries))

	// Stop scheduler
	scheduler.Stop()
}
