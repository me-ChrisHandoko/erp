package dto

import (
	"time"
)

// ============================================================================
// STOCK TRANSFER DTOs
// Inter-warehouse stock transfer management
// ============================================================================

// CreateStockTransferRequest - Request to create a new stock transfer
type CreateStockTransferRequest struct {
	TransferDate      string                      `json:"transferDate" binding:"required"`
	SourceWarehouseID string                      `json:"sourceWarehouseId" binding:"required,uuid"`
	DestWarehouseID   string                      `json:"destWarehouseId" binding:"required,uuid"`
	Notes             *string                     `json:"notes" binding:"omitempty"`
	Items             []CreateStockTransferItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateStockTransferItemRequest - Stock transfer item
type CreateStockTransferItemRequest struct {
	ProductID string  `json:"productId" binding:"required,uuid"`
	Quantity  string  `json:"quantity" binding:"required"`
	BatchID   *string `json:"batchId" binding:"omitempty"`
	Notes     *string `json:"notes" binding:"omitempty"`
}

// UpdateStockTransferRequest - Request to update existing stock transfer (DRAFT only)
type UpdateStockTransferRequest struct {
	TransferDate      *string                      `json:"transferDate" binding:"omitempty"`
	SourceWarehouseID *string                      `json:"sourceWarehouseId" binding:"omitempty,uuid"`
	DestWarehouseID   *string                      `json:"destWarehouseId" binding:"omitempty,uuid"`
	Notes             *string                      `json:"notes" binding:"omitempty"`
	Items             *[]CreateStockTransferItemRequest `json:"items" binding:"omitempty,dive"`
}

// ShipTransferRequest - Request to ship a transfer (DRAFT → SHIPPED)
type ShipTransferRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// ReceiveTransferRequest - Request to receive a transfer (SHIPPED → RECEIVED)
type ReceiveTransferRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// CancelTransferRequest - Request to cancel a transfer (SHIPPED → CANCELLED)
type CancelTransferRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
}

// StockTransferItemResponse - Response DTO for stock transfer item
type StockTransferItemResponse struct {
	ID       string                  `json:"id"`
	ProductID string                 `json:"productId"`
	Product  *ProductBasicResponse   `json:"product,omitempty"`
	Quantity string                  `json:"quantity"`
	BatchID  *string                 `json:"batchId,omitempty"`
	Notes    *string                 `json:"notes,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// ProductBasicResponse - Basic product info for transfers
type ProductBasicResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// WarehouseBasicResponse - Basic warehouse info for transfers
type WarehouseBasicResponse struct {
	ID   string  `json:"id"`
	Code string  `json:"code"`
	Name string  `json:"name"`
}

// StockTransferResponse - Response DTO for stock transfer
type StockTransferResponse struct {
	ID               string                       `json:"id"`
	TransferNumber   string                       `json:"transferNumber"`
	TransferDate     string                       `json:"transferDate"`
	SourceWarehouseID string                      `json:"sourceWarehouseId"`
	SourceWarehouse  *WarehouseBasicResponse     `json:"sourceWarehouse,omitempty"`
	DestWarehouseID   string                      `json:"destWarehouseId"`
	DestWarehouse    *WarehouseBasicResponse     `json:"destWarehouse,omitempty"`
	Status           string                       `json:"status"`
	ShippedBy        *string                      `json:"shippedBy,omitempty"`
	ShippedAt        *time.Time                   `json:"shippedAt,omitempty"`
	ReceivedBy       *string                      `json:"receivedBy,omitempty"`
	ReceivedAt       *time.Time                   `json:"receivedAt,omitempty"`
	Notes            *string                      `json:"notes,omitempty"`
	Items            []StockTransferItemResponse  `json:"items,omitempty"`
	CreatedAt        time.Time                    `json:"createdAt"`
	UpdatedAt        time.Time                    `json:"updatedAt"`
}

// StockTransferListResponse - Response DTO for stock transfer list with pagination
type StockTransferListResponse struct {
	Success    bool                    `json:"success"`
	Data       []StockTransferResponse `json:"data"`
	Pagination PaginationInfo          `json:"pagination"`
}

// StockTransferQuery - Query parameters for listing stock transfers
type StockTransferQuery struct {
	Page              int     `form:"page" binding:"omitempty,min=1"`
	PageSize          int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search            string  `form:"search" binding:"omitempty"`
	Status            *string `form:"status" binding:"omitempty,oneof=DRAFT SHIPPED RECEIVED CANCELLED"`
	SourceWarehouseID *string `form:"source_warehouse_id" binding:"omitempty,uuid"`
	DestWarehouseID   *string `form:"dest_warehouse_id" binding:"omitempty,uuid"`
	DateFrom          *string `form:"date_from" binding:"omitempty"`
	DateTo            *string `form:"date_to" binding:"omitempty"`
	SortBy            string  `form:"sort_by" binding:"omitempty,oneof=transferNumber transferDate status createdAt"`
	SortOrder         string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
