// Package models - Delivery models
package models

import (
	"time"

	"github.com/lucsky/cuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Delivery - Delivery order header
type Delivery struct {
	ID                string          `gorm:"type:varchar(255);primaryKey"`
	TenantID          string          `gorm:"type:varchar(255);not null;index"`
	DeliveryNumber    string          `gorm:"type:varchar(100);not null;uniqueIndex"`
	DeliveryDate      time.Time       `gorm:"type:timestamp;not null;index"`
	SalesOrderID      string          `gorm:"type:varchar(255);not null;index"`
	WarehouseID       string          `gorm:"type:varchar(255);not null;index"` // Source warehouse
	CustomerID        string          `gorm:"type:varchar(255);not null;index"`
	Type              DeliveryType    `gorm:"type:varchar(20);default:'NORMAL';index"`
	Status            DeliveryStatus  `gorm:"type:varchar(20);default:'PREPARED';index"`
	DeliveryAddress   *string         `gorm:"type:text"`
	DriverName        *string         `gorm:"type:varchar(255)"`
	VehicleNumber     *string         `gorm:"type:varchar(50)"`
	DepartureTime     *time.Time      `gorm:"type:timestamp"`
	ArrivalTime       *time.Time      `gorm:"type:timestamp"`
	ReceivedBy        *string         `gorm:"type:varchar(255)"` // Customer's receiver name
	ReceivedAt        *time.Time      `gorm:"type:timestamp"`
	SignatureURL      *string         `gorm:"type:varchar(500)"` // POD signature image
	PhotoURL          *string         `gorm:"type:varchar(500)"` // POD photo
	TTNKNumber        *string         `gorm:"type:varchar(100)"` // Expedition tracking number
	ExpeditionService *string         `gorm:"type:varchar(100)"` // JNE, Sicepat, etc.
	Notes             *string         `gorm:"type:text"`
	CreatedAt         time.Time       `gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant      Tenant          `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	SalesOrder  SalesOrder      `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:RESTRICT"`
	Warehouse   Warehouse       `gorm:"foreignKey:WarehouseID;constraint:OnDelete:RESTRICT"`
	Customer    Customer        `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT"`
	Items       []DeliveryItem  `gorm:"foreignKey:DeliveryID"`
	// Note: Invoices may reference this delivery
}

// TableName specifies the table name for Delivery model
func (Delivery) TableName() string {
	return "deliveries"
}

// BeforeCreate hook to generate CUID for ID field
func (d *Delivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = cuid.New()
	}
	return nil
}

// DeliveryItem - Delivery line items with batch tracking
type DeliveryItem struct {
	ID               string          `gorm:"type:varchar(255);primaryKey"`
	DeliveryID       string          `gorm:"type:varchar(255);not null;index"`
	SalesOrderItemID string          `gorm:"type:varchar(255);not null;index"`
	ProductID        string          `gorm:"type:varchar(255);not null;index"`
	ProductUnitID    *string         `gorm:"type:varchar(255);index"`
	BatchID          *string         `gorm:"type:varchar(255);index"` // Required if product.isBatchTracked
	Quantity         decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	Notes            *string         `gorm:"type:text"`
	CreatedAt        time.Time       `gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Delivery       Delivery        `gorm:"foreignKey:DeliveryID;constraint:OnDelete:CASCADE"`
	SalesOrderItem SalesOrderItem  `gorm:"foreignKey:SalesOrderItemID;constraint:OnDelete:RESTRICT"`
	Product        Product         `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT"`
	ProductUnit    *ProductUnit    `gorm:"foreignKey:ProductUnitID"`
	Batch          *ProductBatch   `gorm:"foreignKey:BatchID"`
}

// TableName specifies the table name for DeliveryItem model
func (DeliveryItem) TableName() string {
	return "delivery_items"
}

// BeforeCreate hook to generate CUID for ID field
func (di *DeliveryItem) BeforeCreate(tx *gorm.DB) error {
	if di.ID == "" {
		di.ID = cuid.New()
	}
	return nil
}
