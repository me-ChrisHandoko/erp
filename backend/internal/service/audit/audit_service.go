package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// Status constants for audit logging (MVP Phase 1)
const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusPartial = "PARTIAL"
)

// AuditContext contains contextual information for audit logging
type AuditContext struct {
	TenantID  *string
	CompanyID *string // MVP Phase 1: Multi-company filtering
	UserID    *string
	RequestID *string // MVP Phase 1: Transaction grouping
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

// ========================================
// PRODUCT AUDIT METHODS (MVP Phase 1)
// ========================================

// LogProductCreated logs when a product is created
func (s *AuditService) LogProductCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	productID string,
	productData map[string]interface{},
) error {
	newValuesJSON, _ := json.Marshal(productData)
	newValuesStr := string(newValuesJSON)
	entityType := "PRODUCT"

	// Create human-readable notes with created fields
	// Filter out empty/default values to show only inputted fields
	createdFields := []string{}
	for key, value := range productData {
		// Skip empty values and defaults
		switch v := value.(type) {
		case string:
			if v != "" && v != "0" && v != "0.00" {
				createdFields = append(createdFields, key)
			}
		case bool:
			// Only include boolean fields if they are true (explicitly set by user)
			// false is the default value and might not be intentional input
			if v {
				createdFields = append(createdFields, key)
			}
		default:
			// Include other types (numbers, etc.)
			createdFields = append(createdFields, key)
		}
	}

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID, // MVP
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID, // MVP
		Action:     "PRODUCT_CREATED",
		EntityType: &entityType,
		EntityID:   &productID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess, // MVP
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogProductUpdated logs when a product is updated
func (s *AuditService) LogProductUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	productID string,
	oldValues, newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "PRODUCT"

	// Create human-readable notes with changed fields
	// Auto-detect all changed fields by comparing old and new values
	changedFields := []string{}
	for key, newValue := range newValues {
		oldValue, exists := oldValues[key]
		if !exists || oldValue != newValue {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Changed fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID, // MVP
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID, // MVP
		Action:     "PRODUCT_UPDATED",
		EntityType: &entityType,
		EntityID:   &productID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess, // MVP
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogProductDeleted logs when a product is deleted (soft delete)
func (s *AuditService) LogProductDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	productID string,
	productData map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(productData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "PRODUCT"

	// Create human-readable notes
	notes := fmt.Sprintf("Product '%s' (code: %s) deactivated",
		productData["name"],
		productData["code"])

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID, // MVP
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID, // MVP
		Action:     "PRODUCT_DELETED",
		EntityType: &entityType,
		EntityID:   &productID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess, // MVP
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ============================================================================
// CUSTOMER AUDIT LOGS
// ============================================================================

// LogCustomerCreated logs when a customer is created
func (s *AuditService) LogCustomerCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	customerID string,
	customerData map[string]interface{},
) error {
	newValuesJSON, _ := json.Marshal(customerData)
	newValuesStr := string(newValuesJSON)
	entityType := "CUSTOMER"

	// Create human-readable notes with created fields
	// Filter out empty/default values to show only inputted fields
	createdFields := []string{}
	for key, value := range customerData {
		// Skip empty values and defaults
		switch v := value.(type) {
		case string:
			if v != "" && v != "0" && v != "0.00" {
				createdFields = append(createdFields, key)
			}
		case bool:
			// Only include boolean fields if they are true (explicitly set by user)
			if v {
				createdFields = append(createdFields, key)
			}
		case int:
			// Only include integer fields if they are not zero (default value)
			if v != 0 {
				createdFields = append(createdFields, key)
			}
		default:
			// Include other types (decimal, etc.)
			createdFields = append(createdFields, key)
		}
	}

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "CUSTOMER_CREATED",
		EntityType: &entityType,
		EntityID:   &customerID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogCustomerUpdated logs when a customer is updated
func (s *AuditService) LogCustomerUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	customerID string,
	oldValues, newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)

	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "CUSTOMER"

	// Create human-readable notes with changed fields
	// Auto-detect all changed fields by comparing old and new values
	changedFields := []string{}
	for key, newValue := range newValues {
		oldValue, exists := oldValues[key]
		if !exists || oldValue != newValue {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Changed fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "CUSTOMER_UPDATED",
		EntityType: &entityType,
		EntityID:   &customerID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogCustomerDeleted logs when a customer is deactivated (soft delete)
func (s *AuditService) LogCustomerDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	customerID string,
	customerData map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(customerData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "CUSTOMER"

	// Create human-readable notes
	notes := fmt.Sprintf("Customer '%s' (code: %s) deactivated",
		customerData["name"],
		customerData["code"])

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "CUSTOMER_DELETED",
		EntityType: &entityType,
		EntityID:   &customerID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogProductOperationFailed logs when a product operation fails
func (s *AuditService) LogProductOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	productID string,
	errorMsg string,
) error {
	entityType := "PRODUCT"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID, // MVP
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID, // MVP
		Action:     action,
		EntityType: &entityType,
		EntityID:   &productID,
		Status:     StatusFailed, // MVP
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}
