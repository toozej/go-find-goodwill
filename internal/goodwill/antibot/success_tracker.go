package antibot

import (
	"sync"
)

// SuccessTracker tracks success rates for user agents
type SuccessTracker struct {
	successRates  map[int]float64 // agentID -> success rate
	attemptCounts map[int]int     // agentID -> total attempts
	mu            sync.Mutex
}

// NewSuccessTracker creates a new success tracker
func NewSuccessTracker() *SuccessTracker {
	return &SuccessTracker{
		successRates:  make(map[int]float64),
		attemptCounts: make(map[int]int),
	}
}

// UpdateSuccess updates success tracking for a user agent
func (st *SuccessTracker) UpdateSuccess(agentID int, success bool) {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Initialize if not exists
	if _, exists := st.successRates[agentID]; !exists {
		st.successRates[agentID] = 1.0 // Start with perfect score
		st.attemptCounts[agentID] = 0
	}

	// Update attempt count
	st.attemptCounts[agentID]++

	// Calculate new success rate using exponential moving average
	currentRate := st.successRates[agentID]

	// Proper exponential moving average formula: alpha * current + (1 - alpha) * previous
	// Using a smoothing factor (alpha) of 0.2 to give more weight to recent results
	alpha := 0.2
	var newValue float64
	if success {
		newValue = 1.0
	} else {
		newValue = 0.0
	}

	// Apply EMA formula
	newRate := alpha*newValue + (1-alpha)*currentRate
	st.successRates[agentID] = newRate
}

// GetSuccessRates gets current success rates
func (st *SuccessTracker) GetSuccessRates() map[int]float64 {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Return a copy of the map
	rates := make(map[int]float64)
	for k, v := range st.successRates {
		rates[k] = v
	}
	return rates
}

// GetAgentSuccessRate gets success rate for specific agent with safe access
func (st *SuccessTracker) GetAgentSuccessRate(agentID int) (float64, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()

	rate, exists := st.successRates[agentID]
	return rate, exists
}

// ResetAgentStats resets statistics for a specific agent
func (st *SuccessTracker) ResetAgentStats(agentID int) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.successRates[agentID] = 1.0 // Reset to perfect score
	st.attemptCounts[agentID] = 0
}

// CleanupOldAgents removes agents that haven't been used recently
func (st *SuccessTracker) CleanupOldAgents() {
	st.mu.Lock()
	defer st.mu.Unlock()

	// In a real implementation, we would track last usage time
	// For now, this is a placeholder for future implementation
}

// GetBestAgentID gets the agent ID with highest success rate
func (st *SuccessTracker) GetBestAgentID() int {
	st.mu.Lock()
	defer st.mu.Unlock()

	bestID := -1
	bestRate := -1.0

	for id, rate := range st.successRates {
		if rate > bestRate {
			bestRate = rate
			bestID = id
		}
	}

	return bestID
}
