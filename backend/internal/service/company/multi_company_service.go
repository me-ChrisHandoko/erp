package company

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"

	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// MultiCompanyService handles multi-company operations for PHASE 2
// Supports 1 Tenant → N Companies architecture
type MultiCompanyService struct {
	db *gorm.DB
}

func NewMultiCompanyService(db *gorm.DB) *MultiCompanyService {
	return &MultiCompanyService{
		db: db,
	}
}

// GetDB returns the database instance for external use
func (s *MultiCompanyService) GetDB() *gorm.DB {
	return s.db
}

// GetCompaniesByTenantID retrieves all companies for a tenant
func (s *MultiCompanyService) GetCompaniesByTenantID(ctx context.Context, tenantID string) ([]models.Company, error) {
	var companies []models.Company

	// Set tenant context for GORM callbacks
	err := s.db.WithContext(ctx).
		Set("tenant_id", tenantID).
		Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("created_at ASC").
		Find(&companies).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get companies: %w", err)
	}

	return companies, nil
}

// GetCompaniesByUserID retrieves all accessible companies for a user
// Considers both Tier 1 (OWNER/TENANT_ADMIN) and Tier 2 (per-company) roles
func (s *MultiCompanyService) GetCompaniesByUserID(ctx context.Context, userID string) ([]models.Company, error) {
	var companies []models.Company

	log.Printf("🔍 DEBUG [GetCompaniesByUserID]: Starting for user: %s", userID)

	// Step 1: Check if user has Tier 1 access (OWNER or TENANT_ADMIN)
	// NOTE: user_tenants table is excluded from tenant isolation (see database/tenant.go whitelist)
	var userTenants []models.UserTenant
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Where("role IN ?", []models.UserRole{models.UserRoleOwner, models.UserRoleTenantAdmin}).
		Find(&userTenants).Error

	if err != nil {
		log.Printf("❌ ERROR [GetCompaniesByUserID]: Failed to query user_tenants: %v", err)
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	log.Printf("✅ DEBUG [GetCompaniesByUserID]: Found %d user tenant records", len(userTenants))

	// If user has Tier 1 access, get ALL companies from their tenants
	if len(userTenants) > 0 {
		tenantIDs := make([]string, len(userTenants))
		for i, ut := range userTenants {
			tenantIDs[i] = ut.TenantID
		}

		log.Printf("🔍 DEBUG [GetCompaniesByUserID]: User has Tier 1 access, fetching companies for tenant IDs: %v", tenantIDs)

		// Set tenant context for GORM callbacks (use first tenant for validation)
		// The WHERE clause will still filter by all tenant IDs
		err = s.db.WithContext(ctx).
			Set("tenant_id", tenantIDs[0]).
			Where("tenant_id IN ? AND is_active = ?", tenantIDs, true).
			Order("created_at ASC").
			Find(&companies).Error

		if err != nil {
			log.Printf("❌ ERROR [GetCompaniesByUserID]: Failed to query companies: %v", err)
			return nil, fmt.Errorf("failed to get companies for tier 1 access: %w", err)
		}

		log.Printf("✅ DEBUG [GetCompaniesByUserID]: Tier 1 - Found %d companies", len(companies))
		return companies, nil
	}

	// Step 2: If no Tier 1 access, get companies via Tier 2 (UserCompanyRole)
	// IMPORTANT: Bypass tenant isolation for authentication - we don't have tenant context yet!
	var userCompanyRoles []models.UserCompanyRole
	err = s.db.WithContext(ctx).
		Set("bypass_tenant", true).
		Preload("Company").
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&userCompanyRoles).Error

	if err != nil {
		log.Printf("❌ ERROR [GetCompaniesByUserID]: Failed to query user_company_roles: %v", err)
		return nil, fmt.Errorf("failed to get user company roles: %w", err)
	}

	log.Printf("✅ DEBUG [GetCompaniesByUserID]: Tier 2 - Found %d user company role records", len(userCompanyRoles))

	// Extract unique companies
	companyMap := make(map[string]models.Company)
	for _, ucr := range userCompanyRoles {
		if ucr.Company.IsActive {
			companyMap[ucr.CompanyID] = ucr.Company
		}
	}

	// Convert map to slice
	for _, company := range companyMap {
		companies = append(companies, company)
	}

	return companies, nil
}

// GetCompanyByID retrieves a single company by ID
func (s *MultiCompanyService) GetCompanyByID(ctx context.Context, companyID string) (*models.Company, error) {
	var company models.Company

	// NOTE: This query needs bypass because it's called before access verification in handlers
	// Access control is enforced at the handler level via CheckUserCompanyAccess
	err := s.db.WithContext(ctx).
		Set("bypass_tenant", true).
		Preload("Banks", "is_active = ?", true).
		Where("id = ? AND is_active = ?", companyID, true).
		First(&company).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("company not found")
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return &company, nil
}

// CreateCompany creates a new company under a tenant (OWNER only)
func (s *MultiCompanyService) CreateCompany(ctx context.Context, tenantID string, req *CreateCompanyRequest) (*models.Company, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, pkgerrors.NewBadRequestError(err.Error())
	}

	// Verify tenant exists
	var tenant models.Tenant
	if err := s.db.WithContext(ctx).Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("tenant not found")
		}
		return nil, fmt.Errorf("failed to verify tenant: %w", err)
	}

	// Create company
	company := &models.Company{
		TenantID:   tenantID,
		Name:       req.Name,
		LegalName:  req.LegalName,
		EntityType: req.EntityType,
		Address:    req.Address,
		City:       req.City,
		Province:   req.Province,
		PostalCode: req.PostalCode,
		Phone:      req.Phone,
		Email:      req.Email,
		Website:    req.Website,
		NPWP:       req.NPWP,
		NIB:        req.NIB,
		IsPKP:      req.IsPKP,
		IsActive:   true,
	}

	// Use transaction for company creation with tenant context
	err := s.db.WithContext(ctx).
		Set("tenant_id", tenantID).
		Transaction(func(tx *gorm.DB) error {
			// Check NPWP uniqueness if provided
			if req.NPWP != nil && *req.NPWP != "" {
				var existingCompany models.Company
				err := tx.Where("npwp = ?", *req.NPWP).First(&existingCompany).Error
			if err == nil {
				return pkgerrors.NewBadRequestError("NPWP already registered to another company")
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to check NPWP uniqueness: %w", err)
			}
		}

		// Create company
		if err := tx.Create(company).Error; err != nil {
			return fmt.Errorf("failed to create company: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return company, nil
}

// UpdateCompany updates company information
func (s *MultiCompanyService) UpdateCompany(ctx context.Context, companyID string, updates map[string]interface{}) (*models.Company, error) {
	// Get existing company
	company, err := s.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Use transaction for atomic update with tenant context
	err = s.db.WithContext(ctx).
		Set("tenant_id", company.TenantID).
		Transaction(func(tx *gorm.DB) error {
			// Validate NPWP uniqueness if being updated
			if npwp, ok := updates["npwp"]; ok {
				if npwpStr, ok := npwp.(string); ok && npwpStr != "" {
					var existingCompany models.Company
					err := tx.Where("npwp = ? AND id != ?", npwpStr, companyID).First(&existingCompany).Error
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
	return s.GetCompanyByID(ctx, companyID)
}

// DeactivateCompany soft deletes a company (sets is_active = false)
func (s *MultiCompanyService) DeactivateCompany(ctx context.Context, companyID string) error {
	// Get company first to obtain tenant_id
	company, err := s.GetCompanyByID(ctx, companyID)
	if err != nil {
		return err
	}

	// Deactivate with tenant context
	result := s.db.WithContext(ctx).
		Set("tenant_id", company.TenantID).
		Model(&models.Company{}).
		Where("id = ?", companyID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return pkgerrors.NewNotFoundError("company not found")
	}

	return nil
}

// CheckUserCompanyAccess verifies if a user has access to a specific company
// Returns the access level (Tier 1 or Tier 2 role)
func (s *MultiCompanyService) CheckUserCompanyAccess(ctx context.Context, userID, companyID string) (*CompanyAccessInfo, error) {
	// Get company first (bypass tenant isolation for access verification)
	var company models.Company
	if err := s.db.WithContext(ctx).
		Set("bypass_tenant", true).
		Where("id = ?", companyID).
		First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("company not found")
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Check Tier 1 access (OWNER/TENANT_ADMIN)
	var userTenant models.UserTenant
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND is_active = ?", userID, company.TenantID, true).
		Where("role IN ?", []models.UserRole{models.UserRoleOwner, models.UserRoleTenantAdmin}).
		First(&userTenant).Error

	if err == nil {
		// User has Tier 1 access - full access to this company
		return &CompanyAccessInfo{
			CompanyID:  companyID,
			TenantID:   company.TenantID,
			AccessTier: 1,
			Role:       userTenant.Role,
			HasAccess:  true,
		}, nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check tier 1 access: %w", err)
	}

	// Check Tier 2 access (per-company role)
	var userCompanyRole models.UserCompanyRole
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND company_id = ? AND is_active = ?", userID, companyID, true).
		First(&userCompanyRole).Error

	if err == nil {
		// User has Tier 2 access
		return &CompanyAccessInfo{
			CompanyID:  companyID,
			TenantID:   company.TenantID,
			AccessTier: 2,
			Role:       userCompanyRole.Role,
			HasAccess:  true,
		}, nil
	}

	if err == gorm.ErrRecordNotFound {
		// No access
		return &CompanyAccessInfo{
			CompanyID:  companyID,
			TenantID:   company.TenantID,
			AccessTier: 0,
			Role:       "",
			HasAccess:  false,
		}, nil
	}

	return nil, fmt.Errorf("failed to check tier 2 access: %w", err)
}

// CreateCompanyRequest represents request to create a new company
type CreateCompanyRequest struct {
	Name       string  `json:"name" validate:"required,min=2,max=255"`
	LegalName  string  `json:"legalName" validate:"required,min=2,max=255"`
	EntityType string  `json:"entityType" validate:"required,oneof=PT CV UD Firma"`
	Address    string  `json:"address" validate:"required"`
	City       string  `json:"city" validate:"required"`
	Province   string  `json:"province" validate:"required"`
	PostalCode *string `json:"postalCode"`
	Phone      string  `json:"phone" validate:"required"`
	Email      string  `json:"email" validate:"required,email"`
	Website    *string `json:"website"`
	NPWP       *string `json:"npwp" validate:"omitempty,len=15"`
	NIB        *string `json:"nib"`
	IsPKP      bool    `json:"isPkp"`
}

// Validate validates the request
func (r *CreateCompanyRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("company name is required")
	}
	if r.LegalName == "" {
		return fmt.Errorf("legal name is required")
	}
	if r.EntityType != "PT" && r.EntityType != "CV" && r.EntityType != "UD" && r.EntityType != "Firma" {
		return fmt.Errorf("entity type must be PT, CV, UD, or Firma")
	}
	if r.Address == "" {
		return fmt.Errorf("address is required")
	}
	if r.City == "" {
		return fmt.Errorf("city is required")
	}
	if r.Province == "" {
		return fmt.Errorf("province is required")
	}
	if r.Phone == "" {
		return fmt.Errorf("phone is required")
	}
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.NPWP != nil && len(*r.NPWP) != 15 {
		return fmt.Errorf("NPWP must be exactly 15 characters")
	}
	return nil
}

// CompanyAccessInfo contains information about user's access to a company
type CompanyAccessInfo struct {
	CompanyID  string
	TenantID   string
	AccessTier int // 0 = no access, 1 = tenant-level, 2 = company-level
	Role       models.UserRole
	HasAccess  bool
}
