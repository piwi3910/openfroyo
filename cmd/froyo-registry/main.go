package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/piwi3910/openfroyo/internal/registry"
)

var (
	version = "1.0.0"
)

func main() {
	configPath := flag.String("config", "/etc/froyo/registry.yml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	listenAddr := flag.String("listen", ":8080", "Listen address")
	modulesDir := flag.String("modules-dir", "/var/lib/froyo-registry/modules", "Modules directory")
	flag.Parse()

	if *showVersion {
		fmt.Printf("froyo-registry version %s\n", version)
		os.Exit(0)
	}

	// Create logger
	logger := log.New(os.Stdout, "[registry] ", log.LstdFlags)

	// Load configuration
	cfg, err := registry.LoadConfig(*configPath)
	if err != nil {
		// Use defaults if config file doesn't exist
		logger.Printf("Warning: Could not load config: %v, using defaults", err)
		cfg = registry.DefaultConfig()
	}

	// Override with CLI flags
	if *listenAddr != ":8080" {
		cfg.Server.Listen = *listenAddr
	}
	if *modulesDir != "/var/lib/froyo-registry/modules" {
		cfg.Storage.ModulesDir = *modulesDir
	}

	logger.Printf("Starting OpenFroyo Module Registry v%s", version)
	logger.Printf("Modules directory: %s", cfg.Storage.ModulesDir)
	logger.Printf("Listen address: %s", cfg.Server.Listen)

	// Create registry server
	srv, err := registry.NewServer(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create registry server: %v", err)
	}

	// Start server in background
	go func() {
		logger.Printf("Registry server listening on %s", cfg.Server.Listen)
		if err := srv.Start(); err != nil {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down...")
	if err := srv.Shutdown(); err != nil {
		logger.Printf("Error during shutdown: %v", err)
	}
	logger.Println("Shutdown complete")
}
