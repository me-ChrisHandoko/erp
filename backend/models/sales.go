// Package models - Sales Order models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SalesOrder - Sales order header
type SalesOrder struct {
	ID               string            `gorm:"type:varchar(255);primaryKey"`
	TenantID         string            `gorm:"type:varchar(255);not null;index"`
	CompanyID        string            `gorm:"type:varchar(255);not null;index:idx_company_sales_order;uniqueIndex:idx_company_so_number"`
	SONumber         string            `gorm:"type:varchar(100);not null;uniqueIndex:idx_company_so_number"`
	SODate           time.Time         `gorm:"type:timestamp;not null;index"`
	CustomerID       string            `gorm:"type:varchar(255);not null;index"`
	Status           SalesOrderStatus  `gorm:"type:varchar(20);default:'DRAFT';index"`
	Subtotal         decimal.Decimal   `gorm:"type:decimal(15,2);default:0"`
	DiscountAmount   decimal.Decimal   `gorm:"type:decimal(15,2);default:0"`
	TaxAmount        decimal.Decimal   `gorm:"type:decimal(15,2);default:0"`
	TotalAmount      decimal.Decimal   `gorm:"type:decimal(15,2);default:0"`
	Notes            *string           `gorm:"type:text"`
	DeliveryAddress  *string           `gorm:"type:text"`
	DeliveryDate     *time.Time        `gorm:"type:timestamp"`
	SalespersonID    *string           `gorm:"type:varchar(255);index"`
	ApprovedBy       *string           `gorm:"type:varchar(255)"`
	ApprovedAt       *time.Time        `gorm:"type:timestamp"`
	CancelledBy      *string           `gorm:"type:varchar(255)"`
	CancelledAt      *time.Time        `gorm:"type:timestamp"`
	CancellationNote *string           `gorm:"type:text"`
	CreatedAt        time.Time         `gorm:"autoCreateTime"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime"`

	// Relations
	Tenant      Tenant            `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company     Company           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Customer    Customer          `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT"`
	Salesperson *User             `gorm:"foreignKey:SalespersonID"`
	Items       []SalesOrderItem  `gorm:"foreignKey:SalesOrderID"`
	// Note: Invoices, Deliveries will reference this SO
}

// TableName specifies the table name for SalesOrder model
func (SalesOrder) TableName() string {
	return "sales_orders"
}

// BeforeCreate hook to generate UUID for ID field
func (so *SalesOrder) BeforeCreate(tx *gorm.DB) error {
	if so.ID == "" {
		so.ID = uuid.New().String()
	}
	return nil
}

// SalesOrderItem - Sales order line items
type SalesOrderItem struct {
	ID            string          `gorm:"type:varchar(255);primaryKey"`
	SalesOrderID  string          `gorm:"type:varchar(255);not null;index"`
	ProductID     string          `gorm:"type:varchar(255);not null;index"`
	ProductUnitID *string         `gorm:"type:varchar(255);index"` // NULL = base unit
	Quantity      decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	UnitPrice     decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	DiscountPct   decimal.Decimal `gorm:"type:decimal(5,2);default:0"`
	DiscountAmt   decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	Subtotal      decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	Notes         *string         `gorm:"type:text"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`

	// Relations
	SalesOrder  SalesOrder   `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE"`
	Product     Product      `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	ProductUnit *ProductUnit `gorm:"foreignKey:ProductUnitID"`
}

// TableName specifies the table name for SalesOrderItem model
func (SalesOrderItem) TableName() string {
	return "sales_order_items"
}

// BeforeCreate hook to generate UUID for ID field
func (soi *SalesOrderItem) BeforeCreate(tx *gorm.DB) error {
	if soi.ID == "" {
		soi.ID = uuid.New().String()
	}
	return nil
}
