package handler

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/sales"
	"backend/models"
	"backend/pkg/errors"
)

// DeliveryHandler handles HTTP requests for delivery management
type DeliveryHandler struct {
	deliveryService *sales.DeliveryService
}

// NewDeliveryHandler creates a new delivery handler
func NewDeliveryHandler(deliveryService *sales.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryService: deliveryService,
	}
}

// ============================================================================
// DELIVERY CRUD ENDPOINTS
// ============================================================================

// CreateDelivery creates a new delivery
// POST /api/v1/deliveries
func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
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

	var req dto.CreateDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service with audit info
	deliveryModel, err := h.deliveryService.CreateDelivery(c.Request.Context(), companyID.(string), tenantID.(string), userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapDeliveryToResponse(deliveryModel)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetDelivery retrieves a delivery by ID
// GET /api/v1/deliveries/:id
func (h *DeliveryHandler) GetDelivery(c *gin.Context) {
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

	deliveryID := c.Param("id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Delivery ID is required"))
		return
	}

	// Get delivery from service
	delivery, err := h.deliveryService.GetDeliveryByID(c.Request.Context(), companyID.(string), tenantID.(string), deliveryID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapDeliveryToResponse(delivery)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListDeliveries retrieves deliveries with pagination and filters
// GET /api/v1/deliveries
func (h *DeliveryHandler) ListDeliveries(c *gin.Context) {
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

	// Bind query parameters to filters
	var filters dto.DeliveryFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError(err.Error()))
		return
	}

	// Call service
	deliveries, total, err := h.deliveryService.ListDeliveries(c.Request.Context(), companyID.(string), tenantID.(string), filters)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map deliveries to responses
	deliveryResponses := make([]dto.DeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		deliveryResponses[i] = h.mapDeliveryToResponse(&delivery)
	}

	// Calculate pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}

	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"data": deliveryResponses,
			"pagination": gin.H{
				"page":        page,
				"page_size":   limit,
				"total_items": total,
				"total_pages": totalPages,
			},
		},
	})
}

// ============================================================================
// DELIVERY STATUS ENDPOINTS
// ============================================================================

// UpdateDeliveryStatus updates delivery status and related fields
// PUT /api/v1/deliveries/:id/status
func (h *DeliveryHandler) UpdateDeliveryStatus(c *gin.Context) {
	companyID, tenantID, _, _, _, deliveryID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.UpdateDeliveryStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	deliveryModel, err := h.deliveryService.UpdateDeliveryStatus(c.Request.Context(), companyID, tenantID, deliveryID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapDeliveryToResponse(deliveryModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Delivery status updated successfully",
	})
}

// StartDelivery moves delivery from PREPARED to IN_TRANSIT
// POST /api/v1/deliveries/:id/start
func (h *DeliveryHandler) StartDelivery(c *gin.Context) {
	companyID, tenantID, _, _, _, deliveryID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.StartDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	deliveryModel, err := h.deliveryService.StartDelivery(c.Request.Context(), companyID, tenantID, deliveryID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapDeliveryToResponse(deliveryModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Delivery started successfully",
	})
}

// CompleteDelivery moves delivery from IN_TRANSIT to DELIVERED
// POST /api/v1/deliveries/:id/complete
func (h *DeliveryHandler) CompleteDelivery(c *gin.Context) {
	companyID, tenantID, _, _, _, deliveryID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.CompleteDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	deliveryModel, err := h.deliveryService.CompleteDelivery(c.Request.Context(), companyID, tenantID, deliveryID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapDeliveryToResponse(deliveryModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Delivery completed successfully",
	})
}

// ConfirmDelivery moves delivery from DELIVERED to CONFIRMED
// POST /api/v1/deliveries/:id/confirm
func (h *DeliveryHandler) ConfirmDelivery(c *gin.Context) {
	companyID, tenantID, _, _, _, deliveryID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	deliveryModel, err := h.deliveryService.ConfirmDelivery(c.Request.Context(), companyID, tenantID, deliveryID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapDeliveryToResponse(deliveryModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Delivery confirmed successfully",
	})
}

// CancelDelivery cancels a delivery
// POST /api/v1/deliveries/:id/cancel
func (h *DeliveryHandler) CancelDelivery(c *gin.Context) {
	companyID, tenantID, _, _, _, deliveryID, ok := h.getContextInfo(c)
	if !ok {
		return
	}

	var req dto.CancelDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	deliveryModel, err := h.deliveryService.CancelDelivery(c.Request.Context(), companyID, tenantID, deliveryID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := h.mapDeliveryToResponse(deliveryModel)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Delivery cancelled successfully",
	})
}

// ============================================================================
// PDF GENERATION ENDPOINTS
// ============================================================================

// DownloadDeliveryPDF generates and downloads delivery note PDF
// GET /api/v1/deliveries/:id/pdf
func (h *DeliveryHandler) DownloadDeliveryPDF(c *gin.Context) {
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

	deliveryID := c.Param("id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Delivery ID is required"))
		return
	}

	// Get delivery from service with all preloaded data
	delivery, err := h.deliveryService.GetDeliveryByID(c.Request.Context(), companyID.(string), tenantID.(string), deliveryID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Generate PDF
	pdfBytes, err := h.deliveryService.GenerateDeliveryNotePDF(delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewInternalError(err))
		return
	}

	// Set headers for PDF download
	filename := fmt.Sprintf("Surat_Jalan_%s.pdf", delivery.DeliveryNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Write PDF bytes to response
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapDeliveryToResponse converts delivery model to response DTO
func (h *DeliveryHandler) mapDeliveryToResponse(delivery *models.Delivery) dto.DeliveryResponse {
	response := dto.DeliveryResponse{
		Id:                delivery.ID,
		DeliveryNumber:    delivery.DeliveryNumber,
		DeliveryDate:      delivery.DeliveryDate,
		SalesOrderId:      delivery.SalesOrderID,
		WarehouseId:       delivery.WarehouseID,
		CustomerId:        delivery.CustomerID,
		Type:              string(delivery.Type),
		Status:            string(delivery.Status),
		DeliveryAddress:   delivery.DeliveryAddress,
		DriverName:        delivery.DriverName,
		VehicleNumber:     delivery.VehicleNumber,
		ExpeditionService: delivery.ExpeditionService,
		TtnkNumber:        delivery.TTNKNumber,
		DepartureTime:     delivery.DepartureTime,
		ArrivalTime:       delivery.ArrivalTime,
		ReceivedBy:        delivery.ReceivedBy,
		ReceivedAt:        delivery.ReceivedAt,
		SignatureUrl:      delivery.SignatureURL,
		PhotoUrl:          delivery.PhotoURL,
		Notes:             delivery.Notes,
		CreatedAt:         delivery.CreatedAt,
		UpdatedAt:         delivery.UpdatedAt,
	}

	// Map sales order if present
	if delivery.SalesOrder.ID != "" {
		response.SalesOrder = &dto.SalesOrderSummary{
			Id:       delivery.SalesOrder.ID,
			SoNumber: delivery.SalesOrder.SONumber,
			SoDate:   delivery.SalesOrder.SODate.Format(time.RFC3339),
			Status:   string(delivery.SalesOrder.Status),
		}
	}

	// Map warehouse if present
	if delivery.Warehouse.ID != "" {
		response.Warehouse = &dto.WarehouseSummary{
			Id:   delivery.Warehouse.ID,
			Code: delivery.Warehouse.Code,
			Name: delivery.Warehouse.Name,
		}
	}

	// Map customer if present
	if delivery.Customer.ID != "" {
		response.Customer = &dto.CustomerSummary{
			Id:    delivery.Customer.ID,
			Code:  delivery.Customer.Code,
			Name:  delivery.Customer.Name,
			Phone: delivery.Customer.Phone,
		}
	}

	// Map items
	if len(delivery.Items) > 0 {
		items := make([]dto.DeliveryItemResponse, len(delivery.Items))
		for i, item := range delivery.Items {
			itemResponse := dto.DeliveryItemResponse{
				Id:               item.ID,
				DeliveryId:       item.DeliveryID,
				SalesOrderItemId: &item.SalesOrderItemID,
				ProductId:        item.ProductID,
				ProductUnitId:    item.ProductUnitID,
				BatchId:          item.BatchID,
				Quantity:         item.Quantity.String(),
				Notes:            item.Notes,
				CreatedAt:        item.CreatedAt,
				UpdatedAt:        item.UpdatedAt,
			}

			// Product info
			if item.Product.ID != "" {
				itemResponse.Product = &dto.ProductSummary{
					Id:       item.Product.ID,
					Code:     item.Product.Code,
					Name:     item.Product.Name,
					BaseUnit: item.Product.BaseUnit,
				}
			}

			// Product unit info
			if item.ProductUnitID != nil && item.ProductUnit != nil && item.ProductUnit.ID != "" {
				itemResponse.ProductUnit = &dto.ProductUnitSummary{
					Id:   item.ProductUnit.ID,
					Name: item.ProductUnit.UnitName,
				}
			}

			// Batch info
			if item.BatchID != nil && item.Batch != nil && item.Batch.ID != "" {
				itemResponse.Batch = &dto.BatchSummary{
					Id:          item.Batch.ID,
					BatchNumber: item.Batch.BatchNumber,
					ExpiryDate:  item.Batch.ExpiryDate,
				}
			}

			items[i] = itemResponse
		}
		response.Items = items
	}

	return response
}

// getContextInfo extracts common context information
func (h *DeliveryHandler) getContextInfo(c *gin.Context) (companyID, tenantID, userID, ipAddress, userAgent, deliveryID string, ok bool) {
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

	deliveryID = c.Param("id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Delivery ID is required"))
		return "", "", "", "", "", "", false
	}

	return companyID, tenantID, userID, ipAddress, userAgent, deliveryID, true
}

// handleValidationError handles validation errors
func (h *DeliveryHandler) handleValidationError(c *gin.Context, err error) {
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
func (h *DeliveryHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, appErr)
		return
	}
	c.JSON(http.StatusInternalServerError, errors.NewInternalError(err))
}
