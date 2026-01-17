package handler

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/sales"
	"backend/models"
	"backend/pkg/errors"
)

// SalesOrderHandler handles HTTP requests for sales order management
type SalesOrderHandler struct {
	salesOrderService *sales.SalesOrderService
}

// NewSalesOrderHandler creates a new sales order handler
func NewSalesOrderHandler(salesOrderService *sales.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{
		salesOrderService: salesOrderService,
	}
}

// ============================================================================
// SALES ORDER CRUD ENDPOINTS
// ============================================================================

// CreateSalesOrder creates a new sales order
// POST /api/v1/sales-orders
func (h *SalesOrderHandler) CreateSalesOrder(c *gin.Context) {
	// Get company ID and tenant ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found. Please provide X-Company-ID header."))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
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

	var req dto.CreateSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service with audit info
	salesOrderModel, err := h.salesOrderService.CreateSalesOrder(c.Request.Context(), companyID.(string), tenantID.(string), userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapSalesOrderToResponse(salesOrderModel)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetSalesOrder retrieves a sales order by ID
// GET /api/v1/sales-orders/:id
func (h *SalesOrderHandler) GetSalesOrder(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
		return
	}

	salesOrderID := c.Param("id")
	if salesOrderID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Sales order ID is required"))
		return
	}

	// Call service
	salesOrderModel, err := h.salesOrderService.GetSalesOrder(c.Request.Context(), companyID.(string), tenantID.(string), salesOrderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapSalesOrderToResponse(salesOrderModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListSalesOrders retrieves paginated sales orders with filters
// GET /api/v1/sales-orders
func (h *SalesOrderHandler) ListSalesOrders(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
		return
	}

	var filters dto.SalesOrderFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	salesOrders, total, err := h.salesOrderService.ListSalesOrders(c.Request.Context(), companyID.(string), tenantID.(string), &filters)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	responses := make([]dto.SalesOrderResponse, len(salesOrders))
	for i, so := range salesOrders {
		responses[i] = h.mapSalesOrderToResponse(&so)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"pagination": gin.H{
			"page":       filters.Page,
			"limit":      filters.Limit,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// UpdateSalesOrder updates an existing sales order
// PUT /api/v1/sales-orders/:id
func (h *SalesOrderHandler) UpdateSalesOrder(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	salesOrderID := c.Param("id")
	if salesOrderID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Sales order ID is required"))
		return
	}

	var req dto.UpdateSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	salesOrderModel, err := h.salesOrderService.UpdateSalesOrder(c.Request.Context(), companyID.(string), tenantID.(string), userIDStr, ipAddress, userAgent, salesOrderID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapSalesOrderToResponse(salesOrderModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// DeleteSalesOrder soft deletes a sales order
// DELETE /api/v1/sales-orders/:id
func (h *SalesOrderHandler) DeleteSalesOrder(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	salesOrderID := c.Param("id")
	if salesOrderID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Sales order ID is required"))
		return
	}

	// Call service
	err := h.salesOrderService.DeleteSalesOrder(c.Request.Context(), companyID.(string), tenantID.(string), userIDStr, ipAddress, userAgent, salesOrderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sales order deleted successfully",
	})
}

// ============================================================================
// SALES ORDER STATUS TRANSITION ENDPOINTS
// ============================================================================

// SubmitSalesOrder transitions from DRAFT to PENDING
// POST /api/v1/sales-orders/:id/submit
func (h *SalesOrderHandler) SubmitSalesOrder(c *gin.Context) {
	h.handleStatusTransition(c, "submit", func(ctx *gin.Context, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID string) (*models.SalesOrder, error) {
		return h.salesOrderService.SubmitSalesOrder(ctx.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID)
	})
}

// ApproveSalesOrder transitions from PENDING to APPROVED
// POST /api/v1/sales-orders/:id/approve
func (h *SalesOrderHandler) ApproveSalesOrder(c *gin.Context) {
	h.handleStatusTransition(c, "approve", func(ctx *gin.Context, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID string) (*models.SalesOrder, error) {
		return h.salesOrderService.ApproveSalesOrder(ctx.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID)
	})
}

// StartProcessingSalesOrder transitions from APPROVED to PROCESSING
// POST /api/v1/sales-orders/:id/start-processing
func (h *SalesOrderHandler) StartProcessingSalesOrder(c *gin.Context) {
	h.handleStatusTransition(c, "start processing", func(ctx *gin.Context, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID string) (*models.SalesOrder, error) {
		return h.salesOrderService.StartProcessingSalesOrder(ctx.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID)
	})
}

// ShipSalesOrder transitions from PROCESSING to SHIPPED
// POST /api/v1/sales-orders/:id/ship
func (h *SalesOrderHandler) ShipSalesOrder(c *gin.Context) {
	companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.ShipSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	salesOrderModel, err := h.salesOrderService.ShipSalesOrder(c.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapSalesOrderToResponse(salesOrderModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Sales order shipped successfully",
	})
}

// DeliverSalesOrder transitions from SHIPPED to DELIVERED
// POST /api/v1/sales-orders/:id/deliver
func (h *SalesOrderHandler) DeliverSalesOrder(c *gin.Context) {
	companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.DeliverSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	salesOrderModel, err := h.salesOrderService.DeliverSalesOrder(c.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapSalesOrderToResponse(salesOrderModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Sales order delivered successfully",
	})
}

// CompleteSalesOrder transitions from DELIVERED to COMPLETED
// POST /api/v1/sales-orders/:id/complete
func (h *SalesOrderHandler) CompleteSalesOrder(c *gin.Context) {
	h.handleStatusTransition(c, "complete", func(ctx *gin.Context, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID string) (*models.SalesOrder, error) {
		return h.salesOrderService.CompleteSalesOrder(ctx.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID)
	})
}

// CancelSalesOrder transitions to CANCELLED
// POST /api/v1/sales-orders/:id/cancel
func (h *SalesOrderHandler) CancelSalesOrder(c *gin.Context) {
	companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.CancelSalesOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	salesOrderModel, err := h.salesOrderService.CancelSalesOrder(c.Request.Context(), companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapSalesOrderToResponse(salesOrderModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Sales order cancelled successfully",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapSalesOrderToResponse maps a models.SalesOrder to dto.SalesOrderResponse
func (h *SalesOrderHandler) mapSalesOrderToResponse(so *models.SalesOrder) dto.SalesOrderResponse {
	response := dto.SalesOrderResponse{
		Id:           so.ID,
		TenantId:     so.TenantID,
		CompanyId:    so.CompanyID,
		OrderNumber:  so.SONumber,
		OrderDate:    so.SODate.Format("2006-01-02T15:04:05Z07:00"),
		CustomerId:   so.CustomerID,
		WarehouseId:  so.WarehouseID,
		Status:       string(so.Status),
		Subtotal:     so.Subtotal.String(),
		Discount:     so.DiscountAmount.String(),
		Tax:          so.TaxAmount.String(),
		ShippingCost: so.ShippingCost.String(),
		TotalAmount:  so.TotalAmount.String(),
		Notes:        so.Notes,
		CreatedAt:    so.CreatedAt,
		UpdatedAt:    so.UpdatedAt,
	}

	// Required date
	if so.RequiredDate != nil {
		requiredDate := so.RequiredDate.Format("2006-01-02T15:04:05Z07:00")
		response.RequiredDate = &requiredDate
	}

	// Delivery date
	if so.DeliveryDate != nil {
		deliveryDate := so.DeliveryDate.Format("2006-01-02T15:04:05Z07:00")
		response.DeliveryDate = &deliveryDate
	}

	// Customer info
	if so.Customer.ID != "" {
		response.CustomerCode = so.Customer.Code
		response.CustomerName = so.Customer.Name
	}

	// Warehouse info
	if so.Warehouse.ID != "" {
		response.WarehouseCode = so.Warehouse.Code
		response.WarehouseName = so.Warehouse.Name
	}

	// Approval info
	if so.ApprovedBy != nil {
		response.ApprovedBy = so.ApprovedBy
	}
	if so.ApprovedAt != nil {
		approvedAt := so.ApprovedAt.Format("2006-01-02T15:04:05Z07:00")
		response.ApprovedAt = &approvedAt
	}

	// Cancellation info
	if so.CancelledBy != nil {
		response.CancelledBy = so.CancelledBy
	}
	if so.CancelledAt != nil {
		cancelledAt := so.CancelledAt.Format("2006-01-02T15:04:05Z07:00")
		response.CancelledAt = &cancelledAt
	}

	// Map items
	if len(so.Items) > 0 {
		items := make([]dto.SalesOrderItemResponse, len(so.Items))
		for i, item := range so.Items {
			itemResponse := dto.SalesOrderItemResponse{
				Id:         item.ID,
				ProductId:  item.ProductID,
				OrderedQty: item.Quantity.String(),
				UnitPrice:  item.UnitPrice.String(),
				Discount:   item.DiscountAmt.String(),
				LineTotal:  item.Subtotal.String(),
				Notes:      item.Notes,
			}

			// Product info
			if item.Product.ID != "" {
				itemResponse.ProductCode = item.Product.Code
				itemResponse.ProductName = item.Product.Name
			}

			// Unit info
			if item.ProductUnitID != nil {
				itemResponse.UnitId = item.ProductUnitID
				if item.ProductUnit != nil && item.ProductUnit.ID != "" {
					itemResponse.UnitName = item.ProductUnit.UnitName
				} else {
					itemResponse.UnitName = item.Product.BaseUnit
				}
			} else {
				itemResponse.UnitName = item.Product.BaseUnit
			}

			items[i] = itemResponse
		}
		response.Items = items
	}

	return response
}

// handleStatusTransition is a helper function for simple status transitions
func (h *SalesOrderHandler) handleStatusTransition(c *gin.Context, actionName string, transitionFunc func(*gin.Context, string, string, string, string, string, string) (*models.SalesOrder, error)) {
	companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	salesOrderModel, err := transitionFunc(c, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapSalesOrderToResponse(salesOrderModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": fmt.Sprintf("Sales order %s successfully", actionName),
	})
}

// getContextInfo extracts common context information
func (h *SalesOrderHandler) getContextInfo(c *gin.Context) (companyID, tenantID, userID, ipAddress, userAgent, salesOrderID string, ok bool) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return "", "", "", "", "", "", false
	}
	companyID = companyIDVal.(string)

	tenantIDVal, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
		return "", "", "", "", "", "", false
	}
	tenantID = tenantIDVal.(string)

	userIDVal, _ := c.Get("user_id")
	if userIDVal != nil {
		userID = userIDVal.(string)
	}

	ipAddress = c.ClientIP()
	userAgent = c.Request.UserAgent()

	salesOrderID = c.Param("id")
	if salesOrderID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Sales order ID is required"))
		return "", "", "", "", "", "", false
	}

	return companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, true
}

// handleValidationError handles validation errors
func (h *SalesOrderHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		formattedErrors := make([]errors.ValidationError, 0, len(validationErrs))
		for _, fe := range validationErrs {
			formattedErrors = append(formattedErrors, errors.ValidationError{
				Field:   fe.Field(),
				Message: fmt.Sprintf("Field '%s' failed validation '%s'", fe.Field(), fe.Tag()),
			})
		}
		c.JSON(http.StatusBadRequest, errors.NewValidationError(formattedErrors))
		return
	}
	c.JSON(http.StatusBadRequest, errors.NewBadRequestError(err.Error()))
}

// handleError handles service errors
func (h *SalesOrderHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, appErr)
		return
	}
	c.JSON(http.StatusInternalServerError, errors.NewInternalError(err))
}
