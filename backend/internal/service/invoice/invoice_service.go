package invoice

import (
	"backend/internal/dto"
	"backend/internal/service/document"
	"backend/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InvoiceService handles invoice business logic
type InvoiceService struct {
	db            *gorm.DB
	docNumberGen  *document.DocumentNumberGenerator
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *InvoiceService {
	return &InvoiceService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ListInvoices retrieves invoices with filters and pagination
func (s *InvoiceService) ListInvoices(tenantID, companyID string, filters dto.InvoiceFilters) (*dto.InvoiceListResponse, error) {
	var invoices []models.Invoice
	var total int64

	// Build query with tenant context
	query := s.db.Set("tenant_id", tenantID).Model(&models.Invoice{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("invoice_number LIKE ? OR EXISTS (SELECT 1 FROM customers c WHERE c.id = invoices.customer_id AND (c.name LIKE ? OR c.code LIKE ?))",
			searchPattern, searchPattern, searchPattern)
	}

	if filters.CustomerID != "" {
		query = query.Where("customer_id = ?", filters.CustomerID)
	}

	if filters.PaymentStatus != "" {
		query = query.Where("payment_status = ?", filters.PaymentStatus)
	}

	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err == nil {
			query = query.Where("invoice_date >= ?", dateFrom)
		}
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err == nil {
			query = query.Where("invoice_date <= ?", dateTo)
		}
	}

	if filters.DueDateFrom != "" {
		dueDateFrom, err := time.Parse("2006-01-02", filters.DueDateFrom)
		if err == nil {
			query = query.Where("due_date >= ?", dueDateFrom)
		}
	}

	if filters.DueDateTo != "" {
		dueDateTo, err := time.Parse("2006-01-02", filters.DueDateTo)
		if err == nil {
			query = query.Where("due_date <= ?", dueDateTo)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Apply sorting
	sortBy := "invoice_date"
	if filters.SortBy != "" {
		sortBy = mapSortField(filters.SortBy)
	}
	sortOrder := "desc"
	if filters.SortOrder != "" {
		sortOrder = filters.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	page := 1
	if filters.Page > 0 {
		page = filters.Page
	}
	limit := 20
	if filters.Limit > 0 {
		limit = filters.Limit
	}
	offset := (page - 1) * limit

	// Fetch invoices with customer relation
	if err := query.Preload("Customer").
		Offset(offset).
		Limit(limit).
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch invoices: %w", err)
	}

	// Convert to response DTOs
	invoiceResponses := make([]dto.InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		invoiceResponses[i] = s.toInvoiceResponse(invoice)
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.InvoiceListResponse{
		Data: invoiceResponses,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetInvoice retrieves a single invoice by ID
func (s *InvoiceService) GetInvoice(tenantID, companyID, invoiceID string) (*dto.InvoiceResponse, error) {
	var invoice models.Invoice

	if err := s.db.Set("tenant_id", tenantID).
		Preload("Customer").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		Preload("Payments").
		Where("id = ? AND company_id = ?", invoiceID, companyID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	response := s.toInvoiceResponseWithDetails(invoice)
	return &response, nil
}

// CreateInvoice creates a new invoice
func (s *InvoiceService) CreateInvoice(companyID, tenantID string, req dto.CreateInvoiceRequest) (*dto.InvoiceResponse, error) {
	// Parse dates
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice date: %w", err)
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, fmt.Errorf("invalid due date: %w", err)
	}

	// Generate invoice number
	ctx := context.Background()
	invoiceNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypeSalesInvoice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Parse amounts
	discountAmount := decimal.Zero
	if req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(req.DiscountAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid discount amount: %w", err)
		}
	}

	taxAmount := decimal.Zero
	if req.TaxAmount != "" {
		taxAmount, err = decimal.NewFromString(req.TaxAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid tax amount: %w", err)
		}
	}

	// Parse Faktur Pajak date if provided
	var fakturPajakDate *time.Time
	if req.FakturPajakDate != nil {
		fpDate, err := time.Parse("2006-01-02", *req.FakturPajakDate)
		if err != nil {
			return nil, fmt.Errorf("invalid faktur pajak date: %w", err)
		}
		fakturPajakDate = &fpDate
	}

	// Create invoice
	invoice := models.Invoice{
		TenantID:        tenantID,
		CompanyID:       companyID,
		InvoiceNumber:   invoiceNumber,
		InvoiceDate:     invoiceDate,
		DueDate:         dueDate,
		CustomerID:      req.CustomerID,
		SalesOrderID:    req.SalesOrderID,
		DeliveryID:      req.DeliveryID,
		DiscountAmount:  discountAmount,
		TaxAmount:       taxAmount,
		Notes:           req.Notes,
		FakturPajakNo:   req.FakturPajakNo,
		FakturPajakDate: fakturPajakDate,
		PaymentStatus:   models.PaymentStatusUnpaid,
		PaidAmount:      decimal.Zero,
	}

	// Calculate subtotal and total from items
	subtotal := decimal.Zero
	for _, itemReq := range req.Items {
		qty, err := decimal.NewFromString(itemReq.Quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity: %w", err)
		}

		unitPrice, err := decimal.NewFromString(itemReq.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid unit price: %w", err)
		}

		discountAmt := decimal.Zero
		if itemReq.DiscountAmt != "" {
			discountAmt, err = decimal.NewFromString(itemReq.DiscountAmt)
			if err != nil {
				return nil, fmt.Errorf("invalid discount amount: %w", err)
			}
		}

		itemSubtotal := qty.Mul(unitPrice).Sub(discountAmt)
		subtotal = subtotal.Add(itemSubtotal)
	}

	invoice.Subtotal = subtotal
	invoice.TotalAmount = subtotal.Sub(invoice.DiscountAmount).Add(invoice.TaxAmount)

	// Start transaction
	tx := s.db.Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create invoice
	if err := tx.Set("tenant_id", tenantID).Create(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create invoice items
	for _, itemReq := range req.Items {
		qty, _ := decimal.NewFromString(itemReq.Quantity)
		unitPrice, _ := decimal.NewFromString(itemReq.UnitPrice)

		discountPct := decimal.Zero
		if itemReq.DiscountPct != "" {
			discountPct, _ = decimal.NewFromString(itemReq.DiscountPct)
		}

		discountAmt := decimal.Zero
		if itemReq.DiscountAmt != "" {
			discountAmt, _ = decimal.NewFromString(itemReq.DiscountAmt)
		}

		itemSubtotal := qty.Mul(unitPrice).Sub(discountAmt)

		item := models.InvoiceItem{
			InvoiceID:        invoice.ID,
			SalesOrderItemID: itemReq.SalesOrderItemID,
			DeliveryItemID:   itemReq.DeliveryItemID,
			ProductID:        itemReq.ProductID,
			ProductUnitID:    itemReq.ProductUnitID,
			Quantity:         qty,
			UnitPrice:        unitPrice,
			DiscountPct:      discountPct,
			DiscountAmt:      discountAmt,
			Subtotal:         itemSubtotal,
			Notes:            itemReq.Notes,
		}

		if err := tx.Set("tenant_id", tenantID).Create(&item).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload invoice with relations
	if err := s.db.Set("tenant_id", tenantID).
		Preload("Customer").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		First(&invoice, invoice.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	response := s.toInvoiceResponseWithDetails(invoice)
	return &response, nil
}

// UpdateInvoice updates an existing invoice
func (s *InvoiceService) UpdateInvoice(tenantID, companyID, invoiceID string, req dto.UpdateInvoiceRequest) (*dto.InvoiceResponse, error) {
	var invoice models.Invoice

	if err := s.db.Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", invoiceID, companyID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Update fields
	if req.InvoiceDate != nil {
		invoiceDate, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice date: %w", err)
		}
		invoice.InvoiceDate = invoiceDate
	}

	if req.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due date: %w", err)
		}
		invoice.DueDate = dueDate
	}

	if req.CustomerID != nil {
		invoice.CustomerID = *req.CustomerID
	}

	if req.DiscountAmount != nil {
		discountAmount, err := decimal.NewFromString(*req.DiscountAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid discount amount: %w", err)
		}
		invoice.DiscountAmount = discountAmount
		// Recalculate total
		invoice.TotalAmount = invoice.Subtotal.Sub(invoice.DiscountAmount).Add(invoice.TaxAmount)
	}

	if req.TaxAmount != nil {
		taxAmount, err := decimal.NewFromString(*req.TaxAmount)
		if err != nil {
			return nil, fmt.Errorf("invalid tax amount: %w", err)
		}
		invoice.TaxAmount = taxAmount
		// Recalculate total
		invoice.TotalAmount = invoice.Subtotal.Sub(invoice.DiscountAmount).Add(invoice.TaxAmount)
	}

	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if req.FakturPajakNo != nil {
		invoice.FakturPajakNo = req.FakturPajakNo
	}

	if req.FakturPajakDate != nil {
		fpDate, err := time.Parse("2006-01-02", *req.FakturPajakDate)
		if err != nil {
			return nil, fmt.Errorf("invalid faktur pajak date: %w", err)
		}
		invoice.FakturPajakDate = &fpDate
	}

	// Save changes
	if err := s.db.Set("tenant_id", tenantID).Save(&invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Reload with relations
	if err := s.db.Set("tenant_id", tenantID).
		Preload("Customer").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		Preload("Payments").
		First(&invoice, invoice.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	response := s.toInvoiceResponseWithDetails(invoice)
	return &response, nil
}

// DeleteInvoice soft deletes an invoice
func (s *InvoiceService) DeleteInvoice(tenantID, companyID, invoiceID string) error {
	result := s.db.Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", invoiceID, companyID).
		Delete(&models.Invoice{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete invoice: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// RecordPayment records a payment against an invoice
func (s *InvoiceService) RecordPayment(tenantID, companyID, invoiceID string, req dto.RecordInvoicePaymentRequest) (*dto.InvoiceResponse, error) {
	var invoice models.Invoice

	if err := s.db.Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", invoiceID, companyID).
		First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment date: %w", err)
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("payment amount must be greater than zero")
	}

	remainingAmount := invoice.TotalAmount.Sub(invoice.PaidAmount)
	if amount.GreaterThan(remainingAmount) {
		return nil, fmt.Errorf("payment amount exceeds remaining balance")
	}

	// Generate payment number
	ctx := context.Background()
	paymentNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypeCustomerPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment number: %w", err)
	}

	// Parse payment method
	var paymentMethod models.PaymentMethod
	switch req.PaymentMethod {
	case "CASH":
		paymentMethod = models.PaymentMethodCash
	case "BANK_TRANSFER":
		paymentMethod = models.PaymentMethodBankTransfer
	case "CHECK":
		paymentMethod = models.PaymentMethodCheck
	case "GIRO":
		paymentMethod = models.PaymentMethodGiro
	case "CREDIT_CARD":
		paymentMethod = models.PaymentMethodCreditCard
	case "DEBIT_CARD":
		paymentMethod = models.PaymentMethodDebitCard
	case "E_WALLET":
		paymentMethod = models.PaymentMethodEWallet
	case "OTHER":
		paymentMethod = models.PaymentMethodOther
	default:
		return nil, fmt.Errorf("invalid payment method")
	}

	// Start transaction
	tx := s.db.Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create payment record
	now := time.Now()
	payment := models.Payment{
		TenantID:      invoice.TenantID,
		PaymentNumber: paymentNumber,
		PaymentDate:   paymentDate,
		CustomerID:    invoice.CustomerID,
		InvoiceID:     invoice.ID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Reference:     req.Reference,
		BankAccountID: req.BankAccountID,
		Notes:         req.Notes,
		ReceivedAt:    &now,
	}

	if err := tx.Set("tenant_id", tenantID).Create(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update invoice paid amount and payment status
	newPaidAmount := invoice.PaidAmount.Add(amount)
	newPaymentStatus := models.PaymentStatusPartial

	if newPaidAmount.GreaterThanOrEqual(invoice.TotalAmount) {
		newPaymentStatus = models.PaymentStatusPaid
	}

	if err := tx.Set("tenant_id", tenantID).Model(&invoice).Updates(map[string]interface{}{
		"paid_amount":    newPaidAmount,
		"payment_status": newPaymentStatus,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload invoice with relations
	if err := s.db.Set("tenant_id", tenantID).
		Preload("Customer").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		Preload("Payments").
		First(&invoice, invoice.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	response := s.toInvoiceResponseWithDetails(invoice)
	return &response, nil
}

// Helper methods

func (s *InvoiceService) toInvoiceResponse(invoice models.Invoice) dto.InvoiceResponse {
	response := dto.InvoiceResponse{
		ID:              invoice.ID,
		InvoiceNumber:   invoice.InvoiceNumber,
		InvoiceDate:     invoice.InvoiceDate.Format("2006-01-02"),
		DueDate:         invoice.DueDate.Format("2006-01-02"),
		CustomerID:      invoice.CustomerID,
		Subtotal:        invoice.Subtotal.String(),
		DiscountAmount:  invoice.DiscountAmount.String(),
		TaxAmount:       invoice.TaxAmount.String(),
		TotalAmount:     invoice.TotalAmount.String(),
		PaidAmount:      invoice.PaidAmount.String(),
		RemainingAmount: invoice.TotalAmount.Sub(invoice.PaidAmount).String(),
		PaymentStatus:   string(invoice.PaymentStatus),
		Notes:           invoice.Notes,
		FakturPajakNo:   invoice.FakturPajakNo,
		CreatedAt:       invoice.CreatedAt,
		UpdatedAt:       invoice.UpdatedAt,
	}

	if invoice.Customer.ID != "" {
		response.CustomerName = invoice.Customer.Name
		response.CustomerCode = &invoice.Customer.Code
	}

	if invoice.FakturPajakDate != nil {
		fpDate := invoice.FakturPajakDate.Format("2006-01-02")
		response.FakturPajakDate = &fpDate
	}

	return response
}

func (s *InvoiceService) toInvoiceResponseWithDetails(invoice models.Invoice) dto.InvoiceResponse {
	response := s.toInvoiceResponse(invoice)

	// Add items
	if len(invoice.Items) > 0 {
		items := make([]dto.InvoiceItemResponse, len(invoice.Items))
		for i, item := range invoice.Items {
			unitName := "PCS" // default
			if item.ProductUnit != nil {
				unitName = item.ProductUnit.UnitName
			}

			items[i] = dto.InvoiceItemResponse{
				ID:               item.ID,
				SalesOrderItemID: item.SalesOrderItemID,
				DeliveryItemID:   item.DeliveryItemID,
				ProductID:        item.ProductID,
				ProductCode:      item.Product.Code,
				ProductName:      item.Product.Name,
				ProductUnitID:    item.ProductUnitID,
				UnitName:         unitName,
				Quantity:         item.Quantity.String(),
				UnitPrice:        item.UnitPrice.String(),
				DiscountPct:      item.DiscountPct.String(),
				DiscountAmt:      item.DiscountAmt.String(),
				Subtotal:         item.Subtotal.String(),
				Notes:            item.Notes,
				CreatedAt:        item.CreatedAt,
				UpdatedAt:        item.UpdatedAt,
			}
		}
		response.Items = items
	}

	// Add payments
	if len(invoice.Payments) > 0 {
		payments := make([]dto.InvoicePaymentResponse, len(invoice.Payments))
		for i, payment := range invoice.Payments {
			paymentResp := dto.InvoicePaymentResponse{
				ID:            payment.ID,
				PaymentNumber: payment.PaymentNumber,
				PaymentDate:   payment.PaymentDate.Format("2006-01-02"),
				Amount:        payment.Amount.String(),
				PaymentMethod: string(payment.PaymentMethod),
				Reference:     payment.Reference,
				BankAccountID: payment.BankAccountID,
				Notes:         payment.Notes,
				ReceivedBy:    payment.ReceivedBy,
				CreatedAt:     payment.CreatedAt,
				UpdatedAt:     payment.UpdatedAt,
			}

			if payment.ReceivedAt != nil {
				receivedAt := payment.ReceivedAt.Format(time.RFC3339)
				paymentResp.ReceivedAt = &receivedAt
			}

			payments[i] = paymentResp
		}
		response.Payments = payments
	}

	return response
}

// mapSortField converts frontend camelCase field names to database snake_case column names
func mapSortField(field string) string {
	fieldMap := map[string]string{
		"invoiceNumber": "invoice_number",
		"invoiceDate":   "invoice_date",
		"dueDate":       "due_date",
		"totalAmount":   "total_amount",
		"customerName":  "customer_name",
		"createdAt":     "created_at",
	}

	if dbField, ok := fieldMap[field]; ok {
		return dbField
	}

	// Default to invoice_date if field not found
	return "invoice_date"
}
