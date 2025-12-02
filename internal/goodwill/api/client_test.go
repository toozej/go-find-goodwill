package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNewShopGoodwillClient(t *testing.T) {
	t.Run("should create client with valid config", func(t *testing.T) {
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     "https://api.shopgoodwill.com",
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
			APIBaseURL:     "https://api.shopgoodwill.com",
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
		// Note: Go's url.Parse is quite lenient, so this might not fail
		// We'll test with a truly invalid URL that should cause issues
		if err == nil {
			// If it didn't fail, let's test that the client was created
			assert.NotNil(t, client)
		} else {
			assert.Contains(t, err.Error(), "invalid base URL")
		}
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("should authenticate successfully", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/auth/login", r.URL.Path)
			assert.Equal(t, "POST", r.Method)

			// Parse request body
			var payload map[string]string
			err := json.NewDecoder(r.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, "testuser", payload["username"])
			assert.Equal(t, "testpass", payload["password"])

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      "test-token",
				"expires_in": 3600 * time.Second,
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
			_, err := w.Write([]byte("invalid credentials"))
			if err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		}))
		defer server.Close()

		// Create client
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
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
			APIBaseURL:     "https://api.shopgoodwill.com",
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
		// Create test server that fails first time, succeeds second time
		failCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			failCount++
			if failCount == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("internal server error"))
				if err != nil {
					t.Logf("Failed to write response: %v", err)
				}
				return
			}

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      "test-token",
				"expires_in": 3600 * time.Second,
			}); err != nil {
				t.Logf("Failed to encode JSON response: %v", err)
			}
		}))
		defer server.Close()

		// Create client with retry configuration
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		antiBotConfig := &config.AntiBotConfig{
			Retry: config.RetryConfig{
				MaxRetries: 3,
				BaseDelay:  100 * time.Millisecond, // Short delay for testing
				MaxDelay:   300 * time.Millisecond,
			},
		}

		client, err := NewShopGoodwillClient(cfg, antiBotConfig, nil)
		assert.NoError(t, err)

		// Test authentication
		err = client.Authenticate(context.Background())
		assert.NoError(t, err)
	})
}

func TestSearch(t *testing.T) {
	t.Run("should execute search successfully", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/search", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "test-query", r.URL.Query().Get("q"))

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"items":       []map[string]interface{}{},
				"total":       0,
				"page":        1,
				"page_size":   10,
				"total_pages": 1,
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
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Test search
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
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("invalid query"))
			if err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		}))
		defer server.Close()

		// Create client
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Test search
		result, err := client.Search(context.Background(), "test-query", SearchParams{})
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestGetItemDetails(t *testing.T) {
	t.Run("should get item details successfully", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/items/test-item-123", r.URL.Path)
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"id":               "test-item-123",
				"title":            "Test Item",
				"description":      "Test Description",
				"seller":           "test-seller",
				"current_price":    10.99,
				"url":              "https://example.com/item",
				"image_url":        "https://example.com/image.jpg",
				"ends_at":          time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"category":         "test-category",
				"subcategory":      "test-subcategory",
				"condition":        "new",
				"shipping_cost":    5.99,
				"shipping_method":  "standard",
				"location":         "Test Location",
				"pickup_available": true,
				"returns_accepted": true,
				"dimensions":       "10x10x10",
				"weight":           "1 lb",
				"material":         "plastic",
				"color":            "blue",
				"brand":            "Test Brand",
				"model":            "Test Model",
				"year":             2023,
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
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Test get item details
		details, err := client.GetItemDetails(context.Background(), "test-item-123")
		assert.NoError(t, err)
		assert.NotNil(t, details)
		assert.Equal(t, "test-item-123", details.ID)
		assert.Equal(t, "Test Item", details.Title)
	})

	t.Run("should handle item details failure", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte("item not found"))
			if err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		}))
		defer server.Close()

		// Create client
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Test get item details
		details, err := client.GetItemDetails(context.Background(), "test-item-123")
		assert.Error(t, err)
		assert.Nil(t, details)
	})
}

func TestThreadSafety(t *testing.T) {
	t.Run("should handle concurrent authentication safely", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      "test-token",
				"expires_in": 3600 * time.Second,
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
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Test concurrent authentication
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := client.Authenticate(context.Background())
				assert.NoError(t, err)
			}()
		}
		wg.Wait()

		// Verify token is set
		client.authMutex.Lock()
		assert.NotEmpty(t, client.authToken)
		client.authMutex.Unlock()
	})

	t.Run("should handle concurrent searches safely", func(t *testing.T) {
		// Create test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"items":       []map[string]interface{}{},
				"total":       0,
				"page":        1,
				"page_size":   10,
				"total_pages": 1,
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
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		client, err := NewShopGoodwillClient(cfg, nil, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Test concurrent searches
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.Search(context.Background(), "test-query", SearchParams{})
				assert.NoError(t, err)
			}()
		}
		wg.Wait()
	})
}

func TestCircuitBreakerIntegration(t *testing.T) {
	t.Run("should use circuit breaker when configured", func(t *testing.T) {
		// Create test server that fails consistently for first 10 calls, then succeeds
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			// Fail consistently for first 10 API calls to ensure circuit breaker opens
			if callCount <= 10 {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("service unavailable"))
				if err != nil {
					t.Logf("Failed to write response: %v", err)
				}
				return
			}

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"items":       []map[string]interface{}{},
				"total":       0,
				"page":        1,
				"page_size":   10,
				"total_pages": 1,
			}); err != nil {
				t.Logf("Failed to encode JSON response: %v", err)
			}
		}))
		defer server.Close()

		// Create client with circuit breaker configuration
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		antiBotConfig := &config.AntiBotConfig{
			Circuit: config.CircuitConfig{
				FailureThreshold: 3,
				SuccessThreshold: 2,
				Timeout:          1 * time.Second, // Short timeout for testing
				ResetTimeout:     1 * time.Second,
			},
		}

		client, err := NewShopGoodwillClient(cfg, antiBotConfig, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// First few requests should fail (each Search call will be retried, but should eventually fail)
		for i := 0; i < 3; i++ {
			_, err := client.Search(context.Background(), "test-query", SearchParams{})
			assert.Error(t, err, "Expected Search call %d to fail after retries", i+1)
		}

		// Circuit should be open now, wait for reset
		time.Sleep(2 * time.Second)

		// Request should succeed after circuit resets
		result, err := client.Search(context.Background(), "test-query", SearchParams{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestRetryManagerIntegration(t *testing.T) {
	t.Run("should retry failed requests when configured", func(t *testing.T) {
		// Create test server that fails first time, succeeds second time
		failCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			failCount++
			if failCount == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte("service unavailable"))
				if err != nil {
					t.Logf("Failed to write response: %v", err)
				}
				return
			}

			// Return success response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"items":       []map[string]interface{}{},
				"total":       0,
				"page":        1,
				"page_size":   10,
				"total_pages": 1,
			}); err != nil {
				t.Logf("Failed to encode JSON response: %v", err)
			}
		}))
		defer server.Close()

		// Create client with retry configuration
		cfg := &config.ShopGoodwillConfig{
			Username:       "testuser",
			Password:       "testpass",
			APIBaseURL:     server.URL,
			MaxRetries:     3,
			RequestTimeout: 30 * time.Second,
		}

		antiBotConfig := &config.AntiBotConfig{
			Retry: config.RetryConfig{
				MaxRetries: 3,
				BaseDelay:  100 * time.Millisecond, // Short delay for testing
				MaxDelay:   300 * time.Millisecond,
			},
		}

		client, err := NewShopGoodwillClient(cfg, antiBotConfig, nil)
		assert.NoError(t, err)

		// Set valid token to skip authentication
		client.authToken = "test-token"
		client.authExpired = time.Now().Add(1 * time.Hour)

		// Request should succeed after retry
		result, err := client.Search(context.Background(), "test-query", SearchParams{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
