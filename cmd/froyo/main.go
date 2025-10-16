package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/openfroyo/openfroyo/cmd/froyo/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	// Setup structured logging
	setupLogging()

	// Create context that cancels on interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Execute root command
	if err := commands.Execute(ctx, Version, Commit, BuildDate); err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

// setupLogging configures zerolog for structured logging
func setupLogging() {
	// Use console writer for human-readable output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set log level from environment or default to Info
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
