package purchaseinvoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/document"
	"backend/models"
)

// PurchaseInvoiceService handles business logic for purchase invoices
type PurchaseInvoiceService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
}

// NewPurchaseInvoiceService creates a new purchase invoice service
func NewPurchaseInvoiceService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *PurchaseInvoiceService {
	return &PurchaseInvoiceService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ============================================================================
// LIST & GET Operations
// ============================================================================

// ListPurchaseInvoices retrieves purchase invoices with filters and pagination
func (s *PurchaseInvoiceService) ListPurchaseInvoices(
	ctx context.Context,
	tenantID, companyID string,
	filters dto.PurchaseInvoiceFilters,
) ([]models.PurchaseInvoice, dto.PaginationResponse, error) {
	var invoices []models.PurchaseInvoice
	var total int64

	// Base query with tenant context set for GORM callbacks
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.PurchaseInvoice{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			"invoice_number ILIKE ? OR supplier_name ILIKE ?",
			searchPattern, searchPattern,
		)
	}

	if filters.SupplierID != "" {
		query = query.Where("supplier_id = ?", filters.SupplierID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.PaymentStatus != "" {
		query = query.Where("payment_status = ?", filters.PaymentStatus)
	}

	// Date range filters
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

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "invoiceNumber"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}

	// Map camelCase to snake_case for database
	sortField := toSnakeCase(sortBy)
	query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&invoices).Error; err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	pagination := dto.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return invoices, pagination, nil
}

// GetPurchaseInvoice retrieves a single purchase invoice by ID with relations
func (s *PurchaseInvoiceService) GetPurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID string,
) (*models.PurchaseInvoice, error) {
	var invoice models.PurchaseInvoice

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Items").
		Preload("Payments").
		Preload("Supplier").
		Where("id = ? AND company_id = ?", invoiceID, companyID).
		First(&invoice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("purchase invoice not found")
		}
		return nil, err
	}

	return &invoice, nil
}

// ============================================================================
// CREATE Operation
// ============================================================================

// CreatePurchaseInvoice creates a new purchase invoice
func (s *PurchaseInvoiceService) CreatePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, userID string,
	req dto.CreatePurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Parse dates
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, errors.New("invalid invoice date format")
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, errors.New("invalid due date format")
	}

	// Validate due date is after invoice date
	if dueDate.Before(invoiceDate) {
		return nil, errors.New("due date must be after invoice date")
	}

	// Generate invoice number using document number generator
	invoiceNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypePurchaseInvoice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Get supplier information for denormalization
	var supplier models.Supplier
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).First(&supplier, "id = ?", req.SupplierID).Error; err != nil {
		return nil, errors.New("supplier not found")
	}

	// Parse tax rate
	taxRate := decimal.NewFromFloat(11.0) // Default 11%
	if req.TaxRate != "" {
		taxRate, err = decimal.NewFromString(req.TaxRate)
		if err != nil {
			return nil, errors.New("invalid tax rate format")
		}
	}

	// Parse discount
	discountAmount := decimal.Zero
	if req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
	}

	// Set payment term days
	paymentTermDays := 30 // Default NET 30
	if req.PaymentTermDays > 0 {
		paymentTermDays = req.PaymentTermDays
	}

	// Parse non-goods costs (biaya tambahan)
	shippingCost := decimal.Zero
	if req.ShippingCost != "" {
		shippingCost, err = decimal.NewFromString(req.ShippingCost)
		if err != nil {
			return nil, errors.New("invalid shipping cost format")
		}
	}

	handlingCost := decimal.Zero
	if req.HandlingCost != "" {
		handlingCost, err = decimal.NewFromString(req.HandlingCost)
		if err != nil {
			return nil, errors.New("invalid handling cost format")
		}
	}

	otherCost := decimal.Zero
	if req.OtherCost != "" {
		otherCost, err = decimal.NewFromString(req.OtherCost)
		if err != nil {
			return nil, errors.New("invalid other cost format")
		}
	}

	var invoice *models.PurchaseInvoice

	// Use transaction for atomic create
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Create purchase invoice
		invoice = &models.PurchaseInvoice{
			TenantID:             tenantID,
			CompanyID:            companyID,
			InvoiceNumber:        invoiceNumber, // Use auto-generated number
			InvoiceDate:          invoiceDate,
			DueDate:              dueDate,
			SupplierID:           req.SupplierID,
			SupplierName:         supplier.Name,
			SupplierCode:         &supplier.Code,
			PurchaseOrderID:      req.PurchaseOrderID,
			GoodsReceiptID:       req.GoodsReceiptID,
			DiscountAmount:       discountAmount,
			TaxRate:              taxRate,
			PaymentTermDays:      paymentTermDays,
			Notes:                req.Notes,
			ShippingCost:         shippingCost,
			HandlingCost:         handlingCost,
			OtherCost:            otherCost,
			OtherCostDescription: req.OtherCostDescription,
			Status:               models.PurchaseInvoiceStatusDraft,
			PaymentStatus:        models.PaymentStatusUnpaid,
			CreatedBy:            userID,
		}

		if err := tx.Create(invoice).Error; err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		// 2. Create invoice items
		for _, itemReq := range req.Items {
			item, err := s.createInvoiceItem(tx, invoice, itemReq)
			if err != nil {
				return fmt.Errorf("failed to create invoice item: %w", err)
			}
			invoice.Items = append(invoice.Items, *item)
		}

		// 3. Calculate totals
		invoice.CalculateTotals()

		// 4. Update invoice with calculated totals
		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to save invoice totals: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with relations
	s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Items").
		First(invoice, invoice.ID)

	return invoice, nil
}

// createInvoiceItem helper function to create an invoice item
func (s *PurchaseInvoiceService) createInvoiceItem(
	tx *gorm.DB,
	invoice *models.PurchaseInvoice,
	req dto.CreatePurchaseInvoiceItemRequest,
) (*models.PurchaseInvoiceItem, error) {
	// Get product information for denormalization
	var product models.Product
	if err := tx.Preload("Units").First(&product, "id = ?", req.ProductID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Get unit information
	var unit models.ProductUnit
	if err := tx.First(&unit, "id = ?", req.UnitID).Error; err != nil {
		return nil, errors.New("product unit not found")
	}

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, errors.New("invalid quantity format")
	}
	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("quantity must be greater than 0")
	}

	// Parse unit price
	unitPrice, err := decimal.NewFromString(req.UnitPrice)
	if err != nil {
		return nil, errors.New("invalid unit price format")
	}
	if unitPrice.LessThan(decimal.Zero) {
		return nil, errors.New("unit price cannot be negative")
	}

	// Parse optional fields
	discountAmount := decimal.Zero
	if req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
	}

	discountPct := decimal.Zero
	if req.DiscountPct != "" {
		discountPct, err = decimal.NewFromString(req.DiscountPct)
		if err != nil {
			return nil, errors.New("invalid discount percentage format")
		}
	}

	taxAmount := decimal.Zero
	if req.TaxAmount != "" {
		taxAmount, err = decimal.NewFromString(req.TaxAmount)
		if err != nil {
			return nil, errors.New("invalid tax amount format")
		}
	}

	// Create invoice item
	item := models.PurchaseInvoiceItem{
		PurchaseInvoiceID:   invoice.ID,
		PurchaseOrderItemID: req.PurchaseOrderItemID,
		GoodsReceiptItemID:  req.GoodsReceiptItemID,
		ProductID:           req.ProductID,
		ProductCode:         product.Code,
		ProductName:         product.Name,
		UnitID:              req.UnitID,
		UnitName:            unit.UnitName,
		Quantity:            quantity,
		UnitPrice:           unitPrice,
		DiscountAmount:      discountAmount,
		DiscountPct:         discountPct,
		TaxAmount:           taxAmount,
		Notes:               req.Notes,
	}

	// Calculate line total
	item.CalculateLineTotal()

	if err := tx.Create(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

// ============================================================================
// UPDATE Operation
// ============================================================================

// UpdatePurchaseInvoice updates an existing purchase invoice
func (s *PurchaseInvoiceService) UpdatePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, userID string,
	req dto.UpdatePurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Only DRAFT invoices can be fully updated
	if invoice.Status != models.PurchaseInvoiceStatusDraft {
		return nil, errors.New("only draft invoices can be updated")
	}

	// Update fields if provided
	if req.InvoiceDate != nil {
		invoiceDate, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return nil, errors.New("invalid invoice date format")
		}
		invoice.InvoiceDate = invoiceDate
	}

	if req.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due date format")
		}
		invoice.DueDate = dueDate
	}

	if req.SupplierID != nil {
		var supplier models.Supplier
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).First(&supplier, "id = ?", *req.SupplierID).Error; err != nil {
			return nil, errors.New("supplier not found")
		}
		invoice.SupplierID = *req.SupplierID
		invoice.SupplierName = supplier.Name
		invoice.SupplierCode = &supplier.Code
	}

	if req.DiscountAmount != nil {
		discountAmount, err := decimal.NewFromString(*req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
		invoice.DiscountAmount = discountAmount
	}

	if req.TaxRate != nil {
		taxRate, err := decimal.NewFromString(*req.TaxRate)
		if err != nil {
			return nil, errors.New("invalid tax rate format")
		}
		invoice.TaxRate = taxRate
	}

	if req.PaymentTermDays != nil {
		invoice.PaymentTermDays = *req.PaymentTermDays
	}

	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if req.Status != nil {
		invoice.Status = models.PurchaseInvoiceStatus(*req.Status)
	}

	// Update non-goods costs (biaya tambahan)
	if req.ShippingCost != nil {
		shippingCost, err := decimal.NewFromString(*req.ShippingCost)
		if err != nil {
			return nil, errors.New("invalid shipping cost format")
		}
		invoice.ShippingCost = shippingCost
	}

	if req.HandlingCost != nil {
		handlingCost, err := decimal.NewFromString(*req.HandlingCost)
		if err != nil {
			return nil, errors.New("invalid handling cost format")
		}
		invoice.HandlingCost = handlingCost
	}

	if req.OtherCost != nil {
		otherCost, err := decimal.NewFromString(*req.OtherCost)
		if err != nil {
			return nil, errors.New("invalid other cost format")
		}
		invoice.OtherCost = otherCost
	}

	if req.OtherCostDescription != nil {
		invoice.OtherCostDescription = req.OtherCostDescription
	}

	invoice.UpdatedBy = &userID

	// Recalculate totals
	invoice.CalculateTotals()

	// Save changes
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(invoice).Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

// ============================================================================
// DELETE Operation
// ============================================================================

// DeletePurchaseInvoice soft deletes a purchase invoice
func (s *PurchaseInvoiceService) DeletePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID string,
) error {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return err
	}

	// Only DRAFT invoices can be deleted
	if invoice.Status != models.PurchaseInvoiceStatusDraft {
		return errors.New("only draft invoices can be deleted")
	}

	// Check if invoice has payments
	if invoice.PaidAmount.GreaterThan(decimal.Zero) {
		return errors.New("cannot delete invoice with payments")
	}

	// Soft delete
	return s.db.WithContext(ctx).Set("tenant_id", tenantID).Delete(invoice).Error
}

// ============================================================================
// WORKFLOW Operations
// ============================================================================

// ApprovePurchaseInvoice approves a purchase invoice
func (s *PurchaseInvoiceService) ApprovePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, approverID string,
	req dto.ApprovePurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Only SUBMITTED invoices can be approved
	if invoice.Status != models.PurchaseInvoiceStatusSubmitted {
		return nil, errors.New("only submitted invoices can be approved")
	}

	// Update status
	now := time.Now()
	invoice.Status = models.PurchaseInvoiceStatusApproved
	invoice.ApprovedBy = &approverID
	invoice.ApprovedAt = &now
	invoice.UpdatedBy = &approverID

	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(invoice).Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

// RejectPurchaseInvoice rejects a purchase invoice
func (s *PurchaseInvoiceService) RejectPurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, rejecterID string,
	req dto.RejectPurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Only SUBMITTED invoices can be rejected
	if invoice.Status != models.PurchaseInvoiceStatusSubmitted {
		return nil, errors.New("only submitted invoices can be rejected")
	}

	// Update status
	now := time.Now()
	invoice.Status = models.PurchaseInvoiceStatusRejected
	invoice.RejectedBy = &rejecterID
	invoice.RejectedAt = &now
	invoice.RejectedReason = &req.Reason
	invoice.UpdatedBy = &rejecterID

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(invoice).Error; err != nil {
		return nil, err
	}

	return invoice, nil
}

// RecordPayment records a payment against a purchase invoice
func (s *PurchaseInvoiceService) RecordPayment(
	ctx context.Context,
	tenantID, companyID, invoiceID, userID string,
	req dto.RecordPaymentRequest,
) (*models.PurchaseInvoicePayment, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Only APPROVED invoices can receive payments
	if invoice.Status != models.PurchaseInvoiceStatusApproved &&
		invoice.Status != models.PurchaseInvoiceStatusPaid {
		return nil, errors.New("only approved invoices can receive payments")
	}

	// Parse payment amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, errors.New("invalid payment amount format")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("payment amount must be greater than 0")
	}

	// Check if payment exceeds remaining amount
	if amount.GreaterThan(invoice.RemainingAmount) {
		return nil, errors.New("payment amount exceeds remaining amount")
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, errors.New("invalid payment date format")
	}

	// Transaction with tenant context
	var payment *models.PurchaseInvoicePayment
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Create payment record
		payment = &models.PurchaseInvoicePayment{
			TenantID:          tenantID,
			CompanyID:         companyID,
			PurchaseInvoiceID: invoiceID,
			PaymentNumber:     req.PaymentNumber,
			PaymentDate:       paymentDate,
			Amount:            amount,
			PaymentMethod:     models.PaymentMethod(req.PaymentMethod),
			Reference:         req.Reference,
			BankAccountID:     req.BankAccountID,
			Notes:             req.Notes,
			CreatedBy:         userID,
		}

		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		// Update invoice paid amount
		invoice.PaidAmount = invoice.PaidAmount.Add(amount)
		invoice.RemainingAmount = invoice.TotalAmount.Sub(invoice.PaidAmount)
		invoice.UpdatePaymentStatus()
		invoice.UpdatedBy = &userID

		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to update invoice: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	switch s {
	case "invoiceNumber":
		return "invoice_number"
	case "invoiceDate":
		return "invoice_date"
	case "dueDate":
		return "due_date"
	case "totalAmount":
		return "total_amount"
	case "createdAt":
		return "created_at"
	default:
		return s
	}
}
