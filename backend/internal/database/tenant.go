package database

import (
	"errors"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"backend/internal/config"
)

// Global tenant configuration
var tenantConfig *config.TenantIsolationConfig

// PRODUCTION-READY: Tables that should be EXCLUDED from tenant isolation
// These are cross-tenant mapping tables or system tables that inherently
// contain data across multiple tenants
var tenantIsolationExcludedTables = map[string]bool{
	// Auth & User Management Tables (user can belong to multiple tenants)
	"user_tenants":        true, // User → Tenant mapping (cross-tenant by design)
	"user_company_roles":  true, // User → Company mapping (cross-tenant by design)
	"refresh_tokens":      true, // Auth tokens are per-user, not per-tenant
	"login_attempts":      true, // Login tracking is per-user/IP, not per-tenant
	"email_verifications": true, // Email verification is per-user, not per-tenant
	"password_resets":     true, // Password reset is per-user, not per-tenant

	// System Tables (no tenant_id column anyway, but explicit for clarity)
	"users":         true, // Users exist across tenants
	"tenants":       true, // Tenant master table itself
	"subscriptions": true, // Subscription is per-tenant (already has tenant_id)
}

// isTableExcludedFromTenantIsolation checks if a table should skip tenant isolation
// This is a SECURITY-CRITICAL function - only add tables here if they are genuinely cross-tenant
func isTableExcludedFromTenantIsolation(tableName string) bool {
	return tenantIsolationExcludedTables[tableName]
}

// TenantScope automatically filters queries by tenant_id
// Usage: db.Scopes(TenantScope(tenantID)).Find(&products)
// This is the SECOND layer of defense (after GORM callbacks)
func TenantScope(tenantID string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// SetTenantSession sets tenant context in GORM session for callback enforcement
// This MUST be called before any tenant-scoped queries
// Enhanced version without PostgreSQL RLS dependency
func SetTenantSession(db *gorm.DB, tenantID string) *gorm.DB {
	// Set tenant_id in GORM session (for callbacks)
	// This is used by RegisterTenantCallbacks to auto-inject tenant filters
	return db.Set("tenant_id", tenantID)
}

// RegisterTenantCallbacks registers GORM callbacks for automatic tenant filtering
// This is the PRIMARY layer of defense with configurable strict mode
// Reference: BACKEND-IMPLEMENTATION.md - Enhanced Dual-Layer Isolation
func RegisterTenantCallbacks(db *gorm.DB, cfg *config.TenantIsolationConfig) {
	// Store config globally for callback access
	tenantConfig = cfg

	// Query callback - auto-inject WHERE tenant_id = ?
	db.Callback().Query().Before("gorm:query").Register("tenant:query", func(db *gorm.DB) {
		// PRODUCTION-READY: Skip tables that are excluded from tenant isolation
		// These are cross-tenant mapping tables (user_tenants, user_company_roles, etc.)
		if db.Statement != nil && db.Statement.Table != "" {
			if isTableExcludedFromTenantIsolation(db.Statement.Table) {
				return // Skip tenant isolation for this table
			}
		}

		// Skip if table doesn't have tenant_id column
		if !hasTenantIDColumn(db) {
			return
		}

		// Get tenant_id from session
		tenantID, exists := db.Get("tenant_id")
		if !exists {
			// Check bypass flag for system/admin operations
			bypass, _ := db.Get("bypass_tenant")
			if bypass == true && tenantConfig.AllowBypass {
				return // Allowed for system operations
			}

			// STRICT MODE: Error instead of silently allowing
			if tenantConfig.StrictMode {
				db.AddError(errors.New("TENANT_CONTEXT_REQUIRED: Cannot query without tenant context"))
				return
			}

			// Log warning in permissive mode
			if tenantConfig.LogWarnings {
				log.Printf("[TENANT WARNING] Query without tenant context - Table: %s, Operation: SELECT\n", db.Statement.Table)
			}
			return
		}

		// Auto-inject tenant_id filter
		// Only if WHERE clause doesn't already filter by tenant_id
		if !hasExistingTenantFilter(db) {
			db.Where("tenant_id = ?", tenantID)
		}
	})

	// Create callback - auto-set tenant_id
	db.Callback().Create().Before("gorm:create").Register("tenant:create", func(db *gorm.DB) {
		// PRODUCTION-READY: Skip tables that are excluded from tenant isolation
		if db.Statement != nil && db.Statement.Table != "" {
			if isTableExcludedFromTenantIsolation(db.Statement.Table) {
				return // Skip tenant isolation for this table
			}
		}

		// Skip if table doesn't have tenant_id column
		if !hasTenantIDColumn(db) {
			return
		}

		// Skip if Statement is not properly initialized (can happen with associations)
		if db.Statement.Schema == nil || db.Statement.Dest == nil {
			return
		}

		// Get tenant_id from session
		tenantID, exists := db.Get("tenant_id")
		if !exists {
			// ALWAYS ERROR on create without tenant context
			// This is critical security - no bypass allowed for create operations
			db.AddError(errors.New("TENANT_CONTEXT_REQUIRED: Cannot create without tenant context"))
			return
		}

		// Auto-set tenant_id field
		db.Statement.SetColumn("tenant_id", tenantID)
	})

	// Prevent tenant_id modification after creation
	db.Callback().Update().Before("gorm:update").Register("tenant:immutable", func(db *gorm.DB) {
		// PRODUCTION-READY: Skip tables that are excluded from tenant isolation
		if db.Statement != nil && db.Statement.Table != "" {
			if isTableExcludedFromTenantIsolation(db.Statement.Table) {
				return // Skip tenant isolation for this table
			}
		}

		// Skip if table doesn't have tenant_id column
		if !hasTenantIDColumn(db) {
			return
		}

		// Check if trying to modify tenant_id
		if db.Statement.Changed("tenant_id") {
			db.AddError(errors.New("FORBIDDEN: Cannot modify tenant_id after creation"))
			return
		}
	})

	// Update callback - ensure not updating other tenants' data
	db.Callback().Update().Before("gorm:update").Register("tenant:update", func(db *gorm.DB) {
		// PRODUCTION-READY: Skip tables that are excluded from tenant isolation
		if db.Statement != nil && db.Statement.Table != "" {
			if isTableExcludedFromTenantIsolation(db.Statement.Table) {
				return // Skip tenant isolation for this table
			}
		}

		// Skip if table doesn't have tenant_id column
		if !hasTenantIDColumn(db) {
			return
		}

		// Get tenant_id from session
		tenantID, exists := db.Get("tenant_id")
		if !exists {
			// Check bypass flag
			bypass, _ := db.Get("bypass_tenant")
			if bypass == true && tenantConfig.AllowBypass {
				return
			}

			// STRICT MODE: Error on update without tenant
			if tenantConfig.StrictMode {
				db.AddError(errors.New("TENANT_CONTEXT_REQUIRED: Cannot update without tenant context"))
				return
			}

			// Log warning
			if tenantConfig.LogWarnings {
				log.Printf("[TENANT WARNING] Update without tenant context - Table: %s\n", db.Statement.Table)
			}
			return
		}

		// Auto-inject tenant_id filter using simple Where()
		if !hasExistingTenantFilter(db) {
			db.Where("tenant_id = ?", tenantID)
		}
	})

	// Delete callback - ensure not deleting other tenants' data
	db.Callback().Delete().Before("gorm:delete").Register("tenant:delete", func(db *gorm.DB) {
		// PRODUCTION-READY: Skip tables that are excluded from tenant isolation
		if db.Statement != nil && db.Statement.Table != "" {
			if isTableExcludedFromTenantIsolation(db.Statement.Table) {
				return // Skip tenant isolation for this table
			}
		}

		// Skip if table doesn't have tenant_id column
		if !hasTenantIDColumn(db) {
			return
		}

		// Get tenant_id from session
		tenantID, exists := db.Get("tenant_id")
		if !exists {
			// Check bypass flag
			bypass, _ := db.Get("bypass_tenant")
			if bypass == true && tenantConfig.AllowBypass {
				return
			}

			// STRICT MODE: Error on delete without tenant
			if tenantConfig.StrictMode {
				db.AddError(errors.New("TENANT_CONTEXT_REQUIRED: Cannot delete without tenant context"))
				return
			}

			// Log warning
			if tenantConfig.LogWarnings {
				log.Printf("[TENANT WARNING] Delete without tenant context - Table: %s\n", db.Statement.Table)
			}
			return
		}

		// Auto-inject tenant_id filter using simple Where()
		if !hasExistingTenantFilter(db) {
			db.Where("tenant_id = ?", tenantID)
		}
	})
}

// hasTenantIDColumn checks if the current table has a tenant_id column
func hasTenantIDColumn(db *gorm.DB) bool {
	if db.Statement == nil || db.Statement.Schema == nil {
		return false
	}

	// Check if schema has tenant_id field
	for _, field := range db.Statement.Schema.Fields {
		if field.DBName == "tenant_id" {
			return true
		}
	}

	return false
}

// hasExistingTenantFilter checks if WHERE clause already filters by tenant_id
// This prevents double-filtering when Scopes() is used together with callbacks
// Optimized version using GORM clauses instead of string matching
func hasExistingTenantFilter(db *gorm.DB) bool {
	if db.Statement == nil || db.Statement.Clauses == nil {
		return false
	}

	// Check GORM clauses for tenant_id filter
	for name, c := range db.Statement.Clauses {
		// Check WHERE clauses
		if name == "WHERE" {
			if whereClause, ok := c.Expression.(clause.Where); ok {
				for _, expr := range whereClause.Exprs {
					if checkExprForTenantID(expr) {
						return true
					}
				}
			}
		}
	}

	return false
}

// checkExprForTenantID recursively checks if an expression filters by tenant_id
func checkExprForTenantID(expr clause.Expression) bool {
	switch e := expr.(type) {
	case clause.Eq:
		// Check equality: tenant_id = ?
		if col, ok := e.Column.(string); ok && col == "tenant_id" {
			return true
		}
		if col, ok := e.Column.(clause.Column); ok && col.Name == "tenant_id" {
			return true
		}
	case clause.AndConditions:
		// Check AND conditions recursively
		for _, subExpr := range e.Exprs {
			if checkExprForTenantID(subExpr) {
				return true
			}
		}
	case clause.OrConditions:
		// Check OR conditions recursively
		for _, subExpr := range e.Exprs {
			if checkExprForTenantID(subExpr) {
				return true
			}
		}
	case clause.Expr:
		// Check raw expression for tenant_id pattern
		sql := e.SQL
		return len(sql) > 0 && (
			containsSubstr(sql, "tenant_id = ?") ||
			containsSubstr(sql, "tenant_id=?") ||
			containsSubstr(sql, `"tenant_id" = ?`) ||
			containsSubstr(sql, "`tenant_id` = ?"))
	}
	return false
}

// containsSubstr is a simple substring check helper
func containsSubstr(s, substr string) bool {
	if len(substr) == 0 || len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
