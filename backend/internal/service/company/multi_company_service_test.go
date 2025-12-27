package company

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

// MultiCompanyServiceTestSuite defines the test suite for MultiCompanyService
type MultiCompanyServiceTestSuite struct {
	suite.Suite
	db      *gorm.DB
	service *MultiCompanyService
	ctx     context.Context
}

// SetupSuite runs once before all tests
func (s *MultiCompanyServiceTestSuite) SetupSuite() {
	// Use in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	s.Require().NoError(err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Tenant{},
		&models.Company{},
		&models.CompanyBank{},
		&models.User{},
		&models.UserTenant{},
		&models.UserCompanyRole{},
	)
	s.Require().NoError(err)

	s.db = db
	s.service = NewMultiCompanyService(db)
	s.ctx = context.Background()
}

// TearDownSuite runs once after all tests
func (s *MultiCompanyServiceTestSuite) TearDownSuite() {
	sqlDB, err := s.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

// SetupTest runs before each test
func (s *MultiCompanyServiceTestSuite) SetupTest() {
	// Clean tables before each test
	s.db.Exec("DELETE FROM user_company_roles")
	s.db.Exec("DELETE FROM user_tenants")
	s.db.Exec("DELETE FROM companies")
	s.db.Exec("DELETE FROM users")
	s.db.Exec("DELETE FROM tenants")
}

// Helper function to create test tenant
func (s *MultiCompanyServiceTestSuite) createTestTenant(name string) *models.Tenant {
	tenant := &models.Tenant{
		Name:      name,
		Subdomain: name,
		Status:    models.TenantStatusActive,
	}
	err := s.db.Create(tenant).Error
	s.Require().NoError(err)
	return tenant
}

// Helper function to create test user
func (s *MultiCompanyServiceTestSuite) createTestUser(email string) *models.User {
	user := &models.User{
		Email:        email,
		PasswordHash: "hashed_password",
	}
	err := s.db.Create(user).Error
	s.Require().NoError(err)
	return user
}

// Helper function to create test company
func (s *MultiCompanyServiceTestSuite) createTestCompany(tenantID, name, npwp string) *models.Company {
	// Remove dots and dashes from NPWP if present (keep only digits)
	cleanNPWP := ""
	for _, char := range npwp {
		if char >= '0' && char <= '9' {
			cleanNPWP += string(char)
		}
	}

	company := &models.Company{
		TenantID:      tenantID,
		Name:          name,
		NPWP:          &cleanNPWP,
		IsActive:      true,
		InvoicePrefix: "INV",
	}
	err := s.db.Create(company).Error
	s.Require().NoError(err)
	return company
}

// Helper function to assign user to tenant
func (s *MultiCompanyServiceTestSuite) assignUserToTenant(userID, tenantID string, role models.UserRole) {
	userTenant := &models.UserTenant{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		IsActive: true,
	}
	err := s.db.Create(userTenant).Error
	s.Require().NoError(err)
}

// Helper function to assign user to company
func (s *MultiCompanyServiceTestSuite) assignUserToCompany(userID, companyID, tenantID string, role models.UserRole) {
	userCompanyRole := &models.UserCompanyRole{
		UserID:    userID,
		CompanyID: companyID,
		TenantID:  tenantID,
		Role:      role,
		IsActive:  true,
	}
	err := s.db.Create(userCompanyRole).Error
	s.Require().NoError(err)
}

// Test_GetCompaniesByTenantID tests retrieving all companies for a tenant
func (s *MultiCompanyServiceTestSuite) Test_GetCompaniesByTenantID() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	company1 := s.createTestCompany(tenant.ID, "Company 1", "12.345.678.9-012.001")
	company2 := s.createTestCompany(tenant.ID, "Company 2", "12.345.678.9-012.002")
	_ = s.createTestCompany(tenant.ID, "Inactive Company", "12.345.678.9-012.003")

	// Mark third company as inactive
	s.db.Model(&models.Company{}).Where("name = ?", "Inactive Company").Update("is_active", false)

	// Execute
	companies, err := s.service.GetCompaniesByTenantID(s.ctx, tenant.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), companies, 2, "Should return only active companies")
	assert.Contains(s.T(), []string{companies[0].ID, companies[1].ID}, company1.ID)
	assert.Contains(s.T(), []string{companies[0].ID, companies[1].ID}, company2.ID)
}

// Test_GetCompaniesByUserID_Tier1Access tests retrieving companies for OWNER user
func (s *MultiCompanyServiceTestSuite) Test_GetCompaniesByUserID_Tier1Access() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("owner@test.com")
	s.assignUserToTenant(user.ID, tenant.ID, models.UserRoleOwner)

	company1 := s.createTestCompany(tenant.ID, "Company 1", "12.345.678.9-012.001")
	company2 := s.createTestCompany(tenant.ID, "Company 2", "12.345.678.9-012.002")

	// Execute
	companies, err := s.service.GetCompaniesByUserID(s.ctx, user.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), companies, 2, "OWNER should see all tenant companies")
	assert.Contains(s.T(), []string{companies[0].ID, companies[1].ID}, company1.ID)
	assert.Contains(s.T(), []string{companies[0].ID, companies[1].ID}, company2.ID)
}

// Test_GetCompaniesByUserID_Tier2Access tests retrieving companies for company-level user
func (s *MultiCompanyServiceTestSuite) Test_GetCompaniesByUserID_Tier2Access() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("admin@test.com")
	s.assignUserToTenant(user.ID, tenant.ID, models.UserRoleAdmin) // Tier 2 role

	company1 := s.createTestCompany(tenant.ID, "Company 1", "12.345.678.9-012.001")
	company2 := s.createTestCompany(tenant.ID, "Company 2", "12.345.678.9-012.002")
	_ = s.createTestCompany(tenant.ID, "Company 3", "12.345.678.9-012.003")

	// Assign user to only company1 and company2
	s.assignUserToCompany(user.ID, company1.ID, tenant.ID, models.UserRoleAdmin)
	s.assignUserToCompany(user.ID, company2.ID, tenant.ID, models.UserRoleFinance)

	// Execute
	companies, err := s.service.GetCompaniesByUserID(s.ctx, user.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), companies, 2, "User should see only assigned companies")
	assert.Contains(s.T(), []string{companies[0].ID, companies[1].ID}, company1.ID)
	assert.Contains(s.T(), []string{companies[0].ID, companies[1].ID}, company2.ID)
}

// Test_GetCompaniesByUserID_NoAccess tests user with no company access
func (s *MultiCompanyServiceTestSuite) Test_GetCompaniesByUserID_NoAccess() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("newuser@test.com")
	s.assignUserToTenant(user.ID, tenant.ID, models.UserRoleAdmin)

	_ = s.createTestCompany(tenant.ID, "Company 1", "12.345.678.9-012.001")

	// Execute
	companies, err := s.service.GetCompaniesByUserID(s.ctx, user.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.Len(s.T(), companies, 0, "User with no company assignments should see no companies")
}

// Test_GetCompanyByID tests retrieving a single company
func (s *MultiCompanyServiceTestSuite) Test_GetCompanyByID() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	company := s.createTestCompany(tenant.ID, "Test Company", "12.345.678.9-012.001")

	// Execute
	result, err := s.service.GetCompanyByID(s.ctx, company.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), company.ID, result.ID)
	assert.Equal(s.T(), company.Name, result.Name)
}

// Test_CreateCompany tests creating a new company
func (s *MultiCompanyServiceTestSuite) Test_CreateCompany() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	npwp := "123456789012001" // Valid 15-digit NPWP

	req := &CreateCompanyRequest{
		Name:       "New Company",
		LegalName:  "PT New Company Indonesia",
		EntityType: "PT",
		Address:    "Jl. Test No. 123",
		City:       "Jakarta",
		Province:   "DKI Jakarta",
		Phone:      "021-12345678",
		Email:      "info@newcompany.com",
		NPWP:       &npwp,
	}

	// Execute
	company, err := s.service.CreateCompany(s.ctx, tenant.ID, req)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), company)
	assert.Equal(s.T(), "New Company", company.Name)
	assert.Equal(s.T(), tenant.ID, company.TenantID)
	assert.Equal(s.T(), npwp, *company.NPWP)
	assert.True(s.T(), company.IsActive)
}

// Test_CreateCompany_DuplicateNPWP tests NPWP uniqueness validation
func (s *MultiCompanyServiceTestSuite) Test_CreateCompany_DuplicateNPWP() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	npwp := "123456789012001" // Valid 15-digit NPWP
	_ = s.createTestCompany(tenant.ID, "Existing Company", npwp)

	req := &CreateCompanyRequest{
		Name:       "New Company",
		LegalName:  "PT New Company Indonesia",
		EntityType: "PT",
		Address:    "Jl. Test No. 123",
		City:       "Jakarta",
		Province:   "DKI Jakarta",
		Phone:      "021-12345678",
		Email:      "info@newcompany.com",
		NPWP:       &npwp,
	}

	// Execute
	company, err := s.service.CreateCompany(s.ctx, tenant.ID, req)

	// Assert
	assert.Error(s.T(), err, "Should fail on duplicate NPWP")
	assert.Nil(s.T(), company)
	assert.Contains(s.T(), err.Error(), "NPWP already registered")
}

// Test_UpdateCompany tests updating company information
func (s *MultiCompanyServiceTestSuite) Test_UpdateCompany() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	company := s.createTestCompany(tenant.ID, "Original Name", "12.345.678.9-012.001")

	updates := map[string]interface{}{
		"name": "Updated Name",
		"city": "Jakarta",
	}

	// Execute
	updated, err := s.service.UpdateCompany(s.ctx, company.ID, updates)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), updated)
	assert.Equal(s.T(), "Updated Name", updated.Name)
	assert.Equal(s.T(), "Jakarta", updated.City)
}

// Test_DeactivateCompany tests soft delete
func (s *MultiCompanyServiceTestSuite) Test_DeactivateCompany() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	company := s.createTestCompany(tenant.ID, "Test Company", "12.345.678.9-012.001")

	// Execute
	err := s.service.DeactivateCompany(s.ctx, company.ID)

	// Assert
	assert.NoError(s.T(), err)

	// Verify company is deactivated
	var result models.Company
	s.db.First(&result, "id = ?", company.ID)
	assert.False(s.T(), result.IsActive, "Company should be deactivated")
}

// Test_CheckUserCompanyAccess_Tier1 tests Tier 1 access check
func (s *MultiCompanyServiceTestSuite) Test_CheckUserCompanyAccess_Tier1() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("owner@test.com")
	s.assignUserToTenant(user.ID, tenant.ID, models.UserRoleOwner)
	company := s.createTestCompany(tenant.ID, "Test Company", "12.345.678.9-012.001")

	// Execute
	accessInfo, err := s.service.CheckUserCompanyAccess(s.ctx, user.ID, company.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), accessInfo)
	assert.True(s.T(), accessInfo.HasAccess)
	assert.Equal(s.T(), 1, accessInfo.AccessTier, "Should have Tier 1 access")
	assert.Equal(s.T(), "OWNER", string(accessInfo.Role))
}

// Test_CheckUserCompanyAccess_Tier2 tests Tier 2 access check
func (s *MultiCompanyServiceTestSuite) Test_CheckUserCompanyAccess_Tier2() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("admin@test.com")
	s.assignUserToTenant(user.ID, tenant.ID, models.UserRoleAdmin)
	company := s.createTestCompany(tenant.ID, "Test Company", "12.345.678.9-012.001")
	s.assignUserToCompany(user.ID, company.ID, tenant.ID, models.UserRoleAdmin)

	// Execute
	accessInfo, err := s.service.CheckUserCompanyAccess(s.ctx, user.ID, company.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), accessInfo)
	assert.True(s.T(), accessInfo.HasAccess)
	assert.Equal(s.T(), 2, accessInfo.AccessTier, "Should have Tier 2 access")
	assert.Equal(s.T(), "ADMIN", string(accessInfo.Role))
}

// Test_CheckUserCompanyAccess_NoAccess tests no access scenario
func (s *MultiCompanyServiceTestSuite) Test_CheckUserCompanyAccess_NoAccess() {
	// Setup
	tenant := s.createTestTenant("test-tenant")
	user := s.createTestUser("user@test.com")
	s.assignUserToTenant(user.ID, tenant.ID, models.UserRoleAdmin)
	company := s.createTestCompany(tenant.ID, "Test Company", "12.345.678.9-012.001")
	// Don't assign user to company

	// Execute
	accessInfo, err := s.service.CheckUserCompanyAccess(s.ctx, user.ID, company.ID)

	// Assert
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), accessInfo)
	assert.False(s.T(), accessInfo.HasAccess)
	assert.Equal(s.T(), 0, accessInfo.AccessTier, "Should have no access")
}

// TestMultiCompanyServiceTestSuite runs the test suite
func TestMultiCompanyServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MultiCompanyServiceTestSuite))
}
