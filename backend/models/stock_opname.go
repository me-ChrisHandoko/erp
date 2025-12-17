// Package models - Stock Opname (Physical Count) models
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// StockOpname - Physical inventory count header
type StockOpname struct {
	ID          string            `gorm:"type:varchar(255);primaryKey"`
	TenantID    string            `gorm:"type:varchar(255);not null;index"`
	OpnameNumber string            `gorm:"type:varchar(100);not null;uniqueIndex"`
	OpnameDate  time.Time         `gorm:"type:timestamp;not null;index"`
	WarehouseID string            `gorm:"type:varchar(255);not null;index"`
	Status      StockOpnameStatus `gorm:"type:varchar(20);default:'DRAFT';index"`
	CountedBy   *string           `gorm:"type:varchar(255)"` // User who performed count
	ApprovedBy  *string           `gorm:"type:varchar(255)"` // User who approved adjustments
	ApprovedAt  *time.Time        `gorm:"type:timestamp"`
	Notes       *string           `gorm:"type:text"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`

	// Relations
	Tenant    Tenant            `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Warehouse Warehouse         `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT"`
	Items     []StockOpnameItem `gorm:"foreignKey:StockOpnameID"`
}

// TableName specifies the table name for StockOpname model
func (StockOpname) TableName() string {
	return "stock_opnames"
}

// BeforeCreate hook to generate CUID for ID field
func (so *StockOpname) BeforeCreate(tx *gorm.DB) error {
	if so.ID == "" {
		so.ID = cuid.New()
	}
	return nil
}

// StockOpnameItem - Physical count line items
type StockOpnameItem struct {
	ID             string          `gorm:"type:varchar(255);primaryKey"`
	StockOpnameID  string          `gorm:"type:varchar(255);not null;index"`
	ProductID      string          `gorm:"type:varchar(255);not null;index"`
	BatchID        *string         `gorm:"type:varchar(255);index"` // For batch-tracked products
	SystemQty      decimal.Decimal `gorm:"type:decimal(15,3);not null"` // Qty per system
	PhysicalQty    decimal.Decimal `gorm:"type:decimal(15,3);not null"` // Actual counted qty
	DifferenceQty  decimal.Decimal `gorm:"type:decimal(15,3);not null"` // Physical - System
	Notes          *string         `gorm:"type:text"`
	CreatedAt      time.Time       `gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime"`

	// Relations
	StockOpname StockOpname   `gorm:"foreignKey:StockOpnameID;constraint:OnDelete:CASCADE"`
	Product     Product       `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	Batch       *ProductBatch `gorm:"foreignKey:BatchID"`
}

// TableName specifies the table name for StockOpnameItem model
func (StockOpnameItem) TableName() string {
	return "stock_opname_items"
}

// BeforeCreate hook to generate CUID for ID field
func (soi *StockOpnameItem) BeforeCreate(tx *gorm.DB) error {
	if soi.ID == "" {
		soi.ID = cuid.New()
	}
	return nil
}
