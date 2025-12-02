package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds database configuration
type DBConfig struct {
	Path               string
	MaxConnections     int
	MaxIdleConnections int
	ConnectionTimeout  time.Duration
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
}

// GormDatabase represents the GORM database connection and provides repository functionality
type GormDatabase struct {
	db     *gorm.DB
	config *DBConfig
}

// NewGormDatabase creates a new GORM database instance
func NewGormDatabase(config *DBConfig) (*GormDatabase, error) {
	if config == nil {
		return nil, errors.New("database config cannot be nil")
	}

	// Set default values
	if config.MaxConnections == 0 {
		config.MaxConnections = 10
	}
	if config.MaxIdleConnections == 0 {
		config.MaxIdleConnections = 5
	}
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 30 * time.Second
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 30 * time.Minute
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 10 * time.Minute
	}

	db := &GormDatabase{
		config: config,
	}

	return db, nil
}

// Connect establishes a connection to the SQLite database using GORM
func (d *GormDatabase) Connect() error {
	// Ensure database directory exists
	err := os.MkdirAll(filepath.Dir(d.config.Path), 0750)
	if err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(), // Use logrus as the underlying logger
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open database connection with GORM
	gormDB, err := gorm.Open(sqlite.Open(d.config.Path), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(d.config.MaxConnections)
	sqlDB.SetMaxIdleConns(d.config.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(d.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(d.config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), d.config.ConnectionTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.db = gormDB
	log.Infof("Connected to GORM database at %s", d.config.Path)

	return nil
}

// Close closes the database connection
func (d *GormDatabase) Close() error {
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying SQL DB for closing: %w", err)
		}

		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}

		d.db = nil
		log.Info("GORM database connection closed")
	}

	return nil
}

// GetDB returns the underlying GORM database connection
func (d *GormDatabase) GetDB() *gorm.DB {
	return d.db
}

// IsConnected checks if database is connected
func (d *GormDatabase) IsConnected() bool {
	return d.db != nil
}

// AutoMigrate runs GORM's AutoMigrate for all models
func (d *GormDatabase) AutoMigrate(models ...interface{}) error {
	if d.db == nil {
		return errors.New("database not connected")
	}

	if err := d.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}

	log.Info("Successfully auto-migrated database models")
	return nil
}

// BeginTransaction starts a new transaction
func (d *GormDatabase) BeginTransaction() (*gorm.DB, error) {
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
