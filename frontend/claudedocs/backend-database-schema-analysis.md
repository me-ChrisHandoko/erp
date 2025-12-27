# Backend Database Schema Analysis - Multi-Company Architecture

**Tanggal**: 2025-12-26
**Analisis**: Backend Database Structure Check
**Fokus**: Apakah backend sudah mendukung multi-company architecture?

---

## 🚨 EXECUTIVE SUMMARY

### Jawaban Singkat: **TIDAK**

Backend database schema saat ini **TIDAK MENDUKUNG** multi-company architecture sama sekali. Semua analisis multi-company yang dilakukan sebelumnya adalah **theoretical design**, bukan current implementation.

**Current State**: Single-company-per-tenant architecture
**Required State**: Multi-company-per-tenant architecture
**Gap**: Fundamental database structure mismatch

---

## 📊 CRITICAL FINDINGS

### 1. Tenant-Company Relationship (FUNDAMENTAL FLAW)

#### Current Implementation

**File**: `backend/models/tenant.go:15`
```go
type Tenant struct {
  ID             string `gorm:"primaryKey"`
  CompanyID      string `gorm:"uniqueIndex;not null;index"` // ← PROBLEM!
  SubscriptionID *string
  // ...

  // Relations
  Company      Company       `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
  Subscription *Subscription `gorm:"foreignKey:SubscriptionID"`
  Users        []UserTenant  `gorm:"foreignKey:TenantID"`
}
```

**File**: `backend/models/company.go:75`
```go
type Company struct {
  ID string `gorm:"primaryKey"`
  // ... fields

  // Relations
  Tenant *Tenant       `gorm:"foreignKey:CompanyID"` // One-to-one
  Banks  []CompanyBank `gorm:"foreignKey:CompanyID"`
}
```

#### Problem Analysis

**❌ Issues Identified**:
1. `Tenant.CompanyID` memiliki constraint `uniqueIndex`
2. Ini **enforces 1:1 relationship** (One Tenant ↔ One Company)
3. Comment di code mengkonfirmasi: `// One-to-one`
4. Relationship direction SALAH (CompanyID di Tenant, harusnya TenantID di Company)

**Current Architecture** (WRONG):
```
Company (1) ←→ (1) Tenant ←→ (1) Subscription
```

**Required Architecture** (CORRECT):
```
Tenant (1) → (N) Companies
Tenant (1) ← (1) Subscription
Users (M) ← (N) Companies (via user_company_roles)
```

#### Required Changes

```go
// tenant.go - REMOVE CompanyID
type Tenant struct {
  ID   string `gorm:"primaryKey"`
  Name string `gorm:"not null"` // Tenant business name
  // ... other fields

  // Relations
  Companies []Company     `gorm:"foreignKey:TenantID"` // ← One-to-Many
  Users     []UserTenant  `gorm:"foreignKey:TenantID"`
}

// company.go - ADD TenantID
type Company struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← ADD THIS
  // ... other fields

  // Relations
  Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
  Banks  []CompanyBank `gorm:"foreignKey:CompanyID"`
}
```

---

### 2. User Permission System (MISSING TABLE)

#### Current Implementation

**File**: `backend/models/user.go:42-56`
```go
// UserTenant - Junction table for User ↔ Tenant with per-tenant role
type UserTenant struct {
  ID       string   `gorm:"primaryKey"`
  UserID   string   `gorm:"not null;index;uniqueIndex:idx_user_tenant"`
  TenantID string   `gorm:"not null;index;uniqueIndex:idx_user_tenant"`
  Role     UserRole `gorm:"type:varchar(20);default:'STAFF';index"` // ← TENANT-level role
  IsActive bool     `gorm:"default:true"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // Relations
  User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
  Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}
```

#### Problem Analysis

**❌ Issues Identified**:
1. Role defined at **TENANT level**, not COMPANY level
2. User has **same role across all companies** (if multiple existed)
3. No granular permission per company
4. **Table `user_company_roles` DOES NOT EXIST**

**Real-World Scenario yang TIDAK BISA di-support**:
- User A: ADMIN di PT Distribusi Utama, STAFF di CV Sembako Jaya ← IMPOSSIBLE
- Current system: User A hanya bisa punya SATU role untuk seluruh tenant

#### Required Changes

**Create New Table**: `user_company_roles`

```go
// user_company_role.go - NEW FILE
package models

type UserCompanyRole struct {
  ID        string   `gorm:"primaryKey"`
  UserID    string   `gorm:"not null;index;uniqueIndex:idx_user_company"`
  CompanyID string   `gorm:"not null;index;uniqueIndex:idx_user_company"` // ← COMPANY-level
  Role      UserRole `gorm:"type:varchar(20);default:'STAFF';index"`
  IsActive  bool     `gorm:"default:true"`
  CreatedAt time.Time
  UpdatedAt time.Time

  // Relations
  User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
  Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

func (UserCompanyRole) TableName() string {
  return "user_company_roles"
}
```

**SQL Migration**:
```sql
CREATE TABLE user_company_roles (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  company_id VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL CHECK (role IN ('OWNER', 'ADMIN', 'FINANCE', 'SALES', 'WAREHOUSE', 'STAFF')),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY idx_user_company (user_id, company_id),
  INDEX idx_user (user_id),
  INDEX idx_company (company_id),
  INDEX idx_role (role),

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);
```

---

### 3. Data Scoping (ALL TABLES AFFECTED - 20+ Tables)

#### Current Implementation Pattern

**ALL transactional tables** hanya punya `TenantID`, **TIDAK ada `CompanyID`**:

**Warehouse** (`backend/models/warehouse.go:15`):
```go
type Warehouse struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← Only TenantID
  Code     string `gorm:"not null"`
  // ...
}
```

**Product** (`backend/models/product.go:15`):
```go
type Product struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← Only TenantID
  Code     string `gorm:"not null"`
  // ...
}
```

**SalesOrder** (`backend/models/sales.go:15`):
```go
type SalesOrder struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← Only TenantID
  // ...
}
```

**Customer** (`backend/models/master.go:15`):
```go
type Customer struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← Only TenantID
  Code     string `gorm:"not null"`
  // ...
}
```

**Supplier** (`backend/models/master.go:60`):
```go
type Supplier struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← Only TenantID
  Code     string `gorm:"not null"`
  // ...
}
```

#### Tables Affected (Complete List)

**Core Master Data**:
1. ❌ `products` - No CompanyID
2. ❌ `product_units` - No CompanyID
3. ❌ `product_batches` - No CompanyID
4. ❌ `product_suppliers` - No CompanyID
5. ❌ `price_list` - No CompanyID
6. ❌ `warehouses` - No CompanyID
7. ❌ `warehouse_stocks` - No CompanyID
8. ❌ `customers` - No CompanyID
9. ❌ `suppliers` - No CompanyID

**Sales & Procurement**:
10. ❌ `sales_orders` - No CompanyID
11. ❌ `sales_order_items` - No CompanyID
12. ❌ `purchase_orders` - No CompanyID (likely)
13. ❌ `purchase_order_items` - No CompanyID (likely)

**Finance**:
14. ❌ `invoices` - No CompanyID (likely)
15. ❌ `invoice_items` - No CompanyID (likely)
16. ❌ `cash_transactions` - No CompanyID
17. ❌ `supplier_payments` - No CompanyID

**Inventory**:
18. ❌ `inventory_movements` - No CompanyID
19. ❌ `deliveries` - No CompanyID (likely)
20. ❌ `delivery_items` - No CompanyID (likely)
21. ❌ `goods_receipts` - No CompanyID
22. ❌ `goods_receipt_items` - No CompanyID (likely)
23. ❌ `stock_opnames` - No CompanyID
24. ❌ `stock_transfers` - No CompanyID

**Total: 20+ tables need modification**

#### Problem Analysis

**❌ Issues Identified**:
1. ALL data scoped to `TenantID` only
2. NO company-level data isolation
3. Cannot separate data between companies within same tenant
4. **CRITICAL SECURITY RISK**: Data leakage potential between companies

**Impact Example**:
```
Scenario:
- Tenant "Distribusi Group"
  - Company A: PT Distribusi Utama
  - Company B: CV Sembako Jaya

Current Backend:
- Product "Beras Premium" has TenantID only
- Both Company A and Company B see the SAME product
- Cannot have separate product catalogs per company
- Cannot isolate inventory between companies
```

#### Required Changes

**Pattern for EVERY transactional model**:

```go
// BEFORE
type Product struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index;uniqueIndex:idx_product_tenant_code"`
  Code     string `gorm:"not null;uniqueIndex:idx_product_tenant_code"`
  // ... fields

  Tenant Tenant `gorm:"foreignKey:TenantID"`
}

// AFTER
type Product struct {
  ID        string `gorm:"primaryKey"`
  TenantID  string `gorm:"not null;index"` // ← Keep for tenant isolation
  CompanyID string `gorm:"not null;index;uniqueIndex:idx_product_company_code"` // ← ADD
  Code      string `gorm:"not null;uniqueIndex:idx_product_company_code"`
  // ... fields

  Tenant  Tenant  `gorm:"foreignKey:TenantID"`
  Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:RESTRICT"` // ← ADD
}
```

**SQL Migration Pattern**:
```sql
-- For EACH transactional table:

-- Step 1: Add column
ALTER TABLE products
  ADD COLUMN company_id VARCHAR(255) AFTER tenant_id;

-- Step 2: Populate from existing data (migration strategy)
UPDATE products p
  JOIN companies c ON c.tenant_id = p.tenant_id
  SET p.company_id = c.id;

-- Step 3: Make NOT NULL
ALTER TABLE products
  MODIFY COLUMN company_id VARCHAR(255) NOT NULL;

-- Step 4: Add index
ALTER TABLE products
  ADD INDEX idx_product_company (company_id);

-- Step 5: Add foreign key
ALTER TABLE products
  ADD CONSTRAINT fk_product_company
  FOREIGN KEY (company_id) REFERENCES companies(id)
  ON DELETE RESTRICT;

-- Step 6: Update composite unique indexes
DROP INDEX idx_product_tenant_code ON products;
CREATE UNIQUE INDEX idx_product_company_code
  ON products(company_id, code);
```

**Apply to ALL 20+ tables listed above**

---

## 🏗️ ARCHITECTURE COMPARISON

### Current Architecture (SINGLE-COMPANY)

```
┌─────────┐
│  User   │
└────┬────┘
     │ M:N (via user_tenants dengan role di TENANT level)
     ↓
┌─────────┐ 1:1 ┌─────────┐
│ Tenant  │←───→│ Company │ ← uniqueIndex enforces 1:1
└────┬────┘     └─────────┘
     │ 1:N
     ↓
┌──────────────────┐
│ Transactional    │
│ Tables:          │
│ - Products       │ ← TenantID only
│ - Warehouses     │ ← TenantID only
│ - SalesOrders    │ ← TenantID only
│ - Customers      │ ← TenantID only
│ (20+ tables)     │ ← TenantID only
└──────────────────┘
```

**Limitations**:
- ❌ Satu tenant = satu company saja (enforced by uniqueIndex)
- ❌ User role di tenant level, tidak bisa beda per company
- ❌ Data tidak bisa di-isolate per company
- ❌ Tidak mendukung business scenario real (multiple PT/CV)

### Required Architecture (MULTI-COMPANY)

```
┌─────────┐
│  User   │
└────┬────┘
     │ M:N (via user_company_roles dengan role di COMPANY level)
     ↓
┌──────────────┐
│UserCompanyRole│ ← NEW TABLE (CRITICAL)
└───────┬───────┘
        │
        ↓
┌─────────┐ 1:N ┌─────────┐
│ Tenant  │────→│Companies│ ← TenantID di Companies (reversed)
└─────────┘     └────┬────┘
                     │ 1:N
                     ↓
             ┌──────────────────┐
             │ Transactional    │
             │ Tables:          │
             │ - Products       │ ← TenantID + CompanyID
             │ - Warehouses     │ ← TenantID + CompanyID
             │ - SalesOrders    │ ← TenantID + CompanyID
             │ - Customers      │ ← TenantID + CompanyID
             │ (20+ tables)     │ ← TenantID + CompanyID
             └──────────────────┘
```

**Benefits**:
- ✅ Satu tenant bisa punya banyak companies
- ✅ User bisa punya role berbeda di setiap company
- ✅ Data ter-isolate per company (data leakage prevention)
- ✅ Mendukung skenario bisnis real (PT + CV dalam 1 tenant)

---

## 📋 REQUIRED DATABASE CHANGES

### Phase 1: Schema Restructuring (BREAKING CHANGES)

#### 1.1. Reverse Tenant-Company Relationship

```sql
-- Step 1: Add TenantID to Companies
ALTER TABLE companies
  ADD COLUMN tenant_id VARCHAR(255) AFTER id;

-- Step 2: Populate TenantID from existing Tenant.CompanyID
UPDATE companies c
  JOIN tenants t ON t.company_id = c.id
  SET c.tenant_id = t.id;

-- Step 3: Verify population
SELECT COUNT(*) as missing_tenant
FROM companies
WHERE tenant_id IS NULL;
-- Should return 0

-- Step 4: Make tenant_id NOT NULL
ALTER TABLE companies
  MODIFY COLUMN tenant_id VARCHAR(255) NOT NULL;

-- Step 5: Add indexes and foreign key
ALTER TABLE companies
  ADD INDEX idx_company_tenant (tenant_id);

ALTER TABLE companies
  ADD CONSTRAINT fk_company_tenant
  FOREIGN KEY (tenant_id) REFERENCES tenants(id)
  ON DELETE CASCADE;

-- Step 6: Remove CompanyID from Tenants
ALTER TABLE tenants DROP FOREIGN KEY IF EXISTS fk_tenant_company;
ALTER TABLE tenants DROP INDEX IF EXISTS company_id;
ALTER TABLE tenants DROP COLUMN company_id;
ALTER TABLE tenants DROP COLUMN subscription_id; -- Move to Company level
```

#### 1.2. Create user_company_roles Table

```sql
CREATE TABLE user_company_roles (
  id VARCHAR(255) PRIMARY KEY,
  user_id VARCHAR(255) NOT NULL,
  company_id VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL
    CHECK (role IN ('OWNER', 'ADMIN', 'FINANCE', 'SALES', 'WAREHOUSE', 'STAFF')),
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY idx_user_company (user_id, company_id),
  INDEX idx_user (user_id),
  INDEX idx_company (company_id),
  INDEX idx_role (role),
  INDEX idx_active (is_active),

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

-- Migrate existing user_tenants to user_company_roles
INSERT INTO user_company_roles (id, user_id, company_id, role, is_active, created_at, updated_at)
SELECT
  CONCAT('ucr_', ut.id) as id,
  ut.user_id,
  c.id as company_id,
  ut.role,
  ut.is_active,
  ut.created_at,
  ut.updated_at
FROM user_tenants ut
JOIN companies c ON c.tenant_id = ut.tenant_id;

-- Verify migration
SELECT
  COUNT(DISTINCT ut.id) as user_tenant_records,
  COUNT(DISTINCT ucr.id) as user_company_records
FROM user_tenants ut
LEFT JOIN user_company_roles ucr ON ucr.user_id = ut.user_id;
-- Should be equal
```

#### 1.3. Add CompanyID to ALL Transactional Tables

**List of Tables to Update**:
1. products
2. product_units
3. product_batches
4. product_suppliers
5. price_list
6. warehouses
7. warehouse_stocks
8. customers
9. suppliers
10. sales_orders
11. sales_order_items
12. purchase_orders
13. purchase_order_items
14. invoices
15. invoice_items
16. cash_transactions
17. supplier_payments
18. inventory_movements
19. deliveries
20. delivery_items
21. goods_receipts
22. goods_receipt_items
23. stock_opnames
24. stock_transfers

**SQL Pattern for Each Table**:
```sql
-- Example: products table

-- Step 1: Add column
ALTER TABLE products
  ADD COLUMN company_id VARCHAR(255) AFTER tenant_id;

-- Step 2: Populate from existing data
-- Assumes each tenant currently has exactly 1 company
UPDATE products p
  JOIN companies c ON c.tenant_id = p.tenant_id
  SET p.company_id = c.id;

-- Step 3: Verify no NULL values
SELECT COUNT(*) as missing_company
FROM products
WHERE company_id IS NULL;
-- Should return 0

-- Step 4: Make NOT NULL
ALTER TABLE products
  MODIFY COLUMN company_id VARCHAR(255) NOT NULL;

-- Step 5: Add index
ALTER TABLE products
  ADD INDEX idx_product_company (company_id);

-- Step 6: Add foreign key
ALTER TABLE products
  ADD CONSTRAINT fk_product_company
  FOREIGN KEY (company_id) REFERENCES companies(id)
  ON DELETE RESTRICT;

-- Step 7: Update composite unique indexes
DROP INDEX idx_product_tenant_code ON products;
CREATE UNIQUE INDEX idx_product_company_code
  ON products(company_id, code);

-- Repeat for ALL 24 tables
```

---

## 🔧 REQUIRED BACKEND CODE CHANGES

### 1. Model Changes

#### 1.1. Update tenant.go

```go
// BEFORE
type Tenant struct {
  ID             string `gorm:"primaryKey"`
  CompanyID      string `gorm:"uniqueIndex;not null"` // ← REMOVE
  SubscriptionID *string
  // ...
  Company      Company `gorm:"foreignKey:CompanyID"` // ← REMOVE
}

// AFTER
type Tenant struct {
  ID   string `gorm:"primaryKey"`
  Name string `gorm:"not null"` // Tenant business/group name
  // ... other fields

  // Relations
  Companies []Company     `gorm:"foreignKey:TenantID"` // ← One-to-Many
  Users     []UserTenant  `gorm:"foreignKey:TenantID"`
}
```

#### 1.2. Update company.go

```go
// BEFORE
type Company struct {
  ID string `gorm:"primaryKey"`
  // ... fields
  Tenant *Tenant `gorm:"foreignKey:CompanyID"` // ← Wrong direction
}

// AFTER
type Company struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index"` // ← ADD

  // ... other fields

  // Relations
  Tenant            Tenant            `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
  Banks             []CompanyBank     `gorm:"foreignKey:CompanyID"`
  UserCompanyRoles  []UserCompanyRole `gorm:"foreignKey:CompanyID"` // ← NEW
}
```

#### 1.3. Create user_company_role.go (NEW FILE)

```go
package models

import (
  "time"
  "github.com/lucsky/cuid"
  "gorm.io/gorm"
)

type UserCompanyRole struct {
  ID        string   `gorm:"primaryKey"`
  UserID    string   `gorm:"not null;index;uniqueIndex:idx_user_company"`
  CompanyID string   `gorm:"not null;index;uniqueIndex:idx_user_company"`
  Role      UserRole `gorm:"type:varchar(20);default:'STAFF';index"`
  IsActive  bool     `gorm:"default:true"`
  CreatedAt time.Time `gorm:"autoCreateTime"`
  UpdatedAt time.Time `gorm:"autoUpdateTime"`

  // Relations
  User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
  Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
}

func (UserCompanyRole) TableName() string {
  return "user_company_roles"
}

func (ucr *UserCompanyRole) BeforeCreate(tx *gorm.DB) error {
  if ucr.ID == "" {
    ucr.ID = cuid.New()
  }
  return nil
}
```

#### 1.4. Update ALL Transactional Models

**Pattern for EVERY model** (apply to 20+ models):

```go
// Example: product.go

// BEFORE
type Product struct {
  ID       string `gorm:"primaryKey"`
  TenantID string `gorm:"not null;index;uniqueIndex:idx_product_tenant_code"`
  Code     string `gorm:"not null;uniqueIndex:idx_product_tenant_code"`
  // ... fields

  Tenant Tenant `gorm:"foreignKey:TenantID"`
}

// AFTER
type Product struct {
  ID        string `gorm:"primaryKey"`
  TenantID  string `gorm:"not null;index"` // ← Keep for tenant isolation
  CompanyID string `gorm:"not null;index;uniqueIndex:idx_product_company_code"` // ← ADD
  Code      string `gorm:"not null;uniqueIndex:idx_product_company_code"`
  // ... fields

  Tenant  Tenant  `gorm:"foreignKey:TenantID"`
  Company Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:RESTRICT"` // ← ADD
}
```

**Apply to**: Product, ProductUnit, ProductBatch, ProductSupplier, PriceList, Warehouse, WarehouseStock, Customer, Supplier, SalesOrder, SalesOrderItem, PurchaseOrder, PurchaseOrderItem, Invoice, InvoiceItem, Delivery, DeliveryItem, GoodsReceipt, GoodsReceiptItem, InventoryMovement, StockOpname, StockTransfer, CashTransaction, SupplierPayment, etc.

### 2. API Changes

#### 2.1. New Endpoints Required

```go
// GET /api/v1/tenant/companies
// List all companies user can access within their tenant
func GetUserCompanies(c *gin.Context)

// POST /api/v1/tenant/companies
// Create new company (OWNER only)
func CreateCompany(c *gin.Context)

// POST /api/v1/auth/switch-company
// Switch active company context
func SwitchCompany(c *gin.Context)
```

#### 2.2. Update JWT Structure

```go
// BEFORE
type JWTClaims struct {
  UserID   string `json:"user_id"`
  TenantID string `json:"tenant_id"`
  Role     string `json:"role"` // ← Single role
  // ...
}

// AFTER
type CompanyAccess struct {
  CompanyID string `json:"company_id"`
  Role      string `json:"role"`
}

type JWTClaims struct {
  UserID         string          `json:"user_id"`
  TenantID       string          `json:"tenant_id"`
  ActiveCompany  string          `json:"active_company"`  // ← Currently selected
  CompanyAccess  []CompanyAccess `json:"company_access"` // ← All accessible companies
  // ...
}
```

#### 2.3. Company Context Middleware

```go
func CompanyContextMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    companyID := c.GetHeader("X-Company-ID")

    // Validate company access
    // Set company context
    // ...
  }
}
```

#### 2.4. Update ALL Query Scopes

**Pattern for EVERY handler** (30+ handlers):

```go
// BEFORE
func GetProducts(c *gin.Context) {
  tenantID := c.GetString("tenant_id")

  var products []models.Product
  db.Where("tenant_id = ?", tenantID).Find(&products)
}

// AFTER
func GetProducts(c *gin.Context) {
  tenantID := c.GetString("tenant_id")
  companyID := c.GetString("company_id") // ← From middleware

  var products []models.Product
  db.Where("tenant_id = ? AND company_id = ?",
           tenantID, companyID). // ← Company-scoped
     Find(&products)
}
```

---

## ⚠️ MIGRATION RISKS & MITIGATION

### High Risks

#### 1. Data Loss Risk
**Risk**: Migration could corrupt existing data
**Mitigation**:
- Full database backup before migration
- Test on staging environment first
- Dry-run migration scripts
- Verify data integrity after each step

#### 2. Downtime Required
**Risk**: Schema changes require production downtime
**Mitigation**:
- Schedule during low-traffic period
- Use blue-green deployment if possible
- Prepare rollback scripts
- Communication plan to users

#### 3. Breaking Changes
**Risk**: ALL API consumers must update
**Mitigation**:
- Version API (v1 vs v2)
- Provide migration guide
- Backward compatibility period
- Coordinate with frontend team

### Medium Risks

#### 4. Performance Degradation
**Risk**: Additional JOINs and WHERE clauses
**Mitigation**:
- Add proper indexes
- Query optimization
- Load testing
- Caching strategy

#### 5. Code Complexity
**Risk**: More complex query logic
**Mitigation**:
- Comprehensive testing
- Code documentation
- Helper functions
- Code review process

---

## 📅 ESTIMATED TIMELINE

### Total: 6-7 Weeks

**Week 1: Database Design & Migration Scripts**
- Design new schema structure
- Write migration SQL scripts
- Test on local development
- Dry-run on staging clone

**Week 2-3: Backend Model & API Changes**
- Update 20+ model files
- Create new endpoints
- Update JWT structure
- Create middleware
- Update 30+ handlers

**Week 4: Backend Testing**
- Unit tests
- Integration tests
- Security tests (data isolation)
- Performance tests

**Week 5-6: Frontend Integration**
- Update types and Redux
- Implement TeamSwitcher
- Permission system
- Dynamic navigation

**Week 7: Deployment & Monitoring**
- Staging deployment
- Production migration
- Monitoring & hotfixes
- User feedback

---

## ✅ CONCLUSION

### Jawaban Lengkap untuk Pertanyaan User

> **Pertanyaan**: "untuk ini apakah sudah melakukan pengecekan pada tabel terkait pada backend?"

### **Jawaban: TIDAK, backend database BELUM MENDUKUNG multi-company architecture**

**Bukti dari Code Analysis**:

1. **✗ Tenant-Company Relationship SALAH**
   - File: `backend/models/tenant.go:15`
   - `CompanyID` dengan `uniqueIndex` → enforces 1:1
   - Harusnya: `TenantID` di Company → allows 1:N

2. **✗ User Permission Table TIDAK ADA**
   - File: `backend/models/user.go` hanya punya `UserTenant`
   - Missing: `UserCompanyRole` table
   - Role di tenant level, bukan company level

3. **✗ Data Scoping COMPLETELY WRONG**
   - 20+ tables hanya punya `TenantID`
   - NO `CompanyID` di ANY transactional table
   - Files checked:
     - `backend/models/product.go:15` - NO CompanyID
     - `backend/models/warehouse.go:15` - NO CompanyID
     - `backend/models/sales.go:15` - NO CompanyID
     - `backend/models/master.go` (Customer & Supplier) - NO CompanyID

**Kesimpulan**:
- Semua analisis multi-company sebelumnya adalah **THEORETICAL DESIGN**
- Backend masih **single-company-per-tenant architecture**
- Untuk implement multi-company, perlu **COMPLETE DATABASE RESTRUCTURING**
- Estimated effort: **6-7 weeks** (3 weeks backend, 2 weeks frontend, 2 weeks testing/deployment)

**Next Steps**:
1. Diskusi dengan Product team: Apakah multi-company mandatory?
2. Jika ya, backend team harus mulai database redesign
3. Frontend harus menunggu backend selesai
4. Koordinasi deployment strategy untuk breaking changes

---

**Document Generated**: 2025-12-26
**Analysis Tool**: Sequential Thinking (Ultrathink mode)
**Files Analyzed**: 10+ backend model files + database schema
