package permission

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// PermissionService handles dual-tier permission system for PHASE 2
// Tier 1: OWNER, TENANT_ADMIN (tenant-level access to all companies)
// Tier 2: ADMIN, FINANCE, SALES, WAREHOUSE, STAFF (per-company access)
type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// AssignUserToCompany assigns a user to a company with a specific role (Tier 2)
// Only ADMIN, FINANCE, SALES, WAREHOUSE, STAFF roles allowed
func (s *PermissionService) AssignUserToCompany(ctx context.Context, req *AssignUserRequest) (*models.UserCompanyRole, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, pkgerrors.NewBadRequestError(err.Error())
	}

	// Validate role is Tier 2 only
	role := models.UserRole(req.Role)
	if !role.IsCompanyLevel() {
		return nil, pkgerrors.NewBadRequestError("only company-level roles allowed (ADMIN, FINANCE, SALES, WAREHOUSE, STAFF)")
	}

	// Verify user exists
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", req.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("user not found")
		}
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	// Verify company exists and get tenant_id
	var company models.Company
	if err := s.db.WithContext(ctx).Where("id = ?", req.CompanyID).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("company not found")
		}
		return nil, fmt.Errorf("failed to verify company: %w", err)
	}

	// Check if assignment already exists
	var existing models.UserCompanyRole
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND company_id = ?", req.UserID, req.CompanyID).
		First(&existing).Error

	if err == nil {
		// Update existing assignment
		existing.Role = role
		existing.IsActive = true

		if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to update user-company role: %w", err)
		}

		return &existing, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing assignment: %w", err)
	}

	// Create new assignment
	ucr := &models.UserCompanyRole{
		UserID:    req.UserID,
		CompanyID: req.CompanyID,
		TenantID:  company.TenantID, // Denormalized for query optimization
		Role:      role,
		IsActive:  true,
	}

	if err := s.db.WithContext(ctx).Create(ucr).Error; err != nil {
		return nil, fmt.Errorf("failed to create user-company role: %w", err)
	}

	return ucr, nil
}

// RemoveUserFromCompany removes a user's access to a company
func (s *PermissionService) RemoveUserFromCompany(ctx context.Context, userID, companyID string) error {
	result := s.db.WithContext(ctx).
		Table("user_company_roles").
		Where("user_id = ? AND company_id = ?", userID, companyID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to remove user from company: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return pkgerrors.NewNotFoundError("user-company assignment not found")
	}

	return nil
}

// GetUserCompanyRoles retrieves all company roles for a user
func (s *PermissionService) GetUserCompanyRoles(ctx context.Context, userID string) ([]models.UserCompanyRole, error) {
	var roles []models.UserCompanyRole

	// NOTE: user_company_roles table is excluded from tenant isolation (see database/tenant.go whitelist)
	err := s.db.WithContext(ctx).
		Preload("Company").
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&roles).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user company roles: %w", err)
	}

	return roles, nil
}

// GetCompanyUsers retrieves all users with access to a company
func (s *PermissionService) GetCompanyUsers(ctx context.Context, companyID string) ([]UserCompanyInfo, error) {
	var userCompanyRoles []models.UserCompanyRole

	err := s.db.WithContext(ctx).
		Preload("User").
		Where("company_id = ? AND is_active = ?", companyID, true).
		Find(&userCompanyRoles).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get company users: %w", err)
	}

	// Convert to UserCompanyInfo
	result := make([]UserCompanyInfo, len(userCompanyRoles))
	for i, ucr := range userCompanyRoles {
		result[i] = UserCompanyInfo{
			UserID:    ucr.UserID,
			UserName:  ucr.User.FullName,
			UserEmail: ucr.User.Email,
			Role:      ucr.Role,
			IsActive:  ucr.IsActive,
		}
	}

	return result, nil
}

// CheckPermission checks if a user has a specific permission for a company
// Considers both Tier 1 and Tier 2 access
func (s *PermissionService) CheckPermission(ctx context.Context, userID, companyID string, requiredPermission Permission) (bool, error) {
	// Get company to find tenant
	var company models.Company
	if err := s.db.WithContext(ctx).Where("id = ?", companyID).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, pkgerrors.NewNotFoundError("company not found")
		}
		return false, fmt.Errorf("failed to get company: %w", err)
	}

	// Check Tier 1 access (OWNER/TENANT_ADMIN) - has all permissions
	var userTenant models.UserTenant
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND is_active = ?", userID, company.TenantID, true).
		Where("role IN ?", []models.UserRole{models.UserRoleOwner, models.UserRoleTenantAdmin}).
		First(&userTenant).Error

	if err == nil {
		// Tier 1 users have all permissions
		return true, nil
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("failed to check tier 1 access: %w", err)
	}

	// Check Tier 2 access (per-company role)
	var userCompanyRole models.UserCompanyRole
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND company_id = ? AND is_active = ?", userID, companyID, true).
		First(&userCompanyRole).Error

	if err == gorm.ErrRecordNotFound {
		// No access
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check tier 2 access: %w", err)
	}

	// Check if role has the required permission
	hasPermission := s.roleHasPermission(userCompanyRole.Role, requiredPermission)
	return hasPermission, nil
}

// roleHasPermission checks if a role has a specific permission
func (s *PermissionService) roleHasPermission(role models.UserRole, permission Permission) bool {
	// Define permission matrix
	permissionMatrix := map[models.UserRole][]Permission{
		models.UserRoleAdmin: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionDeleteData,
			PermissionApproveTransactions,
			PermissionManageUsers,
			PermissionViewReports,
			PermissionManageSettings,
		},
		models.UserRoleFinance: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionApproveTransactions,
			PermissionViewReports,
		},
		models.UserRoleSales: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionViewReports,
		},
		models.UserRoleWarehouse: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
		},
		models.UserRoleStaff: {
			PermissionViewData,
		},
	}

	// Get permissions for this role
	permissions, exists := permissionMatrix[role]
	if !exists {
		return false
	}

	// Check if permission is in the list
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetUserPermissionsForCompany returns all permissions a user has for a company
func (s *PermissionService) GetUserPermissionsForCompany(ctx context.Context, userID, companyID string) ([]Permission, error) {
	// Get company to find tenant
	var company models.Company
	if err := s.db.WithContext(ctx).Where("id = ?", companyID).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("company not found")
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Check Tier 1 access (OWNER/TENANT_ADMIN) - has all permissions
	var userTenant models.UserTenant
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND is_active = ?", userID, company.TenantID, true).
		Where("role IN ?", []models.UserRole{models.UserRoleOwner, models.UserRoleTenantAdmin}).
		First(&userTenant).Error

	if err == nil {
		// Tier 1 users have all permissions
		return []Permission{
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionDeleteData,
			PermissionApproveTransactions,
			PermissionManageUsers,
			PermissionViewReports,
			PermissionManageSettings,
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

	if err == gorm.ErrRecordNotFound {
		// No access
		return []Permission{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check tier 2 access: %w", err)
	}

	// Get permissions for this role
	permissionMatrix := map[models.UserRole][]Permission{
		models.UserRoleAdmin: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionDeleteData,
			PermissionApproveTransactions,
			PermissionManageUsers,
			PermissionViewReports,
			PermissionManageSettings,
		},
		models.UserRoleFinance: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionApproveTransactions,
			PermissionViewReports,
		},
		models.UserRoleSales: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
			PermissionViewReports,
		},
		models.UserRoleWarehouse: {
			PermissionViewData,
			PermissionCreateData,
			PermissionEditData,
		},
		models.UserRoleStaff: {
			PermissionViewData,
		},
	}

	permissions, exists := permissionMatrix[userCompanyRole.Role]
	if !exists {
		return []Permission{}, nil
	}

	return permissions, nil
}

// AssignUserRequest represents request to assign user to company
type AssignUserRequest struct {
	UserID    string `json:"userId" validate:"required"`
	CompanyID string `json:"companyId" validate:"required"`
	Role      string `json:"role" validate:"required,oneof=ADMIN FINANCE SALES WAREHOUSE STAFF"`
}

// Validate validates the request
func (r *AssignUserRequest) Validate() error {
	if r.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if r.CompanyID == "" {
		return fmt.Errorf("company ID is required")
	}
	if r.Role == "" {
		return fmt.Errorf("role is required")
	}

	// Validate role is company-level
	validRoles := []string{"ADMIN", "FINANCE", "SALES", "WAREHOUSE", "STAFF"}
	isValid := false
	for _, validRole := range validRoles {
		if r.Role == validRole {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid role: must be ADMIN, FINANCE, SALES, WAREHOUSE, or STAFF")
	}

	return nil
}

// UserCompanyInfo contains user information with company role
type UserCompanyInfo struct {
	UserID    string
	UserName  string
	UserEmail string
	Role      models.UserRole
	IsActive  bool
}

// Permission represents a specific permission action
type Permission string

const (
	PermissionViewData             Permission = "VIEW_DATA"
	PermissionCreateData           Permission = "CREATE_DATA"
	PermissionEditData             Permission = "EDIT_DATA"
	PermissionDeleteData           Permission = "DELETE_DATA"
	PermissionApproveTransactions  Permission = "APPROVE_TRANSACTIONS"
	PermissionManageUsers          Permission = "MANAGE_USERS"
	PermissionViewReports          Permission = "VIEW_REPORTS"
	PermissionManageSettings       Permission = "MANAGE_SETTINGS"
)
