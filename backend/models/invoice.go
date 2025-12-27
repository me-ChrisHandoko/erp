// Package models - Invoice and Payment models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Invoice - Customer invoice
type Invoice struct {
	ID              string          `gorm:"type:varchar(255);primaryKey"`
	TenantID        string          `gorm:"type:varchar(255);not null;index"`
	CompanyID       string          `gorm:"type:varchar(255);not null;index:idx_company_invoice;uniqueIndex:idx_company_invoice_number"`
	InvoiceNumber   string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_invoice_number"`
	InvoiceDate     time.Time       `gorm:"type:timestamp;not null;index"`
	DueDate         time.Time       `gorm:"type:timestamp;not null;index"`
	CustomerID      string          `gorm:"type:varchar(255);not null;index"`
	SalesOrderID    *string         `gorm:"type:varchar(255);index"`
	DeliveryID      *string         `gorm:"type:varchar(255);index"`
	Subtotal        decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	DiscountAmount  decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	TaxAmount       decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	PaidAmount      decimal.Decimal `gorm:"type:decimal(15,2);default:0;index"`
	PaymentStatus   PaymentStatus   `gorm:"type:varchar(20);default:'UNPAID';index"`
	Notes           *string         `gorm:"type:text"`
	FakturPajakNo   *string         `gorm:"type:varchar(100);uniqueIndex"` // Tax invoice number
	FakturPajakDate *time.Time      `gorm:"type:timestamp"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant      Tenant         `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company     Company        `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Customer    Customer       `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT"`
	SalesOrder  *SalesOrder    `gorm:"foreignKey:SalesOrderID"`
	Delivery    *Delivery      `gorm:"foreignKey:DeliveryID"`
	Items       []InvoiceItem  `gorm:"foreignKey:InvoiceID"`
	Payments    []Payment      `gorm:"foreignKey:InvoiceID"`
}

// TableName specifies the table name for Invoice model
func (Invoice) TableName() string {
	return "invoices"
}

// BeforeCreate hook to generate UUID for ID field
func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.New().String()
	}
	return nil
}

// InvoiceItem - Invoice line items
type InvoiceItem struct {
	ID               string          `gorm:"type:varchar(255);primaryKey"`
	InvoiceID        string          `gorm:"type:varchar(255);not null;index"`
	SalesOrderItemID *string         `gorm:"type:varchar(255);index"`
	DeliveryItemID   *string         `gorm:"type:varchar(255);index"`
	ProductID        string          `gorm:"type:varchar(255);not null;index"`
	ProductUnitID    *string         `gorm:"type:varchar(255);index"`
	Quantity         decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	UnitPrice        decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	DiscountPct      decimal.Decimal `gorm:"type:decimal(5,2);default:0"`
	DiscountAmt      decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	Subtotal         decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	Notes            *string         `gorm:"type:text"`
	CreatedAt        time.Time       `gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Invoice         Invoice          `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE"`
	SalesOrderItem  *SalesOrderItem  `gorm:"foreignKey:SalesOrderItemID"`
	DeliveryItem    *DeliveryItem    `gorm:"foreignKey:DeliveryItemID"`
	Product         Product          `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	ProductUnit     *ProductUnit     `gorm:"foreignKey:ProductUnitID"`
}

// TableName specifies the table name for InvoiceItem model
func (InvoiceItem) TableName() string {
	return "invoice_items"
}

// BeforeCreate hook to generate UUID for ID field
func (ii *InvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if ii.ID == "" {
		ii.ID = uuid.New().String()
	}
	return nil
}

// Payment - Customer payment against invoice
type Payment struct {
	ID            string          `gorm:"type:varchar(255);primaryKey"`
	TenantID      string          `gorm:"type:varchar(255);not null;index"`
	PaymentNumber string          `gorm:"type:varchar(100);not null;uniqueIndex"`
	PaymentDate   time.Time       `gorm:"type:timestamp;not null;index"`
	CustomerID    string          `gorm:"type:varchar(255);not null;index"`
	InvoiceID     string          `gorm:"type:varchar(255);not null;index"`
	Amount        decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	PaymentMethod PaymentMethod   `gorm:"type:varchar(20);not null;index"`
	Reference     *string         `gorm:"type:varchar(100)"` // Transfer reference, check number, etc.
	BankAccountID *string         `gorm:"type:varchar(255);index"`
	Notes         *string         `gorm:"type:text"`
	ReceivedBy    *string         `gorm:"type:varchar(255)"` // User who received payment
	ReceivedAt    *time.Time      `gorm:"type:timestamp"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant      Tenant       `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Customer    Customer     `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT"`
	Invoice     Invoice      `gorm:"foreignKey:InvoiceID;constraint:OnDelete:RESTRICT"`
	BankAccount *CompanyBank `gorm:"foreignKey:BankAccountID"`
	Checks      []PaymentCheck `gorm:"foreignKey:PaymentID"`
}

// TableName specifies the table name for Payment model
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate hook to generate UUID for ID field
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// PaymentCheck - Check/Giro tracking for payment
type PaymentCheck struct {
	ID          string      `gorm:"type:varchar(255);primaryKey"`
	PaymentID   string      `gorm:"type:varchar(255);not null;index"`
	CheckNumber string      `gorm:"type:varchar(100);not null;uniqueIndex"`
	CheckDate   time.Time   `gorm:"type:timestamp;not null"`
	DueDate     time.Time   `gorm:"type:timestamp;not null;index"`
	Amount      decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	BankName    string      `gorm:"type:varchar(255);not null"`
	AccountName *string     `gorm:"type:varchar(255)"`
	Status      CheckStatus `gorm:"type:varchar(20);default:'ISSUED';index"`
	ClearedDate *time.Time  `gorm:"type:timestamp"`
	BouncedDate *time.Time  `gorm:"type:timestamp"`
	BouncedNote *string     `gorm:"type:text"`
	Notes       *string     `gorm:"type:text"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`

	// Relations
	Payment Payment `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PaymentCheck model
func (PaymentCheck) TableName() string {
	return "payment_checks"
}

// BeforeCreate hook to generate UUID for ID field
func (pc *PaymentCheck) BeforeCreate(tx *gorm.DB) error {
	if pc.ID == "" {
		pc.ID = uuid.New().String()
	}
	return nil
}
