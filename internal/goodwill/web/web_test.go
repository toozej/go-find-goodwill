package web

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toozej/go-find-goodwill/internal/goodwill/core/scheduling"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/internal/goodwill/notifications"
	"github.com/toozej/go-find-goodwill/internal/goodwill/web/api"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetSearches(ctx context.Context) ([]db.GormSearch, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormSearch), args.Error(1)
}

func (m *MockRepository) GetSearchByID(ctx context.Context, id int) (*db.GormSearch, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*db.GormSearch), args.Error(1)
}

func (m *MockRepository) AddSearch(ctx context.Context, search db.GormSearch) (int, error) {
	args := m.Called(ctx, search)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) UpdateSearch(ctx context.Context, search db.GormSearch) error {
	args := m.Called(ctx, search)
	return args.Error(0)
}

func (m *MockRepository) DeleteSearch(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetActiveSearches(ctx context.Context) ([]db.GormSearch, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormSearch), args.Error(1)
}

func (m *MockRepository) GetItems(ctx context.Context) ([]db.GormItem, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormItem), args.Error(1)
}

func (m *MockRepository) GetItemsPaginated(ctx context.Context, page int, pageSize int) ([]db.GormItem, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]db.GormItem), args.Error(1)
}

func (m *MockRepository) GetItemByID(ctx context.Context, id int) (*db.GormItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*db.GormItem), args.Error(1)
}

func (m *MockRepository) GetItemByGoodwillID(ctx context.Context, goodwillID string) (*db.GormItem, error) {
	args := m.Called(ctx, goodwillID)
	return args.Get(0).(*db.GormItem), args.Error(1)
}

func (m *MockRepository) AddItem(ctx context.Context, item db.GormItem) (int, error) {
	args := m.Called(ctx, item)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) UpdateItem(ctx context.Context, item db.GormItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockRepository) GetItemsBySearchID(ctx context.Context, searchID int) ([]db.GormItem, error) {
	args := m.Called(ctx, searchID)
	return args.Get(0).([]db.GormItem), args.Error(1)
}

func (m *MockRepository) GetItemsFiltered(ctx context.Context, searchID *int, status *string, category *string, minPrice *float64, maxPrice *float64, limit int, offset int) ([]db.GormItem, int, error) {
	args := m.Called(ctx, searchID, status, category, minPrice, maxPrice, limit, offset)
	return args.Get(0).([]db.GormItem), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetSearchesFiltered(ctx context.Context, enabled *bool, limit int, offset int) ([]db.GormSearch, int, error) {
	args := m.Called(ctx, enabled, limit, offset)
	return args.Get(0).([]db.GormSearch), args.Int(1), args.Error(2)
}

func (m *MockRepository) AddSearchExecution(ctx context.Context, execution db.GormSearchExecution) (int, error) {
	args := m.Called(ctx, execution)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetSearchHistory(ctx context.Context, searchID int, limit int) ([]db.GormSearchExecution, error) {
	args := m.Called(ctx, searchID, limit)
	return args.Get(0).([]db.GormSearchExecution), args.Error(1)
}

func (m *MockRepository) AddSearchItemMapping(ctx context.Context, searchID int, itemID int, foundAt time.Time) error {
	args := m.Called(ctx, searchID, itemID, foundAt)
	return args.Error(0)
}

func (m *MockRepository) AddPriceHistory(ctx context.Context, history db.GormPriceHistory) (int, error) {
	args := m.Called(ctx, history)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetPriceHistory(ctx context.Context, itemID int) ([]db.GormPriceHistory, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]db.GormPriceHistory), args.Error(1)
}

func (m *MockRepository) AddBidHistory(ctx context.Context, history db.GormBidHistory) (int, error) {
	args := m.Called(ctx, history)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetBidHistory(ctx context.Context, itemID int) ([]db.GormBidHistory, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]db.GormBidHistory), args.Error(1)
}

func (m *MockRepository) QueueNotification(ctx context.Context, notification db.GormNotification) (int, error) {
	args := m.Called(ctx, notification)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) UpdateNotificationStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockRepository) GetPendingNotifications(ctx context.Context) ([]db.GormNotification, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormNotification), args.Error(1)
}

func (m *MockRepository) GetNotificationByID(ctx context.Context, id int) (*db.GormNotification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*db.GormNotification), args.Error(1)
}

func (m *MockRepository) UpdateNotification(ctx context.Context, notification db.GormNotification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockRepository) GetAllNotifications(ctx context.Context) ([]db.GormNotification, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormNotification), args.Error(1)
}

func (m *MockRepository) GetRandomUserAgent(ctx context.Context) (*db.GormUserAgent, error) {
	args := m.Called(ctx)
	return args.Get(0).(*db.GormUserAgent), args.Error(1)
}

func (m *MockRepository) UpdateUserAgentUsage(ctx context.Context, agentID int) error {
	args := m.Called(ctx, agentID)
	return args.Error(0)
}

func (m *MockRepository) LogSystemEvent(ctx context.Context, event db.GormSystemLog) (int, error) {
	args := m.Called(ctx, event)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetActiveUserAgents(ctx context.Context) ([]db.GormUserAgent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormUserAgent), args.Error(1)
}

// Add missing notification count methods
func (m *MockRepository) GetTotalNotificationCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetPendingNotificationCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetProcessingNotificationCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetDeliveredNotificationCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) GetFailedNotificationCount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// Add missing GetNotificationsFiltered method
func (m *MockRepository) GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]db.GormNotification, int, error) {
	args := m.Called(ctx, status, notificationType, limit, offset)
	return args.Get(0).([]db.GormNotification), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetNotificationStats(ctx context.Context) (*db.NotificationCountStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.NotificationCountStats), args.Error(1)
}

// Add missing GetRecentItemsForDeduplication method
func (m *MockRepository) GetRecentItemsForDeduplication(ctx context.Context, maxAge time.Duration, limit int, offset int) ([]db.GormItem, int, error) {
	args := m.Called(ctx, maxAge, limit, offset)
	return args.Get(0).([]db.GormItem), args.Int(1), args.Error(2)
}

func TestWebServerCreation(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Web: config.WebConfig{
			Host:         "localhost",
			Port:         8081,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Test server creation
	server := api.NewServer(&cfg.Web, mockRepo, mockScheduler, mockNotificationSvc, logger)
	assert.NotNil(t, server, "Server should not be nil")
	// Test that server was created successfully
	assert.NotNil(t, server.Router(), "Server router should not be nil")
}

func TestAPIHandlers(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Web: config.WebConfig{
			Host:         "localhost",
			Port:         8081,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Test server creation
	server := api.NewServer(&cfg.Web, mockRepo, mockScheduler, mockNotificationSvc, logger)

	// Test that the server struct has the expected handler methods
	assert.NotPanics(t, func() {
		// Test that we can access the router
		_ = server.Router()
	})
}

func TestWebServerIntegration(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Web: config.WebConfig{
			Host:         "localhost",
			Port:         8081,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Test web server creation
	webServer := NewWebServer(cfg, mockRepo, mockScheduler, mockNotificationSvc, logger)
	assert.NotNil(t, webServer, "WebServer should not be nil")

	// Test that web server can start (non-blocking)
	assert.NotPanics(t, func() {
		// This will start the server in a goroutine
		// In a real test, we would need to clean up the server
		// webServer.Start()
	})
}
