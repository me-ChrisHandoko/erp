// Package models - Stock Transfer (Inter-warehouse) models
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// StockTransfer - Inter-warehouse stock transfer header
type StockTransfer struct {
	ID                string              `gorm:"type:varchar(255);primaryKey"`
	TenantID          string              `gorm:"type:varchar(255);not null;index"`
	TransferNumber    string              `gorm:"type:varchar(100);not null;uniqueIndex"`
	TransferDate      time.Time           `gorm:"type:timestamp;not null;index"`
	SourceWarehouseID string              `gorm:"type:varchar(255);not null;index"`
	DestWarehouseID   string              `gorm:"type:varchar(255);not null;index"`
	Status            StockTransferStatus `gorm:"type:varchar(20);default:'DRAFT';index"`
	ShippedBy         *string             `gorm:"type:varchar(255)"` // User who shipped
	ShippedAt         *time.Time          `gorm:"type:timestamp"`
	ReceivedBy        *string             `gorm:"type:varchar(255)"` // User who received
	ReceivedAt        *time.Time          `gorm:"type:timestamp"`
	Notes             *string             `gorm:"type:text"`
	CreatedAt         time.Time           `gorm:"autoCreateTime"`
	UpdatedAt         time.Time           `gorm:"autoUpdateTime"`

	// Relations
	Tenant           Tenant              `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	SourceWarehouse  Warehouse           `gorm:"foreignKey:SourceWarehouseID;constraint:OnDelete:RESTRICT"`
	DestWarehouse    Warehouse           `gorm:"foreignKey:DestWarehouseID;constraint:OnDelete:RESTRICT"`
	Items            []StockTransferItem `gorm:"foreignKey:StockTransferID"`
}

// TableName specifies the table name for StockTransfer model
func (StockTransfer) TableName() string {
	return "stock_transfers"
}

// BeforeCreate hook to generate CUID for ID field
func (st *StockTransfer) BeforeCreate(tx *gorm.DB) error {
	if st.ID == "" {
		st.ID = cuid.New()
	}
	return nil
}

// StockTransferItem - Stock transfer line items
type StockTransferItem struct {
	ID              string          `gorm:"type:varchar(255);primaryKey"`
	StockTransferID string          `gorm:"type:varchar(255);not null;index"`
	ProductID       string          `gorm:"type:varchar(255);not null;index"`
	BatchID         *string         `gorm:"type:varchar(255);index"` // For batch-tracked products
	Quantity        decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	Notes           *string         `gorm:"type:text"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`

	// Relations
	StockTransfer StockTransfer `gorm:"foreignKey:StockTransferID;constraint:OnDelete:CASCADE"`
	Product       Product       `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	Batch         *ProductBatch `gorm:"foreignKey:BatchID"`
}

// TableName specifies the table name for StockTransferItem model
func (StockTransferItem) TableName() string {
	return "stock_transfer_items"
}

// BeforeCreate hook to generate CUID for ID field
func (sti *StockTransferItem) BeforeCreate(tx *gorm.DB) error {
	if sti.ID == "" {
		sti.ID = cuid.New()
	}
	return nil
}
