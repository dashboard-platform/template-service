// Package config provides functionality for loading and managing application configuration.
// It retrieves configuration values from environment variables and ensures that all required
// settings are properly initialized. This package is essential for setting up the application's
// runtime environment, including database connections, JWT secrets, and server settings.
package config

import (
	"errors"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config represents the application configuration.
// It contains environment-specific settings such as the environment name,
// server port, JWT secret, and database URL.
type Config struct {
	Port string // The port on which the server will run.
	Env  string // The current environment (e.g., "dev", "prod").
	DSN  string // The Data Source Name for connecting to the database.
}

const (
	envKey  = "ENV"  // Environment variable key for the environment name.
	portEnv = "PORT" // Environment variable key for the server port.
	dsnEnv  = "DSN"  // Database URL environment variable key.

	defaultEnvKey = "dev" // Default environment name if none is provided.
)

// Load retrieves the application configuration from environment variables.
// It ensures that all required configuration values are set and returns an error
// if any mandatory value is missing.
//
// Returns:
//   - Config: The loaded application configuration.
//   - error: An error if any required configuration value is missing.
func Load() (Config, error) {
	var c Config

	c.Env = os.Getenv(envKey)
	if c.Env == "" {
		c.Env = defaultEnvKey
	}

	c.Port = getEnv(portEnv)
	if c.Port == "" {
		return Config{}, errors.New("empty key")
	}

	c.DSN = getEnv(dsnEnv)
	if c.DSN == "" {
		return Config{}, errors.New("empty dsn")
	}

	return c, nil
}

// getEnv retrieves the value of an environment variable.
// If the variable is not set, it logs an error and returns an empty string.
//
// Parameters:
//   - key: The name of the environment variable to retrieve.
//
// Returns:
//   - string: The value of the environment variable, or an empty string if not set.
func getEnv(key string) string {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	val := os.Getenv(key)
	if val == "" {
		log.Error().Str("var", key).Msg("Failed to load environment")
	}
	return val
}
