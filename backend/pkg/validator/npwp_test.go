package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Company model for NPWP testing (minimal version)
type Company struct {
	ID       string `gorm:"type:varchar(255);primaryKey"`
	TenantID string `gorm:"type:varchar(255);not null;unique"`
	Name     string `gorm:"type:varchar(255);not null"`
	NPWP     *string `gorm:"type:varchar(50)"` // Pointer to allow NULL
	IsActive bool   `gorm:"default:true"`
}

func setupNPWPTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&Company{})
	require.NoError(t, err)

	// Create the partial unique index (simulating migration 000004)
	err = db.Exec(`
		CREATE UNIQUE INDEX idx_companies_npwp_unique
		ON companies(npwp)
		WHERE npwp IS NOT NULL
	`).Error
	require.NoError(t, err)

	return db
}

func stringPtr(s string) *string {
	return &s
}

func TestNPWP_UniqueConstraint_PreventsDuplicates(t *testing.T) {
	db := setupNPWPTestDB(t)

	npwp1 := stringPtr("01.234.567.8-901.000")

	// Create first company with NPWP
	company1 := Company{
		ID:       "company-1",
		TenantID: "tenant-1",
		Name:     "PT Test 1",
		NPWP:     npwp1,
		IsActive: true,
	}
	err := db.Create(&company1).Error
	require.NoError(t, err)

	// Try to create second company with SAME NPWP
	company2 := Company{
		ID:       "company-2",
		TenantID: "tenant-2",
		Name:     "PT Test 2",
		NPWP:     npwp1, // Duplicate NPWP
		IsActive: true,
	}
	err = db.Create(&company2).Error

	// Should fail with UNIQUE constraint error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed")
}

func TestNPWP_UniqueConstraint_AllowsMultipleNulls(t *testing.T) {
	db := setupNPWPTestDB(t)

	// Create first company with NULL NPWP
	company1 := Company{
		ID:       "company-1",
		TenantID: "tenant-1",
		Name:     "PT Test 1",
		NPWP:     nil, // NULL NPWP
		IsActive: true,
	}
	err := db.Create(&company1).Error
	require.NoError(t, err)

	// Create second company with NULL NPWP
	company2 := Company{
		ID:       "company-2",
		TenantID: "tenant-2",
		Name:     "PT Test 2",
		NPWP:     nil, // NULL NPWP
		IsActive: true,
	}
	err = db.Create(&company2).Error

	// Should succeed - multiple NULLs are allowed
	assert.NoError(t, err)

	// Verify both companies exist
	var count int64
	db.Model(&Company{}).Where("npwp IS NULL").Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestNPWP_UniqueConstraint_AllowsDifferentNPWPs(t *testing.T) {
	db := setupNPWPTestDB(t)

	npwp1 := stringPtr("01.234.567.8-901.000")
	npwp2 := stringPtr("02.345.678.9-012.000")

	// Create first company
	company1 := Company{
		ID:       "company-1",
		TenantID: "tenant-1",
		Name:     "PT Test 1",
		NPWP:     npwp1,
		IsActive: true,
	}
	err := db.Create(&company1).Error
	require.NoError(t, err)

	// Create second company with DIFFERENT NPWP
	company2 := Company{
		ID:       "company-2",
		TenantID: "tenant-2",
		Name:     "PT Test 2",
		NPWP:     npwp2,
		IsActive: true,
	}
	err = db.Create(&company2).Error

	// Should succeed - different NPWPs are allowed
	assert.NoError(t, err)

	// Verify both companies exist
	var count int64
	db.Model(&Company{}).Where("npwp IS NOT NULL").Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestNPWP_UniqueConstraint_UpdateToDuplicateFails(t *testing.T) {
	db := setupNPWPTestDB(t)

	npwp1 := stringPtr("01.234.567.8-901.000")
	npwp2 := stringPtr("02.345.678.9-012.000")

	// Create two companies with different NPWPs
	company1 := Company{
		ID:       "company-1",
		TenantID: "tenant-1",
		Name:     "PT Test 1",
		NPWP:     npwp1,
		IsActive: true,
	}
	err := db.Create(&company1).Error
	require.NoError(t, err)

	company2 := Company{
		ID:       "company-2",
		TenantID: "tenant-2",
		Name:     "PT Test 2",
		NPWP:     npwp2,
		IsActive: true,
	}
	err = db.Create(&company2).Error
	require.NoError(t, err)

	// Try to update company2's NPWP to match company1's NPWP
	err = db.Model(&Company{}).Where("id = ?", company2.ID).Update("npwp", npwp1).Error

	// Should fail with UNIQUE constraint error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UNIQUE constraint failed")
}

func TestNPWP_UniqueConstraint_UpdateToNullSucceeds(t *testing.T) {
	db := setupNPWPTestDB(t)

	npwp1 := stringPtr("01.234.567.8-901.000")

	// Create company with NPWP
	company1 := Company{
		ID:       "company-1",
		TenantID: "tenant-1",
		Name:     "PT Test 1",
		NPWP:     npwp1,
		IsActive: true,
	}
	err := db.Create(&company1).Error
	require.NoError(t, err)

	// Update NPWP to NULL
	err = db.Model(&Company{}).Where("id = ?", company1.ID).Update("npwp", nil).Error

	// Should succeed
	assert.NoError(t, err)

	// Verify NPWP is now NULL
	var updatedCompany Company
	db.First(&updatedCompany, "id = ?", company1.ID)
	assert.Nil(t, updatedCompany.NPWP)
}

func TestNPWP_UniqueConstraint_RealWorldScenario(t *testing.T) {
	db := setupNPWPTestDB(t)

	// Scenario: Two companies, one with NPWP, one without
	npwp1 := stringPtr("01.234.567.8-901.000")

	company1 := Company{
		ID:       "company-1",
		TenantID: "tenant-1",
		Name:     "PT Test 1 (has NPWP)",
		NPWP:     npwp1,
		IsActive: true,
	}
	err := db.Create(&company1).Error
	require.NoError(t, err)

	company2 := Company{
		ID:       "company-2",
		TenantID: "tenant-2",
		Name:     "PT Test 2 (no NPWP yet)",
		NPWP:     nil,
		IsActive: true,
	}
	err = db.Create(&company2).Error
	require.NoError(t, err)

	// Later, company2 tries to register with same NPWP
	err = db.Model(&Company{}).Where("id = ?", company2.ID).Update("npwp", npwp1).Error
	assert.Error(t, err, "Should prevent duplicate NPWP registration")

	// Company2 gets their own unique NPWP
	npwp2 := stringPtr("02.345.678.9-012.000")
	err = db.Model(&Company{}).Where("id = ?", company2.ID).Update("npwp", npwp2).Error
	assert.NoError(t, err, "Should allow unique NPWP registration")

	// Verify final state
	var companies []Company
	db.Find(&companies)
	assert.Equal(t, 2, len(companies))
	assert.NotEqual(t, companies[0].NPWP, companies[1].NPWP, "NPWPs should be different")
}
