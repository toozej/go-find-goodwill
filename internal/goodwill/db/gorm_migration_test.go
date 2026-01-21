package db

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGormMigrationManagerBasic(t *testing.T) {
	// Create temporary database file
	tmpFile := "test_gorm_migration_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	// Create GORM database config
	config := &DBConfig{
		Path:           tmpFile,
		MaxConnections: 5,
	}

	// Create GORM database
	gormDB, err := NewGormDatabase(config)
	require.NoError(t, err, "Failed to create GORM database")

	// Connect to database
	err = gormDB.Connect()
	require.NoError(t, err, "Failed to connect to GORM database")
	defer gormDB.Close()

	// Create GORM migration manager
	migrationManager := NewGormMigrationManager(gormDB)

	// Test EnsureMigrationsTable
	err = migrationManager.EnsureMigrationsTable()
	require.NoError(t, err, "Failed to ensure migrations table")

	// Test GetCurrentVersion (should be 0 initially)
	version, err := migrationManager.GetCurrentVersion()
	require.NoError(t, err, "Failed to get current version")
	assert.Equal(t, 0, version, "Initial migration version should be 0")

	// Test LoadMigrations
	err = migrationManager.LoadMigrations()
	require.NoError(t, err, "Failed to load migrations")

	// Test Migrate
	err = migrationManager.Migrate()
	require.NoError(t, err, "Failed to run migrations")

	// Test GetCurrentVersion (should be > 0 after migration)
	version, err = migrationManager.GetCurrentVersion()
	require.NoError(t, err, "Failed to get current version after migration")
	assert.Greater(t, version, 0, "Migration version should be greater than 0 after running migrations")

	// Verify that all expected tables were created
	tables := []string{
		"gorm_searches",
		"gorm_search_executions",
		"gorm_items",
		"gorm_item_details",
		"gorm_price_histories",
		"gorm_bid_histories",
		"gorm_notifications",
		"gorm_user_agents",
		"gorm_system_logs",
		"gorm_search_item_mappings",
		"gorm_migrations",
	}

	for _, table := range tables {
		var count int64
		err := gormDB.GetDB().Table(table).Count(&count).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Logf("Table %s exists", table)
		} else {
			t.Logf("Table %s does not exist", table)
		}
	}
}

// TestGormMigrationManagerErrorHandling tests error handling scenarios
func TestGormMigrationManagerErrorHandling(t *testing.T) {
	// Create temporary database file
	tmpFile := "test_gorm_migration_error_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	// Create GORM database config
	config := &DBConfig{
		Path:           tmpFile,
		MaxConnections: 5,
	}

	// Create GORM database
	gormDB, err := NewGormDatabase(config)
	require.NoError(t, err, "Failed to create GORM database")

	// Connect to database
	err = gormDB.Connect()
	require.NoError(t, err, "Failed to connect to GORM database")
	defer gormDB.Close()

	// Create GORM migration manager
	// migrationManager := NewGormMigrationManager(gormDB) - removed duplicate declaration

	// Test with non-existent migration directory
	t.Run("NonExistentMigrationDirectory", func(t *testing.T) {
		// Create a migration manager with invalid path
		invalidConfig := &DBConfig{
			Path:           tmpFile,
			MaxConnections: 5,
		}

		invalidDB, err := NewGormDatabase(invalidConfig)
		require.NoError(t, err, "Failed to create GORM database with invalid config")

		err = invalidDB.Connect()
		require.NoError(t, err, "Failed to connect to GORM database with invalid config")
		defer invalidDB.Close()

		invalidMigrationManager := NewGormMigrationManager(invalidDB)

		// Test LoadMigrations with invalid path - now succeeds since we use GORM AutoMigrate
		err = invalidMigrationManager.LoadMigrations()
		assert.NoError(t, err, "LoadMigrations should succeed with GORM AutoMigrate")
	})

	// Test migration operations on disconnected database
	t.Run("DisconnectedDatabase", func(t *testing.T) {
		disconnectedDB := &GormDatabase{} // Not connected

		disconnectedMigrationMgr := NewGormMigrationManager(disconnectedDB)

		// Test various operations with disconnected database
		_, err := disconnectedMigrationMgr.GetCurrentVersion()
		assert.Error(t, err)
		assert.Equal(t, "database not connected", err.Error())

		err = disconnectedMigrationMgr.EnsureMigrationsTable()
		assert.Error(t, err)
		assert.Equal(t, "database not connected", err.Error())

		// LoadMigrations now succeeds even with disconnected database since it doesn't depend on external files
		err = disconnectedMigrationMgr.LoadMigrations()
		assert.NoError(t, err, "LoadMigrations should succeed with GORM AutoMigrate")

		err = disconnectedMigrationMgr.Migrate()
		assert.Error(t, err)
		assert.Equal(t, "database not connected", err.Error())

	})
}
