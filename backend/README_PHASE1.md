# Phase 1 Implementation - GORM Migration
## Multi-Tenant ERP System - Core Models

**Status:** ✅ COMPLETED (2025-12-15)
**Test Coverage:** 10/10 PASSED
**Schema Parity:** 100% with Prisma

---

## Quick Start

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Run Migration
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

    // Run Phase 1 migration
    if err := db.AutoMigratePhase1(database); err != nil {
        log.Fatal(err)
    }

    log.Println("✅ Phase 1 migration complete!")
}
```

### 3. Run Tests
```bash
go test -v ./models/
```

Expected output:
```
✅ TestSchemaGeneration
✅ TestCUIDGeneration
✅ TestUserCreation
✅ TestCompanyCreation
✅ TestTenantWithSubscription
✅ TestUserTenantJunction
✅ TestUniqueConstraints
✅ TestCascadeDelete
✅ TestDecimalPrecision
✅ TestEnumValues

PASS
ok      backend/models    1.233s
```

---

## What's Included

### 📦 Models (7 total)
1. **User** - Application users with multi-tenant access
2. **UserTenant** - Junction table (User ↔ Tenant) with roles
3. **Company** - Legal entity with Indonesian tax compliance
4. **CompanyBank** - Company bank accounts
5. **Tenant** - PT/CV subscription instances
6. **Subscription** - Billing & payment tracking
7. **SubscriptionPayment** - Payment history

### 🎯 Features Implemented
- ✅ CUID auto-generation for all models
- ✅ Multi-tenant architecture (User can access multiple tenants)
- ✅ Indonesian tax compliance (NPWP, PKP, PPN 11%)
- ✅ Subscription billing with custom pricing per tenant
- ✅ 14-day trial period support
- ✅ Role-based access control per tenant
- ✅ Decimal precision for money amounts
- ✅ Cascade deletion rules
- ✅ Comprehensive test suite

### 📝 Enum Types (All 17 defined)
- UserRole (OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF)
- TenantStatus (TRIAL, ACTIVE, SUSPENDED, CANCELLED, EXPIRED)
- SubscriptionStatus (ACTIVE, PAST_DUE, CANCELLED, EXPIRED)
- SubscriptionPaymentStatus (PENDING, PAID, FAILED, REFUNDED, CANCELLED)
- + 13 more for future phases

---

## Usage Examples

### Create User
```go
user := &models.User{
    Email:    "john@company.com",
    Username: "johndoe",
    Password: "$2a$10$hashedPassword", // bcrypt
    Name:     "John Doe",
}
db.Create(user)
// user.ID is auto-generated via CUID
```

### Create Company with Tax Setup
```go
npwp := "01.234.567.8-901.234"
company := &models.Company{
    Name:       "CV Maju Bersama",
    LegalName:  "CV Maju Bersama Sejahtera",
    EntityType: "CV",
    Address:    "Jl. Sudirman No. 123",
    City:       "Jakarta Selatan",
    Province:   "DKI Jakarta",
    Phone:      "021-7654321",
    Email:      "info@majubersama.co.id",
    NPWP:       &npwp,
    IsPKP:      true,
    PPNRate:    decimal.NewFromFloat(11.0),
}
db.Create(company)
```

### Create Tenant with Trial
```go
trialEnds := time.Now().AddDate(0, 0, 14) // 14 days
tenant := &models.Tenant{
    CompanyID:   company.ID,
    Status:      models.TenantStatusTrial,
    TrialEndsAt: &trialEnds,
}
db.Create(tenant)
```

### Assign User to Tenant with Role
```go
userTenant := &models.UserTenant{
    UserID:   user.ID,
    TenantID: tenant.ID,
    Role:     models.UserRoleOwner,
    IsActive: true,
}
db.Create(userTenant)
```

### Query User's Tenants
```go
var userTenants []models.UserTenant
db.Preload("Tenant").
   Preload("Tenant.Company").
   Where("user_id = ? AND is_active = ?", userID, true).
   Find(&userTenants)

for _, ut := range userTenants {
    fmt.Printf("Company: %s, Role: %s\n",
        ut.Tenant.Company.Name, ut.Role)
}
```

### Complete Onboarding Flow
See `examples/phase1_usage.go` for full example including:
- Company creation with tax setup
- Tenant with trial period
- Owner user creation
- Role assignment
- Bank account setup
- Subscription activation

---

## Database Schema

### Tables Created (7 total)
```
users
├── id (varchar 255, PK)
├── email (varchar 255, unique)
├── username (varchar 255, unique)
├── password (varchar 255)
├── name (varchar 255)
├── is_system_admin (boolean, default false)
├── is_active (boolean, default true)
├── created_at (datetime)
└── updated_at (datetime)

companies
├── id (varchar 255, PK)
├── name, legal_name, entity_type
├── address, city, province, postal_code, country
├── phone, email, website
├── npwp (varchar 50, unique) ← Indonesian Tax ID
├── is_pkp (boolean, default false)
├── ppn_rate (decimal 5,2, default 11)
├── faktur_pajak_series, sppkp_number
├── logo_url, primary_color, secondary_color
├── invoice_prefix, invoice_number_format, invoice_footer, invoice_terms
├── so_prefix, so_number_format
├── po_prefix, po_number_format
├── currency (default IDR), timezone (default Asia/Jakarta), locale (default id-ID)
├── business_hours_start, business_hours_end, working_days
├── is_active, created_at, updated_at

company_banks
├── id (varchar 255, PK)
├── company_id (varchar 255, FK → companies, CASCADE)
├── bank_name, account_number, account_name, branch_name
├── is_primary (boolean, default false)
├── check_prefix
├── is_active, created_at, updated_at

subscriptions
├── id (varchar 255, PK)
├── price (decimal 15,2, default 300000)
├── billing_cycle (default MONTHLY)
├── status (varchar 20, default ACTIVE)
├── current_period_start, current_period_end, next_billing_date
├── payment_method, last_payment_date, last_payment_amount
├── grace_period_ends
├── auto_renew (boolean, default true)
├── cancelled_at, cancellation_reason
├── created_at, updated_at

tenants
├── id (varchar 255, PK)
├── company_id (varchar 255, unique, FK → companies, CASCADE)
├── subscription_id (varchar 255, FK → subscriptions)
├── status (varchar 20, default TRIAL)
├── trial_ends_at (datetime)
├── notes (text)
├── created_at, updated_at

subscription_payments
├── id (varchar 255, PK)
├── subscription_id (varchar 255, FK → subscriptions, CASCADE)
├── amount (decimal 15,2)
├── payment_date, payment_method, status
├── reference, invoice_number (unique)
├── period_start, period_end
├── paid_at, notes
├── created_at, updated_at

user_tenants (junction table)
├── id (varchar 255, PK)
├── user_id (varchar 255, FK → users, CASCADE)
├── tenant_id (varchar 255, FK → tenants, CASCADE)
├── role (varchar 20, default STAFF)
├── is_active (boolean, default true)
├── created_at, updated_at
└── UNIQUE(user_id, tenant_id)
```

---

## Key Patterns

### 1. CUID Generation
All models use CUID (Collision-resistant Unique ID) via BeforeCreate hook:
```go
func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
    if m.ID == "" {
        m.ID = cuid.New()
    }
    return nil
}
```

### 2. Multi-Tenant Access
A user can access multiple tenants with different roles:
```
User "john@company.com"
├── Tenant A (CV Maju) → Role: OWNER
├── Tenant B (PT Sejahtera) → Role: ADMIN
└── Tenant C (CV Berkah) → Role: FINANCE
```

### 3. Indonesian Tax Compliance
```go
company.NPWP = "01.234.567.8-901.234" // Tax ID
company.IsPKP = true                  // Taxable entrepreneur
company.PPNRate = 11.0                // 11% VAT (2025)
```

### 4. Subscription Lifecycle
```
TRIAL (14 days free)
    ↓ (payment)
ACTIVE (paid subscription)
    ↓ (payment failed)
SUSPENDED (grace period 7 days)
    ↓ (grace period expired)
EXPIRED
    OR
CANCELLED (user cancels)
```

### 5. Cascade Deletion
```go
// Delete company → automatically deletes company_banks
db.Select("Banks").Delete(&company)

// Delete tenant → automatically deletes user_tenants
db.Select("Users").Delete(&tenant)
```

---

## Testing

### Run All Tests
```bash
go test -v ./models/
```

### Run Specific Test
```bash
go test -v ./models/ -run TestUserCreation
```

### Test Coverage
```bash
go test -cover ./models/
```

### Test with Race Detector
```bash
go test -race ./models/
```

---

## Migration from Prisma

### Data Migration (if needed)
1. Export data from Prisma database
2. Transform IDs from Prisma CUID to Go CUID format (should match)
3. Import to GORM database
4. Verify relationships and constraints

### Schema Comparison
```bash
# Generate Prisma SQL
npx prisma migrate diff --from-schema-datamodel schema.prisma --to-empty --script

# Generate GORM SQL
# Use GORM Migrator to get CREATE TABLE statements
```

---

## Production Checklist

### Before Deployment
- [ ] Switch from SQLite to PostgreSQL
- [ ] Test on PostgreSQL database
- [ ] Enable connection pooling
- [ ] Set up database backups
- [ ] Configure logging
- [ ] Implement password hashing (bcrypt/argon2)
- [ ] Add input validation
- [ ] Set up monitoring
- [ ] Load test subscription flows
- [ ] Test multi-tenant isolation

### Security
- [ ] Hash all passwords before storage
- [ ] Encrypt sensitive fields (NPWP)
- [ ] Implement rate limiting
- [ ] Add audit logging
- [ ] Validate tenant isolation in all queries
- [ ] Set up HTTPS/TLS
- [ ] Configure CORS properly

### Performance
- [ ] Add indexes based on query patterns
- [ ] Enable query logging in development
- [ ] Optimize N+1 queries with Preload
- [ ] Consider read replicas for scaling
- [ ] Cache frequently accessed data

---

## Next Steps

### Phase 2: Product & Inventory (Estimated: 2 days)
- [ ] Product, ProductUnit, ProductBatch
- [ ] Warehouse, WarehouseStock
- [ ] PriceList, ProductSupplier

### Phase 3: Transactions (Estimated: 2 days)
- [ ] SalesOrder, Invoice, Payment
- [ ] PurchaseOrder, GoodsReceipt
- [ ] Delivery, SupplierPayment

### Phase 4: Supporting Modules (Estimated: 1 day)
- [ ] InventoryMovement, StockOpname, StockTransfer
- [ ] CashTransaction
- [ ] Setting, AuditLog

### Phase 5: Testing & Production (Estimated: 1.5 days)
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Production deployment

---

## Troubleshooting

### CUID Not Generated
**Problem:** ID field is empty after Create
**Solution:** Ensure BeforeCreate hook is defined and model embeds it correctly

### CASCADE Delete Not Working
**Problem:** Child records not deleted with parent
**Solution:** Use `db.Select("RelationName").Delete()` for has-many relations

### Unique Constraint Violation
**Problem:** Duplicate email/username error
**Solution:** Check for existing records before Create, handle error gracefully

### Foreign Key Constraint Failed (SQLite)
**Problem:** Cannot delete parent with existing children
**Solution:** Enable foreign keys with `PRAGMA foreign_keys = ON`

### Decimal Precision Loss
**Problem:** Money amounts not accurate
**Solution:** Use `decimal.Decimal` type, never float64 for money

---

## Documentation

- **Migration Guide:** `claudedocs/prisma-to-gorm-migration-guide.md`
- **Phase 1 Summary:** `claudedocs/PHASE1_IMPLEMENTATION_SUMMARY.md`
- **Usage Examples:** `examples/phase1_usage.go`
- **Test Suite:** `models/models_test.go`

---

## Support

### Resources
- [GORM Documentation](https://gorm.io/docs/)
- [Prisma Schema Reference](https://www.prisma.io/docs/reference/api-reference/prisma-schema-reference)
- [Go Decimal Package](https://github.com/shopspring/decimal)
- [CUID Package](https://github.com/lucsky/cuid)

### Common Issues
See [Troubleshooting](#troubleshooting) section above

---

**Phase 1 Status:** ✅ COMPLETED
**Last Updated:** 2025-12-15
**Version:** 1.0
