// Package models - Purchase Order models
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PurchaseOrder - Purchase order header
type PurchaseOrder struct {
	ID                 string          `gorm:"type:varchar(255);primaryKey"`
	TenantID           string          `gorm:"type:varchar(255);not null;index"`
	PONumber           string          `gorm:"type:varchar(100);not null;uniqueIndex"`
	PODate             time.Time       `gorm:"type:timestamp;not null;index"`
	SupplierID         string          `gorm:"type:varchar(255);not null;index"`
	WarehouseID        string          `gorm:"type:varchar(255);not null;index"` // Destination warehouse
	Status             PurchaseOrderStatus `gorm:"type:varchar(20);default:'DRAFT';index"`
	Subtotal           decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	DiscountAmount     decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	TaxAmount          decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	TotalAmount        decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	Notes              *string         `gorm:"type:text"`
	ExpectedDeliveryAt *time.Time      `gorm:"type:timestamp"`
	RequestedBy        *string         `gorm:"type:varchar(255);index"`
	ApprovedBy         *string         `gorm:"type:varchar(255)"`
	ApprovedAt         *time.Time      `gorm:"type:timestamp"`
	CancelledBy        *string         `gorm:"type:varchar(255)"`
	CancelledAt        *time.Time      `gorm:"type:timestamp"`
	CancellationNote   *string         `gorm:"type:text"`
	CreatedAt          time.Time       `gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant      Tenant              `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Supplier    Supplier            `gorm:"foreignKey:SupplierID;constraint:OnDelete:RESTRICT"`
	Warehouse   Warehouse           `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT"`
	Requester   *User               `gorm:"foreignKey:RequestedBy"`
	Items       []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID"`
	// Note: GoodsReceipts, SupplierPayments will reference this PO
}

// TableName specifies the table name for PurchaseOrder model
func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

// BeforeCreate hook to generate CUID for ID field
func (po *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if po.ID == "" {
		po.ID = cuid.New()
	}
	return nil
}

// PurchaseOrderItem - Purchase order line items
type PurchaseOrderItem struct {
	ID              string          `gorm:"type:varchar(255);primaryKey"`
	PurchaseOrderID string          `gorm:"type:varchar(255);not null;index"`
	ProductID       string          `gorm:"type:varchar(255);not null;index"`
	ProductUnitID   *string         `gorm:"type:varchar(255);index"` // NULL = base unit
	Quantity        decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	UnitPrice       decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	DiscountPct     decimal.Decimal `gorm:"type:decimal(5,2);default:0"`
	DiscountAmt     decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	Subtotal        decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	ReceivedQty     decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // Quantity received so far
	Notes           *string         `gorm:"type:text"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`

	// Relations
	PurchaseOrder PurchaseOrder `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:CASCADE"`
	Product       Product       `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	ProductUnit   *ProductUnit  `gorm:"foreignKey:ProductUnitID"`
}

// TableName specifies the table name for PurchaseOrderItem model
func (PurchaseOrderItem) TableName() string {
	return "purchase_order_items"
}

// BeforeCreate hook to generate CUID for ID field
func (poi *PurchaseOrderItem) BeforeCreate(tx *gorm.DB) error {
	if poi.ID == "" {
		poi.ID = cuid.New()
	}
	return nil
}
