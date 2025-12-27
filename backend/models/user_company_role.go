// Package models - User-Company role mapping for multi-company permissions
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserCompanyRole - Junction table for User ↔ Company with per-company role assignment
// This implements Tier 2 of the dual-tier permission system
// Allows users to have different roles in different companies within the same tenant
//
// Example:
// - User "Siti Rahayu" in Tenant "PT Multi Bisnis Group":
//   - PT Distribusi Utama: ADMIN role
//   - CV Sembako Jaya: STAFF role
//   - PT Retail Nusantara: NO ACCESS (no record)
type UserCompanyRole struct {
	ID        string    `gorm:"type:varchar(255);primaryKey"`
	UserID    string    `gorm:"type:varchar(255);not null;index:idx_user_company;uniqueIndex:idx_user_company_unique"`
	CompanyID string    `gorm:"type:varchar(255);not null;index:idx_user_company;uniqueIndex:idx_user_company_unique"`
	TenantID  string    `gorm:"type:varchar(255);not null;index:idx_tenant_user_company"` // Denormalized for query optimization
	Role      UserRole  `gorm:"type:varchar(20);not null;index:idx_company_role;check:role IN ('ADMIN','FINANCE','SALES','WAREHOUSE','STAFF')"` // Only Tier 2 roles allowed
	IsActive  bool      `gorm:"default:true;index:idx_user_company_active"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Tenant  Tenant  `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for UserCompanyRole model
func (UserCompanyRole) TableName() string {
	return "user_company_roles"
}

// BeforeCreate hook to generate UUID for ID field and validate role
func (ucr *UserCompanyRole) BeforeCreate(tx *gorm.DB) error {
	if ucr.ID == "" {
		ucr.ID = uuid.New().String()
	}

	// Validate that only Tier 2 (company-level) roles are used
	role := UserRole(ucr.Role)
	if !role.IsCompanyLevel() {
		return gorm.ErrInvalidData // Only ADMIN, FINANCE, SALES, WAREHOUSE, STAFF allowed
	}

	return nil
}

// BeforeUpdate hook to validate role changes
func (ucr *UserCompanyRole) BeforeUpdate(tx *gorm.DB) error {
	// Validate that only Tier 2 (company-level) roles are used
	role := UserRole(ucr.Role)
	if !role.IsCompanyLevel() {
		return gorm.ErrInvalidData // Only ADMIN, FINANCE, SALES, WAREHOUSE, STAFF allowed
	}

	return nil
}
