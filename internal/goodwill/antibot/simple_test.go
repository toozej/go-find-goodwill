package antibot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// TestSimpleAntiBotComponents tests basic anti-bot components
func TestSimpleAntiBotComponents(t *testing.T) {
	t.Run("RateLimiter", func(t *testing.T) {
		limiter := NewRateLimiter(60, 10)
		assert.NotNil(t, limiter)

		// Should allow requests initially
		assert.True(t, limiter.AllowRequest())

		// Should have tokens available
		assert.True(t, limiter.GetAvailableTokens() > 0)
	})

	t.Run("TimingManager", func(t *testing.T) {
		cfg := config.TimingConfig{
			BaseInterval: 15 * time.Minute,
			MinJitter:    2 * time.Minute,
			MaxJitter:    5 * time.Minute,
		}

		tm := NewTimingManager(cfg)
		assert.NotNil(t, tm)

		// Should return reasonable delay
		delay := tm.GetAdaptiveDelay()
		assert.True(t, delay > 0)
		assert.True(t, delay <= cfg.BaseInterval+cfg.MaxJitter)
	})

	t.Run("SuccessTracker", func(t *testing.T) {
		st := NewSuccessTracker()
		assert.NotNil(t, st)

		// Test success tracking
		st.UpdateSuccess(1, true)
		st.UpdateSuccess(1, true)
		st.UpdateSuccess(1, false)

		// Should have some success rate
		rates := st.GetSuccessRates()
		assert.True(t, len(rates) > 0)
	})

	t.Run("RetryManager", func(t *testing.T) {
		rm := NewRetryManager(3, 1*time.Second, 10*time.Second)
		assert.NotNil(t, rm)

		// Should allow retry initially
		assert.True(t, rm.ShouldRetry())

		// Test delay calculation
		delay := rm.GetRetryDelay()
		assert.True(t, delay >= 1*time.Second)
		assert.True(t, delay <= 10*time.Second)
	})

	t.Run("CircuitBreaker", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 1, 1*time.Second, 5*time.Second)
		assert.NotNil(t, cb)

		// Should be closed initially
		assert.Equal(t, StateClosed, cb.GetState())

		// Test state transitions
		cb.handleFailure()
		cb.handleFailure()
		assert.Equal(t, StateOpen, cb.GetState())
	})

	t.Run("CacheManager", func(t *testing.T) {
		cm := NewCacheManager(5*time.Minute, 1*time.Minute)
		assert.NotNil(t, cm)

		// Test basic operations
		cm.Set("test", "value", 1*time.Minute)
		value, ok := cm.Get("test")
		assert.True(t, ok)
		assert.Equal(t, "value", value)

		// Cleanup
		cm.Stop()
	})

	t.Run("AntiBotSystemShutdown", func(t *testing.T) {
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

		// Test shutdown doesn't panic
		antiBotSystem.Shutdown()
	})
}
