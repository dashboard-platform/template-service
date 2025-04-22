// Package logger provides utilities for initializing and managing application logging.
// It supports environment-specific configurations and component-specific loggers.
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Init initializes the base logger for the application.
//
// Parameters:
//   - env: The current environment (e.g., "dev", "prod").
//
// Returns:
//   - zerolog.Logger: The initialized logger instance.
func Init(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "dev" {
		return zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
	}

	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// NewComponentLogger creates a logger for a specific application component.
//
// Parameters:
//   - logger: The base logger instance.
//   - name: The name of the component.
//
// Returns:
//   - zerolog.Logger: The component-specific logger instance.
func NewComponentLogger(logger zerolog.Logger, name string) zerolog.Logger {
	return logger.With().Str("component", name).Logger()
}
