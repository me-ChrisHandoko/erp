package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/supplier"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// SupplierHandler - HTTP handlers for supplier management endpoints
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 3 - Supplier Management
type SupplierHandler struct {
	supplierService *supplier.SupplierService
}

// NewSupplierHandler creates a new supplier handler instance
func NewSupplierHandler(supplierService *supplier.SupplierService) *SupplierHandler {
	return &SupplierHandler{
		supplierService: supplierService,
	}
}

// ============================================================================
// CREATE SUPPLIER
// ============================================================================

// CreateSupplier handles POST /api/v1/suppliers
// @Summary Create a new supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param request body dto.CreateSupplierRequest true "Supplier creation request"
// @Success 201 {object} dto.SupplierResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/suppliers [post]
// @Security BearerAuth
func (h *SupplierHandler) CreateSupplier(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Parse request body
	var req dto.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user ID from JWT middleware for audit logging
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create supplier
	supplierModel, err := h.supplierService.CreateSupplier(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapSupplierToResponse(supplierModel)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// LIST SUPPLIERS
// ============================================================================

// ListSuppliers handles GET /api/v1/suppliers
// @Summary List suppliers with filtering and pagination
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 20, max: 100)"
// @Param search query string false "Search by code or name"
// @Param type query string false "Filter by type (MANUFACTURER, DISTRIBUTOR, WHOLESALER)"
// @Param city query string false "Filter by city"
// @Param province query string false "Filter by province"
// @Param isPKP query bool false "Filter by PKP status"
// @Param isActive query bool false "Filter by active status (default: true)"
// @Param hasOverdue query bool false "Filter suppliers with overdue amounts"
// @Param sortBy query string false "Sort by field (code, name, createdAt, currentOutstanding, overdueAmount)"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Success 200 {object} dto.SupplierListResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/suppliers [get]
// @Security BearerAuth
func (h *SupplierHandler) ListSuppliers(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Parse query parameters
	var query dto.SupplierListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List suppliers
	response, err := h.supplierService.ListSuppliers(c.Request.Context(), tenantID.(string), companyID.(string), &query)
	if err != nil {
		// Log the actual error for debugging
		println("❌ [ListSuppliers] Error:", err.Error())
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// GET SUPPLIER BY ID
// ============================================================================

// GetSupplier handles GET /api/v1/suppliers/:id
// @Summary Get supplier by ID
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID"
// @Success 200 {object} dto.SupplierResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/suppliers/{id} [get]
// @Security BearerAuth
func (h *SupplierHandler) GetSupplier(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get supplier ID from path
	supplierID := c.Param("id")
	if supplierID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Supplier ID is required"))
		return
	}

	// Get supplier
	supplierModel, err := h.supplierService.GetSupplierByID(c.Request.Context(), tenantID.(string), companyID.(string), supplierID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapSupplierToResponse(supplierModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// UPDATE SUPPLIER
// ============================================================================

// UpdateSupplier handles PUT /api/v1/suppliers/:id
// @Summary Update an existing supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID"
// @Param request body dto.UpdateSupplierRequest true "Supplier update request"
// @Success 200 {object} dto.SupplierResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/suppliers/{id} [put]
// @Security BearerAuth
func (h *SupplierHandler) UpdateSupplier(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get supplier ID from path
	supplierID := c.Param("id")
	if supplierID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Supplier ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user ID from JWT middleware for audit logging
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Update supplier
	supplierModel, err := h.supplierService.UpdateSupplier(c.Request.Context(), tenantID.(string), companyID.(string), supplierID, userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapSupplierToResponse(supplierModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// DELETE SUPPLIER
// ============================================================================

// DeleteSupplier handles DELETE /api/v1/suppliers/:id
// @Summary Soft delete a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID"
// @Success 204 "No Content"
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/suppliers/{id} [delete]
// @Security BearerAuth
func (h *SupplierHandler) DeleteSupplier(c *gin.Context) {
	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get supplier ID from path
	supplierID := c.Param("id")
	if supplierID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Supplier ID is required"))
		return
	}

	// Get user ID from JWT middleware for audit logging
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Delete supplier
	err := h.supplierService.DeleteSupplier(c.Request.Context(), tenantID.(string), companyID.(string), supplierID, userIDStr, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	c.Status(http.StatusNoContent)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// handleValidationError handles validation errors from request binding
func (h *SupplierHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getSupplierValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid request format"))
}

// getSupplierValidationErrorMessage returns a user-friendly error message for validation errors
func getSupplierValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "min":
		return fieldErr.Field() + " must be at least " + fieldErr.Param() + " characters"
	case "max":
		return fieldErr.Field() + " must be at most " + fieldErr.Param() + " characters"
	case "email":
		return fieldErr.Field() + " must be a valid email address"
	case "oneof":
		return fieldErr.Field() + " must be one of: " + fieldErr.Param()
	default:
		return fieldErr.Field() + " is invalid"
	}
}

// mapSupplierToResponse converts Supplier model to SupplierResponse DTO
func mapSupplierToResponse(supplier *models.Supplier) dto.SupplierResponse {
	return dto.SupplierResponse{
		ID:                 supplier.ID,
		Code:               supplier.Code,
		Name:               supplier.Name,
		Type:               supplier.Type,
		Phone:              supplier.Phone,
		Email:              supplier.Email,
		Address:            supplier.Address,
		City:               supplier.City,
		Province:           supplier.Province,
		PostalCode:         supplier.PostalCode,
		NPWP:               supplier.NPWP,
		IsPKP:              supplier.IsPKP,
		ContactPerson:      supplier.ContactPerson,
		ContactPhone:       supplier.ContactPhone,
		PaymentTerm:        supplier.PaymentTerm,
		CreditLimit:        supplier.CreditLimit.String(),
		CurrentOutstanding: supplier.CurrentOutstanding.String(),
		OverdueAmount:      supplier.OverdueAmount.String(),
		LastTransactionAt:  supplier.LastTransactionAt,
		Notes:              supplier.Notes,
		IsActive:           supplier.IsActive,
		CreatedAt:          supplier.CreatedAt,
		UpdatedAt:          supplier.UpdatedAt,
	}
}
