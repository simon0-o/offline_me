package main

import (
	"context"
	nethttp "net/http" // Changed from nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/simon0-o/offline_me/backend/application/usecase"
	"github.com/simon0-o/offline_me/backend/infrastructure/cronjob"
	"github.com/simon0-o/offline_me/backend/infrastructure/persistence"
	"github.com/simon0-o/offline_me/backend/interfaces/http"
	// Aliased to avoid conflict with net/http
)

func main() {
	logger := log.NewStdLogger(os.Stdout)
	helper := log.NewHelper(logger)

	// Initialize SQLite database
	store, err := persistence.NewSQLiteStore("../worktime.db")
	if err != nil {
		helper.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	// Initialize dependencies (Clean Architecture layers)
	workUsecase := usecase.NewWorkUsecase(store)
	workHandler := http.NewWorkHandler(workUsecase, logger)

	// Initialize and start cronjob scheduler
	scheduler := cronjob.NewScheduler(store)
	scheduler.Start()
	defer scheduler.Stop()

	// Setup HTTP router and server
	router := http.SetupRouter(workHandler)
	server := &nethttp.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		helper.Infof("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			helper.Fatalf("Server start error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	helper.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		helper.Errorf("Server forced to shutdown: %v", err)
	}

	helper.Info("Server exited")
}
