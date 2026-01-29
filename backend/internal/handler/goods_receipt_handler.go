package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/goodsreceipt"
	pkgerrors "backend/pkg/errors"
)

// GoodsReceiptHandler - HTTP handlers for goods receipt management endpoints
type GoodsReceiptHandler struct {
	goodsReceiptService *goodsreceipt.GoodsReceiptService
}

// NewGoodsReceiptHandler creates a new goods receipt handler instance
func NewGoodsReceiptHandler(goodsReceiptService *goodsreceipt.GoodsReceiptService) *GoodsReceiptHandler {
	return &GoodsReceiptHandler{
		goodsReceiptService: goodsReceiptService,
	}
}

// ============================================================================
// CREATE GOODS RECEIPT
// ============================================================================

// CreateGoodsReceipt handles POST /api/v1/goods-receipts
func (h *GoodsReceiptHandler) CreateGoodsReceipt(c *gin.Context) {
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
	var req dto.CreateGoodsReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create goods receipt
	goodsReceipt, err := h.goodsReceiptService.CreateGoodsReceipt(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// LIST GOODS RECEIPTS
// ============================================================================

// ListGoodsReceipts handles GET /api/v1/goods-receipts
func (h *GoodsReceiptHandler) ListGoodsReceipts(c *gin.Context) {
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
	var query dto.GoodsReceiptListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List goods receipts
	response, err := h.goodsReceiptService.ListGoodsReceipts(c.Request.Context(), tenantID.(string), companyID.(string), &query)
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
// GET GOODS RECEIPT BY ID
// ============================================================================

// GetGoodsReceipt handles GET /api/v1/goods-receipts/:id
func (h *GoodsReceiptHandler) GetGoodsReceipt(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Get goods receipt
	goodsReceipt, err := h.goodsReceiptService.GetGoodsReceiptByID(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO (include items for detail view)
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// UPDATE GOODS RECEIPT
// ============================================================================

// UpdateGoodsReceipt handles PUT /api/v1/goods-receipts/:id
func (h *GoodsReceiptHandler) UpdateGoodsReceipt(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateGoodsReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Update goods receipt
	goodsReceipt, err := h.goodsReceiptService.UpdateGoodsReceipt(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// DELETE GOODS RECEIPT
// ============================================================================

// DeleteGoodsReceipt handles DELETE /api/v1/goods-receipts/:id
func (h *GoodsReceiptHandler) DeleteGoodsReceipt(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Delete goods receipt
	err := h.goodsReceiptService.DeleteGoodsReceipt(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID)
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

// ReceiveGoods handles POST /api/v1/goods-receipts/:id/receive
func (h *GoodsReceiptHandler) ReceiveGoods(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Parse optional request body
	var req dto.ReceiveGoodsRequest
	c.ShouldBindJSON(&req)

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Receive goods
	goodsReceipt, err := h.goodsReceiptService.ReceiveGoods(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Goods received successfully",
	})
}

// InspectGoods handles POST /api/v1/goods-receipts/:id/inspect
func (h *GoodsReceiptHandler) InspectGoods(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Parse optional request body
	var req dto.InspectGoodsRequest
	c.ShouldBindJSON(&req)

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Inspect goods
	goodsReceipt, err := h.goodsReceiptService.InspectGoods(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Goods inspected successfully",
	})
}

// AcceptGoods handles POST /api/v1/goods-receipts/:id/accept
func (h *GoodsReceiptHandler) AcceptGoods(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Parse optional request body
	var req dto.AcceptGoodsRequest
	c.ShouldBindJSON(&req)

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Accept goods
	goodsReceipt, err := h.goodsReceiptService.AcceptGoods(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Goods accepted and stock updated successfully",
	})
}

// RejectGoods handles POST /api/v1/goods-receipts/:id/reject
func (h *GoodsReceiptHandler) RejectGoods(c *gin.Context) {
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

	// Get goods receipt ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	// Parse request body
	var req dto.RejectGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Reject goods
	goodsReceipt, err := h.goodsReceiptService.RejectGoods(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Goods rejected successfully",
	})
}

// ============================================================================
// REJECTION DISPOSITION MANAGEMENT (Odoo+M3 Model)
// ============================================================================

// UpdateRejectionDisposition handles PUT /api/v1/goods-receipts/:id/items/:itemId/disposition
func (h *GoodsReceiptHandler) UpdateRejectionDisposition(c *gin.Context) {
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

	// Get goods receipt ID and item ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Item ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateRejectionDispositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Update rejection disposition
	goodsReceipt, err := h.goodsReceiptService.UpdateRejectionDisposition(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, itemID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Rejection disposition updated successfully",
	})
}

// ResolveDisposition handles POST /api/v1/goods-receipts/:id/items/:itemId/resolve-disposition
func (h *GoodsReceiptHandler) ResolveDisposition(c *gin.Context) {
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

	// Get goods receipt ID and item ID from path
	goodsReceiptID := c.Param("id")
	if goodsReceiptID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Goods Receipt ID is required"))
		return
	}

	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Item ID is required"))
		return
	}

	// Parse request body (optional notes)
	var req dto.ResolveDispositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User-Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Resolve disposition
	goodsReceipt, err := h.goodsReceiptService.ResolveDisposition(c.Request.Context(), tenantID.(string), companyID.(string), goodsReceiptID, itemID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := h.goodsReceiptService.MapToResponse(c.Request.Context(), goodsReceipt, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Rejection disposition resolved successfully",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// handleValidationError handles validation errors from request binding
func (h *GoodsReceiptHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getGoodsReceiptValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid request format"))
}

// getGoodsReceiptValidationErrorMessage returns a user-friendly error message for validation errors
func getGoodsReceiptValidationErrorMessage(fieldErr validator.FieldError) string {
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
