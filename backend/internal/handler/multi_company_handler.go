package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/company"
	"backend/internal/service/permission"
	"backend/internal/util"
	"backend/models"
	"backend/pkg/errors"
	customValidator "backend/pkg/validator"
)

// MultiCompanyHandler handles HTTP requests for multi-company management
// Reference: multi-company-architecture-analysis.md - PHASE 3
type MultiCompanyHandler struct {
	multiCompanyService *company.MultiCompanyService
	permissionService   *permission.PermissionService
	validator           *customValidator.Validator
}

// NewMultiCompanyHandler creates a new multi-company handler
func NewMultiCompanyHandler(
	multiCompanyService *company.MultiCompanyService,
	permissionService *permission.PermissionService,
) *MultiCompanyHandler {
	return &MultiCompanyHandler{
		multiCompanyService: multiCompanyService,
		permissionService:   permissionService,
		validator:           customValidator.New(),
	}
}

// ListCompanies retrieves all accessible companies for current user
// GET /api/v1/companies
// Returns companies based on user's Tier 1 (all tenant companies) or Tier 2 (specific companies) access
func (h *MultiCompanyHandler) ListCompanies(c *gin.Context) {
	// Get user ID from context (set by JWTAuthMiddleware)
	companyCtx := util.NewCompanyContext()
	userID := companyCtx.MustGetUserID(c)

	log.Printf("🔍 DEBUG [ListCompanies]: User ID: %s", userID)

	// Get accessible companies via service
	companies, err := h.multiCompanyService.GetCompaniesByUserID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("❌ ERROR [ListCompanies]: Failed to get companies: %v", err)
		h.handleError(c, err)
		return
	}

	log.Printf("✅ DEBUG [ListCompanies]: Found %d companies", len(companies))

	// Debug: Print company IDs and roles
	for i, company := range companies {
		log.Printf("   Company %d: ID=%s, Name=%s", i, company.ID, company.Name)
	}

	// Get user's tenant role from database
	// This is needed because companyCtx.GetUserRole(c) doesn't work reliably
	var userTenantRole string
	if len(companies) > 0 {
		// Get tenant ID from first company (all companies belong to same tenant)
		tenantID := companies[0].TenantID

		// Check if user has Tier 1 access (OWNER or TENANT_ADMIN)
		var userTenant models.UserTenant
		err := h.multiCompanyService.GetDB().Where(
			"user_id = ? AND tenant_id = ? AND is_active = ? AND role IN ?",
			userID, tenantID, true, []string{string(models.UserRoleOwner), string(models.UserRoleTenantAdmin)},
		).First(&userTenant).Error

		if err == nil {
			// User has Tier 1 access
			userTenantRole = string(userTenant.Role)
			log.Printf("   ✅ User has Tier 1 access: %s", userTenantRole)
		} else {
			log.Printf("   ℹ️  User has Tier 2 access only")
		}
	}

	// Get user roles for each company
	userCompanyRoles, err := h.permissionService.GetUserCompanyRoles(c.Request.Context(), userID)
	if err != nil {
		// Log error but continue with basic company list
		userCompanyRoles = []models.UserCompanyRole{}
	}

	// Build role map for quick lookup
	roleMap := make(map[string]string)
	for _, ucr := range userCompanyRoles {
		roleMap[ucr.CompanyID] = string(ucr.Role)
	}

	// Map to response DTOs
	response := make([]dto.CompanyListResponse, 0, len(companies))
	for _, company := range companies {
		companyResp := dto.CompanyListResponse{
			ID:         company.ID,
			TenantID:   company.TenantID,
			Name:       company.Name,
			LegalName:  company.LegalName,
			City:       company.City,
			Province:   company.Province,
			IsPKP:      company.IsPKP,
			PPNRate:    company.PPNRate.InexactFloat64(),
			IsActive:   company.IsActive,
			AccessTier: 2, // Default to company-level access
		}

		// Set optional fields
		if company.NPWP != nil {
			companyResp.NPWP = *company.NPWP
		}
		if company.LogoURL != nil {
			companyResp.LogoURL = *company.LogoURL
		}

		// Set user role based on access tier
		if userTenantRole != "" {
			// Tier 1: Use tenant-level role (OWNER or TENANT_ADMIN)
			companyResp.AccessTier = 1
			companyResp.UserRole = userTenantRole
		} else if role, exists := roleMap[company.ID]; exists {
			// Tier 2: Use company-specific role
			companyResp.AccessTier = 2
			companyResp.UserRole = role
		}

		response = append(response, companyResp)
	}

	// Debug: Print final response
	for i, resp := range response {
		log.Printf("   Response %d: ID=%s, Role=%s, AccessTier=%d", i, resp.ID, resp.UserRole, resp.AccessTier)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateCompany creates a new company (OWNER only)
// POST /api/v1/companies
// Requires Tier 1 access (OWNER or TENANT_ADMIN)
func (h *MultiCompanyHandler) CreateCompany(c *gin.Context) {
	// Get tenant ID from context
	companyCtx := util.NewCompanyContext()
	tenantID := companyCtx.MustGetTenantID(c)

	var req dto.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Build service request
	serviceReq := &company.CreateCompanyRequest{
		Name:       req.Name,
		LegalName:  getStringValue(req.LegalName),
		EntityType: "CV", // Default to CV, TODO: get from request
		Address:    getStringValue(req.Address),
		City:       getStringValue(req.City),
		Province:   getStringValue(req.Province),
		PostalCode: req.PostalCode,
		Phone:      getStringValue(req.Phone),
		Email:      getStringValue(req.Email),
		Website:    req.Website,
		NPWP:       req.NPWP,
		NIB:        req.NIB,
		IsPKP:      getBoolValue(req.IsPKP),
	}

	// Call service
	newCompany, err := h.multiCompanyService.CreateCompany(c.Request.Context(), tenantID, serviceReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTO
	response := h.mapCompanyToDetailResponse(newCompany, tenantID, "", 1)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Company created successfully",
	})
}

// GetCompany retrieves a single company by ID
// GET /api/v1/companies/:id
// User must have access to this company
func (h *MultiCompanyHandler) GetCompany(c *gin.Context) {
	// Get company ID from URL
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company ID is required"))
		return
	}

	// Get company via service
	companyData, err := h.multiCompanyService.GetCompanyByID(c.Request.Context(), companyID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Get user's access information
	companyCtx := util.NewCompanyContext()
	userID := companyCtx.MustGetUserID(c)

	accessInfo, err := h.multiCompanyService.CheckUserCompanyAccess(c.Request.Context(), userID, companyID)
	if err != nil || !accessInfo.HasAccess {
		c.JSON(http.StatusForbidden, errors.NewAuthorizationError("You don't have access to this company"))
		return
	}

	// Map to response DTO
	response := h.mapCompanyToDetailResponse(companyData, accessInfo.TenantID, string(accessInfo.Role), accessInfo.AccessTier)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateCompany updates company information
// PATCH /api/v1/companies/:id
// Requires ADMIN role or higher
func (h *MultiCompanyHandler) UpdateCompany(c *gin.Context) {
	// Get company ID from URL
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company ID is required"))
		return
	}

	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Convert DTO to map for service
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.LegalName != nil {
		updates["legal_name"] = *req.LegalName
	}
	if req.NPWP != nil {
		updates["npwp"] = *req.NPWP
	}
	if req.NIB != nil {
		updates["nib"] = *req.NIB
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.Province != nil {
		updates["province"] = *req.Province
	}
	if req.PostalCode != nil {
		updates["postal_code"] = *req.PostalCode
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}
	if req.LogoURL != nil {
		updates["logo_url"] = *req.LogoURL
	}
	if req.IsPKP != nil {
		updates["is_pkp"] = *req.IsPKP
	}
	if req.PPNRate != nil {
		updates["ppn_rate"] = *req.PPNRate
	}
	if req.InvoicePrefix != nil {
		updates["invoice_prefix"] = *req.InvoicePrefix
	}
	if req.InvoiceNumberFormat != nil {
		updates["invoice_number_format"] = *req.InvoiceNumberFormat
	}
	if req.FakturPajakSeries != nil {
		updates["faktur_pajak_series"] = *req.FakturPajakSeries
	}
	if req.SPPKPNumber != nil {
		updates["sppkp_number"] = *req.SPPKPNumber
	}

	// Call service
	updatedCompany, err := h.multiCompanyService.UpdateCompany(c.Request.Context(), companyID, updates)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Get user's access information for response
	companyCtx := util.NewCompanyContext()
	userID := companyCtx.MustGetUserID(c)
	accessInfo, _ := h.multiCompanyService.CheckUserCompanyAccess(c.Request.Context(), userID, companyID)

	// Map to response DTO
	response := h.mapCompanyToDetailResponse(updatedCompany, accessInfo.TenantID, string(accessInfo.Role), accessInfo.AccessTier)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Company updated successfully",
	})
}

// DeactivateCompany soft deletes a company (OWNER only)
// DELETE /api/v1/companies/:id
// Requires Tier 1 access (OWNER only)
func (h *MultiCompanyHandler) DeactivateCompany(c *gin.Context) {
	// Get company ID from URL
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company ID is required"))
		return
	}

	// Call service
	err := h.multiCompanyService.DeactivateCompany(c.Request.Context(), companyID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Company deactivated successfully",
	})
}

// Helper functions

// getStringValue returns the string value from a pointer, or empty string if nil
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getBoolValue returns the bool value from a pointer, or false if nil
func getBoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// mapCompanyToDetailResponse converts service model to detailed DTO response
func (h *MultiCompanyHandler) mapCompanyToDetailResponse(companyData *models.Company, tenantID, userRole string, accessTier int) *dto.CompanyDetailResponse {
	// Use base company mapping from company_handler
	baseResponse := &dto.CompanyResponse{
		ID:            companyData.ID,
		Name:          companyData.Name,
		LegalName:     companyData.LegalName,
		Address:       companyData.Address,
		City:          companyData.City,
		Province:      companyData.Province,
		Phone:         companyData.Phone,
		Email:         companyData.Email,
		IsPKP:         companyData.IsPKP,
		PPNRate:       companyData.PPNRate.InexactFloat64(),
		InvoicePrefix: companyData.InvoicePrefix,
		IsActive:      companyData.IsActive,
	}

	// Set optional fields
	if companyData.NPWP != nil {
		baseResponse.NPWP = *companyData.NPWP
	}
	if companyData.NIB != nil {
		baseResponse.NIB = *companyData.NIB
	}
	if companyData.PostalCode != nil {
		baseResponse.PostalCode = *companyData.PostalCode
	}
	if companyData.Website != nil {
		baseResponse.Website = *companyData.Website
	}
	if companyData.LogoURL != nil {
		baseResponse.LogoURL = *companyData.LogoURL
	}

	// Build detailed response
	response := &dto.CompanyDetailResponse{
		CompanyResponse: *baseResponse,
		TenantID:        tenantID,
		UserRole:        userRole,
		AccessTier:      accessTier,
		CreatedAt:       companyData.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       companyData.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return response
}

// handleValidationError formats and returns validation errors
func (h *MultiCompanyHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		formattedErrors := make([]errors.ValidationError, 0, len(validationErrs))

		for _, fieldErr := range validationErrs {
			formattedErrors = append(formattedErrors, errors.ValidationError{
				Field:   getJSONFieldName(fieldErr),
				Message: formatValidationMessage(fieldErr),
			})
		}

		appErr := errors.NewValidationError(formattedErrors)
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

	validationErrors := []errors.ValidationError{
		{
			Field:   "request",
			Message: err.Error(),
		},
	}
	appErr := errors.NewValidationError(validationErrors)
	c.JSON(appErr.StatusCode, gin.H{
		"success": false,
		"error":   appErr,
	})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *MultiCompanyHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}
