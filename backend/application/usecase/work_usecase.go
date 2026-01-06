// Package usecase contains the application business logic and use cases.
package usecase

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/simon0-o/offline_me/backend/domain"
	"github.com/simon0-o/offline_me/backend/infrastructure/client"
	"github.com/simon0-o/offline_me/backend/interfaces/dto"
)

// WorkUsecase handles work tracking business logic
type WorkUsecase struct {
	repo               domain.Repository
	attendanceProvider domain.AttendanceProvider
}

// NewWorkUsecase creates a new work usecase instance
func NewWorkUsecase(repo domain.Repository) *WorkUsecase {
	return &WorkUsecase{
		repo:               repo,
		attendanceProvider: client.NewHRAPIClient(),
	}
}

// CheckIn processes a check-in request
func (uc *WorkUsecase) CheckIn(req *dto.CheckInRequest) (*dto.CheckInResponse, error) {
	today := req.CheckInTime.Format("2006-01-02")
	config, err := uc.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Check if already checked in today
	existingSession := uc.repo.GetTodaySession(today)

	var session *domain.WorkSession
	if existingSession != nil {
		// Re-check-in: update existing session
		existingSession.CheckIn = req.CheckInTime
		existingSession.WorkHours = config.DefaultWorkHours
		existingSession.CheckOut = nil // Reset checkout time
		session = existingSession
		slog.Info("[CheckIn] Re-checking in", "date", today, "time", req.CheckInTime)
	} else {
		// New check-in: create new session
		session = &domain.WorkSession{
			ID:        uuid.New().String(),
			Date:      today,
			CheckIn:   req.CheckInTime,
			WorkHours: config.DefaultWorkHours,
		}
		slog.Info("[CheckIn] New check-in", "date", today, "time", req.CheckInTime)
	}

	if err := uc.repo.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return &dto.CheckInResponse{
		SessionID:    session.ID,
		CheckInTime:  session.CheckIn,
		CheckOutTime: session.CalculateExpectedCheckOut(),
		WorkHours:    session.WorkHours,
	}, nil
}

// CheckOut processes a check-out request
func (uc *WorkUsecase) CheckOut(req *dto.CheckOutRequest) (*dto.CheckOutResponse, error) {
	today := req.CheckOutTime.Format("2006-01-02")

	// Get today's session
	session := uc.repo.GetTodaySession(today)
	if session == nil {
		return nil, fmt.Errorf("no check-in found for %s", today)
	}

	// Update checkout time
	session.CheckOut = &req.CheckOutTime
	if err := uc.repo.SaveSession(session); err != nil {
		return nil, fmt.Errorf("failed to save check-out: %w", err)
	}

	slog.Info("[CheckOut] Checked out", "time", req.CheckOutTime, "overtime_minutes", session.CalculateOvertime())

	return &dto.CheckOutResponse{
		SessionID:       session.ID,
		CheckInTime:     session.CheckIn,
		CheckOutTime:    req.CheckOutTime,
		OvertimeMinutes: session.CalculateOvertime(),
	}, nil
}

// GetStatus retrieves the current work status
func (uc *WorkUsecase) GetStatus() (*dto.StatusResponse, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	session := uc.repo.GetTodaySession(today)
	config, err := uc.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if session == nil {
		return &dto.StatusResponse{
			HasCheckedIn:    false,
			CurrentTime:     now,
			WorkHours:       config.DefaultWorkHours,
			IsCheckOutTime:  false,
			OvertimeMinutes: 0,
		}, nil
	}

	expectedCheckOut := session.CalculateExpectedCheckOut()
	isCheckOutTime := now.After(expectedCheckOut)

	return &dto.StatusResponse{
		HasCheckedIn:     true,
		CheckInTime:      &session.CheckIn,
		CheckOutTime:     session.CheckOut,
		ExpectedCheckOut: &expectedCheckOut,
		CurrentTime:      now,
		WorkHours:        session.WorkHours,
		IsCheckOutTime:   isCheckOutTime,
		OvertimeMinutes:  session.CalculateOvertime(),
	}, nil
}

// GetTodayCheckIn retrieves or auto-fetches today's check-in information
func (uc *WorkUsecase) GetTodayCheckIn(req *dto.TodayCheckInRequest) (*dto.TodayCheckInResponse, error) {
	// Check if already checked in
	session := uc.repo.GetTodaySession(req.Date)
	if session != nil && !req.ReCheckIn {
		return &dto.TodayCheckInResponse{
			HasCheckedIn: true,
			CheckInTime:  &session.CheckIn,
		}, nil
	}

	config, err := uc.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Try to auto-fetch from HR API if enabled
	if config.ShouldAutoFetch() {
		return uc.autoFetchCheckIn(req.Date, session, config)
	}

	return &dto.TodayCheckInResponse{
		HasCheckedIn:     false,
		CheckInTime:      nil,
		CanAutoFetch:     config.HasAPIConfig(),
		AutoFetchEnabled: config.AutoFetchEnabled,
	}, nil
}

// autoFetchCheckIn fetches check-in time from HR API and creates a session
func (uc *WorkUsecase) autoFetchCheckIn(date string, existingSession *domain.WorkSession, config *domain.WorkConfig) (*dto.TodayCheckInResponse, error) {
	checkInTime, _, err := uc.attendanceProvider.FetchAttendanceStatus(config, date)
	if err != nil {
		slog.Info("[AutoFetch] Failed to fetch check-in time", "error", err)
		return &dto.TodayCheckInResponse{
			HasCheckedIn:     false,
			CheckInTime:      nil,
			CanAutoFetch:     true,
			AutoFetchEnabled: true,
			APIError:         err.Error(),
		}, nil
	}

	slog.Info("[AutoFetch] Successfully fetched check-in time", "time", *checkInTime)

	// Determine session ID
	sessionID := uuid.New().String()
	if existingSession != nil {
		sessionID = existingSession.ID
	}

	// Create/update session with fetched time
	session := &domain.WorkSession{
		ID:        sessionID,
		Date:      date,
		CheckIn:   *checkInTime,
		WorkHours: config.DefaultWorkHours,
	}

	if err := uc.repo.SaveSession(session); err != nil {
		slog.Info("[AutoFetch] Failed to save session", "error", err)
		return &dto.TodayCheckInResponse{
			HasCheckedIn:     false,
			CheckInTime:      checkInTime,
			CanAutoFetch:     true,
			AutoFetchEnabled: true,
			APIError:         fmt.Sprintf("Failed to save session: %v", err),
		}, nil
	}

	return &dto.TodayCheckInResponse{
		HasCheckedIn:     true,
		CheckInTime:      checkInTime,
		CanAutoFetch:     true,
		AutoFetchEnabled: true,
	}, nil
}

// UpdateConfig updates the work configuration
func (uc *WorkUsecase) UpdateConfig(req *dto.ConfigRequest) error {
	config, err := uc.repo.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Update configuration fields
	if req.WorkHours > 0 {
		if req.WorkHours > domain.MaxWorkMinutesPerDay {
			return fmt.Errorf("work hours cannot exceed %d minutes (24 hours)", domain.MaxWorkMinutesPerDay)
		}
		config.DefaultWorkHours = req.WorkHours
	}

	config.CheckInAPIURL = req.CheckInAPIURL
	config.AutoFetchEnabled = req.AutoFetchEnabled
	config.PAuth = req.PAuth
	config.PRToken = req.PRToken
	config.CheckInWebhookURL = req.CheckInWebhookURL
	config.CheckOutWebhookURL = req.CheckOutWebhookURL

	// Update existing session's work hours if checked in today
	if req.WorkHours > 0 {
		today := time.Now().Format("2006-01-02")
		if session := uc.repo.GetTodaySession(today); session != nil {
			session.WorkHours = req.WorkHours
			if err := uc.repo.SaveSession(session); err != nil {
				return fmt.Errorf("failed to update session work hours: %w", err)
			}
			slog.Info("[UpdateConfig] Updated today's session work hours", "minutes", req.WorkHours)
		}
	}

	if err := uc.repo.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	slog.Info("[UpdateConfig] Configuration updated successfully")
	return nil
}

// GetConfig retrieves the current work configuration
func (uc *WorkUsecase) GetConfig() (*dto.ConfigResponse, error) {
	config, err := uc.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return &dto.ConfigResponse{
		WorkHours:          config.DefaultWorkHours,
		CheckInAPIURL:      config.CheckInAPIURL,
		AutoFetchEnabled:   config.AutoFetchEnabled,
		PAuth:              config.PAuth,
		PRToken:            config.PRToken,
		CheckInWebhookURL:  config.CheckInWebhookURL,
		CheckOutWebhookURL: config.CheckOutWebhookURL,
	}, nil
}

// GetMonthlyStats retrieves monthly overtime statistics
func (uc *WorkUsecase) GetMonthlyStats() (*dto.MonthlyStatsResponse, error) {
	now := time.Now()
	currentMonth := now.Format("2006-01")
	lastMonth := now.AddDate(0, -1, 0).Format("2006-01")

	// Get current month stats
	currentSessions, err := uc.repo.GetSessionsByMonth(currentMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get current month sessions: %w", err)
	}

	// Get last month stats
	lastSessions, err := uc.repo.GetSessionsByMonth(lastMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get last month sessions: %w", err)
	}

	// Calculate statistics
	currentStats := domain.CalculateStats(currentSessions, currentMonth)
	lastStats := domain.CalculateStats(lastSessions, lastMonth)

	return &dto.MonthlyStatsResponse{
		CurrentMonth: dto.MonthStats{
			YearMonth:       currentStats.YearMonth,
			TotalDays:       currentStats.TotalDays,
			CheckedOutDays:  currentStats.CheckedOutDays,
			OvertimeMinutes: currentStats.OvertimeMinutes,
		},
		LastMonth: dto.MonthStats{
			YearMonth:       lastStats.YearMonth,
			TotalDays:       lastStats.TotalDays,
			CheckedOutDays:  lastStats.CheckedOutDays,
			OvertimeMinutes: lastStats.OvertimeMinutes,
		},
	}, nil
}
