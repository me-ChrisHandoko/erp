// Package models - Product and inventory models
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Product - Product master with multi-unit support
type Product struct {
	ID             string          `gorm:"type:varchar(255);primaryKey"`
	TenantID       string          `gorm:"type:varchar(255);not null;index"`
	CompanyID      string          `gorm:"type:varchar(255);not null;index:idx_company_product;uniqueIndex:idx_company_product_code"`
	Code           string          `gorm:"type:varchar(100);not null;index;uniqueIndex:idx_company_product_code"` // SKU
	Name           string          `gorm:"type:varchar(255);not null;index"`
	Category       *string         `gorm:"type:varchar(100)"`
	BaseUnit       string          `gorm:"type:varchar(20);default:'PCS'"` // Unit terkecil
	BaseCost       decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	BasePrice      decimal.Decimal `gorm:"type:decimal(15,2);default:0"`
	CurrentStock   decimal.Decimal `gorm:"type:decimal(15,3);default:0"` // DEPRECATED: Use WarehouseStock
	MinimumStock   decimal.Decimal `gorm:"type:decimal(15,3);default:0"`
	Description    *string         `gorm:"type:text"`
	Barcode        *string         `gorm:"type:varchar(100);uniqueIndex"`
	IsBatchTracked bool            `gorm:"default:false;index"` // Requires batch/lot tracking
	IsPerishable   bool            `gorm:"default:false"`       // Has expiry date
	IsActive       bool            `gorm:"default:true"`
	CreatedAt      time.Time       `gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Tenant             Tenant            `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
	Company            Company           `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
	Units              []ProductUnit     `gorm:"foreignKey:ProductID"`
	PriceList          []PriceList       `gorm:"foreignKey:ProductID"`
	Batches            []ProductBatch    `gorm:"foreignKey:ProductID"`
	ProductSuppliers   []ProductSupplier `gorm:"foreignKey:ProductID"`
	WarehouseStocks    []WarehouseStock  `gorm:"foreignKey:ProductID"`
	// Note: SalesOrderItems, InvoiceItems, etc. will be added in Phase 3
}

// TableName specifies the table name for Product model
func (Product) TableName() string {
	return "products"
}

// BeforeCreate hook to generate UUID for ID field
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// ProductBatch - Batch/lot tracking for sembako (food items)
// CRITICAL for perishable items with expiry dates
type ProductBatch struct {
	ID               string          `gorm:"type:varchar(255);primaryKey"`
	BatchNumber      string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_batch_product"` // e.g., "BATCH-2025-001"
	ProductID        string          `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_batch_product"`
	WarehouseStockID string          `gorm:"type:varchar(255);not null;index"`
	ManufactureDate  *time.Time      `gorm:"type:timestamp"`
	ExpiryDate       *time.Time      `gorm:"type:timestamp;index"` // CRITICAL for sembako
	Quantity         decimal.Decimal `gorm:"type:decimal(15,3);not null"`
	SupplierID       *string         `gorm:"type:varchar(255)"`
	GoodsReceiptID   *string         `gorm:"type:varchar(255);index"`
	ReceiptDate      time.Time       `gorm:"type:timestamp;not null"`
	Status           BatchStatus     `gorm:"type:varchar(20);default:'AVAILABLE';index"`
	QualityStatus    *string         `gorm:"type:varchar(20);default:'GOOD'"` // GOOD, DAMAGED, QUARANTINE
	ReferenceNumber  *string         `gorm:"type:varchar(100)"`               // Supplier's batch/lot number
	Notes            *string         `gorm:"type:text"`
	CreatedAt        time.Time       `gorm:"autoCreateTime"`
	UpdatedAt        time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Product        Product        `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	WarehouseStock WarehouseStock `gorm:"foreignKey:WarehouseStockID;constraint:OnDelete:CASCADE"`
	// Note: GoodsReceipt, DeliveryItems will be added in Phase 3
}

// TableName specifies the table name for ProductBatch model
func (ProductBatch) TableName() string {
	return "product_batches"
}

// BeforeCreate hook to generate UUID for ID field
func (pb *ProductBatch) BeforeCreate(tx *gorm.DB) error {
	if pb.ID == "" {
		pb.ID = uuid.New().String()
	}
	return nil
}

// ProductUnit - Multi-unit conversion for distribusi
// Allows product to be sold/bought in different units (PCS, KARTON, LUSIN, etc.)
type ProductUnit struct {
	ID             string           `gorm:"type:varchar(255);primaryKey"`
	ProductID      string           `gorm:"type:varchar(255);not null;index;uniqueIndex:idx_product_unit"`
	UnitName       string           `gorm:"type:varchar(50);not null;uniqueIndex:idx_product_unit"` // "PCS", "KARTON", "LUSIN"
	ConversionRate decimal.Decimal  `gorm:"type:decimal(15,3);not null"`                            // Konversi ke base unit (1 KARTON = 24 PCS)
	IsBaseUnit     bool             `gorm:"default:false"`
	BuyPrice       *decimal.Decimal `gorm:"type:decimal(15,2)"` // Harga beli per unit
	SellPrice      *decimal.Decimal `gorm:"type:decimal(15,2)"` // Harga jual per unit
	Barcode        *string          `gorm:"type:varchar(100);index"` // Barcode per unit
	SKU            *string          `gorm:"type:varchar(100)"`       // SKU khusus untuk unit ini
	Weight         *decimal.Decimal `gorm:"type:decimal(10,3)"`      // Berat per unit (kg)
	Volume         *decimal.Decimal `gorm:"type:decimal(10,3)"`      // Volume per unit (m³)
	Description    *string          `gorm:"type:text"`
	IsActive       bool             `gorm:"default:true"`
	CreatedAt      time.Time        `gorm:"autoCreateTime"`
	UpdatedAt      time.Time        `gorm:"autoUpdateTime"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for ProductUnit model
func (ProductUnit) TableName() string {
	return "product_units"
}

// BeforeCreate hook to generate UUID for ID field
func (pu *ProductUnit) BeforeCreate(tx *gorm.DB) error {
	if pu.ID == "" {
		pu.ID = uuid.New().String()
	}
	return nil
}

// PriceList - Multi-customer pricing
// Allows different pricing per customer or default pricing
type PriceList struct {
	ID            string          `gorm:"type:varchar(255);primaryKey"`
	ProductID     string          `gorm:"type:varchar(255);not null;index:idx_product_customer"`
	CustomerID    *string         `gorm:"type:varchar(255);index:idx_product_customer"` // NULL = default price
	Price         decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	MinQty        decimal.Decimal `gorm:"type:decimal(15,3);default:0"`
	EffectiveFrom time.Time       `gorm:"type:timestamp;not null"`
	EffectiveTo   *time.Time      `gorm:"type:timestamp"`
	IsActive      bool            `gorm:"default:true"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	// Note: Customer relation will be added in master.go
}

// TableName specifies the table name for PriceList model
func (PriceList) TableName() string {
	return "price_list"
}

// BeforeCreate hook to generate UUID for ID field
func (pl *PriceList) BeforeCreate(tx *gorm.DB) error {
	if pl.ID == "" {
		pl.ID = uuid.New().String()
	}
	return nil
}

// ProductSupplier - Supplier-Product relationship
// Tracks which supplier supplies which product with pricing and lead time
type ProductSupplier struct {
	ID            string          `gorm:"type:varchar(255);primaryKey"`
	ProductID     string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_product_supplier"`
	SupplierID    string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_product_supplier"`
	SupplierPrice decimal.Decimal `gorm:"type:decimal(15,2);not null"`
	LeadTime      int             `gorm:"type:int;default:7"` // days
	IsPrimary     bool            `gorm:"default:false"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	// Note: Supplier relation will be added in master.go
}

// TableName specifies the table name for ProductSupplier model
func (ProductSupplier) TableName() string {
	return "product_suppliers"
}

// BeforeCreate hook to generate UUID for ID field
func (ps *ProductSupplier) BeforeCreate(tx *gorm.DB) error {
	if ps.ID == "" {
		ps.ID = uuid.New().String()
	}
	return nil
}
