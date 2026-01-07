package warehouse

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// ============================================================================
// WAREHOUSE VALIDATION
// ============================================================================

// validateCodeUniqueness validates warehouse code uniqueness per company
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *WarehouseService) validateCodeUniqueness(ctx context.Context, tenantID, companyID, code, excludeWarehouseID string) error {
	var existing models.Warehouse
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("company_id = ? AND code = ?", companyID, code)

	if excludeWarehouseID != "" {
		query = query.Where("id != ?", excludeWarehouseID)
	}

	err := query.First(&existing).Error
	if err == nil {
		return pkgerrors.NewBadRequestError("warehouse code already exists")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check code uniqueness: %w", err)
	}

	return nil
}

// validateManagerExists validates that the specified manager exists
func (s *WarehouseService) validateManagerExists(managerID string) error {
	var user models.User
	err := s.db.Where("id = ?", managerID).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return pkgerrors.NewBadRequestError("manager user not found")
	} else if err != nil {
		return fmt.Errorf("failed to check manager existence: %w", err)
	}

	return nil
}

// validateDeleteWarehouse validates warehouse deletion
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *WarehouseService) validateDeleteWarehouse(ctx context.Context, tenantID string, warehouse *models.Warehouse) error {
	// 1. Cannot delete if warehouse has stock
	var totalStock decimal.Decimal
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.WarehouseStock{}).
		Where("warehouse_id = ?", warehouse.ID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error

	if err != nil {
		return fmt.Errorf("failed to check warehouse stock: %w", err)
	}

	if totalStock.GreaterThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError(
			fmt.Sprintf("cannot delete warehouse with stock: %s. Please transfer stock out first.", totalStock.String()),
		)
	}

	// 2. TODO Phase 3: Cannot delete if warehouse has pending deliveries
	// 3. TODO Phase 3: Cannot delete if warehouse has pending goods receipts
	// 4. TODO Phase 3: Cannot delete if warehouse has pending stock transfers

	return nil
}
