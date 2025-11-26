// Package http provides HTTP handlers and routing for the work tracking API.
package http

import (
	"net/http"
)

// SetupRouter configures and returns the HTTP router with all routes
func SetupRouter(workHandler *WorkHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Register API routes with CORS middleware
	mux.HandleFunc("/api/checkin", corsMiddleware(workHandler.CheckIn))
	mux.HandleFunc("/api/checkout", corsMiddleware(workHandler.CheckOut))
	mux.HandleFunc("/api/status", corsMiddleware(workHandler.GetStatus))
	mux.HandleFunc("/api/today-checkin", corsMiddleware(workHandler.GetTodayCheckIn))
	mux.HandleFunc("/api/monthly-stats", corsMiddleware(workHandler.GetMonthlyStats))
	mux.HandleFunc("/api/config", corsMiddleware(handleConfig(workHandler)))

	// Serve Next.js static files
	fs := http.FileServer(http.Dir("../../frontend/out"))
	mux.Handle("/", fs)

	return mux
}

// handleConfig handles both GET and POST for /api/config
func handleConfig(workHandler *WorkHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			workHandler.GetConfig(w, r)
		case http.MethodPost:
			workHandler.UpdateConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// corsMiddleware adds CORS headers to allow cross-origin requests
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS for development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
