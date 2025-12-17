// Package models - Multi-tenancy and subscription models
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Tenant - Represents 1 PT/CV subscription instance
type Tenant struct {
	ID             string       `gorm:"type:varchar(255);primaryKey"`
	CompanyID      string       `gorm:"type:varchar(255);uniqueIndex;not null;index"`
	SubscriptionID *string      `gorm:"type:varchar(255);index"`
	Status         TenantStatus `gorm:"type:varchar(20);default:'TRIAL';index"`
	TrialEndsAt    *time.Time   `gorm:"type:timestamp"`
	Notes          *string      `gorm:"type:text"`
	CreatedAt      time.Time    `gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `gorm:"autoUpdateTime"`

	// Relations
	Company      Company       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Subscription *Subscription `gorm:"foreignKey:SubscriptionID"`
	Users        []UserTenant  `gorm:"foreignKey:TenantID"`
	// Note: Other relations (SalesOrders, Invoices, etc.) will be added in their respective model files
}

// TableName specifies the table name for Tenant model
func (Tenant) TableName() string {
	return "tenants"
}

// BeforeCreate hook to generate CUID for ID field
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = cuid.New()
	}
	return nil
}

// Subscription - Billing & payment tracking with custom pricing
type Subscription struct {
	ID                 string             `gorm:"type:varchar(255);primaryKey"`
	Price              decimal.Decimal    `gorm:"type:decimal(15,2);default:300000"` // Monthly price per PT/CV
	BillingCycle       string             `gorm:"type:varchar(20);default:'MONTHLY'"` // MONTHLY, QUARTERLY, YEARLY
	Status             SubscriptionStatus `gorm:"type:varchar(20);default:'ACTIVE';index"`
	CurrentPeriodStart time.Time          `gorm:"type:timestamp;not null"`
	CurrentPeriodEnd   time.Time          `gorm:"type:timestamp;not null"`
	NextBillingDate    time.Time          `gorm:"type:timestamp;not null;index"`
	PaymentMethod      *string            `gorm:"type:varchar(50)"` // "TRANSFER", "VA", "CREDIT_CARD", "QRIS"
	LastPaymentDate    *time.Time         `gorm:"type:timestamp"`
	LastPaymentAmount  *decimal.Decimal   `gorm:"type:decimal(15,2)"`
	GracePeriodEnds    *time.Time         `gorm:"type:timestamp;index"`
	AutoRenew          bool               `gorm:"default:true"`
	CancelledAt        *time.Time         `gorm:"type:timestamp"`
	CancellationReason *string            `gorm:"type:text"`
	CreatedAt          time.Time          `gorm:"autoCreateTime"`
	UpdatedAt          time.Time          `gorm:"autoUpdateTime"`

	// Relations
	Tenants  []Tenant              `gorm:"foreignKey:SubscriptionID"`
	Payments []SubscriptionPayment `gorm:"foreignKey:SubscriptionID"`
}

// TableName specifies the table name for Subscription model
func (Subscription) TableName() string {
	return "subscriptions"
}

// BeforeCreate hook to generate CUID for ID field
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = cuid.New()
	}
	return nil
}

// SubscriptionPayment - Payment history
type SubscriptionPayment struct {
	ID             string                    `gorm:"type:varchar(255);primaryKey"`
	SubscriptionID string                    `gorm:"type:varchar(255);not null;index"`
	Amount         decimal.Decimal           `gorm:"type:decimal(15,2);not null"`
	PaymentDate    time.Time                 `gorm:"type:timestamp;not null;index"`
	PaymentMethod  string                    `gorm:"type:varchar(50);not null"` // "TRANSFER", "VA_BCA", "CC_VISA", "QRIS"
	Status         SubscriptionPaymentStatus `gorm:"type:varchar(20);default:'PENDING';index"`
	Reference      *string                   `gorm:"type:varchar(255)"` // Transfer proof, transaction ID, VA number
	InvoiceNumber  *string                   `gorm:"type:varchar(100);uniqueIndex"` // Invoice for this payment
	PeriodStart    time.Time                 `gorm:"type:timestamp;not null"`
	PeriodEnd      time.Time                 `gorm:"type:timestamp;not null"`
	PaidAt         *time.Time                `gorm:"type:timestamp;index"` // Actual payment timestamp
	Notes          *string                   `gorm:"type:text"`
	CreatedAt      time.Time                 `gorm:"autoCreateTime"`
	UpdatedAt      time.Time                 `gorm:"autoUpdateTime"`

	// Relations
	Subscription Subscription `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for SubscriptionPayment model
func (SubscriptionPayment) TableName() string {
	return "subscription_payments"
}

// BeforeCreate hook to generate CUID for ID field
func (sp *SubscriptionPayment) BeforeCreate(tx *gorm.DB) error {
	if sp.ID == "" {
		sp.ID = cuid.New()
	}
	return nil
}
