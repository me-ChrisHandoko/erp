package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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

// filterNonEmptyFields extracts only fields that have meaningful values (non-empty, non-nil, non-zero)
// This ensures audit logs only record fields that were actually inputted by the user
func filterNonEmptyFields(data map[string]interface{}) []string {
	fields := []string{}
	for key, value := range data {
		if value == nil {
			continue
		}

		// Use reflection to handle pointer types
		v := reflect.ValueOf(value)

		// Dereference pointer if necessary
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
			value = v.Interface()
		}

		// Check based on underlying type
		switch val := value.(type) {
		case string:
			if val != "" && val != "0" && val != "0.00" {
				fields = append(fields, key)
			}
		case bool:
			// Only include boolean fields if true (explicitly set)
			if val {
				fields = append(fields, key)
			}
		case int, int8, int16, int32, int64:
			if v.Int() != 0 {
				fields = append(fields, key)
			}
		case uint, uint8, uint16, uint32, uint64:
			if v.Uint() != 0 {
				fields = append(fields, key)
			}
		case float32, float64:
			if v.Float() != 0 {
				fields = append(fields, key)
			}
		default:
			// For other types (arrays, structs, etc.), include if not zero value
			if !v.IsZero() {
				fields = append(fields, key)
			}
		}
	}
	return fields
}

// getFieldValue extracts a field value from either a map or struct by trying both struct field name and json tag
func getFieldValue(data interface{}, structFieldName, mapKey string) string {
	if data == nil {
		return ""
	}

	// If it's a map, try the mapKey
	if m, ok := data.(map[string]interface{}); ok {
		if val, ok := m[mapKey].(string); ok {
			return val
		}
		return ""
	}

	// If it's a struct, use reflection to get the field
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	field := v.FieldByName(structFieldName)
	if !field.IsValid() {
		return ""
	}

	if field.Kind() == reflect.String {
		return field.String()
	}

	return ""
}

// extractFieldNames extracts field names from either a map or struct for audit logging
func extractFieldNames(data interface{}) []string {
	if data == nil {
		return []string{}
	}

	// If it's a map, use filterNonEmptyFields
	if m, ok := data.(map[string]interface{}); ok {
		return filterNonEmptyFields(m)
	}

	// If it's a struct, extract field names using reflection
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return []string{}
	}

	t := v.Type()
	fields := []string{}
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		// Extract just the field name from json tag (before any comma)
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		fields = append(fields, jsonTag)
	}
	return fields
}

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
	// Use helper function to filter only non-empty values (handles pointer types correctly)
	createdFields := filterNonEmptyFields(productData)

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
		if !exists || !reflect.DeepEqual(oldValue, newValue) {
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
	// Use helper function to filter only non-empty values (handles pointer types correctly)
	createdFields := filterNonEmptyFields(customerData)

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
		if !exists || !reflect.DeepEqual(oldValue, newValue) {
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

// ==================== Product-Supplier Audit Methods ====================

// LogProductSupplierAdded logs when a supplier is linked to a product
func (s *AuditService) LogProductSupplierAdded(
	ctx context.Context,
	auditCtx *AuditContext,
	productID string,
	productSupplierID string,
	supplierData map[string]interface{},
) error {
	newValuesJSON, _ := json.Marshal(supplierData)
	newValuesStr := string(newValuesJSON)
	entityType := "PRODUCT"

	// Create human-readable notes
	supplierName := ""
	if name, ok := supplierData["supplier_name"].(string); ok {
		supplierName = name
	}
	supplierCode := ""
	if code, ok := supplierData["supplier_code"].(string); ok {
		supplierCode = code
	}

	notes := fmt.Sprintf("Added supplier %s (%s) to product", supplierName, supplierCode)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PRODUCT_UPDATED",
		EntityType: &entityType,
		EntityID:   &productID,
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

// LogProductSupplierUpdated logs when a product-supplier relationship is updated
func (s *AuditService) LogProductSupplierUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	productID string,
	productSupplierID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesJSON, _ := json.Marshal(newValues)
	newValuesStr := string(newValuesJSON)
	entityType := "PRODUCT"

	// Create human-readable notes with changed fields
	changedFields := []string{}
	for key := range newValues {
		if oldValues[key] != newValues[key] {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Updated supplier fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PRODUCT_UPDATED",
		EntityType: &entityType,
		EntityID:   &productID,
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

// LogProductSupplierDeleted logs when a supplier is removed from a product
func (s *AuditService) LogProductSupplierDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	productID string,
	productSupplierID string,
	supplierData map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(supplierData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "PRODUCT"

	// Create human-readable notes
	supplierName := ""
	if name, ok := supplierData["supplier_name"].(string); ok {
		supplierName = name
	}
	supplierCode := ""
	if code, ok := supplierData["supplier_code"].(string); ok {
		supplierCode = code
	}

	notes := fmt.Sprintf("Removed supplier %s (%s) from product", supplierName, supplierCode)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PRODUCT_UPDATED",
		EntityType: &entityType,
		EntityID:   &productID,
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

// LogProductSupplierOperationFailed logs when a product-supplier operation fails
func (s *AuditService) LogProductSupplierOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	productID string,
	errorMsg string,
) error {
	entityType := "PRODUCT"
	notes := fmt.Sprintf("Supplier operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &productID,
		Status:     StatusFailed,
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

// ==================== Supplier Audit Methods ====================

// LogSupplierCreated logs when a supplier is created
func (s *AuditService) LogSupplierCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	supplierID string,
	supplierData map[string]interface{},
) error {
	newValuesJSON, _ := json.Marshal(supplierData)
	newValuesStr := string(newValuesJSON)
	entityType := "SUPPLIER"

	// Create human-readable notes with created fields
	// Use helper function to filter only non-empty values (handles pointer types correctly)
	createdFields := filterNonEmptyFields(supplierData)

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "SUPPLIER_CREATED",
		EntityType: &entityType,
		EntityID:   &supplierID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogSupplierUpdated logs when a supplier is updated
func (s *AuditService) LogSupplierUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	supplierID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "SUPPLIER"

	// Create human-readable notes
	changedFields := []string{}
	for key := range newValues {
		if oldValues[key] != newValues[key] {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Updated fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "SUPPLIER_UPDATED",
		EntityType: &entityType,
		EntityID:   &supplierID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogSupplierDeleted logs when a supplier is deleted
func (s *AuditService) LogSupplierDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	supplierID string,
	supplierData map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(supplierData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "SUPPLIER"

	notes := fmt.Sprintf("Supplier deleted: %s", supplierID)
	if name, ok := supplierData["name"].(string); ok {
		notes = fmt.Sprintf("Supplier deleted: %s (ID: %s)", name, supplierID)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "SUPPLIER_DELETED",
		EntityType: &entityType,
		EntityID:   &supplierID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogSupplierOperationFailed logs when a supplier operation fails
func (s *AuditService) LogSupplierOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	supplierID string,
	errorMsg string,
) error {
	entityType := "SUPPLIER"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &supplierID,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ==================== Warehouse Audit Methods ====================

// LogWarehouseCreated logs when a warehouse is created
func (s *AuditService) LogWarehouseCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	warehouseID string,
	warehouseData map[string]interface{},
) error {
	newValuesJSON, _ := json.Marshal(warehouseData)
	newValuesStr := string(newValuesJSON)
	entityType := "WAREHOUSE"

	// Create human-readable notes with created fields
	// Use helper function to filter only non-empty values (handles pointer types correctly)
	createdFields := filterNonEmptyFields(warehouseData)

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "WAREHOUSE_CREATED",
		EntityType: &entityType,
		EntityID:   &warehouseID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogWarehouseUpdated logs when a warehouse is updated
func (s *AuditService) LogWarehouseUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	warehouseID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "WAREHOUSE"

	// Create human-readable notes
	changedFields := []string{}
	for key := range newValues {
		if oldValues[key] != newValues[key] {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Updated fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "WAREHOUSE_UPDATED",
		EntityType: &entityType,
		EntityID:   &warehouseID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogWarehouseDeleted logs when a warehouse is deleted
func (s *AuditService) LogWarehouseDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	warehouseID string,
	warehouseData map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(warehouseData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "WAREHOUSE"

	notes := fmt.Sprintf("Warehouse deleted: %s", warehouseID)
	if name, ok := warehouseData["name"].(string); ok {
		notes = fmt.Sprintf("Warehouse deleted: %s (ID: %s)", name, warehouseID)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "WAREHOUSE_DELETED",
		EntityType: &entityType,
		EntityID:   &warehouseID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogWarehouseOperationFailed logs when a warehouse operation fails
func (s *AuditService) LogWarehouseOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	warehouseID string,
	errorMsg string,
) error {
	entityType := "WAREHOUSE"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &warehouseID,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ==================== Stock Transfer Audit Methods ====================

// LogStockTransferCreated logs when a stock transfer is created
func (s *AuditService) LogStockTransferCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	transferID string,
	transferData map[string]interface{},
) error {
	newValuesJSON, _ := json.Marshal(transferData)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_TRANSFER"

	// Create human-readable notes with created fields
	createdFields := filterNonEmptyFields(transferData)

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_TRANSFER_CREATED",
		EntityType: &entityType,
		EntityID:   &transferID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockTransferUpdated logs when a stock transfer is updated
func (s *AuditService) LogStockTransferUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	transferID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_TRANSFER"

	// Create human-readable notes
	changedFields := []string{}
	for key := range newValues {
		if oldValues[key] != newValues[key] {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Updated fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_TRANSFER_UPDATED",
		EntityType: &entityType,
		EntityID:   &transferID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockTransferDeleted logs when a stock transfer is deleted
func (s *AuditService) LogStockTransferDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	transferID string,
	transferData map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(transferData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "STOCK_TRANSFER"

	notes := fmt.Sprintf("Stock transfer deleted: %s", transferID)
	if number, ok := transferData["transfer_number"].(string); ok {
		notes = fmt.Sprintf("Stock transfer deleted: %s (ID: %s)", number, transferID)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_TRANSFER_DELETED",
		EntityType: &entityType,
		EntityID:   &transferID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockTransferShipped logs when a stock transfer is shipped
func (s *AuditService) LogStockTransferShipped(
	ctx context.Context,
	auditCtx *AuditContext,
	transferID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_TRANSFER"

	notes := "Stock transfer shipped (DRAFT → SHIPPED)"
	if number, ok := newValues["transfer_number"].(string); ok {
		notes = fmt.Sprintf("Stock transfer %s shipped (DRAFT → SHIPPED)", number)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_TRANSFER_SHIPPED",
		EntityType: &entityType,
		EntityID:   &transferID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockTransferReceived logs when a stock transfer is received
func (s *AuditService) LogStockTransferReceived(
	ctx context.Context,
	auditCtx *AuditContext,
	transferID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_TRANSFER"

	notes := "Stock transfer received (SHIPPED → RECEIVED)"
	if number, ok := newValues["transfer_number"].(string); ok {
		notes = fmt.Sprintf("Stock transfer %s received (SHIPPED → RECEIVED)", number)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_TRANSFER_RECEIVED",
		EntityType: &entityType,
		EntityID:   &transferID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockTransferCancelled logs when a stock transfer is cancelled
func (s *AuditService) LogStockTransferCancelled(
	ctx context.Context,
	auditCtx *AuditContext,
	transferID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
	reason string,
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_TRANSFER"

	notes := fmt.Sprintf("Stock transfer cancelled (SHIPPED → CANCELLED). Reason: %s", reason)
	if number, ok := newValues["transfer_number"].(string); ok {
		notes = fmt.Sprintf("Stock transfer %s cancelled (SHIPPED → CANCELLED). Reason: %s", number, reason)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_TRANSFER_CANCELLED",
		EntityType: &entityType,
		EntityID:   &transferID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockTransferOperationFailed logs when a stock transfer operation fails
func (s *AuditService) LogStockTransferOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	transferID string,
	errorMsg string,
) error {
	entityType := "STOCK_TRANSFER"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &transferID,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// GetAuditLogsByEntityID retrieves audit logs for a specific entity
func (s *AuditService) GetAuditLogsByEntityID(
	ctx context.Context,
	tenantID string,
	entityType string,
	entityID string,
	limit int,
	offset int,
) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	// Set tenant context for GORM tenant isolation
	db := s.db.WithContext(ctx)
	if tenantID != "" {
		db = db.Set("tenant_id", tenantID)
	}

	query := db.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND entity_type = ? AND entity_id = ?", tenantID, entityType, entityID)

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

// ==================== Stock Opname Audit Methods ====================

// LogStockOpnameCreated logs when a stock opname is created
func (s *AuditService) LogStockOpnameCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	opnameID string,
	opnameData interface{},
) error {
	newValuesJSON, _ := json.Marshal(opnameData)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_OPNAME"

	// Create human-readable notes with created fields
	createdFields := extractFieldNames(opnameData)

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_OPNAME_CREATED",
		EntityType: &entityType,
		EntityID:   &opnameID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockOpnameUpdated logs when a stock opname is updated
func (s *AuditService) LogStockOpnameUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	opnameID string,
	oldValues interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_OPNAME"

	notes := "Stock opname updated"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_OPNAME_UPDATED",
		EntityType: &entityType,
		EntityID:   &opnameID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockOpnameDeleted logs when a stock opname is deleted
func (s *AuditService) LogStockOpnameDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	opnameID string,
	opnameData interface{},
) error {
	oldValuesJSON, _ := json.Marshal(opnameData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "STOCK_OPNAME"

	notes := fmt.Sprintf("Stock opname deleted: %s", opnameID)
	// Extract opname_number from struct or map
	opnameNumber := getFieldValue(opnameData, "OpnameNumber", "opname_number")
	if opnameNumber != "" {
		notes = fmt.Sprintf("Stock opname deleted: %s (ID: %s)", opnameNumber, opnameID)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_OPNAME_DELETED",
		EntityType: &entityType,
		EntityID:   &opnameID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockOpnameApproved logs when a stock opname is approved
func (s *AuditService) LogStockOpnameApproved(
	ctx context.Context,
	auditCtx *AuditContext,
	opnameID string,
	oldValues map[string]interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_OPNAME"

	notes := "Stock opname approved (COMPLETED → APPROVED)"
	opnameNumber := getFieldValue(newValues, "OpnameNumber", "opname_number")
	if opnameNumber != "" {
		notes = fmt.Sprintf("Stock opname %s approved (COMPLETED → APPROVED)", opnameNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_OPNAME_APPROVED",
		EntityType: &entityType,
		EntityID:   &opnameID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockOpnameBatchUpdated logs when multiple stock opname items are updated in a batch
func (s *AuditService) LogStockOpnameBatchUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	opnameID string,
	oldValues interface{},
	newValues interface{},
	itemCount int,
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_OPNAME"

	// Create human-readable notes
	notes := fmt.Sprintf("Stock opname items updated (%d items)", itemCount)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_OPNAME_UPDATED",
		EntityType: &entityType,
		EntityID:   &opnameID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockOpnameItemUpdated logs when a stock opname item is updated
func (s *AuditService) LogStockOpnameItemUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	opnameID string,
	itemID string,
	oldValues interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "STOCK_OPNAME"

	// Create human-readable notes
	notes := fmt.Sprintf("Stock opname item updated (item: %s)", itemID)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "STOCK_OPNAME_UPDATED",
		EntityType: &entityType,
		EntityID:   &opnameID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogStockOpnameOperationFailed logs when a stock opname operation fails
func (s *AuditService) LogStockOpnameOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	opnameID string,
	errorMsg string,
) error {
	entityType := "STOCK_OPNAME"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &opnameID,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ==================== Warehouse Stock Audit Methods ====================

// LogWarehouseStockUpdated logs when warehouse stock settings are updated
func (s *AuditService) LogWarehouseStockUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	stockID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "WAREHOUSE_STOCK"

	// Create human-readable notes with changed fields
	changedFields := []string{}
	for key, newValue := range newValues {
		oldValue, exists := oldValues[key]
		if !exists || !reflect.DeepEqual(oldValue, newValue) {
			changedFields = append(changedFields, key)
		}
	}

	notes := ""
	if len(changedFields) > 0 {
		notes = fmt.Sprintf("Updated fields: [%s]", strings.Join(changedFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "WAREHOUSE_STOCK_UPDATED",
		EntityType: &entityType,
		EntityID:   &stockID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInitialStockCreated logs when initial stock is created for a warehouse
func (s *AuditService) LogInitialStockCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	warehouseID string,
	stockData interface{},
	totalItems, createdStocks, updatedStocks int,
) error {
	newValuesJSON, _ := json.Marshal(stockData)
	newValuesStr := string(newValuesJSON)
	entityType := "WAREHOUSE_STOCK"

	notes := fmt.Sprintf("Initial stock setup for warehouse. Total items: %d, Created: %d, Updated: %d",
		totalItems,
		createdStocks,
		updatedStocks,
	)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INITIAL_STOCK_CREATED",
		EntityType: &entityType,
		EntityID:   &warehouseID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInitialStockOperationFailed logs when an initial stock operation fails
func (s *AuditService) LogInitialStockOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	warehouseID string,
	errorMsg string,
	requestData interface{},
) error {
	entityType := "WAREHOUSE_STOCK"
	notes := fmt.Sprintf("Initial stock setup failed: %s", errorMsg)

	var newValuesStr *string
	if requestData != nil {
		requestJSON, _ := json.Marshal(requestData)
		reqStr := string(requestJSON)
		newValuesStr = &reqStr
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INITIAL_STOCK_FAILED",
		EntityType: &entityType,
		EntityID:   &warehouseID,
		NewValues:  newValuesStr,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ==================== Inventory Adjustment Audit Methods ====================

// LogInventoryAdjustmentCreated logs when an inventory adjustment is created
func (s *AuditService) LogInventoryAdjustmentCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	adjustmentID string,
	adjustmentData interface{},
) error {
	newValuesJSON, _ := json.Marshal(adjustmentData)
	newValuesStr := string(newValuesJSON)
	entityType := "INVENTORY_ADJUSTMENT"

	// Create human-readable notes with created fields
	createdFields := extractFieldNames(adjustmentData)

	notes := ""
	if len(createdFields) > 0 {
		notes = fmt.Sprintf("Created fields: [%s]", strings.Join(createdFields, ", "))
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INVENTORY_ADJUSTMENT_CREATED",
		EntityType: &entityType,
		EntityID:   &adjustmentID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInventoryAdjustmentUpdated logs when an inventory adjustment is updated
func (s *AuditService) LogInventoryAdjustmentUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	adjustmentID string,
	oldValues interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "INVENTORY_ADJUSTMENT"

	notes := "Inventory adjustment updated"

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INVENTORY_ADJUSTMENT_UPDATED",
		EntityType: &entityType,
		EntityID:   &adjustmentID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInventoryAdjustmentDeleted logs when an inventory adjustment is soft deleted
func (s *AuditService) LogInventoryAdjustmentDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	adjustmentID string,
	adjustmentNumber string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesJSON, _ := json.Marshal(newValues)
	newValuesStr := string(newValuesJSON)
	entityType := "INVENTORY_ADJUSTMENT"

	notes := fmt.Sprintf("Inventory adjustment %s deleted", adjustmentNumber)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INVENTORY_ADJUSTMENT_DELETED",
		EntityType: &entityType,
		EntityID:   &adjustmentID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInventoryAdjustmentApproved logs when an inventory adjustment is approved
func (s *AuditService) LogInventoryAdjustmentApproved(
	ctx context.Context,
	auditCtx *AuditContext,
	adjustmentID string,
	adjustmentNumber string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "INVENTORY_ADJUSTMENT"

	notes := "Inventory adjustment approved (DRAFT → APPROVED)"
	if adjustmentNumber != "" {
		notes = fmt.Sprintf("Inventory adjustment %s approved (DRAFT → APPROVED)", adjustmentNumber)
	}
	// Append approval notes if provided
	if approvalNotes, ok := newValues["notes"].(string); ok && approvalNotes != "" {
		notes = fmt.Sprintf("%s. Notes: %s", notes, approvalNotes)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INVENTORY_ADJUSTMENT_APPROVED",
		EntityType: &entityType,
		EntityID:   &adjustmentID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInventoryAdjustmentCancelled logs when an inventory adjustment is cancelled
func (s *AuditService) LogInventoryAdjustmentCancelled(
	ctx context.Context,
	auditCtx *AuditContext,
	adjustmentID string,
	adjustmentNumber string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
	reason string,
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "INVENTORY_ADJUSTMENT"

	notes := fmt.Sprintf("Inventory adjustment cancelled (DRAFT → CANCELLED). Reason: %s", reason)
	if adjustmentNumber != "" {
		notes = fmt.Sprintf("Inventory adjustment %s cancelled (DRAFT → CANCELLED). Reason: %s", adjustmentNumber, reason)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "INVENTORY_ADJUSTMENT_CANCELLED",
		EntityType: &entityType,
		EntityID:   &adjustmentID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogInventoryAdjustmentOperationFailed logs when an inventory adjustment operation fails
func (s *AuditService) LogInventoryAdjustmentOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	adjustmentID string,
	errorMsg string,
) error {
	entityType := "INVENTORY_ADJUSTMENT"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &adjustmentID,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ==================== Purchase Order Audit Methods ====================

// LogPurchaseOrderCreated logs when a purchase order is created
func (s *AuditService) LogPurchaseOrderCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	purchaseOrderData interface{},
) error {
	newValuesJSON, _ := json.Marshal(purchaseOrderData)
	newValuesStr := string(newValuesJSON)
	entityType := "PURCHASE_ORDER"

	poNumber := getFieldValue(purchaseOrderData, "PONumber", "po_number")
	notes := "Purchase order created (→ DRAFT)"
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order %s created (→ DRAFT)", poNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_CREATED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderUpdated logs when a purchase order is updated
func (s *AuditService) LogPurchaseOrderUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	oldValues interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "PURCHASE_ORDER"

	poNumber := getFieldValue(newValues, "PONumber", "po_number")
	notes := "Purchase order updated"
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order %s updated", poNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_UPDATED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderDeleted logs when a purchase order is deleted
func (s *AuditService) LogPurchaseOrderDeleted(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	purchaseOrderData interface{},
) error {
	oldValuesJSON, _ := json.Marshal(purchaseOrderData)
	oldValuesStr := string(oldValuesJSON)
	entityType := "PURCHASE_ORDER"

	notes := fmt.Sprintf("Purchase order deleted: %s", purchaseOrderID)
	// Extract po_number from struct or map
	poNumber := getFieldValue(purchaseOrderData, "PONumber", "po_number")
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order deleted: %s (ID: %s)", poNumber, purchaseOrderID)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_DELETED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		OldValues:  &oldValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderConfirmed logs when a purchase order is confirmed (DRAFT -> CONFIRMED)
func (s *AuditService) LogPurchaseOrderConfirmed(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	oldValues map[string]interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "PURCHASE_ORDER"

	notes := "Purchase order confirmed (DRAFT → CONFIRMED)"
	poNumber := getFieldValue(newValues, "PONumber", "po_number")
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order %s confirmed (DRAFT → CONFIRMED)", poNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_CONFIRMED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderCompleted logs when a purchase order is completed (CONFIRMED -> COMPLETED)
func (s *AuditService) LogPurchaseOrderCompleted(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	oldValues map[string]interface{},
	newValues interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "PURCHASE_ORDER"

	notes := "Purchase order completed (CONFIRMED → COMPLETED)"
	poNumber := getFieldValue(newValues, "PONumber", "po_number")
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order %s completed (CONFIRMED → COMPLETED)", poNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_COMPLETED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderCancelled logs when a purchase order is cancelled
func (s *AuditService) LogPurchaseOrderCancelled(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	oldValues map[string]interface{},
	newValues interface{},
	reason string,
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "PURCHASE_ORDER"

	notes := fmt.Sprintf("Purchase order cancelled. Reason: %s", reason)
	poNumber := getFieldValue(newValues, "PONumber", "po_number")
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order %s cancelled. Reason: %s", poNumber, reason)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_CANCELLED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderShortClosed logs when a purchase order is short closed (SAP DCI model)
func (s *AuditService) LogPurchaseOrderShortClosed(
	ctx context.Context,
	auditCtx *AuditContext,
	purchaseOrderID string,
	oldValues map[string]interface{},
	newValues interface{},
	reason string,
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "PURCHASE_ORDER"

	notes := fmt.Sprintf("Purchase order short closed. Reason: %s", reason)
	poNumber := getFieldValue(newValues, "PONumber", "po_number")
	if poNumber != "" {
		notes = fmt.Sprintf("Purchase order %s short closed. Reason: %s", poNumber, reason)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "PURCHASE_ORDER_SHORT_CLOSED",
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogPurchaseOrderOperationFailed logs when a purchase order operation fails
func (s *AuditService) LogPurchaseOrderOperationFailed(
	ctx context.Context,
	auditCtx *AuditContext,
	action string,
	purchaseOrderID string,
	errorMsg string,
) error {
	entityType := "PURCHASE_ORDER"
	notes := fmt.Sprintf("Operation failed: %s", errorMsg)

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     action,
		EntityType: &entityType,
		EntityID:   &purchaseOrderID,
		Status:     StatusFailed,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// ============================================================================
// GOODS RECEIPT AUDIT LOGGING
// ============================================================================

// LogGoodsReceiptCreated logs when a new goods receipt is created from PO
func (s *AuditService) LogGoodsReceiptCreated(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	newValues interface{},
) error {
	newValuesJSON, _ := json.Marshal(newValues)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT"

	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	notes := "Goods receipt created (→ PENDING)"
	if grnNumber != "" {
		notes = fmt.Sprintf("Goods receipt %s created (→ PENDING)", grnNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_CREATED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogGoodsReceiptReceived logs when goods are received (PENDING → RECEIVED)
func (s *AuditService) LogGoodsReceiptReceived(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT"

	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	notes := "Goods receipt received (PENDING → RECEIVED)"
	if grnNumber != "" {
		notes = fmt.Sprintf("Goods receipt %s received (PENDING → RECEIVED)", grnNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_RECEIVED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogGoodsReceiptInspected logs when goods are inspected (RECEIVED → INSPECTED)
func (s *AuditService) LogGoodsReceiptInspected(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT"

	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	notes := "Goods receipt inspected (RECEIVED → INSPECTED)"
	if grnNumber != "" {
		notes = fmt.Sprintf("Goods receipt %s inspected (RECEIVED → INSPECTED)", grnNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_INSPECTED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogGoodsReceiptAccepted logs when goods are accepted (INSPECTED → ACCEPTED/PARTIAL)
func (s *AuditService) LogGoodsReceiptAccepted(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT"

	newStatus := "ACCEPTED"
	if status, ok := newValues["status"].(string); ok {
		newStatus = status
	}
	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	notes := fmt.Sprintf("Goods receipt accepted (INSPECTED → %s)", newStatus)
	if grnNumber != "" {
		notes = fmt.Sprintf("Goods receipt %s accepted (INSPECTED → %s)", grnNumber, newStatus)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_ACCEPTED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogGoodsReceiptRejected logs when goods are rejected (INSPECTED → REJECTED)
func (s *AuditService) LogGoodsReceiptRejected(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT"

	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	notes := "Goods receipt rejected (INSPECTED → REJECTED)"
	if grnNumber != "" {
		notes = fmt.Sprintf("Goods receipt %s rejected (INSPECTED → REJECTED)", grnNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_REJECTED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogGoodsReceiptDispositionUpdated logs when a rejection disposition is set/updated for a goods receipt item
func (s *AuditService) LogGoodsReceiptDispositionUpdated(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT_ITEM"

	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	disposition := getFieldValue(newValues, "RejectionDisposition", "rejection_disposition")
	notes := fmt.Sprintf("Rejection disposition set to %s", disposition)
	if grnNumber != "" {
		notes = fmt.Sprintf("Rejection disposition for GRN %s item set to %s", grnNumber, disposition)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_DISPOSITION_UPDATED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}

// LogGoodsReceiptDispositionResolved logs when a rejection disposition is marked as resolved
func (s *AuditService) LogGoodsReceiptDispositionResolved(
	ctx context.Context,
	auditCtx *AuditContext,
	goodsReceiptID string,
	oldValues map[string]interface{},
	newValues map[string]interface{},
) error {
	oldValuesJSON, _ := json.Marshal(oldValues)
	newValuesJSON, _ := json.Marshal(newValues)
	oldValuesStr := string(oldValuesJSON)
	newValuesStr := string(newValuesJSON)
	entityType := "GOODS_RECEIPT_ITEM"

	grnNumber := getFieldValue(newValues, "GRNNumber", "grn_number")
	disposition := getFieldValue(newValues, "RejectionDisposition", "rejection_disposition")
	notes := fmt.Sprintf("Rejection disposition %s resolved", disposition)
	if grnNumber != "" {
		notes = fmt.Sprintf("Rejection disposition %s for GRN %s resolved", disposition, grnNumber)
	}

	auditLog := &models.AuditLog{
		TenantID:   auditCtx.TenantID,
		CompanyID:  auditCtx.CompanyID,
		UserID:     auditCtx.UserID,
		RequestID:  auditCtx.RequestID,
		Action:     "GOODS_RECEIPT_DISPOSITION_RESOLVED",
		EntityType: &entityType,
		EntityID:   &goodsReceiptID,
		OldValues:  &oldValuesStr,
		NewValues:  &newValuesStr,
		Status:     StatusSuccess,
		IPAddress:  auditCtx.IPAddress,
		UserAgent:  auditCtx.UserAgent,
		Notes:      &notes,
	}

	db := s.db.WithContext(ctx)
	if auditCtx.TenantID != nil {
		db = db.Set("tenant_id", *auditCtx.TenantID)
	}
	return db.Create(auditLog).Error
}
