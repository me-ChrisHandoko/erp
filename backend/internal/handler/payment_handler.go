package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/payment"
	pkgerrors "backend/pkg/errors"
)

// PaymentHandler - HTTP handlers for customer payment management endpoints
type PaymentHandler struct {
	paymentService *payment.PaymentService
}

// NewPaymentHandler creates a new payment handler instance
func NewPaymentHandler(paymentService *payment.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// ============================================================================
// LIST PAYMENTS
// ============================================================================

// ListPayments handles GET /api/v1/payments
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Parse query parameters
	var filters dto.PaymentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get payments
	result, err := h.paymentService.ListPayments(companyID.(string), tenantID.(string), filters)
	if err != nil {
		// Log the actual error for debugging
		println("ERROR in ListPayments:", err.Error())
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
		"data":    result,
	})
}

// ============================================================================
// GET PAYMENT
// ============================================================================

// GetPayment handles GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get tenant_id from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Get payment
	result, err := h.paymentService.GetPayment(companyID.(string), tenantID.(string), paymentID)
	if err != nil {
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Payment not found."))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ============================================================================
// CREATE PAYMENT
// ============================================================================

// CreatePayment handles POST /api/v1/payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
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

	// Parse request body
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Create payment with audit info
	paymentResp, err := h.paymentService.CreatePayment(c.Request.Context(), companyID.(string), tenantID.(string), userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
		return
	}

	// Return response in standard API format
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    paymentResp,
		"message": "Payment created successfully",
	})
}

// ============================================================================
// UPDATE PAYMENT
// ============================================================================

// UpdatePayment handles PUT /api/v1/payments/:id
func (h *PaymentHandler) UpdatePayment(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get tenant_id from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
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

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Parse request body
	var req dto.UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Update payment with audit info
	paymentResp, err := h.paymentService.UpdatePayment(c.Request.Context(), companyID.(string), tenantID.(string), paymentID, userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Payment not found."))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paymentResp,
		"message": "Payment updated successfully",
	})
}

// ============================================================================
// VOID PAYMENT
// ============================================================================

// VoidPayment handles DELETE /api/v1/payments/:id
func (h *PaymentHandler) VoidPayment(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get tenant_id from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
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

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Parse request body (optional)
	var req dto.VoidPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// It's okay if body is empty
		req = dto.VoidPaymentRequest{}
	}

	// Void payment with audit info
	err := h.paymentService.VoidPayment(c.Request.Context(), companyID.(string), tenantID.(string), paymentID, userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Payment not found."))
			return
		}
		if err.Error() == "can only void payments from today" {
			c.JSON(http.StatusForbidden, pkgerrors.NewAuthorizationError("Can only void payments from today."))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment voided successfully",
	})
}

// ============================================================================
// UPDATE CHECK STATUS
// ============================================================================

// UpdateCheckStatus handles PATCH /api/v1/payments/:id/check-status
func (h *PaymentHandler) UpdateCheckStatus(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	// Get tenant_id from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
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

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Parse request body
	var req dto.UpdateCheckStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Update check status with audit info
	paymentResp, err := h.paymentService.UpdateCheckStatus(c.Request.Context(), companyID.(string), tenantID.(string), paymentID, userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if err.Error() == "payment not found" {
			c.JSON(http.StatusNotFound, pkgerrors.NewNotFoundError("Payment not found."))
			return
		}
		if err.Error() == "payment does not have check records" {
			c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Payment does not have check/giro records."))
			return
		}
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    paymentResp,
		"message": "Check status updated successfully",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func (h *PaymentHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range validationErrs {
			errors[e.Field()] = e.Tag()
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Validation failed",
				"details": errors,
			},
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
}
