package goodsreceipt

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// GoodsReceiptService - Business logic for goods receipt management
type GoodsReceiptService struct {
	db *gorm.DB
}

// NewGoodsReceiptService creates a new goods receipt service instance
func NewGoodsReceiptService(db *gorm.DB) *GoodsReceiptService {
	return &GoodsReceiptService{
		db: db,
	}
}

// ============================================================================
// CREATE GOODS RECEIPT
// ============================================================================

// CreateGoodsReceipt creates a new goods receipt from a purchase order
func (s *GoodsReceiptService) CreateGoodsReceipt(ctx context.Context, tenantID, companyID, userID string, req *dto.CreateGoodsReceiptRequest) (*models.GoodsReceipt, error) {
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

	// Load purchase order items for validation
	var poItems []models.PurchaseOrderItem
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
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

			// Parse quantities
			receivedQty, err := decimal.NewFromString(itemReq.ReceivedQty)
			if err != nil || receivedQty.LessThan(decimal.Zero) {
				return pkgerrors.NewBadRequestError("invalid receivedQty")
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
		baseQuery = baseQuery.Where("grn_number LIKE ?", searchPattern)
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
		responses[i] = s.MapToResponse(&gr, false)
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

	// Load items with product info
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		Preload("ProductUnit").
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

		if err := tx.Save(goodsReceipt).Error; err != nil {
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
func (s *GoodsReceiptService) ReceiveGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.ReceiveGoodsRequest) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusPending {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in PENDING status")
	}

	now := time.Now()
	goodsReceipt.Status = models.GoodsReceiptStatusReceived
	goodsReceipt.ReceivedBy = &userID
	goodsReceipt.ReceivedAt = &now
	if req != nil && req.Notes != nil {
		goodsReceipt.Notes = req.Notes
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(goodsReceipt).Error; err != nil {
		return nil, fmt.Errorf("failed to update goods receipt status: %w", err)
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// InspectGoods marks goods as inspected (RECEIVED → INSPECTED)
func (s *GoodsReceiptService) InspectGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.InspectGoodsRequest) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusReceived {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in RECEIVED status")
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
			goodsReceipt.Notes = req.Notes
		}

		return tx.Save(goodsReceipt).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to inspect goods: %w", err)
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// AcceptGoods accepts goods and updates stock (INSPECTED → ACCEPTED)
func (s *GoodsReceiptService) AcceptGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.AcceptGoodsRequest) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusInspected {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in INSPECTED status")
	}

	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Update warehouse stock for each accepted item
		for _, item := range goodsReceipt.Items {
			if item.AcceptedQty.GreaterThan(decimal.Zero) {
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

				// Update received quantity on purchase order item
				if err := tx.Model(&models.PurchaseOrderItem{}).
					Where("id = ?", item.PurchaseOrderItemID).
					Update("received_qty", gorm.Expr("received_qty + ?", item.AcceptedQty)).Error; err != nil {
					return fmt.Errorf("failed to update PO item received qty: %w", err)
				}
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
			goodsReceipt.Notes = req.Notes
		}

		return tx.Save(goodsReceipt).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to accept goods: %w", err)
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// RejectGoods rejects goods (INSPECTED → REJECTED)
func (s *GoodsReceiptService) RejectGoods(ctx context.Context, tenantID, companyID, goodsReceiptID, userID string, req *dto.RejectGoodsRequest) (*models.GoodsReceipt, error) {
	goodsReceipt, err := s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
	if err != nil {
		return nil, err
	}

	if goodsReceipt.Status != models.GoodsReceiptStatusInspected {
		return nil, pkgerrors.NewBadRequestError("goods receipt must be in INSPECTED status")
	}

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
		goodsReceipt.Notes = &notes

		return tx.Save(goodsReceipt).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to reject goods: %w", err)
	}

	return s.GetGoodsReceiptByID(ctx, tenantID, companyID, goodsReceiptID)
}

// ============================================================================
// RESPONSE MAPPING
// ============================================================================

// MapToResponse maps a GoodsReceipt model to response DTO
func (s *GoodsReceiptService) MapToResponse(gr *models.GoodsReceipt, includeItems bool) dto.GoodsReceiptResponse {
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
		response.Items = make([]dto.GoodsReceiptItemResponse, len(gr.Items))
		for i, item := range gr.Items {
			response.Items[i] = dto.GoodsReceiptItemResponse{
				ID:                  item.ID,
				GoodsReceiptID:      item.GoodsReceiptID,
				PurchaseOrderItemID: item.PurchaseOrderItemID,
				ProductID:           item.ProductID,
				ProductUnitID:       item.ProductUnitID,
				BatchNumber:         item.BatchNumber,
				ManufactureDate:     item.ManufactureDate,
				ExpiryDate:          item.ExpiryDate,
				OrderedQty:          item.OrderedQty.String(),
				ReceivedQty:         item.ReceivedQty.String(),
				AcceptedQty:         item.AcceptedQty.String(),
				RejectedQty:         item.RejectedQty.String(),
				RejectionReason:     item.RejectionReason,
				QualityNote:         item.QualityNote,
				Notes:               item.Notes,
				CreatedAt:           item.CreatedAt,
				UpdatedAt:           item.UpdatedAt,
			}

			// Map product
			if item.Product.ID != "" {
				response.Items[i].Product = &dto.GoodsReceiptProductResponse{
					ID:       item.Product.ID,
					Code:     item.Product.Code,
					Name:     item.Product.Name,
					BaseUnit: item.Product.BaseUnit,
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
		}
	}

	return response
}
