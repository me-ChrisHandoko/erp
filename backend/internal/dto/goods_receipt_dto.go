package dto

import (
	"time"
)

// ============================================================================
// GOODS RECEIPT MANAGEMENT DTOs
// Reference: Phase 3 - Goods Receipt Management (Penerimaan Barang)
// ============================================================================

// CreateGoodsReceiptItemRequest - Request to create a goods receipt item
type CreateGoodsReceiptItemRequest struct {
	PurchaseOrderItemID string  `json:"purchaseOrderItemId" binding:"required,uuid"`
	ProductID           string  `json:"productId" binding:"required,uuid"`
	ProductUnitID       *string `json:"productUnitId" binding:"omitempty,uuid"`
	BatchNumber         *string `json:"batchNumber" binding:"omitempty,max=100"`
	ManufactureDate     *string `json:"manufactureDate" binding:"omitempty"`     // ISO date
	ExpiryDate          *string `json:"expiryDate" binding:"omitempty"`          // ISO date
	ReceivedQty         string  `json:"receivedQty" binding:"required"`          // decimal as string
	AcceptedQty         string  `json:"acceptedQty" binding:"omitempty"`         // decimal as string
	RejectedQty         string  `json:"rejectedQty" binding:"omitempty"`         // decimal as string
	RejectionReason     *string `json:"rejectionReason" binding:"omitempty"`
	QualityNote         *string `json:"qualityNote" binding:"omitempty"`
	Notes               *string `json:"notes" binding:"omitempty"`
}

// CreateGoodsReceiptRequest - Request to create a new goods receipt
type CreateGoodsReceiptRequest struct {
	PurchaseOrderID  string                          `json:"purchaseOrderId" binding:"required,uuid"`
	GRNDate          string                          `json:"grnDate" binding:"required"` // ISO date string
	SupplierInvoice  *string                         `json:"supplierInvoice" binding:"omitempty,max=100"`
	SupplierDONumber *string                         `json:"supplierDONumber" binding:"omitempty,max=100"`
	Notes            *string                         `json:"notes" binding:"omitempty"`
	Items            []CreateGoodsReceiptItemRequest `json:"items" binding:"required,min=1,dive"`
}

// UpdateGoodsReceiptItemRequest - Request to update a goods receipt item
type UpdateGoodsReceiptItemRequest struct {
	ID              *string `json:"id" binding:"omitempty,uuid"` // Existing item ID (if updating)
	ReceivedQty     string  `json:"receivedQty" binding:"required"`
	AcceptedQty     string  `json:"acceptedQty" binding:"omitempty"`
	RejectedQty     string  `json:"rejectedQty" binding:"omitempty"`
	BatchNumber     *string `json:"batchNumber" binding:"omitempty,max=100"`
	ManufactureDate *string `json:"manufactureDate" binding:"omitempty"`
	ExpiryDate      *string `json:"expiryDate" binding:"omitempty"`
	RejectionReason *string `json:"rejectionReason" binding:"omitempty"`
	QualityNote     *string `json:"qualityNote" binding:"omitempty"`
	Notes           *string `json:"notes" binding:"omitempty"`
}

// UpdateGoodsReceiptRequest - Request to update an existing goods receipt
type UpdateGoodsReceiptRequest struct {
	GRNDate          *string                         `json:"grnDate" binding:"omitempty"`
	SupplierInvoice  *string                         `json:"supplierInvoice" binding:"omitempty,max=100"`
	SupplierDONumber *string                         `json:"supplierDONumber" binding:"omitempty,max=100"`
	Notes            *string                         `json:"notes" binding:"omitempty"`
	Items            []UpdateGoodsReceiptItemRequest `json:"items" binding:"omitempty,dive"`
}

// ReceiveGoodsRequest - Request to mark goods as received
type ReceiveGoodsRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// InspectGoodsRequest - Request to mark goods as inspected
type InspectGoodsRequest struct {
	Notes *string                         `json:"notes" binding:"omitempty"`
	Items []UpdateGoodsReceiptItemRequest `json:"items" binding:"omitempty,dive"`
}

// AcceptGoodsRequest - Request to accept goods (stock will be updated)
type AcceptGoodsRequest struct {
	Notes *string `json:"notes" binding:"omitempty"`
}

// RejectGoodsRequest - Request to reject goods
type RejectGoodsRequest struct {
	RejectionReason string `json:"rejectionReason" binding:"required,min=1,max=500"`
}

// GoodsReceiptProductResponse - Product info for goods receipt items
type GoodsReceiptProductResponse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	BaseUnit string `json:"baseUnit"`
}

// GoodsReceiptProductUnitResponse - Product unit info for goods receipt items
type GoodsReceiptProductUnitResponse struct {
	ID             string `json:"id"`
	UnitName       string `json:"unitName"`
	ConversionRate string `json:"conversionRate"`
}

// GoodsReceiptItemResponse - Response DTO for goods receipt item
type GoodsReceiptItemResponse struct {
	ID                  string                           `json:"id"`
	GoodsReceiptID      string                           `json:"goodsReceiptId"`
	PurchaseOrderItemID string                           `json:"purchaseOrderItemId"`
	ProductID           string                           `json:"productId"`
	Product             *GoodsReceiptProductResponse     `json:"product,omitempty"`
	ProductUnitID       *string                          `json:"productUnitId,omitempty"`
	ProductUnit         *GoodsReceiptProductUnitResponse `json:"productUnit,omitempty"`
	BatchNumber         *string                          `json:"batchNumber,omitempty"`
	ManufactureDate     *time.Time                       `json:"manufactureDate,omitempty"`
	ExpiryDate          *time.Time                       `json:"expiryDate,omitempty"`
	OrderedQty          string                           `json:"orderedQty"`
	ReceivedQty         string                           `json:"receivedQty"`
	AcceptedQty         string                           `json:"acceptedQty"`
	RejectedQty         string                           `json:"rejectedQty"`
	RejectionReason     *string                          `json:"rejectionReason,omitempty"`
	QualityNote         *string                          `json:"qualityNote,omitempty"`
	Notes               *string                          `json:"notes,omitempty"`
	CreatedAt           time.Time                        `json:"createdAt"`
	UpdatedAt           time.Time                        `json:"updatedAt"`
}

// GoodsReceiptPurchaseOrderResponse - Purchase order info for goods receipt
type GoodsReceiptPurchaseOrderResponse struct {
	ID       string `json:"id"`
	PONumber string `json:"poNumber"`
	PODate   string `json:"poDate"`
}

// GoodsReceiptResponse - Response DTO for goods receipt
type GoodsReceiptResponse struct {
	ID               string                             `json:"id"`
	GRNNumber        string                             `json:"grnNumber"`
	GRNDate          time.Time                          `json:"grnDate"`
	PurchaseOrderID  string                             `json:"purchaseOrderId"`
	PurchaseOrder    *GoodsReceiptPurchaseOrderResponse `json:"purchaseOrder,omitempty"`
	WarehouseID      string                             `json:"warehouseId"`
	Warehouse        *WarehouseBasicResponse            `json:"warehouse,omitempty"`
	SupplierID       string                             `json:"supplierId"`
	Supplier         *SupplierBasicResponse             `json:"supplier,omitempty"`
	Status           string                             `json:"status"`
	ReceivedBy       *string                            `json:"receivedBy,omitempty"`
	Receiver         *UserBasicResponse                 `json:"receiver,omitempty"`
	ReceivedAt       *time.Time                         `json:"receivedAt,omitempty"`
	InspectedBy      *string                            `json:"inspectedBy,omitempty"`
	Inspector        *UserBasicResponse                 `json:"inspector,omitempty"`
	InspectedAt      *time.Time                         `json:"inspectedAt,omitempty"`
	SupplierInvoice  *string                            `json:"supplierInvoice,omitempty"`
	SupplierDONumber *string                            `json:"supplierDONumber,omitempty"`
	Notes            *string                            `json:"notes,omitempty"`
	Items            []GoodsReceiptItemResponse         `json:"items,omitempty"`
	CreatedAt        time.Time                          `json:"createdAt"`
	UpdatedAt        time.Time                          `json:"updatedAt"`
}

// GoodsReceiptListResponse - Response DTO for goods receipt list with pagination
type GoodsReceiptListResponse struct {
	Success    bool                   `json:"success"`
	Data       []GoodsReceiptResponse `json:"data"`
	Pagination PaginationInfo         `json:"pagination"`
}

// GoodsReceiptListQuery - Query parameters for listing goods receipts
type GoodsReceiptListQuery struct {
	Page            int     `form:"page" binding:"omitempty,min=1"`
	PageSize        int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search          string  `form:"search" binding:"omitempty"` // Search by GRN number
	Status          *string `form:"status" binding:"omitempty,oneof=PENDING RECEIVED INSPECTED ACCEPTED REJECTED PARTIAL"`
	PurchaseOrderID *string `form:"purchase_order_id" binding:"omitempty,uuid"`
	SupplierID      *string `form:"supplier_id" binding:"omitempty,uuid"`
	WarehouseID     *string `form:"warehouse_id" binding:"omitempty,uuid"`
	DateFrom        *string `form:"date_from" binding:"omitempty"` // ISO date string
	DateTo          *string `form:"date_to" binding:"omitempty"`   // ISO date string
	SortBy          string  `form:"sort_by" binding:"omitempty,oneof=grnNumber grnDate status createdAt"`
	SortOrder       string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
