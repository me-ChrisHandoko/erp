package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/deliverytolerance"
	pkgerrors "backend/pkg/errors"
)

// DeliveryToleranceHandler - HTTP handlers for delivery tolerance settings
type DeliveryToleranceHandler struct {
	toleranceService *deliverytolerance.DeliveryToleranceService
}

// NewDeliveryToleranceHandler creates a new delivery tolerance handler instance
func NewDeliveryToleranceHandler(toleranceService *deliverytolerance.DeliveryToleranceService) *DeliveryToleranceHandler {
	return &DeliveryToleranceHandler{
		toleranceService: toleranceService,
	}
}

// ============================================================================
// CREATE DELIVERY TOLERANCE
// ============================================================================

// CreateDeliveryTolerance handles POST /api/v1/delivery-tolerances
func (h *DeliveryToleranceHandler) CreateDeliveryTolerance(c *gin.Context) {
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

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Parse request body
	var req dto.CreateDeliveryToleranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create tolerance
	tolerance, err := h.toleranceService.CreateDeliveryTolerance(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.toleranceService.MapToResponse(tolerance)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// LIST DELIVERY TOLERANCES
// ============================================================================

// ListDeliveryTolerances handles GET /api/v1/delivery-tolerances
func (h *DeliveryToleranceHandler) ListDeliveryTolerances(c *gin.Context) {
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
	var query dto.DeliveryToleranceListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List tolerances
	response, err := h.toleranceService.ListDeliveryTolerances(c.Request.Context(), tenantID.(string), companyID.(string), &query)
	if err != nil {
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
// GET DELIVERY TOLERANCE
// ============================================================================

// GetDeliveryTolerance handles GET /api/v1/delivery-tolerances/:id
func (h *DeliveryToleranceHandler) GetDeliveryTolerance(c *gin.Context) {
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

	// Get tolerance ID from path
	toleranceID := c.Param("id")
	if toleranceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tolerance ID is required"))
		return
	}

	// Get tolerance
	tolerance, err := h.toleranceService.GetDeliveryToleranceByID(c.Request.Context(), tenantID.(string), companyID.(string), toleranceID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.toleranceService.MapToResponse(tolerance)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// UPDATE DELIVERY TOLERANCE
// ============================================================================

// UpdateDeliveryTolerance handles PUT /api/v1/delivery-tolerances/:id
func (h *DeliveryToleranceHandler) UpdateDeliveryTolerance(c *gin.Context) {
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

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get tolerance ID from path
	toleranceID := c.Param("id")
	if toleranceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tolerance ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateDeliveryToleranceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Update tolerance
	tolerance, err := h.toleranceService.UpdateDeliveryTolerance(c.Request.Context(), tenantID.(string), companyID.(string), toleranceID, userIDStr, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.toleranceService.MapToResponse(tolerance)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Delivery tolerance updated successfully",
	})
}

// ============================================================================
// DELETE DELIVERY TOLERANCE
// ============================================================================

// DeleteDeliveryTolerance handles DELETE /api/v1/delivery-tolerances/:id
func (h *DeliveryToleranceHandler) DeleteDeliveryTolerance(c *gin.Context) {
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

	// Get tolerance ID from path
	toleranceID := c.Param("id")
	if toleranceID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tolerance ID is required"))
		return
	}

	// Delete tolerance
	err := h.toleranceService.DeleteDeliveryTolerance(c.Request.Context(), tenantID.(string), companyID.(string), toleranceID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Delivery tolerance deleted successfully",
	})
}

// ============================================================================
// GET EFFECTIVE TOLERANCE
// ============================================================================

// GetEffectiveTolerance handles GET /api/v1/delivery-tolerances/effective
func (h *DeliveryToleranceHandler) GetEffectiveTolerance(c *gin.Context) {
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

	// Get product ID from query
	productID := c.Query("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("product_id query parameter is required"))
		return
	}

	// Get effective tolerance
	response, err := h.toleranceService.GetEffectiveTolerance(c.Request.Context(), tenantID.(string), companyID.(string), productID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// handleValidationError handles validation errors from request binding
func (h *DeliveryToleranceHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getToleranceValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid request format"))
}

// getToleranceValidationErrorMessage returns a user-friendly error message for validation errors
func getToleranceValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "min":
		return fieldErr.Field() + " must be at least " + fieldErr.Param()
	case "max":
		return fieldErr.Field() + " must be at most " + fieldErr.Param()
	case "uuid":
		return fieldErr.Field() + " must be a valid UUID"
	case "oneof":
		return fieldErr.Field() + " must be one of: " + fieldErr.Param()
	default:
		return fieldErr.Field() + " is invalid"
	}
}
