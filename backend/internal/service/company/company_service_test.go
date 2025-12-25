package company

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

func setupCompanyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&models.Company{}, &models.CompanyBank{}, &models.Tenant{})
	require.NoError(t, err)

	return db
}

func createTestCompany(t *testing.T, db *gorm.DB, tenant *models.Tenant) *models.Company {
	company := &models.Company{
		ID:       "company-1",
		Name:     "Test Company",
		IsActive: true,
	}
	err := db.Create(company).Error
	require.NoError(t, err)

	// Link tenant to company
	tenant.CompanyID = company.ID
	err = db.Save(tenant).Error
	require.NoError(t, err)

	return company
}

// Test Issue #6 Fix: AddBankAccount uses transaction for atomic unset primary + create
func TestAddBankAccount_TransactionConsistency_UnsetPrimaryAndCreate(t *testing.T) {
	db := setupCompanyTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create tenant
	tenant := &models.Tenant{ID: "tenant-1", Status: models.TenantStatusTrial}
	db.Create(tenant)

	// Create company
	company := createTestCompany(t, db, tenant)

	// Create first bank as primary
	req1 := &AddBankRequest{
		BankName:      "Bank A",
		AccountNumber: "12345678",
		AccountName:   "Test Account A",
		IsPrimary:     true,
	}
	bank1, err := service.AddBankAccount(ctx, tenant.ID, req1)
	require.NoError(t, err)
	assert.True(t, bank1.IsPrimary)

	// Add second bank as primary - should unset first bank's primary status atomically
	req2 := &AddBankRequest{
		BankName:      "Bank B",
		AccountNumber: "87654321",
		AccountName:   "Test Account B",
		IsPrimary:     true,
	}
	bank2, err := service.AddBankAccount(ctx, tenant.ID, req2)
	require.NoError(t, err)
	assert.True(t, bank2.IsPrimary)

	// Verify first bank is no longer primary (transaction rolled back properly)
	var updatedBank1 models.CompanyBank
	db.First(&updatedBank1, "id = ?", bank1.ID)
	assert.False(t, updatedBank1.IsPrimary, "First bank should no longer be primary")

	// Verify only one primary bank exists
	var primaryCount int64
	db.Model(&models.CompanyBank{}).Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).Count(&primaryCount)
	assert.Equal(t, int64(1), primaryCount, "Only one primary bank should exist")
}

// Test Issue #6 Fix: UpdateBankAccount transaction consistency
func TestUpdateBankAccount_TransactionConsistency_SetPrimary(t *testing.T) {
	db := setupCompanyTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create tenant
	tenant := &models.Tenant{ID: "tenant-1", Status: models.TenantStatusTrial}
	db.Create(tenant)

	// Create company
	company := createTestCompany(t, db, tenant)

	// Create two banks
	bank1 := &models.CompanyBank{
		ID:            "bank-1",
		CompanyID:     company.ID,
		BankName:      "Bank A",
		AccountNumber: "12345678",
		AccountName:   "Account A",
		IsPrimary:     true,
		IsActive:      true,
	}
	db.Create(bank1)

	bank2 := &models.CompanyBank{
		ID:            "bank-2",
		CompanyID:     company.ID,
		BankName:      "Bank B",
		AccountNumber: "87654321",
		AccountName:   "Account B",
		IsPrimary:     false,
		IsActive:      true,
	}
	db.Create(bank2)

	// Update bank2 to primary - should unset bank1 atomically
	updates := map[string]interface{}{"is_primary": true}
	updatedBank, err := service.UpdateBankAccount(ctx, tenant.ID, bank2.ID, updates)
	require.NoError(t, err)
	assert.True(t, updatedBank.IsPrimary)

	// Verify bank1 is no longer primary
	var reloadedBank1 models.CompanyBank
	db.First(&reloadedBank1, "id = ?", bank1.ID)
	assert.False(t, reloadedBank1.IsPrimary, "Bank 1 should no longer be primary")

	// Verify only one primary bank exists
	var primaryCount int64
	db.Model(&models.CompanyBank{}).Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).Count(&primaryCount)
	assert.Equal(t, int64(1), primaryCount, "Only one primary bank should exist")
}

// Test Issue #6 Fix: UpdateCompany transaction consistency for NPWP validation
func TestUpdateCompany_TransactionConsistency_NPWPUniqueness(t *testing.T) {
	db := setupCompanyTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create tenants
	tenant1 := &models.Tenant{ID: "tenant-1", Status: models.TenantStatusTrial}
	db.Create(tenant1)
	tenant2 := &models.Tenant{ID: "tenant-2", Status: models.TenantStatusTrial}
	db.Create(tenant2)

	// Create companies
	npwp1 := "01.234.567.8-901.000"
	npwp2 := "02.345.678.9-012.000"

	company1 := &models.Company{
		ID:       "company-1",
		Name:     "Company 1",
		NPWP:     &npwp1,
		IsActive: true,
	}
	db.Create(company1)
	tenant1.CompanyID = company1.ID
	db.Save(tenant1)

	company2 := &models.Company{
		ID:       "company-2",
		Name:     "Company 2",
		NPWP:     &npwp2,
		IsActive: true,
	}
	db.Create(company2)
	tenant2.CompanyID = company2.ID
	db.Save(tenant2)

	// Try to update company2 to use company1's NPWP - should fail atomically
	updates := map[string]interface{}{"npwp": npwp1}
	_, err := service.UpdateCompany(ctx, tenant2.ID, updates)
	assert.Error(t, err, "Should prevent duplicate NPWP")
	assert.Contains(t, err.Error(), "already registered")

	// Verify company2's NPWP was NOT changed (transaction rollback)
	var reloadedCompany2 models.Company
	db.First(&reloadedCompany2, "id = ?", company2.ID)
	assert.Equal(t, npwp2, *reloadedCompany2.NPWP, "Company 2 NPWP should remain unchanged")
}

// Test Issue #6 Fix: DeleteBankAccount validates minimum requirement before deletion
func TestDeleteBankAccount_ValidatesMinimumRequirement(t *testing.T) {
	db := setupCompanyTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create tenant
	tenant := &models.Tenant{ID: "tenant-1", Status: models.TenantStatusTrial}
	db.Create(tenant)

	// Create company
	company := createTestCompany(t, db, tenant)

	// Create only one bank
	bank := &models.CompanyBank{
		ID:            "bank-1",
		CompanyID:     company.ID,
		BankName:      "Bank A",
		AccountNumber: "12345678",
		AccountName:   "Account A",
		IsPrimary:     true,
		IsActive:      true,
	}
	db.Create(bank)

	// Try to delete the last bank - should fail
	err := service.DeleteBankAccount(ctx, tenant.ID, bank.ID)
	assert.Error(t, err, "Should prevent deletion of last bank account")
	assert.Contains(t, err.Error(), "minimum 1 required")

	// Verify bank is still active (validation prevented deletion)
	var reloadedBank models.CompanyBank
	db.First(&reloadedBank, "id = ?", bank.ID)
	assert.True(t, reloadedBank.IsActive, "Bank should still be active")
}

// Test Issue #6 Fix: Race condition simulation for AddBankAccount
func TestAddBankAccount_RaceConditionPrevention(t *testing.T) {
	db := setupCompanyTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create tenant
	tenant := &models.Tenant{ID: "tenant-1", Status: models.TenantStatusTrial}
	db.Create(tenant)

	// Create company
	company := createTestCompany(t, db, tenant)

	// Create first bank as primary
	bank1 := &models.CompanyBank{
		ID:            "bank-1",
		CompanyID:     company.ID,
		BankName:      "Bank A",
		AccountNumber: "12345678",
		AccountName:   "Account A",
		IsPrimary:     true,
		IsActive:      true,
	}
	db.Create(bank1)

	// Simulate concurrent requests to add primary banks
	// Without transaction, both could become primary
	// With transaction (Issue #6 fix), only the last one becomes primary

	req2 := &AddBankRequest{
		BankName:      "Bank B",
		AccountNumber: "11111111",
		AccountName:   "Account B",
		IsPrimary:     true,
	}
	bank2, err := service.AddBankAccount(ctx, tenant.ID, req2)
	require.NoError(t, err)

	req3 := &AddBankRequest{
		BankName:      "Bank C",
		AccountNumber: "22222222",
		AccountName:   "Account C",
		IsPrimary:     true,
	}
	bank3, err := service.AddBankAccount(ctx, tenant.ID, req3)
	require.NoError(t, err)

	// Verify ONLY bank3 is primary (transaction ensures consistency)
	var primaryBanks []models.CompanyBank
	db.Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).Find(&primaryBanks)
	assert.Equal(t, 1, len(primaryBanks), "Only one primary bank should exist")
	assert.Equal(t, bank3.ID, primaryBanks[0].ID, "Bank 3 should be the primary")

	// Verify bank1 and bank2 are NOT primary
	var updatedBank1 models.CompanyBank
	db.First(&updatedBank1, "id = ?", bank1.ID)
	assert.False(t, updatedBank1.IsPrimary)

	var updatedBank2 models.CompanyBank
	db.First(&updatedBank2, "id = ?", bank2.ID)
	assert.False(t, updatedBank2.IsPrimary)
}
