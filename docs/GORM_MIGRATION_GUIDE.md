# GORM Migration Guide for go-find-goodwill

## Introduction

This guide documents the transition from file-based SQL migrations to GORM's AutoMigrate system, explaining the migration process, benefits, and usage patterns.

## Table of Contents

1. [Migration Overview](#migration-overview)
2. [Before and After Comparison](#before-and-after-comparison)
3. [Migration Process](#migration-process)
4. [GORM AutoMigrate System](#gorm-automigrate-system)
5. [Legacy Compatibility](#legacy-compatibility)
6. [Migration Best Practices](#migration-best-practices)
7. [Troubleshooting](#troubleshooting)
8. [Rollback Procedures](#rollback-procedures)

## Migration Overview

### What Changed

The go-find-goodwill application has migrated from:

**File-Based SQL Migrations:**
- Manual SQL files (`*.up.sql`, `*.down.sql`)
- Explicit version tracking in migrations table
- Manual schema updates requiring SQL knowledge
- Separate migration execution process

**To GORM AutoMigrate:**
- Schema defined in Go structs with GORM tags
- Automatic table creation and updates
- Automatic version tracking
- Schema evolution handled by GORM
- Integrated with application code

### Benefits of GORM Migration

1. **Automatic Schema Management**: GORM handles table creation and updates
2. **Type Safety**: Schema defined in Go code with compile-time checks
3. **ORM Capabilities**: Full object-relational mapping support
4. **Relationship Management**: Automatic handling of relationships
5. **Improved Developer Experience**: No need to write SQL for common operations
6. **Better Error Handling**: GORM-specific error types and handling
7. **Transaction Support**: Built-in transaction management

## Before and After Comparison

### Schema Definition

**Before (SQL Files):**
```sql
-- migrations/001_init_schema.up.sql
CREATE TABLE searches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    query TEXT NOT NULL,
    enabled BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**After (GORM Models):**
```go
type GormSearch struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:255;not null"`
    Query     string    `gorm:"size:500;not null"`
    Enabled   bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Migration Execution

**Before (Manual Process):**
```go
// Load and execute SQL files manually
files, err := ioutil.ReadDir(migrationsPath)
for _, file := range files {
    if strings.HasSuffix(file.Name(), ".up.sql") {
        // Execute SQL file
    }
}
```

**After (AutoMigrate):**
```go
// Automatic migration
models := GetAllModels()
if err := db.AutoMigrate(models...); err != nil {
    return err
}
```

### Query Patterns

**Before (Raw SQL):**
```go
rows, err := db.Query("SELECT id, name, query FROM searches WHERE enabled = 1")
defer rows.Close()

var searches []Search
for rows.Next() {
    var s Search
    err := rows.Scan(&s.ID, &s.Name, &s.Query)
    // ...
}
```

**After (GORM ORM):**
```go
var searches []GormSearch
result := db.Where("enabled = ?", true).Find(&searches)
```

## Migration Process

### Step-by-Step Migration

1. **Backup Existing Database**
   ```bash
   # Create backup of existing database
   sqlite3 production.db .dump > backup.sql
   ```

2. **Update Dependencies**
   ```bash
   # Add GORM dependencies
   go get -u gorm.io/gorm
   go get -u gorm.io/driver/sqlite
   ```

3. **Define GORM Models**
   ```go
   // Create gorm_models.go with struct definitions
   type GormSearch struct {
       ID      uint   `gorm:"primaryKey"`
       Name    string `gorm:"size:255;not null"`
       // ... other fields
   }
   ```

4. **Implement GORM Database Layer**
   ```go
   // Create gorm_db.go with connection management
   type GormDatabase struct {
       db *gorm.DB
   }
   ```

5. **Create GORM Repository**
   ```go
   // Create gorm_repository.go implementing Repository interface
   type GormRepository struct {
       db *GormDatabase
   }
   ```

6. **Implement Migration Manager**
   ```go
   // Create gorm_migration.go for AutoMigrate management
   type GormMigrationManager struct {
       db *GormDatabase
   }
   ```

7. **Add Conversion Functions**
   ```go
   // Add functions to convert between legacy and GORM models
   func ConvertToGormSearch(search Search) GormSearch
   func ConvertFromGormSearch(gormSearch GormSearch) Search
   ```

8. **Update Application Code**
   ```go
   // Replace legacy DB calls with GORM repository calls
   // repo := NewGormRepository(db)
   // searches, err := repo.GetSearches()
   ```

9. **Test Thoroughly**
   ```bash
   # Run comprehensive tests
   go test ./internal/goodwill/db/...
   ```

10. **Deploy to Production**
    ```bash
    # Monitor migration process
    # Verify data integrity
    # Rollback if issues occur
    ```

## GORM AutoMigrate System

### How AutoMigrate Works

1. **Model Analysis**: GORM analyzes all registered models
2. **Schema Comparison**: Compares current database schema with model definitions
3. **Schema Updates**: Creates missing tables, adds missing columns, updates constraints
4. **Version Tracking**: Records migration versions in `gorm_migrations` table

### AutoMigrate Configuration

```go
// GetAllModels returns all models for AutoMigrate
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

// Run AutoMigrate
models := GetAllModels()
if err := db.AutoMigrate(models...); err != nil {
    return fmt.Errorf("auto-migrate failed: %w", err)
}
```

### Migration Version Tracking

```go
// GormMigration tracks migration versions
type GormMigration struct {
    ID        uint      `gorm:"primaryKey"`
    Version   int       `gorm:"unique;not null"`
    Name      string    `gorm:"size:255;not null"`
    AppliedAt time.Time `gorm:"autoCreateTime"`
}

// MigrationManager handles version tracking
type GormMigrationManager struct {
    db         *GormDatabase
    migrations []GormMigration
}
```

## Legacy Compatibility

### Conversion Functions

The system provides bidirectional conversion between legacy and GORM models:

```go
// Convert legacy models to GORM
func ConvertToGormSearch(search Search) GormSearch {
    return GormSearch{
        ID:      uint(search.ID),
        Name:    search.Name,
        Query:   search.Query,
        Enabled: search.Enabled,
        // ... other fields
    }
}

// Convert GORM models to legacy
func ConvertFromGormSearch(gormSearch GormSearch) Search {
    return Search{
        ID:      int(gormSearch.ID),
        Name:    gormSearch.Name,
        Query:   gormSearch.Query,
        Enabled: gormSearch.Enabled,
        // ... other fields
    }
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

### Dual Database Support

During transition period, the system supports both legacy and GORM databases:

```go
// Database interface supports both implementations
type Database interface {
    // Common interface methods
}

// LegacyDatabase implements Database interface
type LegacyDatabase struct {
    // Legacy implementation
}

// GormDatabase implements Database interface
type GormDatabase struct {
    // GORM implementation
}
```

## Migration Best Practices

### Pre-Migration Checklist

1. **Database Backup**: Create complete backup before migration
2. **Test Environment**: Test migration in staging first
3. **Data Validation**: Verify data integrity after migration
4. **Performance Testing**: Test query performance with GORM
5. **Error Handling**: Ensure proper error handling is in place
6. **Monitoring**: Set up monitoring for migration process

### Migration Strategy

1. **Phased Approach**: Migrate tables incrementally
2. **Dual Write**: Write to both systems during transition
3. **Data Validation**: Compare data between systems
4. **Performance Tuning**: Optimize GORM configuration
5. **Rollback Plan**: Have rollback procedure ready

### Post-Migration Tasks

1. **Monitor Performance**: Check query performance
2. **Validate Data**: Ensure all data migrated correctly
3. **Update Documentation**: Document new GORM patterns
4. **Train Team**: Educate team on GORM usage
5. **Cleanup**: Remove legacy migration files after successful transition

## Troubleshooting

### Common Migration Issues

| Issue | Solution |
|-------|----------|
| Schema mismatch | Check GORM tags match legacy schema |
| Data type conversion | Ensure proper type mapping |
| Relationship errors | Verify foreign key definitions |
| Performance issues | Add proper indexes, optimize queries |
| Connection problems | Check connection pool settings |
| Transaction failures | Verify transaction isolation levels |

### Debugging Techniques

```go
// Enable GORM logging
gormDB, err := gorm.Open(sqlite.Open(path), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

// Log slow queries
config := &gorm.Config{
    Logger: logger.New(
        log.New(),
        logger.Config{
            SlowThreshold: 200 * time.Millisecond,
            LogLevel: logger.Info,
            Colorful: true,
        },
    ),
}
```

### Error Handling Patterns

```go
// Handle specific GORM errors
result := db.First(&search, 1)
if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    // Handle not found
} else if result.Error != nil {
    // Handle other errors
}

// Transaction error handling
err := db.WithTransaction(func(tx *gorm.DB) error {
    // Transaction operations
    return nil
})

if err != nil {
    // Handle transaction failure
}
```

## Rollback Procedures

### GORM Migration Rollback

```go
// Rollback last migration
func (m *GormMigrationManager) Rollback() error {
    // Get current version
    currentVersion, err := m.GetCurrentVersion()
    if err != nil {
        return err
    }

    if currentVersion == 0 {
        return errors.New("no migrations to rollback")
    }

    // Find and remove migration
    var migration GormMigration
    if err := m.db.GetDB().Where("version = ?", currentVersion).First(&migration).Error; err != nil {
        return err
    }

    if err := m.db.GetDB().Delete(&migration).Error; err != nil {
        return err
    }

    return nil
}
```

### Emergency Rollback

1. **Stop Application**: Prevent further database access
2. **Restore Backup**: Restore from pre-migration backup
3. **Verify Data**: Check data integrity after restore
4. **Re-deploy Legacy**: Deploy previous version if needed
5. **Investigate Issue**: Determine root cause of migration failure

### Partial Rollback

```go
// Rollback specific changes
func RollbackSpecificMigration(db *gorm.DB, version int) error {
    // Manual schema adjustments
    // May require SQL execution for complex changes

    // Remove migration record
    if err := db.Where("version = ?", version).Delete(&GormMigration{}).Error; err != nil {
        return err
    }

    return nil
}
```

## Performance Considerations

### GORM Optimization

1. **Connection Pooling**: Configure appropriate pool sizes
2. **Batch Operations**: Use batch inserts/updates
3. **Indexing**: Ensure proper indexes are defined
4. **Query Optimization**: Use `Select()` to load only needed fields
5. **Eager Loading**: Use `Preload()` to avoid N+1 queries

### Migration Performance

1. **Batch Size**: Process data in reasonable batch sizes
2. **Transaction Size**: Keep transactions manageable
3. **Memory Usage**: Monitor memory during large migrations
4. **Timeout Settings**: Configure appropriate timeouts
5. **Progress Tracking**: Log migration progress

## Testing Migration

### Test Setup

```go
func setupMigrationTest() (*GormDatabase, error) {
    // Create test database
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

    return db, nil
}
```

### Migration Tests

```go
func TestGormMigration(t *testing.T) {
    db, err := setupMigrationTest()
    require.NoError(t, err)
    defer db.Close()

    // Test migration process
    migrationManager := NewGormMigrationManager(db)

    // Test initial migration
    err = migrationManager.Migrate()
    require.NoError(t, err)

    // Verify migration version
    version, err := migrationManager.GetCurrentVersion()
    require.NoError(t, err)
    assert.NotZero(t, version)

    // Test rollback
    err = migrationManager.Rollback()
    require.NoError(t, err)
}
```

### Data Integrity Tests

```go
func TestDataIntegrityAfterMigration(t *testing.T) {
    // Setup legacy and GORM databases
    legacyDB, err := setupLegacyTestDB()
    require.NoError(t, err)
    defer legacyDB.Close()

    gormDB, err := setupGormTestDB()
    require.NoError(t, err)
    defer gormDB.Close()

    // Migrate data
    err = MigrateData(legacyDB, gormDB)
    require.NoError(t, err)

    // Compare data
    legacySearches, err := legacyDB.GetSearches()
    require.NoError(t, err)

    gormSearches, err := gormDB.GetSearches()
    require.NoError(t, err)

    // Verify counts match
    assert.Equal(t, len(legacySearches), len(gormSearches))

    // Verify data matches
    for i, legacySearch := range legacySearches {
        gormSearch := ConvertToGormSearch(legacySearch)
        assert.Equal(t, gormSearch.Name, gormSearches[i].Name)
        assert.Equal(t, gormSearch.Query, gormSearches[i].Query)
    }
}
```

## Monitoring and Maintenance

### Migration Monitoring

1. **Log Migration Progress**: Track migration steps and timing
2. **Monitor Performance**: Watch for slow queries or timeouts
3. **Validate Data**: Check data integrity during and after migration
4. **Error Tracking**: Log and monitor migration errors
5. **Resource Usage**: Monitor CPU, memory, and disk usage

### Post-Migration Maintenance

1. **Regular Backups**: Continue regular database backups
2. **Performance Tuning**: Optimize GORM configuration
3. **Schema Updates**: Use GORM AutoMigrate for future changes
4. **Documentation Updates**: Keep documentation current
5. **Team Training**: Ensure team understands GORM patterns

## Conclusion

The migration to GORM provides significant benefits including automatic schema management, ORM capabilities, and improved developer productivity. This guide covers the complete migration process, from planning to execution, with comprehensive examples and best practices for working with the new GORM implementation.

For additional information, refer to:
- [GORM Official Documentation](https://gorm.io/docs/)
- [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) - Complete schema specification
- [GORM_USAGE_GUIDE.md](GORM_USAGE_GUIDE.md) - Usage patterns and examples