package inventoryadjustment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// InventoryAdjustmentService handles business logic for inventory adjustments
type InventoryAdjustmentService struct {
	db *gorm.DB
}

// NewInventoryAdjustmentService creates a new inventory adjustment service instance
func NewInventoryAdjustmentService(db *gorm.DB) *InventoryAdjustmentService {
	return &InventoryAdjustmentService{
		db: db,
	}
}

// CreateInventoryAdjustment creates a new inventory adjustment (DRAFT status)
func (s *InventoryAdjustmentService) CreateInventoryAdjustment(
	ctx context.Context,
	tenantID, companyID, userID string,
	req *dto.CreateInventoryAdjustmentRequest,
) (*models.InventoryAdjustment, error) {
	// Validate warehouse belongs to company
	var warehouse models.Warehouse
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ? AND tenant_id = ?", req.WarehouseID, companyID, tenantID).
		First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewBadRequestError("Warehouse not found or does not belong to this company")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Parse adjustment date
	adjustmentDate, err := time.Parse("2006-01-02", req.AdjustmentDate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("Invalid adjustment date format")
	}

	// Start transaction
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate adjustment number
	adjustmentNumber, err := s.generateAdjustmentNumber(tx, tenantID, companyID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Calculate totals and validate items
	totalItems := len(req.Items)
	totalValue := decimal.Zero

	// Validate all products exist and belong to company
	for _, itemReq := range req.Items {
		var product models.Product
		if err := tx.Where("id = ? AND company_id = ?", itemReq.ProductID, companyID).
			First(&product).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Product not found: %s", itemReq.ProductID))
			}
			return nil, pkgerrors.NewInternalError(err)
		}

		qty, err := decimal.NewFromString(itemReq.QuantityAdjusted)
		if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError("Invalid quantity adjusted")
		}

		unitCost, err := decimal.NewFromString(itemReq.UnitCost)
		if err != nil || unitCost.LessThan(decimal.Zero) {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError("Invalid unit cost")
		}

		// For DECREASE type, quantity should be positive in request (will be applied as negative)
		if qty.LessThanOrEqual(decimal.Zero) {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError("Quantity adjusted must be greater than zero")
		}

		totalValue = totalValue.Add(qty.Mul(unitCost).Abs())
	}

	// Create adjustment header
	adjustment := &models.InventoryAdjustment{
		TenantID:         tenantID,
		CompanyID:        companyID,
		AdjustmentNumber: adjustmentNumber,
		AdjustmentDate:   adjustmentDate,
		WarehouseID:      req.WarehouseID,
		AdjustmentType:   models.InventoryAdjustmentType(req.AdjustmentType),
		Reason:           models.InventoryAdjustmentReason(req.Reason),
		Status:           models.InventoryAdjustmentStatusDraft,
		Notes:            req.Notes,
		TotalItems:       totalItems,
		TotalValue:       totalValue,
		CreatedBy:        userID,
	}

	if err := tx.Create(adjustment).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Create adjustment items
	for _, itemReq := range req.Items {
		qty, _ := decimal.NewFromString(itemReq.QuantityAdjusted)
		unitCost, _ := decimal.NewFromString(itemReq.UnitCost)

		// Get current stock quantity
		var warehouseStock models.WarehouseStock
		currentQty := decimal.Zero
		err := tx.Where("warehouse_id = ? AND product_id = ?", req.WarehouseID, itemReq.ProductID).
			First(&warehouseStock).Error
		if err == nil {
			currentQty = warehouseStock.Quantity
		} else if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}

		// Calculate quantity after based on adjustment type
		var qtyAdjusted, qtyAfter decimal.Decimal
		if req.AdjustmentType == "INCREASE" {
			qtyAdjusted = qty
			qtyAfter = currentQty.Add(qty)
		} else {
			qtyAdjusted = qty.Neg() // Make it negative for DECREASE
			qtyAfter = currentQty.Sub(qty)
		}

		item := &models.InventoryAdjustmentItem{
			AdjustmentID:     adjustment.ID,
			ProductID:        itemReq.ProductID,
			BatchID:          itemReq.BatchID,
			QuantityBefore:   currentQty,
			QuantityAdjusted: qtyAdjusted,
			QuantityAfter:    qtyAfter,
			UnitCost:         unitCost,
			TotalValue:       qty.Mul(unitCost),
			Notes:            itemReq.Notes,
		}

		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	// Reload with associations
	return s.GetInventoryAdjustmentByID(ctx, tenantID, companyID, adjustment.ID)
}

// ListInventoryAdjustments lists inventory adjustments with filtering
func (s *InventoryAdjustmentService) ListInventoryAdjustments(
	ctx context.Context,
	tenantID string,
	companyID string,
	query *dto.InventoryAdjustmentQuery,
) ([]models.InventoryAdjustment, *dto.PaginationInfo, error) {
	var adjustments []models.InventoryAdjustment
	var total int64

	db := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.AdjustmentType != nil {
		db = db.Where("adjustment_type = ?", *query.AdjustmentType)
	}
	if query.Reason != nil {
		db = db.Where("reason = ?", *query.Reason)
	}
	if query.WarehouseID != nil {
		db = db.Where("warehouse_id = ?", *query.WarehouseID)
	}
	if query.Search != "" {
		db = db.Where("adjustment_number LIKE ?", "%"+query.Search+"%")
	}
	if query.DateFrom != nil {
		db = db.Where("adjustment_date >= ?", *query.DateFrom)
	}
	if query.DateTo != nil {
		db = db.Where("adjustment_date <= ?", *query.DateTo)
	}

	// Count total
	if err := db.Model(&models.InventoryAdjustment{}).Count(&total).Error; err != nil {
		return nil, nil, pkgerrors.NewInternalError(err)
	}

	// Apply sorting - map camelCase API fields to snake_case DB columns
	sortColumn := "adjustment_number"
	if query.SortBy == "adjustmentNumber" {
		sortColumn = "adjustment_number"
	} else if query.SortBy == "adjustmentDate" {
		sortColumn = "adjustment_date"
	} else if query.SortBy == "status" {
		sortColumn = "status"
	} else if query.SortBy == "createdAt" {
		sortColumn = "created_at"
	} else if query.SortBy != "" {
		sortColumn = query.SortBy
	}

	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	db = db.Order(fmt.Sprintf("%s %s", sortColumn, sortOrder))

	// Apply pagination
	offset := (query.Page - 1) * query.PageSize
	db = db.Offset(offset).Limit(query.PageSize)

	// Preload associations
	db = db.Preload("Warehouse").
		Preload("CreatedByUser").
		Preload("ApprovedByUser").
		Preload("Items.Product")

	if err := db.Find(&adjustments).Error; err != nil {
		return nil, nil, pkgerrors.NewInternalError(err)
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	pagination := &dto.PaginationInfo{
		Page:       query.Page,
		Limit:      query.PageSize,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return adjustments, pagination, nil
}

// GetInventoryAdjustmentByID retrieves a single inventory adjustment
func (s *InventoryAdjustmentService) GetInventoryAdjustmentByID(
	ctx context.Context,
	tenantID, companyID, adjustmentID string,
) (*models.InventoryAdjustment, error) {
	var adjustment models.InventoryAdjustment

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", adjustmentID, companyID).
		Preload("Warehouse").
		Preload("CreatedByUser").
		Preload("ApprovedByUser").
		Preload("Items.Product").
		First(&adjustment).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Inventory adjustment not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	return &adjustment, nil
}

// UpdateInventoryAdjustment updates a DRAFT adjustment
func (s *InventoryAdjustmentService) UpdateInventoryAdjustment(
	ctx context.Context,
	tenantID, companyID, adjustmentID string,
	req *dto.UpdateInventoryAdjustmentRequest,
) (*models.InventoryAdjustment, error) {
	// Get adjustment without preloading
	var adjustment models.InventoryAdjustment
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", adjustmentID, companyID).
		First(&adjustment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Inventory adjustment not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only DRAFT can be updated
	if adjustment.Status != models.InventoryAdjustmentStatusDraft {
		return nil, pkgerrors.NewBadRequestError("Only DRAFT adjustments can be updated")
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update fields
	if req.AdjustmentDate != nil {
		adjustmentDate, err := time.Parse("2006-01-02", *req.AdjustmentDate)
		if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError("Invalid adjustment date")
		}
		adjustment.AdjustmentDate = adjustmentDate
	}
	if req.WarehouseID != nil {
		// Validate warehouse belongs to company
		var warehouse models.Warehouse
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", *req.WarehouseID, companyID, tenantID).
			First(&warehouse).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, pkgerrors.NewBadRequestError("Warehouse not found")
			}
			return nil, pkgerrors.NewInternalError(err)
		}
		adjustment.WarehouseID = *req.WarehouseID
	}
	if req.AdjustmentType != nil {
		adjustment.AdjustmentType = models.InventoryAdjustmentType(*req.AdjustmentType)
	}
	if req.Reason != nil {
		adjustment.Reason = models.InventoryAdjustmentReason(*req.Reason)
	}
	if req.Notes != nil {
		adjustment.Notes = req.Notes
	}

	// Update header fields
	updateFields := map[string]interface{}{
		"adjustment_date": adjustment.AdjustmentDate,
		"warehouse_id":    adjustment.WarehouseID,
		"adjustment_type": adjustment.AdjustmentType,
		"reason":          adjustment.Reason,
		"updated_at":      time.Now(),
	}
	if req.Notes != nil {
		updateFields["notes"] = adjustment.Notes
	}

	if err := tx.Model(&adjustment).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Update items if provided
	if req.Items != nil {
		// Delete existing items
		if err := tx.Where("adjustment_id = ?", adjustmentID).Delete(&models.InventoryAdjustmentItem{}).Error; err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}

		// Recalculate totals
		totalItems := len(*req.Items)
		totalValue := decimal.Zero

		// Create new items
		for _, itemReq := range *req.Items {
			qty, err := decimal.NewFromString(itemReq.QuantityAdjusted)
			if err != nil || qty.LessThanOrEqual(decimal.Zero) {
				tx.Rollback()
				return nil, pkgerrors.NewBadRequestError("Invalid quantity adjusted")
			}

			unitCost, err := decimal.NewFromString(itemReq.UnitCost)
			if err != nil || unitCost.LessThan(decimal.Zero) {
				tx.Rollback()
				return nil, pkgerrors.NewBadRequestError("Invalid unit cost")
			}

			// Get current stock quantity
			var warehouseStock models.WarehouseStock
			currentQty := decimal.Zero
			err = tx.Where("warehouse_id = ? AND product_id = ?", adjustment.WarehouseID, itemReq.ProductID).
				First(&warehouseStock).Error
			if err == nil {
				currentQty = warehouseStock.Quantity
			} else if err != gorm.ErrRecordNotFound {
				tx.Rollback()
				return nil, pkgerrors.NewInternalError(err)
			}

			// Calculate quantity after based on adjustment type
			var qtyAdjusted, qtyAfter decimal.Decimal
			if adjustment.AdjustmentType == models.InventoryAdjustmentTypeIncrease {
				qtyAdjusted = qty
				qtyAfter = currentQty.Add(qty)
			} else {
				qtyAdjusted = qty.Neg()
				qtyAfter = currentQty.Sub(qty)
			}

			item := &models.InventoryAdjustmentItem{
				AdjustmentID:     adjustment.ID,
				ProductID:        itemReq.ProductID,
				BatchID:          itemReq.BatchID,
				QuantityBefore:   currentQty,
				QuantityAdjusted: qtyAdjusted,
				QuantityAfter:    qtyAfter,
				UnitCost:         unitCost,
				TotalValue:       qty.Mul(unitCost),
				Notes:            itemReq.Notes,
			}

			if err := tx.Create(item).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewInternalError(err)
			}

			totalValue = totalValue.Add(qty.Mul(unitCost))
		}

		// Update totals
		if err := tx.Model(&adjustment).Updates(map[string]interface{}{
			"total_items": totalItems,
			"total_value": totalValue,
		}).Error; err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetInventoryAdjustmentByID(ctx, tenantID, companyID, adjustmentID)
}

// DeleteInventoryAdjustment deletes a DRAFT adjustment
func (s *InventoryAdjustmentService) DeleteInventoryAdjustment(
	ctx context.Context,
	tenantID, companyID, adjustmentID string,
) error {
	adjustment, err := s.GetInventoryAdjustmentByID(ctx, tenantID, companyID, adjustmentID)
	if err != nil {
		return err
	}

	// Only DRAFT can be deleted
	if adjustment.Status != models.InventoryAdjustmentStatusDraft {
		return pkgerrors.NewBadRequestError("Only DRAFT adjustments can be deleted")
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Delete items first
	if err := tx.Where("adjustment_id = ?", adjustmentID).Delete(&models.InventoryAdjustmentItem{}).Error; err != nil {
		tx.Rollback()
		return pkgerrors.NewInternalError(err)
	}

	// Delete adjustment
	if err := tx.Delete(adjustment).Error; err != nil {
		tx.Rollback()
		return pkgerrors.NewInternalError(err)
	}

	return tx.Commit().Error
}

// ApproveInventoryAdjustment approves an adjustment (DRAFT → APPROVED)
func (s *InventoryAdjustmentService) ApproveInventoryAdjustment(
	ctx context.Context,
	tenantID, companyID, adjustmentID, userID string,
	req *dto.ApproveAdjustmentRequest,
) (*models.InventoryAdjustment, error) {
	// Get adjustment without preloading
	var adjustment models.InventoryAdjustment
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", adjustmentID, companyID).
		First(&adjustment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Inventory adjustment not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only DRAFT can be approved
	if adjustment.Status != models.InventoryAdjustmentStatusDraft {
		return nil, pkgerrors.NewBadRequestError("Only DRAFT adjustments can be approved")
	}

	// Load items for stock update
	var items []models.InventoryAdjustmentItem
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("adjustment_id = ?", adjustmentID).
		Find(&items).Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	if len(items) == 0 {
		return nil, pkgerrors.NewBadRequestError("Cannot approve adjustment with no items")
	}

	// Validate stock for DECREASE adjustments
	if adjustment.AdjustmentType == models.InventoryAdjustmentTypeDecrease {
		for _, item := range items {
			var stock models.WarehouseStock
			err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
				Where("warehouse_id = ? AND product_id = ?", adjustment.WarehouseID, item.ProductID).
				First(&stock).Error

			if err == gorm.ErrRecordNotFound {
				return nil, pkgerrors.NewBadRequestError("Product not found in warehouse")
			}
			if err != nil {
				return nil, pkgerrors.NewInternalError(err)
			}

			// QuantityAdjusted is negative for DECREASE, so use Abs
			if stock.Quantity.LessThan(item.QuantityAdjusted.Abs()) {
				return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Insufficient stock. Available: %s, Required: %s",
					stock.Quantity.String(), item.QuantityAdjusted.Abs().String()))
			}
		}
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update adjustment status
	now := time.Now()
	updateFields := map[string]interface{}{
		"status":      models.InventoryAdjustmentStatusApproved,
		"approved_by": userID,
		"approved_at": now,
		"updated_at":  now,
	}

	if req != nil && req.Notes != nil {
		updateFields["notes"] = *req.Notes
	}

	if err := tx.Model(&adjustment).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Apply stock changes
	for _, item := range items {
		var stock models.WarehouseStock

		err := tx.Where("warehouse_id = ? AND product_id = ?", adjustment.WarehouseID, item.ProductID).
			First(&stock).Error

		if err == gorm.ErrRecordNotFound {
			// Create new stock record for INCREASE
			if adjustment.AdjustmentType == models.InventoryAdjustmentTypeIncrease {
				stock = models.WarehouseStock{
					ID:          uuid.New().String(),
					WarehouseID: adjustment.WarehouseID,
					ProductID:   item.ProductID,
					Quantity:    item.QuantityAdjusted,
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				if err := tx.Create(&stock).Error; err != nil {
					tx.Rollback()
					return nil, pkgerrors.NewInternalError(err)
				}
			} else {
				// Should not happen due to validation above
				tx.Rollback()
				return nil, pkgerrors.NewBadRequestError("Product not found in warehouse")
			}
		} else if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		} else {
			// Update existing stock
			// QuantityAdjusted is positive for INCREASE, negative for DECREASE
			newQuantity := stock.Quantity.Add(item.QuantityAdjusted)
			if err := tx.Model(&stock).Update("quantity", newQuantity).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewInternalError(err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetInventoryAdjustmentByID(ctx, tenantID, companyID, adjustmentID)
}

// CancelInventoryAdjustment cancels an adjustment (DRAFT → CANCELLED)
func (s *InventoryAdjustmentService) CancelInventoryAdjustment(
	ctx context.Context,
	tenantID, companyID, adjustmentID, userID string,
	req *dto.CancelAdjustmentRequest,
) (*models.InventoryAdjustment, error) {
	// Get adjustment without preloading
	var adjustment models.InventoryAdjustment
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", adjustmentID, companyID).
		First(&adjustment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Inventory adjustment not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only DRAFT can be cancelled
	if adjustment.Status != models.InventoryAdjustmentStatusDraft {
		return nil, pkgerrors.NewBadRequestError("Only DRAFT adjustments can be cancelled")
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update adjustment status
	now := time.Now()
	updateFields := map[string]interface{}{
		"status":        models.InventoryAdjustmentStatusCancelled,
		"cancelled_by":  userID,
		"cancelled_at":  now,
		"cancel_reason": req.Reason,
		"updated_at":    now,
	}

	if err := tx.Model(&adjustment).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetInventoryAdjustmentByID(ctx, tenantID, companyID, adjustmentID)
}

// generateAdjustmentNumber generates unique adjustment number for company
func (s *InventoryAdjustmentService) generateAdjustmentNumber(tx *gorm.DB, tenantID, companyID string) (string, error) {
	var count int64
	currentYear := time.Now().Year()
	prefix := fmt.Sprintf("ADJ-%d-", currentYear)

	if err := tx.Model(&models.InventoryAdjustment{}).
		Where("company_id = ? AND tenant_id = ? AND adjustment_number LIKE ?", companyID, tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return "", pkgerrors.NewInternalError(err)
	}

	return fmt.Sprintf("%s%05d", prefix, count+1), nil
}
