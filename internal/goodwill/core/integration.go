package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/antibot"
	"github.com/toozej/go-find-goodwill/internal/goodwill/api"
	"github.com/toozej/go-find-goodwill/internal/goodwill/core/deduplication"
	"github.com/toozej/go-find-goodwill/internal/goodwill/core/scheduling"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/internal/goodwill/notifications"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// Application represents the main application instance
type Application struct {
	config           *config.Config
	gormDatabase     *db.GormDatabase
	apiClient        *api.ShopGoodwillClient
	deduplicationSvc *deduplication.DeduplicationService
	scheduler        *scheduling.Scheduler
	notificationSvc  *notifications.NotificationIntegration
	repository       db.Repository
	shutdownChan     chan struct{}
}

// NewApplication creates a new application instance
func NewApplication() (*Application, error) {
	// Load configuration
	cfg, _ := config.GetEnvVars()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Setup logging
	setupLogging(&cfg.Logging)

	// Create GORM database
	gormDatabase, err := db.NewGormDatabase(&db.DBConfig{
		Path:              cfg.Database.Path,
		MaxConnections:    cfg.Database.MaxConnections,
		ConnectionTimeout: cfg.Database.ConnectionTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM database: %w", err)
	}

	// Connect to GORM database
	if err := gormDatabase.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to GORM database: %w", err)
	}

	// Create GORM migration manager and run migrations
	gormMigrationManager := db.NewGormMigrationManager(gormDatabase)
	if err := gormMigrationManager.EnsureMigrationsTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	if err := gormMigrationManager.LoadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	if err := gormMigrationManager.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create GORM repository
	repo := db.NewGormRepository(gormDatabase)

	// Create anti-bot system
	antiBotSystem, err := antibot.NewAntiBotSystem(&cfg.AntiBot, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create anti-bot system: %w", err)
	}

	// Create API client
	apiClient, err := api.NewShopGoodwillClient(&cfg.ShopGoodwill, &cfg.AntiBot, antiBotSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// SearchManager removed - using scheduler only

	// Create deduplication service
	deduplicationSvc := deduplication.NewDeduplicationService(repo, nil)

	// Create scheduler
	scheduler := scheduling.NewScheduler(&cfg, repo, apiClient)

	// SearchEnhancer removed - unused component

	// Create notification service
	notificationSvc, err := notifications.NewNotificationIntegration(&cfg, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification integration: %w", err)
	}

	app := &Application{
		config:           &cfg,
		gormDatabase:     gormDatabase,
		apiClient:        apiClient,
		deduplicationSvc: deduplicationSvc,
		scheduler:        scheduler,
		notificationSvc:  notificationSvc,
		repository:       db.NewGormRepository(gormDatabase),
		shutdownChan:     make(chan struct{}),
	}

	return app, nil
}

// setupLogging configures the logging system
func setupLogging(cfg *config.LoggingConfig) {
	// Set log level
	level, err := log.ParseLevel(cfg.Level)
	if err != nil {
		log.Warnf("Invalid log level '%s', defaulting to 'info'", cfg.Level)
		level = log.InfoLevel
	}
	log.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set log output
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Warnf("Failed to open log file '%s': %v", cfg.File, err)
		} else {
			log.SetOutput(file)
		}
	}
}

// Start starts the application
func (a *Application) Start() error {
	log.Info("Starting go-find-goodwill application")

	// SearchManager removed - using scheduler only

	// Start notification service
	a.notificationSvc.Start()

	// Start scheduler
	a.scheduler.Start()

	// Set up signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	select {
	case <-signalChan:
		log.Info("Received shutdown signal")
		a.Shutdown()
		return nil
	case <-a.shutdownChan:
		log.Info("Received internal shutdown signal")
		a.Shutdown()
		return nil
	}
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown() {
	log.Info("Shutting down application")

	// SearchManager removed - using scheduler only

	// Stop notification service
	a.notificationSvc.Stop()

	// Stop scheduler
	a.scheduler.Stop()

	// Close GORM database connection
	if err := a.gormDatabase.Close(); err != nil {
		log.Errorf("Failed to close GORM database: %v", err)
	}

	close(a.shutdownChan)
	log.Info("Application shutdown complete")
}

// GetAPIClient returns the API client
func (a *Application) GetAPIClient() *api.ShopGoodwillClient {
	return a.apiClient
}

// GetGormDatabase returns the GORM database
func (a *Application) GetGormDatabase() *db.GormDatabase {
	return a.gormDatabase
}

// GetConfig returns the configuration
func (a *Application) GetConfig() *config.Config {
	return a.config
}

// GetDeduplicationService returns the deduplication service
func (a *Application) GetDeduplicationService() *deduplication.DeduplicationService {
	return a.deduplicationSvc
}

// GetScheduler returns the scheduler
func (a *Application) GetScheduler() *scheduling.Scheduler {
	return a.scheduler
}

// GetRepository returns the GORM repository
func (a *Application) GetRepository() db.Repository {
	// Create GORM repository once and reuse it
	if a.repository == nil {
		a.repository = db.NewGormRepository(a.gormDatabase)
	}
	return a.repository
}

// GetNotificationService returns the notification service
func (a *Application) GetNotificationService() *notifications.NotificationIntegration {
	return a.notificationSvc
}

// RunApplication runs the complete application
func RunApplication() error {
	// Create application
	app, err := NewApplication()
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	// Start application
	if err := app.Start(); err != nil {
		return fmt.Errorf("application failed: %w", err)
	}

	return nil
}

// InitializeDatabase initializes the database with sample data for testing
func (a *Application) InitializeDatabase() error {
	log.Info("Initializing database with sample data")

	// Add sample user agents for anti-bot measures
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
	}

	for _, ua := range userAgents {
		userAgent := db.GormUserAgent{
			UserAgent: ua,
			IsActive:  true,
		}
		if err := a.gormDatabase.GetDB().FirstOrCreate(&userAgent, "user_agent = ?", ua).Error; err != nil {
			log.Errorf("Failed to add user agent %s: %v", ua, err)
		}
	}

	// Add sample searches
	sampleSearches := []db.GormSearch{
		{
			Name:                      "Vintage Cameras",
			Query:                     "vintage camera",
			Enabled:                   true,
			NotificationThresholdDays: 1,
			CategoryFilter:            "Cameras & Photography",
			SortBy:                    "ends_soonest",
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		},
		{
			Name:                      "Collectible Watches",
			Query:                     "rolex OR omega OR vintage watch",
			Enabled:                   true,
			NotificationThresholdDays: 2,
			CategoryFilter:            "Jewelry & Watches",
			SortBy:                    "ends_soonest",
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		},
		{
			Name:                      "Electronics Deals",
			Query:                     "laptop OR tablet OR smartphone",
			Enabled:                   true,
			NotificationThresholdDays: 1,
			CategoryFilter:            "Electronics",
			SortBy:                    "ends_soonest",
			CreatedAt:                 time.Now(),
			UpdatedAt:                 time.Now(),
		},
	}

	for _, search := range sampleSearches {
		gormSearch := search
		if err := a.gormDatabase.GetDB().FirstOrCreate(&gormSearch, "name = ?", search.Name).Error; err != nil {
			log.Errorf("Failed to add sample search %s: %v", search.Name, err)
		} else {
			log.Infof("Added sample search: %s", search.Name)
		}
	}

	return nil
}

// HealthCheck performs a health check of all components
func (a *Application) HealthCheck() error {
	log.Info("Performing health check")

	// Check GORM database connection
	if !a.gormDatabase.IsConnected() {
		return fmt.Errorf("GORM database is not connected")
	}

	// Test API client authentication
	if err := a.apiClient.Authenticate(context.Background()); err != nil {
		return fmt.Errorf("API client authentication failed: %w", err)
	}

	log.Info("Health check passed")
	return nil
}
