package api

import (
	"time"
)

// NotificationRequest represents a test notification request
type NotificationRequest struct {
	Type      string `json:"type"`
	ItemID    int    `json:"item_id"`
	SearchID  int    `json:"search_id"`
	Recipient string `json:"recipient"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID               int        `json:"id"`
	ItemID           int        `json:"item_id"`
	SearchID         int        `json:"search_id"`
	NotificationType string     `json:"notification_type"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	SentAt           *time.Time `json:"sent_at"`
	DeliveredAt      *time.Time `json:"delivered_at"`
	ErrorMessage     string     `json:"error_message"`
	RetryCount       int        `json:"retry_count"`
}

// NotificationListResponse represents a list of notifications with pagination
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int                    `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}
