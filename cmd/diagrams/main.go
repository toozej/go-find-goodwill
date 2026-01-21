// Package main provides diagram generation utilities for the go-find-goodwill project.
//
// This application generates architectural and component diagrams for the go-find-goodwill
// template using the go-diagrams library. It creates visual representations of the
// project structure and component relationships to aid in documentation and understanding.
//
// The generated diagrams are saved as .dot files in the docs/diagrams/go-diagrams/
// directory and can be converted to various image formats using Graphviz.
//
// Usage:
//
//	go run cmd/diagrams/main.go
//
// This will generate:
//   - architecture.dot: High-level architecture showing user interaction flow
//   - components.dot: Component relationships and dependencies
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/generic"
	"github.com/blushft/go-diagrams/nodes/programming"
)

// main is the entry point for the diagram generation utility.
//
// This function orchestrates the entire diagram generation process:
//  1. Creates the output directory structure
//  2. Changes to the appropriate working directory
//  3. Generates architecture and component diagrams
//  4. Reports successful completion
//
// The function will terminate with log.Fatal if any critical operation fails,
// such as directory creation, navigation, or diagram rendering.
func main() {
	// Ensure output directory exists
	if err := os.MkdirAll("docs/diagrams", 0750); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Change to docs/diagrams directory
	if err := os.Chdir("docs/diagrams"); err != nil {
		log.Fatal("Failed to change directory:", err)
	}

	// Generate architecture diagram
	generateArchitectureDiagram()

	// Generate component diagram
	generateComponentDiagram()

	fmt.Println("Diagram .dot files generated successfully in ./docs/diagrams/go-diagrams/")
}

// generateArchitectureDiagram creates a high-level architecture diagram showing
// the interaction flow between users and the go-find-goodwill application components.
//
// The diagram illustrates:
//   - User interaction with both CLI and Web UI
//   - Core service handling both interfaces
//   - Integration with ShopGoodwill API client
//   - Database layer (SQLite/GORM)
//   - Notification service
//   - Anti-bot measures
//   - Scheduling system
//
// The diagram is rendered in top-to-bottom (TB) direction and saved as
// "architecture.dot" in the current working directory. The function will
// terminate the program with log.Fatal if diagram creation or rendering fails.
func generateArchitectureDiagram() {
	d, err := diagram.New(diagram.Filename("architecture"), diagram.Label("go-find-goodwill Architecture"), diagram.Direction("TB"))
	if err != nil {
		log.Fatal(err)
	}

	// Define components
	user := generic.Blank.Blank(diagram.NodeLabel("User"))
	cli := programming.Language.Go(diagram.NodeLabel("CLI Interface"))
	webUI := programming.Language.Go(diagram.NodeLabel("Web UI"))
	coreService := programming.Language.Go(diagram.NodeLabel("Core Service"))
	apiClient := programming.Language.Go(diagram.NodeLabel("ShopGoodwill API Client"))
	database := programming.Language.Go(diagram.NodeLabel("Database Layer\n(SQLite/GORM)"))
	notification := programming.Language.Go(diagram.NodeLabel("Notification Service"))
	antiBot := programming.Language.Go(diagram.NodeLabel("Anti-Bot Measures"))
	scheduler := programming.Language.Go(diagram.NodeLabel("Scheduling System"))
	config := generic.Blank.Blank(diagram.NodeLabel("Configuration\n(env/godotenv)"))
	logging := generic.Blank.Blank(diagram.NodeLabel("Logging\n(logrus)"))

	// Create connections showing the flow
	d.Connect(user, cli, diagram.Forward())
	d.Connect(user, webUI, diagram.Forward())
	d.Connect(cli, coreService, diagram.Forward())
	d.Connect(webUI, coreService, diagram.Forward())
	d.Connect(coreService, apiClient, diagram.Forward())
	d.Connect(coreService, database, diagram.Forward())
	d.Connect(coreService, notification, diagram.Forward())
	d.Connect(coreService, scheduler, diagram.Forward())
	d.Connect(apiClient, antiBot, diagram.Forward())
	d.Connect(coreService, config, diagram.Forward())
	d.Connect(coreService, logging, diagram.Forward())

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}

// generateComponentDiagram creates a detailed component diagram showing the
// relationships and dependencies between different packages in the go-find-goodwill project.
//
// The diagram illustrates:
//   - main.go as the entry point
//   - cmd/go-find-goodwill package handling CLI operations
//   - Web server components (web/main.go, API handlers, UI handlers)
//   - Core service components (search manager, deduplication, scheduling)
//   - Database layer (GORM database, repository)
//   - Notification service components
//   - Anti-bot measures components
//   - Configuration, version, and man packages
//   - Data flow between components
//
// The diagram is rendered in left-to-right (LR) direction and saved as
// "components.dot" in the current working directory. The function will
// terminate the program with log.Fatal if diagram creation or rendering fails.
func generateComponentDiagram() {
	d, err := diagram.New(diagram.Filename("components"), diagram.Label("go-find-goodwill Components"), diagram.Direction("LR"))
	if err != nil {
		log.Fatal(err)
	}

	// Main components
	main := programming.Language.Go(diagram.NodeLabel("main.go"))
	rootCmd := programming.Language.Go(diagram.NodeLabel("cmd/go-find-goodwill\nroot.go"))
	config := programming.Language.Go(diagram.NodeLabel("pkg/config\nconfig.go"))
	version := programming.Language.Go(diagram.NodeLabel("pkg/version\nversion.go"))
	man := programming.Language.Go(diagram.NodeLabel("pkg/man\nman.go"))

	// Web server components
	webMain := programming.Language.Go(diagram.NodeLabel("internal/goodwill/web\nmain.go"))
	apiHandlers := programming.Language.Go(diagram.NodeLabel("internal/goodwill/web/api\n*_handlers.go"))
	uiHandlers := programming.Language.Go(diagram.NodeLabel("internal/goodwill/web/ui\nhandler.go"))

	// Core service components
	scheduling := programming.Language.Go(diagram.NodeLabel("internal/goodwill/core/scheduling\nscheduler.go"))

	// Database components
	gormDB := programming.Language.Go(diagram.NodeLabel("internal/goodwill/db\ngorm_db.go"))
	repository := programming.Language.Go(diagram.NodeLabel("internal/goodwill/db\ngorm_repository.go"))

	// Notification components
	notificationService := programming.Language.Go(diagram.NodeLabel("internal/goodwill/notifications\nnotification_service.go"))

	// Anti-bot components
	antiBotManager := programming.Language.Go(diagram.NodeLabel("internal/goodwill/antibot\nantibot.go"))
	rateLimiter := programming.Language.Go(diagram.NodeLabel("internal/goodwill/antibot\nrate_limiter.go"))
	timingManager := programming.Language.Go(diagram.NodeLabel("internal/goodwill/antibot\ntiming_manager.go"))

	// API client component
	apiClient := programming.Language.Go(diagram.NodeLabel("internal/goodwill/api\nclient.go"))

	// Create connections showing the flow
	d.Connect(main, rootCmd, diagram.Forward())
	d.Connect(rootCmd, config, diagram.Forward())
	d.Connect(rootCmd, version, diagram.Forward())
	d.Connect(rootCmd, man, diagram.Forward())
	d.Connect(rootCmd, webMain, diagram.Forward())

	// Web server connections
	d.Connect(webMain, apiHandlers, diagram.Forward())
	d.Connect(webMain, uiHandlers, diagram.Forward())

	// Core service connections
	d.Connect(webMain, scheduling, diagram.Forward())
	d.Connect(scheduling, apiClient, diagram.Forward())

	// Database connections
	d.Connect(webMain, gormDB, diagram.Forward())
	d.Connect(gormDB, repository, diagram.Forward())
	d.Connect(scheduling, repository, diagram.Forward())

	// Notification connections
	d.Connect(webMain, notificationService, diagram.Forward())
	d.Connect(scheduling, notificationService, diagram.Forward())

	// Anti-bot connections
	d.Connect(apiClient, antiBotManager, diagram.Forward())
	d.Connect(antiBotManager, rateLimiter, diagram.Forward())
	d.Connect(antiBotManager, timingManager, diagram.Forward())

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}
