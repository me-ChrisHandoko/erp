package dto

import (
	"time"
)

// ============================================================================
// PURCHASE ORDER MANAGEMENT DTOs
// Reference: Phase 3 - Purchase Order Management
// ============================================================================

// CreatePurchaseOrderItemRequest - Request to create a purchase order item
type CreatePurchaseOrderItemRequest struct {
	ProductID     string  `json:"productId" binding:"required,uuid"`
	ProductUnitID *string `json:"productUnitId" binding:"omitempty,uuid"`
	Quantity      string  `json:"quantity" binding:"required"` // decimal as string
	UnitPrice     string  `json:"unitPrice" binding:"required"` // decimal as string
	DiscountPct   *string `json:"discountPct" binding:"omitempty"` // decimal as string
	Notes         *string `json:"notes" binding:"omitempty"`
}

// CreatePurchaseOrderRequest - Request to create a new purchase order
type CreatePurchaseOrderRequest struct {
	SupplierID         string                            `json:"supplierId" binding:"required,uuid"`
	WarehouseID        string                            `json:"warehouseId" binding:"required,uuid"`
	PODate             string                            `json:"poDate" binding:"required"` // ISO date string
	ExpectedDeliveryAt *string                           `json:"expectedDeliveryAt" binding:"omitempty"`
	DiscountAmount     *string                           `json:"discountAmount" binding:"omitempty"` // Overall discount
	TaxAmount          *string                           `json:"taxAmount" binding:"omitempty"` // PPN
	Notes              *string                           `json:"notes" binding:"omitempty"`
	Items              []CreatePurchaseOrderItemRequest  `json:"items" binding:"required,min=1,dive"`
}

// UpdatePurchaseOrderItemRequest - Request to update a purchase order item
type UpdatePurchaseOrderItemRequest struct {
	ID            *string `json:"id" binding:"omitempty,uuid"` // Existing item ID (if updating)
	ProductID     string  `json:"productId" binding:"required,uuid"`
	ProductUnitID *string `json:"productUnitId" binding:"omitempty,uuid"`
	Quantity      string  `json:"quantity" binding:"required"`
	UnitPrice     string  `json:"unitPrice" binding:"required"`
	DiscountPct   *string `json:"discountPct" binding:"omitempty"`
	Notes         *string `json:"notes" binding:"omitempty"`
}

// UpdatePurchaseOrderRequest - Request to update an existing purchase order
type UpdatePurchaseOrderRequest struct {
	SupplierID         *string                           `json:"supplierId" binding:"omitempty,uuid"`
	WarehouseID        *string                           `json:"warehouseId" binding:"omitempty,uuid"`
	PODate             *string                           `json:"poDate" binding:"omitempty"`
	ExpectedDeliveryAt *string                           `json:"expectedDeliveryAt" binding:"omitempty"`
	DiscountAmount     *string                           `json:"discountAmount" binding:"omitempty"`
	TaxAmount          *string                           `json:"taxAmount" binding:"omitempty"`
	Notes              *string                           `json:"notes" binding:"omitempty"`
	Items              []UpdatePurchaseOrderItemRequest  `json:"items" binding:"omitempty,dive"`
}

// CancelPurchaseOrderRequest - Request to cancel a purchase order
type CancelPurchaseOrderRequest struct {
	CancellationNote string `json:"cancellationNote" binding:"required,min=1,max=500"`
}

// PurchaseOrderProductResponse - Product info for purchase order items
type PurchaseOrderProductResponse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	BaseUnit string `json:"baseUnit"`
}

// PurchaseOrderProductUnitResponse - Product unit info for purchase order items
type PurchaseOrderProductUnitResponse struct {
	ID             string `json:"id"`
	UnitName       string `json:"unitName"`
	ConversionRate string `json:"conversionRate"`
}

// PurchaseOrderItemResponse - Response DTO for purchase order item
type PurchaseOrderItemResponse struct {
	ID            string                            `json:"id"`
	ProductID     string                            `json:"productId"`
	Product       *PurchaseOrderProductResponse     `json:"product,omitempty"`
	ProductUnitID *string                           `json:"productUnitId,omitempty"`
	ProductUnit   *PurchaseOrderProductUnitResponse `json:"productUnit,omitempty"`
	Quantity      string                            `json:"quantity"`
	UnitPrice     string                            `json:"unitPrice"`
	DiscountPct   string                            `json:"discountPct"`
	DiscountAmt   string                            `json:"discountAmt"`
	Subtotal      string                            `json:"subtotal"`
	ReceivedQty   string                            `json:"receivedQty"`
	Notes         *string                           `json:"notes,omitempty"`
	CreatedAt     time.Time                         `json:"createdAt"`
	UpdatedAt     time.Time                         `json:"updatedAt"`
}

// PurchaseOrderResponse - Response DTO for purchase order
type PurchaseOrderResponse struct {
	ID                 string                      `json:"id"`
	PONumber           string                      `json:"poNumber"`
	PODate             time.Time                   `json:"poDate"`
	SupplierID         string                      `json:"supplierId"`
	Supplier           *SupplierBasicResponse      `json:"supplier,omitempty"`
	WarehouseID        string                      `json:"warehouseId"`
	Warehouse          *WarehouseBasicResponse     `json:"warehouse,omitempty"`
	Status             string                      `json:"status"`
	Subtotal           string                      `json:"subtotal"`
	DiscountAmount     string                      `json:"discountAmount"`
	TaxAmount          string                      `json:"taxAmount"`
	TotalAmount        string                      `json:"totalAmount"`
	Notes              *string                     `json:"notes,omitempty"`
	ExpectedDeliveryAt *time.Time                  `json:"expectedDeliveryAt,omitempty"`
	RequestedBy        *string                     `json:"requestedBy,omitempty"`
	Requester          *UserBasicResponse          `json:"requester,omitempty"`
	ApprovedBy         *string                     `json:"approvedBy,omitempty"`
	ApprovedAt         *time.Time                  `json:"approvedAt,omitempty"`
	CancelledBy        *string                     `json:"cancelledBy,omitempty"`
	CancelledAt        *time.Time                  `json:"cancelledAt,omitempty"`
	CancellationNote   *string                     `json:"cancellationNote,omitempty"`
	Items              []PurchaseOrderItemResponse `json:"items,omitempty"`
	CreatedAt          time.Time                   `json:"createdAt"`
	UpdatedAt          time.Time                   `json:"updatedAt"`
}

// PurchaseOrderListResponse - Response DTO for purchase order list with pagination
type PurchaseOrderListResponse struct {
	Success    bool                    `json:"success"`
	Data       []PurchaseOrderResponse `json:"data"`
	Pagination PaginationInfo          `json:"pagination"`
}

// PurchaseOrderListQuery - Query parameters for listing purchase orders
type PurchaseOrderListQuery struct {
	Page        int     `form:"page" binding:"omitempty,min=1"`
	PageSize    int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search      string  `form:"search" binding:"omitempty"` // Search by PO number
	Status      *string `form:"status" binding:"omitempty,oneof=DRAFT CONFIRMED COMPLETED CANCELLED"`
	SupplierID  *string `form:"supplier_id" binding:"omitempty,uuid"`
	WarehouseID *string `form:"warehouse_id" binding:"omitempty,uuid"`
	DateFrom    *string `form:"date_from" binding:"omitempty"` // ISO date string
	DateTo      *string `form:"date_to" binding:"omitempty"`   // ISO date string
	SortBy      string  `form:"sort_by" binding:"omitempty,oneof=poNumber poDate totalAmount status createdAt"`
	SortOrder   string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// BASIC RESPONSE DTOs (for nested objects)
// Note: WarehouseBasicResponse, ProductBasicResponse are defined in stock_transfer_dto.go
// Note: ProductUnitResponse is defined in product_dto.go
// Note: UserBasicResponse is defined in inventory_adjustment_dto.go
// ============================================================================

// SupplierBasicResponse - Minimal supplier info for nested responses
type SupplierBasicResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
