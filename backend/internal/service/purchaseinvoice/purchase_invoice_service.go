package purchaseinvoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"backend/internal/dto"
	"backend/internal/service/audit"
	"backend/internal/service/document"
	"backend/models"
)

// PurchaseInvoiceService handles business logic for purchase invoices
type PurchaseInvoiceService struct {
	db           *gorm.DB
	docNumberGen *document.DocumentNumberGenerator
	auditService *audit.AuditService
}

// NewPurchaseInvoiceService creates a new purchase invoice service
func NewPurchaseInvoiceService(db *gorm.DB, docNumberGen *document.DocumentNumberGenerator, auditService *audit.AuditService) *PurchaseInvoiceService {
	return &PurchaseInvoiceService{
		db:           db,
		docNumberGen: docNumberGen,
		auditService: auditService,
	}
}

// ============================================================================
// LIST & GET Operations
// ============================================================================

// ListPurchaseInvoices retrieves purchase invoices with filters and pagination
func (s *PurchaseInvoiceService) ListPurchaseInvoices(
	ctx context.Context,
	tenantID, companyID string,
	filters dto.PurchaseInvoiceFilters,
) ([]models.PurchaseInvoice, dto.PaginationResponse, error) {
	var invoices []models.PurchaseInvoice
	var total int64

	// Base query with tenant context set for GORM callbacks
	query := s.db.WithContext(ctx).Set("tenant_id", tenantID).Model(&models.PurchaseInvoice{}).
		Where("company_id = ?", companyID)

	// Apply filters
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			"invoice_number ILIKE ? OR supplier_name ILIKE ?",
			searchPattern, searchPattern,
		)
	}

	if filters.SupplierID != "" {
		query = query.Where("supplier_id = ?", filters.SupplierID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.PaymentStatus != "" {
		query = query.Where("payment_status = ?", filters.PaymentStatus)
	}

	// Date range filters
	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err == nil {
			query = query.Where("invoice_date >= ?", dateFrom)
		}
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err == nil {
			query = query.Where("invoice_date <= ?", dateTo)
		}
	}

	if filters.DueDateFrom != "" {
		dueDateFrom, err := time.Parse("2006-01-02", filters.DueDateFrom)
		if err == nil {
			query = query.Where("due_date >= ?", dueDateFrom)
		}
	}

	if filters.DueDateTo != "" {
		dueDateTo, err := time.Parse("2006-01-02", filters.DueDateTo)
		if err == nil {
			query = query.Where("due_date <= ?", dueDateTo)
		}
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "invoiceNumber"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "asc"
	}

	// Map camelCase to snake_case for database
	sortField := toSnakeCase(sortBy)
	query = query.Order(fmt.Sprintf("%s %s", sortField, sortOrder))

	// Apply pagination
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&invoices).Error; err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Calculate total pages
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	pagination := dto.PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return invoices, pagination, nil
}

// GetPurchaseInvoice retrieves a single purchase invoice by ID with relations
func (s *PurchaseInvoiceService) GetPurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID string,
) (*models.PurchaseInvoice, error) {
	var invoice models.PurchaseInvoice

	err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Items").
		Preload("Payments").
		Preload("Supplier").
		Preload("PurchaseOrder").
		Preload("GoodsReceipt").
		Where("id = ? AND company_id = ?", invoiceID, companyID).
		First(&invoice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("purchase invoice not found")
		}
		return nil, err
	}

	return &invoice, nil
}

// ============================================================================
// CREATE Operation
// ============================================================================

// CreatePurchaseInvoice creates a new purchase invoice
func (s *PurchaseInvoiceService) CreatePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, userID, ipAddress, userAgent string,
	req dto.CreatePurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Parse dates
	invoiceDate, err := time.Parse("2006-01-02", req.InvoiceDate)
	if err != nil {
		return nil, errors.New("invalid invoice date format")
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, errors.New("invalid due date format")
	}

	// Validate due date is after invoice date
	if dueDate.Before(invoiceDate) {
		return nil, errors.New("due date must be after invoice date")
	}

	// Generate invoice number using document number generator
	invoiceNumber, err := s.docNumberGen.GenerateNumber(ctx, tenantID, companyID, document.DocTypePurchaseInvoice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Get supplier information for denormalization
	var supplier models.Supplier
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).First(&supplier, "id = ?", req.SupplierID).Error; err != nil {
		return nil, errors.New("supplier not found")
	}

	// Get company settings for invoice control policy and tolerance
	var company models.Company
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).First(&company, "id = ?", companyID).Error; err != nil {
		return nil, errors.New("company not found")
	}

	// Parse tax rate
	taxRate := decimal.NewFromFloat(11.0) // Default 11%
	if req.TaxRate != "" {
		taxRate, err = decimal.NewFromString(req.TaxRate)
		if err != nil {
			return nil, errors.New("invalid tax rate format")
		}
	}

	// Parse discount
	discountAmount := decimal.Zero
	if req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
	}

	// Calculate payment term days from due date - invoice date
	// This ensures paymentTermDays is consistent with the actual dates provided
	paymentTermDays := int(dueDate.Sub(invoiceDate).Hours() / 24)
	if paymentTermDays < 0 {
		paymentTermDays = 0
	}
	// Override with request value if explicitly provided
	if req.PaymentTermDays > 0 {
		paymentTermDays = req.PaymentTermDays
	}

	// Parse non-goods costs (biaya tambahan)
	shippingCost := decimal.Zero
	if req.ShippingCost != "" {
		shippingCost, err = decimal.NewFromString(req.ShippingCost)
		if err != nil {
			return nil, errors.New("invalid shipping cost format")
		}
	}

	handlingCost := decimal.Zero
	if req.HandlingCost != "" {
		handlingCost, err = decimal.NewFromString(req.HandlingCost)
		if err != nil {
			return nil, errors.New("invalid handling cost format")
		}
	}

	otherCost := decimal.Zero
	if req.OtherCost != "" {
		otherCost, err = decimal.NewFromString(req.OtherCost)
		if err != nil {
			return nil, errors.New("invalid other cost format")
		}
	}

	var invoice *models.PurchaseInvoice

	// Use transaction for atomic create
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 0. Pre-load PO for validation
		var purchaseOrder *models.PurchaseOrder

		// Load Purchase Order if provided
		if req.PurchaseOrderID != nil && *req.PurchaseOrderID != "" {
			purchaseOrder = &models.PurchaseOrder{}
			if err := tx.Preload("Items").First(purchaseOrder, "id = ?", *req.PurchaseOrderID).Error; err != nil {
				return errors.New("purchase order not found")
			}
		}

		// Validate Goods Receipt exists if provided
		if req.GoodsReceiptID != nil && *req.GoodsReceiptID != "" {
			var count int64
			if err := tx.Model(&models.GoodsReceipt{}).Where("id = ?", *req.GoodsReceiptID).Count(&count).Error; err != nil || count == 0 {
				return errors.New("goods receipt not found")
			}
		}

		// Validate invoice quantities against PO if PO reference exists
		if purchaseOrder != nil {

			// Build a map of PO items for quick lookup
			poItemMap := make(map[string]*models.PurchaseOrderItem)
			for i := range purchaseOrder.Items {
				poItemMap[purchaseOrder.Items[i].ID] = &purchaseOrder.Items[i]
			}

			// Validate each invoice item against PO item's remaining qty
			// Apply company settings: InvoiceControlPolicy and InvoiceTolerancePct
			for _, itemReq := range req.Items {
				if itemReq.PurchaseOrderItemID != nil && *itemReq.PurchaseOrderItemID != "" {
					poItem, exists := poItemMap[*itemReq.PurchaseOrderItemID]
					if !exists {
						return fmt.Errorf("PO item not found: %s", *itemReq.PurchaseOrderItemID)
					}

					// Parse invoice quantity
					invoiceQty, err := decimal.NewFromString(itemReq.Quantity)
					if err != nil {
						return errors.New("invalid quantity format")
					}

					// Determine max invoiceable qty based on control policy
					var maxInvoiceableQty decimal.Decimal
					if company.InvoiceControlPolicy == models.InvoiceControlPolicyReceived {
						// Three-way matching: can only invoice qty that has been received (GRN)
						maxInvoiceableQty = poItem.ReceivedQty.Sub(poItem.InvoicedQty)
						if maxInvoiceableQty.LessThan(decimal.Zero) {
							maxInvoiceableQty = decimal.Zero
						}
					} else {
						// Default (ORDERED): can invoice based on PO qty
						maxInvoiceableQty = poItem.GetRemainingQtyToInvoice()
					}

					// Apply tolerance percentage if set
					if company.InvoiceTolerancePct.GreaterThan(decimal.Zero) {
						// Calculate tolerance: maxQty * (1 + tolerancePct/100)
						toleranceMultiplier := decimal.NewFromInt(1).Add(company.InvoiceTolerancePct.Div(decimal.NewFromInt(100)))
						maxInvoiceableQty = maxInvoiceableQty.Mul(toleranceMultiplier)
					}

					// Check if invoice qty exceeds max allowed qty
					if invoiceQty.GreaterThan(maxInvoiceableQty) {
						policyDesc := "PO"
						if company.InvoiceControlPolicy == models.InvoiceControlPolicyReceived {
							policyDesc = "received (GRN)"
						}
						return fmt.Errorf("invoice quantity (%.3f) exceeds max invoiceable %s quantity (%.3f) for product %s",
							invoiceQty.InexactFloat64(), policyDesc, maxInvoiceableQty.InexactFloat64(), poItem.ProductID)
					}
				}
			}
		}

		// 1. Create purchase invoice
		invoice = &models.PurchaseInvoice{
			TenantID:             tenantID,
			CompanyID:            companyID,
			InvoiceNumber:        invoiceNumber, // Use auto-generated number
			InvoiceDate:          invoiceDate,
			DueDate:              dueDate,
			SupplierID:           req.SupplierID,
			SupplierName:         supplier.Name,
			SupplierCode:         &supplier.Code,
			PurchaseOrderID:      req.PurchaseOrderID,
			GoodsReceiptID:       req.GoodsReceiptID,
			DiscountAmount:       discountAmount,
			TaxRate:              taxRate,
			PaymentTermDays:      paymentTermDays,
			Notes:                req.Notes,
			ShippingCost:         shippingCost,
			HandlingCost:         handlingCost,
			OtherCost:            otherCost,
			OtherCostDescription: req.OtherCostDescription,
			Status:               models.PurchaseInvoiceStatusDraft,
			PaymentStatus:        models.PaymentStatusUnpaid,
			CreatedBy:            userID,
		}

		if err := tx.Create(invoice).Error; err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		// 2. Create invoice items and track PO item quantities
		poItemUpdates := make(map[string]decimal.Decimal) // PO Item ID -> qty to add to InvoicedQty
		for _, itemReq := range req.Items {
			item, err := s.createInvoiceItem(tx, invoice, itemReq)
			if err != nil {
				return fmt.Errorf("failed to create invoice item: %w", err)
			}
			invoice.Items = append(invoice.Items, *item)

			// Track quantity to update on PO item
			if itemReq.PurchaseOrderItemID != nil && *itemReq.PurchaseOrderItemID != "" {
				qty, _ := decimal.NewFromString(itemReq.Quantity)
				if existing, ok := poItemUpdates[*itemReq.PurchaseOrderItemID]; ok {
					poItemUpdates[*itemReq.PurchaseOrderItemID] = existing.Add(qty)
				} else {
					poItemUpdates[*itemReq.PurchaseOrderItemID] = qty
				}
			}
		}

		// 3. Calculate totals
		invoice.CalculateTotals()

		// 4. Update invoice with calculated totals
		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to save invoice totals: %w", err)
		}

		// 5. Update PO item InvoicedQty and PO InvoiceStatus
		if purchaseOrder != nil && len(poItemUpdates) > 0 {
			for poItemID, addQty := range poItemUpdates {
				// Update InvoicedQty on PO item
				if err := tx.Model(&models.PurchaseOrderItem{}).
					Where("id = ?", poItemID).
					Update("invoiced_qty", gorm.Expr("invoiced_qty + ?", addQty)).Error; err != nil {
					return fmt.Errorf("failed to update PO item invoiced qty: %w", err)
				}
			}

			// Reload PO items to recalculate invoice status
			if err := tx.Preload("Items").First(purchaseOrder, "id = ?", purchaseOrder.ID).Error; err != nil {
				return fmt.Errorf("failed to reload purchase order: %w", err)
			}

			// Update PO invoice status
			purchaseOrder.UpdateInvoiceStatus()
			if err := tx.Model(purchaseOrder).Update("invoice_status", purchaseOrder.InvoiceStatus).Error; err != nil {
				return fmt.Errorf("failed to update PO invoice status: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload with relations
	s.db.WithContext(ctx).Set("tenant_id", tenantID).
		Preload("Items").
		First(invoice, invoice.ID)

	// Log audit trail for purchase invoice creation
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

		// Prepare audit data including header and items
		auditData := map[string]interface{}{
			"invoice_number":         invoice.InvoiceNumber,
			"invoice_date":           invoice.InvoiceDate.Format("2006-01-02"),
			"due_date":               invoice.DueDate.Format("2006-01-02"),
			"supplier_id":            invoice.SupplierID,
			"supplier_name":          invoice.SupplierName,
			"supplier_code":          invoice.SupplierCode,
			"purchase_order_id":      invoice.PurchaseOrderID,
			"goods_receipt_id":       invoice.GoodsReceiptID,
			"discount_amount":        invoice.DiscountAmount.String(),
			"tax_rate":               invoice.TaxRate.String(),
			"tax_amount":             invoice.TaxAmount.String(),
			"subtotal_amount":        invoice.SubtotalAmount.String(),
			"total_amount":           invoice.TotalAmount.String(),
			"shipping_cost":          invoice.ShippingCost.String(),
			"handling_cost":          invoice.HandlingCost.String(),
			"other_cost":             invoice.OtherCost.String(),
			"other_cost_description": invoice.OtherCostDescription,
			"payment_term_days":      invoice.PaymentTermDays,
			"notes":                  invoice.Notes,
			"status":                 string(invoice.Status),
			"payment_status":         string(invoice.PaymentStatus),
		}

		// Add items to audit data
		items := make([]map[string]interface{}, len(invoice.Items))
		for i, item := range invoice.Items {
			items[i] = map[string]interface{}{
				"product_id":      item.ProductID,
				"product_code":    item.ProductCode,
				"product_name":    item.ProductName,
				"unit_id":         item.UnitID,
				"unit_name":       item.UnitName,
				"quantity":        item.Quantity.String(),
				"unit_price":      item.UnitPrice.String(),
				"discount_amount": item.DiscountAmount.String(),
				"discount_pct":    item.DiscountPct.String(),
				"tax_amount":      item.TaxAmount.String(),
				"line_total":      item.LineTotal.String(),
				"notes":           item.Notes,
			}
		}
		auditData["items"] = items

		if err := s.auditService.LogPurchaseInvoiceCreated(ctx, auditCtx, invoice.ID, auditData); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice creation: %v\n", err)
		}
	}

	return invoice, nil
}

// createInvoiceItem helper function to create an invoice item
func (s *PurchaseInvoiceService) createInvoiceItem(
	tx *gorm.DB,
	invoice *models.PurchaseInvoice,
	req dto.CreatePurchaseInvoiceItemRequest,
) (*models.PurchaseInvoiceItem, error) {
	// Get product information for denormalization
	var product models.Product
	if err := tx.Preload("Units").First(&product, "id = ?", req.ProductID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Get unit information - if UnitID not provided, use product's base unit
	var unit models.ProductUnit
	if req.UnitID != "" {
		if err := tx.First(&unit, "id = ?", req.UnitID).Error; err != nil {
			return nil, errors.New("product unit not found")
		}
	} else {
		// Look up the product's base unit from preloaded units
		found := false
		for _, u := range product.Units {
			if u.IsBaseUnit {
				unit = u
				found = true
				break
			}
		}
		if !found {
			// Fallback to database query if not in preloaded units
			if err := tx.Where("product_id = ? AND is_base_unit = true", req.ProductID).First(&unit).Error; err != nil {
				return nil, errors.New("product base unit not found")
			}
		}
	}

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, errors.New("invalid quantity format")
	}
	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("quantity must be greater than 0")
	}

	// Parse unit price
	unitPrice, err := decimal.NewFromString(req.UnitPrice)
	if err != nil {
		return nil, errors.New("invalid unit price format")
	}
	if unitPrice.LessThan(decimal.Zero) {
		return nil, errors.New("unit price cannot be negative")
	}

	// Parse optional fields
	discountAmount := decimal.Zero
	if req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
	}

	discountPct := decimal.Zero
	if req.DiscountPct != "" {
		discountPct, err = decimal.NewFromString(req.DiscountPct)
		if err != nil {
			return nil, errors.New("invalid discount percentage format")
		}
	}

	taxAmount := decimal.Zero
	if req.TaxAmount != "" {
		taxAmount, err = decimal.NewFromString(req.TaxAmount)
		if err != nil {
			return nil, errors.New("invalid tax amount format")
		}
	}

	// Create invoice item
	item := models.PurchaseInvoiceItem{
		PurchaseInvoiceID:   invoice.ID,
		PurchaseOrderItemID: req.PurchaseOrderItemID,
		GoodsReceiptItemID:  req.GoodsReceiptItemID,
		ProductID:           req.ProductID,
		ProductCode:         product.Code,
		ProductName:         product.Name,
		UnitID:              unit.ID, // Use looked up unit ID (either from request or base unit)
		UnitName:            unit.UnitName,
		Quantity:            quantity,
		UnitPrice:           unitPrice,
		DiscountAmount:      discountAmount,
		DiscountPct:         discountPct,
		TaxAmount:           taxAmount,
		Notes:               req.Notes,
	}

	// Calculate line total
	item.CalculateLineTotal()

	if err := tx.Create(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

// updateInvoiceItem helper function to create/update an invoice item during update operation
func (s *PurchaseInvoiceService) updateInvoiceItem(
	tx *gorm.DB,
	invoice *models.PurchaseInvoice,
	req dto.UpdatePurchaseInvoiceItemRequest,
) (*models.PurchaseInvoiceItem, error) {
	// Extract tenant_id from current session to preserve it in new sessions
	tenantID, _ := tx.Get("tenant_id")

	// Helper function to create a fresh session with tenant context preserved
	freshSession := func() *gorm.DB {
		return tx.Session(&gorm.Session{NewDB: true}).Set("tenant_id", tenantID)
	}

	// Get product information for denormalization
	// Use a fresh session to avoid carrying WHERE conditions from previous queries
	var product models.Product
	if err := freshSession().Preload("Units").First(&product, "id = ?", req.ProductID).Error; err != nil {
		return nil, errors.New("product not found")
	}

	// Get unit information - if UnitID not provided, use product's base unit
	var unit models.ProductUnit
	if req.UnitID != "" {
		// Use a fresh session to avoid carrying WHERE conditions from product query
		if err := freshSession().First(&unit, "id = ?", req.UnitID).Error; err != nil {
			return nil, errors.New("product unit not found")
		}
	} else {
		// Look up the product's base unit from preloaded units
		found := false
		for _, u := range product.Units {
			if u.IsBaseUnit {
				unit = u
				found = true
				break
			}
		}
		if !found {
			// Fallback to database query if not in preloaded units
			// Use a fresh session to avoid carrying WHERE conditions
			if err := freshSession().Where("product_id = ? AND is_base_unit = true", req.ProductID).First(&unit).Error; err != nil {
				return nil, errors.New("product base unit not found")
			}
		}
	}

	// Parse quantity
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, errors.New("invalid quantity format")
	}
	if quantity.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("quantity must be greater than 0")
	}

	// Parse unit price
	unitPrice, err := decimal.NewFromString(req.UnitPrice)
	if err != nil {
		return nil, errors.New("invalid unit price format")
	}
	if unitPrice.LessThan(decimal.Zero) {
		return nil, errors.New("unit price cannot be negative")
	}

	// Parse optional fields
	discountAmount := decimal.Zero
	if req.DiscountAmount != "" {
		discountAmount, err = decimal.NewFromString(req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
	}

	discountPct := decimal.Zero
	if req.DiscountPct != "" {
		discountPct, err = decimal.NewFromString(req.DiscountPct)
		if err != nil {
			return nil, errors.New("invalid discount percentage format")
		}
	}

	taxAmount := decimal.Zero
	if req.TaxAmount != "" {
		taxAmount, err = decimal.NewFromString(req.TaxAmount)
		if err != nil {
			return nil, errors.New("invalid tax amount format")
		}
	}

	// Create invoice item
	item := models.PurchaseInvoiceItem{
		PurchaseInvoiceID:   invoice.ID,
		PurchaseOrderItemID: req.PurchaseOrderItemID, // Preserve PO item linkage
		GoodsReceiptItemID:  req.GoodsReceiptItemID,  // Preserve GRN item linkage for invoiced qty tracking
		ProductID:           req.ProductID,
		ProductCode:         product.Code,
		ProductName:         product.Name,
		UnitID:              unit.ID,
		UnitName:            unit.UnitName,
		Quantity:            quantity,
		UnitPrice:           unitPrice,
		DiscountAmount:      discountAmount,
		DiscountPct:         discountPct,
		TaxAmount:           taxAmount,
		Notes:               req.Notes,
	}

	// Calculate line total
	item.CalculateLineTotal()

	// Use fresh session for Create to avoid any stale query state
	if err := freshSession().Create(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil
}

// ============================================================================
// UPDATE Operation
// ============================================================================

// UpdatePurchaseInvoice updates an existing purchase invoice
func (s *PurchaseInvoiceService) UpdatePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, userID, ipAddress, userAgent string,
	req dto.UpdatePurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Only DRAFT invoices can be fully updated
	if invoice.Status != models.PurchaseInvoiceStatusDraft {
		return nil, errors.New("only draft invoices can be updated")
	}

	// Capture old values for audit logging (including items)
	oldValues := map[string]interface{}{
		"invoice_date":           invoice.InvoiceDate.Format("2006-01-02"),
		"due_date":               invoice.DueDate.Format("2006-01-02"),
		"supplier_id":            invoice.SupplierID,
		"supplier_name":          invoice.SupplierName,
		"discount_amount":        invoice.DiscountAmount.String(),
		"tax_rate":               invoice.TaxRate.String(),
		"payment_term_days":      invoice.PaymentTermDays,
		"notes":                  invoice.Notes,
		"status":                 string(invoice.Status),
		"shipping_cost":          invoice.ShippingCost.String(),
		"handling_cost":          invoice.HandlingCost.String(),
		"other_cost":             invoice.OtherCost.String(),
		"other_cost_description": invoice.OtherCostDescription,
		"subtotal_amount":        invoice.SubtotalAmount.String(),
		"tax_amount":             invoice.TaxAmount.String(),
		"total_amount":           invoice.TotalAmount.String(),
	}

	// Capture old items for audit logging
	oldItems := make([]map[string]interface{}, len(invoice.Items))
	for i, item := range invoice.Items {
		oldItems[i] = map[string]interface{}{
			"id":              item.ID,
			"product_id":      item.ProductID,
			"product_code":    item.ProductCode,
			"product_name":    item.ProductName,
			"unit_id":         item.UnitID,
			"unit_name":       item.UnitName,
			"quantity":        item.Quantity.String(),
			"unit_price":      item.UnitPrice.String(),
			"discount_amount": item.DiscountAmount.String(),
			"discount_pct":    item.DiscountPct.String(),
			"tax_amount":      item.TaxAmount.String(),
			"line_total":      item.LineTotal.String(),
			"notes":           item.Notes,
		}
	}
	oldValues["items"] = oldItems

	// Update fields if provided
	if req.InvoiceDate != nil {
		invoiceDate, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return nil, errors.New("invalid invoice date format")
		}
		invoice.InvoiceDate = invoiceDate
	}

	if req.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due date format")
		}
		invoice.DueDate = dueDate
	}

	// Validate due date is not before invoice date after updates
	if req.InvoiceDate != nil || req.DueDate != nil {
		if invoice.DueDate.Before(invoice.InvoiceDate) {
			return nil, errors.New("due date must be after invoice date")
		}
	}

	if req.SupplierID != nil {
		var supplier models.Supplier
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).First(&supplier, "id = ?", *req.SupplierID).Error; err != nil {
			return nil, errors.New("supplier not found")
		}
		invoice.SupplierID = *req.SupplierID
		invoice.SupplierName = supplier.Name
		invoice.SupplierCode = &supplier.Code
	}

	if req.DiscountAmount != nil {
		discountAmount, err := decimal.NewFromString(*req.DiscountAmount)
		if err != nil {
			return nil, errors.New("invalid discount amount format")
		}
		invoice.DiscountAmount = discountAmount
	}

	if req.TaxRate != nil {
		taxRate, err := decimal.NewFromString(*req.TaxRate)
		if err != nil {
			return nil, errors.New("invalid tax rate format")
		}
		invoice.TaxRate = taxRate
	}

	// Recalculate payment term days if dates changed and paymentTermDays not explicitly provided
	if req.PaymentTermDays != nil {
		invoice.PaymentTermDays = *req.PaymentTermDays
	} else if req.InvoiceDate != nil || req.DueDate != nil {
		// Recalculate from current dates
		paymentTermDays := int(invoice.DueDate.Sub(invoice.InvoiceDate).Hours() / 24)
		if paymentTermDays < 0 {
			paymentTermDays = 0
		}
		invoice.PaymentTermDays = paymentTermDays
	}

	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if req.Status != nil {
		invoice.Status = models.PurchaseInvoiceStatus(*req.Status)
	}

	// Update non-goods costs (biaya tambahan)
	if req.ShippingCost != nil {
		shippingCost, err := decimal.NewFromString(*req.ShippingCost)
		if err != nil {
			return nil, errors.New("invalid shipping cost format")
		}
		invoice.ShippingCost = shippingCost
	}

	if req.HandlingCost != nil {
		handlingCost, err := decimal.NewFromString(*req.HandlingCost)
		if err != nil {
			return nil, errors.New("invalid handling cost format")
		}
		invoice.HandlingCost = handlingCost
	}

	if req.OtherCost != nil {
		otherCost, err := decimal.NewFromString(*req.OtherCost)
		if err != nil {
			return nil, errors.New("invalid other cost format")
		}
		invoice.OtherCost = otherCost
	}

	if req.OtherCostDescription != nil {
		invoice.OtherCostDescription = req.OtherCostDescription
	}

	invoice.UpdatedBy = &userID

	// Update items if provided
	if len(req.Items) > 0 {
		// Delete existing items
		if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
			Where("purchase_invoice_id = ?", invoice.ID).
			Delete(&models.PurchaseInvoiceItem{}).Error; err != nil {
			return nil, errors.New("failed to delete existing invoice items")
		}

		// Create new items
		var newItems []models.PurchaseInvoiceItem
		for _, itemReq := range req.Items {
			item, err := s.updateInvoiceItem(s.db.WithContext(ctx).Set("tenant_id", tenantID), invoice, itemReq)
			if err != nil {
				return nil, err
			}
			newItems = append(newItems, *item)
		}
		invoice.Items = newItems
	}

	// Recalculate totals
	invoice.CalculateTotals()

	// Save changes
	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(invoice).Error; err != nil {
		return nil, err
	}

	// Log audit trail for purchase invoice update
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

		// Prepare new values for comparison (including items)
		newValues := map[string]interface{}{
			"invoice_date":           invoice.InvoiceDate.Format("2006-01-02"),
			"due_date":               invoice.DueDate.Format("2006-01-02"),
			"supplier_id":            invoice.SupplierID,
			"supplier_name":          invoice.SupplierName,
			"discount_amount":        invoice.DiscountAmount.String(),
			"tax_rate":               invoice.TaxRate.String(),
			"payment_term_days":      invoice.PaymentTermDays,
			"notes":                  invoice.Notes,
			"status":                 string(invoice.Status),
			"shipping_cost":          invoice.ShippingCost.String(),
			"handling_cost":          invoice.HandlingCost.String(),
			"other_cost":             invoice.OtherCost.String(),
			"other_cost_description": invoice.OtherCostDescription,
			"subtotal_amount":        invoice.SubtotalAmount.String(),
			"tax_amount":             invoice.TaxAmount.String(),
			"total_amount":           invoice.TotalAmount.String(),
		}

		// Capture new items for audit logging
		newItems := make([]map[string]interface{}, len(invoice.Items))
		for i, item := range invoice.Items {
			newItems[i] = map[string]interface{}{
				"id":              item.ID,
				"product_id":      item.ProductID,
				"product_code":    item.ProductCode,
				"product_name":    item.ProductName,
				"unit_id":         item.UnitID,
				"unit_name":       item.UnitName,
				"quantity":        item.Quantity.String(),
				"unit_price":      item.UnitPrice.String(),
				"discount_amount": item.DiscountAmount.String(),
				"discount_pct":    item.DiscountPct.String(),
				"tax_amount":      item.TaxAmount.String(),
				"line_total":      item.LineTotal.String(),
				"notes":           item.Notes,
			}
		}
		newValues["items"] = newItems

		if err := s.auditService.LogPurchaseInvoiceUpdated(ctx, auditCtx, invoice.ID, oldValues, newValues); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice update: %v\n", err)
		}
	}

	return invoice, nil
}

// ============================================================================
// DELETE Operation
// ============================================================================

// DeletePurchaseInvoice soft deletes a purchase invoice
func (s *PurchaseInvoiceService) DeletePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, userID, ipAddress, userAgent string,
) error {
	// Get existing invoice with items
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return err
	}

	// Only DRAFT invoices can be deleted
	if invoice.Status != models.PurchaseInvoiceStatusDraft {
		return errors.New("only draft invoices can be deleted")
	}

	// Check if invoice has payments
	if invoice.PaidAmount.GreaterThan(decimal.Zero) {
		return errors.New("cannot delete invoice with payments")
	}

	// Use transaction for atomic delete and PO qty revert
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Collect PO item qty updates from invoice items
		poItemUpdates := make(map[string]decimal.Decimal) // PO Item ID -> qty to subtract from InvoicedQty
		for _, item := range invoice.Items {
			if item.PurchaseOrderItemID != nil && *item.PurchaseOrderItemID != "" {
				if existing, ok := poItemUpdates[*item.PurchaseOrderItemID]; ok {
					poItemUpdates[*item.PurchaseOrderItemID] = existing.Add(item.Quantity)
				} else {
					poItemUpdates[*item.PurchaseOrderItemID] = item.Quantity
				}
			}
		}

		// 2. Soft delete the invoice
		if err := tx.Delete(invoice).Error; err != nil {
			return fmt.Errorf("failed to delete invoice: %w", err)
		}

		// 3. Revert InvoicedQty on PO items and update PO InvoiceStatus
		if invoice.PurchaseOrderID != nil && *invoice.PurchaseOrderID != "" && len(poItemUpdates) > 0 {
			for poItemID, subtractQty := range poItemUpdates {
				// Subtract InvoicedQty on PO item (ensure it doesn't go below 0)
				if err := tx.Model(&models.PurchaseOrderItem{}).
					Where("id = ?", poItemID).
					Update("invoiced_qty", gorm.Expr("GREATEST(invoiced_qty - ?, 0)", subtractQty)).Error; err != nil {
					return fmt.Errorf("failed to revert PO item invoiced qty: %w", err)
				}
			}

			// Reload PO and update invoice status
			var purchaseOrder models.PurchaseOrder
			if err := tx.Preload("Items").First(&purchaseOrder, "id = ?", *invoice.PurchaseOrderID).Error; err == nil {
				purchaseOrder.UpdateInvoiceStatus()
				if err := tx.Model(&purchaseOrder).Update("invoice_status", purchaseOrder.InvoiceStatus).Error; err != nil {
					return fmt.Errorf("failed to update PO invoice status: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Log audit trail for purchase invoice deletion
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

		oldValues := map[string]interface{}{
			"status": string(invoice.Status),
		}

		newValues := map[string]interface{}{
			"deleted": true,
		}

		if err := s.auditService.LogPurchaseInvoiceDeleted(ctx, auditCtx, invoice.ID, invoice.InvoiceNumber, oldValues, newValues); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice deletion: %v\n", err)
		}
	}

	return nil
}

// ============================================================================
// WORKFLOW Operations
// ============================================================================

// SubmitPurchaseInvoice submits a purchase invoice for approval
func (s *PurchaseInvoiceService) SubmitPurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, submitterID, ipAddress, userAgent string,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Capture old status for audit
	oldStatus := string(invoice.Status)

	// Only DRAFT invoices can be submitted
	if invoice.Status != models.PurchaseInvoiceStatusDraft {
		return nil, errors.New("only draft invoices can be submitted")
	}

	// Update status
	invoice.Status = models.PurchaseInvoiceStatusSubmitted
	invoice.UpdatedBy = &submitterID

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(invoice).Error; err != nil {
		return nil, err
	}

	// Log audit trail for purchase invoice submission
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &submitterID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]interface{}{
			"status": oldStatus,
		}

		newValues := map[string]interface{}{
			"status": string(invoice.Status),
		}

		if err := s.auditService.LogPurchaseInvoiceSubmitted(ctx, auditCtx, invoice.ID, invoice.InvoiceNumber, oldValues, newValues); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice submission: %v\n", err)
		}
	}

	return invoice, nil
}

// ApprovePurchaseInvoice approves a purchase invoice
func (s *PurchaseInvoiceService) ApprovePurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, approverID, ipAddress, userAgent string,
	req dto.ApprovePurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Capture old status for audit
	oldStatus := string(invoice.Status)

	// Only SUBMITTED invoices can be approved
	if invoice.Status != models.PurchaseInvoiceStatusSubmitted {
		return nil, errors.New("only submitted invoices can be approved")
	}

	// Update status
	now := time.Now()
	invoice.Status = models.PurchaseInvoiceStatusApproved
	invoice.ApprovedBy = &approverID
	invoice.ApprovedAt = &now
	invoice.UpdatedBy = &approverID

	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).Save(invoice).Error; err != nil {
		return nil, err
	}

	// Log audit trail for purchase invoice approval
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &approverID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]interface{}{
			"status": oldStatus,
		}

		newValues := map[string]interface{}{
			"status": string(invoice.Status),
		}

		if err := s.auditService.LogPurchaseInvoiceApproved(ctx, auditCtx, invoice.ID, invoice.InvoiceNumber, oldValues, newValues); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice approval: %v\n", err)
		}
	}

	return invoice, nil
}

// RejectPurchaseInvoice rejects a purchase invoice and reverts invoiced quantities
func (s *PurchaseInvoiceService) RejectPurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, rejecterID, ipAddress, userAgent string,
	req dto.RejectPurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice with items
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Capture old status for audit
	oldStatus := string(invoice.Status)

	// Only SUBMITTED invoices can be rejected
	if invoice.Status != models.PurchaseInvoiceStatusSubmitted {
		return nil, errors.New("only submitted invoices can be rejected")
	}

	// Use transaction for atomic reject and PO qty revert
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Collect PO item qty updates from invoice items
		poItemUpdates := make(map[string]decimal.Decimal) // PO Item ID -> qty to subtract from InvoicedQty
		for _, item := range invoice.Items {
			if item.PurchaseOrderItemID != nil && *item.PurchaseOrderItemID != "" {
				if existing, ok := poItemUpdates[*item.PurchaseOrderItemID]; ok {
					poItemUpdates[*item.PurchaseOrderItemID] = existing.Add(item.Quantity)
				} else {
					poItemUpdates[*item.PurchaseOrderItemID] = item.Quantity
				}
			}
		}

		// 2. Update invoice status to REJECTED
		now := time.Now()
		invoice.Status = models.PurchaseInvoiceStatusRejected
		invoice.RejectedBy = &rejecterID
		invoice.RejectedAt = &now
		invoice.RejectedReason = &req.Reason
		invoice.UpdatedBy = &rejecterID

		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to update invoice status: %w", err)
		}

		// 3. Revert InvoicedQty on PO items and update PO InvoiceStatus
		if invoice.PurchaseOrderID != nil && *invoice.PurchaseOrderID != "" && len(poItemUpdates) > 0 {
			for poItemID, subtractQty := range poItemUpdates {
				// Subtract InvoicedQty on PO item (ensure it doesn't go below 0)
				if err := tx.Model(&models.PurchaseOrderItem{}).
					Where("id = ?", poItemID).
					Update("invoiced_qty", gorm.Expr("GREATEST(invoiced_qty - ?, 0)", subtractQty)).Error; err != nil {
					return fmt.Errorf("failed to revert PO item invoiced qty: %w", err)
				}
			}

			// Reload PO and update invoice status
			var purchaseOrder models.PurchaseOrder
			if err := tx.Preload("Items").First(&purchaseOrder, "id = ?", *invoice.PurchaseOrderID).Error; err == nil {
				purchaseOrder.UpdateInvoiceStatus()
				if err := tx.Model(&purchaseOrder).Update("invoice_status", purchaseOrder.InvoiceStatus).Error; err != nil {
					return fmt.Errorf("failed to update PO invoice status: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Log audit trail for purchase invoice rejection
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &rejecterID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]interface{}{
			"status": oldStatus,
		}

		newValues := map[string]interface{}{
			"status": string(invoice.Status),
			"reason": req.Reason,
		}

		if err := s.auditService.LogPurchaseInvoiceRejected(ctx, auditCtx, invoice.ID, invoice.InvoiceNumber, oldValues, newValues, req.Reason); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice rejection: %v\n", err)
		}
	}

	return invoice, nil
}

// CancelPurchaseInvoice cancels an approved purchase invoice and reverts invoiced quantities
func (s *PurchaseInvoiceService) CancelPurchaseInvoice(
	ctx context.Context,
	tenantID, companyID, invoiceID, cancellerID, ipAddress, userAgent string,
	req dto.CancelPurchaseInvoiceRequest,
) (*models.PurchaseInvoice, error) {
	// Get existing invoice with items
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Capture old status for audit
	oldStatus := string(invoice.Status)

	// Only APPROVED invoices can be cancelled
	if invoice.Status != models.PurchaseInvoiceStatusApproved {
		return nil, errors.New("only approved invoices can be cancelled")
	}

	// Cannot cancel if there are payments
	if invoice.PaidAmount.GreaterThan(decimal.Zero) {
		return nil, errors.New("cannot cancel invoice with payments")
	}

	// Use transaction for atomic cancel and PO qty revert
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// 1. Collect PO item qty updates from invoice items
		poItemUpdates := make(map[string]decimal.Decimal) // PO Item ID -> qty to subtract from InvoicedQty
		for _, item := range invoice.Items {
			if item.PurchaseOrderItemID != nil && *item.PurchaseOrderItemID != "" {
				if existing, ok := poItemUpdates[*item.PurchaseOrderItemID]; ok {
					poItemUpdates[*item.PurchaseOrderItemID] = existing.Add(item.Quantity)
				} else {
					poItemUpdates[*item.PurchaseOrderItemID] = item.Quantity
				}
			}
		}

		// 2. Update invoice status to CANCELLED
		now := time.Now()
		invoice.Status = models.PurchaseInvoiceStatusCancelled
		invoice.CancelledBy = &cancellerID
		invoice.CancelledAt = &now
		invoice.CancellationReason = &req.Reason
		invoice.UpdatedBy = &cancellerID

		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to update invoice status: %w", err)
		}

		// 3. Revert InvoicedQty on PO items and update PO InvoiceStatus
		if invoice.PurchaseOrderID != nil && *invoice.PurchaseOrderID != "" && len(poItemUpdates) > 0 {
			for poItemID, subtractQty := range poItemUpdates {
				// Subtract InvoicedQty on PO item (ensure it doesn't go below 0)
				if err := tx.Model(&models.PurchaseOrderItem{}).
					Where("id = ?", poItemID).
					Update("invoiced_qty", gorm.Expr("GREATEST(invoiced_qty - ?, 0)", subtractQty)).Error; err != nil {
					return fmt.Errorf("failed to revert PO item invoiced qty: %w", err)
				}
			}

			// Reload PO and update invoice status
			var purchaseOrder models.PurchaseOrder
			if err := tx.Preload("Items").First(&purchaseOrder, "id = ?", *invoice.PurchaseOrderID).Error; err == nil {
				purchaseOrder.UpdateInvoiceStatus()
				if err := tx.Model(&purchaseOrder).Update("invoice_status", purchaseOrder.InvoiceStatus).Error; err != nil {
					return fmt.Errorf("failed to update PO invoice status: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Log audit trail for purchase invoice cancellation
	if s.auditService != nil {
		requestID := uuid.New().String()
		auditCtx := &audit.AuditContext{
			TenantID:  &tenantID,
			CompanyID: &companyID,
			UserID:    &cancellerID,
			RequestID: &requestID,
			IPAddress: &ipAddress,
			UserAgent: &userAgent,
		}

		oldValues := map[string]interface{}{
			"status": oldStatus,
		}

		newValues := map[string]interface{}{
			"status": string(invoice.Status),
			"reason": req.Reason,
		}

		if err := s.auditService.LogPurchaseInvoiceCancelled(ctx, auditCtx, invoice.ID, invoice.InvoiceNumber, oldValues, newValues, req.Reason); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice cancellation: %v\n", err)
		}
	}

	return invoice, nil
}

// RecordPayment records a payment against a purchase invoice
func (s *PurchaseInvoiceService) RecordPayment(
	ctx context.Context,
	tenantID, companyID, invoiceID, userID, ipAddress, userAgent string,
	req dto.RecordPaymentRequest,
) (*models.PurchaseInvoicePayment, error) {
	// Get existing invoice
	invoice, err := s.GetPurchaseInvoice(ctx, tenantID, companyID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Only APPROVED invoices can receive payments
	if invoice.Status != models.PurchaseInvoiceStatusApproved &&
		invoice.Status != models.PurchaseInvoiceStatusPaid {
		return nil, errors.New("only approved invoices can receive payments")
	}

	// Parse payment amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, errors.New("invalid payment amount format")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("payment amount must be greater than 0")
	}

	// Check if payment exceeds remaining amount
	if amount.GreaterThan(invoice.RemainingAmount) {
		return nil, errors.New("payment amount exceeds remaining amount")
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, errors.New("invalid payment date format")
	}

	// Transaction with tenant context
	var payment *models.PurchaseInvoicePayment
	err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
		// Create payment record
		payment = &models.PurchaseInvoicePayment{
			TenantID:          tenantID,
			CompanyID:         companyID,
			PurchaseInvoiceID: invoiceID,
			PaymentNumber:     req.PaymentNumber,
			PaymentDate:       paymentDate,
			Amount:            amount,
			PaymentMethod:     models.PaymentMethod(req.PaymentMethod),
			Reference:         req.Reference,
			BankAccountID:     req.BankAccountID,
			Notes:             req.Notes,
			CreatedBy:         userID,
		}

		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		// Update invoice paid amount
		invoice.PaidAmount = invoice.PaidAmount.Add(amount)
		invoice.RemainingAmount = invoice.TotalAmount.Sub(invoice.PaidAmount)
		invoice.UpdatePaymentStatus()
		invoice.UpdatedBy = &userID

		if err := tx.Save(invoice).Error; err != nil {
			return fmt.Errorf("failed to update invoice: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Log audit trail for payment recording
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

		paymentData := map[string]interface{}{
			"payment_id":        payment.ID,
			"payment_number":    payment.PaymentNumber,
			"payment_date":      payment.PaymentDate.Format("2006-01-02"),
			"amount":            payment.Amount.String(),
			"payment_method":    string(payment.PaymentMethod),
			"reference":         payment.Reference,
			"bank_account_id":   payment.BankAccountID,
			"notes":             payment.Notes,
			"invoice_id":        invoiceID,
			"invoice_number":    invoice.InvoiceNumber,
			"paid_amount":       invoice.PaidAmount.String(),
			"remaining_amount":  invoice.RemainingAmount.String(),
			"payment_status":    string(invoice.PaymentStatus),
		}

		if err := s.auditService.LogPurchaseInvoicePaymentRecorded(ctx, auditCtx, invoiceID, invoice.InvoiceNumber, paymentData); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("[AUDIT] Failed to log purchase invoice payment: %v\n", err)
		}
	}

	return payment, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// toSnakeCase converts camelCase to snake_case
func toSnakeCase(s string) string {
	switch s {
	case "invoiceNumber":
		return "invoice_number"
	case "invoiceDate":
		return "invoice_date"
	case "dueDate":
		return "due_date"
	case "totalAmount":
		return "total_amount"
	case "createdAt":
		return "created_at"
	default:
		return s
	}
}
