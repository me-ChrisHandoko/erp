// Package models - Purchase Order models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PurchaseOrder - Purchase order header
type PurchaseOrder struct {
	ID                 string              `gorm:"type:varchar(255);primaryKey"`
	TenantID           string              `gorm:"type:varchar(255);not null;index"`
	CompanyID          string              `gorm:"type:varchar(255);not null;index:idx_company_purchase_order;uniqueIndex:idx_company_po_number"`
	PONumber           string              `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_po_number"`
	PODate             time.Time           `gorm:"type:timestamp;not null;index"`
	SupplierID         string              `gorm:"type:varchar(255);not null;index"`
	WarehouseID        string              `gorm:"type:varchar(255);not null;index"` // Destination warehouse
	Status             PurchaseOrderStatus `gorm:"type:varchar(20);default:'DRAFT';index"`
	InvoiceStatus      POInvoiceStatus     `gorm:"type:varchar(25);default:'NOT_INVOICED';index"` // Invoice tracking status (like Odoo)
	Subtotal           decimal.Decimal     `gorm:"type:decimal(15,2);default:0"`
	DiscountAmount     decimal.Decimal     `gorm:"type:decimal(15,2);default:0"`
	TaxAmount          decimal.Decimal     `gorm:"type:decimal(15,2);default:0"`
	TotalAmount        decimal.Decimal     `gorm:"type:decimal(15,2);default:0"`
	Notes              *string             `gorm:"type:text"`
	ExpectedDeliveryAt *time.Time          `gorm:"type:timestamp"`
	RequestedBy        *string             `gorm:"type:varchar(255);index"`
	ApprovedBy         *string             `gorm:"type:varchar(255)"`
	ApprovedAt         *time.Time          `gorm:"type:timestamp"`
	CancelledBy        *string             `gorm:"type:varchar(255)"`
	CancelledAt        *time.Time          `gorm:"type:timestamp"`
	CancellationNote   *string             `gorm:"type:text"`
	ShortClosedBy      *string             `gorm:"type:varchar(255)"` // User who short closed (SAP DCI model)
	ShortClosedAt      *time.Time          `gorm:"type:timestamp"`    // Timestamp of short close
	ShortCloseReason   *string             `gorm:"type:text"`         // Reason for short closing
	CreatedAt          time.Time           `gorm:"autoCreateTime"`
	UpdatedAt          time.Time           `gorm:"autoUpdateTime"`

	// Relations
	Tenant      Tenant              `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company     Company             `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
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

// BeforeCreate hook to generate UUID for ID field
func (po *PurchaseOrder) BeforeCreate(tx *gorm.DB) error {
	if po.ID == "" {
		po.ID = uuid.New().String()
	}
	return nil
}

// UpdateInvoiceStatus updates the invoice status based on items' invoiced quantities
// Call this after updating InvoicedQty on PO items
func (po *PurchaseOrder) UpdateInvoiceStatus() {
	if len(po.Items) == 0 {
		po.InvoiceStatus = POInvoiceStatusNotInvoiced
		return
	}

	totalOrdered := decimal.Zero
	totalInvoiced := decimal.Zero

	for _, item := range po.Items {
		totalOrdered = totalOrdered.Add(item.Quantity)
		totalInvoiced = totalInvoiced.Add(item.InvoicedQty)
	}

	if totalInvoiced.IsZero() {
		po.InvoiceStatus = POInvoiceStatusNotInvoiced
	} else if totalInvoiced.GreaterThanOrEqual(totalOrdered) {
		po.InvoiceStatus = POInvoiceStatusFullyInvoiced
	} else {
		po.InvoiceStatus = POInvoiceStatusPartiallyInvoiced
	}
}

// GetRemainingQtyToInvoice returns total remaining quantity that can still be invoiced
func (po *PurchaseOrder) GetRemainingQtyToInvoice() decimal.Decimal {
	remaining := decimal.Zero
	for _, item := range po.Items {
		itemRemaining := item.Quantity.Sub(item.InvoicedQty)
		if itemRemaining.GreaterThan(decimal.Zero) {
			remaining = remaining.Add(itemRemaining)
		}
	}
	return remaining
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
	ReceivedQty     decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // Quantity received so far (from GRN)
	InvoicedQty     decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // Quantity invoiced so far (from Purchase Invoice)
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

// BeforeCreate hook to generate UUID for ID field
func (poi *PurchaseOrderItem) BeforeCreate(tx *gorm.DB) error {
	if poi.ID == "" {
		poi.ID = uuid.New().String()
	}
	return nil
}

// GetRemainingQtyToInvoice returns the remaining quantity that can still be invoiced for this item
func (poi *PurchaseOrderItem) GetRemainingQtyToInvoice() decimal.Decimal {
	remaining := poi.Quantity.Sub(poi.InvoicedQty)
	if remaining.LessThan(decimal.Zero) {
		return decimal.Zero
	}
	return remaining
}

// CanInvoiceQty checks if the given quantity can be invoiced for this item
// Returns true if qty <= remaining qty to invoice
func (poi *PurchaseOrderItem) CanInvoiceQty(qty decimal.Decimal) bool {
	return qty.LessThanOrEqual(poi.GetRemainingQtyToInvoice())
}

// IsFullyInvoiced returns true if the item has been fully invoiced
func (poi *PurchaseOrderItem) IsFullyInvoiced() bool {
	return poi.InvoicedQty.GreaterThanOrEqual(poi.Quantity)
}
