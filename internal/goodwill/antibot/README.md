# Anti-Bot System Documentation

## Overview

The anti-bot system provides comprehensive protection against bot detection and rate limiting for the go-find-goodwill application. It includes user agent rotation, request throttling, adaptive timing, and response analysis.

## Components

### 1. AntiBotSystem (Main System)

The main anti-bot system that coordinates all anti-bot measures:

- **User Agent Rotation**: Automatically rotates user agents based on success rates
- **Rate Limiting**: Implements token bucket algorithm for request throttling
- **Adaptive Timing**: Human-like timing patterns to avoid detection
- **Response Analysis**: Detects blocking patterns in API responses

### 2. RateLimiter

Token bucket implementation for request throttling:

- Configurable requests per minute and burst limits
- Thread-safe implementation
- Automatic token refill based on time

### 3. TimingManager

Human-like timing patterns:

- Base interval with configurable jitter
- Adaptive variance based on success/failure
- Multiple timing patterns for realism

### 4. SuccessTracker

User agent success rate tracking:

- Tracks success rates for each user agent
- Exponential moving average for smooth metrics
- Automatic selection of best-performing agents

### 5. RetryManager

Exponential backoff retry logic:

- Configurable max retries and delay parameters
- Jitter to avoid thundering herd problem
- Context-aware execution with cancellation

### 6. CircuitBreaker

Circuit breaker pattern for API failures:

- Three states: Closed, Open, Half-Open
- Configurable failure and success thresholds
- Automatic state transitions with timeouts

### 7. CacheManager

Caching for frequent operations:

- Time-based expiration (TTL)
- Automatic cleanup of expired entries
- Fallback functions for cache misses

### 8. HealthMonitor

System health monitoring:

- Component-based health tracking
- Overall system health status
- Regular health check cycles

## Usage Examples

### Basic Anti-Bot System Setup

```go
// Create configuration
cfg := &config.AntiBotConfig{
    UserAgent: config.UserAgentConfig{
        RotationEnabled: true,
        RotationInterval: 1 * time.Hour,
        RequestsPerUA: 20,
        MinSuccessRate: 0.8,
    },
    Timing: config.TimingConfig{
        BaseInterval: 15 * time.Minute,
        MinJitter: 2 * time.Minute,
        MaxJitter: 5 * time.Minute,
    },
    Throttling: config.ThrottlingConfig{
        RequestsPerMinute: 60,
        BurstLimit: 10,
    },
}

// Create anti-bot system
antiBotSystem, err := antibot.NewAntiBotSystem(cfg, repository)
if err != nil {
    log.Fatalf("Failed to create anti-bot system: %v", err)
}
```

### User Agent Rotation

```go
// Get user agent with rotation
agent, err := antiBotSystem.GetUserAgentWithRotation()
if err != nil {
    log.Errorf("Failed to get user agent: %v", err)
    return
}

// Use the user agent in HTTP request
req.Header.Set("User-Agent", agent.UserAgent)

// Update success tracking
antiBotSystem.UpdateUserAgentSuccess(agent.ID, true)
```

### Rate Limiting

```go
// Check rate limit before making request
if !antiBotSystem.CheckRateLimit() {
    log.Warn("Rate limit exceeded, waiting...")
    time.Sleep(1 * time.Second)
    return
}
```

### Adaptive Timing

```go
// Get adaptive delay for human-like timing
delay := antiBotSystem.GetAdaptiveDelay()
log.Infof("Waiting %v before next request", delay)
time.Sleep(delay)
```

### Response Analysis

```go
// Analyze API response for blocking
resp, err := httpClient.Do(req)
if err != nil {
    log.Errorf("Request failed: %v", err)
    return
}

body, _ := io.ReadAll(resp.Body)
isBlocked := antiBotSystem.AnalyzeResponse(resp.StatusCode, string(body))
if isBlocked {
    log.Warn("Detected blocking response, adjusting strategy")
    // Implement fallback strategy
}
```

### Retry Logic

```go
// Create retry manager
retryManager := antibot.NewRetryManager(3, 1*time.Second, 30*time.Second)

// Execute with retry
err := retryManager.ExecuteWithRetry(context.Background(), "api-request", func() error {
    // Make API request
    resp, err := httpClient.Do(req)
    if err != nil {
        return err
    }
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("API returned status %d", resp.StatusCode)
    }
    return nil
})

if err != nil {
    log.Errorf("API request failed after retries: %v", err)
}
```

### Circuit Breaker

```go
// Create circuit breaker
cb := antibot.NewCircuitBreaker("api-cb", 3, 2, 30*time.Second, 5*time.Minute)

// Execute with circuit breaker protection
err := cb.Execute(func() error {
    // Make API request
    resp, err := httpClient.Do(req)
    if err != nil {
        return err
    }
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("API returned status %d", resp.StatusCode)
    }
    return nil
})

if err != nil {
    if cb.GetState() == antibot.StateOpen {
        log.Warn("Circuit breaker is open, using fallback")
        // Implement fallback logic
    } else {
        log.Errorf("API request failed: %v", err)
    }
}
```

### Caching

```go
// Create cache manager
cacheManager := antibot.NewCacheManager(10*time.Minute, 5*time.Minute)
defer cacheManager.Stop()

// Get with fallback
result, err := cacheManager.GetWithFallback("search-results", 5*time.Minute, func() (interface{}, error) {
    // Fetch from API
    results, err := apiClient.Search(ctx, query)
    if err != nil {
        return nil, err
    }
    return results, nil
})

if err != nil {
    log.Errorf("Failed to get search results: %v", err)
    return
}

// Use cached result
searchResults := result.(*api.SearchResponse)
```

## Configuration

The anti-bot system is configured through the `AntiBotConfig` structure:

```yaml
antibot:
  user_agent:
    rotation_enabled: true
    rotation_interval: 1h
    requests_per_ua: 20
    min_success_rate: 0.8
  timing:
    base_interval: 15m
    min_jitter: 2m
    max_jitter: 5m
    human_like_variation: true
  throttling:
    requests_per_minute: 60
    burst_limit: 10
```

## Best Practices

1. **Gradual Rollout**: Start with conservative settings and gradually increase aggression
2. **Monitoring**: Implement comprehensive logging and monitoring
3. **Fallback Strategies**: Always have fallback mechanisms for when anti-bot measures are detected
4. **Regular Updates**: Keep user agent lists and patterns updated
5. **Testing**: Test anti-bot measures in staging before production

## Integration Points

The anti-bot system integrates with:

- **API Client**: User agent rotation and request timing
- **Search Manager**: Rate limiting and adaptive timing
- **Database**: Caching and batch operations
- **Logging**: Comprehensive error logging and monitoring