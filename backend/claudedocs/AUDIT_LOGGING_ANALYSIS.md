# 📊 ANALISIS KOMPREHENSIF: Audit Logging untuk Operasi Product

**Tanggal Analisis:** 2026-01-05
**Scope:** Product INSERT, UPDATE, DELETE Operations
**Status:** 🔄 IN PROGRESS (Model Updated, Implementation Pending)

---

## 📑 TABLE OF CONTENTS

1. [Executive Summary](#executive-summary)
2. [MVP Implementation Strategy](#mvp-implementation-strategy) ⭐ NEW
3. [Temuan Kritis](#temuan-kritis)
4. [Struktur Tabel Saat Ini](#struktur-tabel-audit_logs-saat-ini)
5. [Field yang Perlu Ditambahkan](#field-yang-perlu-ditambahkan)
6. [Rekomendasi Implementasi](#rekomendasi-implementasi)
7. [Implementation Roadmap - MVP Phased Approach](#implementation-roadmap) ⭐ UPDATED
8. [Next Steps](#next-steps)

---

## 🎯 EXECUTIVE SUMMARY

### Pertanyaan Awal
> "Sebagai programmer profesional lakukan analisa, pada operasi product saat INSERT dan UPDATE belum ada pembuatan record pada tabel audit_logs, apakah ada dari field di tabel audit_logs yang perlu ditambah?"

### Jawaban
✅ **KONFIRMASI:** Operasi product INSERT dan UPDATE **TIDAK ADA** audit logging
✅ **KONFIRMASI:** Ada **9 field yang perlu ditambahkan** ke tabel audit_logs

### Critical Findings
- 🔴 **ZERO audit logging** untuk product operations (CREATE/UPDATE/DELETE)
- 🔴 **CompanyID field MISSING** (critical for multi-company scenarios)
- 🔴 **No transaction grouping** (RequestID/TransactionID needed)
- 🔴 **No operation status tracking** (SUCCESS/FAILED)
- 🟡 **No changed fields tracking** (efficiency issue for updates)

### 🚀 MVP Approach (RECOMMENDED)

**Current Progress:** ✅ Day 1 Complete (2026-01-05)

**MVP Phase 1 - CRITICAL Fields Only (3 fields):**
- ✅ **CompanyID** - Multi-company filtering & compliance *(Model Updated)*
- ✅ **RequestID** - Transaction grouping & debugging *(Model Updated)*
- ✅ **Status** - Success/failure tracking *(Model Updated)*

**Next Steps:**
1. Run migration: `go run cmd/migrate/main.go`
2. Add Status constants & update AuditContext
3. Implement audit methods (Day 3-4)
4. Integrate with ProductService (Day 5)

**Timeline:** Week 1-2 remaining → Production ready!

**Phase 2 & 3:** Add remaining fields based on actual needs (see detailed roadmap below)

---

## 🚀 MVP IMPLEMENTATION STRATEGY

### Why MVP Approach?

**Benefits:**
- ✅ **Faster Time to Production** - 1-2 minggu vs 4 minggu
- ✅ **Lower Risk** - Incremental changes, easier rollback
- ✅ **Validate Early** - Get feedback before building everything
- ✅ **Flexibility** - Pivot based on actual usage patterns
- ✅ **Cost Effective** - Don't build what you don't need

### 3-Phase Approach

```
MVP Phase 1 (1-2 weeks)     Phase 2 (2-3 weeks)         Phase 3 (Future)
━━━━━━━━━━━━━━━━━━━━━━    ━━━━━━━━━━━━━━━━━━━━       ━━━━━━━━━━━━━━━━
🔴 CRITICAL (3 fields)     🟡 HIGH (4 fields)          🟢 NICE-TO-HAVE (3 fields)

✅ CompanyID                ✅ ErrorMessage             ✅ Duration
✅ RequestID                ✅ ChangedFields            ✅ Severity
✅ Status                   ✅ Module                   ✅ Metadata
                            ✅ SubModule

→ Production Ready!         → Enhanced Capabilities     → Advanced Features
```

### MVP Phase 1: CRITICAL Fields (WAJIB)

**Status:** ✅ Model Updated (2026-01-05) | ⏳ Implementation Pending

**3 Fields yang HARUS ada untuk production:**

| Field | Why CRITICAL? | Without This... |
|-------|---------------|-----------------|
| **CompanyID** | Multi-company filtering | ❌ Audit logs tercampur antar companies<br>❌ Compliance audit GAGAL<br>❌ Tidak bisa filter per company |
| **RequestID** | Transaction grouping | ❌ Tidak bisa group related operations<br>❌ Debugging nightmare<br>❌ Tidak tahu product + units + stocks = 1 transaction |
| **Status** | Success/failure tracking | ❌ Tidak bisa detect failed operations<br>❌ Missing error patterns<br>❌ Assume everything always succeeds |

**✅ Model Update (COMPLETED - models/system.go):**
```go
type AuditLog struct {
    ID            string    `gorm:"type:varchar(255);primaryKey"`
    TenantID      *string   `gorm:"type:varchar(255);index"`
    CompanyID     *string   `gorm:"type:varchar(255);index"` // ✅ ADDED
    UserID        *string   `gorm:"type:varchar(255);index"`
    RequestID     *string   `gorm:"type:varchar(100);index"` // ✅ ADDED
    Action        string    `gorm:"type:varchar(100);not null;index"`
    EntityType    *string   `gorm:"type:varchar(100);index"`
    EntityID      *string   `gorm:"type:varchar(255);index"`
    OldValues     *string   `gorm:"type:text"`
    NewValues     *string   `gorm:"type:text"`
    IPAddress     *string   `gorm:"type:varchar(45)"`
    UserAgent     *string   `gorm:"type:varchar(500)"`
    Status        string    `gorm:"type:varchar(20);default:'SUCCESS';index"` // ✅ ADDED
    Notes         *string   `gorm:"type:text"`
    CreatedAt     time.Time `gorm:"autoCreateTime;index"`

    // Relations
    Tenant  *Tenant  `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
    Company *Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"` // ✅ ADDED
    User    *User    `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
```

**Database Migration:**
> ℹ️ **Note:** Sistem belum berjalan - tidak perlu ALTER statement.
> GORM akan otomatis create/update tabel saat migration.

```bash
# Apply migration
go run cmd/migrate/main.go
```

GORM akan generate schema dengan 3 field baru + indexes + foreign key constraint.

**Implementation (MVP Phase 1):**
```go
// Minimal audit method
func (s *AuditService) LogProductCreated(
    ctx context.Context,
    auditCtx *AuditContext,
    productID string,
    productData map[string]interface{},
) error {
    newValuesJSON, _ := json.Marshal(productData)
    newValuesStr := string(newValuesJSON)
    entityType := "PRODUCT"

    auditLog := &models.AuditLog{
        TenantID:   auditCtx.TenantID,
        CompanyID:  auditCtx.CompanyID,  // MVP
        UserID:     auditCtx.UserID,
        RequestID:  auditCtx.RequestID,  // MVP
        Action:     "PRODUCT_CREATED",
        EntityType: &entityType,
        EntityID:   &productID,
        NewValues:  &newValuesStr,
        Status:     "SUCCESS",           // MVP
        IPAddress:  auditCtx.IPAddress,
        UserAgent:  auditCtx.UserAgent,
    }

    return s.db.WithContext(ctx).Create(auditLog).Error
}
```

### Phase 2: HIGH Priority (Optional, Based on Feedback)

**Add after MVP is stable:**
- ErrorMessage - Error details for failed operations
- ChangedFields - Efficiency for UPDATE operations
- Module/SubModule - Categorization for reporting

### Phase 3: NICE-TO-HAVE (Future Enhancement)

**Add when needed:**
- Duration - Performance monitoring
- Severity - Alerting system
- Metadata - Flexible extensibility

### Decision Point After MVP

**After MVP Phase 1 deployment, evaluate:**

✅ **Continue to Phase 2 if:**
- MVP working well
- Need better error tracking
- Want efficient UPDATE logging
- Need categorization for reporting

⏸️ **Stay at MVP if:**
- Basic audit trail sufficient
- No immediate need for advanced features
- Want to observe usage patterns first

---

## 🔴 TEMUAN KRITIS

### 1. Audit Logging TIDAK Diimplementasikan ❌

#### Lokasi yang Dianalisis:

**✅ Infrastructure Sudah Ada:**
- `models/system.go:39-70` - Model AuditLog sudah lengkap
- `internal/service/audit/audit_service.go` - Service sudah tersedia

**❌ Product Operations TIDAK ADA Audit Logging:**

| File | Method | Line | Status |
|------|--------|------|--------|
| `internal/handler/product_handler.go` | CreateProduct | 37-71 | ❌ NO AUDIT |
| `internal/handler/product_handler.go` | UpdateProduct | 178-218 | ❌ NO AUDIT |
| `internal/handler/product_handler.go` | DeleteProduct | 223-247 | ❌ NO AUDIT |
| `internal/service/product/product_service.go` | CreateProduct | 32-183 | ❌ NO AUDIT |
| `internal/service/product/product_service.go` | UpdateProduct | 281-367 | ❌ NO AUDIT |
| `internal/service/product/product_service.go` | DeleteProduct | 369-389 | ❌ NO AUDIT |

#### Audit Service Saat Ini Hanya Mencakup:
- ✅ User role changes (`LogUserRoleChange`)
- ✅ User additions/removals (`LogUserAdded`, `LogUserRemoved`, `LogUserReactivated`)
- ✅ Bank account operations (`LogBankAccountAdded`, `LogBankAccountUpdated`, `LogBankAccountDeleted`)
- ✅ Company profile updates (`LogCompanyUpdated`)
- ❌ **Product operations** (TIDAK ADA)

---

## 📋 STRUKTUR TABEL `audit_logs` SAAT INI

### Current Schema (models/system.go:39-70)

```go
type AuditLog struct {
    ID         string    `gorm:"type:varchar(255);primaryKey"`
    TenantID   *string   `gorm:"type:varchar(255);index"`
    UserID     *string   `gorm:"type:varchar(255);index"`
    Action     string    `gorm:"type:varchar(100);not null;index"`
    EntityType *string   `gorm:"type:varchar(100);index"`
    EntityID   *string   `gorm:"type:varchar(255);index"`
    OldValues  *string   `gorm:"type:text"` // JSON
    NewValues  *string   `gorm:"type:text"` // JSON
    IPAddress  *string   `gorm:"type:varchar(45)"`
    UserAgent  *string   `gorm:"type:varchar(500)"`
    Notes      *string   `gorm:"type:text"`
    CreatedAt  time.Time `gorm:"autoCreateTime;index"`

    // Relations
    Tenant *Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
    User   *User   `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}
```

### Field Analysis

| Field | Status | Purpose | Assessment |
|-------|--------|---------|------------|
| ID | ✅ | Unique identifier | Good |
| TenantID | ✅ | Multi-tenant isolation | Good but INCOMPLETE (missing CompanyID) |
| UserID | ✅ | Who performed action | Good |
| Action | ✅ | Operation type | Good |
| EntityType | ✅ | What was affected | Good |
| EntityID | ✅ | ID of affected record | Good |
| OldValues | ✅ | Previous state (JSON) | Good but INEFFICIENT for updates |
| NewValues | ✅ | New state (JSON) | Good but INEFFICIENT for updates |
| IPAddress | ✅ | Request origin | Good |
| UserAgent | ✅ | Client information | Good |
| Notes | ✅ | Human-readable description | Good |
| CreatedAt | ✅ | Timestamp | Good |
| **CompanyID** | ❌ | **MISSING** | **CRITICAL** |
| **RequestID** | ❌ | **MISSING** | **CRITICAL** |
| **Status** | ❌ | **MISSING** | **CRITICAL** |
| **ChangedFields** | ❌ | **MISSING** | **HIGH** |
| **Module** | ❌ | **MISSING** | **HIGH** |
| **ErrorMessage** | ❌ | **MISSING** | **HIGH** |

---

## ⚠️ FIELD YANG PERLU DITAMBAHKAN

### 🔴 PRIORITY: CRITICAL (Harus Segera)

---

#### 1. CompanyID ⭐⭐⭐⭐⭐ SANGAT PENTING

```go
CompanyID *string `gorm:"type:varchar(255);index"`
```

**Alasan:**
- Product model memiliki **BOTH** `TenantID` dan `CompanyID`
  ```go
  // models/product.go:15-16
  TenantID  string `gorm:"type:varchar(255);not null;index"`
  CompanyID string `gorm:"type:varchar(255);not null;index"`
  ```
- **Multi-company scenario:** Satu tenant bisa memiliki multiple companies
- Audit logs saat ini **HANYA** punya `TenantID`, tidak bisa filter per company
- **Compliance requirement:** Harus tahu operasi dilakukan di company mana

**Use Case Real-World:**
```
Scenario: PT ABC Group (1 Tenant) memiliki 3 companies:
  - PT ABC Sembako (company_id: comp-001)
  - PT XYZ Distribusi (company_id: comp-002)
  - PT 123 Foods (company_id: comp-003)

User membuat product "Beras Premium" di PT ABC Sembako
→ Audit HARUS mencatat: tenant_id = "tenant-abc" AND company_id = "comp-001"

Query audit per company:
SELECT * FROM audit_logs
WHERE tenant_id = 'tenant-abc'
  AND company_id = 'comp-001'
  AND entity_type = 'PRODUCT';
```

**Impact jika TIDAK ada:**
- ❌ Tidak bisa filter audit logs per company
- ❌ Compliance audit gagal (tidak bisa prove per-company compliance)
- ❌ Error investigation sulit (data tercampur antar companies)
- ❌ Reporting tidak akurat (gabungan semua companies dalam 1 tenant)

**Priority:** 🔴 **CRITICAL** - MUST HAVE for Phase 1

---

#### 2. RequestID / TransactionID

```go
RequestID *string `gorm:"type:varchar(100);index"`
```

**Alasan:**
Product CREATE operation adalah **COMPLEX TRANSACTION** dengan multiple insertions:

```go
// internal/service/product/product_service.go:60-175
err = s.db.Transaction(func(tx *gorm.DB) error {
    // 1. Create Product
    tx.Create(product)

    // 2. Create Base Unit
    tx.Create(baseUnit)

    // 3. Create Additional Units (loop)
    for _, unitReq := range req.Units {
        tx.Create(unit)
    }

    // 4. Initialize Warehouse Stocks (loop)
    for _, wh := range warehouses {
        tx.Create(whStock)
    }
})
```

**Problem tanpa RequestID:**
Satu operasi "Create Product" menghasilkan **6+ audit entries terpisah**:
- 1x PRODUCT_CREATED
- 1x PRODUCT_UNIT_CREATED (base unit)
- 3x PRODUCT_UNIT_CREATED (additional units)
- 2x WAREHOUSE_STOCK_INITIALIZED (2 warehouses)

**Tidak ada cara untuk group entries ini!**

**Use Case dengan RequestID:**
```sql
-- Semua operasi dalam satu request
SELECT * FROM audit_logs
WHERE request_id = 'req-12345-67890'
ORDER BY created_at;

Result:
┌────────────────────────┬─────────────────┬──────────────────┐
│ Action                 │ EntityType      │ EntityID         │
├────────────────────────┼─────────────────┼──────────────────┤
│ PRODUCT_CREATED        │ PRODUCT         │ prod-001         │
│ PRODUCT_UNIT_CREATED   │ PRODUCT_UNIT    │ unit-001 (PCS)   │
│ PRODUCT_UNIT_CREATED   │ PRODUCT_UNIT    │ unit-002 (KARTON)│
│ PRODUCT_UNIT_CREATED   │ PRODUCT_UNIT    │ unit-003 (LUSIN) │
│ WAREHOUSE_STOCK_INIT   │ WAREHOUSE_STOCK │ ws-001 (MAIN)    │
│ WAREHOUSE_STOCK_INIT   │ WAREHOUSE_STOCK │ ws-002 (BRANCH)  │
└────────────────────────┴─────────────────┴──────────────────┘

-- Tracking flow lengkap untuk debugging
-- Error investigation: "Kenapa warehouse stock tidak ke-initialize?"
```

**Implementation:**
```go
// Generate unique RequestID per HTTP request
requestID := uuid.New().String()

// Pass ke semua audit log calls dalam transaction
auditCtx := &AuditContext{
    RequestID: &requestID,
    // ... other fields
}
```

**Priority:** 🔴 **CRITICAL** - Essential for complex operations

---

#### 3. Status

```go
Status string `gorm:"type:varchar(20);default:'SUCCESS';index"`
// Values: SUCCESS, FAILED, PARTIAL
```

**Alasan:**
Saat ini audit_logs **ASSUME SEMUA OPERASI SUKSES**. Tidak ada cara untuk track:
- Validation failures
- Database constraint errors
- Business logic rejections
- Partial successes

**Problem Real-World:**

```go
// internal/service/product/product_service.go:281-367
func (s *ProductService) UpdateProduct(...) (*models.Product, error) {
    // Validation
    if err := s.validateUpdateProduct(...); err != nil {
        return nil, err // ❌ Error tidak tercatat di audit!
    }

    // Update
    if err := s.db.Updates(updates).Error; err != nil {
        return nil, err // ❌ Error tidak tercatat di audit!
    }

    // ✅ Success path only gets audited (jika diimplementasikan)
}
```

**Use Case dengan Status:**
```sql
-- Failed operations analysis
SELECT action, entity_type, error_message, COUNT(*) as failures
FROM audit_logs
WHERE status = 'FAILED'
  AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY action, entity_type, error_message
ORDER BY failures DESC;

Result:
┌────────────────┬─────────────┬──────────────────────────┬──────────┐
│ Action         │ EntityType  │ ErrorMessage             │ Failures │
├────────────────┼─────────────┼──────────────────────────┼──────────┤
│ PRODUCT_UPDATED│ PRODUCT     │ Code already exists      │ 45       │
│ PRODUCT_CREATED│ PRODUCT     │ Duplicate barcode        │ 23       │
│ PRODUCT_UPDATED│ PRODUCT     │ Invalid decimal format   │ 12       │
└────────────────┴─────────────┴──────────────────────────┴──────────┘

-- User error patterns (siapa yang sering error?)
SELECT user_id, COUNT(*) as error_count
FROM audit_logs
WHERE status = 'FAILED'
GROUP BY user_id
ORDER BY error_count DESC;
```

**Status Values:**
- `SUCCESS` - Operation completed successfully
- `FAILED` - Operation completely failed (rolled back)
- `PARTIAL` - Some parts succeeded, some failed (complex transactions)

**Priority:** 🔴 **CRITICAL** - Essential for error tracking

---

### 🟡 PRIORITY: HIGH (Sangat Disarankan)

---

#### 4. ErrorMessage

```go
ErrorMessage *string `gorm:"type:text"`
```

**Alasan:**
Jika `Status = FAILED`, **WAJIB** store error detail untuk debugging.

**Integration dengan Status:**
```go
// When operation fails
auditLog := &models.AuditLog{
    Status:       "FAILED",
    ErrorMessage: &errorMsg, // "invalid baseCost format: expected decimal"
    // ... other fields
}
```

**Use Case:**
```sql
-- Debugging specific error
SELECT created_at, user_id, entity_id, error_message
FROM audit_logs
WHERE status = 'FAILED'
  AND entity_type = 'PRODUCT'
  AND error_message ILIKE '%duplicate%'
ORDER BY created_at DESC
LIMIT 20;

-- Error pattern analysis
SELECT
    SUBSTRING(error_message FROM 1 FOR 50) as error_pattern,
    COUNT(*) as occurrences,
    MIN(created_at) as first_seen,
    MAX(created_at) as last_seen
FROM audit_logs
WHERE status = 'FAILED'
GROUP BY error_pattern
HAVING COUNT(*) > 5
ORDER BY occurrences DESC;
```

**Priority:** 🟡 **HIGH** - Critical for debugging

---

#### 5. ChangedFields

```go
ChangedFields *string `gorm:"type:text"`
// JSON array: ["basePrice", "baseCost", "name"]
```

**Alasan:**
Update operation menggunakan **selective field updates**:

```go
// internal/service/product/product_service.go:294-354
updates := make(map[string]interface{})

if req.Name != nil {
    updates["name"] = *req.Name
}
if req.BasePrice != nil {
    updates["base_price"] = basePrice
}
// ... only changed fields
```

**Current Inefficiency:**
Storing **entire object** even if only 1 field changed!

```json
// User hanya mengubah basePrice dari 10000 → 12000
{
  "oldValues": "{\"id\":\"prod-001\",\"code\":\"BRS-001\",\"name\":\"Beras Premium\",\"category\":\"SEMBAKO\",\"baseUnit\":\"KG\",\"baseCost\":\"8000\",\"basePrice\":\"10000\",\"minimumStock\":\"10\",\"isBatchTracked\":true}",  // 200+ chars
  "newValues": "{\"id\":\"prod-001\",\"code\":\"BRS-001\",\"name\":\"Beras Premium\",\"category\":\"SEMBAKO\",\"baseUnit\":\"KG\",\"baseCost\":\"8000\",\"basePrice\":\"12000\",\"minimumStock\":\"10\",\"isBatchTracked\":true}"   // 200+ chars
}

// Total: 400+ chars untuk perubahan 1 field!
```

**With ChangedFields:**
```json
{
  "changedFields": ["basePrice"],
  "oldValues": "{\"basePrice\":\"10000\"}",  // Only changed field
  "newValues": "{\"basePrice\":\"12000\"}"   // Only changed field
}

// Total: ~80 chars - 80% reduction!
```

**Query Benefits:**
```sql
-- Siapa yang mengubah harga minggu ini? (PostgreSQL)
SELECT user_id, entity_id, old_values, new_values, created_at
FROM audit_logs
WHERE changed_fields::jsonb ? 'basePrice'
  AND created_at >= NOW() - INTERVAL '7 days'
  AND entity_type = 'PRODUCT'
ORDER BY created_at DESC;

-- Field change frequency (most changed fields)
SELECT
    jsonb_array_elements_text(changed_fields::jsonb) as field_name,
    COUNT(*) as change_count
FROM audit_logs
WHERE entity_type = 'PRODUCT'
  AND action = 'PRODUCT_UPDATED'
GROUP BY field_name
ORDER BY change_count DESC;

Result:
┌─────────────┬──────────────┐
│ field_name  │ change_count │
├─────────────┼──────────────┤
│ basePrice   │ 1,245        │
│ baseCost    │ 892          │
│ name        │ 156          │
│ description │ 89           │
└─────────────┴──────────────┘
```

**Priority:** 🟡 **HIGH** - Efficiency & query performance

---

#### 6. Module / SubModule

```go
Module    *string `gorm:"type:varchar(50);index"`
SubModule *string `gorm:"type:varchar(50);index"`
```

**Alasan:**
Kategorisasi audit logs berdasarkan **business domain** untuk:
- Filtering by module
- Reporting by business area
- Access control (user hanya bisa lihat audit logs untuk module tertentu)

**Module Hierarchy:**
```
MASTER_DATA
  ├─ PRODUCT
  ├─ CUSTOMER
  ├─ SUPPLIER
  └─ WAREHOUSE

INVENTORY
  ├─ STOCK_MOVEMENT
  ├─ STOCK_OPNAME
  └─ BATCH_TRACKING

SALES
  ├─ SALES_ORDER
  ├─ DELIVERY
  └─ INVOICE

PURCHASE
  ├─ PURCHASE_ORDER
  ├─ GOODS_RECEIPT
  └─ SUPPLIER_PAYMENT

FINANCE
  ├─ CASH_TRANSACTION
  ├─ PAYMENT
  └─ RECEIVABLE
```

**Use Case:**
```sql
-- Audit activity by module (monthly report)
SELECT
    module,
    sub_module,
    COUNT(*) as total_operations,
    COUNT(DISTINCT user_id) as unique_users
FROM audit_logs
WHERE created_at >= DATE_TRUNC('month', NOW())
GROUP BY module, sub_module
ORDER BY total_operations DESC;

Result:
┌─────────────┬──────────────────┬──────────────────┬──────────────┐
│ Module      │ SubModule        │ Total Operations │ Unique Users │
├─────────────┼──────────────────┼──────────────────┼──────────────┤
│ MASTER_DATA │ PRODUCT          │ 1,543            │ 12           │
│ SALES       │ SALES_ORDER      │ 892              │ 8            │
│ INVENTORY   │ STOCK_MOVEMENT   │ 756              │ 5            │
│ MASTER_DATA │ CUSTOMER         │ 432              │ 7            │
└─────────────┴──────────────────┴──────────────────┴──────────────┘

-- User activity by module (access patterns)
SELECT user_id, module, COUNT(*) as operations
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY user_id, module
ORDER BY user_id, operations DESC;
```

**Priority:** 🟡 **HIGH** - Essential for reporting & filtering

---

### 🟢 PRIORITY: NICE-TO-HAVE (Opsional, Future Enhancement)

---

#### 7. Duration

```go
Duration *int64 `gorm:"type:bigint"` // milliseconds
```

**Alasan:**
- Performance monitoring: Operasi mana yang lambat?
- Anomaly detection: Update product tiba-tiba 10x lebih lambat
- Optimization: Identify bottlenecks

**Implementation:**
```go
startTime := time.Now()

// ... operation ...

duration := time.Since(startTime).Milliseconds()

auditLog.Duration = &duration
```

**Use Case:**
```sql
-- Slow operations (> 5 seconds)
SELECT action, entity_type, entity_id, duration, created_at
FROM audit_logs
WHERE duration > 5000
  AND created_at >= NOW() - INTERVAL '24 hours'
ORDER BY duration DESC;

-- Average duration by action type
SELECT
    action,
    AVG(duration) as avg_ms,
    MAX(duration) as max_ms,
    COUNT(*) as operations
FROM audit_logs
WHERE duration IS NOT NULL
GROUP BY action
ORDER BY avg_ms DESC;
```

**Priority:** 🟢 **NICE-TO-HAVE** - Useful for optimization

---

#### 8. Severity

```go
Severity string `gorm:"type:varchar(20);default:'INFO'"`
// Values: INFO, WARNING, CRITICAL
```

**Alasan:**
- Prioritization: Mana yang perlu immediate attention?
- Alerting: Send notification untuk CRITICAL changes
- Filtering: Focus on important changes

**Severity Levels:**
```
INFO     - Normal operations (create, read, minor updates)
WARNING  - Important changes (price changes, deletions, role changes)
CRITICAL - High-impact changes (mass updates, security changes, system config)
```

**Auto-assignment Logic:**
```go
severity := "INFO" // default

// Price changes always WARNING
if contains(changedFields, "basePrice") || contains(changedFields, "baseCost") {
    severity = "WARNING"
}

// Deletions always WARNING
if action == "PRODUCT_DELETED" {
    severity = "WARNING"
}

// Mass operations CRITICAL
if entityCount > 100 {
    severity = "CRITICAL"
}
```

**Use Case:**
```sql
-- Critical changes today
SELECT action, entity_type, user_id, notes, created_at
FROM audit_logs
WHERE severity = 'CRITICAL'
  AND created_at >= CURRENT_DATE
ORDER BY created_at DESC;

-- Alert monitoring
SELECT severity, COUNT(*) as count
FROM audit_logs
WHERE created_at >= NOW() - INTERVAL '1 hour'
GROUP BY severity;
```

**Priority:** 🟢 **NICE-TO-HAVE** - Useful for monitoring

---

#### 9. Metadata

```go
Metadata *string `gorm:"type:jsonb"` // PostgreSQL JSONB
```

**Alasan:**
**Extensibility** - Store domain-specific data tanpa schema changes!

Different entities need different audit data:
- **Product:** Stock level at time of change, price change %, affected customers
- **Invoice:** Payment status, due date, outstanding amount
- **Customer:** Credit limit changes, overdue days
- **Warehouse:** Stock movement reason, batch expiry info

**Examples:**

**Product Price Change:**
```json
{
  "priceChangePercent": 20.5,
  "previousStockLevel": "100.000",
  "affectedPriceLists": 5,
  "affectedCustomers": 15,
  "competitorPriceComparison": {
    "competitorA": "11500",
    "competitorB": "12200",
    "ourNewPrice": "12000"
  }
}
```

**Product Deletion:**
```json
{
  "deletionReason": "Discontinued product",
  "remainingStock": "25.000",
  "replacementProductId": "prod-789",
  "affectedSalesOrders": 3,
  "lastSoldDate": "2025-12-15"
}
```

**Batch Tracking Change:**
```json
{
  "batchTrackingEnabled": true,
  "expiryDateRequired": true,
  "existingStockCount": "500.000",
  "migrationRequired": true,
  "estimatedMigrationTime": "2 hours"
}
```

**Query with Metadata (PostgreSQL):**
```sql
-- Products with price increases > 10%
SELECT entity_id,
       metadata->>'priceChangePercent' as increase,
       metadata->>'affectedCustomers' as customers
FROM audit_logs
WHERE entity_type = 'PRODUCT'
  AND action = 'PRODUCT_UPDATED'
  AND (metadata->>'priceChangePercent')::numeric > 10
ORDER BY created_at DESC;

-- Deletions with remaining stock
SELECT entity_id,
       metadata->>'deletionReason' as reason,
       metadata->>'remainingStock' as stock
FROM audit_logs
WHERE action = 'PRODUCT_DELETED'
  AND (metadata->>'remainingStock')::numeric > 0;
```

**Priority:** 🟢 **NICE-TO-HAVE** - Great for future extensibility

---

## 📝 REKOMENDASI IMPLEMENTASI

### Phase 1: Database Schema Enhancement

#### Migration SQL

```sql
-- Add new columns to audit_logs table
ALTER TABLE audit_logs
ADD COLUMN company_id VARCHAR(255),
ADD COLUMN request_id VARCHAR(100),
ADD COLUMN status VARCHAR(20) DEFAULT 'SUCCESS',
ADD COLUMN error_message TEXT,
ADD COLUMN changed_fields TEXT,
ADD COLUMN module VARCHAR(50),
ADD COLUMN sub_module VARCHAR(50),
ADD COLUMN duration BIGINT,
ADD COLUMN severity VARCHAR(20) DEFAULT 'INFO',
ADD COLUMN metadata JSONB;

-- Add indexes for new columns
CREATE INDEX idx_audit_logs_company_id ON audit_logs(company_id);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id);
CREATE INDEX idx_audit_logs_status ON audit_logs(status);
CREATE INDEX idx_audit_logs_module ON audit_logs(module);
CREATE INDEX idx_audit_logs_sub_module ON audit_logs(sub_module);
CREATE INDEX idx_audit_logs_severity ON audit_logs(severity);

-- Add foreign key constraint for company_id
ALTER TABLE audit_logs
ADD CONSTRAINT fk_audit_logs_company
FOREIGN KEY (company_id)
REFERENCES companies(id)
ON DELETE CASCADE;

-- Add check constraint for status
ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_status
CHECK (status IN ('SUCCESS', 'FAILED', 'PARTIAL'));

-- Add check constraint for severity
ALTER TABLE audit_logs
ADD CONSTRAINT chk_audit_logs_severity
CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL'));
```

---

### Phase 2: Update Go Models

#### Updated AuditLog Model

```go
// models/system.go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// AuditLog - Enhanced audit trail with comprehensive tracking
type AuditLog struct {
    ID            string    `gorm:"type:varchar(255);primaryKey"`
    TenantID      *string   `gorm:"type:varchar(255);index"`
    CompanyID     *string   `gorm:"type:varchar(255);index"` // 🆕 ADDED
    UserID        *string   `gorm:"type:varchar(255);index"`
    RequestID     *string   `gorm:"type:varchar(100);index"` // 🆕 ADDED
    Action        string    `gorm:"type:varchar(100);not null;index"`
    EntityType    *string   `gorm:"type:varchar(100);index"`
    EntityID      *string   `gorm:"type:varchar(255);index"`
    OldValues     *string   `gorm:"type:text"`
    NewValues     *string   `gorm:"type:text"`
    ChangedFields *string   `gorm:"type:text"` // 🆕 ADDED - JSON array
    Status        string    `gorm:"type:varchar(20);default:'SUCCESS';index"` // 🆕 ADDED
    ErrorMessage  *string   `gorm:"type:text"` // 🆕 ADDED
    Module        *string   `gorm:"type:varchar(50);index"` // 🆕 ADDED
    SubModule     *string   `gorm:"type:varchar(50);index"` // 🆕 ADDED
    Duration      *int64    `gorm:"type:bigint"` // 🆕 ADDED - milliseconds
    Severity      string    `gorm:"type:varchar(20);default:'INFO'"` // 🆕 ADDED
    Metadata      *string   `gorm:"type:jsonb"` // 🆕 ADDED
    IPAddress     *string   `gorm:"type:varchar(45)"`
    UserAgent     *string   `gorm:"type:varchar(500)"`
    Notes         *string   `gorm:"type:text"`
    CreatedAt     time.Time `gorm:"autoCreateTime;index"`

    // Relations
    Tenant  *Tenant  `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
    Company *Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"` // 🆕 ADDED
    User    *User    `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
    return "audit_logs"
}

// BeforeCreate hook to generate UUID for ID field
func (al *AuditLog) BeforeCreate(tx *gorm.DB) error {
    if al.ID == "" {
        al.ID = uuid.New().String()
    }
    return nil
}

// Constants for Status
const (
    AuditStatusSuccess = "SUCCESS"
    AuditStatusFailed  = "FAILED"
    AuditStatusPartial = "PARTIAL"
)

// Constants for Severity
const (
    AuditSeverityInfo     = "INFO"
    AuditSeverityWarning  = "WARNING"
    AuditSeverityCritical = "CRITICAL"
)
```

#### Updated AuditContext

```go
// internal/service/audit/audit_service.go

// AuditContext contains contextual information for audit logging
type AuditContext struct {
    TenantID  *string
    CompanyID *string // 🆕 ADDED
    UserID    *string
    RequestID *string // 🆕 ADDED
    IPAddress *string
    UserAgent *string
}

// AuditService handles audit logging for sensitive operations
type AuditService struct {
    db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
    return &AuditService{db: db}
}
```

---

### Phase 3: Add Product Audit Methods

```go
// internal/service/audit/audit_service.go

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "gorm.io/gorm"
    "backend/models"
)

// LogProductCreated logs when a product is created
func (s *AuditService) LogProductCreated(
    ctx context.Context,
    auditCtx *AuditContext,
    productID string,
    productData map[string]interface{},
) error {
    startTime := time.Now()

    newValuesJSON, _ := json.Marshal(productData)
    newValuesStr := string(newValuesJSON)
    entityType := "PRODUCT"
    module := "MASTER_DATA"
    subModule := "PRODUCT"
    notes := fmt.Sprintf("Product created: %v", productData["name"])

    auditLog := &models.AuditLog{
        TenantID:   auditCtx.TenantID,
        CompanyID:  auditCtx.CompanyID,
        UserID:     auditCtx.UserID,
        RequestID:  auditCtx.RequestID,
        Action:     "PRODUCT_CREATED",
        EntityType: &entityType,
        EntityID:   &productID,
        NewValues:  &newValuesStr,
        Status:     models.AuditStatusSuccess,
        Module:     &module,
        SubModule:  &subModule,
        Severity:   models.AuditSeverityInfo,
        IPAddress:  auditCtx.IPAddress,
        UserAgent:  auditCtx.UserAgent,
        Notes:      &notes,
    }

    err := s.db.WithContext(ctx).Create(auditLog).Error

    // Update duration
    duration := time.Since(startTime).Milliseconds()
    auditLog.Duration = &duration
    s.db.WithContext(ctx).Model(auditLog).Update("duration", duration)

    return err
}

// LogProductUpdated logs when a product is updated
func (s *AuditService) LogProductUpdated(
    ctx context.Context,
    auditCtx *AuditContext,
    productID string,
    oldValues, newValues map[string]interface{},
    changedFields []string,
) error {
    startTime := time.Now()

    // Only store changed fields in old/new values
    oldValuesFiltered := make(map[string]interface{})
    newValuesFiltered := make(map[string]interface{})
    for _, field := range changedFields {
        if val, ok := oldValues[field]; ok {
            oldValuesFiltered[field] = val
        }
        if val, ok := newValues[field]; ok {
            newValuesFiltered[field] = val
        }
    }

    oldValuesJSON, _ := json.Marshal(oldValuesFiltered)
    newValuesJSON, _ := json.Marshal(newValuesFiltered)
    changedFieldsJSON, _ := json.Marshal(changedFields)

    oldValuesStr := string(oldValuesJSON)
    newValuesStr := string(newValuesJSON)
    changedFieldsStr := string(changedFieldsJSON)
    entityType := "PRODUCT"
    module := "MASTER_DATA"
    subModule := "PRODUCT"

    // Determine severity based on changed fields
    severity := models.AuditSeverityInfo
    if contains(changedFields, "basePrice") || contains(changedFields, "baseCost") {
        severity = models.AuditSeverityWarning
    }

    notes := fmt.Sprintf("Product updated: %d fields changed", len(changedFields))

    auditLog := &models.AuditLog{
        TenantID:      auditCtx.TenantID,
        CompanyID:     auditCtx.CompanyID,
        UserID:        auditCtx.UserID,
        RequestID:     auditCtx.RequestID,
        Action:        "PRODUCT_UPDATED",
        EntityType:    &entityType,
        EntityID:      &productID,
        OldValues:     &oldValuesStr,
        NewValues:     &newValuesStr,
        ChangedFields: &changedFieldsStr,
        Status:        models.AuditStatusSuccess,
        Module:        &module,
        SubModule:     &subModule,
        Severity:      severity,
        IPAddress:     auditCtx.IPAddress,
        UserAgent:     auditCtx.UserAgent,
        Notes:         &notes,
    }

    err := s.db.WithContext(ctx).Create(auditLog).Error

    duration := time.Since(startTime).Milliseconds()
    auditLog.Duration = &duration
    s.db.WithContext(ctx).Model(auditLog).Update("duration", duration)

    return err
}

// LogProductDeleted logs when a product is soft-deleted
func (s *AuditService) LogProductDeleted(
    ctx context.Context,
    auditCtx *AuditContext,
    productID string,
    productData map[string]interface{},
    reason string,
) error {
    startTime := time.Now()

    oldValuesJSON, _ := json.Marshal(productData)
    oldValuesStr := string(oldValuesJSON)
    entityType := "PRODUCT"
    module := "MASTER_DATA"
    subModule := "PRODUCT"
    notes := fmt.Sprintf("Product deleted: %v (Reason: %s)", productData["name"], reason)

    // Metadata for deletion context
    metadata := map[string]interface{}{
        "deletionReason": reason,
        "wasActive":      productData["isActive"],
    }
    metadataJSON, _ := json.Marshal(metadata)
    metadataStr := string(metadataJSON)

    auditLog := &models.AuditLog{
        TenantID:   auditCtx.TenantID,
        CompanyID:  auditCtx.CompanyID,
        UserID:     auditCtx.UserID,
        RequestID:  auditCtx.RequestID,
        Action:     "PRODUCT_DELETED",
        EntityType: &entityType,
        EntityID:   &productID,
        OldValues:  &oldValuesStr,
        Status:     models.AuditStatusSuccess,
        Module:     &module,
        SubModule:  &subModule,
        Severity:   models.AuditSeverityWarning, // Deletions always WARNING
        Metadata:   &metadataStr,
        IPAddress:  auditCtx.IPAddress,
        UserAgent:  auditCtx.UserAgent,
        Notes:      &notes,
    }

    err := s.db.WithContext(ctx).Create(auditLog).Error

    duration := time.Since(startTime).Milliseconds()
    auditLog.Duration = &duration
    s.db.WithContext(ctx).Model(auditLog).Update("duration", duration)

    return err
}

// LogProductOperationFailed logs when a product operation fails
func (s *AuditService) LogProductOperationFailed(
    ctx context.Context,
    auditCtx *AuditContext,
    action string,
    productID string,
    errorMsg string,
) error {
    entityType := "PRODUCT"
    module := "MASTER_DATA"
    subModule := "PRODUCT"
    notes := fmt.Sprintf("Product operation failed: %s", errorMsg)

    auditLog := &models.AuditLog{
        TenantID:     auditCtx.TenantID,
        CompanyID:    auditCtx.CompanyID,
        UserID:       auditCtx.UserID,
        RequestID:    auditCtx.RequestID,
        Action:       action,
        EntityType:   &entityType,
        EntityID:     &productID,
        Status:       models.AuditStatusFailed,
        ErrorMessage: &errorMsg,
        Module:       &module,
        SubModule:    &subModule,
        Severity:     models.AuditSeverityWarning,
        IPAddress:    auditCtx.IPAddress,
        UserAgent:    auditCtx.UserAgent,
        Notes:        &notes,
    }

    return s.db.WithContext(ctx).Create(auditLog).Error
}

// Helper function
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

---

### Phase 4: Integrate Audit in ProductService

```go
// internal/service/product/product_service.go

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/shopspring/decimal"
    "gorm.io/gorm"

    "backend/internal/dto"
    "backend/internal/service/audit" // 🆕 ADDED
    "backend/models"
    pkgerrors "backend/pkg/errors"
)

type ProductService struct {
    db           *gorm.DB
    auditService *audit.AuditService // 🆕 ADDED
}

func NewProductService(db *gorm.DB, auditService *audit.AuditService) *ProductService {
    return &ProductService{
        db:           db,
        auditService: auditService, // 🆕 ADDED
    }
}

// CreateProduct with audit logging
func (s *ProductService) CreateProduct(
    ctx context.Context,
    companyID string,
    tenantID string,
    req *dto.CreateProductRequest,
) (*models.Product, error) {
    startTime := time.Now()
    requestID := uuid.New().String() // 🆕 Generate RequestID

    // Get audit context from HTTP context (set by middleware)
    auditCtx := s.getAuditContext(ctx, tenantID, companyID, requestID)

    // Parse decimal fields
    baseCost, err := decimal.NewFromString(req.BaseCost)
    if err != nil {
        // 🆕 LOG FAILED OPERATION
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_CREATED", "",
            fmt.Sprintf("invalid baseCost format: %v", err),
        )
        return nil, pkgerrors.NewBadRequestError("invalid baseCost format")
    }

    basePrice, err := decimal.NewFromString(req.BasePrice)
    if err != nil {
        // 🆕 LOG FAILED OPERATION
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_CREATED", "",
            fmt.Sprintf("invalid basePrice format: %v", err),
        )
        return nil, pkgerrors.NewBadRequestError("invalid basePrice format")
    }

    minimumStock := decimal.Zero
    if req.MinimumStock != "" {
        minimumStock, err = decimal.NewFromString(req.MinimumStock)
        if err != nil {
            s.auditService.LogProductOperationFailed(
                ctx, auditCtx, "PRODUCT_CREATED", "",
                fmt.Sprintf("invalid minimumStock format: %v", err),
            )
            return nil, pkgerrors.NewBadRequestError("invalid minimumStock format")
        }
    }

    // Validate request
    if err := s.validateCreateProduct(ctx, companyID, tenantID, req, baseCost, basePrice, minimumStock); err != nil {
        // 🆕 LOG FAILED VALIDATION
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_CREATED", "",
            fmt.Sprintf("validation failed: %v", err),
        )
        return nil, err
    }

    var product *models.Product

    // Use transaction for atomic create
    err = s.db.WithContext(ctx).Set("tenant_id", tenantID).Transaction(func(tx *gorm.DB) error {
        // 1. Create product
        product = &models.Product{
            CompanyID:      companyID,
            TenantID:       tenantID,
            Code:           req.Code,
            Name:           req.Name,
            Category:       req.Category,
            BaseUnit:       req.BaseUnit,
            BaseCost:       baseCost,
            BasePrice:      basePrice,
            CurrentStock:   decimal.Zero,
            MinimumStock:   minimumStock,
            Description:    req.Description,
            Barcode:        req.Barcode,
            IsBatchTracked: req.IsBatchTracked,
            IsPerishable:   req.IsPerishable,
            IsActive:       true,
        }

        if err := tx.Create(product).Error; err != nil {
            return fmt.Errorf("failed to create product: %w", err)
        }

        // 🆕 LOG PRODUCT CREATION
        productData := map[string]interface{}{
            "code":           product.Code,
            "name":           product.Name,
            "category":       product.Category,
            "baseUnit":       product.BaseUnit,
            "baseCost":       product.BaseCost.String(),
            "basePrice":      product.BasePrice.String(),
            "minimumStock":   product.MinimumStock.String(),
            "isBatchTracked": product.IsBatchTracked,
            "isPerishable":   product.IsPerishable,
        }
        if err := s.auditService.LogProductCreated(ctx, auditCtx, product.ID, productData); err != nil {
            // Log error but don't fail transaction
            fmt.Printf("Failed to log product creation: %v\n", err)
        }

        // 2. Create base unit entry
        baseUnit := &models.ProductUnit{
            ProductID:      product.ID,
            UnitName:       req.BaseUnit,
            ConversionRate: decimal.NewFromInt(1),
            IsBaseUnit:     true,
            BuyPrice:       &baseCost,
            SellPrice:      &basePrice,
            IsActive:       true,
        }

        if err := tx.Create(baseUnit).Error; err != nil {
            return fmt.Errorf("failed to create base unit: %w", err)
        }

        // 3. Create additional units
        for _, unitReq := range req.Units {
            conversionRate, err := decimal.NewFromString(unitReq.ConversionRate)
            if err != nil {
                return pkgerrors.NewBadRequestError(
                    fmt.Sprintf("invalid conversionRate for unit %s", unitReq.UnitName),
                )
            }

            unit := &models.ProductUnit{
                ProductID:      product.ID,
                UnitName:       unitReq.UnitName,
                ConversionRate: conversionRate,
                IsBaseUnit:     false,
                Barcode:        unitReq.Barcode,
                SKU:            unitReq.SKU,
                Description:    unitReq.Description,
                IsActive:       true,
            }

            // Parse optional decimal fields
            if unitReq.BuyPrice != nil {
                buyPrice, err := decimal.NewFromString(*unitReq.BuyPrice)
                if err != nil {
                    return pkgerrors.NewBadRequestError(
                        fmt.Sprintf("invalid buyPrice for unit %s", unitReq.UnitName),
                    )
                }
                unit.BuyPrice = &buyPrice
            }

            if unitReq.SellPrice != nil {
                sellPrice, err := decimal.NewFromString(*unitReq.SellPrice)
                if err != nil {
                    return pkgerrors.NewBadRequestError(
                        fmt.Sprintf("invalid sellPrice for unit %s", unitReq.UnitName),
                    )
                }
                unit.SellPrice = &sellPrice
            }

            if unitReq.Weight != nil {
                weight, err := decimal.NewFromString(*unitReq.Weight)
                if err != nil {
                    return pkgerrors.NewBadRequestError(
                        fmt.Sprintf("invalid weight for unit %s", unitReq.UnitName),
                    )
                }
                unit.Weight = &weight
            }

            if unitReq.Volume != nil {
                volume, err := decimal.NewFromString(*unitReq.Volume)
                if err != nil {
                    return pkgerrors.NewBadRequestError(
                        fmt.Sprintf("invalid volume for unit %s", unitReq.UnitName),
                    )
                }
                unit.Volume = &volume
            }

            if err := tx.Create(unit).Error; err != nil {
                return fmt.Errorf("failed to create product unit: %w", err)
            }
        }

        // 4. Initialize warehouse stocks
        var warehouses []models.Warehouse
        if err := tx.Where("company_id = ? AND is_active = ?", companyID, true).Find(&warehouses).Error; err != nil {
            return fmt.Errorf("failed to get warehouses: %w", err)
        }

        for _, wh := range warehouses {
            whStock := &models.WarehouseStock{
                WarehouseID:  wh.ID,
                ProductID:    product.ID,
                Quantity:     decimal.Zero,
                MinimumStock: minimumStock,
            }

            if err := tx.Create(whStock).Error; err != nil {
                return fmt.Errorf("failed to initialize warehouse stock: %w", err)
            }
        }

        return nil
    })

    if err != nil {
        // 🆕 LOG TRANSACTION FAILURE
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_CREATED", "",
            fmt.Sprintf("transaction failed: %v", err),
        )
        return nil, err
    }

    // ✅ Operation succeeded
    duration := time.Since(startTime).Milliseconds()
    fmt.Printf("✅ Product created successfully in %dms (RequestID: %s)\n", duration, requestID)

    // Reload product with relations
    return s.GetProduct(ctx, companyID, tenantID, product.ID)
}

// UpdateProduct with audit logging
func (s *ProductService) UpdateProduct(
    ctx context.Context,
    companyID, tenantID, productID string,
    req *dto.UpdateProductRequest,
) (*models.Product, error) {
    startTime := time.Now()
    requestID := uuid.New().String()
    auditCtx := s.getAuditContext(ctx, tenantID, companyID, requestID)

    // Get existing product
    product, err := s.GetProduct(ctx, companyID, tenantID, productID)
    if err != nil {
        return nil, err
    }

    // 🆕 Capture old values BEFORE update
    oldValues := map[string]interface{}{
        "code":           product.Code,
        "name":           product.Name,
        "category":       product.Category,
        "baseUnit":       product.BaseUnit,
        "baseCost":       product.BaseCost.String(),
        "basePrice":      product.BasePrice.String(),
        "minimumStock":   product.MinimumStock.String(),
        "description":    product.Description,
        "barcode":        product.Barcode,
        "isBatchTracked": product.IsBatchTracked,
        "isPerishable":   product.IsPerishable,
        "isActive":       product.IsActive,
    }

    // Validate updates
    if err := s.validateUpdateProduct(ctx, companyID, product.TenantID, productID, req); err != nil {
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_UPDATED", productID,
            fmt.Sprintf("validation failed: %v", err),
        )
        return nil, err
    }

    // Build updates map
    updates := make(map[string]interface{})
    changedFields := []string{} // 🆕 Track changed fields

    if req.Name != nil {
        updates["name"] = *req.Name
        changedFields = append(changedFields, "name")
    }

    if req.Category != nil {
        updates["category"] = *req.Category
        changedFields = append(changedFields, "category")
    }

    if req.Code != nil {
        updates["code"] = *req.Code
        changedFields = append(changedFields, "code")
    }

    if req.BaseUnit != nil {
        updates["base_unit"] = *req.BaseUnit
        changedFields = append(changedFields, "baseUnit")
    }

    if req.BaseCost != nil {
        baseCost, err := decimal.NewFromString(*req.BaseCost)
        if err != nil {
            s.auditService.LogProductOperationFailed(
                ctx, auditCtx, "PRODUCT_UPDATED", productID,
                fmt.Sprintf("invalid baseCost format: %v", err),
            )
            return nil, pkgerrors.NewBadRequestError("invalid baseCost format")
        }
        updates["base_cost"] = baseCost
        changedFields = append(changedFields, "baseCost")
    }

    if req.BasePrice != nil {
        basePrice, err := decimal.NewFromString(*req.BasePrice)
        if err != nil {
            s.auditService.LogProductOperationFailed(
                ctx, auditCtx, "PRODUCT_UPDATED", productID,
                fmt.Sprintf("invalid basePrice format: %v", err),
            )
            return nil, pkgerrors.NewBadRequestError("invalid basePrice format")
        }
        updates["base_price"] = basePrice
        changedFields = append(changedFields, "basePrice")
    }

    if req.MinimumStock != nil {
        minimumStock, err := decimal.NewFromString(*req.MinimumStock)
        if err != nil {
            s.auditService.LogProductOperationFailed(
                ctx, auditCtx, "PRODUCT_UPDATED", productID,
                fmt.Sprintf("invalid minimumStock format: %v", err),
            )
            return nil, pkgerrors.NewBadRequestError("invalid minimumStock format")
        }
        updates["minimum_stock"] = minimumStock
        changedFields = append(changedFields, "minimumStock")
    }

    if req.Description != nil {
        updates["description"] = *req.Description
        changedFields = append(changedFields, "description")
    }

    if req.Barcode != nil {
        updates["barcode"] = *req.Barcode
        changedFields = append(changedFields, "barcode")
    }

    if req.IsBatchTracked != nil {
        updates["is_batch_tracked"] = *req.IsBatchTracked
        changedFields = append(changedFields, "isBatchTracked")
    }

    if req.IsPerishable != nil {
        updates["is_perishable"] = *req.IsPerishable
        changedFields = append(changedFields, "isPerishable")
    }

    if req.IsActive != nil {
        updates["is_active"] = *req.IsActive
        changedFields = append(changedFields, "isActive")
    }

    // 🆕 If no fields changed, skip update
    if len(changedFields) == 0 {
        return product, nil
    }

    // Update product
    if err := s.db.WithContext(ctx).Set("tenant_id", tenantID).
        Model(&models.Product{}).
        Where("id = ? AND company_id = ?", productID, companyID).
        Updates(updates).Error; err != nil {

        // 🆕 LOG FAILED UPDATE
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_UPDATED", productID,
            fmt.Sprintf("database update failed: %v", err),
        )
        return nil, fmt.Errorf("failed to update product: %w", err)
    }

    // Reload product
    updatedProduct, err := s.GetProduct(ctx, companyID, tenantID, productID)
    if err != nil {
        return nil, err
    }

    // 🆕 Capture new values AFTER update
    newValues := map[string]interface{}{
        "code":           updatedProduct.Code,
        "name":           updatedProduct.Name,
        "category":       updatedProduct.Category,
        "baseUnit":       updatedProduct.BaseUnit,
        "baseCost":       updatedProduct.BaseCost.String(),
        "basePrice":      updatedProduct.BasePrice.String(),
        "minimumStock":   updatedProduct.MinimumStock.String(),
        "description":    updatedProduct.Description,
        "barcode":        updatedProduct.Barcode,
        "isBatchTracked": updatedProduct.IsBatchTracked,
        "isPerishable":   updatedProduct.IsPerishable,
        "isActive":       updatedProduct.IsActive,
    }

    // 🆕 LOG SUCCESSFUL UPDATE
    if err := s.auditService.LogProductUpdated(
        ctx, auditCtx, productID, oldValues, newValues, changedFields,
    ); err != nil {
        fmt.Printf("Failed to log product update: %v\n", err)
    }

    duration := time.Since(startTime).Milliseconds()
    fmt.Printf("✅ Product updated successfully in %dms (RequestID: %s, Changed: %v)\n",
        duration, requestID, changedFields)

    return updatedProduct, nil
}

// DeleteProduct with audit logging
func (s *ProductService) DeleteProduct(
    ctx context.Context,
    companyID, productID string,
) error {
    startTime := time.Now()
    requestID := uuid.New().String()

    // Get tenant from product
    product, err := s.GetProduct(ctx, companyID, "", productID)
    if err != nil {
        return err
    }

    auditCtx := s.getAuditContext(ctx, product.TenantID, companyID, requestID)

    // Validate deletion
    if err := s.validateDeleteProduct(ctx, product); err != nil {
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_DELETED", productID,
            fmt.Sprintf("validation failed: %v", err),
        )
        return err
    }

    // 🆕 Capture product data before deletion
    productData := map[string]interface{}{
        "code":           product.Code,
        "name":           product.Name,
        "category":       product.Category,
        "basePrice":      product.BasePrice.String(),
        "baseCost":       product.BaseCost.String(),
        "isBatchTracked": product.IsBatchTracked,
        "isPerishable":   product.IsPerishable,
        "isActive":       product.IsActive,
    }

    // Soft delete
    if err := s.db.WithContext(ctx).Model(product).Update("is_active", false).Error; err != nil {
        s.auditService.LogProductOperationFailed(
            ctx, auditCtx, "PRODUCT_DELETED", productID,
            fmt.Sprintf("soft delete failed: %v", err),
        )
        return fmt.Errorf("failed to delete product: %w", err)
    }

    // 🆕 LOG SUCCESSFUL DELETION
    if err := s.auditService.LogProductDeleted(
        ctx, auditCtx, productID, productData, "User requested deletion",
    ); err != nil {
        fmt.Printf("Failed to log product deletion: %v\n", err)
    }

    duration := time.Since(startTime).Milliseconds()
    fmt.Printf("✅ Product deleted successfully in %dms (RequestID: %s)\n", duration, requestID)

    return nil
}

// 🆕 Helper: Extract audit context from HTTP context
func (s *ProductService) getAuditContext(
    ctx context.Context,
    tenantID, companyID, requestID string,
) *audit.AuditContext {
    // These would be set by middleware from JWT claims and HTTP headers
    userID, _ := ctx.Value("user_id").(string)
    ipAddress, _ := ctx.Value("ip_address").(string)
    userAgent, _ := ctx.Value("user_agent").(string)

    return &audit.AuditContext{
        TenantID:  &tenantID,
        CompanyID: &companyID,
        UserID:    &userID,
        RequestID: &requestID,
        IPAddress: &ipAddress,
        UserAgent: &userAgent,
    }
}
```

---

## 📅 IMPLEMENTATION ROADMAP - MVP PHASED APPROACH

> **⚡ RECOMMENDED:** Start with MVP Phase 1 (1-2 weeks), validate, then decide on Phase 2

---

## 🎯 MVP PHASE 1: CRITICAL Foundation (1-2 Weeks) ⭐ START HERE

**Goal:** Production-ready audit logging dengan minimal fields

**Current Status:**
- ✅ Model Updated (2026-01-05)
- ⏳ Implementation Pending

### Week 1: MVP Implementation (Jan 5-12, 2026)

#### **✅ Day 1: Model Update (COMPLETED - Jan 5, 2026)**
```bash
# Completed Tasks:
- [x] Update AuditLog struct (3 fields: CompanyID, RequestID, Status)
- [x] Add Company relation untuk foreign key
- [x] File updated: models/system.go
```

**Model Changes Applied:**
```go
// ✅ 3 fields added to AuditLog:
CompanyID *string `gorm:"type:varchar(255);index"`
RequestID *string `gorm:"type:varchar(100);index"`
Status    string  `gorm:"type:varchar(20);default:'SUCCESS';index"`

// ✅ Company relation added:
Company *Company `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE"`
```

#### **⏳ Day 2: Migration & Constants (NEXT)**
```bash
# Pending Tasks:
- [ ] Run migration: go run cmd/migrate/main.go
- [ ] Verify tabel audit_logs created/updated
- [ ] Add Status constants (SUCCESS, FAILED, PARTIAL)
- [ ] Update AuditContext struct (add CompanyID, RequestID)
- [ ] Test compilation
```

#### **Day 3-4: Implement Audit Methods (Basic)**
```bash
# Tasks:
- [ ] LogProductCreated (minimal version)
- [ ] LogProductUpdated (minimal version)
- [ ] LogProductDeleted (minimal version)
- [ ] LogProductOperationFailed (basic error tracking)
```

**Focus:** Simple, working implementation. NO advanced features yet.

#### **Day 5: Integration & Testing**
```bash
# Tasks:
- [ ] Integrate audit calls in ProductService.CreateProduct
- [ ] Integrate audit calls in ProductService.UpdateProduct
- [ ] Integrate audit calls in ProductService.DeleteProduct
- [ ] Unit tests for audit methods
- [ ] Integration tests (create product → verify audit)
- [ ] Test failure scenarios
```

**Deliverable:** ✅ Production-ready MVP with 3 critical fields

---

### Week 2: Validation & Deployment (Jan 15-19, 2026)

#### **Day 1-2: Comprehensive Testing**
```bash
# Tasks:
- [ ] Load testing (100+ operations)
- [ ] Verify audit log creation
- [ ] Test query performance
- [ ] Validate tenant isolation
- [ ] Test RequestID grouping
```

#### **Day 3: Staging Deployment**
```bash
# Tasks:
- [ ] Deploy to staging
- [ ] Run smoke tests
- [ ] Monitor error rates
- [ ] Gather QA feedback
```

#### **Day 4: Production Deployment**
```bash
# Tasks:
- [ ] Create rollback plan
- [ ] Deploy to production
- [ ] Monitor system health
- [ ] Verify audit logs being created
```

#### **Day 5: Post-Deployment & Evaluation**
```bash
# Tasks:
- [ ] Monitor audit log growth
- [ ] Check query performance
- [ ] Document lessons learned
- [ ] 🎯 DECISION POINT: Continue to Phase 2?
```

**Deliverable:** ✅ MVP in production, gathering real usage data

---

## 🔄 PHASE 2: Enhanced Capabilities (2-3 Weeks) - Optional

> **⏸️ PAUSE:** Only proceed if MVP proves insufficient or specific needs identified

**Add 4 HIGH priority fields:**
- ErrorMessage - Error details
- ChangedFields - Efficient UPDATE tracking
- Module - Categorization
- SubModule - Sub-categorization

### Week 3-4: Phase 2 Implementation

#### **Week 3: Schema & Implementation**
```bash
Day 1-2: Database migration (4 fields)
Day 3-4: Update audit methods (add new fields)
Day 5: Testing
```

#### **Week 4: Integration & Deployment**
```bash
Day 1-2: Integration with ProductService
Day 3: Staging deployment
Day 4: Production deployment
Day 5: Monitoring & evaluation
```

**Deliverable:** ✅ Enhanced audit system with 7 fields total

---

## 🚀 PHASE 3: Advanced Features (Future) - When Needed

> **💡 TIP:** Most projects never need Phase 3. Implement only if specific use cases arise.

**Add 3 NICE-TO-HAVE fields:**
- Duration - Performance monitoring
- Severity - Alerting system
- Metadata - Extensibility

**Timeline:** TBD based on actual needs

---

## 📊 Phase Comparison

| Aspect | MVP Phase 1 | Phase 2 | Phase 3 |
|--------|-------------|---------|---------|
| **Fields Added** | 3 (CRITICAL) | +4 (HIGH) | +3 (NICE) |
| **Timeline** | 1-2 weeks | 2-3 weeks | TBD |
| **Risk** | Low | Medium | Low |
| **Production Ready** | ✅ YES | ✅ YES | ✅ YES |
| **Rollback Complexity** | Simple | Medium | Simple |
| **Use Cases** | 90% needs | 95% needs | 99% needs |

---

## 🎯 Decision Framework

### After MVP Phase 1:

**✅ Continue to Phase 2 if:**
- [ ] Need detailed error messages for debugging
- [ ] UPDATE operations are frequent and need efficiency
- [ ] Need audit log reporting by module
- [ ] Have resources and time available

**⏸️ Stay at MVP Phase 1 if:**
- [ ] Basic audit trail is sufficient
- [ ] No complaints from users
- [ ] Want to focus on other features
- [ ] Resource constraints

**❌ Skip Phase 2 if:**
- [ ] MVP fully meets requirements
- [ ] No identified use cases for additional fields
- [ ] Cost/benefit doesn't justify enhancement

---

---

## 🎯 NEXT STEPS

### Immediate Actions (This Week)

1. **Review & Approve**
   - [ ] Review all 9 proposed fields
   - [ ] Prioritize: Which fields are MVP vs. Phase 2?
   - [ ] Approve database schema changes
   - [ ] Sign-off on implementation approach

2. **Technical Preparation**
   - [ ] Backup production database
   - [ ] Create feature branch: `feature/product-audit-logging`
   - [ ] Set up development database for testing
   - [ ] Prepare rollback plan

3. **Resource Allocation**
   - [ ] Assign developer(s) to implementation
   - [ ] Schedule code review sessions
   - [ ] Allocate QA resources for testing
   - [ ] Plan deployment window

### Decision Points

**🔴 CRITICAL Decisions Needed:**
1. **CompanyID**: ✅ APPROVE or ❌ DEFER?
2. **RequestID**: ✅ APPROVE or ❌ DEFER?
3. **Status**: ✅ APPROVE or ❌ DEFER?

**🟡 HIGH Priority Decisions:**
4. **ChangedFields**: ✅ APPROVE or ❌ DEFER?
5. **Module/SubModule**: ✅ APPROVE or ❌ DEFER?
6. **ErrorMessage**: ✅ APPROVE or ❌ DEFER?

**🟢 NICE-TO-HAVE Decisions:**
7. **Duration**: ✅ APPROVE or ❌ DEFER to Phase 2?
8. **Severity**: ✅ APPROVE or ❌ DEFER to Phase 2?
9. **Metadata**: ✅ APPROVE or ❌ DEFER to Phase 2?

### Questions for Stakeholders

1. **Compliance Requirements:**
   - What regulatory compliance needs do we have for audit trails?
   - Are there specific retention policies for audit logs?
   - Do we need to support audit log exports for external auditors?

2. **Performance Concerns:**
   - What is acceptable performance overhead for audit logging? (<5%, <10%, <15%?)
   - What is expected audit log growth rate? (Records per day)
   - How long should we retain audit logs? (30 days, 90 days, 1 year, forever?)

3. **Access Control:**
   - Who should have access to audit logs?
   - Should audit logs be visible in the main application UI?
   - Do we need a separate audit log viewer/dashboard?

4. **Alerting & Monitoring:**
   - Should we alert on failed operations?
   - Do we need real-time audit log monitoring?
   - What metrics should we track? (Operation success rate, average duration, error patterns)

---

## 📊 SUMMARY

### Current State: ❌ NO AUDIT LOGGING
- Product INSERT: **NOT LOGGED**
- Product UPDATE: **NOT LOGGED**
- Product DELETE: **NOT LOGGED**

### Proposed Enhancement: ✅ COMPREHENSIVE AUDIT SYSTEM
- **9 new fields** to enhance audit_logs table
- **4 new audit methods** for product operations
- **Complete integration** with ProductService
- **Error tracking** for failed operations
- **Transaction grouping** via RequestID
- **Performance tracking** via Duration
- **Flexible metadata** for extensibility

### Expected Benefits
✅ **Compliance**: Complete audit trail for regulatory requirements
✅ **Debugging**: Track failed operations with error details
✅ **Security**: Know who changed what, when, and from where
✅ **Performance**: Identify slow operations and bottlenecks
✅ **Reporting**: Generate audit reports by module, user, company
✅ **Accountability**: Clear responsibility for all operations

### Implementation Effort
- **4 weeks** total implementation time
- **Low risk** - non-breaking changes
- **Minimal performance impact** - estimated <5% overhead
- **High value** - essential for production ERP system

---

## 📞 CONTACT & SUPPORT

**Questions or Clarifications?**
- Create GitHub issue: `[AUDIT] Your question here`
- Contact: Backend Team Lead
- Documentation: See `CLAUDE.md` for audit patterns

**Review Status:** 🟡 PENDING APPROVAL

---

*Document prepared by: Claude Code Analysis System*
*Date: 2026-01-05*
*Version: 1.0*
