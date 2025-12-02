package db

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// GormRepository implements the Repository interface using GORM models directly
type GormRepository struct {
	db *GormDatabase
}

// NewGormRepository creates a new GORM repository
func NewGormRepository(db *GormDatabase) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

// ensureDB checks if the database connection is available
func (r *GormRepository) ensureDB() (*gorm.DB, error) {
	gormDB := r.db.GetDB()
	if gormDB == nil {
		return nil, errors.New("database not connected")
	}
	return gormDB, nil
}

// GetSearches implements Repository.GetSearches
func (r *GormRepository) GetSearches(ctx context.Context) ([]GormSearch, error) {
	return r.GetSearchesPaginated(ctx, 1, 100) // Default pagination: page 1, 100 items per page
}

// GetSearchesPaginated implements paginated search retrieval
func (r *GormRepository) GetSearchesPaginated(ctx context.Context, page int, pageSize int) ([]GormSearch, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	var gormSearches []GormSearch
	result := gormDB.WithContext(ctx).Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&gormSearches)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query paginated searches: %w", result.Error)
	}

	return gormSearches, nil
}

// GetSearchesFiltered implements database-level filtering and pagination for searches
func (r *GormRepository) GetSearchesFiltered(ctx context.Context, enabled *bool, limit int, offset int) ([]GormSearch, int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, 0, err
	}

	// Start with base query
	query := gormDB.WithContext(ctx).Model(&GormSearch{})

	// Apply filters
	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	// Count total matching searches
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered searches: %w", err)
	}

	// Apply pagination
	query = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	// Execute query
	var gormSearches []GormSearch
	result := query.Find(&gormSearches)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query filtered searches: %w", result.Error)
	}

	return gormSearches, int(total), nil
}

// GetSearchByID implements Repository.GetSearchByID
func (r *GormRepository) GetSearchByID(ctx context.Context, id int) (*GormSearch, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormSearch GormSearch
	result := gormDB.WithContext(ctx).First(&gormSearch, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query search: %w", result.Error)
	}

	return &gormSearch, nil
}

// AddSearch implements Repository.AddSearch
func (r *GormRepository) AddSearch(ctx context.Context, search GormSearch) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&search)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert search: %w", result.Error)
	}

	return search.ID, nil
}

// UpdateSearch implements Repository.UpdateSearch
func (r *GormRepository) UpdateSearch(ctx context.Context, search GormSearch) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	result := gormDB.WithContext(ctx).Save(&search)
	if result.Error != nil {
		return fmt.Errorf("failed to update search: %w", result.Error)
	}

	return nil
}

// DeleteSearch implements Repository.DeleteSearch
func (r *GormRepository) DeleteSearch(ctx context.Context, id int) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	result := gormDB.WithContext(ctx).Delete(&GormSearch{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete search: %w", result.Error)
	}

	return nil
}

// GetActiveSearches implements Repository.GetActiveSearches
func (r *GormRepository) GetActiveSearches(ctx context.Context) ([]GormSearch, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormSearches []GormSearch
	result := gormDB.WithContext(ctx).Where("enabled = ?", true).Find(&gormSearches)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query active searches: %w", result.Error)
	}

	return gormSearches, nil
}

// GetItems implements Repository.GetItems
func (r *GormRepository) GetItems(ctx context.Context) ([]GormItem, error) {
	return r.GetItemsPaginated(ctx, 1, 100) // Default pagination: page 1, 100 items per page
}

// GetItemsPaginated implements Repository.GetItemsPaginated
func (r *GormRepository) GetItemsPaginated(ctx context.Context, page int, pageSize int) ([]GormItem, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	var gormItems []GormItem
	result := gormDB.WithContext(ctx).Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&gormItems)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query paginated items: %w", result.Error)
	}

	return gormItems, nil
}

// GetItemsFiltered implements database-level filtering and pagination for items
func (r *GormRepository) GetItemsFiltered(ctx context.Context, searchID *int, status *string, category *string, minPrice *float64, maxPrice *float64, limit int, offset int) ([]GormItem, int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, 0, err
	}

	// Start with base query
	query := gormDB.WithContext(ctx).Model(&GormItem{})

	// Apply filters
	if searchID != nil {
		query = query.Joins("JOIN gorm_search_item_mappings ON gorm_search_item_mappings.item_id = gorm_items.id").
			Where("gorm_search_item_mappings.search_id = ?", *searchID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if category != nil {
		query = query.Where("category = ?", *category)
	}
	if minPrice != nil {
		query = query.Where("current_price >= ?", *minPrice)
	}
	if maxPrice != nil {
		query = query.Where("current_price <= ?", *maxPrice)
	}

	// Count total matching items
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered items: %w", err)
	}

	// Apply pagination
	query = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	// Execute query
	var gormItems []GormItem
	result := query.Find(&gormItems)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query filtered items: %w", result.Error)
	}

	return gormItems, int(total), nil
}

// GetRecentItemsForDeduplication implements database-level filtering for recent items
func (r *GormRepository) GetRecentItemsForDeduplication(ctx context.Context, maxAge time.Duration, limit int, offset int) ([]GormItem, int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, 0, err
	}

	// Calculate the cutoff time
	cutoffTime := time.Now().Add(-maxAge)

	// Start with base query
	query := gormDB.WithContext(ctx).Model(&GormItem{})

	// Apply recency filter
	query = query.Where("last_seen >= ?", cutoffTime)

	// Count total matching items
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count recent items: %w", err)
	}

	// Apply pagination
	query = query.Order("last_seen DESC").
		Offset(offset).
		Limit(limit)

	// Execute query
	var gormItems []GormItem
	result := query.Find(&gormItems)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query recent items: %w", result.Error)
	}

	return gormItems, int(total), nil
}

// GetItemByID implements Repository.GetItemByID
func (r *GormRepository) GetItemByID(ctx context.Context, id int) (*GormItem, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormItem GormItem
	result := gormDB.WithContext(ctx).First(&gormItem, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query item: %w", result.Error)
	}

	return &gormItem, nil
}

// GetItemByGoodwillID implements Repository.GetItemByGoodwillID
func (r *GormRepository) GetItemByGoodwillID(ctx context.Context, goodwillID string) (*GormItem, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormItem GormItem
	result := gormDB.WithContext(ctx).Where("goodwill_id = ?", goodwillID).First(&gormItem)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query item: %w", result.Error)
	}

	return &gormItem, nil
}

// AddItem implements Repository.AddItem
func (r *GormRepository) AddItem(ctx context.Context, item GormItem) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&item)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert item: %w", result.Error)
	}

	return item.ID, nil
}

// UpdateItem implements Repository.UpdateItem
func (r *GormRepository) UpdateItem(ctx context.Context, item GormItem) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	result := gormDB.WithContext(ctx).Save(&item)
	if result.Error != nil {
		return fmt.Errorf("failed to update item: %w", result.Error)
	}

	return nil
}

// GetItemsBySearchID implements Repository.GetItemsBySearchID
func (r *GormRepository) GetItemsBySearchID(ctx context.Context, searchID int) ([]GormItem, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormItems []GormItem
	result := gormDB.WithContext(ctx).Joins("JOIN gorm_search_item_mappings ON gorm_search_item_mappings.item_id = gorm_items.id").
		Where("gorm_search_item_mappings.search_id = ?", searchID).
		Find(&gormItems)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query items by search ID: %w", result.Error)
	}

	return gormItems, nil
}

// AddSearchExecution implements Repository.AddSearchExecution
func (r *GormRepository) AddSearchExecution(ctx context.Context, execution GormSearchExecution) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&execution)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert search execution: %w", result.Error)
	}

	return execution.ID, nil
}

// GetSearchHistory implements Repository.GetSearchHistory
func (r *GormRepository) GetSearchHistory(ctx context.Context, searchID int, limit int) ([]GormSearchExecution, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var query *gorm.DB
	if limit > 0 {
		query = gormDB.WithContext(ctx).Where("search_id = ?", searchID).
			Order("executed_at DESC").
			Limit(limit)
	} else {
		query = gormDB.WithContext(ctx).Where("search_id = ?", searchID).
			Order("executed_at DESC")
	}

	var gormExecutions []GormSearchExecution
	result := query.Find(&gormExecutions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query search history: %w", result.Error)
	}

	return gormExecutions, nil
}

// AddSearchItemMapping implements Repository.AddSearchItemMapping
func (r *GormRepository) AddSearchItemMapping(ctx context.Context, searchID int, itemID int, foundAt time.Time) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	mapping := GormSearchItemMapping{
		SearchID: searchID,
		ItemID:   itemID,
		FoundAt:  foundAt,
	}
	result := gormDB.WithContext(ctx).Create(&mapping)
	if result.Error != nil {
		return fmt.Errorf("failed to insert search-item mapping: %w", result.Error)
	}

	return nil
}

// AddPriceHistory implements Repository.AddPriceHistory
func (r *GormRepository) AddPriceHistory(ctx context.Context, history GormPriceHistory) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&history)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert price history: %w", result.Error)
	}

	return history.ID, nil
}

// GetPriceHistory implements Repository.GetPriceHistory
func (r *GormRepository) GetPriceHistory(ctx context.Context, itemID int) ([]GormPriceHistory, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormHistories []GormPriceHistory
	result := gormDB.WithContext(ctx).Where("item_id = ?", itemID).
		Order("recorded_at DESC").
		Find(&gormHistories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query price history: %w", result.Error)
	}

	return gormHistories, nil
}

// AddBidHistory implements Repository.AddBidHistory
func (r *GormRepository) AddBidHistory(ctx context.Context, history GormBidHistory) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&history)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert bid history: %w", result.Error)
	}

	return history.ID, nil
}

// GetBidHistory implements Repository.GetBidHistory
func (r *GormRepository) GetBidHistory(ctx context.Context, itemID int) ([]GormBidHistory, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormHistories []GormBidHistory
	result := gormDB.WithContext(ctx).Where("item_id = ?", itemID).
		Order("recorded_at DESC").
		Find(&gormHistories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query bid history: %w", result.Error)
	}

	return gormHistories, nil
}

// QueueNotification implements Repository.QueueNotification
func (r *GormRepository) QueueNotification(ctx context.Context, notification GormNotification) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&notification)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert notification: %w", result.Error)
	}

	return notification.ID, nil
}

// UpdateNotification implements Repository.UpdateNotification
func (r *GormRepository) UpdateNotification(ctx context.Context, notification GormNotification) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	result := gormDB.WithContext(ctx).Save(&notification)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification: %w", result.Error)
	}

	return nil
}

// UpdateNotificationStatus implements Repository.UpdateNotificationStatus
func (r *GormRepository) UpdateNotificationStatus(ctx context.Context, id int, status string) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	result := gormDB.WithContext(ctx).Model(&GormNotification{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification status: %w", result.Error)
	}

	return nil
}

// GetPendingNotifications implements Repository.GetPendingNotifications
func (r *GormRepository) GetPendingNotifications(ctx context.Context) ([]GormNotification, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormNotifications []GormNotification
	result := gormDB.WithContext(ctx).Where("status = ?", "queued").
		Order("created_at ASC").
		Find(&gormNotifications)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query pending notifications: %w", result.Error)
	}

	return gormNotifications, nil
}

// GetNotificationByID implements Repository.GetNotificationByID
func (r *GormRepository) GetNotificationByID(ctx context.Context, id int) (*GormNotification, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormNotification GormNotification
	result := gormDB.WithContext(ctx).First(&gormNotification, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query notification: %w", result.Error)
	}

	return &gormNotification, nil
}

// GetAllNotifications implements Repository.GetAllNotifications
func (r *GormRepository) GetAllNotifications(ctx context.Context) ([]GormNotification, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormNotifications []GormNotification
	result := gormDB.WithContext(ctx).Order("created_at DESC").Find(&gormNotifications)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", result.Error)
	}

	return gormNotifications, nil
}

// GetNotificationsFiltered implements database-level filtering and pagination for notifications
func (r *GormRepository) GetNotificationsFiltered(ctx context.Context, status *string, notificationType *string, limit int, offset int) ([]GormNotification, int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, 0, err
	}

	// Start with base query
	query := gormDB.WithContext(ctx).Model(&GormNotification{})

	// Apply filters
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if notificationType != nil {
		query = query.Where("notification_type = ?", *notificationType)
	}

	// Count total matching notifications
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count filtered notifications: %w", err)
	}

	// Apply pagination
	query = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	// Execute query
	var gormNotifications []GormNotification
	result := query.Find(&gormNotifications)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query filtered notifications: %w", result.Error)
	}

	return gormNotifications, int(total), nil
}

// NotificationCountStats holds counts of notifications by status
type NotificationCountStats struct {
	Total      int
	Pending    int
	Processing int
	Delivered  int
	Failed     int
}

// GetNotificationStats returns all notification counts in a single query
func (r *GormRepository) GetNotificationStats(ctx context.Context) (*NotificationCountStats, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	type Result struct {
		Status string
		Count  int
	}

	var results []Result
	if err := gormDB.WithContext(ctx).Model(&GormNotification{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	stats := &NotificationCountStats{}
	for _, res := range results {
		count := res.Count
		stats.Total += count
		switch res.Status {
		case "queued", "pending":
			stats.Pending += count
		case "processing":
			stats.Processing += count
		case "delivered":
			stats.Delivered += count
		case "failed":
			stats.Failed += count
		}
	}

	return stats, nil
}

// Deprecated: methods below are kept for interface compatibility but delegate to GetNotificationStats
// They should be removed once the interface is updated.

// GetTotalNotificationCount implements Repository.GetTotalNotificationCount
func (r *GormRepository) GetTotalNotificationCount(ctx context.Context) (int, error) {
	stats, err := r.GetNotificationStats(ctx)
	if err != nil {
		return 0, err
	}
	return stats.Total, nil
}

// GetPendingNotificationCount implements Repository.GetPendingNotificationCount
func (r *GormRepository) GetPendingNotificationCount(ctx context.Context) (int, error) {
	stats, err := r.GetNotificationStats(ctx)
	if err != nil {
		return 0, err
	}
	return stats.Pending, nil
}

// GetProcessingNotificationCount implements Repository.GetProcessingNotificationCount
func (r *GormRepository) GetProcessingNotificationCount(ctx context.Context) (int, error) {
	stats, err := r.GetNotificationStats(ctx)
	if err != nil {
		return 0, err
	}
	return stats.Processing, nil
}

// GetDeliveredNotificationCount implements Repository.GetDeliveredNotificationCount
func (r *GormRepository) GetDeliveredNotificationCount(ctx context.Context) (int, error) {
	stats, err := r.GetNotificationStats(ctx)
	if err != nil {
		return 0, err
	}
	return stats.Delivered, nil
}

// GetFailedNotificationCount implements Repository.GetFailedNotificationCount
func (r *GormRepository) GetFailedNotificationCount(ctx context.Context) (int, error) {
	stats, err := r.GetNotificationStats(ctx)
	if err != nil {
		return 0, err
	}
	return stats.Failed, nil
}

// GetRandomUserAgent implements Repository.GetRandomUserAgent
func (r *GormRepository) GetRandomUserAgent(ctx context.Context) (*GormUserAgent, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	// Optimized approach: Count total, generate random offset
	var count int64
	if err := gormDB.WithContext(ctx).Model(&GormUserAgent{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to count user agents: %w", err)
	}

	if count == 0 {
		return nil, nil
	}

	// Secure random number generation
	n, err := rand.Int(rand.Reader, big.NewInt(count))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}
	offset := int(n.Int64())

	var gormAgent GormUserAgent
	result := gormDB.WithContext(ctx).Where("is_active = ?", true).
		Offset(offset).
		Limit(1).
		First(&gormAgent)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query user agent: %w", result.Error)
	}

	return &gormAgent, nil
}

// UpdateUserAgentUsage implements Repository.UpdateUserAgentUsage
func (r *GormRepository) UpdateUserAgentUsage(ctx context.Context, agentID int) error {
	gormDB, err := r.ensureDB()
	if err != nil {
		return err
	}

	result := gormDB.WithContext(ctx).Model(&GormUserAgent{}).
		Where("id = ?", agentID).
		Updates(map[string]interface{}{
			"last_used":   time.Now(),
			"usage_count": gorm.Expr("usage_count + 1"),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update user agent usage: %w", result.Error)
	}

	return nil
}

func (r *GormRepository) GetActiveUserAgents(ctx context.Context) ([]GormUserAgent, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return nil, err
	}

	var gormAgents []GormUserAgent
	result := gormDB.WithContext(ctx).Where("is_active = ?", true).
		Order("usage_count ASC").
		Find(&gormAgents)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query active user agents: %w", result.Error)
	}

	return gormAgents, nil
}

// LogSystemEvent implements Repository.LogSystemEvent
func (r *GormRepository) LogSystemEvent(ctx context.Context, event GormSystemLog) (int, error) {
	gormDB, err := r.ensureDB()
	if err != nil {
		return 0, err
	}

	result := gormDB.WithContext(ctx).Create(&event)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to insert system log: %w", result.Error)
	}

	return event.ID, nil
}
