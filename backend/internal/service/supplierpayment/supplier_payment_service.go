package supplierpayment

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

// SupplierPaymentService handles business logic for supplier payments
type SupplierPaymentService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
}

// NewSupplierPaymentService creates a new supplier payment service
func NewSupplierPaymentService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *SupplierPaymentService {
	return &SupplierPaymentService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ============================================================================
// LIST & GET Operations
// ============================================================================

// ListSupplierPayments retrieves supplier payments with filters and pagination
func (s *SupplierPaymentService) ListSupplierPayments(
	ctx context.Context,
	tenantID, companyID string,
	filters dto.SupplierPaymentFilters,
) ([]models.SupplierPayment, dto.PaginationResponse, error) {
	var payments []models.SupplierPayment
	var total int64

	// Set default pagination
	filters.SetDefaultPagination()

	// Base query with tenant context set for GORM callbacks
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.SupplierPayment{}).
		Where("company_id = ?", companyID).
		Preload("Supplier").
		Preload("PurchaseOrder").
		Preload("BankAccount")

	// Apply filters
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			s.db.Where("payment_number ILIKE ?", searchPattern).
				Or("reference ILIKE ?", searchPattern).
				Or("EXISTS (SELECT 1 FROM suppliers WHERE suppliers.id = supplier_payments.supplier_id AND (suppliers.name ILIKE ? OR suppliers.code ILIKE ?))", searchPattern, searchPattern),
		)
	}

	if filters.SupplierID != "" {
		query = query.Where("supplier_id = ?", filters.SupplierID)
	}

	if filters.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filters.PaymentMethod)
	}

	// Date range filters
	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err == nil {
			query = query.Where("payment_date >= ?", dateFrom)
		}
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err == nil {
			query = query.Where("payment_date <= ?", dateTo)
		}
	}

	// Amount range filters
	if filters.AmountMin != "" {
		amountMin, err := decimal.NewFromString(filters.AmountMin)
		if err == nil {
			query = query.Where("amount >= ?", amountMin)
		}
	}

	if filters.AmountMax != "" {
		amountMax, err := decimal.NewFromString(filters.AmountMax)
		if err == nil {
			query = query.Where("amount <= ?", amountMax)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Apply sorting
	// For supplier_name, we need to join with suppliers table
	if filters.SortBy == "supplier_name" {
		query = query.Joins("LEFT JOIN suppliers ON suppliers.id = supplier_payments.supplier_id").
			Order("suppliers.name " + filters.SortOrder)
	} else {
		orderClause := filters.SortBy + " " + filters.SortOrder
		query = query.Order(orderClause)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Execute query
	if err := query.Find(&payments).Error; err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Calculate pagination
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))
	pagination := dto.PaginationResponse{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return payments, pagination, nil
}

// GetSupplierPayment retrieves a specific supplier payment by ID
func (s *SupplierPaymentService) GetSupplierPayment(
	ctx context.Context,
	tenantID, companyID, paymentID string,
) (*models.SupplierPayment, error) {
	var payment models.SupplierPayment

	// Query with tenant context
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", paymentID, companyID).
		Preload("Supplier").
		Preload("PurchaseOrder").
		Preload("BankAccount").
		First(&payment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("supplier payment not found")
		}
		return nil, err
	}

	return &payment, nil
}

// ============================================================================
// CREATE Operation
// ============================================================================

// CreateSupplierPayment creates a new supplier payment
func (s *SupplierPaymentService) CreateSupplierPayment(
	ctx context.Context,
	tenantID, companyID, userID string,
	req dto.CreateSupplierPaymentRequest,
) (*models.SupplierPayment, error) {
	// Validate supplier exists
	var supplier models.Supplier
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", req.SupplierID, companyID).
		First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("supplier not found")
		}
		return nil, err
	}

	// Validate purchase order if provided
	if req.PurchaseOrderID != nil {
		var po models.PurchaseOrder
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("id = ? AND company_id = ? AND supplier_id = ?", *req.PurchaseOrderID, companyID, req.SupplierID).
			First(&po).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("purchase order not found or does not belong to the specified supplier")
			}
			return nil, err
		}
	}

	// Validate bank account if provided
	if req.BankAccountID != nil {
		var bankAccount models.CompanyBank
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("id = ? AND company_id = ?", *req.BankAccountID, companyID).
			First(&bankAccount).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("bank account not found")
			}
			return nil, err
		}
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid payment date format: %w", err)
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be greater than zero")
	}

	// Generate payment number
	paymentNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypeSupplierPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment number: %w", err)
	}

	// Create payment
	payment := models.SupplierPayment{
		TenantID:        tenantID,
		CompanyID:       companyID,
		PaymentNumber:   paymentNumber,
		PaymentDate:     paymentDate,
		SupplierID:      req.SupplierID,
		PurchaseOrderID: req.PurchaseOrderID,
		Amount:          amount,
		PaymentMethod:   models.PaymentMethod(req.PaymentMethod),
		Reference:       req.Reference,
		BankAccountID:   req.BankAccountID,
		Notes:           req.Notes,
	}

	// Save payment in transaction
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		// Preload relations for response
		if err := tx.Preload("Supplier").Preload("PurchaseOrder").Preload("BankAccount").
			First(&payment, "id = ?", payment.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &payment, nil
}

// ============================================================================
// UPDATE Operation
// ============================================================================

// UpdateSupplierPayment updates an existing supplier payment
func (s *SupplierPaymentService) UpdateSupplierPayment(
	ctx context.Context,
	tenantID, companyID, userID, paymentID string,
	req dto.UpdateSupplierPaymentRequest,
) (*models.SupplierPayment, error) {
	// Get existing payment
	payment, err := s.GetSupplierPayment(ctx, tenantID, companyID, paymentID)
	if err != nil {
		return nil, err
	}

	// Check if payment is approved (cannot edit approved payments)
	if payment.ApprovedBy != nil {
		return nil, errors.New("cannot update approved payment")
	}

	// Update fields if provided
	if req.PaymentDate != nil {
		paymentDate, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			return nil, fmt.Errorf("invalid payment date format: %w", err)
		}
		payment.PaymentDate = paymentDate
	}

	if req.SupplierID != nil {
		// Validate supplier exists
		var supplier models.Supplier
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("id = ? AND company_id = ?", *req.SupplierID, companyID).
			First(&supplier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("supplier not found")
			}
			return nil, err
		}
		payment.SupplierID = *req.SupplierID
	}

	if req.Amount != nil {
		amount, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			return nil, fmt.Errorf("invalid amount format: %w", err)
		}
		if amount.LessThanOrEqual(decimal.Zero) {
			return nil, errors.New("amount must be greater than zero")
		}
		payment.Amount = amount
	}

	if req.PaymentMethod != nil {
		payment.PaymentMethod = models.PaymentMethod(*req.PaymentMethod)
	}

	if req.Reference != nil {
		payment.Reference = req.Reference
	}

	if req.BankAccountID != nil {
		// Validate bank account if provided
		if *req.BankAccountID != "" {
			var bankAccount models.CompanyBank
			if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
				Where("id = ? AND company_id = ?", *req.BankAccountID, companyID).
				First(&bankAccount).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errors.New("bank account not found")
				}
				return nil, err
			}
		}
		payment.BankAccountID = req.BankAccountID
	}

	if req.Notes != nil {
		payment.Notes = req.Notes
	}

	if req.PurchaseOrderID != nil {
		payment.PurchaseOrderID = req.PurchaseOrderID
	}

	// Save updates
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(&payment).Error; err != nil {
		return nil, err
	}

	// Reload with relations
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Supplier").Preload("PurchaseOrder").Preload("BankAccount").
		First(&payment, "id = ?", payment.ID).Error; err != nil {
		return nil, err
	}

	return payment, nil
}

// ============================================================================
// DELETE Operation
// ============================================================================

// DeleteSupplierPayment soft deletes a supplier payment
func (s *SupplierPaymentService) DeleteSupplierPayment(
	ctx context.Context,
	tenantID, companyID, userID, paymentID string,
) error {
	// Get payment
	payment, err := s.GetSupplierPayment(ctx, tenantID, companyID, paymentID)
	if err != nil {
		return err
	}

	// Check if payment is approved (cannot delete approved payments)
	if payment.ApprovedBy != nil {
		return errors.New("cannot delete approved payment")
	}

	// Soft delete
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Delete(&payment).Error; err != nil {
		return err
	}

	return nil
}

// ============================================================================
// APPROVAL Operations
// ============================================================================

// ApproveSupplierPayment approves a supplier payment
func (s *SupplierPaymentService) ApproveSupplierPayment(
	ctx context.Context,
	tenantID, companyID, userID, paymentID string,
	req dto.ApproveSupplierPaymentRequest,
) (*models.SupplierPayment, error) {
	// Get payment
	payment, err := s.GetSupplierPayment(ctx, tenantID, companyID, paymentID)
	if err != nil {
		return nil, err
	}

	// Check if already approved
	if payment.ApprovedBy != nil {
		return nil, errors.New("payment is already approved")
	}

	// Set approval info
	now := time.Now()
	payment.ApprovedBy = &userID
	payment.ApprovedAt = &now

	// Save updates
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(&payment).Error; err != nil {
		return nil, err
	}

	// Reload with relations
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Supplier").Preload("PurchaseOrder").Preload("BankAccount").
		First(&payment, "id = ?", payment.ID).Error; err != nil {
		return nil, err
	}

	return payment, nil
}
