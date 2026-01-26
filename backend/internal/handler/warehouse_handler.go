package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/warehouse"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// WarehouseHandler - HTTP handlers for warehouse management endpoints
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 4 - Warehouse Management
type WarehouseHandler struct {
	warehouseService *warehouse.WarehouseService
}

// NewWarehouseHandler creates a new warehouse handler instance
func NewWarehouseHandler(warehouseService *warehouse.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{
		warehouseService: warehouseService,
	}
}

// ============================================================================
// WAREHOUSE CRUD ENDPOINTS
// ============================================================================

// CreateWarehouse handles POST /api/v1/warehouses
// @Summary Create a new warehouse
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param request body dto.CreateWarehouseRequest true "Warehouse creation request"
// @Success 201 {object} dto.WarehouseResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouses [post]
// @Security BearerAuth
func (h *WarehouseHandler) CreateWarehouse(c *gin.Context) {
	// Get company ID from context (set by CompanyContextMiddleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context (set by CompanyContextMiddleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Parse request body
	var req dto.CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user ID from JWT middleware for audit logging
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create warehouse
	warehouseModel, err := h.warehouseService.CreateWarehouse(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapWarehouseToResponse(warehouseModel)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListWarehouses handles GET /api/v1/warehouses
// @Summary List warehouses with filtering and pagination
// @Tags Warehouses
// @Accept json
// @Produce json
// @Success 200 {object} dto.WarehouseListResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouses [get]
// @Security BearerAuth
func (h *WarehouseHandler) ListWarehouses(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Parse query parameters
	var query dto.WarehouseListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List warehouses
	response, err := h.warehouseService.ListWarehouses(c.Request.Context(), tenantID.(string), companyID.(string), &query)
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

// GetWarehouse handles GET /api/v1/warehouses/:id
// @Summary Get warehouse by ID
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID"
// @Success 200 {object} dto.WarehouseResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouses/{id} [get]
// @Security BearerAuth
func (h *WarehouseHandler) GetWarehouse(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get warehouse ID from path
	warehouseID := c.Param("id")
	if warehouseID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Warehouse ID is required"))
		return
	}

	// Get warehouse
	warehouseModel, err := h.warehouseService.GetWarehouseByID(c.Request.Context(), tenantID.(string), companyID.(string), warehouseID)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapWarehouseToResponse(warehouseModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetWarehouseStockStatus handles GET /api/v1/warehouses/stock-status
// @Summary Get warehouse stock initialization status
// @Tags Warehouses
// @Accept json
// @Produce json
// @Success 200 {object} dto.WarehouseStockStatusListResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouses/stock-status [get]
// @Security BearerAuth
func (h *WarehouseHandler) GetWarehouseStockStatus(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get warehouse stock status
	statusList, err := h.warehouseService.GetWarehouseStockStatus(c.Request.Context(), tenantID.(string), companyID.(string))
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	c.JSON(http.StatusOK, dto.WarehouseStockStatusListResponse{
		Success:    true,
		Warehouses: statusList,
	})
}

// UpdateWarehouse handles PUT /api/v1/warehouses/:id
// @Summary Update an existing warehouse
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID"
// @Param request body dto.UpdateWarehouseRequest true "Warehouse update request"
// @Success 200 {object} dto.WarehouseResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouses/{id} [put]
// @Security BearerAuth
func (h *WarehouseHandler) UpdateWarehouse(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get warehouse ID from path
	warehouseID := c.Param("id")
	if warehouseID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Warehouse ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get user ID from JWT middleware for audit logging
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Update warehouse
	warehouseModel, err := h.warehouseService.UpdateWarehouse(c.Request.Context(), tenantID.(string), companyID.(string), warehouseID, userIDStr, ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO
	response := mapWarehouseToResponse(warehouseModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// DeleteWarehouse handles DELETE /api/v1/warehouses/:id
// @Summary Soft delete a warehouse
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID"
// @Success 204 "No Content"
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouses/{id} [delete]
// @Security BearerAuth
func (h *WarehouseHandler) DeleteWarehouse(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get warehouse ID from path
	warehouseID := c.Param("id")
	if warehouseID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Warehouse ID is required"))
		return
	}

	// Get user ID from JWT middleware for audit logging
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Delete warehouse
	err := h.warehouseService.DeleteWarehouse(c.Request.Context(), tenantID.(string), companyID.(string), warehouseID, userIDStr, ipAddress, userAgent)
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
// WAREHOUSE STOCK ENDPOINTS
// ============================================================================

// ListWarehouseStocks handles GET /api/v1/warehouse-stocks
// @Summary List warehouse stocks with filtering and pagination
// @Tags Warehouse Stocks
// @Accept json
// @Produce json
// @Success 200 {object} dto.WarehouseStockListResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouse-stocks [get]
// @Security BearerAuth
func (h *WarehouseHandler) ListWarehouseStocks(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Parse query parameters
	var query dto.WarehouseStockListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// List warehouse stocks
	response, err := h.warehouseService.ListWarehouseStocks(c.Request.Context(), tenantID.(string), companyID.(string), &query)
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

// UpdateWarehouseStock handles PUT /api/v1/warehouse-stocks/:id
// @Summary Update warehouse stock settings (not quantity)
// @Tags Warehouse Stocks
// @Accept json
// @Produce json
// @Param id path string true "Warehouse Stock ID"
// @Param request body dto.UpdateWarehouseStockRequest true "Warehouse stock update request"
// @Success 200 {object} dto.WarehouseStockResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 404 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouse-stocks/{id} [put]
// @Security BearerAuth
func (h *WarehouseHandler) UpdateWarehouseStock(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get user ID from context (for audit logging)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("User context not found."))
		return
	}

	// Get stock ID from path
	stockID := c.Param("id")
	if stockID == "" {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Stock ID is required"))
		return
	}

	// Parse request body
	var req dto.UpdateWarehouseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP address and User-Agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Update warehouse stock
	stockModel, err := h.warehouseService.UpdateWarehouseStock(c.Request.Context(), tenantID.(string), companyID.(string), stockID, userID.(string), ipAddress, userAgent, &req)
	if err != nil {
		if appErr, ok := err.(*pkgerrors.AppError); ok {
			c.JSON(appErr.StatusCode, appErr)
			return
		}
		c.JSON(http.StatusInternalServerError, pkgerrors.NewInternalError(err))
		return
	}

	// Map to response DTO (Product should already be loaded in service layer)
	response := mapWarehouseStockToResponse(stockModel)

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// INITIAL STOCK SETUP ENDPOINTS
// ============================================================================

// CreateInitialStock handles POST /api/v1/warehouse-stocks/initial-setup
// @Summary Create initial stock for a warehouse
// @Tags Warehouse Stocks
// @Accept json
// @Produce json
// @Param request body dto.InitialStockSetupRequest true "Initial stock setup request"
// @Success 201 {object} dto.InitialStockSetupResponse
// @Failure 400 {object} pkgerrors.ErrorResponse
// @Failure 500 {object} pkgerrors.ErrorResponse
// @Router /api/v1/warehouse-stocks/initial-setup [post]
// @Security BearerAuth
func (h *WarehouseHandler) CreateInitialStock(c *gin.Context) {
	// Get company ID from context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Company context not found. Please set X-Company-ID header."))
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Tenant context not found."))
		return
	}

	// Get user ID from context (for audit logging)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("User context not found."))
		return
	}

	// Parse request body
	var req dto.InitialStockSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP address and User-Agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Create initial stocks
	response, err := h.warehouseService.CreateInitialStock(
		c.Request.Context(),
		tenantID.(string),
		companyID.(string),
		userID.(string),
		ipAddress,
		userAgent,
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

	c.JSON(http.StatusCreated, response)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// handleValidationError handles validation errors from request binding
func (h *WarehouseHandler) handleValidationError(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldErr := range validationErrs {
			errors[fieldErr.Field()] = getWarehouseValidationErrorMessage(fieldErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validation failed",
			"fields": errors,
		})
		return
	}
	c.JSON(http.StatusBadRequest, pkgerrors.NewBadRequestError("Invalid request format"))
}

// getWarehouseValidationErrorMessage returns a user-friendly error message for validation errors
func getWarehouseValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "min":
		return fieldErr.Field() + " must be at least " + fieldErr.Param() + " characters"
	case "max":
		return fieldErr.Field() + " must be at most " + fieldErr.Param() + " characters"
	case "email":
		return fieldErr.Field() + " must be a valid email address"
	case "oneof":
		return fieldErr.Field() + " must be one of: " + fieldErr.Param()
	default:
		return fieldErr.Field() + " is invalid"
	}
}

// mapWarehouseToResponse converts Warehouse model to WarehouseResponse DTO
func mapWarehouseToResponse(warehouse *models.Warehouse) dto.WarehouseResponse {
	var capacity *string
	if warehouse.Capacity != nil {
		cap := warehouse.Capacity.String()
		capacity = &cap
	}

	return dto.WarehouseResponse{
		ID:         warehouse.ID,
		Code:       warehouse.Code,
		Name:       warehouse.Name,
		Type:       string(warehouse.Type),
		Address:    warehouse.Address,
		City:       warehouse.City,
		Province:   warehouse.Province,
		PostalCode: warehouse.PostalCode,
		Phone:      warehouse.Phone,
		Email:      warehouse.Email,
		ManagerID:  warehouse.ManagerID,
		Capacity:   capacity,
		IsActive:   warehouse.IsActive,
		CreatedAt:  warehouse.CreatedAt,
		UpdatedAt:  warehouse.UpdatedAt,
	}
}

// mapWarehouseStockToResponse converts WarehouseStock model to WarehouseStockResponse DTO
func mapWarehouseStockToResponse(stock *models.WarehouseStock) dto.WarehouseStockResponse {
	var lastCountQty *string
	if stock.LastCountQty != nil {
		qty := stock.LastCountQty.String()
		lastCountQty = &qty
	}

	productCode := ""
	productName := ""
	if stock.Product.ID != "" {
		productCode = stock.Product.Code
		productName = stock.Product.Name
	}

	return dto.WarehouseStockResponse{
		ID:            stock.ID,
		WarehouseID:   stock.WarehouseID,
		ProductID:     stock.ProductID,
		ProductCode:   productCode,
		ProductName:   productName,
		Quantity:      stock.Quantity.String(),
		MinimumStock:  stock.MinimumStock.String(),
		MaximumStock:  stock.MaximumStock.String(),
		Location:      stock.Location,
		LastCountDate: stock.LastCountDate,
		LastCountQty:  lastCountQty,
		CreatedAt:     stock.CreatedAt,
		UpdatedAt:     stock.UpdatedAt,
	}
}
