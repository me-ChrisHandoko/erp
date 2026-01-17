package sales

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/document"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

type SalesOrderService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
}

func NewSalesOrderService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *SalesOrderService {
	return &SalesOrderService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

// CreateSalesOrder creates a new sales order with items
func (s *SalesOrderService) CreateSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, req *dto.CreateSalesOrderRequest) (*models.SalesOrder, error) {
	// Parse decimal fields
	subtotal, err := decimal.NewFromString(req.Subtotal)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid subtotal format")
	}

	discount, err := decimal.NewFromString(req.Discount)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid discount format")
	}

	tax, err := decimal.NewFromString(req.Tax)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid tax format")
	}

	shippingCost, err := decimal.NewFromString(req.ShippingCost)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid shippingCost format")
	}

	totalAmount, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid totalAmount format")
	}

	// Parse dates
	orderDate, err := time.Parse("2006-01-02", req.OrderDate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid orderDate format (use YYYY-MM-DD)")
	}

	var requiredDate *time.Time
	if req.RequiredDate != nil && *req.RequiredDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.RequiredDate)
		if err != nil {
			return nil, pkgerrors.NewBadRequestError("invalid requiredDate format (use YYYY-MM-DD)")
		}
		requiredDate = &parsed
	}

	var salesOrder *models.SalesOrder

	// Use transaction for atomic create
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Verify customer exists
		var customer models.Customer
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", req.CustomerId, companyID, tenantID).First(&customer).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Customer not found")
			}
			return fmt.Errorf("failed to verify customer: %w", err)
		}

		// 2. Verify warehouse exists
		var warehouse models.Warehouse
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", req.WarehouseId, companyID, tenantID).First(&warehouse).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Warehouse not found")
			}
			return fmt.Errorf("failed to verify warehouse: %w", err)
		}

		// 3. Generate SO number
		soNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypeSalesOrder)
		if err != nil {
			return fmt.Errorf("failed to generate SO number: %w", err)
		}

		// 4. Create sales order
		salesOrder = &models.SalesOrder{
			TenantID:       tenantID,
			CompanyID:      companyID,
			SONumber:       soNumber,
			SODate:         orderDate,
			RequiredDate:   requiredDate,
			CustomerID:     req.CustomerId,
			WarehouseID:    req.WarehouseId,
			Status:         models.SalesOrderStatusDraft,
			Subtotal:       subtotal,
			DiscountAmount: discount,
			TaxAmount:      tax,
			ShippingCost:   shippingCost,
			TotalAmount:    totalAmount,
			Notes:          req.Notes,
		}

		if err := tx.Create(salesOrder).Error; err != nil {
			return fmt.Errorf("failed to create sales order: %w", err)
		}

		// 5. Create sales order items
		for _, itemReq := range req.Items {
			// Parse item decimal fields
			quantity, err := decimal.NewFromString(itemReq.OrderedQty)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid quantity for product %s", itemReq.ProductId))
			}

			unitPrice, err := decimal.NewFromString(itemReq.UnitPrice)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid unitPrice for product %s", itemReq.ProductId))
			}

			itemDiscount, err := decimal.NewFromString(itemReq.Discount)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid discount for product %s", itemReq.ProductId))
			}

			lineTotal, err := decimal.NewFromString(itemReq.LineTotal)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid lineTotal for product %s", itemReq.ProductId))
			}

			// Verify product exists
			var product models.Product
			if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", itemReq.ProductId, companyID, tenantID).First(&product).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewNotFoundError(fmt.Sprintf("Product %s not found", itemReq.ProductId))
				}
				return fmt.Errorf("failed to verify product: %w", err)
			}

			item := &models.SalesOrderItem{
				SalesOrderID: salesOrder.ID,
				ProductID:    itemReq.ProductId,
				Quantity:     quantity,
				UnitPrice:    unitPrice,
				DiscountAmt:  itemDiscount,
				Subtotal:     lineTotal,
				Notes:        itemReq.Notes,
			}

			// Handle unit if specified
			if itemReq.UnitId != "" {
				var productUnit models.ProductUnit
				if err := tx.Where("id = ? AND product_id = ?", itemReq.UnitId, itemReq.ProductId).First(&productUnit).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return pkgerrors.NewNotFoundError(fmt.Sprintf("Product unit %s not found", itemReq.UnitId))
					}
					return fmt.Errorf("failed to verify product unit: %w", err)
				}
				item.ProductUnitID = &itemReq.UnitId
			}

			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("failed to create sales order item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with relations
	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrder.ID)
}

// GetSalesOrder retrieves a single sales order by ID with all relations
func (s *SalesOrderService) GetSalesOrder(ctx context.Context, companyID string, tenantID string, salesOrderID string) (*models.SalesOrder, error) {
	var salesOrder models.SalesOrder

	err := s.db.WithContext(ctx).
		Preload("Customer").
		Preload("Warehouse").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).
		First(&salesOrder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Sales order not found")
		}
		return nil, fmt.Errorf("failed to get sales order: %w", err)
	}

	return &salesOrder, nil
}

// ListSalesOrders retrieves paginated sales orders with filters
func (s *SalesOrderService) ListSalesOrders(ctx context.Context, companyID string, tenantID string, filters *dto.SalesOrderFilters) ([]models.SalesOrder, int, error) {
	// Set defaults
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.SortBy == "" {
		filters.SortBy = "orderDate"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "desc"
	}

	// Build query
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.SalesOrder{}).
		Where("company_id = ? AND tenant_id = ?", companyID, tenantID)

	// Apply filters
	if filters.Search != "" {
		query = query.Where("so_number ILIKE ?", "%"+filters.Search+"%")
	}

	if filters.Status != nil && *filters.Status != "" {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.CustomerId != "" {
		query = query.Where("customer_id = ?", filters.CustomerId)
	}

	if filters.WarehouseId != "" {
		query = query.Where("warehouse_id = ?", filters.WarehouseId)
	}

	if filters.FromDate != nil && *filters.FromDate != "" {
		fromDate, err := time.Parse("2006-01-02", *filters.FromDate)
		if err == nil {
			query = query.Where("so_date >= ?", fromDate)
		}
	}

	if filters.ToDate != nil && *filters.ToDate != "" {
		toDate, err := time.Parse("2006-01-02", *filters.ToDate)
		if err == nil {
			// Add 1 day to include the entire toDate
			toDate = toDate.Add(24 * time.Hour)
			query = query.Where("so_date < ?", toDate)
		}
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sales orders: %w", err)
	}

	// Apply sorting
	sortColumn := filters.SortBy
	if sortColumn == "orderNumber" {
		sortColumn = "so_number"
	} else if sortColumn == "orderDate" {
		sortColumn = "so_date"
	} else if sortColumn == "totalAmount" {
		sortColumn = "total_amount"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortColumn, filters.SortOrder))

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Execute query with preload
	var salesOrders []models.SalesOrder
	if err := query.Preload("Customer").Preload("Warehouse").Find(&salesOrders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list sales orders: %w", err)
	}

	return salesOrders, int(total), nil
}

// UpdateSalesOrder updates an existing sales order (only DRAFT can be edited)
func (s *SalesOrderService) UpdateSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string, req *dto.UpdateSalesOrderRequest) (*models.SalesOrder, error) {
	var salesOrder *models.SalesOrder

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Get existing sales order
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// 2. Check if order is editable (only DRAFT)
		if salesOrder.Status != models.SalesOrderStatusDraft {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot edit sales order with status %s. Only DRAFT orders can be edited.", salesOrder.Status))
		}

		// 3. Update fields
		updates := make(map[string]interface{})

		if req.CustomerId != nil {
			var customer models.Customer
			if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", *req.CustomerId, companyID, tenantID).First(&customer).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewNotFoundError("Customer not found")
				}
				return fmt.Errorf("failed to verify customer: %w", err)
			}
			updates["customer_id"] = *req.CustomerId
		}

		if req.WarehouseId != nil {
			var warehouse models.Warehouse
			if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", *req.WarehouseId, companyID, tenantID).First(&warehouse).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewNotFoundError("Warehouse not found")
				}
				return fmt.Errorf("failed to verify warehouse: %w", err)
			}
			updates["warehouse_id"] = *req.WarehouseId
		}

		if req.OrderDate != nil {
			orderDate, err := time.Parse("2006-01-02", *req.OrderDate)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid orderDate format (use YYYY-MM-DD)")
			}
			updates["so_date"] = orderDate
		}

		if req.RequiredDate != nil {
			if *req.RequiredDate == "" {
				updates["required_date"] = nil
			} else {
				requiredDate, err := time.Parse("2006-01-02", *req.RequiredDate)
				if err != nil {
					return pkgerrors.NewBadRequestError("invalid requiredDate format (use YYYY-MM-DD)")
				}
				updates["required_date"] = requiredDate
			}
		}

		if req.Notes != nil {
			updates["notes"] = req.Notes
		}

		if req.Subtotal != nil {
			subtotal, err := decimal.NewFromString(*req.Subtotal)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid subtotal format")
			}
			updates["subtotal"] = subtotal
		}

		if req.Discount != nil {
			discount, err := decimal.NewFromString(*req.Discount)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid discount format")
			}
			updates["discount_amount"] = discount
		}

		if req.Tax != nil {
			tax, err := decimal.NewFromString(*req.Tax)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid tax format")
			}
			updates["tax_amount"] = tax
		}

		if req.ShippingCost != nil {
			shippingCost, err := decimal.NewFromString(*req.ShippingCost)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid shippingCost format")
			}
			updates["shipping_cost"] = shippingCost
		}

		if req.TotalAmount != nil {
			totalAmount, err := decimal.NewFromString(*req.TotalAmount)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid totalAmount format")
			}
			updates["total_amount"] = totalAmount
		}

		// 4. Update sales order
		if len(updates) > 0 {
			if err := tx.Model(&salesOrder).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update sales order: %w", err)
			}
		}

		// 5. Update items if provided
		if req.Items != nil && len(req.Items) > 0 {
			// Delete existing items
			if err := tx.Where("sales_order_id = ?", salesOrderID).Delete(&models.SalesOrderItem{}).Error; err != nil {
				return fmt.Errorf("failed to delete existing items: %w", err)
			}

			// Create new items
			for _, itemReq := range req.Items {
				quantity, err := decimal.NewFromString(*itemReq.OrderedQty)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid quantity for product"))
				}

				unitPrice, err := decimal.NewFromString(*itemReq.UnitPrice)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid unitPrice"))
				}

				itemDiscount, err := decimal.NewFromString(*itemReq.Discount)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid discount"))
				}

				lineTotal, err := decimal.NewFromString(*itemReq.LineTotal)
				if err != nil {
					return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid lineTotal"))
				}

				item := &models.SalesOrderItem{
					SalesOrderID: salesOrderID,
					ProductID:    *itemReq.ProductId,
					Quantity:     quantity,
					UnitPrice:    unitPrice,
					DiscountAmt:  itemDiscount,
					Subtotal:     lineTotal,
					Notes:        itemReq.Notes,
				}

				if itemReq.UnitId != nil {
					item.ProductUnitID = itemReq.UnitId
				}

				if err := tx.Create(item).Error; err != nil {
					return fmt.Errorf("failed to create sales order item: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with relations
	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrderID)
}

// DeleteSalesOrder soft deletes a sales order (only DRAFT can be deleted)
func (s *SalesOrderService) DeleteSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string) error {
	return s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Get sales order
		var salesOrder models.SalesOrder
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// 2. Check if order is deletable (only DRAFT)
		if salesOrder.Status != models.SalesOrderStatusDraft {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot delete sales order with status %s. Only DRAFT orders can be deleted.", salesOrder.Status))
		}

		// 3. Delete sales order (cascade will delete items)
		if err := tx.Delete(&salesOrder).Error; err != nil {
			return fmt.Errorf("failed to delete sales order: %w", err)
		}

		return nil
	})
}

// ============================================================================
// STATUS TRANSITION OPERATIONS
// ============================================================================

// SubmitSalesOrder transitions from DRAFT to PENDING
func (s *SalesOrderService) SubmitSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string) (*models.SalesOrder, error) {
	return s.transitionStatus(ctx, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, models.SalesOrderStatusDraft, models.SalesOrderStatusPending, "Submitted sales order")
}

// ApproveSalesOrder transitions from PENDING to APPROVED
func (s *SalesOrderService) ApproveSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string) (*models.SalesOrder, error) {
	var salesOrder *models.SalesOrder

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get sales order
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// Check current status
		if salesOrder.Status != models.SalesOrderStatusPending {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot approve sales order with status %s", salesOrder.Status))
		}

		// Update status and approval info
		now := time.Now()
		updates := map[string]interface{}{
			"status":      models.SalesOrderStatusApproved,
			"approved_by": userID,
			"approved_at": now,
		}

		if err := tx.Model(&salesOrder).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to approve sales order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrderID)
}

// StartProcessingSalesOrder transitions from APPROVED to PROCESSING
func (s *SalesOrderService) StartProcessingSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string) (*models.SalesOrder, error) {
	return s.transitionStatus(ctx, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, models.SalesOrderStatusApproved, models.SalesOrderStatusProcessing, "Started processing sales order")
}

// ShipSalesOrder transitions from PROCESSING to SHIPPED
func (s *SalesOrderService) ShipSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string, req *dto.ShipSalesOrderRequest) (*models.SalesOrder, error) {
	var salesOrder *models.SalesOrder

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get sales order
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// Check current status
		if salesOrder.Status != models.SalesOrderStatusProcessing {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot ship sales order with status %s", salesOrder.Status))
		}

		// Update status and delivery info
		updates := map[string]interface{}{
			"status": models.SalesOrderStatusShipped,
		}

		if req.DeliveryDate != nil && *req.DeliveryDate != "" {
			deliveryDate, err := time.Parse("2006-01-02", *req.DeliveryDate)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid deliveryDate format (use YYYY-MM-DD)")
			}
			updates["delivery_date"] = deliveryDate
		}

		if req.DeliveryAddress != nil {
			updates["delivery_address"] = *req.DeliveryAddress
		}

		if req.Notes != nil {
			updates["notes"] = *req.Notes
		}

		if err := tx.Model(&salesOrder).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to ship sales order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrderID)
}

// DeliverSalesOrder transitions from SHIPPED to DELIVERED
func (s *SalesOrderService) DeliverSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string, req *dto.DeliverSalesOrderRequest) (*models.SalesOrder, error) {
	var salesOrder *models.SalesOrder

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get sales order
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// Check current status
		if salesOrder.Status != models.SalesOrderStatusShipped {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot mark as delivered sales order with status %s", salesOrder.Status))
		}

		// Update status and delivery info
		updates := map[string]interface{}{
			"status": models.SalesOrderStatusDelivered,
		}

		if req.DeliveryDate != nil && *req.DeliveryDate != "" {
			deliveryDate, err := time.Parse("2006-01-02", *req.DeliveryDate)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid deliveryDate format (use YYYY-MM-DD)")
			}
			updates["delivery_date"] = deliveryDate
		} else {
			// Set to now if not provided
			now := time.Now()
			updates["delivery_date"] = now
		}

		if req.Notes != nil {
			updates["notes"] = *req.Notes
		}

		if err := tx.Model(&salesOrder).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to deliver sales order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrderID)
}

// CompleteSalesOrder transitions from DELIVERED to COMPLETED
func (s *SalesOrderService) CompleteSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string) (*models.SalesOrder, error) {
	return s.transitionStatus(ctx, companyID, tenantID, userID, ipAddress, userAgent, salesOrderID, models.SalesOrderStatusDelivered, models.SalesOrderStatusCompleted, "Completed sales order")
}

// CancelSalesOrder transitions to CANCELLED from any non-final status
func (s *SalesOrderService) CancelSalesOrder(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string, req *dto.CancelSalesOrderRequest) (*models.SalesOrder, error) {
	var salesOrder *models.SalesOrder

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get sales order
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// Check if cancellable (cannot cancel COMPLETED or already CANCELLED)
		if salesOrder.Status == models.SalesOrderStatusCompleted || salesOrder.Status == models.SalesOrderStatusCancelled {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot cancel sales order with status %s", salesOrder.Status))
		}

		// Update status and cancellation info
		now := time.Now()
		updates := map[string]interface{}{
			"status":            models.SalesOrderStatusCancelled,
			"cancelled_by":      userID,
			"cancelled_at":      now,
			"cancellation_note": req.Reason,
		}

		if err := tx.Model(&salesOrder).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to cancel sales order: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrderID)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// transitionStatus is a helper function for simple status transitions
func (s *SalesOrderService) transitionStatus(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, salesOrderID string, fromStatus models.SalesOrderStatus, toStatus models.SalesOrderStatus, auditMessage string) (*models.SalesOrder, error) {
	var salesOrder *models.SalesOrder

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get sales order
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ?", salesOrderID, companyID, tenantID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to get sales order: %w", err)
		}

		// Check current status
		if salesOrder.Status != fromStatus {
			return pkgerrors.NewBadRequestError(fmt.Sprintf("Cannot transition from %s to %s. Current status is %s", fromStatus, toStatus, salesOrder.Status))
		}

		// Update status
		if err := tx.Model(&salesOrder).Update("status", toStatus).Error; err != nil {
			return fmt.Errorf("failed to update sales order status: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.GetSalesOrder(ctx, companyID, tenantID, salesOrderID)
}
