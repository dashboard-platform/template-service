// Package logger provides utilities for initializing and managing application logging.
// It supports environment-specific configurations and component-specific loggers.
package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInit verifies that the Init function initializes a logger for different environments.
func TestInit(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{name: "dev", env: "dev"},
		{name: "prod", env: "prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Init(tt.env)
			require.NotNil(t, logger)
		})
	}
}
