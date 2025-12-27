// Package models - System configuration and audit models
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Setting - System and tenant-specific configuration
type Setting struct {
	ID        string    `gorm:"type:varchar(255);primaryKey"`
	TenantID  *string   `gorm:"type:varchar(255);index"` // NULL for system-wide settings
	Key       string    `gorm:"type:varchar(255);not null;index"`
	Value     *string   `gorm:"type:text"` // JSON or plain text
	DataType  string    `gorm:"type:varchar(50);default:'STRING'"` // STRING, NUMBER, BOOLEAN, JSON
	IsPublic  bool      `gorm:"default:false"` // Whether setting is exposed to frontend
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	Tenant *Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for Setting model
func (Setting) TableName() string {
	return "settings"
}

// BeforeCreate hook to generate UUID for ID field
func (s *Setting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// AuditLog - Comprehensive audit trail for all critical operations
type AuditLog struct {
	ID            string    `gorm:"type:varchar(255);primaryKey"`
	TenantID      *string   `gorm:"type:varchar(255);index"` // NULL for system operations
	UserID        *string   `gorm:"type:varchar(255);index"` // NULL for system operations
	Action        string    `gorm:"type:varchar(100);not null;index"` // CREATE, UPDATE, DELETE, LOGIN, etc.
	EntityType    *string   `gorm:"type:varchar(100);index"` // Product, Invoice, User, etc.
	EntityID      *string   `gorm:"type:varchar(255);index"` // ID of affected entity
	OldValues     *string   `gorm:"type:text"` // JSON of old values
	NewValues     *string   `gorm:"type:text"` // JSON of new values
	IPAddress     *string   `gorm:"type:varchar(45)"` // IPv4 or IPv6
	UserAgent     *string   `gorm:"type:varchar(500)"`
	Notes         *string   `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index"`

	// Relations
	Tenant *Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	User   *User   `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to generate UUID for ID field
func (al *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == "" {
		al.ID = uuid.New().String()
	}
	return nil
}
