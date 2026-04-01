package api

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
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

const (
	// aesKey is the fixed key used by ShopGoodwill for credential obfuscation
	aesKey = "0123456789123456"
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
	authURL := c.baseURL.JoinPath("SignIn", "Login")
	if authURL == nil {
		return fmt.Errorf("failed to construct auth URL")
	}

	// Encrypt credentials
	encUser, err := c.encryptCredentials(c.config.Username)
	if err != nil {
		return fmt.Errorf("failed to encrypt username: %w", err)
	}
	encPass, err := c.encryptCredentials(c.config.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Create request payload
	payload := map[string]interface{}{
		"userName":   encUser,
		"password":   encPass,
		"remember":   false,
		"appVersion": c.config.AppVersion,
		"browser":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	// Create request with anti-bot features to get necessary headers like User-Agent
	req, err := c.createRequestWithAntiBot(ctx, "POST", authURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set required content type header
	req.Header.Set("Content-Type", "application/json")

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
		Status      bool   `json:"status"`
		Message     string `json:"message"`
		AccessToken string `json:"accessToken"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Ensure body is closed
	if err := resp.Body.Close(); err != nil {
		log.Errorf("Failed to close auth response body: %v", err)
	}

	if !authResponse.Status {
		return fmt.Errorf("authentication failed: %s", authResponse.Message)
	}

	// Set authentication token and expiration (already thread-safe due to defer unlock)
	c.authToken = authResponse.AccessToken
	// New API doesn't specify expiration in simple login, set to 2 hours
	c.authExpired = time.Now().Add(2 * time.Hour)

	log.Infof("Successfully authenticated with ShopGoodwill API")

	return nil
}

// encryptCredentials encrypts a string using AES-CBC with the fixed ShopGoodwill key
func (c *ShopGoodwillClient) encryptCredentials(text string) (string, error) {
	key := []byte(aesKey)
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	padding := block.BlockSize() - (len(plaintext) % block.BlockSize())
	// Ensure padding is within byte range to satisfy gosec G115
	if padding < 0 || padding > 255 {
		return "", fmt.Errorf("invalid padding size: %d", padding)
	}
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext = append(plaintext, padtext...)

	ciphertext := make([]byte, len(plaintext))
	// Use the key as the IV as well (based on reverse engineering of the ShopGoodwill API obfuscation scheme)
	// #nosec G407
	mode := cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(ciphertext, plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Search executes a search query on ShopGoodwill
func (c *ShopGoodwillClient) Search(ctx context.Context, query string, params SearchParams) (*SearchResponse, error) {
	// Ensure we're authenticated
	if err := c.Authenticate(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Prepare search URL
	searchURL := c.baseURL.JoinPath("Search", "ItemListing")
	if searchURL == nil {
		return nil, fmt.Errorf("failed to construct search URL")
	}

	// Create search payload
	payload := map[string]interface{}{
		"searchText":                      query,
		"selectedCategoryIds":             params.Category,
		"selectedSellerIds":               params.Seller,
		"lowPrice":                        fmt.Sprintf("%.2f", params.MinPrice),
		"highPrice":                       fmt.Sprintf("%.2f", params.MaxPrice),
		"sortColumn":                      "1", // Default to ending soon
		"page":                            fmt.Sprintf("%d", params.Page),
		"pageSize":                        fmt.Sprintf("%d", params.PageSize),
		"sortDescending":                  "false",
		"searchPickupOnly":                "false",
		"searchNoPickupOnly":              "false",
		"searchOneCentShippingOnly":       "false",
		"searchDescriptions":              "false",
		"searchClosedAuctions":            "false",
		"searchCanadaShipping":            "false",
		"searchInternationalShippingOnly": "false",
		"searchUSOnlyShipping":            "false",
		"useBuyerPrefs":                   "true",
	}

	if params.SortBy == "price_asc" {
		payload["sortColumn"] = "2"
		payload["sortDescending"] = "false"
	} else if params.SortBy == "price_desc" {
		payload["sortColumn"] = "2"
		payload["sortDescending"] = "true"
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search payload: %w", err)
	}

	// Create request with anti-bot features
	req, err := c.createRequestWithAntiBot(ctx, "POST", searchURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
	itemURL := c.baseURL.JoinPath("ItemDetail", "GetItemDetailModelByItemId", itemID)
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
	SearchResults struct {
		Items            []SearchItem `json:"items"`
		TotalResultCount int          `json:"totalResultCount"`
		PageCount        int          `json:"pageCount"`
		CurrentPage      int          `json:"currentPage"`
	} `json:"searchResults"`
}

// SearchItem represents an item in search results
type SearchItem struct {
	ItemID       int      `json:"itemId"`
	Title        string   `json:"title"`
	SellerName   string   `json:"sellerName"`
	CurrentPrice float64  `json:"currentPrice"`
	BuyItNow     *float64 `json:"buyItNowPrice"`
	ImageName    string   `json:"imageName"`
	EndTime      string   `json:"endTime"`
	CategoryName string   `json:"categoryName"`
	Condition    string   `json:"condition"`
	WatchCount   int      `json:"watchCount"`
	BidCount     int      `json:"bids"`
}

// ItemDetails represents detailed item information
type ItemDetails struct {
	ItemID          int      `json:"itemId"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	SellerName      string   `json:"sellerName"`
	CurrentPrice    float64  `json:"currentPrice"`
	BuyItNowPrice   *float64 `json:"buyItNowPrice"`
	ImageName       string   `json:"imageName"`
	EndTime         string   `json:"endTime"`
	CategoryName    string   `json:"categoryName"`
	ConditionName   string   `json:"conditionName"`
	ShippingCost    *float64 `json:"shippingCost"`
	ShippingMethod  string   `json:"shippingMethod"`
	Location        string   `json:"location"`
	PickupAvailable bool     `json:"pickupAvailable"`
	ReturnsAccepted bool     `json:"returnsAccepted"`
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
	endsAt, err := time.Parse(time.RFC3339, item.EndTime)
	if err != nil {
		// Fallback for different format if needed, but the API seems to use RFC3339
		endsAt, err = time.Parse("2006-01-02T15:04:05Z", item.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse endTime %q: %w", item.EndTime, err)
		}
	}

	// Create database item
	dbItem := &db.GormItem{
		GoodwillID:   fmt.Sprintf("%d", item.ItemID),
		Title:        item.Title,
		Seller:       item.SellerName,
		CurrentPrice: item.CurrentPrice,
		BuyNowPrice:  item.BuyItNow,
		URL:          fmt.Sprintf("https://www.shopgoodwill.com/Item/%d", item.ItemID),
		ImageURL:     item.ImageName,
		EndsAt:       &endsAt,
		Status:       "active",
		Category:     item.CategoryName,
		Condition:    item.Condition,
		WatchCount:   item.WatchCount,
		BidCount:     item.BidCount,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
	}

	return dbItem, nil
}

// ParseItemDetailsToDBItem converts ItemDetails to database Item model
func ParseItemDetailsToDBItem(details ItemDetails) (*db.GormItem, error) {
	// Parse ends_at timestamp
	endsAt, err := time.Parse(time.RFC3339, details.EndTime)
	if err != nil {
		endsAt, err = time.Parse("2006-01-02T15:04:05Z", details.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse endTime %q: %w", details.EndTime, err)
		}
	}

	// Create database item
	dbItem := &db.GormItem{
		GoodwillID:      fmt.Sprintf("%d", details.ItemID),
		Title:           details.Title,
		Seller:          details.SellerName,
		CurrentPrice:    details.CurrentPrice,
		BuyNowPrice:     details.BuyItNowPrice,
		URL:             fmt.Sprintf("https://www.shopgoodwill.com/Item/%d", details.ItemID),
		ImageURL:        details.ImageName,
		EndsAt:          &endsAt,
		Status:          "active",
		Category:        details.CategoryName,
		Condition:       details.ConditionName,
		ShippingCost:    details.ShippingCost,
		ShippingMethod:  details.ShippingMethod,
		Location:        details.Location,
		PickupAvailable: details.PickupAvailable,
		ReturnsAccepted: details.ReturnsAccepted,
		Description:     details.Description,
		Dimensions:      details.Dimensions,
		Weight:          details.Weight,
		Material:        details.Material,
		Color:           details.Color,
		Brand:           details.Brand,
		Model:           details.Model,
		Year:            details.Year,
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

	// Validate request is targeting allowed host to prevent SSRF attacks
	if err := c.validateRequestHost(req); err != nil {
		return nil, err
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
	// Validate request is targeting allowed host to prevent SSRF attacks
	if err := c.validateRequestHost(req); err != nil {
		return nil, err
	}

	// #nosec G704 -- Host is validated by validateRequestHost(req) above
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

// validateRequestHost ensures the request is targeting an allowed host to prevent SSRF attacks
func (c *ShopGoodwillClient) validateRequestHost(req *http.Request) error {
	if req.URL == nil {
		return fmt.Errorf("request URL is nil")
	}

	// Get the request host (handle both host:port and just host)
	reqHost := req.URL.Hostname()
	if reqHost == "" {
		return fmt.Errorf("request host is empty")
	}

	// Get the allowed host from baseURL
	allowedHost := c.baseURL.Hostname()
	if allowedHost == "" {
		return fmt.Errorf("base URL host is not configured")
	}

	// Validate the request is going to the allowed host
	if reqHost != allowedHost {
		return fmt.Errorf("request host %q is not allowed (expected %q): potential SSRF attack", reqHost, allowedHost)
	}

	return nil
}
