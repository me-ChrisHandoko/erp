package customer

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

// CustomerService - Business logic for customer management
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 2
type CustomerService struct {
	db *gorm.DB
}

// NewCustomerService creates a new customer service instance
func NewCustomerService(db *gorm.DB) *CustomerService {
	return &CustomerService{db: db}
}

// ============================================================================
// CREATE CUSTOMER
// ============================================================================

// CreateCustomer creates a new customer
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *CustomerService) CreateCustomer(ctx context.Context, companyID string, req *dto.CreateCustomerRequest) (*models.Customer, error) {
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

	// Create customer
	customer := &models.Customer{
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

	if err := s.db.WithContext(ctx).Create(customer).Error; err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// ============================================================================
// LIST CUSTOMERS
// ============================================================================

// ListCustomers retrieves customers with filtering, sorting, and pagination
func (s *CustomerService) ListCustomers(ctx context.Context, companyID string, query *dto.CustomerListQuery) (*dto.CustomerListResponse, error) {
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
	baseQuery := s.db.WithContext(ctx).Model(&models.Customer{}).
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
		// Default: only show active customers
		baseQuery = baseQuery.Where("is_active = ?", true)
	}

	if query.HasOverdue != nil && *query.HasOverdue {
		baseQuery = baseQuery.Where("overdue_amount > 0")
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	var customers []models.Customer
	if err := baseQuery.Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	// Map to response DTOs
	customerResponses := make([]dto.CustomerResponse, len(customers))
	for i, customer := range customers {
		customerResponses[i] = mapCustomerToResponse(&customer)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.CustomerListResponse{
		Customers:  customerResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ============================================================================
// GET CUSTOMER BY ID
// ============================================================================

// GetCustomerByID retrieves a customer by ID
func (s *CustomerService) GetCustomerByID(ctx context.Context, companyID, customerID string) (*models.Customer, error) {
	var customer models.Customer
	err := s.db.WithContext(ctx).
		Where("company_id = ? AND id = ?", companyID, customerID).
		First(&customer).Error

	if err == gorm.ErrRecordNotFound {
		return nil, pkgerrors.NewNotFoundError("customer not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

// ============================================================================
// UPDATE CUSTOMER
// ============================================================================

// UpdateCustomer updates an existing customer
func (s *CustomerService) UpdateCustomer(ctx context.Context, companyID, customerID string, req *dto.UpdateCustomerRequest) (*models.Customer, error) {
	// Get existing customer
	customer, err := s.GetCustomerByID(ctx, companyID, customerID)
	if err != nil {
		return nil, err
	}

	// Validate code uniqueness if updating code
	if req.Code != nil && *req.Code != customer.Code {
		if err := s.validateCodeUniqueness(companyID, *req.Code, customerID); err != nil {
			return nil, err
		}
		customer.Code = *req.Code
	}

	// Update fields
	if req.Name != nil {
		customer.Name = *req.Name
	}

	if req.Type != nil {
		customer.Type = req.Type
	}

	if req.Phone != nil {
		customer.Phone = req.Phone
	}

	if req.Email != nil {
		customer.Email = req.Email
	}

	if req.Address != nil {
		customer.Address = req.Address
	}

	if req.City != nil {
		customer.City = req.City
	}

	if req.Province != nil {
		customer.Province = req.Province
	}

	if req.PostalCode != nil {
		customer.PostalCode = req.PostalCode
	}

	if req.NPWP != nil {
		customer.NPWP = req.NPWP
	}

	if req.IsPKP != nil {
		customer.IsPKP = *req.IsPKP
	}

	if req.ContactPerson != nil {
		customer.ContactPerson = req.ContactPerson
	}

	if req.ContactPhone != nil {
		customer.ContactPhone = req.ContactPhone
	}

	if req.PaymentTerm != nil {
		customer.PaymentTerm = *req.PaymentTerm
	}

	if req.CreditLimit != nil && *req.CreditLimit != "" {
		creditLimit, err := decimal.NewFromString(*req.CreditLimit)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid creditLimit format")
		}
		if creditLimit.LessThan(decimal.Zero) {
			return nil, pkgerrors.NewBadRequestError("credit limit cannot be negative")
		}
		customer.CreditLimit = creditLimit
	}

	if req.Notes != nil {
		customer.Notes = req.Notes
	}

	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
	}

	// Save updates
	if err := s.db.WithContext(ctx).Save(customer).Error; err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// ============================================================================
// DELETE CUSTOMER (SOFT DELETE)
// ============================================================================

// DeleteCustomer soft deletes a customer
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *CustomerService) DeleteCustomer(ctx context.Context, companyID, customerID string) error {
	// Get customer
	customer, err := s.GetCustomerByID(ctx, companyID, customerID)
	if err != nil {
		return err
	}

	// Validate deletion
	if err := s.validateDeleteCustomer(ctx, customer); err != nil {
		return err
	}

	// Soft delete (set IsActive = false)
	customer.IsActive = false
	if err := s.db.WithContext(ctx).Save(customer).Error; err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapCustomerToResponse converts Customer model to CustomerResponse DTO
func mapCustomerToResponse(customer *models.Customer) dto.CustomerResponse {
	return dto.CustomerResponse{
		ID:                 customer.ID,
		Code:               customer.Code,
		Name:               customer.Name,
		Type:               customer.Type,
		Phone:              customer.Phone,
		Email:              customer.Email,
		Address:            customer.Address,
		City:               customer.City,
		Province:           customer.Province,
		PostalCode:         customer.PostalCode,
		NPWP:               customer.NPWP,
		IsPKP:              customer.IsPKP,
		ContactPerson:      customer.ContactPerson,
		ContactPhone:       customer.ContactPhone,
		PaymentTerm:        customer.PaymentTerm,
		CreditLimit:        customer.CreditLimit.String(),
		CurrentOutstanding: customer.CurrentOutstanding.String(),
		OverdueAmount:      customer.OverdueAmount.String(),
		LastTransactionAt:  customer.LastTransactionAt,
		Notes:              customer.Notes,
		IsActive:           customer.IsActive,
		CreatedAt:          customer.CreatedAt,
		UpdatedAt:          customer.UpdatedAt,
	}
}
