// Package models - Cash Transaction (Buku Kas) models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CashTransaction - Cash book (Buku Kas) transaction tracking
type CashTransaction struct {
	ID                string              `gorm:"type:varchar(255);primaryKey"`
	TenantID          string              `gorm:"type:varchar(255);not null;index"`
	CompanyID         string              `gorm:"type:varchar(255);not null;index:idx_company_cash_transaction;uniqueIndex:idx_company_transaction_number"`
	TransactionNumber string              `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_transaction_number"`
	TransactionDate   time.Time           `gorm:"type:timestamp;not null;index"`
	Type              CashTransactionType `gorm:"type:varchar(20);not null;index"` // CASH_IN, CASH_OUT
	Category          CashCategory        `gorm:"type:varchar(50);not null;index"` // SALES, PURCHASE, EXPENSE, etc.
	Amount          decimal.Decimal      `gorm:"type:decimal(15,2);not null"`
	BalanceBefore   decimal.Decimal      `gorm:"type:decimal(15,2);not null"`
	BalanceAfter    decimal.Decimal      `gorm:"type:decimal(15,2);not null"`
	Description     string               `gorm:"type:varchar(500);not null"`
	ReferenceType   *string              `gorm:"type:varchar(50)"` // PAYMENT, SUPPLIER_PAYMENT, etc.
	ReferenceID     *string              `gorm:"type:varchar(255);index"`
	ReferenceNumber *string              `gorm:"type:varchar(100)"` // PAY-001, SUPP-PAY-001, etc.
	Notes           *string              `gorm:"type:text"`
	CreatedBy       *string              `gorm:"type:varchar(255)"` // User who created transaction
	CreatedAt       time.Time            `gorm:"autoCreateTime"`
	UpdatedAt       time.Time            `gorm:"autoUpdateTime"`

	// Relations
	Tenant  Tenant  `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for CashTransaction model
func (CashTransaction) TableName() string {
	return "cash_transactions"
}

// BeforeCreate hook to generate UUID for ID field
func (ct *CashTransaction) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == "" {
		ct.ID = uuid.New().String()
	}
	return nil
}
