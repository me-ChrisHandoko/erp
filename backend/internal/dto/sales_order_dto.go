package dto

import "time"

// ============================================================================
// SALES ORDER REQUEST DTOs
// ============================================================================

// CreateSalesOrderRequest represents sales order creation request
type CreateSalesOrderRequest struct {
	CustomerId    string                     `json:"customerId" binding:"required"`
	WarehouseId   string                     `json:"warehouseId" binding:"required"`
	OrderDate     string                     `json:"orderDate" binding:"required"` // ISO 8601 date
	RequiredDate  *string                    `json:"requiredDate" binding:"omitempty"` // ISO 8601 date, optional
	Notes         *string                    `json:"notes" binding:"omitempty"`
	Subtotal      string                     `json:"subtotal" binding:"required"` // decimal as string
	Discount      string                     `json:"discount" binding:"required"` // decimal as string
	Tax           string                     `json:"tax" binding:"required"` // decimal as string
	ShippingCost  string                     `json:"shippingCost" binding:"required"` // decimal as string
	TotalAmount   string                     `json:"totalAmount" binding:"required"` // decimal as string
	Items         []CreateSalesOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// UpdateSalesOrderRequest represents sales order update request
type UpdateSalesOrderRequest struct {
	CustomerId    *string                        `json:"customerId" binding:"omitempty"`
	WarehouseId   *string                        `json:"warehouseId" binding:"omitempty"`
	OrderDate     *string                        `json:"orderDate" binding:"omitempty"` // ISO 8601 date
	RequiredDate  *string                        `json:"requiredDate" binding:"omitempty"` // ISO 8601 date
	Notes         *string                        `json:"notes" binding:"omitempty"`
	Subtotal      *string                        `json:"subtotal" binding:"omitempty"` // decimal as string
	Discount      *string                        `json:"discount" binding:"omitempty"` // decimal as string
	Tax           *string                        `json:"tax" binding:"omitempty"` // decimal as string
	ShippingCost  *string                        `json:"shippingCost" binding:"omitempty"` // decimal as string
	TotalAmount   *string                        `json:"totalAmount" binding:"omitempty"` // decimal as string
	Items         []UpdateSalesOrderItemRequest  `json:"items" binding:"omitempty,dive"`
}

// SalesOrderFilters represents sales order list filters
type SalesOrderFilters struct {
	Search      string  `form:"search"` // Search by order number, customer name
	Status      *string `form:"status"` // Filter by status
	CustomerId  string  `form:"customer_id"` // Filter by customer
	WarehouseId string  `form:"warehouse_id"` // Filter by warehouse
	FromDate    *string `form:"from_date"` // ISO 8601 date
	ToDate      *string `form:"to_date"` // ISO 8601 date
	Page        int     `form:"page" binding:"omitempty,min=1"`
	Limit       int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy      string  `form:"sort_by" binding:"omitempty,oneof=orderNumber orderDate totalAmount"`
	SortOrder   string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// SALES ORDER ITEM REQUEST DTOs
// ============================================================================

// CreateSalesOrderItemRequest represents sales order item creation
type CreateSalesOrderItemRequest struct {
	ProductId   string  `json:"productId" binding:"required"`
	UnitId      string  `json:"unitId" binding:"required"`
	OrderedQty  string  `json:"orderedQty" binding:"required"` // decimal as string
	UnitPrice   string  `json:"unitPrice" binding:"required"` // decimal as string
	Discount    string  `json:"discount" binding:"required"` // decimal as string
	LineTotal   string  `json:"lineTotal" binding:"required"` // decimal as string
	Notes       *string `json:"notes" binding:"omitempty"`
}

// UpdateSalesOrderItemRequest represents sales order item update
type UpdateSalesOrderItemRequest struct {
	Id          *string `json:"id" binding:"omitempty"` // For updating existing items
	ProductId   *string `json:"productId" binding:"omitempty"`
	UnitId      *string `json:"unitId" binding:"omitempty"`
	OrderedQty  *string `json:"orderedQty" binding:"omitempty"` // decimal as string
	UnitPrice   *string `json:"unitPrice" binding:"omitempty"` // decimal as string
	Discount    *string `json:"discount" binding:"omitempty"` // decimal as string
	LineTotal   *string `json:"lineTotal" binding:"omitempty"` // decimal as string
	Notes       *string `json:"notes" binding:"omitempty"`
}

// ============================================================================
// SALES ORDER RESPONSE DTOs
// ============================================================================

// SalesOrderResponse represents sales order information response
type SalesOrderResponse struct {
	Id             string                     `json:"id"`
	TenantId       string                     `json:"tenantId"`
	CompanyId      string                     `json:"companyId"`
	OrderNumber    string                     `json:"orderNumber"`
	OrderDate      string                     `json:"orderDate"` // ISO 8601
	RequiredDate   *string                    `json:"requiredDate,omitempty"` // ISO 8601
	DeliveryDate   *string                    `json:"deliveryDate,omitempty"` // ISO 8601
	CustomerId     string                     `json:"customerId"`
	CustomerCode   string                     `json:"customerCode"`
	CustomerName   string                     `json:"customerName"`
	WarehouseId    string                     `json:"warehouseId"`
	WarehouseCode  string                     `json:"warehouseCode"`
	WarehouseName  string                     `json:"warehouseName"`
	Status         string                     `json:"status"`
	Subtotal       string                     `json:"subtotal"` // decimal as string
	Discount       string                     `json:"discount"` // decimal as string
	Tax            string                     `json:"tax"` // decimal as string
	ShippingCost   string                     `json:"shippingCost"` // decimal as string
	TotalAmount    string                     `json:"totalAmount"` // decimal as string
	Notes          *string                    `json:"notes,omitempty"`
	Items          []SalesOrderItemResponse   `json:"items,omitempty"`
	ApprovedBy     *string                    `json:"approvedBy,omitempty"`
	ApprovedAt     *string                    `json:"approvedAt,omitempty"` // ISO 8601
	CancelledBy    *string                    `json:"cancelledBy,omitempty"`
	CancelledAt    *string                    `json:"cancelledAt,omitempty"` // ISO 8601
	CreatedAt      time.Time                  `json:"createdAt"`
	UpdatedAt      time.Time                  `json:"updatedAt"`
}

// SalesOrderItemResponse represents sales order item information
type SalesOrderItemResponse struct {
	Id           string  `json:"id"`
	ProductId    string  `json:"productId"`
	ProductCode  string  `json:"productCode"`
	ProductName  string  `json:"productName"`
	UnitId       *string `json:"unitId,omitempty"`
	UnitName     string  `json:"unitName"` // Always present (base unit if no unit specified)
	OrderedQty   string  `json:"orderedQty"` // decimal as string
	UnitPrice    string  `json:"unitPrice"` // decimal as string
	Discount     string  `json:"discount"` // decimal as string
	LineTotal    string  `json:"lineTotal"` // decimal as string
	Notes        *string `json:"notes,omitempty"`
}

// ============================================================================
// PAGINATION & LIST RESPONSE
// ============================================================================

// SalesOrderListResponse represents paginated sales order list
type SalesOrderListResponse struct {
	Success    bool                 `json:"success"`
	Data       []SalesOrderResponse `json:"data"`
	Pagination PaginationInfo       `json:"pagination"`
}

// ============================================================================
// STANDARD API RESPONSES
// ============================================================================

// SalesOrderDetailResponse represents single sales order response
type SalesOrderDetailResponse struct {
	Success bool               `json:"success"`
	Data    SalesOrderResponse `json:"data"`
}

// ============================================================================
// STATUS TRANSITION DTOs
// ============================================================================

// SubmitSalesOrderRequest represents submitting draft to pending
type SubmitSalesOrderRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// ApproveSalesOrderRequest represents approving pending order
type ApproveSalesOrderRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// StartProcessingSalesOrderRequest represents starting order processing
type StartProcessingSalesOrderRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// ShipSalesOrderRequest represents shipping the order
type ShipSalesOrderRequest struct {
	DeliveryDate    *string `json:"deliveryDate" binding:"omitempty"` // ISO 8601
	DeliveryAddress *string `json:"deliveryAddress" binding:"omitempty"`
	Notes           *string `json:"notes" binding:"omitempty"`
}

// DeliverSalesOrderRequest represents marking order as delivered
type DeliverSalesOrderRequest struct {
	DeliveryDate *string `json:"deliveryDate" binding:"omitempty"` // ISO 8601
	Notes        *string `json:"notes" binding:"omitempty"`
}

// CompleteSalesOrderRequest represents completing the order
type CompleteSalesOrderRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// CancelSalesOrderRequest represents cancelling the order
type CancelSalesOrderRequest struct {
	Reason string `json:"reason" binding:"required,min=5"`
}
