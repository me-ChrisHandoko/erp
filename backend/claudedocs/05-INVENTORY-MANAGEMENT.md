# Module Group 5: Inventory Management & Control

**Implementation Priority:** WEEK 7 (Final MVP Module)
**Dependencies:** All previous modules ✅ (Foundation, Master Data, Sales Flow, Purchase Flow)
**Modules:** Inventory Movement Tracking, Stock Opname, Stock Transfer

---

## Overview

This module group provides **inventory control and reconciliation** features. Essential for maintaining stock accuracy and preventing discrepancies.

### Business Context
- **Inventory Movements:** Audit trail of all stock changes (deliveries, receipts, adjustments)
- **Stock Opname:** Physical inventory count to reconcile with system stock
- **Stock Transfer:** Inter-warehouse transfers for multi-warehouse operations

### Why Week 7 (Last)?
1. **Dependent on transactions** - movements auto-created by sales/purchase flows
2. **Control mechanism** - validates and corrects existing stock data
3. **Lower priority for MVP** - can function with basic movement tracking initially
4. **Complexity** - requires understanding of all previous flows

---

## Module 1: Inventory Movement Tracking

### Purpose
Auto-create movement records for all stock changes with complete audit trail (stock before/after).

### Key Features (MVP)
1. **Auto-Create Movements** - from deliveries, goods receipts, adjustments
2. **Movement History** - per product, per warehouse, per batch
3. **Stock Before/After** - audit trail for reconciliation
4. **Movement Types** - IN, OUT, ADJUSTMENT, TRANSFER_IN, TRANSFER_OUT
5. **Reference Tracking** - link to source transaction (delivery, GRN, etc.)

### Database Model (Already Defined)

```go
type InventoryMovement struct {
    ID            string
    TenantID      string
    WarehouseID   string
    ProductID     string
    BatchID       *string         // Optional: for batch-tracked products
    MovementType  string          // IN, OUT, ADJUSTMENT, TRANSFER_IN, TRANSFER_OUT
    ReferenceType string          // DELIVERY, GOODS_RECEIPT, STOCK_OPNAME, STOCK_TRANSFER
    ReferenceID   string          // ID of the source transaction
    Quantity      decimal.Decimal // Positive = increase, Negative = decrease (in base units)
    StockBefore   decimal.Decimal // Stock level before movement
    StockAfter    decimal.Decimal // Stock level after movement
    UnitCost      *decimal.Decimal // Cost per unit (for costing/valuation)
    MovementDate  time.Time
    Notes         *string
    CreatedBy     *string
    CreatedAt     time.Time

    // Relations
    Warehouse Warehouse
    Product   Product
    Batch     *ProductBatch
}
```

---

### API Endpoints

#### 1. List Inventory Movements
```http
GET /api/v1/inventory/movements
GET /api/v1/inventory/movements?productId=clprod1&warehouseId=clwh1&fromDate=2025-01-01&toDate=2025-01-31

Query Parameters:
  - productId: string
  - warehouseId: string
  - batchId: string
  - movementType: IN | OUT | ADJUSTMENT | TRANSFER_IN | TRANSFER_OUT
  - fromDate: date
  - toDate: date
  - page, limit, sortBy, sortOrder

Response (200):
{
  "success": true,
  "data": [
    {
      "id": "clmov1",
      "warehouse": {
        "id": "clwh1",
        "code": "GD-01",
        "name": "Gudang Utama"
      },
      "product": {
        "id": "clprod1",
        "code": "BRS-001",
        "name": "Beras Premium 5kg"
      },
      "batch": {
        "id": "clbatch1",
        "batchNumber": "BATCH-2025-001",
        "expiryDate": null
      },
      "movementType": "IN",
      "referenceType": "GOODS_RECEIPT",
      "referenceId": "clgrn1",
      "quantity": "500.000",
      "stockBefore": "0.000",
      "stockAfter": "500.000",
      "unitCost": "12000.00",
      "movementDate": "2025-01-10T00:00:00Z",
      "notes": "Initial stock from GRN/001",
      "createdAt": "2025-01-10T08:00:00Z"
    },
    {
      "id": "clmov2",
      "warehouse": {
        "id": "clwh1",
        "code": "GD-01",
        "name": "Gudang Utama"
      },
      "product": {
        "id": "clprod1",
        "code": "BRS-001",
        "name": "Beras Premium 5kg"
      },
      "movementType": "OUT",
      "referenceType": "DELIVERY",
      "referenceId": "cldel1",
      "quantity": "-250.000",
      "stockBefore": "500.000",
      "stockAfter": "250.000",
      "movementDate": "2025-01-15T00:00:00Z",
      "notes": "Delivery to Toko Sembako Jaya",
      "createdAt": "2025-01-15T10:00:00Z"
    }
  ],
  "pagination": { /* ... */ },
  "summary": {
    "totalIn": "500.000",
    "totalOut": "-250.000",
    "netMovement": "250.000",
    "currentStock": "250.000"
  }
}
```

#### 2. Get Product Movement History
```http
GET /api/v1/inventory/movements/product/:productId
GET /api/v1/inventory/movements/product/:productId?warehouseId=clwh1&fromDate=2025-01-01

Response (200):
{
  "success": true,
  "data": {
    "product": {
      "id": "clprod1",
      "code": "BRS-001",
      "name": "Beras Premium 5kg"
    },
    "movements": [ /* Same as list above */ ],
    "summary": {
      "openingStock": "0.000",
      "totalIn": "500.000",
      "totalOut": "-250.000",
      "closingStock": "250.000",
      "warehouses": [
        {
          "warehouseId": "clwh1",
          "warehouseName": "Gudang Utama",
          "currentStock": "250.000"
        }
      ]
    }
  }
}
```

#### 3. Manual Stock Adjustment (ADMIN only)
```http
POST /api/v1/inventory/movements/adjustment

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}

Middleware: RequireRoleMiddleware("OWNER", "ADMIN", "WAREHOUSE")

Request:
{
  "warehouseId": "clwh1",
  "productId": "clprod1",
  "batchId": "clbatch1",          // Optional
  "adjustmentQty": "-10.000",     // Negative = decrease, Positive = increase
  "notes": "Stock damaged during handling",
  "adjustmentDate": "2025-01-16T00:00:00Z"
}

Business Logic:
1. Get current warehouse stock
2. Calculate new stock = current + adjustment
3. Validate new stock >= 0
4. Create inventory movement (type = ADJUSTMENT)
5. Update warehouse stock

Response (201):
{
  "success": true,
  "data": {
    "id": "clmov3",
    "movementType": "ADJUSTMENT",
    "referenceType": "MANUAL_ADJUSTMENT",
    "quantity": "-10.000",
    "stockBefore": "250.000",
    "stockAfter": "240.000",
    "notes": "Stock damaged during handling",
    "createdBy": "cluser1",
    "createdAt": "2025-01-16T09:00:00Z"
  }
}

Validation:
- Adjustment cannot result in negative stock
- OWNER/ADMIN/WAREHOUSE role required
- Notes mandatory for adjustments (audit requirement)
```

---

### Business Logic

```go
// Auto-create movement (called from DeliveryService, GoodsReceiptService, etc.)
func (s *InventoryService) CreateMovement(ctx context.Context, req *CreateMovementRequest) error {
    // Get current warehouse stock
    var whStock models.WarehouseStock
    s.db.Where("warehouse_id = ? AND product_id = ?", req.WarehouseID, req.ProductID).First(&whStock)

    stockBefore := whStock.Quantity
    stockAfter := stockBefore.Add(req.Quantity) // Positive for IN, negative for OUT

    // Validate
    if stockAfter.LessThan(decimal.Zero) {
        return errors.New("insufficient stock")
    }

    // Create movement record
    movement := &models.InventoryMovement{
        TenantID:      req.TenantID,
        WarehouseID:   req.WarehouseID,
        ProductID:     req.ProductID,
        BatchID:       req.BatchID,
        MovementType:  req.MovementType,
        ReferenceType: req.ReferenceType,
        ReferenceID:   req.ReferenceID,
        Quantity:      req.Quantity,
        StockBefore:   stockBefore,
        StockAfter:    stockAfter,
        UnitCost:      req.UnitCost,
        MovementDate:  req.MovementDate,
        Notes:         req.Notes,
        CreatedBy:     req.CreatedBy,
    }

    return s.db.Create(movement).Error
}
```

---

## Module 2: Stock Opname (Physical Inventory Count)

### Purpose
Record physical stock counts and reconcile with system stock, generating adjustment movements for discrepancies.

### Key Features (MVP)
1. **Create Stock Opname** - schedule physical count
2. **Record Counted Quantities** - per product, per warehouse
3. **Calculate Variance** - system stock vs physical count
4. **Generate Adjustments** - auto-create inventory movements for differences
5. **Opname History** - track historical counts

### Database Model (Already Defined)

```go
type StockOpname struct {
    ID          string
    TenantID    string
    OpnameNumber string          // Auto-generated
    OpnameDate  time.Time
    WarehouseID string
    Status      StockOpnameStatus // DRAFT, IN_PROGRESS, COMPLETED
    Notes       *string
    CreatedBy   string
    ApprovedBy  *string
    ApprovedAt  *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // Relations
    Warehouse Warehouse
    Items     []StockOpnameItem
}

type StockOpnameItem struct {
    ID               string
    StockOpnameID    string
    ProductID        string
    BatchID          *string
    SystemQty        decimal.Decimal // Stock in system
    PhysicalQty      decimal.Decimal // Counted stock
    Variance         decimal.Decimal // Difference (physical - system)
    VariancePct      decimal.Decimal // Percentage
    Notes            *string
    CountedBy        *string
    CountedAt        *time.Time

    // Relations
    Product Product
    Batch   *ProductBatch
}
```

**Enums:**
```go
type StockOpnameStatus string
const (
    OpnameStatusDraft      StockOpnameStatus = "DRAFT"
    OpnameStatusInProgress StockOpnameStatus = "IN_PROGRESS"
    OpnameStatusCompleted  StockOpnameStatus = "COMPLETED"
)
```

---

### API Endpoints

#### 1. Create Stock Opname
```http
POST /api/v1/inventory/stock-opname

Request:
{
  "opnameDate": "2025-01-20T00:00:00Z",
  "warehouseId": "clwh1",
  "notes": "Monthly stock count - January 2025"
}

Business Logic:
1. Generate opname number
2. Create stock opname record (status = DRAFT)
3. Auto-populate items from current warehouse stocks
4. Set systemQty from WarehouseStock.Quantity
5. PhysicalQty initially NULL (to be filled during count)

Response (201):
{
  "success": true,
  "data": {
    "id": "clopn1",
    "opnameNumber": "OPN/2025/001",
    "opnameDate": "2025-01-20T00:00:00Z",
    "warehouse": {
      "id": "clwh1",
      "code": "GD-01",
      "name": "Gudang Utama"
    },
    "status": "DRAFT",
    "items": [
      {
        "id": "clopni1",
        "product": {
          "id": "clprod1",
          "code": "BRS-001",
          "name": "Beras Premium 5kg"
        },
        "systemQty": "240.000",
        "physicalQty": null,        // To be filled
        "variance": null
      },
      // ... other products
    ],
    "notes": "Monthly stock count - January 2025",
    "createdBy": "cluser1",
    "createdAt": "2025-01-20T08:00:00Z"
  }
}
```

#### 2. Update Physical Count
```http
PUT /api/v1/inventory/stock-opname/:id/items/:itemId

Request:
{
  "physicalQty": "235.000",
  "notes": "5kg shortage found",
  "countedBy": "John Warehouse",
  "countedAt": "2025-01-20T10:30:00Z"
}

Business Logic:
1. Update physicalQty
2. Calculate variance = physicalQty - systemQty
3. Calculate variancePct = (variance / systemQty) × 100

Response (200):
{
  "success": true,
  "data": {
    "id": "clopni1",
    "product": {
      "id": "clprod1",
      "code": "BRS-001",
      "name": "Beras Premium 5kg"
    },
    "systemQty": "240.000",
    "physicalQty": "235.000",
    "variance": "-5.000",          // 5kg shortage
    "variancePct": "-2.08",        // -2.08%
    "notes": "5kg shortage found",
    "countedBy": "John Warehouse",
    "countedAt": "2025-01-20T10:30:00Z"
  }
}
```

#### 3. Complete Stock Opname
```http
POST /api/v1/inventory/stock-opname/:id/complete

Middleware: RequireRoleMiddleware("OWNER", "ADMIN")

Business Logic:
1. Validate all items have physicalQty
2. For each item with variance != 0:
   - Create inventory movement (type = ADJUSTMENT)
   - Update warehouse stock = physicalQty
3. Update opname status = COMPLETED

Response (200):
{
  "success": true,
  "data": {
    "id": "clopn1",
    "status": "COMPLETED",
    "approvedBy": "cluser1",
    "approvedAt": "2025-01-20T11:00:00Z",
    "summary": {
      "totalItems": 50,
      "itemsWithVariance": 5,
      "totalVariance": "-12.500",  // Total shortage
      "adjustmentsMade": 5
    }
  }
}
```

---

### Business Logic

```go
// Complete stock opname → generate adjustments
func (s *StockOpnameService) CompleteOpname(ctx context.Context, tenantID, opnameID string) error {
    var opname models.StockOpname
    s.db.Preload("Items.Product").Preload("Warehouse").First(&opname, opnameID)

    // Validate all items counted
    for _, item := range opname.Items {
        if item.PhysicalQty == nil {
            return errors.New("not all items have been counted")
        }
    }

    tx := s.db.Begin()

    adjustmentCount := 0

    for _, item := range opname.Items {
        if item.Variance.IsZero() {
            continue // No variance, skip
        }

        // Update warehouse stock
        var whStock models.WarehouseStock
        tx.Where("warehouse_id = ? AND product_id = ?", opname.WarehouseID, item.ProductID).First(&whStock)

        stockBefore := whStock.Quantity
        whStock.Quantity = item.PhysicalQty
        tx.Save(&whStock)

        // Create inventory movement
        movement := &models.InventoryMovement{
            TenantID:      tenantID,
            WarehouseID:   opname.WarehouseID,
            ProductID:     item.ProductID,
            BatchID:       item.BatchID,
            MovementType:  "ADJUSTMENT",
            ReferenceType: "STOCK_OPNAME",
            ReferenceID:   opnameID,
            Quantity:      item.Variance,  // Can be positive or negative
            StockBefore:   stockBefore,
            StockAfter:    item.PhysicalQty,
            MovementDate:  opname.OpnameDate,
            Notes:         item.Notes,
        }
        tx.Create(movement)

        adjustmentCount++
    }

    // Update opname status
    opname.Status = models.OpnameStatusCompleted
    now := time.Now()
    opname.ApprovedAt = &now
    tx.Save(&opname)

    tx.Commit()

    return nil
}
```

---

## Module 3: Stock Transfer (Inter-Warehouse)

### Purpose
Transfer stock between warehouses with complete tracking and dual movements.

### Key Features (MVP)
1. **Create Transfer Request** - from warehouse to warehouse
2. **Status Tracking** - PENDING → IN_TRANSIT → RECEIVED
3. **Dual Movements** - OUT from source, IN to destination
4. **Batch Tracking** - maintain batch information during transfer

### Database Model (Already Defined)

```go
type StockTransfer struct {
    ID                  string
    TenantID            string
    TransferNumber      string          // Auto-generated
    TransferDate        time.Time
    SourceWarehouseID   string
    DestinationWarehouseID string
    Status              StockTransferStatus // PENDING, IN_TRANSIT, RECEIVED, CANCELLED
    Notes               *string
    RequestedBy         string
    ApprovedBy          *string
    ApprovedAt          *time.Time
    ReceivedBy          *string
    ReceivedAt          *time.Time
    CreatedAt           time.Time
    UpdatedAt           time.Time

    // Relations
    SourceWarehouse      Warehouse
    DestinationWarehouse Warehouse
    Items                []StockTransferItem
}

type StockTransferItem struct {
    ID              string
    StockTransferID string
    ProductID       string
    ProductUnitID   *string
    BatchID         *string
    Quantity        decimal.Decimal // In selected unit
    ReceivedQty     *decimal.Decimal // Actual received (may differ)
    Notes           *string
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/inventory/stock-transfers            # Create transfer
GET    /api/v1/inventory/stock-transfers             # List
GET    /api/v1/inventory/stock-transfers/:id         # Details
POST   /api/v1/inventory/stock-transfers/:id/receive # Receive at destination
DELETE /api/v1/inventory/stock-transfers/:id         # Cancel (PENDING only)
```

### Business Logic

```go
// Receive stock transfer → create dual movements
func (s *StockTransferService) ReceiveTransfer(ctx context.Context, tenantID, transferID string) error {
    transfer := getStockTransfer(transferID)

    if transfer.Status != TransferStatusInTransit {
        return errors.New("can only receive IN_TRANSIT transfers")
    }

    tx := s.db.Begin()

    for _, item := range transfer.Items {
        baseQty := item.Quantity.Mul(item.ProductUnit.ConversionRate)

        // Movement OUT from source warehouse (already created when transfer initiated)
        // Now create movement IN to destination warehouse

        var destStock models.WarehouseStock
        err := tx.Where("warehouse_id = ? AND product_id = ?", transfer.DestinationWarehouseID, item.ProductID).First(&destStock).Error

        if err != nil {
            // Create if doesn't exist
            destStock = models.WarehouseStock{
                WarehouseID: transfer.DestinationWarehouseID,
                ProductID:   item.ProductID,
                Quantity:    baseQty,
            }
            tx.Create(&destStock)
        } else {
            stockBefore := destStock.Quantity
            destStock.Quantity = destStock.Quantity.Add(baseQty)
            tx.Save(&destStock)

            // Create movement IN
            movement := &models.InventoryMovement{
                TenantID:      tenantID,
                WarehouseID:   transfer.DestinationWarehouseID,
                ProductID:     item.ProductID,
                BatchID:       item.BatchID,
                MovementType:  "TRANSFER_IN",
                ReferenceType: "STOCK_TRANSFER",
                ReferenceID:   transferID,
                Quantity:      baseQty,
                StockBefore:   stockBefore,
                StockAfter:    destStock.Quantity,
                MovementDate:  transfer.TransferDate,
            }
            tx.Create(movement)
        }
    }

    // Update transfer status
    transfer.Status = TransferStatusReceived
    now := time.Now()
    transfer.ReceivedAt = &now
    tx.Save(&transfer)

    tx.Commit()
    return nil
}
```

---

## Implementation Priority (Week 7)

**Day 1:** Inventory Movement listing and history
**Day 2:** Manual stock adjustment (ADMIN feature)
**Day 3-4:** Stock Opname (create, count, complete)
**Day 5:** Stock Transfer (create, receive)

---

## Testing Checklist

### Inventory Movements
- [ ] Auto-created from delivery (OUT movement)
- [ ] Auto-created from goods receipt (IN movement)
- [ ] Manual adjustment (ADJUSTMENT movement)
- [ ] Stock before/after correctly recorded
- [ ] Movement history filterable by product/warehouse/date

### Stock Opname
- [ ] Create opname → auto-populates products
- [ ] Record physical counts
- [ ] Calculate variance (physical - system)
- [ ] Complete opname → generates adjustments
- [ ] Warehouse stock updated to physical count

### Stock Transfer
- [ ] Create transfer request
- [ ] Initiate transfer → OUT movement from source
- [ ] Receive transfer → IN movement to destination
- [ ] Batch information preserved during transfer
- [ ] Cannot transfer more than available stock

### Multi-Tenancy
- [ ] Tenant A cannot see Tenant B's movements/opname/transfers
- [ ] All movements filtered by tenantID

---

## Summary

**Inventory Management** provides:
1. ✅ Complete audit trail of all stock movements
2. ✅ Physical stock count reconciliation
3. ✅ Inter-warehouse stock transfers
4. ✅ Stock adjustment capabilities

**Critical Business Rules:**
- All stock changes must create inventory movements
- Stock before/after tracked for audit purposes
- Physical counts reconcile system stock via adjustments
- Transfers create dual movements (OUT + IN)

**Estimated Completion:** 1 week (5 days)

---

## MVP Completion

After completing Inventory Management (Week 7), the **complete MVP is ready** with:

1. ✅ **Foundation Setup** - Company & tenant configuration
2. ✅ **Master Data** - Products, customers, suppliers, warehouses
3. ✅ **Sales Flow** - Order → delivery → invoice → payment
4. ✅ **Purchase Flow** - Order → receipt → supplier payment
5. ✅ **Inventory Control** - Movements, stock counts, transfers

**Total Implementation Time:** 7 weeks (35 working days)

**Next Phase (Post-MVP):**
- Advanced reporting & analytics
- Automated email notifications
- Mobile app integration
- Barcode scanning
- Advanced pricing rules
- Production/manufacturing module
