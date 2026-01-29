package dto

// CompanyResponse represents company profile information
// Reference: CLAUDE.md - Company Profile Management
type CompanyResponse struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	LegalName      string             `json:"legalName,omitempty"`
	EntityType     string             `json:"entityType,omitempty"`
	NPWP           string             `json:"npwp,omitempty"`
	NIB            string             `json:"nib,omitempty"`
	Address        string             `json:"address,omitempty"`
	City           string             `json:"city,omitempty"`
	Province       string             `json:"province,omitempty"`
	PostalCode     string             `json:"postalCode,omitempty"`
	Phone          string             `json:"phone,omitempty"`
	Email          string             `json:"email,omitempty"`
	Website        string             `json:"website,omitempty"`
	LogoURL        string             `json:"logoUrl,omitempty"`
	IsPKP          bool               `json:"isPkp"`
	PPNRate        float64            `json:"ppnRate"`
	InvoicePrefix  string             `json:"invoicePrefix,omitempty"`
	// Purchase Invoice Settings (3-way matching)
	InvoiceControlPolicy string  `json:"invoiceControlPolicy,omitempty"` // ORDERED or RECEIVED
	InvoiceTolerancePct  float64 `json:"invoiceTolerancePct"`            // Tolerance % for over-invoicing
	IsActive             bool    `json:"isActive"`
	Banks                []CompanyBankInfo  `json:"banks,omitempty"`
}

// CompanyBankInfo represents bank account information in company response
type CompanyBankInfo struct {
	ID            string `json:"id"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	BranchName    string `json:"branchName,omitempty"`
	IsPrimary     bool   `json:"isPrimary"`
	CheckPrefix   string `json:"checkPrefix,omitempty"`
	IsActive      bool   `json:"isActive"`
}

// UpdateCompanyRequest represents company profile update request
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Day 1-4 Tasks
type UpdateCompanyRequest struct {
	Name                 *string  `json:"name" binding:"omitempty,min=2,max=255" validate:"omitempty,min=2,max=255"`
	LegalName            *string  `json:"legalName" binding:"omitempty,min=2,max=255" validate:"omitempty,min=2,max=255"`
	EntityType           *string  `json:"entityType" binding:"omitempty,oneof=PT CV UD Firma" validate:"omitempty,oneof=PT CV UD Firma"`
	NPWP                 *string  `json:"npwp" binding:"omitempty,min=15,max=20" validate:"omitempty,min=15,max=20"`
	NIB                  *string  `json:"nib" binding:"omitempty,max=50" validate:"omitempty,max=50"`
	Address              *string  `json:"address" binding:"omitempty,max=500" validate:"omitempty,max=500"`
	City                 *string  `json:"city" binding:"omitempty,max=100" validate:"omitempty,max=100"`
	Province             *string  `json:"province" binding:"omitempty,max=100" validate:"omitempty,max=100"`
	PostalCode           *string  `json:"postalCode" binding:"omitempty,max=10" validate:"omitempty,max=10"`
	Phone                *string  `json:"phone" binding:"omitempty,max=20" validate:"omitempty,max=20"`
	Email                *string  `json:"email" binding:"omitempty,email,max=255" validate:"omitempty,email,max=255"`
	Website              *string  `json:"website" binding:"omitempty,url,max=255" validate:"omitempty,url,max=255"`
	LogoURL              *string  `json:"logoUrl" binding:"omitempty,url,max=500" validate:"omitempty,url,max=500"`
	IsPKP                *bool    `json:"isPkp" binding:"omitempty" validate:"omitempty"`
	PPNRate              *float64 `json:"ppnRate" binding:"omitempty,min=0,max=100" validate:"omitempty,min=0,max=100"`
	InvoicePrefix        *string  `json:"invoicePrefix" binding:"omitempty,max=10" validate:"omitempty,max=10"`
	InvoiceNumberFormat  *string  `json:"invoiceNumberFormat" binding:"omitempty,max=50" validate:"omitempty,max=50"`
	FakturPajakSeries    *string  `json:"fakturPajakSeries" binding:"omitempty,max=20" validate:"omitempty,max=20"`
	SPPKPNumber          *string  `json:"sppkpNumber" binding:"omitempty,max=50" validate:"omitempty,max=50"`
	// Purchase Invoice Settings (3-way matching like SAP/Odoo)
	InvoiceControlPolicy *string  `json:"invoiceControlPolicy" binding:"omitempty,oneof=ORDERED RECEIVED" validate:"omitempty,oneof=ORDERED RECEIVED"`
	InvoiceTolerancePct  *float64 `json:"invoiceTolerancePct" binding:"omitempty,min=0,max=100" validate:"omitempty,min=0,max=100"`
}

// AddBankAccountRequest represents bank account addition request
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #10
type AddBankAccountRequest struct {
	BankName      string  `json:"bankName" binding:"required,min=2,max=100" validate:"required,min=2,max=100"`
	AccountNumber string  `json:"accountNumber" binding:"required,min=8,max=50" validate:"required,min=8,max=50"`
	AccountName   string  `json:"accountName" binding:"required,min=3,max=255" validate:"required,min=3,max=255"`
	BranchName    *string `json:"branchName" binding:"omitempty,max=255" validate:"omitempty,max=255"`
	IsPrimary     bool    `json:"isPrimary"`
	CheckPrefix   *string `json:"checkPrefix" binding:"omitempty,max=20" validate:"omitempty,max=20"`
}

// UpdateBankAccountRequest represents bank account update request
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #10
type UpdateBankAccountRequest struct {
	BankName      *string `json:"bankName" binding:"omitempty,min=2,max=100" validate:"omitempty,min=2,max=100"`
	AccountNumber *string `json:"accountNumber" binding:"omitempty,min=8,max=50" validate:"omitempty,min=8,max=50"`
	AccountName   *string `json:"accountName" binding:"omitempty,min=3,max=255" validate:"omitempty,min=3,max=255"`
	BranchName    *string `json:"branchName" binding:"omitempty,max=255" validate:"omitempty,max=255"`
	IsPrimary     *bool   `json:"isPrimary" binding:"omitempty" validate:"omitempty"`
	CheckPrefix   *string `json:"checkPrefix" binding:"omitempty,max=20" validate:"omitempty,max=20"`
}

// BankAccountResponse represents bank account information response
type BankAccountResponse struct {
	ID            string `json:"id"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	BranchName    string `json:"branchName,omitempty"`
	IsPrimary     bool   `json:"isPrimary"`
	CheckPrefix   string `json:"checkPrefix,omitempty"`
	IsActive      bool   `json:"isActive"`
}

// BankAccountFilters represents bank account list filters
// Follows the same pattern as ProductFilters for consistency
type BankAccountFilters struct {
	Search    string `form:"search"`                                                 // Search by bank name or account number
	IsPrimary *bool  `form:"is_primary"`                                             // Filter by primary status
	IsActive  *bool  `form:"is_active"`                                              // Filter by active status
	Page      int    `form:"page" binding:"omitempty,min=1"`                         // Page number (default: 1)
	Limit     int    `form:"page_size" binding:"omitempty,min=1,max=100"`            // Page size (default: 20)
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=bankName createdAt"`   // Sort field
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`          // Sort order (asc/desc)
}
