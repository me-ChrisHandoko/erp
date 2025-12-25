package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/internal/service/tenant"
	"backend/models"
	"backend/pkg/errors"
)

// TenantHandler handles tenant management endpoints
// Reference: 01-TENANT-COMPANY-SETUP.md lines 732-929
type TenantHandler struct {
	tenantService *tenant.TenantService
	cfg           *config.Config
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *tenant.TenantService, cfg *config.Config) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		cfg:           cfg,
	}
}

// GetTenantDetails retrieves complete tenant information
// GET /api/v1/tenant
// Requires authentication + tenant context
// Reference: 01-TENANT-COMPANY-SETUP.md lines 736-771
func (h *TenantHandler) GetTenantDetails(c *gin.Context) {
	// Get tenant ID from context (set by TenantContextMiddleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found"))
		return
	}

	// Call service
	details, err := h.tenantService.GetTenantDetails(c.Request.Context(), tenantID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    details,
	})
}

// ListTenantUsers retrieves all users in a tenant with optional filters
// GET /api/v1/tenant/users
// GET /api/v1/tenant/users?role=ADMIN&isActive=true
// Requires authentication + tenant context + OWNER/ADMIN role
// Reference: 01-TENANT-COMPANY-SETUP.md lines 774-820
func (h *TenantHandler) ListTenantUsers(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found"))
		return
	}

	// Parse query parameters
	var role *string
	var isActive *bool

	if roleParam := c.Query("role"); roleParam != "" {
		role = &roleParam
	}

	if isActiveParam := c.Query("isActive"); isActiveParam != "" {
		if isActiveParam == "true" {
			isActiveVal := true
			isActive = &isActiveVal
		} else if isActiveParam == "false" {
			isActiveVal := false
			isActive = &isActiveVal
		}
	}

	// Call service
	users, err := h.tenantService.ListTenantUsers(c.Request.Context(), tenantID.(string), role, isActive)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// InviteUser invites a new user to the tenant or links existing user
// POST /api/v1/tenant/users/invite
// Requires authentication + tenant context + OWNER/ADMIN role + CSRF
// Reference: 01-TENANT-COMPANY-SETUP.md lines 823-869
func (h *TenantHandler) InviteUser(c *gin.Context) {
	// Get tenant ID and user ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	// Parse request body
	var req dto.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Create audit context
	tenantIDStr := tenantID.(string)
	userIDStr := userID.(string)
	ipAddr := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	auditCtx := &audit.AuditContext{
		TenantID:  &tenantIDStr,
		UserID:    &userIDStr,
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
	}

	// Call service
	response, err := h.tenantService.InviteUser(c.Request.Context(), tenantID.(string), &req, auditCtx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateUserRole updates a user's role in the tenant
// PUT /api/v1/tenant/users/:userTenantId/role
// Requires authentication + tenant context + OWNER/ADMIN role + CSRF
// Reference: 01-TENANT-COMPANY-SETUP.md lines 872-904
func (h *TenantHandler) UpdateUserRole(c *gin.Context) {
	// Get tenant ID and user ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	// Get userTenantId from path parameter
	userTenantID := c.Param("userTenantId")
	if userTenantID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("User tenant ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Create audit context
	tenantIDStr := tenantID.(string)
	userIDStr := userID.(string)
	ipAddr := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	auditCtx := &audit.AuditContext{
		TenantID:  &tenantIDStr,
		UserID:    &userIDStr,
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
	}

	// Call service (using existing UpdateUserRole method)
	err := h.tenantService.UpdateUserRole(
		c.Request.Context(),
		tenantID.(string),
		userTenantID,
		models.UserRole(req.Role),
		auditCtx,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Fetch updated user tenant to return in response
	users, err := h.tenantService.ListTenantUsers(c.Request.Context(), tenantID.(string), nil, nil)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Find the updated user tenant
	var updatedUserTenant *dto.TenantUserInfo
	for i := range users {
		if users[i].ID == userTenantID {
			updatedUserTenant = &users[i]
			break
		}
	}

	if updatedUserTenant == nil {
		c.JSON(http.StatusNotFound, errors.NewNotFoundError("User tenant not found after update"))
		return
	}

	response := &dto.UpdateUserRoleResponse{
		ID:          updatedUserTenant.ID,
		TenantID:    updatedUserTenant.TenantID,
		Email:       updatedUserTenant.Email,
		Name:        updatedUserTenant.Name,
		Role:        updatedUserTenant.Role,
		IsActive:    updatedUserTenant.IsActive,
		CreatedAt:   updatedUserTenant.CreatedAt,
		UpdatedAt:   updatedUserTenant.UpdatedAt,
		LastLoginAt: updatedUserTenant.LastLoginAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// RemoveUser removes a user from the tenant (soft delete)
// DELETE /api/v1/tenant/users/:userTenantId
// Requires authentication + tenant context + OWNER/ADMIN role + CSRF
// Reference: 01-TENANT-COMPANY-SETUP.md lines 907-929
func (h *TenantHandler) RemoveUser(c *gin.Context) {
	// Get tenant ID and user ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
		return
	}

	// Get userTenantId from path parameter
	userTenantID := c.Param("userTenantId")
	if userTenantID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("User tenant ID is required"))
		return
	}

	// Create audit context
	tenantIDStr := tenantID.(string)
	userIDStr := userID.(string)
	ipAddr := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	auditCtx := &audit.AuditContext{
		TenantID:  &tenantIDStr,
		UserID:    &userIDStr,
		IPAddress: &ipAddr,
		UserAgent: &userAgent,
	}

	// Call service (using existing RemoveUserFromTenant method)
	err := h.tenantService.RemoveUserFromTenant(c.Request.Context(), tenantID.(string), userTenantID, auditCtx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User removed from tenant successfully",
	})
}

// Helper methods

func (h *TenantHandler) handleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *errors.AppError:
		c.JSON(e.StatusCode, e)
	default:
		c.JSON(http.StatusInternalServerError, errors.NewInternalError(err))
	}
}

func (h *TenantHandler) handleValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, errors.NewBadRequestError(err.Error()))
}
