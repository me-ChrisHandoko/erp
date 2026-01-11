package dto

import (
	"time"
)

// ============================================================================
// WAREHOUSE MANAGEMENT DTOs
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 4 - Warehouse Management
// ============================================================================

// CreateWarehouseRequest - Request to create a new warehouse
type CreateWarehouseRequest struct {
	Code       string  `json:"code" binding:"required,min=1,max=50"`
	Name       string  `json:"name" binding:"required,min=2,max=255"`
	Type       string  `json:"type" binding:"required,oneof=MAIN BRANCH CONSIGNMENT TRANSIT"`
	Address    *string `json:"address" binding:"omitempty"`
	City       *string `json:"city" binding:"omitempty,max=100"`
	Province   *string `json:"province" binding:"omitempty,max=100"`
	PostalCode *string `json:"postalCode" binding:"omitempty,max=50"`
	Phone      *string `json:"phone" binding:"omitempty,max=50"`
	Email      *string `json:"email" binding:"omitempty,email,max=255"`
	ManagerID  *string `json:"managerID" binding:"omitempty"`
	Capacity   *string `json:"capacity" binding:"omitempty"` // decimal as string
}

// UpdateWarehouseRequest - Request to update an existing warehouse
type UpdateWarehouseRequest struct {
	Code       *string `json:"code" binding:"omitempty,min=1,max=50"`
	Name       *string `json:"name" binding:"omitempty,min=2,max=255"`
	Type       *string `json:"type" binding:"omitempty,oneof=MAIN BRANCH CONSIGNMENT TRANSIT"`
	Address    *string `json:"address" binding:"omitempty"`
	City       *string `json:"city" binding:"omitempty,max=100"`
	Province   *string `json:"province" binding:"omitempty,max=100"`
	PostalCode *string `json:"postalCode" binding:"omitempty,max=50"`
	Phone      *string `json:"phone" binding:"omitempty,max=50"`
	Email      *string `json:"email" binding:"omitempty,email,max=255"`
	ManagerID  *string `json:"managerID" binding:"omitempty"`
	Capacity   *string `json:"capacity" binding:"omitempty"`
	IsActive   *bool   `json:"isActive" binding:"omitempty"`
}

// WarehouseResponse - Response DTO for warehouse data
type WarehouseResponse struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Address    *string    `json:"address,omitempty"`
	City       *string    `json:"city,omitempty"`
	Province   *string    `json:"province,omitempty"`
	PostalCode *string    `json:"postalCode,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	Email      *string    `json:"email,omitempty"`
	ManagerID  *string    `json:"managerID,omitempty"`
	Capacity   *string    `json:"capacity,omitempty"`
	IsActive   bool       `json:"isActive"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// WarehouseListResponse - Response DTO for warehouse list with pagination
type WarehouseListResponse struct {
	Success    bool                `json:"success"`
	Data       []WarehouseResponse `json:"data"`
	Pagination PaginationInfo      `json:"pagination"`
}

// WarehouseListQuery - Query parameters for listing warehouses
type WarehouseListQuery struct {
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PageSize  int     `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Search    string  `form:"search" binding:"omitempty"` // Search by code or name
	Type      *string `form:"type" binding:"omitempty,oneof=MAIN BRANCH CONSIGNMENT TRANSIT"`
	City      *string `form:"city" binding:"omitempty"`
	Province  *string `form:"province" binding:"omitempty"`
	ManagerID *string `form:"managerID" binding:"omitempty"`
	IsActive  *bool   `form:"isActive" binding:"omitempty"`
	SortBy    string  `form:"sortBy" binding:"omitempty,oneof=code name type createdAt"`
	SortOrder string  `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// WAREHOUSE STOCK DTOs
// ============================================================================

// WarehouseStockResponse - Response DTO for warehouse stock data
type WarehouseStockResponse struct {
	ID            string     `json:"id"`
	WarehouseID   string     `json:"warehouseID"`
	WarehouseName string     `json:"warehouseName"`
	WarehouseCode string     `json:"warehouseCode"`
	ProductID     string     `json:"productID"`
	ProductCode   string     `json:"productCode"`
	ProductName   string     `json:"productName"`
	Quantity      string     `json:"quantity"`
	MinimumStock  string     `json:"minimumStock"`
	MaximumStock  string     `json:"maximumStock"`
	Location      *string    `json:"location,omitempty"`
	LastCountDate *time.Time `json:"lastCountDate,omitempty"`
	LastCountQty  *string    `json:"lastCountQty,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// WarehouseStockListResponse - Response DTO for warehouse stock list with pagination
type WarehouseStockListResponse struct {
	Stocks     []WarehouseStockResponse `json:"stocks"`
	TotalCount int64                    `json:"totalCount"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"pageSize"`
	TotalPages int                      `json:"totalPages"`
}

// WarehouseStockListQuery - Query parameters for listing warehouse stocks
type WarehouseStockListQuery struct {
	Page         int     `form:"page" binding:"omitempty,min=1"`
	PageSize     int     `form:"pageSize" binding:"omitempty,min=1,max=1000"` // Increased to 1000 for Initial Stock Setup
	WarehouseID  *string `form:"warehouseID" binding:"omitempty"`
	ProductID    *string `form:"productID" binding:"omitempty"`
	Search       string  `form:"search" binding:"omitempty"` // Search by product code or name
	LowStock     *bool   `form:"lowStock" binding:"omitempty"` // Filter products below minimum stock
	ZeroStock    *bool   `form:"zeroStock" binding:"omitempty"` // Filter products with zero stock
	SortBy       string  `form:"sortBy" binding:"omitempty,oneof=productCode productName quantity createdAt"`
	SortOrder    string  `form:"sortOrder" binding:"omitempty,oneof=asc desc"`
}

// UpdateWarehouseStockRequest - Request to update warehouse stock settings
type UpdateWarehouseStockRequest struct {
	MinimumStock *string `json:"minimumStock" binding:"omitempty"`
	MaximumStock *string `json:"maximumStock" binding:"omitempty"`
	Location     *string `json:"location" binding:"omitempty,max=100"`
}

// ============================================================================
// WAREHOUSE STATISTICS DTOs (for future use)
// ============================================================================

// WarehouseStatisticsResponse - Statistics about warehouse capacity and utilization
type WarehouseStatisticsResponse struct {
	TotalWarehouses        int    `json:"totalWarehouses"`
	ActiveWarehouses       int    `json:"activeWarehouses"`
	TotalCapacity          string `json:"totalCapacity"`
	TotalProducts          int    `json:"totalProducts"`
	ProductsBelowMinimum   int    `json:"productsBelowMinimum"`
	ProductsWithZeroStock  int    `json:"productsWithZeroStock"`
}

// WarehouseStockStatusResponse - Response DTO for warehouse stock initialization status
type WarehouseStockStatusResponse struct {
	WarehouseID    string     `json:"warehouseId"`
	WarehouseName  string     `json:"warehouseName"`
	WarehouseCode  string     `json:"warehouseCode,omitempty"`
	HasInitialStock bool      `json:"hasInitialStock"`
	TotalProducts  int        `json:"totalProducts"`
	TotalValue     string     `json:"totalValue"`
	LastUpdated    *time.Time `json:"lastUpdated,omitempty"`
}

// WarehouseStockStatusListResponse - Response DTO for warehouse stock status list
type WarehouseStockStatusListResponse struct {
	Success    bool                           `json:"success"`
	Warehouses []WarehouseStockStatusResponse `json:"warehouses"`
}
