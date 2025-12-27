package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/service/tenant"
	"backend/internal/util"
	"backend/models"
	"backend/pkg/errors"
)

// CompanyUserHandler handles HTTP requests for company-scoped user management
// Uses CompanyContextMiddleware to ensure proper company isolation
type CompanyUserHandler struct {
	tenantService *tenant.TenantService
	config        *config.Config
}

// NewCompanyUserHandler creates a new company user handler
func NewCompanyUserHandler(
	tenantService *tenant.TenantService,
	config *config.Config,
) *CompanyUserHandler {
	return &CompanyUserHandler{
		tenantService: tenantService,
		config:        config,
	}
}

// ListCompanyUsers retrieves users for the active company
// GET /api/v1/company/users
// Requires: CompanyContextMiddleware (validates X-Company-ID header)
// Returns: Users who have UserCompanyRole for the active company
func (h *CompanyUserHandler) ListCompanyUsers(c *gin.Context) {
	// Get company context (set by CompanyContextMiddleware)
	companyCtx := util.NewCompanyContext()
	companyID, exists := companyCtx.GetCompanyID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Company context not found"),
		})
		return
	}

	tenantID, exists := companyCtx.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Tenant context not found"),
		})
		return
	}

	log.Printf("🔍 DEBUG [ListCompanyUsers]: CompanyID=%s, TenantID=%s", companyID, tenantID)

	// Parse filters from query parameters
	filters := dto.GetUsersFilters{}
	if role := c.Query("role"); role != "" {
		filters.Role = (*models.UserRole)(&role)
	}
	if isActive := c.Query("isActive"); isActive != "" {
		active := isActive == "true"
		filters.IsActive = &active
	}

	// Get company users from service
	users, err := h.tenantService.ListCompanyUsers(c.Request.Context(), companyID, tenantID, filters)
	if err != nil {
		log.Printf("❌ ERROR [ListCompanyUsers]: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err,
		})
		return
	}

	log.Printf("✅ SUCCESS [ListCompanyUsers]: Found %d users for company %s", len(users), companyID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
	})
}

// InviteCompanyUser invites a new user to the active company
// POST /api/v1/company/users/invite
// Requires: CompanyContextMiddleware + OWNER/ADMIN role
func (h *CompanyUserHandler) InviteCompanyUser(c *gin.Context) {
	// Get company context
	companyCtx := util.NewCompanyContext()
	companyID, exists := companyCtx.GetCompanyID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Company context not found"),
		})
		return
	}

	tenantID, exists := companyCtx.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Tenant context not found"),
		})
		return
	}

	// Parse request body (ShouldBindJSON validates using binding tags)
	var req dto.InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	log.Printf("🔍 DEBUG [InviteCompanyUser]: Email=%s, Role=%s, CompanyID=%s", req.Email, req.Role, companyID)

	// Invite user via service (creates user and UserCompanyRole)
	userTenant, err := h.tenantService.InviteUserToCompany(c.Request.Context(), tenantID, companyID, req)
	if err != nil {
		log.Printf("❌ ERROR [InviteCompanyUser]: %v", err)
		h.handleError(c, err)
		return
	}

	log.Printf("✅ SUCCESS [InviteCompanyUser]: User %s invited to company %s", req.Email, companyID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    userTenant,
	})
}

// UpdateCompanyUserRole updates user's role in the active company
// PUT /api/v1/company/users/:userTenantId/role
// Requires: CompanyContextMiddleware + OWNER/ADMIN role
func (h *CompanyUserHandler) UpdateCompanyUserRole(c *gin.Context) {
	// Get company context
	companyCtx := util.NewCompanyContext()
	companyID, exists := companyCtx.GetCompanyID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Company context not found"),
		})
		return
	}

	tenantID, exists := companyCtx.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Tenant context not found"),
		})
		return
	}

	// Get user ID from path
	userTenantID := c.Param("userTenantId")
	if userTenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("User ID is required"),
		})
		return
	}

	// Parse request body (ShouldBindJSON validates using binding tags)
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	log.Printf("🔍 DEBUG [UpdateCompanyUserRole]: UserID=%s, NewRole=%s, CompanyID=%s", userTenantID, req.Role, companyID)

	// Update role via service (updates UserCompanyRole for specific company)
	updatedUser, err := h.tenantService.UpdateUserRoleInCompany(c.Request.Context(), tenantID, companyID, userTenantID, models.UserRole(req.Role))
	if err != nil {
		log.Printf("❌ ERROR [UpdateCompanyUserRole]: %v", err)
		h.handleError(c, err)
		return
	}

	log.Printf("✅ SUCCESS [UpdateCompanyUserRole]: User %s role updated to %s in company %s", userTenantID, req.Role, companyID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedUser,
	})
}

// RemoveCompanyUser removes user from the active company
// DELETE /api/v1/company/users/:userTenantId
// Requires: CompanyContextMiddleware + OWNER/ADMIN role
// Note: Soft deletes UserCompanyRole, doesn't delete the user account
func (h *CompanyUserHandler) RemoveCompanyUser(c *gin.Context) {
	// Get company context
	companyCtx := util.NewCompanyContext()
	companyID, exists := companyCtx.GetCompanyID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Company context not found"),
		})
		return
	}

	tenantID, exists := companyCtx.GetTenantID(c)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("Tenant context not found"),
		})
		return
	}

	// Get user ID from path
	userTenantID := c.Param("userTenantId")
	if userTenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errors.NewBadRequestError("User ID is required"),
		})
		return
	}

	log.Printf("🔍 DEBUG [RemoveCompanyUser]: UserID=%s, CompanyID=%s", userTenantID, companyID)

	// Remove user from company via service (soft delete UserCompanyRole)
	err := h.tenantService.RemoveUserFromCompany(c.Request.Context(), tenantID, companyID, userTenantID)
	if err != nil {
		log.Printf("❌ ERROR [RemoveCompanyUser]: %v", err)
		h.handleError(c, err)
		return
	}

	log.Printf("✅ SUCCESS [RemoveCompanyUser]: User %s removed from company %s", userTenantID, companyID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "User removed from company successfully",
		},
	})
}

// handleValidationError handles validation errors
func (h *CompanyUserHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMessages := make([]string, 0)
		for _, fieldError := range validationErrors {
			errorMessages = append(errorMessages, fieldError.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Validation failed",
				"details": errorMessages,
			},
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": err.Error(),
		},
	})
}

// handleError handles service errors
func (h *CompanyUserHandler) handleError(c *gin.Context, err error) {
	// Check error type and return appropriate HTTP status
	errMsg := err.Error()
	statusCode := http.StatusInternalServerError

	if strings.Contains(errMsg, "not found") {
		statusCode = http.StatusNotFound
	} else if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
		statusCode = http.StatusConflict
	} else if strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "permission") {
		statusCode = http.StatusForbidden
	}

	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   err,
	})
}
