package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
)

// handleGetNotifications handles GET /api/v1/notifications
func (s *Server) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	status := queryParams.Get("status")
	notificationType := queryParams.Get("type")
	limitStr := queryParams.Get("limit")
	offsetStr := queryParams.Get("offset")

	// Parse pagination
	limit := 20
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			s.handleError(w, fmt.Errorf("invalid limit"), http.StatusBadRequest)
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			s.handleError(w, fmt.Errorf("invalid offset"), http.StatusBadRequest)
			return
		}
	}

	// Prepare filter parameters
	var statusFilter *string
	var typeFilter *string

	if status != "" {
		statusFilter = &status
	}

	if notificationType != "" {
		typeFilter = &notificationType
	}

	// Use database-level filtering and pagination
	paginatedNotifications, totalCount, err := s.repo.GetNotificationsFiltered(
		r.Context(),
		statusFilter,
		typeFilter,
		limit,
		offset,
	)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response NotificationListResponse
	response.Total = totalCount
	response.Limit = limit
	response.Offset = offset

	for _, notification := range paginatedNotifications {
		// Safe conversion from uint to int with overflow checking
		notificationID := notification.ID
		itemID := notification.ItemID
		searchID := notification.SearchID

		response.Notifications = append(response.Notifications, NotificationResponse{
			ID:               notificationID,
			ItemID:           itemID,
			SearchID:         searchID,
			NotificationType: notification.NotificationType,
			Status:           notification.Status,
			CreatedAt:        notification.CreatedAt,
			UpdatedAt:        notification.UpdatedAt,
			SentAt:           notification.SentAt,
			DeliveredAt:      notification.DeliveredAt,
			ErrorMessage:     notification.ErrorMessage,
			RetryCount:       int(notification.RetryCount),
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}

// handleTestNotification handles POST /api/v1/notifications/test
func (s *Server) handleTestNotification(w http.ResponseWriter, r *http.Request) {
	var request NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.Type == "" {
		s.handleError(w, fmt.Errorf("type is required"), http.StatusBadRequest)
		return
	}
	if request.Recipient == "" {
		s.handleError(w, fmt.Errorf("recipient is required"), http.StatusBadRequest)
		return
	}

	// Create test notification with safe conversion from int to uint
	itemID := request.ItemID
	searchID := request.SearchID

	notification := db.GormNotification{
		ItemID:           itemID,
		SearchID:         searchID,
		NotificationType: request.Type,
		Status:           "queued",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Queue notification
	notificationID, err := s.repo.QueueNotification(context.Background(), notification)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Get the queued notification (for verification)
	_, err = s.repo.GetNotificationByID(context.Background(), notificationID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"status":          "queued",
		"message":         "Test notification queued successfully",
		"notification_id": notificationID,
		"recipient":       request.Recipient,
		"type":            request.Type,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}
