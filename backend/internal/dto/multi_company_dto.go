package dto

// PHASE 3: Multi-Company DTOs
// Reference: multi-company-architecture-analysis.md - PHASE 3

// ============================================================================
// Request DTOs
// ============================================================================

// CreateCompanyRequest represents request to create a new company
type CreateCompanyRequest struct {
	Name                 string   `json:"name" binding:"required,min=2,max=255" validate:"required,min=2,max=255"`
	LegalName            *string  `json:"legalName" binding:"omitempty,min=2,max=255" validate:"omitempty,min=2,max=255"`
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
	PPNRate              *float64 `json:"ppnRate" binding:"omitempty,gte=0,lte=100" validate:"omitempty,gte=0,lte=100"`
	InvoicePrefix        *string  `json:"invoicePrefix" binding:"omitempty,max=10" validate:"omitempty,max=10"`
	InvoiceNumberFormat  *string  `json:"invoiceNumberFormat" binding:"omitempty,max=50" validate:"omitempty,max=50"`
	FakturPajakSeries    *string  `json:"fakturPajakSeries" binding:"omitempty,max=20" validate:"omitempty,max=20"`
	SPPKPNumber          *string  `json:"sppkpNumber" binding:"omitempty,max=50" validate:"omitempty,max=50"`
}

// UpdateCompanyRequest represents request to update company (reuses existing DTO)
// Defined in company_dto.go

// SwitchCompanyRequest represents request to switch active company
type SwitchCompanyRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid" validate:"required,uuid"`
}

// ============================================================================
// Response DTOs
// ============================================================================

// CompanyListResponse represents a company in list view
type CompanyListResponse struct {
	ID               string  `json:"id"`
	TenantID         string  `json:"tenantId"`
	Name             string  `json:"name"`
	LegalName        string  `json:"legalName,omitempty"`
	NPWP             string  `json:"npwp,omitempty"`
	City             string  `json:"city,omitempty"`
	Province         string  `json:"province,omitempty"`
	IsPKP            bool    `json:"isPkp"`
	IsActive         bool    `json:"isActive"`
	UserRole         string  `json:"userRole,omitempty"`         // User's role in this company (Tier 2)
	AccessTier       int     `json:"accessTier"`                 // 1=Tenant-level, 2=Company-level
}

// CompanyDetailResponse represents detailed company information
// Reuses CompanyResponse from company_dto.go with additional fields
type CompanyDetailResponse struct {
	CompanyResponse
	TenantID   string `json:"tenantId"`
	UserRole   string `json:"userRole,omitempty"`
	AccessTier int    `json:"accessTier"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// SwitchCompanyResponse represents response after company switch
type SwitchCompanyResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"` // Empty when sent via httpOnly cookie
	CompanyID    string `json:"company_id"`
	CompanyName  string `json:"company_name"`
	Message      string `json:"message"`
}

// CompanyAccessInfo represents user's access information to a company
type CompanyAccessInfo struct {
	CompanyID  string `json:"company_id"`
	TenantID   string `json:"tenant_id"`
	AccessTier int    `json:"access_tier"` // 0=no access, 1=tenant-level, 2=company-level
	Role       string `json:"role"`        // OWNER, TENANT_ADMIN, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF
	HasAccess  bool   `json:"has_access"`
}
