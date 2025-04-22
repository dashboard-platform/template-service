package middleware

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// RequestLogger logs details about incoming HTTP requests and their responses.
// It logs the method, path, status, latency, and user ID (if available).
//
// Parameters:
//   - logger: A zerolog.Logger instance for logging.
//
// Returns:
//   - fiber.Handler: The middleware handler function.
func RequestLogger(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		stop := time.Now()

		status := c.Response().StatusCode()
		msg := "request"

		if err != nil {
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
				msg = fe.Message
			} else {
				status = fiber.StatusInternalServerError
				msg = "Internal Server Error"
			}
		}

		event := logger.Info()
		if err != nil || status >= 400 {
			event = logger.Error()
		}

		if err != nil {
			event = event.Err(err)
		}

		userID := c.Locals("user_id")
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			event = event.Str("user_id", userIDStr)
		}

		event.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", stop.Sub(start)).
			Str("ip", c.IP()).
			Msg("request")

		if err != nil {
			return c.Status(status).JSON(fiber.Map{
				"error": msg,
			})
		}

		return nil
	}
}
