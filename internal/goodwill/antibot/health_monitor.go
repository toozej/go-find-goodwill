package antibot

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// HealthMonitor handles health checks and monitoring
type HealthMonitor struct {
	components    map[string]HealthStatus
	mu            sync.RWMutex
	checkInterval time.Duration
	stopChan      chan struct{}
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status       string
	LastCheck    time.Time
	ErrorMessage string
	Details      map[string]interface{}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval time.Duration) *HealthMonitor {
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second // default
	}

	hm := &HealthMonitor{
		components:    make(map[string]HealthStatus),
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}

	// Start monitoring
	go hm.startMonitoring()

	return hm
}

// RegisterComponent registers a component for health monitoring
func (hm *HealthMonitor) RegisterComponent(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.components[name] = HealthStatus{
		Status:    "unknown",
		LastCheck: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// UpdateComponentStatus updates a component's health status
func (hm *HealthMonitor) UpdateComponentStatus(name, status, errorMessage string, details map[string]interface{}) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if component, exists := hm.components[name]; exists {
		component.Status = status
		component.LastCheck = time.Now()
		component.ErrorMessage = errorMessage
		component.Details = details
		hm.components[name] = component
	}
}

// GetComponentStatus gets a component's health status
func (hm *HealthMonitor) GetComponentStatus(name string) (HealthStatus, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	status, exists := hm.components[name]
	return status, exists
}

// GetAllStatuses gets all component statuses
func (hm *HealthMonitor) GetAllStatuses() map[string]HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Return a copy
	statuses := make(map[string]HealthStatus)
	for k, v := range hm.components {
		statuses[k] = v
	}
	return statuses
}

// GetOverallHealth gets overall system health
func (hm *HealthMonitor) GetOverallHealth() string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Check if any critical components are unhealthy
	for _, status := range hm.components {
		if status.Status == "critical" || status.Status == "unhealthy" {
			return "unhealthy"
		}
	}

	// Check if any components are degraded
	for _, status := range hm.components {
		if status.Status == "degraded" {
			return "degraded"
		}
	}

	// Check if all components are healthy
	allHealthy := true
	for _, status := range hm.components {
		if status.Status != "healthy" {
			allHealthy = false
			break
		}
	}

	if allHealthy {
		return "healthy"
	}

	return "unknown"
}

// startMonitoring starts the monitoring goroutine
func (hm *HealthMonitor) startMonitoring() {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hm.checkComponentHealth()
		case <-hm.stopChan:
			return
		}
	}
}

// checkComponentHealth checks health of all components
func (hm *HealthMonitor) checkComponentHealth() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for name := range hm.components {
		// In a real implementation, we would have specific health checks
		// For now, we'll just update the last check time
		component := hm.components[name]
		component.LastCheck = time.Now()

		// Simulate some health degradation over time
		if time.Since(component.LastCheck) > hm.checkInterval*2 {
			component.Status = "degraded"
			component.ErrorMessage = "Component not responding"
		}

		hm.components[name] = component
	}

	log.Debug("Completed health check cycle")
}

// Stop stops the health monitor
func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
}

// LogHealthEvent logs a health-related event
func (hm *HealthMonitor) LogHealthEvent(component, eventType, message string) {
	_ = map[string]interface{}{
		"component":  component,
		"event_type": eventType,
		"message":    message,
		"timestamp":  time.Now(),
	}

	log.Infof("Health event: %s - %s: %s", component, eventType, message)
}

// AddHealthCheck adds a custom health check function
func (hm *HealthMonitor) AddHealthCheck(name string, checkFunc func() HealthStatus) {
	// In a real implementation, we would store and execute custom health checks
	// For now, this is a placeholder
	hm.RegisterComponent(name)
}
