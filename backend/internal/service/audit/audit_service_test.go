package audit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

func setupAuditTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(&models.AuditLog{})
	require.NoError(t, err)

	return db
}

func stringPtr(s string) *string {
	return &s
}

// Test Issue #7 Fix: Audit logging for user role change
func TestLogUserRoleChange_CreatesAuditLog(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	userID := "user-1"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	auditCtx := &AuditContext{
		TenantID:  &tenantID,
		UserID:    &userID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	// Log role change
	err := service.LogUserRoleChange(ctx, auditCtx, "user-tenant-1", models.UserRoleStaff, models.UserRoleAdmin)
	assert.NoError(t, err)

	// Verify audit log was created
	var auditLog models.AuditLog
	err = db.First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, tenantID, *auditLog.TenantID)
	assert.Equal(t, userID, *auditLog.UserID)
	assert.Equal(t, "USER_ROLE_CHANGED", auditLog.Action)
	assert.Equal(t, "USER_TENANT", *auditLog.EntityType)
	assert.Equal(t, "user-tenant-1", *auditLog.EntityID)
	assert.Contains(t, *auditLog.Notes, "STAFF")
	assert.Contains(t, *auditLog.Notes, "ADMIN")
	assert.Equal(t, ipAddress, *auditLog.IPAddress)
	assert.Equal(t, userAgent, *auditLog.UserAgent)
}

// Test Issue #7 Fix: Audit logging for user addition
func TestLogUserAdded_CreatesAuditLog(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	userID := "user-1"

	auditCtx := &AuditContext{
		TenantID: &tenantID,
		UserID:   &userID,
	}

	// Log user addition
	err := service.LogUserAdded(ctx, auditCtx, "user-tenant-1", models.UserRoleAdmin)
	assert.NoError(t, err)

	// Verify audit log was created
	var auditLog models.AuditLog
	err = db.First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, "USER_ADDED_TO_TENANT", auditLog.Action)
	assert.Equal(t, "USER_TENANT", *auditLog.EntityType)
	assert.Contains(t, *auditLog.Notes, "ADMIN")
	assert.Contains(t, *auditLog.NewValues, "ADMIN")
}

// Test Issue #7 Fix: Audit logging for user removal
func TestLogUserRemoved_CreatesAuditLog(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	auditCtx := &AuditContext{
		TenantID: &tenantID,
	}

	// Log user removal
	err := service.LogUserRemoved(ctx, auditCtx, "user-tenant-1", models.UserRoleStaff)
	assert.NoError(t, err)

	// Verify audit log was created
	var auditLog models.AuditLog
	err = db.First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, "USER_REMOVED_FROM_TENANT", auditLog.Action)
	assert.Contains(t, *auditLog.Notes, "removed")
	assert.Contains(t, *auditLog.OldValues, "true") // is_active was true
	assert.Contains(t, *auditLog.NewValues, "false") // is_active now false
}

// Test Issue #7 Fix: Audit logging for user reactivation
func TestLogUserReactivated_CreatesAuditLog(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	auditCtx := &AuditContext{
		TenantID: &tenantID,
	}

	// Log user reactivation with role change
	err := service.LogUserReactivated(ctx, auditCtx, "user-tenant-1", models.UserRoleStaff, models.UserRoleAdmin)
	assert.NoError(t, err)

	// Verify audit log was created
	var auditLog models.AuditLog
	err = db.First(&auditLog).Error
	require.NoError(t, err)

	assert.Equal(t, "USER_REACTIVATED", auditLog.Action)
	assert.Contains(t, *auditLog.Notes, "reactivated")
	assert.Contains(t, *auditLog.Notes, "STAFF")
	assert.Contains(t, *auditLog.Notes, "ADMIN")
	assert.Contains(t, *auditLog.OldValues, "false") // was inactive
	assert.Contains(t, *auditLog.NewValues, "true")  // now active
}

// Test Issue #7 Fix: Get audit logs with filtering
func TestGetAuditLogs_FiltersCorrectly(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	userID1 := "user-1"
	userID2 := "user-2"

	// Create multiple audit logs
	auditCtx1 := &AuditContext{TenantID: &tenantID, UserID: &userID1}
	service.LogUserRoleChange(ctx, auditCtx1, "ut-1", models.UserRoleStaff, models.UserRoleAdmin)

	auditCtx2 := &AuditContext{TenantID: &tenantID, UserID: &userID2}
	service.LogUserAdded(ctx, auditCtx2, "ut-2", models.UserRoleStaff)

	auditCtx3 := &AuditContext{TenantID: &tenantID, UserID: &userID1}
	service.LogUserRemoved(ctx, auditCtx3, "ut-3", models.UserRoleStaff)

	// Test: Get all logs for tenant
	logs, total, err := service.GetAuditLogs(ctx, tenantID, map[string]interface{}{}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Equal(t, 3, len(logs))

	// Test: Filter by action
	logs, total, err = service.GetAuditLogs(ctx, tenantID, map[string]interface{}{
		"action": "USER_ROLE_CHANGED",
	}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, "USER_ROLE_CHANGED", logs[0].Action)

	// Test: Filter by user_id
	logs, total, err = service.GetAuditLogs(ctx, tenantID, map[string]interface{}{
		"user_id": userID1,
	}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(logs))

	// Test: Filter by entity_type
	logs, total, err = service.GetAuditLogs(ctx, tenantID, map[string]interface{}{
		"entity_type": "USER_TENANT",
	}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
}

// Test Issue #7 Fix: Audit logs support pagination
func TestGetAuditLogs_Pagination(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	userID := "user-1"
	auditCtx := &AuditContext{TenantID: &tenantID, UserID: &userID}

	// Create 5 audit logs
	for i := 0; i < 5; i++ {
		service.LogUserRoleChange(ctx, auditCtx, "ut-"+string(rune('1'+i)), models.UserRoleStaff, models.UserRoleAdmin)
	}

	// Test: First page (limit 2, offset 0)
	logs, total, err := service.GetAuditLogs(ctx, tenantID, map[string]interface{}{}, 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 2, len(logs))

	// Test: Second page (limit 2, offset 2)
	logs, total, err = service.GetAuditLogs(ctx, tenantID, map[string]interface{}{}, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 2, len(logs))

	// Test: Last page (limit 2, offset 4)
	logs, total, err = service.GetAuditLogs(ctx, tenantID, map[string]interface{}{}, 2, 4)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 1, len(logs))
}

// Test Issue #7 Fix: Audit logs ordered by created_at DESC
func TestGetAuditLogs_OrderedByCreatedAt(t *testing.T) {
	db := setupAuditTestDB(t)
	service := NewAuditService(db)
	ctx := context.Background()

	tenantID := "tenant-1"
	userID := "user-1"
	auditCtx := &AuditContext{TenantID: &tenantID, UserID: &userID}

	// Create logs in sequence
	service.LogUserAdded(ctx, auditCtx, "ut-1", models.UserRoleStaff)
	service.LogUserRoleChange(ctx, auditCtx, "ut-1", models.UserRoleStaff, models.UserRoleAdmin)
	service.LogUserRemoved(ctx, auditCtx, "ut-1", models.UserRoleAdmin)

	// Get logs
	logs, _, err := service.GetAuditLogs(ctx, tenantID, map[string]interface{}{}, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(logs))

	// Verify order (most recent first)
	assert.Equal(t, "USER_REMOVED_FROM_TENANT", logs[0].Action)
	assert.Equal(t, "USER_ROLE_CHANGED", logs[1].Action)
	assert.Equal(t, "USER_ADDED_TO_TENANT", logs[2].Action)
}
