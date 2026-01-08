package warehouse

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// WarehouseService - Business logic for warehouse management
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 4
type WarehouseService struct {
	db           *gorm.DB
	auditService *audit.AuditService
}

// NewWarehouseService creates a new warehouse service instance
func NewWarehouseService(db *gorm.DB, auditService *audit.AuditService) *WarehouseService {
	return &WarehouseService{
		db:           db,
		auditService: auditService,
	}
}

// ============================================================================
// CREATE WAREHOUSE
// ============================================================================

// CreateWarehouse creates a new warehouse
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *WarehouseService) CreateWarehouse(ctx context.Context, tenantID, companyID, userID, ipAddress, userAgent string, req *dto.CreateWarehouseRequest) (*models.Warehouse, error) {
	// Parse capacity
	var capacity *decimal.Decimal
	if req.Capacity != nil && *req.Capacity != "" {
		cap, err := decimal.NewFromString(*req.Capacity)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid capacity format")
		}
		if cap.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("capacity cannot be negative")
		}
		capacity = &cap
	}

	// Validate code uniqueness per company
	if err := s.validateCodeUniqueness(ctx, tenantID, companyID, req.Code, ""); err != nil {
		return nil, err
	}

	// Validate manager exists if provided
	if req.ManagerID != nil && *req.ManagerID != "" {
		if err := s.validateManagerExists(*req.ManagerID); err != nil {
			return nil, err
		}
	}

	// Create warehouse
	warehouse := &models.Warehouse{
		TenantID:   tenantID,
		CompanyID:  companyID,
		Code:       req.Code,
		Name:       req.Name,
		Type:       models.WarehouseType(req.Type),
		Address:    req.Address,
		City:       req.City,
		Province:   req.Province,
		PostalCode: req.PostalCode,
		Phone:      req.Phone,
		Email:      req.Email,
		ManagerID:  req.ManagerID,
		Capacity:   capacity,
		IsActive:   true,
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Create(warehouse).Error; err != nil {
		return nil, fmt.Errorf("failed to create warehouse: %w", err)
	}

	// Audit logging - Log successful warehouse creation
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	warehouseData := map[string]interface{}{
		"code":        warehouse.Code,
		"name":        warehouse.Name,
		"type":        string(warehouse.Type),
		"address":     warehouse.Address,
		"city":        warehouse.City,
		"province":    warehouse.Province,
		"postal_code": warehouse.PostalCode,
		"phone":       warehouse.Phone,
		"email":       warehouse.Email,
		"manager_id":  warehouse.ManagerID,
		"is_active":   warehouse.IsActive,
	}

	if warehouse.Capacity != nil {
		warehouseData["capacity"] = warehouse.Capacity.String()
	}

	if err := s.auditService.LogWarehouseCreated(ctx, auditCtx, warehouse.ID, warehouseData); err != nil {
		fmt.Printf("WARNING: Failed to create audit log: %v\n", err)
	}

	return warehouse, nil
}

// ============================================================================
// LIST WAREHOUSES
// ============================================================================

// ListWarehouses retrieves warehouses with filtering, sorting, and pagination
func (s *WarehouseService) ListWarehouses(ctx context.Context, tenantID, companyID string, query *dto.WarehouseListQuery) (*dto.WarehouseListResponse, error) {
	// Set defaults
	page := 1
	if query.Page > 0 {
		page = query.Page
	}

	pageSize := 20
	if query.PageSize > 0 {
		pageSize = query.PageSize
	}

	sortBy := "created_at"
	if query.SortBy != "" {
		sortBy = query.SortBy
	}

	sortOrder := "desc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}

	// Build base query with tenant context
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.Warehouse{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		baseQuery = baseQuery.Where("code LIKE ? OR name LIKE ?", searchPattern, searchPattern)
	}

	if query.Type != nil {
		baseQuery = baseQuery.Where("type = ?", *query.Type)
	}

	if query.City != nil {
		baseQuery = baseQuery.Where("city = ?", *query.City)
	}

	if query.Province != nil {
		baseQuery = baseQuery.Where("province = ?", *query.Province)
	}

	if query.ManagerID != nil {
		baseQuery = baseQuery.Where("manager_id = ?", *query.ManagerID)
	}

	if query.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *query.IsActive)
	} else {
		// Default: only show active warehouses
		baseQuery = baseQuery.Where("is_active = ?", true)
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count warehouses: %w", err)
	}

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	var warehouses []models.Warehouse
	if err := baseQuery.Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Find(&warehouses).Error; err != nil {
		return nil, fmt.Errorf("failed to list warehouses: %w", err)
	}

	// Map to response DTOs
	warehouseResponses := make([]dto.WarehouseResponse, len(warehouses))
	for i, warehouse := range warehouses {
		warehouseResponses[i] = mapWarehouseToResponse(&warehouse)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.WarehouseListResponse{
		Success: true,
		Data:    warehouseResponses,
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      pageSize,
			Total:      int(totalCount),
			TotalPages: totalPages,
		},
	}, nil
}

// ============================================================================
// GET WAREHOUSE BY ID
// ============================================================================

// GetWarehouseByID retrieves a warehouse by ID
func (s *WarehouseService) GetWarehouseByID(ctx context.Context, tenantID, companyID, warehouseID string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ? AND id = ?", companyID, warehouseID).
		First(&warehouse).Error

	if err == gorm.ErrRecordNotFound {
		return nil, pkgerrors.NewNotFoundError("warehouse not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	return &warehouse, nil
}

// ============================================================================
// UPDATE WAREHOUSE
// ============================================================================

// UpdateWarehouse updates an existing warehouse
func (s *WarehouseService) UpdateWarehouse(ctx context.Context, tenantID, companyID, warehouseID, userID, ipAddress, userAgent string, req *dto.UpdateWarehouseRequest) (*models.Warehouse, error) {
	// Get existing warehouse
	warehouse, err := s.GetWarehouseByID(ctx, tenantID, companyID, warehouseID)
	if err != nil {
		return nil, err
	}

	// Capture old values for audit logging
	oldValues := map[string]interface{}{
		"code":        warehouse.Code,
		"name":        warehouse.Name,
		"type":        string(warehouse.Type),
		"address":     warehouse.Address,
		"city":        warehouse.City,
		"province":    warehouse.Province,
		"postal_code": warehouse.PostalCode,
		"phone":       warehouse.Phone,
		"email":       warehouse.Email,
		"manager_id":  warehouse.ManagerID,
		"is_active":   warehouse.IsActive,
	}
	if warehouse.Capacity != nil {
		oldValues["capacity"] = warehouse.Capacity.String()
	}

	// Validate code uniqueness if updating code
	if req.Code != nil && *req.Code != warehouse.Code {
		if err := s.validateCodeUniqueness(ctx, tenantID, companyID, *req.Code, warehouseID); err != nil {
			return nil, err
		}
		warehouse.Code = *req.Code
	}

	// Validate manager exists if updating manager
	if req.ManagerID != nil && *req.ManagerID != "" && (warehouse.ManagerID == nil || *req.ManagerID != *warehouse.ManagerID) {
		if err := s.validateManagerExists(*req.ManagerID); err != nil {
			return nil, err
		}
		warehouse.ManagerID = req.ManagerID
	}

	// Update fields
	if req.Name != nil {
		warehouse.Name = *req.Name
	}

	if req.Type != nil {
		warehouse.Type = models.WarehouseType(*req.Type)
	}

	if req.Address != nil {
		warehouse.Address = req.Address
	}

	if req.City != nil {
		warehouse.City = req.City
	}

	if req.Province != nil {
		warehouse.Province = req.Province
	}

	if req.PostalCode != nil {
		warehouse.PostalCode = req.PostalCode
	}

	if req.Phone != nil {
		warehouse.Phone = req.Phone
	}

	if req.Email != nil {
		warehouse.Email = req.Email
	}

	if req.Capacity != nil && *req.Capacity != "" {
		capacity, err := decimal.NewFromString(*req.Capacity)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid capacity format")
		}
		if capacity.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("capacity cannot be negative")
		}
		warehouse.Capacity = &capacity
	}

	if req.IsActive != nil {
		warehouse.IsActive = *req.IsActive
	}

	// Save updates
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(warehouse).Error; err != nil {
		return nil, fmt.Errorf("failed to update warehouse: %w", err)
	}

	// Audit logging - Log successful warehouse update
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	newValues := map[string]interface{}{
		"code":        warehouse.Code,
		"name":        warehouse.Name,
		"type":        string(warehouse.Type),
		"address":     warehouse.Address,
		"city":        warehouse.City,
		"province":    warehouse.Province,
		"postal_code": warehouse.PostalCode,
		"phone":       warehouse.Phone,
		"email":       warehouse.Email,
		"manager_id":  warehouse.ManagerID,
		"is_active":   warehouse.IsActive,
	}
	if warehouse.Capacity != nil {
		newValues["capacity"] = warehouse.Capacity.String()
	}

	if err := s.auditService.LogWarehouseUpdated(ctx, auditCtx, warehouse.ID, oldValues, newValues); err != nil {
		fmt.Printf("WARNING: Failed to create audit log: %v\n", err)
	}

	return warehouse, nil
}

// ============================================================================
// DELETE WAREHOUSE (SOFT DELETE)
// ============================================================================

// DeleteWarehouse soft deletes a warehouse
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *WarehouseService) DeleteWarehouse(ctx context.Context, tenantID, companyID, warehouseID, userID, ipAddress, userAgent string) error {
	// Get warehouse
	warehouse, err := s.GetWarehouseByID(ctx, tenantID, companyID, warehouseID)
	if err != nil {
		return err
	}

	// Capture warehouse data for audit logging before deletion
	warehouseData := map[string]interface{}{
		"code":        warehouse.Code,
		"name":        warehouse.Name,
		"type":        string(warehouse.Type),
		"address":     warehouse.Address,
		"city":        warehouse.City,
		"province":    warehouse.Province,
		"postal_code": warehouse.PostalCode,
		"phone":       warehouse.Phone,
		"email":       warehouse.Email,
		"manager_id":  warehouse.ManagerID,
		"is_active":   warehouse.IsActive,
	}
	if warehouse.Capacity != nil {
		warehouseData["capacity"] = warehouse.Capacity.String()
	}

	// Validate deletion
	if err := s.validateDeleteWarehouse(ctx, tenantID, warehouse); err != nil {
		return err
	}

	// Soft delete (set IsActive = false)
	warehouse.IsActive = false
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(warehouse).Error; err != nil {
		return fmt.Errorf("failed to delete warehouse: %w", err)
	}

	// Audit logging - Log successful warehouse deletion
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	if err := s.auditService.LogWarehouseDeleted(ctx, auditCtx, warehouse.ID, warehouseData); err != nil {
		fmt.Printf("WARNING: Failed to create audit log: %v\n", err)
	}

	return nil
}

// ============================================================================
// WAREHOUSE STOCK MANAGEMENT
// ============================================================================

// ListWarehouseStocks retrieves warehouse stocks with filtering and pagination
func (s *WarehouseService) ListWarehouseStocks(ctx context.Context, tenantID, companyID string, query *dto.WarehouseStockListQuery) (*dto.WarehouseStockListResponse, error) {
	// Set defaults
	page := 1
	if query.Page > 0 {
		page = query.Page
	}

	pageSize := 20
	if query.PageSize > 0 {
		pageSize = query.PageSize
	}

	sortBy := "products.code"
	if query.SortBy == "productCode" {
		sortBy = "products.code"
	} else if query.SortBy == "productName" {
		sortBy = "products.name"
	} else if query.SortBy == "quantity" {
		sortBy = "warehouse_stocks.quantity"
	} else if query.SortBy == "createdAt" {
		sortBy = "warehouse_stocks.created_at"
	}

	sortOrder := "asc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}

	// Build base query
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.WarehouseStock{}).
		Joins("JOIN warehouses ON warehouses.id = warehouse_stocks.warehouse_id").
		Joins("JOIN products ON products.id = warehouse_stocks.product_id").
		Where("warehouses.company_id = ? AND products.company_id = ?", companyID, companyID)

	// Apply filters
	if query.WarehouseID != nil {
		baseQuery = baseQuery.Where("warehouse_stocks.warehouse_id = ?", *query.WarehouseID)
	}

	if query.ProductID != nil {
		baseQuery = baseQuery.Where("warehouse_stocks.product_id = ?", *query.ProductID)
	}

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		baseQuery = baseQuery.Where("products.code LIKE ? OR products.name LIKE ?", searchPattern, searchPattern)
	}

	if query.LowStock != nil && *query.LowStock {
		baseQuery = baseQuery.Where("warehouse_stocks.quantity < warehouse_stocks.minimum_stock AND warehouse_stocks.quantity > 0")
	}

	if query.ZeroStock != nil && *query.ZeroStock {
		baseQuery = baseQuery.Where("warehouse_stocks.quantity = 0")
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count warehouse stocks: %w", err)
	}

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	var stocks []models.WarehouseStock
	if err := baseQuery.Order(orderClause).
		Preload("Product").
		Limit(pageSize).
		Offset(offset).
		Find(&stocks).Error; err != nil {
		return nil, fmt.Errorf("failed to list warehouse stocks: %w", err)
	}

	// Map to response DTOs
	stockResponses := make([]dto.WarehouseStockResponse, len(stocks))
	for i, stock := range stocks {
		stockResponses[i] = mapWarehouseStockToResponse(&stock)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.WarehouseStockListResponse{
		Stocks:     stockResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateWarehouseStock updates warehouse stock settings (not quantity)
func (s *WarehouseService) UpdateWarehouseStock(ctx context.Context, tenantID, companyID, stockID string, req *dto.UpdateWarehouseStockRequest) (*models.WarehouseStock, error) {
	// Get existing warehouse stock
	var stock models.WarehouseStock
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Joins("JOIN warehouses ON warehouses.id = warehouse_stocks.warehouse_id").
		Where("warehouses.company_id = ? AND warehouse_stocks.id = ?", companyID, stockID).
		First(&stock).Error

	if err == gorm.ErrRecordNotFound {
		return nil, pkgerrors.NewNotFoundError("warehouse stock not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse stock: %w", err)
	}

	// Update fields
	if req.MinimumStock != nil && *req.MinimumStock != "" {
		minimumStock, err := decimal.NewFromString(*req.MinimumStock)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid minimumStock format")
		}
		if minimumStock.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("minimum stock cannot be negative")
		}
		stock.MinimumStock = minimumStock
	}

	if req.MaximumStock != nil && *req.MaximumStock != "" {
		maximumStock, err := decimal.NewFromString(*req.MaximumStock)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid maximumStock format")
		}
		if maximumStock.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("maximum stock cannot be negative")
		}
		stock.MaximumStock = maximumStock
	}

	if req.Location != nil {
		stock.Location = req.Location
	}

	// Save updates
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(&stock).Error; err != nil {
		return nil, fmt.Errorf("failed to update warehouse stock: %w", err)
	}

	// Reload with Product relation
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Preload("Product").Where("id = ?", stock.ID).First(&stock).Error; err != nil {
		return nil, fmt.Errorf("failed to reload warehouse stock: %w", err)
	}

	return &stock, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapWarehouseToResponse converts Warehouse model to WarehouseResponse DTO
func mapWarehouseToResponse(warehouse *models.Warehouse) dto.WarehouseResponse {
	var capacity *string
	if warehouse.Capacity != nil {
		cap := warehouse.Capacity.String()
		capacity = &cap
	}

	return dto.WarehouseResponse{
		ID:         warehouse.ID,
		Code:       warehouse.Code,
		Name:       warehouse.Name,
		Type:       string(warehouse.Type),
		Address:    warehouse.Address,
		City:       warehouse.City,
		Province:   warehouse.Province,
		PostalCode: warehouse.PostalCode,
		Phone:      warehouse.Phone,
		Email:      warehouse.Email,
		ManagerID:  warehouse.ManagerID,
		Capacity:   capacity,
		IsActive:   warehouse.IsActive,
		CreatedAt:  warehouse.CreatedAt,
		UpdatedAt:  warehouse.UpdatedAt,
	}
}

// mapWarehouseStockToResponse converts WarehouseStock model to WarehouseStockResponse DTO
func mapWarehouseStockToResponse(stock *models.WarehouseStock) dto.WarehouseStockResponse {
	var lastCountQty *string
	if stock.LastCountQty != nil {
		qty := stock.LastCountQty.String()
		lastCountQty = &qty
	}

	productCode := ""
	productName := ""
	if stock.Product.ID != "" {
		productCode = stock.Product.Code
		productName = stock.Product.Name
	}

	return dto.WarehouseStockResponse{
		ID:            stock.ID,
		WarehouseID:   stock.WarehouseID,
		ProductID:     stock.ProductID,
		ProductCode:   productCode,
		ProductName:   productName,
		Quantity:      stock.Quantity.String(),
		MinimumStock:  stock.MinimumStock.String(),
		MaximumStock:  stock.MaximumStock.String(),
		Location:      stock.Location,
		LastCountDate: stock.LastCountDate,
		LastCountQty:  lastCountQty,
		CreatedAt:     stock.CreatedAt,
		UpdatedAt:     stock.UpdatedAt,
	}
}
