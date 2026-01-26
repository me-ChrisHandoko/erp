package dto

import (
	"time"
)

// ============================================================================
// DELIVERY TOLERANCE SETTINGS DTOs
// Reference: SAP Model - Hierarchical tolerance configuration
// ============================================================================

// CreateDeliveryToleranceRequest - Request to create a delivery tolerance setting
type CreateDeliveryToleranceRequest struct {
	Level                  string  `json:"level" binding:"required,oneof=COMPANY CATEGORY PRODUCT"`
	CategoryName           *string `json:"categoryName" binding:"omitempty,max=100"` // Required for CATEGORY level (matches Product.Category)
	ProductID              *string `json:"productId" binding:"omitempty,uuid"`       // Required for PRODUCT level
	UnderDeliveryTolerance string  `json:"underDeliveryTolerance" binding:"required"` // Percentage as string (e.g., "5.00")
	OverDeliveryTolerance  string  `json:"overDeliveryTolerance" binding:"required"`  // Percentage as string
	UnlimitedOverDelivery  bool    `json:"unlimitedOverDelivery"`
	IsActive               *bool   `json:"isActive"`
	Notes                  *string `json:"notes" binding:"omitempty,max=500"`
}

// UpdateDeliveryToleranceRequest - Request to update a delivery tolerance setting
type UpdateDeliveryToleranceRequest struct {
	UnderDeliveryTolerance *string `json:"underDeliveryTolerance" binding:"omitempty"` // Percentage as string
	OverDeliveryTolerance  *string `json:"overDeliveryTolerance" binding:"omitempty"`  // Percentage as string
	UnlimitedOverDelivery  *bool   `json:"unlimitedOverDelivery"`
	IsActive               *bool   `json:"isActive"`
	Notes                  *string `json:"notes" binding:"omitempty,max=500"`
}

// DeliveryToleranceProductResponse - Product info for tolerance response
type DeliveryToleranceProductResponse struct {
	ID       string  `json:"id"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Category *string `json:"category,omitempty"`
	BaseUnit string  `json:"baseUnit"`
}

// DeliveryToleranceResponse - Response DTO for delivery tolerance
type DeliveryToleranceResponse struct {
	ID                     string                            `json:"id"`
	Level                  string                            `json:"level"`
	CategoryName           *string                           `json:"categoryName,omitempty"` // For CATEGORY level (matches Product.Category)
	ProductID              *string                           `json:"productId,omitempty"`
	Product                *DeliveryToleranceProductResponse `json:"product,omitempty"`
	UnderDeliveryTolerance string                            `json:"underDeliveryTolerance"` // Percentage as string
	OverDeliveryTolerance  string                            `json:"overDeliveryTolerance"`  // Percentage as string
	UnlimitedOverDelivery  bool                              `json:"unlimitedOverDelivery"`
	IsActive               bool                              `json:"isActive"`
	Notes                  *string                           `json:"notes,omitempty"`
	CreatedAt              time.Time                         `json:"createdAt"`
	UpdatedAt              time.Time                         `json:"updatedAt"`
	CreatedBy              *string                           `json:"createdBy,omitempty"`
	UpdatedBy              *string                           `json:"updatedBy,omitempty"`
}

// DeliveryToleranceListResponse - Response DTO for delivery tolerance list
type DeliveryToleranceListResponse struct {
	Success    bool                        `json:"success"`
	Data       []DeliveryToleranceResponse `json:"data"`
	Pagination PaginationInfo              `json:"pagination"`
}

// DeliveryToleranceListQuery - Query parameters for listing delivery tolerances
type DeliveryToleranceListQuery struct {
	Page         int     `form:"page" binding:"omitempty,min=1"`
	PageSize     int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Level        *string `form:"level" binding:"omitempty,oneof=COMPANY CATEGORY PRODUCT"`
	CategoryName *string `form:"category_name" binding:"omitempty,max=100"` // Filter by category name
	ProductID    *string `form:"product_id" binding:"omitempty,uuid"`
	IsActive     *bool   `form:"is_active"`
	SortBy       string  `form:"sort_by" binding:"omitempty,oneof=level createdAt updatedAt"`
	SortOrder    string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// GetEffectiveToleranceRequest - Request to get effective tolerance for a product
type GetEffectiveToleranceRequest struct {
	ProductID string `form:"product_id" binding:"required,uuid"`
}

// EffectiveToleranceResponse - Response with resolved tolerance for a product
// This returns the actual tolerance that will be applied based on hierarchy
type EffectiveToleranceResponse struct {
	ProductID              string  `json:"productId"`
	ProductCode            string  `json:"productCode"`
	ProductName            string  `json:"productName"`
	UnderDeliveryTolerance string  `json:"underDeliveryTolerance"`
	OverDeliveryTolerance  string  `json:"overDeliveryTolerance"`
	UnlimitedOverDelivery  bool    `json:"unlimitedOverDelivery"`
	ResolvedFrom           string  `json:"resolvedFrom"` // PRODUCT, CATEGORY, COMPANY, or DEFAULT
	ToleranceID            *string `json:"toleranceId,omitempty"` // ID of the tolerance setting used
}
