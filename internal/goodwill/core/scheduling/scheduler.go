package scheduling

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/api"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// queuedSearch represents a search in the execution queue with its timer
type queuedSearch struct {
	search db.GormSearch
	timer  *time.Timer
}

// Scheduler handles advanced scheduling with cron-like functionality and jitter
type Scheduler struct {
	config               *config.Config
	repo                 db.Repository
	apiClient            *api.ShopGoodwillClient
	searchQueue          chan db.GormSearch
	shutdownChan         chan struct{}
	workersWg            sync.WaitGroup
	mu                   sync.Mutex
	activeSearches       map[int]bool
	searchExecutionQueue []queuedSearch
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.Config, repo db.Repository, apiClient *api.ShopGoodwillClient) *Scheduler {
	return &Scheduler{
		config:               cfg,
		repo:                 repo,
		apiClient:            apiClient,
		searchQueue:          make(chan db.GormSearch, 100),
		shutdownChan:         make(chan struct{}),
		activeSearches:       make(map[int]bool),
		searchExecutionQueue: make([]queuedSearch, 0),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	log.Info("Starting advanced scheduler")

	// Start scheduler workers
	numWorkers := 3
	for i := 0; i < numWorkers; i++ {
		s.workersWg.Add(1)
		go s.schedulerWorker(i)
	}

	// Start main scheduling loop
	s.workersWg.Add(1)
	go s.schedulingLoop()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Info("Stopping scheduler")

	close(s.shutdownChan)
	s.workersWg.Wait()
	log.Info("Scheduler stopped")
}

// schedulingLoop runs the main scheduling logic
func (s *Scheduler) schedulingLoop() {
	defer s.workersWg.Done()

	// Calculate base interval with jitter
	baseInterval := time.Duration(s.config.Search.IntervalMinutes) * time.Minute
	ticker := time.NewTicker(baseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownChan:
			log.Info("Scheduling loop shutting down")
			return

		case <-ticker.C:
			s.executeScheduledSearches()
		}
	}
}

// executeScheduledSearches executes searches based on scheduling logic
func (s *Scheduler) executeScheduledSearches() {
	log.Info("Executing scheduled searches with advanced timing")

	// Get active searches
	searches, err := s.repo.GetActiveSearches(context.Background())
	if err != nil {
		log.Errorf("Failed to get active searches: %v", err)
		return
	}

	if len(searches) == 0 {
		log.Info("No active searches to execute")
		return
	}

	log.Infof("Found %d active searches to schedule", len(searches))

	// Schedule each search with appropriate timing
	for _, search := range searches {
		// Calculate timing with jitter
		executionTime := s.calculateExecutionTimeWithJitter()

		// Add to execution queue
		s.addToExecutionQueue(search, executionTime)
	}
}

// calculateExecutionTimeWithJitter calculates execution time with randomized jitter
func (s *Scheduler) calculateExecutionTimeWithJitter() time.Duration {
	// Base delay
	baseDelay := time.Duration(s.config.AntiBot.Timing.BaseInterval)

	// Add jitter (random variation)
	jitter, err := s.getRandomDuration(s.config.AntiBot.Timing.MaxJitter - s.config.AntiBot.Timing.MinJitter)
	if err != nil {
		log.Errorf("Failed to generate random jitter: %v", err)
		jitter = (s.config.AntiBot.Timing.MaxJitter - s.config.AntiBot.Timing.MinJitter) / 2
	}
	jitter += s.config.AntiBot.Timing.MinJitter

	// Apply human-like variation if enabled
	if s.config.AntiBot.Timing.HumanLikeVariation {
		// Add some additional randomness to make it more human-like
		humanVariation, err := s.getRandomDuration(30 * time.Second)
		if err != nil {
			log.Errorf("Failed to generate human variation: %v", err)
			humanVariation = 15 * time.Second
		}
		return baseDelay + jitter + humanVariation
	}

	return baseDelay + jitter
}

// getRandomDuration generates a random duration up to the specified maximum
func (s *Scheduler) getRandomDuration(max time.Duration) (time.Duration, error) {
	if max <= 0 {
		return 0, nil
	}

	// Generate random int64 using crypto/rand
	maxInt := big.NewInt(int64(max))
	randomInt, err := rand.Int(rand.Reader, maxInt)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}

	return time.Duration(randomInt.Int64()), nil
}

// addToExecutionQueue adds a search to the execution queue
func (s *Scheduler) addToExecutionQueue(search db.GormSearch, delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Debugf("Queued search '%s' (ID: %d) for execution in %v", search.Name, search.ID, delay)

	// Create timer for execution
	timer := time.AfterFunc(delay, func() {
		// When timer fires, try to remove from queue tracker if possible
		// ideally we'd remove it from the slice, but that requires locking and finding indices
		// simpler: just execute. The slice is for bounding pending executions.
		select {
		case s.searchQueue <- search:
			log.Debugf("Search '%s' (ID: %d) added to execution queue", search.Name, search.ID)
		case <-s.shutdownChan:
			log.Debugf("Scheduler shutting down, search '%s' (ID: %d) not executed", search.Name, search.ID)
		}
	})

	// Add to queue
	s.searchExecutionQueue = append(s.searchExecutionQueue, queuedSearch{
		search: search,
		timer:  timer,
	})

	// Bound the queue size to prevent memory leak
	if len(s.searchExecutionQueue) > s.config.AntiBot.Timing.MaxQueueSize {
		log.Warnf("Execution queue size exceeded maximum (%d), clearing oldest entries",
			s.config.AntiBot.Timing.MaxQueueSize)

		// Calculate how many to remove
		removeCount := len(s.searchExecutionQueue) - s.config.AntiBot.Timing.MaxQueueSize
		for i := 0; i < removeCount; i++ {
			// Stop timer of evicted item
			if s.searchExecutionQueue[i].timer != nil {
				s.searchExecutionQueue[i].timer.Stop()
			}
		}

		// Keep only the most recent items
		s.searchExecutionQueue = s.searchExecutionQueue[removeCount:]
	}
}

// processSearchExecution processes a search execution
func (s *Scheduler) processSearchExecution(search db.GormSearch) {
	// Mark search as active
	s.markSearchActive(search.ID, true)
	defer s.markSearchActive(search.ID, false)

	log.Infof("Processing search execution for '%s' (ID: %d)", search.Name, search.ID)

	// Execute the search with retry logic
	err := s.executeSearchWithRetry(search)
	if err != nil {
		log.Errorf("Search execution failed for '%s' (ID: %d): %v", search.Name, search.ID, err)

		// Apply exponential backoff for retry
		s.scheduleRetry(search)
	} else {
		log.Infof("Search execution completed for '%s' (ID: %d)", search.Name, search.ID)
	}
}

// executeSearchWithRetry executes a search with configurable retry logic
func (s *Scheduler) executeSearchWithRetry(search db.GormSearch) error {
	// Use configurable retry parameters
	maxRetries := s.config.AntiBot.Retry.MaxRetries
	retryDelay := s.config.AntiBot.Retry.BaseDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := s.executeSearchAttempt(search)
		if err == nil {
			return nil // Success
		}

		if attempt < maxRetries {
			log.Warnf("Search attempt %d failed for '%s' (ID: %d): %v. Retrying in %v...",
				attempt+1, search.Name, search.ID, err, retryDelay)

			// Exponential backoff with configurable max delay
			time.Sleep(retryDelay)
			retryDelay *= 2
			if retryDelay > s.config.AntiBot.Retry.MaxDelay {
				retryDelay = s.config.AntiBot.Retry.MaxDelay
			}
		}
	}

	return fmt.Errorf("search failed after %d retries", maxRetries)
}

// processSearchResults processes search results and stores new items
func (s *Scheduler) processSearchResults(search db.GormSearch, response *api.SearchResponse) int {
	newItemsFound := 0

	for _, item := range response.Items {
		// Check if item already exists
		existingItem, err := s.repo.GetItemByGoodwillID(context.Background(), item.ID)
		if err != nil {
			log.Errorf("Failed to check for existing item %s: %v", item.ID, err)
			continue
		}

		if existingItem != nil {
			// Item already exists, update if needed
			if existingItem.CurrentPrice != item.CurrentPrice {
				existingItem.CurrentPrice = item.CurrentPrice
				existingItem.UpdatedAt = time.Now()
				existingItem.LastSeen = time.Now()

				if err := s.repo.UpdateItem(context.Background(), *existingItem); err != nil {
					log.Errorf("Failed to update existing item %s: %v", item.ID, err)
				} else {
					// Record price history
					priceHistory := db.GormPriceHistory{
						ItemID:     existingItem.ID,
						Price:      item.CurrentPrice,
						PriceType:  "current",
						RecordedAt: time.Now(),
					}

					_, err := s.repo.AddPriceHistory(context.Background(), priceHistory)
					if err != nil {
						log.Errorf("Failed to record price history for item %s: %v", item.ID, err)
					}
				}
			}
			continue
		}

		// New item found
		newItemsFound++

		// Convert search item to database item
		dbItem, err := api.ParseSearchItemToDBItem(item)
		if err != nil {
			log.Errorf("Failed to parse search item %s: %v", item.ID, err)
			continue
		}

		// Add new item to database
		itemID, err := s.repo.AddItem(context.Background(), *dbItem)
		if err != nil {
			log.Errorf("Failed to add new item %s: %v", item.ID, err)
			continue
		}

		// Record initial price history
		priceHistory := db.GormPriceHistory{
			ItemID:     itemID,
			Price:      item.CurrentPrice,
			PriceType:  "current",
			RecordedAt: time.Now(),
		}

		_, err = s.repo.AddPriceHistory(context.Background(), priceHistory)
		if err != nil {
			log.Errorf("Failed to record initial price history for item %s: %v", item.ID, err)
		}

		// Create search-item mapping
		err = s.repo.AddSearchItemMapping(context.Background(), search.ID, itemID, time.Now())
		if err != nil {
			log.Errorf("Failed to create search-item mapping for item %s: %v", item.ID, err)
		}

		log.Infof("Added new item: %s - %s", item.ID, item.Title)
	}

	return newItemsFound
}

// executeSearchAttempt executes a single search attempt
func (s *Scheduler) executeSearchAttempt(search db.GormSearch) error {
	log.Infof("Executing search attempt for '%s' (ID: %d)", search.Name, search.ID)
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Prepare search parameters
	var minPrice, maxPrice float64
	if search.MinPrice != nil {
		minPrice = *search.MinPrice
	}
	if search.MaxPrice != nil {
		maxPrice = *search.MaxPrice
	}

	params := api.SearchParams{
		Category:  search.CategoryFilter,
		Seller:    search.SellerFilter,
		Condition: search.ConditionFilter,
		Shipping:  search.ShippingFilter,
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		SortBy:    search.SortBy,
		Page:      1,
		PageSize:  50,
	}

	// Execute actual search using API client
	searchResponse, err := s.apiClient.Search(ctx, search.Query, params)
	if err != nil {
		return fmt.Errorf("search execution failed: %w", err)
	}

	log.Infof("Search '%s' completed: found %d items", search.Name, searchResponse.Total)

	// Process search results
	newItemsFound := s.processSearchResults(search, searchResponse)

	// Record successful execution
	execution := db.GormSearchExecution{
		SearchID:      search.ID,
		ExecutedAt:    time.Now(),
		Status:        "success",
		ItemsFound:    searchResponse.Total,
		NewItemsFound: newItemsFound,
		DurationMS:    int(time.Since(startTime).Milliseconds()),
	}

	_, err = s.repo.AddSearchExecution(context.Background(), execution)
	if err != nil {
		return fmt.Errorf("failed to record search execution: %w", err)
	}

	return nil
}

// scheduleRetry schedules a search for retry with configurable backoff
func (s *Scheduler) scheduleRetry(search db.GormSearch) {
	// Use configurable retry parameters
	retryDelay := s.config.AntiBot.Retry.BaseDelay

	// Add some jitter to avoid thundering herd
	jitter, err := s.getRandomDuration(30 * time.Second)
	if err != nil {
		log.Errorf("Failed to generate retry jitter: %v", err)
		jitter = 15 * time.Second
	}
	retryDelay += jitter

	// Ensure retry delay doesn't exceed maximum
	if retryDelay > s.config.AntiBot.Retry.MaxDelay {
		retryDelay = s.config.AntiBot.Retry.MaxDelay
	}

	log.Warnf("Scheduling retry for search '%s' (ID: %d) in %v", search.Name, search.ID, retryDelay)

	// Schedule the retry
	time.AfterFunc(retryDelay, func() {
		select {
		case s.searchQueue <- search:
			log.Infof("Retry scheduled for search '%s' (ID: %d)", search.Name, search.ID)
		case <-s.shutdownChan:
			log.Infof("Scheduler shutting down, retry for search '%s' (ID: %d) cancelled", search.Name, search.ID)
		}
	})
}

// markSearchActive marks a search as active/inactive
func (s *Scheduler) markSearchActive(searchID int, active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if active {
		s.activeSearches[searchID] = true
	} else {
		delete(s.activeSearches, searchID)
	}
}

// IsSearchActive checks if a search is currently being executed
func (s *Scheduler) IsSearchActive(searchID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.activeSearches[searchID]
}

// GetQueueStats returns statistics about the scheduling queue
func (s *Scheduler) GetQueueStats() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := make(map[string]interface{})

	stats["queue_size"] = len(s.searchExecutionQueue)
	stats["active_searches"] = len(s.activeSearches)
	stats["pending_executions"] = len(s.searchExecutionQueue)

	return stats
}

// schedulerWorker handles search execution in a worker
func (s *Scheduler) schedulerWorker(workerID int) {
	defer s.workersWg.Done()

	log.Infof("Scheduler worker %d started", workerID)

	for {
		select {
		case search := <-s.searchQueue:
			s.processSearchExecution(search)
		case <-s.shutdownChan:
			log.Infof("Scheduler worker %d shutting down", workerID)
			return
		}
	}
}

// GetActiveSearches returns currently active searches
func (s *Scheduler) GetActiveSearches() ([]int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	activeIDs := make([]int, 0, len(s.activeSearches))
	for searchID := range s.activeSearches {
		activeIDs = append(activeIDs, searchID)
	}

	return activeIDs, nil
}

// TriggerSearch manually triggers a search execution immediately
func (s *Scheduler) TriggerSearch(searchID int) error {
	// First check if search exists and is not already running
	// We need to fetch the search first to pass it to the queue
	// Note: We avoid holding the lock while calling other methods that might lock

	search, err := s.repo.GetSearchByID(context.Background(), searchID)
	if err != nil {
		return fmt.Errorf("failed to get search: %w", err)
	}
	if search == nil {
		return fmt.Errorf("search not found")
	}

	if !search.Enabled {
		return fmt.Errorf("search is disabled")
	}

	s.mu.Lock()
	if s.activeSearches[searchID] {
		s.mu.Unlock()
		return fmt.Errorf("search is already active")
	}
	s.mu.Unlock()

	// Add to execution queue immediately
	s.addToExecutionQueue(*search, 0)

	return nil
}

// ClearQueue clears the execution queue
func (s *Scheduler) ClearQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, qs := range s.searchExecutionQueue {
		if qs.timer != nil {
			qs.timer.Stop()
		}
	}

	s.searchExecutionQueue = make([]queuedSearch, 0)
	log.Info("Execution queue cleared")
}
