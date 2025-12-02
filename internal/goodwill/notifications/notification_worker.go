package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
)

// NotificationWorker handles background synchronization between database and in-memory queue
type NotificationWorker struct {
	repo              db.Repository
	notificationQueue chan db.GormNotification
	shutdownChan      chan struct{}
	workerWg          sync.WaitGroup
	pollInterval      time.Duration
	processingTimeout time.Duration
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(repo db.Repository, notificationQueue chan db.GormNotification) *NotificationWorker {
	return &NotificationWorker{
		repo:              repo,
		notificationQueue: notificationQueue,
		shutdownChan:      make(chan struct{}),
		pollInterval:      5 * time.Second,  // Default: poll every 5 seconds
		processingTimeout: 30 * time.Second, // Default: 30 second timeout for processing
	}
}

// SetPollInterval sets the database polling interval
func (w *NotificationWorker) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// SetPollIntervalForTesting sets a very short poll interval for testing
func (w *NotificationWorker) SetPollIntervalForTesting() {
	w.pollInterval = 100 * time.Millisecond // Very short interval for testing
}

// SetProcessingTimeout sets the processing timeout
func (w *NotificationWorker) SetProcessingTimeout(timeout time.Duration) {
	w.processingTimeout = timeout
}

// Start starts the notification worker
func (w *NotificationWorker) Start() {
	log.Info("Starting notification worker")

	w.workerWg.Add(1)
	go w.databaseToMemorySyncWorker()
}

// Stop stops the notification worker
func (w *NotificationWorker) Stop() {
	log.Info("Stopping notification worker")

	// Signal shutdown
	close(w.shutdownChan)

	// Wait for worker to finish
	w.workerWg.Wait()
	log.Info("Notification worker stopped")
}

// databaseToMemorySyncWorker continuously synchronizes database notifications to in-memory queue
func (w *NotificationWorker) databaseToMemorySyncWorker() {
	defer w.workerWg.Done()

	log.Info("Database-to-memory sync worker started")

	for {
		select {
		case <-w.shutdownChan:
			log.Info("Database-to-memory sync worker shutting down")
			return
		case <-time.After(w.pollInterval):
			w.syncPendingNotifications()
		}
	}
}

// syncPendingNotifications fetches pending notifications from database and queues them in memory
func (w *NotificationWorker) syncPendingNotifications() {
	ctx, cancel := context.WithTimeout(context.Background(), w.processingTimeout)
	defer cancel()

	log.Debug("Syncing pending notifications from database")

	// Get pending notifications from database
	pendingNotifications, err := w.repo.GetPendingNotifications(ctx)
	if err != nil {
		log.Errorf("Failed to get pending notifications: %v", err)
		return
	}

	if len(pendingNotifications) == 0 {
		log.Debug("No pending notifications found")
		return
	}

	log.Infof("Found %d pending notifications to process", len(pendingNotifications))

	// Queue each pending notification to in-memory channel
	for _, notification := range pendingNotifications {
		// Mark as processing in database FIRST to avoid race condition with processor
		// (Processor might finish and mark as delivered before we mark as processing if we did it after)
		err := w.repo.UpdateNotificationStatus(ctx, notification.ID, "processing")
		if err != nil {
			log.Errorf("Failed to update notification %d status to processing: %v", notification.ID, err)
			continue // Skip if we can't update status
		}

		select {
		case w.notificationQueue <- notification:
			log.Debugf("Queued notification %d for processing", notification.ID)
		case <-w.shutdownChan:
			log.Warn("Notification worker shutting down, cannot queue notification")
			// Note: Item is stuck in 'processing' state, but that's better than overwriting 'delivered'
			return
		}
	}
}

// QueueNotificationForProcessing queues a notification for processing
func (w *NotificationWorker) QueueNotificationForProcessing(ctx context.Context, notification db.GormNotification) error {
	// Queue notification in database first
	notificationID, err := w.repo.QueueNotification(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to queue notification in database: %w", err)
	}

	// Update the notification with the new ID
	notification.ID = notificationID
	log.Infof("Queued notification %d in database", notificationID)

	// The database queue is sufficient - no need for immediate memory queuing
	// This removes special casing that was only useful for tests

	return nil
}

// GetNotificationStats returns statistics about notification processing
func (w *NotificationWorker) GetNotificationStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get counts directly from database using single efficient query
	repoStats, err := w.repo.GetNotificationStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	stats["total"] = repoStats.Total
	stats["pending"] = repoStats.Pending
	stats["processing"] = repoStats.Processing
	stats["delivered"] = repoStats.Delivered
	stats["failed"] = repoStats.Failed

	// Calculate success rate
	if repoStats.Total > 0 {
		stats["success_rate"] = float64(repoStats.Delivered) / float64(repoStats.Total)
	} else {
		stats["success_rate"] = 0.0
	}

	return stats, nil
}
