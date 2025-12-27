package util

import (
	"github.com/gin-gonic/gin"
)

// Context keys for company context
const (
	// CompanyIDKey stores the active company ID in context
	CompanyIDKey = "company_id"

	// TenantIDKey stores the tenant ID in context
	TenantIDKey = "tenant_id"

	// UserIDKey stores the user ID in context
	UserIDKey = "user_id"

	// CompanyAccessKey stores the company access information
	CompanyAccessKey = "company_access"

	// UserRoleKey stores the user's role for the active company
	UserRoleKey = "user_role"
)

// CompanyContext provides helper methods for managing company context
type CompanyContext struct{}

// NewCompanyContext creates a new CompanyContext instance
func NewCompanyContext() *CompanyContext {
	return &CompanyContext{}
}

// SetCompanyID sets the company ID in the context
func (cc *CompanyContext) SetCompanyID(c *gin.Context, companyID string) {
	c.Set(CompanyIDKey, companyID)
}

// GetCompanyID retrieves the company ID from the context
func (cc *CompanyContext) GetCompanyID(c *gin.Context) (string, bool) {
	companyID, exists := c.Get(CompanyIDKey)
	if !exists {
		return "", false
	}

	companyIDStr, ok := companyID.(string)
	if !ok {
		return "", false
	}

	return companyIDStr, true
}

// MustGetCompanyID retrieves the company ID from context or panics
func (cc *CompanyContext) MustGetCompanyID(c *gin.Context) string {
	companyID, exists := cc.GetCompanyID(c)
	if !exists {
		panic("company ID not found in context")
	}
	return companyID
}

// SetTenantID sets the tenant ID in the context
func (cc *CompanyContext) SetTenantID(c *gin.Context, tenantID string) {
	c.Set(TenantIDKey, tenantID)
}

// GetTenantID retrieves the tenant ID from the context
func (cc *CompanyContext) GetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get(TenantIDKey)
	if !exists {
		return "", false
	}

	tenantIDStr, ok := tenantID.(string)
	if !ok {
		return "", false
	}

	return tenantIDStr, true
}

// MustGetTenantID retrieves the tenant ID from context or panics
func (cc *CompanyContext) MustGetTenantID(c *gin.Context) string {
	tenantID, exists := cc.GetTenantID(c)
	if !exists {
		panic("tenant ID not found in context")
	}
	return tenantID
}

// SetUserID sets the user ID in the context
func (cc *CompanyContext) SetUserID(c *gin.Context, userID string) {
	c.Set(UserIDKey, userID)
}

// GetUserID retrieves the user ID from the context
func (cc *CompanyContext) GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}

	return userIDStr, true
}

// MustGetUserID retrieves the user ID from context or panics
func (cc *CompanyContext) MustGetUserID(c *gin.Context) string {
	userID, exists := cc.GetUserID(c)
	if !exists {
		panic("user ID not found in context")
	}
	return userID
}

// SetUserRole sets the user's role for the active company
func (cc *CompanyContext) SetUserRole(c *gin.Context, role string) {
	c.Set(UserRoleKey, role)
}

// GetUserRole retrieves the user's role from the context
func (cc *CompanyContext) GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", false
	}

	return roleStr, true
}

// SetCompanyAccess sets the company access information in the context
func (cc *CompanyContext) SetCompanyAccess(c *gin.Context, access interface{}) {
	c.Set(CompanyAccessKey, access)
}

// GetCompanyAccess retrieves the company access information from the context
func (cc *CompanyContext) GetCompanyAccess(c *gin.Context) (interface{}, bool) {
	access, exists := c.Get(CompanyAccessKey)
	return access, exists
}

// Helper function to get company ID from context (backward compatibility)
func GetCompanyIDFromContext(c *gin.Context) (string, bool) {
	cc := NewCompanyContext()
	return cc.GetCompanyID(c)
}

// Helper function to get tenant ID from context (backward compatibility)
func GetTenantIDFromContext(c *gin.Context) (string, bool) {
	cc := NewCompanyContext()
	return cc.GetTenantID(c)
}

// Helper function to get user ID from context (backward compatibility)
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	cc := NewCompanyContext()
	return cc.GetUserID(c)
}
