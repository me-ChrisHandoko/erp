// Package models - Purchase Invoice (Supplier Invoice) models for procurement
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PurchaseInvoiceStatus represents the workflow status of a purchase invoice
type PurchaseInvoiceStatus string

const (
	PurchaseInvoiceStatusDraft     PurchaseInvoiceStatus = "DRAFT"
	PurchaseInvoiceStatusSubmitted PurchaseInvoiceStatus = "SUBMITTED"
	PurchaseInvoiceStatusApproved  PurchaseInvoiceStatus = "APPROVED"
	PurchaseInvoiceStatusRejected  PurchaseInvoiceStatus = "REJECTED"
	PurchaseInvoiceStatusPaid      PurchaseInvoiceStatus = "PAID"
	PurchaseInvoiceStatusCancelled PurchaseInvoiceStatus = "CANCELLED"
)

// Note: PaymentStatus and related constants are defined in enums.go
// We reuse the existing PaymentStatus type instead of creating a duplicate

// PurchaseInvoice represents a supplier invoice (procurement invoice)
type PurchaseInvoice struct {
	// Primary Key
	ID string `gorm:"type:varchar(255);primaryKey"`

	// Multi-Tenancy
	TenantID  string `gorm:"type:varchar(255);not null;index:idx_purchase_invoice_tenant"`
	CompanyID string `gorm:"type:varchar(255);not null;index:idx_purchase_invoice_company;uniqueIndex:idx_company_purchase_invoice_number"`

	// Invoice Information
	InvoiceNumber string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_purchase_invoice_number"`
	InvoiceDate   time.Time `gorm:"type:timestamp;not null;index:idx_purchase_invoice_date"`
	DueDate       time.Time `gorm:"type:timestamp;not null;index:idx_purchase_invoice_due_date"`

	// Supplier Reference
	SupplierID   string  `gorm:"type:varchar(255);not null;index:idx_purchase_invoice_supplier"`
	SupplierName string  `gorm:"type:varchar(255);not null"` // Denormalized for reporting
	SupplierCode *string `gorm:"type:varchar(100)"`          // Denormalized

	// Purchase Order Reference (optional - can create invoice without PO)
	PurchaseOrderID *string `gorm:"type:varchar(255);index:idx_purchase_invoice_po"`
	PONumber        *string `gorm:"type:varchar(100)"`

	// Goods Receipt Reference (optional)
	GoodsReceiptID *string `gorm:"type:varchar(255);index:idx_purchase_invoice_gr"`
	GRNumber       *string `gorm:"type:varchar(100)"`

	// Financial Information
	SubtotalAmount  decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0"` // Before tax & discount
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(20,4);default:0"`
	TaxAmount       decimal.Decimal `gorm:"type:decimal(20,4);default:0"`   // PPN
	TaxRate         decimal.Decimal `gorm:"type:decimal(5,2);default:11.00"` // 11% PPN
	TotalAmount     decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0"`
	PaidAmount      decimal.Decimal `gorm:"type:decimal(20,4);default:0;index:idx_purchase_invoice_paid_amount"`
	RemainingAmount decimal.Decimal `gorm:"type:decimal(20,4);not null;default:0"`

	// Payment Terms
	PaymentTermDays int     `gorm:"default:30"` // e.g., NET 30
	Notes           *string `gorm:"type:text"`

	// Status & Workflow
	Status        PurchaseInvoiceStatus `gorm:"type:varchar(20);not null;default:'DRAFT';index:idx_purchase_invoice_status"`
	PaymentStatus PaymentStatus         `gorm:"type:varchar(20);not null;default:'UNPAID';index:idx_purchase_invoice_payment_status"`

	// Approval Workflow
	ApprovedBy *string    `gorm:"type:varchar(255)"`
	ApprovedAt *time.Time `gorm:"type:timestamp"`
	RejectedBy *string    `gorm:"type:varchar(255)"`
	RejectedAt *time.Time `gorm:"type:timestamp"`
	RejectedReason *string `gorm:"type:text"`

	// Tax Document (Faktur Pajak)
	TaxInvoiceNumber *string    `gorm:"type:varchar(100);uniqueIndex:idx_tax_invoice_number"` // Nomor Faktur Pajak
	TaxInvoiceDate   *time.Time `gorm:"type:timestamp"`

	// Audit Fields
	CreatedBy string  `gorm:"type:varchar(255);not null"`
	UpdatedBy *string `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_purchase_invoice_created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_purchase_invoice_deleted_at"`

	// Relations
	Tenant        Tenant                  `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company       Company                 `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Supplier      Supplier                `gorm:"foreignKey:SupplierID;constraint:OnDelete:RESTRICT"`
	PurchaseOrder *PurchaseOrder          `gorm:"foreignKey:PurchaseOrderID"`
	GoodsReceipt  *GoodsReceipt           `gorm:"foreignKey:GoodsReceiptID"`
	Items         []PurchaseInvoiceItem   `gorm:"foreignKey:PurchaseInvoiceID;constraint:OnDelete:CASCADE"`
	Payments      []PurchaseInvoicePayment `gorm:"foreignKey:PurchaseInvoiceID"`
	Creator       User                    `gorm:"foreignKey:CreatedBy;constraint:OnDelete:RESTRICT"`
	Updater       *User                   `gorm:"foreignKey:UpdatedBy"`
	Approver      *User                   `gorm:"foreignKey:ApprovedBy"`
	Rejecter      *User                   `gorm:"foreignKey:RejectedBy"`
}

// TableName specifies the table name for PurchaseInvoice model
func (PurchaseInvoice) TableName() string {
	return "purchase_invoices"
}

// BeforeCreate hook to generate UUID for ID field
func (pi *PurchaseInvoice) BeforeCreate(tx *gorm.DB) error {
	if pi.ID == "" {
		pi.ID = uuid.New().String()
	}
	return nil
}

// CalculateTotals calculates subtotal, tax, and total amounts from line items
func (pi *PurchaseInvoice) CalculateTotals() {
	var subtotal decimal.Decimal = decimal.Zero
	var discount decimal.Decimal = pi.DiscountAmount

	for _, item := range pi.Items {
		subtotal = subtotal.Add(item.LineTotal)
	}

	pi.SubtotalAmount = subtotal
	taxableAmount := subtotal.Sub(discount)
	pi.TaxAmount = taxableAmount.Mul(pi.TaxRate).Div(decimal.NewFromInt(100))
	pi.TotalAmount = taxableAmount.Add(pi.TaxAmount)
	pi.RemainingAmount = pi.TotalAmount.Sub(pi.PaidAmount)
}

// UpdatePaymentStatus updates payment status based on paid amount
func (pi *PurchaseInvoice) UpdatePaymentStatus() {
	if pi.PaidAmount.IsZero() {
		// Check if overdue
		if time.Now().After(pi.DueDate) && pi.Status != PurchaseInvoiceStatusPaid {
			pi.PaymentStatus = PaymentStatusOverdue
		} else {
			pi.PaymentStatus = PaymentStatusUnpaid
		}
	} else if pi.PaidAmount.GreaterThanOrEqual(pi.TotalAmount) {
		pi.PaymentStatus = PaymentStatusPaid
		if pi.Status != PurchaseInvoiceStatusPaid {
			pi.Status = PurchaseInvoiceStatusPaid
		}
	} else {
		pi.PaymentStatus = PaymentStatusPartial
	}
}

// PurchaseInvoiceItem represents line items in a purchase invoice
type PurchaseInvoiceItem struct {
	// Primary Key
	ID string `gorm:"type:varchar(255);primaryKey"`

	// Foreign Keys
	PurchaseInvoiceID string  `gorm:"type:varchar(255);not null;index:idx_purchase_invoice_item_invoice"`
	PurchaseOrderItemID *string `gorm:"type:varchar(255);index:idx_purchase_invoice_item_po_item"`
	GoodsReceiptItemID  *string `gorm:"type:varchar(255);index:idx_purchase_invoice_item_gr_item"`

	// Product Reference
	ProductID   string  `gorm:"type:varchar(255);not null;index:idx_purchase_invoice_item_product"`
	ProductCode string  `gorm:"type:varchar(100);not null"` // Denormalized
	ProductName string  `gorm:"type:varchar(255);not null"` // Denormalized

	// Unit & Quantity
	UnitID   string          `gorm:"type:varchar(255);not null"`
	UnitName string          `gorm:"type:varchar(50);not null"` // Denormalized
	Quantity decimal.Decimal `gorm:"type:decimal(20,4);not null"`

	// Pricing
	UnitPrice      decimal.Decimal `gorm:"type:decimal(20,4);not null"`
	DiscountAmount decimal.Decimal `gorm:"type:decimal(20,4);default:0"`
	DiscountPct    decimal.Decimal `gorm:"type:decimal(5,2);default:0"` // Optional percentage
	TaxAmount      decimal.Decimal `gorm:"type:decimal(20,4);default:0"`
	LineTotal      decimal.Decimal `gorm:"type:decimal(20,4);not null"` // (Qty * Price) - Discount + Tax

	// Notes
	Notes *string `gorm:"type:text"`

	// Audit Fields
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	PurchaseInvoice   PurchaseInvoice   `gorm:"foreignKey:PurchaseInvoiceID;constraint:OnDelete:CASCADE"`
	PurchaseOrderItem *PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderItemID"`
	GoodsReceiptItem  *GoodsReceiptItem  `gorm:"foreignKey:GoodsReceiptItemID"`
	Product           Product            `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	Unit              ProductUnit        `gorm:"foreignKey:UnitID;constraint:OnDelete:RESTRICT"`
}

// TableName specifies the table name for PurchaseInvoiceItem model
func (PurchaseInvoiceItem) TableName() string {
	return "purchase_invoice_items"
}

// BeforeCreate hook to generate UUID for ID field
func (pii *PurchaseInvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if pii.ID == "" {
		pii.ID = uuid.New().String()
	}
	return nil
}

// CalculateLineTotal calculates the line total: (Quantity * UnitPrice) - Discount + Tax
func (pii *PurchaseInvoiceItem) CalculateLineTotal() {
	subtotal := pii.Quantity.Mul(pii.UnitPrice)
	pii.LineTotal = subtotal.Sub(pii.DiscountAmount).Add(pii.TaxAmount)
}

// PurchaseInvoicePayment represents payment transactions against a purchase invoice
type PurchaseInvoicePayment struct {
	// Primary Key
	ID string `gorm:"type:varchar(255);primaryKey"`

	// Foreign Keys
	TenantID          string `gorm:"type:varchar(255);not null;index:idx_purchase_payment_tenant"`
	CompanyID         string `gorm:"type:varchar(255);not null;index:idx_purchase_payment_company"`
	PurchaseInvoiceID string `gorm:"type:varchar(255);not null;index:idx_purchase_payment_invoice"`

	// Payment Information
	PaymentNumber string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_purchase_payment_number"`
	PaymentDate   time.Time       `gorm:"type:timestamp;not null;index:idx_purchase_payment_date"`
	Amount        decimal.Decimal `gorm:"type:decimal(20,4);not null"`
	PaymentMethod PaymentMethod   `gorm:"type:varchar(20);not null;index:idx_purchase_payment_method"`
	Reference     *string         `gorm:"type:varchar(255)"` // Bank transfer reference, check number, etc.

	// Bank Account (optional)
	BankAccountID *string `gorm:"type:varchar(255);index:idx_purchase_payment_bank"`

	// Notes
	Notes *string `gorm:"type:text"`

	// Audit Fields
	CreatedBy string  `gorm:"type:varchar(255);not null"`
	UpdatedBy *string `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_purchase_payment_deleted_at"`

	// Relations
	Tenant          Tenant          `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company         Company         `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	PurchaseInvoice PurchaseInvoice `gorm:"foreignKey:PurchaseInvoiceID;constraint:OnDelete:RESTRICT"`
	BankAccount     *CompanyBank    `gorm:"foreignKey:BankAccountID"`
	Creator         User            `gorm:"foreignKey:CreatedBy;constraint:OnDelete:RESTRICT"`
	Updater         *User           `gorm:"foreignKey:UpdatedBy"`
}

// TableName specifies the table name for PurchaseInvoicePayment model
func (PurchaseInvoicePayment) TableName() string {
	return "purchase_invoice_payments"
}

// BeforeCreate hook to generate UUID for ID field
func (pip *PurchaseInvoicePayment) BeforeCreate(tx *gorm.DB) error {
	if pip.ID == "" {
		pip.ID = uuid.New().String()
	}
	return nil
}
