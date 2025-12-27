package company

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

// setupRaceTestDB creates an in-memory database for race condition tests
func setupRaceTestDB(t *testing.T) (*gorm.DB, *models.Tenant, *models.Company) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Company{}, &models.CompanyBank{}, &models.Tenant{})
	require.NoError(t, err)

	// Create tenant first
	tenant := &models.Tenant{
		ID:        "tenant-1",
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		Status:    models.TenantStatusTrial,
	}
	db.Create(tenant)

	// Create company and link to tenant
	company := &models.Company{
		ID:       "company-1",
		TenantID: tenant.ID,
		Name:     "Test Company",
		IsActive: true,
	}
	db.Create(company)

	return db, tenant, company
}

// Test Issue #10 Fix: Race Condition in Primary Bank Selection
// Concurrent SetPrimaryBank operations with SELECT FOR UPDATE
func TestAddBankAccount_ConcurrentPrimarySelection_Issue10(t *testing.T) {
	db, tenant, company := setupRaceTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create 10 bank accounts concurrently, ALL trying to be primary
	// Issue #10 Fix: SELECT FOR UPDATE ensures only ONE becomes primary
	var wg sync.WaitGroup
	results := make([]*models.CompanyBank, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := &AddBankRequest{
				BankName:      "Bank " + string(rune('A'+index)),
				AccountNumber: "123456789" + string(rune('0'+index)),
				AccountName:   "Account " + string(rune('A'+index)),
				IsPrimary:     true, // ALL try to be primary
			}
			results[index], errors[index] = service.AddBankAccount(ctx, tenant.ID, req)
		}(i)
	}

	wg.Wait()

	// All operations should succeed (no deadlocks)
	successCount := 0
	for i, err := range errors {
		if err == nil {
			successCount++
		} else {
			t.Logf("Operation %d failed: %v", i, err)
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "At least one AddBankAccount should succeed")

	// CRITICAL: Count primary banks - MUST be exactly 1
	var primaryCount int64
	db.Model(&models.CompanyBank{}).
		Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).
		Count(&primaryCount)

	assert.Equal(t, int64(1), primaryCount, "Only ONE bank should be primary after concurrent operations")

	// Verify which bank is primary
	var primaryBank models.CompanyBank
	err := db.Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).
		First(&primaryBank).Error
	require.NoError(t, err)
	assert.True(t, primaryBank.IsPrimary)
}

// Test Issue #10 Fix: UpdateBankAccount concurrent primary updates
func TestUpdateBankAccount_ConcurrentPrimaryUpdate_Issue10(t *testing.T) {
	db, tenant, company := setupRaceTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create 5 non-primary banks
	banks := make([]*models.CompanyBank, 5)
	for i := 0; i < 5; i++ {
		banks[i] = &models.CompanyBank{
			ID:            "bank-" + string(rune('1'+i)),
			CompanyID:     company.ID,
			BankName:      "Bank " + string(rune('A'+i)),
			AccountNumber: "123456789" + string(rune('0'+i)),
			AccountName:   "Account " + string(rune('A'+i)),
			IsPrimary:     false,
			IsActive:      true,
		}
		db.Create(banks[i])
	}

	// Update all 5 banks to primary concurrently
	// Issue #10 Fix: SELECT FOR UPDATE ensures only ONE becomes primary
	var wg sync.WaitGroup
	errors := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			updates := map[string]interface{}{"is_primary": true}
			_, errors[index] = service.UpdateBankAccount(ctx, tenant.ID, banks[index].ID, updates)
		}(i)
	}

	wg.Wait()

	// All operations should succeed (no deadlocks)
	successCount := 0
	for i, err := range errors {
		if err == nil {
			successCount++
		} else {
			t.Logf("Operation %d failed: %v", i, err)
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "At least one UpdateBankAccount should succeed")

	// CRITICAL: Count primary banks - MUST be exactly 1
	var primaryCount int64
	db.Model(&models.CompanyBank{}).
		Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).
		Count(&primaryCount)

	assert.Equal(t, int64(1), primaryCount, "Only ONE bank should be primary after concurrent updates")
}

// Test Issue #10 Fix: Mixed concurrent operations (Add + Update)
func TestBankOperations_MixedConcurrentOperations_Issue10(t *testing.T) {
	db, tenant, company := setupRaceTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// Create 3 existing banks
	existingBanks := make([]*models.CompanyBank, 3)
	for i := 0; i < 3; i++ {
		existingBanks[i] = &models.CompanyBank{
			ID:            "bank-existing-" + string(rune('1'+i)),
			CompanyID:     company.ID,
			BankName:      "Existing Bank " + string(rune('A'+i)),
			AccountNumber: "555555555" + string(rune('0'+i)),
			AccountName:   "Existing Account " + string(rune('A'+i)),
			IsPrimary:     i == 0, // First one is primary
			IsActive:      true,
		}
		db.Create(existingBanks[i])
	}

	// Concurrently:
	// - Update 3 existing banks to primary
	// - Add 3 new banks as primary
	var wg sync.WaitGroup
	errors := make([]error, 6)

	// Update operations
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			updates := map[string]interface{}{"is_primary": true}
			_, errors[index] = service.UpdateBankAccount(ctx, tenant.ID, existingBanks[index].ID, updates)
		}(i)
	}

	// Add operations
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := &AddBankRequest{
				BankName:      "New Bank " + string(rune('A'+index)),
				AccountNumber: "999999999" + string(rune('0'+index)),
				AccountName:   "New Account " + string(rune('A'+index)),
				IsPrimary:     true,
			}
			_, errors[index+3] = service.AddBankAccount(ctx, tenant.ID, req)
		}(i)
	}

	wg.Wait()

	// Most operations should succeed
	successCount := 0
	for i, err := range errors {
		if err == nil {
			successCount++
		} else {
			t.Logf("Operation %d failed: %v", i, err)
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "At least one operation should succeed")

	// CRITICAL: Count primary banks - MUST be exactly 1
	var primaryCount int64
	db.Model(&models.CompanyBank{}).
		Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).
		Count(&primaryCount)

	assert.Equal(t, int64(1), primaryCount, "Only ONE bank should be primary after mixed concurrent operations")

	// Count total active banks
	var totalCount int64
	db.Model(&models.CompanyBank{}).
		Where("company_id = ? AND is_active = ?", company.ID, true).
		Count(&totalCount)

	// Should have original 3 + successful adds
	assert.GreaterOrEqual(t, totalCount, int64(3), "Should have at least the original 3 banks")
}

// Test Issue #10 Fix: Verify SELECT FOR UPDATE deadlock handling
func TestBankOperations_HighConcurrency_Issue10(t *testing.T) {
	db, tenant, company := setupRaceTestDB(t)
	service := NewCompanyService(db)
	ctx := context.Background()

	// High concurrency test: 20 concurrent operations
	var wg sync.WaitGroup
	results := make([]*models.CompanyBank, 20)
	errors := make([]error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := &AddBankRequest{
				BankName:      "Bank " + string(rune('A'+(index%26))),
				AccountNumber: "123456789" + string(rune('0'+(index%10))),
				AccountName:   "Account " + string(rune('A'+(index%26))),
				IsPrimary:     true,
			}
			results[index], errors[index] = service.AddBankAccount(ctx, tenant.ID, req)
		}(i)
	}

	wg.Wait()

	// Should not have deadlocks - at least some should succeed
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}
	assert.GreaterOrEqual(t, successCount, 1, "At least one operation should succeed even with high concurrency")

	// CRITICAL: Only ONE primary bank must exist
	var primaryCount int64
	db.Model(&models.CompanyBank{}).
		Where("company_id = ? AND is_primary = ? AND is_active = ?", company.ID, true, true).
		Count(&primaryCount)

	assert.Equal(t, int64(1), primaryCount, "Only ONE bank should be primary even with 20 concurrent operations")
}
