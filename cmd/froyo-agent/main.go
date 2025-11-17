package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/piwi3910/openfroyo/internal/agent"
	"github.com/piwi3910/openfroyo/internal/agent/config"
)

const (
	version = "1.0.0"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "/etc/froyo/agent.yml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("froyo-agent version %s\n", version)
		os.Exit(0)
	}

	// Create logger
	logger := log.New(os.Stdout, "[froyo-agent] ", log.LstdFlags)

	logger.Printf("OpenFroyo Agent v%s", version)
	logger.Printf("Loading configuration from: %s", *configPath)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate required configuration
	if cfg.Agent.ID == "" {
		logger.Fatal("agent.id is required in configuration")
	}
	if cfg.Agent.Datacenter == "" {
		logger.Fatal("agent.datacenter is required in configuration")
	}

	// Set version if not in config
	if cfg.Agent.Version == "" {
		cfg.Agent.Version = version
	}

	logger.Printf("Agent ID: %s", cfg.Agent.ID)
	logger.Printf("Datacenter: %s", cfg.Agent.Datacenter)
	logger.Printf("NATS Servers: %v", cfg.NATS.Servers)

	// Create agent
	ag, err := agent.New(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create agent: %v", err)
	}

	// Run agent (blocks until shutdown)
	if err := ag.Run(); err != nil {
		logger.Fatalf("Agent failed: %v", err)
	}

	logger.Println("Agent shutdown complete")
}
