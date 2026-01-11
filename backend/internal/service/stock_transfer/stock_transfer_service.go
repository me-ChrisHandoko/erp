package stock_transfer

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

// StockTransferService handles business logic for stock transfers
type StockTransferService struct {
	db *gorm.DB
}

// NewStockTransferService creates a new stock transfer service instance
func NewStockTransferService(db *gorm.DB) *StockTransferService {
	return &StockTransferService{
		db: db,
	}
}

// CreateStockTransfer creates a new stock transfer (DRAFT status)
func (s *StockTransferService) CreateStockTransfer(
	ctx context.Context,
	tenantID, companyID, userID string,
	req *dto.CreateStockTransferRequest,
) (*models.StockTransfer, error) {
	fmt.Printf("🔵 [Service.CreateStockTransfer] Called with tenantID=%s, companyID=%s\n", tenantID, companyID)
	fmt.Printf("🔍 [Service.CreateStockTransfer] Request: %+v\n", req)

	// Validation: source != destination
	if req.SourceWarehouseID == req.DestWarehouseID {
		fmt.Println("❌ [Service.CreateStockTransfer] Source and destination are the same")
		return nil, pkgerrors.NewBadRequestError("Source and destination warehouses must be different")
	}

	// Validation: both warehouses must belong to the same company
	var sourceWarehouse, destWarehouse models.Warehouse

	// Check source warehouse ownership
	// IMPORTANT: Set tenant context in GORM session for callback enforcement
	fmt.Printf("🔍 [Service.CreateStockTransfer] Checking source warehouse: %s\n", req.SourceWarehouseID)
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ? AND tenant_id = ?", req.SourceWarehouseID, companyID, tenantID).
		First(&sourceWarehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("❌ [Service.CreateStockTransfer] Source warehouse not found")
			return nil, pkgerrors.NewBadRequestError("Source warehouse not found or does not belong to this company")
		}
		fmt.Printf("❌ [Service.CreateStockTransfer] Source warehouse query error: %v\n", err)
		return nil, pkgerrors.NewInternalError(err)
	}
	fmt.Println("✅ [Service.CreateStockTransfer] Source warehouse found")

	// Check destination warehouse ownership
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ? AND tenant_id = ?", req.DestWarehouseID, companyID, tenantID).
		First(&destWarehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewBadRequestError("Destination warehouse not found or does not belong to this company")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Parse transfer date
	transferDate, err := time.Parse("2006-01-02", req.TransferDate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("Invalid transfer date format")
	}

	// Start transaction with tenant context
	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate transfer number
	transferNumber, err := s.generateTransferNumber(tx, tenantID, companyID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create transfer header
	transfer := &models.StockTransfer{
		TenantID:          tenantID,
		CompanyID:         companyID,
		TransferNumber:    transferNumber,
		TransferDate:      transferDate,
		SourceWarehouseID: req.SourceWarehouseID,
		DestWarehouseID:   req.DestWarehouseID,
		Status:            models.StockTransferStatusDraft,
		Notes:             req.Notes,
	}

	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Create transfer items
	for _, itemReq := range req.Items {
		qty, err := decimal.NewFromString(itemReq.Quantity)
		if err != nil || qty.LessThanOrEqual(decimal.Zero) {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError("Invalid quantity")
		}

		item := &models.StockTransferItem{
			StockTransferID: transfer.ID,
			ProductID:       itemReq.ProductID,
			Quantity:        qty,
			BatchID:         itemReq.BatchID,
			Notes:           itemReq.Notes,
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
	return s.GetStockTransferByID(ctx, tenantID, companyID, transfer.ID)
}

// ListStockTransfers lists stock transfers with filtering
func (s *StockTransferService) ListStockTransfers(
	ctx context.Context,
	tenantID string,
	companyID string,
	query *dto.StockTransferQuery,
) ([]models.StockTransfer, *dto.PaginationInfo, error) {
	var transfers []models.StockTransfer
	var total int64

	db := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("company_id = ?", companyID)

	// Apply filters
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.SourceWarehouseID != nil {
		db = db.Where("source_warehouse_id = ?", *query.SourceWarehouseID)
	}
	if query.DestWarehouseID != nil {
		db = db.Where("dest_warehouse_id = ?", *query.DestWarehouseID)
	}
	if query.Search != "" {
		db = db.Where("transfer_number LIKE ?", "%"+query.Search+"%")
	}

	// Count total
	if err := db.Model(&models.StockTransfer{}).Count(&total).Error; err != nil {
		return nil, nil, pkgerrors.NewInternalError(err)
	}

	// Apply sorting - map camelCase API fields to snake_case DB columns
	sortColumn := "transfer_number"
	if query.SortBy == "transferNumber" {
		sortColumn = "transfer_number"
	} else if query.SortBy == "transferDate" {
		sortColumn = "transfer_date"
	} else if query.SortBy == "status" {
		sortColumn = "status"
	} else if query.SortBy == "createdAt" {
		sortColumn = "created_at"
	} else if query.SortBy != "" {
		sortColumn = query.SortBy // Allow direct DB column names as fallback
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
	db = db.Preload("SourceWarehouse").
		Preload("DestWarehouse").
		Preload("Items.Product")

	if err := db.Find(&transfers).Error; err != nil {
		return nil, nil, pkgerrors.NewInternalError(err)
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	pagination := &dto.PaginationInfo{
		Page:       query.Page,
		Limit:      query.PageSize,
		Total:      int(total),
		TotalPages: totalPages,
	}

	return transfers, pagination, nil
}

// GetStockTransferByID retrieves a single stock transfer
func (s *StockTransferService) GetStockTransferByID(
	ctx context.Context,
	tenantID, companyID, transferID string,
) (*models.StockTransfer, error) {
	var transfer models.StockTransfer

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", transferID, companyID).
		Preload("SourceWarehouse").
		Preload("DestWarehouse").
		Preload("Items.Product").
		First(&transfer).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Stock transfer not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	return &transfer, nil
}

// UpdateStockTransfer updates a DRAFT transfer
func (s *StockTransferService) UpdateStockTransfer(
	ctx context.Context,
	tenantID, companyID, transferID string,
	req *dto.UpdateStockTransferRequest,
) (*models.StockTransfer, error) {
	// Get transfer WITHOUT preloading associations to avoid GORM trying to save them
	var transfer models.StockTransfer
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", transferID, companyID).
		First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Stock transfer not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only DRAFT can be updated
	if transfer.Status != models.StockTransferStatusDraft {
		return nil, pkgerrors.NewBadRequestError("Only DRAFT transfers can be updated")
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update fields
	if req.TransferDate != nil {
		transferDate, err := time.Parse("2006-01-02", *req.TransferDate)
		if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewBadRequestError("Invalid transfer date")
		}
		transfer.TransferDate = transferDate
	}
	if req.SourceWarehouseID != nil {
		transfer.SourceWarehouseID = *req.SourceWarehouseID
	}
	if req.DestWarehouseID != nil {
		transfer.DestWarehouseID = *req.DestWarehouseID
	}
	if req.Notes != nil {
		transfer.Notes = req.Notes
	}

	// Validate source != destination
	if transfer.SourceWarehouseID == transfer.DestWarehouseID {
		tx.Rollback()
		return nil, pkgerrors.NewBadRequestError("Source and destination must be different")
	}

	// Validate warehouse ownership if warehouse IDs were changed
	if req.SourceWarehouseID != nil || req.DestWarehouseID != nil {
		var sourceWarehouse, destWarehouse models.Warehouse

		// Check source warehouse ownership
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("id = ? AND company_id = ? AND tenant_id = ?", transfer.SourceWarehouseID, companyID, transfer.TenantID).
			First(&sourceWarehouse).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, pkgerrors.NewBadRequestError("Source warehouse not found or does not belong to this company")
			}
			return nil, pkgerrors.NewInternalError(err)
		}

		// Check destination warehouse ownership
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("id = ? AND company_id = ? AND tenant_id = ?", transfer.DestWarehouseID, companyID, transfer.TenantID).
			First(&destWarehouse).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, pkgerrors.NewBadRequestError("Destination warehouse not found or does not belong to this company")
			}
			return nil, pkgerrors.NewInternalError(err)
		}
	}

	// Update only the transfer fields, not associations
	// Use Updates to avoid saving preloaded associations (warehouses)
	updateFields := map[string]interface{}{
		"transfer_date":       transfer.TransferDate,
		"source_warehouse_id": transfer.SourceWarehouseID,
		"dest_warehouse_id":   transfer.DestWarehouseID,
		"updated_at":          time.Now(),
	}

	// Only update notes if it was provided in the request
	if req.Notes != nil {
		updateFields["notes"] = transfer.Notes
	}

	if err := tx.Model(transfer).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Update items if provided
	if req.Items != nil {
		// Delete existing items
		if err := tx.Where("stock_transfer_id = ?", transferID).Delete(&models.StockTransferItem{}).Error; err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}

		// Create new items
		for _, itemReq := range *req.Items {
			qty, err := decimal.NewFromString(itemReq.Quantity)
			if err != nil || qty.LessThanOrEqual(decimal.Zero) {
				tx.Rollback()
				return nil, pkgerrors.NewBadRequestError("Invalid quantity")
			}

			item := &models.StockTransferItem{
				StockTransferID: transfer.ID,
				ProductID:       itemReq.ProductID,
				Quantity:        qty,
				BatchID:         itemReq.BatchID,
				Notes:           itemReq.Notes,
			}

			if err := tx.Create(item).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewInternalError(err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetStockTransferByID(ctx, tenantID, companyID, transferID)
}

// DeleteStockTransfer deletes a DRAFT transfer
func (s *StockTransferService) DeleteStockTransfer(
	ctx context.Context,
	tenantID, companyID, transferID string,
) error {
	transfer, err := s.GetStockTransferByID(ctx, tenantID, companyID, transferID)
	if err != nil {
		return err
	}

	// Only DRAFT can be deleted
	if transfer.Status != models.StockTransferStatusDraft {
		return pkgerrors.NewBadRequestError("Only DRAFT transfers can be deleted")
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Delete items first (CASCADE should handle this, but explicit is better)
	if err := tx.Where("stock_transfer_id = ?", transferID).Delete(&models.StockTransferItem{}).Error; err != nil {
		tx.Rollback()
		return pkgerrors.NewInternalError(err)
	}

	// Delete transfer
	if err := tx.Delete(transfer).Error; err != nil {
		tx.Rollback()
		return pkgerrors.NewInternalError(err)
	}

	return tx.Commit().Error
}

// ShipStockTransfer ships a transfer (DRAFT → SHIPPED)
func (s *StockTransferService) ShipStockTransfer(
	ctx context.Context,
	tenantID, companyID, transferID, userID string,
	req *dto.ShipTransferRequest,
) (*models.StockTransfer, error) {
	// Get transfer WITHOUT preloading associations to avoid GORM trying to save them
	var transfer models.StockTransfer
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", transferID, companyID).
		First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Stock transfer not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only DRAFT can be shipped
	if transfer.Status != models.StockTransferStatusDraft {
		return nil, pkgerrors.NewBadRequestError("Only DRAFT transfers can be shipped")
	}

	// Load transfer items to process inventory movements
	var items []models.StockTransferItem
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("stock_transfer_id = ?", transferID).
		Find(&items).Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	// Check stock availability in source warehouse
	for _, item := range items {
		var stock models.WarehouseStock
		err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("warehouse_id = ? AND product_id = ?", transfer.SourceWarehouseID, item.ProductID).
			First(&stock).Error

		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Product not found in source warehouse"))
		}
		if err != nil {
			return nil, pkgerrors.NewInternalError(err)
		}

		if stock.Quantity.LessThan(item.Quantity) {
			return nil, pkgerrors.NewBadRequestError(fmt.Sprintf("Insufficient stock for product. Available: %s, Required: %s", stock.Quantity.String(), item.Quantity.String()))
		}
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update transfer status using Updates to avoid saving associations
	now := time.Now()
	updateFields := map[string]interface{}{
		"status":      models.StockTransferStatusShipped,
		"shipped_by":  userID,
		"shipped_at":  now,
		"updated_at":  now,
	}

	// Only update notes if provided in request
	if req.Notes != nil {
		updateFields["notes"] = *req.Notes
	}

	if err := tx.Model(&transfer).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Reduce stock at source warehouse
	for _, item := range items {
		var stock models.WarehouseStock
		if err := tx.Where("warehouse_id = ? AND product_id = ?", transfer.SourceWarehouseID, item.ProductID).
			First(&stock).Error; err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}

		// Decrease quantity
		newQuantity := stock.Quantity.Sub(item.Quantity)
		if err := tx.Model(&stock).Update("quantity", newQuantity).Error; err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetStockTransferByID(ctx, tenantID, companyID, transferID)
}

// ReceiveStockTransfer receives a transfer (SHIPPED → RECEIVED)
func (s *StockTransferService) ReceiveStockTransfer(
	ctx context.Context,
	tenantID, companyID, transferID, userID string,
	req *dto.ReceiveTransferRequest,
) (*models.StockTransfer, error) {
	// Get transfer WITHOUT preloading associations to avoid GORM trying to save them
	var transfer models.StockTransfer
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", transferID, companyID).
		First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Stock transfer not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only SHIPPED can be received
	if transfer.Status != models.StockTransferStatusShipped {
		return nil, pkgerrors.NewBadRequestError("Only SHIPPED transfers can be received")
	}

	// Load transfer items to process inventory movements
	var items []models.StockTransferItem
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("stock_transfer_id = ?", transferID).
		Find(&items).Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update transfer status using Updates to avoid saving associations
	now := time.Now()
	updateFields := map[string]interface{}{
		"status":       models.StockTransferStatusReceived,
		"received_by":  userID,
		"received_at":  now,
		"updated_at":   now,
	}

	// Only update notes if provided in request
	if req.Notes != nil {
		updateFields["notes"] = *req.Notes
	}

	if err := tx.Model(&transfer).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// Add stock at destination warehouse
	for _, item := range items {
		var stock models.WarehouseStock

		// Try to find existing stock record
		err := tx.Where("warehouse_id = ? AND product_id = ?", transfer.DestWarehouseID, item.ProductID).
			First(&stock).Error

		if err == gorm.ErrRecordNotFound {
			// Create new stock record if it doesn't exist
			stock = models.WarehouseStock{
				ID:          uuid.New().String(),
				WarehouseID: transfer.DestWarehouseID,
				ProductID:   item.ProductID,
				Quantity:    item.Quantity,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := tx.Create(&stock).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewInternalError(err)
			}
		} else if err != nil {
			tx.Rollback()
			return nil, pkgerrors.NewInternalError(err)
		} else {
			// Update existing stock by adding the transferred quantity
			newQuantity := stock.Quantity.Add(item.Quantity)
			if err := tx.Model(&stock).Update("quantity", newQuantity).Error; err != nil {
				tx.Rollback()
				return nil, pkgerrors.NewInternalError(err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetStockTransferByID(ctx, tenantID, companyID, transferID)
}

// CancelStockTransfer cancels a transfer (SHIPPED → CANCELLED)
func (s *StockTransferService) CancelStockTransfer(
	ctx context.Context,
	tenantID, companyID, transferID string,
	req *dto.CancelTransferRequest,
) (*models.StockTransfer, error) {
	// Get transfer WITHOUT preloading associations to avoid GORM trying to save them
	var transfer models.StockTransfer
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", transferID, companyID).
		First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Stock transfer not found")
		}
		return nil, pkgerrors.NewInternalError(err)
	}

	// Only SHIPPED can be cancelled
	if transfer.Status != models.StockTransferStatusShipped {
		return nil, pkgerrors.NewBadRequestError("Only SHIPPED transfers can be cancelled")
	}

	tx := s.db.WithContext(ctx).Set("tenant_id", tenantID).Begin()

	// Update transfer status using Updates to avoid saving associations
	cancelNote := fmt.Sprintf("CANCELLED: %s", req.Reason)
	updateFields := map[string]interface{}{
		"status":     models.StockTransferStatusCancelled,
		"notes":      cancelNote,
		"updated_at": time.Now(),
	}

	if err := tx.Model(&transfer).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, pkgerrors.NewInternalError(err)
	}

	// TODO: Reverse inventory movements (return stock to source warehouse)

	if err := tx.Commit().Error; err != nil {
		return nil, pkgerrors.NewInternalError(err)
	}

	return s.GetStockTransferByID(ctx, tenantID, companyID, transferID)
}

// generateTransferNumber generates unique transfer number for company
func (s *StockTransferService) generateTransferNumber(tx *gorm.DB, tenantID, companyID string) (string, error) {
	var count int64
	currentYear := time.Now().Year()
	prefix := fmt.Sprintf("TRF-%d-", currentYear)

	if err := tx.Model(&models.StockTransfer{}).
		Where("company_id = ? AND tenant_id = ? AND transfer_number LIKE ?", companyID, tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return "", pkgerrors.NewInternalError(err)
	}

	return fmt.Sprintf("%s%05d", prefix, count+1), nil
}
