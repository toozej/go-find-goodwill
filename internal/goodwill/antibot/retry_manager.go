package antibot

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// RetryManager handles retry logic with exponential backoff
type RetryManager struct {
	maxRetries    int
	baseDelay     time.Duration
	maxDelay      time.Duration
	jitterFactor  float64
	retryCount    int
	lastRetryTime time.Time
	mu            sync.Mutex
}

// NewRetryManager creates a new retry manager
func NewRetryManager(maxRetries int, baseDelay, maxDelay time.Duration) *RetryManager {
	if maxRetries <= 0 {
		maxRetries = 3 // default
	}
	if baseDelay <= 0 {
		baseDelay = 1 * time.Second // default
	}
	if maxDelay <= 0 || maxDelay < baseDelay {
		maxDelay = 30 * time.Second // default
	}

	return &RetryManager{
		maxRetries:    maxRetries,
		baseDelay:     baseDelay,
		maxDelay:      maxDelay,
		jitterFactor:  0.2, // 20% jitter
		retryCount:    0,
		lastRetryTime: time.Now(),
	}
}

// ShouldRetry determines if a retry should be attempted
func (rm *RetryManager) ShouldRetry() bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.retryCount < rm.maxRetries
}

// GetRetryDelay gets the delay before next retry with exponential backoff
func (rm *RetryManager) GetRetryDelay() time.Duration {
	// Calculate exponential backoff: baseDelay * 2^retryCount
	backoff := rm.baseDelay * time.Duration(math.Pow(2, float64(rm.retryCount)))

	// Apply jitter to avoid thundering herd problem
	// Generate random float between -1 and 1 using crypto/rand
	randomFloat, err := getRandomFloat64()
	if err != nil {
		log.Errorf("Failed to generate random float for jitter: %v", err)
		randomFloat = 0.0 // Fallback to no jitter
	}

	jitter := time.Duration(float64(backoff) * rm.jitterFactor * (randomFloat*2 - 1))
	delay := backoff + jitter

	// Ensure delay doesn't exceed max delay
	if delay > rm.maxDelay {
		delay = rm.maxDelay
	}

	// Ensure minimum delay
	if delay < rm.baseDelay {
		delay = rm.baseDelay
	}

	return delay
}

// RecordRetry records a retry attempt
func (rm *RetryManager) RecordRetry() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.retryCount++
	rm.lastRetryTime = time.Now()
}

// Reset resets the retry counter
func (rm *RetryManager) Reset() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.retryCount = 0
	rm.lastRetryTime = time.Now()
}

// GetRetryCount gets current retry count
func (rm *RetryManager) GetRetryCount() int {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.retryCount
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operationName string, fn func() error) error {
	rm.Reset()

	for rm.ShouldRetry() {
		err := fn()
		if err == nil {
			return nil // Success
		}

		// Log the error
		log.Errorf("Operation '%s' failed (attempt %d/%d): %v",
			operationName, rm.retryCount+1, rm.maxRetries, err)

		if !rm.ShouldRetry() {
			break
		}

		// Wait for retry delay
		delay := rm.GetRetryDelay()
		log.Infof("Retrying operation '%s' in %v (attempt %d/%d)",
			operationName, delay, rm.retryCount+1, rm.maxRetries)

		select {
		case <-time.After(delay):
			rm.RecordRetry()
			continue
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}
	}

	return fmt.Errorf("operation '%s' failed after %d attempts", operationName, rm.retryCount)
}
