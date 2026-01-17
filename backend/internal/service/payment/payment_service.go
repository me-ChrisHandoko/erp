package payment

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

// PaymentService handles customer payment business logic
type PaymentService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
}

// NewPaymentService creates a new payment service
func NewPaymentService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *PaymentService {
	return &PaymentService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ListPayments retrieves customer payments with filters and pagination
func (s *PaymentService) ListPayments(companyID, tenantID string, filters dto.PaymentFilters) (*dto.PaymentListResponse, error) {
	var payments []models.Payment
	var total int64

	// Build query with tenant context
	// Add explicit tenant_id filter with table prefix to prevent ambiguous column error
	// GORM's hasExistingTenantFilter() will detect this and skip auto-injection
	query := s.db.Set("tenant_id", tenantID).Model(&models.Payment{}).
		Joins("JOIN invoices ON payments.invoice_id = invoices.id").
		Joins("JOIN customers ON payments.customer_id = customers.id").
		Where("payments.tenant_id = ? AND invoices.company_id = ?", tenantID, companyID)

	// Apply filters
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			"payments.payment_number LIKE ? OR "+
				"EXISTS (SELECT 1 FROM customers c WHERE c.id = payments.customer_id AND (c.name LIKE ? OR c.code LIKE ?)) OR "+
				"EXISTS (SELECT 1 FROM invoices i WHERE i.id = payments.invoice_id AND i.invoice_number LIKE ?)",
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	if filters.CustomerID != "" {
		query = query.Where("payments.customer_id = ?", filters.CustomerID)
	}

	if filters.InvoiceID != "" {
		query = query.Where("payments.invoice_id = ?", filters.InvoiceID)
	}

	if filters.PaymentMethod != "" {
		query = query.Where("payments.payment_method = ?", filters.PaymentMethod)
	}

	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err == nil {
			query = query.Where("payments.payment_date >= ?", dateFrom)
		}
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err == nil {
			query = query.Where("payments.payment_date <= ?", dateTo)
		}
	}

	if filters.AmountMin != "" {
		amountMin, err := decimal.NewFromString(filters.AmountMin)
		if err == nil {
			query = query.Where("payments.amount >= ?", amountMin)
		}
	}

	if filters.AmountMax != "" {
		amountMax, err := decimal.NewFromString(filters.AmountMax)
		if err == nil {
			query = query.Where("payments.amount <= ?", amountMax)
		}
	}

	// Filter by check status if provided
	if filters.CheckStatus != "" {
		query = query.Where(
			"EXISTS (SELECT 1 FROM payment_checks pc WHERE pc.payment_id = payments.id AND pc.status = ?)",
			filters.CheckStatus)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	// Apply sorting
	sortBy := "payments.payment_date"
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

	// Fetch payments with relations
	if err := query.
		Preload("Customer").
		Preload("Invoice").
		Preload("BankAccount").
		Preload("Checks").
		Offset(offset).
		Limit(limit).
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payments: %w", err)
	}

	// Convert to response DTOs
	paymentResponses := make([]dto.PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = s.toPaymentResponse(payment)
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.PaymentListResponse{
		Data: paymentResponses,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetPayment retrieves a single payment by ID
func (s *PaymentService) GetPayment(companyID, tenantID, paymentID string) (*dto.PaymentResponse, error) {
	var payment models.Payment

	// Set tenant context for GORM callbacks
	if err := s.db.Set("tenant_id", tenantID).
		Preload("Customer").
		Preload("Invoice").
		Preload("BankAccount").
		Preload("Checks").
		Joins("JOIN invoices ON payments.invoice_id = invoices.id").
		Where("payments.id = ? AND payments.tenant_id = ? AND invoices.company_id = ?", paymentID, tenantID, companyID).
		First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	response := s.toPaymentResponse(payment)
	return &response, nil
}

// CreatePayment creates a new customer payment
func (s *PaymentService) CreatePayment(ctx context.Context, companyID, tenantID, userID, ipAddress, userAgent string, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, errors.New("invalid payment date format")
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, errors.New("invalid amount format")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be greater than zero")
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Verify customer exists and belongs to company
	var customer models.Customer
	if err := tx.Where("id = ? AND company_id = ?", req.CustomerID, companyID).
		First(&customer).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("customer not found")
		}
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}

	// Verify invoice exists and belongs to same customer and company
	var invoice models.Invoice
	if err := tx.Where("id = ? AND customer_id = ? AND company_id = ?", req.InvoiceID, req.CustomerID, companyID).
		First(&invoice).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found or does not belong to this customer")
		}
		return nil, fmt.Errorf("failed to verify invoice: %w", err)
	}

	// Verify payment doesn't exceed remaining balance
	remaining := invoice.TotalAmount.Sub(invoice.PaidAmount)
	if amount.GreaterThan(remaining) {
		tx.Rollback()
		return nil, fmt.Errorf("payment amount (%s) exceeds remaining invoice balance (%s)", amount.String(), remaining.String())
	}

	// Generate payment number
	paymentNumber, err := s.docNumberGen.GeneratePaymentNumber(ctx, tx, companyID, paymentDate)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate payment number: %w", err)
	}

	// Create payment record
	payment := models.Payment{
		TenantID:      tenantID,
		PaymentNumber: paymentNumber,
		PaymentDate:   paymentDate,
		CustomerID:    req.CustomerID,
		InvoiceID:     req.InvoiceID,
		Amount:        amount,
		PaymentMethod: models.PaymentMethod(req.PaymentMethod),
		Reference:     req.Reference,
		BankAccountID: req.BankAccountID,
		Notes:         req.Notes,
		ReceivedBy:    &userID,
		ReceivedAt:    &paymentDate,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Create check record if payment method is CHECK or GIRO
	if (req.PaymentMethod == "CHECK" || req.PaymentMethod == "GIRO") && req.CheckNumber != nil && req.CheckDate != nil {
		checkDate, err := time.Parse("2006-01-02", *req.CheckDate)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("invalid check date format")
		}

		check := models.PaymentCheck{
			PaymentID:   payment.ID,
			CheckNumber: *req.CheckNumber,
			CheckDate:   paymentDate,
			DueDate:     checkDate,
			Amount:      amount,
			BankName:    "N/A", // TODO: Get from bank account if available
			Status:      models.CheckStatusIssued,
		}

		if err := tx.Create(&check).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create check record: %w", err)
		}
	}

	// Update invoice paid amount and status
	newPaidAmount := invoice.PaidAmount.Add(amount)
	newPaymentStatus := models.PaymentStatusPartial

	if newPaidAmount.GreaterThanOrEqual(invoice.TotalAmount) {
		newPaymentStatus = models.PaymentStatusPaid
	}

	if err := tx.Model(&invoice).Updates(map[string]interface{}{
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

	// Fetch created payment with relations
	return s.GetPayment(companyID, tenantID, payment.ID)
}

// UpdatePayment updates an existing customer payment
func (s *PaymentService) UpdatePayment(ctx context.Context, companyID, tenantID, paymentID, userID, ipAddress, userAgent string, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch existing payment
	var payment models.Payment
	if err := tx.Preload("Invoice").
		Joins("JOIN invoices ON payments.invoice_id = invoices.id").
		Where("payments.id = ? AND invoices.company_id = ?", paymentID, companyID).
		First(&payment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	// Store old amount for invoice update
	oldAmount := payment.Amount

	// Prepare updates
	updates := make(map[string]interface{})

	if req.PaymentDate != nil {
		paymentDate, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("invalid payment date format")
		}
		updates["payment_date"] = paymentDate
	}

	if req.Amount != nil {
		newAmount, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("invalid amount format")
		}
		if newAmount.LessThanOrEqual(decimal.Zero) {
			tx.Rollback()
			return nil, errors.New("amount must be greater than zero")
		}
		updates["amount"] = newAmount
	}

	if req.PaymentMethod != nil {
		updates["payment_method"] = *req.PaymentMethod
	}

	if req.Reference != nil {
		updates["reference"] = *req.Reference
	}

	if req.BankAccountID != nil {
		updates["bank_account_id"] = *req.BankAccountID
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	// Update payment
	if len(updates) > 0 {
		if err := tx.Model(&payment).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update payment: %w", err)
		}
	}

	// If amount changed, update invoice
	if req.Amount != nil {
		newAmount, _ := decimal.NewFromString(*req.Amount)
		amountDiff := newAmount.Sub(oldAmount)

		// Update invoice paid amount
		newPaidAmount := payment.Invoice.PaidAmount.Add(amountDiff)

		// Verify new paid amount doesn't exceed total
		if newPaidAmount.GreaterThan(payment.Invoice.TotalAmount) {
			tx.Rollback()
			return nil, errors.New("new payment amount would exceed invoice total")
		}

		newPaymentStatus := models.PaymentStatusPartial
		if newPaidAmount.GreaterThanOrEqual(payment.Invoice.TotalAmount) {
			newPaymentStatus = models.PaymentStatusPaid
		} else if newPaidAmount.LessThanOrEqual(decimal.Zero) {
			newPaymentStatus = models.PaymentStatusUnpaid
		}

		if err := tx.Model(&payment.Invoice).Updates(map[string]interface{}{
			"paid_amount":    newPaidAmount,
			"payment_status": newPaymentStatus,
		}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update invoice: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch updated payment
	return s.GetPayment(companyID, tenantID, paymentID)
}

// VoidPayment voids a payment (same-day delete only)
func (s *PaymentService) VoidPayment(ctx context.Context, companyID, tenantID, paymentID, userID, ipAddress, userAgent string, req *dto.VoidPaymentRequest) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch payment
	var payment models.Payment
	if err := tx.Preload("Invoice").
		Joins("JOIN invoices ON payments.invoice_id = invoices.id").
		Where("payments.id = ? AND invoices.company_id = ?", paymentID, companyID).
		First(&payment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("payment not found")
		}
		return fmt.Errorf("failed to fetch payment: %w", err)
	}

	// Check if payment is from today (same-day void only)
	today := time.Now().Truncate(24 * time.Hour)
	paymentDay := payment.PaymentDate.Truncate(24 * time.Hour)
	if !paymentDay.Equal(today) {
		tx.Rollback()
		return errors.New("can only void payments from today")
	}

	// Update invoice paid amount and status
	newPaidAmount := payment.Invoice.PaidAmount.Sub(payment.Amount)

	newPaymentStatus := models.PaymentStatusPartial
	if newPaidAmount.LessThanOrEqual(decimal.Zero) {
		newPaymentStatus = models.PaymentStatusUnpaid
	}

	if err := tx.Model(&payment.Invoice).Updates(map[string]interface{}{
		"paid_amount":    newPaidAmount,
		"payment_status": newPaymentStatus,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// Delete payment (cascade will delete checks)
	if err := tx.Delete(&payment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateCheckStatus updates check/giro status
func (s *PaymentService) UpdateCheckStatus(ctx context.Context, companyID, tenantID, paymentID, userID, ipAddress, userAgent string, req *dto.UpdateCheckStatusRequest) (*dto.PaymentResponse, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch payment
	var payment models.Payment
	if err := tx.Preload("Checks").
		Joins("JOIN invoices ON payments.invoice_id = invoices.id").
		Where("payments.id = ? AND invoices.company_id = ?", paymentID, companyID).
		First(&payment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	// Verify payment has check records
	if len(payment.Checks) == 0 {
		tx.Rollback()
		return nil, errors.New("payment does not have check records")
	}

	// Update check status
	now := time.Now()
	updates := map[string]interface{}{
		"status": req.CheckStatus,
	}

	if req.CheckStatus == dto.CheckStatusCleared {
		updates["cleared_date"] = now
	} else if req.CheckStatus == dto.CheckStatusBounced {
		updates["bounced_date"] = now
		if req.Notes != nil {
			updates["bounced_note"] = *req.Notes
		}
	}

	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}

	// Update all checks for this payment
	if err := tx.Model(&models.PaymentCheck{}).
		Where("payment_id = ?", paymentID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update check status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch updated payment
	return s.GetPayment(companyID, tenantID, paymentID)
}

// Helper functions

func mapSortField(sortBy string) string {
	mapping := map[string]string{
		"paymentNumber": "payments.payment_number",
		"paymentDate":   "payments.payment_date",
		"customerName":  "customers.name",
		"amount":        "payments.amount",
		"createdAt":     "payments.created_at",
	}
	if field, ok := mapping[sortBy]; ok {
		return field
	}
	return "payments.payment_date"
}

func (s *PaymentService) toPaymentResponse(payment models.Payment) dto.PaymentResponse {
	response := dto.PaymentResponse{
		ID:            payment.ID,
		PaymentNumber: payment.PaymentNumber,
		PaymentDate:   payment.PaymentDate.Format("2006-01-02"),
		CustomerID:    payment.CustomerID,
		InvoiceID:     payment.InvoiceID,
		Amount:        payment.Amount.String(),
		PaymentMethod: string(payment.PaymentMethod),
		Reference:     payment.Reference,
		BankAccountID: payment.BankAccountID,
		Notes:         payment.Notes,
		CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
	}

	// Add customer info
	if payment.Customer.ID != "" {
		response.CustomerName = payment.Customer.Name
		customerCode := payment.Customer.Code
		response.CustomerCode = &customerCode
	}

	// Add invoice info
	if payment.Invoice.ID != "" {
		response.InvoiceNumber = payment.Invoice.InvoiceNumber
	}

	// Add bank account info
	if payment.BankAccount != nil && payment.BankAccount.ID != "" {
		bankName := fmt.Sprintf("%s - %s", payment.BankAccount.BankName, payment.BankAccount.AccountNumber)
		response.BankAccountName = &bankName
	}

	// Add check info if exists
	if len(payment.Checks) > 0 {
		check := payment.Checks[0]
		response.CheckNumber = &check.CheckNumber
		checkDate := check.DueDate.Format("2006-01-02")
		response.CheckDate = &checkDate
		checkStatus := string(check.Status)
		response.CheckStatus = &checkStatus
	}

	// Add audit info
	if payment.ReceivedBy != nil {
		response.CreatedBy = *payment.ReceivedBy
	} else {
		response.CreatedBy = "System"
	}

	return response
}
