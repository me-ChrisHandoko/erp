package product

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

type ProductService struct {
	db           *gorm.DB
	auditService *audit.AuditService
}

func NewProductService(db *gorm.DB, auditService *audit.AuditService) *ProductService {
	return &ProductService{
		db:           db,
		auditService: auditService,
	}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

// CreateProduct creates a new product with multi-unit support and warehouse stock initialization
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 365-456 (Business Logic)
// CRITICAL: Creates base unit entry and initializes warehouse stocks in transaction
func (s *ProductService) CreateProduct(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, req *dto.CreateProductRequest) (*models.Product, error) {
	// Parse decimal fields
	baseCost, err := decimal.NewFromString(req.BaseCost)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid baseCost format")
	}

	basePrice, err := decimal.NewFromString(req.BasePrice)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid basePrice format")
	}

	minimumStock := decimal.Zero
	if req.MinimumStock != "" {
		minimumStock, err = decimal.NewFromString(req.MinimumStock)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid minimumStock format")
		}
	}

	// Validate request
	if err := s.validateCreateProduct(ctx, companyID, tenantID, req, baseCost, basePrice, minimumStock); err != nil {
		return nil, err
	}

	var product *models.Product

	// Use transaction for atomic create
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Create product
		product = &models.Product{
			CompanyID:      companyID,
			TenantID:       tenantID,
			Code:           req.Code,
			Name:           req.Name,
			Category:       req.Category,
			BaseUnit:       req.BaseUnit,
			BaseCost:       baseCost,
			BasePrice:      basePrice,
			CurrentStock:   decimal.Zero, // DEPRECATED: Never update this field
			MinimumStock:   minimumStock,
			Description:    req.Description,
			Barcode:        req.Barcode,
			IsBatchTracked: req.IsBatchTracked,
			IsPerishable:   req.IsPerishable,
			IsActive:       true,
		}

		if err := tx.Create(product).Error; err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		// 2. Create base unit entry (CRITICAL - Issue #4 fix)
		baseUnit := &models.ProductUnit{
			ProductID:      product.ID,
			UnitName:       req.BaseUnit,
			ConversionRate: decimal.NewFromInt(1), // Base unit conversion = 1
			IsBaseUnit:     true,
			BuyPrice:       &baseCost,
			SellPrice:      &basePrice,
			IsActive:       true,
		}

		if err := tx.Create(baseUnit).Error; err != nil {
			return fmt.Errorf("failed to create base unit: %w", err)
		}

		// 3. Create additional units
		for _, unitReq := range req.Units {
			conversionRate, err := decimal.NewFromString(unitReq.ConversionRate)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid conversionRate for unit %s", unitReq.UnitName))
			}

			unit := &models.ProductUnit{
				ProductID:      product.ID,
				UnitName:       unitReq.UnitName,
				ConversionRate: conversionRate,
				IsBaseUnit:     false,
				Barcode:        unitReq.Barcode,
				SKU:            unitReq.SKU,
				Description:    unitReq.Description,
				IsActive:       true,
			}

			// Parse optional decimal fields
			if unitReq.BuyPrice != nil {
				buyPrice, err := decimal.NewFromString(*unitReq.BuyPrice)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid buyPrice for unit %s", unitReq.UnitName))
				}
				unit.BuyPrice = &buyPrice
			}

			if unitReq.SellPrice != nil {
				sellPrice, err := decimal.NewFromString(*unitReq.SellPrice)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid sellPrice for unit %s", unitReq.UnitName))
				}
				unit.SellPrice = &sellPrice
			}

			if unitReq.Weight != nil {
				weight, err := decimal.NewFromString(*unitReq.Weight)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid weight for unit %s", unitReq.UnitName))
				}
				unit.Weight = &weight
			}

			if unitReq.Volume != nil {
				volume, err := decimal.NewFromString(*unitReq.Volume)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid volume for unit %s", unitReq.UnitName))
				}
				unit.Volume = &volume
			}

			if err := tx.Create(unit).Error; err != nil {
				return fmt.Errorf("failed to create product unit: %w", err)
			}
		}

		// 4. Initialize warehouse stocks (zero stock for all active warehouses)
		var warehouses []models.Warehouse
		if err := tx.Where("company_id = ? AND is_active = ?", companyID, true).Find(&warehouses).Error; err != nil {
			return fmt.Errorf("failed to get warehouses: %w", err)
		}

		for _, wh := range warehouses {
			whStock := &models.WarehouseStock{
				WarehouseID:  wh.ID,
				ProductID:    product.ID,
				Quantity:     decimal.Zero,
				MinimumStock: minimumStock,
			}

			if err := tx.Create(whStock).Error; err != nil {
				return fmt.Errorf("failed to initialize warehouse stock: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		// Audit: Log failed operation
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}
		if auditErr := s.auditService.LogProductOperationFailed(ctx, auditCtx, "PRODUCT_CREATED", "", err.Error()); auditErr != nil {
			fmt.Printf("WARNING: Failed to log failed product creation: %v\n", auditErr)
		}
		return nil, err
	}

	// Audit: Log successful creation
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	productData := map[string]interface{}{
		"code":             product.Code,
		"name":             product.Name,
		"description":      stringPtrToValue(product.Description),
		"barcode":          stringPtrToValue(product.Barcode),
		"category":         stringPtrToValue(product.Category),
		"base_unit":        product.BaseUnit,
		"base_cost":        product.BaseCost.String(),
		"base_price":       product.BasePrice.String(),
		"minimum_stock":    product.MinimumStock.String(),
		"is_batch_tracked": product.IsBatchTracked,
		"is_perishable":    product.IsPerishable,
	}

	if err := s.auditService.LogProductCreated(ctx, auditCtx, product.ID, productData); err != nil {
		// Log error but don't fail the create operation
		fmt.Printf("WARNING: Failed to create audit log for product creation: %v\n", err)
	}

	// Reload product with relations
	return s.GetProduct(ctx, companyID, tenantID, product.ID)
}

// GetProduct retrieves a product by ID with all relations
func (s *ProductService) GetProduct(ctx context.Context, companyID, tenantID, productID string) (*models.Product, error) {
	var product models.Product

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Units", "is_active = ?", true).
		Preload("ProductSuppliers.Supplier").
		Preload("WarehouseStocks.Warehouse", "is_active = ?", true).
		Where("company_id = ? AND id = ?", companyID, productID).
		First(&product).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// ListProducts retrieves paginated products with filters
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 117-179 (List Products API)
// tenantID is passed from handler (from middleware context) to enable GORM tenant isolation
func (s *ProductService) ListProducts(ctx context.Context, tenantID, companyID string, filters *dto.ProductFilters) ([]models.Product, int64, error) {
	// Set defaults
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.SortBy == "" {
		filters.SortBy = "created_at"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "desc"
	}

	// Build query with tenant context set for GORM callbacks
	// tenantID is explicitly passed from handler (no bypass needed!)
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.Product{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if filters.Search != "" {
		query = query.Where(
			"code ILIKE ? OR name ILIKE ? OR barcode ILIKE ?",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
			"%"+filters.Search+"%",
		)
	}

	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.IsBatchTracked != nil {
		query = query.Where("is_batch_tracked = ?", *filters.IsBatchTracked)
	}

	if filters.IsPerishable != nil {
		query = query.Where("is_perishable = ?", *filters.IsPerishable)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get paginated results
	var products []models.Product
	offset := (filters.Page - 1) * filters.Limit

	err := query.
		Preload("Units", "is_active = ?", true).
		Preload("WarehouseStocks.Warehouse", "is_active = ?", true).
		Order(fmt.Sprintf("%s %s", filters.SortBy, filters.SortOrder)).
		Limit(filters.Limit).
		Offset(offset).
		Find(&products).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(ctx context.Context, companyID, tenantID, productID string, userID string, ipAddress string, userAgent string, req *dto.UpdateProductRequest) (*models.Product, error) {
	// Get existing product
	product, err := s.GetProduct(ctx, companyID, tenantID, productID)
	if err != nil {
		return nil, err
	}

	// Validate updates
	if err := s.validateUpdateProduct(ctx, companyID, product.TenantID, productID, req); err != nil {
		return nil, err
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Category != nil {
		updates["category"] = *req.Category
	}

	if req.Code != nil {
		updates["code"] = *req.Code
	}

	if req.BaseUnit != nil {
		updates["base_unit"] = *req.BaseUnit
	}

	if req.BaseCost != nil {
		baseCost, err := decimal.NewFromString(*req.BaseCost)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid baseCost format")
		}
		updates["base_cost"] = baseCost
	}

	if req.BasePrice != nil {
		basePrice, err := decimal.NewFromString(*req.BasePrice)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid basePrice format")
		}
		updates["base_price"] = basePrice
	}

	if req.MinimumStock != nil {
		minimumStock, err := decimal.NewFromString(*req.MinimumStock)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid minimumStock format")
		}
		updates["minimum_stock"] = minimumStock
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Barcode != nil {
		updates["barcode"] = *req.Barcode
	}

	if req.IsBatchTracked != nil {
		updates["is_batch_tracked"] = *req.IsBatchTracked
	}

	if req.IsPerishable != nil {
		updates["is_perishable"] = *req.IsPerishable
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// Capture old values for audit
	oldValues := map[string]interface{}{
		"code":             product.Code,
		"name":             product.Name,
		"description":      stringPtrToValue(product.Description),
		"barcode":          stringPtrToValue(product.Barcode),
		"category":         stringPtrToValue(product.Category),
		"base_unit":        product.BaseUnit,
		"base_cost":        product.BaseCost.String(),
		"base_price":       product.BasePrice.String(),
		"minimum_stock":    product.MinimumStock.String(),
		"is_batch_tracked": product.IsBatchTracked,
		"is_perishable":    product.IsPerishable,
		"is_active":        product.IsActive,
	}

	// Update product directly without preloaded associations
	// Don't use the preloaded product object to avoid association auto-save issues
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Model(&models.Product{}).
		Where("id = ? AND company_id = ?", productID, companyID).
		Updates(updates).Error; err != nil {

		// Audit: Log failed operation
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}
		if auditErr := s.auditService.LogProductOperationFailed(ctx, auditCtx, "PRODUCT_UPDATED", productID, err.Error()); auditErr != nil {
			fmt.Printf("WARNING: Failed to log failed product update: %v\n", auditErr)
		}

		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Reload product
	updatedProduct, err := s.GetProduct(ctx, companyID, tenantID, productID)
	if err != nil {
		return nil, err
	}

	// 🔍 DEBUG: Verify this code is running
	fmt.Println("🔍 DEBUG: UpdateProduct - About to create audit log...")
	fmt.Printf("🔍 DEBUG: Product ID: %s, TenantID: %s, CompanyID: %s\n", productID, tenantID, companyID)

	// Audit: Log successful update
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
		"code":             updatedProduct.Code,
		"name":             updatedProduct.Name,
		"description":      stringPtrToValue(updatedProduct.Description),
		"barcode":          stringPtrToValue(updatedProduct.Barcode),
		"category":         stringPtrToValue(updatedProduct.Category),
		"base_unit":        updatedProduct.BaseUnit,
		"base_cost":        updatedProduct.BaseCost.String(),
		"base_price":       updatedProduct.BasePrice.String(),
		"minimum_stock":    updatedProduct.MinimumStock.String(),
		"is_batch_tracked": updatedProduct.IsBatchTracked,
		"is_perishable":    updatedProduct.IsPerishable,
		"is_active":        updatedProduct.IsActive,
	}

	// 🔍 DEBUG: Check if auditService is nil
	if s.auditService == nil {
		fmt.Println("❌ ERROR: auditService is NIL!")
	} else {
		fmt.Println("✅ DEBUG: auditService is NOT nil, calling LogProductUpdated...")
	}

	if err := s.auditService.LogProductUpdated(ctx, auditCtx, productID, oldValues, newValues); err != nil {
		// Log error but don't fail the update operation
		fmt.Printf("⚠️ WARNING: Failed to create audit log for product update: %v\n", err)
	} else {
		fmt.Println("✅ SUCCESS: Audit log created successfully!")
	}

	return updatedProduct, nil
}

// DeleteProduct soft deletes a product (sets is_active = false)
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 328-343 (Business Logic)
func (s *ProductService) DeleteProduct(ctx context.Context, companyID, productID string, userID string, ipAddress string, userAgent string) error {
	// Get product
	product, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return err
	}

	// Validate deletion
	if err := s.validateDeleteProduct(ctx, product); err != nil {
		return err
	}

	// Capture product data for audit
	productData := map[string]interface{}{
		"code":             product.Code,
		"name":             product.Name,
		"category":         product.Category,
		"base_unit":        product.BaseUnit,
		"base_cost":        product.BaseCost.String(),
		"base_price":       product.BasePrice.String(),
		"is_batch_tracked": product.IsBatchTracked,
		"is_perishable":    product.IsPerishable,
	}

	// Soft delete
	if err := s.db.WithContext(ctx).Model(product).Update("is_active", false).Error; err != nil {
		// Audit: Log failed operation
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &product.TenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}
		if auditErr := s.auditService.LogProductOperationFailed(ctx, auditCtx, "PRODUCT_DELETED", productID, err.Error()); auditErr != nil {
			fmt.Printf("WARNING: Failed to log failed product deletion: %v\n", auditErr)
		}

		return fmt.Errorf("failed to delete product: %w", err)
	}

	// Audit: Log successful deletion
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &product.TenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	if err := s.auditService.LogProductDeleted(ctx, auditCtx, productID, productData); err != nil {
		// Log error but don't fail the delete operation
		fmt.Printf("WARNING: Failed to create audit log for product deletion: %v\n", err)
	}

	return nil
}

// ============================================================================
// PRODUCT UNIT OPERATIONS
// ============================================================================

// AddProductUnit adds a new unit to a product
func (s *ProductService) AddProductUnit(ctx context.Context, companyID, productID string, req *dto.CreateProductUnitRequest) (*models.ProductUnit, error) {
	// Verify product exists and belongs to company
	product, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return nil, err
	}

	// Parse conversion rate
	conversionRate, err := decimal.NewFromString(req.ConversionRate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid conversionRate format")
	}

	if conversionRate.LessThanOrEqual(decimal.Zero) {
		return nil, pkgerrors.NewBadRequestError("conversionRate must be greater than 0")
	}

	// Check unit name uniqueness for this product
	var existingUnit models.ProductUnit
	err = s.db.Where("product_id = ? AND unit_name = ?", productID, req.UnitName).First(&existingUnit).Error
	if err == nil {
		return nil, pkgerrors.NewBadRequestError("unit name already exists for this product")
	}

	// Create unit
	unit := &models.ProductUnit{
		ProductID:      product.ID,
		UnitName:       req.UnitName,
		ConversionRate: conversionRate,
		IsBaseUnit:     false,
		Barcode:        req.Barcode,
		SKU:            req.SKU,
		Description:    req.Description,
		IsActive:       true,
	}

	// Parse optional decimal fields
	if req.BuyPrice != nil {
		buyPrice, err := decimal.NewFromString(*req.BuyPrice)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid buyPrice format")
		}
		unit.BuyPrice = &buyPrice
	}

	if req.SellPrice != nil {
		sellPrice, err := decimal.NewFromString(*req.SellPrice)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid sellPrice format")
		}
		unit.SellPrice = &sellPrice
	}

	if req.Weight != nil {
		weight, err := decimal.NewFromString(*req.Weight)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid weight format")
		}
		unit.Weight = &weight
	}

	if req.Volume != nil {
		volume, err := decimal.NewFromString(*req.Volume)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid volume format")
		}
		unit.Volume = &volume
	}

	if err := s.db.WithContext(ctx).Create(unit).Error; err != nil {
		return nil, fmt.Errorf("failed to create product unit: %w", err)
	}

	return unit, nil
}

// UpdateProductUnit updates a product unit
func (s *ProductService) UpdateProductUnit(ctx context.Context, companyID, productID, unitID string, req *dto.UpdateProductUnitRequest) (*models.ProductUnit, error) {
	// Verify product exists
	_, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return nil, err
	}

	// Get unit
	var unit models.ProductUnit
	err = s.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", unitID, productID).
		First(&unit).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("product unit not found")
		}
		return nil, fmt.Errorf("failed to get product unit: %w", err)
	}

	// Cannot update base unit
	if unit.IsBaseUnit {
		return nil, pkgerrors.NewBadRequestError("cannot update base unit")
	}

	// Build updates
	updates := make(map[string]interface{})

	if req.UnitName != nil {
		updates["unit_name"] = *req.UnitName
	}

	if req.ConversionRate != nil {
		conversionRate, err := decimal.NewFromString(*req.ConversionRate)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid conversionRate format")
		}
		if conversionRate.LessThanOrEqual(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("conversionRate must be greater than 0")
		}
		updates["conversion_rate"] = conversionRate
	}

	if req.BuyPrice != nil {
		buyPrice, err := decimal.NewFromString(*req.BuyPrice)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid buyPrice format")
		}
		updates["buy_price"] = buyPrice
	}

	if req.SellPrice != nil {
		sellPrice, err := decimal.NewFromString(*req.SellPrice)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid sellPrice format")
		}
		updates["sell_price"] = sellPrice
	}

	if req.Barcode != nil {
		updates["barcode"] = *req.Barcode
	}

	if req.SKU != nil {
		updates["sku"] = *req.SKU
	}

	if req.Weight != nil {
		weight, err := decimal.NewFromString(*req.Weight)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid weight format")
		}
		updates["weight"] = weight
	}

	if req.Volume != nil {
		volume, err := decimal.NewFromString(*req.Volume)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid volume format")
		}
		updates["volume"] = volume
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	// Update
	if err := s.db.WithContext(ctx).Model(&unit).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product unit: %w", err)
	}

	// Reload
	err = s.db.WithContext(ctx).Where("id = ?", unitID).First(&unit).Error
	if err != nil {
		return nil, fmt.Errorf("failed to reload product unit: %w", err)
	}

	return &unit, nil
}

// DeleteProductUnit soft deletes a product unit
func (s *ProductService) DeleteProductUnit(ctx context.Context, companyID, productID, unitID string) error {
	// Verify product exists
	_, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return err
	}

	// Get unit
	var unit models.ProductUnit
	err = s.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", unitID, productID).
		First(&unit).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgerrors.NewNotFoundError("product unit not found")
		}
		return fmt.Errorf("failed to get product unit: %w", err)
	}

	// Cannot delete base unit
	if unit.IsBaseUnit {
		return pkgerrors.NewBadRequestError("cannot delete base unit")
	}

	// Soft delete
	if err := s.db.WithContext(ctx).Model(&unit).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to delete product unit: %w", err)
	}

	return nil
}

// ============================================================================
// PRODUCT SUPPLIER OPERATIONS
// ============================================================================

// AddProductSupplier links a supplier to a product
func (s *ProductService) AddProductSupplier(ctx context.Context, companyID, productID string, req *dto.AddProductSupplierRequest) (*models.ProductSupplier, error) {
	// Verify product exists
	_, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return nil, err
	}

	// Verify supplier exists and belongs to company
	var supplier models.Supplier
	err = s.db.WithContext(ctx).
		Where("company_id = ? AND id = ? AND is_active = ?", companyID, req.SupplierID, true).
		First(&supplier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("supplier not found")
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Check if already linked
	var existing models.ProductSupplier
	err = s.db.Where("product_id = ? AND supplier_id = ?", productID, req.SupplierID).First(&existing).Error
	if err == nil {
		return nil, pkgerrors.NewBadRequestError("supplier already linked to this product")
	}

	// Parse supplier price
	supplierPrice, err := decimal.NewFromString(req.SupplierPrice)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid supplierPrice format")
	}

	leadTime := req.LeadTime
	if leadTime == 0 {
		leadTime = 7 // Default 7 days
	}

	// Create relationship
	productSupplier := &models.ProductSupplier{
		ProductID:     productID,
		SupplierID:    req.SupplierID,
		SupplierPrice: supplierPrice,
		LeadTime:      leadTime,
		IsPrimary:     req.IsPrimary,
	}

	// If primary, unset other primary suppliers
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.IsPrimary {
			if err := tx.Model(&models.ProductSupplier{}).
				Where("product_id = ? AND is_primary = ?", productID, true).
				Update("is_primary", false).Error; err != nil {
				return fmt.Errorf("failed to unset primary suppliers: %w", err)
			}
		}

		if err := tx.Create(productSupplier).Error; err != nil {
			return fmt.Errorf("failed to create product supplier: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return productSupplier, nil
}

// UpdateProductSupplier updates a product-supplier relationship
func (s *ProductService) UpdateProductSupplier(ctx context.Context, companyID, productID, productSupplierID string, req *dto.UpdateProductSupplierRequest) (*models.ProductSupplier, error) {
	// Verify product exists
	_, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return nil, err
	}

	// Get product supplier
	var ps models.ProductSupplier
	err = s.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", productSupplierID, productID).
		First(&ps).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("product supplier not found")
		}
		return nil, fmt.Errorf("failed to get product supplier: %w", err)
	}

	// Build updates
	updates := make(map[string]interface{})

	if req.SupplierPrice != nil {
		supplierPrice, err := decimal.NewFromString(*req.SupplierPrice)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid supplierPrice format")
		}
		updates["supplier_price"] = supplierPrice
	}

	if req.LeadTime != nil {
		updates["lead_time"] = *req.LeadTime
	}

	if req.IsPrimary != nil {
		updates["is_primary"] = *req.IsPrimary
	}

	// Update with transaction
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If setting as primary, unset other primary suppliers
		if req.IsPrimary != nil && *req.IsPrimary {
			if err := tx.Model(&models.ProductSupplier{}).
				Where("product_id = ? AND is_primary = ? AND id != ?", productID, true, productSupplierID).
				Update("is_primary", false).Error; err != nil {
				return fmt.Errorf("failed to unset primary suppliers: %w", err)
			}
		}

		if err := tx.Model(&ps).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update product supplier: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload
	err = s.db.WithContext(ctx).Where("id = ?", productSupplierID).First(&ps).Error
	if err != nil {
		return nil, fmt.Errorf("failed to reload product supplier: %w", err)
	}

	return &ps, nil
}

// DeleteProductSupplier removes a supplier from a product
func (s *ProductService) DeleteProductSupplier(ctx context.Context, companyID, productID, productSupplierID string) error {
	// Verify product exists
	_, err := s.GetProduct(ctx, companyID, "", productID)
	if err != nil {
		return err
	}

	// Delete relationship
	result := s.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", productSupplierID, productID).
		Delete(&models.ProductSupplier{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete product supplier: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return pkgerrors.NewNotFoundError("product supplier not found")
	}

	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// stringPtrToValue converts *string to string, returns empty string if nil
func stringPtrToValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
