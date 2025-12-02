package antibot

import (
	"slices"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// TimingManager handles adaptive timing with human-like patterns
type TimingManager struct {
	config           config.TimingConfig
	humanPatterns    []time.Duration
	currentPattern   int
	adaptiveVariance float64
	mu               sync.Mutex
}

// NewTimingManager creates a new timing manager
func NewTimingManager(cfg config.TimingConfig) *TimingManager {
	tm := &TimingManager{
		config:           cfg,
		humanPatterns:    generateHumanPatterns(cfg),
		adaptiveVariance: 1.0,
	}

	return tm
}

// GetAdaptiveDelay gets adaptive delay with human-like timing
func (tm *TimingManager) GetAdaptiveDelay() time.Duration {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Get base pattern
	baseDelay := tm.humanPatterns[tm.currentPattern]

	// Apply adaptive variance
	adaptiveDelay := time.Duration(float64(baseDelay) * tm.adaptiveVariance)

	// Add human-like jitter
	jitter := tm.getHumanJitter()
	finalDelay := adaptiveDelay + jitter

	// Move to next pattern
	tm.currentPattern = (tm.currentPattern + 1) % len(tm.humanPatterns)

	return finalDelay
}

// AdjustVariance adjusts the timing variance based on success/failure
func (tm *TimingManager) AdjustVariance(success bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// Update adaptive variance based on success/failure
	if success {
		// Gradually decrease variance when successful (normalize)
		// We want to return towards 1.0 or lower
		tm.adaptiveVariance = slices.Max([]float64{tm.adaptiveVariance - 0.1, 1.0})
	} else {
		// Increase variance when failing (slow down / match human unpredictability)
		// We want to increase delay significantly
		tm.adaptiveVariance = slices.Min([]float64{tm.adaptiveVariance + 0.5, 5.0})
	}
}

// getHumanJitter generates human-like random jitter
func (tm *TimingManager) getHumanJitter() time.Duration {
	// Generate jitter within configured range
	maxJitter := int(tm.config.MaxJitter.Seconds() - tm.config.MinJitter.Seconds())
	if maxJitter <= 0 {
		return tm.config.MinJitter
	}

	jitterSeconds, err := getRandomInt(maxJitter)
	if err != nil {
		log.Errorf("Failed to generate human jitter: %v", err)
		return tm.config.MinJitter
	}

	return time.Duration(jitterSeconds+int(tm.config.MinJitter.Seconds())) * time.Second
}

// generateHumanPatterns generates realistic human timing patterns
func generateHumanPatterns(cfg config.TimingConfig) []time.Duration {
	// Base patterns with some variation
	patterns := []time.Duration{
		cfg.BaseInterval,
		cfg.BaseInterval + cfg.MinJitter,
		cfg.BaseInterval - cfg.MinJitter/2,
		cfg.BaseInterval + cfg.MaxJitter/2,
		cfg.BaseInterval - cfg.MaxJitter/3,
	}

	// Add some random human-like variations
	for i := 0; i < 3; i++ {
		variation, err := getRandomInt(int(cfg.MaxJitter.Seconds()))
		if err != nil {
			log.Errorf("Failed to generate human pattern variation: %v", err)
			variation = int(cfg.MaxJitter.Seconds()) / 2
		}
		patterns = append(patterns, cfg.BaseInterval+time.Duration(variation)*time.Second)
	}

	return patterns
}
