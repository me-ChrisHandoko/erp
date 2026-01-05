package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/customer"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// CustomerHandler - HTTP handlers for customer management endpoints
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 2 - Customer Management
type CustomerHandler struct {
	customerService *customer.CustomerService
}

// NewCustomerHandler creates a new customer handler instance
func NewCustomerHandler(customerService *customer.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// ============================================================================
// CREATE CUSTOMER
// ============================================================================

// CreateCustomer handles POST /api/v1/customers
// @Summary Create a new customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body dto.CreateCustomerRequest true "Customer creation request"
// @Success 201 {object} dto.CustomerResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/customers [post]
// @Security BearerAuth
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	// Get tenant_id from context (set by middleware)
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
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create customer
	customerModel, err := h.customerService.CreateCustomer(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapCustomerToResponse(customerModel)

	// Return response in standard API format (matching products endpoint)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// LIST CUSTOMERS
// ============================================================================

// ListCustomers handles GET /api/v1/customers
// @Summary List customers with filtering and pagination
// @Tags Customers
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 20, max: 100)"
// @Param search query string false "Search by code or name"
// @Param type query string false "Filter by type (RETAIL, WHOLESALE, DISTRIBUTOR)"
// @Param city query string false "Filter by city"
// @Param province query string false "Filter by province"
// @Param isPKP query bool false "Filter by PKP status"
// @Param isActive query bool false "Filter by active status (default: true)"
// @Param hasOverdue query bool false "Filter customers with overdue amounts"
// @Param sortBy query string false "Sort by field (code, name, createdAt, currentOutstanding, overdueAmount)"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Success 200 {object} dto.CustomerListResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/customers [get]
// @Security BearerAuth
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// Get tenant_id from context (set by middleware)
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
	var query dto.CustomerListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List customers
	response, err := h.customerService.ListCustomers(c.Request.Context(), tenantID.(string), companyID.(string), &query)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response in standard API format (matching products endpoint and frontend expectations)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response.Customers,
		"pagination": gin.H{
			"page":       response.Page,
			"pageSize":   response.PageSize,
			"totalItems": response.TotalCount,
			"totalPages": response.TotalPages,
			"hasMore":    response.Page < response.TotalPages,
		},
	})
}

// ============================================================================
// GET CUSTOMER BY ID
// ============================================================================

// GetCustomer handles GET /api/v1/customers/:id
// @Summary Get customer by ID
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/customers/{id} [get]
// @Security BearerAuth
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	// Get tenant_id from context (set by middleware)
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

	// Get customer ID from path
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Customer ID is required"))
		return
	}

	// Get customer
	customerModel, err := h.customerService.GetCustomerByID(c.Request.Context(), tenantID.(string), companyID.(string), customerID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapCustomerToResponse(customerModel)

	// Return response in standard API format (matching products endpoint)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// UPDATE CUSTOMER
// ============================================================================

// UpdateCustomer handles PUT /api/v1/customers/:id
// @Summary Update an existing customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body dto.UpdateCustomerRequest true "Customer update request"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/customers/{id} [put]
// @Security BearerAuth
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	// Get tenant_id from context (set by middleware)
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

	// Get customer ID from path
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Customer ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Update customer
	customerModel, err := h.customerService.UpdateCustomer(c.Request.Context(), tenantID.(string), companyID.(string), customerID, userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapCustomerToResponse(customerModel)

	// Return response in standard API format (matching products endpoint)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// DELETE CUSTOMER
// ============================================================================

// DeleteCustomer handles DELETE /api/v1/customers/:id
// @Summary Soft delete a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 204 "No Content"
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/customers/{id} [delete]
// @Security BearerAuth
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	// Get tenant_id from context (set by middleware)
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

	// Get customer ID from path
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Customer ID is required"))
		return
	}

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Delete customer
	err := h.customerService.DeleteCustomer(c.Request.Context(), tenantID.(string), companyID.(string), customerID, userIDStr, ipAddress, userAgent)
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
func (h *CustomerHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid request format"))
}

// getValidationErrorMessage returns a user-friendly error message for validation errors
func getValidationErrorMessage(fieldErr validator.FieldError) string {
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

// mapCustomerToResponse converts Customer model to CustomerResponse DTO
func mapCustomerToResponse(customer *models.Customer) dto.CustomerResponse {
	return dto.CustomerResponse{
		ID:                 customer.ID,
		Code:               customer.Code,
		Name:               customer.Name,
		Type:               customer.Type,
		Phone:              customer.Phone,
		Email:              customer.Email,
		Address:            customer.Address,
		City:               customer.City,
		Province:           customer.Province,
		PostalCode:         customer.PostalCode,
		NPWP:               customer.NPWP,
		IsPKP:              customer.IsPKP,
		ContactPerson:      customer.ContactPerson,
		ContactPhone:       customer.ContactPhone,
		PaymentTerm:        customer.PaymentTerm,
		CreditLimit:        customer.CreditLimit.String(),
		CurrentOutstanding: customer.CurrentOutstanding.String(),
		OverdueAmount:      customer.OverdueAmount.String(),
		LastTransactionAt:  customer.LastTransactionAt,
		Notes:              customer.Notes,
		IsActive:           customer.IsActive,
		CreatedAt:          customer.CreatedAt,
		UpdatedAt:          customer.UpdatedAt,
	}
}
