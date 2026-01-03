package customer

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// ============================================================================
// CUSTOMER VALIDATION
// ============================================================================

// validateCodeUniqueness validates customer code uniqueness per company
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *CustomerService) validateCodeUniqueness(companyID, code, excludeCustomerID string) error {
	var existing models.Customer
	query := s.db.Where("company_id = ? AND code = ?", companyID, code)

	if excludeCustomerID != "" {
		query = query.Where("id != ?", excludeCustomerID)
	}

	err := query.First(&existing).Error
	if err == nil {
		return pkgerrors.NewBadRequestError("customer code already exists")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check code uniqueness: %w", err)
	}

	return nil
}

// validateDeleteCustomer validates customer deletion
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *CustomerService) validateDeleteCustomer(ctx context.Context, customer *models.Customer) error {
	// 1. Cannot delete if customer has outstanding balance
	if customer.CurrentOutstanding.GreaterThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError(
			fmt.Sprintf("cannot delete customer with outstanding balance: %s. Please clear outstanding first.",
				customer.CurrentOutstanding.String()),
		)
	}

	// 2. Cannot delete if customer has overdue amount
	if customer.OverdueAmount.GreaterThan(decimal.Zero) {
		return pkgerrors.NewBadRequestError(
			fmt.Sprintf("cannot delete customer with overdue amount: %s. Please clear overdue first.",
				customer.OverdueAmount.String()),
		)
	}

	// 3. TODO Phase 3: Cannot delete if customer has pending sales orders
	// 4. TODO Phase 3: Cannot delete if customer has unpaid invoices

	return nil
}
