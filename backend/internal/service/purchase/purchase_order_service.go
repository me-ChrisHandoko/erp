package purchase

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/document"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// PurchaseOrderService - Business logic for purchase order management
type PurchaseOrderService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
}

// NewPurchaseOrderService creates a new purchase order service instance
func NewPurchaseOrderService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *PurchaseOrderService {
	return &PurchaseOrderService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ============================================================================
// CREATE PURCHASE ORDER
// ============================================================================

// CreatePurchaseOrder creates a new purchase order
func (s *PurchaseOrderService) CreatePurchaseOrder(ctx context.Context, tenantID, companyID, userID string, req *dto.CreatePurchaseOrderRequest) (*models.PurchaseOrder, error) {
	// Parse PO date
	poDate, err := time.Parse("2006-01-02", req.PODate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid poDate format, expected YYYY-MM-DD")
	}

	// Parse expected delivery date if provided
	var expectedDeliveryAt *time.Time
	if req.ExpectedDeliveryAt != nil && *req.ExpectedDeliveryAt != "" {
		parsed, err := time.Parse("2006-01-02", *req.ExpectedDeliveryAt)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid expectedDeliveryAt format, expected YYYY-MM-DD")
		}
		expectedDeliveryAt = &parsed
	}

	// Validate supplier exists
	var supplier models.Supplier
	if err := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND is_active = true", req.SupplierID, companyID).First(&supplier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewBadRequestError("supplier not found or inactive")
		}
		return nil, fmt.Errorf("failed to validate supplier: %w", err)
	}

	// Validate warehouse exists
	var warehouse models.Warehouse
	if err := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND is_active = true", req.WarehouseID, companyID).First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewBadRequestError("warehouse not found or inactive")
		}
		return nil, fmt.Errorf("failed to validate warehouse: %w", err)
	}

	// Generate PO number using document number generator
	poNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypePurchaseOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PO number: %w", err)
	}

	// Parse discount and tax amounts
	discountAmount := decimal.Zero
	if req.DiscountAmount != nil && *req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(*req.DiscountAmount)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid discountAmount format")
		}
	}

	taxAmount := decimal.Zero
	if req.TaxAmount != nil && *req.TaxAmount != "" {
		taxAmount, err = decimal.NewFromString(*req.TaxAmount)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid taxAmount format")
		}
	}

	// Create purchase order in transaction
	var purchaseOrder *models.PurchaseOrder

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create purchase order header
		purchaseOrder = &models.PurchaseOrder{
			TenantID:           tenantID,
			CompanyID:          companyID,
			PONumber:           poNumber,
			PODate:             poDate,
			SupplierID:         req.SupplierID,
			WarehouseID:        req.WarehouseID,
			Status:             models.PurchaseOrderStatusDraft,
			DiscountAmount:     discountAmount,
			TaxAmount:          taxAmount,
			Notes:              req.Notes,
			ExpectedDeliveryAt: expectedDeliveryAt,
			RequestedBy:        &userID,
		}

		if err := tx.Create(purchaseOrder).Error; err != nil {
			return fmt.Errorf("failed to create purchase order: %w", err)
		}

		// Create purchase order items
		subtotal := decimal.Zero
		for _, itemReq := range req.Items {
			item, itemSubtotal, err := s.createPurchaseOrderItem(ctx, tx, companyID, purchaseOrder.ID, &itemReq)
			if err != nil {
				return err
			}
			subtotal = subtotal.Add(itemSubtotal)
			purchaseOrder.Items = append(purchaseOrder.Items, *item)
		}

		// Update totals
		purchaseOrder.Subtotal = subtotal
		purchaseOrder.TotalAmount = subtotal.Sub(discountAmount).Add(taxAmount)

		if err := tx.Save(purchaseOrder).Error; err != nil {
			return fmt.Errorf("failed to update purchase order totals: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load relations
	if err := s.loadPurchaseOrderRelations(ctx, tenantID, purchaseOrder); err != nil {
		return nil, err
	}

	return purchaseOrder, nil
}

// createPurchaseOrderItem creates a single purchase order item
func (s *PurchaseOrderService) createPurchaseOrderItem(ctx context.Context, tx *gorm.DB, companyID, purchaseOrderID string, req *dto.CreatePurchaseOrderItemRequest) (*models.PurchaseOrderItem, decimal.Decimal, error) {
	// Validate product exists
	var product models.Product
	if err := tx.WithContext(ctx).Where("id = ? AND company_id = ? AND is_active = true", req.ProductID, companyID).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, decimal.Zero, pkgerrors.NewBadRequestError(fmt.Sprintf("product %s not found or inactive", req.ProductID))
		}
		return nil, decimal.Zero, fmt.Errorf("failed to validate product: %w", err)
	}

	// Validate product unit if provided
	if req.ProductUnitID != nil && *req.ProductUnitID != "" {
		var productUnit models.ProductUnit
		if err := tx.WithContext(ctx).Where("id = ? AND product_id = ?", *req.ProductUnitID, req.ProductID).First(&productUnit).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, decimal.Zero, pkgerrors.NewBadRequestError("product unit not found")
			}
			return nil, decimal.Zero, fmt.Errorf("failed to validate product unit: %w", err)
		}
	}

	// Parse quantities and prices
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, decimal.Zero, pkgerrors.NewBadRequestError("invalid quantity format")
	}
	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, decimal.Zero, pkgerrors.NewBadRequestError("quantity must be positive")
	}

	unitPrice, err := decimal.NewFromString(req.UnitPrice)
	if err != nil {
		return nil, decimal.Zero, pkgerrors.NewBadRequestError("invalid unitPrice format")
	}
	if unitPrice.LessThan(decimal.Zero) {
		return nil, decimal.Zero, pkgerrors.NewBadRequestError("unitPrice cannot be negative")
	}

	discountPct := decimal.Zero
	if req.DiscountPct != nil && *req.DiscountPct != "" {
		discountPct, err = decimal.NewFromString(*req.DiscountPct)
		if err != nil {
			return nil, decimal.Zero, pkgerrors.NewBadRequestError("invalid discountPct format")
		}
		if discountPct.LessThan(decimal.Zero) || discountPct.GreaterThan(decimal.NewFromInt(100)) {
			return nil, decimal.Zero, pkgerrors.NewBadRequestError("discountPct must be between 0 and 100")
		}
	}

	// Calculate subtotal
	lineTotal := quantity.Mul(unitPrice)
	discountAmt := lineTotal.Mul(discountPct).Div(decimal.NewFromInt(100))
	subtotal := lineTotal.Sub(discountAmt)

	// Create item
	item := &models.PurchaseOrderItem{
		PurchaseOrderID: purchaseOrderID,
		ProductID:       req.ProductID,
		ProductUnitID:   req.ProductUnitID,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		DiscountPct:     discountPct,
		DiscountAmt:     discountAmt,
		Subtotal:        subtotal,
		ReceivedQty:     decimal.Zero,
		Notes:           req.Notes,
	}

	if err := tx.Create(item).Error; err != nil {
		return nil, decimal.Zero, fmt.Errorf("failed to create purchase order item: %w", err)
	}

	return item, subtotal, nil
}


// ============================================================================
// LIST PURCHASE ORDERS
// ============================================================================

// ListPurchaseOrders retrieves purchase orders with filtering and pagination
func (s *PurchaseOrderService) ListPurchaseOrders(ctx context.Context, tenantID, companyID string, query *dto.PurchaseOrderListQuery) (*dto.PurchaseOrderListResponse, error) {
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
		// Map camelCase to snake_case
		sortByMap := map[string]string{
			"poNumber":    "po_number",
			"poDate":      "po_date",
			"totalAmount": "total_amount",
			"status":      "status",
			"createdAt":   "created_at",
		}
		if mapped, ok := sortByMap[query.SortBy]; ok {
			sortBy = mapped
		} else {
			sortBy = query.SortBy
		}
	}

	sortOrder := "desc"
	if query.SortOrder != "" {
		sortOrder = query.SortOrder
	}

	// Build base query with tenant context set for GORM callbacks
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.PurchaseOrder{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		baseQuery = baseQuery.Where("po_number LIKE ?", searchPattern)
	}

	if query.Status != nil {
		baseQuery = baseQuery.Where("status = ?", *query.Status)
	}

	if query.SupplierID != nil {
		baseQuery = baseQuery.Where("supplier_id = ?", *query.SupplierID)
	}

	if query.WarehouseID != nil {
		baseQuery = baseQuery.Where("warehouse_id = ?", *query.WarehouseID)
	}

	if query.DateFrom != nil && *query.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", *query.DateFrom)
		if err == nil {
			baseQuery = baseQuery.Where("po_date >= ?", dateFrom)
		}
	}

	if query.DateTo != nil && *query.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", *query.DateTo)
		if err == nil {
			baseQuery = baseQuery.Where("po_date <= ?", dateTo)
		}
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count purchase orders: %w", err)
	}

	// Apply sorting and pagination
	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	var purchaseOrders []models.PurchaseOrder
	if err := baseQuery.
		Preload("Supplier").
		Preload("Warehouse").
		Preload("Requester").
		Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to list purchase orders: %w", err)
	}

	// Map to response DTOs
	responses := make([]dto.PurchaseOrderResponse, len(purchaseOrders))
	for i, po := range purchaseOrders {
		responses[i] = s.mapPurchaseOrderToResponse(&po, false)
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.PurchaseOrderListResponse{
		Success: true,
		Data:    responses,
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      pageSize,
			Total:      int(totalCount),
			TotalPages: totalPages,
		},
	}, nil
}

// ============================================================================
// GET PURCHASE ORDER BY ID
// ============================================================================

// GetPurchaseOrderByID retrieves a purchase order by ID with all relations
func (s *PurchaseOrderService) GetPurchaseOrderByID(ctx context.Context, tenantID, companyID, purchaseOrderID string) (*models.PurchaseOrder, error) {
	var purchaseOrder models.PurchaseOrder
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ? AND id = ?", companyID, purchaseOrderID).
		First(&purchaseOrder).Error

	if err == gorm.ErrRecordNotFound {
		return nil, pkgerrors.NewNotFoundError("purchase order not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	// Load relations
	if err := s.loadPurchaseOrderRelations(ctx, tenantID, &purchaseOrder); err != nil {
		return nil, err
	}

	return &purchaseOrder, nil
}

// loadPurchaseOrderRelations loads all relations for a purchase order
func (s *PurchaseOrderService) loadPurchaseOrderRelations(ctx context.Context, tenantID string, po *models.PurchaseOrder) error {
	// Load supplier
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("id = ?", po.SupplierID).First(&po.Supplier).Error; err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to load supplier: %w", err)
	}

	// Load warehouse
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("id = ?", po.WarehouseID).First(&po.Warehouse).Error; err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to load warehouse: %w", err)
	}

	// Load requester
	if po.RequestedBy != nil {
		var user models.User
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("id = ?", *po.RequestedBy).First(&user).Error; err == nil {
			po.Requester = &user
		}
	}

	// Load items with product info
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		Preload("ProductUnit").
		Where("purchase_order_id = ?", po.ID).
		Find(&po.Items).Error; err != nil {
		return fmt.Errorf("failed to load items: %w", err)
	}

	return nil
}

// ============================================================================
// UPDATE PURCHASE ORDER
// ============================================================================

// UpdatePurchaseOrder updates an existing purchase order (only DRAFT status)
func (s *PurchaseOrderService) UpdatePurchaseOrder(ctx context.Context, tenantID, companyID, purchaseOrderID string, req *dto.UpdatePurchaseOrderRequest) (*models.PurchaseOrder, error) {
	// Get existing purchase order
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, tenantID, companyID, purchaseOrderID)
	if err != nil {
		return nil, err
	}

	// Check status - can only update DRAFT
	if purchaseOrder.Status != models.PurchaseOrderStatusDraft {
		return nil, pkgerrors.NewBadRequestError("can only update purchase orders in DRAFT status")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update header fields
		if req.SupplierID != nil {
			// Validate supplier
			var supplier models.Supplier
			if err := tx.Where("id = ? AND company_id = ? AND is_active = true", *req.SupplierID, companyID).First(&supplier).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewBadRequestError("supplier not found or inactive")
				}
				return err
			}
			purchaseOrder.SupplierID = *req.SupplierID
		}

		if req.WarehouseID != nil {
			// Validate warehouse
			var warehouse models.Warehouse
			if err := tx.Where("id = ? AND company_id = ? AND is_active = true", *req.WarehouseID, companyID).First(&warehouse).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewBadRequestError("warehouse not found or inactive")
				}
				return err
			}
			purchaseOrder.WarehouseID = *req.WarehouseID
		}

		if req.PODate != nil {
			poDate, err := time.Parse("2006-01-02", *req.PODate)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid poDate format")
			}
			purchaseOrder.PODate = poDate
		}

		if req.ExpectedDeliveryAt != nil {
			if *req.ExpectedDeliveryAt == "" {
				purchaseOrder.ExpectedDeliveryAt = nil
			} else {
				parsed, err := time.Parse("2006-01-02", *req.ExpectedDeliveryAt)
				if err != nil {
					return pkgerrors.NewBadRequestError("invalid expectedDeliveryAt format")
				}
				purchaseOrder.ExpectedDeliveryAt = &parsed
			}
		}

		if req.DiscountAmount != nil {
			discountAmount, err := decimal.NewFromString(*req.DiscountAmount)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid discountAmount format")
			}
			purchaseOrder.DiscountAmount = discountAmount
		}

		if req.TaxAmount != nil {
			taxAmount, err := decimal.NewFromString(*req.TaxAmount)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid taxAmount format")
			}
			purchaseOrder.TaxAmount = taxAmount
		}

		if req.Notes != nil {
			purchaseOrder.Notes = req.Notes
		}

		// Update items if provided
		if req.Items != nil && len(req.Items) > 0 {
			// Delete existing items
			if err := tx.Where("purchase_order_id = ?", purchaseOrderID).Delete(&models.PurchaseOrderItem{}).Error; err != nil {
				return fmt.Errorf("failed to delete existing items: %w", err)
			}

			// Create new items
			purchaseOrder.Items = []models.PurchaseOrderItem{}
			subtotal := decimal.Zero
			for _, itemReq := range req.Items {
				createReq := &dto.CreatePurchaseOrderItemRequest{
					ProductID:     itemReq.ProductID,
					ProductUnitID: itemReq.ProductUnitID,
					Quantity:      itemReq.Quantity,
					UnitPrice:     itemReq.UnitPrice,
					DiscountPct:   itemReq.DiscountPct,
					Notes:         itemReq.Notes,
				}
				item, itemSubtotal, err := s.createPurchaseOrderItem(ctx, tx, companyID, purchaseOrderID, createReq)
				if err != nil {
					return err
				}
				subtotal = subtotal.Add(itemSubtotal)
				purchaseOrder.Items = append(purchaseOrder.Items, *item)
			}
			purchaseOrder.Subtotal = subtotal
		}

		// Recalculate total
		purchaseOrder.TotalAmount = purchaseOrder.Subtotal.Sub(purchaseOrder.DiscountAmount).Add(purchaseOrder.TaxAmount)

		// Save updates
		if err := tx.Save(purchaseOrder).Error; err != nil {
			return fmt.Errorf("failed to update purchase order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload relations
	if err := s.loadPurchaseOrderRelations(ctx, tenantID, purchaseOrder); err != nil {
		return nil, err
	}

	return purchaseOrder, nil
}

// ============================================================================
// DELETE PURCHASE ORDER
// ============================================================================

// DeletePurchaseOrder deletes a purchase order (only DRAFT status)
func (s *PurchaseOrderService) DeletePurchaseOrder(ctx context.Context, tenantID, companyID, purchaseOrderID string) error {
	// Get purchase order
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, tenantID, companyID, purchaseOrderID)
	if err != nil {
		return err
	}

	// Check status - can only delete DRAFT
	if purchaseOrder.Status != models.PurchaseOrderStatusDraft {
		return pkgerrors.NewBadRequestError("can only delete purchase orders in DRAFT status")
	}

	// Delete in transaction
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete items first
		if err := tx.Where("purchase_order_id = ?", purchaseOrderID).Delete(&models.PurchaseOrderItem{}).Error; err != nil {
			return fmt.Errorf("failed to delete purchase order items: %w", err)
		}

		// Delete purchase order
		if err := tx.Delete(purchaseOrder).Error; err != nil {
			return fmt.Errorf("failed to delete purchase order: %w", err)
		}

		return nil
	})
}

// ============================================================================
// STATUS TRANSITIONS
// ============================================================================

// ConfirmPurchaseOrder confirms a purchase order (DRAFT -> CONFIRMED)
func (s *PurchaseOrderService) ConfirmPurchaseOrder(ctx context.Context, tenantID, companyID, purchaseOrderID, userID string) (*models.PurchaseOrder, error) {
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, tenantID, companyID, purchaseOrderID)
	if err != nil {
		return nil, err
	}

	if purchaseOrder.Status != models.PurchaseOrderStatusDraft {
		return nil, pkgerrors.NewBadRequestError("can only confirm purchase orders in DRAFT status")
	}

	// Validate has items
	if len(purchaseOrder.Items) == 0 {
		return nil, pkgerrors.NewBadRequestError("cannot confirm purchase order without items")
	}

	now := time.Now()
	purchaseOrder.Status = models.PurchaseOrderStatusConfirmed
	purchaseOrder.ApprovedBy = &userID
	purchaseOrder.ApprovedAt = &now

	if err := s.db.WithContext(ctx).Save(purchaseOrder).Error; err != nil {
		return nil, fmt.Errorf("failed to confirm purchase order: %w", err)
	}

	return purchaseOrder, nil
}

// CompletePurchaseOrder completes a purchase order (CONFIRMED -> COMPLETED)
func (s *PurchaseOrderService) CompletePurchaseOrder(ctx context.Context, tenantID, companyID, purchaseOrderID string) (*models.PurchaseOrder, error) {
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, tenantID, companyID, purchaseOrderID)
	if err != nil {
		return nil, err
	}

	if purchaseOrder.Status != models.PurchaseOrderStatusConfirmed {
		return nil, pkgerrors.NewBadRequestError("can only complete purchase orders in CONFIRMED status")
	}

	purchaseOrder.Status = models.PurchaseOrderStatusCompleted

	if err := s.db.WithContext(ctx).Save(purchaseOrder).Error; err != nil {
		return nil, fmt.Errorf("failed to complete purchase order: %w", err)
	}

	return purchaseOrder, nil
}

// CancelPurchaseOrder cancels a purchase order (DRAFT/CONFIRMED -> CANCELLED)
func (s *PurchaseOrderService) CancelPurchaseOrder(ctx context.Context, tenantID, companyID, purchaseOrderID, userID string, req *dto.CancelPurchaseOrderRequest) (*models.PurchaseOrder, error) {
	purchaseOrder, err := s.GetPurchaseOrderByID(ctx, tenantID, companyID, purchaseOrderID)
	if err != nil {
		return nil, err
	}

	if purchaseOrder.Status != models.PurchaseOrderStatusDraft && purchaseOrder.Status != models.PurchaseOrderStatusConfirmed {
		return nil, pkgerrors.NewBadRequestError("can only cancel purchase orders in DRAFT or CONFIRMED status")
	}

	now := time.Now()
	purchaseOrder.Status = models.PurchaseOrderStatusCancelled
	purchaseOrder.CancelledBy = &userID
	purchaseOrder.CancelledAt = &now
	purchaseOrder.CancellationNote = &req.CancellationNote

	if err := s.db.WithContext(ctx).Save(purchaseOrder).Error; err != nil {
		return nil, fmt.Errorf("failed to cancel purchase order: %w", err)
	}

	return purchaseOrder, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// mapPurchaseOrderToResponse converts PurchaseOrder model to PurchaseOrderResponse DTO
func (s *PurchaseOrderService) mapPurchaseOrderToResponse(po *models.PurchaseOrder, includeItems bool) dto.PurchaseOrderResponse {
	response := dto.PurchaseOrderResponse{
		ID:                 po.ID,
		PONumber:           po.PONumber,
		PODate:             po.PODate,
		SupplierID:         po.SupplierID,
		WarehouseID:        po.WarehouseID,
		Status:             string(po.Status),
		Subtotal:           po.Subtotal.String(),
		DiscountAmount:     po.DiscountAmount.String(),
		TaxAmount:          po.TaxAmount.String(),
		TotalAmount:        po.TotalAmount.String(),
		Notes:              po.Notes,
		ExpectedDeliveryAt: po.ExpectedDeliveryAt,
		RequestedBy:        po.RequestedBy,
		ApprovedBy:         po.ApprovedBy,
		ApprovedAt:         po.ApprovedAt,
		CancelledBy:        po.CancelledBy,
		CancelledAt:        po.CancelledAt,
		CancellationNote:   po.CancellationNote,
		CreatedAt:          po.CreatedAt,
		UpdatedAt:          po.UpdatedAt,
	}

	// Map supplier if loaded
	if po.Supplier.ID != "" {
		response.Supplier = &dto.SupplierBasicResponse{
			ID:   po.Supplier.ID,
			Code: po.Supplier.Code,
			Name: po.Supplier.Name,
		}
	}

	// Map warehouse if loaded
	if po.Warehouse.ID != "" {
		response.Warehouse = &dto.WarehouseBasicResponse{
			ID:   po.Warehouse.ID,
			Code: po.Warehouse.Code,
			Name: po.Warehouse.Name,
		}
	}

	// Map requester if loaded
	if po.Requester != nil {
		response.Requester = &dto.UserBasicResponse{
			ID:       po.Requester.ID,
			FullName: po.Requester.FullName,
		}
	}

	// Map items if requested
	if includeItems && len(po.Items) > 0 {
		response.Items = make([]dto.PurchaseOrderItemResponse, len(po.Items))
		for i, item := range po.Items {
			response.Items[i] = s.mapPurchaseOrderItemToResponse(&item)
		}
	}

	return response
}

// mapPurchaseOrderItemToResponse converts PurchaseOrderItem model to PurchaseOrderItemResponse DTO
func (s *PurchaseOrderService) mapPurchaseOrderItemToResponse(item *models.PurchaseOrderItem) dto.PurchaseOrderItemResponse {
	response := dto.PurchaseOrderItemResponse{
		ID:          item.ID,
		ProductID:   item.ProductID,
		Quantity:    item.Quantity.String(),
		UnitPrice:   item.UnitPrice.String(),
		DiscountPct: item.DiscountPct.String(),
		DiscountAmt: item.DiscountAmt.String(),
		Subtotal:    item.Subtotal.String(),
		ReceivedQty: item.ReceivedQty.String(),
		Notes:       item.Notes,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	// Map product if loaded (including tracking flags for goods receipt validation)
	if item.Product.ID != "" {
		response.Product = &dto.PurchaseOrderProductResponse{
			ID:             item.Product.ID,
			Code:           item.Product.Code,
			Name:           item.Product.Name,
			BaseUnit:       item.Product.BaseUnit,
			IsBatchTracked: item.Product.IsBatchTracked,
			IsPerishable:   item.Product.IsPerishable,
		}
	}

	// Map product unit if loaded
	if item.ProductUnit != nil && item.ProductUnit.ID != "" {
		response.ProductUnitID = &item.ProductUnit.ID
		response.ProductUnit = &dto.PurchaseOrderProductUnitResponse{
			ID:             item.ProductUnit.ID,
			UnitName:       item.ProductUnit.UnitName,
			ConversionRate: item.ProductUnit.ConversionRate.String(),
		}
	}

	return response
}

// MapToResponse is a public method to map purchase order to response
func (s *PurchaseOrderService) MapToResponse(po *models.PurchaseOrder, includeItems bool) dto.PurchaseOrderResponse {
	return s.mapPurchaseOrderToResponse(po, includeItems)
}

// GenerateRequestID generates a unique request ID for audit logging
func GenerateRequestID() string {
	return uuid.New().String()
}
