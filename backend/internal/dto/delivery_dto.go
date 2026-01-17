package dto

import "time"

// ============================================================================
// DELIVERY REQUEST DTOs
// ============================================================================

// CreateDeliveryRequest represents delivery creation request
type CreateDeliveryRequest struct {
	SalesOrderId     string                     `json:"salesOrderId" binding:"required"`
	DeliveryDate     string                     `json:"deliveryDate" binding:"required"` // ISO 8601 date
	WarehouseId      string                     `json:"warehouseId" binding:"required"`
	CustomerId       string                     `json:"customerId" binding:"required"`
	Type             string                     `json:"type" binding:"required,oneof=NORMAL RETURN REPLACEMENT"`
	DeliveryAddress  *string                    `json:"deliveryAddress" binding:"omitempty"`
	DriverName       *string                    `json:"driverName" binding:"omitempty"`
	VehicleNumber    *string                    `json:"vehicleNumber" binding:"omitempty"`
	ExpeditionService *string                   `json:"expeditionService" binding:"omitempty"`
	TtnkNumber       *string                    `json:"ttnkNumber" binding:"omitempty"` // Tracking number
	Notes            *string                    `json:"notes" binding:"omitempty"`
	Items            []CreateDeliveryItemRequest `json:"items" binding:"required,min=1,dive"`
}

// CreateDeliveryItemRequest represents delivery item creation
type CreateDeliveryItemRequest struct {
	SalesOrderItemId *string `json:"salesOrderItemId" binding:"omitempty"`
	ProductId        string  `json:"productId" binding:"required"`
	ProductUnitId    *string `json:"productUnitId" binding:"omitempty"`
	BatchId          *string `json:"batchId" binding:"omitempty"`
	Quantity         string  `json:"quantity" binding:"required"` // decimal as string
	Notes            *string `json:"notes" binding:"omitempty"`
}

// UpdateDeliveryStatusRequest represents delivery status update
type UpdateDeliveryStatusRequest struct {
	Status        *string `json:"status" binding:"omitempty,oneof=PREPARED IN_TRANSIT DELIVERED CONFIRMED CANCELLED"`
	DepartureTime *string `json:"departureTime" binding:"omitempty"` // ISO 8601 datetime
	ArrivalTime   *string `json:"arrivalTime" binding:"omitempty"`   // ISO 8601 datetime
	ReceivedBy    *string `json:"receivedBy" binding:"omitempty"`
	SignatureUrl  *string `json:"signatureUrl" binding:"omitempty"`
	PhotoUrl      *string `json:"photoUrl" binding:"omitempty"`
}

// StartDeliveryRequest represents start delivery request (PREPARED -> IN_TRANSIT)
type StartDeliveryRequest struct {
	DepartureTime string `json:"departureTime" binding:"required"` // ISO 8601 datetime
}

// CompleteDeliveryRequest represents complete delivery request (IN_TRANSIT -> DELIVERED)
type CompleteDeliveryRequest struct {
	ReceivedBy   *string `json:"receivedBy" binding:"omitempty"`
	SignatureUrl *string `json:"signatureUrl" binding:"omitempty"`
	PhotoUrl     *string `json:"photoUrl" binding:"omitempty"`
}

// CancelDeliveryRequest represents cancel delivery request
type CancelDeliveryRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// DeliveryFilters represents delivery list filters
type DeliveryFilters struct {
	Search      string  `form:"search"` // Search by delivery number, customer name
	Status      *string `form:"status"` // Filter by status
	Type        *string `form:"type"`   // Filter by delivery type
	CustomerId  string  `form:"customer_id"` // Filter by customer
	WarehouseId string  `form:"warehouse_id"` // Filter by warehouse
	FromDate    *string `form:"from_date"` // ISO 8601 date
	ToDate      *string `form:"to_date"` // ISO 8601 date
	Page        int     `form:"page" binding:"omitempty,min=1"`
	Limit       int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy      string  `form:"sort_by" binding:"omitempty,oneof=deliveryNumber deliveryDate"`
	SortOrder   string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// DELIVERY RESPONSE DTOs
// ============================================================================

// DeliveryResponse represents delivery response
type DeliveryResponse struct {
	Id                string                   `json:"id"`
	DeliveryNumber    string                   `json:"deliveryNumber"`
	DeliveryDate      time.Time                `json:"deliveryDate"`
	SalesOrderId      string                   `json:"salesOrderId"`
	WarehouseId       string                   `json:"warehouseId"`
	CustomerId        string                   `json:"customerId"`
	Type              string                   `json:"type"`
	Status            string                   `json:"status"`
	DeliveryAddress   *string                  `json:"deliveryAddress,omitempty"`
	DriverName        *string                  `json:"driverName,omitempty"`
	VehicleNumber     *string                  `json:"vehicleNumber,omitempty"`
	ExpeditionService *string                  `json:"expeditionService,omitempty"`
	TtnkNumber        *string                  `json:"ttnkNumber,omitempty"`
	DepartureTime     *time.Time               `json:"departureTime,omitempty"`
	ArrivalTime       *time.Time               `json:"arrivalTime,omitempty"`
	ReceivedBy        *string                  `json:"receivedBy,omitempty"`
	ReceivedAt        *time.Time               `json:"receivedAt,omitempty"`
	SignatureUrl      *string                  `json:"signatureUrl,omitempty"`
	PhotoUrl          *string                  `json:"photoUrl,omitempty"`
	Notes             *string                  `json:"notes,omitempty"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`

	// Relations
	SalesOrder *SalesOrderSummary `json:"salesOrder,omitempty"`
	Warehouse  *WarehouseSummary  `json:"warehouse,omitempty"`
	Customer   *CustomerSummary   `json:"customer,omitempty"`
	Items      []DeliveryItemResponse `json:"items,omitempty"`
}

// DeliveryItemResponse represents delivery item response
type DeliveryItemResponse struct {
	Id               string            `json:"id"`
	DeliveryId       string            `json:"deliveryId"`
	SalesOrderItemId *string           `json:"salesOrderItemId,omitempty"`
	ProductId        string            `json:"productId"`
	ProductUnitId    *string           `json:"productUnitId,omitempty"`
	BatchId          *string           `json:"batchId,omitempty"`
	Quantity         string            `json:"quantity"` // decimal as string
	Notes            *string           `json:"notes,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`

	// Relations
	Product     *ProductSummary     `json:"product,omitempty"`
	ProductUnit *ProductUnitSummary `json:"productUnit,omitempty"`
	Batch       *BatchSummary       `json:"batch,omitempty"`
}

// DeliverySummary represents minimal delivery info for references
type DeliverySummary struct {
	Id             string    `json:"id"`
	DeliveryNumber string    `json:"deliveryNumber"`
	DeliveryDate   time.Time `json:"deliveryDate"`
	Status         string    `json:"status"`
	Type           string    `json:"type"`
}

// DeliveryListResponse represents paginated delivery list
type DeliveryListResponse struct {
	Data       []DeliveryResponse `json:"data"`
	Pagination PaginationInfo     `json:"pagination"`
}

// ============================================================================
// SUMMARY DTOs (for nested references)
// ============================================================================

// SalesOrderSummary represents minimal sales order info for references
type SalesOrderSummary struct {
	Id       string `json:"id"`
	SoNumber string `json:"soNumber"`
	SoDate   string `json:"soDate"` // ISO 8601
	Status   string `json:"status"`
}

// WarehouseSummary represents minimal warehouse info for references
type WarehouseSummary struct {
	Id   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// CustomerSummary represents minimal customer info for references
type CustomerSummary struct {
	Id    string  `json:"id"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Phone *string `json:"phone,omitempty"`
}

// ProductSummary represents minimal product info for references
type ProductSummary struct {
	Id       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	BaseUnit string `json:"baseUnit"`
}

// ProductUnitSummary represents minimal product unit info for references
type ProductUnitSummary struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// BatchSummary represents minimal batch info for references
type BatchSummary struct {
	Id          string     `json:"id"`
	BatchNumber string     `json:"batchNumber"`
	ExpiryDate  *time.Time `json:"expiryDate,omitempty"`
}
