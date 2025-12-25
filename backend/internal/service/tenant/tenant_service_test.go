package tenant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

func setupTenantTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&models.Company{}, &models.Tenant{}, &models.User{}, &models.UserTenant{})
	require.NoError(t, err)

	return db
}

func createTestTenant(t *testing.T, db *gorm.DB) *models.Tenant {
	// Create company first
	company := &models.Company{
		ID:       "company-1",
		Name:     "Test Company",
		IsActive: true,
	}
	err := db.Create(company).Error
	require.NoError(t, err)

	// Create tenant linked to company
	tenant := &models.Tenant{
		ID:        "tenant-1",
		CompanyID: company.ID,
		Status:    models.TenantStatusTrial,
	}
	err = db.Create(tenant).Error
	require.NoError(t, err)
	return tenant
}

func createTestUser(t *testing.T, db *gorm.DB, id string) *models.User {
	user := &models.User{
		ID:           id,
		Email:        id + "@test.com",
		Username:     id,
		PasswordHash: "hashed_password",
		FullName:     "User " + id,
		IsActive:     true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestUserTenant(t *testing.T, db *gorm.DB, userID, tenantID string, role models.UserRole) *models.UserTenant {
	userTenant := &models.UserTenant{
		ID:       userID + "-" + tenantID,
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		IsActive: true,
	}
	err := db.Create(userTenant).Error
	require.NoError(t, err)
	return userTenant
}

// Test Issue #6 Fix: RemoveUserFromTenant uses transaction for atomic admin count check + delete
func TestRemoveUserFromTenant_TransactionConsistency_AdminCountCheck(t *testing.T) {
	db := setupTenantTestDB(t)
	service := NewTenantService(db)
	ctx := context.Background()

	// Create tenant
	tenant := createTestTenant(t, db)

	// Create users
	user1 := createTestUser(t, db, "user-1")
	user2 := createTestUser(t, db, "user-2")
	user3 := createTestUser(t, db, "user-3")

	// Create user-tenant links: 2 admins + 1 staff
	userTenant1 := createTestUserTenant(t, db, user1.ID, tenant.ID, models.UserRoleAdmin)
	userTenant2 := createTestUserTenant(t, db, user2.ID, tenant.ID, models.UserRoleAdmin)
	userTenant3 := createTestUserTenant(t, db, user3.ID, tenant.ID, models.UserRoleStaff)

	// Remove one admin - should succeed (1 admin remains)
	err := service.RemoveUserFromTenant(ctx, tenant.ID, userTenant1.ID, nil)
	assert.NoError(t, err)

	// Verify user1 is deactivated
	var updatedUserTenant1 models.UserTenant
	db.First(&updatedUserTenant1, "id = ?", userTenant1.ID)
	assert.False(t, updatedUserTenant1.IsActive)

	// Try to remove the last admin - should fail
	err = service.RemoveUserFromTenant(ctx, tenant.ID, userTenant2.ID, nil)
	assert.Error(t, err, "Should prevent removal of last admin")
	assert.Contains(t, err.Error(), "last ADMIN")

	// Verify user2 is still active (transaction prevented deletion)
	var stillActiveUserTenant2 models.UserTenant
	db.First(&stillActiveUserTenant2, "id = ?", userTenant2.ID)
	assert.True(t, stillActiveUserTenant2.IsActive, "Last admin should remain active")

	// Can still remove staff user
	err = service.RemoveUserFromTenant(ctx, tenant.ID, userTenant3.ID, nil)
	assert.NoError(t, err)
}

// Test Issue #6 Fix: RemoveUserFromTenant prevents OWNER removal
func TestRemoveUserFromTenant_PreventOwnerRemoval(t *testing.T) {
	db := setupTenantTestDB(t)
	service := NewTenantService(db)
	ctx := context.Background()

	// Create tenant
	tenant := createTestTenant(t, db)

	// Create owner
	owner := createTestUser(t, db, "owner")
	ownerTenant := createTestUserTenant(t, db, owner.ID, tenant.ID, models.UserRoleOwner)

	// Try to remove owner - should fail
	err := service.RemoveUserFromTenant(ctx, tenant.ID, ownerTenant.ID, nil)
	assert.Error(t, err, "Should prevent removal of OWNER")
	assert.Contains(t, err.Error(), "OWNER")

	// Verify owner is still active
	var stillActiveOwner models.UserTenant
	db.First(&stillActiveOwner, "id = ?", ownerTenant.ID)
	assert.True(t, stillActiveOwner.IsActive, "OWNER should remain active")
}

// Test Issue #6 Fix: UpdateUserRole uses transaction for atomic admin count validation
func TestUpdateUserRole_TransactionConsistency_AdminCountCheck(t *testing.T) {
	db := setupTenantTestDB(t)
	service := NewTenantService(db)
	ctx := context.Background()

	// Create tenant
	tenant := createTestTenant(t, db)

	// Create users
	user1 := createTestUser(t, db, "user-1")
	user2 := createTestUser(t, db, "user-2")

	// Create user-tenant links: 2 admins
	userTenant1 := createTestUserTenant(t, db, user1.ID, tenant.ID, models.UserRoleAdmin)
	userTenant2 := createTestUserTenant(t, db, user2.ID, tenant.ID, models.UserRoleAdmin)

	// Change one admin to staff - should succeed (1 admin remains)
	err := service.UpdateUserRole(ctx, tenant.ID, userTenant1.ID, models.UserRoleStaff, nil)
	assert.NoError(t, err)

	// Verify role changed
	var updatedUserTenant1 models.UserTenant
	db.First(&updatedUserTenant1, "id = ?", userTenant1.ID)
	assert.Equal(t, models.UserRoleStaff, updatedUserTenant1.Role)

	// Try to change last admin to staff - should fail
	err = service.UpdateUserRole(ctx, tenant.ID, userTenant2.ID, models.UserRoleStaff, nil)
	assert.Error(t, err, "Should prevent changing last admin role")
	assert.Contains(t, err.Error(), "last ADMIN")

	// Verify role was NOT changed (transaction rollback)
	var stillAdminUserTenant2 models.UserTenant
	db.First(&stillAdminUserTenant2, "id = ?", userTenant2.ID)
	assert.Equal(t, models.UserRoleAdmin, stillAdminUserTenant2.Role, "Last admin role should remain unchanged")
}

// Test Issue #6 Fix: UpdateUserRole prevents OWNER role changes
func TestUpdateUserRole_PreventOwnerRoleChanges(t *testing.T) {
	db := setupTenantTestDB(t)
	service := NewTenantService(db)
	ctx := context.Background()

	// Create tenant
	tenant := createTestTenant(t, db)

	// Create owner and admin
	owner := createTestUser(t, db, "owner")
	ownerTenant := createTestUserTenant(t, db, owner.ID, tenant.ID, models.UserRoleOwner)

	admin := createTestUser(t, db, "admin")
	adminTenant := createTestUserTenant(t, db, admin.ID, tenant.ID, models.UserRoleAdmin)

	// Try to change owner to admin - should fail
	err := service.UpdateUserRole(ctx, tenant.ID, ownerTenant.ID, models.UserRoleAdmin, nil)
	assert.Error(t, err, "Should prevent changing OWNER role")
	assert.Contains(t, err.Error(), "OWNER")

	// Try to change admin to owner - should fail
	err = service.UpdateUserRole(ctx, tenant.ID, adminTenant.ID, models.UserRoleOwner, nil)
	assert.Error(t, err, "Should prevent assigning OWNER role")
	assert.Contains(t, err.Error(), "OWNER")

	// Verify roles unchanged
	var stillOwner models.UserTenant
	db.First(&stillOwner, "id = ?", ownerTenant.ID)
	assert.Equal(t, models.UserRoleOwner, stillOwner.Role)

	var stillAdmin models.UserTenant
	db.First(&stillAdmin, "id = ?", adminTenant.ID)
	assert.Equal(t, models.UserRoleAdmin, stillAdmin.Role)
}

// Test Issue #6 Fix: Race condition simulation for RemoveUserFromTenant
func TestRemoveUserFromTenant_RaceConditionPrevention(t *testing.T) {
	db := setupTenantTestDB(t)
	service := NewTenantService(db)
	ctx := context.Background()

	// Create tenant
	tenant := createTestTenant(t, db)

	// Create users
	user1 := createTestUser(t, db, "user-1")
	user2 := createTestUser(t, db, "user-2")

	// Create user-tenant links: 2 admins
	userTenant1 := createTestUserTenant(t, db, user1.ID, tenant.ID, models.UserRoleAdmin)
	userTenant2 := createTestUserTenant(t, db, user2.ID, tenant.ID, models.UserRoleAdmin)

	// Remove first admin
	err := service.RemoveUserFromTenant(ctx, tenant.ID, userTenant1.ID, nil)
	require.NoError(t, err)

	// Verify only 1 active admin remains
	var activeAdminCount int64
	db.Model(&models.UserTenant{}).
		Where("tenant_id = ? AND role = ? AND is_active = ?", tenant.ID, models.UserRoleAdmin, true).
		Count(&activeAdminCount)
	assert.Equal(t, int64(1), activeAdminCount)

	// Try to remove the last admin - should fail (transaction ensures consistency)
	err = service.RemoveUserFromTenant(ctx, tenant.ID, userTenant2.ID, nil)
	assert.Error(t, err)

	// Verify user2 is still active (transaction prevented deletion)
	var stillActiveUserTenant2 models.UserTenant
	db.First(&stillActiveUserTenant2, "id = ?", userTenant2.ID)
	assert.True(t, stillActiveUserTenant2.IsActive, "Last admin should remain active")
}

// Test Issue #6 Fix: AddUserToTenant reactivates deactivated users
func TestAddUserToTenant_ReactivatesDeactivatedUser(t *testing.T) {
	db := setupTenantTestDB(t)
	service := NewTenantService(db)
	ctx := context.Background()

	// Create tenant
	tenant := createTestTenant(t, db)

	// Create user
	user := createTestUser(t, db, "user-1")

	// Create deactivated user-tenant link
	userTenant := &models.UserTenant{
		ID:       user.ID + "-" + tenant.ID,
		UserID:   user.ID,
		TenantID: tenant.ID,
		Role:     models.UserRoleStaff,
		IsActive: false, // Deactivated
	}
	db.Create(userTenant)
	// Explicitly set to false (override GORM default)
	db.Model(userTenant).Update("is_active", false)

	// Verify user-tenant is inactive before attempting to reactivate
	var checkUserTenant models.UserTenant
	db.Where("user_id = ? AND tenant_id = ?", user.ID, tenant.ID).First(&checkUserTenant)
	t.Logf("Before reactivation: IsActive=%v, Role=%s", checkUserTenant.IsActive, checkUserTenant.Role)

	// Add user again with different role - should reactivate
	reactivated, err := service.AddUserToTenant(ctx, tenant.ID, user.ID, models.UserRoleAdmin, nil)
	require.NoError(t, err)
	assert.True(t, reactivated.IsActive)
	assert.Equal(t, models.UserRoleAdmin, reactivated.Role, "Role should be updated")

	// Verify only one user-tenant link exists
	var count int64
	db.Model(&models.UserTenant{}).Where("user_id = ? AND tenant_id = ?", user.ID, tenant.ID).Count(&count)
	assert.Equal(t, int64(1), count, "Should not create duplicate user-tenant link")
}
