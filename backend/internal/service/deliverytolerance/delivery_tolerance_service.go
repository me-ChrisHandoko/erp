package deliverytolerance

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

// DeliveryToleranceService - Business logic for delivery tolerance settings
type DeliveryToleranceService struct {
	db           *gorm.DB
	auditService *audit.AuditService
}

// NewDeliveryToleranceService creates a new delivery tolerance service instance
func NewDeliveryToleranceService(db *gorm.DB, auditService *audit.AuditService) *DeliveryToleranceService {
	return &DeliveryToleranceService{
		db:           db,
		auditService: auditService,
	}
}

// ============================================================================
// CREATE DELIVERY TOLERANCE
// ============================================================================

// CreateDeliveryTolerance creates a new delivery tolerance setting
func (s *DeliveryToleranceService) CreateDeliveryTolerance(ctx context.Context, tenantID, companyID, userID string, req *dto.CreateDeliveryToleranceRequest, ipAddress, userAgent string) (*models.DeliveryTolerance, error) {
	// Parse tolerance percentages
	underTolerance, err := decimal.NewFromString(req.UnderDeliveryTolerance)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid underDeliveryTolerance format")
	}
	overTolerance, err := decimal.NewFromString(req.OverDeliveryTolerance)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid overDeliveryTolerance format")
	}

	// Validate tolerance range (0-100%)
	hundred := decimal.NewFromInt(100)
	zero := decimal.Zero
	if underTolerance.LessThan(zero) || underTolerance.GreaterThan(hundred) {
		return nil, pkgerrors.NewBadRequestError("underDeliveryTolerance must be between 0 and 100")
	}
	if overTolerance.LessThan(zero) || overTolerance.GreaterThan(hundred) {
		return nil, pkgerrors.NewBadRequestError("overDeliveryTolerance must be between 0 and 100")
	}

	// Validate level-specific requirements
	level := models.DeliveryToleranceLevel(req.Level)
	switch level {
	case models.ToleranceLevelCompany:
		// Company level should not have category or product
		if req.CategoryName != nil || req.ProductID != nil {
			return nil, pkgerrors.NewBadRequestError("COMPANY level tolerance should not have categoryName or productId")
		}
	case models.ToleranceLevelCategory:
		// Category level requires category name
		if req.CategoryName == nil || *req.CategoryName == "" {
			return nil, pkgerrors.NewBadRequestError("CATEGORY level tolerance requires categoryName")
		}
		if req.ProductID != nil {
			return nil, pkgerrors.NewBadRequestError("CATEGORY level tolerance should not have productId")
		}
	case models.ToleranceLevelProduct:
		// Product level requires product ID
		if req.ProductID == nil {
			return nil, pkgerrors.NewBadRequestError("PRODUCT level tolerance requires productId")
		}
		if req.CategoryName != nil {
			return nil, pkgerrors.NewBadRequestError("PRODUCT level tolerance should not have categoryName")
		}
		// Validate product exists
		var product models.Product
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("id = ? AND company_id = ?", *req.ProductID, companyID).
			First(&product).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, pkgerrors.NewBadRequestError("product not found")
			}
			return nil, fmt.Errorf("failed to validate product: %w", err)
		}
	default:
		return nil, pkgerrors.NewBadRequestError("invalid tolerance level")
	}

	// Check for duplicate tolerance setting
	existsQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Model(&models.DeliveryTolerance{}).
		Where("company_id = ? AND level = ?", companyID, level)

	if req.CategoryName != nil {
		existsQuery = existsQuery.Where("category_name = ?", *req.CategoryName)
	} else {
		existsQuery = existsQuery.Where("category_name = ''")
	}

	if req.ProductID != nil {
		existsQuery = existsQuery.Where("product_id = ?", *req.ProductID)
	} else {
		existsQuery = existsQuery.Where("product_id = ''")
	}

	var count int64
	if err := existsQuery.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check for duplicate tolerance: %w", err)
	}
	if count > 0 {
		return nil, pkgerrors.NewConflictError("tolerance setting already exists for this level/category/product combination")
	}

	// Set default isActive
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Convert CategoryName from *string to string (empty string if nil)
	categoryName := ""
	if req.CategoryName != nil {
		categoryName = *req.CategoryName
	}

	// Convert ProductID from *string to string (empty string if nil)
	productID := ""
	if req.ProductID != nil {
		productID = *req.ProductID
	}

	// Create tolerance setting
	tolerance := &models.DeliveryTolerance{
		ID:                     uuid.New().String(),
		TenantID:               tenantID,
		CompanyID:              companyID,
		Level:                  level,
		CategoryName:           categoryName,
		ProductID:              productID,
		UnderDeliveryTolerance: underTolerance,
		OverDeliveryTolerance:  overTolerance,
		UnlimitedOverDelivery:  req.UnlimitedOverDelivery,
		IsActive:               isActive,
		Notes:                  req.Notes,
		CreatedBy:              &userID,
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Create(tolerance).Error; err != nil {
		return nil, fmt.Errorf("failed to create delivery tolerance: %w", err)
	}

	// Load relations
	return s.GetDeliveryToleranceByID(ctx, tenantID, companyID, tolerance.ID)
}

// ============================================================================
// GET DELIVERY TOLERANCE
// ============================================================================

// GetDeliveryToleranceByID retrieves a delivery tolerance by ID
func (s *DeliveryToleranceService) GetDeliveryToleranceByID(ctx context.Context, tenantID, companyID, toleranceID string) (*models.DeliveryTolerance, error) {
	var tolerance models.DeliveryTolerance
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		Where("id = ? AND company_id = ?", toleranceID, companyID).
		First(&tolerance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("delivery tolerance not found")
		}
		return nil, fmt.Errorf("failed to get delivery tolerance: %w", err)
	}
	return &tolerance, nil
}

// ============================================================================
// LIST DELIVERY TOLERANCES
// ============================================================================

// ListDeliveryTolerances lists delivery tolerances with filtering and pagination
func (s *DeliveryToleranceService) ListDeliveryTolerances(ctx context.Context, tenantID, companyID string, query *dto.DeliveryToleranceListQuery) (*dto.DeliveryToleranceListResponse, error) {
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
		sortByMap := map[string]string{
			"level":     "level",
			"createdAt": "created_at",
			"updatedAt": "updated_at",
		}
		if mapped, ok := sortByMap[query.SortBy]; ok {
			sortBy = mapped
		}
	}

	sortOrder := "desc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}

	// Build base query
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.DeliveryTolerance{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Level != nil {
		baseQuery = baseQuery.Where("level = ?", *query.Level)
	}

	if query.CategoryName != nil {
		baseQuery = baseQuery.Where("category_name = ?", *query.CategoryName)
	}

	if query.ProductID != nil {
		baseQuery = baseQuery.Where("product_id = ?", *query.ProductID)
	}

	if query.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *query.IsActive)
	}

	// Count total
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count delivery tolerances: %w", err)
	}

	// Fetch data with pagination
	var tolerances []models.DeliveryTolerance
	offset := (page - 1) * pageSize
	if err := baseQuery.
		Preload("Product").
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Offset(offset).
		Limit(pageSize).
		Find(&tolerances).Error; err != nil {
		return nil, fmt.Errorf("failed to list delivery tolerances: %w", err)
	}

	// Map to response
	responses := make([]dto.DeliveryToleranceResponse, len(tolerances))
	for i, t := range tolerances {
		responses[i] = s.MapToResponse(&t)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.DeliveryToleranceListResponse{
		Success: true,
		Data:    responses,
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      pageSize,
			Total:      int(totalCount),
			TotalPages: totalPages,
		},
	}, nil
}

// ============================================================================
// UPDATE DELIVERY TOLERANCE
// ============================================================================

// UpdateDeliveryTolerance updates an existing delivery tolerance
func (s *DeliveryToleranceService) UpdateDeliveryTolerance(ctx context.Context, tenantID, companyID, toleranceID, userID string, req *dto.UpdateDeliveryToleranceRequest) (*models.DeliveryTolerance, error) {
	tolerance, err := s.GetDeliveryToleranceByID(ctx, tenantID, companyID, toleranceID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.UnderDeliveryTolerance != nil {
		underTolerance, err := decimal.NewFromString(*req.UnderDeliveryTolerance)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid underDeliveryTolerance format")
		}
		hundred := decimal.NewFromInt(100)
		zero := decimal.Zero
		if underTolerance.LessThan(zero) || underTolerance.GreaterThan(hundred) {
			return nil, pkgerrors.NewBadRequestError("underDeliveryTolerance must be between 0 and 100")
		}
		tolerance.UnderDeliveryTolerance = underTolerance
	}

	if req.OverDeliveryTolerance != nil {
		overTolerance, err := decimal.NewFromString(*req.OverDeliveryTolerance)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid overDeliveryTolerance format")
		}
		hundred := decimal.NewFromInt(100)
		zero := decimal.Zero
		if overTolerance.LessThan(zero) || overTolerance.GreaterThan(hundred) {
			return nil, pkgerrors.NewBadRequestError("overDeliveryTolerance must be between 0 and 100")
		}
		tolerance.OverDeliveryTolerance = overTolerance
	}

	if req.UnlimitedOverDelivery != nil {
		tolerance.UnlimitedOverDelivery = *req.UnlimitedOverDelivery
	}

	if req.IsActive != nil {
		tolerance.IsActive = *req.IsActive
	}

	if req.Notes != nil {
		tolerance.Notes = req.Notes
	}

	tolerance.UpdatedBy = &userID

	// Save changes
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(tolerance).Error; err != nil {
		return nil, fmt.Errorf("failed to update delivery tolerance: %w", err)
	}

	return s.GetDeliveryToleranceByID(ctx, tenantID, companyID, toleranceID)
}

// ============================================================================
// DELETE DELIVERY TOLERANCE
// ============================================================================

// DeleteDeliveryTolerance deletes a delivery tolerance
func (s *DeliveryToleranceService) DeleteDeliveryTolerance(ctx context.Context, tenantID, companyID, toleranceID string) error {
	tolerance, err := s.GetDeliveryToleranceByID(ctx, tenantID, companyID, toleranceID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Delete(tolerance).Error; err != nil {
		return fmt.Errorf("failed to delete delivery tolerance: %w", err)
	}

	return nil
}

// ============================================================================
// GET EFFECTIVE TOLERANCE
// ============================================================================

// GetEffectiveTolerance gets the effective tolerance for a product
// Resolution order: Product > Category > Company > Default
func (s *DeliveryToleranceService) GetEffectiveTolerance(ctx context.Context, tenantID, companyID, productID string) (*dto.EffectiveToleranceResponse, error) {
	// Get product info
	var product models.Product
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", productID, companyID).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Try to find product-level tolerance
	var productTolerance models.DeliveryTolerance
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ? AND level = ? AND product_id = ? AND is_active = ?", companyID, models.ToleranceLevelProduct, productID, true).
		First(&productTolerance).Error
	if err == nil {
		return &dto.EffectiveToleranceResponse{
			ProductID:              product.ID,
			ProductCode:            product.Code,
			ProductName:            product.Name,
			UnderDeliveryTolerance: productTolerance.UnderDeliveryTolerance.String(),
			OverDeliveryTolerance:  productTolerance.OverDeliveryTolerance.String(),
			UnlimitedOverDelivery:  productTolerance.UnlimitedOverDelivery,
			ResolvedFrom:           "PRODUCT",
			ToleranceID:            &productTolerance.ID,
		}, nil
	}

	// Try to find category-level tolerance
	if product.Category != nil && *product.Category != "" {
		var categoryTolerance models.DeliveryTolerance
		err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("company_id = ? AND level = ? AND category_name = ? AND is_active = ?", companyID, models.ToleranceLevelCategory, *product.Category, true).
			First(&categoryTolerance).Error
		if err == nil {
			return &dto.EffectiveToleranceResponse{
				ProductID:              product.ID,
				ProductCode:            product.Code,
				ProductName:            product.Name,
				UnderDeliveryTolerance: categoryTolerance.UnderDeliveryTolerance.String(),
				OverDeliveryTolerance:  categoryTolerance.OverDeliveryTolerance.String(),
				UnlimitedOverDelivery:  categoryTolerance.UnlimitedOverDelivery,
				ResolvedFrom:           "CATEGORY",
				ToleranceID:            &categoryTolerance.ID,
			}, nil
		}
	}

	// Try to find company-level tolerance
	var companyTolerance models.DeliveryTolerance
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ? AND level = ? AND is_active = ?", companyID, models.ToleranceLevelCompany, true).
		First(&companyTolerance).Error
	if err == nil {
		return &dto.EffectiveToleranceResponse{
			ProductID:              product.ID,
			ProductCode:            product.Code,
			ProductName:            product.Name,
			UnderDeliveryTolerance: companyTolerance.UnderDeliveryTolerance.String(),
			OverDeliveryTolerance:  companyTolerance.OverDeliveryTolerance.String(),
			UnlimitedOverDelivery:  companyTolerance.UnlimitedOverDelivery,
			ResolvedFrom:           "COMPANY",
			ToleranceID:            &companyTolerance.ID,
		}, nil
	}

	// Return default (0% tolerance)
	return &dto.EffectiveToleranceResponse{
		ProductID:              product.ID,
		ProductCode:            product.Code,
		ProductName:            product.Name,
		UnderDeliveryTolerance: "0",
		OverDeliveryTolerance:  "0",
		UnlimitedOverDelivery:  false,
		ResolvedFrom:           "DEFAULT",
		ToleranceID:            nil,
	}, nil
}

// ============================================================================
// RESPONSE MAPPING
// ============================================================================

// MapToResponse maps a DeliveryTolerance model to response DTO
func (s *DeliveryToleranceService) MapToResponse(t *models.DeliveryTolerance) dto.DeliveryToleranceResponse {
	// Convert CategoryName from string to *string (nil if empty)
	var categoryNamePtr *string
	if t.CategoryName != "" {
		categoryNamePtr = &t.CategoryName
	}

	// Convert ProductID from string to *string (nil if empty)
	var productIDPtr *string
	if t.ProductID != "" {
		productIDPtr = &t.ProductID
	}

	response := dto.DeliveryToleranceResponse{
		ID:                     t.ID,
		Level:                  string(t.Level),
		CategoryName:           categoryNamePtr,
		ProductID:              productIDPtr,
		UnderDeliveryTolerance: t.UnderDeliveryTolerance.String(),
		OverDeliveryTolerance:  t.OverDeliveryTolerance.String(),
		UnlimitedOverDelivery:  t.UnlimitedOverDelivery,
		IsActive:               t.IsActive,
		Notes:                  t.Notes,
		CreatedAt:              t.CreatedAt,
		UpdatedAt:              t.UpdatedAt,
		CreatedBy:              t.CreatedBy,
		UpdatedBy:              t.UpdatedBy,
	}

	// Map product if loaded
	if t.Product != nil && t.Product.ID != "" {
		response.Product = &dto.DeliveryToleranceProductResponse{
			ID:       t.Product.ID,
			Code:     t.Product.Code,
			Name:     t.Product.Name,
			Category: t.Product.Category,
			BaseUnit: t.Product.BaseUnit,
		}
	}

	return response
}
