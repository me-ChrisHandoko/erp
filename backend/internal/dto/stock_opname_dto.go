package dto

import "time"

// ============================================================================
// STOCK OPNAME REQUEST DTOs
// ============================================================================

// CreateStockOpnameRequest represents stock opname creation request
type CreateStockOpnameRequest struct {
	OpnameDate  string                         `json:"opnameDate" binding:"required"`
	WarehouseID string                         `json:"warehouseId" binding:"required"`
	Notes       *string                        `json:"notes" binding:"omitempty"`
	Items       []CreateStockOpnameItemRequest `json:"items" binding:"required,min=1,dive"`
}

// UpdateStockOpnameRequest represents stock opname update request
type UpdateStockOpnameRequest struct {
	OpnameDate *string `json:"opnameDate" binding:"omitempty"`
	Status     *string `json:"status" binding:"omitempty,oneof=draft in_progress completed"`
	Notes      *string `json:"notes" binding:"omitempty"`
}

// ApproveStockOpnameRequest represents stock opname approval request
type ApproveStockOpnameRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// CreateStockOpnameItemRequest represents stock opname item creation
type CreateStockOpnameItemRequest struct {
	ProductID   string  `json:"productId" binding:"required"`
	ExpectedQty string  `json:"expectedQty" binding:"required"` // decimal as string
	ActualQty   string  `json:"actualQty" binding:"required"`   // decimal as string
	Notes       *string `json:"notes" binding:"omitempty"`
}

// UpdateStockOpnameItemRequest represents stock opname item update
type UpdateStockOpnameItemRequest struct {
	ActualQty *string `json:"actualQty" binding:"omitempty"` // decimal as string
	Notes     *string `json:"notes" binding:"omitempty"`
}

// BatchUpdateStockOpnameItemRequest represents a single item in batch update
type BatchUpdateStockOpnameItemRequest struct {
	ItemID    string  `json:"itemId" binding:"required"`
	ActualQty *string `json:"actualQty" binding:"omitempty"` // decimal as string
	Notes     *string `json:"notes" binding:"omitempty"`
}

// BatchUpdateStockOpnameItemsRequest represents batch update for multiple items
type BatchUpdateStockOpnameItemsRequest struct {
	Items []BatchUpdateStockOpnameItemRequest `json:"items" binding:"required,min=1,dive"`
}

// StockOpnameFilters represents stock opname list filters
type StockOpnameFilters struct {
	Search      string `form:"search"`
	WarehouseID string `form:"warehouse_id"`
	Status      string `form:"status" binding:"omitempty,oneof=draft in_progress completed approved"`
	DateFrom    string `form:"date_from"`
	DateTo      string `form:"date_to"`
	Page        int    `form:"page" binding:"omitempty,min=1"`
	Limit       int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy      string `form:"sort_by" binding:"omitempty,oneof=opnameNumber opnameDate warehouseName status"`
	SortOrder   string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// STOCK OPNAME RESPONSE DTOs
// ============================================================================

// StockOpnameResponse represents stock opname information response
type StockOpnameResponse struct {
	ID               string                    `json:"id"`
	CompanyID        string                    `json:"companyId"`
	TenantID         string                    `json:"tenantId"`
	OpnameNumber     string                    `json:"opnameNumber"`
	OpnameDate       string                    `json:"opnameDate"` // ISO 8601 format
	WarehouseID      string                    `json:"warehouseId"`
	WarehouseName    *string                   `json:"warehouseName,omitempty"`
	Status           string                    `json:"status"`
	TotalItems       int                       `json:"totalItems"`
	TotalExpectedQty string                    `json:"totalExpectedQty"` // decimal as string
	TotalActualQty   string                    `json:"totalActualQty"`   // decimal as string
	TotalDifference  string                    `json:"totalDifference"`  // decimal as string
	Notes            *string                   `json:"notes,omitempty"`
	CreatedBy        string                    `json:"createdBy"`
	CreatedByName    *string                   `json:"createdByName,omitempty"`
	CreatedAt        time.Time                 `json:"createdAt"`
	ApprovedBy       *string                   `json:"approvedBy,omitempty"`
	ApprovedByName   *string                   `json:"approvedByName,omitempty"`
	ApprovedAt       *time.Time                `json:"approvedAt,omitempty"`
	Items            []StockOpnameItemResponse `json:"items,omitempty"`
	UpdatedAt        time.Time                 `json:"updatedAt"`
}

// StockOpnameItemResponse represents stock opname item information
type StockOpnameItemResponse struct {
	ID          string  `json:"id"`
	OpnameID    string  `json:"opnameId"`
	ProductID   string  `json:"productId"`
	ProductCode *string `json:"productCode,omitempty"`
	ProductName *string `json:"productName,omitempty"`
	ExpectedQty string  `json:"expectedQty"` // decimal as string
	ActualQty   string  `json:"actualQty"`   // decimal as string
	Difference  string  `json:"difference"`  // decimal as string
	Notes       *string `json:"notes,omitempty"`
}

// ============================================================================
// PAGINATION & LIST RESPONSE
// ============================================================================

// StockOpnameListResponse represents paginated stock opname list
type StockOpnameListResponse struct {
	Success    bool                  `json:"success"`
	Data       []StockOpnameResponse `json:"data"`
	Pagination PaginationInfo        `json:"pagination"`
}

// StockOpnameDetailResponse represents single stock opname response
type StockOpnameDetailResponse struct {
	Success bool                `json:"success"`
	Data    StockOpnameResponse `json:"data"`
}

// StockOpnameItemDetailResponse represents single stock opname item response
type StockOpnameItemDetailResponse struct {
	Success bool                    `json:"success"`
	Data    StockOpnameItemResponse `json:"data"`
}

// ImportProductsResponse represents bulk import response
type ImportProductsResponse struct {
	Success    bool   `json:"success"`
	ItemsAdded int    `json:"itemsAdded"`
	Message    string `json:"message"`
}
