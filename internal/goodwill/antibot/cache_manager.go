package antibot

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// CacheManager handles caching for frequent operations
type CacheManager struct {
	cache           map[string]CacheEntry
	mu              sync.RWMutex
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// CacheEntry represents a cached value
type CacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
}

// NewCacheManager creates a new cache manager
func NewCacheManager(defaultTTL, cleanupInterval time.Duration) *CacheManager {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute // default
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 10 * time.Minute // default
	}

	cm := &CacheManager{
		cache:           make(map[string]CacheEntry),
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go cm.startCleanup()

	return cm
}

// Get gets a value from cache
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mu.RLock()
	entry, exists := cm.cache[key]
	if !exists {
		cm.mu.RUnlock()
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		cm.mu.RUnlock()
		return nil, false
	}

	// Update access time without upgrading lock - simpler approach
	// This reduces contention by avoiding lock upgrades
	entry.AccessedAt = time.Now()

	cm.mu.RUnlock()
	return entry.Value, true
}

// Set sets a value in cache
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	} else {
		expiresAt = time.Now().Add(cm.defaultTTL)
	}

	cm.cache[key] = CacheEntry{
		Value:      value,
		ExpiresAt:  expiresAt,
		AccessedAt: time.Now(),
	}
}

// Delete deletes a value from cache
func (cm *CacheManager) Delete(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.cache, key)
}

// Clear clears the entire cache
func (cm *CacheManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cache = make(map[string]CacheEntry)
}

// GetKeys gets all cache keys
func (cm *CacheManager) GetKeys() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	keys := make([]string, 0, len(cm.cache))
	for key := range cm.cache {
		keys = append(keys, key)
	}
	return keys
}

// GetSize gets current cache size
func (cm *CacheManager) GetSize() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.cache)
}

// startCleanup starts the cleanup goroutine
func (cm *CacheManager) startCleanup() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.cleanupExpired()
		case <-cm.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes expired entries from cache
func (cm *CacheManager) cleanupExpired() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for key, entry := range cm.cache {
		if now.After(entry.ExpiresAt) {
			delete(cm.cache, key)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		log.Debugf("Cache cleanup removed %d expired entries", expiredCount)
	}
}

// Stop stops the cleanup goroutine
func (cm *CacheManager) Stop() {
	close(cm.stopCleanup)
}

// GetWithFallback gets value from cache or uses fallback function
func (cm *CacheManager) GetWithFallback(key string, ttl time.Duration, fallback func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, ok := cm.Get(key); ok {
		return value, nil
	}

	// Use fallback function
	value, err := fallback()
	if err != nil {
		return nil, err
	}

	// Cache the result
	cm.Set(key, value, ttl)

	return value, nil
}
