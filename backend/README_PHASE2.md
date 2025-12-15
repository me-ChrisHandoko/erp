# Phase 2 Implementation - Product & Inventory
## Multi-Tenant ERP System - GORM Migration

**Status:** ✅ COMPLETED (2025-12-16)
**Test Coverage:** 14/14 PASSED
**Schema Parity:** 100% with Prisma

---

## Quick Start

### 1. Run Phase 2 Migration

```go
package main

import (
    "backend/db"
    "log"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // Open database
    database, err := gorm.Open(sqlite.Open("erp.db"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Enable foreign keys (SQLite only)
    database.Exec("PRAGMA foreign_keys = ON")

    // Run Phase 1 migration (dependencies)
    if err := db.AutoMigratePhase1(database); err != nil {
        log.Fatal(err)
    }

    // Run Phase 2 migration
    if err := db.AutoMigratePhase2(database); err != nil {
        log.Fatal(err)
    }

    log.Println("✅ Phase 2 migration complete!")
}
```

### 2. Run Tests

```bash
# Run all Phase 2 tests
go test -v ./models/ -run "Phase2"

# Run all tests (Phase 1 + Phase 2)
go test -v ./models/
```

Expected output:
```
✅ TestPhase2SchemaGeneration
✅ TestCustomerCreation
✅ TestSupplierCreation
✅ TestWarehouseCreation
✅ TestProductCreation
✅ TestProductUnitConversion
✅ TestProductBatchTracking
✅ TestPriceListCustomerPricing
✅ TestProductSupplierRelationship
✅ TestWarehouseStockTracking
✅ TestUniqueConstraintsPhase2
✅ TestCascadeDeletePhase2
✅ TestDecimalPrecisionPhase2
✅ TestEnumValuesPhase2

PASS
ok      backend/models    0.437s
```

---

## What's Included

### 📦 Models (9 total)

#### Master Data
1. **Customer** - Customer master with outstanding tracking
2. **Supplier** - Supplier master with payment terms

#### Warehouse Management
3. **Warehouse** - Multi-warehouse management (MAIN, BRANCH, CONSIGNMENT, TRANSIT)
4. **WarehouseStock** - Stock per warehouse per product (actual stock tracking)

#### Product & Inventory
5. **Product** - Product master with batch tracking flags
6. **ProductUnit** - Multi-unit conversion (PCS, KARTON, LUSIN, SACK, etc.)
7. **ProductBatch** - Batch/lot tracking with expiry dates (FEFO/FIFO)
8. **PriceList** - Customer-specific pricing
9. **ProductSupplier** - Product-supplier junction with pricing

### 🎯 Features Implemented

- ✅ Multi-warehouse inventory tracking
- ✅ Multi-unit product system (base unit + conversions)
- ✅ Batch/lot tracking with expiry dates (FEFO/FIFO)
- ✅ Customer-specific pricing with effective dates
- ✅ Outstanding amount tracking (receivables/payables)
- ✅ Product-supplier relationships with lead time
- ✅ Warehouse location tracking ("RAK-A-01", "ZONE-B")
- ✅ Decimal precision for quantities (15,3) and money (15,2)
- ✅ Tenant isolation with CASCADE deletion
- ✅ Composite unique constraints [tenantID, code]

---

## Usage Examples

### Create Product with Multi-Unit

```go
// 1. Create base product
product := &models.Product{
    TenantID:       tenantID,
    Code:           "PROD-BERAS-001",
    Name:           "Beras Premium 5kg",
    Category:       stringPtr("BERAS"),
    BaseUnit:       "KG",
    BaseCost:       decimal.NewFromFloat(12000),
    BasePrice:      decimal.NewFromFloat(15000),
    MinimumStock:   decimal.NewFromFloat(100),
    IsBatchTracked: true,
    IsPerishable:   false,
    IsActive:       true,
}
db.Create(product)

// 2. Create SACK unit (1 SACK = 50 KG)
sackUnit := &models.ProductUnit{
    ProductID:      product.ID,
    UnitName:       "SACK",
    ConversionRate: decimal.NewFromFloat(50),
    BuyPrice:       decimalPtr(decimal.NewFromFloat(550000)),
    SellPrice:      decimalPtr(decimal.NewFromFloat(600000)),
    IsActive:       true,
}
db.Create(sackUnit)
```

### Create Warehouse and Stock

```go
// 1. Create warehouse
warehouse := &models.Warehouse{
    TenantID: tenantID,
    Code:     "WH-MAIN",
    Name:     "Gudang Utama",
    Type:     models.WarehouseTypeMain,
    Address:  stringPtr("Jl. Gudang Raya No. 123"),
    City:     stringPtr("Jakarta"),
    Province: stringPtr("DKI Jakarta"),
    Capacity: decimalPtr(decimal.NewFromFloat(1000.00)),
    IsActive: true,
}
db.Create(warehouse)

// 2. Create warehouse stock record
warehouseStock := &models.WarehouseStock{
    WarehouseID:  warehouse.ID,
    ProductID:    product.ID,
    Quantity:     decimal.NewFromFloat(0),
    MinimumStock: decimal.NewFromFloat(50),
    MaximumStock: decimal.NewFromFloat(500),
    Location:     stringPtr("RAK-A-01"),
}
db.Create(warehouseStock)
```

### Create Batch with Expiry (FEFO/FIFO)

```go
// Create batch for perishable product
mfgDate := time.Now()
expDate := time.Now().AddDate(0, 6, 0) // 6 months shelf life

batch := &models.ProductBatch{
    BatchNumber:      "BATCH-2025-12-001",
    ProductID:        product.ID,
    WarehouseStockID: warehouseStock.ID,
    ManufactureDate:  &mfgDate,
    ExpiryDate:       &expDate,
    Quantity:         decimal.NewFromFloat(100),
    Status:           models.BatchStatusAvailable,
    QualityStatus:    stringPtr("GOOD"),
    ReceiptDate:      time.Now(),
}
db.Create(batch)

// Update warehouse stock
warehouseStock.Quantity = warehouseStock.Quantity.Add(batch.Quantity)
db.Save(warehouseStock)
```

### Query Batches by Expiry (FEFO)

```go
// Get batches by nearest expiry date (First Expired, First Out)
var batches []models.ProductBatch
db.Where("product_id = ? AND status = ?", productID, models.BatchStatusAvailable).
   Order("expiry_date ASC"). // Nearest expiry first
   Find(&batches)

// Allocate from batch with nearest expiry
if len(batches) > 0 {
    batch := batches[0]
    deliveryItem.BatchID = &batch.ID
}
```

### Customer-Specific Pricing

```go
// Set customer-specific pricing (VIP discount)
vipPriceList := &models.PriceList{
    ProductID:     product.ID,
    CustomerID:    &vipCustomer.ID,
    Price:         decimal.NewFromFloat(14000), // Rp 1,000 discount
    MinQty:        decimal.NewFromFloat(10),
    EffectiveFrom: time.Now(),
    IsActive:      true,
}
db.Create(vipPriceList)

// Default pricing for all other customers
defaultPriceList := &models.PriceList{
    ProductID:     product.ID,
    CustomerID:    nil, // NULL = default
    Price:         decimal.NewFromFloat(15000),
    MinQty:        decimal.NewFromFloat(1),
    EffectiveFrom: time.Now(),
    IsActive:      true,
}
db.Create(defaultPriceList)
```

### Outstanding Amount Tracking

```go
// Create customer with credit limit
customer := &models.Customer{
    TenantID:           tenantID,
    Code:               "CUST-001",
    Name:               "Toko Berkah",
    PaymentTerm:        30, // Net 30 days
    CreditLimit:        decimal.NewFromFloat(50000000),
    CurrentOutstanding: decimal.NewFromFloat(0),
    OverdueAmount:      decimal.NewFromFloat(0),
    IsActive:           true,
}
db.Create(customer)

// Query customers with outstanding balance
var customersWithDebt []models.Customer
db.Where("current_outstanding > ?", 0).Find(&customersWithDebt)

// Query customers with overdue payments
var customersOverdue []models.Customer
db.Where("overdue_amount > ?", 0).Find(&customersOverdue)
```

---

## Database Schema

### Tables Created (9 total)

```
customers
├── id (varchar 255, PK)
├── tenant_id (varchar 255, FK → tenants, CASCADE)
├── code (varchar 100, unique with tenant_id)
├── name, type, contact info
├── payment_term (int, default 0 days)
├── credit_limit (decimal 15,2, default 0)
├── current_outstanding (decimal 15,2, default 0, indexed)
├── overdue_amount (decimal 15,2, default 0, indexed)
└── created_at, updated_at

suppliers
├── Same structure as customers
└── For supplier payables tracking

warehouses
├── id (varchar 255, PK)
├── tenant_id (varchar 255, FK → tenants, CASCADE)
├── code (varchar 50, unique with tenant_id)
├── name, type (MAIN, BRANCH, CONSIGNMENT, TRANSIT)
├── address, city, province, postal_code
├── manager_id (varchar 255, FK → users, optional)
├── capacity (decimal 15,2)
└── created_at, updated_at

products
├── id (varchar 255, PK)
├── tenant_id (varchar 255, FK → tenants, CASCADE)
├── code (varchar 100, unique with tenant_id)
├── name, category, base_unit
├── base_cost, base_price (decimal 15,2)
├── current_stock (decimal 15,3) ← DEPRECATED
├── minimum_stock (decimal 15,3)
├── barcode (varchar 100, unique globally)
├── is_batch_tracked, is_perishable (boolean)
└── created_at, updated_at

product_units
├── id (varchar 255, PK)
├── product_id (varchar 255, FK → products, CASCADE)
├── unit_name (varchar 50, unique with product_id)
├── conversion_rate (decimal 15,3) ← e.g., 1 KARTON = 40 PCS
├── buy_price, sell_price (decimal 15,2, optional)
├── barcode (varchar 100, indexed)
├── weight, volume (decimal 10,3, optional)
└── created_at, updated_at

price_list
├── id (varchar 255, PK)
├── product_id (varchar 255, FK → products, CASCADE)
├── customer_id (varchar 255, FK → customers, NULL = default)
├── price (decimal 15,2)
├── min_qty (decimal 15,3, default 0)
├── effective_from, effective_to (datetime)
└── created_at, updated_at

product_suppliers
├── id (varchar 255, PK)
├── product_id (varchar 255, unique with supplier_id)
├── supplier_id (varchar 255, unique with product_id)
├── supplier_price (decimal 15,2)
├── lead_time (int, default 7 days)
├── is_primary (boolean, default false)
└── created_at, updated_at

warehouse_stocks ← ACTUAL STOCK TABLE
├── id (varchar 255, PK)
├── warehouse_id (varchar 255, unique with product_id)
├── product_id (varchar 255, unique with warehouse_id)
├── quantity (decimal 15,3, indexed)
├── minimum_stock, maximum_stock (decimal 15,3)
├── location (varchar 100) ← e.g., "RAK-A-01", "ZONE-B"
├── last_count_date, last_count_qty (datetime, decimal)
└── created_at, updated_at

product_batches
├── id (varchar 255, PK)
├── batch_number (varchar 100, unique with product_id)
├── product_id (varchar 255, FK → products, CASCADE)
├── warehouse_stock_id (varchar 255, FK → warehouse_stocks, CASCADE)
├── manufacture_date (datetime, optional)
├── expiry_date (datetime, indexed) ← CRITICAL for FEFO
├── quantity (decimal 15,3)
├── status (varchar 20, AVAILABLE/RESERVED/EXPIRED/DAMAGED/RECALLED/SOLD)
├── quality_status (varchar 20, GOOD/DAMAGED/QUARANTINE)
└── created_at, updated_at
```

---

## Key Patterns

### 1. Multi-Unit Conversion

All stock operations use **base units** internally:

```go
// Customer orders "2 KARTON" (1 KARTON = 40 PCS)
orderQty := 2
conversionRate := 40
baseQty := orderQty * conversionRate // 80 PCS

// Update stock in base units
warehouseStock.Quantity -= baseQty
```

### 2. FEFO/FIFO Batch Allocation

For perishable products, allocate from batch with **nearest expiry date**:

```go
// Query batches ordered by expiry date
db.Where("product_id = ? AND status = ?", productID, "AVAILABLE").
   Order("expiry_date ASC"). // First Expired, First Out
   Find(&batches)

// Allocate from first batch (nearest expiry)
batch := batches[0]
```

### 3. Warehouse Stock vs Product Stock

**DEPRECATED:** `Product.currentStock` field
**CORRECT:** Use `WarehouseStock` table

```go
// WRONG: Do not use Product.currentStock
product.CurrentStock -= qty

// RIGHT: Update WarehouseStock per warehouse
warehouseStock := getWarehouseStock(warehouseID, productID)
warehouseStock.Quantity -= qty
```

### 4. Tenant Isolation

All queries **MUST** filter by `tenantID`:

```go
// CORRECT: Include tenantID filter
products, err := db.Product.FindMany(
    db.Product.TenantID.Equals(tenantID),
    db.Product.IsActive.Equals(true),
).Exec(ctx)

// WRONG: Missing tenantID (data leak!)
products, err := db.Product.FindMany(
    db.Product.IsActive.Equals(true),
).Exec(ctx)
```

---

## Testing

### Run All Tests

```bash
# Phase 2 only
go test -v ./models/ -run "Phase2"

# All phases
go test -v ./models/
```

### Test Coverage

Phase 2 test suite validates:
- ✅ Schema generation (9 tables)
- ✅ CUID auto-generation
- ✅ Model creation with all fields
- ✅ Multi-unit conversion
- ✅ Batch tracking with expiry dates
- ✅ Customer-specific pricing
- ✅ Product-supplier relationships
- ✅ Warehouse stock tracking
- ✅ Composite unique constraints
- ✅ CASCADE deletion from Tenant
- ✅ Decimal precision (15,2 and 15,3)
- ✅ Enum values (WarehouseType, BatchStatus)

---

## Production Checklist

### Before Deployment

- [ ] Switch from SQLite to PostgreSQL
- [ ] Test on PostgreSQL database
- [ ] Verify multi-tenant isolation
- [ ] Test FEFO batch allocation logic
- [ ] Validate unit conversion calculations
- [ ] Load test with realistic product catalog
- [ ] Test outstanding amount calculations
- [ ] Verify CASCADE deletion on production backup

### Security

- [ ] Implement tenant isolation middleware
- [ ] Validate tenantID in all queries
- [ ] Encrypt sensitive customer data (NPWP)
- [ ] Add audit logging for stock movements
- [ ] Implement rate limiting for API endpoints

### Performance

- [ ] Add indexes based on query patterns
- [ ] Optimize batch expiry queries
- [ ] Cache product unit conversions
- [ ] Consider read replicas for reporting
- [ ] Monitor slow queries and optimize

---

## Next Steps

### Phase 3: Transactions (Estimated: 2 days)

- [ ] SalesOrder, SalesOrderItem
- [ ] Invoice, InvoiceItem, Payment
- [ ] PurchaseOrder, PurchaseOrderItem
- [ ] GoodsReceipt, GoodsReceiptItem
- [ ] Delivery, DeliveryItem
- [ ] SupplierPayment, PaymentCheck

### Phase 4: Supporting Modules (Estimated: 1 day)

- [ ] InventoryMovement (stock audit trail)
- [ ] StockOpname, StockOpnameItem (physical count)
- [ ] StockTransfer, StockTransferItem (inter-warehouse)
- [ ] CashTransaction (Buku Kas)
- [ ] Setting, AuditLog

---

## Troubleshooting

### Import Cycle Error

**Problem:** `import cycle not allowed in test`
**Solution:** Tests should not import `backend/db`, use local migration functions instead

### Decimal Precision Loss

**Problem:** Money amounts showing incorrect decimals
**Solution:** Always use `decimal.Decimal` type, never `float64`

### Unique Constraint Violation

**Problem:** Duplicate code error when code should be unique per tenant
**Solution:** Ensure composite unique index `[tenantID, code]` is applied

### CASCADE Deletion Not Working

**Problem:** Child records not deleted with parent
**Solution:** Use `db.Select("RelationName").Delete()` for has-many relations

### Batch Not Found for Tracked Product

**Problem:** Product requires batch but none available
**Solution:** Check `product.IsBatchTracked` flag before requiring batch

---

## Documentation

- **Migration Guide:** `claudedocs/prisma-to-gorm-migration-guide.md`
- **Phase 1 Summary:** `claudedocs/PHASE1_IMPLEMENTATION_SUMMARY.md`
- **Phase 2 Summary:** `claudedocs/PHASE2_IMPLEMENTATION_SUMMARY.md`
- **Phase 1 README:** `README_PHASE1.md`
- **Phase 2 README:** `README_PHASE2.md` (this file)
- **Test Suite:** `models/phase2_test.go`

---

**Phase 2 Status:** ✅ COMPLETED
**Last Updated:** 2025-12-16
**Version:** 1.0
