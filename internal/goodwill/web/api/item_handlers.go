package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// handleGetItems handles GET /api/v1/items
func (s *Server) handleGetItems(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	searchIDStr := queryParams.Get("search_id")
	status := queryParams.Get("status")
	minPriceStr := queryParams.Get("min_price")
	maxPriceStr := queryParams.Get("max_price")
	category := queryParams.Get("category")
	limitStr := queryParams.Get("limit")
	offsetStr := queryParams.Get("offset")

	// Parse search_id filter
	var searchIDFilter *int
	if searchIDStr != "" {
		searchID, err := strconv.Atoi(searchIDStr)
		if err != nil {
			s.handleError(w, fmt.Errorf("invalid search_id"), http.StatusBadRequest)
			return
		}
		searchIDFilter = &searchID
	}

	// Parse min_price filter
	var minPriceFilter *float64
	if minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			s.handleError(w, fmt.Errorf("invalid min_price"), http.StatusBadRequest)
			return
		}
		minPriceFilter = &minPrice
	}

	// Parse max_price filter
	var maxPriceFilter *float64
	if maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			s.handleError(w, fmt.Errorf("invalid max_price"), http.StatusBadRequest)
			return
		}
		maxPriceFilter = &maxPrice
	}

	// Parse pagination
	limit := 20
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			s.handleError(w, fmt.Errorf("invalid limit"), http.StatusBadRequest)
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			s.handleError(w, fmt.Errorf("invalid offset"), http.StatusBadRequest)
			return
		}
	}

	// Use database-level filtering and pagination
	var statusFilter *string
	if status != "" {
		statusFilter = &status
	}

	var categoryFilter *string
	if category != "" {
		categoryFilter = &category
	}

	paginatedItems, totalCount, err := s.repo.GetItemsFiltered(
		r.Context(),
		searchIDFilter,
		statusFilter,
		categoryFilter,
		minPriceFilter,
		maxPriceFilter,
		limit,
		offset,
	)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response ItemListResponse
	response.Total = totalCount
	response.Limit = limit
	response.Offset = offset

	for _, item := range paginatedItems {
		itemID := item.ID
		response.Items = append(response.Items, ItemResponse{
			ID:              itemID,
			GoodwillID:      item.GoodwillID,
			Title:           item.Title,
			Seller:          item.Seller,
			CurrentPrice:    item.CurrentPrice,
			BuyNowPrice:     item.BuyNowPrice,
			URL:             item.URL,
			ImageURL:        item.ImageURL,
			EndsAt:          item.EndsAt,
			Status:          item.Status,
			Category:        item.Category,
			Subcategory:     item.Subcategory,
			Condition:       item.Condition,
			ShippingCost:    item.ShippingCost,
			ShippingMethod:  item.ShippingMethod,
			Description:     item.Description,
			Location:        item.Location,
			PickupAvailable: item.PickupAvailable,
			ReturnsAccepted: item.ReturnsAccepted,
			WatchCount:      item.WatchCount,
			BidCount:        item.BidCount,
			ViewCount:       item.ViewCount,
			FirstSeen:       item.FirstSeen,
			LastSeen:        item.LastSeen,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError)
		return
	}
}

// handleGetItem handles GET /api/v1/items/{id}
func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.handleError(w, fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.handleError(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}

	// Get item from repository
	itemID := id
	item, err := s.repo.GetItemByID(r.Context(), itemID)
	if err != nil {
		s.handleError(w, err, http.StatusNotFound)
		return
	}

	// Convert to response format
	itemResponseID := item.ID
	response := ItemResponse{
		ID:              itemResponseID,
		GoodwillID:      item.GoodwillID,
		Title:           item.Title,
		Seller:          item.Seller,
		CurrentPrice:    item.CurrentPrice,
		BuyNowPrice:     item.BuyNowPrice,
		URL:             item.URL,
		ImageURL:        item.ImageURL,
		EndsAt:          item.EndsAt,
		Status:          item.Status,
		Category:        item.Category,
		Subcategory:     item.Subcategory,
		Condition:       item.Condition,
		ShippingCost:    item.ShippingCost,
		ShippingMethod:  item.ShippingMethod,
		Description:     item.Description,
		Location:        item.Location,
		PickupAvailable: item.PickupAvailable,
		ReturnsAccepted: item.ReturnsAccepted,
		WatchCount:      item.WatchCount,
		BidCount:        item.BidCount,
		ViewCount:       item.ViewCount,
		FirstSeen:       item.FirstSeen,
		LastSeen:        item.LastSeen,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError)
		return
	}
}

// handleGetItemHistory handles GET /api/v1/items/{id}/history
func (s *Server) handleGetItemHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.handleError(w, fmt.Errorf("id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		s.handleError(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}

	// Get price history
	itemID := id
	priceHistory, err := s.repo.GetPriceHistory(r.Context(), itemID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Get bid history
	bidHistory, err := s.repo.GetBidHistory(r.Context(), itemID)
	if err != nil {
		s.handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var response ItemHistoryResponse

	for _, ph := range priceHistory {
		response.PriceHistory = append(response.PriceHistory, PriceHistoryResponse{
			Price:      ph.Price,
			PriceType:  ph.PriceType,
			RecordedAt: ph.RecordedAt,
		})
	}

	for _, bh := range bidHistory {
		response.BidHistory = append(response.BidHistory, BidHistoryResponse{
			BidAmount:  bh.BidAmount,
			Bidder:     bh.Bidder,
			BidderID:   bh.BidderID,
			RecordedAt: bh.RecordedAt,
		})
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.handleError(w, fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError)
		return
	}
}
