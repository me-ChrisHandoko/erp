package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/invoice"
	pkgerrors "backend/pkg/errors"
)

// InvoiceHandler - HTTP handlers for invoice management endpoints
type InvoiceHandler struct {
	invoiceService *invoice.InvoiceService
}

// NewInvoiceHandler creates a new invoice handler instance
func NewInvoiceHandler(invoiceService *invoice.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
	}
}

// ============================================================================
// CREATE INVOICE
// ============================================================================

// CreateInvoice handles POST /api/v1/invoices
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
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
	var req dto.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Create invoice
	invoiceResp, err := h.invoiceService.CreateInvoice(companyID.(string), tenantID.(string), req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response in standard API format
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    invoiceResp,
	})
}

// ============================================================================
// LIST INVOICES
// ============================================================================

// ListInvoices handles GET /api/v1/invoices
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Parse query parameters
	var filters dto.InvoiceFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get invoices
	result, err := h.invoiceService.ListInvoices(tenantID.(string), companyID.(string), filters)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response in standard API format (matching products endpoint)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Data,
		"pagination": result.Pagination,
	})
}

// ============================================================================
// GET INVOICE
// ============================================================================

// GetInvoice handles GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get invoice ID from URL parameter
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invoice ID is required"))
		return
	}

	// Get invoice
	invoiceResp, err := h.invoiceService.GetInvoice(tenantID.(string), companyID.(string), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Invoice not found"))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response in standard API format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    invoiceResp,
	})
}

// ============================================================================
// UPDATE INVOICE
// ============================================================================

// UpdateInvoice handles PUT /api/v1/invoices/:id
func (h *InvoiceHandler) UpdateInvoice(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get invoice ID from URL parameter
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invoice ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Update invoice
	invoiceResp, err := h.invoiceService.UpdateInvoice(tenantID.(string), companyID.(string), invoiceID, req)
	if err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Invoice not found"))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response in standard API format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    invoiceResp,
	})
}

// ============================================================================
// DELETE INVOICE
// ============================================================================

// DeleteInvoice handles DELETE /api/v1/invoices/:id
func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get invoice ID from URL parameter
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invoice ID is required"))
		return
	}

	// Delete invoice
	if err := h.invoiceService.DeleteInvoice(tenantID.(string), companyID.(string), invoiceID); err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Invoice not found"))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invoice deleted successfully",
	})
}

// ============================================================================
// RECORD PAYMENT
// ============================================================================

// RecordPayment handles POST /api/v1/invoices/:id/payments
func (h *InvoiceHandler) RecordPayment(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get invoice ID from URL parameter
	invoiceID := c.Param("id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invoice ID is required"))
		return
	}

	// Parse request body
	var req dto.RecordInvoicePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Record payment
	invoiceResp, err := h.invoiceService.RecordPayment(tenantID.(string), companyID.(string), invoiceID, req)
	if err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Invoice not found"))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response in standard API format
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    invoiceResp,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// handleValidationError handles validation errors from request binding
func (h *InvoiceHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		details := make([]pkgerrors.ValidationError, 0, len(validationErrs))
		for _, e := range validationErrs {
			details = append(details, pkgerrors.ValidationError{
				Field:   e.Field(),
				Message: getValidationErrorMsg(e),
			})
		}
		appErr := pkgerrors.NewValidationError(details)
		c.JSON(appErr.StatusCode, appErr)
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
}

// getValidationErrorMsg returns a user-friendly validation error message
func getValidationErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too small"
	case "max":
		return "Value is too large"
	case "uuid":
		return "Invalid UUID format"
	case "oneof":
		return "Invalid value"
	default:
		return "Invalid value"
	}
}
