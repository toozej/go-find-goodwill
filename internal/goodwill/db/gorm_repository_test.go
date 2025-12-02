package db

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGormRepositorySearchOperations tests search-related operations
func TestGormRepositorySearchOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test AddSearch
	search := GormSearch{
		Name:                      "Test Search",
		Query:                     "test query",
		RegexPattern:              "test.*",
		Enabled:                   true,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
		NotificationThresholdDays: 7,
		MinPrice:                  floatPtr(10.0),
		MaxPrice:                  floatPtr(100.0),
		CategoryFilter:            "Electronics",
		SortBy:                    "price",
	}

	searchID, err := repo.AddSearch(context.Background(), search)
	assert.NoError(t, err)
	assert.Greater(t, searchID, 0)

	// Test GetSearchByID
	retrievedSearch, err := repo.GetSearchByID(context.Background(), searchID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSearch)
	assert.Equal(t, "Test Search", retrievedSearch.Name)

	// Test UpdateSearch
	search.ID = searchID
	search.Name = "Updated Test Search"
	err = repo.UpdateSearch(context.Background(), search)
	assert.NoError(t, err)

	updatedSearch, err := repo.GetSearchByID(context.Background(), searchID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Search", updatedSearch.Name)

	// Test GetSearches
	allSearches, err := repo.GetSearches(context.Background())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(allSearches), 1)

	// Test GetActiveSearches
	activeSearches, err := repo.GetActiveSearches(context.Background())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeSearches), 1)

	// Test DeleteSearch
	err = repo.DeleteSearch(context.Background(), searchID)
	assert.NoError(t, err)

	deletedSearch, err := repo.GetSearchByID(context.Background(), searchID)
	assert.NoError(t, err)
	assert.Nil(t, deletedSearch)
}

// TestGormRepositoryItemOperations tests item-related operations
func TestGormRepositoryItemOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test AddItem
	item := GormItem{
		GoodwillID:      "test-123",
		Title:           "Test Item",
		Seller:          "Test Seller",
		CurrentPrice:    25.99,
		BuyNowPrice:     floatPtr(30.99),
		URL:             "https://example.com/test",
		ImageURL:        "https://example.com/test.jpg",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
		Status:          "active",
		Category:        "Electronics",
		Subcategory:     "Computers",
		Condition:       "New",
		ShippingCost:    floatPtr(5.99),
		ShippingMethod:  "Standard",
		Description:     "Test description",
		Location:        "Test Location",
		PickupAvailable: true,
		ReturnsAccepted: true,
		WatchCount:      10,
		BidCount:        5,
		ViewCount:       100,
	}

	itemID, err := repo.AddItem(context.Background(), item)
	assert.NoError(t, err)
	assert.Greater(t, itemID, 0)

	// Test GetItemByID
	retrievedItem, err := repo.GetItemByID(context.Background(), itemID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedItem)
	assert.Equal(t, "Test Item", retrievedItem.Title)

	// Test GetItemByGoodwillID
	retrievedItemByGW, err := repo.GetItemByGoodwillID(context.Background(), "test-123")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedItemByGW)
	assert.Equal(t, "Test Item", retrievedItemByGW.Title)

	// Test UpdateItem
	item.ID = itemID
	item.Title = "Updated Test Item"
	err = repo.UpdateItem(context.Background(), item)
	assert.NoError(t, err)

	updatedItem, err := repo.GetItemByID(context.Background(), itemID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Item", updatedItem.Title)

	// Test GetItems
	allItems, err := repo.GetItems(context.Background())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(allItems), 1)
}

// TestGormRepositorySearchExecutionOperations tests search execution operations
func TestGormRepositorySearchExecutionOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// First add a search
	search := GormSearch{
		Name:      "Test Search",
		Query:     "test query",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	searchID, err := repo.AddSearch(context.Background(), search)
	assert.NoError(t, err)

	// Test AddSearchExecution
	execution := GormSearchExecution{
		SearchID:      searchID,
		ExecutedAt:    time.Now(),
		Status:        "completed",
		ItemsFound:    10,
		NewItemsFound: 5,
		ErrorMessage:  "",
		DurationMS:    1000,
	}

	executionID, err := repo.AddSearchExecution(context.Background(), execution)
	assert.NoError(t, err)
	assert.Greater(t, executionID, 0)

	// Test GetSearchHistory
	history, err := repo.GetSearchHistory(context.Background(), searchID, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(history), 1)
	assert.Equal(t, "completed", history[0].Status)
}

// TestGormRepositoryNotificationOperations tests notification operations
func TestGormRepositoryNotificationOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// First add a search and item
	search := GormSearch{
		Name:      "Test Search",
		Query:     "test query",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	searchID, err := repo.AddSearch(context.Background(), search)
	assert.NoError(t, err)

	item := GormItem{
		GoodwillID:   "test-notification-123",
		Title:        "Test Item for Notification",
		Seller:       "Test Seller",
		CurrentPrice: 25.99,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
		Status:       "active",
	}

	itemID, err := repo.AddItem(context.Background(), item)
	assert.NoError(t, err)

	// Test QueueNotification
	notification := GormNotification{
		ItemID:           itemID,
		SearchID:         searchID,
		NotificationType: "price_drop",
		Status:           "queued",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	notificationID, err := repo.QueueNotification(context.Background(), notification)
	assert.NoError(t, err)
	assert.Greater(t, notificationID, 0)

	// Test GetPendingNotifications
	pendingNotifications, err := repo.GetPendingNotifications(context.Background())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(pendingNotifications), 1)

	// Test UpdateNotificationStatus
	err = repo.UpdateNotificationStatus(context.Background(), notificationID, "sent")
	assert.NoError(t, err)

	// Test GetNotificationByID
	retrievedNotification, err := repo.GetNotificationByID(context.Background(), notificationID)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedNotification)
	assert.Equal(t, "sent", retrievedNotification.Status)

	// Test UpdateNotification
	notification.ID = notificationID
	notification.Status = "delivered"
	notification.DeliveredAt = timePtr(time.Now())
	err = repo.UpdateNotification(context.Background(), notification)
	assert.NoError(t, err)

	updatedNotification, err := repo.GetNotificationByID(context.Background(), notificationID)
	assert.NoError(t, err)
	assert.Equal(t, "delivered", updatedNotification.Status)

	// Test GetAllNotifications
	allNotifications, err := repo.GetAllNotifications(context.Background())
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(allNotifications), 1)
}

// TestGormRepositoryPriceHistoryOperations tests price history operations
func TestGormRepositoryPriceHistoryOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// First add an item
	item := GormItem{
		GoodwillID:   "test-price-123",
		Title:        "Test Item for Price History",
		Seller:       "Test Seller",
		CurrentPrice: 25.99,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
	}

	itemID, err := repo.AddItem(context.Background(), item)
	assert.NoError(t, err)

	// Test AddPriceHistory
	priceHistory := GormPriceHistory{
		ItemID:     itemID,
		Price:      25.99,
		PriceType:  "current",
		RecordedAt: time.Now(),
	}

	historyID, err := repo.AddPriceHistory(context.Background(), priceHistory)
	assert.NoError(t, err)
	assert.Greater(t, historyID, 0)

	// Test GetPriceHistory
	priceHistories, err := repo.GetPriceHistory(context.Background(), itemID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(priceHistories), 1)
	assert.Equal(t, "current", priceHistories[0].PriceType)
}

// TestGormRepositoryBidHistoryOperations tests bid history operations
func TestGormRepositoryBidHistoryOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// First add an item
	item := GormItem{
		GoodwillID:   "test-bid-123",
		Title:        "Test Item for Bid History",
		Seller:       "Test Seller",
		CurrentPrice: 25.99,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
	}

	itemID, err := repo.AddItem(context.Background(), item)
	assert.NoError(t, err)

	// Test AddBidHistory
	bidHistory := GormBidHistory{
		ItemID:     itemID,
		BidAmount:  20.50,
		Bidder:     "test_bidder",
		BidderID:   "bidder_123",
		RecordedAt: time.Now(),
	}

	historyID, err := repo.AddBidHistory(context.Background(), bidHistory)
	assert.NoError(t, err)
	assert.Greater(t, historyID, 0)

	// Test GetBidHistory
	bidHistories, err := repo.GetBidHistory(context.Background(), itemID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(bidHistories), 1)
	assert.Equal(t, "test_bidder", bidHistories[0].Bidder)
}

// TestGormRepositoryUserAgentOperations tests user agent operations
func TestGormRepositoryUserAgentOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test GetRandomUserAgent (should return nil if no agents exist)
	agent, err := repo.GetRandomUserAgent(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, agent)

	// Add a user agent directly via GORM for testing
	db := gormDB.GetDB()
	userAgent := GormUserAgent{
		UserAgent:  "Mozilla/5.0 (Test Agent)",
		IsActive:   true,
		UsageCount: 0,
	}
	result := db.Create(&userAgent)
	assert.NoError(t, result.Error)

	// Test GetRandomUserAgent again
	agent, err = repo.GetRandomUserAgent(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "Mozilla/5.0 (Test Agent)", agent.UserAgent)

	// Test UpdateUserAgentUsage
	err = repo.UpdateUserAgentUsage(context.Background(), userAgent.ID)
	assert.NoError(t, err)

	// Verify usage count was updated
	var updatedAgent GormUserAgent
	result = db.First(&updatedAgent, userAgent.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1, updatedAgent.UsageCount)
}

// TestGormRepositorySystemLogOperations tests system log operations
func TestGormRepositorySystemLogOperations(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test LogSystemEvent
	systemLog := GormSystemLog{
		Timestamp:  time.Now(),
		Level:      "INFO",
		Component:  "test_component",
		Message:    "Test system log message",
		Details:    "Test details",
		StackTrace: "",
	}

	logID, err := repo.LogSystemEvent(context.Background(), systemLog)
	assert.NoError(t, err)
	assert.Greater(t, logID, 0)
}

// TestGormRepositorySearchItemMapping tests search-item mapping operations
func TestGormRepositorySearchItemMapping(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Add a search and item first
	search := GormSearch{
		Name:      "Test Search",
		Query:     "test query",
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	searchID, err := repo.AddSearch(context.Background(), search)
	assert.NoError(t, err)

	item := GormItem{
		GoodwillID:   "test-mapping-123",
		Title:        "Test Item for Mapping",
		Seller:       "Test Seller",
		CurrentPrice: 25.99,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
	}

	itemID, err := repo.AddItem(context.Background(), item)
	assert.NoError(t, err)

	// Test AddSearchItemMapping
	foundAt := time.Now()
	err = repo.AddSearchItemMapping(context.Background(), searchID, itemID, foundAt)
	assert.NoError(t, err)

	// Test GetItemsBySearchID
	items, err := repo.GetItemsBySearchID(context.Background(), searchID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 1)
	assert.Equal(t, "Test Item for Mapping", items[0].Title)
}

// TestGormRepositoryErrorHandling tests error handling scenarios
func TestGormRepositoryErrorHandling(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test with disconnected database
	disconnectedRepo := &GormRepository{
		db: &GormDatabase{},
	}

	// Test various operations with disconnected database
	_, err := disconnectedRepo.GetSearches(context.Background())
	assert.Error(t, err)
	assert.Equal(t, "database not connected", err.Error())

	_, err = disconnectedRepo.GetSearchByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Equal(t, "database not connected", err.Error())

	_, err = disconnectedRepo.AddSearch(context.Background(), GormSearch{})
	assert.Error(t, err)
	assert.Equal(t, "database not connected", err.Error())

	err = disconnectedRepo.UpdateSearch(context.Background(), GormSearch{})
	assert.Error(t, err)
	assert.Equal(t, "database not connected", err.Error())

	err = disconnectedRepo.DeleteSearch(context.Background(), 1)
	assert.Error(t, err)
	assert.Equal(t, "database not connected", err.Error())

	// Test non-existent record retrieval
	_, err = repo.GetSearchByID(context.Background(), 999999)
	assert.NoError(t, err) // Should return nil, nil for non-existent search
	// assert.Nil(t, _)

	_, err = repo.GetItemByID(context.Background(), 999999)
	assert.NoError(t, err) // Should return nil, nil for non-existent item
	// assert.Nil(t, _)

	_, err = repo.GetNotificationByID(context.Background(), 999999)
	assert.NoError(t, err) // Should return nil, nil for non-existent notification
	// assert.Nil(t, _)
}

// Helper functions for tests
func setupTestGormDB(t *testing.T) (*GormDatabase, func()) {
	t.Helper()

	// Create in-memory database for testing
	config := &DBConfig{
		Path: ":memory:",
	}

	gormDB, err := NewGormDatabase(config)
	require.NoError(t, err)

	err = gormDB.Connect()
	require.NoError(t, err)

	// Auto-migrate all models
	err = gormDB.AutoMigrate(GetAllModels()...)
	require.NoError(t, err)

	cleanup := func() {
		// For in-memory DB, just close the connection
		_ = gormDB.Close()
	}

	return gormDB, cleanup
}

func floatPtr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// TestGormRepositoryPerformance tests performance characteristics
func TestGormRepositoryPerformance(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test bulk operations performance
	t.Run("BulkOperations", func(t *testing.T) {
		startTime := time.Now()

		// Create multiple searches
		for i := 0; i < 10; i++ {
			search := GormSearch{
				Name:      fmt.Sprintf("Performance Test Search %d", i),
				Query:     fmt.Sprintf("performance query %d", i),
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err := repo.AddSearch(context.Background(), search)
			assert.NoError(t, err)
		}

		// Create multiple items
		for i := 0; i < 20; i++ {
			item := GormItem{
				GoodwillID:   fmt.Sprintf("PERF_TEST_%d", i),
				Title:        fmt.Sprintf("Performance Test Item %d", i),
				CurrentPrice: float64(10 + i),
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
				FirstSeen:    time.Now(),
				LastSeen:     time.Now(),
			}

			_, err := repo.AddItem(context.Background(), item)
			assert.NoError(t, err)
		}

		// Measure time for bulk operations
		elapsed := time.Since(startTime)
		t.Logf("Bulk operations completed in %v", elapsed)

		// Verify data was created
		searches, err := repo.GetSearches(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 10, len(searches))

		items, err := repo.GetItems(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 20, len(items))
	})

	// Test query performance
	t.Run("QueryPerformance", func(t *testing.T) {
		startTime := time.Now()

		// Perform multiple queries
		for i := 0; i < 5; i++ {
			searches, err := repo.GetSearches(context.Background())
			assert.NoError(t, err)

			items, err := repo.GetItems(context.Background())
			assert.NoError(t, err)

			activeSearches, err := repo.GetActiveSearches(context.Background())
			assert.NoError(t, err)

			// Should have data from previous test
			assert.GreaterOrEqual(t, len(searches), 10)
			assert.GreaterOrEqual(t, len(items), 20)
			assert.GreaterOrEqual(t, len(activeSearches), 10)
		}

		elapsed := time.Since(startTime)
		t.Logf("Query operations completed in %v", elapsed)
	})
}

// TestGormRepositoryEdgeCases tests edge cases and boundary conditions
func TestGormRepositoryEdgeCases(t *testing.T) {
	// Setup test database
	gormDB, cleanup := setupTestGormDB(t)
	defer cleanup()

	repo := NewGormRepository(gormDB)

	// Test empty queries
	t.Run("EmptyQueries", func(t *testing.T) {
		searches, err := repo.GetSearches(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(searches), "Should have no searches initially")

		items, err := repo.GetItems(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(items), "Should have no items initially")

		activeSearches, err := repo.GetActiveSearches(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(activeSearches), "Should have no active searches initially")
	})

	// Test non-existent record retrieval
	t.Run("NonExistentRecords", func(t *testing.T) {
		search, err := repo.GetSearchByID(context.Background(), 999999)
		assert.NoError(t, err)
		assert.Nil(t, search)

		item, err := repo.GetItemByID(context.Background(), 999999)
		assert.NoError(t, err) // Should return nil, nil for non-existent item
		assert.Nil(t, item)

		notification, err := repo.GetNotificationByID(context.Background(), 999999)
		assert.NoError(t, err) // Should return nil, nil for non-existent notification
		assert.Nil(t, notification)
	})

	// Test empty string queries
	t.Run("EmptyStringQueries", func(t *testing.T) {
		item, err := repo.GetItemByGoodwillID(context.Background(), "")
		assert.NoError(t, err)
		assert.Nil(t, item)

		// Test with whitespace
		item, err = repo.GetItemByGoodwillID(context.Background(), "   ")
		assert.NoError(t, err)
		assert.Nil(t, item)
	})

	// Test large data values
	t.Run("LargeDataValues", func(t *testing.T) {
		// Create search with long name and query
		longName := "Very Long Search Name " + strings.Repeat("X", 500)
		longQuery := "Very Long Query " + strings.Repeat("Y", 1000)

		search := GormSearch{
			Name:      longName,
			Query:     longQuery,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		searchID, err := repo.AddSearch(context.Background(), search)
		assert.NoError(t, err)
		assert.Greater(t, searchID, 0)

		// Retrieve and verify
		retrievedSearch, err := repo.GetSearchByID(context.Background(), searchID)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedSearch)
		assert.Equal(t, longName, retrievedSearch.Name)
		assert.Equal(t, longQuery, retrievedSearch.Query)
	})
}
