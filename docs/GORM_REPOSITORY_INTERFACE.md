# GORM Repository Interface Documentation

## Introduction

This document provides comprehensive documentation for the GORM repository interface and usage patterns in the go-find-goodwill application.

## Table of Contents

1. [Repository Interface Overview](#repository-interface-overview)
2. [GORM Repository Implementation](#gorm-repository-implementation)
3. [Interface Methods](#interface-methods)
4. [Usage Patterns](#usage-patterns)
5. [Error Handling](#error-handling)
6. [Transaction Patterns](#transaction-patterns)
7. [Testing Patterns](#testing-patterns)
8. [Best Practices](#best-practices)

## Repository Interface Overview

### Interface Definition

```go
// Repository interface defines database operations
type Repository interface {
    // Search operations
    GetSearches() ([]Search, error)
    GetSearchByID(id int) (*Search, error)
    AddSearch(search Search) (int, error)
    UpdateSearch(search Search) error
    DeleteSearch(id int) error
    GetActiveSearches() ([]Search, error)

    // Item operations
    GetItems() ([]Item, error)
    GetItemByID(id int) (*Item, error)
    GetItemByGoodwillID(goodwillID string) (*Item, error)
    AddItem(item Item) (int, error)
    UpdateItem(item Item) error
    GetItemsBySearchID(searchID int) ([]Item, error)

    // Search history
    AddSearchExecution(execution SearchExecution) (int, error)
    GetSearchHistory(searchID int, limit int) ([]SearchExecution, error)

    // Price history
    AddPriceHistory(history PriceHistory) (int, error)
    GetPriceHistory(itemID int) ([]PriceHistory, error)

    // Bid history
    AddBidHistory(history BidHistory) (int, error)
    GetBidHistory(itemID int) ([]BidHistory, error)

    // Notifications
    QueueNotification(notification Notification) (int, error)
    UpdateNotification(notification Notification) error
    UpdateNotificationStatus(id int, status string) error
    GetPendingNotifications() ([]Notification, error)
    GetNotificationByID(id int) (*Notification, error)
    GetAllNotifications() ([]Notification, error)

    // Anti-bot
    GetRandomUserAgent() (*UserAgent, error)
    UpdateUserAgentUsage(agentID int) error

    // System
    LogSystemEvent(event SystemLog) (int, error)
}
```

### GORM Implementation

```go
// GormRepository implements Repository using GORM
type GormRepository struct {
    db *GormDatabase
}

func NewGormRepository(db *GormDatabase) *GormRepository {
    return &GormRepository{
        db: db,
    }
}
```

## GORM Repository Implementation

### Core Structure

```go
type GormRepository struct {
    db *GormDatabase  // GORM database connection
}

// Constructor
func NewGormRepository(db *GormDatabase) *GormRepository {
    return &GormRepository{db: db}
}
```

### Database Connection Management

```go
// GetDB returns the underlying GORM database connection
func (r *GormRepository) getDB() (*gorm.DB, error) {
    gormDB := r.db.GetDB()
    if gormDB == nil {
        return nil, errors.New("database not connected")
    }
    return gormDB, nil
}
```

## Interface Methods

### Search Operations

```go
// GetSearches returns all searches
func (r *GormRepository) GetSearches() ([]Search, error) {
    gormDB, err := r.getDB()
    if err != nil {
        return nil, err
    }

    var gormSearches []GormSearch
    result := gormDB.Find(&gormSearches)
    if result.Error != nil {
        return nil, fmt.Errorf("failed to query searches: %w", result.Error)
    }

    var searches []Search
    for _, gormSearch := range gormSearches {
        searches = append(searches, ConvertFromGormSearch(gormSearch))
    }

    return searches, nil
}

// GetSearchByID returns a single search by ID
func (r *GormRepository) GetSearchByID(id int) (*Search, error) {
    gormDB, err := r.getDB()
    if err != nil {
        return nil, err
    }

    var gormSearch GormSearch
    result := gormDB.First(&gormSearch, id)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("search with id %d not found: %w", id, result.Error)
        }
        return nil, fmt.Errorf("failed to query search: %w", result.Error)
    }

    search := ConvertFromGormSearch(gormSearch)
    return &search, nil
}
```

### Item Operations

```go
// GetItems returns all items
func (r *GormRepository) GetItems() ([]Item, error) {
    gormDB, err := r.getDB()
    if err != nil {
        return nil, err
    }

    var gormItems []GormItem
    result := gormDB.Order("created_at DESC").Find(&gormItems)
    if result.Error != nil {
        return nil, fmt.Errorf("failed to query items: %w", result.Error)
    }

    var items []Item
    for _, gormItem := range gormItems {
        items = append(items, ConvertFromGormItem(gormItem))
    }

    return items, nil
}

// GetItemByGoodwillID finds item by Goodwill ID
func (r *GormRepository) GetItemByGoodwillID(goodwillID string) (*Item, error) {
    gormDB, err := r.getDB()
    if err != nil {
        return nil, err
    }

    var gormItem GormItem
    result := gormDB.Where("goodwill_id = ?", goodwillID).First(&gormItem)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to query item: %w", result.Error)
    }

    item := ConvertFromGormItem(gormItem)
    return &item, nil
}
```

### Notification Operations

```go
// QueueNotification adds a new notification
func (r *GormRepository) QueueNotification(notification Notification) (int, error) {
    gormDB, err := r.getDB()
    if err != nil {
        return 0, err
    }

    gormNotification := ConvertToGormNotification(notification)
    result := gormDB.Create(&gormNotification)
    if result.Error != nil {
        return 0, fmt.Errorf("failed to insert notification: %w", result.Error)
    }

    return int(gormNotification.ID), nil
}

// GetPendingNotifications returns queued notifications
func (r *GormRepository) GetPendingNotifications() ([]Notification, error) {
    gormDB, err := r.getDB()
    if err != nil {
        return nil, err
    }

    var gormNotifications []GormNotification
    result := gormDB.Where("status = ?", "queued").
                   Order("created_at ASC").
                   Find(&gormNotifications)
    if result.Error != nil {
        return nil, fmt.Errorf("failed to query pending notifications: %w", result.Error)
    }

    var notifications []Notification
    for _, gormNotification := range gormNotifications {
        notifications = append(notifications, ConvertFromGormNotification(gormNotification))
    }

    return notifications, nil
}
```

## Usage Patterns

### Basic CRUD Operations

```go
// Create a new search
repo := NewGormRepository(db)
search := Search{Name: "Electronics", Query: "laptop OR tablet"}
id, err := repo.AddSearch(search)
if err != nil {
    // Handle error
}

// Get search by ID
search, err := repo.GetSearchByID(id)
if err != nil {
    // Handle error
}

// Update search
search.Query = "laptop OR tablet OR smartphone"
err = repo.UpdateSearch(*search)
if err != nil {
    // Handle error
}

// Delete search
err = repo.DeleteSearch(id)
if err != nil {
    // Handle error
}
```

### Complex Query Patterns

```go
// Get active searches with items
activeSearches, err := repo.GetActiveSearches()
if err != nil {
    // Handle error
}

// Get items by search ID
items, err := repo.GetItemsBySearchID(searchID)
if err != nil {
    // Handle error
}

// Get price history for item
priceHistory, err := repo.GetPriceHistory(itemID)
if err != nil {
    // Handle error
}
```

### Transaction Usage

```go
// Complex operation with transaction
err := repo.WithTransaction(func(tx *gorm.DB) error {
    // Create search
    gormSearch := ConvertToGormSearch(search)
    if err := tx.Create(&gormSearch).Error; err != nil {
        return fmt.Errorf("failed to create search: %w", err)
    }

    // Create search execution
    gormExecution := ConvertToGormSearchExecution(execution)
    gormExecution.SearchID = gormSearch.ID
    if err := tx.Create(&gormExecution).Error; err != nil {
        return fmt.Errorf("failed to create execution: %w", err)
    }

    // Update search last checked
    if err := tx.Model(&gormSearch).
               Update("last_checked", gormExecution.ExecutedAt).Error; err != nil {
        return fmt.Errorf("failed to update search: %w", err)
    }

    return nil
})

if err != nil {
    // Handle transaction error
}
```

## Error Handling

### Common Error Patterns

```go
// Handle record not found
search, err := repo.GetSearchByID(999)
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        // Handle not found case
    } else {
        // Handle other errors
    }
}

// Handle database connection errors
items, err := repo.GetItems()
if err != nil {
    if strings.Contains(err.Error(), "database not connected") {
        // Handle connection error
    } else {
        // Handle other errors
    }
}
```

### GORM-Specific Error Handling

```go
// Handle GORM-specific errors
func (r *GormRepository) handleGormError(result *gorm.DB, operation string) error {
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return fmt.Errorf("%s: record not found", operation)
        } else if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
            return fmt.Errorf("%s: duplicate key violation", operation)
        } else if errors.Is(result.Error, gorm.ErrInvalidTransaction) {
            return fmt.Errorf("%s: invalid transaction", operation)
        } else {
            return fmt.Errorf("%s: database error: %w", operation, result.Error)
        }
    }
    return nil
}
```

### Repository Error Handling

```go
// Standard error handling pattern
func (r *GormRepository) safeOperation(operation string, fn func() error) error {
    gormDB, err := r.getDB()
    if err != nil {
        return fmt.Errorf("%s: %w", operation, err)
    }

    if err := fn(); err != nil {
        return fmt.Errorf("%s: %w", operation, err)
    }

    return nil
}
```

## Transaction Patterns

### Basic Transaction Usage

```go
// Simple transaction example
func (r *GormRepository) CreateSearchWithExecution(search Search, execution SearchExecution) error {
    return r.db.WithTransaction(func(tx *gorm.DB) error {
        // Create search
        gormSearch := ConvertToGormSearch(search)
        if err := tx.Create(&gormSearch).Error; err != nil {
            return fmt.Errorf("failed to create search: %w", err)
        }

        // Create execution
        gormExecution := ConvertToGormSearchExecution(execution)
        gormExecution.SearchID = gormSearch.ID
        if err := tx.Create(&gormExecution).Error; err != nil {
            return fmt.Errorf("failed to create execution: %w", err)
        }

        return nil
    })
}
```

### Complex Transaction with Rollback

```go
// Complex transaction with error handling
func (r *GormRepository) ProcessNewItemWithHistory(item Item, priceHistory PriceHistory, notification Notification) error {
    return r.db.WithTransaction(func(tx *gorm.DB) error {
        // Add item
        gormItem := ConvertToGormItem(item)
        if err := tx.Create(&gormItem).Error; err != nil {
            return fmt.Errorf("failed to create item: %w", err)
        }

        // Add price history
        gormPriceHistory := ConvertToGormPriceHistory(priceHistory)
        gormPriceHistory.ItemID = gormItem.ID
        if err := tx.Create(&gormPriceHistory).Error; err != nil {
            return fmt.Errorf("failed to create price history: %w", err)
        }

        // Queue notification
        gormNotification := ConvertToGormNotification(notification)
        gormNotification.ItemID = gormItem.ID
        if err := tx.Create(&gormNotification).Error; err != nil {
            return fmt.Errorf("failed to create notification: %w", err)
        }

        // Update item with notification info
        if err := tx.Model(&gormItem).
                   Update("notification_status", "queued").Error; err != nil {
            return fmt.Errorf("failed to update item: %w", err)
        }

        return nil
    })
}
```

### Nested Transactions

```go
// Nested transaction example
func (r *GormRepository) ComplexOperation() error {
    return r.db.WithTransaction(func(tx *gorm.DB) error {
        // Outer transaction operations

        // Nested transaction (uses same underlying transaction)
        err := r.db.WithTransaction(func(nestedTx *gorm.DB) error {
            // Nested operations
            return nil
        })

        if err != nil {
            return err
        }

        // More outer operations
        return nil
    })
}
```

## Testing Patterns

### Test Setup

```go
// Setup test repository
func setupTestRepository() (*GormRepository, error) {
    config := &Config{
        Path: ":memory:",
        MaxConnections: 5,
    }

    db, err := NewGormDatabase(config)
    if err != nil {
        return nil, err
    }

    if err := db.Connect(); err != nil {
        return nil, err
    }

    // AutoMigrate test schema
    if err := db.AutoMigrate(GetAllModels()...); err != nil {
        return nil, err
    }

    return NewGormRepository(db), nil
}
```

### Repository Tests

```go
// Test CRUD operations
func TestGormRepositoryCRUD(t *testing.T) {
    repo, err := setupTestRepository()
    require.NoError(t, err)

    // Test Create
    search := Search{Name: "Test", Query: "test"}
    id, err := repo.AddSearch(search)
    require.NoError(t, err)
    assert.NotZero(t, id)

    // Test Read
    retrieved, err := repo.GetSearchByID(id)
    require.NoError(t, err)
    assert.Equal(t, "Test", retrieved.Name)

    // Test Update
    retrieved.Query = "updated"
    err = repo.UpdateSearch(*retrieved)
    require.NoError(t, err)

    // Test Delete
    err = repo.DeleteSearch(id)
    require.NoError(t, err)

    // Verify deletion
    _, err = repo.GetSearchByID(id)
    assert.Error(t, err)
}
```

### Transaction Tests

```go
// Test transaction rollback
func TestRepositoryTransactionRollback(t *testing.T) {
    repo, err := setupTestRepository()
    require.NoError(t, err)

    // This should fail and rollback
    err = repo.WithTransaction(func(tx *gorm.DB) error {
        // Create search
        search := GormSearch{Name: "Test", Query: "test"}
        if err := tx.Create(&search).Error; err != nil {
            return err
        }

        // This will fail (duplicate name)
        duplicate := GormSearch{Name: "Test", Query: "test"}
        if err := tx.Create(&duplicate).Error; err != nil {
            return err
        }

        return nil
    })

    // Should have error
    assert.Error(t, err)

    // Verify no search was created
    searches, err := repo.GetSearches()
    require.NoError(t, err)
    assert.Empty(t, searches)
}
```

### Error Handling Tests

```go
// Test error handling
func TestRepositoryErrorHandling(t *testing.T) {
    repo, err := setupTestRepository()
    require.NoError(t, err)

    // Test not found error
    _, err = repo.GetSearchByID(999)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")

    // Test database connection error
    // (Would need to mock disconnected database)
}
```

## Best Practices

### Repository Design

1. **Single Responsibility**: Each repository method should have a single purpose
2. **Consistent Error Handling**: Use consistent error handling patterns
3. **Proper Transaction Usage**: Use transactions for related operations
4. **Input Validation**: Validate inputs before database operations
5. **Proper Resource Management**: Ensure database connections are properly managed

### Query Optimization

1. **Use Selective Loading**: Load only needed fields with `Select()`
2. **Use Eager Loading**: Prevent N+1 queries with `Preload()`
3. **Use Batch Operations**: For bulk data operations
4. **Add Proper Indexes**: Ensure frequently queried fields are indexed
5. **Monitor Slow Queries**: Use GORM's slow query logging

### Error Handling

1. **Handle Specific Errors**: Check for specific GORM error types
2. **Provide Context**: Include operation context in error messages
3. **Use Error Wrapping**: Use `fmt.Errorf` with `%w` for error wrapping
4. **Log Errors Appropriately**: Log errors at appropriate levels
5. **Handle Connection Errors**: Properly handle database connection issues

### Transaction Usage

1. **Keep Transactions Short**: Avoid long-running transactions
2. **Group Related Operations**: Put related operations in same transaction
3. **Handle Rollbacks Properly**: Ensure proper rollback on errors
4. **Avoid Nested Transactions**: GORM doesn't support true nested transactions
5. **Monitor Transaction Performance**: Watch for transaction timeouts

### Testing

1. **Test All CRUD Operations**: Ensure all methods are tested
2. **Test Error Conditions**: Test error handling paths
3. **Test Transaction Behavior**: Verify transaction rollback works
4. **Test Performance**: Check query performance
5. **Test Data Integrity**: Verify data consistency

## Repository Interface Evolution

### Future Enhancements

1. **Async Operations**: Add async operation support
2. **Bulk Operations**: Add more bulk operation methods
3. **Advanced Querying**: Add more complex query methods
4. **Caching**: Add caching layer integration
5. **Metrics**: Add performance metrics collection

### Deprecation Strategy

1. **Mark Deprecated Methods**: Use comments to mark deprecated methods
2. **Maintain Backward Compatibility**: Keep old methods working
3. **Provide Migration Path**: Document how to migrate to new methods
4. **Eventual Removal**: Remove deprecated methods after grace period

## Conclusion

The GORM repository interface provides a clean, type-safe way to interact with the database, offering automatic schema management, ORM capabilities, and transaction support. This documentation covers all aspects of the repository interface, from basic CRUD operations to advanced transaction patterns and testing strategies.

For additional information, refer to:
- [GORM Official Documentation](https://gorm.io/docs/)
- [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) - Complete schema specification
- [GORM_USAGE_GUIDE.md](GORM_USAGE_GUIDE.md) - Usage patterns and examples
- [GORM_MIGRATION_GUIDE.md](GORM_MIGRATION_GUIDE.md) - Migration documentation