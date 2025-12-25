package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CompanyBank model for testing (minimal version)
type CompanyBank struct {
	ID        string `gorm:"type:varchar(255);primaryKey"`
	CompanyID string `gorm:"type:varchar(255);not null"`
	BankName  string `gorm:"type:varchar(100);not null"`
	IsPrimary bool   `gorm:"default:false"`
	IsActive  bool   `gorm:"default:true"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&CompanyBank{})
	require.NoError(t, err)

	return db
}

func createTestBanks(t *testing.T, db *gorm.DB, companyID string, count int) []CompanyBank {
	banks := make([]CompanyBank, count)
	for i := 0; i < count; i++ {
		bank := CompanyBank{
			ID:        generateID(companyID, i),
			CompanyID: companyID,
			BankName:  "Test Bank",
			IsPrimary: i == 0, // First bank is primary
			IsActive:  true,
		}
		err := db.Create(&bank).Error
		require.NoError(t, err)
		banks[i] = bank
	}
	return banks
}

func generateID(companyID string, i int) string {
	return companyID + "-bank-" + string(rune('a'+i))
}

func TestValidateMinimumBankAccount_WithMultipleBanks(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 3 banks
	banks := createTestBanks(t, db, companyID, 3)

	// Deleting one bank should be allowed (2 will remain)
	err := validator.ValidateMinimumBankAccount(context.Background(), companyID, banks[0].ID)
	assert.NoError(t, err)
}

func TestValidateMinimumBankAccount_WithTwoBanks(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 2 banks
	banks := createTestBanks(t, db, companyID, 2)

	// Deleting one bank should be allowed (1 will remain)
	err := validator.ValidateMinimumBankAccount(context.Background(), companyID, banks[0].ID)
	assert.NoError(t, err)
}

func TestValidateMinimumBankAccount_WithOneBank(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 1 bank
	banks := createTestBanks(t, db, companyID, 1)

	// Deleting the last bank should be BLOCKED
	err := validator.ValidateMinimumBankAccount(context.Background(), companyID, banks[0].ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete last bank account")
	assert.Contains(t, err.Error(), "minimum 1 required")
}

func TestValidateMinimumBankAccount_WithNoBanks(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// No banks exist - validation should fail
	err := validator.ValidateMinimumBankAccount(context.Background(), companyID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete last bank account")
}

func TestValidateMinimumBankAccount_WithInactiveBanks(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 2 banks
	banks := createTestBanks(t, db, companyID, 2)

	// Deactivate one bank (soft delete)
	db.Model(&CompanyBank{}).Where("id = ?", banks[1].ID).Update("is_active", false)

	// Now only 1 active bank remains
	// Deleting it should be BLOCKED
	err := validator.ValidateMinimumBankAccount(context.Background(), companyID, banks[0].ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete last bank account")
}

func TestValidateCompanyHasBankAccounts_Success(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 1 bank
	createTestBanks(t, db, companyID, 1)

	// Validation should pass
	err := validator.ValidateCompanyHasBankAccounts(context.Background(), companyID)
	assert.NoError(t, err)
}

func TestValidateCompanyHasBankAccounts_NoBanks(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// No banks exist
	err := validator.ValidateCompanyHasBankAccounts(context.Background(), companyID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have at least one active bank account")
}

func TestValidateCompanyHasBankAccounts_OnlyInactiveBanks(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 2 banks but deactivate both
	banks := createTestBanks(t, db, companyID, 2)
	db.Model(&CompanyBank{}).Where("company_id = ?", companyID).Update("is_active", false)

	// Validation should fail (no active banks)
	err := validator.ValidateCompanyHasBankAccounts(context.Background(), companyID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have at least one active bank account")

	// Verify banks exist but are inactive
	var count int64
	db.Model(&CompanyBank{}).Where("id IN ?", []string{banks[0].ID, banks[1].ID}).Count(&count)
	assert.Equal(t, int64(2), count) // Banks exist
}

func TestValidateUniquePrimaryBank_NoPrimaryExists(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 2 banks, both non-primary
	banks := createTestBanks(t, db, companyID, 2)
	db.Model(&CompanyBank{}).Where("company_id = ?", companyID).Update("is_primary", false)

	// Setting one as primary should be allowed
	err := validator.ValidateUniquePrimaryBank(context.Background(), companyID, banks[0].ID)
	assert.NoError(t, err)
}

func TestValidateUniquePrimaryBank_OnePrimaryExists(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 2 banks, first one is primary (from createTestBanks)
	banks := createTestBanks(t, db, companyID, 2)

	// Trying to set second bank as primary should FAIL
	// (first bank is already primary)
	err := validator.ValidateUniquePrimaryBank(context.Background(), companyID, banks[1].ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Another bank account is already set as primary")
}

func TestValidateUniquePrimaryBank_SameBankUpdate(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)
	companyID := "company-1"

	// Create 1 bank (primary)
	banks := createTestBanks(t, db, companyID, 1)

	// Updating the same bank to remain primary should be allowed
	err := validator.ValidateUniquePrimaryBank(context.Background(), companyID, banks[0].ID)
	assert.NoError(t, err)
}

func TestValidateMinimumBankAccount_DifferentCompanies(t *testing.T) {
	db := setupTestDB(t)
	validator := NewBusinessValidator(db)

	// Company 1 has 1 bank
	company1 := "company-1"
	banks1 := createTestBanks(t, db, company1, 1)

	// Company 2 has 3 banks
	company2 := "company-2"
	createTestBanks(t, db, company2, 3)

	// Deleting last bank from company 1 should be BLOCKED
	err := validator.ValidateMinimumBankAccount(context.Background(), company1, banks1[0].ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot delete last bank account")

	// But company 2 should still have 3 banks (not affected)
	var count int64
	db.Model(&CompanyBank{}).Where("company_id = ? AND is_active = ?", company2, true).Count(&count)
	assert.Equal(t, int64(3), count)
}
