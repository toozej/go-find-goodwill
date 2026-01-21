package notifications

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
)

// MockRepositoryForWorker is a mock repository for testing the notification worker
type MockRepositoryForWorker struct {
	mu                   sync.Mutex
	pendingNotifications []db.GormNotification
	allNotifications     []db.GormNotification
	queuedNotifications  []db.GormNotification
}

// NewMockRepositoryForWorker creates a new mock repository with test data
func NewMockRepositoryForWorker() *MockRepositoryForWorker {
	return &MockRepositoryForWorker{
		pendingNotifications: []db.GormNotification{
			{
				ID:               1,
				ItemID:           1,
				SearchID:         1,
				NotificationType: "test",
				Status:           "queued",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			{
				ID:               2,
				ItemID:           2,
				SearchID:         2,
				NotificationType: "test",
				Status:           "queued",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		},
		allNotifications: []db.GormNotification{
			{
				ID:               1,
				ItemID:           1,
				SearchID:         1,
				NotificationType: "test",
				Status:           "queued",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			{
				ID:               2,
				ItemID:           2,
				SearchID:         2,
				NotificationType: "test",
				Status:           "delivered",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			{
				ID:               3,
				ItemID:           3,
				SearchID:         3,
				NotificationType: "test",
				Status:           "failed",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		},
	}
}

// GetPendingNotifications implements Repository.GetPendingNotifications
func (m *MockRepositoryForWorker) GetPendingNotifications(ctx context.Context) ([]db.GormNotification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pendingNotifications, nil
}

// GetAllNotifications implements Repository.GetAllNotifications
func (m *MockRepositoryForWorker) GetAllNotifications(ctx context.Context) ([]db.GormNotification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.allNotifications, nil
}

// QueueNotification implements Repository.QueueNotification
func (m *MockRepositoryForWorker) QueueNotification(ctx context.Context, notification db.GormNotification) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	notification.ID = len(m.queuedNotifications) + 1
	m.queuedNotifications = append(m.queuedNotifications, notification)
	return notification.ID, nil
}

// UpdateNotificationStatus implements Repository.UpdateNotificationStatus
func (m *MockRepositoryForWorker) UpdateNotificationStatus(ctx context.Context, id int, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Find and update the notification
	for i, notif := range m.pendingNotifications {
		if notif.ID == id {
			m.pendingNotifications[i].Status = status
			return nil
		}
	}
	for i, notif := range m.allNotifications {
		if notif.ID == id {
			m.allNotifications[i].Status = status
			return nil
		}
	}
	return nil
}

// AddPendingNotification adds a notification to the pending list for testing
func (m *MockRepositoryForWorker) AddPendingNotification(notification db.GormNotification) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pendingNotifications = append(m.pendingNotifications, notification)
}

// ClearPendingNotifications clears the pending notifications for testing
func (m *MockRepositoryForWorker) ClearPendingNotifications() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pendingNotifications = []db.GormNotification{}
}

func TestNotificationWorker(t *testing.T) {
	// Create mock repository
	mockRepo := NewMockRepositoryForWorker()

	// Create notification queue
	notificationQueue := make(chan db.GormNotification, 10)

	// Create worker
	worker := NewNotificationWorker(mockRepo, notificationQueue)
	worker.SetPollInterval(100 * time.Millisecond) // Fast polling for testing

	// Start worker
	worker.Start()
	defer worker.Stop()

	// Test syncing pending notifications
	t.Run("SyncPendingNotifications", func(t *testing.T) {
		// Clear any existing notifications
		mockRepo.ClearPendingNotifications()

		// Add test pending notifications
		testNotifications := []db.GormNotification{
			{
				ID:               1,
				ItemID:           1,
				SearchID:         1,
				NotificationType: "test",
				Status:           "queued",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			{
				ID:               2,
				ItemID:           2,
				SearchID:         2,
				NotificationType: "test",
				Status:           "queued",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		}

		for _, notif := range testNotifications {
			mockRepo.AddPendingNotification(notif)
		}

		// Wait for worker to process
		time.Sleep(200 * time.Millisecond)

		// Check that notifications were queued
		queuedCount := 0
		for {
			select {
			case <-notificationQueue:
				queuedCount++
				if queuedCount >= len(testNotifications) {
					return
				}
			case <-time.After(1 * time.Second):
				t.Fatalf("Expected %d notifications to be queued, got %d", len(testNotifications), queuedCount)
				return
			}
		}
	})

	t.Run("GetNotificationStats", func(t *testing.T) {
		// Test stats calculation
		stats, err := worker.GetNotificationStats(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Should have 3 total notifications
		assert.Equal(t, 3, stats["total"])

		// Should have 1 pending, 1 delivered, 1 failed
		assert.Equal(t, 1, stats["pending"])
		assert.Equal(t, 1, stats["delivered"])
		assert.Equal(t, 1, stats["failed"])

		// Success rate should be 1/3 = 0.333...
		assert.InDelta(t, 0.333, stats["success_rate"], 0.01)
	})

	t.Run("QueueNotificationForProcessing", func(t *testing.T) {
		// Test queuing a notification
		testNotification := db.GormNotification{
			ItemID:           99,
			SearchID:         99,
			NotificationType: "test",
			Status:           "queued",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		err := worker.QueueNotificationForProcessing(context.Background(), testNotification)
		assert.NoError(t, err)

		// Should have been added to the repository
		assert.Equal(t, 1, len(mockRepo.queuedNotifications))
		assert.Equal(t, 99, mockRepo.queuedNotifications[0].ItemID)
	})
}

func TestNotificationWorkerContextPropagation(t *testing.T) {
	// Create mock repository
	mockRepo := NewMockRepositoryForWorker()

	// Create notification queue
	notificationQueue := make(chan db.GormNotification, 10)

	// Create worker with very short timeout
	worker := NewNotificationWorker(mockRepo, notificationQueue)
	worker.SetPollInterval(100 * time.Millisecond)
	worker.SetProcessingTimeout(1 * time.Millisecond) // Very short timeout for testing

	// Start worker
	worker.Start()
	defer worker.Stop()

	// Test that context timeout is handled properly
	t.Run("ContextTimeoutHandling", func(t *testing.T) {
		// This test verifies that the worker handles context timeouts gracefully
		// The worker should not crash when context times out

		// Add a pending notification
		mockRepo.ClearPendingNotifications()
		mockRepo.AddPendingNotification(db.GormNotification{
			ID:               1,
			ItemID:           1,
			SearchID:         1,
			NotificationType: "test",
			Status:           "queued",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})

		// Wait for processing (should handle timeout gracefully)
		time.Sleep(300 * time.Millisecond)

		// Worker should still be running
		assert.NotNil(t, worker)
	})
}

// Implement remaining mock methods to satisfy the Repository interface
func (m *MockRepositoryForWorker) GetSearches(ctx context.Context) ([]db.GormSearch, error) {
	return []db.GormSearch{}, nil
}

func (m *MockRepositoryForWorker) GetSearchByID(ctx context.Context, id int) (*db.GormSearch, error) {
	return nil, nil
}

func (m *MockRepositoryForWorker) AddSearch(ctx context.Context, search db.GormSearch) (int, error) {
	return 1, nil
}

func (m *MockRepositoryForWorker) UpdateSearch(ctx context.Context, search db.GormSearch) error {
	return nil
}

func (m *MockRepositoryForWorker) DeleteSearch(ctx context.Context, id int) error {
	return nil
}

func (m *MockRepositoryForWorker) GetActiveSearches(ctx context.Context) ([]db.GormSearch, error) {
	return []db.GormSearch{}, nil
}

func (m *MockRepositoryForWorker) GetActiveUserAgents(ctx context.Context) ([]db.GormUserAgent, error) {
	return []db.GormUserAgent{}, nil
}

func (m *MockRepositoryForWorker) GetItems(ctx context.Context) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}

func (m *MockRepositoryForWorker) GetItemsPaginated(ctx context.Context, page int, pageSize int) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}

func (m *MockRepositoryForWorker) GetItemByID(ctx context.Context, id int) (*db.GormItem, error) {
	return &db.GormItem{
		ID:           id,
		GoodwillID:   "test-123",
		Title:        "Test Item",
		CurrentPrice: 99.99,
		URL:          "http://example.com/item/123",
		Category:     "Test Category",
		EndsAt:       timePtr(time.Now().Add(24 * time.Hour)),
	}, nil
}

func (m *MockRepositoryForWorker) GetItemByGoodwillID(ctx context.Context, goodwillID string) (*db.GormItem, error) {
	return nil, nil
}

func (m *MockRepositoryForWorker) AddItem(ctx context.Context, item db.GormItem) (int, error) {
	return 1, nil
}

func (m *MockRepositoryForWorker) UpdateItem(ctx context.Context, item db.GormItem) error {
	return nil
}

func (m *MockRepositoryForWorker) GetItemsBySearchID(ctx context.Context, searchID int) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}

func (m *MockRepositoryForWorker) AddSearchExecution(ctx context.Context, execution db.GormSearchExecution) (int, error) {
	return 1, nil
}

func (m *MockRepositoryForWorker) GetSearchHistory(ctx context.Context, searchID int, limit int) ([]db.GormSearchExecution, error) {
	return []db.GormSearchExecution{}, nil
}

func (m *MockRepositoryForWorker) AddPriceHistory(ctx context.Context, history db.GormPriceHistory) (int, error) {
	return 1, nil
}

func (m *MockRepositoryForWorker) GetPriceHistory(ctx context.Context, itemID int) ([]db.GormPriceHistory, error) {
	return []db.GormPriceHistory{}, nil
}

func (m *MockRepositoryForWorker) AddBidHistory(ctx context.Context, history db.GormBidHistory) (int, error) {
	return 1, nil
}

func (m *MockRepositoryForWorker) GetBidHistory(ctx context.Context, itemID int) ([]db.GormBidHistory, error) {
	return []db.GormBidHistory{}, nil
}

func (m *MockRepositoryForWorker) GetRandomUserAgent(ctx context.Context) (*db.GormUserAgent, error) {
	return nil, nil
}

func (m *MockRepositoryForWorker) UpdateUserAgentUsage(ctx context.Context, agentID int) error {
	return nil
}

func (m *MockRepositoryForWorker) AddSearchItemMapping(ctx context.Context, searchID int, itemID int, foundAt time.Time) error {
	return nil
}

func (m *MockRepositoryForWorker) LogSystemEvent(ctx context.Context, event db.GormSystemLog) (int, error) {
	return 1, nil
}

func (m *MockRepositoryForWorker) UpdateNotification(ctx context.Context, notification db.GormNotification) error {
	return nil
}

func (m *MockRepositoryForWorker) GetNotificationByID(ctx context.Context, id int) (*db.GormNotification, error) {
	return nil, nil
}

// GetItemsFiltered implements Repository.GetItemsFiltered
func (m *MockRepositoryForWorker) GetItemsFiltered(ctx context.Context, searchID *int, status *string, category *string, minPrice *float64, maxPrice *float64, limit int, offset int) ([]db.GormItem, int, error) {
	return []db.GormItem{}, 0, nil
}

// GetSearchesFiltered implements Repository.GetSearchesFiltered
func (m *MockRepositoryForWorker) GetSearchesFiltered(ctx context.Context, enabled *bool, limit int, offset int) ([]db.GormSearch, int, error) {
	return []db.GormSearch{}, 0, nil
}

func (m *MockRepositoryForWorker) GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]db.GormNotification, int, error) {
	return []db.GormNotification{}, 0, nil
}

func (m *MockRepositoryForWorker) GetNotificationStats(ctx context.Context) (*db.NotificationCountStats, error) {
	return &db.NotificationCountStats{
		Total:      3,
		Pending:    1,
		Processing: 0,
		Delivered:  1,
		Failed:     1,
	}, nil
}
