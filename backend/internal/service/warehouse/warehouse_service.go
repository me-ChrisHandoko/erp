package warehouse

import (
	"context"
	"fmt"
	"math"
	"time"

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
		Preload("Warehouse").
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
func (s *WarehouseService) UpdateWarehouseStock(ctx context.Context, tenantID, companyID, stockID, userID, ipAddress, userAgent string, req *dto.UpdateWarehouseStockRequest) (*models.WarehouseStock, error) {
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

	// Capture old values for audit logging
	oldValues := map[string]interface{}{
		"minimum_stock": stock.MinimumStock.String(),
		"maximum_stock": stock.MaximumStock.String(),
		"location":      stock.Location,
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

	// Audit logging - Log successful warehouse stock update
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
		"minimum_stock": stock.MinimumStock.String(),
		"maximum_stock": stock.MaximumStock.String(),
		"location":      stock.Location,
	}

	if err := s.auditService.LogWarehouseStockUpdated(ctx, auditCtx, stock.ID, oldValues, newValues); err != nil {
		fmt.Printf("WARNING: Failed to create audit log for warehouse stock: %v\n", err)
	}

	// Reload with Product relation
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Preload("Product").Where("id = ?", stock.ID).First(&stock).Error; err != nil {
		return nil, fmt.Errorf("failed to reload warehouse stock: %w", err)
	}

	return &stock, nil
}

// ============================================================================
// WAREHOUSE STOCK STATUS
// ============================================================================

// GetWarehouseStockStatus retrieves stock initialization status for all warehouses in a company
// Returns information about whether warehouses have initial stock setup, total products, and total value
func (s *WarehouseService) GetWarehouseStockStatus(ctx context.Context, tenantID, companyID string) ([]dto.WarehouseStockStatusResponse, error) {
	// Query to get warehouse stock status
	// Uses LEFT JOIN to include warehouses without stocks
	// Aggregates stock information per warehouse
	type queryResult struct {
		WarehouseID   string
		WarehouseName string
		WarehouseCode string
		TotalProducts int
		TotalValue    string
		LastUpdated   *time.Time
	}

	var results []queryResult

	err := s.db.WithContext(ctx).
		Set("tenant_id", tenantID).
		Table("warehouses").
		Select(`
			warehouses.id as warehouse_id,
			warehouses.name as warehouse_name,
			warehouses.code as warehouse_code,
			COUNT(DISTINCT warehouse_stocks.product_id) as total_products,
			COALESCE(SUM(warehouse_stocks.quantity * 0), '0') as total_value,
			MAX(warehouse_stocks.updated_at) as last_updated
		`).
		Joins("LEFT JOIN warehouse_stocks ON warehouse_stocks.warehouse_id = warehouses.id").
		Where("warehouses.company_id = ? AND warehouses.is_active = ?", companyID, true).
		Group("warehouses.id, warehouses.name, warehouses.code").
		Order("warehouses.name ASC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query warehouse stock status: %w", err)
	}

	// Map query results to DTO response
	statusList := make([]dto.WarehouseStockStatusResponse, 0, len(results))
	for _, result := range results {
		statusList = append(statusList, dto.WarehouseStockStatusResponse{
			WarehouseID:     result.WarehouseID,
			WarehouseName:   result.WarehouseName,
			WarehouseCode:   result.WarehouseCode,
			HasInitialStock: result.TotalProducts > 0,
			TotalProducts:   result.TotalProducts,
			TotalValue:      result.TotalValue,
			LastUpdated:     result.LastUpdated,
		})
	}

	return statusList, nil
}

// ============================================================================
// INITIAL STOCK SETUP
// ============================================================================

// CreateInitialStock creates initial warehouse stocks for a warehouse
// This is used for one-time initial stock setup when migrating data or setting up a new warehouse
func (s *WarehouseService) CreateInitialStock(
	ctx context.Context,
	tenantID, companyID, userID, ipAddress, userAgent string,
	req *dto.InitialStockSetupRequest,
) (*dto.InitialStockSetupResponse, error) {
	// Create audit context early for failure logging
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	// Helper function for failure logging
	logFailure := func(errorMsg string) {
		reqData := map[string]interface{}{
			"warehouse_id": req.WarehouseID,
			"total_items":  len(req.Items),
		}
		if err := s.auditService.LogInitialStockOperationFailed(ctx, auditCtx, req.WarehouseID, errorMsg, reqData); err != nil {
			fmt.Printf("WARNING: Failed to create failure audit log: %v\n", err)
		}
	}

	// Validate warehouse exists and belongs to company
	warehouse, err := s.GetWarehouseByID(ctx, tenantID, companyID, req.WarehouseID)
	if err != nil {
		logFailure(fmt.Sprintf("Warehouse validation failed: %v", err))
		return nil, err
	}

	if !warehouse.IsActive {
		logFailure("Warehouse is not active")
		return nil, pkgerrors.NewBadRequestError("Warehouse is not active")
	}

	// Get all product IDs from request
	productIDs := make([]string, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	// Validate all products exist and belong to company
	var products []models.Product
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ? AND id IN ?", companyID, productIDs).
		Find(&products).Error; err != nil {
		logFailure(fmt.Sprintf("Failed to fetch products: %v", err))
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	productMap := make(map[string]*models.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	// Validate all products were found
	for _, item := range req.Items {
		if _, exists := productMap[item.ProductID]; !exists {
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Product with ID %s not found", item.ProductID))
		}
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	if tx.Error != nil {
		logFailure(fmt.Sprintf("Failed to start transaction: %v", tx.Error))
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	createdCount := 0
	updatedCount := 0
	totalValue := decimal.Zero

	// Collect item details for audit logging
	type itemAuditData struct {
		ProductID    string  `json:"product_id"`
		ProductCode  string  `json:"product_code"`
		ProductName  string  `json:"product_name"`
		Quantity     string  `json:"quantity"`
		CostPerUnit  string  `json:"cost_per_unit"`
		Value        string  `json:"value"`
		Location     *string `json:"location,omitempty"`
		MinimumStock string  `json:"minimum_stock"`
		MaximumStock string  `json:"maximum_stock"`
		Notes        *string `json:"notes,omitempty"`
		Action       string  `json:"action"`
		StockBefore  string  `json:"stock_before"`
		StockAfter   string  `json:"stock_after"`
	}
	itemsAuditData := make([]itemAuditData, 0, len(req.Items))

	for _, item := range req.Items {
		// Parse quantity
		qty, err := decimal.NewFromString(item.Quantity)
		if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Invalid quantity for product %s", item.ProductID))
		}
		if qty.LessThanOrEqual(decimal.Zero) {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Quantity must be positive for product %s", item.ProductID))
		}

		// Parse cost per unit
		cost, err := decimal.NewFromString(item.CostPerUnit)
		if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Invalid cost for product %s", item.ProductID))
		}
		if cost.LessThanOrEqual(decimal.Zero) {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Cost must be positive for product %s", item.ProductID))
		}

		// Parse optional fields
		var minStock, maxStock decimal.Decimal
		if item.MinimumStock != nil && *item.MinimumStock != "" {
			minStock, err = decimal.NewFromString(*item.MinimumStock)
			if err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Invalid minimum stock for product %s", item.ProductID))
			}
		}
		if item.MaximumStock != nil && *item.MaximumStock != "" {
			maxStock, err = decimal.NewFromString(*item.MaximumStock)
			if err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Invalid maximum stock for product %s", item.ProductID))
			}
		}

		// Check if warehouse stock already exists
		var existingStock models.WarehouseStock
		err = tx.Where("warehouse_id = ? AND product_id = ?", req.WarehouseID, item.ProductID).
			First(&existingStock).Error

		if err == gorm.ErrRecordNotFound {
			// Create new warehouse stock
			newStock := &models.WarehouseStock{
				WarehouseID:  req.WarehouseID,
				ProductID:    item.ProductID,
				Quantity:     qty,
				MinimumStock: minStock,
				MaximumStock: maxStock,
				Location:     item.Location,
			}

			if err := tx.Create(newStock).Error; err != nil {
				tx.Rollback()
				logFailure(fmt.Sprintf("Failed to create warehouse stock for product %s: %v", item.ProductID, err))
				return nil, fmt.Errorf("failed to create warehouse stock: %w", err)
			}

			// Create inventory movement
			refType := "INITIAL_STOCK"
			movement := &models.InventoryMovement{
				TenantID:      tenantID,
				CompanyID:     companyID,
				MovementDate:  time.Now(),
				WarehouseID:   req.WarehouseID,
				ProductID:     item.ProductID,
				MovementType:  models.MovementTypeInitial,
				Quantity:      qty,
				StockBefore:   decimal.Zero,
				StockAfter:    qty,
				ReferenceType: &refType,
				ReferenceID:   &newStock.ID,
				Notes:         req.Notes,
				CreatedBy:     &userID,
			}

			if err := tx.Create(movement).Error; err != nil {
				tx.Rollback()
				logFailure(fmt.Sprintf("Failed to create inventory movement for product %s: %v", item.ProductID, err))
				return nil, fmt.Errorf("failed to create inventory movement: %w", err)
			}

			// Collect audit data for created item
			product := productMap[item.ProductID]
			itemValue := qty.Mul(cost)
			itemsAuditData = append(itemsAuditData, itemAuditData{
				ProductID:    item.ProductID,
				ProductCode:  product.Code,
				ProductName:  product.Name,
				Quantity:     qty.String(),
				CostPerUnit:  cost.String(),
				Value:        itemValue.String(),
				Location:     item.Location,
				MinimumStock: minStock.String(),
				MaximumStock: maxStock.String(),
				Notes:        item.Notes,
				Action:       "created",
				StockBefore:  "0",
				StockAfter:   qty.String(),
			})

			createdCount++
		} else if err != nil {
			tx.Rollback()
			logFailure(fmt.Sprintf("Failed to check existing stock for product %s: %v", item.ProductID, err))
			return nil, fmt.Errorf("failed to check existing stock: %w", err)
		} else {
			// Update existing warehouse stock
			stockBefore := existingStock.Quantity
			existingStock.Quantity = existingStock.Quantity.Add(qty)
			if item.MinimumStock != nil && *item.MinimumStock != "" {
				existingStock.MinimumStock = minStock
			}
			if item.MaximumStock != nil && *item.MaximumStock != "" {
				existingStock.MaximumStock = maxStock
			}
			if item.Location != nil {
				existingStock.Location = item.Location
			}

			if err := tx.Save(&existingStock).Error; err != nil {
				tx.Rollback()
				logFailure(fmt.Sprintf("Failed to update warehouse stock for product %s: %v", item.ProductID, err))
				return nil, fmt.Errorf("failed to update warehouse stock: %w", err)
			}

			// Create inventory movement
			refType := "INITIAL_STOCK"
			movement := &models.InventoryMovement{
				TenantID:      tenantID,
				CompanyID:     companyID,
				MovementDate:  time.Now(),
				WarehouseID:   req.WarehouseID,
				ProductID:     item.ProductID,
				MovementType:  models.MovementTypeInitial,
				Quantity:      qty,
				StockBefore:   stockBefore,
				StockAfter:    existingStock.Quantity,
				ReferenceType: &refType,
				ReferenceID:   &existingStock.ID,
				Notes:         req.Notes,
				CreatedBy:     &userID,
			}

			if err := tx.Create(movement).Error; err != nil {
				tx.Rollback()
				logFailure(fmt.Sprintf("Failed to create inventory movement for product %s: %v", item.ProductID, err))
				return nil, fmt.Errorf("failed to create inventory movement: %w", err)
			}

			// Collect audit data for updated item
			product := productMap[item.ProductID]
			itemValue := qty.Mul(cost)
			itemsAuditData = append(itemsAuditData, itemAuditData{
				ProductID:    item.ProductID,
				ProductCode:  product.Code,
				ProductName:  product.Name,
				Quantity:     qty.String(),
				CostPerUnit:  cost.String(),
				Value:        itemValue.String(),
				Location:     item.Location,
				MinimumStock: minStock.String(),
				MaximumStock: maxStock.String(),
				Notes:        item.Notes,
				Action:       "updated",
				StockBefore:  stockBefore.String(),
				StockAfter:   existingStock.Quantity.String(),
			})

			updatedCount++
		}

		// Calculate total value
		itemValue := qty.Mul(cost)
		totalValue = totalValue.Add(itemValue)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logFailure(fmt.Sprintf("Transaction commit failed: %v", err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Audit logging for success - struct ensures JSON field ordering
	type initialStockAuditData struct {
		WarehouseID   string          `json:"warehouse_id"`
		WarehouseName string          `json:"warehouse_name"`
		WarehouseCode string          `json:"warehouse_code"`
		UpdatedStocks int             `json:"updated_stocks"`
		TotalItems    int             `json:"total_items"`
		TotalValue    string          `json:"total_value"`
		CreatedStocks int             `json:"created_stocks"`
		Notes         *string         `json:"notes,omitempty"`
		Items         []itemAuditData `json:"items"`
	}

	auditData := initialStockAuditData{
		WarehouseID:   req.WarehouseID,
		WarehouseName: warehouse.Name,
		WarehouseCode: warehouse.Code,
		UpdatedStocks: updatedCount,
		TotalItems:    len(req.Items),
		TotalValue:    totalValue.String(),
		CreatedStocks: createdCount,
		Notes:         req.Notes,
		Items:         itemsAuditData,
	}

	if err := s.auditService.LogInitialStockCreated(ctx, auditCtx, req.WarehouseID, auditData, len(req.Items), createdCount, updatedCount); err != nil {
		fmt.Printf("WARNING: Failed to create audit log for initial stock: %v\n", err)
	}

	return &dto.InitialStockSetupResponse{
		Success:       true,
		Message:       fmt.Sprintf("Successfully created initial stock for %d items", len(req.Items)),
		TotalItems:    len(req.Items),
		TotalValue:    totalValue.String(),
		CreatedStocks: createdCount,
		UpdatedStocks: updatedCount,
	}, nil
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

	warehouseName := ""
	warehouseCode := ""
	if stock.Warehouse.ID != "" {
		warehouseName = stock.Warehouse.Name
		warehouseCode = stock.Warehouse.Code
	}

	return dto.WarehouseStockResponse{
		ID:            stock.ID,
		WarehouseID:   stock.WarehouseID,
		WarehouseName: warehouseName,
		WarehouseCode: warehouseCode,
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
