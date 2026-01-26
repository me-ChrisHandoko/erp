// Package models - Delivery Tolerance Settings
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// DeliveryToleranceLevel - Hierarchy level for tolerance settings (SAP Model)
type DeliveryToleranceLevel string

const (
	ToleranceLevelCompany  DeliveryToleranceLevel = "COMPANY"
	ToleranceLevelCategory DeliveryToleranceLevel = "CATEGORY"
	ToleranceLevelProduct  DeliveryToleranceLevel = "PRODUCT"
)

// DeliveryTolerance - Configurable delivery tolerance settings
// Based on SAP's hierarchical tolerance configuration model:
// - Company level: Default tolerances for all products
// - Category level: Override for specific product categories (using category string from Product)
// - Product level: Override for specific products
//
// Resolution order: Product > Category > Company
type DeliveryTolerance struct {
	ID        string                 `gorm:"type:varchar(255);primaryKey"`
	TenantID  string                 `gorm:"type:varchar(255);not null;index:idx_dt_tenant"`
	CompanyID string                 `gorm:"type:varchar(255);not null;uniqueIndex:idx_dt_unique,priority:1"`
	Level     DeliveryToleranceLevel `gorm:"type:varchar(20);not null;index:idx_dt_level;uniqueIndex:idx_dt_unique,priority:2"`

	// Reference to category (string) or product (foreign key)
	// For CATEGORY level: uses category string matching Product.Category field
	// For PRODUCT level: uses foreign key to products table
	// Using empty string instead of NULL for unique constraint compatibility
	CategoryName string `gorm:"type:varchar(100);not null;default:'';uniqueIndex:idx_dt_unique,priority:3"` // For category-level tolerance
	ProductID    string `gorm:"type:varchar(255);not null;default:'';uniqueIndex:idx_dt_unique,priority:4"` // For product-level tolerance

	// Tolerance percentages (SAP Model)
	// Example: 5% means if ordered 100, acceptable range is 95-105
	UnderDeliveryTolerance decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0"` // Percentage (0-100)
	OverDeliveryTolerance  decimal.Decimal `gorm:"type:decimal(5,2);not null;default:0"` // Percentage (0-100)

	// SAP-style unlimited over-delivery flag
	// When true, any over-delivery is accepted (ignores OverDeliveryTolerance)
	UnlimitedOverDelivery bool `gorm:"type:boolean;not null;default:false"`

	// Active flag for soft enable/disable
	IsActive bool `gorm:"type:boolean;not null;default:true"`

	// Notes for documentation
	Notes *string `gorm:"type:text"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedBy *string   `gorm:"type:varchar(255)"`
	UpdatedBy *string   `gorm:"type:varchar(255)"`

	// Relations
	Tenant  Tenant   `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company Company  `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Product *Product `gorm:"foreignKey:ProductID;references:ID;constraint:false"` // Optional relation, empty string means no product - no FK constraint to allow empty string
}

// TableName specifies the table name for DeliveryTolerance model
func (DeliveryTolerance) TableName() string {
	return "delivery_tolerances"
}

// BeforeCreate hook to generate UUID and validate level/reference consistency
func (dt *DeliveryTolerance) BeforeCreate(tx *gorm.DB) error {
	if dt.ID == "" {
		dt.ID = uuid.New().String()
	}
	return nil
}
