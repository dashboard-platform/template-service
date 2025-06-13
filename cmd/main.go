package main

import (
	"time"

	"github.com/dashboard-platform/template-service/internal/config"
	"github.com/dashboard-platform/template-service/internal/database"
	"github.com/dashboard-platform/template-service/internal/handler"
	"github.com/dashboard-platform/template-service/internal/logger"
	"github.com/dashboard-platform/template-service/internal/middleware"
	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Load the configuration from environment variables.
	c, err := config.Load()
	if err != nil {
		return
	}

	// Initialize the logger with the loaded configuration
	baseLogger := logger.Init(c.Env)
	httpLogger := logger.NewComponentLogger(baseLogger, "http")

	db, err := database.Init(c.DSN, baseLogger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
		return
	}

	if err := db.AutoMigrate(); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
		return
	}

	app := fiber.New()
	// Middlewares
	app.Use(
		// Add security headers.
		helmet.New(),

		// Add custom request logger middleware.
		middleware.RequestLogger(httpLogger),
	)

	globalLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
	})

	h := handler.New(db)

	app.Post("/templates", globalLimiter, h.CreateTemplate)
	app.Get("/templates", globalLimiter, h.GetTemplates)
	app.Post("/templates/history", globalLimiter, h.CreateHistory)
	app.Get("/templates/history", globalLimiter, h.GetHistory)
	app.Get("/templates/:id", globalLimiter, h.GetTemplateByID)
	app.Post("/templates/:id/update", globalLimiter, h.UpdateTemplate)
	app.Post("/templates/:id/delete", globalLimiter, h.DeleteTemplate)
	app.Post("/templates/:id/preview", limiter.New(limiter.Config{
		Max:        1000,
		Expiration: 1 * time.Minute,
	}), h.PreviewTemplate)
	// Start the HTTP server.
	log.Info().Msgf("Template Service started on %s", c.Port)
	if err = app.Listen(c.Port); err != nil {
		log.Error().Msgf("Error starting  template service: %v", err)
		return
	}
}
