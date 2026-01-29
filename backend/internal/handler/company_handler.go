package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/company"
	"backend/models"
	"backend/pkg/errors"
	customValidator "backend/pkg/validator"
)

// CompanyHandler handles HTTP requests for company profile management
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Day 1-4 Tasks
// PHASE 5: Updated to use MultiCompanyService for multi-company support
type CompanyHandler struct {
	companyService      *company.CompanyService
	multiCompanyService *company.MultiCompanyService
	validator           *customValidator.Validator
}

// NewCompanyHandler creates a new company handler
// PHASE 5: Now requires MultiCompanyService for multi-company operations
func NewCompanyHandler(companyService *company.CompanyService, multiCompanyService *company.MultiCompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService:      companyService,
		multiCompanyService: multiCompanyService,
		validator:           customValidator.New(),
	}
}

// GetCompanyProfile retrieves company profile for current active company
// GET /api/v1/company
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md lines 35-49
// PHASE 5: Updated to use active company ID from context (multi-company support)
func (h *CompanyHandler) GetCompanyProfile(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Call service to get company by ID
	companyData, err := h.multiCompanyService.GetCompanyByID(c.Request.Context(), companyID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTO
	response := h.mapCompanyToResponse(companyData)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateCompanyProfile updates company profile
// PUT /api/v1/company
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md lines 50-132
// PHASE 5: Updated to use company_id from CompanyContextMiddleware
func (h *CompanyHandler) UpdateCompanyProfile(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
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
	if req.EntityType != nil {
		updates["entity_type"] = *req.EntityType
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
	// Purchase Invoice Settings (3-way matching)
	if req.InvoiceControlPolicy != nil {
		updates["invoice_control_policy"] = *req.InvoiceControlPolicy
	}
	if req.InvoiceTolerancePct != nil {
		updates["invoice_tolerance_pct"] = *req.InvoiceTolerancePct
	}

	// Call service with company ID (using MultiCompanyService for PHASE 5)
	updatedCompany, err := h.multiCompanyService.UpdateCompany(c.Request.Context(), companyID.(string), updates)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTO
	response := h.mapCompanyToResponse(updatedCompany)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Company profile updated successfully",
	})
}

// AddBankAccount adds a new bank account
// POST /api/v1/company/banks
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #10
// PHASE 5: Updated to use company_id from CompanyContextMiddleware
func (h *CompanyHandler) AddBankAccount(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	var req dto.AddBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Convert DTO to service request
	serviceReq := &company.AddBankRequest{
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		BranchName:    req.BranchName,
		IsPrimary:     req.IsPrimary,
		CheckPrefix:   req.CheckPrefix,
	}

	// Call service with company ID
	bank, err := h.companyService.AddBankAccount(c.Request.Context(), companyID.(string), serviceReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTO
	response := h.mapBankToResponse(bank)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Bank account added successfully",
	})
}

// UpdateBankAccount updates an existing bank account
// PUT /api/v1/company/banks/:id
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #10
// PHASE 5: Updated to use company_id from CompanyContextMiddleware
func (h *CompanyHandler) UpdateBankAccount(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get bank ID from URL
	bankID := c.Param("id")
	if bankID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Bank ID is required"))
		return
	}

	var req dto.UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Convert DTO to map for service
	updates := make(map[string]interface{})
	if req.BankName != nil {
		updates["bank_name"] = *req.BankName
	}
	if req.AccountNumber != nil {
		updates["account_number"] = *req.AccountNumber
	}
	if req.AccountName != nil {
		updates["account_name"] = *req.AccountName
	}
	if req.BranchName != nil {
		updates["branch_name"] = *req.BranchName
	}
	if req.IsPrimary != nil {
		updates["is_primary"] = *req.IsPrimary
	}
	if req.CheckPrefix != nil {
		updates["check_prefix"] = *req.CheckPrefix
	}

	// Call service with company ID
	updatedBank, err := h.companyService.UpdateBankAccount(c.Request.Context(), companyID.(string), bankID, updates)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTO
	response := h.mapBankToResponse(updatedBank)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Bank account updated successfully",
	})
}

// DeleteBankAccount soft deletes a bank account
// DELETE /api/v1/company/banks/:id
// Reference: ANALYSIS-01-TENANT-COMPANY-SETUP.md Issue #4
// PHASE 5: Updated to use company_id from CompanyContextMiddleware
func (h *CompanyHandler) DeleteBankAccount(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get bank ID from URL
	bankID := c.Param("id")
	if bankID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Bank ID is required"))
		return
	}

	// Call service with company ID
	err := h.companyService.DeleteBankAccount(c.Request.Context(), companyID.(string), bankID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bank account deleted successfully",
	})
}

// GetBankAccounts retrieves paginated bank accounts with filters for current company
// GET /api/v1/company/banks
// PHASE 5: Updated to use company_id from CompanyContextMiddleware
// Updated to support pagination and filtering (follows Products pattern)
func (h *CompanyHandler) GetBankAccounts(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Bind query params to filters struct
	var filters dto.BankAccountFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Set defaults if not provided (backward compatibility)
	if filters.Page == 0 {
		filters.Page = 1
	}
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.SortBy == "" {
		filters.SortBy = "bankName"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "asc"
	}

	// Call service with filters
	banks, total, err := h.companyService.ListBankAccounts(
		c.Request.Context(),
		companyID.(string),
		&filters,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTOs
	response := make([]dto.BankAccountResponse, 0, len(banks))
	for _, bank := range banks {
		response = append(response, *h.mapBankToResponse(bank))
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	// Response with pagination metadata (CONSISTENT with Products pattern)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"pagination": dto.PaginationInfo{
			Page:       filters.Page,
			Limit:      filters.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}

// GetBankAccount retrieves a single bank account by ID
// GET /api/v1/company/banks/:id
// PHASE 5: Updated to use company_id from CompanyContextMiddleware
func (h *CompanyHandler) GetBankAccount(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get bank ID from URL
	bankID := c.Param("id")
	if bankID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Bank ID is required"))
		return
	}

	// Call service with company ID
	bank, err := h.companyService.GetBankAccountByID(c.Request.Context(), companyID.(string), bankID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response DTO
	response := h.mapBankToResponse(bank)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Helper functions

// mapCompanyToResponse converts service model to DTO response
func (h *CompanyHandler) mapCompanyToResponse(companyData interface{}) *dto.CompanyResponse {
	// Type assertion to *models.Company
	companyModel, ok := companyData.(*models.Company)
	if !ok {
		// Fallback: return minimal response
		return &dto.CompanyResponse{}
	}

	response := &dto.CompanyResponse{
		ID:            companyModel.ID,
		Name:          companyModel.Name,
		LegalName:     companyModel.LegalName,
		EntityType:    companyModel.EntityType,
		Address:       companyModel.Address,
		City:          companyModel.City,
		Province:      companyModel.Province,
		Phone:         companyModel.Phone,
		Email:         companyModel.Email,
		IsPKP:         companyModel.IsPKP,
		PPNRate:       companyModel.PPNRate.InexactFloat64(),
		InvoicePrefix: companyModel.InvoicePrefix,
		// Purchase Invoice Settings (3-way matching)
		InvoiceControlPolicy: string(companyModel.InvoiceControlPolicy),
		InvoiceTolerancePct:  companyModel.InvoiceTolerancePct.InexactFloat64(),
		IsActive:             companyModel.IsActive,
	}

	// Set optional fields
	if companyModel.NPWP != nil {
		response.NPWP = *companyModel.NPWP
	}
	if companyModel.NIB != nil {
		response.NIB = *companyModel.NIB
	}
	if companyModel.PostalCode != nil {
		response.PostalCode = *companyModel.PostalCode
	}
	if companyModel.Website != nil {
		response.Website = *companyModel.Website
	}
	if companyModel.LogoURL != nil {
		response.LogoURL = *companyModel.LogoURL
	}

	// Map banks if preloaded
	if len(companyModel.Banks) > 0 {
		response.Banks = make([]dto.CompanyBankInfo, 0, len(companyModel.Banks))
		for _, bank := range companyModel.Banks {
			bankInfo := dto.CompanyBankInfo{
				ID:            bank.ID,
				BankName:      bank.BankName,
				AccountNumber: bank.AccountNumber,
				AccountName:   bank.AccountName,
				IsPrimary:     bank.IsPrimary,
				IsActive:      bank.IsActive,
			}
			if bank.BranchName != nil {
				bankInfo.BranchName = *bank.BranchName
			}
			if bank.CheckPrefix != nil {
				bankInfo.CheckPrefix = *bank.CheckPrefix
			}
			response.Banks = append(response.Banks, bankInfo)
		}
	}

	return response
}

// mapBankToResponse converts service bank model to DTO response
func (h *CompanyHandler) mapBankToResponse(bankData interface{}) *dto.BankAccountResponse {
	// Type assertion to *models.CompanyBank
	bankModel, ok := bankData.(*models.CompanyBank)
	if !ok {
		// Fallback: return minimal response
		return &dto.BankAccountResponse{}
	}

	response := &dto.BankAccountResponse{
		ID:            bankModel.ID,
		BankName:      bankModel.BankName,
		AccountNumber: bankModel.AccountNumber,
		AccountName:   bankModel.AccountName,
		IsPrimary:     bankModel.IsPrimary,
		IsActive:      bankModel.IsActive,
	}

	// Set optional fields
	if bankModel.BranchName != nil {
		response.BranchName = *bankModel.BranchName
	}
	if bankModel.CheckPrefix != nil {
		response.CheckPrefix = *bankModel.CheckPrefix
	}

	return response
}

// handleValidationError formats and returns validation errors
func (h *CompanyHandler) handleValidationError(c *gin.Context, err error) {
	// Check if it's a validator error
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		// Format each validation error
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

	// Not a validation error, return generic error
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
func (h *CompanyHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

	// Unknown error - return internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}
