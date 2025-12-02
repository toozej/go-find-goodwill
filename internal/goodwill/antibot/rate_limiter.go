package antibot

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate       int        // requests per minute
	burst      int        // burst limit
	tokens     int        // current tokens
	lastRefill time.Time  // last refill time
	mu         sync.Mutex // mutex for thread safety
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int) *RateLimiter {
	if rate <= 0 {
		rate = 60 // default: 60 requests per minute
	}
	if burst <= 0 {
		burst = rate // default: burst equals rate
	}

	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     burst,
		lastRefill: time.Now(),
	}
}

// AllowRequest checks if a request should be allowed
func (rl *RateLimiter) AllowRequest() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time passed
	rl.refillTokens()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// refillTokens refills tokens based on time passed
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	duration := now.Sub(rl.lastRefill)
	seconds := duration.Seconds()

	// Calculate tokens to add (rate tokens per minute)
	tokensToAdd := int(seconds * float64(rl.rate) / 60.0)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.burst {
			rl.tokens = rl.burst
		}
		rl.lastRefill = now
	}
}

// GetAvailableTokens returns available tokens
func (rl *RateLimiter) GetAvailableTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()
	return rl.tokens
}
