# PHASE 1: Multi-Company Architecture Migration

## Quick Start

### 1. Run Migration
```bash
cd /Users/christianhandoko/Development/work/erp/backend
go run cmd/migrate/phase1_multi_company.go
```

This will:
- Create/update core tables (tenants, companies, user_company_roles)
- Add CompanyID to all transactional tables
- Validate schema structure

### 2. Seed Test Data
```bash
go run cmd/seed/phase1_seed.go
```

This creates:
- 1 Tenant: "PT Multi Bisnis Group"
- 3 Companies: PT Distribusi Utama, CV Sembako Jaya, PT Retail Nusantara
- 5 Users with various permission levels
- Sample master data (warehouses, products, customers)

### 3. Validate Schema
```bash
go run cmd/validate/phase1_validate.go
```

This verifies:
- Tenant-Company relationship (1:N)
- UserCompanyRole constraints
- CompanyID in all transactional tables
- Data integrity
- Permission system setup

## Test Credentials

All users have password: `password123`

**Users**:
1. `budi@example.com` - OWNER (access to all 3 companies)
2. `siti@example.com` - TENANT_ADMIN + ADMIN at PT Distribusi, STAFF at CV Sembako
3. `ahmad@example.com` - FINANCE only at CV Sembako
4. `joko@example.com` - WAREHOUSE at PT Distribusi and CV Sembako
5. `dewi@example.com` - SALES at all 3 companies

## Troubleshooting

### "Failed to connect to database"
Make sure `.env` file exists with correct database configuration:
```env
DB_DRIVER=sqlite  # or postgres
DATABASE_URL=postgres://user:pass@localhost:5432/dbname  # if using postgres
```

### "Table already exists"
GORM AutoMigrate will update existing tables. It's safe to run multiple times.

### "Validation failed"
Check the specific error message. Common issues:
- Missing CompanyID in transactional tables
- Incorrect Tenant-Company relationship
- Invalid role constraints in user_company_roles

## Next Steps

After successful migration and validation:
1. Continue with remaining transactional models (Delivery, GoodsReceipt, etc.)
2. Proceed to PHASE 2: Backend API implementation
3. Proceed to PHASE 3: Frontend integration

## Documentation

See `claudedocs/PHASE1_Implementation_Summary.md` for complete implementation details.
