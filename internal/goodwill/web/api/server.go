package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/core/scheduling"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/internal/goodwill/notifications"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// Server represents the web server
type Server struct {
	config          *config.WebConfig
	repo            db.Repository
	scheduler       *scheduling.Scheduler
	notificationSvc *notifications.NotificationIntegration
	httpServer      *http.Server
	router          *http.ServeMux
	log             *log.Logger
}

// Router returns the HTTP router
func (s *Server) Router() *http.ServeMux {
	return s.router
}

// NewServer creates a new web server instance
func NewServer(cfg *config.WebConfig, repo db.Repository, scheduler *scheduling.Scheduler, notificationSvc *notifications.NotificationIntegration, logger *log.Logger) *Server {
	return &Server{
		config:          cfg,
		repo:            repo,
		scheduler:       scheduler,
		notificationSvc: notificationSvc,
		log:             logger,
		router:          http.NewServeMux(),
	}
}

// Start starts the web server (blocking)
func (s *Server) Start() error {
	// Setup routes
	s.setupRoutes()

	// Setup CSRF Protection
	var csrfKey []byte
	if s.config.CSRFAuthKey != "" {
		csrfKey = []byte(s.config.CSRFAuthKey)
		// Ensure it's 32 bytes
		if len(csrfKey) > 32 {
			csrfKey = csrfKey[:32]
		}
	} else {
		// Generate random 32-byte key
		csrfKey = make([]byte, 32)
		if _, err := rand.Read(csrfKey); err != nil {
			return fmt.Errorf("failed to generate secure CSRF key: %w", err)
		}
		s.log.Warn("No CSRFAuthKey configured. Generated random key (sessions will invalidate on restart).")
	}

	csrfMiddleware := csrf.Protect(
		csrfKey,
		csrf.Secure(!s.config.TLS.Enabled && os.Getenv("ENV") != "development"), // True in prod, false otherwise
		csrf.Path("/"),
		// The API uses JSON, and we also have forms. By default gorilla/csrf checks X-CSRF-Token or csrf_token.
	)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      csrfMiddleware(s.router),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.log.Infof("Starting web server on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.log.Errorf("Server error: %v", err)
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the web server
func (s *Server) Shutdown() error {
	s.log.Info("Shutting down server...")

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Errorf("Server shutdown error: %v", err)
		return err
	}

	s.log.Info("Server shutdown complete")
	return nil
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API v1 routes
	apiV1 := http.NewServeMux()

	// Search endpoints
	apiV1.HandleFunc("GET /searches", s.handleGetSearches)
	apiV1.HandleFunc("POST /searches", s.handleCreateSearch)
	apiV1.HandleFunc("GET /searches/{id}", s.handleGetSearch)
	apiV1.HandleFunc("PUT /searches/{id}", s.handleUpdateSearch)
	apiV1.HandleFunc("DELETE /searches/{id}", s.handleDeleteSearch)
	apiV1.HandleFunc("POST /searches/{id}/execute", s.handleExecuteSearch)

	// Item endpoints
	apiV1.HandleFunc("GET /items", s.handleGetItems)
	apiV1.HandleFunc("GET /items/{id}", s.handleGetItem)
	apiV1.HandleFunc("GET /items/{id}/history", s.handleGetItemHistory)

	// Notification endpoints
	apiV1.HandleFunc("GET /notifications", s.handleGetNotifications)
	apiV1.HandleFunc("POST /notifications/test", s.handleTestNotification)

	// System endpoints
	apiV1.HandleFunc("GET /system/status", s.handleHealthCheck)

	// Mount API v1 under /api/v1
	s.router.Handle("/api/v1/", http.StripPrefix("/api/v1", apiV1))

	// Health check endpoint
	s.router.HandleFunc("GET /health", s.handleHealthCheck)
}

// APIError represents an API error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// handleError handles API errors consistently
func (s *Server) handleError(w http.ResponseWriter, err error, statusCode int) {
	s.log.Errorf("API Error: %v", err)

	response := APIError{
		Code:    statusCode,
		Message: http.StatusText(statusCode),
	}

	if os.Getenv("DEBUG") == "true" {
		response.Details = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.log.Errorf("Failed to encode error response: %v", err)
	}
}

// handleHealthCheck handles health check requests
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	dbStatus := "healthy"
	dbError := ""

	// Test database connection by getting stats (single query)
	notifStats, err := s.repo.GetNotificationStats(r.Context())
	if err != nil {
		dbStatus = "unhealthy"
		dbError = err.Error()
		s.log.Errorf("Database health check failed: %v", err)
	}

	// Get system statistics (from the optimized query above mostly, or call again? call once is better)
	// If the first call failed, we probably don't have stats.
	var totalNotifications, pendingNotifications, processingNotifications, deliveredNotifications, failedNotifications int
	if notifStats != nil {
		totalNotifications = notifStats.Total
		pendingNotifications = notifStats.Pending
		processingNotifications = notifStats.Processing
		deliveredNotifications = notifStats.Delivered
		failedNotifications = notifStats.Failed
	}

	// Determine overall system status
	systemStatus := "healthy"
	if dbStatus == "unhealthy" {
		systemStatus = "degraded"
	}

	// Prepare health check response
	healthResponse := map[string]interface{}{
		"status":    systemStatus,
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"database": map[string]interface{}{
			"status": dbStatus,
			"error":  dbError,
		},
		"notifications": map[string]interface{}{
			"total":      totalNotifications,
			"pending":    pendingNotifications,
			"processing": processingNotifications,
			"delivered":  deliveredNotifications,
			"failed":     failedNotifications,
		},
		"components": map[string]string{
			"api":           "healthy",
			"scheduler":     "healthy",
			"notifications": "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		s.log.Errorf("Failed to encode health check response: %v", err)
	}
}
