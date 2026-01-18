package dto

import "time"

// ============================================================================
// PRODUCT REQUEST DTOs
// ============================================================================

// CreateProductRequest represents product creation request
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 260-282
type CreateProductRequest struct {
	Code           string                     `json:"code" binding:"required,min=1,max=100"`
	Name           string                     `json:"name" binding:"required,min=2,max=255"`
	Category       *string                    `json:"category" binding:"omitempty,max=100"`
	BaseUnit       string                     `json:"baseUnit" binding:"required,max=20"`
	BaseCost       string                     `json:"baseCost" binding:"required"` // decimal as string
	BasePrice      string                     `json:"basePrice" binding:"required"` // decimal as string
	MinimumStock   string                     `json:"minimumStock" binding:"omitempty"` // decimal as string, default 0
	Description    *string                    `json:"description" binding:"omitempty"`
	Barcode        *string                    `json:"barcode" binding:"omitempty,max=100"`
	IsBatchTracked bool                       `json:"isBatchTracked"`
	IsPerishable   bool                       `json:"isPerishable"`
	Units          []CreateProductUnitRequest `json:"units" binding:"omitempty,dive"`
}

// UpdateProductRequest represents product update request
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 310-325
type UpdateProductRequest struct {
	Code           *string                         `json:"code" binding:"omitempty,min=1,max=100"` // Added: Allow code updates with validation
	Name           *string                         `json:"name" binding:"omitempty,min=2,max=255"`
	Category       *string                         `json:"category" binding:"omitempty,max=100"`
	BaseUnit       *string                         `json:"baseUnit" binding:"omitempty,max=20"` // Added: Allow baseUnit updates
	BaseCost       *string                         `json:"baseCost" binding:"omitempty"` // decimal as string
	BasePrice      *string                         `json:"basePrice" binding:"omitempty"` // decimal as string
	MinimumStock   *string                         `json:"minimumStock" binding:"omitempty"` // decimal as string
	Description    *string                         `json:"description" binding:"omitempty"`
	Barcode        *string                         `json:"barcode" binding:"omitempty,max=100"`
	IsBatchTracked *bool                           `json:"isBatchTracked" binding:"omitempty"`
	IsPerishable   *bool                           `json:"isPerishable" binding:"omitempty"`
	IsActive       *bool                           `json:"isActive" binding:"omitempty"`
	Suppliers      *UpdateProductSuppliersRequest  `json:"suppliers" binding:"omitempty"` // Optional: supplier changes
	Units          *UpdateProductUnitsRequest      `json:"units" binding:"omitempty"` // Optional: unit changes
}

// UpdateProductSuppliersRequest represents supplier changes in product update
type UpdateProductSuppliersRequest struct {
	Add    []AddProductSupplierRequest    `json:"add" binding:"omitempty,dive"`
	Update []UpdateProductSupplierItem    `json:"update" binding:"omitempty,dive"`
	Delete []string                        `json:"delete" binding:"omitempty"` // ProductSupplier IDs to delete
}

// UpdateProductSupplierItem represents a supplier update with ID
type UpdateProductSupplierItem struct {
	ID            string  `json:"id" binding:"required"` // ProductSupplier ID
	SupplierPrice *string `json:"supplierPrice" binding:"omitempty"`
	LeadTime      *int    `json:"leadTime" binding:"omitempty,min=0"`
	IsPrimary     *bool   `json:"isPrimary" binding:"omitempty"`
}

// UpdateProductUnitsRequest represents unit changes in product update
type UpdateProductUnitsRequest struct {
	Update []UpdateProductUnitItem `json:"update" binding:"omitempty,dive"`
}

// UpdateProductUnitItem represents a unit update with ID
type UpdateProductUnitItem struct {
	ID       string  `json:"id" binding:"required"` // ProductUnit ID
	UnitName *string `json:"unitName" binding:"omitempty,min=1,max=50"`
}

// ProductFilters represents product list filters
type ProductFilters struct {
	Search         string  `form:"search"`
	Category       string  `form:"category"`
	SupplierID     string  `form:"supplier_id"` // Filter products by supplier
	IsActive       *bool   `form:"is_active"`
	IsBatchTracked *bool   `form:"is_batch_tracked"`
	IsPerishable   *bool   `form:"is_perishable"`
	Page           int     `form:"page" binding:"omitempty,min=1"`
	Limit          int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy         string  `form:"sort_by" binding:"omitempty,oneof=code name createdAt"`
	SortOrder      string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// PRODUCT UNIT REQUEST DTOs
// ============================================================================

// CreateProductUnitRequest represents product unit creation request
type CreateProductUnitRequest struct {
	UnitName       string  `json:"unitName" binding:"required,min=1,max=50"`
	ConversionRate string  `json:"conversionRate" binding:"required"` // decimal as string, must be > 0
	BuyPrice       *string `json:"buyPrice" binding:"omitempty"` // decimal as string
	SellPrice      *string `json:"sellPrice" binding:"omitempty"` // decimal as string
	Barcode        *string `json:"barcode" binding:"omitempty,max=100"`
	SKU            *string `json:"sku" binding:"omitempty,max=100"`
	Weight         *string `json:"weight" binding:"omitempty"` // decimal as string (kg)
	Volume         *string `json:"volume" binding:"omitempty"` // decimal as string (m³)
	Description    *string `json:"description" binding:"omitempty"`
}

// UpdateProductUnitRequest represents product unit update request
type UpdateProductUnitRequest struct {
	UnitName       *string `json:"unitName" binding:"omitempty,min=1,max=50"`
	ConversionRate *string `json:"conversionRate" binding:"omitempty"` // decimal as string
	BuyPrice       *string `json:"buyPrice" binding:"omitempty"` // decimal as string
	SellPrice      *string `json:"sellPrice" binding:"omitempty"` // decimal as string
	Barcode        *string `json:"barcode" binding:"omitempty,max=100"`
	SKU            *string `json:"sku" binding:"omitempty,max=100"`
	Weight         *string `json:"weight" binding:"omitempty"` // decimal as string
	Volume         *string `json:"volume" binding:"omitempty"` // decimal as string
	Description    *string `json:"description" binding:"omitempty"`
	IsActive       *bool   `json:"isActive" binding:"omitempty"`
}

// ============================================================================
// PRODUCT SUPPLIER REQUEST DTOs
// ============================================================================

// AddProductSupplierRequest represents linking supplier to product
type AddProductSupplierRequest struct {
	SupplierID    string `json:"supplierId" binding:"required"`
	SupplierPrice string `json:"supplierPrice" binding:"required"` // decimal as string
	LeadTime      int    `json:"leadTime" binding:"omitempty,min=0"` // days, default 7
	IsPrimary     bool   `json:"isPrimary"`
}

// UpdateProductSupplierRequest represents updating product-supplier relationship
type UpdateProductSupplierRequest struct {
	SupplierPrice *string `json:"supplierPrice" binding:"omitempty"` // decimal as string
	LeadTime      *int    `json:"leadTime" binding:"omitempty,min=0"`
	IsPrimary     *bool   `json:"isPrimary" binding:"omitempty"`
}

// ============================================================================
// PRODUCT RESPONSE DTOs
// ============================================================================

// ProductResponse represents product information response
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 133-178
type ProductResponse struct {
	ID             string                    `json:"id"`
	Code           string                    `json:"code"`
	Name           string                    `json:"name"`
	Category       *string                   `json:"category,omitempty"`
	BaseUnit       string                    `json:"baseUnit"`
	BaseCost       string                    `json:"baseCost"` // decimal as string
	BasePrice      string                    `json:"basePrice"` // decimal as string
	MinimumStock   string                    `json:"minimumStock"` // decimal as string
	Description    *string                   `json:"description,omitempty"`
	Barcode        *string                   `json:"barcode,omitempty"`
	IsBatchTracked bool                      `json:"isBatchTracked"`
	IsPerishable   bool                      `json:"isPerishable"`
	IsActive       bool                      `json:"isActive"`
	Units          []ProductUnitResponse     `json:"units,omitempty"`
	Suppliers      []ProductSupplierResponse `json:"suppliers,omitempty"`
	CurrentStock   *CurrentStockResponse     `json:"currentStock,omitempty"`
	CreatedAt      time.Time                 `json:"createdAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
}

// ProductUnitResponse represents product unit information
type ProductUnitResponse struct {
	ID             string  `json:"id"`
	UnitName       string  `json:"unitName"`
	ConversionRate string  `json:"conversionRate"` // decimal as string
	IsBaseUnit     bool    `json:"isBaseUnit"`
	BuyPrice       *string `json:"buyPrice,omitempty"` // decimal as string
	SellPrice      *string `json:"sellPrice,omitempty"` // decimal as string
	Barcode        *string `json:"barcode,omitempty"`
	SKU            *string `json:"sku,omitempty"`
	Weight         *string `json:"weight,omitempty"` // decimal as string
	Volume         *string `json:"volume,omitempty"` // decimal as string
	Description    *string `json:"description,omitempty"`
	IsActive       bool    `json:"isActive"`
}

// ProductSupplierResponse represents product-supplier relationship
type ProductSupplierResponse struct {
	ID                string `json:"id"`
	SupplierID        string `json:"supplierId"`
	SupplierCode      string `json:"supplierCode"`
	SupplierName      string `json:"supplierName"`
	SupplierPrice     string `json:"supplierPrice"` // decimal as string
	LeadTimeDays      int    `json:"leadTimeDays"` // days
	IsPrimarySupplier bool   `json:"isPrimarySupplier"`
}

// CurrentStockResponse represents aggregated stock information
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 159-169
type CurrentStockResponse struct {
	Total      string                    `json:"total"` // decimal as string
	Warehouses []WarehouseStockInfo      `json:"warehouses,omitempty"`
}

// WarehouseStockInfo represents stock in specific warehouse
type WarehouseStockInfo struct {
	WarehouseID   string `json:"warehouseId"`
	WarehouseName string `json:"warehouseName"`
	Quantity      string `json:"quantity"` // decimal as string
}

// ============================================================================
// PAGINATION & LIST RESPONSE
// ============================================================================

// ProductListResponse represents paginated product list
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 132-179
type ProductListResponse struct {
	Success    bool              `json:"success"`
	Data       []ProductResponse `json:"data"`
	Pagination PaginationInfo    `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// ============================================================================
// STANDARD API RESPONSES
// ============================================================================

// ProductDetailResponse represents single product response
type ProductDetailResponse struct {
	Success bool            `json:"success"`
	Data    ProductResponse `json:"data"`
}

// MessageResponse represents simple message response
type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
