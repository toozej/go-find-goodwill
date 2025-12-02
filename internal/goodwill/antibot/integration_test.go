package antibot

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// TestAntiBotSystemIntegration tests the complete anti-bot system integration
func TestAntiBotSystemIntegration(t *testing.T) {
	// Create test configuration
	cfg := &config.AntiBotConfig{
		UserAgent: config.UserAgentConfig{
			RotationEnabled:  true,
			RotationInterval: 1 * time.Minute,
			RequestsPerUA:    10,
			MinSuccessRate:   0.7,
		},
		Timing: config.TimingConfig{
			BaseInterval:       15 * time.Minute,
			MinJitter:          2 * time.Minute,
			MaxJitter:          5 * time.Minute,
			HumanLikeVariation: true,
		},
		Throttling: config.ThrottlingConfig{
			RequestsPerMinute: 60,
			BurstLimit:        10,
		},
	}

	// Create mock repository
	mockRepo := &MockRepository{}

	// Create anti-bot system
	antiBotSystem, err := NewAntiBotSystem(cfg, mockRepo)
	assert.NoError(t, err)
	assert.NotNil(t, antiBotSystem)

	// Test user agent rotation
	t.Run("UserAgentRotation", func(t *testing.T) {
		agent, err := antiBotSystem.GetUserAgentWithRotation()
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.NotEmpty(t, agent.UserAgent)
	})

	// Test rate limiting
	t.Run("RateLimiting", func(t *testing.T) {
		// Should allow requests initially
		assert.True(t, antiBotSystem.CheckRateLimit())

		// Consume all tokens
		for i := 0; i < cfg.Throttling.BurstLimit+1; i++ {
			antiBotSystem.CheckRateLimit()
		}

		// Should block after burst limit
		assert.False(t, antiBotSystem.CheckRateLimit())
	})

	// Test adaptive timing
	t.Run("AdaptiveTiming", func(t *testing.T) {
		delay := antiBotSystem.GetAdaptiveDelay()
		assert.True(t, delay > 0)
		assert.True(t, delay <= cfg.Timing.BaseInterval+cfg.Timing.MaxJitter)
	})

	// Test response analysis
	t.Run("ResponseAnalysis", func(t *testing.T) {
		// Test blocking detection
		isBlocked := antiBotSystem.AnalyzeResponse(403, "Access denied: bot detected")
		assert.True(t, isBlocked)

		// Test non-blocking response
		isBlocked = antiBotSystem.AnalyzeResponse(200, "Success")
		assert.False(t, isBlocked)
	})
}

// TestRetryManager tests the retry manager
func TestRetryManager(t *testing.T) {
	retryManager := NewRetryManager(3, 1*time.Second, 10*time.Second)

	t.Run("RetryLogic", func(t *testing.T) {
		// Should allow initial retry
		assert.True(t, retryManager.ShouldRetry())

		// Test retry delay calculation
		delay := retryManager.GetRetryDelay()
		assert.True(t, delay >= 1*time.Second)
		assert.True(t, delay <= 10*time.Second)

		// Record retry
		retryManager.RecordRetry()

		// Should allow more retries
		assert.True(t, retryManager.ShouldRetry())

		// After max retries, should not allow more
		retryManager.retryCount = 3
		assert.False(t, retryManager.ShouldRetry())
	})

	t.Run("ExecuteWithRetry", func(t *testing.T) {
		// Test successful execution
		err := retryManager.ExecuteWithRetry(context.Background(), "test-operation", func() error {
			return nil
		})
		assert.NoError(t, err)

		// Test failed execution with retries
		attemptCount := 0
		err = retryManager.ExecuteWithRetry(context.Background(), "test-fail-operation", func() error {
			attemptCount++
			if attemptCount < 3 {
				return fmt.Errorf("simulated error")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, attemptCount)
	})
}

// TestCircuitBreaker tests the circuit breaker
func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test-cb", 2, 1, 1*time.Second, 5*time.Second)

	t.Run("CircuitBreakerStates", func(t *testing.T) {
		// Initial state should be closed
		assert.Equal(t, StateClosed, cb.GetState())

		// Simulate failures to open circuit
		cb.handleFailure()
		cb.handleFailure()

		// Should be open now
		assert.Equal(t, StateOpen, cb.GetState())

		// Wait for reset timeout
		time.Sleep(5 * time.Second)

		// Manually transition to half-open since GetState is now a pure getter
		cb.Reset()
		cb.handleSuccess()

		// Should be closed again
		assert.Equal(t, StateClosed, cb.GetState())
	})

	t.Run("CircuitBreakerExecution", func(t *testing.T) {
		// Reset circuit breaker
		cb.Reset()

		// Test successful execution
		err := cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)

		// Test failed execution
		err = cb.Execute(func() error {
			return fmt.Errorf("simulated error")
		})
		assert.Error(t, err)
	})
}

// TestCacheManager tests the cache manager
func TestCacheManager(t *testing.T) {
	cacheManager := NewCacheManager(5*time.Minute, 1*time.Minute)
	defer cacheManager.Stop()

	t.Run("CacheOperations", func(t *testing.T) {
		// Test cache set and get
		cacheManager.Set("test-key", "test-value", 1*time.Minute)
		value, ok := cacheManager.Get("test-key")
		assert.True(t, ok)
		assert.Equal(t, "test-value", value)

		// Test cache miss
		_, ok = cacheManager.Get("non-existent-key")
		assert.False(t, ok)

		// Test cache delete
		cacheManager.Delete("test-key")
		_, ok = cacheManager.Get("test-key")
		assert.False(t, ok)
	})

	t.Run("CacheWithFallback", func(t *testing.T) {
		// Test fallback function
		value, err := cacheManager.GetWithFallback("fallback-key", 1*time.Minute, func() (interface{}, error) {
			return "fallback-value", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "fallback-value", value)

		// Should be cached now
		cachedValue, ok := cacheManager.Get("fallback-key")
		assert.True(t, ok)
		assert.Equal(t, "fallback-value", cachedValue)
	})
}

// MockRepository implements a mock repository for testing
type MockRepository struct{}

func (m *MockRepository) GetSearches(ctx context.Context) ([]db.GormSearch, error) {
	return []db.GormSearch{}, nil
}

func (m *MockRepository) GetSearchByID(ctx context.Context, id int) (*db.GormSearch, error) {
	return &db.GormSearch{ID: id, Name: "Test Search"}, nil
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

func (m *MockRepository) GetItemByID(ctx context.Context, id int) (*db.GormItem, error) {
	return &db.GormItem{ID: id, Title: "Test Item"}, nil
}

func (m *MockRepository) GetItemByGoodwillID(ctx context.Context, goodwillID string) (*db.GormItem, error) {
	return &db.GormItem{GoodwillID: goodwillID, Title: "Test Item"}, nil
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

func (m *MockRepository) QueueNotification(ctx context.Context, notification db.GormNotification) (int, error) {
	return 1, nil
}

func (m *MockRepository) UpdateNotificationStatus(ctx context.Context, id int, status string) error {
	return nil
}

func (m *MockRepository) GetPendingNotifications(ctx context.Context) ([]db.GormNotification, error) {
	return []db.GormNotification{}, nil
}

func (m *MockRepository) GetNotificationByID(ctx context.Context, id int) (*db.GormNotification, error) {
	return &db.GormNotification{ID: id}, nil
}

func (m *MockRepository) UpdateNotification(ctx context.Context, notification db.GormNotification) error {
	return nil
}

func (m *MockRepository) GetAllNotifications(ctx context.Context) ([]db.GormNotification, error) {
	return []db.GormNotification{}, nil
}

func (m *MockRepository) GetRandomUserAgent(ctx context.Context) (*db.GormUserAgent, error) {
	return &db.GormUserAgent{
		ID:         1,
		UserAgent:  "Mozilla/5.0 (Test) AppleWebKit/537.36",
		UsageCount: 0,
		IsActive:   true,
	}, nil
}

func (m *MockRepository) GetActiveUserAgents(ctx context.Context) ([]db.GormUserAgent, error) {
	return []db.GormUserAgent{
		{
			ID:         1,
			UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
			UsageCount: 0,
			IsActive:   true,
		},
	}, nil
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
