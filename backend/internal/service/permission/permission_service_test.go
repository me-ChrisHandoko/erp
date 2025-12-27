package permission

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

// PermissionServiceTestSuite defines the test suite for PermissionService
type PermissionServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *PermissionService
	ctx     context.Context
}

// SetupSuite runs once before all tests
func (s *PermissionServiceTestSuite) SetupSuite() {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Tenant{},
		&models.Company{},
		&models.User{},
		&models.UserTenant{},
		&models.UserCompanyRole{},
	)
	s.Require().NoError(err)

	s.db = db
	s.service = NewPermissionService(db)
	s.ctx = context.Background()
}

// TearDownSuite runs once after all tests
func (s *PermissionServiceTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest runs before each test
func (s *PermissionServiceTestSuite) SetupTest() {
	// Clean tables before each test
	s.db.Exec("DELETE FROM user_company_roles")
	s.db.Exec("DELETE FROM user_tenants")
	s.db.Exec("DELETE FROM companies")
	s.db.Exec("DELETE FROM users")
	s.db.Exec("DELETE FROM tenants")
}

// Helper functions
func (s *PermissionServiceTestSuite) createTestTenant(name string) *models.Tenant {
	tenant := &models.Tenant{
		Name:      name,
		Subdomain: name,
		Status:    models.TenantStatusActive,
	}
	err := s.db.Create(tenant).Error
	s.Require().NoError(err)
	return tenant
}

func (s *PermissionServiceTestSuite) createTestUser(email string) *models.User {
	// Generate username from email (part before @)
	username := email[:strings.Index(email, "@")]
	fullName := "Test User " + username

	user := &models.User{
		Email:        email,
		Username:     username,
		FullName:     fullName,
		PasswordHash: "hashed_password",
		IsActive:     true,
	}
	err := s.db.Create(user).Error
	s.Require().NoError(err)
	return user
}

func (s *PermissionServiceTestSuite) createTestCompany(tenantID, name string) *models.Company {
	// Generate unique NPWP based on company name (hash-like but deterministic)
	// Use simple sum of characters to create unique 15-digit NPWP
	sum := 0
	for _, char := range name {
		sum += int(char)
	}
	npwp := fmt.Sprintf("%015d", 123456789000000+sum) // Ensure 15 digits and uniqueness

	company := &models.Company{
		TenantID:      tenantID,
		Name:          name,
		NPWP:          &npwp,
		IsActive:      true,
		InvoicePrefix: "INV",
	}
	err := s.db.Create(company).Error
	s.Require().NoError(err)
	return company
}

// Test_AssignUserToCompany tests assigning user to company with role
func (s *PermissionServiceTestSuite) Test_AssignUserToCompany() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("user@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	req := &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleAdmin),
	}

	// Execute
	result, err := s.service.AssignUserToCompany(s.ctx, req)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), user.ID, result.UserID)
	assert.Equal(s.T(), company.ID, result.CompanyID)
	assert.Equal(s.T(), models.UserRoleAdmin, result.Role)
	assert.True(s.T(), result.IsActive)
}

// Test_AssignUserToCompany_Tier1Role tests that Tier 1 roles cannot be assigned
func (s *PermissionServiceTestSuite) Test_AssignUserToCompany_Tier1Role() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("user@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	req := &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleOwner), // Tier 1 role - should fail
	}

	// Execute
	result, err := s.service.AssignUserToCompany(s.ctx, req)

	// Assert
	assert.Error(s.T(), err, "Should fail when assigning Tier 1 role")
	assert.Nil(s.T(), result)
	assert.Contains(s.T(), err.Error(), "invalid role")
}

// Test_RemoveUserFromCompany tests removing user from company
func (s *PermissionServiceTestSuite) Test_RemoveUserFromCompany() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("user@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign user first
	req := &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleAdmin),
	}
	_, err := s.service.AssignUserToCompany(s.ctx, req)
	s.Require().NoError(err)

	// Execute
	err = s.service.RemoveUserFromCompany(s.ctx, user.ID, company.ID)

	// Assert
	assert.NoError(s.T(), err)

	// Verify user is removed
	var ucr models.UserCompanyRole
	result := s.db.Where("user_id = ? AND company_id = ? AND is_active = ?",
		user.ID, company.ID, true).First(&ucr)
	assert.Error(s.T(), result.Error, "Should not find active assignment")
}

// Test_GetUserCompanyRoles tests retrieving all company roles for a user
func (s *PermissionServiceTestSuite) Test_GetUserCompanyRoles() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("user@test.com")
	company1 := s.createTestCompany(tenant.ID, "Company 1")
	company2 := s.createTestCompany(tenant.ID, "Company 2")

	// Assign user to both companies
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company1.ID,
		Role:      string(models.UserRoleAdmin),
	})
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company2.ID,
		Role:      string(models.UserRoleFinance),
	})

	// Execute
	roles, err := s.service.GetUserCompanyRoles(s.ctx, user.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), roles, 2)
}

// Test_CheckPermission_Owner tests that OWNER has all permissions
func (s *PermissionServiceTestSuite) Test_CheckPermission_Owner() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("owner@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Create Tier 1 user (OWNER)
	userTenant := &models.UserTenant{
		UserID:   user.ID,
		TenantID: tenant.ID,
		Role:     models.UserRoleOwner,
		IsActive: true,
	}
	s.db.Create(userTenant)

	// Execute - check various permissions
	permissions := []Permission{
		PermissionViewData,
		PermissionCreateData,
		PermissionEditData,
		PermissionDeleteData,
		PermissionApproveTransactions,
		PermissionManageUsers,
		PermissionViewReports,
		PermissionManageSettings,
	}

	for _, perm := range permissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.True(s.T(), hasPermission, "OWNER should have permission: "+string(perm))
	}
}

// Test_CheckPermission_Admin tests ADMIN role permissions
func (s *PermissionServiceTestSuite) Test_CheckPermission_Admin() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("admin@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign as ADMIN
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleAdmin),
	})

	// Execute - ADMIN should have all permissions except tenant-level ones
	permissions := []Permission{
		PermissionViewData,
		PermissionCreateData,
		PermissionEditData,
		PermissionDeleteData,
		PermissionApproveTransactions,
		PermissionManageUsers,
		PermissionViewReports,
		PermissionManageSettings,
	}

	for _, perm := range permissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.True(s.T(), hasPermission, "ADMIN should have permission: "+string(perm))
	}
}

// Test_CheckPermission_Finance tests FINANCE role permissions
func (s *PermissionServiceTestSuite) Test_CheckPermission_Finance() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("finance@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign as FINANCE
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleFinance),
	})

	// Execute - check allowed permissions
	allowedPermissions := []Permission{
		PermissionViewData,
		PermissionCreateData,
		PermissionEditData,
		PermissionApproveTransactions,
		PermissionViewReports,
	}

	for _, perm := range allowedPermissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.True(s.T(), hasPermission, "FINANCE should have permission: "+string(perm))
	}

	// Check denied permissions
	deniedPermissions := []Permission{
		PermissionDeleteData,
		PermissionManageUsers,
		PermissionManageSettings,
	}

	for _, perm := range deniedPermissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.False(s.T(), hasPermission, "FINANCE should NOT have permission: "+string(perm))
	}
}

// Test_CheckPermission_Sales tests SALES role permissions
func (s *PermissionServiceTestSuite) Test_CheckPermission_Sales() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("sales@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign as SALES
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleSales),
	})

	// Execute - check allowed permissions
	allowedPermissions := []Permission{
		PermissionViewData,
		PermissionCreateData,
		PermissionEditData,
		PermissionViewReports,
	}

	for _, perm := range allowedPermissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.True(s.T(), hasPermission, "SALES should have permission: "+string(perm))
	}

	// Check denied permissions
	deniedPermissions := []Permission{
		PermissionDeleteData,
		PermissionApproveTransactions,
		PermissionManageUsers,
		PermissionManageSettings,
	}

	for _, perm := range deniedPermissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.False(s.T(), hasPermission, "SALES should NOT have permission: "+string(perm))
	}
}

// Test_CheckPermission_Staff tests STAFF role permissions (view only)
func (s *PermissionServiceTestSuite) Test_CheckPermission_Staff() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("staff@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign as STAFF
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleStaff),
	})

	// Execute - STAFF should only have VIEW permission
	hasView, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, PermissionViewData)
	assert.NoError(s.T(), err)
	assert.True(s.T(), hasView, "STAFF should have VIEW permission")

	// Check denied permissions
	deniedPermissions := []Permission{
		PermissionCreateData,
		PermissionEditData,
		PermissionDeleteData,
		PermissionApproveTransactions,
		PermissionManageUsers,
		PermissionViewReports,
		PermissionManageSettings,
	}

	for _, perm := range deniedPermissions {
		hasPermission, err := s.service.CheckPermission(s.ctx, user.ID, company.ID, perm)
		assert.NoError(s.T(), err)
		assert.False(s.T(), hasPermission, "STAFF should NOT have permission: "+string(perm))
	}
}

// Test_GetUserPermissionsForCompany tests getting all permissions for user
func (s *PermissionServiceTestSuite) Test_GetUserPermissionsForCompany() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("admin@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign as ADMIN
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleAdmin),
	})

	// Execute
	permissions, err := s.service.GetUserPermissionsForCompany(s.ctx, user.ID, company.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), permissions)
	assert.Contains(s.T(), permissions, PermissionViewData)
	assert.Contains(s.T(), permissions, PermissionCreateData)
	assert.Contains(s.T(), permissions, PermissionEditData)
}

// Test_GetCompanyUsers tests retrieving all users with access to a company
func (s *PermissionServiceTestSuite) Test_GetCompanyUsers() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user1 := s.createTestUser("user1@test.com")
	user2 := s.createTestUser("user2@test.com")
	company := s.createTestCompany(tenant.ID, "Test Company")

	// Assign both users
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user1.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleAdmin),
	})
	s.service.AssignUserToCompany(s.ctx, &AssignUserRequest{
		UserID:    user2.ID,
		CompanyID: company.ID,
		Role:      string(models.UserRoleFinance),
	})

	// Execute
	users, err := s.service.GetCompanyUsers(s.ctx, company.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 2)

	// Verify user IDs
	userIDs := []string{users[0].UserID, users[1].UserID}
	assert.Contains(s.T(), userIDs, user1.ID)
	assert.Contains(s.T(), userIDs, user2.ID)
}

// TestPermissionServiceTestSuite runs the test suite
func TestPermissionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionServiceTestSuite))
}
