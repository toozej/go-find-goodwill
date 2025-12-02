package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// MockRepository is a simple mock for testing
type MockRepository struct{}

func (m *MockRepository) GetItemByID(ctx context.Context, id int) (*db.GormItem, error) {
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

func (m *MockRepository) UpdateNotificationStatus(ctx context.Context, id int, status string) error {
	return nil
}

func (m *MockRepository) GetPendingNotifications(ctx context.Context) ([]db.GormNotification, error) {
	return []db.GormNotification{}, nil
}

func (m *MockRepository) GetAllNotifications(ctx context.Context) ([]db.GormNotification, error) {
	return []db.GormNotification{}, nil
}

func (m *MockRepository) GetNotificationByID(ctx context.Context, id int) (*db.GormNotification, error) {
	return nil, nil
}

func (m *MockRepository) UpdateNotification(ctx context.Context, notification db.GormNotification) error {
	return nil
}

func (m *MockRepository) QueueNotification(ctx context.Context, notification db.GormNotification) (int, error) {
	return 1, nil
}

func (m *MockRepository) GetSearches(ctx context.Context) ([]db.GormSearch, error) {
	return []db.GormSearch{}, nil
}

func (m *MockRepository) GetSearchByID(ctx context.Context, id int) (*db.GormSearch, error) {
	return nil, nil
}

func (m *MockRepository) AddSearch(ctx context.Context, search db.GormSearch) (int, error) {
	return 1, nil
}

func (m *MockRepository) UpdateSearch(ctx context.Context, search db.GormSearch) error {
	return nil
}

func (m *MockRepository) DeleteSearch(ctx context.Context, id int) error {
	return nil
}

func (m *MockRepository) GetActiveSearches(ctx context.Context) ([]db.GormSearch, error) {
	return []db.GormSearch{}, nil
}

func (m *MockRepository) GetItems(ctx context.Context) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}

func (m *MockRepository) GetItemByGoodwillID(ctx context.Context, goodwillID string) (*db.GormItem, error) {
	return nil, nil
}

func (m *MockRepository) AddItem(ctx context.Context, item db.GormItem) (int, error) {
	return 1, nil
}

func (m *MockRepository) UpdateItem(ctx context.Context, item db.GormItem) error {
	return nil
}

func (m *MockRepository) GetItemsBySearchID(ctx context.Context, searchID int) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}

func (m *MockRepository) AddSearchExecution(ctx context.Context, execution db.GormSearchExecution) (int, error) {
	return 1, nil
}

func (m *MockRepository) GetSearchHistory(ctx context.Context, searchID int, limit int) ([]db.GormSearchExecution, error) {
	return []db.GormSearchExecution{}, nil
}

func (m *MockRepository) AddPriceHistory(ctx context.Context, history db.GormPriceHistory) (int, error) {
	return 1, nil
}

func (m *MockRepository) GetPriceHistory(ctx context.Context, itemID int) ([]db.GormPriceHistory, error) {
	return []db.GormPriceHistory{}, nil
}

func (m *MockRepository) AddBidHistory(ctx context.Context, history db.GormBidHistory) (int, error) {
	return 1, nil
}

func (m *MockRepository) GetBidHistory(ctx context.Context, itemID int) ([]db.GormBidHistory, error) {
	return []db.GormBidHistory{}, nil
}

func (m *MockRepository) GetRandomUserAgent(ctx context.Context) (*db.GormUserAgent, error) {
	return nil, nil
}

func (m *MockRepository) UpdateUserAgentUsage(ctx context.Context, agentID int) error {
	return nil
}

func (m *MockRepository) LogSystemEvent(ctx context.Context, event db.GormSystemLog) (int, error) {
	return 1, nil
}

func (m *MockRepository) AddSearchItemMapping(ctx context.Context, searchID int, itemID int, foundAt time.Time) error {
	return nil
}

func (m *MockRepository) GetItemsPaginated(ctx context.Context, page int, pageSize int) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}

func (m *MockRepository) GetItemsFiltered(ctx context.Context, searchID *int, status *string, category *string, minPrice *float64, maxPrice *float64, limit int, offset int) ([]db.GormItem, int, error) {
	return []db.GormItem{}, 0, nil
}

func (m *MockRepository) GetSearchesFiltered(ctx context.Context, enabled *bool, limit int, offset int) ([]db.GormSearch, int, error) {
	return []db.GormSearch{}, 0, nil
}

// Add missing GetActiveUserAgents method
func (m *MockRepository) GetActiveUserAgents(ctx context.Context) ([]db.GormUserAgent, error) {
	return []db.GormUserAgent{}, nil
}

// Add missing notification count methods
func (m *MockRepository) GetTotalNotificationCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockRepository) GetPendingNotificationCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockRepository) GetProcessingNotificationCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockRepository) GetDeliveredNotificationCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockRepository) GetFailedNotificationCount(ctx context.Context) (int, error) {
	return 0, nil
}

// Add missing GetNotificationsFiltered method
func (m *MockRepository) GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]db.GormNotification, int, error) {
	return []db.GormNotification{}, 0, nil
}

// Add missing GetRecentItemsForDeduplication method
func (m *MockRepository) GetRecentItemsForDeduplication(ctx context.Context, maxAge time.Duration, limit int, offset int) ([]db.GormItem, int, error) {
	return []db.GormItem{}, 0, nil
}

func (m *MockRepository) GetNotificationStats(ctx context.Context) (*db.NotificationCountStats, error) {
	return &db.NotificationCountStats{}, nil
}

func TestNotificationIntegration(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Gotify: config.GotifyConfig{
				Enabled:  true,
				URL:      "http://localhost:8080",
				Token:    "test-token",
				Priority: 5,
			},
			Slack: config.SlackConfig{
				Enabled:   true,
				Token:     "slack-token",
				ChannelID: "C123456",
			},
		},
	}

	// Create mock repository
	mockRepo := &MockRepository{}

	// Create notification integration
	integration, err := NewNotificationIntegration(cfg, mockRepo)
	assert.NoError(t, err)
	assert.NotNil(t, integration)

	// Start the integration
	integration.Start()
	defer integration.Stop()

	// Test queuing a notification for the new system
	item := &db.GormItem{
		ID:           1,
		GoodwillID:   "test-123",
		Title:        "Vintage Camera",
		CurrentPrice: 99.99,
		URL:          "http://example.com/item/123",
		Category:     "Cameras",
		EndsAt:       timePtr(time.Now().Add(24 * time.Hour)),
	}

	search := &db.GormSearch{
		ID:    1,
		Name:  "Test Search",
		Query: "vintage camera",
	}

	// Queue notification for new system
	err = integration.QueueNotificationForNewSystem(context.Background(), item, search)
	assert.NoError(t, err, "QueueNotificationForNewSystem should not return an error")

	// Give it a moment to process
	time.Sleep(100 * time.Millisecond)

	// Test backward compatibility - queue notification using old system
	oldNotification := db.GormNotification{
		ItemID:           1,
		SearchID:         1,
		NotificationType: "gotify",
		Status:           "queued",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = integration.QueueNotification(context.Background(), oldNotification)
	assert.NoError(t, err, "QueueNotification should not return an error")

	// Test stats
	stats, err := integration.GetNotificationStats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestNotificationManagerWithAllServices(t *testing.T) {
	// Create a comprehensive config with all services
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Gotify: config.GotifyConfig{
				Enabled:  true,
				URL:      "http://localhost:8080",
				Token:    "test-token",
				Priority: 5,
			},
			Slack: config.SlackConfig{
				Enabled:   true,
				Token:     "slack-token",
				ChannelID: "C123456",
			},
			Discord: config.DiscordConfig{
				Enabled:   true,
				Token:     "discord-token",
				ChannelID: "D123456",
			},
			Pushover: config.PushoverConfig{
				Enabled:     true,
				Token:       "pushover-token",
				RecipientID: "user123",
			},
			Pushbullet: config.PushbulletConfig{
				Enabled:        true,
				Token:          "pushbullet-token",
				DeviceNickname: "my-device",
			},
		},
	}

	// Create notification manager
	manager, err := NewNotificationManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Should have Gotify + Nikoksr (with Slack, Discord, Pushover, Pushbullet)
	assert.Equal(t, 2, len(manager.notifiers))
	assert.Equal(t, true, manager.condense)
}
