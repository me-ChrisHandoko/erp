package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/inventoryadjustment"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// InventoryAdjustmentHandler - HTTP handlers for inventory adjustment endpoints
type InventoryAdjustmentHandler struct {
	adjustmentService *inventoryadjustment.InventoryAdjustmentService
}

// NewInventoryAdjustmentHandler creates a new inventory adjustment handler instance
func NewInventoryAdjustmentHandler(adjustmentService *inventoryadjustment.InventoryAdjustmentService) *InventoryAdjustmentHandler {
	return &InventoryAdjustmentHandler{
		adjustmentService: adjustmentService,
	}
}

// ============================================================================
// INVENTORY ADJUSTMENT CRUD ENDPOINTS
// ============================================================================

// CreateInventoryAdjustment handles POST /api/v1/inventory-adjustments
func (h *InventoryAdjustmentHandler) CreateInventoryAdjustment(c *gin.Context) {
	// Get context values
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Parse request
	var req dto.CreateInventoryAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	// Create adjustment
	adjustment, err := h.adjustmentService.CreateInventoryAdjustment(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		userIDStr,
		&req,
		ipAddress,
		userAgent,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapInventoryAdjustmentToResponse(adjustment)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListInventoryAdjustments handles GET /api/v1/inventory-adjustments
func (h *InventoryAdjustmentHandler) ListInventoryAdjustments(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	// Parse query parameters
	var query dto.InventoryAdjustmentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid query parameters"))
		return
	}

	// Set defaults
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}
	if query.SortBy == "" {
		query.SortBy = "adjustmentNumber"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	adjustments, pagination, err := h.adjustmentService.ListInventoryAdjustments(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		&query,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Get status counts for statistics cards
	statusCounts, err := h.adjustmentService.GetStatusCounts(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
	)
	if err != nil {
		// Log error but don't fail the request - status counts are optional
		statusCounts = nil
	}

	responses := make([]dto.InventoryAdjustmentResponse, len(adjustments))
	for i, adjustment := range adjustments {
		responses[i] = mapInventoryAdjustmentToResponse(&adjustment)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         responses,
		"pagination":   pagination,
		"statusCounts": statusCounts,
	})
}

// GetInventoryAdjustment handles GET /api/v1/inventory-adjustments/:id
func (h *InventoryAdjustmentHandler) GetInventoryAdjustment(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	adjustmentID := c.Param("id")
	if adjustmentID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Adjustment ID is required"))
		return
	}

	adjustment, err := h.adjustmentService.GetInventoryAdjustmentByID(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		adjustmentID,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapInventoryAdjustmentToResponse(adjustment)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateInventoryAdjustment handles PUT /api/v1/inventory-adjustments/:id
func (h *InventoryAdjustmentHandler) UpdateInventoryAdjustment(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	adjustmentID := c.Param("id")
	if adjustmentID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Adjustment ID is required"))
		return
	}

	var req dto.UpdateInventoryAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	adjustment, err := h.adjustmentService.UpdateInventoryAdjustment(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		adjustmentID,
		userIDStr,
		&req,
		ipAddress,
		userAgent,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapInventoryAdjustmentToResponse(adjustment)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// DeleteInventoryAdjustment handles DELETE /api/v1/inventory-adjustments/:id
func (h *InventoryAdjustmentHandler) DeleteInventoryAdjustment(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	adjustmentID := c.Param("id")
	if adjustmentID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Adjustment ID is required"))
		return
	}

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	err := h.adjustmentService.DeleteInventoryAdjustment(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		adjustmentID,
		userIDStr,
		ipAddress,
		userAgent,
	)
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
		"message": "Inventory adjustment deleted successfully",
	})
}

// ============================================================================
// STATUS TRANSITION ENDPOINTS
// ============================================================================

// ApproveInventoryAdjustment handles POST /api/v1/inventory-adjustments/:id/approve
func (h *InventoryAdjustmentHandler) ApproveInventoryAdjustment(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	adjustmentID := c.Param("id")
	if adjustmentID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Adjustment ID is required"))
		return
	}

	var req dto.ApproveAdjustmentRequest
	// Bind JSON, but it's okay if body is empty
	_ = c.ShouldBindJSON(&req)

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	adjustment, err := h.adjustmentService.ApproveInventoryAdjustment(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		adjustmentID,
		userIDStr,
		&req,
		ipAddress,
		userAgent,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapInventoryAdjustmentToResponse(adjustment)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CancelInventoryAdjustment handles POST /api/v1/inventory-adjustments/:id/cancel
func (h *InventoryAdjustmentHandler) CancelInventoryAdjustment(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	adjustmentID := c.Param("id")
	if adjustmentID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Adjustment ID is required"))
		return
	}

	var req dto.CancelAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get client info for audit logging
	ipAddress, userAgent := h.getClientInfo(c)

	adjustment, err := h.adjustmentService.CancelInventoryAdjustment(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		adjustmentID,
		userIDStr,
		&req,
		ipAddress,
		userAgent,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapInventoryAdjustmentToResponse(adjustment)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getClientInfo extracts IP address and User-Agent from request for audit logging
func (h *InventoryAdjustmentHandler) getClientInfo(c *gin.Context) (ipAddress, userAgent string) {
	// Get IP address (check forwarded headers first for proxy support)
	ipAddress = c.ClientIP()

	// Get User-Agent
	userAgent = c.Request.UserAgent()

	return ipAddress, userAgent
}

func (h *InventoryAdjustmentHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		formattedErrors := make([]pkgerrors.ValidationError, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			formattedErrors = append(formattedErrors, pkgerrors.ValidationError{
				Field:   fieldErr.Field(),
				Message: fieldErr.Error(),
			})
		}
		c.JSON(http.StatusBadRequest, pkgerrors.NewValidationError(formattedErrors))
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError(err.Error()))
}

func mapInventoryAdjustmentToResponse(adjustment *models.InventoryAdjustment) dto.InventoryAdjustmentResponse {
	response := dto.InventoryAdjustmentResponse{
		ID:               adjustment.ID,
		AdjustmentNumber: adjustment.AdjustmentNumber,
		AdjustmentDate:   adjustment.AdjustmentDate.Format("2006-01-02"),
		WarehouseID:      adjustment.WarehouseID,
		AdjustmentType:   string(adjustment.AdjustmentType),
		Reason:           string(adjustment.Reason),
		Status:           string(adjustment.Status),
		Notes:            adjustment.Notes,
		TotalItems:       adjustment.TotalItems,
		TotalValue:       adjustment.TotalValue.String(),
		CreatedBy:        adjustment.CreatedBy,
		ApprovedBy:       adjustment.ApprovedBy,
		ApprovedAt:       adjustment.ApprovedAt,
		CancelledBy:      adjustment.CancelledBy,
		CancelledAt:      adjustment.CancelledAt,
		CancelReason:     adjustment.CancelReason,
		CreatedAt:        adjustment.CreatedAt,
		UpdatedAt:        adjustment.UpdatedAt,
	}

	// Map warehouse
	if adjustment.Warehouse.ID != "" {
		response.Warehouse = &dto.WarehouseBasicResponse{
			ID:   adjustment.Warehouse.ID,
			Code: adjustment.Warehouse.Code,
			Name: adjustment.Warehouse.Name,
		}
	}

	// Map created by user
	if adjustment.CreatedByUser.ID != "" {
		response.CreatedByUser = &dto.UserBasicResponse{
			ID:       adjustment.CreatedByUser.ID,
			Email:    adjustment.CreatedByUser.Email,
			FullName: adjustment.CreatedByUser.FullName,
		}
	}

	// Map approved by user
	if adjustment.ApprovedByUser != nil && adjustment.ApprovedByUser.ID != "" {
		response.ApprovedByUser = &dto.UserBasicResponse{
			ID:       adjustment.ApprovedByUser.ID,
			Email:    adjustment.ApprovedByUser.Email,
			FullName: adjustment.ApprovedByUser.FullName,
		}
	}

	// Map items
	if len(adjustment.Items) > 0 {
		response.Items = make([]dto.InventoryAdjustmentItemResponse, len(adjustment.Items))
		for i, item := range adjustment.Items {
			response.Items[i] = dto.InventoryAdjustmentItemResponse{
				ID:               item.ID,
				ProductID:        item.ProductID,
				BatchID:          item.BatchID,
				QuantityBefore:   item.QuantityBefore.String(),
				QuantityAdjusted: item.QuantityAdjusted.String(),
				QuantityAfter:    item.QuantityAfter.String(),
				UnitCost:         item.UnitCost.String(),
				TotalValue:       item.TotalValue.String(),
				Notes:            item.Notes,
				CreatedAt:        item.CreatedAt,
				UpdatedAt:        item.UpdatedAt,
			}

			// Map product if available
			if item.Product.ID != "" {
				response.Items[i].Product = &dto.ProductBasicResponse{
					ID:   item.Product.ID,
					Code: item.Product.Code,
					Name: item.Product.Name,
				}
			}
		}
	}

	return response
}
