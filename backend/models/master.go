// Package models - Customer and Supplier master data
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Customer - Customer master with outstanding tracking
type Customer struct {
	ID                 string          `gorm:"type:varchar(255);primaryKey"`
	TenantID           string          `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_customer_tenant_code"`
	Code               string          `gorm:"type:varchar(100);not null;index;uniqueIndex:idx_customer_tenant_code"`
	Name               string          `gorm:"type:varchar(255);not null;index"`
	Type               *string         `gorm:"type:varchar(50)"` // RETAIL, WHOLESALE, DISTRIBUTOR
	Phone              *string         `gorm:"type:varchar(50)"`
	Email              *string         `gorm:"type:varchar(255)"`
	Address            *string         `gorm:"type:text"`
	City               *string         `gorm:"type:varchar(100)"`
	Province           *string         `gorm:"type:varchar(100)"`
	PostalCode         *string         `gorm:"type:varchar(50)"`
	NPWP               *string         `gorm:"type:varchar(50)"`
	IsPKP              bool            `gorm:"default:false"`
	ContactPerson      *string         `gorm:"type:varchar(255)"`
	ContactPhone       *string         `gorm:"type:varchar(50)"`
	PaymentTerm        int             `gorm:"type:int;default:0"` // Days (0 = cash)
	CreditLimit        decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	CurrentOutstanding decimal.Decimal `gorm:"type:decimal(15,2);default:0;index"`
	OverdueAmount      decimal.Decimal `gorm:"type:decimal(15,2);default:0;index"`
	LastTransactionAt  *time.Time      `gorm:"type:timestamp"`
	Notes              *string         `gorm:"type:text"`
	IsActive           bool            `gorm:"default:true"`
	CreatedAt          time.Time       `gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	// Note: SalesOrders, Invoices, Payments, PriceList will be added in Phase 3
}

// TableName specifies the table name for Customer model
func (Customer) TableName() string {
	return "customers"
}

// BeforeCreate hook to generate CUID for ID field
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = cuid.New()
	}
	return nil
}

// Supplier - Supplier master with outstanding tracking
type Supplier struct {
	ID                 string          `gorm:"type:varchar(255);primaryKey"`
	TenantID           string          `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_supplier_tenant_code"`
	Code               string          `gorm:"type:varchar(100);not null;index;uniqueIndex:idx_supplier_tenant_code"`
	Name               string          `gorm:"type:varchar(255);not null;index"`
	Type               *string         `gorm:"type:varchar(50)"` // MANUFACTURER, DISTRIBUTOR, WHOLESALER
	Phone              *string         `gorm:"type:varchar(50)"`
	Email              *string         `gorm:"type:varchar(255)"`
	Address            *string         `gorm:"type:text"`
	City               *string         `gorm:"type:varchar(100)"`
	Province           *string         `gorm:"type:varchar(100)"`
	PostalCode         *string         `gorm:"type:varchar(50)"`
	NPWP               *string         `gorm:"type:varchar(50)"`
	IsPKP              bool            `gorm:"default:false"`
	ContactPerson      *string         `gorm:"type:varchar(255)"`
	ContactPhone       *string         `gorm:"type:varchar(50)"`
	PaymentTerm        int             `gorm:"type:int;default:0"` // Days (0 = cash)
	CreditLimit        decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	CurrentOutstanding decimal.Decimal `gorm:"type:decimal(15,2);default:0;index"`
	OverdueAmount      decimal.Decimal `gorm:"type:decimal(15,2);default:0;index"`
	LastTransactionAt  *time.Time      `gorm:"type:timestamp"`
	Notes              *string         `gorm:"type:text"`
	IsActive           bool            `gorm:"default:true"`
	CreatedAt          time.Time       `gorm:"autoCreateTime"`
	UpdatedAt          time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant          Tenant            `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	ProductSuppliers []ProductSupplier `gorm:"foreignKey:SupplierID"`
	// Note: PurchaseOrders, GoodsReceipts, SupplierPayments will be added in Phase 3
}

// TableName specifies the table name for Supplier model
func (Supplier) TableName() string {
	return "suppliers"
}

// BeforeCreate hook to generate CUID for ID field
func (s *Supplier) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = cuid.New()
	}
	return nil
}
