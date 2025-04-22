// Package handler provides HTTP handlers for the authentication service.
// These handlers process incoming requests, interact with the service layer, and return responses.
package handler

import (
	"github.com/aymerick/raymond"
	"github.com/dashboard-platform/template-service/internal/database"
	"github.com/dashboard-platform/template-service/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// HTTPHandler represents the HTTP handlers for the authentication service.
// It includes methods for health checks, user registration, login, and retrieving user details.
type HTTPHandler struct {
	db *database.Database
}

// New creates a new instance of HTTPHandler.
//
// Parameters:
//   - authService: The authentication service implementation.
//   - jwtObj: The JWT utility object.
//
// Returns:
//   - HTTPHandler: A new instance of the HTTPHandler.
func New(db *database.Database) HTTPHandler {
	return HTTPHandler{
		db: db,
	}
}

// Healthcheck handles the health check endpoint.
//
// Returns:
//   - fiber.StatusOK: If the service is running.
func (h *HTTPHandler) Healthcheck(ctx *fiber.Ctx) error {
	log.Info().Msg("Healthcheck called")

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "auth-service is alive",
	})
}

func (h *HTTPHandler) CreateTemplate(ctx *fiber.Ctx) error {
	var data models.CreateTemplateAPI

	if err := ctx.BodyParser(&data); err != nil {
		log.Error().Err(err).Msg("error reading/parsing HTTP request body data")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userIDStr := ctx.Get("X-User-ID")
	if userIDStr == "" {
		log.Error().Msg("X-User-ID header is missing")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "X-User-ID header is required",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("error parsing X-User-ID header")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid X-User-ID header",
		})
	}

	id, err := h.db.CreateTemplate(userID, data)
	if err != nil {
		log.Error().Err(err).Msg("error creating template")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create template",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(models.Response{
		Error: false,
		Data: fiber.Map{
			"id":   id,
			"name": data.Name,
		},
	})
}

func (h *HTTPHandler) GetTemplates(ctx *fiber.Ctx) error {
	userID := ctx.Get("X-User-ID")
	if userID == "" {
		log.Error().Msg("X-User-ID header is missing")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "X-User-ID header is required",
		})
	}

	templates, err := h.db.GetTemplates(userID)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving templates")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve templates",
		})
	}

	var dto []models.TemplateDTO
	for _, t := range templates {
		dto = append(dto, t.ToDTO())
	}

	return ctx.Status(fiber.StatusOK).JSON(models.Response{
		Error: false,
		Data: fiber.Map{
			"templates": dto,
		},
	})
}

func (h *HTTPHandler) GetTemplateByID(ctx *fiber.Ctx) error {
	userID := ctx.Get("X-User-ID")
	if userID == "" {
		log.Error().Msg("X-User-ID header is missing")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "X-User-ID header is required",
		})
	}

	templateID := ctx.Params("id")
	if templateID == "" {
		log.Error().Msg("template ID is missing")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Template ID is required",
		})
	}

	template, err := h.db.GetTemplateByID(userID, templateID)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving template by ID")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve template",
		})
	}

	dto := template.ToDTO()

	return ctx.Status(fiber.StatusOK).JSON(models.Response{
		Error: false,
		Data: fiber.Map{
			"template": dto,
		},
	})
}

func (h *HTTPHandler) PreviewTemplate(ctx *fiber.Ctx) error {
	templateID := ctx.Params("id")
	if templateID == "" {
		log.Error().Msg("template ID is missing")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Template ID is required",
		})
	}

	userID := ctx.Get("X-User-ID")
	if userID == "" {
		log.Error().Msg("X-User-ID header is missing")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "X-User-ID header is required",
		})
	}

	var req struct {
		Values map[string]interface{} `json:"values"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("error reading/parsing HTTP request body data")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	template, err := h.db.GetTemplateByID(userID, templateID)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving template by ID")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve template",
		})
	}

	if len(template.Versions) == 0 {
		log.Error().Msg("template has no versions")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Template has no versions",
		})
	}

	content := template.Versions[0].Content

	result, err := raymond.Render(content, req.Values)
	if err != nil {
		log.Error().Err(err).Msg("error rendering template")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to render template",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(models.Response{
		Error: false,
		Data: fiber.Map{
			"preview_html": result,
		},
	})
}
