package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
)

// handleGetSearches handles GET /api/v1/searches
func (s *Server) handleGetSearches(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	enabledStr := queryParams.Get("enabled")
	limitStr := queryParams.Get("limit")
	offsetStr := queryParams.Get("offset")

	// Parse enabled filter
	var enabledFilter *bool
	if enabledStr != "" {
		enabled, err := strconv.ParseBool(enabledStr)
		if err != nil {
			s.handleError(w, err, http.StatusBadRequest)
			return
		}
		enabledFilter = &enabled
	}

	// Parse pagination
	limit := 20
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			s.handleError(w, err, http.StatusBadRequest)
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			s.handleError(w, err, http.StatusBadRequest)
			return
		}
	}

	// Use database-level filtering and pagination
	paginatedSearches, totalCount, err := s.repo.GetSearchesFiltered(
		r.Context(),
		enabledFilter,
		limit,
		offset,
	)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response SearchListResponse
	response.Total = totalCount
	response.Limit = limit
	response.Offset = offset

	for _, search := range paginatedSearches {
		searchID := search.ID
		response.Searches = append(response.Searches, SearchResponse{
			ID:                        searchID,
			Name:                      search.Name,
			Query:                     search.Query,
			RegexPattern:              search.RegexPattern,
			Enabled:                   search.Enabled,
			CreatedAt:                 search.CreatedAt,
			UpdatedAt:                 search.UpdatedAt,
			LastChecked:               search.LastChecked,
			NotificationThresholdDays: search.NotificationThresholdDays,
			MinPrice:                  search.MinPrice,
			MaxPrice:                  search.MaxPrice,
			CategoryFilter:            search.CategoryFilter,
			SellerFilter:              search.SellerFilter,
			ShippingFilter:            search.ShippingFilter,
			ConditionFilter:           search.ConditionFilter,
			SortBy:                    search.SortBy,
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}

// handleCreateSearch handles POST /api/v1/searches
func (s *Server) handleCreateSearch(w http.ResponseWriter, r *http.Request) {
	var request SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.Name == "" {
		s.handleError(w, fmt.Errorf("name is required"), http.StatusBadRequest)
		return
	}
	if request.Query == "" {
		s.handleError(w, fmt.Errorf("query is required"), http.StatusBadRequest)
		return
	}

	// Create search model
	search := db.GormSearch{
		Name:                      request.Name,
		Query:                     request.Query,
		RegexPattern:              request.RegexPattern,
		Enabled:                   request.Enabled,
		NotificationThresholdDays: request.NotificationThresholdDays,
		MinPrice:                  request.MinPrice,
		MaxPrice:                  request.MaxPrice,
		CategoryFilter:            request.CategoryFilter,
		SellerFilter:              request.SellerFilter,
		ShippingFilter:            request.ShippingFilter,
		ConditionFilter:           request.ConditionFilter,
		SortBy:                    request.SortBy,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	// Add search to repository
	searchID, err := s.repo.AddSearch(r.Context(), search)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Get the created search
	createdSearch, err := s.repo.GetSearchByID(r.Context(), searchID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Return response
	createdSearchID := createdSearch.ID
	response := SearchResponse{
		ID:                        createdSearchID,
		Name:                      createdSearch.Name,
		Query:                     createdSearch.Query,
		RegexPattern:              createdSearch.RegexPattern,
		Enabled:                   createdSearch.Enabled,
		CreatedAt:                 createdSearch.CreatedAt,
		UpdatedAt:                 createdSearch.UpdatedAt,
		LastChecked:               createdSearch.LastChecked,
		NotificationThresholdDays: createdSearch.NotificationThresholdDays,
		MinPrice:                  createdSearch.MinPrice,
		MaxPrice:                  createdSearch.MaxPrice,
		CategoryFilter:            createdSearch.CategoryFilter,
		SellerFilter:              createdSearch.SellerFilter,
		ShippingFilter:            createdSearch.ShippingFilter,
		ConditionFilter:           createdSearch.ConditionFilter,
		SortBy:                    createdSearch.SortBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}

// handleGetSearch handles GET /api/v1/searches/{id}
func (s *Server) handleGetSearch(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.handleError(w, fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	// Get search from repository
	searchID := id
	search, err := s.repo.GetSearchByID(r.Context(), searchID)
	if err != nil {
		s.handleError(w, err, http.StatusNotFound)
		return
	}

	// Convert to response format
	fetchedSearchID := search.ID
	response := SearchResponse{
		ID:                        fetchedSearchID,
		Name:                      search.Name,
		Query:                     search.Query,
		RegexPattern:              search.RegexPattern,
		Enabled:                   search.Enabled,
		CreatedAt:                 search.CreatedAt,
		UpdatedAt:                 search.UpdatedAt,
		LastChecked:               search.LastChecked,
		NotificationThresholdDays: search.NotificationThresholdDays,
		MinPrice:                  search.MinPrice,
		MaxPrice:                  search.MaxPrice,
		CategoryFilter:            search.CategoryFilter,
		SellerFilter:              search.SellerFilter,
		ShippingFilter:            search.ShippingFilter,
		ConditionFilter:           search.ConditionFilter,
		SortBy:                    search.SortBy,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}

// handleUpdateSearch handles PUT /api/v1/searches/{id}
func (s *Server) handleUpdateSearch(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.handleError(w, fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	var request SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	// Get existing search
	searchID := id
	existingSearch, err := s.repo.GetSearchByID(r.Context(), searchID)
	if err != nil {
		s.handleError(w, err, http.StatusNotFound)
		return
	}

	// Update fields
	if request.Name != "" {
		existingSearch.Name = request.Name
	}
	if request.Query != "" {
		existingSearch.Query = request.Query
	}
	if request.RegexPattern != "" {
		existingSearch.RegexPattern = request.RegexPattern
	}
	existingSearch.Enabled = request.Enabled
	existingSearch.NotificationThresholdDays = request.NotificationThresholdDays
	existingSearch.MinPrice = request.MinPrice
	existingSearch.MaxPrice = request.MaxPrice
	existingSearch.CategoryFilter = request.CategoryFilter
	existingSearch.SellerFilter = request.SellerFilter
	existingSearch.ShippingFilter = request.ShippingFilter
	existingSearch.ConditionFilter = request.ConditionFilter
	existingSearch.SortBy = request.SortBy
	existingSearch.UpdatedAt = time.Now()

	// Update search in repository
	if err := s.repo.UpdateSearch(r.Context(), *existingSearch); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Return updated search
	updatedSearchID := existingSearch.ID
	response := SearchResponse{
		ID:                        updatedSearchID,
		Name:                      existingSearch.Name,
		Query:                     existingSearch.Query,
		RegexPattern:              existingSearch.RegexPattern,
		Enabled:                   existingSearch.Enabled,
		CreatedAt:                 existingSearch.CreatedAt,
		UpdatedAt:                 existingSearch.UpdatedAt,
		LastChecked:               existingSearch.LastChecked,
		NotificationThresholdDays: existingSearch.NotificationThresholdDays,
		MinPrice:                  existingSearch.MinPrice,
		MaxPrice:                  existingSearch.MaxPrice,
		CategoryFilter:            existingSearch.CategoryFilter,
		SellerFilter:              existingSearch.SellerFilter,
		ShippingFilter:            existingSearch.ShippingFilter,
		ConditionFilter:           existingSearch.ConditionFilter,
		SortBy:                    existingSearch.SortBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}

// handleDeleteSearch handles DELETE /api/v1/searches/{id}
func (s *Server) handleDeleteSearch(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.handleError(w, fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	// Delete search from repository
	searchID := id
	if err := s.repo.DeleteSearch(r.Context(), searchID); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// handleExecuteSearch handles POST /api/v1/searches/{id}/execute
func (s *Server) handleExecuteSearch(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.handleError(w, fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.handleError(w, err, http.StatusBadRequest)
		return
	}

	// Get search from repository
	searchID := id
	search, err := s.repo.GetSearchByID(r.Context(), searchID)
	if err != nil {
		s.handleError(w, err, http.StatusNotFound)
		return
	}

	// Check if search is enabled
	if !search.Enabled {
		s.handleError(w, fmt.Errorf("search is disabled"), http.StatusBadRequest)
		return
	}

	// Trigger search execution via scheduler
	err = s.scheduler.TriggerSearch(searchID)
	if err != nil {
		// Differentiate between "already active" vs other errors
		if err.Error() == "search is already active" {
			response := map[string]interface{}{
				"status":    "already_running",
				"search_id": searchID,
				"message":   "Search is already being executed",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			if err := json.NewEncoder(w).Encode(response); err != nil {
				s.log.Errorf("Failed to encode response: %v", err)
			}
			return
		}

		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Create a dummy response for compatibility or success message
	// The scheduler creates the execution record asynchronously when processed
	// But the user might expect an execution ID immediately.
	// Since we queue it, we don't have the execution ID yet (it's created by worker).
	// We'll return accepted status.

	response := map[string]interface{}{
		"status":      "queued",
		"search_id":   searchID,
		"search_name": search.Name,
		"message":     "Search execution queued successfully",
		"queued_at":   time.Now(),
		"queue_stats": s.scheduler.GetQueueStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}
}
