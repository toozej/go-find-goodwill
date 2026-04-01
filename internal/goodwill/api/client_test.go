package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toozej/go-find-goodwill/internal/goodwill/antibot"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockRepository implements a mock repository for testing
type MockRepository struct {
	mock.Mock
}

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
func (m *MockRepository) GetSearchesFiltered(ctx context.Context, enabled *bool, limit int, offset int) ([]db.GormSearch, int, error) {
	return []db.GormSearch{}, 0, nil
}
func (m *MockRepository) GetItems(ctx context.Context) ([]db.GormItem, error) {
	return []db.GormItem{}, nil
}
func (m *MockRepository) GetItemsPaginated(ctx context.Context, page int, pageSize int) ([]db.GormItem, error) {
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
func (m *MockRepository) GetItemsFiltered(ctx context.Context, searchID *int, status *string, category *string, minPrice *float64, maxPrice *float64, limit int, offset int) ([]db.GormItem, int, error) {
	return []db.GormItem{}, 0, nil
}
func (m *MockRepository) AddSearchExecution(ctx context.Context, execution db.GormSearchExecution) (int, error) {
	return 1, nil
}
func (m *MockRepository) GetSearchHistory(ctx context.Context, searchID int, limit int) ([]db.GormSearchExecution, error) {
	return []db.GormSearchExecution{}, nil
}
func (m *MockRepository) AddSearchItemMapping(ctx context.Context, searchID int, itemID int, foundAt time.Time) error {
	return nil
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
func (m *MockRepository) GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]db.GormNotification, int, error) {
	return []db.GormNotification{}, 0, nil
}
func (m *MockRepository) GetNotificationStats(ctx context.Context) (*db.NotificationCountStats, error) {
	return &db.NotificationCountStats{}, nil
}
func (m *MockRepository) GetRandomUserAgent(ctx context.Context) (*db.GormUserAgent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return &db.GormUserAgent{UserAgent: "Mozilla/5.0 (Test) AppleWebKit/537.36"}, args.Error(1)
	}
	return args.Get(0).(*db.GormUserAgent), args.Error(1)
}
func (m *MockRepository) GetActiveUserAgents(ctx context.Context) ([]db.GormUserAgent, error) {
	return []db.GormUserAgent{}, nil
}
func (m *MockRepository) UpdateUserAgentUsage(ctx context.Context, agentID int) error {
	return nil
}
func (m *MockRepository) LogSystemEvent(ctx context.Context, event db.GormSystemLog) (int, error) {
	return 1, nil
}

func TestNewShopGoodwillClient(t *testing.T) {
	t.Run("should create client with valid config", func(t *testing.T) {
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     "https://buyerapi.shopgoodwill.com/api",
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		antiBotConfig := &config.AntiBotConfig{
			Retry: config.RetryConfig{
				MaxRetries: 3,
				BaseDelay:  1 * time.Second,
				MaxDelay:   30 * time.Second,
			},
			Circuit: config.CircuitConfig{
				FailureThreshold: 3,
				SuccessThreshold: 2,
				Timeout:          30 * time.Second,
				ResetTimeout:     5 * time.Minute,
			},
		}

		client, err := NewShopGoodwillClient(cfg, antiBotConfig, nil)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
		assert.NotNil(t, client.retryManager)
		assert.NotNil(t, client.circuitBreaker)
	})

	t.Run("should create client with nil anti-bot config", func(t *testing.T) {
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     "https://buyerapi.shopgoodwill.com/api",
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.httpClient)
		assert.NotNil(t, client.retryManager)
		assert.NotNil(t, client.circuitBreaker)
	})

	t.Run("should return error with nil config", func(t *testing.T) {
		client, err := NewShopGoodwillClient(nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("should return error with invalid base URL", func(t *testing.T) {
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     "invalid-url",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		if err != nil {
			assert.Contains(t, err.Error(), "invalid base URL")
		} else {
			assert.NotNil(t, client)
		}
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("should authenticate successfully", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/SignIn/Login", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			// Verify anti-bot browser headers are present
			assert.NotEmpty(t, r.Header.Get("User-Agent"))
			assert.Contains(t, r.Header.Get("Accept"), "application/json")
			assert.Equal(t, "en-US,en;q=0.9", r.Header.Get("Accept-Language"))
			assert.Equal(t, "cors", r.Header.Get("Sec-Fetch-Mode"))
			assert.Equal(t, "same-origin", r.Header.Get("Sec-Fetch-Site"))
			assert.Equal(t, "empty", r.Header.Get("Sec-Fetch-Dest"))

			// Parse request body
			var payload map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.NotEmpty(t, payload["userName"])
			assert.NotEmpty(t, payload["password"])
			assert.Equal(t, "0ac533a6087baed7", payload["appVersion"])

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      true,
				"accessToken": "test-token",
			}); err != nil {
				t.Logf("Failed to encode JSON response: %v", err)
			}
		}))
		defer server.Close()

		// Create client
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Test authentication
		err = client.Authenticate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("should handle authentication failure", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid credentials"))
		}))
		defer server.Close()

		// Create client
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Test authentication
		err = client.Authenticate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "operation 'authentication' failed after 3 attempts")
	})

	t.Run("should use cached token if valid", func(t *testing.T) {
		// Create client
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     "https://buyerapi.shopgoodwill.com/api",
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Set valid cached token
		client.authToken = "cached-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Test authentication (should not make network call)
		err = client.Authenticate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("should retry on failure when retry manager is configured", func(t *testing.T) {
		failCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			failCount++
			if failCount == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal server error"))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      true,
				"accessToken": "test-token",
			})
		}))
		defer server.Close()

		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		antiBotConfig := &config.AntiBotConfig{
			Retry: config.RetryConfig{
				MaxRetries: 3,
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   50 * time.Millisecond,
			},
		}

		client, err := NewShopGoodwillClient(cfg, antiBotConfig, nil)
		assert.NoError(t, err)

		err = client.Authenticate(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 2, failCount)
	})

	t.Run("should use rotated user agent from anti-bot system during authentication", func(t *testing.T) {
		expectedUA := "Mozilla/5.0 (Rotated-UA)"

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expectedUA, r.Header.Get("User-Agent"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      true,
				"accessToken": "test-token",
			})
		}))
		defer server.Close()

		mockRepo := &MockRepository{}
		mockRepo.On("GetRandomUserAgent", mock.Anything).Return(&db.GormUserAgent{
			ID:        1,
			UserAgent: expectedUA,
			IsActive:  true,
		}, nil)

		cfg := &config.ShopGoodwillConfig{
			Username:   "testuser",
			Password:   "testpass",
			APIBaseURL: server.URL,
			AppVersion: "0ac533a6087baed7",
		}

		antiBotCfg := &config.AntiBotConfig{
			UserAgent: config.UserAgentConfig{
				RotationEnabled:  false,
				RotationInterval: 1 * time.Minute,
			},
		}

		abs, _ := antibot.NewAntiBotSystem(antiBotCfg, mockRepo)
		client, err := NewShopGoodwillClient(cfg, nil, abs)
		assert.NoError(t, err)

		err = client.Authenticate(context.Background())
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestSearch(t *testing.T) {
	t.Run("should execute search successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/Search/ItemListing", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"searchResults": map[string]interface{}{
					"items":            []map[string]interface{}{},
					"totalResultCount": 0,
					"pageCount":        1,
					"currentPage":      1,
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, _ := NewShopGoodwillClient(cfg, nil, nil)
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		params := SearchParams{
			Category: "test-category",
			Page:     1,
			PageSize: 10,
		}

		result, err := client.Search(context.Background(), "test-query", params)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("should handle search failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("invalid query"))
		}))
		defer server.Close()

		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, _ := NewShopGoodwillClient(cfg, nil, nil)
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		result, err := client.Search(context.Background(), "test-query", SearchParams{})
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetItemDetails(t *testing.T) {
	t.Run("should get item details successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/ItemDetail/GetItemDetailModelByItemId/123", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"itemId":          123,
				"title":           "Test Item",
				"description":     "Test Description",
				"sellerName":      "test-seller",
				"currentPrice":    10.99,
				"imageName":       "https://example.com/image.jpg",
				"endTime":         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"categoryName":    "test-category",
				"conditionName":   "new",
				"shippingCost":    5.99,
				"shippingMethod":  "standard",
				"location":        "Test Location",
				"pickupAvailable": true,
				"returnsAccepted": true,
				"dimensions":      "10x10x10",
				"weight":          "1 lb",
				"material":        "plastic",
				"color":           "blue",
				"brand":           "Test Brand",
				"model":           "Test Model",
				"year":            2023,
			})
		}))
		defer server.Close()

		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, _ := NewShopGoodwillClient(cfg, nil, nil)
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		item, err := client.GetItemDetails(context.Background(), "123")
		assert.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, 123, item.ItemID)
	})

	t.Run("should handle item details failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("item not found"))
		}))
		defer server.Close()

		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			AppVersion:     "0ac533a6087baed7",
			MaxRetries:     1,
			RequestTimeout: 30 * time.Second,
		}

		client, _ := NewShopGoodwillClient(cfg, nil, nil)
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		item, err := client.GetItemDetails(context.Background(), "123")
		assert.Error(t, err)
		assert.Nil(t, item)
	})
}
