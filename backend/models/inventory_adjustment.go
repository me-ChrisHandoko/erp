// Package models - Inventory Adjustment (Stock Adjustment) models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InventoryAdjustment - Stock adjustment header
type InventoryAdjustment struct {
	ID               string                    `gorm:"type:varchar(255);primaryKey"`
	TenantID         string                    `gorm:"type:varchar(255);not null;index"`
	CompanyID        string                    `gorm:"type:varchar(255);not null;index:idx_company_adjustment;uniqueIndex:idx_company_adjustment_number"`
	AdjustmentNumber string                    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_adjustment_number"`
	AdjustmentDate   time.Time                 `gorm:"type:timestamp;not null;index"`
	WarehouseID      string                    `gorm:"type:varchar(255);not null;index"`
	AdjustmentType   InventoryAdjustmentType   `gorm:"type:varchar(20);not null"`
	Reason           InventoryAdjustmentReason `gorm:"type:varchar(20);not null"`
	Status           InventoryAdjustmentStatus `gorm:"type:varchar(20);default:'DRAFT';index"`
	Notes            *string                   `gorm:"type:text"`
	TotalItems       int                       `gorm:"type:int;default:0"`
	TotalValue       decimal.Decimal           `gorm:"type:decimal(15,2);default:0"`
	CreatedBy        string                    `gorm:"type:varchar(255);not null"` // User who created
	ApprovedBy       *string                   `gorm:"type:varchar(255)"`          // User who approved
	ApprovedAt       *time.Time                `gorm:"type:timestamp"`
	CancelledBy      *string                   `gorm:"type:varchar(255)"` // User who cancelled
	CancelledAt      *time.Time                `gorm:"type:timestamp"`
	CancelReason     *string                   `gorm:"type:text"` // Reason for cancellation
	CreatedAt        time.Time                 `gorm:"autoCreateTime"`
	UpdatedAt        time.Time                 `gorm:"autoUpdateTime"`

	// Relations
	Tenant         Tenant                    `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company        Company                   `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Warehouse      Warehouse                 `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT"`
	Items          []InventoryAdjustmentItem `gorm:"foreignKey:AdjustmentID"`
	CreatedByUser  User                      `gorm:"foreignKey:CreatedBy;constraint:OnDelete:RESTRICT"`
	ApprovedByUser *User                     `gorm:"foreignKey:ApprovedBy;constraint:OnDelete:SET NULL"`
}

// TableName specifies the table name for InventoryAdjustment model
func (InventoryAdjustment) TableName() string {
	return "inventory_adjustments"
}

// BeforeCreate hook to generate UUID for ID field
func (ia *InventoryAdjustment) BeforeCreate(tx *gorm.DB) error {
	if ia.ID == "" {
		ia.ID = uuid.New().String()
	}
	return nil
}

// InventoryAdjustmentItem - Stock adjustment line items
type InventoryAdjustmentItem struct {
	ID               string          `gorm:"type:varchar(255);primaryKey"`
	AdjustmentID     string          `gorm:"type:varchar(255);not null;index"`
	ProductID        string          `gorm:"type:varchar(255);not null;index"`
	BatchID          *string         `gorm:"type:varchar(255);index"` // For batch-tracked products
	QuantityBefore   decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	QuantityAdjusted decimal.Decimal `gorm:"type:decimal(15,3);not null"` // Can be positive or negative
	QuantityAfter    decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	UnitCost         decimal.Decimal `gorm:"type:decimal(15,2);not null"` // Cost per unit for valuation
	TotalValue       decimal.Decimal `gorm:"type:decimal(15,2);not null"` // QuantityAdjusted * UnitCost
	Notes            *string         `gorm:"type:text"`
	CreatedAt        time.Time       `gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Adjustment InventoryAdjustment `gorm:"foreignKey:AdjustmentID;constraint:OnDelete:CASCADE"`
	Product    Product             `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	Batch      *ProductBatch       `gorm:"foreignKey:BatchID"`
}

// TableName specifies the table name for InventoryAdjustmentItem model
func (InventoryAdjustmentItem) TableName() string {
	return "inventory_adjustment_items"
}

// BeforeCreate hook to generate UUID for ID field
func (iai *InventoryAdjustmentItem) BeforeCreate(tx *gorm.DB) error {
	if iai.ID == "" {
		iai.ID = uuid.New().String()
	}
	return nil
}
