# Seed Data Documentation

This document describes the test users created by the database seeding script for development and testing purposes.

## Quick Start

To seed the database with test users:

```bash
# Using Make
make seed

# Or directly with Go
go run cmd/seed/main.go
```

## Default Password

**All test users share the same password for convenience:**
- Password: `Password123!`

## Test Data Overview

The seeding script creates:
- 1 System Administrator (cross-tenant access)
- 2 Companies (PT and CV)
- 2 Tenants (1 TRIAL, 1 ACTIVE)
- 11 Total Users (10 regular + 1 system admin)
- 1 Multi-tenant User (consultant)

---

## Company 1: PT Maju Jaya Distribusi

**Tenant Status:** TRIAL (14 days remaining)
**Company Type:** PT (Perseroan Terbatas)
**Location:** Jakarta Utara, DKI Jakarta

### Users

| Email                        | Username           | Name                | Role      |
|------------------------------|-------------------|---------------------|-----------|
| owner.maju@example.com       | owner_maju        | Budi Santoso        | OWNER     |
| admin.maju@example.com       | admin_maju        | Siti Aminah         | ADMIN     |
| finance.maju@example.com     | finance_maju      | Andi Wijaya         | FINANCE   |
| sales.maju@example.com       | sales_maju        | Dewi Lestari        | SALES     |
| warehouse.maju@example.com   | warehouse_maju    | Joko Susilo         | WAREHOUSE |
| staff.maju@example.com       | staff_maju        | Nina Kusuma         | STAFF     |

### Role Permissions

- **OWNER**: Full system access, company settings, user management
- **ADMIN**: Full operational access, cannot modify company settings
- **FINANCE**: Financial transactions, invoices, payments, cash management
- **SALES**: Sales orders, customer management, deliveries
- **WAREHOUSE**: Inventory, stock movements, goods receipts
- **STAFF**: Read-only access to most features

---

## Company 2: CV Berkah Mandiri

**Tenant Status:** ACTIVE (with paid subscription)
**Subscription:** Rp 300,000/month (auto-renew enabled)
**Company Type:** CV (Commanditaire Vennootschap)
**Location:** Surabaya, Jawa Timur

### Users

| Email                        | Username           | Name                | Role      |
|------------------------------|-------------------|---------------------|-----------|
| owner.berkah@example.com     | owner_berkah      | Hendra Gunawan      | OWNER     |
| admin.berkah@example.com     | admin_berkah      | Rina Wijayanti      | ADMIN     |
| finance.berkah@example.com   | finance_berkah    | Tono Hartono        | FINANCE   |
| sales.berkah@example.com     | sales_berkah      | Maya Sari           | SALES     |

---

## Special Users

### System Administrator
- **Email:** superadmin@example.com
- **Username:** superadmin
- **Name:** System Administrator
- **Role:** SYSADMIN (cross-tenant access)
- **Permissions:** Can access and manage all tenants

### Multi-Tenant User
- **Email:** consultant@example.com
- **Username:** consultant
- **Name:** Ahmad Konsultan
- **Role:** STAFF (in both companies)
- **Access:** Can switch between PT Maju Jaya and CV Berkah Mandiri

---

## Testing Scenarios

### 1. Test Multi-Tenant Access
Login as `consultant@example.com` to verify:
- User can see both companies in tenant selector
- Switching tenants properly isolates data
- STAFF role permissions work correctly

### 2. Test Role-Based Access Control
Login as different roles to verify:
- OWNER can access all features
- FINANCE can only access financial modules
- SALES can only access sales modules
- WAREHOUSE can only access inventory modules
- STAFF has read-only access

### 3. Test Trial vs Active Subscription
Compare access between:
- PT Maju Jaya (TRIAL) - Should have trial banner/warnings
- CV Berkah Mandiri (ACTIVE) - Should have full access

### 4. Test System Admin Access
Login as `superadmin@example.com` to verify:
- Can view all tenants
- Can switch between any tenant
- Has full administrative privileges

### 5. Test Authentication Flow
Use any user to test:
- Login with correct credentials
- Login with wrong password (should lock after 5 attempts)
- Password reset flow
- Email verification flow

---

## Database Schema

### Related Tables

The seed data populates these tables:
- `users` - User accounts
- `user_tenants` - User-to-tenant relationships with roles
- `tenants` - Tenant records
- `companies` - Company profiles
- `subscriptions` - Billing and subscription data

### Tenant Isolation

Every user can access one or more tenants through the `user_tenants` junction table. Each access has a specific role that determines permissions within that tenant.

**Example:** The consultant user has 2 records in `user_tenants`:
1. `user_id: consultant_id, tenant_id: tenant1_id, role: STAFF`
2. `user_id: consultant_id, tenant_id: tenant2_id, role: STAFF`

---

## Security Notes

### Password Hashing
- All passwords are hashed using **Argon2id**
- Configuration from `.env`:
  - Memory: 64 MB
  - Iterations: 3
  - Parallelism: 4
  - Salt Length: 16 bytes
  - Key Length: 32 bytes

### Brute Force Protection
The system implements 4-tier lockout:
- Tier 1: 3 attempts → 5 min lockout
- Tier 2: 5 attempts → 15 min lockout
- Tier 3: 10 attempts → 1 hour lockout
- Tier 4: 15 attempts → 24 hour lockout

### Testing Brute Force Protection
1. Login with wrong password 3 times
2. Verify account is locked for 5 minutes
3. Wait for lockout to expire
4. Test progressive lockout tiers

---

## Cleaning Up Test Data

To remove all seeded data and start fresh:

```bash
# Reset database (WARNING: Destructive!)
make migrate-reset

# Or manually delete records
DELETE FROM user_tenants;
DELETE FROM users;
DELETE FROM tenants;
DELETE FROM subscriptions;
DELETE FROM companies;
```

---

## Extending Seed Data

To add more test data:

1. Edit `cmd/seed/main.go`
2. Add new user/company/tenant records in `seedTestData()` function
3. Follow existing patterns for data creation
4. Update this documentation

### Example: Adding a New User

```go
newUser := &models.User{
    Email:    "newuser@example.com",
    Username: "newuser",
    Password: hashedPassword,
    Name:     "New User Name",
    IsActive: true,
}
db.Create(newUser)

// Link to tenant
userTenant := &models.UserTenant{
    UserID:   newUser.ID,
    TenantID: tenant1.ID,
    Role:     models.UserRoleSales,
    IsActive: true,
}
db.Create(userTenant)
```

---

## Troubleshooting

### Error: "User already exists"
The seed script doesn't check for duplicates. To re-seed:
1. Drop and recreate database
2. Run migrations
3. Run seed script

### Error: "Failed to hash password"
Check Argon2 configuration in `.env` file:
```env
ARGON2_MEMORY=65536
ARGON2_ITERATIONS=3
ARGON2_PARALLELISM=4
ARGON2_SALT_LENGTH=16
ARGON2_KEY_LENGTH=32
```

### Error: "Database connection failed"
Verify `DATABASE_URL` in `.env`:
```env
DATABASE_URL=postgresql://user:password@localhost:5432/erp_db?sslmode=disable
```

---

## API Testing with Seed Users

### Login Request Example

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "owner.maju@example.com",
    "password": "Password123!"
  }'
```

### Expected Response

```json
{
  "user": {
    "id": "clh123...",
    "email": "owner.maju@example.com",
    "name": "Budi Santoso",
    "tenants": [
      {
        "id": "clh456...",
        "name": "PT Maju Jaya Distribusi",
        "role": "OWNER"
      }
    ]
  },
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "clh789..."
}
```

---

## Summary

This seed data provides a comprehensive testing environment with:
- ✅ Multiple user roles per tenant
- ✅ Multi-tenant user scenario
- ✅ System administrator account
- ✅ Trial and active subscription states
- ✅ Realistic Indonesian company data
- ✅ Secure password hashing with Argon2id

Use these credentials to test authentication, authorization, multi-tenancy, and role-based access control features.
