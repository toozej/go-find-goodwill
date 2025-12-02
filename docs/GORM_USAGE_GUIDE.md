# GORM Usage Guide for go-find-goodwill

## Introduction

This guide provides comprehensive documentation for working with GORM in the go-find-goodwill application, covering model relationships, AutoMigrate usage, transaction patterns, and best practices.

## Table of Contents

1. [GORM Model Definitions](#gorm-model-definitions)
2. [Model Relationships](#model-relationships)
3. [AutoMigrate Usage](#automigrate-usage)
4. [Transaction Patterns](#transaction-patterns)
5. [Repository Interface](#repository-interface)
6. [Query Patterns](#query-patterns)
7. [Error Handling](#error-handling)
8. [Performance Optimization](#performance-optimization)
9. [Migration Patterns](#migration-patterns)
10. [Testing with GORM](#testing-with-gorm)

## GORM Model Definitions

### Struct Tags

GORM uses struct tags to define database schema:

```go
type GormSearch struct {
    ID    uint   `gorm:"primaryKey"`          // Primary key
    Name  string `gorm:"size:255;not null"`   // VARCHAR(255) NOT NULL
    Query string `gorm:"size:500;not null"`  // VARCHAR(500) NOT NULL
    Enabled bool   `gorm:"default:true"`      // BOOLEAN DEFAULT TRUE
}
```

### Common GORM Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `primaryKey` | Marks field as primary key | `gorm:"primaryKey"` |
| `size` | Sets column size | `gorm:"size:255"` |
| `not null` | Sets NOT NULL constraint | `gorm:"not null"` |
| `default` | Sets default value | `gorm:"default:true"` |
| `unique` | Sets UNIQUE constraint | `gorm:"unique"` |
| `type` | Sets column type | `gorm:"type:text"` |
| `autoCreateTime` | Auto-set on creation | `gorm:"autoCreateTime"` |
| `autoUpdateTime` | Auto-set on update | `gorm:"autoUpdateTime"` |

### Automatic Fields

GORM automatically manages these fields:
- `CreatedAt`: Set on record creation
- `UpdatedAt`: Updated on record update
- `DeletedAt`: Used for soft delete functionality

## Model Relationships

### One-to-Many Relationships

```go
type GormSearch struct {
    ID                  uint   `gorm:"primaryKey"`
    SearchExecutions    []GormSearchExecution `gorm:"foreignKey:SearchID"`
    SearchItemMappings  []GormSearchItemMapping `gorm:"foreignKey:SearchID"`
}

type GormSearchExecution struct {
    ID       uint   `gorm:"primaryKey"`
    SearchID uint   `gorm:"not null"`
    Search   GormSearch `gorm:"foreignKey:SearchID"`
}
```

### Many-to-Many Relationships

```go
type GormSearchItemMapping struct {
    ID       uint `gorm:"primaryKey"`
    SearchID uint `gorm:"not null"`
    ItemID   uint `gorm:"not null"`

    // Relationships
    Search GormSearch `gorm:"foreignKey:SearchID"`
    Item   GormItem   `gorm:"foreignKey:ItemID"`
}
```

### Relationship Query Patterns

#### Eager Loading with Preload

```go
// Load search with its executions
var search GormSearch
result := db.Preload("SearchExecutions").First(&search, 1)

// Load item with all related data
var item GormItem
result := db.Preload("PriceHistories").
           Preload("BidHistories").
           Preload("Notifications").
           First(&item, 1)
```

#### Conditional Relationship Loading

```go
// Load only recent price history
var item GormItem
result := db.Preload("PriceHistories", "recorded_at > ?", time.Now().Add(-7*24*time.Hour)).
           First(&item, 1)
```

## AutoMigrate Usage

### Basic AutoMigrate

```go
// AutoMigrate all models
models := GetAllModels()
if err := db.AutoMigrate(models...); err != nil {
    return fmt.Errorf("failed to auto-migrate: %w", err)
}
```

### Model Registration

```go
// GetAllModels returns all GORM models for AutoMigrate
func GetAllModels() []interface{} {
    return []interface{}{
        &GormSearch{},
        &GormSearchExecution{},
        &GormItem{},
        &GormItemDetails{},
        &GormPriceHistory{},
        &GormBidHistory{},
        &GormNotification{},
        &GormUserAgent{},
        &GormSystemLog{},
        &GormSearchItemMapping{},
        &GormMigration{},
    }
}
```

### Migration Management

```go
// Migrate runs GORM AutoMigrate for all models and tracks versions
func (m *GormMigrationManager) Migrate() error {
    // Get current version
    currentVersion, err := m.GetCurrentVersion()
    if err != nil {
        return fmt.Errorf("failed to get current version: %w", err)
    }

    // Run AutoMigrate for all models
    models := GetAllModels()
    if err := m.db.AutoMigrate(models...); err != nil {
        return fmt.Errorf("failed to auto-migrate models: %w", err)
    }

    // Record new migrations
    for _, migration := range m.migrations {
        if migration.Version <= currentVersion {
            continue
        }

        // Record migration in database
        gormMigration := GormMigration{
            Version: migration.Version,
            Name:    migration.Name,
        }

        if err := m.db.GetDB().Create(&gormMigration).Error; err != nil {
            return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
        }
    }

    return nil
}
```

## Transaction Patterns

### Basic Transaction Usage

```go
// Simple transaction example
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Error; err != nil {
    return err
}

// Perform operations
if err := tx.Create(&search).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&execution).Error; err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error
```

### Repository Transaction Helper

```go
// WithTransaction executes a function within a transaction
func (d *GormDatabase) WithTransaction(fn func(tx *gorm.DB) error) error {
    tx, err := d.BeginTransaction()
    if err != nil {
        return err
    }

    defer func() {
        if p := recover(); p != nil {
            if rbErr := tx.Rollback(); rbErr != nil {
                log.Errorf("Transaction panicked and failed to rollback: %v", rbErr)
            }
            panic(p)
        }
    }()

    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            log.Errorf("Transaction failed and rollback error: %v", rbErr)
        }
        return fmt.Errorf("transaction failed: %w", err)
    }

    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

### Transaction Usage Example

```go
// Complex operation requiring transaction
func (r *GormRepository) CreateSearchWithInitialExecution(search Search, execution SearchExecution) error {
    return r.db.WithTransaction(func(tx *gorm.DB) error {
        // Create search
        gormSearch := ConvertToGormSearch(search)
        if err := tx.Create(&gormSearch).Error; err != nil {
            return fmt.Errorf("failed to create search: %w", err)
        }

        // Create execution with search ID
        gormExecution := ConvertToGormSearchExecution(execution)
        gormExecution.SearchID = gormSearch.ID
        if err := tx.Create(&gormExecution).Error; err != nil {
            return fmt.Errorf("failed to create execution: %w", err)
        }

        // Update search last checked time
        if err := tx.Model(&gormSearch).
                   Update("last_checked", gormExecution.ExecutedAt).Error; err != nil {
            return fmt.Errorf("failed to update search: %w", err)
        }

        return nil
    })
}
```

## Repository Interface

### GORM Repository Structure

```go
type GormRepository struct {
    db *GormDatabase
}

func NewGormRepository(db *GormDatabase) *GormRepository {
    return &GormRepository{
        db: db,
    }
}
```

### Common Repository Methods

```go
// CRUD Operations
func (r *GormRepository) GetSearches() ([]Search, error)
func (r *GormRepository) GetSearchByID(id int) (*Search, error)
func (r *GormRepository) AddSearch(search Search) (int, error)
func (r *GormRepository) UpdateSearch(search Search) error
func (r *GormRepository) DeleteSearch(id int) error

// Item Operations
func (r *GormRepository) GetItems() ([]Item, error)
func (r *GormRepository) GetItemByID(id int) (*Item, error)
func (r *GormRepository) AddItem(item Item) (int, error)
func (r *GormRepository) UpdateItem(item Item) error
```

### Conversion Functions

```go
// Convert between legacy and GORM models
func ConvertToGormSearch(search Search) GormSearch
func ConvertFromGormSearch(gormSearch GormSearch) Search
func ConvertToGormItem(item Item) GormItem
func ConvertFromGormItem(gormItem GormItem) Item
```

## Query Patterns

### Basic CRUD Queries

```go
// Create
gormSearch := GormSearch{Name: "Test", Query: "test"}
result := db.Create(&gormSearch)

// Read
var search GormSearch
result := db.First(&search, 1)

// Update
search.Query = "updated"
result := db.Save(&search)

// Delete
result := db.Delete(&search)
```

### Advanced Query Patterns

```go
// Find with conditions
var searches []GormSearch
result := db.Where("enabled = ?", true).
             Order("name ASC").
             Find(&searches)

// Find with OR conditions
var items []GormItem
result := db.Where("status = ?", "active").
             Or("status = ?", "pending").
             Find(&items)

// Find with IN clause
var items []GormItem
result := db.Where("category IN ?", []string{"Electronics", "Furniture"}).
             Find(&items)

// Find with LIKE
var searches []GormSearch
result := db.Where("name LIKE ?", "%camera%").
             Find(&searches)
```

### Relationship Queries

```go
// Find searches with their executions
var searches []GormSearch
result := db.Preload("SearchExecutions").
             Where("enabled = ?", true).
             Find(&searches)

// Find items with price history
var items []GormItem
result := db.Preload("PriceHistories", func(db *gorm.DB) *gorm.DB {
    return db.Order("recorded_at DESC").Limit(5)
}).
Find(&items)
```

### Joins and Complex Queries

```go
// Join query example
var items []GormItem
result := db.Joins("JOIN gorm_search_item_mappings ON gorm_search_item_mappings.item_id = gorm_items.id").
           Where("gorm_search_item_mappings.search_id = ?", searchID).
           Find(&items)

// Subquery example
subQuery := db.Select("item_id").Where("search_id = ?", searchID).Table("gorm_search_item_mappings")
result := db.Where("id IN (?)", subQuery).Find(&items)
```

## Error Handling

### Common GORM Errors

```go
// Handle record not found
var search GormSearch
result := db.First(&search, 1)
if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    // Handle not found case
}

// Handle database errors
if result.Error != nil {
    return fmt.Errorf("database operation failed: %w", result.Error)
}

// Check for specific error types
if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
    // Handle duplicate key error
}
```

### Repository Error Handling

```go
func (r *GormRepository) GetSearchByID(id int) (*Search, error) {
    gormDB := r.db.GetDB()
    if gormDB == nil {
        return nil, errors.New("database not connected")
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

## Performance Optimization

### Connection Pooling

```go
// Configure connection pool
sqlDB, err := gormDB.DB()
if err != nil {
    return err
}

sqlDB.SetMaxOpenConns(25)      // Maximum open connections
sqlDB.SetMaxIdleConns(10)      // Maximum idle connections
sqlDB.SetConnMaxLifetime(30 * time.Minute)  // Connection lifetime
sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Idle connection timeout
```

### Batch Operations

```go
// Batch insert
items := []GormItem{/* ... */}
result := db.Create(&items)

// Batch update
result := db.Model(&GormItem{}).
           Where("status = ?", "active").
           Updates(map[string]interface{}{
               "last_seen": time.Now(),
               "updated_at": time.Now(),
           })
```

### Index Optimization

```go
// Add indexes using GORM tags
type GormItem struct {
    GoodwillID string `gorm:"size:100;unique;not null"`  // Unique index
    Status     string `gorm:"size:50;index"`            // Regular index
    EndsAt     *time.Time `gorm:"index"`                // Index for date filtering
}

// Composite index example
type GormSearchItemMapping struct {
    SearchID uint `gorm:"not null;index:idx_search_item,unique"`
    ItemID   uint `gorm:"not null;index:idx_search_item,unique"`
}
```

### Query Optimization

```go
// Use Select to load only needed fields
var item GormItem
result := db.Select("id", "goodwill_id", "title", "current_price").
           First(&item, 1)

// Use Limit and Offset for pagination
var items []GormItem
result := db.Limit(50).Offset(100).Find(&items)

// Use Distinct for unique results
var categories []string
result := db.Model(&GormItem{}).
           Distinct("category").
           Find(&categories)
```

## Migration Patterns

### Migration Manager

```go
// GormMigrationManager handles database migrations
type GormMigrationManager struct {
    db         *GormDatabase
    migrations []GormMigration
}

// Migrate runs AutoMigrate and tracks versions
func (m *GormMigrationManager) Migrate() error {
    // Get current version
    currentVersion, err := m.GetCurrentVersion()
    if err != nil {
        return err
    }

    // AutoMigrate all models
    models := GetAllModels()
    if err := m.db.AutoMigrate(models...); err != nil {
        return err
    }

    // Record new migrations
    for _, migration := range m.migrations {
        if migration.Version > currentVersion {
            if err := m.recordMigration(migration); err != nil {
                return err
            }
        }
    }

    return nil
}
```

### Legacy Migration Import

```go
// ImportLegacyMigrations imports from legacy system
func (m *GormMigrationManager) ImportLegacyMigrations(legacyDB *Database) error {
    // Query legacy migrations
    var legacyMigrations []Migration
    rows, err := legacyDB.GetDB().Query("SELECT version, name, applied_at FROM migrations ORDER BY version ASC")
    if err != nil {
        return err
    }
    defer rows.Close()

    // Import each migration
    for rows.Next() {
        var migration Migration
        // ... scan and process

        gormMigration := GormMigration{
            Version:   migration.Version,
            Name:      migration.Name,
            AppliedAt: migration.AppliedAt,
        }

        if err := m.db.GetDB().Create(&gormMigration).Error; err != nil {
            return err
        }
    }

    return nil
}
```

## Testing with GORM

### Test Setup

```go
// Setup test database
func setupTestDB() (*GormDatabase, error) {
    config := &Config{
        Path:             ":memory:",
        MaxConnections:   5,
        ConnectionTimeout: 5 * time.Second,
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

    return db, nil
}
```

### Test Patterns

```go
// Test CRUD operations
func TestGormSearchCRUD(t *testing.T) {
    db, err := setupTestDB()
    require.NoError(t, err)
    defer db.Close()

    repo := NewGormRepository(db)

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
}
```

### Transaction Testing

```go
// Test transaction rollback
func TestTransactionRollback(t *testing.T) {
    db, err := setupTestDB()
    require.NoError(t, err)
    defer db.Close()

    repo := NewGormRepository(db)

    // This should fail and rollback
    err = repo.WithTransaction(func(tx *gorm.DB) error {
        // Create search
        search := GormSearch{Name: "Test", Query: "test"}
        if err := tx.Create(&search).Error; err != nil {
            return err
        }

        // This will fail (duplicate)
        duplicate := GormSearch{Name: "Test", Query: "test"}
        if err := tx.Create(&duplicate).Error; err != nil {
            return err
        }

        return nil
    })

    // Should have error
    assert.Error(t, err)

    // Verify no search was created
    var count int64
    err = db.GetDB().Model(&GormSearch{}).Count(&count).Error
    require.NoError(t, err)
    assert.Equal(t, int64(0), count)
}
```

## Best Practices

### Model Design

1. **Use Proper Tags**: Always specify appropriate GORM tags for schema definition
2. **Define Relationships**: Explicitly define relationships between models
3. **Use Pointers for Nullable Fields**: Use `*time.Time`, `*float64` for nullable columns
4. **Leverage Automatic Fields**: Use `CreatedAt`, `UpdatedAt` for automatic timestamping

### Query Design

1. **Use Preload for Relationships**: Avoid N+1 query problems with `Preload()`
2. **Limit Query Scope**: Use `Select()` to load only needed fields
3. **Use Transactions**: Group related operations in transactions
4. **Handle Errors Properly**: Check for specific GORM error types

### Performance

1. **Optimize Connection Pool**: Configure appropriate connection pool settings
2. **Use Batch Operations**: Prefer batch inserts/updates for bulk operations
3. **Add Proper Indexes**: Ensure frequently queried fields are indexed
4. **Monitor Slow Queries**: Use GORM's slow query logging

### Migration

1. **Test Migrations**: Always test migrations on staging environments first
2. **Backup Before Migration**: Ensure database backups before running migrations
3. **Monitor Migration Process**: Log migration progress and errors
4. **Handle Schema Changes Carefully**: Plan for data migration when changing schemas

This comprehensive GORM usage guide provides developers with the knowledge needed to effectively work with the GORM implementation in go-find-goodwill, covering all aspects from basic CRUD operations to advanced transaction patterns and performance optimization.