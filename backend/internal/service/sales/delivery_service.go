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

type DeliveryService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
}

func NewDeliveryService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator) *DeliveryService {
	return &DeliveryService{
		db:           db,
		docNumberGen: docNumberGen,
	}
}

// ============================================================================
// CRUD OPERATIONS
// ============================================================================

// CreateDelivery creates a new delivery with items
func (s *DeliveryService) CreateDelivery(ctx context.Context, companyID string, tenantID string, userID string, ipAddress string, userAgent string, req *dto.CreateDeliveryRequest) (*models.Delivery, error) {
	// Parse delivery date
	deliveryDate, err := time.Parse("2006-01-02", req.DeliveryDate)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid deliveryDate format (use YYYY-MM-DD)")
	}

	// Validate delivery type
	var deliveryType models.DeliveryType
	switch req.Type {
	case "NORMAL":
		deliveryType = models.DeliveryTypeNormal
	case "RETURN":
		deliveryType = models.DeliveryTypeReturn
	case "REPLACEMENT":
		deliveryType = models.DeliveryTypeReplacement
	default:
		return nil, pkgerrors.NewBadRequestError("invalid delivery type")
	}

	var delivery *models.Delivery

	// Use transaction for atomic create
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Verify sales order exists and is APPROVED
		var salesOrder models.SalesOrder
		if err := tx.Where("id = ? AND company_id = ?", req.SalesOrderId, companyID).First(&salesOrder).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Sales order not found")
			}
			return fmt.Errorf("failed to verify sales order: %w", err)
		}

		if salesOrder.Status != models.SalesOrderStatusApproved {
			return pkgerrors.NewBadRequestError("sales order must be APPROVED before creating delivery")
		}

		// 2. Verify customer exists
		var customer models.Customer
		if err := tx.Where("id = ? AND company_id = ?", req.CustomerId, companyID).First(&customer).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Customer not found")
			}
			return fmt.Errorf("failed to verify customer: %w", err)
		}

		// 3. Verify warehouse exists
		var warehouse models.Warehouse
		if err := tx.Where("id = ? AND company_id = ?", req.WarehouseId, companyID).First(&warehouse).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Warehouse not found")
			}
			return fmt.Errorf("failed to verify warehouse: %w", err)
		}

		// 4. Generate delivery number
		deliveryNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypeDelivery)
		if err != nil {
			return fmt.Errorf("failed to generate delivery number: %w", err)
		}

		// 5. Create delivery
		delivery = &models.Delivery{
			TenantID:          tenantID,
			CompanyID:         companyID,
			DeliveryNumber:    deliveryNumber,
			DeliveryDate:      deliveryDate,
			SalesOrderID:      req.SalesOrderId,
			WarehouseID:       req.WarehouseId,
			CustomerID:        req.CustomerId,
			Type:              deliveryType,
			Status:            models.DeliveryStatusPrepared,
			DeliveryAddress:   req.DeliveryAddress,
			DriverName:        req.DriverName,
			VehicleNumber:     req.VehicleNumber,
			ExpeditionService: req.ExpeditionService,
			TTNKNumber:        req.TtnkNumber,
			Notes:             req.Notes,
		}

		if err := tx.Create(delivery).Error; err != nil {
			return fmt.Errorf("failed to create delivery: %w", err)
		}

		// 6. Create delivery items
		for _, itemReq := range req.Items {
			// Parse quantity
			quantity, err := decimal.NewFromString(itemReq.Quantity)
			if err != nil {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid quantity for product %s", itemReq.ProductId))
			}

			// Verify product exists
			var product models.Product
			if err := tx.Where("id = ? AND company_id = ?", itemReq.ProductId, companyID).First(&product).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return pkgerrors.NewNotFoundError(fmt.Sprintf("Product %s not found", itemReq.ProductId))
				}
				return fmt.Errorf("failed to verify product: %w", err)
			}

			// Verify sales order item if provided
			var salesOrderItemID string
			if itemReq.SalesOrderItemId != nil && *itemReq.SalesOrderItemId != "" {
				var soItem models.SalesOrderItem
				if err := tx.Where("id = ? AND sales_order_id = ?", *itemReq.SalesOrderItemId, req.SalesOrderId).First(&soItem).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return pkgerrors.NewNotFoundError(fmt.Sprintf("Sales order item %s not found", *itemReq.SalesOrderItemId))
					}
					return fmt.Errorf("failed to verify sales order item: %w", err)
				}
				salesOrderItemID = *itemReq.SalesOrderItemId
			} else {
				// Find matching sales order item by product
				var soItem models.SalesOrderItem
				if err := tx.Where("sales_order_id = ? AND product_id = ?", req.SalesOrderId, itemReq.ProductId).First(&soItem).Error; err != nil {
					return pkgerrors.NewNotFoundError(fmt.Sprintf("Sales order item for product %s not found", itemReq.ProductId))
				}
				salesOrderItemID = soItem.ID
			}

			// Verify product unit if provided
			if itemReq.ProductUnitId != nil && *itemReq.ProductUnitId != "" {
				var productUnit models.ProductUnit
				if err := tx.Where("id = ? AND product_id = ?", *itemReq.ProductUnitId, itemReq.ProductId).First(&productUnit).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return pkgerrors.NewNotFoundError(fmt.Sprintf("Product unit %s not found", *itemReq.ProductUnitId))
					}
					return fmt.Errorf("failed to verify product unit: %w", err)
				}
			}

			// Create delivery item
			deliveryItem := &models.DeliveryItem{
				DeliveryID:       delivery.ID,
				SalesOrderItemID: salesOrderItemID,
				ProductID:        itemReq.ProductId,
				ProductUnitID:    itemReq.ProductUnitId,
				BatchID:          itemReq.BatchId,
				Quantity:         quantity,
				Notes:            itemReq.Notes,
			}

			if err := tx.Create(deliveryItem).Error; err != nil {
				return fmt.Errorf("failed to create delivery item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload delivery with relations
	return s.GetDeliveryByID(ctx, companyID, tenantID, delivery.ID)
}

// GetDeliveryByID retrieves a delivery by ID with full relations
func (s *DeliveryService) GetDeliveryByID(ctx context.Context, companyID string, tenantID string, deliveryID string) (*models.Delivery, error) {
	var delivery models.Delivery

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("SalesOrder").
		Preload("SalesOrder.Customer").
		Preload("Customer").
		Preload("Warehouse").
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		Preload("Items.Batch").
		Where("company_id = ? AND id = ?", companyID, deliveryID).
		First(&delivery).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgerrors.NewNotFoundError("Delivery not found")
		}
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	return &delivery, nil
}

// ListDeliveries retrieves deliveries with pagination and filters
func (s *DeliveryService) ListDeliveries(ctx context.Context, companyID string, tenantID string, filters dto.DeliveryFilters) ([]models.Delivery, int64, error) {
	var deliveries []models.Delivery
	var total int64

	// Build query with tenant context set for GORM callbacks
	// tenantID is explicitly passed from handler (no bypass needed!)
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.Delivery{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if filters.Status != nil && *filters.Status != "" {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.Type != nil && *filters.Type != "" {
		query = query.Where("type = ?", *filters.Type)
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
			query = query.Where("delivery_date >= ?", fromDate)
		}
	}

	if filters.ToDate != nil && *filters.ToDate != "" {
		toDate, err := time.Parse("2006-01-02", *filters.ToDate)
		if err == nil {
			query = query.Where("delivery_date <= ?", toDate)
		}
	}

	// Search
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Joins("LEFT JOIN customers ON customers.id = deliveries.customer_id").
			Where("deliveries.delivery_number LIKE ? OR customers.name LIKE ?", searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count deliveries: %w", err)
	}

	// Sorting
	sortBy := "delivery_date"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "deliveryNumber":
			sortBy = "delivery_number"
		case "deliveryDate":
			sortBy = "delivery_date"
		}
	}

	sortOrder := "desc"
	if filters.SortOrder != "" && filters.SortOrder == "asc" {
		sortOrder = "asc"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}

	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Execute query with preloads
	err := query.
		Preload("SalesOrder").
		Preload("Customer").
		Preload("Warehouse").
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.ProductUnit").
		Preload("Items.Batch").
		Limit(limit).
		Offset(offset).
		Find(&deliveries).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list deliveries: %w", err)
	}

	return deliveries, total, nil
}

// ============================================================================
// STATUS MANAGEMENT
// ============================================================================

// UpdateDeliveryStatus updates delivery status and related fields
func (s *DeliveryService) UpdateDeliveryStatus(ctx context.Context, companyID string, tenantID string, deliveryID string, req *dto.UpdateDeliveryStatusRequest) (*models.Delivery, error) {
	var delivery *models.Delivery

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get delivery
		if err := tx.Where("id = ? AND company_id = ?", deliveryID, companyID).First(&delivery).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Delivery not found")
			}
			return fmt.Errorf("failed to get delivery: %w", err)
		}

		// Check if delivery can be updated
		if delivery.Status == models.DeliveryStatusCancelled {
			return pkgerrors.NewBadRequestError("cannot update cancelled delivery")
		}

		if delivery.Status == models.DeliveryStatusConfirmed {
			return pkgerrors.NewBadRequestError("cannot update confirmed delivery")
		}

		updates := make(map[string]interface{})

		// Update status if provided
		if req.Status != nil && *req.Status != "" {
			var newStatus models.DeliveryStatus
			switch *req.Status {
			case "PREPARED":
				newStatus = models.DeliveryStatusPrepared
			case "IN_TRANSIT":
				newStatus = models.DeliveryStatusInTransit
			case "DELIVERED":
				newStatus = models.DeliveryStatusDelivered
			case "CONFIRMED":
				newStatus = models.DeliveryStatusConfirmed
			case "CANCELLED":
				newStatus = models.DeliveryStatusCancelled
			default:
				return pkgerrors.NewBadRequestError("invalid status")
			}

			// Validate status transition
			if !s.isValidStatusTransition(delivery.Status, newStatus) {
				return pkgerrors.NewBadRequestError(fmt.Sprintf("invalid status transition from %s to %s", delivery.Status, newStatus))
			}

			updates["status"] = newStatus
		}

		// Update timestamps
		if req.DepartureTime != nil && *req.DepartureTime != "" {
			departureTime, err := time.Parse(time.RFC3339, *req.DepartureTime)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid departureTime format (use RFC3339)")
			}
			updates["departure_time"] = departureTime
		}

		if req.ArrivalTime != nil && *req.ArrivalTime != "" {
			arrivalTime, err := time.Parse(time.RFC3339, *req.ArrivalTime)
			if err != nil {
				return pkgerrors.NewBadRequestError("invalid arrivalTime format (use RFC3339)")
			}
			updates["arrival_time"] = arrivalTime
		}

		// Update POD fields
		if req.ReceivedBy != nil {
			updates["received_by"] = req.ReceivedBy
			if *req.ReceivedBy != "" {
				now := time.Now()
				updates["received_at"] = &now
			}
		}

		if req.SignatureUrl != nil {
			updates["signature_url"] = req.SignatureUrl
		}

		if req.PhotoUrl != nil {
			updates["photo_url"] = req.PhotoUrl
		}

		// Apply updates
		if len(updates) > 0 {
			if err := tx.Model(&delivery).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update delivery: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload delivery with relations
	return s.GetDeliveryByID(ctx, companyID, tenantID, deliveryID)
}

// StartDelivery moves delivery from PREPARED to IN_TRANSIT
func (s *DeliveryService) StartDelivery(ctx context.Context, companyID string, tenantID string, deliveryID string, req *dto.StartDeliveryRequest) (*models.Delivery, error) {
	// Parse departure time
	departureTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		return nil, pkgerrors.NewBadRequestError("invalid departureTime format (use RFC3339)")
	}

	var delivery *models.Delivery

	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get delivery
		if err := tx.Where("id = ? AND company_id = ?", deliveryID, companyID).First(&delivery).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Delivery not found")
			}
			return fmt.Errorf("failed to get delivery: %w", err)
		}

		// Validate status
		if delivery.Status != models.DeliveryStatusPrepared {
			return pkgerrors.NewBadRequestError("delivery must be in PREPARED status to start")
		}

		// Update status and departure time
		updates := map[string]interface{}{
			"status":         models.DeliveryStatusInTransit,
			"departure_time": departureTime,
		}

		if err := tx.Model(&delivery).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to start delivery: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload delivery with relations
	return s.GetDeliveryByID(ctx, companyID, tenantID, deliveryID)
}

// CompleteDelivery moves delivery from IN_TRANSIT to DELIVERED
func (s *DeliveryService) CompleteDelivery(ctx context.Context, companyID string, tenantID string, deliveryID string, req *dto.CompleteDeliveryRequest) (*models.Delivery, error) {
	var delivery *models.Delivery

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get delivery
		if err := tx.Where("id = ? AND company_id = ?", deliveryID, companyID).First(&delivery).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Delivery not found")
			}
			return fmt.Errorf("failed to get delivery: %w", err)
		}

		// Validate status
		if delivery.Status != models.DeliveryStatusInTransit {
			return pkgerrors.NewBadRequestError("delivery must be in IN_TRANSIT status to complete")
		}

		// Update status and POD fields
		now := time.Now()
		updates := map[string]interface{}{
			"status":       models.DeliveryStatusDelivered,
			"arrival_time": now,
		}

		if req.ReceivedBy != nil && *req.ReceivedBy != "" {
			updates["received_by"] = req.ReceivedBy
			updates["received_at"] = now
		}

		if req.SignatureUrl != nil && *req.SignatureUrl != "" {
			updates["signature_url"] = req.SignatureUrl
		}

		if req.PhotoUrl != nil && *req.PhotoUrl != "" {
			updates["photo_url"] = req.PhotoUrl
		}

		if err := tx.Model(&delivery).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to complete delivery: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload delivery with relations
	return s.GetDeliveryByID(ctx, companyID, tenantID, deliveryID)
}

// ConfirmDelivery moves delivery from DELIVERED to CONFIRMED
func (s *DeliveryService) ConfirmDelivery(ctx context.Context, companyID string, tenantID string, deliveryID string) (*models.Delivery, error) {
	var delivery *models.Delivery

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get delivery
		if err := tx.Where("id = ? AND company_id = ?", deliveryID, companyID).First(&delivery).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Delivery not found")
			}
			return fmt.Errorf("failed to get delivery: %w", err)
		}

		// Validate status
		if delivery.Status != models.DeliveryStatusDelivered {
			return pkgerrors.NewBadRequestError("delivery must be in DELIVERED status to confirm")
		}

		// Update status
		if err := tx.Model(&delivery).Update("status", models.DeliveryStatusConfirmed).Error; err != nil {
			return fmt.Errorf("failed to confirm delivery: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload delivery with relations
	return s.GetDeliveryByID(ctx, companyID, tenantID, deliveryID)
}

// CancelDelivery cancels a delivery
func (s *DeliveryService) CancelDelivery(ctx context.Context, companyID string, tenantID string, deliveryID string, req *dto.CancelDeliveryRequest) (*models.Delivery, error) {
	var delivery *models.Delivery

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Get delivery
		if err := tx.Where("id = ? AND company_id = ?", deliveryID, companyID).First(&delivery).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkgerrors.NewNotFoundError("Delivery not found")
			}
			return fmt.Errorf("failed to get delivery: %w", err)
		}

		// Check if delivery can be cancelled
		if delivery.Status == models.DeliveryStatusCancelled {
			return pkgerrors.NewBadRequestError("delivery is already cancelled")
		}

		if delivery.Status == models.DeliveryStatusConfirmed {
			return pkgerrors.NewBadRequestError("cannot cancel confirmed delivery")
		}

		// Update status and notes
		updates := map[string]interface{}{
			"status": models.DeliveryStatusCancelled,
		}

		if req.Notes != nil && *req.Notes != "" {
			updates["notes"] = req.Notes
		}

		if err := tx.Model(&delivery).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to cancel delivery: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload delivery with relations
	return s.GetDeliveryByID(ctx, companyID, tenantID, deliveryID)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// isValidStatusTransition checks if status transition is valid
func (s *DeliveryService) isValidStatusTransition(from, to models.DeliveryStatus) bool {
	validTransitions := map[models.DeliveryStatus][]models.DeliveryStatus{
		models.DeliveryStatusPrepared: {
			models.DeliveryStatusPrepared,
			models.DeliveryStatusInTransit,
			models.DeliveryStatusCancelled,
		},
		models.DeliveryStatusInTransit: {
			models.DeliveryStatusInTransit,
			models.DeliveryStatusDelivered,
			models.DeliveryStatusCancelled,
		},
		models.DeliveryStatusDelivered: {
			models.DeliveryStatusDelivered,
			models.DeliveryStatusConfirmed,
			models.DeliveryStatusCancelled,
		},
		models.DeliveryStatusConfirmed: {
			models.DeliveryStatusConfirmed,
		},
		models.DeliveryStatusCancelled: {
			models.DeliveryStatusCancelled,
		},
	}

	allowedStatuses, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == to {
			return true
		}
	}

	return false
}
