package handler

import (
	"encoding/json"
	"log"
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
	log.Printf("🔍 DEBUG [CreatePurchaseOrderHandler]: Request received")

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		log.Printf("❌ DEBUG [CreatePurchaseOrderHandler]: tenant_id not found in context")
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}
	log.Printf("✅ DEBUG [CreatePurchaseOrderHandler]: tenant_id=%s", tenantID)

	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		log.Printf("❌ DEBUG [CreatePurchaseOrderHandler]: company_id not found in context")
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}
	log.Printf("✅ DEBUG [CreatePurchaseOrderHandler]: company_id=%s", companyID)

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}
	log.Printf("✅ DEBUG [CreatePurchaseOrderHandler]: user_id=%s", userIDStr)

	// Parse request body
	var req dto.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ DEBUG [CreatePurchaseOrderHandler]: Failed to bind JSON: %v", err)
		h.handleValidationError(c, err)
		return
	}

	// Log request body for debugging
	reqJSON, _ := json.Marshal(req)
	log.Printf("✅ DEBUG [CreatePurchaseOrderHandler]: Request body: %s", string(reqJSON))

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Create purchase order
	log.Printf("🔍 DEBUG [CreatePurchaseOrderHandler]: Calling service.CreatePurchaseOrder...")
	purchaseOrder, err := h.purchaseOrderService.CreatePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		log.Printf("❌ DEBUG [CreatePurchaseOrderHandler]: Service returned error: %v", err)
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			log.Printf("❌ DEBUG [CreatePurchaseOrderHandler]: AppError - StatusCode=%d, Message=%s", appErr.StatusCode, appErr.Message)
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		log.Printf("❌ DEBUG [CreatePurchaseOrderHandler]: Internal error - returning 500")
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}
	log.Printf("✅ DEBUG [CreatePurchaseOrderHandler]: Purchase order created successfully - ID=%s", purchaseOrder.ID)

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
	var req dto.UpdatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Update purchase order
	purchaseOrder, err := h.purchaseOrderService.UpdatePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, &req, ipAddress, userAgent)
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

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Delete purchase order
	err := h.purchaseOrderService.DeletePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, ipAddress, userAgent)
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

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Confirm purchase order
	purchaseOrder, err := h.purchaseOrderService.ConfirmPurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, ipAddress, userAgent)
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

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Complete purchase order
	purchaseOrder, err := h.purchaseOrderService.CompletePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, ipAddress, userAgent)
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

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Cancel purchase order
	purchaseOrder, err := h.purchaseOrderService.CancelPurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, &req, ipAddress, userAgent)
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

// ShortClosePurchaseOrder handles POST /api/v1/purchase-orders/:id/short-close
// Short close allows closing a PO even if not fully delivered (SAP DCI model)
func (h *PurchaseOrderHandler) ShortClosePurchaseOrder(c *gin.Context) {
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
	var req dto.ShortClosePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Short close purchase order
	purchaseOrder, err := h.purchaseOrderService.ShortClosePurchaseOrder(c.Request.Context(), tenantID.(string), companyID.(string), purchaseOrderID, userIDStr, &req, ipAddress, userAgent)
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
		"message": "Purchase order short closed successfully",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getClientInfo extracts IP address and User Agent from request context
func (h *PurchaseOrderHandler) getClientInfo(c *gin.Context) (ipAddress, userAgent string) {
	ipAddress = c.ClientIP()
	userAgent = c.Request.UserAgent()
	return
}

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
