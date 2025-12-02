package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGormDatabaseOperations(t *testing.T) {
	// Create temporary database file
	tmpFile := "test_gorm_db_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	// Create database config
	config := &DBConfig{
		Path:              tmpFile,
		MaxConnections:    2,
		ConnectionTimeout: 5 * time.Second,
	}

	// Create GORM database
	gormDB, err := NewGormDatabase(config)
	require.NoError(t, err, "Failed to create GORM database")

	// Connect to database
	err = gormDB.Connect()
	require.NoError(t, err, "Failed to connect to GORM database")
	defer gormDB.Close()

	// Test database connection
	assert.True(t, gormDB.IsConnected(), "GORM database should be connected")

	// Test GetDB
	db := gormDB.GetDB()
	assert.NotNil(t, db, "GORM database connection should not be nil")
}

func TestGormMigrationManager(t *testing.T) {
	// Create temporary database file
	tmpFile := "test_gorm_migration_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	// Create database config
	config := &DBConfig{
		Path:              tmpFile,
		MaxConnections:    2,
		ConnectionTimeout: 5 * time.Second,
	}

	// Create GORM database
	gormDB, err := NewGormDatabase(config)
	require.NoError(t, err, "Failed to create GORM database")

	// Connect to database
	err = gormDB.Connect()
	require.NoError(t, err, "Failed to connect to GORM database")
	defer gormDB.Close()

	// Create GORM migration manager
	gormMigrationManager := NewGormMigrationManager(gormDB)

	// Test EnsureMigrationsTable
	err = gormMigrationManager.EnsureMigrationsTable()
	require.NoError(t, err, "Failed to ensure GORM migrations table")

	// Test LoadMigrations
	err = gormMigrationManager.LoadMigrations()
	require.NoError(t, err, "Failed to load GORM migrations")

	// Test GetCurrentVersion (should be 0 initially)
	version, err := gormMigrationManager.GetCurrentVersion()
	require.NoError(t, err, "Failed to get GORM current version")
	assert.Equal(t, 0, version, "Initial GORM migration version should be 0")

	// Test Migrate
	err = gormMigrationManager.Migrate()
	require.NoError(t, err, "Failed to run GORM migrations")

	// Test GetCurrentVersion (should be > 0 after migration)
	version, err = gormMigrationManager.GetCurrentVersion()
	require.NoError(t, err, "Failed to get GORM current version after migration")
	assert.Greater(t, version, 0, "GORM migration version should be greater than 0 after running migrations")
}

func TestGormRepository(t *testing.T) {
	// Create temporary database file
	tmpFile := "test_gorm_repo_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	// Create database config
	config := &DBConfig{
		Path:              tmpFile,
		MaxConnections:    2,
		ConnectionTimeout: 5 * time.Second,
	}

	// Create GORM database
	gormDB, err := NewGormDatabase(config)
	require.NoError(t, err, "Failed to create GORM database")

	// Connect to database
	err = gormDB.Connect()
	require.NoError(t, err, "Failed to connect to GORM database")
	defer gormDB.Close()

	// Create GORM migration manager and run migrations
	gormMigrationManager := NewGormMigrationManager(gormDB)
	err = gormMigrationManager.EnsureMigrationsTable()
	require.NoError(t, err, "Failed to ensure GORM migrations table")

	err = gormMigrationManager.LoadMigrations()
	require.NoError(t, err, "Failed to load GORM migrations")

	err = gormMigrationManager.Migrate()
	require.NoError(t, err, "Failed to run GORM migrations")

	// Create GORM repository
	repo := NewGormRepository(gormDB)

	// Test AddSearch and GetSearches
	search := GormSearch{
		Name:                      "Test GORM Search",
		Query:                     "test gorm query",
		Enabled:                   true,
		NotificationThresholdDays: 1,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	searchID, err := repo.AddSearch(context.Background(), search)
	require.NoError(t, err, "Failed to add GORM search")
	assert.Greater(t, searchID, 0, "GORM search ID should be greater than 0")

	searches, err := repo.GetSearches(context.Background())
	require.NoError(t, err, "Failed to get GORM searches")
	assert.Equal(t, 1, len(searches), "Should have 1 GORM search")
	assert.Equal(t, "Test GORM Search", searches[0].Name, "GORM search name should match")

	// Test GetSearchByID
	retrievedSearch, err := repo.GetSearchByID(context.Background(), searchID)
	require.NoError(t, err, "Failed to get GORM search by ID")
	assert.NotNil(t, retrievedSearch, "Retrieved GORM search should not be nil")
	assert.Equal(t, "Test GORM Search", retrievedSearch.Name, "GORM search name should match")

	// Test UpdateSearch
	retrievedSearch.Name = "Updated Test GORM Search"
	err = repo.UpdateSearch(context.Background(), *retrievedSearch)
	require.NoError(t, err, "Failed to update GORM search")

	updatedSearch, err := repo.GetSearchByID(context.Background(), searchID)
	require.NoError(t, err, "Failed to get updated GORM search")
	assert.Equal(t, "Updated Test GORM Search", updatedSearch.Name, "GORM search name should be updated")

	// Test GetActiveSearches
	activeSearches, err := repo.GetActiveSearches(context.Background())
	require.NoError(t, err, "Failed to get GORM active searches")
	assert.Equal(t, 1, len(activeSearches), "Should have 1 GORM active search")

	// Test AddItem and GetItemByGoodwillID
	item := GormItem{
		GoodwillID:   "GORM_TEST123",
		Title:        "Test GORM Item",
		CurrentPrice: 10.99,
		URL:          "https://example.com/gorm-test",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
	}

	itemID, err := repo.AddItem(context.Background(), item)
	require.NoError(t, err, "Failed to add GORM item")
	assert.Greater(t, itemID, 0, "GORM item ID should be greater than 0")

	retrievedItem, err := repo.GetItemByGoodwillID(context.Background(), "GORM_TEST123")
	require.NoError(t, err, "Failed to get GORM item by Goodwill ID")
	assert.NotNil(t, retrievedItem, "Retrieved GORM item should not be nil")
	assert.Equal(t, "Test GORM Item", retrievedItem.Title, "GORM item title should match")

	// Test AddSearchExecution
	execution := GormSearchExecution{
		SearchID:      searchID,
		ExecutedAt:    time.Now(),
		Status:        "success",
		ItemsFound:    1,
		NewItemsFound: 1,
	}

	executionID, err := repo.AddSearchExecution(context.Background(), execution)
	require.NoError(t, err, "Failed to add GORM search execution")
	assert.Greater(t, executionID, 0, "GORM execution ID should be greater than 0")

	// Test GetSearchHistory
	history, err := repo.GetSearchHistory(context.Background(), searchID, 10)
	require.NoError(t, err, "Failed to get GORM search history")
	assert.Equal(t, 1, len(history), "Should have 1 GORM search execution in history")
}
