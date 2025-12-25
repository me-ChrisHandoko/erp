# Module Group 4: Purchase Flow (Procurement Cycle)

**Implementation Priority:** WEEK 6 (Inventory Replenishment)
**Dependencies:** Company Profile ✅, Master Data ✅ (Products, Suppliers, Warehouses), Sales Flow ✅
**Modules:** Purchase Order, Goods Receipt, Supplier Payment

---

## Overview

This module group implements the **procurement cycle** from ordering to supplier payment. Essential for inventory replenishment and cost management.

### Business Flow
```
Purchase Order (DRAFT → CONFIRMED)
  ↓
Goods Receipt (PENDING → RECEIVED → ACCEPTED)
  ↓
Batch Recording (for batch-tracked products)
  ↓
Update Warehouse Stock
  ↓
Supplier Payment (cash, transfer, check/giro)
  ↓
Update Supplier Outstanding
```

### Why Week 6?
1. **Inventory replenishment** - needed when sales deplete stock
2. **Cost tracking** - purchase cost affects profitability
3. **Supplier management** - payables tracking
4. **Simpler than sales flow** - similar patterns, less complexity

---

## Module 1: Purchase Order Management

### Purpose
Record supplier orders with items, pricing, and approval workflow.

### Key Features (MVP)
1. **Create/Edit Purchase Order** - DRAFT status, add/remove items
2. **Supplier Pricing** - from ProductSupplier or manual entry
3. **Confirm Order** - DRAFT → CONFIRMED (send to supplier)
4. **Status Tracking** - DRAFT, CONFIRMED, COMPLETED, CANCELLED
5. **Auto-Number Generation** - from company.PONumberFormat

### Database Model

```go
type PurchaseOrder struct {
    ID              string
    TenantID        string
    PONumber        string          // Auto-generated: PO/001
    PODate          time.Time
    SupplierID      string
    WarehouseID     string          // Destination warehouse
    Status          PurchaseOrderStatus // DRAFT, CONFIRMED, COMPLETED, CANCELLED
    Subtotal        decimal.Decimal
    DiscountAmount  decimal.Decimal
    TaxAmount       decimal.Decimal
    TotalAmount     decimal.Decimal
    ExpectedDate    *time.Time      // Expected delivery date
    Notes           *string
    ApprovedBy      *string
    ApprovedAt      *time.Time
    CancelledBy      *string
    CancelledAt     *time.Time

    // Relations
    Supplier  Supplier
    Warehouse Warehouse
    Items     []PurchaseOrderItem
}

type PurchaseOrderItem struct {
    ID             string
    PurchaseOrderID string
    ProductID      string
    ProductUnitID  *string
    Quantity       decimal.Decimal
    UnitCost       decimal.Decimal
    Subtotal       decimal.Decimal
    Notes          *string
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/purchase-orders              # Create
GET    /api/v1/purchase-orders               # List with filters
GET    /api/v1/purchase-orders/:id           # Details
PUT    /api/v1/purchase-orders/:id           # Update (DRAFT only)
POST   /api/v1/purchase-orders/:id/confirm   # Confirm
DELETE /api/v1/purchase-orders/:id           # Cancel
```

### Business Logic

```go
// Create PO with supplier pricing
func (s *PurchaseOrderService) CreatePO(ctx context.Context, tenantID string, req *CreatePORequest) (*models.PurchaseOrder, error) {
    supplier := getSupplier(req.SupplierID)

    // Check supplier credit limit
    if supplier.CurrentOutstanding.Add(req.TotalAmount).GreaterThan(supplier.CreditLimit) {
        return nil, errors.New("supplier credit limit exceeded")
    }

    // Generate PO number
    poNumber := GeneratePONumber(company, tenantID)

    po := &models.PurchaseOrder{
        TenantID:     tenantID,
        PONumber:     poNumber,
        PODate:       req.PODate,
        SupplierID:   req.SupplierID,
        WarehouseID:  req.WarehouseID,
        Status:       models.POStatusDraft,
        ExpectedDate: req.ExpectedDate,
        Notes:        req.Notes,
    }

    tx := s.db.Begin()

    // Create PO
    tx.Create(po)

    // Create items
    for _, itemReq := range req.Items {
        // Get supplier price from ProductSupplier or use manual entry
        unitCost := itemReq.UnitCost
        if unitCost.IsZero() {
            // Auto-fetch from ProductSupplier if exists
            var ps models.ProductSupplier
            err := tx.Where("product_id = ? AND supplier_id = ?", itemReq.ProductID, req.SupplierID).First(&ps).Error
            if err == nil {
                unitCost = ps.SupplierPrice
            }
        }

        item := &models.PurchaseOrderItem{
            PurchaseOrderID: po.ID,
            ProductID:       itemReq.ProductID,
            ProductUnitID:   itemReq.ProductUnitID,
            Quantity:        itemReq.Quantity,
            UnitCost:        unitCost,
            Subtotal:        itemReq.Quantity.Mul(unitCost),
            Notes:           itemReq.Notes,
        }
        tx.Create(item)

        po.Subtotal = po.Subtotal.Add(item.Subtotal)
    }

    // Calculate tax (if applicable)
    if supplier.IsPKP {
        po.TaxAmount = po.Subtotal.Mul(company.PPNRate.Div(decimal.NewFromInt(100)))
    }

    po.TotalAmount = po.Subtotal.Sub(po.DiscountAmount).Add(po.TaxAmount)
    tx.Save(po)

    tx.Commit()

    return po, nil
}
```

---

## Module 2: Goods Receipt Management

### Purpose
Record product receipts from suppliers with quality inspection and batch tracking.

### Key Features (MVP)
1. **Create Goods Receipt from PO** - auto-fill items
2. **Batch Recording** - assign batch number, expiry date (for perishables)
3. **Quality Inspection** - accepted/rejected quantities
4. **Status Tracking** - PENDING → RECEIVED → ACCEPTED/REJECTED
5. **Update Warehouse Stock** - add stock when accepted

### Database Model

```go
type GoodsReceipt struct {
    ID                string
    TenantID          string
    GRNNumber         string          // Auto-generated
    ReceiptDate       time.Time
    PurchaseOrderID   *string
    SupplierID        string
    WarehouseID       string
    Status            GoodsReceiptStatus // PENDING, RECEIVED, ACCEPTED, REJECTED, PARTIAL
    Notes             *string
    ReceivedBy        *string         // Staff who received
    InspectedBy       *string
    InspectionDate    *time.Time
    InspectionNotes   *string

    // Relations
    PurchaseOrder *PurchaseOrder
    Supplier      Supplier
    Warehouse     Warehouse
    Items         []GoodsReceiptItem
}

type GoodsReceiptItem struct {
    ID               string
    GoodsReceiptID   string
    ProductID        string
    ProductUnitID    *string
    OrderedQty       decimal.Decimal  // From PO
    ReceivedQty      decimal.Decimal  // Actual received
    AcceptedQty      decimal.Decimal  // After QC
    RejectedQty      decimal.Decimal  // Failed QC
    BatchNumber      *string          // For batch-tracked products
    ManufactureDate  *time.Time
    ExpiryDate       *time.Time       // CRITICAL for perishables
    QualityStatus    *string          // GOOD, DAMAGED, QUARANTINE
    Notes            *string
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/goods-receipts              # Create from PO
GET    /api/v1/goods-receipts               # List
GET    /api/v1/goods-receipts/:id           # Details
PUT    /api/v1/goods-receipts/:id           # Update (PENDING only)
POST   /api/v1/goods-receipts/:id/accept    # Accept → update stock
DELETE /api/v1/goods-receipts/:id           # Cancel
```

### Business Logic

```go
// Accept goods receipt → Add stock
func (s *GoodsReceiptService) AcceptGoodsReceipt(ctx context.Context, tenantID, grnID string) error {
    grn := getGoodsReceipt(grnID)

    if grn.Status != GRNStatusReceived {
        return errors.New("can only accept RECEIVED goods receipts")
    }

    tx := s.db.Begin()

    for _, item := range grn.Items {
        // Convert to base units
        baseQty := item.AcceptedQty.Mul(item.ProductUnit.ConversionRate)

        // Update warehouse stock
        var whStock models.WarehouseStock
        err := tx.Where("warehouse_id = ? AND product_id = ?", grn.WarehouseID, item.ProductID).First(&whStock).Error

        if err != nil {
            // Create if doesn't exist
            whStock = models.WarehouseStock{
                WarehouseID: grn.WarehouseID,
                ProductID:   item.ProductID,
                Quantity:    baseQty,
            }
            tx.Create(&whStock)
        } else {
            stockBefore := whStock.Quantity
            whStock.Quantity = whStock.Quantity.Add(baseQty)
            tx.Save(&whStock)

            // Create inventory movement
            movement := &models.InventoryMovement{
                TenantID:      tenantID,
                WarehouseID:   grn.WarehouseID,
                ProductID:     item.ProductID,
                MovementType:  "IN",
                ReferenceType: "GOODS_RECEIPT",
                ReferenceID:   grnID,
                Quantity:      baseQty,
                StockBefore:   stockBefore,
                StockAfter:    whStock.Quantity,
                MovementDate:  grn.ReceiptDate,
            }
            tx.Create(movement)
        }

        // Create batch record (if product is batch-tracked)
        product := getProduct(item.ProductID)
        if product.IsBatchTracked {
            batch := &models.ProductBatch{
                BatchNumber:      item.BatchNumber,
                ProductID:        item.ProductID,
                WarehouseStockID: whStock.ID,
                ManufactureDate:  item.ManufactureDate,
                ExpiryDate:       item.ExpiryDate,
                Quantity:         baseQty,
                SupplierID:       &grn.SupplierID,
                GoodsReceiptID:   &grnID,
                ReceiptDate:      grn.ReceiptDate,
                Status:           models.BatchStatusAvailable,
                QualityStatus:    item.QualityStatus,
            }
            tx.Create(batch)
        }
    }

    // Update GRN status
    grn.Status = GRNStatusAccepted
    grn.InspectionDate = time.Now()
    tx.Save(grn)

    // Update PO status (if all GRNs accepted)
    if grn.PurchaseOrderID != nil {
        var po models.PurchaseOrder
        tx.First(&po, *grn.PurchaseOrderID)
        allAccepted := checkAllGRNsAccepted(tx, *grn.PurchaseOrderID)
        if allAccepted {
            po.Status = models.POStatusCompleted
            tx.Save(&po)
        }
    }

    tx.Commit()
    return nil
}
```

---

## Module 3: Supplier Payment

### Purpose
Record supplier payments and apply to purchase orders, update outstanding payables.

### Key Features (MVP)
1. **Record Payment** - cash, transfer, check/giro
2. **Apply to POs** - full or partial payment
3. **Update Supplier Outstanding** - deduct from CurrentOutstanding
4. **Check/Giro Tracking** - for non-cash payments

### Database Model

```go
type SupplierPayment struct {
    ID              string
    TenantID        string
    PaymentNumber   string          // Auto-generated
    PaymentDate     time.Time
    SupplierID      string
    PurchaseOrderID *string
    Amount          decimal.Decimal
    PaymentMethod   PaymentMethod   // CASH, TRANSFER, CHECK, GIRO
    BankAccountID   *string         // Company bank from which payment made
    Reference       *string
    CheckNumber     *string
    CheckDate       *time.Time
    CheckStatus     *CheckStatus
    Notes           *string

    // Relations
    Supplier      Supplier
    PurchaseOrder *PurchaseOrder
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/supplier-payments            # Record payment
GET    /api/v1/supplier-payments             # List
GET    /api/v1/supplier-payments/:id         # Details
DELETE /api/v1/supplier-payments/:id         # Void (same day only)
```

### Business Logic

```go
// Record supplier payment
func (s *SupplierPaymentService) RecordPayment(ctx context.Context, tenantID string, req *RecordSupplierPaymentRequest) (*models.SupplierPayment, error) {
    po := getPurchaseOrder(req.PurchaseOrderID)

    // Calculate unpaid amount
    var totalPaid decimal.Decimal
    s.db.Model(&models.SupplierPayment{}).
        Where("purchase_order_id = ?", req.PurchaseOrderID).
        Select("COALESCE(SUM(amount), 0)").
        Scan(&totalPaid)

    unpaidAmount := po.TotalAmount.Sub(totalPaid)

    if req.Amount.GreaterThan(unpaidAmount) {
        return nil, errors.New("payment amount exceeds unpaid amount")
    }

    tx := s.db.Begin()

    // Create payment record
    payment := &models.SupplierPayment{
        TenantID:        tenantID,
        PaymentNumber:   GeneratePaymentNumber(company, tenantID),
        PaymentDate:     req.PaymentDate,
        SupplierID:      po.SupplierID,
        PurchaseOrderID: &req.PurchaseOrderID,
        Amount:          req.Amount,
        PaymentMethod:   req.PaymentMethod,
        BankAccountID:   req.BankAccountID,
        Reference:       req.Reference,
        Notes:           req.Notes,
    }
    tx.Create(payment)

    // Update supplier outstanding
    var supplier models.Supplier
    tx.First(&supplier, po.SupplierID)
    supplier.CurrentOutstanding = supplier.CurrentOutstanding.Sub(req.Amount)
    supplier.LastTransactionAt = time.Now()
    tx.Save(&supplier)

    tx.Commit()

    return payment, nil
}
```

---

## Implementation Priority (Week 6)

**Day 1-2:** Purchase Order CRUD + supplier pricing
**Day 3-4:** Goods Receipt CRUD + batch recording
**Day 4:** Accept GRN → update stock
**Day 5:** Supplier Payment + outstanding updates

---

## Testing Checklist

### Purchase Order
- [ ] Create PO with supplier pricing
- [ ] Auto-fetch price from ProductSupplier
- [ ] Confirm PO
- [ ] Cancel PO
- [ ] Cannot confirm if supplier credit limit exceeded

### Goods Receipt
- [ ] Create GRN from PO
- [ ] Record batch information (batch number, expiry date)
- [ ] Accept GRN → adds stock
- [ ] Creates batch records for batch-tracked products
- [ ] Creates inventory movement
- [ ] Updates PO status to COMPLETED

### Supplier Payment
- [ ] Record full payment → PO fully paid
- [ ] Record partial payment
- [ ] Updates supplier outstanding correctly
- [ ] Cannot overpay PO

### Batch Tracking
- [ ] GRN creates batches with expiry dates
- [ ] Batches shown in product details
- [ ] FEFO logic (First Expired, First Out) for delivery (manual for MVP)

---

## Next Module Group

After completing Purchase Flow (Week 6), proceed to:
**→ `05-INVENTORY-MANAGEMENT.md`** (Stock Movements, Stock Opname, Stock Transfer)

---

## Summary

**Purchase Flow** implements the procurement cycle:
1. ✅ Purchase Order → Supplier orders with pricing
2. ✅ Goods Receipt → Stock replenishment with batch tracking
3. ✅ Supplier Payment → Payables management

**Critical Business Rules:**
- Supplier credit limit enforcement
- Batch information recording (expiry dates for perishables)
- Automatic supplier outstanding updates
- Stock increase only after GRN acceptance

**Estimated Completion:** 1 week (5 days)
