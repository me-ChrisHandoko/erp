// Package models - Supplier Payment models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SupplierPayment - Payment to supplier
type SupplierPayment struct {
	ID              string          `gorm:"type:varchar(255);primaryKey"`
	TenantID        string          `gorm:"type:varchar(255);not null;index"`
	CompanyID       string          `gorm:"type:varchar(255);not null;index:idx_company_supplier_payment;uniqueIndex:idx_company_payment_number"`
	PaymentNumber   string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_payment_number"`
	PaymentDate     time.Time       `gorm:"type:timestamp;not null;index"`
	SupplierID      string          `gorm:"type:varchar(255);not null;index"`
	PurchaseOrderID *string         `gorm:"type:varchar(255);index"` // Optional PO reference
	Amount          decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	PaymentMethod   PaymentMethod   `gorm:"type:varchar(20);not null;index"`
	Reference       *string         `gorm:"type:varchar(100)"` // Transfer reference, check number, etc.
	BankAccountID   *string         `gorm:"type:varchar(255);index"`
	Notes           *string         `gorm:"type:text"`
	ApprovedBy      *string         `gorm:"type:varchar(255)"` // User who approved payment
	ApprovedAt      *time.Time      `gorm:"type:timestamp"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant        Tenant        `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company       Company       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Supplier      Supplier      `gorm:"foreignKey:SupplierID;constraint:OnDelete:RESTRICT"`
	PurchaseOrder *PurchaseOrder `gorm:"foreignKey:PurchaseOrderID"`
	BankAccount   *CompanyBank  `gorm:"foreignKey:BankAccountID"`
}

// TableName specifies the table name for SupplierPayment model
func (SupplierPayment) TableName() string {
	return "supplier_payments"
}

// BeforeCreate hook to generate UUID for ID field
func (sp *SupplierPayment) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == "" {
		sp.ID = uuid.New().String()
	}
	return nil
}
