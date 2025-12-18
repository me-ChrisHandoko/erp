# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Go-based Multi-Tenant ERP System** for Indonesian food distribution (Distribusi Sembako), implementing a SaaS subscription model with comprehensive warehouse, inventory, and financial management capabilities.

**Key Technologies:**
- Go 1.25.4
- GORM (Go ORM library)
- PostgreSQL/SQLite database
- Multi-tenant architecture with tenant isolation

## Architecture & Design Principles

### Multi-Tenancy Architecture

**Tenant Isolation Model:**
- Each PT/CV (company) operates as a separate tenant with isolated data
- All transactional tables include `tenantId` for data isolation
- `UserTenant` junction table enables users to access multiple tenants with different roles per tenant
- System admins have `isSystemAdmin` flag for cross-tenant management

**Subscription System:**
- Custom pricing per tenant (default: Rp 300,000/month)
- Trial period: 14 days free
- Grace period: 7 days after billing date
- Payment tracking with multiple methods (TRANSFER, VA, CREDIT_CARD, QRIS)
- Auto-renewal with cancellation support

### Database Schema Structure (Phase 1 Implementation)

**Core Modules:**
1. **User Management**: Multi-tenant access with per-tenant roles (OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF)
2. **Company Profile**: Indonesian tax compliance (NPWP, PKP status, PPN rates)
3. **Multi-Tenancy**: Tenant management, subscriptions, payments
4. **Warehouse Management**: Multi-warehouse support (MAIN, BRANCH, CONSIGNMENT, TRANSIT)
5. **Customer & Supplier**: Outstanding tracking with overdue monitoring
6. **Product & Inventory**: Batch/lot tracking with expiry dates for perishables
7. **Sales Order Flow**: SO → Delivery → Invoice with simplified status workflow
8. **Purchase Order Flow**: PO → Goods Receipt → Supplier Payment
9. **Inventory Control**: Movements, stock opname, inter-warehouse transfers
10. **Financial Management**: Cash transactions with running balance (Buku Kas)

### Critical Domain Concepts

**Multi-Unit Product System:**
- Products have a `baseUnit` (smallest unit: PCS, KG, LITER)
- `ProductUnit` table defines conversion rates (e.g., 1 KARTON = 24 PCS)
- Each unit can have different pricing, barcodes, and SKUs
- All stock calculations use base units internally

**Batch/Lot Tracking (Phase 0):**
- Enabled per-product via `isBatchTracked` and `isPerishable` flags
- Critical for food items with expiry dates
- `ProductBatch` tracks manufacture date, expiry date, supplier source
- FIFO/FEFO inventory management for perishable goods
- Batch status: AVAILABLE, RESERVED, EXPIRED, DAMAGED, RECALLED, SOLD

**Warehouse Stock Management:**
- Stock stored per warehouse via `WarehouseStock`
- Supports minimum/maximum stock levels
- Location tracking within warehouse (e.g., "RAK-A-01", "ZONE-B")
- Last count date and quantity tracking

**Goods Receipt Workflow (GRN):**
- Links PO to warehouse stock receipt
- Quality inspection tracking with accepted/rejected quantities
- Batch information captured during receipt
- Multiple GRN statuses: PENDING → RECEIVED → INSPECTED → ACCEPTED/REJECTED/PARTIAL

**Delivery & Fulfillment:**
- Simplified status flow: PREPARED → IN_TRANSIT → DELIVERED → CONFIRMED → CANCELLED
- Batch-level tracking for traceability
- Logistics info: driver, vehicle, departure/arrival times
- POD (Proof of Delivery): signature, photo proof, received by/at
- TTNK support for expedition services (JNE, Sicepat, JNT)

**Payment Tracking:**
- Customer receivables (piutang): tracked via `currentOutstanding`, `overdueAmount`
- Supplier payables (hutang): same tracking mechanism
- Check/Giro support with status tracking (ISSUED, CLEARED, BOUNCED, CANCELLED)
- Invoice payment status: UNPAID, PARTIAL, PAID, OVERDUE

**Indonesian Tax Compliance:**
- NPWP (Tax ID) support
- PKP status (Pengusaha Kena Pajak)
- Configurable PPN rate (default 11% as of 2025)
- Faktur Pajak series and SPPKP number tracking

## Development Commands

### Database Operations

**Run Database Migrations:**
```bash
go run cmd/migrate/main.go
```

**Seed Database with Test Data:**
```bash
go run cmd/seed/main.go
```

**Auto-Migrate Models (Development):**
```go
// In code using GORM AutoMigrate
db.AutoMigrate(&models.User{}, &models.Company{}, &models.Tenant{}, ...)
```

**Reset Database (CAUTION - Development only):**
```bash
rm dev.db  # For SQLite
# Or DROP DATABASE for PostgreSQL
```

### Go Development

```bash
# Initialize Go modules (if starting fresh)
go mod init backend

# Install dependencies
go mod tidy

# Run the application
go run main.go

# Build the application
go build -o bin/erp-backend

# Run tests
go test ./...

# Run specific test
go test ./path/to/package -v
```

## Code Patterns & Conventions

### Tenant Isolation Enforcement

**CRITICAL:** Every query on transactional tables MUST include `tenantId` filter to prevent cross-tenant data leakage.

```go
// CORRECT: Include tenantId in all queries
products, err := db.Product.FindMany(
    db.Product.TenantID.Equals(tenantID),
    db.Product.IsActive.Equals(true),
).Exec(ctx)

// WRONG: Missing tenantId filter
products, err := db.Product.FindMany(
    db.Product.IsActive.Equals(true),
).Exec(ctx)
```

### Stock Calculation Pattern

All stock operations use base units internally. Convert from transaction units to base units using `ProductUnit.conversionRate`:

```go
// Example: Convert from KARTON to PCS
// If 1 KARTON = 24 PCS (conversionRate = 24)
baseQuantity := orderQuantity * conversionRate

// Update stock in base units
warehouseStock.Quantity += baseQuantity
```

### Batch Tracking Implementation

For batch-tracked products (`isBatchTracked = true`), always specify `batchId` in inventory movements and delivery items:

```go
// Check if product requires batch tracking
if product.IsBatchTracked {
    // Must provide batchId for all inventory operations
    movement := db.InventoryMovement.Create(
        db.InventoryMovement.TenantID.Set(tenantID),
        db.InventoryMovement.ProductID.Set(productID),
        db.InventoryMovement.BatchID.Set(batchID), // Required
        // ... other fields
    )
}
```

### Expiry Date Monitoring

For perishable products (`isPerishable = true`), implement FEFO (First Expired, First Out) logic:

```go
// Query batches by expiry date (nearest expiry first)
batches, err := db.ProductBatch.FindMany(
    db.ProductBatch.ProductID.Equals(productID),
    db.ProductBatch.Status.Equals(db.BatchStatusAvailable),
).OrderBy(
    db.ProductBatch.ExpiryDate.Order(db.ASC),
).Exec(ctx)
```

### Invoice Numbering Format

Follow the configured format pattern from `Company.invoiceNumberFormat`:
- Default: `{PREFIX}/{NUMBER}/{MONTH}/{YEAR}` → "INV/001/12/2025"
- Ensure number padding (e.g., 001, 002, not 1, 2)
- Extract month/year from invoice date

### Role-Based Access Control (RBAC)

Implement per-tenant role checking via `UserTenant.role`:

```go
// Get user's role for specific tenant
userTenant, err := db.UserTenant.FindFirst(
    db.UserTenant.UserID.Equals(userID),
    db.UserTenant.TenantID.Equals(tenantID),
    db.UserTenant.IsActive.Equals(true),
).Exec(ctx)

// Role hierarchy: OWNER > ADMIN > FINANCE > SALES > WAREHOUSE > STAFF
// Example permissions:
// - OWNER/ADMIN: Full access
// - FINANCE: Invoice, payments, cash transactions
// - SALES: Sales orders, customers, deliveries
// - WAREHOUSE: Inventory, stock movements, goods receipts
// - STAFF: Read-only access
```

## Data Integrity Rules

### Outstanding Amount Tracking

Maintain `currentOutstanding` and `overdueAmount` consistency:

1. **Customer outstanding increases** when invoice created
2. **Customer outstanding decreases** when payment recorded
3. **Overdue amount** calculated daily for invoices past `dueDate`
4. **Supplier outstanding** follows same pattern for purchase orders

### Stock Movement Audit Trail

Every stock change MUST create an `InventoryMovement` record with:
- `stockBefore`: Stock quantity before change
- `stockAfter`: Stock quantity after change
- `referenceType` and `referenceId`: Source of change (e.g., "GOODS_RECEIPT", "DELIVERY")

### Subscription Payment Validation

Before allowing tenant access:
1. Check `Tenant.status` is ACTIVE or TRIAL
2. If TRIAL, verify `trialEndsAt > now()`
3. If ACTIVE, verify subscription not in PAST_DUE or EXPIRED status
4. Check `Subscription.gracePeriodEnds` if payment overdue

## Testing Considerations

### Critical Test Scenarios

1. **Tenant Isolation:** Verify queries cannot access other tenant's data
2. **Batch FEFO Logic:** Test expiry date ordering for batch selection
3. **Unit Conversion:** Validate base unit calculations across different units
4. **Outstanding Calculation:** Verify receivables/payables accuracy
5. **Stock Movement Audit:** Ensure all stock changes create movement records
6. **Subscription Status:** Test grace period and suspension workflows
7. **Multi-unit Pricing:** Validate correct price calculation per unit type
8. **GRN Quality Control:** Test accepted vs rejected quantity handling

## Security & Compliance

### Authentication & Authorization

- Hash passwords using bcrypt or argon2 (never store plaintext)
- Implement JWT or session-based authentication
- Validate `tenantId` in JWT claims to prevent tenant spoofing
- Check `UserTenant.isActive` and `User.isActive` before granting access

### Indonesian Tax Compliance

- Store NPWP in standardized format (XX.XXX.XXX.X-XXX.XXX)
- Apply PPN rate from `Company.ppnRate` (default 11%)
- Generate Faktur Pajak numbers using `Company.fakturPajakSeries`
- Track PKP status for tax-exempt vs taxable transactions

### Data Privacy

- Implement soft deletes (keep `isActive` flag) for audit trail
- Never hard delete transactional data
- Log all sensitive operations in `AuditLog`
- Encrypt sensitive fields (bank account numbers, tax IDs)

## Performance Optimization

### Database Indexing

Schema already includes critical indexes. Key patterns:
- Composite indexes on `[tenantId, code]` for unique constraints
- Date indexes for reporting (`invoiceDate`, `deliveryDate`, `opnameDate`)
- Status indexes for workflow queries
- Foreign key indexes for join optimization

### Query Optimization

- Always filter by `tenantId` first (indexed)
- Use batch operations for bulk inserts/updates
- Implement pagination for large result sets
- Cache frequently accessed configuration (Company settings)
- Pre-calculate totals (avoid SUM queries on large tables)

## Phase 0 vs Phase 1 Implementation

**Phase 0 Features (Already in Schema):**
- Multi-warehouse management
- Batch/lot tracking with expiry dates
- Goods Receipt workflow with quality inspection
- Stock opname (physical inventory count)
- Inter-warehouse stock transfers
- Outstanding tracking for customers/suppliers
- Simplified status workflows (7 states → 4-5 states)

**Phase 1 Features (Current):**
- Multi-tenancy with subscription system
- Custom pricing per tenant
- Trial period and grace period management
- Subscription payment tracking
- User-tenant access control

## Common Pitfalls to Avoid

1. **Missing Tenant Filter:** Always include `tenantId` in queries
2. **Wrong Unit Calculations:** Convert to base units before stock operations
3. **Batch Tracking Skip:** Check `isBatchTracked` before allowing batch-less operations
4. **Expiry Date Ignore:** Validate expiry dates for perishable products
5. **Outstanding Mismatch:** Update customer/supplier outstanding amounts when creating invoices/payments
6. **Stock Before/After Skip:** Always record stock before/after in inventory movements
7. **Number Format Violations:** Follow configured number formats (invoice, SO, PO, GRN)
8. **Hard Deletes:** Use soft deletes (isActive = false) for transactional data
