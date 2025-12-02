package web

import (
	"embed"
	"log"

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
	log *logrus.Logger
}

// NewWebServer creates a new web server with both API and UI
func NewWebServer(cfg *config.Config, repo db.Repository, scheduler *scheduling.Scheduler, notificationSvc *notifications.NotificationIntegration, logger *logrus.Logger) *WebServer {
	return &WebServer{
		config:          cfg,
		repo:            repo,
		scheduler:       scheduler,
		notificationSvc: notificationSvc,
		log:             logger,
	}
}

// Start starts the web server (non-blocking)
func (ws *WebServer) Start() error {
	// Create API server with all services
	apiServer := webapi.NewServer(&ws.config.Web, ws.repo, ws.scheduler, ws.notificationSvc, ws.log)

	// Create UI handler
	uiHandler, err := ui.NewUIHandler(staticFS, ws.log, ws.repo)
	if err != nil {
		return err
	}

	// Setup UI routes
	uiHandler.SetupRoutes(apiServer.Router())

	// Start the server (non-blocking)
	apiServer.Start()
	return nil
}

// Shutdown gracefully shuts down the web server
func (ws *WebServer) Shutdown() error {
	// Create API server for shutdown
	apiServer := webapi.NewServer(&ws.config.Web, ws.repo, ws.scheduler, ws.notificationSvc, ws.log)
	return apiServer.Shutdown()
}

// RunWebServer starts the web server as a standalone application
func RunWebServer(cfg config.Config) {

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
}
