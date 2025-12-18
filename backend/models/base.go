// Package models contains GORM model definitions for ERP system
// Multi-tenant ERP with comprehensive business logic
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for most models
// Note: Not using gorm.Model because ID is string (cuid), not uint
type BaseModel struct {
	ID        string    `gorm:"type:varchar(255);primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// BeforeCreate hook to generate CUID for ID field
// This hook is automatically called by GORM before inserting a new record
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = cuid.New()
	}
	return nil
}

// BaseModelWithoutTimestamps for models that don't have timestamps
// Used for junction tables or simple reference tables
type BaseModelWithoutTimestamps struct {
	ID string `gorm:"type:varchar(255);primaryKey"`
}

// BeforeCreate hook for models without timestamps
func (m *BaseModelWithoutTimestamps) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = cuid.New()
	}
	return nil
}
