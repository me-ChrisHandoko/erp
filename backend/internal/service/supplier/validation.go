package supplier

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// ============================================================================
// SUPPLIER VALIDATION
// ============================================================================

// validateCodeUniqueness validates supplier code uniqueness per company
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *SupplierService) validateCodeUniqueness(companyID, code, excludeSupplierID string) error {
	var existing models.Supplier
	query := s.db.Where("company_id = ? AND code = ?", companyID, code)

	if excludeSupplierID != "" {
		query = query.Where("id != ?", excludeSupplierID)
	}

	err := query.First(&existing).Error
	if err == nil {
		return pkgerrors.NewBadRequestError("supplier code already exists")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check code uniqueness: %w", err)
	}

	return nil
}

// validateDeleteSupplier validates supplier deletion
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *SupplierService) validateDeleteSupplier(ctx context.Context, supplier *models.Supplier) error {
	// 1. Cannot delete if supplier has outstanding balance
	if supplier.CurrentOutstanding.GreaterThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError(
			fmt.Sprintf("cannot delete supplier with outstanding balance: %s. Please clear outstanding first.",
				supplier.CurrentOutstanding.String()),
		)
	}

	// 2. Cannot delete if supplier has overdue amount
	if supplier.OverdueAmount.GreaterThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError(
			fmt.Sprintf("cannot delete supplier with overdue amount: %s. Please clear overdue first.",
				supplier.OverdueAmount.String()),
		)
	}

	// 3. TODO Phase 3: Cannot delete if supplier has pending purchase orders
	// 4. TODO Phase 3: Cannot delete if supplier has unpaid goods receipts
	// 5. Check if supplier is linked to any products
	var productCount int64
	err := s.db.WithContext(ctx).
		Model(&models.ProductSupplier{}).
		Where("supplier_id = ?", supplier.ID).
		Count(&productCount).Error

	if err != nil {
		return fmt.Errorf("failed to check supplier product links: %w", err)
	}

	if productCount > 0 {
		return pkgerrors.NewBadRequestError(
			fmt.Sprintf("cannot delete supplier with %d linked product(s). Please remove product links first.", productCount),
		)
	}

	return nil
}
