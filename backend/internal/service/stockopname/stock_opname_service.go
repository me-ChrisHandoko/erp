package stockopname

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

type StockOpnameService struct {
	db *gorm.DB
}

func NewStockOpnameService(db *gorm.DB) *StockOpnameService {
	return &StockOpnameService{
		db: db,
	}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

// CreateStockOpname creates a new stock opname with items
func (s *StockOpnameService) CreateStockOpname(ctx context.Context, companyID string, tenantID string, userID string, req *dto.CreateStockOpnameRequest) (*models.StockOpname, error) {
	// Parse opname date
	opnameDate, err := time.Parse("2006-01-02", req.OpnameDate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid opnameDate format, expected YYYY-MM-DD")
	}

	// Validate warehouse exists
	var warehouse models.Warehouse
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND company_id = ?", req.WarehouseID, companyID).
		First(&warehouse).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("warehouse not found")
		}
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}

	var opname *models.StockOpname

	// Use transaction for atomic create
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Generate opname number
		opnameNumber, err := s.generateOpnameNumber(tx, companyID, opnameDate)
		if err != nil {
			return fmt.Errorf("failed to generate opname number: %w", err)
		}

		// Create stock opname header
		opname = &models.StockOpname{
			TenantID:     tenantID,
			CompanyID:    companyID,
			OpnameNumber: opnameNumber,
			OpnameDate:   opnameDate,
			WarehouseID:  req.WarehouseID,
			Status:       models.StockOpnameStatusDraft,
			Notes:        req.Notes,
		}

		if userID != "" {
			opname.CountedBy = &userID
		}

		if err := tx.Create(opname).Error; err != nil {
			return fmt.Errorf("failed to create stock opname: %w", err)
		}

		// Create stock opname items
		for _, itemReq := range req.Items {
			// Parse quantities
			expectedQty, err := decimal.NewFromString(itemReq.ExpectedQty)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid expectedQty for product %s", itemReq.ProductID))
			}

			actualQty, err := decimal.NewFromString(itemReq.ActualQty)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid actualQty for product %s", itemReq.ProductID))
			}

			// Calculate difference
			difference := actualQty.Sub(expectedQty)

			// Validate product exists
			var product models.Product
			if err := tx.Where("id = ? AND company_id = ?", itemReq.ProductID, companyID).
				First(&product).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewNotFoundError(fmt.Sprintf("product %s not found", itemReq.ProductID))
				}
				return fmt.Errorf("failed to get product: %w", err)
			}

			item := &models.StockOpnameItem{
				StockOpnameID: opname.ID,
				ProductID:     itemReq.ProductID,
				SystemQty:     expectedQty,
				PhysicalQty:   actualQty,
				DifferenceQty: difference,
				Notes:         itemReq.Notes,
			}

			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("failed to create stock opname item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload opname with relations
	return s.GetStockOpname(ctx, companyID, tenantID, opname.ID)
}

// GetStockOpname retrieves a stock opname by ID with all relations
func (s *StockOpnameService) GetStockOpname(ctx context.Context, companyID, tenantID, opnameID string) (*models.StockOpname, error) {
	var opname models.StockOpname

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Warehouse").
		Preload("Items.Product").
		Where("company_id = ? AND id = ?", companyID, opnameID).
		First(&opname).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("stock opname not found")
		}
		return nil, fmt.Errorf("failed to get stock opname: %w", err)
	}

	return &opname, nil
}

// ListStockOpnames retrieves paginated stock opnames with filters
func (s *StockOpnameService) ListStockOpnames(ctx context.Context, tenantID, companyID string, filters *dto.StockOpnameFilters) ([]models.StockOpname, int64, error) {
	// Set defaults
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.SortBy == "" {
		filters.SortBy = "opnameDate"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "desc"
	}

	// Build query
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.StockOpname{}).
		Preload("Warehouse").
		Preload("Items").
		Where("company_id = ?", companyID)

	// Apply filters
	if filters.Search != "" {
		query = query.Where("opname_number ILIKE ?", "%"+filters.Search+"%")
	}

	if filters.WarehouseID != "" {
		query = query.Where("warehouse_id = ?", filters.WarehouseID)
	}

	if filters.Status != "" {
		// Map frontend status to backend enum
		var status models.StockOpnameStatus
		switch filters.Status {
		case "draft":
			status = models.StockOpnameStatusDraft
		case "in_progress":
			status = models.StockOpnameStatusInProgress
		case "completed":
			status = models.StockOpnameStatusCompleted
		case "approved":
			status = models.StockOpnameStatusApproved
		case "cancelled":
			status = models.StockOpnameStatusCancelled
		default:
			status = models.StockOpnameStatus(filters.Status)
		}
		query = query.Where("status = ?", status)
	}

	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err == nil {
			query = query.Where("opname_date >= ?", dateFrom)
		}
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err == nil {
			query = query.Where("opname_date <= ?", dateTo)
		}
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count stock opnames: %w", err)
	}

	// Sort
	sortField := filters.SortBy
	// Map frontend field names to database columns
	switch sortField {
	case "opnameNumber":
		sortField = "opname_number"
	case "opnameDate":
		sortField = "opname_date"
	case "warehouseName":
		sortField = "warehouse_id" // Will need to join for actual name sorting
	}

	query = query.Order(fmt.Sprintf("%s %s", sortField, filters.SortOrder))

	// Paginate
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Execute query
	var opnames []models.StockOpname
	if err := query.Find(&opnames).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list stock opnames: %w", err)
	}

	return opnames, total, nil
}

// UpdateStockOpname updates a stock opname
func (s *StockOpnameService) UpdateStockOpname(ctx context.Context, companyID, tenantID, opnameID string, req *dto.UpdateStockOpnameRequest) (*models.StockOpname, error) {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if opname.Status == models.StockOpnameStatusApproved {
		return nil, pkgerrors.NewBadRequestError("cannot update approved stock opname")
	}

	if opname.Status == models.StockOpnameStatusCancelled {
		return nil, pkgerrors.NewBadRequestError("cannot update cancelled stock opname")
	}

	// Update fields
	updates := make(map[string]interface{})

	if req.OpnameDate != nil {
		opnameDate, err := time.Parse("2006-01-02", *req.OpnameDate)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid opnameDate format")
		}
		updates["opname_date"] = opnameDate
	}

	if req.Status != nil {
		// Map frontend status to backend enum
		var status models.StockOpnameStatus
		switch *req.Status {
		case "draft":
			status = models.StockOpnameStatusDraft
		case "in_progress":
			status = models.StockOpnameStatusInProgress
		case "completed":
			status = models.StockOpnameStatusCompleted
		default:
			return nil, pkgerrors.NewBadRequestError("invalid status value")
		}
		updates["status"] = status
	}

	if req.Notes != nil {
		updates["notes"] = req.Notes
	}

	// Update in database
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Model(&models.StockOpname{}).
		Where("id = ? AND company_id = ?", opnameID, companyID).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update stock opname: %w", err)
	}

	// Reload and return
	return s.GetStockOpname(ctx, companyID, tenantID, opnameID)
}

// DeleteStockOpname deletes a stock opname (only draft can be deleted)
func (s *StockOpnameService) DeleteStockOpname(ctx context.Context, companyID, tenantID, opnameID string) error {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return err
	}

	// Only allow deleting draft opnames
	if opname.Status != models.StockOpnameStatusDraft {
		return pkgerrors.NewBadRequestError("only draft stock opnames can be deleted")
	}

	// Delete in transaction (cascade will delete items)
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Delete items first (explicit deletion for safety)
		if err := tx.Where("stock_opname_id = ?", opnameID).Delete(&models.StockOpnameItem{}).Error; err != nil {
			return fmt.Errorf("failed to delete stock opname items: %w", err)
		}

		// Delete header
		if err := tx.Where("id = ? AND company_id = ?", opnameID, companyID).Delete(&models.StockOpname{}).Error; err != nil {
			return fmt.Errorf("failed to delete stock opname: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// ApproveStockOpname approves a stock opname and posts stock adjustments
func (s *StockOpnameService) ApproveStockOpname(ctx context.Context, companyID, tenantID, opnameID, userID string, req *dto.ApproveStockOpnameRequest) (*models.StockOpname, error) {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if opname.Status != models.StockOpnameStatusCompleted {
		return nil, pkgerrors.NewBadRequestError("only completed stock opnames can be approved")
	}

	// Approve in transaction
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Update opname status
		now := time.Now()
		updates := map[string]interface{}{
			"status":      models.StockOpnameStatusApproved,
			"approved_by": userID,
			"approved_at": now,
		}

		if req.Notes != nil {
			updates["notes"] = req.Notes
		}

		if err := tx.Model(&models.StockOpname{}).
			Where("id = ? AND company_id = ?", opnameID, companyID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update stock opname: %w", err)
		}

		// Post stock adjustments
		for _, item := range opname.Items {
			// Skip if no difference
			if item.DifferenceQty.IsZero() {
				continue
			}

			// Update warehouse stock
			var whStock models.WarehouseStock
			if err := tx.Where("warehouse_id = ? AND product_id = ?", opname.WarehouseID, item.ProductID).
				First(&whStock).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewNotFoundError(fmt.Sprintf("warehouse stock not found for product %s", item.ProductID))
				}
				return fmt.Errorf("failed to get warehouse stock: %w", err)
			}

			// Adjust quantity
			newQuantity := whStock.Quantity.Add(item.DifferenceQty)
			if newQuantity.IsNegative() {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("adjustment would result in negative stock for product %s", item.ProductID))
			}

			if err := tx.Model(&whStock).Update("quantity", newQuantity).Error; err != nil {
				return fmt.Errorf("failed to update warehouse stock: %w", err)
			}

			// TODO: Create stock movement record when StockMovement model is available
			// movement := &models.StockMovement{
			// 	TenantID:          tenantID,
			// 	CompanyID:         companyID,
			// 	MovementDate:      now,
			// 	MovementType:      models.MovementTypeAdjustment,
			// 	ProductID:         item.ProductID,
			// 	SourceWarehouseID: &opname.WarehouseID,
			// 	Quantity:          item.DifferenceQty.Abs(),
			// 	ReferenceType:     "stock_opname",
			// 	ReferenceID:       &opnameID,
			// 	Notes:             item.Notes,
			// }
			//
			// if err := tx.Create(movement).Error; err != nil {
			// 	return fmt.Errorf("failed to create stock movement: %w", err)
			// }
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload and return
	return s.GetStockOpname(ctx, companyID, tenantID, opnameID)
}

// ============================================================================
// ITEM OPERATIONS
// ============================================================================

// AddStockOpnameItem adds a new item to a stock opname
func (s *StockOpnameService) AddStockOpnameItem(ctx context.Context, companyID, tenantID, opnameID string, req *dto.CreateStockOpnameItemRequest) (*models.StockOpnameItem, error) {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if opname.Status == models.StockOpnameStatusApproved {
		return nil, pkgerrors.NewBadRequestError("cannot add items to approved stock opname")
	}

	if opname.Status == models.StockOpnameStatusCancelled {
		return nil, pkgerrors.NewBadRequestError("cannot add items to cancelled stock opname")
	}

	// Parse quantities
	expectedQty, err := decimal.NewFromString(req.ExpectedQty)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid expectedQty format")
	}

	actualQty, err := decimal.NewFromString(req.ActualQty)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid actualQty format")
	}

	// Calculate difference
	difference := actualQty.Sub(expectedQty)

	// Create item
	item := &models.StockOpnameItem{
		StockOpnameID: opnameID,
		ProductID:     req.ProductID,
		SystemQty:     expectedQty,
		PhysicalQty:   actualQty,
		DifferenceQty: difference,
		Notes:         req.Notes,
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Create(item).Error; err != nil {
		return nil, fmt.Errorf("failed to create stock opname item: %w", err)
	}

	// Reload with product info
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		First(item, "id = ?", item.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload stock opname item: %w", err)
	}

	return item, nil
}

// UpdateStockOpnameItem updates a stock opname item
func (s *StockOpnameService) UpdateStockOpnameItem(ctx context.Context, companyID, tenantID, opnameID, itemID string, req *dto.UpdateStockOpnameItemRequest) (*models.StockOpnameItem, error) {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if opname.Status == models.StockOpnameStatusApproved {
		return nil, pkgerrors.NewBadRequestError("cannot update items in approved stock opname")
	}

	if opname.Status == models.StockOpnameStatusCancelled {
		return nil, pkgerrors.NewBadRequestError("cannot update items in cancelled stock opname")
	}

	// Get existing item
	var item models.StockOpnameItem
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND stock_opname_id = ?", itemID, opnameID).
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("stock opname item not found")
		}
		return nil, fmt.Errorf("failed to get stock opname item: %w", err)
	}

	// Update fields
	updates := make(map[string]interface{})

	if req.ActualQty != nil {
		actualQty, err := decimal.NewFromString(*req.ActualQty)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid actualQty format")
		}

		// Recalculate difference
		difference := actualQty.Sub(item.SystemQty)

		updates["physical_qty"] = actualQty
		updates["difference_qty"] = difference
	}

	if req.Notes != nil {
		updates["notes"] = req.Notes
	}

	// Update in database
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Model(&item).
		Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update stock opname item: %w", err)
	}

	// Reload with product info
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		First(&item, "id = ?", itemID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload stock opname item: %w", err)
	}

	return &item, nil
}

// DeleteStockOpnameItem deletes a stock opname item
func (s *StockOpnameService) DeleteStockOpnameItem(ctx context.Context, companyID, tenantID, opnameID, itemID string) error {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return err
	}

	// Validate status
	if opname.Status == models.StockOpnameStatusApproved {
		return pkgerrors.NewBadRequestError("cannot delete items from approved stock opname")
	}

	if opname.Status == models.StockOpnameStatusCancelled {
		return pkgerrors.NewBadRequestError("cannot delete items from cancelled stock opname")
	}

	// Delete item
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Where("id = ? AND stock_opname_id = ?", itemID, opnameID).
		Delete(&models.StockOpnameItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete stock opname item: %w", err)
	}

	return nil
}

// ImportWarehouseProducts imports all products from a warehouse to stock opname
func (s *StockOpnameService) ImportWarehouseProducts(ctx context.Context, companyID, tenantID, opnameID string) (int, error) {
	// Get existing opname
	opname, err := s.GetStockOpname(ctx, companyID, tenantID, opnameID)
	if err != nil {
		return 0, err
	}

	// Validate status
	if opname.Status == models.StockOpnameStatusApproved {
		return 0, pkgerrors.NewBadRequestError("cannot import products to approved stock opname")
	}

	if opname.Status == models.StockOpnameStatusCancelled {
		return 0, pkgerrors.NewBadRequestError("cannot import products to cancelled stock opname")
	}

	// Get warehouse stocks
	var stocks []models.WarehouseStock
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Product").
		Where("warehouse_id = ?", opname.WarehouseID).
		Find(&stocks).Error; err != nil {
		return 0, fmt.Errorf("failed to get warehouse stocks: %w", err)
	}

	// Create items in transaction
	itemsAdded := 0
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		for _, stock := range stocks {
			// Check if product already exists in opname
			var existing models.StockOpnameItem
			if err := tx.Where("stock_opname_id = ? AND product_id = ?", opnameID, stock.ProductID).
				First(&existing).Error; err == nil {
				// Already exists, skip
				continue
			}

			// Create new item
			item := &models.StockOpnameItem{
				StockOpnameID: opnameID,
				ProductID:     stock.ProductID,
				SystemQty:     stock.Quantity,
				PhysicalQty:   decimal.Zero, // User will fill this
				DifferenceQty: decimal.Zero.Sub(stock.Quantity), // Negative until user fills actual
			}

			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("failed to create stock opname item: %w", err)
			}

			itemsAdded++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return itemsAdded, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateOpnameNumber generates a unique opname number
func (s *StockOpnameService) generateOpnameNumber(tx *gorm.DB, companyID string, opnameDate time.Time) (string, error) {
	// Format: OPN-YYYYMMDD-XXX
	prefix := fmt.Sprintf("OPN-%s-", opnameDate.Format("20060102"))

	// Get last number for the day
	var lastOpname models.StockOpname
	err := tx.Where("company_id = ? AND opname_number LIKE ?", companyID, prefix+"%").
		Order("opname_number DESC").
		First(&lastOpname).Error

	sequence := 1
	if err == nil {
		// Parse last sequence number
		var lastSeq int
		fmt.Sscanf(lastOpname.OpnameNumber, prefix+"%d", &lastSeq)
		sequence = lastSeq + 1
	} else if err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("failed to get last opname number: %w", err)
	}

	return fmt.Sprintf("%s%03d", prefix, sequence), nil
}
