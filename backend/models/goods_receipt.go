// Package models - Goods Receipt models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GoodsReceipt - Goods receipt note (GRN)
type GoodsReceipt struct {
	ID               string             `gorm:"type:varchar(255);primaryKey"`
	TenantID         string             `gorm:"type:varchar(255);not null;index"`
	CompanyID        string             `gorm:"type:varchar(255);not null;index:idx_company_goods_receipt;uniqueIndex:idx_company_grn_number"`
	GRNNumber        string             `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_grn_number"`
	GRNDate          time.Time          `gorm:"type:timestamp;not null;index"`
	PurchaseOrderID  string             `gorm:"type:varchar(255);not null;index"`
	WarehouseID      string             `gorm:"type:varchar(255);not null;index"` // Destination warehouse
	SupplierID       string             `gorm:"type:varchar(255);not null;index"`
	Status           GoodsReceiptStatus `gorm:"type:varchar(20);default:'PENDING';index"`
	ReceivedBy       *string            `gorm:"type:varchar(255)"` // User who received
	ReceivedAt       *time.Time         `gorm:"type:timestamp"`
	InspectedBy      *string            `gorm:"type:varchar(255)"` // User who inspected
	InspectedAt      *time.Time         `gorm:"type:timestamp"`
	SupplierInvoice  *string            `gorm:"type:varchar(100)"` // Supplier's invoice number
	SupplierDONumber *string            `gorm:"type:varchar(100)"` // Supplier's delivery order number
	Notes            *string            `gorm:"type:text"`
	CreatedAt        time.Time          `gorm:"autoCreateTime"`
	UpdatedAt        time.Time          `gorm:"autoUpdateTime"`

	// Relations
	Tenant        Tenant             `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company       Company            `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	PurchaseOrder PurchaseOrder      `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:RESTRICT"`
	Warehouse     Warehouse          `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT"`
	Supplier      Supplier           `gorm:"foreignKey:SupplierID;constraint:OnDelete:RESTRICT"`
	Items         []GoodsReceiptItem `gorm:"foreignKey:GoodsReceiptID"`
}

// TableName specifies the table name for GoodsReceipt model
func (GoodsReceipt) TableName() string {
	return "goods_receipts"
}

// BeforeCreate hook to generate UUID for ID field
func (gr *GoodsReceipt) BeforeCreate(tx *gorm.DB) error {
	if gr.ID == "" {
		gr.ID = uuid.New().String()
	}
	return nil
}

// GoodsReceiptItem - Goods receipt line items with quality inspection
type GoodsReceiptItem struct {
	ID                 string          `gorm:"type:varchar(255);primaryKey"`
	GoodsReceiptID     string          `gorm:"type:varchar(255);not null;index"`
	PurchaseOrderItemID string          `gorm:"type:varchar(255);not null;index"`
	ProductID          string          `gorm:"type:varchar(255);not null;index"`
	ProductUnitID      *string         `gorm:"type:varchar(255);index"`
	BatchNumber        *string         `gorm:"type:varchar(100);index"` // Batch/lot number from supplier
	ManufactureDate    *time.Time      `gorm:"type:timestamp"`
	ExpiryDate         *time.Time      `gorm:"type:timestamp;index"` // For perishable products
	OrderedQty         decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	ReceivedQty        decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // Physically received
	AcceptedQty        decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // Passed quality inspection
	RejectedQty        decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // Failed quality inspection
	RejectionReason    *string         `gorm:"type:text"`
	QualityNote        *string         `gorm:"type:text"`
	Notes              *string         `gorm:"type:text"`
	CreatedAt          time.Time       `gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime"`

	// Relations
	GoodsReceipt       GoodsReceipt      `gorm:"foreignKey:GoodsReceiptID;constraint:OnDelete:CASCADE"`
	PurchaseOrderItem  PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderItemID;constraint:OnDelete:RESTRICT"`
	Product            Product           `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	ProductUnit        *ProductUnit      `gorm:"foreignKey:ProductUnitID"`
}

// TableName specifies the table name for GoodsReceiptItem model
func (GoodsReceiptItem) TableName() string {
	return "goods_receipt_items"
}

// BeforeCreate hook to generate UUID for ID field
func (gri *GoodsReceiptItem) BeforeCreate(tx *gorm.DB) error {
	if gri.ID == "" {
		gri.ID = uuid.New().String()
	}
	return nil
}
