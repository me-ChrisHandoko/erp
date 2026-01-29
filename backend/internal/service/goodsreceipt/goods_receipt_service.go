package goodsreceipt

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/internal/service/deliverytolerance"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// GoodsReceiptService - Business logic for goods receipt management
type GoodsReceiptService struct {
	db               *gorm.DB
	auditService     *audit.AuditService
	toleranceService *deliverytolerance.DeliveryToleranceService
}

// NewGoodsReceiptService creates a new goods receipt service instance
func NewGoodsReceiptService(db *gorm.DB, auditService *audit.AuditService, toleranceService *deliverytolerance.DeliveryToleranceService) *GoodsReceiptService {
	return &GoodsReceiptService{
		db:               db,
		auditService:     auditService,
		toleranceService: toleranceService,
	}
}

// ToleranceValidationResult holds the result of tolerance validation
type ToleranceValidationResult struct {
	IsValid              bool
	EffectiveTolerance   *decimal.Decimal // Under tolerance percentage
	OverTolerance        *decimal.Decimal // Over tolerance percentage
	UnlimitedOver        bool
	ResolvedFrom         string // PRODUCT, CATEGORY, COMPANY, DEFAULT
	MinAllowedQty        decimal.Decimal
	MaxAllowedQty        decimal.Decimal
	DeviationPercent     decimal.Decimal
	ViolationType        string // "UNDER", "OVER", ""
	ValidationMessage    string
}

// validateDeliveryTolerance checks if received quantity is within configured tolerance
func (s *GoodsReceiptService) validateDeliveryTolerance(ctx context.Context, tenantID, companyID, productID string, orderedQty, receivedQty decimal.Decimal) (*ToleranceValidationResult, error) {
	result := &ToleranceValidationResult{
		IsValid:       true,
		ResolvedFrom:  "DEFAULT",
		UnlimitedOver: false,
	}

	// Default tolerance: 0% (no tolerance)
	defaultTolerance := decimal.Zero
	result.EffectiveTolerance = &defaultTolerance
	result.OverTolerance = &defaultTolerance

	// Try to get effective tolerance from service
	if s.toleranceService != nil {
		effectiveTol, err := s.toleranceService.GetEffectiveTolerance(ctx, tenantID, companyID, productID)
		if err != nil {
			// Log warning but continue with default tolerance
			log.Printf("Warning: Failed to get effective tolerance for product %s: %v", productID, err)
		} else if effectiveTol != nil {
			underTol, _ := decimal.NewFromString(effectiveTol.UnderDeliveryTolerance)
			overTol, _ := decimal.NewFromString(effectiveTol.OverDeliveryTolerance)
			result.EffectiveTolerance = &underTol
			result.OverTolerance = &overTol
			result.UnlimitedOver = effectiveTol.UnlimitedOverDelivery
			result.ResolvedFrom = effectiveTol.ResolvedFrom
		}
	}

	// Calculate allowed quantity range
	hundred := decimal.NewFromInt(100)

	// Min allowed = orderedQty * (1 - underTolerance/100)
	underFactor := hundred.Sub(*result.EffectiveTolerance).Div(hundred)
	result.MinAllowedQty = orderedQty.Mul(underFactor)

	// Max allowed = orderedQty * (1 + overTolerance/100) unless unlimited
	if result.UnlimitedOver {
		// Set a very large number for unlimited over delivery
		result.MaxAllowedQty = decimal.NewFromFloat(math.MaxFloat64)
	} else {
		overFactor := hundred.Add(*result.OverTolerance).Div(hundred)
		result.MaxAllowedQty = orderedQty.Mul(overFactor)
	}

	// Calculate deviation percentage
	if orderedQty.IsPositive() {
		deviation := receivedQty.Sub(orderedQty)
		result.DeviationPercent = deviation.Div(orderedQty).Mul(hundred)
	}

	// Check if received quantity is within tolerance
	if receivedQty.LessThan(result.MinAllowedQty) {
		result.IsValid = false
		result.ViolationType = "UNDER"
		result.ValidationMessage = fmt.Sprintf(
			"Jumlah diterima (%s) kurang dari batas toleransi bawah (%s). Toleransi under-delivery: %s%%, dari: %s",
			receivedQty.String(),
			result.MinAllowedQty.Round(2).String(),
			result.EffectiveTolerance.String(),
			result.ResolvedFrom,
		)
	} else if !result.UnlimitedOver && receivedQty.GreaterThan(result.MaxAllowedQty) {
		result.IsValid = false
		result.ViolationType = "OVER"
		result.ValidationMessage = fmt.Sprintf(
			"Jumlah diterima (%s) melebihi batas toleransi atas (%s). Toleransi over-delivery: %s%%, dari: %s",
			receivedQty.String(),
			result.MaxAllowedQty.Round(2).String(),
			result.OverTolerance.String(),
			result.ResolvedFrom,
		)
	}

	return result, nil
}

// ============================================================================
// CREATE GOODS RECEIPT
// ============================================================================

// CreateGoodsReceipt creates a new goods receipt from a purchase order
func (s *GoodsReceiptService) CreateGoodsReceipt(ctx context.Context, tenantID, companyID, userID string, req *dto.CreateGoodsReceiptRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	// Parse GRN date
	grnDate, err := time.Parse("2006-01-02", req.GRNDate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid grnDate format, expected YYYY-MM-DD")
	}

	// Validate purchase order exists and is CONFIRMED
	var purchaseOrder models.PurchaseOrder
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", req.PurchaseOrderID, companyID).
		First(&purchaseOrder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewBadRequestError("purchase order not found")
		}
		return nil, fmt.Errorf("failed to validate purchase order: %w", err)
	}

	if purchaseOrder.Status != models.PurchaseOrderStatusConfirmed {
		return nil, pkgerrors.NewBadRequestError("purchase order must be in CONFIRMED status to create goods receipt")
	}

	// Load purchase order items for validation (with Product for audit log)
	var poItems []models.PurchaseOrderItem
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		Where("purchase_order_id = ?", purchaseOrder.ID).
		Find(&poItems).Error; err != nil {
		return nil, fmt.Errorf("failed to load purchase order items: %w", err)
	}

	// Create map of PO items for quick lookup
	poItemMap := make(map[string]models.PurchaseOrderItem)
	for _, item := range poItems {
		poItemMap[item.ID] = item
	}

	// Generate GRN number
	grnNumber, err := s.generateGRNNumber(ctx, tenantID, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GRN number: %w", err)
	}

	// Create goods receipt in transaction
	var goodsReceipt *models.GoodsReceipt

	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Create goods receipt header
		goodsReceipt = &models.GoodsReceipt{
			ID:               uuid.New().String(),
			TenantID:         tenantID,
			CompanyID:        companyID,
			GRNNumber:        grnNumber,
			GRNDate:          grnDate,
			PurchaseOrderID:  purchaseOrder.ID,
			WarehouseID:      purchaseOrder.WarehouseID,
			SupplierID:       purchaseOrder.SupplierID,
			Status:           models.GoodsReceiptStatusPending,
			SupplierInvoice:  req.SupplierInvoice,
			SupplierDONumber: req.SupplierDONumber,
			Notes:            req.Notes,
			ItemCount:        len(req.Items),
		}

		if err := tx.Create(goodsReceipt).Error; err != nil {
			return fmt.Errorf("failed to create goods receipt: %w", err)
		}

		// Create goods receipt items
		for _, itemReq := range req.Items {
			// Validate PO item exists
			poItem, exists := poItemMap[itemReq.PurchaseOrderItemID]
			if !exists {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("purchase order item not found: %s", itemReq.PurchaseOrderItemID))
			}

			// Validate product matches
			if poItem.ProductID != itemReq.ProductID {
				return pkgerrors.NewBadRequestError("product ID does not match purchase order item")
			}

			// Load product to check tracking flags
			var product models.Product
			if err := tx.Where("id = ?", itemReq.ProductID).First(&product).Error; err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("product not found: %s", itemReq.ProductID))
			}

			// Validate batch number for batch-tracked products
			if product.IsBatchTracked {
				if itemReq.BatchNumber == nil || *itemReq.BatchNumber == "" {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("batch number is required for batch-tracked product: %s", product.Name))
				}
			}

			// Validate expiry date for perishable products
			if product.IsPerishable {
				if itemReq.ExpiryDate == nil || *itemReq.ExpiryDate == "" {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("expiry date is required for perishable product: %s", product.Name))
				}
			}

			// Parse quantities
			receivedQty, err := decimal.NewFromString(itemReq.ReceivedQty)
			if err != nil || receivedQty.LessThan(decimal.Zero) {
				return pkgerrors.NewBadRequestError("invalid receivedQty")
			}

			// Validate: cannot receive more than remaining qty (defense-in-depth)
			// remainingQty = orderedQty - alreadyReceivedQty
			remainingQty := poItem.Quantity.Sub(poItem.ReceivedQty)
			if receivedQty.GreaterThan(remainingQty) {
				return pkgerrors.NewBadRequestError(fmt.Sprintf(
					"received qty (%s) exceeds remaining qty (%s) for PO item",
					receivedQty.String(), remainingQty.String(),
				))
			}

			// Validate delivery tolerance (SAP Model)
			// Check if received quantity is within configured tolerance for this product
			toleranceResult, toleranceErr := s.validateDeliveryTolerance(ctx, tenantID, companyID, poItem.ProductID, remainingQty, receivedQty)
			if toleranceErr != nil {
				log.Printf("Warning: Tolerance validation error for product %s: %v", poItem.ProductID, toleranceErr)
				// Continue without tolerance validation if there's an error
			} else if toleranceResult != nil && !toleranceResult.IsValid {
				return pkgerrors.NewBadRequestError(fmt.Sprintf(
					"Toleransi pengiriman tidak valid untuk produk %s: %s",
					product.Name,
					toleranceResult.ValidationMessage,
				))
			}

			acceptedQty := receivedQty
			if itemReq.AcceptedQty != "" {
				acceptedQty, err = decimal.NewFromString(itemReq.AcceptedQty)
				if err != nil || acceptedQty.LessThan(decimal.Zero) {
					return pkgerrors.NewBadRequestError("invalid acceptedQty")
				}
			}

			rejectedQty := decimal.Zero
			if itemReq.RejectedQty != "" {
				rejectedQty, err = decimal.NewFromString(itemReq.RejectedQty)
				if err != nil || rejectedQty.LessThan(decimal.Zero) {
					return pkgerrors.NewBadRequestError("invalid rejectedQty")
				}
			}

			// Parse dates
			var manufactureDate, expiryDate *time.Time
			if itemReq.ManufactureDate != nil && *itemReq.ManufactureDate != "" {
				parsed, err := time.Parse("2006-01-02", *itemReq.ManufactureDate)
				if err != nil {
					return pkgerrors.NewBadRequestError("invalid manufactureDate format")
				}
				manufactureDate = &parsed
			}
			if itemReq.ExpiryDate != nil && *itemReq.ExpiryDate != "" {
				parsed, err := time.Parse("2006-01-02", *itemReq.ExpiryDate)
				if err != nil {
					return pkgerrors.NewBadRequestError("invalid expiryDate format")
				}
				expiryDate = &parsed
			}

			item := &models.GoodsReceiptItem{
				ID:                  uuid.New().String(),
				GoodsReceiptID:      goodsReceipt.ID,
				PurchaseOrderItemID: poItem.ID,
				ProductID:           poItem.ProductID,
				ProductUnitID:       poItem.ProductUnitID,
				BatchNumber:         itemReq.BatchNumber,
				ManufactureDate:     manufactureDate,
				ExpiryDate:          expiryDate,
				OrderedQty:          poItem.Quantity,
				ReceivedQty:         receivedQty,
				AcceptedQty:         acceptedQty,
				RejectedQty:         rejectedQty,
				RejectionReason:     itemReq.RejectionReason,
				QualityNote:         itemReq.QualityNote,
				Notes:               itemReq.Notes,
			}

			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("failed to create goods receipt item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Audit logging - log goods receipt creation
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		// Build items detail for audit log (struct to preserve JSON field order)
		type auditItemDetail struct {
			ProductID       string  `json:"product_id"`
			ProductCode     string  `json:"product_code"`
			ProductName     string  `json:"product_name"`
			ProductUnitID   *string `json:"product_unit_id,omitempty"`
			BatchNumber     *string `json:"batch_number,omitempty"`
			ManufactureDate *string `json:"manufacture_date,omitempty"`
			ExpiryDate      *string `json:"expiry_date,omitempty"`
			ReceivedQty     string  `json:"received_qty"`
			AcceptedQty     *string `json:"accepted_qty,omitempty"`
			RejectedQty     *string `json:"rejected_qty,omitempty"`
			Notes           *string `json:"notes,omitempty"`
		}

		itemsDetail := make([]auditItemDetail, len(req.Items))
		for i, item := range req.Items {
			// Get product info from PO item map
			productCode := ""
			productName := ""
			if poItem, exists := poItemMap[item.PurchaseOrderItemID]; exists {
				productCode = poItem.Product.Code
				productName = poItem.Product.Name
			}

			itemsDetail[i] = auditItemDetail{
				ProductID:       item.ProductID,
				ProductCode:     productCode,
				ProductName:     productName,
				ProductUnitID:   item.ProductUnitID,
				BatchNumber:     item.BatchNumber,
				ManufactureDate: item.ManufactureDate,
				ExpiryDate:      item.ExpiryDate,
				ReceivedQty:     item.ReceivedQty,
				Notes:           item.Notes,
			}
			if item.AcceptedQty != "" {
				itemsDetail[i].AcceptedQty = &item.AcceptedQty
			}
			if item.RejectedQty != "" {
				itemsDetail[i].RejectedQty = &item.RejectedQty
			}
		}

		// Struct to preserve JSON field order
		type auditNewValues struct {
			GRNNumber        string            `json:"grn_number"`
			GRNDate          string            `json:"grn_date"`
			PurchaseOrderID  string            `json:"purchase_order_id"`
			SupplierID       string            `json:"supplier_id"`
			WarehouseID      string            `json:"warehouse_id"`
			SupplierInvoice  *string           `json:"supplier_invoice,omitempty"`
			SupplierDONumber *string           `json:"supplier_do_number,omitempty"`
			Notes            *string           `json:"notes,omitempty"`
			Status           string            `json:"status"`
			ItemCount        int               `json:"item_count"`
			Items            []auditItemDetail `json:"items"`
		}

		newValues := auditNewValues{
			GRNNumber:        goodsReceipt.GRNNumber,
			GRNDate:          goodsReceipt.GRNDate.Format("2006-01-02"),
			PurchaseOrderID:  goodsReceipt.PurchaseOrderID,
			SupplierID:       goodsReceipt.SupplierID,
			WarehouseID:      goodsReceipt.WarehouseID,
			SupplierInvoice:  goodsReceipt.SupplierInvoice,
			SupplierDONumber: goodsReceipt.SupplierDONumber,
			Notes:            goodsReceipt.Notes,
			Status:           string(goodsReceipt.Status),
			ItemCount:        len(req.Items),
			Items:            itemsDetail,
		}

		if err := s.auditService.LogGoodsReceiptCreated(ctx, auditCtx, goodsReceipt.ID, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for goods receipt created: %v", err)
		}
	}

	// Reload with relations
	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceipt.ID)
}

// generateGRNNumber generates a unique GRN number for the company
func (s *GoodsReceiptService) generateGRNNumber(ctx context.Context, tenantID, companyID string) (string, error) {
	year := time.Now().Year()
	month := time.Now().Month()

	// Get the count of GRNs this month for sequence number
	var count int64
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.GoodsReceipt{}).
		Where("company_id = ? AND EXTRACT(YEAR FROM created_at) = ? AND EXTRACT(MONTH FROM created_at) = ?",
			companyID, year, int(month)).
		Count(&count).Error; err != nil {
		return "", err
	}

	// Format: GRN-YYYYMM-XXXX
	grnNumber := fmt.Sprintf("GRN-%d%02d-%04d", year, month, count+1)
	return grnNumber, nil
}

// ============================================================================
// LIST GOODS RECEIPTS
// ============================================================================

// ListGoodsReceipts retrieves goods receipts with filtering and pagination
func (s *GoodsReceiptService) ListGoodsReceipts(ctx context.Context, tenantID, companyID string, query *dto.GoodsReceiptListQuery) (*dto.GoodsReceiptListResponse, error) {
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
			"grnNumber": "grn_number",
			"grnDate":   "grn_date",
			"status":    "status",
			"createdAt": "created_at",
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
	baseQuery := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.GoodsReceipt{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		// Search by GRN number or PO number (via subquery)
		baseQuery = baseQuery.Where(
			"grn_number LIKE ? OR purchase_order_id IN (SELECT id FROM purchase_orders WHERE po_number LIKE ? AND company_id = ?)",
			searchPattern, searchPattern, companyID,
		)
	}

	if query.Status != nil {
		baseQuery = baseQuery.Where("status = ?", *query.Status)
	}

	if query.PurchaseOrderID != nil {
		baseQuery = baseQuery.Where("purchase_order_id = ?", *query.PurchaseOrderID)
	}

	if query.SupplierID != nil {
		baseQuery = baseQuery.Where("supplier_id = ?", *query.SupplierID)
	}

	if query.WarehouseID != nil {
		baseQuery = baseQuery.Where("warehouse_id = ?", *query.WarehouseID)
	}

	if query.DateFrom != nil {
		baseQuery = baseQuery.Where("grn_date >= ?", *query.DateFrom)
	}

	if query.DateTo != nil {
		baseQuery = baseQuery.Where("grn_date <= ?", *query.DateTo)
	}

	// Count total records
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count goods receipts: %w", err)
	}

	// Apply sorting and pagination
	var goodsReceipts []models.GoodsReceipt
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	offset := (page - 1) * pageSize

	if err := baseQuery.
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&goodsReceipts).Error; err != nil {
		return nil, fmt.Errorf("failed to list goods receipts: %w", err)
	}

	// Load relations for each goods receipt
	for i := range goodsReceipts {
		if err := s.loadGoodsReceiptRelations(ctx, tenantID, &goodsReceipts[i]); err != nil {
			return nil, err
		}
	}

	// Map to response DTOs
	responses := make([]dto.GoodsReceiptResponse, len(goodsReceipts))
	for i, gr := range goodsReceipts {
		responses[i] = s.MapToResponse(ctx, &gr, false)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &dto.GoodsReceiptListResponse{
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
// GET GOODS RECEIPT BY ID
// ============================================================================

// GetGoodsReceiptByID retrieves a goods receipt by ID with all relations
func (s *GoodsReceiptService) GetGoodsReceiptByID(ctx context.Context, tenantID, companyID, goodsReceiptID string) (*models.GoodsReceipt, error) {
	var goodsReceipt models.GoodsReceipt
	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ? AND id = ?", companyID, goodsReceiptID).
		First(&goodsReceipt).Error

	if err == gorm.ErrRecordNotFound {
		return nil, pkgerrors.NewNotFoundError("goods receipt not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get goods receipt: %w", err)
	}

	// Load relations
	if err := s.loadGoodsReceiptRelations(ctx, tenantID, &goodsReceipt); err != nil {
		return nil, err
	}

	return &goodsReceipt, nil
}

// loadGoodsReceiptRelations loads all relations for a goods receipt
func (s *GoodsReceiptService) loadGoodsReceiptRelations(ctx context.Context, tenantID string, gr *models.GoodsReceipt) error {
	// Load purchase order
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("id = ?", gr.PurchaseOrderID).First(&gr.PurchaseOrder).Error; err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to load purchase order: %w", err)
	}

	// Load supplier
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("id = ?", gr.SupplierID).First(&gr.Supplier).Error; err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to load supplier: %w", err)
	}

	// Load warehouse
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Where("id = ?", gr.WarehouseID).First(&gr.Warehouse).Error; err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to load warehouse: %w", err)
	}

	// Load items with product info and disposition resolver
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		Preload("ProductUnit").
		Preload("DispositionResolver").
		Where("goods_receipt_id = ?", gr.ID).
		Find(&gr.Items).Error; err != nil {
		return fmt.Errorf("failed to load items: %w", err)
	}

	return nil
}

// ============================================================================
// UPDATE GOODS RECEIPT
// ============================================================================

// UpdateGoodsReceipt updates an existing goods receipt (only PENDING status)
func (s *GoodsReceiptService) UpdateGoodsReceipt(ctx context.Context, tenantID, companyID, goodsReceiptID string, req *dto.UpdateGoodsReceiptRequest) (*models.GoodsReceipt, error) {
	// Get existing goods receipt
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	// Check status - can only update PENDING
	if goodsReceipt.Status != models.GoodsReceiptStatusPending {
		return nil, pkgerrors.NewBadRequestError("can only update goods receipts in PENDING status")
	}

	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Update header fields
		if req.GRNDate != nil {
			grnDate, err := time.Parse("2006-01-02", *req.GRNDate)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid grnDate format")
			}
			goodsReceipt.GRNDate = grnDate
		}

		if req.SupplierInvoice != nil {
			goodsReceipt.SupplierInvoice = req.SupplierInvoice
		}

		if req.SupplierDONumber != nil {
			goodsReceipt.SupplierDONumber = req.SupplierDONumber
		}

		if req.Notes != nil {
			goodsReceipt.Notes = req.Notes
		}

		if err := tx.Model(goodsReceipt).
			Select("grn_date", "supplier_invoice", "supplier_do_number", "notes", "updated_at").
			Updates(goodsReceipt).Error; err != nil {
			return fmt.Errorf("failed to update goods receipt: %w", err)
		}

		// Update items if provided
		if req.Items != nil {
			for _, itemReq := range req.Items {
				if itemReq.ID != nil {
					// Update existing item
					var item models.GoodsReceiptItem
					if err := tx.Where("id = ? AND goods_receipt_id = ?", *itemReq.ID, goodsReceiptID).First(&item).Error; err != nil {
						return pkgerrors.NewBadRequestError(fmt.Sprintf("item not found: %s", *itemReq.ID))
					}

					receivedQty, err := decimal.NewFromString(itemReq.ReceivedQty)
					if err != nil {
						return pkgerrors.NewBadRequestError("invalid receivedQty")
					}
					item.ReceivedQty = receivedQty

					if itemReq.AcceptedQty != "" {
						acceptedQty, err := decimal.NewFromString(itemReq.AcceptedQty)
						if err != nil {
							return pkgerrors.NewBadRequestError("invalid acceptedQty")
						}
						item.AcceptedQty = acceptedQty
					}

					if itemReq.RejectedQty != "" {
						rejectedQty, err := decimal.NewFromString(itemReq.RejectedQty)
						if err != nil {
							return pkgerrors.NewBadRequestError("invalid rejectedQty")
						}
						item.RejectedQty = rejectedQty
					}

					if itemReq.BatchNumber != nil {
						item.BatchNumber = itemReq.BatchNumber
					}
					if itemReq.RejectionReason != nil {
						item.RejectionReason = itemReq.RejectionReason
					}
					if itemReq.QualityNote != nil {
						item.QualityNote = itemReq.QualityNote
					}
					if itemReq.Notes != nil {
						item.Notes = itemReq.Notes
					}

					if err := tx.Save(&item).Error; err != nil {
						return fmt.Errorf("failed to update item: %w", err)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// ============================================================================
// DELETE GOODS RECEIPT
// ============================================================================

// DeleteGoodsReceipt deletes a goods receipt (only PENDING status)
func (s *GoodsReceiptService) DeleteGoodsReceipt(ctx context.Context, tenantID, companyID, goodsReceiptID string) error {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return err
	}

	// Check status - can only delete PENDING
	if goodsReceipt.Status != models.GoodsReceiptStatusPending {
		return pkgerrors.NewBadRequestError("can only delete goods receipts in PENDING status")
	}

	return s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Delete items first
		if err := tx.Where("goods_receipt_id = ?", goodsReceiptID).Delete(&models.GoodsReceiptItem{}).Error; err != nil {
			return fmt.Errorf("failed to delete goods receipt items: %w", err)
		}

		// Delete goods receipt
		if err := tx.Delete(goodsReceipt).Error; err != nil {
			return fmt.Errorf("failed to delete goods receipt: %w", err)
		}

		return nil
	})
}

// ============================================================================
// STATUS TRANSITIONS
// ============================================================================

// ReceiveGoods marks goods as received (PENDING → RECEIVED)
func (s *GoodsReceiptService) ReceiveGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.ReceiveGoodsRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusPending {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in PENDING status")
	}

	// Capture old status for audit
	oldStatus := string(goodsReceipt.Status)

	now := time.Now()
	goodsReceipt.Status = models.GoodsReceiptStatusReceived
	goodsReceipt.ReceivedBy = &userID
	goodsReceipt.ReceivedAt = &now
	if req != nil && req.Notes != nil {
		goodsReceipt.ReceiveNotes = req.Notes
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Model(goodsReceipt).
		Select("status", "received_by", "received_at", "receive_notes", "updated_at").
		Updates(goodsReceipt).Error; err != nil {
		return nil, fmt.Errorf("failed to update goods receipt status: %w", err)
	}

	// Audit logging - only log status change
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]any{
			"status": oldStatus,
		}

		newValues := map[string]any{
			"grn_number":    goodsReceipt.GRNNumber,
			"status":        string(goodsReceipt.Status),
			"receive_notes": goodsReceipt.ReceiveNotes,
		}

		if err := s.auditService.LogGoodsReceiptReceived(ctx, auditCtx, goodsReceipt.ID, oldValues, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for goods receipt received: %v", err)
		}
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// InspectGoods marks goods as inspected (RECEIVED → INSPECTED)
func (s *GoodsReceiptService) InspectGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.InspectGoodsRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusReceived {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in RECEIVED status")
	}

	// Capture old status for audit
	oldStatus := string(goodsReceipt.Status)

	// Capture old item states for audit
	type oldItemState struct {
		ID          string `json:"id"`
		ProductID   string `json:"product_id"`
		ReceivedQty string `json:"received_qty"`
		AcceptedQty string `json:"accepted_qty"`
		RejectedQty string `json:"rejected_qty"`
	}
	oldItemStates := make([]oldItemState, len(goodsReceipt.Items))
	for i, item := range goodsReceipt.Items {
		oldItemStates[i] = oldItemState{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ReceivedQty: item.ReceivedQty.String(),
			AcceptedQty: item.AcceptedQty.String(),
			RejectedQty: item.RejectedQty.String(),
		}
	}

	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Update items if provided
		if req != nil && req.Items != nil {
			for _, itemReq := range req.Items {
				if itemReq.ID != nil {
					var item models.GoodsReceiptItem
					if err := tx.Where("id = ? AND goods_receipt_id = ?", *itemReq.ID, goodsReceiptID).First(&item).Error; err != nil {
						continue
					}

					if itemReq.AcceptedQty != "" {
						acceptedQty, _ := decimal.NewFromString(itemReq.AcceptedQty)
						item.AcceptedQty = acceptedQty
					}
					if itemReq.RejectedQty != "" {
						rejectedQty, _ := decimal.NewFromString(itemReq.RejectedQty)
						item.RejectedQty = rejectedQty
					}
					if itemReq.RejectionReason != nil {
						item.RejectionReason = itemReq.RejectionReason
					}
					if itemReq.QualityNote != nil {
						item.QualityNote = itemReq.QualityNote
					}

					if err := tx.Save(&item).Error; err != nil {
						return err
					}
				}
			}
		}

		// Update status
		now := time.Now()
		goodsReceipt.Status = models.GoodsReceiptStatusInspected
		goodsReceipt.InspectedBy = &userID
		goodsReceipt.InspectedAt = &now
		if req != nil && req.Notes != nil {
			goodsReceipt.InspectionNotes = req.Notes
		}

		return tx.Model(goodsReceipt).
			Select("status", "inspected_by", "inspected_at", "inspection_notes", "updated_at").
			Updates(goodsReceipt).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to inspect goods: %w", err)
	}

	// Reload goods receipt to get updated items
	updatedGoodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	// Audit logging - include items detail
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		// Build new item states for audit log
		type newItemState struct {
			ID              string  `json:"id"`
			ProductID       string  `json:"product_id"`
			ProductName     string  `json:"product_name,omitempty"`
			ReceivedQty     string  `json:"received_qty"`
			AcceptedQty     string  `json:"accepted_qty"`
			RejectedQty     string  `json:"rejected_qty"`
			RejectionReason *string `json:"rejection_reason,omitempty"`
			QualityNote     *string `json:"quality_note,omitempty"`
		}
		newItemStates := make([]newItemState, len(updatedGoodsReceipt.Items))
		for i, item := range updatedGoodsReceipt.Items {
			newItemStates[i] = newItemState{
				ID:              item.ID,
				ProductID:       item.ProductID,
				ProductName:     item.Product.Name,
				ReceivedQty:     item.ReceivedQty.String(),
				AcceptedQty:     item.AcceptedQty.String(),
				RejectedQty:     item.RejectedQty.String(),
				RejectionReason: item.RejectionReason,
				QualityNote:     item.QualityNote,
			}
		}

		oldValues := map[string]any{
			"status": oldStatus,
			"items":  oldItemStates,
		}

		newValues := map[string]any{
			"grn_number":       updatedGoodsReceipt.GRNNumber,
			"status":           string(updatedGoodsReceipt.Status),
			"inspection_notes": updatedGoodsReceipt.InspectionNotes,
			"item_count":       len(updatedGoodsReceipt.Items),
			"items":            newItemStates,
		}

		if err := s.auditService.LogGoodsReceiptInspected(ctx, auditCtx, updatedGoodsReceipt.ID, oldValues, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for goods receipt inspected: %v", err)
		}
	}

	return updatedGoodsReceipt, nil
}

// AcceptGoods accepts goods and updates stock (INSPECTED → ACCEPTED)
func (s *GoodsReceiptService) AcceptGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.AcceptGoodsRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusInspected {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in INSPECTED status")
	}

	// Capture old status for audit
	oldStatus := string(goodsReceipt.Status)

	// Track if PO was auto-completed for audit logging
	var poCompleted bool
	var poNumber string

	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Update warehouse stock for each accepted item
		for _, item := range goodsReceipt.Items {
			if item.AcceptedQty.GreaterThan(decimal.Zero) {
				// Load product to check tracking flags
				var product models.Product
				if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
					return fmt.Errorf("failed to load product: %w", err)
				}

				var stock models.WarehouseStock
				err := tx.Where("warehouse_id = ? AND product_id = ?", goodsReceipt.WarehouseID, item.ProductID).
					First(&stock).Error

				if err == gorm.ErrRecordNotFound {
					// Create new stock record
					stock = models.WarehouseStock{
						ID:          uuid.New().String(),
						WarehouseID: goodsReceipt.WarehouseID,
						ProductID:   item.ProductID,
						Quantity:    item.AcceptedQty,
					}
					if err := tx.Create(&stock).Error; err != nil {
						return fmt.Errorf("failed to create stock: %w", err)
					}
				} else if err != nil {
					return fmt.Errorf("failed to get stock: %w", err)
				} else {
					// Update existing stock
					stock.Quantity = stock.Quantity.Add(item.AcceptedQty)
					if err := tx.Save(&stock).Error; err != nil {
						return fmt.Errorf("failed to update stock: %w", err)
					}
				}

				// Create or update ProductBatch record for batch-tracked products
				if product.IsBatchTracked && item.BatchNumber != nil && *item.BatchNumber != "" {
					// Check if batch already exists for this product and batch number
					var existingBatch models.ProductBatch
					err := tx.Where("product_id = ? AND batch_number = ?", item.ProductID, *item.BatchNumber).
						First(&existingBatch).Error

					if err == gorm.ErrRecordNotFound {
						// Create new batch
						qualityStatus := "GOOD"
						batch := models.ProductBatch{
							ID:               uuid.New().String(),
							BatchNumber:      *item.BatchNumber,
							ProductID:        item.ProductID,
							WarehouseStockID: stock.ID,
							ManufactureDate:  item.ManufactureDate,
							ExpiryDate:       item.ExpiryDate,
							Quantity:         item.AcceptedQty,
							GoodsReceiptID:   &goodsReceipt.ID,
							ReceiptDate:      time.Now(),
							Status:           "AVAILABLE",
							QualityStatus:    &qualityStatus,
						}
						if err := tx.Create(&batch).Error; err != nil {
							return fmt.Errorf("failed to create product batch: %w", err)
						}
					} else if err != nil {
						return fmt.Errorf("failed to check existing batch: %w", err)
					} else {
						// Update existing batch quantity
						existingBatch.Quantity = existingBatch.Quantity.Add(item.AcceptedQty)
						// Update expiry date if provided and newer
						if item.ExpiryDate != nil && (existingBatch.ExpiryDate == nil || item.ExpiryDate.After(*existingBatch.ExpiryDate)) {
							existingBatch.ExpiryDate = item.ExpiryDate
						}
						if err := tx.Save(&existingBatch).Error; err != nil {
							return fmt.Errorf("failed to update product batch quantity: %w", err)
						}
					}
				}

				// Update received quantity on purchase order item
				if err := tx.Model(&models.PurchaseOrderItem{}).
					Where("id = ?", item.PurchaseOrderItemID).
					Update("received_qty", gorm.Expr("received_qty + ?", item.AcceptedQty)).Error; err != nil {
					return fmt.Errorf("failed to update PO item received qty: %w", err)
				}
			}
		}

		// Check if all PO items are fully received and auto-complete the PO
		var po models.PurchaseOrder
		if err := tx.Where("id = ?", goodsReceipt.PurchaseOrderID).First(&po).Error; err != nil {
			return fmt.Errorf("failed to load purchase order: %w", err)
		}

		if po.Status == models.PurchaseOrderStatusConfirmed {
			// Load all PO items to check if fully received
			var poItems []models.PurchaseOrderItem
			if err := tx.Where("purchase_order_id = ?", po.ID).Find(&poItems).Error; err != nil {
				return fmt.Errorf("failed to load PO items: %w", err)
			}

			// Check if all items are fully received (received_qty >= quantity)
			allFullyReceived := true
			for _, poItem := range poItems {
				if poItem.ReceivedQty.LessThan(poItem.Quantity) {
					allFullyReceived = false
					break
				}
			}

			// If all items are fully received, update PO status to COMPLETED
			if allFullyReceived {
				if err := tx.Model(&models.PurchaseOrder{}).
					Where("id = ?", po.ID).
					Update("status", models.PurchaseOrderStatusCompleted).Error; err != nil {
					return fmt.Errorf("failed to update PO status to COMPLETED: %w", err)
				}
				poCompleted = true
				poNumber = po.PONumber
			}
		}

		// Determine final status based on rejections
		hasRejections := false
		for _, item := range goodsReceipt.Items {
			if item.RejectedQty.GreaterThan(decimal.Zero) {
				hasRejections = true
				break
			}
		}

		if hasRejections {
			goodsReceipt.Status = models.GoodsReceiptStatusPartial
		} else {
			goodsReceipt.Status = models.GoodsReceiptStatusAccepted
		}

		if req != nil && req.Notes != nil {
			goodsReceipt.AcceptanceNotes = req.Notes
		}

		return tx.Model(goodsReceipt).
			Select("status", "acceptance_notes", "updated_at").
			Updates(goodsReceipt).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to accept goods: %w", err)
	}

	// Audit logging - only log status change
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]any{
			"status": oldStatus,
		}

		newValues := map[string]any{
			"grn_number":       goodsReceipt.GRNNumber,
			"status":           string(goodsReceipt.Status),
			"acceptance_notes": goodsReceipt.AcceptanceNotes,
		}

		if err := s.auditService.LogGoodsReceiptAccepted(ctx, auditCtx, goodsReceipt.ID, oldValues, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for goods receipt accepted: %v", err)
		}

		// Log PO completion if it was auto-completed
		if poCompleted {
			poOldValues := map[string]any{
				"status": string(models.PurchaseOrderStatusConfirmed),
			}
			poNewValues := map[string]any{
				"po_number": poNumber,
				"status":    string(models.PurchaseOrderStatusCompleted),
			}
			if err := s.auditService.LogPurchaseOrderCompleted(ctx, auditCtx, goodsReceipt.PurchaseOrderID, poOldValues, poNewValues); err != nil {
				log.Printf("WARNING: Failed to create audit log for PO auto-completed: %v", err)
			}
		}
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// RejectGoods rejects goods (INSPECTED → REJECTED)
func (s *GoodsReceiptService) RejectGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.RejectGoodsRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusInspected {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in INSPECTED status")
	}

	// Capture old status for audit
	oldStatus := string(goodsReceipt.Status)

	// Update all items as rejected
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		for _, item := range goodsReceipt.Items {
			item.RejectedQty = item.ReceivedQty
			item.AcceptedQty = decimal.Zero
			item.RejectionReason = &req.RejectionReason
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}

		goodsReceipt.Status = models.GoodsReceiptStatusRejected
		notes := req.RejectionReason
		goodsReceipt.RejectionNotes = &notes

		return tx.Model(goodsReceipt).
			Select("status", "rejection_notes", "updated_at").
			Updates(goodsReceipt).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to reject goods: %w", err)
	}

	// Audit logging - only log status change
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]any{
			"status": oldStatus,
		}

		newValues := map[string]any{
			"grn_number":       goodsReceipt.GRNNumber,
			"status":           string(goodsReceipt.Status),
			"rejection_notes":  goodsReceipt.RejectionNotes,
			"rejection_reason": req.RejectionReason,
		}

		if err := s.auditService.LogGoodsReceiptRejected(ctx, auditCtx, goodsReceipt.ID, oldValues, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for goods receipt rejected: %v", err)
		}
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// ============================================================================
// REJECTION DISPOSITION MANAGEMENT (Odoo+M3 Model)
// ============================================================================

// UpdateRejectionDisposition updates the rejection disposition for a goods receipt item
// This tracks what will happen to rejected items: PENDING_REPLACEMENT, CREDIT_REQUESTED, RETURNED, WRITTEN_OFF
func (s *GoodsReceiptService) UpdateRejectionDisposition(ctx context.Context, tenantID, companyID, goodsReceiptID, itemID, userID string, req *dto.UpdateRejectionDispositionRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	// Get goods receipt
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	// Find the item
	var targetItem *models.GoodsReceiptItem
	for i := range goodsReceipt.Items {
		if goodsReceipt.Items[i].ID == itemID {
			targetItem = &goodsReceipt.Items[i]
			break
		}
	}

	if targetItem == nil {
		return nil, pkgerrors.NewNotFoundError("goods receipt item not found")
	}

	// Validate item has rejected quantity
	if targetItem.RejectedQty.IsZero() {
		return nil, pkgerrors.NewBadRequestError("can only set disposition for items with rejected quantity")
	}

	// Parse disposition
	disposition := models.RejectionDisposition(req.RejectionDisposition)

	// Update the item
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		targetItem.RejectionDisposition = &disposition
		targetItem.DispositionNotes = req.DispositionNotes

		return tx.Model(targetItem).
			Select("rejection_disposition", "disposition_notes", "updated_at").
			Updates(targetItem).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update rejection disposition: %w", err)
	}

	// Audit logging
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		newValues := map[string]any{
			"grn_number":            goodsReceipt.GRNNumber,
			"item_id":               itemID,
			"product_id":            targetItem.ProductID,
			"rejection_disposition": string(disposition),
			"disposition_notes":     req.DispositionNotes,
		}

		if err := s.auditService.LogGoodsReceiptDispositionUpdated(ctx, auditCtx, goodsReceiptID, nil, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for rejection disposition update: %v", err)
		}
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// ResolveDisposition marks a rejection disposition as resolved
// This is called when the supplier has sent replacement, credit note has been received, goods returned, or written off
func (s *GoodsReceiptService) ResolveDisposition(ctx context.Context, tenantID, companyID, goodsReceiptID, itemID, userID string, req *dto.ResolveDispositionRequest, ipAddress, userAgent string) (*models.GoodsReceipt, error) {
	// Get goods receipt
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	// Find the item
	var targetItem *models.GoodsReceiptItem
	for i := range goodsReceipt.Items {
		if goodsReceipt.Items[i].ID == itemID {
			targetItem = &goodsReceipt.Items[i]
			break
		}
	}

	if targetItem == nil {
		return nil, pkgerrors.NewNotFoundError("goods receipt item not found")
	}

	// Validate item has disposition set
	if targetItem.RejectionDisposition == nil {
		return nil, pkgerrors.NewBadRequestError("no rejection disposition set for this item")
	}

	// Validate disposition is not already resolved
	if targetItem.DispositionResolvedAt != nil {
		return nil, pkgerrors.NewBadRequestError("disposition has already been resolved")
	}

	// Update the item
	now := time.Now()
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		targetItem.DispositionResolvedAt = &now
		targetItem.DispositionResolvedBy = &userID
		if req.DispositionResolvedNotes != nil {
			targetItem.DispositionResolvedNotes = req.DispositionResolvedNotes
		}

		return tx.Model(targetItem).
			Select("disposition_resolved_at", "disposition_resolved_by", "disposition_resolved_notes", "updated_at").
			Updates(targetItem).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to resolve disposition: %w", err)
	}

	// Audit logging
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &userID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		newValues := map[string]any{
			"grn_number":                 goodsReceipt.GRNNumber,
			"item_id":                    itemID,
			"product_id":                 targetItem.ProductID,
			"rejection_disposition":      string(*targetItem.RejectionDisposition),
			"disposition_resolved_at":    now.Format(time.RFC3339),
			"disposition_resolved_by":    userID,
			"disposition_resolved_notes": targetItem.DispositionResolvedNotes,
		}

		if err := s.auditService.LogGoodsReceiptDispositionResolved(ctx, auditCtx, goodsReceiptID, nil, newValues); err != nil {
			log.Printf("WARNING: Failed to create audit log for disposition resolved: %v", err)
		}
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// ============================================================================
// RESPONSE MAPPING
// ============================================================================

// MapToResponse maps a GoodsReceipt model to response DTO
// Includes invoicedQty calculation for each GRN item
func (s *GoodsReceiptService) MapToResponse(ctx context.Context, gr *models.GoodsReceipt, includeItems bool) dto.GoodsReceiptResponse {
	response := dto.GoodsReceiptResponse{
		ID:               gr.ID,
		GRNNumber:        gr.GRNNumber,
		GRNDate:          gr.GRNDate,
		PurchaseOrderID:  gr.PurchaseOrderID,
		WarehouseID:      gr.WarehouseID,
		SupplierID:       gr.SupplierID,
		Status:           string(gr.Status),
		ReceivedBy:       gr.ReceivedBy,
		ReceivedAt:       gr.ReceivedAt,
		InspectedBy:      gr.InspectedBy,
		InspectedAt:      gr.InspectedAt,
		SupplierInvoice:  gr.SupplierInvoice,
		SupplierDONumber: gr.SupplierDONumber,
		Notes:            gr.Notes,
		ReceiveNotes:     gr.ReceiveNotes,
		InspectionNotes:  gr.InspectionNotes,
		AcceptanceNotes:  gr.AcceptanceNotes,
		RejectionNotes:   gr.RejectionNotes,
		ItemCount:        gr.ItemCount, // From database field
		CreatedAt:        gr.CreatedAt,
		UpdatedAt:        gr.UpdatedAt,
	}

	// Map purchase order
	if gr.PurchaseOrder.ID != "" {
		response.PurchaseOrder = &dto.GoodsReceiptPurchaseOrderResponse{
			ID:       gr.PurchaseOrder.ID,
			PONumber: gr.PurchaseOrder.PONumber,
			PODate:   gr.PurchaseOrder.PODate.Format("2006-01-02"),
		}
	}

	// Map warehouse
	if gr.Warehouse.ID != "" {
		response.Warehouse = &dto.WarehouseBasicResponse{
			ID:   gr.Warehouse.ID,
			Code: gr.Warehouse.Code,
			Name: gr.Warehouse.Name,
		}
	}

	// Map supplier
	if gr.Supplier.ID != "" {
		response.Supplier = &dto.SupplierBasicResponse{
			ID:   gr.Supplier.ID,
			Code: gr.Supplier.Code,
			Name: gr.Supplier.Name,
		}
	}

	// Map items
	if includeItems && len(gr.Items) > 0 {
		// Query invoiced quantities for all GRN items
		// Use two approaches to handle both new and old data:
		// 1. Via goods_receipt_item_id (granular, per-item tracking)
		// 2. Via product_id from invoices linked to this GRN (fallback for older data)
		invoicedQtyMap := make(map[string]float64)
		if len(gr.Items) > 0 {
			itemIDs := make([]string, len(gr.Items))
			productIDs := make([]string, len(gr.Items))
			itemProductMap := make(map[string]string) // GRN item ID -> product ID
			for i, item := range gr.Items {
				itemIDs[i] = item.ID
				productIDs[i] = item.ProductID
				itemProductMap[item.ID] = item.ProductID
			}

			// Approach 1: Query via goods_receipt_item_id (granular tracking)
			// IMPORTANT: Exclude soft-deleted invoices and REJECTED/CANCELLED invoices
			type InvoicedQtyResult struct {
				GoodsReceiptItemID string
				TotalInvoiced      float64
			}
			var results []InvoicedQtyResult
			s.db.WithContext(ctx).
				Table("purchase_invoice_items").
				Select("purchase_invoice_items.goods_receipt_item_id, COALESCE(SUM(purchase_invoice_items.quantity), 0) as total_invoiced").
				Joins("JOIN purchase_invoices ON purchase_invoice_items.purchase_invoice_id = purchase_invoices.id").
				Where("purchase_invoice_items.goods_receipt_item_id IN ?", itemIDs).
				Where("purchase_invoices.deleted_at IS NULL").
				Where("purchase_invoices.status NOT IN ?", []string{"REJECTED", "CANCELLED"}).
				Group("purchase_invoice_items.goods_receipt_item_id").
				Scan(&results)

			for _, r := range results {
				invoicedQtyMap[r.GoodsReceiptItemID] = r.TotalInvoiced
			}

			// Approach 2: Fallback - query via product_id from invoices linked to this GRN
			// This handles invoices created before item-level GRN linkage was implemented
			// IMPORTANT: Exclude soft-deleted invoices and REJECTED/CANCELLED invoices
			type ProductInvoicedResult struct {
				ProductID     string
				TotalInvoiced float64
			}
			var productResults []ProductInvoicedResult
			s.db.WithContext(ctx).
				Table("purchase_invoice_items").
				Select("purchase_invoice_items.product_id, COALESCE(SUM(purchase_invoice_items.quantity), 0) as total_invoiced").
				Joins("JOIN purchase_invoices ON purchase_invoice_items.purchase_invoice_id = purchase_invoices.id").
				Where("purchase_invoices.goods_receipt_id = ?", gr.ID).
				Where("purchase_invoices.deleted_at IS NULL").
				Where("purchase_invoices.status NOT IN ?", []string{"REJECTED", "CANCELLED"}).
				Where("purchase_invoice_items.goods_receipt_item_id IS NULL"). // Only items without GRN item linkage
				Where("purchase_invoice_items.product_id IN ?", productIDs).
				Group("purchase_invoice_items.product_id").
				Scan(&productResults)

			// Map product-level invoiced qty to GRN items (fallback for items without direct linkage)
			productInvoicedMap := make(map[string]float64)
			for _, r := range productResults {
				productInvoicedMap[r.ProductID] = r.TotalInvoiced
			}

			// Add product-level invoiced qty to items that don't have direct linkage
			for itemID, productID := range itemProductMap {
				if invoicedQtyMap[itemID] == 0 && productInvoicedMap[productID] > 0 {
					invoicedQtyMap[itemID] = productInvoicedMap[productID]
				}
			}
		}

		response.Items = make([]dto.GoodsReceiptItemResponse, len(gr.Items))
		for i, item := range gr.Items {
			// Map rejection disposition fields
			var rejectionDisposition *string
			if item.RejectionDisposition != nil {
				disposition := string(*item.RejectionDisposition)
				rejectionDisposition = &disposition
			}

			// Get invoiced qty for this item
			invoicedQty := invoicedQtyMap[item.ID]

			response.Items[i] = dto.GoodsReceiptItemResponse{
				ID:                    item.ID,
				GoodsReceiptID:        item.GoodsReceiptID,
				PurchaseOrderItemID:   item.PurchaseOrderItemID,
				ProductID:             item.ProductID,
				ProductUnitID:         item.ProductUnitID,
				BatchNumber:           item.BatchNumber,
				ManufactureDate:       item.ManufactureDate,
				ExpiryDate:            item.ExpiryDate,
				OrderedQty:            item.OrderedQty.String(),
				ReceivedQty:           item.ReceivedQty.String(),
				AcceptedQty:           item.AcceptedQty.String(),
				InvoicedQty:           fmt.Sprintf("%.4f", invoicedQty),
				RejectedQty:           item.RejectedQty.String(),
				RejectionReason:       item.RejectionReason,
				RejectionDisposition:  rejectionDisposition,
				DispositionResolved:      item.DispositionResolvedAt != nil,
				DispositionResolvedAt:    item.DispositionResolvedAt,
				DispositionResolvedBy:    item.DispositionResolvedBy,
				DispositionNotes:         item.DispositionNotes,
				DispositionResolvedNotes: item.DispositionResolvedNotes,
				QualityNote:              item.QualityNote,
				Notes:                 item.Notes,
				CreatedAt:             item.CreatedAt,
				UpdatedAt:             item.UpdatedAt,
			}

			// Map product with tracking flags
			if item.Product.ID != "" {
				response.Items[i].Product = &dto.GoodsReceiptProductResponse{
					ID:             item.Product.ID,
					Code:           item.Product.Code,
					Name:           item.Product.Name,
					BaseUnit:       item.Product.BaseUnit,
					IsBatchTracked: item.Product.IsBatchTracked,
					IsPerishable:   item.Product.IsPerishable,
				}
			}

			// Map product unit
			if item.ProductUnit != nil && item.ProductUnit.ID != "" {
				response.Items[i].ProductUnit = &dto.GoodsReceiptProductUnitResponse{
					ID:             item.ProductUnit.ID,
					UnitName:       item.ProductUnit.UnitName,
					ConversionRate: item.ProductUnit.ConversionRate.String(),
				}
			}

			// Map disposition resolver (user who resolved rejected goods disposition)
			if item.DispositionResolver != nil && item.DispositionResolver.ID != "" {
				response.Items[i].DispositionResolver = &dto.UserBasicResponse{
					ID:       item.DispositionResolver.ID,
					Email:    item.DispositionResolver.Email,
					FullName: item.DispositionResolver.FullName,
				}
			}
		}
	}

	// Calculate invoice status for this GRN (always, regardless of includeItems)
	// This is needed for the list view to show correct invoice badges
	var totalAccepted, totalInvoiced float64

	// Query total accepted qty from GRN items
	type AcceptedResult struct {
		TotalAccepted float64
	}
	var acceptedResult AcceptedResult
	s.db.WithContext(ctx).
		Table("goods_receipt_items").
		Select("COALESCE(SUM(accepted_qty), 0) as total_accepted").
		Where("goods_receipt_id = ?", gr.ID).
		Scan(&acceptedResult)
	totalAccepted = acceptedResult.TotalAccepted

	// Query total invoiced qty from purchase_invoice_items for this GRN
	// Use two approaches and take the maximum to handle both:
	// 1. Items with goods_receipt_item_id set (granular tracking)
	// 2. Items from invoices linked to this GRN via goods_receipt_id (fallback for older data)
	type InvoicedResult struct {
		TotalInvoiced float64
	}

	// Approach 1: Via goods_receipt_item_id (granular, per-item tracking)
	// IMPORTANT: Exclude soft-deleted invoices and REJECTED/CANCELLED invoices
	var invoicedByItem InvoicedResult
	s.db.WithContext(ctx).
		Table("purchase_invoice_items").
		Select("COALESCE(SUM(purchase_invoice_items.quantity), 0) as total_invoiced").
		Joins("JOIN goods_receipt_items ON purchase_invoice_items.goods_receipt_item_id = goods_receipt_items.id").
		Joins("JOIN purchase_invoices ON purchase_invoice_items.purchase_invoice_id = purchase_invoices.id").
		Where("goods_receipt_items.goods_receipt_id = ?", gr.ID).
		Where("purchase_invoices.deleted_at IS NULL").
		Where("purchase_invoices.status NOT IN ?", []string{"REJECTED", "CANCELLED"}).
		Scan(&invoicedByItem)

	// Approach 2: Via purchase_invoices.goods_receipt_id (fallback for invoices without item-level GRN linkage)
	// IMPORTANT: Exclude soft-deleted invoices and REJECTED/CANCELLED invoices
	var invoicedByInvoice InvoicedResult
	s.db.WithContext(ctx).
		Table("purchase_invoice_items").
		Select("COALESCE(SUM(purchase_invoice_items.quantity), 0) as total_invoiced").
		Joins("JOIN purchase_invoices ON purchase_invoice_items.purchase_invoice_id = purchase_invoices.id").
		Where("purchase_invoices.goods_receipt_id = ?", gr.ID).
		Where("purchase_invoices.deleted_at IS NULL").
		Where("purchase_invoices.status NOT IN ?", []string{"REJECTED", "CANCELLED"}).
		Scan(&invoicedByInvoice)

	// Take the maximum of both approaches
	// This handles cases where item-level linkage is missing but invoice-level is present
	totalInvoiced = invoicedByItem.TotalInvoiced
	if invoicedByInvoice.TotalInvoiced > totalInvoiced {
		totalInvoiced = invoicedByInvoice.TotalInvoiced
	}

	// Calculate invoice status
	response.TotalAcceptedQty = fmt.Sprintf("%.4f", totalAccepted)
	response.TotalInvoicedQty = fmt.Sprintf("%.4f", totalInvoiced)

	if totalAccepted <= 0 {
		response.InvoiceStatus = "NONE"
	} else if totalInvoiced <= 0 {
		response.InvoiceStatus = "NONE"
	} else if totalInvoiced >= totalAccepted {
		response.InvoiceStatus = "FULL"
	} else {
		response.InvoiceStatus = "PARTIAL"
	}

	return response
}
