package ui

import (
	"time"

	"github.com/toozej/go-find-goodwill/internal/goodwill/web/api"
)

// DashboardData represents data for the dashboard page
type DashboardData struct {
	Title               string
	TotalSearches       int
	ActiveSearches      int
	TotalItems          int
	ActiveItems         int
	RecentNotifications []api.NotificationResponse
	SearchStats         []SearchStat
	FlashMessages       []FlashMessage
}

// FlashMessage represents a flash message to display in the UI
type FlashMessage struct {
	Type    string // success, info, warning, danger
	Message string
}

// SearchStat represents statistics for a search
type SearchStat struct {
	SearchID   int
	SearchName string
	ItemCount  int
	LastRun    *time.Time
}

// SearchesData represents data for the searches page
type SearchesData struct {
	Title         string
	Searches      []api.SearchResponse
	Total         int
	Limit         int
	Offset        int
	FlashMessages []FlashMessage
}

// ItemsData represents data for the items page
type ItemsData struct {
	Title         string
	Items         []api.ItemResponse
	Total         int
	Limit         int
	Offset        int
	FlashMessages []FlashMessage
}

// NotificationsData represents data for the notifications page
type NotificationsData struct {
	Title         string
	Notifications []api.NotificationResponse
	Total         int
	Limit         int
	Offset        int
	FlashMessages []FlashMessage
}

// SettingsData represents data for the settings page
type SettingsData struct {
	Title         string
	FlashMessages []FlashMessage
	// Settings data would go here
}

// LoginData represents data for the login page
type LoginData struct {
	Title         string
	FlashMessages []FlashMessage
	// Login form data would go here
}
