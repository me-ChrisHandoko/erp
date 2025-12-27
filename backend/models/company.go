// Package models - Company profile and settings
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Company - Legal entity profile & settings (PT, CV, UD, Firma)
// CORRECT ARCHITECTURE: Company belongs to 1 Tenant (N Companies → 1 Tenant)
type Company struct {
	ID       string `gorm:"type:varchar(255);primaryKey"`
	TenantID string `gorm:"type:varchar(255);not null;index:idx_company_tenant"` // FK to tenants table

	// Legal Entity Information
	Name       string `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_company"` // Unique per tenant
	LegalName  string `gorm:"type:varchar(255);not null"`
	EntityType string `gorm:"type:varchar(50);default:'CV';check:entity_type IN ('PT','CV','UD','Firma')"` // CV, PT, UD, Firma

	// Address
	Address    string  `gorm:"type:text;not null"`
	City       string  `gorm:"type:varchar(255);not null"`
	Province   string  `gorm:"type:varchar(255);not null"`
	PostalCode *string `gorm:"type:varchar(50)"`
	Country    string  `gorm:"type:varchar(100);default:'Indonesia'"`

	// Contact
	Phone   string  `gorm:"type:varchar(50);not null"`
	Email   string  `gorm:"type:varchar(255);not null"`
	Website *string `gorm:"type:varchar(255)"`

	// Indonesian Tax Compliance
	NPWP              *string         `gorm:"type:varchar(50);uniqueIndex"` // Nomor Pokok Wajib Pajak
	NIB               *string         `gorm:"type:varchar(50)"`             // Nomor Induk Berusaha
	IsPKP             bool            `gorm:"default:false"`                // Pengusaha Kena Pajak
	PPNRate           decimal.Decimal `gorm:"type:decimal(5,2);default:11"` // PPN rate (11% in 2025)
	FakturPajakSeries *string         `gorm:"type:varchar(50)"`             // Series Faktur Pajak
	SPPKPNumber       *string         `gorm:"type:varchar(50)"`             // Surat Pengukuhan PKP

	// Branding
	LogoURL        *string `gorm:"type:varchar(255)"`
	PrimaryColor   *string `gorm:"type:varchar(20);default:'#1E40AF'"`
	SecondaryColor *string `gorm:"type:varchar(20);default:'#64748B'"`

	// Invoice Settings
	InvoicePrefix       string  `gorm:"type:varchar(20);default:'INV'"`
	InvoiceNumberFormat string  `gorm:"type:varchar(100);default:'{PREFIX}/{NUMBER}/{MONTH}/{YEAR}'"`
	InvoiceFooter       *string `gorm:"type:text"`
	InvoiceTerms        *string `gorm:"type:text"`

	// Sales Order Settings
	SOPrefix       string `gorm:"type:varchar(20);default:'SO'"`
	SONumberFormat string `gorm:"type:varchar(100);default:'{PREFIX}{NUMBER}'"`

	// Purchase Order Settings
	POPrefix       string `gorm:"type:varchar(20);default:'PO'"`
	PONumberFormat string `gorm:"type:varchar(100);default:'{PREFIX}{NUMBER}'"`

	// System Settings
	Currency string `gorm:"type:varchar(10);default:'IDR'"`
	Timezone string `gorm:"type:varchar(50);default:'Asia/Jakarta'"`
	Locale   string `gorm:"type:varchar(10);default:'id-ID'"`

	// Business Hours
	BusinessHoursStart *string `gorm:"type:varchar(10);default:'08:00'"` // HH:mm format
	BusinessHoursEnd   *string `gorm:"type:varchar(10);default:'17:00'"`
	WorkingDays        *string `gorm:"type:varchar(50);default:'1,2,3,4,5'"` // 0=Sunday, 1=Monday

	IsActive  bool      `gorm:"default:true;index:idx_company_active"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	// Relations
	Tenant           Tenant            `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"` // N Companies → 1 Tenant
	Banks            []CompanyBank     `gorm:"foreignKey:CompanyID"`
	UserCompanyRoles []UserCompanyRole `gorm:"foreignKey:CompanyID"` // User access mapping per company
}

// TableName specifies the table name for Company model
func (Company) TableName() string {
	return "companies"
}

// BeforeCreate hook to generate UUID for ID field
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// CompanyBank - Company bank accounts
type CompanyBank struct {
	ID            string    `gorm:"type:varchar(255);primaryKey"`
	CompanyID     string    `gorm:"type:varchar(255);not null;index"`
	BankName      string    `gorm:"type:varchar(100);not null"` // "BCA", "Mandiri", "BRI"
	AccountNumber string    `gorm:"type:varchar(100);not null"`
	AccountName   string    `gorm:"type:varchar(255);not null"`
	BranchName    *string   `gorm:"type:varchar(255)"`
	IsPrimary     bool      `gorm:"default:false;index"` // Primary bank for invoices
	CheckPrefix   *string   `gorm:"type:varchar(20)"`    // Prefix for check numbers
	IsActive      bool      `gorm:"default:true"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`

	// Relations
	Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for CompanyBank model
func (CompanyBank) TableName() string {
	return "company_banks"
}

// BeforeCreate hook to generate UUID for ID field
func (cb *CompanyBank) BeforeCreate(tx *gorm.DB) error {
	if cb.ID == "" {
		cb.ID = uuid.New().String()
	}
	return nil
}
