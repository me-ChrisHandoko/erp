package supplier

import (
	"context"
	"fmt"
	"math"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// SupplierService - Business logic for supplier management
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 3
type SupplierService struct {
	db *gorm.DB
}

// NewSupplierService creates a new supplier service instance
func NewSupplierService(db *gorm.DB) *SupplierService {
	return &SupplierService{db: db}
}

// ============================================================================
// CREATE SUPPLIER
// ============================================================================

// CreateSupplier creates a new supplier
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *SupplierService) CreateSupplier(ctx context.Context, companyID string, req *dto.CreateSupplierRequest) (*models.Supplier, error) {
	// Parse credit limit
	creditLimit := decimal.Zero
	if req.CreditLimit != nil && *req.CreditLimit != "" {
		var err error
		creditLimit, err = decimal.NewFromString(*req.CreditLimit)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid creditLimit format")
		}
		if creditLimit.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("credit limit cannot be negative")
		}
	}

	// Validate payment term
	if req.PaymentTerm != nil && *req.PaymentTerm < 0 {
		return nil, pkgerrors.NewBadRequestError("payment term cannot be negative")
	}

	// Validate code uniqueness per company
	if err := s.validateCodeUniqueness(companyID, req.Code, ""); err != nil {
		return nil, err
	}

	// Set defaults
	isPKP := false
	if req.IsPKP != nil {
		isPKP = *req.IsPKP
	}

	paymentTerm := 0
	if req.PaymentTerm != nil {
		paymentTerm = *req.PaymentTerm
	}

	// Create supplier
	supplier := &models.Supplier{
		CompanyID:          companyID,
		Code:               req.Code,
		Name:               req.Name,
		Type:               req.Type,
		Phone:              req.Phone,
		Email:              req.Email,
		Address:            req.Address,
		City:               req.City,
		Province:           req.Province,
		PostalCode:         req.PostalCode,
		NPWP:               req.NPWP,
		IsPKP:              isPKP,
		ContactPerson:      req.ContactPerson,
		ContactPhone:       req.ContactPhone,
		PaymentTerm:        paymentTerm,
		CreditLimit:        creditLimit,
		CurrentOutstanding: decimal.Zero,
		OverdueAmount:      decimal.Zero,
		Notes:              req.Notes,
		IsActive:           true,
	}

	if err := s.db.WithContext(ctx).Create(supplier).Error; err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

// ============================================================================
// LIST SUPPLIERS
// ============================================================================

// ListSuppliers retrieves suppliers with filtering, sorting, and pagination
func (s *SupplierService) ListSuppliers(ctx context.Context, companyID string, query *dto.SupplierListQuery) (*dto.SupplierListResponse, error) {
	// Set defaults
	page := 1
	if query.Page > 0 {
		page = query.Page
	}

	pageSize := 20
	if query.PageSize > 0 {
		pageSize = query.PageSize
	}

	sortBy := "created_at"
	if query.SortBy != "" {
		sortBy = query.SortBy
	}

	sortOrder := "desc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}

	// Build base query
	baseQuery := s.db.WithContext(ctx).Model(&models.Supplier{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		baseQuery = baseQuery.Where("code LIKE ? OR name LIKE ?", searchPattern, searchPattern)
	}

	if query.Type != nil {
		baseQuery = baseQuery.Where("type = ?", *query.Type)
	}

	if query.City != nil {
		baseQuery = baseQuery.Where("city = ?", *query.City)
	}

	if query.Province != nil {
		baseQuery = baseQuery.Where("province = ?", *query.Province)
	}

	if query.IsPKP != nil {
		baseQuery = baseQuery.Where("is_pkp = ?", *query.IsPKP)
	}

	if query.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *query.IsActive)
	} else {
		// Default: only show active suppliers
		baseQuery = baseQuery.Where("is_active = ?", true)
	}

	if query.HasOverdue != nil && *query.HasOverdue {
		baseQuery = baseQuery.Where("overdue_amount > 0")
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count suppliers: %w", err)
	}

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	var suppliers []models.Supplier
	if err := baseQuery.Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Find(&suppliers).Error; err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}

	// Map to response DTOs
	supplierResponses := make([]dto.SupplierResponse, len(suppliers))
	for i, supplier := range suppliers {
		supplierResponses[i] = mapSupplierToResponse(&supplier)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.SupplierListResponse{
		Suppliers:  supplierResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ============================================================================
// GET SUPPLIER BY ID
// ============================================================================

// GetSupplierByID retrieves a supplier by ID
func (s *SupplierService) GetSupplierByID(ctx context.Context, companyID, supplierID string) (*models.Supplier, error) {
	var supplier models.Supplier
	err := s.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, supplierID).
		First(&supplier).Error

	if err == gorm.ErrRecordNotFound {
		return nil, pkgerrors.NewNotFoundError("supplier not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return &supplier, nil
}

// ============================================================================
// UPDATE SUPPLIER
// ============================================================================

// UpdateSupplier updates an existing supplier
func (s *SupplierService) UpdateSupplier(ctx context.Context, companyID, supplierID string, req *dto.UpdateSupplierRequest) (*models.Supplier, error) {
	// Get existing supplier
	supplier, err := s.GetSupplierByID(ctx, companyID, supplierID)
	if err != nil {
		return nil, err
	}

	// Validate code uniqueness if updating code
	if req.Code != nil && *req.Code != supplier.Code {
		if err := s.validateCodeUniqueness(companyID, *req.Code, supplierID); err != nil {
			return nil, err
		}
		supplier.Code = *req.Code
	}

	// Update fields
	if req.Name != nil {
		supplier.Name = *req.Name
	}

	if req.Type != nil {
		supplier.Type = req.Type
	}

	if req.Phone != nil {
		supplier.Phone = req.Phone
	}

	if req.Email != nil {
		supplier.Email = req.Email
	}

	if req.Address != nil {
		supplier.Address = req.Address
	}

	if req.City != nil {
		supplier.City = req.City
	}

	if req.Province != nil {
		supplier.Province = req.Province
	}

	if req.PostalCode != nil {
		supplier.PostalCode = req.PostalCode
	}

	if req.NPWP != nil {
		supplier.NPWP = req.NPWP
	}

	if req.IsPKP != nil {
		supplier.IsPKP = *req.IsPKP
	}

	if req.ContactPerson != nil {
		supplier.ContactPerson = req.ContactPerson
	}

	if req.ContactPhone != nil {
		supplier.ContactPhone = req.ContactPhone
	}

	if req.PaymentTerm != nil {
		supplier.PaymentTerm = *req.PaymentTerm
	}

	if req.CreditLimit != nil && *req.CreditLimit != "" {
		creditLimit, err := decimal.NewFromString(*req.CreditLimit)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid creditLimit format")
		}
		if creditLimit.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("credit limit cannot be negative")
		}
		supplier.CreditLimit = creditLimit
	}

	if req.Notes != nil {
		supplier.Notes = req.Notes
	}

	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}

	// Save updates
	if err := s.db.WithContext(ctx).Save(supplier).Error; err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	return supplier, nil
}

// ============================================================================
// DELETE SUPPLIER (SOFT DELETE)
// ============================================================================

// DeleteSupplier soft deletes a supplier
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *SupplierService) DeleteSupplier(ctx context.Context, companyID, supplierID string) error {
	// Get supplier
	supplier, err := s.GetSupplierByID(ctx, companyID, supplierID)
	if err != nil {
		return err
	}

	// Validate deletion
	if err := s.validateDeleteSupplier(ctx, supplier); err != nil {
		return err
	}

	// Soft delete (set IsActive = false)
	supplier.IsActive = false
	if err := s.db.WithContext(ctx).Save(supplier).Error; err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapSupplierToResponse converts Supplier model to SupplierResponse DTO
func mapSupplierToResponse(supplier *models.Supplier) dto.SupplierResponse {
	return dto.SupplierResponse{
		ID:                 supplier.ID,
		Code:               supplier.Code,
		Name:               supplier.Name,
		Type:               supplier.Type,
		Phone:              supplier.Phone,
		Email:              supplier.Email,
		Address:            supplier.Address,
		City:               supplier.City,
		Province:           supplier.Province,
		PostalCode:         supplier.PostalCode,
		NPWP:               supplier.NPWP,
		IsPKP:              supplier.IsPKP,
		ContactPerson:      supplier.ContactPerson,
		ContactPhone:       supplier.ContactPhone,
		PaymentTerm:        supplier.PaymentTerm,
		CreditLimit:        supplier.CreditLimit.String(),
		CurrentOutstanding: supplier.CurrentOutstanding.String(),
		OverdueAmount:      supplier.OverdueAmount.String(),
		LastTransactionAt:  supplier.LastTransactionAt,
		Notes:              supplier.Notes,
		IsActive:           supplier.IsActive,
		CreatedAt:          supplier.CreatedAt,
		UpdatedAt:          supplier.UpdatedAt,
	}
}
