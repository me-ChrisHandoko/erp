package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/dto"
	"backend/internal/service/supplierpayment"
	"backend/models"
)

// SupplierPaymentHandler handles HTTP requests for supplier payments
type SupplierPaymentHandler struct {
	service *supplierpayment.SupplierPaymentService
}

// NewSupplierPaymentHandler creates a new supplier payment handler
func NewSupplierPaymentHandler(service *supplierpayment.SupplierPaymentService) *SupplierPaymentHandler {
	return &SupplierPaymentHandler{service: service}
}

// ============================================================================
// LIST & GET Handlers
// ============================================================================

// ListSupplierPayments handles GET /api/v1/supplier-payments
func (h *SupplierPaymentHandler) ListSupplierPayments(c *gin.Context) {
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
	var filters dto.SupplierPaymentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Get supplier payments
	payments, pagination, err := h.service.ListSupplierPayments(c.Request.Context(), tenantID.(string), companyID.(string), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to retrieve supplier payments",
			"details": err.Error(),
		})
		return
	}

	// Convert to response DTOs
	responseData := make([]dto.SupplierPaymentResponse, len(payments))
	for i, payment := range payments {
		responseData[i] = mapSupplierPaymentToResponse(payment)
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"data":       responseData,
			"pagination": pagination,
		},
	})
}

// GetSupplierPayment handles GET /api/v1/supplier-payments/:id
func (h *SupplierPaymentHandler) GetSupplierPayment(c *gin.Context) {
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

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Get payment
	payment, err := h.service.GetSupplierPayment(c.Request.Context(), tenantID.(string), companyID.(string), paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Supplier payment not found",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mapSupplierPaymentToResponse(*payment),
	})
}

// ============================================================================
// CREATE Handler
// ============================================================================

// CreateSupplierPayment handles POST /api/v1/supplier-payments
func (h *SupplierPaymentHandler) CreateSupplierPayment(c *gin.Context) {
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

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User context not found",
		})
		return
	}

	// Parse request body
	var req dto.CreateSupplierPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create payment
	payment, err := h.service.CreateSupplierPayment(c.Request.Context(), tenantID.(string), companyID.(string), userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create supplier payment",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    mapSupplierPaymentToResponse(*payment),
	})
}

// ============================================================================
// UPDATE Handler
// ============================================================================

// UpdateSupplierPayment handles PUT /api/v1/supplier-payments/:id
func (h *SupplierPaymentHandler) UpdateSupplierPayment(c *gin.Context) {
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

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User context not found",
		})
		return
	}

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Parse request body
	var req dto.UpdateSupplierPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Update payment
	payment, err := h.service.UpdateSupplierPayment(c.Request.Context(), tenantID.(string), companyID.(string), userID.(string), paymentID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update supplier payment",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mapSupplierPaymentToResponse(*payment),
	})
}

// ============================================================================
// DELETE Handler
// ============================================================================

// DeleteSupplierPayment handles DELETE /api/v1/supplier-payments/:id
func (h *SupplierPaymentHandler) DeleteSupplierPayment(c *gin.Context) {
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

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User context not found",
		})
		return
	}

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Delete payment
	if err := h.service.DeleteSupplierPayment(c.Request.Context(), tenantID.(string), companyID.(string), userID.(string), paymentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete supplier payment",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Supplier payment deleted successfully",
	})
}

// ============================================================================
// APPROVAL Handlers
// ============================================================================

// ApproveSupplierPayment handles POST /api/v1/supplier-payments/:id/approve
func (h *SupplierPaymentHandler) ApproveSupplierPayment(c *gin.Context) {
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

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User context not found",
		})
		return
	}

	// Get payment ID from URL
	paymentID := c.Param("id")

	// Parse request body
	var req dto.ApproveSupplierPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Approve payment
	payment, err := h.service.ApproveSupplierPayment(c.Request.Context(), tenantID.(string), companyID.(string), userID.(string), paymentID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to approve supplier payment",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mapSupplierPaymentToResponse(*payment),
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapSupplierPaymentToResponse converts a SupplierPayment model to a response DTO
func mapSupplierPaymentToResponse(payment models.SupplierPayment) dto.SupplierPaymentResponse {
	response := dto.SupplierPaymentResponse{
		ID:            payment.ID,
		PaymentNumber: payment.PaymentNumber,
		PaymentDate:   payment.PaymentDate.Format("2006-01-02"),
		SupplierID:    payment.SupplierID,
		Amount:        payment.Amount.String(),
		PaymentMethod: string(payment.PaymentMethod),
		Reference:     payment.Reference,
		BankAccountID: payment.BankAccountID,
		Notes:         payment.Notes,
		CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
	}

	// Add supplier information
	if payment.Supplier.ID != "" {
		response.SupplierName = payment.Supplier.Name
		response.SupplierCode = &payment.Supplier.Code
	}

	// Add purchase order information
	if payment.PurchaseOrder != nil && payment.PurchaseOrder.ID != "" {
		response.PurchaseOrderID = &payment.PurchaseOrder.ID
		response.PONumber = &payment.PurchaseOrder.PONumber
	}

	// Add bank account information
	if payment.BankAccount != nil && payment.BankAccount.ID != "" {
		bankName := payment.BankAccount.BankName + " - " + payment.BankAccount.AccountNumber
		response.BankAccountName = &bankName
	}

	// Add approval information
	if payment.ApprovedBy != nil {
		response.ApprovedBy = payment.ApprovedBy
		if payment.ApprovedAt != nil {
			approvedAtStr := payment.ApprovedAt.Format(time.RFC3339)
			response.ApprovedAt = &approvedAtStr
		}
	}

	// TODO: Add CreatedBy and UpdatedBy user information from audit logs

	return response
}
