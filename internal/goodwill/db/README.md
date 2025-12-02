# Database Layer - GORM Implementation

## Overview

This directory contains the GORM-based database implementation for go-find-goodwill, providing ORM capabilities, automatic schema management, and transaction support.

## GORM Database Structure

### Key Components

1. **GORM Models**: Struct-based schema definitions with relationships
2. **GORM Database**: Connection management and AutoMigrate support
3. **GORM Repository**: CRUD operations with GORM optimizations
4. **GORM Migration**: Automatic schema migration and version tracking

### Files Overview

| File | Purpose |
|------|---------|
| `gorm_db.go` | GORM database connection and configuration |
| `gorm_models.go` | GORM model definitions with struct tags |
| `gorm_repository.go` | GORM repository implementation |
| `gorm_migration.go` | GORM migration manager |
| `repository.go` | Repository interface definition |

## GORM Models

### Core Models

- **GormSearch**: Search queries and configurations
- **GormSearchExecution**: Search execution history
- **GormItem**: Goodwill item listings
- **GormItemDetails**: Detailed item information
- **GormPriceHistory**: Item price history tracking
- **GormBidHistory**: Item bid history tracking
- **GormNotification**: Notification system
- **GormUserAgent**: Anti-bot user agent management
- **GormSystemLog**: System logging
- **GormSearchItemMapping**: Many-to-many search-item relationships

### Model Features

- **Struct Tags**: GORM uses tags for schema definition (`gorm:"primaryKey"`, `gorm:"size:255"`, etc.)
- **Automatic Timestamps**: `CreatedAt`, `UpdatedAt` fields are auto-managed
- **Relationships**: One-to-many and many-to-many relationships
- **Soft Delete**: Optional soft delete functionality
- **Validation**: Built-in validation support

## GORM Database Connection

### Configuration

```go
config := &Config{
    Path:             "database.db",
    MaxConnections:   10,
    ConnectionTimeout: 30 * time.Second,
    MigrationPath:    "internal/goodwill/db/migrations",
}

db, err := NewGormDatabase(config)
if err != nil {
    // Handle error
}

if err := db.Connect(); err != nil {
    // Handle error
}
```

### Connection Pooling

- **Max Open Connections**: Configurable (default: 10)
- **Max Idle Connections**: Half of max connections
- **Connection Lifetime**: 30 minutes
- **Idle Timeout**: 10 minutes

## GORM AutoMigrate

### Automatic Schema Management

```go
// AutoMigrate all models
models := GetAllModels()
if err := db.AutoMigrate(models...); err != nil {
    // Handle error
}
```

### Model Registration

```go
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

## GORM Repository

### Repository Usage

```go
// Create repository
repo := NewGormRepository(db)

// CRUD Operations
searches, err := repo.GetSearches()
search, err := repo.GetSearchByID(1)
id, err := repo.AddSearch(search)
err := repo.UpdateSearch(search)
err := repo.DeleteSearch(1)
```

### Transaction Support

```go
// Transaction example
err := repo.WithTransaction(func(tx *gorm.DB) error {
    // Perform operations within transaction
    if err := tx.Create(&search).Error; err != nil {
        return err
    }

    if err := tx.Create(&execution).Error; err != nil {
        return err
    }

    return nil
})
```

## GORM Migration System

### Migration Process

1. **AutoMigrate**: GORM automatically creates/updates tables
2. **Version Tracking**: Migration versions stored in `gorm_migrations` table
3. **Schema Evolution**: GORM handles schema changes automatically

### Migration Commands

```go
// Create migration manager
migrationManager := NewGormMigrationManager(db)

// Run migrations
if err := migrationManager.Migrate(); err != nil {
    // Handle error
}

// Rollback last migration
if err := migrationManager.Rollback(); err != nil {
    // Handle error
}
```

## Testing

### Test Setup

```go
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
func TestGormRepository(t *testing.T) {
    db, err := setupTestDB()
    require.NoError(t, err)
    defer db.Close()

    repo := NewGormRepository(db)

    // Test CRUD operations
    // ...
}
```

## Performance Optimization

### Connection Management

- Use connection pooling for better performance
- Configure appropriate connection limits
- Monitor and tune connection settings

### Query Optimization

- Use `Select()` to load only needed fields
- Use `Preload()` for relationship loading
- Use batch operations for bulk data
- Add proper indexes for frequently queried fields

### Transaction Usage

- Group related operations in transactions
- Keep transactions short and focused
- Handle errors and rollbacks properly
- Avoid long-running transactions

## Best Practices

### Model Design

1. Use appropriate GORM tags for schema definition
2. Define relationships explicitly
3. Use pointers for nullable fields
4. Leverage automatic timestamp fields

### Query Design

1. Use `Preload()` to avoid N+1 queries
2. Limit query scope with `Select()`
3. Use transactions for related operations
4. Handle GORM-specific errors properly

### Migration

1. Test migrations thoroughly
2. Backup database before migrations
3. Monitor migration progress
4. Handle schema changes carefully

## Documentation

For comprehensive GORM usage documentation, see:

- [DATABASE_SCHEMA.md](../../../../DATABASE_SCHEMA.md) - GORM schema specification
- [GORM_USAGE_GUIDE.md](../../../../GORM_USAGE_GUIDE.md) - Complete usage guide
- [GORM Official Documentation](https://gorm.io/docs/) - Official GORM docs

## Migration from Legacy System

The database layer has completed its transition from file-based SQL migrations to GORM AutoMigrate:

**Before (Legacy):**
- Manual SQL files in `migrations/`
- Explicit `.up.sql` and `.down.sql` files
- Manual version tracking

**After (GORM):**
- Schema defined in Go structs
- Automatic table creation/updates
- Automatic version tracking
- Schema evolution handled by GORM

The migration to GORM is now complete. All database operations use GORM exclusively with no legacy code remaining.