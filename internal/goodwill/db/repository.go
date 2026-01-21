package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository defines the database repository interface using GORM models directly
type Repository interface {
	// Search operations
	GetSearches(ctx context.Context) ([]GormSearch, error)
	GetSearchByID(ctx context.Context, id int) (*GormSearch, error)
	AddSearch(ctx context.Context, search GormSearch) (int, error)
	UpdateSearch(ctx context.Context, search GormSearch) error
	DeleteSearch(ctx context.Context, id int) error
	GetActiveSearches(ctx context.Context) ([]GormSearch, error)
	GetSearchesFiltered(ctx context.Context, enabled *bool, limit int, offset int) ([]GormSearch, int, error)

	// Item operations
	GetItems(ctx context.Context) ([]GormItem, error)
	GetItemsPaginated(ctx context.Context, page int, pageSize int) ([]GormItem, error)
	GetItemByID(ctx context.Context, id int) (*GormItem, error)
	GetItemByGoodwillID(ctx context.Context, goodwillID string) (*GormItem, error)
	AddItem(ctx context.Context, item GormItem) (int, error)
	UpdateItem(ctx context.Context, item GormItem) error
	GetItemsBySearchID(ctx context.Context, searchID int) ([]GormItem, error)
	GetItemsFiltered(ctx context.Context, searchID *int, status *string, category *string, minPrice *float64, maxPrice *float64, limit int, offset int) ([]GormItem, int, error)

	// Search history
	AddSearchExecution(ctx context.Context, execution GormSearchExecution) (int, error)
	GetSearchHistory(ctx context.Context, searchID int, limit int) ([]GormSearchExecution, error)

	// Search-Item mapping
	AddSearchItemMapping(ctx context.Context, searchID int, itemID int, foundAt time.Time) error

	// Price history
	AddPriceHistory(ctx context.Context, history GormPriceHistory) (int, error)
	GetPriceHistory(ctx context.Context, itemID int) ([]GormPriceHistory, error)

	// Bid history
	AddBidHistory(ctx context.Context, history GormBidHistory) (int, error)
	GetBidHistory(ctx context.Context, itemID int) ([]GormBidHistory, error)

	// Notifications
	QueueNotification(ctx context.Context, notification GormNotification) (int, error)
	UpdateNotificationStatus(ctx context.Context, id int, status string) error
	GetPendingNotifications(ctx context.Context) ([]GormNotification, error)
	GetNotificationByID(ctx context.Context, id int) (*GormNotification, error)
	UpdateNotification(ctx context.Context, notification GormNotification) error
	GetAllNotifications(ctx context.Context) ([]GormNotification, error)
	// Filtered notifications for efficient database operations
	GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]GormNotification, int, error)
	GetNotificationStats(ctx context.Context) (*NotificationCountStats, error)

	// Anti-bot
	GetRandomUserAgent(ctx context.Context) (*GormUserAgent, error)
	GetActiveUserAgents(ctx context.Context) ([]GormUserAgent, error)
	UpdateUserAgentUsage(ctx context.Context, agentID int) error

	// System
	LogSystemEvent(ctx context.Context, event GormSystemLog) (int, error)
}

// Standard error for "not found" cases
var ErrRecordNotFound = gorm.ErrRecordNotFound
