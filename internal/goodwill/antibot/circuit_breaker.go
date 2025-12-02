package antibot

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name             string
	state            CircuitBreakerState
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	resetTimeout     time.Duration
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	lastStateChange  time.Time
	probeInProgress  bool
	mu               sync.Mutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold, successThreshold int, timeout, resetTimeout time.Duration) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 3
	}
	if successThreshold <= 0 {
		successThreshold = 2
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if resetTimeout <= 0 {
		resetTimeout = 5 * time.Minute
	}

	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		resetTimeout:     resetTimeout,
		failureCount:     0,
		successCount:     0,
		lastStateChange:  time.Now(),
		probeInProgress:  false,
	}
}

// AllowRequest checks if a request is currently allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	return cb.checkAndTransitionState() == nil
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check state and handle transitions without lock if possible, but state check needs lock
	if err := cb.checkAndTransitionState(); err != nil {
		return err
	}

	// Double-check Half-Open state logic for single probe
	cb.mu.Lock()
	if cb.state == StateHalfOpen {
		if cb.probeInProgress {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker half-open: probe already in progress")
		}
		cb.probeInProgress = true
	}
	cb.mu.Unlock()

	// Execute with panic recovery
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic in circuit breaker execution: %v", r)
				// Panic is treated as a failure
			}
		}()
		err = fn()
	}()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Reset probe flag if we were in half-open (or if state changed, doesn't hurt to reset)
	if cb.probeInProgress {
		cb.probeInProgress = false
	}

	if err != nil {
		cb.handleFailure()
		return err
	}

	cb.handleSuccess()
	return nil
}

// checkAndTransitionState checks circuit breaker state and handles transitions
func (cb *CircuitBreaker) checkAndTransitionState() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit is open
	if cb.state == StateOpen {
		// Check if timeout has expired and we should try half-open
		if time.Since(cb.lastStateChange) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
			log.Infof("Circuit breaker '%s' transitioning to half-open state", cb.name)
		} else {
			return fmt.Errorf("circuit breaker '%s' is open (last failure: %v ago)", cb.name, time.Since(cb.lastFailureTime))
		}
	}

	return nil
}

// handleFailure handles a failure
func (cb *CircuitBreaker) handleFailure() {
	cb.failureCount++
	cb.successCount = 0
	cb.lastFailureTime = time.Now()

	log.Warnf("Circuit breaker '%s' recorded failure (%d/%d)", cb.name, cb.failureCount, cb.failureThreshold)

	// Transition to open state if threshold reached
	if cb.state != StateOpen && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
		log.Warnf("Circuit breaker '%s' transitioned to OPEN state", cb.name)
	}
}

// handleSuccess handles a success
func (cb *CircuitBreaker) handleSuccess() {
	cb.successCount++
	cb.failureCount = 0

	// If in half-open state, transition back to closed after success threshold
	if cb.state == StateHalfOpen {
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.lastStateChange = time.Now()
			log.Infof("Circuit breaker '%s' transitioned to CLOSED state (recovered)", cb.name)
		}
	}
}

// GetState gets current state (pure getter - no state transitions)
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastStateChange = time.Now()
	cb.probeInProgress = false // Reset probe status on manual reset
	log.Infof("Circuit breaker '%s' manually reset to CLOSED state", cb.name)
}

// IsAvailable checks if circuit breaker is available
func (cb *CircuitBreaker) IsAvailable() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// This method only checks the current state without triggering transitions
	// or managing probeInProgress. For execution, use Execute().
	return cb.state != StateOpen
}

// GetFailureCount gets current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failureCount
}

// GetSuccessCount gets current success count
func (cb *CircuitBreaker) GetSuccessCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.successCount
}
