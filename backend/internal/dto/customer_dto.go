package dto

import (
	"time"
)

// ============================================================================
// CUSTOMER MANAGEMENT DTOs
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 2 - Customer Management
// ============================================================================

// CreateCustomerRequest - Request to create a new customer
type CreateCustomerRequest struct {
	Code           string  `json:"code" binding:"required,min=1,max=100"`
	Name           string  `json:"name" binding:"required,min=2,max=255"`
	Type           *string `json:"customerType" binding:"omitempty,oneof=Retail Grosir Distributor"`
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
	PaymentTerm    *int    `json:"creditTermDays" binding:"omitempty,min=0"`     // Days (0 = cash)
	CreditLimit    *string `json:"creditLimit" binding:"omitempty"`              // decimal as string
	Notes          *string `json:"notes" binding:"omitempty"`
}

// UpdateCustomerRequest - Request to update an existing customer
type UpdateCustomerRequest struct {
	Code           *string `json:"code" binding:"omitempty,min=1,max=100"`
	Name           *string `json:"name" binding:"omitempty,min=2,max=255"`
	Type           *string `json:"customerType" binding:"omitempty,oneof=Retail Grosir Distributor"`
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
	PaymentTerm    *int    `json:"creditTermDays" binding:"omitempty,min=0"`
	CreditLimit    *string `json:"creditLimit" binding:"omitempty"`
	Notes          *string `json:"notes" binding:"omitempty"`
	IsActive       *bool   `json:"isActive" binding:"omitempty"`
}

// CustomerResponse - Response DTO for customer data
type CustomerResponse struct {
	ID                 string                   `json:"id"`
	Code               string                   `json:"code"`
	Name               string                   `json:"name"`
	Type               *string                  `json:"customerType,omitempty"`
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
	PaymentTerm        int                      `json:"creditTermDays"`
	CreditLimit        string                   `json:"creditLimit"`
	CurrentOutstanding string                   `json:"currentOutstanding"`
	OverdueAmount      string                   `json:"overdueAmount"`
	LastTransactionAt  *time.Time               `json:"lastTransactionAt,omitempty"`
	Notes              *string                  `json:"notes,omitempty"`
	IsActive           bool                     `json:"isActive"`
	CreatedAt          time.Time                `json:"createdAt"`
	UpdatedAt          time.Time                `json:"updatedAt"`
}

// CustomerListResponse - Response DTO for customer list with pagination
type CustomerListResponse struct {
	Customers  []CustomerResponse `json:"customers"`
	TotalCount int64              `json:"totalCount"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	TotalPages int                `json:"totalPages"`
}

// CustomerListQuery - Query parameters for listing customers
type CustomerListQuery struct {
	Page        int     `form:"page" binding:"omitempty,min=1"`
	PageSize    int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search      string  `form:"search" binding:"omitempty"`                                      // Search by code or name
	Type        *string `form:"type" binding:"omitempty,oneof=RETAIL WHOLESALE DISTRIBUTOR"`
	City        *string `form:"city" binding:"omitempty"`
	Province    *string `form:"province" binding:"omitempty"`
	IsPKP       *bool   `form:"is_pkp" binding:"omitempty"`
	IsActive    *bool   `form:"is_active" binding:"omitempty"`
	HasOverdue  *bool   `form:"has_overdue" binding:"omitempty"`  // Filter customers with overdue amounts > 0
	SortBy      string  `form:"sort_by" binding:"omitempty,oneof=code name createdAt currentOutstanding overdueAmount"`
	SortOrder   string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// CUSTOMER STATISTICS DTOs (for future use)
// ============================================================================

// CustomerStatisticsResponse - Statistics about customer outstanding and overdue
type CustomerStatisticsResponse struct {
	TotalCustomers          int    `json:"totalCustomers"`
	ActiveCustomers         int    `json:"activeCustomers"`
	TotalOutstanding        string `json:"totalOutstanding"`
	TotalOverdue            string `json:"totalOverdue"`
	CustomersWithOverdue    int    `json:"customersWithOverdue"`
	CustomersExceedingLimit int    `json:"customersExceedingLimit"`
}
