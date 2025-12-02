package notifications

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// NotificationIntegration provides a unified notification system with direct repository access
type NotificationIntegration struct {
	notificationManager *NotificationManager
	config              *config.Config
	repo                db.Repository
	notificationQueue   chan db.GormNotification
	worker              *NotificationWorker
	shutdownChan        chan struct{}
}

// NewNotificationIntegration creates a new notification integration with direct repository access
func NewNotificationIntegration(cfg *config.Config, repo db.Repository) (*NotificationIntegration, error) {
	// Create the notification manager
	notificationManager, err := NewNotificationManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification manager: %w", err)
	}

	integration := &NotificationIntegration{
		notificationManager: notificationManager,
		config:              cfg,
		repo:                repo,
		notificationQueue:   make(chan db.GormNotification, 100),
		shutdownChan:        make(chan struct{}),
	}

	// Create the notification worker
	integration.worker = NewNotificationWorker(repo, integration.notificationQueue)

	return integration, nil
}

// Start starts the notification integration service
func (ni *NotificationIntegration) Start() {
	log.Info("Starting notification integration service")

	// Start the notification worker (database-to-memory sync)
	ni.worker.Start()

	// Start the notification processor
	go ni.notificationProcessor()
}

// Stop stops the notification integration service
func (ni *NotificationIntegration) Stop() {
	log.Info("Stopping notification integration service")

	// Stop the notification worker
	ni.worker.Stop()

	// Signal shutdown to processor
	close(ni.shutdownChan)
}

// notificationProcessor processes notifications from the in-memory queue
func (ni *NotificationIntegration) notificationProcessor() {
	log.Info("Notification processor started")

	for {
		select {
		case notification := <-ni.notificationQueue:
			ni.processNotification(context.Background(), notification)
		case <-ni.shutdownChan:
			log.Info("Notification processor shutting down")
			return
		}
	}
}

// processNotification processes a notification using the notification manager
func (ni *NotificationIntegration) processNotification(ctx context.Context, notification db.GormNotification) {
	log.Infof("Processing notification for item %d", notification.ItemID)

	// Get item details with proper context
	item, err := ni.repo.GetItemByID(ctx, notification.ItemID)
	if err != nil {
		log.Errorf("Failed to get item for notification: %v", err)
		ni.markNotificationFailed(ctx, notification.ID, fmt.Sprintf("Failed to get item: %v", err))
		return
	}

	if item == nil {
		log.Errorf("Item not found for notification %d", notification.ID)
		ni.markNotificationFailed(ctx, notification.ID, "Item not found")
		return
	}

	// Get search details with proper context
	search, err := ni.repo.GetSearchByID(ctx, notification.SearchID)
	if err != nil {
		log.Errorf("Failed to get search for notification: %v", err)
		ni.markNotificationFailed(ctx, notification.ID, fmt.Sprintf("Failed to get search: %v", err))
		return
	}

	if search == nil {
		log.Errorf("Search not found for notification %d", notification.ID)
		ni.markNotificationFailed(ctx, notification.ID, "Search not found")
		return
	}

	// Send notification using the notification manager
	err = ni.notificationManager.NotifyFound(ctx, item, search)
	if err != nil {
		log.Errorf("Failed to send notification: %v", err)
		ni.markNotificationFailed(ctx, notification.ID, fmt.Sprintf("Failed to send notification: %v", err))
		return
	}

	// Mark as delivered
	err = ni.repo.UpdateNotificationStatus(ctx, notification.ID, "delivered")
	if err != nil {
		log.Errorf("Failed to update notification status to delivered: %v", err)
		return
	}

	log.Infof("Successfully processed and delivered notification %d", notification.ID)
}

// markNotificationFailed marks a notification as failed with error message
func (ni *NotificationIntegration) markNotificationFailed(ctx context.Context, notificationID int, errorMessage string) {
	// Update status to failed
	err := ni.repo.UpdateNotificationStatus(ctx, notificationID, "failed")
	if err != nil {
		log.Errorf("Failed to update notification %d status to failed: %v", notificationID, err)
		return
	}

	// Update the notification with error message
	notification, err := ni.repo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		log.Errorf("Failed to get notification %d for error update: %v", notificationID, err)
		return
	}

	if notification != nil {
		notification.ErrorMessage = errorMessage
		notification.UpdatedAt = time.Now()
		err = ni.repo.UpdateNotification(ctx, *notification)
		if err != nil {
			log.Errorf("Failed to update notification %d with error message: %v", notificationID, err)
		}
	}
}

// QueueNotification queues a notification for processing using the new system
func (ni *NotificationIntegration) QueueNotification(ctx context.Context, notification db.GormNotification) error {
	return ni.worker.QueueNotificationForProcessing(ctx, notification)
}

// QueueNotificationForNewSystem queues a notification for processing using the new system
func (ni *NotificationIntegration) QueueNotificationForNewSystem(ctx context.Context, item *db.GormItem, search *db.GormSearch) error {
	notification := db.GormNotification{
		ItemID:           item.ID,
		SearchID:         search.ID,
		NotificationType: "new_system",
		Status:           "queued",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return ni.QueueNotification(ctx, notification)
}

// GetNotificationStats returns statistics about notification processing
func (ni *NotificationIntegration) GetNotificationStats(ctx context.Context) (map[string]interface{}, error) {
	return ni.worker.GetNotificationStats(ctx)
}

// GetPendingNotifications returns pending notifications
func (ni *NotificationIntegration) GetPendingNotifications(ctx context.Context) ([]db.GormNotification, error) {
	return ni.repo.GetPendingNotifications(ctx)
}

// GetAllNotifications returns all notifications
func (ni *NotificationIntegration) GetAllNotifications(ctx context.Context) ([]db.GormNotification, error) {
	return ni.repo.GetAllNotifications(ctx)
}

// GetNotificationByID returns a notification by ID
func (ni *NotificationIntegration) GetNotificationByID(ctx context.Context, id int) (*db.GormNotification, error) {
	return ni.repo.GetNotificationByID(ctx, id)
}

// UpdateNotificationStatus updates notification status
func (ni *NotificationIntegration) UpdateNotificationStatus(ctx context.Context, id int, status string) error {
	return ni.repo.UpdateNotificationStatus(ctx, id, status)
}

// UpdateNotification updates a notification
func (ni *NotificationIntegration) UpdateNotification(ctx context.Context, notification db.GormNotification) error {
	return ni.repo.UpdateNotification(ctx, notification)
}
