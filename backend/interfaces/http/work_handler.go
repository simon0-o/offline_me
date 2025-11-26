package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/simon0-o/offline_me/backend/domain"
	"github.com/simon0-o/offline_me/backend/interfaces/dto"
)

// WorkUsecase defines the interface for work-related business logic
type WorkUsecase interface {
	CheckIn(req *dto.CheckInRequest) (*dto.CheckInResponse, error)
	CheckOut(req *dto.CheckOutRequest) (*dto.CheckOutResponse, error)
	GetStatus() (*dto.StatusResponse, error)
	GetTodayCheckIn(req *dto.TodayCheckInRequest) (*dto.TodayCheckInResponse, error)
	UpdateConfig(req *dto.ConfigRequest) error
	GetConfig() (*dto.ConfigResponse, error)
	GetMonthlyStats() (*dto.MonthlyStatsResponse, error)
}

// WorkHandler handles HTTP requests for work tracking
type WorkHandler struct {
	uc  WorkUsecase
	log *log.Helper
}

// NewWorkHandler creates a new work handler instance
func NewWorkHandler(uc WorkUsecase, logger log.Logger) *WorkHandler {
	return &WorkHandler{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CheckIn handles check-in requests
func (h *WorkHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorf("Invalid check-in request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.uc.CheckIn(&req)
	if err != nil {
		h.log.Errorf("Check-in failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, resp)
}

// CheckOut handles check-out requests
func (h *WorkHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorf("Invalid check-out request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.uc.CheckOut(&req)
	if err != nil {
		h.log.Errorf("Check-out failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.respondJSON(w, resp)
}

// GetStatus handles status requests
func (h *WorkHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := h.uc.GetStatus()
	if err != nil {
		h.log.Errorf("Failed to get status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, resp)
}

// GetTodayCheckIn handles today's check-in retrieval/auto-fetch requests
func (h *WorkHandler) GetTodayCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.TodayCheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorf("Invalid today-checkin request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.uc.GetTodayCheckIn(&req)
	if err != nil {
		h.log.Errorf("Failed to get today check-in: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, resp)
}

// UpdateConfig handles configuration update requests
func (h *WorkHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Errorf("Invalid config request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate work hours
	if req.WorkHours > 0 && req.WorkHours > domain.MaxWorkMinutesPerDay {
		http.Error(w, "Work hours cannot exceed 24 hours (1440 minutes)", http.StatusBadRequest)
		return
	}

	if err := h.uc.UpdateConfig(&req); err != nil {
		h.log.Errorf("Failed to update config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]string{"status": "success"})
}

// GetConfig handles configuration retrieval requests
func (h *WorkHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := h.uc.GetConfig()
	if err != nil {
		h.log.Errorf("Failed to get config: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, config)
}

// GetMonthlyStats handles monthly statistics requests
func (h *WorkHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.uc.GetMonthlyStats()
	if err != nil {
		h.log.Errorf("Failed to get monthly stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, stats)
}

// respondJSON writes a JSON response
func (h *WorkHandler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Errorf("Failed to encode JSON response: %v", err)
	}
}
