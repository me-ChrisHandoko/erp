package dto

import (
	"time"
)

// ============================================================================
// SUPPLIER MANAGEMENT DTOs
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 3 - Supplier Management
// ============================================================================

// CreateSupplierRequest - Request to create a new supplier
type CreateSupplierRequest struct {
	Code           string  `json:"code" binding:"required,min=1,max=100"`
	Name           string  `json:"name" binding:"required,min=2,max=255"`
	Type           *string `json:"type" binding:"omitempty,oneof=MANUFACTURER DISTRIBUTOR WHOLESALER"`
	Phone          *string `json:"phone" binding:"omitempty,max=50"`
	Email          *string `json:"email" binding:"omitempty,email,max=255"`
	Address        *string `json:"address" binding:"omitempty"`
	City           *string `json:"city" binding:"omitempty,max=100"`
	Province       *string `json:"province" binding:"omitempty,max=100"`
	PostalCode     *string `json:"postalCode" binding:"omitempty,max=50"`
	NPWP           *string `json:"npwp" binding:"omitempty,max=50"`
	IsPKP          *bool   `json:"isPKP" binding:"omitempty"`
	ContactPerson  *string `json:"contactPerson" binding:"omitempty,max=255"`
	ContactPhone   *string `json:"contactPhone" binding:"omitempty,max=50"`
	PaymentTerm    *int    `json:"paymentTerm" binding:"omitempty,min=0"`        // Days (0 = cash)
	CreditLimit    *string `json:"creditLimit" binding:"omitempty"`              // decimal as string
	Notes          *string `json:"notes" binding:"omitempty"`
}

// UpdateSupplierRequest - Request to update an existing supplier
type UpdateSupplierRequest struct {
	Code           *string `json:"code" binding:"omitempty,min=1,max=100"`
	Name           *string `json:"name" binding:"omitempty,min=2,max=255"`
	Type           *string `json:"type" binding:"omitempty,oneof=MANUFACTURER DISTRIBUTOR WHOLESALER"`
	Phone          *string `json:"phone" binding:"omitempty,max=50"`
	Email          *string `json:"email" binding:"omitempty,email,max=255"`
	Address        *string `json:"address" binding:"omitempty"`
	City           *string `json:"city" binding:"omitempty,max=100"`
	Province       *string `json:"province" binding:"omitempty,max=100"`
	PostalCode     *string `json:"postalCode" binding:"omitempty,max=50"`
	NPWP           *string `json:"npwp" binding:"omitempty,max=50"`
	IsPKP          *bool   `json:"isPKP" binding:"omitempty"`
	ContactPerson  *string `json:"contactPerson" binding:"omitempty,max=255"`
	ContactPhone   *string `json:"contactPhone" binding:"omitempty,max=50"`
	PaymentTerm    *int    `json:"paymentTerm" binding:"omitempty,min=0"`
	CreditLimit    *string `json:"creditLimit" binding:"omitempty"`
	Notes          *string `json:"notes" binding:"omitempty"`
	IsActive       *bool   `json:"isActive" binding:"omitempty"`
}

// SupplierResponse - Response DTO for supplier data
type SupplierResponse struct {
	ID                 string                   `json:"id"`
	Code               string                   `json:"code"`
	Name               string                   `json:"name"`
	Type               *string                  `json:"type,omitempty"`
	Phone              *string                  `json:"phone,omitempty"`
	Email              *string                  `json:"email,omitempty"`
	Address            *string                  `json:"address,omitempty"`
	City               *string                  `json:"city,omitempty"`
	Province           *string                  `json:"province,omitempty"`
	PostalCode         *string                  `json:"postalCode,omitempty"`
	NPWP               *string                  `json:"npwp,omitempty"`
	IsPKP              bool                     `json:"isPKP"`
	ContactPerson      *string                  `json:"contactPerson,omitempty"`
	ContactPhone       *string                  `json:"contactPhone,omitempty"`
	PaymentTerm        int                      `json:"paymentTerm"`
	CreditLimit        string                   `json:"creditLimit"`
	CurrentOutstanding string                   `json:"currentOutstanding"`
	OverdueAmount      string                   `json:"overdueAmount"`
	LastTransactionAt  *time.Time               `json:"lastTransactionAt,omitempty"`
	Notes              *string                  `json:"notes,omitempty"`
	IsActive           bool                     `json:"isActive"`
	CreatedAt          time.Time                `json:"createdAt"`
	UpdatedAt          time.Time                `json:"updatedAt"`
}

// SupplierListResponse - Response DTO for supplier list with pagination
type SupplierListResponse struct {
	Suppliers  []SupplierResponse `json:"suppliers"`
	TotalCount int64              `json:"totalCount"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

// SupplierListQuery - Query parameters for listing suppliers
type SupplierListQuery struct {
	Page        int     `form:"page" binding:"omitempty,min=1"`
	PageSize    int     `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Search      string  `form:"search" binding:"omitempty"`                                              // Search by code or name
	Type        *string `form:"type" binding:"omitempty,oneof=MANUFACTURER DISTRIBUTOR WHOLESALER"`
	City        *string `form:"city" binding:"omitempty"`
	Province    *string `form:"province" binding:"omitempty"`
	IsPKP       *bool   `form:"isPKP" binding:"omitempty"`
	IsActive    *bool   `form:"isActive" binding:"omitempty"`
	HasOverdue  *bool   `form:"hasOverdue" binding:"omitempty"`  // Filter suppliers with overdue amounts > 0
	SortBy      string  `form:"sortBy" binding:"omitempty,oneof=code name createdAt currentOutstanding overdueAmount"`
	SortOrder   string  `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// SUPPLIER STATISTICS DTOs (for future use)
// ============================================================================

// SupplierStatisticsResponse - Statistics about supplier outstanding and overdue
type SupplierStatisticsResponse struct {
	TotalSuppliers          int    `json:"totalSuppliers"`
	ActiveSuppliers         int    `json:"activeSuppliers"`
	TotalOutstanding        string `json:"totalOutstanding"`
	TotalOverdue            string `json:"totalOverdue"`
	SuppliersWithOverdue    int    `json:"suppliersWithOverdue"`
	SuppliersExceedingLimit int    `json:"suppliersExceedingLimit"`
}
