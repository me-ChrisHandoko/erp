package supplier

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

// SupplierService - Business logic for supplier management
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Module 3
type SupplierService struct {
	db           *gorm.DB
	auditService *audit.AuditService
}

// NewSupplierService creates a new supplier service instance
func NewSupplierService(db *gorm.DB, auditService *audit.AuditService) *SupplierService {
	return &SupplierService{
		db:           db,
		auditService: auditService,
	}
}

// ============================================================================
// CREATE SUPPLIER
// ============================================================================

// CreateSupplier creates a new supplier
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.1 (Validation Rules)
func (s *SupplierService) CreateSupplier(ctx context.Context, tenantID, companyID, userID, ipAddress, userAgent string, req *dto.CreateSupplierRequest) (*models.Supplier, error) {
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
	if err := s.validateCodeUniqueness(ctx, tenantID, companyID, req.Code, ""); err != nil {
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
		TenantID:           tenantID,
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

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Create(supplier).Error; err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	// Audit logging - Log successful supplier creation
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	supplierData := map[string]interface{}{
		"code":            supplier.Code,
		"name":            supplier.Name,
		"type":            supplier.Type,
		"phone":           supplier.Phone,
		"email":           supplier.Email,
		"address":         supplier.Address,
		"city":            supplier.City,
		"province":        supplier.Province,
		"postal_code":     supplier.PostalCode,
		"npwp":            supplier.NPWP,
		"is_pkp":          supplier.IsPKP,
		"contact_person":  supplier.ContactPerson,
		"contact_phone":   supplier.ContactPhone,
		"payment_term":    supplier.PaymentTerm,
		"credit_limit":    supplier.CreditLimit.String(),
		"notes":           supplier.Notes,
		"is_active":       supplier.IsActive,
	}

	if err := s.auditService.LogSupplierCreated(ctx, auditCtx, supplier.ID, supplierData); err != nil {
		fmt.Printf("WARNING: Failed to create audit log: %v\n", err)
	}

	return supplier, nil
}

// ============================================================================
// LIST SUPPLIERS
// ============================================================================

// ListSuppliers retrieves suppliers with filtering, sorting, and pagination
func (s *SupplierService) ListSuppliers(ctx context.Context, tenantID, companyID string, query *dto.SupplierListQuery) (*dto.SupplierListResponse, error) {
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
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.Supplier{}).
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

	// Filter by isActive status
	// If IsActive is nil, show ALL suppliers (both active and inactive)
	// If IsActive is provided, filter by the specified status
	if query.IsActive != nil {
		baseQuery = baseQuery.Where("is_active = ?", *query.IsActive)
	}
	// No default filter - show all when not specified

	if query.HasOverdue != nil && *query.HasOverdue {
		baseQuery = baseQuery.Where("overdue_amount > 0")
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		println("❌ [ListSuppliers] Count error:", err.Error())
		return nil, fmt.Errorf("failed to count suppliers: %w", err)
	}
	println("✅ [ListSuppliers] Total count:", totalCount)

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	var suppliers []models.Supplier
	if err := baseQuery.Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Find(&suppliers).Error; err != nil {
		println("❌ [ListSuppliers] Query error:", err.Error())
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}
	println("✅ [ListSuppliers] Found", len(suppliers), "suppliers")

	// Map to response DTOs
	supplierResponses := make([]dto.SupplierResponse, len(suppliers))
	for i, supplier := range suppliers {
		supplierResponses[i] = mapSupplierToResponse(&supplier)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.SupplierListResponse{
		Success: true,
		Data:    supplierResponses,
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      pageSize,
			Total:      int(totalCount),
			TotalPages: totalPages,
		},
	}, nil
}

// ============================================================================
// GET SUPPLIER BY ID
// ============================================================================

// GetSupplierByID retrieves a supplier by ID
func (s *SupplierService) GetSupplierByID(ctx context.Context, tenantID, companyID, supplierID string) (*models.Supplier, error) {
	var supplier models.Supplier
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
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
func (s *SupplierService) UpdateSupplier(ctx context.Context, tenantID, companyID, supplierID, userID, ipAddress, userAgent string, req *dto.UpdateSupplierRequest) (*models.Supplier, error) {
	// Get existing supplier
	supplier, err := s.GetSupplierByID(ctx, tenantID, companyID, supplierID)
	if err != nil {
		return nil, err
	}

	// Capture old values for audit logging
	oldValues := map[string]interface{}{
		"code":            supplier.Code,
		"name":            supplier.Name,
		"type":            supplier.Type,
		"phone":           supplier.Phone,
		"email":           supplier.Email,
		"address":         supplier.Address,
		"city":            supplier.City,
		"province":        supplier.Province,
		"postal_code":     supplier.PostalCode,
		"npwp":            supplier.NPWP,
		"is_pkp":          supplier.IsPKP,
		"contact_person":  supplier.ContactPerson,
		"contact_phone":   supplier.ContactPhone,
		"payment_term":    supplier.PaymentTerm,
		"credit_limit":    supplier.CreditLimit.String(),
		"notes":           supplier.Notes,
		"is_active":       supplier.IsActive,
	}

	// Validate code uniqueness if updating code
	if req.Code != nil && *req.Code != supplier.Code {
		if err := s.validateCodeUniqueness(ctx, tenantID, companyID, *req.Code, supplierID); err != nil {
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
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(supplier).Error; err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	// Audit logging - Log successful supplier update
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	newValues := map[string]interface{}{
		"code":            supplier.Code,
		"name":            supplier.Name,
		"type":            supplier.Type,
		"phone":           supplier.Phone,
		"email":           supplier.Email,
		"address":         supplier.Address,
		"city":            supplier.City,
		"province":        supplier.Province,
		"postal_code":     supplier.PostalCode,
		"npwp":            supplier.NPWP,
		"is_pkp":          supplier.IsPKP,
		"contact_person":  supplier.ContactPerson,
		"contact_phone":   supplier.ContactPhone,
		"payment_term":    supplier.PaymentTerm,
		"credit_limit":    supplier.CreditLimit.String(),
		"notes":           supplier.Notes,
		"is_active":       supplier.IsActive,
	}

	if err := s.auditService.LogSupplierUpdated(ctx, auditCtx, supplier.ID, oldValues, newValues); err != nil {
		fmt.Printf("WARNING: Failed to create audit log: %v\n", err)
	}

	return supplier, nil
}

// ============================================================================
// DELETE SUPPLIER (SOFT DELETE)
// ============================================================================

// DeleteSupplier soft deletes a supplier
// Reference: ANALYSIS-02-MASTER-DATA-MANAGEMENT.md Section 5.3 (Soft Delete Rules)
func (s *SupplierService) DeleteSupplier(ctx context.Context, tenantID, companyID, supplierID, userID, ipAddress, userAgent string) error {
	// Get supplier
	supplier, err := s.GetSupplierByID(ctx, tenantID, companyID, supplierID)
	if err != nil {
		return err
	}

	// Capture supplier data for audit logging before deletion
	supplierData := map[string]interface{}{
		"code":            supplier.Code,
		"name":            supplier.Name,
		"type":            supplier.Type,
		"phone":           supplier.Phone,
		"email":           supplier.Email,
		"address":         supplier.Address,
		"city":            supplier.City,
		"province":        supplier.Province,
		"postal_code":     supplier.PostalCode,
		"npwp":            supplier.NPWP,
		"is_pkp":          supplier.IsPKP,
		"contact_person":  supplier.ContactPerson,
		"contact_phone":   supplier.ContactPhone,
		"payment_term":    supplier.PaymentTerm,
		"credit_limit":    supplier.CreditLimit.String(),
		"notes":           supplier.Notes,
		"is_active":       supplier.IsActive,
	}

	// Validate deletion
	if err := s.validateDeleteSupplier(ctx, tenantID, supplier); err != nil {
		return err
	}

	// Soft delete (set IsActive = false)
	supplier.IsActive = false
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(supplier).Error; err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	// Audit logging - Log successful supplier deletion
	requestID := uuid.New().String()
	auditCtx := &audit.AuditContext{
		TenantID:  &tenantID,
		CompanyID: &companyID,
		UserID:    &userID,
		RequestID: &requestID,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
	}

	if err := s.auditService.LogSupplierDeleted(ctx, auditCtx, supplier.ID, supplierData); err != nil {
		fmt.Printf("WARNING: Failed to create audit log: %v\n", err)
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
