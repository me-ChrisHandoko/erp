package handler

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"backend/internal/dto"
	"backend/internal/service/product"
	"backend/models"
	"backend/pkg/errors"
)

// ProductHandler handles HTTP requests for product management
// Reference: 02-MASTER-DATA-MANAGEMENT.md Module 1: Product Management
type ProductHandler struct {
	productService *product.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *product.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// ============================================================================
// PRODUCT CRUD ENDPOINTS
// ============================================================================

// CreateProduct creates a new product
// POST /api/v1/products
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 257-307
func (h *ProductHandler) CreateProduct(c *gin.Context) {
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

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	productModel, err := h.productService.CreateProduct(c.Request.Context(), companyID.(string), tenantID.(string), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapProductToResponse(productModel)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetProduct retrieves a product by ID
// GET /api/v1/products/:id
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 182-254
func (h *ProductHandler) GetProduct(c *gin.Context) {
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

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	// Call service
	productModel, err := h.productService.GetProduct(c.Request.Context(), companyID.(string), tenantID.(string), productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapProductToResponse(productModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListProducts retrieves paginated products with filters
// GET /api/v1/products
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 117-179
func (h *ProductHandler) ListProducts(c *gin.Context) {
	fmt.Println("🔍 DEBUG [ListProducts]: Handler started")

	companyID, exists := c.Get("company_id")
	if !exists {
		fmt.Println("❌ ERROR [ListProducts]: Company context not found")
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}
	fmt.Printf("✅ DEBUG [ListProducts]: Company ID retrieved: %v\n", companyID)

	var filters dto.ProductFilters
	fmt.Printf("🔍 DEBUG [ListProducts]: About to bind query params: %s\n", c.Request.URL.RawQuery)
	if err := c.ShouldBindQuery(&filters); err != nil {
		fmt.Printf("❌ ERROR [ListProducts]: Query binding failed: %v\n", err)
		h.handleValidationError(c, err)
		return
	}
	fmt.Printf("✅ DEBUG [ListProducts]: Query params bound - Page: %d, Limit: %d, SortBy: %s, SortOrder: %s\n",
		filters.Page, filters.Limit, filters.SortBy, filters.SortOrder)

	// Get tenant_id from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		fmt.Println("❌ ERROR [ListProducts]: Tenant context not found")
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Tenant context not found."))
		return
	}
	fmt.Printf("✅ DEBUG [ListProducts]: Tenant ID retrieved: %v\n", tenantID)

	// Call service with explicit tenant_id (no bypass needed!)
	fmt.Println("🔍 DEBUG [ListProducts]: Calling service...")
	products, total, err := h.productService.ListProducts(c.Request.Context(), tenantID.(string), companyID.(string), &filters)
	if err != nil {
		fmt.Printf("❌ ERROR [ListProducts]: Service returned error: %v\n", err)
		h.handleError(c, err)
		return
	}
	fmt.Printf("✅ DEBUG [ListProducts]: Service returned %d products (total: %d)\n", len(products), total)

	// Map to responses
	responses := make([]dto.ProductResponse, 0, len(products))
	for i := range products {
		responses = append(responses, *h.mapProductToResponse(&products[i]))
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(filters.Limit)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"pagination": dto.PaginationInfo{
			Page:       filters.Page,
			Limit:      filters.Limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}

// UpdateProduct updates a product
// PUT /api/v1/products/:id
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 310-325
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
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

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	productModel, err := h.productService.UpdateProduct(c.Request.Context(), companyID.(string), tenantID.(string), productID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapProductToResponse(productModel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Product updated successfully",
	})
}

// DeleteProduct soft deletes a product
// DELETE /api/v1/products/:id
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 328-343
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	// Call service
	err := h.productService.DeleteProduct(c.Request.Context(), companyID.(string), productID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product deactivated successfully",
	})
}

// ============================================================================
// PRODUCT UNIT ENDPOINTS
// ============================================================================

// AddProductUnit adds a new unit to a product
// POST /api/v1/products/:id/units
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 346-350
func (h *ProductHandler) AddProductUnit(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	var req dto.CreateProductUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	unit, err := h.productService.AddProductUnit(c.Request.Context(), companyID.(string), productID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapProductUnitToResponse(unit)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Product unit added successfully",
	})
}

// UpdateProductUnit updates a product unit
// PUT /api/v1/products/:id/units/:unitId
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 346-350
func (h *ProductHandler) UpdateProductUnit(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	unitID := c.Param("unitId")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Unit ID is required"))
		return
	}

	var req dto.UpdateProductUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	unit, err := h.productService.UpdateProductUnit(c.Request.Context(), companyID.(string), productID, unitID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := h.mapProductUnitToResponse(unit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Product unit updated successfully",
	})
}

// DeleteProductUnit soft deletes a product unit
// DELETE /api/v1/products/:id/units/:unitId
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 346-350
func (h *ProductHandler) DeleteProductUnit(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	unitID := c.Param("unitId")
	if unitID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Unit ID is required"))
		return
	}

	// Call service
	err := h.productService.DeleteProductUnit(c.Request.Context(), companyID.(string), productID, unitID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product unit deleted successfully",
	})
}

// ============================================================================
// PRODUCT SUPPLIER ENDPOINTS
// ============================================================================

// AddProductSupplier links a supplier to a product
// POST /api/v1/products/:id/suppliers
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 352-357
func (h *ProductHandler) AddProductSupplier(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	var req dto.AddProductSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	ps, err := h.productService.AddProductSupplier(c.Request.Context(), companyID.(string), productID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response (simple response since we don't have supplier name yet)
	response := gin.H{
		"id":            ps.ID,
		"productId":     ps.ProductID,
		"supplierId":    ps.SupplierID,
		"supplierPrice": ps.SupplierPrice.String(),
		"leadTime":      ps.LeadTime,
		"isPrimary":     ps.IsPrimary,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
		"message": "Supplier linked to product successfully",
	})
}

// UpdateProductSupplier updates a product-supplier relationship
// PUT /api/v1/products/:id/suppliers/:supplierId
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 352-357
func (h *ProductHandler) UpdateProductSupplier(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	productSupplierID := c.Param("supplierId")
	if productSupplierID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product Supplier ID is required"))
		return
	}

	var req dto.UpdateProductSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Call service
	ps, err := h.productService.UpdateProductSupplier(c.Request.Context(), companyID.(string), productID, productSupplierID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Map to response
	response := gin.H{
		"id":            ps.ID,
		"productId":     ps.ProductID,
		"supplierId":    ps.SupplierID,
		"supplierPrice": ps.SupplierPrice.String(),
		"leadTime":      ps.LeadTime,
		"isPrimary":     ps.IsPrimary,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Product supplier updated successfully",
	})
}

// DeleteProductSupplier removes a supplier from a product
// DELETE /api/v1/products/:id/suppliers/:supplierId
// Reference: 02-MASTER-DATA-MANAGEMENT.md lines 352-357
func (h *ProductHandler) DeleteProductSupplier(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Company context not found."))
		return
	}

	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product ID is required"))
		return
	}

	productSupplierID := c.Param("supplierId")
	if productSupplierID == "" {
		c.JSON(http.StatusBadRequest, errors.NewBadRequestError("Product Supplier ID is required"))
		return
	}

	// Call service
	err := h.productService.DeleteProductSupplier(c.Request.Context(), companyID.(string), productID, productSupplierID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Supplier removed from product successfully",
	})
}

// ============================================================================
// MAPPER FUNCTIONS
// ============================================================================

// mapProductToResponse converts product model to response DTO
func (h *ProductHandler) mapProductToResponse(product *models.Product) *dto.ProductResponse {
	response := &dto.ProductResponse{
		ID:             product.ID,
		Code:           product.Code,
		Name:           product.Name,
		Category:       product.Category,
		BaseUnit:       product.BaseUnit,
		BaseCost:       product.BaseCost.String(),
		BasePrice:      product.BasePrice.String(),
		MinimumStock:   product.MinimumStock.String(),
		Description:    product.Description,
		Barcode:        product.Barcode,
		IsBatchTracked: product.IsBatchTracked,
		IsPerishable:   product.IsPerishable,
		IsActive:       product.IsActive,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	}

	// Map units
	if len(product.Units) > 0 {
		response.Units = make([]dto.ProductUnitResponse, 0, len(product.Units))
		for _, unit := range product.Units {
			response.Units = append(response.Units, *h.mapProductUnitToResponse(&unit))
		}
	}

	// Map suppliers (only if preloaded)
	if len(product.ProductSuppliers) > 0 {
		response.Suppliers = make([]dto.ProductSupplierResponse, 0, len(product.ProductSuppliers))
		for _, ps := range product.ProductSuppliers {
			// Get supplier name if available (preloaded)
			supplierName := ""
			// Supplier might not be preloaded in all cases
			// Will be empty if not preloaded - frontend should fetch if needed

			response.Suppliers = append(response.Suppliers, dto.ProductSupplierResponse{
				ID:            ps.ID,
				SupplierID:    ps.SupplierID,
				SupplierName:  supplierName,
				SupplierPrice: ps.SupplierPrice.String(),
				LeadTime:      ps.LeadTime,
				IsPrimary:     ps.IsPrimary,
			})
		}
	}

	// Map current stock
	if len(product.WarehouseStocks) > 0 {
		totalStock := product.WarehouseStocks[0].Quantity.Copy()
		warehouses := make([]dto.WarehouseStockInfo, 0, len(product.WarehouseStocks))

		for i, ws := range product.WarehouseStocks {
			warehouseName := ""
			if ws.Warehouse.Name != "" {
				warehouseName = ws.Warehouse.Name
			}

			warehouses = append(warehouses, dto.WarehouseStockInfo{
				WarehouseID:   ws.WarehouseID,
				WarehouseName: warehouseName,
				Quantity:      ws.Quantity.String(),
			})

			// Sum total stock (skip first since we already initialized with it)
			if i > 0 {
				totalStock = totalStock.Add(ws.Quantity)
			}
		}

		response.CurrentStock = &dto.CurrentStockResponse{
			Total:      totalStock.String(),
			Warehouses: warehouses,
		}
	}

	return response
}

// mapProductUnitToResponse converts product unit model to response DTO
func (h *ProductHandler) mapProductUnitToResponse(unit *models.ProductUnit) *dto.ProductUnitResponse {
	response := &dto.ProductUnitResponse{
		ID:             unit.ID,
		UnitName:       unit.UnitName,
		ConversionRate: unit.ConversionRate.String(),
		IsBaseUnit:     unit.IsBaseUnit,
		Barcode:        unit.Barcode,
		SKU:            unit.SKU,
		Description:    unit.Description,
		IsActive:       unit.IsActive,
	}

	if unit.BuyPrice != nil {
		buyPrice := unit.BuyPrice.String()
		response.BuyPrice = &buyPrice
	}

	if unit.SellPrice != nil {
		sellPrice := unit.SellPrice.String()
		response.SellPrice = &sellPrice
	}

	if unit.Weight != nil {
		weight := unit.Weight.String()
		response.Weight = &weight
	}

	if unit.Volume != nil {
		volume := unit.Volume.String()
		response.Volume = &volume
	}

	return response
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

// handleValidationError formats and returns validation errors
func (h *ProductHandler) handleValidationError(c *gin.Context, err error) {
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
func (h *ProductHandler) handleError(c *gin.Context, err error) {
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
