package product

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// ============================================================================
// PRODUCT VALIDATION
// ============================================================================

// validateCreateProduct validates product creation request
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Product Validation Rules)
func (s *ProductService) validateCreateProduct(
	ctx context.Context,
	companyID string,
	tenantID string,
	req *dto.CreateProductRequest,
	baseCost, basePrice, minimumStock decimal.Decimal,
) error {
	// 1. Code uniqueness per company
	var existing models.Product
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("company_id = ? AND code = ?", companyID, req.Code).First(&existing).Error
	if err == nil {
		return pkgerrors.NewBadRequestError("product code already exists")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check code uniqueness: %w", err)
	}

	// 2. BasePrice >= BaseCost
	if basePrice.LessThan(baseCost) {
		return pkgerrors.NewBadRequestError("base price must be greater than or equal to base cost")
	}

	// 3. MinimumStock >= 0
	if minimumStock.LessThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError("minimum stock cannot be negative")
	}

	// 4. Barcode uniqueness across products and product_units (Issue #3 fix)
	if req.Barcode != nil && *req.Barcode != "" {
		if err := s.validateBarcodeUniqueness(ctx, tenantID, *req.Barcode, ""); err != nil {
			return err
		}
	}

	// 5. Unit conversion rates must be > 0
	for _, unit := range req.Units {
		conversionRate, err := decimal.NewFromString(unit.ConversionRate)
		if err != nil {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid conversionRate for unit %s", unit.UnitName))
		}
		if conversionRate.LessThanOrEqual(decimal.Zero) {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("conversion rate for unit %s must be greater than 0", unit.UnitName))
		}
	}

	// 6. Unit name uniqueness within product
	unitNames := make(map[string]bool)
	unitNames[req.BaseUnit] = true // Base unit is always created

	for _, unit := range req.Units {
		if unitNames[unit.UnitName] {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("duplicate unit name: %s", unit.UnitName))
		}
		unitNames[unit.UnitName] = true
	}

	// 7. Validate unit barcodes uniqueness
	for _, unit := range req.Units {
		if unit.Barcode != nil && *unit.Barcode != "" {
			if err := s.validateBarcodeUniqueness(ctx, tenantID, *unit.Barcode, ""); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateUpdateProduct validates product update request
func (s *ProductService) validateUpdateProduct(ctx context.Context, companyID, tenantID, productID string, req *dto.UpdateProductRequest) error {
	// Get current product to check validation against existing values
	product, err := s.GetProduct(ctx, companyID, tenantID, productID)
	if err != nil {
		return err
	}

	// Calculate effective base cost and price after update
	baseCost := product.BaseCost
	if req.BaseCost != nil {
		newBaseCost, err := decimal.NewFromString(*req.BaseCost)
		if err != nil {
			return pkgerrors.NewBadRequestError("invalid baseCost format")
		}
		baseCost = newBaseCost
	}

	basePrice := product.BasePrice
	if req.BasePrice != nil {
		newBasePrice, err := decimal.NewFromString(*req.BasePrice)
		if err != nil {
			return pkgerrors.NewBadRequestError("invalid basePrice format")
		}
		basePrice = newBasePrice
	}

	// Validate base price >= base cost (check if either is being updated)
	if req.BasePrice != nil || req.BaseCost != nil {
		if basePrice.LessThan(baseCost) {
			return pkgerrors.NewBadRequestError("base price must be greater than or equal to base cost")
		}
	}

	// 2. If updating code, validate uniqueness within company
	if req.Code != nil && *req.Code != "" {
		// Only validate if code actually changed
		if *req.Code != product.Code {
			var existing models.Product
			err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
				Where("company_id = ? AND code = ? AND id != ?", companyID, *req.Code, productID).
				First(&existing).Error

			if err == nil {
				return pkgerrors.NewBadRequestError("kode produk sudah digunakan oleh produk lain dalam perusahaan ini")
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to check code uniqueness: %w", err)
			}
		}
	}

	// 3. If updating barcode, validate uniqueness
	if req.Barcode != nil && *req.Barcode != "" {
		if err := s.validateBarcodeUniqueness(ctx, tenantID, *req.Barcode, productID); err != nil {
			return err
		}
	}

	// 4. MinimumStock >= 0
	if req.MinimumStock != nil {
		minimumStock, err := decimal.NewFromString(*req.MinimumStock)
		if err != nil {
			return pkgerrors.NewBadRequestError("invalid minimumStock format")
		}
		if minimumStock.LessThan(decimal.Zero) {
			return pkgerrors.NewBadRequestError("minimum stock cannot be negative")
		}
	}

	return nil
}

// validateDeleteProduct validates product deletion
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *ProductService) validateDeleteProduct(ctx context.Context, product *models.Product) error {
	// 1. Cannot delete if product has stock in any warehouse
	var totalStock decimal.Decimal
	err := s.db.WithContext(ctx).Model(&models.WarehouseStock{}).
		Joins("JOIN products ON products.id = warehouse_stocks.product_id").
		Where("products.company_id = ? AND products.id = ?", product.CompanyID, product.ID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error

	if err != nil {
		return fmt.Errorf("failed to check product stock: %w", err)
	}

	if totalStock.GreaterThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError("cannot delete product with stock in warehouses. Please transfer stock out first.")
	}

	// 2. TODO Phase 3: Cannot delete if product has pending sales orders
	// 3. TODO Phase 3: Cannot delete if product has outstanding invoices

	return nil
}

// validateBarcodeUniqueness validates barcode uniqueness across products and product_units
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Issue #3
func (s *ProductService) validateBarcodeUniqueness(ctx context.Context, tenantID string, barcode string, excludeProductID string) error {
	// Check in products table
	var existingProduct models.Product
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("barcode = ?", barcode)
	if excludeProductID != "" {
		query = query.Where("id != ?", excludeProductID)
	}

	err := query.First(&existingProduct).Error
	if err == nil {
		return pkgerrors.NewBadRequestError("barcode already exists in another product")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check barcode uniqueness in products: %w", err)
	}

	// Check in product_units table
	var existingUnit models.ProductUnit
	unitQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("barcode = ?", barcode)
	if excludeProductID != "" {
		unitQuery = unitQuery.Where("product_id != ?", excludeProductID)
	}

	err = unitQuery.First(&existingUnit).Error
	if err == nil {
		return pkgerrors.NewBadRequestError("barcode already exists in another product unit")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check barcode uniqueness in product units: %w", err)
	}

	return nil
}
