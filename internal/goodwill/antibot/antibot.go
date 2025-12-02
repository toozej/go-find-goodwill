package antibot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// AntiBotSystem handles anti-bot detection and prevention
type AntiBotSystem struct {
	config         *config.AntiBotConfig
	repository     db.Repository
	userAgentCache []db.GormUserAgent
	cacheMutex     sync.Mutex
	rateLimiter    *RateLimiter
	timingManager  *TimingManager
	successTracker *SuccessTracker
	shutdownChan   chan struct{}
}

// NewAntiBotSystem creates a new anti-bot system
func NewAntiBotSystem(cfg *config.AntiBotConfig, repo db.Repository) (*AntiBotSystem, error) {
	if cfg == nil {
		return nil, fmt.Errorf("anti-bot config cannot be nil")
	}
	if repo == nil {
		return nil, fmt.Errorf("repository cannot be nil")
	}

	// Initialize components
	rateLimiter := NewRateLimiter(cfg.Throttling.RequestsPerMinute, cfg.Throttling.BurstLimit)
	timingManager := NewTimingManager(cfg.Timing)
	successTracker := NewSuccessTracker()

	system := &AntiBotSystem{
		config:         cfg,
		repository:     repo,
		rateLimiter:    rateLimiter,
		timingManager:  timingManager,
		successTracker: successTracker,
		shutdownChan:   make(chan struct{}),
	}

	// Initialize user agent cache
	go system.cacheUserAgents()

	return system, nil
}

// Shutdown stops the anti-bot system and cleans up resources
func (a *AntiBotSystem) Shutdown() {
	close(a.shutdownChan)
}

// GetUserAgentWithRotation gets a user agent with rotation logic
func (a *AntiBotSystem) GetUserAgentWithRotation() (*db.GormUserAgent, error) {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	// If we have cached user agents, use rotation logic
	if len(a.userAgentCache) > 0 {
		// Find user agent with best success rate
		bestAgent := a.findBestUserAgent()
		if bestAgent != nil {
			return bestAgent, nil
		}
	}

	// Fallback to random user agent from database
	return a.repository.GetRandomUserAgent(context.Background())
}

// UpdateUserAgentSuccess updates user agent success tracking
func (a *AntiBotSystem) UpdateUserAgentSuccess(agentID int, success bool) {
	a.successTracker.UpdateSuccess(agentID, success)
}

// CheckRateLimit checks if request should be allowed based on rate limiting
func (a *AntiBotSystem) CheckRateLimit() bool {
	return a.rateLimiter.AllowRequest()
}

// GetAdaptiveDelay gets adaptive delay for human-like timing
func (a *AntiBotSystem) GetAdaptiveDelay() time.Duration {
	return a.timingManager.GetAdaptiveDelay()
}

// AnalyzeResponse analyzes API response for blocking detection
func (a *AntiBotSystem) AnalyzeResponse(statusCode int, responseBody string) bool {
	// Check for common blocking patterns
	blockingPatterns := []string{
		"access denied",
		"bot detected",
		"automated request",
		"403 forbidden",
		"429 too many requests",
		"captcha",
		"verification required",
		"cloudflare",
		"security check",
	}

	responseLower := strings.ToLower(responseBody)
	for _, pattern := range blockingPatterns {
		if strings.Contains(responseLower, pattern) {
			return true
		}
	}

	// Check status codes that indicate blocking
	if statusCode == 403 || statusCode == 429 || statusCode == 401 {
		return true
	}

	return false
}

// cacheUserAgents caches user agents from database
func (a *AntiBotSystem) cacheUserAgents() {
	ticker := time.NewTicker(a.config.UserAgent.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.refreshUserAgentCache()
		case <-a.shutdownChan:
			return
		}
	}
}

// refreshUserAgentCache refreshes the user agent cache
func (a *AntiBotSystem) refreshUserAgentCache() {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	// Fetch active user agents from database
	userAgents, err := a.repository.GetActiveUserAgents(context.Background())
	if err != nil {
		log.Errorf("Failed to fetch user agents from database: %v", err)
		// Fallback to default user agents if database fetch fails
		a.userAgentCache = []db.GormUserAgent{
			{ID: 1, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", UsageCount: 0, IsActive: true},
			{ID: 2, UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", UsageCount: 0, IsActive: true},
			{ID: 3, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0", UsageCount: 0, IsActive: true},
			{ID: 4, UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15", UsageCount: 0, IsActive: true},
			{ID: 5, UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", UsageCount: 0, IsActive: true},
		}
	} else {
		a.userAgentCache = userAgents
	}

	log.Infof("Refreshed user agent cache with %d agents", len(a.userAgentCache))
}

// findBestUserAgent finds the best user agent based on success rate
func (a *AntiBotSystem) findBestUserAgent() *db.GormUserAgent {
	// Get success rates from tracker
	successRates := a.successTracker.GetSuccessRates()

	// Find agent with highest success rate that meets minimum threshold
	var bestAgent *db.GormUserAgent
	bestRate := -1.0

	for _, agent := range a.userAgentCache {
		agentIDInt := agent.ID
		rate := successRates[agentIDInt]
		if rate >= a.config.UserAgent.MinSuccessRate && rate > bestRate {
			bestRate = rate
			bestAgent = &agent
		}
	}

	// If no agent meets threshold, return random one
	if bestAgent == nil && len(a.userAgentCache) > 0 {
		randomIndex, err := getRandomInt(len(a.userAgentCache))
		if err != nil {
			log.Errorf("Failed to generate random index for user agent: %v", err)
			return &a.userAgentCache[0] // Fallback to first agent
		}
		return &a.userAgentCache[randomIndex]
	}

	return bestAgent
}
