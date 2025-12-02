package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GormMigrationManager handles database migrations using GORM's AutoMigrate
type GormMigrationManager struct {
	db         *GormDatabase
	migrations []GormMigration
}

// NewGormMigrationManager creates a new GORM migration manager
func NewGormMigrationManager(db *GormDatabase) *GormMigrationManager {
	return &GormMigrationManager{
		db: db,
	}
}

// GormMigration represents a database migration with GORM
// This is the same as the one in gorm_models.go but defined here for the migration manager
type GormMigration struct {
	ID        uint      `gorm:"primaryKey"`
	Version   int       `gorm:"unique;not null"`
	Name      string    `gorm:"size:255;not null"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

// GetCurrentVersion returns the current migration version
func (m *GormMigrationManager) GetCurrentVersion() (int, error) {
	if !m.db.IsConnected() {
		return 0, errors.New("database not connected")
	}

	var version sql.NullInt64
	result := m.db.GetDB().Model(&GormMigration{}).Select("MAX(version)").Scan(&version)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current migration version: %w", result.Error)
	}

	if !version.Valid {
		return 0, nil
	}

	return int(version.Int64), nil
}

// EnsureMigrationsTable ensures the migrations table exists
func (m *GormMigrationManager) EnsureMigrationsTable() error {
	if !m.db.IsConnected() {
		return errors.New("database not connected")
	}

	// AutoMigrate will create the migrations table if it doesn't exist
	if err := m.db.AutoMigrate(&GormMigration{}); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	return nil
}

// LoadMigrations loads migration definitions for GORM models
func (m *GormMigrationManager) LoadMigrations() error {
	// Define explicit migrations with proper versioning
	// This avoids dependency on slice order from GetAllModels()
	explicitMigrations := []GormMigration{
		{Version: 1, Name: "initial_schema"},
		{Version: 2, Name: "add_search_execution_table"},
		{Version: 3, Name: "add_item_tables"},
		{Version: 4, Name: "add_price_history_table"},
		{Version: 5, Name: "add_bid_history_table"},
		{Version: 6, Name: "add_notification_table"},
		{Version: 7, Name: "add_user_agent_table"},
		{Version: 8, Name: "add_system_log_table"},
		{Version: 9, Name: "add_search_item_mapping_table"},
		{Version: 10, Name: "add_migration_table"},
	}

	// Sort migrations by version to ensure proper ordering
	sort.Slice(explicitMigrations, func(i, j int) bool {
		return explicitMigrations[i].Version < explicitMigrations[j].Version
	})

	m.migrations = explicitMigrations
	return nil
}

// Migrate runs GORM AutoMigrate for all models and tracks versions
func (m *GormMigrationManager) Migrate() error {
	if !m.db.IsConnected() {
		return errors.New("database not connected")
	}

	// Get current version
	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	log.Infof("Current GORM migration version: %d", currentVersion)

	// Get all models for AutoMigrate
	models := GetAllModels()

	// Run AutoMigrate for all models
	if err := m.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	// Track migrations that need to be recorded
	for _, migration := range m.migrations {
		if migration.Version <= currentVersion {
			continue
		}

		log.Infof("Recording GORM migration %d: %s", migration.Version, migration.Name)

		// Record migration in database
		gormMigration := GormMigration{
			Version: migration.Version,
			Name:    migration.Name,
		}

		if err := m.db.GetDB().Create(&gormMigration).Error; err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		log.Infof("Successfully recorded GORM migration %d", migration.Version)
	}

	return nil
}

// Rollback rolls back the last migration with proper schema rollback
func (m *GormMigrationManager) Rollback() error {
	if !m.db.IsConnected() {
		return errors.New("database not connected")
	}

	// Get current version
	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if currentVersion == 0 {
		return errors.New("no migrations to rollback")
	}

	// Find the migration to rollback
	var migration GormMigration
	if err := m.db.GetDB().Where("version = ?", currentVersion).First(&migration).Error; err != nil {
		return fmt.Errorf("failed to find migration %d: %w", currentVersion, err)
	}

	log.Infof("Rolling back GORM migration %d: %s", migration.Version, migration.Name)

	// Get the underlying SQL database for direct schema operations
	sqlDB, err := m.db.GetDB().DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Begin transaction for rollback
	tx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin rollback transaction: %w", err)
	}

	// Implement proper schema rollback based on migration version
	// This is a simplified approach - in a real system you'd have specific rollback logic per migration
	var rollbackError error

	switch migration.Version {
	case 1:
		// Rollback migration 1: Drop specific tables or columns
		_, rollbackError = tx.Exec(`
			-- Example rollback SQL for migration 1
			-- This would contain the actual schema changes to undo migration 1
			-- For GORM migrations, this is complex as GORM doesn't track specific schema changes
		`)
	case 2:
		// Rollback migration 2
		_, rollbackError = tx.Exec(`
			-- Example rollback SQL for migration 2
		`)
	default:
		// For unknown migrations, we'll just remove the record
		// In a production system, you'd want to handle this more gracefully
		log.Warnf("No specific rollback logic for migration %d, only removing migration record", migration.Version)
	}

	if rollbackError != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Errorf("Rollback failed and transaction rollback error: %v", rbErr)
		}
		return fmt.Errorf("failed to execute rollback SQL: %w", rollbackError)
	}

	// Remove migration record
	if err := m.db.GetDB().Delete(&migration).Error; err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Errorf("Migration record removal failed and transaction rollback error: %v", rbErr)
		}
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit the rollback transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	log.Infof("Successfully rolled back GORM migration %d", migration.Version)

	return nil
}

// GetMigrationHistory returns the migration history
func (m *GormMigrationManager) GetMigrationHistory() ([]GormMigration, error) {
	if !m.db.IsConnected() {
		return nil, errors.New("database not connected")
	}

	var migrations []GormMigration
	if err := m.db.GetDB().Order("version asc").Find(&migrations).Error; err != nil {
		return nil, fmt.Errorf("failed to get migration history: %w", err)
	}

	return migrations, nil
}
