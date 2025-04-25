// Package database provides functionality for managing database connections and operations.
// It includes methods for initializing the database, performing migrations, and executing
// CRUD operations on user data. This package is essential for interacting with the application's
// persistent storage layer.

package database

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dashboard-platform/template-service/models"
	"github.com/google/uuid"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

// Database represents the database connection and provides methods for interacting with it.
// It includes a GORM database instance and a logger for logging database operations.
type Database struct {
	db     *gorm.DB       // GORM database instance.
	logger zerolog.Logger // Logger for database operations.
}

// Init initializes a new database connection using the provided DSN (Data Source Name).
// It retries the connection multiple times if the database is not ready.
//
// Parameters:
//   - dsn: The Data Source Name for connecting to the database.
//   - logger: A logger instance for logging database operations.
//
// Returns:
//   - *Database: A pointer to the initialized Database instance.
//   - error: An error if the connection fails after retries.
func Init(dsn string, logger zerolog.Logger) (*Database, error) {
	var (
		db  *gorm.DB
		err error
	)

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLog.Default.LogMode(gormLog.Silent),
		})

		if err == nil {
			break
		}

		log.Warn().Err(err).Msgf("DB is not ready, retrying... (%d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database after retries")
		os.Exit(1)
	}

	return &Database{
		db:     db,
		logger: logger,
	}, nil
}

// AutoMigrate performs database migrations for the User model.
//
// Returns:
//   - error: An error if the migration fails.
func (d *Database) AutoMigrate() error {
	if err := d.db.AutoMigrate(&models.TemplateVersion{}); err != nil {
		return err
	}

	if err := d.db.AutoMigrate(&models.TemplateField{}); err != nil {
		return err
	}

	if err := d.db.AutoMigrate(&models.Template{}); err != nil {
		return err
	}

	if err := d.db.AutoMigrate(&models.TemplateHistory{}); err != nil {
		return err
	}

	d.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_template_version_unique ON template_versions (template_id, version);")
	d.db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_template_field_key ON template_fields (template_id, key);")

	return nil
}

func (d *Database) CreateTemplate(userID uuid.UUID, input models.CreateTemplateAPI) (uuid.UUID, error) {
	var templateID uuid.UUID
	err := d.db.Transaction(func(tx *gorm.DB) error {
		templateID = uuid.New()

		// Create Template
		template := models.Template{
			ID:          templateID,
			UserID:      userID,
			Name:        input.Name,
			Description: input.Description,
			Type:        input.Type,
			Category:    input.Category,
			IsPublic:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := tx.Create(&template).Error; err != nil {
			return err
		}

		// Create first version (v1)
		version := models.TemplateVersion{
			ID:         uuid.New(),
			TemplateID: templateID,
			Version:    1,
			Content:    input.Content,
			CreatedAt:  time.Now(),
		}

		if err := tx.Create(&version).Error; err != nil {
			return err
		}

		// Create fields
		for _, f := range input.Fields {
			field := models.TemplateField{
				ID:         uuid.New(),
				TemplateID: templateID,
				Key:        f.Key,
				Label:      f.Label,
				Type:       f.Type,
				Required:   f.Required,
				CreatedAt:  time.Now(),
			}
			if field.Type == "" {
				field.Type = "text" // default
			}

			if f.Options != nil {
				var jsonCheck interface{}
				if err := json.Unmarshal(f.Options, &jsonCheck); err != nil {
					return err
				}
			}

			if err := tx.Create(&field).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	return templateID, nil
}

func (d *Database) GetTemplates(userIDStr string) ([]models.Template, error) {
	var templates []models.Template

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	if err := d.db.
		Preload("Fields").
		Preload("Versions").
		Where("user_id = ?", userID).
		Find(&templates).Error; err != nil {
		return nil, err
	}

	return templates, nil
}

func (d *Database) GetTemplateByID(userIDStr, templateIDStr string) (models.Template, error) {
	var template models.Template

	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return template, err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return template, err
	}

	if err := d.db.
		Preload("Fields").
		Preload("Versions").
		Where("id = ? AND user_id = ?", templateID, userID).
		First(&template).Error; err != nil {
		return template, err
	}

	return template, nil
}

func (d *Database) CreateHistory(template models.TemplateDTO, userID uuid.UUID) error {
	templateID, err := uuid.Parse(template.ID)
	if err != nil {
		return err
	}

	data := models.TemplateHistory{
		ID:           uuid.New(),
		UserID:       userID,
		TemplateID:   templateID,
		TemplateName: template.Name,
		Version:      template.Version.Version,
	}

	return d.db.Create(&data).Error
}

func (d *Database) GetHistory(userID uuid.UUID) ([]models.TemplateHistory, error) {
	var history []models.TemplateHistory

	err := d.db.Where("user_id = ?", userID).Find(&history).Error
	if err != nil {
		return nil, err
	}

	return history, nil
}
