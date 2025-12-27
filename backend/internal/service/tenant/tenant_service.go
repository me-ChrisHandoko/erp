package tenant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/models"
	"backend/pkg/email"
	pkgerrors "backend/pkg/errors"
	"backend/pkg/security"
)

type TenantService struct {
	db             *gorm.DB
	auditService   *audit.AuditService
	cfg            *config.Config
	passwordHasher *security.PasswordHasher
	emailService   *email.EmailService
}

func NewTenantService(
	db *gorm.DB,
	cfg *config.Config,
	passwordHasher *security.PasswordHasher,
	emailService *email.EmailService,
) *TenantService {
	return &TenantService{
		db:             db,
		auditService:   audit.NewAuditService(db),
		cfg:            cfg,
		passwordHasher: passwordHasher,
		emailService:   emailService,
	}
}

// RemoveUserFromTenant removes a user from tenant with transaction safety and audit logging
// Issue #6 Fix: Uses transaction to ensure atomic admin count check + delete
// Issue #7 Fix: Logs RBAC operation to audit trail
func (s *TenantService) RemoveUserFromTenant(ctx context.Context, tenantID, userTenantID string, auditCtx *audit.AuditContext) error {
	var userTenant models.UserTenant

	// Set tenant session for querying user_tenants table (tenant-scoped)
	tenantDB := database.SetTenantSession(s.db, tenantID)

	// Get existing user-tenant link
	err := tenantDB.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userTenantID, tenantID).
		First(&userTenant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgerrors.NewNotFoundError("user-tenant link not found")
		}
		return fmt.Errorf("failed to get user-tenant link: %w", err)
	}

	// Cannot remove OWNER
	if userTenant.Role == models.UserRoleOwner {
		return pkgerrors.NewBadRequestError("cannot remove OWNER from tenant")
	}

	// Use transaction to ensure atomic operation
	err = tenantDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if this is the last ADMIN
		if userTenant.Role == models.UserRoleAdmin {
			var adminCount int64
			err := tx.Model(&models.UserTenant{}).
				Where("tenant_id = ? AND role = ? AND is_active = ? AND id != ?",
					tenantID, models.UserRoleAdmin, true, userTenantID).
				Count(&adminCount).Error

			if err != nil {
				return fmt.Errorf("failed to count admins: %w", err)
			}

			if adminCount < 1 {
				return pkgerrors.NewBadRequestError("cannot remove last ADMIN from tenant - minimum 1 ADMIN required")
			}
		}

		// Soft delete user-tenant link
		if err := tx.Model(&userTenant).Update("is_active", false).Error; err != nil {
			return fmt.Errorf("failed to deactivate user-tenant link: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Issue #7 Fix: Log audit trail for RBAC operation
	if auditCtx != nil {
		if logErr := s.auditService.LogUserRemoved(ctx, auditCtx, userTenantID, userTenant.Role); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to create audit log: %v\n", logErr)
		}
	}

	return nil
}

// UpdateUserRole updates user role in tenant with transaction safety and audit logging
// Issue #6 Fix: Uses transaction to ensure atomic admin count validation
// Issue #7 Fix: Logs RBAC operation to audit trail
func (s *TenantService) UpdateUserRole(ctx context.Context, tenantID, userTenantID string, newRole models.UserRole, auditCtx *audit.AuditContext) error {
	var userTenant models.UserTenant

	// Set tenant session for querying user_tenants table (tenant-scoped)
	tenantDB := database.SetTenantSession(s.db, tenantID)

	// Get existing user-tenant link
	err := tenantDB.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userTenantID, tenantID).
		First(&userTenant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgerrors.NewNotFoundError("user-tenant link not found")
		}
		return fmt.Errorf("failed to get user-tenant link: %w", err)
	}

	// Store old role for audit logging
	oldRole := userTenant.Role

	// Cannot change OWNER role
	if userTenant.Role == models.UserRoleOwner || newRole == models.UserRoleOwner {
		return pkgerrors.NewBadRequestError("cannot change OWNER role")
	}

	// Use transaction to ensure atomic operation
	err = tenantDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If changing from ADMIN to non-ADMIN, check if this is the last ADMIN
		if userTenant.Role == models.UserRoleAdmin && newRole != models.UserRoleAdmin {
			var adminCount int64
			err := tx.Model(&models.UserTenant{}).
				Where("tenant_id = ? AND role = ? AND is_active = ? AND id != ?",
					tenantID, models.UserRoleAdmin, true, userTenantID).
				Count(&adminCount).Error

			if err != nil {
				return fmt.Errorf("failed to count admins: %w", err)
			}

			if adminCount < 1 {
				return pkgerrors.NewBadRequestError("cannot change last ADMIN role - minimum 1 ADMIN required")
			}
		}

		// Update role
		if err := tx.Model(&userTenant).Update("role", newRole).Error; err != nil {
			return fmt.Errorf("failed to update user role: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Issue #7 Fix: Log audit trail for RBAC operation
	if auditCtx != nil {
		if logErr := s.auditService.LogUserRoleChange(ctx, auditCtx, userTenantID, oldRole, newRole); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to create audit log: %v\n", logErr)
		}
	}

	return nil
}

// AddUserToTenant adds a user to tenant with role assignment and audit logging
// Issue #7 Fix: Logs RBAC operation to audit trail
func (s *TenantService) AddUserToTenant(ctx context.Context, tenantID, userID string, role models.UserRole, auditCtx *audit.AuditContext) (*models.UserTenant, error) {
	// Validate role
	if role == models.UserRoleOwner {
		return nil, pkgerrors.NewBadRequestError("cannot assign OWNER role through this endpoint")
	}

	// Check if user-tenant link already exists
	var existing models.UserTenant
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		First(&existing).Error

	if err == nil {
		// Link exists - reactivate if inactive
		if !existing.IsActive {
			oldRole := existing.Role

			err = s.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
				"is_active": true,
				"role":      role,
			}).Error
			if err != nil {
				return nil, fmt.Errorf("failed to reactivate user-tenant link: %w", err)
			}
			// Reload to get updated values
			var reactivated models.UserTenant
			err = s.db.WithContext(ctx).Where("id = ?", existing.ID).First(&reactivated).Error
			if err != nil {
				return nil, fmt.Errorf("failed to reload user-tenant link: %w", err)
			}

			// Issue #7 Fix: Log audit trail for user reactivation
			if auditCtx != nil {
				if logErr := s.auditService.LogUserReactivated(ctx, auditCtx, existing.ID, oldRole, role); logErr != nil {
					// Log error but don't fail the operation
					fmt.Printf("Warning: Failed to create audit log: %v\n", logErr)
				}
			}

			return &reactivated, nil
		}
		return nil, pkgerrors.NewBadRequestError("user already added to tenant")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user-tenant link: %w", err)
	}

	// Create new user-tenant link
	userTenant := &models.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(userTenant).Error; err != nil {
		return nil, fmt.Errorf("failed to add user to tenant: %w", err)
	}

	// Issue #7 Fix: Log audit trail for new user addition
	if auditCtx != nil {
		if logErr := s.auditService.LogUserAdded(ctx, auditCtx, userTenant.ID, role); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to create audit log: %v\n", logErr)
		}
	}

	return userTenant, nil
}

// GetTenantDetails retrieves complete tenant information
// Reference: 01-TENANT-COMPANY-SETUP.md lines 736-771
// Updated for PHASE 3: Multi-company architecture (tenant.Company relationship removed)
func (s *TenantService) GetTenantDetails(ctx context.Context, tenantID string) (*dto.GetTenantDetailsResponse, error) {
	var tenant models.Tenant

	// Fetch tenant with subscription
	err := s.db.WithContext(ctx).
		Preload("Subscription").
		Where("id = ?", tenantID).
		First(&tenant).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Tenant not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	response := &dto.GetTenantDetailsResponse{
		ID:          tenant.ID,
		Status:      string(tenant.Status),
		TrialEndsAt: tenant.TrialEndsAt,
		// Company field removed - use /api/v1/companies endpoint to get tenant's companies
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	// Add subscription info if exists
	if tenant.Subscription != nil {
		sub := tenant.Subscription
		subInfo := &dto.SubscriptionInfo{
			ID:                 sub.ID,
			Price:              sub.Price.String(),
			BillingCycle:       sub.BillingCycle,
			Status:             string(sub.Status),
			CurrentPeriodStart: sub.CurrentPeriodStart,
			CurrentPeriodEnd:   sub.CurrentPeriodEnd,
			NextBillingDate:    sub.NextBillingDate,
			PaymentMethod:      sub.PaymentMethod,
			LastPaymentDate:    sub.LastPaymentDate,
			AutoRenew:          sub.AutoRenew,
		}

		if sub.LastPaymentAmount != nil {
			amount := sub.LastPaymentAmount.String()
			subInfo.LastPaymentAmount = &amount
		}

		response.Subscription = subInfo
	}

	return response, nil
}

// ListTenantUsers retrieves all users in a tenant with optional filters
// Reference: 01-TENANT-COMPANY-SETUP.md lines 774-820
func (s *TenantService) ListTenantUsers(ctx context.Context, tenantID string, role *string, isActive *bool) ([]dto.TenantUserInfo, error) {
	var userTenants []models.UserTenant

	// Set tenant session for querying user_tenants table (tenant-scoped)
	tenantDB := database.SetTenantSession(s.db, tenantID)

	query := tenantDB.WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	// Apply filters
	if role != nil {
		query = query.Where("role = ?", *role)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Order("created_at ASC").Find(&userTenants).Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	// Collect user IDs to load user data separately
	userIDs := make([]string, 0, len(userTenants))
	for _, ut := range userTenants {
		userIDs = append(userIDs, ut.UserID)
	}

	// Load users separately (users table has no tenant_id, so use regular db)
	var usersData []models.User
	if len(userIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&usersData).Error; err != nil {
			return nil, pkgerrors.NewInternalError(err)
		}
	}

	// Create a map for quick user lookup
	userMap := make(map[string]models.User)
	for _, user := range usersData {
		userMap[user.ID] = user
	}

	// Build flat response
	users := make([]dto.TenantUserInfo, 0, len(userTenants))
	for _, ut := range userTenants {
		// Get user data from map
		user, exists := userMap[ut.UserID]

		userInfo := dto.TenantUserInfo{
			ID:        ut.ID,
			TenantID:  ut.TenantID,
			Email:     "",
			Name:      "",
			Role:      string(ut.Role),
			IsActive:  ut.IsActive,
			CreatedAt: ut.CreatedAt,
			UpdatedAt: ut.UpdatedAt,
		}

		// Populate user fields if user exists
		if exists {
			userInfo.Email = user.Email
			userInfo.Name = user.FullName
		}

		users = append(users, userInfo)
	}

	return users, nil
}

// InviteUser invites a new user to the tenant or links existing user
// Reference: 01-TENANT-COMPANY-SETUP.md lines 823-869
func (s *TenantService) InviteUser(ctx context.Context, tenantID string, req *dto.InviteUserRequest, auditCtx *audit.AuditContext) (*dto.InviteUserResponse, error) {
	// Validate role - cannot invite OWNER
	if req.Role == "OWNER" {
		return nil, pkgerrors.NewBadRequestError("Cannot invite another OWNER. Only one OWNER per tenant is allowed.")
	}

	// Check if user already exists
	var existingUser models.User
	userExists := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error == nil

	var user models.User
	var userTenant models.UserTenant
	var invitationToken *string

	// Set tenant session for querying user_tenants table (tenant-scoped)
	tenantDB := database.SetTenantSession(s.db, tenantID)

	err := tenantDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if userExists {
			// User exists - check if already linked to this tenant
			var existingLink models.UserTenant
			linkExists := tx.Where("user_id = ? AND tenant_id = ?", existingUser.ID, tenantID).
				First(&existingLink).Error == nil

			if linkExists && existingLink.IsActive {
				return pkgerrors.NewBadRequestError("User is already a member of this tenant")
			}

			if linkExists && !existingLink.IsActive {
				// Reactivate existing link
				existingLink.IsActive = true
				existingLink.Role = models.UserRole(req.Role)
				if err := tx.Save(&existingLink).Error; err != nil {
					return pkgerrors.NewInternalError(err)
				}
				userTenant = existingLink
			} else {
				// Create new UserTenant link
				userTenant = models.UserTenant{
					UserID:   existingUser.ID,
					TenantID: tenantID,
					Role:     models.UserRole(req.Role),
					IsActive: true,
				}

				if err := tx.Create(&userTenant).Error; err != nil {
					return pkgerrors.NewInternalError(err)
				}
			}

			user = existingUser
		} else {
			// Create new user
			// Generate temporary password (will be replaced when user verifies email)
			tempPassword := generateRandomPassword(16)
			hashedPassword, err := s.passwordHasher.HashPassword(tempPassword)
			if err != nil {
				return pkgerrors.NewInternalError(err)
			}

			// Generate email verification token
			token, err := generateVerificationToken()
			if err != nil {
				return pkgerrors.NewInternalError(err)
			}
			invitationToken = &token

			user = models.User{
				Email:        req.Email,
				Username:     req.Email, // Use email as username
				PasswordHash: hashedPassword,
				FullName:     req.Name,
				IsActive:     false, // Will be activated after email verification
			}

			if err := tx.Create(&user).Error; err != nil {
				return pkgerrors.NewInternalError(err)
			}

			// Create UserTenant link
			userTenant = models.UserTenant{
				UserID:   user.ID,
				TenantID: tenantID,
				Role:     models.UserRole(req.Role),
				IsActive: true,
			}

			if err := tx.Create(&userTenant).Error; err != nil {
				return pkgerrors.NewInternalError(err)
			}

			// Send invitation email
			// TODO: Implement email sending with verification link
			// s.emailService.SendInvitationEmail(req.Email, req.FullName, *invitationToken)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Log audit trail
	if auditCtx != nil {
		if logErr := s.auditService.LogUserAdded(ctx, auditCtx, userTenant.ID, userTenant.Role); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to create audit log: %v\n", logErr)
		}
	}

	// Build flat response
	response := &dto.InviteUserResponse{
		ID:              userTenant.ID,
		TenantID:        tenantID,
		Email:           user.Email,
		Name:            user.FullName,
		Role:            req.Role,
		IsActive:        userTenant.IsActive,
		InvitationToken: invitationToken,
		CreatedAt:       userTenant.CreatedAt,
		UpdatedAt:       userTenant.UpdatedAt,
		LastLoginAt:     nil, // New user has no login yet
	}

	return response, nil
}

// Helper functions

// generateRandomPassword generates a random password
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	password := make([]byte, length)
	for i := range password {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		password[i] = charset[int(randomByte[0])%len(charset)]
	}
	return string(password)
}

// generateVerificationToken generates a secure verification token
func generateVerificationToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// ============================================================================
// COMPANY-SCOPED USER MANAGEMENT METHODS
// These methods work with UserCompanyRole for multi-company isolation
// ============================================================================

// ListCompanyUsers retrieves users for a specific company
// Filters users via UserCompanyRole junction table
func (s *TenantService) ListCompanyUsers(ctx context.Context, companyID, tenantID string, filters dto.GetUsersFilters) ([]dto.TenantUserInfo, error) {
	var users []dto.TenantUserInfo

	// Build query with UserCompanyRole join
	// Note: ID is from user_company_roles (the junction table), not users table
	query := s.db.WithContext(ctx).
		Table("user_company_roles").
		Select(`
			user_company_roles.id,
			user_company_roles.tenant_id,
			users.email,
			users.name as name,
			user_company_roles.role,
			user_company_roles.is_active,
			user_company_roles.created_at,
			user_company_roles.updated_at
		`).
		Joins("INNER JOIN users ON user_company_roles.user_id = users.id").
		Where("user_company_roles.company_id = ?", companyID).
		Where("user_company_roles.tenant_id = ?", tenantID)

	// Apply filters
	if filters.Role != nil {
		query = query.Where("user_company_roles.role = ?", *filters.Role)
	}
	if filters.IsActive != nil {
		query = query.Where("user_company_roles.is_active = ?", *filters.IsActive)
	}

	// Execute query
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list company users: %w", err)
	}

	return users, nil
}

// InviteUserToCompany invites a user to a specific company
// Creates User (if new) and UserCompanyRole
func (s *TenantService) InviteUserToCompany(ctx context.Context, tenantID, companyID string, req dto.InviteUserRequest) (*dto.TenantUserInfo, error) {
	// Validate role - cannot invite OWNER at company level
	if req.Role == "OWNER" {
		return nil, pkgerrors.NewBadRequestError("Cannot invite OWNER at company level. OWNER is tenant-level only.")
	}

	// Check if user already exists
	var existingUser models.User
	userExists := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error == nil

	var user models.User
	var userCompanyRole models.UserCompanyRole

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if userExists {
			// User exists - check if already has role in this company
			var existingRole models.UserCompanyRole
			roleExists := tx.Where("user_id = ? AND company_id = ?", existingUser.ID, companyID).
				First(&existingRole).Error == nil

			if roleExists && existingRole.IsActive {
				return pkgerrors.NewBadRequestError("User already has access to this company")
			}

			if roleExists && !existingRole.IsActive {
				// Reactivate existing role
				existingRole.IsActive = true
				existingRole.Role = models.UserRole(req.Role)
				if err := tx.Save(&existingRole).Error; err != nil {
					return pkgerrors.NewInternalError(err)
				}
				userCompanyRole = existingRole
			} else {
				// Create new UserCompanyRole
				userCompanyRole = models.UserCompanyRole{
					UserID:    existingUser.ID,
					CompanyID: companyID,
					TenantID:  tenantID,
					Role:      models.UserRole(req.Role),
					IsActive:  true,
				}

				if err := tx.Create(&userCompanyRole).Error; err != nil {
					return pkgerrors.NewInternalError(err)
				}
			}

			user = existingUser
		} else {
			// Create new user
			tempPassword := generateRandomPassword(16)
			hashedPassword, err := s.passwordHasher.HashPassword(tempPassword)
			if err != nil {
				return pkgerrors.NewInternalError(err)
			}

			user = models.User{
				Email:        req.Email,
				Username:     req.Email,
				PasswordHash: hashedPassword,
				FullName:     req.Name,
				IsActive:     false, // Activated after email verification
			}

			if err := tx.Create(&user).Error; err != nil {
				return pkgerrors.NewInternalError(err)
			}

			// Create UserCompanyRole
			userCompanyRole = models.UserCompanyRole{
				UserID:    user.ID,
				CompanyID: companyID,
				TenantID:  tenantID,
				Role:      models.UserRole(req.Role),
				IsActive:  true,
			}

			if err := tx.Create(&userCompanyRole).Error; err != nil {
				return pkgerrors.NewInternalError(err)
			}

			// TODO: Send invitation email
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return user info
	response := &dto.TenantUserInfo{
		ID:          userCompanyRole.ID,
		TenantID:    tenantID,
		Email:       user.Email,
		Name:        user.FullName,
		Role:        string(userCompanyRole.Role),
		IsActive:    userCompanyRole.IsActive,
		CreatedAt:   userCompanyRole.CreatedAt,
		UpdatedAt:   userCompanyRole.UpdatedAt,
		LastLoginAt: nil, // User model doesn't track last login yet
	}

	return response, nil
}

// UpdateUserRoleInCompany updates user's role in a specific company
// Modifies UserCompanyRole for the given company
func (s *TenantService) UpdateUserRoleInCompany(ctx context.Context, tenantID, companyID, userCompanyRoleID string, newRole models.UserRole) (*dto.TenantUserInfo, error) {
	// Cannot set OWNER at company level
	if newRole == models.UserRoleOwner {
		return nil, pkgerrors.NewBadRequestError("Cannot assign OWNER role at company level")
	}

	var userCompanyRole models.UserCompanyRole

	// Get existing role
	err := s.db.WithContext(ctx).
		Where("id = ? AND company_id = ? AND tenant_id = ?", userCompanyRoleID, companyID, tenantID).
		First(&userCompanyRole).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("user role in company not found")
		}
		return nil, fmt.Errorf("failed to get user company role: %w", err)
	}

	// Update role
	userCompanyRole.Role = newRole
	if err := s.db.WithContext(ctx).Save(&userCompanyRole).Error; err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	// Get user details
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userCompanyRole.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}

	// Return updated user info
	response := &dto.TenantUserInfo{
		ID:          userCompanyRole.ID,
		TenantID:    tenantID,
		Email:       user.Email,
		Name:        user.FullName,
		Role:        string(userCompanyRole.Role),
		IsActive:    userCompanyRole.IsActive,
		CreatedAt:   userCompanyRole.CreatedAt,
		UpdatedAt:   userCompanyRole.UpdatedAt,
		LastLoginAt: nil, // User model doesn't track last login yet
	}

	return response, nil
}

// RemoveUserFromCompany removes user's access to a specific company
// Soft deletes UserCompanyRole (sets is_active = false)
func (s *TenantService) RemoveUserFromCompany(ctx context.Context, tenantID, companyID, userCompanyRoleID string) error {
	var userCompanyRole models.UserCompanyRole

	// Get existing role
	err := s.db.WithContext(ctx).
		Where("id = ? AND company_id = ? AND tenant_id = ?", userCompanyRoleID, companyID, tenantID).
		First(&userCompanyRole).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkgerrors.NewNotFoundError("user role in company not found")
		}
		return fmt.Errorf("failed to get user company role: %w", err)
	}

	// Soft delete by setting is_active to false
	userCompanyRole.IsActive = false
	if err := s.db.WithContext(ctx).Save(&userCompanyRole).Error; err != nil {
		return fmt.Errorf("failed to remove user from company: %w", err)
	}

	return nil
}
