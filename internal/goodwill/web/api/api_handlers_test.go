package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toozej/go-find-goodwill/internal/goodwill/core/scheduling"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/internal/goodwill/notifications"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// MockRepository is a mock implementation of the Repository interface for API handler testing
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
	return args.Get(0).(int), args.Error(1)
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
	return args.Get(0).(int), args.Error(1)
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
	return args.Get(0).(int), args.Error(1)
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
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRepository) GetPriceHistory(ctx context.Context, itemID int) ([]db.GormPriceHistory, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]db.GormPriceHistory), args.Error(1)
}

func (m *MockRepository) AddBidHistory(ctx context.Context, history db.GormBidHistory) (int, error) {
	args := m.Called(ctx, history)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRepository) GetBidHistory(ctx context.Context, itemID int) ([]db.GormBidHistory, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]db.GormBidHistory), args.Error(1)
}

func (m *MockRepository) QueueNotification(ctx context.Context, notification db.GormNotification) (int, error) {
	args := m.Called(ctx, notification)
	return args.Get(0).(int), args.Error(1)
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

func (m *MockRepository) GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]db.GormNotification, int, error) {
	args := m.Called(ctx, status, notificationType, limit, offset)
	return args.Get(0).([]db.GormNotification), args.Int(1), args.Error(2)
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
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRepository) GetActiveUserAgents(ctx context.Context) ([]db.GormUserAgent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.GormUserAgent), args.Error(1)
}

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

func (m *MockRepository) GetRecentItemsForDeduplication(ctx context.Context, maxAge time.Duration, limit int, offset int) ([]db.GormItem, int, error) {
	args := m.Called(ctx, maxAge, limit, offset)
	return args.Get(0).([]db.GormItem), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetNotificationStats(ctx context.Context) (*db.NotificationCountStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.NotificationCountStats), args.Error(1)
}

func TestItemsHandlerWithDatabaseFiltering(t *testing.T) {
	// Create test configuration
	cfg := &config.WebConfig{
		Host:         "localhost",
		Port:         8081,
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  120,
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Create server
	server := NewServer(cfg, mockRepo, mockScheduler, mockNotificationSvc, logger)

	// Test data
	testItems := []db.GormItem{
		{
			ID:           1,
			GoodwillID:   "item1",
			Title:        "Test Item 1",
			Status:       "active",
			Category:     "Electronics",
			CurrentPrice: 100.0,
		},
		{
			ID:           2,
			GoodwillID:   "item2",
			Title:        "Test Item 2",
			Status:       "sold",
			Category:     "Furniture",
			CurrentPrice: 200.0,
		},
	}

	// Mock the GetItemsFiltered method
	mockRepo.On("GetItemsFiltered",
		mock.Anything, // context
		mock.Anything, // searchID *int (nil)
		mock.Anything, // status *string (nil)
		mock.Anything, // category *string (nil)
		mock.Anything, // minPrice *float64 (nil)
		mock.Anything, // maxPrice *float64 (nil)
		20,            // limit
		0,             // offset
	).Return(testItems, len(testItems), nil)

	// Create a test request
	req, err := http.NewRequest("GET", "/api/v1/items", nil)
	assert.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	server.handleGetItems(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse the response
	var response ItemListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, 2, len(response.Items))

	// Verify the items
	assert.Equal(t, "item1", response.Items[0].GoodwillID)
	assert.Equal(t, "item2", response.Items[1].GoodwillID)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestSearchesHandlerWithDatabaseFiltering(t *testing.T) {
	// Create test configuration
	cfg := &config.WebConfig{
		Host:         "localhost",
		Port:         8081,
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  120,
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Create server
	server := NewServer(cfg, mockRepo, mockScheduler, mockNotificationSvc, logger)

	// Test data
	testSearches := []db.GormSearch{
		{
			ID:      1,
			Name:    "Test Search 1",
			Query:   "test query 1",
			Enabled: true,
		},
		{
			ID:      2,
			Name:    "Test Search 2",
			Query:   "test query 2",
			Enabled: false,
		},
	}

	// Mock the GetSearchesFiltered method
	mockRepo.On("GetSearchesFiltered",
		mock.Anything, // context
		mock.Anything, // enabled *bool (nil)
		20,            // limit
		0,             // offset
	).Return(testSearches, len(testSearches), nil)

	// Create a test request
	req, err := http.NewRequest("GET", "/api/v1/searches", nil)
	assert.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	server.handleGetSearches(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse the response
	var response SearchListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, 2, len(response.Searches))

	// Verify the searches
	assert.Equal(t, "Test Search 1", response.Searches[0].Name)
	assert.Equal(t, "Test Search 2", response.Searches[1].Name)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestItemsHandlerWithFiltering(t *testing.T) {
	// Create test configuration
	cfg := &config.WebConfig{
		Host:         "localhost",
		Port:         8081,
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  120,
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Create server
	server := NewServer(cfg, mockRepo, mockScheduler, mockNotificationSvc, logger)

	// Test data - only active items
	testItems := []db.GormItem{
		{
			ID:           1,
			GoodwillID:   "item1",
			Title:        "Test Item 1",
			Status:       "active",
			Category:     "Electronics",
			CurrentPrice: 100.0,
		},
	}

	// Mock the GetItemsFiltered method with status filter
	statusFilter := "active"
	mockRepo.On("GetItemsFiltered",
		mock.Anything, // context
		mock.Anything, // searchID *int (nil)
		&statusFilter, // status *string
		mock.Anything, // category *string (nil)
		mock.Anything, // minPrice *float64 (nil)
		mock.Anything, // maxPrice *float64 (nil)
		20,            // limit
		0,             // offset
	).Return(testItems, len(testItems), nil)

	// Create a test request with status filter
	req, err := http.NewRequest("GET", "/api/v1/items?status=active", nil)
	assert.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	server.handleGetItems(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response
	var response ItemListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response - should only have active items
	assert.Equal(t, 1, response.Total)
	assert.Equal(t, 1, len(response.Items))
	assert.Equal(t, "active", response.Items[0].Status)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

func TestNotificationsHandlerWithDatabaseFiltering(t *testing.T) {
	// Create test configuration
	cfg := &config.WebConfig{
		Host:         "localhost",
		Port:         8081,
		ReadTimeout:  30,
		WriteTimeout: 30,
		IdleTimeout:  120,
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock repository
	mockRepo := new(MockRepository)

	// Create mock services
	mockScheduler := new(scheduling.Scheduler)
	mockNotificationSvc := new(notifications.NotificationIntegration)

	// Create server
	server := NewServer(cfg, mockRepo, mockScheduler, mockNotificationSvc, logger)

	// Test data - only notifications that match the filter
	filteredNotifications := []db.GormNotification{
		{
			ID:               1,
			ItemID:           1,
			SearchID:         1,
			NotificationType: "new_item",
			Status:           "pending",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// Mock the GetNotificationsFiltered method
	statusFilter := "pending"
	notificationTypeFilter := "new_item"
	mockRepo.On("GetNotificationsFiltered",
		mock.Anything,           // context
		&statusFilter,           // status *string
		&notificationTypeFilter, // notificationType *string
		20,                      // limit
		0,                       // offset
	).Return(filteredNotifications, len(filteredNotifications), nil)

	// Create a test request with filters
	req, err := http.NewRequest("GET", "/api/v1/notifications?status=pending&type=new_item", nil)
	assert.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	server.handleGetNotifications(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse the response
	var response NotificationListResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response
	assert.Equal(t, 1, response.Total)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, 1, len(response.Notifications))

	// Verify the notifications
	assert.Equal(t, "pending", response.Notifications[0].Status)
	assert.Equal(t, "new_item", response.Notifications[0].NotificationType)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
