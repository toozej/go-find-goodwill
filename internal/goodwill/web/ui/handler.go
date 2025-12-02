package ui

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/internal/goodwill/web/api"
)

// UIHandler represents the UI handler
type UIHandler struct {
	templates *template.Template
	staticFS  http.FileSystem
	log       *logrus.Logger
	repo      db.Repository
}

// NewUIHandler creates a new UI handler
func NewUIHandler(staticFS embed.FS, logger *logrus.Logger, repo db.Repository) (*UIHandler, error) {
	// Parse templates
	templates, err := parseTemplates(staticFS)
	if err != nil {
		return nil, err
	}

	return &UIHandler{
		templates: templates,
		staticFS:  http.FS(staticFS),
		log:       logger,
		repo:      repo,
	}, nil
}

// parseTemplates parses all HTML templates using template.ParseFS
func parseTemplates(staticFS embed.FS) (*template.Template, error) {
	// Use template.ParseFS to parse all templates in the templates directory
	tmpl := template.New("")

	// Parse all templates in the templates directory
	tmpl, err := tmpl.ParseFS(staticFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// SetupRoutes configures UI routes
func (h *UIHandler) SetupRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(h.staticFS)))

	// Page routes
	mux.HandleFunc("GET /dashboard", h.handleDashboard)
	mux.HandleFunc("GET /searches", h.handleSearches)
	mux.HandleFunc("GET /items", h.handleItems)
	mux.HandleFunc("GET /notifications", h.handleNotifications)
	mux.HandleFunc("GET /settings", h.handleSettings)
	mux.HandleFunc("GET /login", h.handleLogin)
	mux.HandleFunc("GET /", h.handleDashboard)
}

// handleDashboard handles the dashboard page
func (h *UIHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get all searches
	searches, err := h.repo.GetSearches(r.Context())
	if err != nil {
		h.log.Errorf("Failed to get searches: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get active searches
	activeSearches, err := h.repo.GetActiveSearches(r.Context())
	if err != nil {
		h.log.Errorf("Failed to get active searches: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get all items
	items, err := h.repo.GetItems(r.Context())
	if err != nil {
		h.log.Errorf("Failed to get items: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get recent notifications
	notifications, _, err := h.repo.GetNotificationsFiltered(r.Context(), nil, nil, 5, 0)
	if err != nil {
		h.log.Errorf("Failed to get recent notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert notifications to API response format
	var notificationResponses []api.NotificationResponse
	for _, notification := range notifications {
		id := notification.ID
		itemID := notification.ItemID
		searchID := notification.SearchID
		retryCount := notification.RetryCount

		notificationResponses = append(notificationResponses, api.NotificationResponse{
			ID:               id,
			ItemID:           itemID,
			SearchID:         searchID,
			NotificationType: notification.NotificationType,
			Status:           notification.Status,
			CreatedAt:        notification.CreatedAt,
			UpdatedAt:        notification.UpdatedAt,
			SentAt:           notification.SentAt,
			DeliveredAt:      notification.DeliveredAt,
			ErrorMessage:     notification.ErrorMessage,
			RetryCount:       retryCount,
		})
	}

	// Get search statistics
	var searchStats []SearchStat
	for _, search := range searches {
		searchItems, err := h.repo.GetItemsBySearchID(r.Context(), search.ID)
		if err != nil {
			h.log.Errorf("Failed to get items for search %d: %v", search.ID, err)
			continue
		}

		searchID := search.ID

		searchStats = append(searchStats, SearchStat{
			SearchID:   searchID,
			SearchName: search.Name,
			ItemCount:  len(searchItems),
			LastRun:    search.LastChecked,
		})
	}

	data := DashboardData{
		Title:               "Dashboard",
		TotalSearches:       len(searches),
		ActiveSearches:      len(activeSearches),
		TotalItems:          len(items),
		ActiveItems:         len(items), // All items are considered active for now
		RecentNotifications: notificationResponses,
		SearchStats:         searchStats,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		h.log.Errorf("Failed to render dashboard: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleSearches handles the searches page
func (h *UIHandler) handleSearches(w http.ResponseWriter, r *http.Request) {
	// Get searches with pagination
	searches, totalCount, err := h.repo.GetSearchesFiltered(r.Context(), nil, 20, 0)
	if err != nil {
		h.log.Errorf("Failed to get searches: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	var searchResponses []api.SearchResponse
	for _, search := range searches {
		searchID := search.ID

		searchResponses = append(searchResponses, api.SearchResponse{
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

	data := SearchesData{
		Title:    "Search Management",
		Searches: searchResponses,
		Total:    totalCount,
		Limit:    20,
		Offset:   0,
	}

	if err := h.templates.ExecuteTemplate(w, "searches.html", data); err != nil {
		h.log.Errorf("Failed to render searches: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleItems handles the items page
func (h *UIHandler) handleItems(w http.ResponseWriter, r *http.Request) {
	// Get items with pagination
	items, totalCount, err := h.repo.GetItemsFiltered(r.Context(), nil, nil, nil, nil, nil, 20, 0)
	if err != nil {
		h.log.Errorf("Failed to get items: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	var itemResponses []api.ItemResponse
	for _, item := range items {
		itemID := item.ID

		itemResponses = append(itemResponses, api.ItemResponse{
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

	data := ItemsData{
		Title:  "Item Management",
		Items:  itemResponses,
		Total:  totalCount,
		Limit:  20,
		Offset: 0,
	}

	if err := h.templates.ExecuteTemplate(w, "items.html", data); err != nil {
		h.log.Errorf("Failed to render items: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleNotifications handles the notifications page
func (h *UIHandler) handleNotifications(w http.ResponseWriter, r *http.Request) {
	// Get notifications with pagination
	notifications, totalCount, err := h.repo.GetNotificationsFiltered(r.Context(), nil, nil, 20, 0)
	if err != nil {
		h.log.Errorf("Failed to get notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	var notificationResponses []api.NotificationResponse
	for _, notification := range notifications {
		id := notification.ID
		itemID := notification.ItemID
		searchID := notification.SearchID
		retryCount := notification.RetryCount

		notificationResponses = append(notificationResponses, api.NotificationResponse{
			ID:               id,
			ItemID:           itemID,
			SearchID:         searchID,
			NotificationType: notification.NotificationType,
			Status:           notification.Status,
			CreatedAt:        notification.CreatedAt,
			UpdatedAt:        notification.UpdatedAt,
			SentAt:           notification.SentAt,
			DeliveredAt:      notification.DeliveredAt,
			ErrorMessage:     notification.ErrorMessage,
			RetryCount:       retryCount,
		})
	}

	data := NotificationsData{
		Title:         "Notification Center",
		Notifications: notificationResponses,
		Total:         totalCount,
		Limit:         20,
		Offset:        0,
	}

	if err := h.templates.ExecuteTemplate(w, "notifications.html", data); err != nil {
		h.log.Errorf("Failed to render notifications: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleSettings handles the settings page
func (h *UIHandler) handleSettings(w http.ResponseWriter, r *http.Request) {
	data := SettingsData{
		Title: "System Settings",
	}

	if err := h.templates.ExecuteTemplate(w, "settings.html", data); err != nil {
		h.log.Errorf("Failed to render settings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleLogin handles the login page
func (h *UIHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	data := LoginData{
		Title: "Login",
	}

	if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		h.log.Errorf("Failed to render login: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
