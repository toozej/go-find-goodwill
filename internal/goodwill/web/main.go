package web

import (
	"embed"
	"io/fs"
	"log"
	"mime"

	"github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/antibot"
	"github.com/toozej/go-find-goodwill/internal/goodwill/api"
	"github.com/toozej/go-find-goodwill/internal/goodwill/core/scheduling"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/internal/goodwill/notifications"
	webapi "github.com/toozej/go-find-goodwill/internal/goodwill/web/api"
	"github.com/toozej/go-find-goodwill/internal/goodwill/web/ui"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

//go:embed ui/templates/* ui/static/*
var staticFS embed.FS

// WebServer represents the complete web server
type WebServer struct {
	config          *config.Config
	repo            db.Repository
	scheduler       *scheduling.Scheduler
	notificationSvc *notifications.NotificationIntegration
	// server field removed - was unused
	// ui field removed - was unused
	log       *logrus.Logger
	apiServer *webapi.Server
}

// NewWebServer creates a new web server with both API and UI
func NewWebServer(cfg *config.Config, repo db.Repository, scheduler *scheduling.Scheduler, notificationSvc *notifications.NotificationIntegration, logger *logrus.Logger) *WebServer {
	return &WebServer{
		config:          cfg,
		repo:            repo,
		scheduler:       scheduler,
		notificationSvc: notificationSvc,
		log:             logger,
		apiServer:       webapi.NewServer(&cfg.Web, repo, scheduler, notificationSvc, logger),
	}
}

// Start starts the web server (blocking)
func (ws *WebServer) Start() error {
	// Start background services
	ws.scheduler.Start()
	ws.notificationSvc.Start()

	// Create UI handler
	templateFS, err := fs.Sub(staticFS, "ui/templates")
	if err != nil {
		return err
	}
	assetFS, err := fs.Sub(staticFS, "ui/static")
	if err != nil {
		return err
	}

	uiHandler, err := ui.NewUIHandler(templateFS, assetFS, ws.log, ws.repo)
	if err != nil {
		return err
	}

	// Setup UI routes
	uiHandler.SetupRoutes(ws.apiServer.Router())

	// Start the server (blocking)
	return ws.apiServer.Start()
}

// Shutdown gracefully shuts down the web server
func (ws *WebServer) Shutdown() error {
	ws.scheduler.Stop()
	ws.notificationSvc.Stop()
	return ws.apiServer.Shutdown()
}

// RunWebServer starts the web server as a standalone application
func RunWebServer(cfg config.Config) {
	// Register MIME types for static assets
	// This is important for distroless images which don't have a full /etc/mime.types
	_ = mime.AddExtensionType(".js", "application/javascript")
	_ = mime.AddExtensionType(".css", "text/css")

	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Initialize GORM database
	dbConfig := &db.DBConfig{
		Path:              cfg.Database.Path,
		MaxConnections:    cfg.Database.MaxConnections,
		ConnectionTimeout: cfg.Database.ConnectionTimeout,
	}

	gormDatabase, err := db.NewGormDatabase(dbConfig)
	if err != nil {
		log.Printf("Failed to create GORM database: %v", err)
		return
	}

	if err := gormDatabase.Connect(); err != nil {
		log.Printf("Failed to connect to GORM database: %v", err)
		return
	}
	defer gormDatabase.Close()

	// Create GORM repository
	repo := db.NewGormRepository(gormDatabase)

	// Run database migrations
	migrationManager := db.NewGormMigrationManager(gormDatabase)
	if err := migrationManager.EnsureMigrationsTable(); err != nil {
		log.Printf("Failed to ensure migrations table: %v", err)
		return
	}
	if err := migrationManager.LoadMigrations(); err != nil {
		log.Printf("Failed to load migrations: %v", err)
		return
	}
	if err := migrationManager.Migrate(); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		return
	}
	if err := migrationManager.Seed(); err != nil {
		log.Printf("Failed to seed database: %v", err)
		return
	}

	// Create core services
	// Create anti-bot system
	antiBotSystem, err := antibot.NewAntiBotSystem(&cfg.AntiBot, repo)
	if err != nil {
		log.Printf("Failed to create anti-bot system: %v", err)
		return
	}

	apiClient, err := api.NewShopGoodwillClient(&cfg.ShopGoodwill, &cfg.AntiBot, antiBotSystem)
	if err != nil {
		log.Printf("Failed to create API client: %v", err)
		return
	}

	// Create scheduler
	scheduler := scheduling.NewScheduler(&cfg, repo, apiClient)

	// Create notification service
	notificationSvc, err := notifications.NewNotificationIntegration(&cfg, repo)
	if err != nil {
		log.Printf("Failed to create notification integration: %v", err)
		return
	}

	// Create and start web server
	webServer := NewWebServer(&cfg, repo, scheduler, notificationSvc, logger)
	if err := webServer.Start(); err != nil {
		log.Printf("Web server failed: %v", err)
		return
	}

	logger.Info("Application exited successfully")
}
