# MVP Roadmap - Post-Authentication Implementation Plan

## Executive Summary

**Status:** ✅ Authentication Module Complete
**Next Phase:** Core Business Modules (MVP)
**Approach:** Build minimum viable features for each module, following business flow priority

---

## What's Already Built ✅

### 1. **Authentication & Security (COMPLETE)**
- ✅ Login with brute force protection (5 failed attempts = 30 min lockout)
- ✅ JWT token management (access + refresh tokens)
- ✅ Password reset flow (forgot password → email → reset)
- ✅ Email verification
- ✅ Multi-tenant context switching
- ✅ Change password (authenticated users)
- ✅ Account unlock (system admin)
- ✅ CSRF protection middleware
- ✅ Rate limiting (auth: 10/min, general: 60/min)
- ✅ Remember me functionality

### 2. **Database Models (COMPLETE)**
All 23 database models defined with relationships:
- ✅ User & UserTenant (multi-tenant access control)
- ✅ Tenant & Subscription (SaaS billing)
- ✅ Company & CompanyBank
- ✅ Customer & Supplier
- ✅ Product, ProductUnit, ProductBatch, PriceList, WarehouseStock
- ✅ SalesOrder, SalesOrderItem
- ✅ PurchaseOrder, PurchaseOrderItem
- ✅ Delivery, DeliveryItem
- ✅ Invoice, InvoiceItem, InvoicePayment
- ✅ GoodsReceipt, GoodsReceiptItem
- ✅ SupplierPayment
- ✅ InventoryMovement, StockOpname, StockTransfer
- ✅ CashTransaction
- ✅ AuditLog

### 3. **Infrastructure (COMPLETE)**
- ✅ Multi-tenant architecture with tenant isolation
- ✅ PostgreSQL/SQLite database with GORM
- ✅ Middleware: JWT auth, tenant context, CSRF, rate limiting, CORS
- ✅ Configuration management (env-based)
- ✅ Error handling with custom AppError types
- ✅ Password hashing (Argon2)
- ✅ Email service integration
- ✅ Background job scheduler
- ✅ Health check endpoints
- ✅ Seed data system

---

## What Needs to Be Built (MVP) 🚧

### **Implementation Priority: Follow Business Flow**

The modules are organized in **5 implementation groups**, each building on the previous:

1. **Foundation Setup** → Tenant & Company configuration
2. **Master Data** → Products, Customers, Suppliers, Warehouses
3. **Sales Flow** → Orders → Deliveries → Invoices → Payments
4. **Purchase Flow** → Orders → Receipts → Supplier Payments
5. **Inventory Control** → Stock movements, transfers, stock counts

---

## MVP Module Groups

### 📁 Group 1: Foundation Setup (Week 1)
**File:** `01-TENANT-COMPANY-SETUP.md`

**Modules:**
1. **Company Profile Management**
   - CRUD for company details (name, NPWP, PKP status, address)
   - Bank account management
   - Invoice/SO/PO number format configuration
   - Tax settings (PPN rate, Faktur Pajak series)

2. **Tenant Management** (OWNER/ADMIN only)
   - View tenant details & subscription status
   - Update tenant settings
   - User-tenant role management (invite users, assign roles)

**Why First:** Required for all subsequent modules - sets up company configuration and multi-tenant access control.

---

### 📁 Group 2: Master Data (Week 2-3)
**File:** `02-MASTER-DATA-MANAGEMENT.md`

**Modules:**
3. **Product Management**
   - CRUD products (code, name, category, base unit)
   - Multi-unit management (conversions: 1 KARTON = 24 PCS)
   - Batch/lot tracking setup (for perishable items)
   - Product pricing (base cost, base price per unit)
   - Barcode management

4. **Customer Management**
   - CRUD customers (code, name, type, contact info)
   - Payment terms & credit limits
   - Outstanding tracking (receivables)
   - Customer-specific pricing (price list)

5. **Supplier Management**
   - CRUD suppliers (code, name, type, contact info)
   - Payment terms & credit limits
   - Outstanding tracking (payables)
   - Product-supplier relationships

6. **Warehouse Management**
   - CRUD warehouses (code, name, type, address)
   - Warehouse stock initialization
   - Stock location management

**Why Second:** Foundation data required for all transactions - can't sell products you haven't defined.

---

### 📁 Group 3: Sales Flow (Week 4-5)
**File:** `03-SALES-FLOW.md`

**Modules:**
7. **Sales Order Module**
   - Create/Edit/Cancel sales orders
   - Add/remove items with pricing
   - Status: DRAFT → CONFIRMED → COMPLETED/CANCELLED
   - Auto-calculate totals (subtotal, discount, tax, total)
   - Generate SO numbers (configurable format)

8. **Delivery Module** (Simplified for MVP)
   - Create delivery from sales order
   - Batch selection (FEFO for perishables)
   - Status: PREPARED → DELIVERED → CONFIRMED
   - Proof of delivery (signature, photo)

9. **Invoice Module**
   - Generate invoice from sales order/delivery
   - Invoice items with tax calculation
   - Generate invoice numbers (configurable format)
   - Invoice status: UNPAID → PARTIAL → PAID → OVERDUE

10. **Customer Payment Module**
    - Record customer payments (cash, transfer, check/giro)
    - Apply payments to invoices
    - Update customer outstanding amounts
    - Payment status tracking

**Why Third:** Core revenue-generating business flow - MVP needs to sell products and collect payments.

---

### 📁 Group 4: Purchase Flow (Week 6)
**File:** `04-PURCHASE-FLOW.md`

**Modules:**
11. **Purchase Order Module**
    - Create/Edit/Cancel purchase orders
    - Add/remove items with supplier pricing
    - Status: DRAFT → CONFIRMED → COMPLETED/CANCELLED
    - Auto-calculate totals
    - Generate PO numbers

12. **Goods Receipt Module** (Simplified for MVP)
    - Create goods receipt from purchase order
    - Record batch information (manufacture date, expiry date)
    - Quality inspection (accepted/rejected quantities)
    - Update warehouse stock
    - Status: PENDING → RECEIVED → ACCEPTED

13. **Supplier Payment Module**
    - Record supplier payments (cash, transfer, check/giro)
    - Apply payments to purchase orders
    - Update supplier outstanding amounts
    - Payment status tracking

**Why Fourth:** Needed to replenish inventory - can be deferred initially if starting with existing stock.

---

### 📁 Group 5: Inventory Management (Week 7)
**File:** `05-INVENTORY-MANAGEMENT.md`

**Modules:**
14. **Inventory Movement Tracking**
    - Auto-create movements from deliveries/receipts
    - View movement history per product/warehouse
    - Stock before/after tracking

15. **Stock Opname (Physical Count)**
    - Record physical stock count
    - Compare with system stock
    - Generate adjustment movements
    - Variance reporting

16. **Stock Transfer** (Inter-warehouse)
    - Transfer stock between warehouses
    - Update stock in both warehouses
    - Movement tracking

**Why Last:** Essential for inventory control but MVP can function with basic stock tracking first.

---

## Implementation Sequence (7 Weeks)

```
Week 1: Foundation Setup
├─ Company Profile CRUD
└─ Tenant Management (user-tenant roles)

Week 2-3: Master Data
├─ Product Management (CRUD + multi-unit)
├─ Customer Management (CRUD + outstanding)
├─ Supplier Management (CRUD + outstanding)
└─ Warehouse Management (CRUD + stock init)

Week 4-5: Sales Flow (REVENUE PATH - PRIORITY)
├─ Sales Order CRUD
├─ Delivery (simplified)
├─ Invoice Generation
└─ Customer Payment

Week 6: Purchase Flow
├─ Purchase Order CRUD
├─ Goods Receipt (simplified)
└─ Supplier Payment

Week 7: Inventory Control
├─ Movement Tracking (auto-create)
├─ Stock Opname
└─ Stock Transfer
```

---

## MVP Scope Boundaries

### ✅ **IN SCOPE (MVP)**
- Basic CRUD for all master data
- Single-warehouse operations initially
- Manual batch selection (no auto-FEFO)
- Simple tax calculation (PPN rate × subtotal)
- Cash & transfer payments only
- Basic reporting (lists, totals)
- Manual stock count entry

### ❌ **OUT OF SCOPE (MVP)**
- Advanced reporting & analytics (Phase 2)
- Automated email notifications (Phase 2)
- Mobile app (Phase 3)
- Barcode scanning (Phase 3)
- Advanced pricing rules (Phase 2)
- Promotion & discount engine (Phase 2)
- Multi-currency support (Phase 4)
- Integration with accounting software (Phase 3)
- Advanced permissions (per-module roles) (Phase 2)
- Automated stock reordering (Phase 3)
- Production/manufacturing module (Phase 4)

---

## Success Criteria (MVP Complete)

1. ✅ **Company can onboard** → Create company profile, configure settings
2. ✅ **Add products** → Define products with units and pricing
3. ✅ **Add customers** → Manage customer database with payment terms
4. ✅ **Create sales orders** → Record customer orders with items
5. ✅ **Deliver products** → Track deliveries with batch information
6. ✅ **Generate invoices** → Create invoices from sales orders
7. ✅ **Collect payments** → Record customer payments and update receivables
8. ✅ **Purchase from suppliers** → Create purchase orders and receive goods
9. ✅ **Pay suppliers** → Record supplier payments and update payables
10. ✅ **Track inventory** → View stock levels and movement history

---

## Technical Approach (All Modules)

### 1. **Backend Structure (per module)**
```
internal/
├── handler/
│   ├── product_handler.go      # HTTP handlers
│   ├── customer_handler.go
│   └── ...
├── service/
│   ├── product/
│   │   ├── product_service.go  # Business logic
│   │   ├── validation.go       # Business rules
│   │   └── models.go           # Service DTOs
│   └── ...
├── dto/
│   ├── product_dto.go          # Request/Response DTOs
│   └── ...
└── router/
    └── router.go               # Route registration
```

### 2. **Standard CRUD Pattern**
```go
// Create
POST   /api/v1/{module}           → Create{Module}
// Read
GET    /api/v1/{module}           → List{Module} (with pagination, filters)
GET    /api/v1/{module}/:id       → Get{Module}
// Update
PUT    /api/v1/{module}/:id       → Update{Module}
PATCH  /api/v1/{module}/:id       → PartialUpdate{Module}
// Delete
DELETE /api/v1/{module}/:id       → Delete{Module} (soft delete: isActive=false)
```

### 3. **Middleware Stack (All Protected Routes)**
```go
protected.Use(middleware.RateLimitMiddleware(redisClient, 60))
protected.Use(middleware.JWTAuthMiddleware(tokenService))
protected.Use(middleware.TenantContextMiddleware(db))
protected.Use(middleware.CSRFMiddleware())
```

### 4. **Tenant Isolation (CRITICAL)**
```go
// ALWAYS filter by tenantID in queries
products, err := db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&products).Error
```

### 5. **Response Format (Standardized)**
```json
{
  "success": true,
  "data": { ... },
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

---

## Documentation Structure

Each module group has a dedicated markdown file with:

1. **Module Overview** - Purpose and business context
2. **Database Models** - Already defined (reference from models/)
3. **API Endpoints** - RESTful routes with request/response examples
4. **Business Logic** - Service layer implementation details
5. **Validation Rules** - Input validation and business constraints
6. **Tenant Isolation** - How to enforce multi-tenant filtering
7. **Testing Checklist** - Critical test scenarios
8. **Implementation Priority** - Order of feature development

---

## Next Steps

1. **Read each module group file** in sequence:
   - `01-TENANT-COMPANY-SETUP.md`
   - `02-MASTER-DATA-MANAGEMENT.md`
   - `03-SALES-FLOW.md`
   - `04-PURCHASE-FLOW.md`
   - `05-INVENTORY-MANAGEMENT.md`

2. **Start with Group 1 (Foundation Setup)** - Week 1 focus

3. **Follow implementation pattern:**
   - Define DTOs (request/response)
   - Implement service layer (business logic + validation)
   - Create handlers (HTTP layer)
   - Register routes
   - Add middleware (tenant context, role-based if needed)
   - Write tests (unit + integration)

4. **Deploy incrementally:**
   - Each module should be deployable independently
   - Use feature flags if needed for gradual rollout

---

## Risk Mitigation

### **Risk 1: Scope Creep**
**Mitigation:** Strictly follow MVP boundaries. Defer non-essential features to Phase 2.

### **Risk 2: Tenant Data Leakage**
**Mitigation:**
- Always include `tenantID` in WHERE clauses
- Use `TenantContextMiddleware` to inject tenant from JWT
- Add integration tests for cross-tenant isolation

### **Risk 3: Complex Business Logic**
**Mitigation:**
- Start with simple workflows (e.g., no auto-FEFO in MVP)
- Document business rules clearly
- Use service layer to isolate complexity

### **Risk 4: Performance Issues**
**Mitigation:**
- Implement pagination for all list endpoints
- Add database indexes (already defined in models)
- Use eager loading for related entities
- Monitor query performance

---

## Questions & Clarifications

Before starting implementation, clarify:

1. **Stock Management:** Start with single warehouse or multi-warehouse from day 1?
   - **Recommendation:** Single warehouse for MVP, add multi-warehouse in Week 7

2. **Batch Tracking:** Mandatory for all products or optional?
   - **Current Design:** Optional via `isBatchTracked` flag - OK for MVP

3. **Tax Calculation:** Support multiple tax rates or single PPN rate?
   - **Current Design:** Company-wide PPN rate (11%) - OK for MVP

4. **Payment Methods:** Cash, transfer, check/giro - all needed in MVP?
   - **Recommendation:** Cash + transfer for MVP, add check/giro in Phase 2

5. **User Roles:** Which roles need access to which modules?
   - **Recommendation:** Defer detailed RBAC to Phase 2, use basic role checks (OWNER/ADMIN for setup, SALES for sales, etc.)

---

## Conclusion

This MVP roadmap prioritizes **business-critical features** in a logical sequence:
1. Setup (company config)
2. Master data (products, customers, suppliers)
3. Revenue flow (sales → invoice → payment)
4. Cost flow (purchase → receipt → payment)
5. Inventory control (movements, counts, transfers)

**Total Estimated Timeline:** 7 weeks (5 sprints)
**Delivery Strategy:** Incremental - each group deliverable independently
**Success Metric:** Complete sales cycle (order → delivery → invoice → payment) functional

Proceed to `01-TENANT-COMPANY-SETUP.md` to begin implementation.
