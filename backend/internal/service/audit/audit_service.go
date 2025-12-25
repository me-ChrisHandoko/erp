package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"backend/models"
)

// AuditService handles audit logging for sensitive operations
type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// AuditContext contains contextual information for audit logging
type AuditContext struct {
	TenantID  *string
	UserID    *string
	IPAddress *string
	UserAgent *string
}

// LogUserRoleChange logs when a user's role is changed
// Issue #7 Fix: Audit trail for RBAC operations
func (s *AuditService) LogUserRoleChange(ctx context.Context, auditCtx *AuditContext, userTenantID string, oldRole, newRole models.UserRole) error {
	changes := fmt.Sprintf("Role changed from %s to %s", oldRole, newRole)

	oldValuesJSON, _ := json.Marshal(map[string]interface{}{
		"role": oldRole,
	})
	newValuesJSON, _ := json.Marshal(map[string]interface{}{
		"role": newRole,
	})

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "USER_TENANT"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "USER_ROLE_CHANGED",
		EntityType: &entityType,
		EntityID:   &userTenantID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &changes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogUserAdded logs when a user is added to a tenant
func (s *AuditService) LogUserAdded(ctx context.Context, auditCtx *AuditContext, userTenantID string, role models.UserRole) error {
	notes := fmt.Sprintf("User added to tenant with role %s", role)

	newValuesJSON, _ := json.Marshal(map[string]interface{}{
		"role":      role,
		"is_active": true,
	})
	newValuesStr := string(newValuesJSON)
	entityType := "USER_TENANT"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "USER_ADDED_TO_TENANT",
		EntityType: &entityType,
		EntityID:   &userTenantID,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogUserRemoved logs when a user is removed from a tenant
func (s *AuditService) LogUserRemoved(ctx context.Context, auditCtx *AuditContext, userTenantID string, role models.UserRole) error {
	notes := fmt.Sprintf("User with role %s removed from tenant", role)

	oldValuesJSON, _ := json.Marshal(map[string]interface{}{
		"role":      role,
		"is_active": true,
	})
	newValuesJSON, _ := json.Marshal(map[string]interface{}{
		"is_active": false,
	})

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "USER_TENANT"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "USER_REMOVED_FROM_TENANT",
		EntityType: &entityType,
		EntityID:   &userTenantID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogUserReactivated logs when a deactivated user is reactivated
func (s *AuditService) LogUserReactivated(ctx context.Context, auditCtx *AuditContext, userTenantID string, oldRole, newRole models.UserRole) error {
	notes := fmt.Sprintf("Deactivated user reactivated with role changed from %s to %s", oldRole, newRole)

	oldValuesJSON, _ := json.Marshal(map[string]interface{}{
		"role":      oldRole,
		"is_active": false,
	})
	newValuesJSON, _ := json.Marshal(map[string]interface{}{
		"role":      newRole,
		"is_active": true,
	})

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "USER_TENANT"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "USER_REACTIVATED",
		EntityType: &entityType,
		EntityID:   &userTenantID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogBankAccountAdded logs when a bank account is added
func (s *AuditService) LogBankAccountAdded(ctx context.Context, auditCtx *AuditContext, bankID string, bankName string, isPrimary bool) error {
	notes := fmt.Sprintf("Bank account '%s' added (primary: %v)", bankName, isPrimary)

	newValuesJSON, _ := json.Marshal(map[string]interface{}{
		"bank_name":  bankName,
		"is_primary": isPrimary,
		"is_active":  true,
	})
	newValuesStr := string(newValuesJSON)
	entityType := "COMPANY_BANK"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "BANK_ACCOUNT_ADDED",
		EntityType: &entityType,
		EntityID:   &bankID,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogBankAccountUpdated logs when a bank account is updated
func (s *AuditService) LogBankAccountUpdated(ctx context.Context, auditCtx *AuditContext, bankID string, oldValues, newValues map[string]interface{}) error {
	notes := "Bank account updated"

	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "COMPANY_BANK"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "BANK_ACCOUNT_UPDATED",
		EntityType: &entityType,
		EntityID:   &bankID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogBankAccountDeleted logs when a bank account is deleted
func (s *AuditService) LogBankAccountDeleted(ctx context.Context, auditCtx *AuditContext, bankID string, bankName string) error {
	notes := fmt.Sprintf("Bank account '%s' deleted", bankName)

	oldValuesJSON, _ := json.Marshal(map[string]interface{}{
		"bank_name": bankName,
		"is_active": true,
	})
	newValuesJSON, _ := json.Marshal(map[string]interface{}{
		"is_active": false,
	})

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "COMPANY_BANK"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "BANK_ACCOUNT_DELETED",
		EntityType: &entityType,
		EntityID:   &bankID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// LogCompanyUpdated logs when company profile is updated
func (s *AuditService) LogCompanyUpdated(ctx context.Context, auditCtx *AuditContext, companyID string, oldValues, newValues map[string]interface{}) error {
	notes := "Company profile updated"

	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "COMPANY"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		UserID:     auditCtx.UserID,
		Action:     "COMPANY_UPDATED",
		EntityType: &entityType,
		EntityID:   &companyID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	return s.db.WithContext(ctx).Create(auditLog).Error
}

// GetAuditLogs retrieves audit logs with filtering
func (s *AuditService) GetAuditLogs(ctx context.Context, tenantID string, filters map[string]interface{}, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.db.WithContext(ctx).Model(&models.AuditLog{})

	// Tenant filter
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if entityType, ok := filters["entity_type"]; ok {
		query = query.Where("entity_type = ?", entityType)
	}
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get paginated results
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}
