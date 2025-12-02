package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGormDatabase(t *testing.T) {
	// Create a temporary database file
	tempDB := "test_gorm.db"
	defer func() {
		if err := os.Remove(tempDB); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove test database: %v", err)
		}
	}()

	// Create config
	config := &DBConfig{
		Path:           tempDB,
		MaxConnections: 5,
	}

	// Test NewGormDatabase
	t.Run("NewGormDatabase", func(t *testing.T) {
		db, err := NewGormDatabase(config)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.Equal(t, config, db.config)
		assert.False(t, db.IsConnected())
	})

	// Test Connect
	t.Run("Connect", func(t *testing.T) {
		db, err := NewGormDatabase(config)
		assert.NoError(t, err)

		err = db.Connect()
		assert.NoError(t, err)
		assert.True(t, db.IsConnected())

		// Test GetDB
		gormDB := db.GetDB()
		assert.NotNil(t, gormDB)

		// Test Close
		err = db.Close()
		assert.NoError(t, err)
		assert.False(t, db.IsConnected())
	})

	// Test AutoMigrate
	t.Run("AutoMigrate", func(t *testing.T) {
		db, err := NewGormDatabase(config)
		assert.NoError(t, err)

		err = db.Connect()
		assert.NoError(t, err)

		// Get all models
		models := GetAllModels()
		assert.NotEmpty(t, models)

		// Test AutoMigrate
		err = db.AutoMigrate(models...)
		assert.NoError(t, err)

		err = db.Close()
		assert.NoError(t, err)
	})

	// Test Transaction support
	t.Run("Transactions", func(t *testing.T) {
		db, err := NewGormDatabase(config)
		assert.NoError(t, err)

		err = db.Connect()
		assert.NoError(t, err)

		// Test BeginTransaction
		tx, err := db.BeginTransaction()
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Test WithTransaction
		err = db.WithTransaction(func(tx *gorm.DB) error {
			// Create a test search within transaction
			search := GormSearch{
				Name:      "Test Search",
				Query:     "test query",
				Enabled:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			result := tx.Create(&search)
			assert.NoError(t, result.Error)
			assert.NotZero(t, search.ID)

			return nil
		})
		assert.NoError(t, err)

		// Verify the search was committed
		var count int64
		gormDB := db.GetDB()
		result := gormDB.Model(&GormSearch{}).Where("name = ?", "Test Search").Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), count)

		err = db.Close()
		assert.NoError(t, err)
	})

	// Test GORM-specific query features
	t.Run("GormQueryFeatures", func(t *testing.T) {
		db, err := NewGormDatabase(config)
		assert.NoError(t, err)

		err = db.Connect()
		assert.NoError(t, err)

		// Auto-migrate models
		err = db.AutoMigrate(GetAllModels()...)
		assert.NoError(t, err)

		// Clean up any existing test data from previous tests
		err = db.GetDB().Where("name LIKE ?", "%Test%").Delete(&GormSearch{}).Error
		assert.NoError(t, err)

		// Create test data
		search := GormSearch{
			Name:      "Query Test Search",
			Query:     "test query",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := db.GetDB().Create(&search)
		assert.NoError(t, result.Error)

		// Test GORM query features
		var foundSearch GormSearch
		err = db.GetDB().Where("name LIKE ?", "%Query%").First(&foundSearch).Error
		assert.NoError(t, err)
		assert.Equal(t, "Query Test Search", foundSearch.Name)

		// Test GORM preloading (if associations exist)
		var searches []GormSearch
		err = db.GetDB().Order("created_at DESC").Find(&searches).Error
		assert.NoError(t, err)
		assert.Equal(t, 1, len(searches))

		err = db.Close()
		assert.NoError(t, err)
	})
}
