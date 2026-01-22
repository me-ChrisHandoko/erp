package dto

import (
	"time"
)

// ============================================================================
// INVENTORY ADJUSTMENT DTOs
// Stock adjustment management (penambahan/pengurangan stok)
// ============================================================================

// CreateInventoryAdjustmentRequest - Request to create a new inventory adjustment
type CreateInventoryAdjustmentRequest struct {
	AdjustmentDate string                               `json:"adjustmentDate" binding:"required"`
	WarehouseID    string                               `json:"warehouseId" binding:"required,uuid"`
	AdjustmentType string                               `json:"adjustmentType" binding:"required,oneof=INCREASE DECREASE"`
	Reason         string                               `json:"reason" binding:"required,oneof=SHRINKAGE DAMAGE EXPIRED THEFT OPNAME CORRECTION RETURN OTHER"`
	Notes          *string                              `json:"notes" binding:"omitempty"`
	Items          []CreateInventoryAdjustmentItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateInventoryAdjustmentItemRequest - Inventory adjustment item
type CreateInventoryAdjustmentItemRequest struct {
	ProductID        string  `json:"productId" binding:"required,uuid"`
	QuantityAdjusted string  `json:"quantityAdjusted" binding:"required"`
	UnitCost         string  `json:"unitCost" binding:"required"`
	BatchID          *string `json:"batchId" binding:"omitempty"`
	Notes            *string `json:"notes" binding:"omitempty"`
}

// UpdateInventoryAdjustmentRequest - Request to update existing inventory adjustment (DRAFT only)
type UpdateInventoryAdjustmentRequest struct {
	AdjustmentDate *string                               `json:"adjustmentDate" binding:"omitempty"`
	WarehouseID    *string                               `json:"warehouseId" binding:"omitempty,uuid"`
	AdjustmentType *string                               `json:"adjustmentType" binding:"omitempty,oneof=INCREASE DECREASE"`
	Reason         *string                               `json:"reason" binding:"omitempty,oneof=SHRINKAGE DAMAGE EXPIRED THEFT OPNAME CORRECTION RETURN OTHER"`
	Notes          *string                               `json:"notes" binding:"omitempty"`
	Items          *[]CreateInventoryAdjustmentItemRequest `json:"items" binding:"omitempty,dive"`
}

// ApproveAdjustmentRequest - Request to approve an adjustment (optional notes)
type ApproveAdjustmentRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// CancelAdjustmentRequest - Request to cancel an adjustment
type CancelAdjustmentRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
}

// InventoryAdjustmentItemResponse - Response DTO for inventory adjustment item
type InventoryAdjustmentItemResponse struct {
	ID               string                `json:"id"`
	ProductID        string                `json:"productId"`
	Product          *ProductBasicResponse `json:"product,omitempty"`
	BatchID          *string               `json:"batchId,omitempty"`
	QuantityBefore   string                `json:"quantityBefore"`
	QuantityAdjusted string                `json:"quantityAdjusted"`
	QuantityAfter    string                `json:"quantityAfter"`
	UnitCost         string                `json:"unitCost"`
	TotalValue       string                `json:"totalValue"`
	Notes            *string               `json:"notes,omitempty"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
}

// UserBasicResponse - Basic user info for audit trail
type UserBasicResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

// InventoryAdjustmentResponse - Response DTO for inventory adjustment
type InventoryAdjustmentResponse struct {
	ID               string                            `json:"id"`
	AdjustmentNumber string                            `json:"adjustmentNumber"`
	AdjustmentDate   string                            `json:"adjustmentDate"`
	WarehouseID      string                            `json:"warehouseId"`
	Warehouse        *WarehouseBasicResponse           `json:"warehouse,omitempty"`
	AdjustmentType   string                            `json:"adjustmentType"`
	Reason           string                            `json:"reason"`
	Status           string                            `json:"status"`
	Notes            *string                           `json:"notes,omitempty"`
	TotalItems       int                               `json:"totalItems"`
	TotalValue       string                            `json:"totalValue"`
	CreatedBy        string                            `json:"createdBy"`
	CreatedByUser    *UserBasicResponse                `json:"createdByUser,omitempty"`
	ApprovedBy       *string                           `json:"approvedBy,omitempty"`
	ApprovedByUser   *UserBasicResponse                `json:"approvedByUser,omitempty"`
	ApprovedAt       *time.Time                        `json:"approvedAt,omitempty"`
	CancelledBy      *string                           `json:"cancelledBy,omitempty"`
	CancelledAt      *time.Time                        `json:"cancelledAt,omitempty"`
	CancelReason     *string                           `json:"cancelReason,omitempty"`
	Items            []InventoryAdjustmentItemResponse `json:"items,omitempty"`
	CreatedAt        time.Time                         `json:"createdAt"`
	UpdatedAt        time.Time                         `json:"updatedAt"`
}

// InventoryAdjustmentStatusCounts - Status counts for statistics cards
type InventoryAdjustmentStatusCounts struct {
	Draft     int `json:"draft"`
	Approved  int `json:"approved"`
	Cancelled int `json:"cancelled"`
}

// InventoryAdjustmentListResponse - Response DTO for inventory adjustment list with pagination
type InventoryAdjustmentListResponse struct {
	Success      bool                             `json:"success"`
	Data         []InventoryAdjustmentResponse    `json:"data"`
	Pagination   PaginationInfo                   `json:"pagination"`
	StatusCounts *InventoryAdjustmentStatusCounts `json:"statusCounts,omitempty"`
}

// InventoryAdjustmentQuery - Query parameters for listing inventory adjustments
type InventoryAdjustmentQuery struct {
	Page           int     `form:"page" binding:"omitempty,min=1"`
	PageSize       int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search         string  `form:"search" binding:"omitempty"`
	Status         *string `form:"status" binding:"omitempty,oneof=DRAFT APPROVED CANCELLED"`
	AdjustmentType *string `form:"adjustment_type" binding:"omitempty,oneof=INCREASE DECREASE"`
	Reason         *string `form:"reason" binding:"omitempty,oneof=SHRINKAGE DAMAGE EXPIRED THEFT OPNAME CORRECTION RETURN OTHER"`
	WarehouseID    *string `form:"warehouse_id" binding:"omitempty,uuid"`
	DateFrom       *string `form:"date_from" binding:"omitempty"`
	DateTo         *string `form:"date_to" binding:"omitempty"`
	SortBy         string  `form:"sort_by" binding:"omitempty,oneof=adjustmentNumber adjustmentDate status createdAt"`
	SortOrder      string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
