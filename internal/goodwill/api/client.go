package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/antibot"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// ShopGoodwillClient represents the ShopGoodwill API client
type ShopGoodwillClient struct {
	config         *config.ShopGoodwillConfig
	httpClient     *http.Client
	baseURL        *url.URL
	authToken      string
	authExpired    time.Time
	authMutex      sync.Mutex
	retryManager   *antibot.RetryManager
	circuitBreaker *antibot.CircuitBreaker
	antiBotSystem  *antibot.AntiBotSystem
}

// NewShopGoodwillClient creates a new ShopGoodwill API client
func NewShopGoodwillClient(cfg *config.ShopGoodwillConfig, antiBotConfig *config.AntiBotConfig, antiBotSystem *antibot.AntiBotSystem) (*ShopGoodwillClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Parse base URL
	baseURL, err := url.Parse(cfg.APIBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: cfg.RequestTimeout,
	}

	// Initialize retry manager and circuit breaker with configuration
	var retryManager *antibot.RetryManager
	var circuitBreaker *antibot.CircuitBreaker

	if antiBotConfig != nil {
		// Use configured retry settings
		retryManager = antibot.NewRetryManager(
			antiBotConfig.Retry.MaxRetries,
			antiBotConfig.Retry.BaseDelay,
			antiBotConfig.Retry.MaxDelay,
		)

		// Use configured circuit breaker settings
		circuitBreaker = antibot.NewCircuitBreaker(
			"api-client",
			antiBotConfig.Circuit.FailureThreshold,
			antiBotConfig.Circuit.SuccessThreshold,
			antiBotConfig.Circuit.Timeout,
			antiBotConfig.Circuit.ResetTimeout,
		)
	} else {
		// Fallback to default values
		retryManager = antibot.NewRetryManager(3, 1*time.Second, 30*time.Second)
		circuitBreaker = antibot.NewCircuitBreaker("api-client", 3, 2, 30*time.Second, 5*time.Minute)
	}

	client := &ShopGoodwillClient{
		config:         cfg,
		httpClient:     httpClient,
		baseURL:        baseURL,
		retryManager:   retryManager,
		circuitBreaker: circuitBreaker,
		antiBotSystem:  antiBotSystem,
	}

	return client, nil
}

// Authenticate authenticates with ShopGoodwill API
func (c *ShopGoodwillClient) Authenticate(ctx context.Context) error {
	// Execute with circuit breaker if available
	if c.circuitBreaker != nil {
		err := c.circuitBreaker.Execute(func() error {
			return c.authenticateInternal(ctx)
		})
		return err
	}

	return c.authenticateInternal(ctx)
}

// authenticateInternal performs the actual authentication logic
func (c *ShopGoodwillClient) authenticateInternal(ctx context.Context) error {
	// Acquire lock immediately to ensure thread-safe access to authToken and authExpired
	c.authMutex.Lock()
	defer c.authMutex.Unlock()

	// Check if token is still valid
	if c.authToken != "" && time.Now().Before(c.authExpired) {
		return nil
	}

	// Prepare authentication request
	authURL := c.baseURL.JoinPath("auth", "login")
	if authURL == nil {
		return fmt.Errorf("failed to construct auth URL")
	}

	// Create request payload
	payload := map[string]string{
		"username": c.config.Username,
		"password": c.config.Password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", authURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request with retry manager
	var resp *http.Response
	var execErr error
	if c.retryManager != nil {
		execErr = c.retryManager.ExecuteWithRetry(ctx, "authentication", func() error {
			var err error
			resp, err = c.executeRequestWithErrorHandling(req)
			return err
		})

		if execErr != nil {
			return execErr
		}
	} else {
		// Fallback to simple execution
		resp, execErr = c.executeRequestWithErrorHandling(req)
		if execErr != nil {
			return fmt.Errorf("failed to execute auth request: %w", execErr)
		}
	}

	// Parse response
	var authResponse struct {
		Token     string        `json:"token"`
		ExpiresIn time.Duration `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Ensure body is closed
	if err := resp.Body.Close(); err != nil {
		log.Errorf("Failed to close auth response body: %v", err)
	}

	// Set authentication token and expiration (already thread-safe due to defer unlock)
	c.authToken = authResponse.Token
	c.authExpired = time.Now().Add(authResponse.ExpiresIn)

	log.Infof("Successfully authenticated with ShopGoodwill API")

	return nil
}

// Search executes a search query on ShopGoodwill
func (c *ShopGoodwillClient) Search(ctx context.Context, query string, params SearchParams) (*SearchResponse, error) {
	// Ensure we're authenticated
	if err := c.Authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Prepare search URL
	searchURL := c.baseURL.JoinPath("search")
	if searchURL == nil {
		return nil, fmt.Errorf("failed to construct search URL")
	}

	// Add query parameters
	q := searchURL.Query()
	q.Set("q", query)

	if params.Category != "" {
		q.Set("category", params.Category)
	}
	if params.Seller != "" {
		q.Set("seller", params.Seller)
	}
	if params.Condition != "" {
		q.Set("condition", params.Condition)
	}
	if params.Shipping != "" {
		q.Set("shipping", params.Shipping)
	}
	if params.MinPrice > 0 {
		q.Set("min_price", fmt.Sprintf("%.2f", params.MinPrice))
	}
	if params.MaxPrice > 0 {
		q.Set("max_price", fmt.Sprintf("%.2f", params.MaxPrice))
	}
	if params.SortBy != "" {
		q.Set("sort_by", params.SortBy)
	}
	if params.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", params.PageSize))
	}

	searchURL.RawQuery = q.Encode()

	// Create request with anti-bot features
	req, err := c.createRequestWithAntiBot(ctx, "GET", searchURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	// Execute request with anti-bot protection
	resp, err := c.executeRequestWithAntiBot(ctx, req, "search")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &searchResponse, nil
}

// GetItemDetails retrieves detailed information for a specific item
func (c *ShopGoodwillClient) GetItemDetails(ctx context.Context, itemID string) (*ItemDetails, error) {
	// Ensure we're authenticated
	if err := c.Authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Prepare item details URL
	itemURL := c.baseURL.JoinPath("items", itemID)
	if itemURL == nil {
		return nil, fmt.Errorf("failed to construct item URL")
	}

	// Create request with anti-bot features
	req, err := c.createRequestWithAntiBot(ctx, "GET", itemURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create item request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	// Execute request with anti-bot protection
	resp, err := c.executeRequestWithAntiBot(ctx, req, "get-item-details")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse response
	var itemDetails ItemDetails
	if err := json.NewDecoder(resp.Body).Decode(&itemDetails); err != nil {
		return nil, fmt.Errorf("failed to decode item details: %w", err)
	}

	return &itemDetails, nil
}

// SearchParams represents search parameters
type SearchParams struct {
	Category  string
	Seller    string
	Condition string
	Shipping  string
	MinPrice  float64
	MaxPrice  float64
	SortBy    string
	Page      int
	PageSize  int
}

// SearchResponse represents a search response from ShopGoodwill API
type SearchResponse struct {
	Items      []SearchItem `json:"items"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// SearchItem represents an item in search results
type SearchItem struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Seller         string   `json:"seller"`
	CurrentPrice   float64  `json:"current_price"`
	BuyNowPrice    *float64 `json:"buy_now_price"`
	URL            string   `json:"url"`
	ImageURL       string   `json:"image_url"`
	EndsAt         string   `json:"ends_at"`
	Category       string   `json:"category"`
	Subcategory    string   `json:"subcategory"`
	Condition      string   `json:"condition"`
	ShippingCost   *float64 `json:"shipping_cost"`
	ShippingMethod string   `json:"shipping_method"`
	WatchCount     int      `json:"watch_count"`
	BidCount       int      `json:"bid_count"`
	ViewCount      int      `json:"view_count"`
}

// ItemDetails represents detailed item information
type ItemDetails struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Seller          string   `json:"seller"`
	CurrentPrice    float64  `json:"current_price"`
	BuyNowPrice     *float64 `json:"buy_now_price"`
	URL             string   `json:"url"`
	ImageURL        string   `json:"image_url"`
	EndsAt          string   `json:"ends_at"`
	Category        string   `json:"category"`
	Subcategory     string   `json:"subcategory"`
	Condition       string   `json:"condition"`
	ShippingCost    *float64 `json:"shipping_cost"`
	ShippingMethod  string   `json:"shipping_method"`
	Location        string   `json:"location"`
	PickupAvailable bool     `json:"pickup_available"`
	ReturnsAccepted bool     `json:"returns_accepted"`
	Dimensions      string   `json:"dimensions"`
	Weight          string   `json:"weight"`
	Material        string   `json:"material"`
	Color           string   `json:"color"`
	Brand           string   `json:"brand"`
	Model           string   `json:"model"`
	Year            *int     `json:"year"`
}

// ParseSearchItemToDBItem converts a SearchItem to a database Item model
func ParseSearchItemToDBItem(item SearchItem) (*db.GormItem, error) {
	// Parse ends_at timestamp
	endsAt, err := time.Parse(time.RFC3339, item.EndsAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ends_at: %w", err)
	}

	// Create database item
	dbItem := &db.GormItem{
		GoodwillID:     item.ID,
		Title:          item.Title,
		Seller:         item.Seller,
		CurrentPrice:   item.CurrentPrice,
		BuyNowPrice:    item.BuyNowPrice,
		URL:            item.URL,
		ImageURL:       item.ImageURL,
		EndsAt:         &endsAt,
		Status:         "active",
		Category:       item.Category,
		Subcategory:    item.Subcategory,
		Condition:      item.Condition,
		ShippingCost:   item.ShippingCost,
		ShippingMethod: item.ShippingMethod,
		WatchCount:     item.WatchCount,
		BidCount:       item.BidCount,
		ViewCount:      item.ViewCount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		FirstSeen:      time.Now(),
		LastSeen:       time.Now(),
	}

	return dbItem, nil
}

// ParseItemDetailsToDBItem converts ItemDetails to database Item model
func ParseItemDetailsToDBItem(details ItemDetails) (*db.GormItem, error) {
	// Parse ends_at timestamp
	endsAt, err := time.Parse(time.RFC3339, details.EndsAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ends_at: %w", err)
	}

	// Create database item
	dbItem := &db.GormItem{
		GoodwillID:      details.ID,
		Title:           details.Title,
		Seller:          details.Seller,
		CurrentPrice:    details.CurrentPrice,
		BuyNowPrice:     details.BuyNowPrice,
		URL:             details.URL,
		ImageURL:        details.ImageURL,
		EndsAt:          &endsAt,
		Status:          "active",
		Category:        details.Category,
		Subcategory:     details.Subcategory,
		Condition:       details.Condition,
		ShippingCost:    details.ShippingCost,
		ShippingMethod:  details.ShippingMethod,
		Location:        details.Location,
		PickupAvailable: details.PickupAvailable,
		ReturnsAccepted: details.ReturnsAccepted,
		Description:     details.Description, // Merged
		Dimensions:      details.Dimensions,  // Merged
		Weight:          details.Weight,      // Merged
		Material:        details.Material,    // Merged
		Color:           details.Color,       // Merged
		Brand:           details.Brand,       // Merged
		Model:           details.Model,       // Merged
		Year:            details.Year,        // Merged
		WatchCount:      0,
		BidCount:        0,
		ViewCount:       0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
	}

	return dbItem, nil
}

// createRequestWithAntiBot creates a request with anti-bot features
func (c *ShopGoodwillClient) createRequestWithAntiBot(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	// Create basic request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use user-agent rotation from anti-bot system if available
	if c.antiBotSystem != nil {
		userAgent, err := c.antiBotSystem.GetUserAgentWithRotation()
		if err != nil {
			log.Warnf("Failed to get rotated user agent, falling back to default: %v", err)
			// Fallback to default user agent
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		} else {
			req.Header.Set("User-Agent", userAgent.UserAgent)
		}
	} else {
		// Fallback to default user agent if no anti-bot system is configured
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	}

	// Set standard headers to mimic browser behavior
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1") // Do Not Track
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	return req, nil
}

// executeRequestWithAntiBot executes a request with anti-bot protection
func (c *ShopGoodwillClient) executeRequestWithAntiBot(ctx context.Context, req *http.Request, operationName string) (*http.Response, error) {
	// Execute with circuit breaker protection if available
	if c.circuitBreaker != nil {
		var resp *http.Response
		err := c.circuitBreaker.Execute(func() error {
			var err error
			resp, err = c.executeRequestWithRetry(ctx, req, operationName)
			return err
		})

		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// Fallback to direct execution with retry
	return c.executeRequestWithRetry(ctx, req, operationName)
}

// executeRequestWithRetry executes a request with retry logic
func (c *ShopGoodwillClient) executeRequestWithRetry(ctx context.Context, req *http.Request, operationName string) (*http.Response, error) {
	if c.retryManager != nil {
		var resp *http.Response
		err := c.retryManager.ExecuteWithRetry(ctx, operationName, func() error {
			var err error
			resp, err = c.executeRequestWithErrorHandling(req)
			return err
		})

		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// Fallback to simple execution
	return c.executeRequestWithErrorHandling(req)
}

// executeRequestWithErrorHandling executes a single request with proper error handling and body management
func (c *ShopGoodwillClient) executeRequestWithErrorHandling(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Don't close body here - let the caller handle it

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Limit the amount of data we read from the error body to prevent memory issues
		// with unexpectedly large responses (e.g., HTML pages)
		limitReader := io.LimitReader(resp.Body, 10240) // 10KB limit
		bodyBytes, _ := io.ReadAll(limitReader)

		// Ensure body is closed before returning error
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close response body for status %d: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
