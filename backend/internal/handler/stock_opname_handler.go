package handler

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/stockopname"
	"backend/models"
	"backend/pkg/errors"
)

// StockOpnameHandler handles HTTP requests for stock opname management
type StockOpnameHandler struct {
	stockOpnameService *stockopname.StockOpnameService
}

// NewStockOpnameHandler creates a new stock opname handler
func NewStockOpnameHandler(stockOpnameService *stockopname.StockOpnameService) *StockOpnameHandler {
	return &StockOpnameHandler{
		stockOpnameService: stockOpnameService,
	}
}

// ============================================================================
// STOCK OPNAME CRUD ENDPOINTS
// ============================================================================

// CreateStockOpname creates a new stock opname
// POST /api/v1/stock-opnames
func (h *StockOpnameHandler) CreateStockOpname(c *gin.Context) {
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

	var req dto.CreateStockOpnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	opnameModel, err := h.stockOpnameService.CreateStockOpname(c.Request.Context(), companyID.(string), tenantID.(string), userIDStr, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapStockOpnameToResponse(opnameModel)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetStockOpname retrieves a stock opname by ID
// GET /api/v1/stock-opnames/:id
func (h *StockOpnameHandler) GetStockOpname(c *gin.Context) {
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

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	// Call service
	opnameModel, err := h.stockOpnameService.GetStockOpname(c.Request.Context(), companyID.(string), tenantID.(string), opnameID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapStockOpnameToResponse(opnameModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListStockOpnames retrieves paginated stock opnames with filters
// GET /api/v1/stock-opnames
func (h *StockOpnameHandler) ListStockOpnames(c *gin.Context) {
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

	var filters dto.StockOpnameFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	opnames, total, err := h.stockOpnameService.ListStockOpnames(c.Request.Context(), tenantID.(string), companyID.(string), &filters)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to responses
	responses := make([]dto.StockOpnameResponse, 0, len(opnames))
	for i := range opnames {
		responses = append(responses, *h.mapStockOpnameToResponse(&opnames[i]))
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	// Get status counts for statistics
	statusCounts, err := h.stockOpnameService.GetStatusCounts(c.Request.Context(), tenantID.(string), companyID.(string))
	if err != nil {
		// Log error but don't fail the request - status counts are optional
		fmt.Printf("⚠️ Failed to get status counts: %v\n", err)
		statusCounts = map[string]int64{
			"draft":       0,
			"in_progress": 0,
			"completed":   0,
			"approved":    0,
			"cancelled":   0,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"pagination": gin.H{
			"page":       filters.Page,
			"pageSize":   filters.Limit,
			"totalItems": int(total),
			"totalPages": totalPages,
			"hasMore":    filters.Page < totalPages,
		},
		"statusCounts": gin.H{
			"draft":       statusCounts["draft"],
			"inProgress":  statusCounts["in_progress"],
			"completed":   statusCounts["completed"],
			"approved":    statusCounts["approved"],
			"cancelled":   statusCounts["cancelled"],
		},
	})
}

// UpdateStockOpname updates a stock opname
// PUT /api/v1/stock-opnames/:id
func (h *StockOpnameHandler) UpdateStockOpname(c *gin.Context) {
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

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	var req dto.UpdateStockOpnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	opnameModel, err := h.stockOpnameService.UpdateStockOpname(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, userIDStr, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapStockOpnameToResponse(opnameModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Stock opname updated successfully",
	})
}

// DeleteStockOpname deletes a stock opname
// DELETE /api/v1/stock-opnames/:id
func (h *StockOpnameHandler) DeleteStockOpname(c *gin.Context) {
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

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	// Call service
	err := h.stockOpnameService.DeleteStockOpname(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, userIDStr, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Stock opname deleted successfully",
	})
}

// ApproveStockOpname approves a stock opname and posts adjustments
// POST /api/v1/stock-opnames/:id/approve
func (h *StockOpnameHandler) ApproveStockOpname(c *gin.Context) {
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

	// Get user ID from JWT middleware
	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	var req dto.ApproveStockOpnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	opnameModel, err := h.stockOpnameService.ApproveStockOpname(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, userIDStr, &req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapStockOpnameToResponse(opnameModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Stock opname approved successfully and adjustments posted",
	})
}

// ============================================================================
// STOCK OPNAME ITEM ENDPOINTS
// ============================================================================

// AddStockOpnameItem adds a new item to a stock opname
// POST /api/v1/stock-opnames/:id/items
func (h *StockOpnameHandler) AddStockOpnameItem(c *gin.Context) {
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

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	var req dto.CreateStockOpnameItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	item, err := h.stockOpnameService.AddStockOpnameItem(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapStockOpnameItemToResponse(item)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Stock opname item added successfully",
	})
}

// UpdateStockOpnameItem updates a stock opname item
// PUT /api/v1/stock-opnames/:id/items/:itemId
func (h *StockOpnameHandler) UpdateStockOpnameItem(c *gin.Context) {
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

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname Item ID is required"))
		return
	}

	var req dto.UpdateStockOpnameItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Call service
	item, err := h.stockOpnameService.UpdateStockOpnameItem(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, itemID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapStockOpnameItemToResponse(item)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Stock opname item updated successfully",
	})
}

// BatchUpdateStockOpnameItems updates multiple stock opname items
// PUT /api/v1/stock-opnames/:id/items/batch
func (h *StockOpnameHandler) BatchUpdateStockOpnameItems(c *gin.Context) {
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

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	var req dto.BatchUpdateStockOpnameItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Get IP and User Agent for audit
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Call service
	items, err := h.stockOpnameService.BatchUpdateStockOpnameItems(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, userIDStr, &req, ipAddress, userAgent)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	var responses []*dto.StockOpnameItemResponse
	for _, item := range items {
		responses = append(responses, h.mapStockOpnameItemToResponse(&item))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"message": "Stock opname items updated successfully",
	})
}

// DeleteStockOpnameItem deletes a stock opname item
// DELETE /api/v1/stock-opnames/:id/items/:itemId
func (h *StockOpnameHandler) DeleteStockOpnameItem(c *gin.Context) {
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

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname Item ID is required"))
		return
	}

	// Call service
	err := h.stockOpnameService.DeleteStockOpnameItem(c.Request.Context(), companyID.(string), tenantID.(string), opnameID, itemID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Stock opname item deleted successfully",
	})
}

// ImportWarehouseProducts imports all products from a warehouse to stock opname
// POST /api/v1/stock-opnames/:id/import-products
func (h *StockOpnameHandler) ImportWarehouseProducts(c *gin.Context) {
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

	opnameID := c.Param("id")
	if opnameID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Stock Opname ID is required"))
		return
	}

	// Call service
	itemsAdded, err := h.stockOpnameService.ImportWarehouseProducts(c.Request.Context(), companyID.(string), tenantID.(string), opnameID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"itemsAdded": itemsAdded,
		"message":    fmt.Sprintf("%d products imported successfully", itemsAdded),
	})
}

// ============================================================================
// MAPPER FUNCTIONS
// ============================================================================

// mapStockOpnameToResponse converts stock opname model to response DTO
func (h *StockOpnameHandler) mapStockOpnameToResponse(opname *models.StockOpname) *dto.StockOpnameResponse {
	response := &dto.StockOpnameResponse{
		ID:           opname.ID,
		CompanyID:    opname.CompanyID,
		TenantID:     opname.TenantID,
		OpnameNumber: opname.OpnameNumber,
		OpnameDate:   opname.OpnameDate.Format("2006-01-02T15:04:05Z07:00"),
		WarehouseID:  opname.WarehouseID,
		Status:       h.mapStatusToString(opname.Status),
		Notes:        opname.Notes,
		CreatedAt:    opname.CreatedAt,
		UpdatedAt:    opname.UpdatedAt,
	}

	// Map warehouse name if available
	if opname.Warehouse.Name != "" {
		response.WarehouseName = &opname.Warehouse.Name
	}

	// Map approved info
	response.ApprovedBy = opname.ApprovedBy
	response.ApprovedAt = opname.ApprovedAt

	// Calculate totals from items
	if len(opname.Items) > 0 {
		response.TotalItems = len(opname.Items)

		totalExpected := opname.Items[0].SystemQty.Copy()
		totalActual := opname.Items[0].PhysicalQty.Copy()
		totalDiff := opname.Items[0].DifferenceQty.Copy()

		for i, item := range opname.Items {
			if i > 0 {
				totalExpected = totalExpected.Add(item.SystemQty)
				totalActual = totalActual.Add(item.PhysicalQty)
				totalDiff = totalDiff.Add(item.DifferenceQty)
			}
		}

		response.TotalExpectedQty = totalExpected.String()
		response.TotalActualQty = totalActual.String()
		response.TotalDifference = totalDiff.String()

		// Map items
		response.Items = make([]dto.StockOpnameItemResponse, 0, len(opname.Items))
		for _, item := range opname.Items {
			response.Items = append(response.Items, *h.mapStockOpnameItemToResponse(&item))
		}
	} else {
		response.TotalItems = 0
		response.TotalExpectedQty = "0"
		response.TotalActualQty = "0"
		response.TotalDifference = "0"
	}

	return response
}

// mapStockOpnameItemToResponse converts stock opname item model to response DTO
func (h *StockOpnameHandler) mapStockOpnameItemToResponse(item *models.StockOpnameItem) *dto.StockOpnameItemResponse {
	response := &dto.StockOpnameItemResponse{
		ID:          item.ID,
		OpnameID:    item.StockOpnameID,
		ProductID:   item.ProductID,
		ExpectedQty: item.SystemQty.String(),
		ActualQty:   item.PhysicalQty.String(),
		Difference:  item.DifferenceQty.String(),
		Notes:       item.Notes,
	}

	// Map product info if preloaded
	if item.Product.Code != "" {
		response.ProductCode = &item.Product.Code
	}
	if item.Product.Name != "" {
		response.ProductName = &item.Product.Name
	}

	return response
}

// mapStatusToString converts StockOpnameStatus enum to string
func (h *StockOpnameHandler) mapStatusToString(status models.StockOpnameStatus) string {
	switch status {
	case models.StockOpnameStatusDraft:
		return "draft"
	case models.StockOpnameStatusInProgress:
		return "in_progress"
	case models.StockOpnameStatusCompleted:
		return "completed"
	case models.StockOpnameStatusApproved:
		return "approved"
	case models.StockOpnameStatusCancelled:
		return "cancelled"
	default:
		return string(status)
	}
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

// handleValidationError formats and returns validation errors
func (h *StockOpnameHandler) handleValidationError(c *gin.Context, err error) {
	// Check if it's a validator error
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		formattedErrors := make([]errors.ValidationError, 0, len(validationErrs))

		for _, fieldErr := range validationErrs {
			formattedErrors = append(formattedErrors, errors.ValidationError{
				Field:   getJSONFieldName(fieldErr),
				Message: formatValidationMessage(fieldErr),
			})
		}

		appErr := errors.NewValidationError(formattedErrors)
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

	// Not a validation error, return generic error
	validationErrors := []errors.ValidationError{
		{
			Field:   "request",
			Message: err.Error(),
		},
	}
	appErr := errors.NewValidationError(validationErrors)
	c.JSON(appErr.StatusCode, gin.H{
		"success": false,
		"error":   appErr,
	})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *StockOpnameHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error":   appErr,
		})
		return
	}

	// Log the actual error for debugging
	fmt.Printf("❌ INTERNAL ERROR: %v\n", err)

	// Unknown error - return internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "An unexpected error occurred",
		},
	})
}
