package customer

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// CustomerService - Business logic for customer management
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 2
type CustomerService struct {
	db           *gorm.DB
	auditService *audit.AuditService
}

// NewCustomerService creates a new customer service instance
func NewCustomerService(db *gorm.DB, auditService *audit.AuditService) *CustomerService {
	return &CustomerService{
		db:           db,
		auditService: auditService,
	}
}

// ============================================================================
// CREATE CUSTOMER
// ============================================================================

// CreateCustomer creates a new customer
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *CustomerService) CreateCustomer(ctx context.Context, tenantID, companyID string, userID string, ipAddress string, userAgent string, req *dto.CreateCustomerRequest) (*models.Customer, error) {
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
	if err := s.validateCodeUniqueness(tenantID, companyID, req.Code, ""); err != nil {
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

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Create(customer).Error; err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Audit logging
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	// Prepare customer data for audit
	customerData := map[string]interface{}{
		"code":                customer.Code,
		"name":                customer.Name,
		"type":                stringPtrToValue(customer.Type),
		"phone":               stringPtrToValue(customer.Phone),
		"email":               stringPtrToValue(customer.Email),
		"address":             stringPtrToValue(customer.Address),
		"city":                stringPtrToValue(customer.City),
		"province":            stringPtrToValue(customer.Province),
		"postal_code":         stringPtrToValue(customer.PostalCode),
		"npwp":                stringPtrToValue(customer.NPWP),
		"is_pkp":              customer.IsPKP,
		"contact_person":      stringPtrToValue(customer.ContactPerson),
		"contact_phone":       stringPtrToValue(customer.ContactPhone),
		"payment_term":        customer.PaymentTerm,
		"credit_limit":        customer.CreditLimit.String(),
		"current_outstanding": customer.CurrentOutstanding.String(),
		"overdue_amount":      customer.OverdueAmount.String(),
		"notes":               stringPtrToValue(customer.Notes),
	}

	// Log customer creation (async, don't block on audit failure)
	go s.auditService.LogCustomerCreated(context.Background(), auditCtx, customer.ID, customerData)

	return customer, nil
}

// ============================================================================
// LIST CUSTOMERS
// ============================================================================

// ListCustomers retrieves customers with filtering, sorting, and pagination
func (s *CustomerService) ListCustomers(ctx context.Context, tenantID, companyID string, query *dto.CustomerListQuery) (*dto.CustomerListResponse, error) {
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

	// Build base query with tenant context set for GORM callbacks
	// tenantID is explicitly passed from handler
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.Customer{}).
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
	}
	// No default filter - show all customers (active and inactive)

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
func (s *CustomerService) GetCustomerByID(ctx context.Context, tenantID, companyID, customerID string) (*models.Customer, error) {
	var customer models.Customer
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
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
func (s *CustomerService) UpdateCustomer(ctx context.Context, tenantID, companyID, customerID string, userID string, ipAddress string, userAgent string, req *dto.UpdateCustomerRequest) (*models.Customer, error) {
	// Get existing customer
	customer, err := s.GetCustomerByID(ctx, tenantID, companyID, customerID)
	if err != nil {
		return nil, err
	}

	// Capture old values for audit
	oldValues := map[string]interface{}{
		"code":                customer.Code,
		"name":                customer.Name,
		"type":                stringPtrToValue(customer.Type),
		"phone":               stringPtrToValue(customer.Phone),
		"email":               stringPtrToValue(customer.Email),
		"address":             stringPtrToValue(customer.Address),
		"city":                stringPtrToValue(customer.City),
		"province":            stringPtrToValue(customer.Province),
		"postal_code":         stringPtrToValue(customer.PostalCode),
		"npwp":                stringPtrToValue(customer.NPWP),
		"is_pkp":              customer.IsPKP,
		"contact_person":      stringPtrToValue(customer.ContactPerson),
		"contact_phone":       stringPtrToValue(customer.ContactPhone),
		"payment_term":        customer.PaymentTerm,
		"credit_limit":        customer.CreditLimit.String(),
		"current_outstanding": customer.CurrentOutstanding.String(),
		"overdue_amount":      customer.OverdueAmount.String(),
		"notes":               stringPtrToValue(customer.Notes),
		"is_active":           customer.IsActive,
	}

	// Validate code uniqueness if updating code
	if req.Code != nil && *req.Code != customer.Code {
		if err := s.validateCodeUniqueness(tenantID, companyID, *req.Code, customerID); err != nil {
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
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(customer).Error; err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Capture new values for audit
	newValues := map[string]interface{}{
		"code":                customer.Code,
		"name":                customer.Name,
		"type":                stringPtrToValue(customer.Type),
		"phone":               stringPtrToValue(customer.Phone),
		"email":               stringPtrToValue(customer.Email),
		"address":             stringPtrToValue(customer.Address),
		"city":                stringPtrToValue(customer.City),
		"province":            stringPtrToValue(customer.Province),
		"postal_code":         stringPtrToValue(customer.PostalCode),
		"npwp":                stringPtrToValue(customer.NPWP),
		"is_pkp":              customer.IsPKP,
		"contact_person":      stringPtrToValue(customer.ContactPerson),
		"contact_phone":       stringPtrToValue(customer.ContactPhone),
		"payment_term":        customer.PaymentTerm,
		"credit_limit":        customer.CreditLimit.String(),
		"current_outstanding": customer.CurrentOutstanding.String(),
		"overdue_amount":      customer.OverdueAmount.String(),
		"notes":               stringPtrToValue(customer.Notes),
		"is_active":           customer.IsActive,
	}

	// Audit logging
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	// Log customer update (async, don't block on audit failure)
	go s.auditService.LogCustomerUpdated(context.Background(), auditCtx, customer.ID, oldValues, newValues)

	return customer, nil
}

// ============================================================================
// DELETE CUSTOMER (SOFT DELETE)
// ============================================================================

// DeleteCustomer soft deletes a customer
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *CustomerService) DeleteCustomer(ctx context.Context, tenantID, companyID, customerID string, userID string, ipAddress string, userAgent string) error {
	// Get customer
	customer, err := s.GetCustomerByID(ctx, tenantID, companyID, customerID)
	if err != nil {
		return err
	}

	// Validate deletion
	if err := s.validateDeleteCustomer(ctx, customer); err != nil {
		return err
	}

	// Prepare customer data for audit (before deletion)
	customerData := map[string]interface{}{
		"code":                customer.Code,
		"name":                customer.Name,
		"type":                stringPtrToValue(customer.Type),
		"phone":               stringPtrToValue(customer.Phone),
		"email":               stringPtrToValue(customer.Email),
		"address":             stringPtrToValue(customer.Address),
		"city":                stringPtrToValue(customer.City),
		"province":            stringPtrToValue(customer.Province),
		"postal_code":         stringPtrToValue(customer.PostalCode),
		"npwp":                stringPtrToValue(customer.NPWP),
		"is_pkp":              customer.IsPKP,
		"contact_person":      stringPtrToValue(customer.ContactPerson),
		"contact_phone":       stringPtrToValue(customer.ContactPhone),
		"payment_term":        customer.PaymentTerm,
		"credit_limit":        customer.CreditLimit.String(),
		"current_outstanding": customer.CurrentOutstanding.String(),
		"overdue_amount":      customer.OverdueAmount.String(),
		"notes":               stringPtrToValue(customer.Notes),
		"is_active":           customer.IsActive,
	}

	// Soft delete (set IsActive = false)
	customer.IsActive = false
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(customer).Error; err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	// Audit logging
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	// Log customer deletion (async, don't block on audit failure)
	go s.auditService.LogCustomerDeleted(context.Background(), auditCtx, customer.ID, customerData)

	return nil
}

// ============================================================================
// GET FREQUENT PRODUCTS
// ============================================================================

// GetFrequentProducts retrieves the most frequently purchased products for a customer
// Optimized endpoint for Quick Add panel in Sales Orders
func (s *CustomerService) GetFrequentProducts(ctx context.Context, tenantID, companyID, customerID string, warehouseID *string, limit int) (*dto.FrequentProductsResponse, error) {
	// Validate customer exists
	customer, err := s.GetCustomerByID(ctx, tenantID, companyID, customerID)
	if err != nil {
		return nil, err
	}

	// Set default limit
	if limit <= 0 {
		limit = 8
	}

	// Define result structure
	type FrequentProductResult struct {
		ProductID     string
		ProductCode   string
		ProductName   string
		Frequency     int
		TotalQty      decimal.Decimal
		LastOrderDate string
		BaseUnit      string
		LatestPrice   decimal.Decimal
	}

	var results []FrequentProductResult

	// Build base query with tenant context
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Table("sales_order_items soi").
		Select(`
			soi.product_id,
			p.code as product_code,
			p.name as product_name,
			COUNT(DISTINCT so.id) as frequency,
			SUM(soi.quantity) as total_qty,
			MAX(so.so_date) as last_order_date,
			p.base_unit,
			(SELECT soi2.unit_price
			 FROM sales_order_items soi2
			 JOIN sales_orders so2 ON soi2.sales_order_id = so2.id
			 WHERE soi2.product_id = soi.product_id
			   AND so2.customer_id = ?
			   AND so2.company_id = ?
			 ORDER BY so2.so_date DESC
			 LIMIT 1) as latest_price
		`, customerID, companyID).
		Joins("JOIN sales_orders so ON soi.sales_order_id = so.id").
		Joins("JOIN products p ON soi.product_id = p.id").
		Where("so.customer_id = ?", customerID).
		Where("so.company_id = ?", companyID).
		Where("so.status IN ('APPROVED', 'COMPLETED', 'DELIVERED')"). // Only count fulfilled orders
		Group("soi.product_id, p.code, p.name, p.base_unit")

	// Optional warehouse filter
	if warehouseID != nil && *warehouseID != "" {
		query = query.Where("so.warehouse_id = ?", *warehouseID)
	}

	// Order by frequency (DESC), total quantity (DESC), last order date (DESC)
	query = query.Order("frequency DESC, total_qty DESC, last_order_date DESC").
		Limit(limit)

	// Execute query
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get frequent products: %w", err)
	}

	// Get date range from sales orders
	type DateRange struct {
		MinDate *string
		MaxDate *string
	}
	var dateRange DateRange

	dateRangeQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Table("sales_orders").
		Select("MIN(so_date) as min_date, MAX(so_date) as max_date").
		Where("customer_id = ?", customerID).
		Where("company_id = ?", companyID).
		Where("status IN ('APPROVED', 'COMPLETED', 'DELIVERED')")

	if warehouseID != nil && *warehouseID != "" {
		dateRangeQuery = dateRangeQuery.Where("warehouse_id = ?", *warehouseID)
	}

	if err := dateRangeQuery.Scan(&dateRange).Error; err != nil {
		return nil, fmt.Errorf("failed to get date range: %w", err)
	}

	// Count total orders
	var totalOrders int64
	ordersCountQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Model(&models.SalesOrder{}).
		Where("customer_id = ?", customerID).
		Where("company_id = ?", companyID).
		Where("status IN ('APPROVED', 'COMPLETED', 'DELIVERED')")

	if warehouseID != nil && *warehouseID != "" {
		ordersCountQuery = ordersCountQuery.Where("warehouse_id = ?", *warehouseID)
	}

	if err := ordersCountQuery.Count(&totalOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Map to response DTO
	frequentProducts := make([]dto.FrequentProductItem, len(results))
	for i, result := range results {
		frequentProducts[i] = dto.FrequentProductItem{
			ProductID:     result.ProductID,
			ProductCode:   result.ProductCode,
			ProductName:   result.ProductName,
			Frequency:     result.Frequency,
			TotalQty:      result.TotalQty.String(),
			LastOrderDate: result.LastOrderDate,
			BaseUnitID:    "", // Not using unit ID, just name
			BaseUnitName:  result.BaseUnit,
			LatestPrice:   result.LatestPrice.String(),
		}
	}

	// Build response
	response := &dto.FrequentProductsResponse{
		CustomerID:       customer.ID,
		CustomerName:     customer.Name,
		FrequentProducts: frequentProducts,
		TotalOrders:      int(totalOrders),
		DateRangeFrom:    "",
		DateRangeTo:      "",
	}

	if warehouseID != nil && *warehouseID != "" {
		response.WarehouseID = *warehouseID
	}

	if dateRange.MinDate != nil {
		response.DateRangeFrom = *dateRange.MinDate
	}

	if dateRange.MaxDate != nil {
		response.DateRangeTo = *dateRange.MaxDate
	}

	return response, nil
}

// GetCustomerCreditInfo retrieves customer credit limit and outstanding balance information
// Used for credit limit validation in sales orders
func (s *CustomerService) GetCustomerCreditInfo(ctx context.Context, tenantID, companyID, customerID string) (*dto.CustomerCreditInfoResponse, error) {
	// Validate customer exists and get customer data
	customer, err := s.GetCustomerByID(ctx, tenantID, companyID, customerID)
	if err != nil {
		return nil, err
	}

	// Calculate outstanding amount from unpaid/partially paid sales invoices
	type OutstandingResult struct {
		TotalOutstanding decimal.Decimal
		TotalOverdue     decimal.Decimal
	}

	var result OutstandingResult

	// TEMPORARY SOLUTION: Use sales_orders as estimation until sales_invoices is implemented
	// Query to calculate outstanding and overdue amounts from sales orders
	// Outstanding: Total amount from orders with status APPROVED or COMPLETED (not yet invoiced/paid)
	// Overdue: Orders where required_date has passed and still not DELIVERED
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Table("sales_orders so").
		Select(`
			COALESCE(SUM(so.total_amount), 0) as total_outstanding,
			COALESCE(SUM(
				CASE
					WHEN so.required_date < NOW() AND so.status NOT IN ('DELIVERED', 'CANCELLED')
					THEN so.total_amount
					ELSE 0
				END
			), 0) as total_overdue
		`).
		Where("so.customer_id = ?", customerID).
		Where("so.company_id = ?", companyID).
		Where("so.status IN ('APPROVED', 'COMPLETED')"). // Orders that are approved but not yet delivered
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate customer outstanding: %w", err)
	}

	// Calculate available credit (credit limit - outstanding)
	creditLimit := customer.CreditLimit
	outstandingAmount := result.TotalOutstanding
	availableCredit := creditLimit.Sub(outstandingAmount)

	// Check if exceeding limit
	isExceedingLimit := outstandingAmount.GreaterThan(creditLimit)

	// Calculate utilization percentage
	var utilizationPercent decimal.Decimal
	if creditLimit.GreaterThan(decimal.Zero) {
		utilizationPercent = outstandingAmount.Div(creditLimit).Mul(decimal.NewFromInt(100))
	} else {
		utilizationPercent = decimal.Zero
	}

	// Build response
	response := &dto.CustomerCreditInfoResponse{
		CustomerID:         customer.ID,
		CustomerName:       customer.Name,
		CustomerCode:       customer.Code,
		CreditLimit:        creditLimit.String(),
		OutstandingAmount:  outstandingAmount.String(),
		AvailableCredit:    availableCredit.String(),
		OverdueAmount:      result.TotalOverdue.String(),
		PaymentTermDays:    customer.PaymentTerm,
		IsExceedingLimit:   isExceedingLimit,
		UtilizationPercent: utilizationPercent.StringFixed(2),
	}

	return response, nil
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

// stringPtrToValue converts *string to string, returns empty string if nil
func stringPtrToValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
