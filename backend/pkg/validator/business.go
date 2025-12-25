package validator

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"backend/pkg/errors"
)

// BusinessValidator provides business logic validation
// This is separate from struct validation and focuses on business rules
type BusinessValidator struct {
	db *gorm.DB
}

// NewBusinessValidator creates a new business validator
func NewBusinessValidator(db *gorm.DB) *BusinessValidator {
	return &BusinessValidator{
		db: db,
	}
}

// ValidateMinimumBankAccount ensures at least 1 active bank account remains
// This validation is critical for invoice generation which requires bank account info
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #4 (lines 140-179)
func (v *BusinessValidator) ValidateMinimumBankAccount(ctx context.Context, companyID string, excludeBankID string) error {
	// Count active bank accounts for company (excluding the one being deleted)
	var count int64
	query := v.db.WithContext(ctx).
		Table("company_banks").
		Where("company_id = ? AND is_active = ?", companyID, true)

	// Exclude the bank being deleted from count
	if excludeBankID != "" {
		query = query.Where("id != ?", excludeBankID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to count bank accounts: %w", err)
	}

	// Minimum 1 bank account required
	if count < 1 {
		return errors.NewBadRequestError("Cannot delete last bank account - minimum 1 required for invoice generation")
	}

	return nil
}

// ValidateCompanyHasBankAccounts ensures company has at least one active bank account
// Used when creating invoices or other operations requiring bank info
func (v *BusinessValidator) ValidateCompanyHasBankAccounts(ctx context.Context, companyID string) error {
	var count int64
	err := v.db.WithContext(ctx).
		Table("company_banks").
		Where("company_id = ? AND is_active = ?", companyID, true).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to count bank accounts: %w", err)
	}

	if count == 0 {
		return errors.NewBadRequestError("Company must have at least one active bank account")
	}

	return nil
}

// ValidateUniquePrimaryBank ensures only one bank account is marked as primary
// When setting isPrimary=true, all other banks should be isPrimary=false
func (v *BusinessValidator) ValidateUniquePrimaryBank(ctx context.Context, companyID string, newPrimaryBankID string) error {
	// Count primary banks excluding the new one
	var count int64
	err := v.db.WithContext(ctx).
		Table("company_banks").
		Where("company_id = ? AND is_primary = ? AND id != ? AND is_active = ?",
			companyID, true, newPrimaryBankID, true).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to count primary bank accounts: %w", err)
	}

	// Should have 0 other primary banks (the new one will become primary)
	if count > 0 {
		return errors.NewBadRequestError("Another bank account is already set as primary. Please unset it first.")
	}

	return nil
}
