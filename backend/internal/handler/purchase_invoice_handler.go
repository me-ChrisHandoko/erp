package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/dto"
	"backend/internal/service/purchaseinvoice"
	"backend/models"
)

// PurchaseInvoiceHandler handles HTTP requests for purchase invoices
type PurchaseInvoiceHandler struct {
	service *purchaseinvoice.PurchaseInvoiceService
}

// NewPurchaseInvoiceHandler creates a new purchase invoice handler
func NewPurchaseInvoiceHandler(service *purchaseinvoice.PurchaseInvoiceService) *PurchaseInvoiceHandler {
	return &PurchaseInvoiceHandler{service: service}
}

// ============================================================================
// LIST & GET Handlers
// ============================================================================

// ListPurchaseInvoices handles GET /api/v1/purchase-invoices
func (h *PurchaseInvoiceHandler) ListPurchaseInvoices(c *gin.Context) {
	// Get tenant and company context
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	// Parse filters from query params
	var filters dto.PurchaseInvoiceFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		// Log binding error
		println("❌ ERROR [BindQuery]:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Log filters for debugging
	println("✅ DEBUG [Filters]:", "page=", filters.Page, "limit=", filters.Limit, "sortBy=", filters.SortBy, "sortOrder=", filters.SortOrder)

	// Get purchase invoices
	invoices, pagination, err := h.service.ListPurchaseInvoices(c.Request.Context(), tenantID.(string), companyID.(string), filters)
	if err != nil {
		// Log service error
		println("❌ ERROR [Service]:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve purchase invoices",
			"details": err.Error(),
		})
		return
	}

	println("✅ DEBUG [Service returned]:", len(invoices), "invoices")

	// Convert to response DTOs
	responseInvoices := make([]dto.PurchaseInvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		responseInvoices[i] = convertToInvoiceResponse(&invoice)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseInvoices,
		"pagination": pagination,
	})
}

// GetPurchaseInvoice handles GET /api/v1/purchase-invoices/:id
func (h *PurchaseInvoiceHandler) GetPurchaseInvoice(c *gin.Context) {
	// Get tenant and company context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	invoiceID := c.Param("id")

	// Get purchase invoice
	invoice, err := h.service.GetPurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ============================================================================
// CREATE Handler
// ============================================================================

// CreatePurchaseInvoice handles POST /api/v1/purchase-invoices
func (h *PurchaseInvoiceHandler) CreatePurchaseInvoice(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Parse request body
	var req dto.CreatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		println("❌ ERROR [CreatePurchaseInvoice] Invalid request body:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Debug log the request
	println("✅ DEBUG [CreatePurchaseInvoice] Request received")
	println("  - SupplierID:", req.SupplierID)
	println("  - InvoiceDate:", req.InvoiceDate)
	println("  - DueDate:", req.DueDate)
	println("  - Items count:", len(req.Items))
	for i, item := range req.Items {
		println("  - Item", i, "ProductID:", item.ProductID, "UnitID:", item.UnitID, "Qty:", item.Quantity)
	}

	// Create purchase invoice
	invoice, err := h.service.CreatePurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), userIDStr, ipAddress, userAgent, req)
	if err != nil {
		println("❌ ERROR [CreatePurchaseInvoice] Service error:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to create purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Purchase invoice created successfully",
		"data":    response,
	})
}

// ============================================================================
// UPDATE Handler
// ============================================================================

// UpdatePurchaseInvoice handles PUT /api/v1/purchase-invoices/:id
func (h *PurchaseInvoiceHandler) UpdatePurchaseInvoice(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Parse request body
	var req dto.UpdatePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update purchase invoice
	invoice, err := h.service.UpdatePurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userIDStr, ipAddress, userAgent, req)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to update purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase invoice updated successfully",
		"data":    response,
	})
}

// ============================================================================
// DELETE Handler
// ============================================================================

// DeletePurchaseInvoice handles DELETE /api/v1/purchase-invoices/:id
func (h *PurchaseInvoiceHandler) DeletePurchaseInvoice(c *gin.Context) {
	// Get tenant and company context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	// Get user ID from context for audit trail
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User context not found",
		})
		return
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Delete purchase invoice
	if err := h.service.DeletePurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userID.(string), ipAddress, userAgent); err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to delete purchase invoice",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase invoice deleted successfully",
	})
}

// ============================================================================
// WORKFLOW Handlers
// ============================================================================

// SubmitPurchaseInvoice handles POST /api/v1/purchase-invoices/:id/submit
func (h *PurchaseInvoiceHandler) SubmitPurchaseInvoice(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Submit purchase invoice
	invoice, err := h.service.SubmitPurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userIDStr, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to submit purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase invoice submitted successfully",
		"data":    response,
	})
}

// ApprovePurchaseInvoice handles POST /api/v1/purchase-invoices/:id/approve
func (h *PurchaseInvoiceHandler) ApprovePurchaseInvoice(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Parse request body (optional notes)
	var req dto.ApprovePurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Ignore binding errors for optional body
		req = dto.ApprovePurchaseInvoiceRequest{}
	}

	// Approve purchase invoice
	invoice, err := h.service.ApprovePurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userIDStr, ipAddress, userAgent, req)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to approve purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase invoice approved successfully",
		"data":    response,
	})
}

// RejectPurchaseInvoice handles POST /api/v1/purchase-invoices/:id/reject
func (h *PurchaseInvoiceHandler) RejectPurchaseInvoice(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Parse request body
	var req dto.RejectPurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Reject purchase invoice
	invoice, err := h.service.RejectPurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userIDStr, ipAddress, userAgent, req)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to reject purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase invoice rejected successfully",
		"data":    response,
	})
}

// RecordPayment handles POST /api/v1/purchase-invoices/:id/payment
func (h *PurchaseInvoiceHandler) RecordPayment(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Parse request body
	var req dto.RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Record payment
	payment, err := h.service.RecordPayment(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userIDStr, ipAddress, userAgent, req)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to record payment",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToPaymentResponse(payment)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Payment recorded successfully",
		"data":    response,
	})
}

// CancelPurchaseInvoice handles POST /api/v1/purchase-invoices/:id/cancel
func (h *PurchaseInvoiceHandler) CancelPurchaseInvoice(c *gin.Context) {
	// Get tenant, company, and user context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tenant context not found",
		})
		return
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Company context not found",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userIDStr := ""
	if userID != nil {
		userIDStr = userID.(string)
	}

	// Get IP address and user agent for audit logging
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	invoiceID := c.Param("id")

	// Parse request body
	var req dto.CancelPurchaseInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Cancel purchase invoice
	invoice, err := h.service.CancelPurchaseInvoice(c.Request.Context(), tenantID.(string), companyID.(string), invoiceID, userIDStr, ipAddress, userAgent, req)
	if err != nil {
		if err.Error() == "purchase invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Purchase invoice not found",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to cancel purchase invoice",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTO
	response := convertToInvoiceResponse(invoice)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Purchase invoice cancelled successfully",
		"data":    response,
	})
}

// ============================================================================
// HELPER FUNCTIONS - Model to DTO Conversion
// ============================================================================

// convertToInvoiceResponse converts model to response DTO
func convertToInvoiceResponse(invoice *models.PurchaseInvoice) dto.PurchaseInvoiceResponse {
	response := dto.PurchaseInvoiceResponse{
		ID:                   invoice.ID,
		InvoiceNumber:        invoice.InvoiceNumber,
		InvoiceDate:          invoice.InvoiceDate.Format("2006-01-02"),
		DueDate:              invoice.DueDate.Format("2006-01-02"),
		SupplierID:           invoice.SupplierID,
		SupplierName:         invoice.SupplierName,
		SupplierCode:         invoice.SupplierCode,
		PurchaseOrderID:      invoice.PurchaseOrderID,
		GoodsReceiptID:       invoice.GoodsReceiptID,
		SubtotalAmount:       invoice.SubtotalAmount.String(),
		DiscountAmount:       invoice.DiscountAmount.String(),
		TaxAmount:            invoice.TaxAmount.String(),
		TaxRate:              invoice.TaxRate.String(),
		TotalAmount:          invoice.TotalAmount.String(),
		PaidAmount:           invoice.PaidAmount.String(),
		RemainingAmount:      invoice.RemainingAmount.String(),
		ShippingCost:         invoice.ShippingCost.String(),
		HandlingCost:         invoice.HandlingCost.String(),
		OtherCost:            invoice.OtherCost.String(),
		OtherCostDescription: invoice.OtherCostDescription,
		TotalNonGoodsCost:    invoice.GetTotalNonGoodsCost().String(),
		PaymentTermDays:      invoice.PaymentTermDays,
		Status:               string(invoice.Status),
		PaymentStatus:        string(invoice.PaymentStatus),
		Notes:                invoice.Notes,
		ApprovedBy:           invoice.ApprovedBy,
		RejectedBy:           invoice.RejectedBy,
		RejectedReason:       invoice.RejectedReason,
		TaxInvoiceNumber:     invoice.TaxInvoiceNumber,
		CreatedBy:            invoice.CreatedBy,
		UpdatedBy:            invoice.UpdatedBy,
		CreatedAt:            invoice.CreatedAt,
		UpdatedAt:            invoice.UpdatedAt,
	}

	// Format optional timestamp fields
	if invoice.ApprovedAt != nil {
		approvedAt := invoice.ApprovedAt.Format(time.RFC3339)
		response.ApprovedAt = &approvedAt
	}

	if invoice.RejectedAt != nil {
		rejectedAt := invoice.RejectedAt.Format(time.RFC3339)
		response.RejectedAt = &rejectedAt
	}

	if invoice.TaxInvoiceDate != nil {
		taxDate := invoice.TaxInvoiceDate.Format("2006-01-02")
		response.TaxInvoiceDate = &taxDate
	}

	// Populate PO and GRN numbers from relations (normalized approach)
	if invoice.PurchaseOrder != nil {
		response.PONumber = &invoice.PurchaseOrder.PONumber
	}
	if invoice.GoodsReceipt != nil {
		response.GRNumber = &invoice.GoodsReceipt.GRNNumber
	}

	// Convert items if present
	if len(invoice.Items) > 0 {
		response.Items = make([]dto.PurchaseInvoiceItemResponse, len(invoice.Items))
		for i, item := range invoice.Items {
			response.Items[i] = convertToItemResponse(&item)
		}
	}

	// Convert payments if present
	if len(invoice.Payments) > 0 {
		response.Payments = make([]dto.PurchaseInvoicePaymentResponse, len(invoice.Payments))
		for i, payment := range invoice.Payments {
			response.Payments[i] = convertToPaymentResponse(&payment)
		}
	}

	return response
}

// convertToItemResponse converts item model to response DTO
func convertToItemResponse(item *models.PurchaseInvoiceItem) dto.PurchaseInvoiceItemResponse {
	return dto.PurchaseInvoiceItemResponse{
		ID:                  item.ID,
		PurchaseOrderItemID: item.PurchaseOrderItemID,
		GoodsReceiptItemID:  item.GoodsReceiptItemID,
		ProductID:           item.ProductID,
		ProductCode:         item.ProductCode,
		ProductName:         item.ProductName,
		UnitID:              item.UnitID,
		UnitName:            item.UnitName,
		Quantity:            item.Quantity.String(),
		UnitPrice:           item.UnitPrice.String(),
		DiscountAmount:      item.DiscountAmount.String(),
		DiscountPct:         item.DiscountPct.String(),
		TaxAmount:           item.TaxAmount.String(),
		LineTotal:           item.LineTotal.String(),
		Notes:               item.Notes,
		CreatedAt:           item.CreatedAt,
		UpdatedAt:           item.UpdatedAt,
	}
}

// convertToPaymentResponse converts payment model to response DTO
func convertToPaymentResponse(payment *models.PurchaseInvoicePayment) dto.PurchaseInvoicePaymentResponse {
	return dto.PurchaseInvoicePaymentResponse{
		ID:            payment.ID,
		PaymentNumber: payment.PaymentNumber,
		PaymentDate:   payment.PaymentDate.Format("2006-01-02"),
		Amount:        payment.Amount.String(),
		PaymentMethod: string(payment.PaymentMethod),
		Reference:     payment.Reference,
		BankAccountID: payment.BankAccountID,
		Notes:         payment.Notes,
		CreatedBy:     payment.CreatedBy,
		UpdatedBy:     payment.UpdatedBy,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}
}
