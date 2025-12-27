// Package models - Warehouse management models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Warehouse - Multi-warehouse management
type Warehouse struct {
	ID        string        `gorm:"type:varchar(255);primaryKey"`
	TenantID  string        `gorm:"type:varchar(255);not null;index"`
	CompanyID string        `gorm:"type:varchar(255);not null;index:idx_company_warehouse;uniqueIndex:idx_company_warehouse_code"`
	Code      string        `gorm:"type:varchar(50);not null;index;uniqueIndex:idx_company_warehouse_code"`
	Name      string        `gorm:"type:varchar(255);not null"`
	Type      WarehouseType `gorm:"type:varchar(20);default:'MAIN';index"`
	Address    *string         `gorm:"type:text"`
	City       *string         `gorm:"type:varchar(100)"`
	Province   *string         `gorm:"type:varchar(100)"`
	PostalCode *string         `gorm:"type:varchar(50)"`
	Phone      *string         `gorm:"type:varchar(50)"`
	Email      *string         `gorm:"type:varchar(255)"`
	ManagerID  *string         `gorm:"type:varchar(255);index"`
	Capacity   *decimal.Decimal `gorm:"type:decimal(15,2)"` // Square meters or volume
	IsActive   bool            `gorm:"default:true"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant  Tenant           `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company Company          `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Manager *User            `gorm:"foreignKey:ManagerID"`
	Stocks  []WarehouseStock `gorm:"foreignKey:WarehouseID"`
	// Note: InventoryMovements, GoodsReceipts, Deliveries, StockOpnames, StockTransfers will be added in Phase 3-4
}

// TableName specifies the table name for Warehouse model
func (Warehouse) TableName() string {
	return "warehouses"
}

// BeforeCreate hook to generate UUID for ID field
func (w *Warehouse) BeforeCreate(tx *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	return nil
}

// WarehouseStock - Stock per warehouse per product
// This is the actual stock tracking table (Product.currentStock is deprecated)
type WarehouseStock struct {
	ID            string          `gorm:"type:varchar(255);primaryKey"`
	WarehouseID   string          `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_warehouse_product"`
	ProductID     string          `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_warehouse_product"`
	Quantity      decimal.Decimal `gorm:"type:decimal(15,3);default:0;index"` // Stock quantity (base unit)
	MinimumStock  decimal.Decimal `gorm:"type:decimal(15,3);default:0"`
	MaximumStock  decimal.Decimal `gorm:"type:decimal(15,3);default:0"`
	Location      *string         `gorm:"type:varchar(100)"` // e.g., "RAK-A-01", "ZONE-B"
	LastCountDate *time.Time      `gorm:"type:timestamp"`
	LastCountQty  *decimal.Decimal `gorm:"type:decimal(15,3)"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Warehouse Warehouse      `gorm:"foreignKey:WarehouseID;constraint:OnDelete:CASCADE"`
	Product   Product        `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Batches   []ProductBatch `gorm:"foreignKey:WarehouseStockID"`
}

// TableName specifies the table name for WarehouseStock model
func (WarehouseStock) TableName() string {
	return "warehouse_stocks"
}

// BeforeCreate hook to generate UUID for ID field
func (ws *WarehouseStock) BeforeCreate(tx *gorm.DB) error {
	if ws.ID == "" {
		ws.ID = uuid.New().String()
	}
	return nil
}
