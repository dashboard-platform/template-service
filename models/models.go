package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Response struct {
	Error bool `json:"error"`
	Data  any  `json:"data"`
}

type Template struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"not null;index"`
	Name        string    `gorm:"not null"`
	Description string
	Type        string `gorm:"not null"` // html, latex, etc.
	Category    string
	IsPublic    bool `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relations
	Versions []TemplateVersion `gorm:"foreignKey:TemplateID"`
	Fields   []TemplateField   `gorm:"foreignKey:TemplateID"`
}

func (t *Template) ToDTO() TemplateDTO {
	fields := make([]FieldDTO, 0, len(t.Fields))
	for _, f := range t.Fields {
		fields = append(fields, FieldDTO{
			Key:      f.Key,
			Label:    f.Label,
			Type:     f.Type,
			Required: f.Required,
			Options:  json.RawMessage(f.Options), // already JSON
		})
	}

	var latest TemplateVersionDTO
	if len(t.Versions) > 0 {
		v := t.Versions[0]
		latest = TemplateVersionDTO{
			Version: v.Version,
			Content: v.Content,
		}
	}

	return TemplateDTO{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		Type:        t.Type,
		Category:    t.Category,
		Fields:      fields,
		Version:     latest,
	}
}

type TemplateVersion struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TemplateID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_template_version_unique"`
	Version    int       `gorm:"not null;uniqueIndex:idx_template_version_unique"`
	Content    string    `gorm:"type:text;not null"`
	CreatedAt  time.Time
}

type TemplateField struct {
	gorm.Model
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TemplateID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Key        string         `gorm:"not null"`
	Label      string         `gorm:"not null"`
	Type       string         `gorm:"not null;default:text"` // text, number, date, etc.
	Required   bool           `gorm:"default:true"`
	Options    datatypes.JSON `gorm:"type:jsonb"` // optional, for select fields
	CreatedAt  time.Time
}

type TemplateDTO struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Type        string             `json:"type"`
	Category    string             `json:"category"`
	Fields      []FieldDTO         `json:"fields"`
	Version     TemplateVersionDTO `json:"version"`
}

type FieldDTO struct {
	Key      string          `json:"key"`
	Label    string          `json:"label"`
	Type     string          `json:"type"`
	Required bool            `json:"required"`
	Options  json.RawMessage `json:"options,omitempty"`
}

type TemplateVersionDTO struct {
	Version int    `json:"version"`
	Content string `json:"content"`
}

type CreateTemplateAPI struct {
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description"`
	Type        string             `json:"type" binding:"required"` // html, latex, etc.
	Category    string             `json:"category"`
	Content     string             `json:"content" binding:"required"`
	Fields      []TemplateFieldAPI `json:"fields"`
}

type TemplateFieldAPI struct {
	Key      string          `json:"key" binding:"required"`
	Label    string          `json:"label" binding:"required"`
	Type     string          `json:"type"` // optional default = "text"
	Required bool            `json:"required"`
	Options  json.RawMessage `json:"options"` // optional for select
}
