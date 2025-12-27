// Package models - User management models
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User - Application user with multi-tenant access
// Note: This model matches the database schema from migration 000001_init_schema.up.sql
type User struct {
	ID            string    `gorm:"type:varchar(255);primaryKey"`
	Email         string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Username      string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash  string    `gorm:"column:password;type:varchar(255);not null"` // Maps to DB column 'password'
	FullName      string    `gorm:"column:name;type:varchar(255);not null"` // Maps to DB column 'name'
	IsSystemAdmin bool      `gorm:"default:false"` // Can manage all tenants
	IsActive      bool      `gorm:"default:true"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`

	// Relations - will be populated by GORM when querying with Preload
	Tenants          []UserTenant      `gorm:"foreignKey:UserID"` // Tier 1: Tenant-level access
	UserCompanyRoles []UserCompanyRole `gorm:"foreignKey:UserID"` // Tier 2: Per-company access
	// Note: Other relations (SalesOrders, PurchaseOrders, etc.) will be added in their respective model files
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to generate UUID for ID field
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// UserTenant - Junction table for User ↔ Tenant with per-tenant role
// This allows a user to access multiple tenants with different roles per tenant
type UserTenant struct {
	ID        string    `gorm:"type:varchar(255);primaryKey"`
	UserID    string    `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_user_tenant"`
	TenantID  string    `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_user_tenant"`
	Role      UserRole  `gorm:"type:varchar(20);default:'STAFF';index"`
	IsActive  bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for UserTenant model
func (UserTenant) TableName() string {
	return "user_tenants"
}

// BeforeCreate hook to generate UUID for ID field
func (ut *UserTenant) BeforeCreate(tx *gorm.DB) error {
	if ut.ID == "" {
		ut.ID = uuid.New().String()
	}
	return nil
}
