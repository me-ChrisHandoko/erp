// Package models - Inventory Movement tracking
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InventoryMovement - Stock movement audit trail
// CRITICAL: Every stock change MUST create an InventoryMovement record
type InventoryMovement struct {
	ID              string          `gorm:"type:varchar(255);primaryKey"`
	TenantID        string          `gorm:"type:varchar(255);not null;index"`
	CompanyID       string          `gorm:"type:varchar(255);not null;index:idx_company_inventory_movement"`
	MovementDate    time.Time       `gorm:"type:timestamp;not null;index"`
	WarehouseID     string          `gorm:"type:varchar(255);not null;index"`
	ProductID       string          `gorm:"type:varchar(255);not null;index"`
	BatchID         *string         `gorm:"type:varchar(255);index"` // Required if product.isBatchTracked
	MovementType    MovementType    `gorm:"type:varchar(20);not null;index"`
	Quantity        decimal.Decimal `gorm:"type:decimal(15,3);not null"` // Positive = IN, Negative = OUT
	StockBefore     decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	StockAfter      decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	ReferenceType   *string         `gorm:"type:varchar(50)"` // GOODS_RECEIPT, DELIVERY, ADJUSTMENT, etc.
	ReferenceID     *string         `gorm:"type:varchar(255);index"`
	ReferenceNumber *string         `gorm:"type:varchar(100)"` // GRN-001, DEL-001, etc.
	Notes           *string         `gorm:"type:text"`
	CreatedBy       *string         `gorm:"type:varchar(255)"` // User who created movement
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant    Tenant        `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company   Company       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Warehouse Warehouse     `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT"`
	Product   Product       `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	Batch     *ProductBatch `gorm:"foreignKey:BatchID"`
}

// TableName specifies the table name for InventoryMovement model
func (InventoryMovement) TableName() string {
	return "inventory_movements"
}

// BeforeCreate hook to generate UUID for ID field
func (im *InventoryMovement) BeforeCreate(tx *gorm.DB) error {
	if im.ID == "" {
		im.ID = uuid.New().String()
	}
	return nil
}
