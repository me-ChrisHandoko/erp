package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/purchase"
	pkgerrors "backend/pkg/errors"
)

// PurchaseOrderHandler - HTTP handlers for purchase order management endpoints
type PurchaseOrderHandler struct {
	purchaseOrderService *purchase.PurchaseOrderService
}

// NewPurchaseOrderHandler creates a new purchase order handler instance
func NewPurchaseOrderHandler(purchaseOrderService *purchase.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		purchaseOrderService: purchaseOrderService,
	}
}

// ============================================================================
// CREATE PURCHASE ORDER
// ============================================================================

// CreatePurchaseOrder handles POST /api/v1/purchase-orders
func (h *PurchaseOrderHandler) CreatePurchaseOrder(c *gin.Context) {
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
	var req dto.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Create purchase order
	purchaseOrder, err := h.purchaseOrderService.CreatePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.purchaseOrderService.MapToResponse(purchaseOrder, true)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// LIST PURCHASE ORDERS
// ============================================================================

// ListPurchaseOrders handles GET /api/v1/purchase-orders
func (h *PurchaseOrderHandler) ListPurchaseOrders(c *gin.Context) {
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
	var query dto.PurchaseOrderListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List purchase orders
	response, err := h.purchaseOrderService.ListPurchaseOrders(c.Request.Context(), tenantID.(string), companyID.(string), &query)
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
// GET PURCHASE ORDER BY ID
// ============================================================================

// GetPurchaseOrder handles GET /api/v1/purchase-orders/:id
func (h *PurchaseOrderHandler) GetPurchaseOrder(c *gin.Context) {
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

	// Get purchase order ID from path
	purchaseOrderID := c.Param("id")
	if purchaseOrderID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Purchase Order ID is required"))
		return
	}

	// Get purchase order
	purchaseOrder, err := h.purchaseOrderService.GetPurchaseOrderByID(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO (include items for detail view)
	response := h.purchaseOrderService.MapToResponse(purchaseOrder, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// UPDATE PURCHASE ORDER
// ============================================================================

// UpdatePurchaseOrder handles PUT /api/v1/purchase-orders/:id
func (h *PurchaseOrderHandler) UpdatePurchaseOrder(c *gin.Context) {
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

	// Get purchase order ID from path
	purchaseOrderID := c.Param("id")
	if purchaseOrderID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Purchase Order ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Update purchase order
	purchaseOrder, err := h.purchaseOrderService.UpdatePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.purchaseOrderService.MapToResponse(purchaseOrder, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// DELETE PURCHASE ORDER
// ============================================================================

// DeletePurchaseOrder handles DELETE /api/v1/purchase-orders/:id
func (h *PurchaseOrderHandler) DeletePurchaseOrder(c *gin.Context) {
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

	// Get purchase order ID from path
	purchaseOrderID := c.Param("id")
	if purchaseOrderID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Purchase Order ID is required"))
		return
	}

	// Delete purchase order
	err := h.purchaseOrderService.DeletePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID)
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
// STATUS TRANSITION ENDPOINTS
// ============================================================================

// ConfirmPurchaseOrder handles POST /api/v1/purchase-orders/:id/confirm
func (h *PurchaseOrderHandler) ConfirmPurchaseOrder(c *gin.Context) {
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

	// Get purchase order ID from path
	purchaseOrderID := c.Param("id")
	if purchaseOrderID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Purchase Order ID is required"))
		return
	}

	// Confirm purchase order
	purchaseOrder, err := h.purchaseOrderService.ConfirmPurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.purchaseOrderService.MapToResponse(purchaseOrder, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Purchase order confirmed successfully",
	})
}

// CompletePurchaseOrder handles POST /api/v1/purchase-orders/:id/complete
func (h *PurchaseOrderHandler) CompletePurchaseOrder(c *gin.Context) {
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

	// Get purchase order ID from path
	purchaseOrderID := c.Param("id")
	if purchaseOrderID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Purchase Order ID is required"))
		return
	}

	// Complete purchase order
	purchaseOrder, err := h.purchaseOrderService.CompletePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.purchaseOrderService.MapToResponse(purchaseOrder, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Purchase order completed successfully",
	})
}

// CancelPurchaseOrder handles POST /api/v1/purchase-orders/:id/cancel
func (h *PurchaseOrderHandler) CancelPurchaseOrder(c *gin.Context) {
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

	// Get purchase order ID from path
	purchaseOrderID := c.Param("id")
	if purchaseOrderID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Purchase Order ID is required"))
		return
	}

	// Parse request body
	var req dto.CancelPurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Cancel purchase order
	purchaseOrder, err := h.purchaseOrderService.CancelPurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.purchaseOrderService.MapToResponse(purchaseOrder, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Purchase order cancelled successfully",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// handleValidationError handles validation errors from request binding
func (h *PurchaseOrderHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getPurchaseOrderValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid request format"))
}

// getPurchaseOrderValidationErrorMessage returns a user-friendly error message for validation errors
func getPurchaseOrderValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "min":
		return fieldErr.Field() + " must have at least " + fieldErr.Param() + " items"
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
