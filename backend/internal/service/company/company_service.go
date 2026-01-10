package company

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"backend/internal/dto"
	"backend/models"
	pkgerrors "backend/pkg/errors"
	"backend/pkg/validator"
)

type CompanyService struct {
	db                *gorm.DB
	businessValidator *validator.BusinessValidator
}

func NewCompanyService(db *gorm.DB) *CompanyService {
	return &CompanyService{
		db:                db,
		businessValidator: validator.NewBusinessValidator(db),
	}
}

// GetCompanyByTenantID retrieves company profile for the given company ID within tenant context
// DEPRECATED: This method is for backward compatibility. Use GetCompanyByID instead.
// In multi-company architecture, tenants can have multiple companies.
func (s *CompanyService) GetCompanyByTenantID(ctx context.Context, companyID string) (*models.Company, error) {
	var company models.Company

	// Get company with banks
	err := s.db.WithContext(ctx).
		Preload("Banks", "is_active = ?", true).
		Where("id = ?", companyID).
		First(&company).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("company not found")
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

// UpdateCompany updates company profile with transaction safety
// Issue #6 Fix: Uses transaction to ensure atomic updates
func (s *CompanyService) UpdateCompany(ctx context.Context, tenantID string, updates map[string]interface{}) (*models.Company, error) {
	// Get existing company
	company, err := s.GetCompanyByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Use transaction for atomic update
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Validate NPWP uniqueness if being updated
		if npwp, ok := updates["npwp"]; ok {
			if npwpStr, ok := npwp.(string); ok && npwpStr != "" {
				// Check if NPWP is already used by another company
				var existingCompany models.Company
				err := tx.Where("npwp = ? AND id != ?", npwpStr, company.ID).First(&existingCompany).Error
				if err == nil {
					return pkgerrors.NewBadRequestError("NPWP already registered to another company")
				} else if err != gorm.ErrRecordNotFound {
					return fmt.Errorf("failed to check NPWP uniqueness: %w", err)
				}
			}
		}

		// Update company
		if err := tx.Model(company).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update company: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload company with updated data
	return s.GetCompanyByTenantID(ctx, tenantID)
}

// AddBankAccount adds a new bank account with transaction safety
// Issue #6 Fix: Uses transaction to ensure atomic unset primary + create new
// AddBankAccount adds a new bank account to a company
// PHASE 5: Updated to accept companyID directly instead of tenantID
func (s *CompanyService) AddBankAccount(ctx context.Context, companyID string, req *AddBankRequest) (*models.CompanyBank, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, pkgerrors.NewBadRequestError(err.Error())
	}

	var bank *models.CompanyBank

	// Use transaction to ensure atomic operation
	// Issue #10 Fix: Add SELECT FOR UPDATE to prevent race conditions on primary bank selection
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If isPrimary = true, lock and unset other primary banks FIRST (within transaction)
		if req.IsPrimary {
			// SELECT FOR UPDATE locks the rows, preventing concurrent modifications
			var existingPrimaryBanks []models.CompanyBank
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("company_id = ? AND is_primary = ? AND is_active = ?", companyID, true, true).
				Find(&existingPrimaryBanks).Error; err != nil {
				return fmt.Errorf("failed to lock primary banks: %w", err)
			}

			// Unset all locked primary banks
			if len(existingPrimaryBanks) > 0 {
				if err := tx.Model(&models.CompanyBank{}).
					Where("company_id = ? AND is_primary = ? AND is_active = ?", companyID, true, true).
					Update("is_primary", false).Error; err != nil {
					return fmt.Errorf("failed to unset primary banks: %w", err)
				}
			}
		}

		// Create bank account
		bank = &models.CompanyBank{
			CompanyID:     companyID,
			BankName:      req.BankName,
			AccountNumber: req.AccountNumber,
			AccountName:   req.AccountName,
			BranchName:    req.BranchName,
			IsPrimary:     req.IsPrimary,
			CheckPrefix:   req.CheckPrefix,
			IsActive:      true,
		}

		if err := tx.Create(bank).Error; err != nil {
			return fmt.Errorf("failed to create bank account: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return bank, nil
}

// UpdateBankAccount updates bank account with transaction safety
// Issue #6 Fix: Uses transaction for atomic primary bank updates
// PHASE 5: Updated to accept companyID directly instead of tenantID
func (s *CompanyService) UpdateBankAccount(ctx context.Context, companyID, bankID string, updates map[string]interface{}) (*models.CompanyBank, error) {
	// Get existing bank
	var bank models.CompanyBank
	if err := s.db.WithContext(ctx).Where("id = ? AND company_id = ?", bankID, companyID).First(&bank).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("bank account not found")
		}
		return nil, fmt.Errorf("failed to get bank account: %w", err)
	}

	// Use transaction for atomic update
	// Issue #10 Fix: Add SELECT FOR UPDATE to prevent race conditions on primary bank selection
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If setting as primary, lock and unset others first (within transaction)
		if isPrimary, ok := updates["is_primary"]; ok {
			if isPrimaryBool, ok := isPrimary.(bool); ok && isPrimaryBool {
				// SELECT FOR UPDATE locks the rows, preventing concurrent modifications
				var existingPrimaryBanks []models.CompanyBank
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("company_id = ? AND id != ? AND is_primary = ? AND is_active = ?", companyID, bankID, true, true).
					Find(&existingPrimaryBanks).Error; err != nil {
					return fmt.Errorf("failed to lock primary banks: %w", err)
				}

				// Unset all locked primary banks
				if len(existingPrimaryBanks) > 0 {
					if err := tx.Model(&models.CompanyBank{}).
						Where("company_id = ? AND id != ? AND is_primary = ? AND is_active = ?", companyID, bankID, true, true).
						Update("is_primary", false).Error; err != nil {
						return fmt.Errorf("failed to unset primary banks: %w", err)
					}
				}
			}
		}

		// Update bank
		if err := tx.Model(&bank).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update bank account: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload bank
	if err := s.db.WithContext(ctx).Where("id = ?", bankID).First(&bank).Error; err != nil {
		return nil, fmt.Errorf("failed to reload bank account: %w", err)
	}

	return &bank, nil
}

// DeleteBankAccount deletes (soft delete) a bank account with validation
// Issue #6 Fix: Validates minimum bank account requirement
// PHASE 5: Updated to accept companyID directly instead of tenantID
func (s *CompanyService) DeleteBankAccount(ctx context.Context, companyID, bankID string) error {
	// Validate minimum bank account requirement
	if err := s.businessValidator.ValidateMinimumBankAccount(ctx, companyID, bankID); err != nil {
		return err
	}

	// Soft delete bank account
	result := s.db.WithContext(ctx).
		Model(&models.CompanyBank{}).
		Where("id = ? AND company_id = ?", bankID, companyID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to delete bank account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return pkgerrors.NewNotFoundError("bank account not found")
	}

	return nil
}

// GetBankAccounts retrieves all active bank accounts for the current company
// PHASE 5: Updated to accept companyID directly instead of tenantID
// DEPRECATED: Use ListBankAccounts for paginated results
func (s *CompanyService) GetBankAccounts(ctx context.Context, companyID string) ([]*models.CompanyBank, error) {
	// Get all active bank accounts
	var banks []*models.CompanyBank
	if err := s.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Order("is_primary DESC, created_at ASC").
		Find(&banks).Error; err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	return banks, nil
}

// ListBankAccounts retrieves paginated bank accounts with filters
// Follows the same pattern as ListProducts for consistency
func (s *CompanyService) ListBankAccounts(ctx context.Context, companyID string, filters *dto.BankAccountFilters) ([]*models.CompanyBank, int64, error) {
	var banks []*models.CompanyBank
	var total int64

	// Build base query
	query := s.db.WithContext(ctx).
		Model(&models.CompanyBank{}).
		Where("company_id = ?", companyID)

	// Apply search filter (search by bank name or account number)
	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where(
			"bank_name LIKE ? OR account_number LIKE ?",
			searchTerm, searchTerm,
		)
	}

	// Apply isPrimary filter
	if filters.IsPrimary != nil {
		query = query.Where("is_primary = ?", *filters.IsPrimary)
	}

	// Apply isActive filter
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		// Default: only show active banks if not explicitly filtering
		query = query.Where("is_active = ?", true)
	}

	// Count total records (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count bank accounts: %w", err)
	}

	// Apply sorting - Map camelCase to snake_case for database columns
	sortByColumn := filters.SortBy
	switch filters.SortBy {
	case "bankName":
		sortByColumn = "bank_name"
	case "createdAt":
		sortByColumn = "created_at"
	}
	orderClause := sortByColumn + " " + filters.SortOrder
	query = query.Order(orderClause)

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Execute query
	if err := query.Find(&banks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list bank accounts: %w", err)
	}

	return banks, total, nil
}

// GetBankAccountByID retrieves a single bank account by ID for the current company
// PHASE 5: Updated to accept companyID directly instead of tenantID
func (s *CompanyService) GetBankAccountByID(ctx context.Context, companyID, bankID string) (*models.CompanyBank, error) {
	// Get bank account
	var bank models.CompanyBank
	if err := s.db.WithContext(ctx).
		Where("id = ? AND company_id = ? AND is_active = ?", bankID, companyID, true).
		First(&bank).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("bank account not found")
		}
		return nil, fmt.Errorf("failed to get bank account: %w", err)
	}

	return &bank, nil
}

// AddBankRequest represents request to add bank account
type AddBankRequest struct {
	BankName      string  `json:"bankName" validate:"required,min=2,max=100"`
	AccountNumber string  `json:"accountNumber" validate:"required,min=8,max=50"`
	AccountName   string  `json:"accountName" validate:"required,min=3,max=255"`
	BranchName    *string `json:"branchName"`
	IsPrimary     bool    `json:"isPrimary"`
	CheckPrefix   *string `json:"checkPrefix" validate:"omitempty,max=20"`
}

// Validate validates the request
func (r *AddBankRequest) Validate() error {
	if r.BankName == "" {
		return fmt.Errorf("bank name is required")
	}
	if len(r.BankName) < 2 || len(r.BankName) > 100 {
		return fmt.Errorf("bank name must be between 2 and 100 characters")
	}
	if r.AccountNumber == "" {
		return fmt.Errorf("account number is required")
	}
	if len(r.AccountNumber) < 8 || len(r.AccountNumber) > 50 {
		return fmt.Errorf("account number must be between 8 and 50 characters")
	}
	if r.AccountName == "" {
		return fmt.Errorf("account name is required")
	}
	if len(r.AccountName) < 3 || len(r.AccountName) > 255 {
		return fmt.Errorf("account name must be between 3 and 255 characters")
	}
	if r.CheckPrefix != nil && len(*r.CheckPrefix) > 20 {
		return fmt.Errorf("check prefix must be at most 20 characters")
	}
	return nil
}
