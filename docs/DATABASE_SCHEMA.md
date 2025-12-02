# go-find-goodwill Database Schema Specification (GORM Edition)

## Overview

This document provides the detailed GORM-based database schema for the go-find-goodwill application, including GORM model definitions, relationships, AutoMigrate usage, and transaction patterns.

## GORM Database Structure

The application has transitioned from file-based SQL migrations to GORM's AutoMigrate system, providing automatic schema management and ORM capabilities.

### GORM Models

The database schema is now defined using GORM structs with appropriate tags for schema definition:

#### 1. GormSearch Model
```go
type GormSearch struct {
    ID                        uint   `gorm:"primaryKey"`
    Name                      string `gorm:"size:255;not null"`
    Query                     string `gorm:"size:500;not null"`
    RegexPattern              string `gorm:"size:500"`
    Enabled                   bool   `gorm:"default:true"`
    CreatedAt                 time.Time
    UpdatedAt                 time.Time
    LastChecked               *time.Time
    NotificationThresholdDays int `gorm:"default:7"`
    MinPrice                  *float64
    MaxPrice                  *float64
    CategoryFilter            string `gorm:"size:100"`
    SellerFilter              string `gorm:"size:100"`
    ShippingFilter            string `gorm:"size:100"`
    ConditionFilter           string `gorm:"size:100"`
    SortBy                    string `gorm:"size:50"`

    // Relationships
    SearchExecutions   []GormSearchExecution   `gorm:"foreignKey:SearchID"`
    SearchItemMappings []GormSearchItemMapping `gorm:"foreignKey:SearchID"`
}
```

#### 2. GormSearchExecution Model
```go
type GormSearchExecution struct {
    ID            uint `gorm:"primaryKey"`
    SearchID      uint `gorm:"not null"`
    ExecutedAt    time.Time
    Status        string `gorm:"size:50;not null"`
    ItemsFound    int
    NewItemsFound int
    ErrorMessage  string `gorm:"size:1000"`
    DurationMS    int

    // Relationships
    Search GormSearch `gorm:"foreignKey:SearchID"`
}
```

#### 3. GormItem Model
```go
type GormItem struct {
    ID              uint    `gorm:"primaryKey"`
    GoodwillID      string  `gorm:"size:100;unique;not null"`
    Title           string  `gorm:"size:500;not null"`
    Seller          string  `gorm:"size:100"`
    CurrentPrice    float64 `gorm:"not null"`
    BuyNowPrice     *float64
    URL             string `gorm:"size:500;not null"`
    ImageURL        string `gorm:"size:500"`
    EndsAt          *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
    FirstSeen       time.Time
    LastSeen        time.Time
    Status          string `gorm:"size:50;default:'active'"`
    Category        string `gorm:"size:100"`
    Subcategory     string `gorm:"size:100"`
    Condition       string `gorm:"size:100"`
    ShippingCost    *float64
    ShippingMethod  string `gorm:"size:100"`
    Description     string `gorm:"type:text"`
    Location        string `gorm:"size:200"`
    PickupAvailable bool   `gorm:"default:false"`
    ReturnsAccepted bool   `gorm:"default:false"`
    WatchCount      int    `gorm:"default:0"`
    BidCount        int    `gorm:"default:0"`
    ViewCount       int    `gorm:"default:0"`

    // Relationships
    PriceHistories     []GormPriceHistory      `gorm:"foreignKey:ItemID"`
    BidHistories       []GormBidHistory        `gorm:"foreignKey:ItemID"`
    Notifications      []GormNotification      `gorm:"foreignKey:ItemID"`
    SearchItemMappings []GormSearchItemMapping `gorm:"foreignKey:ItemID"`
}
```

#### 4. GormItemDetails Model
```go
type GormItemDetails struct {
    ID             uint   `gorm:"primaryKey"`
    ItemID         uint   `gorm:"not null"`
    Description    string `gorm:"type:text"`
    Condition      string `gorm:"size:100"`
    ShippingCost   *float64
    ShippingMethod string `gorm:"size:100"`
    Category       string `gorm:"size:100"`
    Subcategory    string `gorm:"size:100"`
    Dimensions     string `gorm:"size:100"`
    Weight         string `gorm:"size:50"`
    Material       string `gorm:"size:100"`
    Color          string `gorm:"size:50"`
    Brand          string `gorm:"size:100"`
    Model          string `gorm:"size:100"`
    Year           *int

    // Relationship
    Item GormItem `gorm:"foreignKey:ItemID"`
}
```

#### 5. GormPriceHistory Model
```go
type GormPriceHistory struct {
    ID         uint    `gorm:"primaryKey"`
    ItemID     uint    `gorm:"not null"`
    Price      float64 `gorm:"not null"`
    PriceType  string  `gorm:"size:50;not null"`
    RecordedAt time.Time

    // Relationship
    Item GormItem `gorm:"foreignKey:ItemID"`
}
```

#### 6. GormBidHistory Model
```go
type GormBidHistory struct {
    ID         uint    `gorm:"primaryKey"`
    ItemID     uint    `gorm:"not null"`
    BidAmount  float64 `gorm:"not null"`
    Bidder     string  `gorm:"size:100"`
    BidderID   string  `gorm:"size:100"`
    RecordedAt time.Time

    // Relationship
    Item GormItem `gorm:"foreignKey:ItemID"`
}
```

#### 7. GormNotification Model
```go
type GormNotification struct {
    ID               uint   `gorm:"primaryKey"`
    ItemID           uint   `gorm:"not null"`
    SearchID         uint   `gorm:"not null"`
    NotificationType string `gorm:"size:100;not null"`
    Status           string `gorm:"size:50;default:'queued'"`
    SentAt           *time.Time
    DeliveredAt      *time.Time
    ErrorMessage     string `gorm:"size:1000"`
    RetryCount       int    `gorm:"default:0"`
    CreatedAt        time.Time
    UpdatedAt        time.Time

    // Relationships
    Item   GormItem   `gorm:"foreignKey:ItemID"`
    Search GormSearch `gorm:"foreignKey:SearchID"`
}
```

#### 8. GormUserAgent Model
```go
type GormUserAgent struct {
    ID         uint   `gorm:"primaryKey"`
    UserAgent  string `gorm:"size:500;not null"`
    LastUsed   *time.Time
    UsageCount int  `gorm:"default:0"`
    IsActive   bool `gorm:"default:true"`
}
```

#### 9. GormSystemLog Model
```go
type GormSystemLog struct {
    ID         uint `gorm:"primaryKey"`
    Timestamp  time.Time
    Level      string `gorm:"size:20;not null"`
    Component  string `gorm:"size:100;not null"`
    Message    string `gorm:"size:1000;not null"`
    Details    string `gorm:"type:text"`
    StackTrace string `gorm:"type:text"`
}
```

#### 10. GormSearchItemMapping Model
```go
type GormSearchItemMapping struct {
    ID       uint `gorm:"primaryKey"`
    SearchID uint `gorm:"not null"`
    ItemID   uint `gorm:"not null"`
    FoundAt  time.Time

    // Relationships
    Search GormSearch `gorm:"foreignKey:SearchID"`
    Item   GormItem   `gorm:"foreignKey:ItemID"`
}
```

#### 11. GormMigration Model
```go
type GormMigration struct {
    ID        uint      `gorm:"primaryKey"`
    Version   int       `gorm:"unique;not null"`
    Name      string    `gorm:"size:255;not null"`
    AppliedAt time.Time `gorm:"autoCreateTime"`
}
```

## GORM Model Relationships

The GORM implementation defines the following relationships:

### One-to-Many Relationships
- **GormSearch → GormSearchExecution**: One search can have many executions
- **GormSearch → GormSearchItemMapping**: One search can map to many items
- **GormItem → GormPriceHistory**: One item can have many price history entries
- **GormItem → GormBidHistory**: One item can have many bid history entries
- **GormItem → GormNotification**: One item can have many notifications
- **GormItem → GormSearchItemMapping**: One item can be mapped to many searches

### Many-to-Many Relationships
- **GormSearch ↔ GormItem**: Many-to-many relationship through GormSearchItemMapping

## GORM AutoMigrate Usage

### AutoMigrate Function
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

// AutoMigrate runs GORM's AutoMigrate for all models
func (d *GormDatabase) AutoMigrate(models ...interface{}) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    if d.db == nil {
        return errors.New("database not connected")
    }

    if err := d.db.AutoMigrate(models...); err != nil {
        return fmt.Errorf("failed to auto-migrate models: %w", err)
    }

    log.Info("Successfully auto-migrated database models")
    return nil
}
```

### Migration Process
1. **Initialization**: Create database file if not exists
2. **AutoMigrate**: GORM automatically creates/updates tables based on model definitions
3. **Version Tracking**: Migration versions are tracked in the `gorm_migrations` table
4. **Schema Evolution**: GORM handles schema changes automatically

## GORM Transaction Patterns

### Basic Transaction Usage
```go
// BeginTransaction starts a new transaction
func (d *GormDatabase) BeginTransaction() (*gorm.DB, error) {
    d.mu.Lock()
    defer d.mu.Unlock()

    if d.db == nil {
        return nil, errors.New("database not connected")
    }

    tx := d.db.Begin()
    if tx.Error != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
    }

    return tx, nil
}

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
// Example of using transactions for related operations
func (r *GormRepository) AddItemWithHistory(item Item, priceHistory PriceHistory) error {
    return r.db.WithTransaction(func(tx *gorm.DB) error {
        // Add item
        gormItem := ConvertToGormItem(item)
        if err := tx.Create(&gormItem).Error; err != nil {
            return fmt.Errorf("failed to create item: %w", err)
        }

        // Add price history with the same transaction
        gormPriceHistory := ConvertToGormPriceHistory(priceHistory)
        gormPriceHistory.ItemID = gormItem.ID
        if err := tx.Create(&gormPriceHistory).Error; err != nil {
            return fmt.Errorf("failed to create price history: %w", err)
        }

        return nil
    })
}
```

## GORM Repository Interface

The GORM repository implements the standard repository interface with GORM-specific optimizations:

```go
type GormRepository struct {
    db *GormDatabase
}

// Repository methods with GORM implementation
func (r *GormRepository) GetSearches() ([]Search, error)
func (r *GormRepository) GetSearchByID(id int) (*Search, error)
func (r *GormRepository) AddSearch(search Search) (int, error)
// ... other repository methods
```

### Key GORM Features Used

1. **Struct Tags**: GORM uses struct tags for schema definition (`gorm:"primaryKey"`, `gorm:"size:255"`, etc.)
2. **Automatic Timestamps**: `CreatedAt`, `UpdatedAt` fields are automatically managed
3. **Relationships**: One-to-many and many-to-many relationships are defined using struct fields
4. **Soft Delete**: GORM supports soft delete functionality
5. **Query Building**: GORM provides a fluent query API
6. **Error Handling**: GORM-specific error handling (e.g., `gorm.ErrRecordNotFound`)

## Migration from File-Based to GORM AutoMigrate

### Transition Overview

The application has migrated from a file-based SQL migration system to GORM's AutoMigrate:

**Before (File-Based):**
- Manual SQL files in `internal/goodwill/db/migrations/`
- Explicit `.up.sql` and `.down.sql` files
- Manual version tracking
- Requires manual schema updates

**After (GORM AutoMigrate):**
- Schema defined in Go structs with GORM tags
- Automatic table creation and updates
- Automatic version tracking
- Schema evolution handled by GORM

### Migration Compatibility

The GORM implementation includes conversion functions to maintain compatibility:

```go
// Conversion functions between legacy and GORM models
func ConvertToGormSearch(search Search) GormSearch
func ConvertFromGormSearch(gormSearch GormSearch) Search
func ConvertToGormItem(item Item) GormItem
func ConvertFromGormItem(gormItem GormItem) Item
// ... other conversion functions
```

### Legacy Migration Import

The system supports importing legacy migrations:

```go
// ImportLegacyMigrations imports migrations from the legacy system
func (m *GormMigrationManager) ImportLegacyMigrations(legacyDB *Database) error {
    // Query legacy migrations and import them into GORM system
}
```

## GORM Database Configuration

### Connection Setup
```go
// NewGormDatabase creates a new GORM database instance
func NewGormDatabase(config *Config) (*GormDatabase, error)

// Connect establishes a connection to the SQLite database using GORM
func (d *GormDatabase) Connect() error
```

### Configuration Options
- **Connection Pooling**: Configurable max connections, idle connections
- **Timeout Settings**: Connection timeout, query timeout
- **Logging**: GORM logger with configurable log levels
- **Migration Path**: Configurable migration file locations

## Performance Considerations

### GORM-Specific Optimizations

1. **Connection Pooling**: Properly configured connection pools
2. **Batch Operations**: GORM supports batch inserts and updates
3. **Eager Loading**: Use GORM's `Preload()` for relationship loading
4. **Indexing**: GORM automatically creates indexes based on model definitions
5. **Query Optimization**: GORM provides query building with proper WHERE, JOIN, and ORDER BY support

### Best Practices

1. **Use Transactions**: For related operations that should succeed/fail together
2. **Batch Operations**: For bulk data operations
3. **Proper Error Handling**: Handle GORM-specific errors appropriately
4. **Connection Management**: Properly manage database connections
5. **Logging**: Utilize GORM's logging capabilities for debugging

## Backup and Recovery

### GORM Database Backup Strategy

1. **Automatic Backups**: Regular database backups using SQLite tools
2. **Schema Export**: Export GORM models for schema documentation
3. **Data Export**: Export data using GORM query capabilities
4. **Recovery Process**: Restore from backup and let GORM AutoMigrate handle schema updates

## GORM Usage Examples

### Basic CRUD Operations
```go
// Create
gormSearch := GormSearch{Name: "Test Search", Query: "test query"}
result := db.Create(&gormSearch)

// Read
var search GormSearch
result := db.First(&search, 1)

// Update
search.Query = "updated query"
result := db.Save(&search)

// Delete
result := db.Delete(&search)
```

### Query Examples
```go
// Find with conditions
var searches []GormSearch
result := db.Where("enabled = ?", true).Find(&searches)

// Find with relationships
var search GormSearch
result := db.Preload("SearchExecutions").First(&search, 1)

// Complex queries
var items []GormItem
result := db.Where("status = ?", "active").
           Order("ends_at ASC").
           Limit(100).
           Find(&items)
```

This comprehensive GORM-based database schema provides the foundation for all data storage and retrieval operations in the go-find-goodwill application, offering automatic schema management, ORM capabilities, and improved developer productivity.