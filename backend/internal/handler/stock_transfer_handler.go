package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/stock_transfer"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// StockTransferHandler - HTTP handlers for stock transfer endpoints
type StockTransferHandler struct {
	stockTransferService *stock_transfer.StockTransferService
}

// NewStockTransferHandler creates a new stock transfer handler instance
func NewStockTransferHandler(stockTransferService *stock_transfer.StockTransferService) *StockTransferHandler {
	return &StockTransferHandler{
		stockTransferService: stockTransferService,
	}
}

// ============================================================================
// STOCK TRANSFER CRUD ENDPOINTS
// ============================================================================

// CreateStockTransfer handles POST /api/v1/stock-transfers
func (h *StockTransferHandler) CreateStockTransfer(c *gin.Context) {
	log.Println("🔵 [CreateStockTransfer] Handler called")

	// Get context values
	companyID, exists := c.Get("company_id")
	if !exists {
		log.Println("❌ [CreateStockTransfer] Company context not found")
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found"))
		return
	}
	log.Printf("✅ [CreateStockTransfer] Company ID: %v\n", companyID)

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		log.Println("❌ [CreateStockTransfer] Tenant context not found")
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found"))
		return
	}
	log.Printf("✅ [CreateStockTransfer] Tenant ID: %v\n", tenantID)

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}
	log.Printf("✅ [CreateStockTransfer] User ID: %s\n", userIDStr)

	// Parse request
	var req dto.CreateStockTransferRequest
	log.Println("🔍 [CreateStockTransfer] Attempting to bind JSON...")
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ [CreateStockTransfer] JSON binding error: %v\n", err)
		h.handleValidationError(c, err)
		return
	}
	log.Printf("✅ [CreateStockTransfer] Request parsed: %+v\n", req)

	// Create transfer
	log.Println("🔍 [CreateStockTransfer] Calling service.CreateStockTransfer...")
	transfer, err := h.stockTransferService.CreateStockTransfer(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		userIDStr,
		&req,
	)
	if err != nil {
		log.Printf("❌ [CreateStockTransfer] Service error: %v\n", err)
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}
	log.Println("✅ [CreateStockTransfer] Transfer created successfully")

	response := mapStockTransferToResponse(transfer)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListStockTransfers handles GET /api/v1/stock-transfers
func (h *StockTransferHandler) ListStockTransfers(c *gin.Context) {
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
	var query dto.StockTransferQuery
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
		query.SortBy = "transfer_number"
	}
	if query.SortOrder == "" {
		query.SortOrder = "desc"
	}

	transfers, pagination, err := h.stockTransferService.ListStockTransfers(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		&query,
	)
	if err != nil {
		// Log the actual error for debugging
		println("❌ ListStockTransfers error:", err.Error())
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	responses := make([]dto.StockTransferResponse, len(transfers))
	for i, transfer := range transfers {
		responses[i] = mapStockTransferToResponse(&transfer)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       responses,
		"pagination": pagination,
	})
}

// GetStockTransfer handles GET /api/v1/stock-transfers/:id
func (h *StockTransferHandler) GetStockTransfer(c *gin.Context) {
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

	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Transfer ID is required"))
		return
	}

	transfer, err := h.stockTransferService.GetStockTransferByID(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		transferID,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapStockTransferToResponse(transfer)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateStockTransfer handles PUT /api/v1/stock-transfers/:id
func (h *StockTransferHandler) UpdateStockTransfer(c *gin.Context) {
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

	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Transfer ID is required"))
		return
	}

	var req dto.UpdateStockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	transfer, err := h.stockTransferService.UpdateStockTransfer(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		transferID,
		&req,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapStockTransferToResponse(transfer)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// DeleteStockTransfer handles DELETE /api/v1/stock-transfers/:id
func (h *StockTransferHandler) DeleteStockTransfer(c *gin.Context) {
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

	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Transfer ID is required"))
		return
	}

	err := h.stockTransferService.DeleteStockTransfer(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		transferID,
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
		"message": "Stock transfer deleted successfully",
	})
}

// ============================================================================
// STATUS TRANSITION ENDPOINTS
// ============================================================================

// ShipStockTransfer handles POST /api/v1/stock-transfers/:id/ship
func (h *StockTransferHandler) ShipStockTransfer(c *gin.Context) {
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

	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Transfer ID is required"))
		return
	}

	var req dto.ShipTransferRequest
	// Bind JSON, but it's okay if body is empty
	_ = c.ShouldBindJSON(&req)

	transfer, err := h.stockTransferService.ShipStockTransfer(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		transferID,
		userIDStr,
		&req,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapStockTransferToResponse(transfer)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ReceiveStockTransfer handles POST /api/v1/stock-transfers/:id/receive
func (h *StockTransferHandler) ReceiveStockTransfer(c *gin.Context) {
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

	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Transfer ID is required"))
		return
	}

	var req dto.ReceiveTransferRequest
	_ = c.ShouldBindJSON(&req)

	transfer, err := h.stockTransferService.ReceiveStockTransfer(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		transferID,
		userIDStr,
		&req,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapStockTransferToResponse(transfer)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CancelStockTransfer handles POST /api/v1/stock-transfers/:id/cancel
func (h *StockTransferHandler) CancelStockTransfer(c *gin.Context) {
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

	transferID := c.Param("id")
	if transferID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Transfer ID is required"))
		return
	}

	var req dto.CancelTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	transfer, err := h.stockTransferService.CancelStockTransfer(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		transferID,
		&req,
	)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	response := mapStockTransferToResponse(transfer)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func (h *StockTransferHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		// Format validation errors
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

func mapStockTransferToResponse(transfer *models.StockTransfer) dto.StockTransferResponse {
	response := dto.StockTransferResponse{
		ID:                transfer.ID,
		TransferNumber:    transfer.TransferNumber,
		TransferDate:      transfer.TransferDate.Format("2006-01-02"),
		SourceWarehouseID: transfer.SourceWarehouseID,
		DestWarehouseID:   transfer.DestWarehouseID,
		Status:            string(transfer.Status),
		ShippedBy:         transfer.ShippedBy,
		ShippedAt:         transfer.ShippedAt,
		ReceivedBy:        transfer.ReceivedBy,
		ReceivedAt:        transfer.ReceivedAt,
		Notes:             transfer.Notes,
		CreatedAt:         transfer.CreatedAt,
		UpdatedAt:         transfer.UpdatedAt,
	}

	// Map source warehouse
	if transfer.SourceWarehouse.ID != "" {
		response.SourceWarehouse = &dto.WarehouseBasicResponse{
			ID:   transfer.SourceWarehouse.ID,
			Code: transfer.SourceWarehouse.Code,
			Name: transfer.SourceWarehouse.Name,
		}
	}

	// Map destination warehouse
	if transfer.DestWarehouse.ID != "" {
		response.DestWarehouse = &dto.WarehouseBasicResponse{
			ID:   transfer.DestWarehouse.ID,
			Code: transfer.DestWarehouse.Code,
			Name: transfer.DestWarehouse.Name,
		}
	}

	// Map items
	if len(transfer.Items) > 0 {
		response.Items = make([]dto.StockTransferItemResponse, len(transfer.Items))
		for i, item := range transfer.Items {
			response.Items[i] = dto.StockTransferItemResponse{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity.String(),
				BatchID:   item.BatchID,
				Notes:     item.Notes,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
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
