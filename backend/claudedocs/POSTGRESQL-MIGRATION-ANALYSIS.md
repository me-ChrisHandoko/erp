# PostgreSQL Migration Analysis & SQLite Cleanup Recommendations

**Analysis Date:** 2025-12-17
**Scope:** Complete codebase analysis for PostgreSQL usage and SQLite removal
**Analysis Type:** Ultra-comprehensive with security and performance considerations

---

## Executive Summary

### Current Database Status

✅ **Schema:** PostgreSQL configured in `schema.prisma` (line 10)
✅ **Environment:** PostgreSQL configured in `.env.example` (lines 12-13)
⚠️ **Default Config:** Still defaults to SQLite in `internal/config/env.go` (line 25)
⚠️ **Production Code:** Dual database support (PostgreSQL + SQLite) still present
✅ **Test Suite:** Properly uses in-memory SQLite for fast, isolated testing

### Migration Status

**Schema Migration:** ✅ COMPLETE
**Environment Configuration:** ✅ COMPLETE
**Production Code Cleanup:** ❌ PENDING
**Documentation Updates:** ❌ PENDING
**Dependency Cleanup:** ❌ PENDING

---

## Detailed Findings

### 1. Configuration Files Analysis

#### ✅ schema.prisma (Correctly Configured)
```prisma
datasource db {
  provider = "postgresql" // ✅ Changed from sqlite
  url      = env("DATABASE_URL")
}
```
**Status:** ✅ No changes needed
**Location:** `/schema.prisma:10`

#### ✅ .env.example (Correctly Configured)
```bash
# Lines 12-13: PostgreSQL primary configuration
DATABASE_DRIVER=postgres
DATABASE_DSN=postgres://postgres:password@localhost:5432/erp_db?sslmode=disable

# Lines 14-16: SQLite commented out (development fallback)
# For SQLite (development):
# DATABASE_DRIVER=sqlite
# DATABASE_DSN=./data/erp.db
```
**Status:** ⚠️ Remove commented SQLite references
**Location:** `/.env.example:14-16`

#### ❌ internal/config/env.go (Incorrect Default)
```go
// Line 25: Default still points to SQLite
Database: DatabaseConfig{
    URL: getEnv("DATABASE_URL", "file:./erp.db"), // ❌ SQLite default
    MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
    MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
    ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
},
```
**Status:** ❌ MUST CHANGE - Critical security/production issue
**Location:** `/internal/config/env.go:25`

**Recommendation:**
```go
URL: getEnv("DATABASE_URL", ""), // ❌ Remove default or use PostgreSQL
```

**Rationale:**
1. **Production Safety:** No default prevents accidental SQLite usage in production
2. **Explicit Configuration:** Forces explicit DATABASE_URL setting
3. **Fail-Fast:** Application fails early with clear error if DATABASE_URL not set

**Alternative (PostgreSQL default for local development):**
```go
URL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/erp_db?sslmode=disable"),
```

---

### 2. Production Code Analysis

#### ❌ Dual Database Support Detection

Two files contain duplicate database detection logic:

**File 1: `internal/config/database.go`**
```go
// Lines 7-8: SQLite import
import (
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite" // ❌ Remove
    "gorm.io/gorm"
)

// Lines 31-36: SQLite detection and connection
if isPostgresURL(cfg.URL) {
    db, err = gorm.Open(postgres.Open(cfg.URL), gormConfig)
} else if isSQLiteURL(cfg.URL) { // ❌ Remove this branch
    db, err = gorm.Open(sqlite.Open(cfg.URL), gormConfig)
} else {
    return nil, fmt.Errorf("unsupported database URL format: %s", cfg.URL)
}

// Lines 79-81: SQLite URL detection function
func isSQLiteURL(url string) bool { // ❌ Remove entire function
    return len(url) > 5 && url[:5] == "file:"
}
```

**File 2: `internal/database/database.go`**
```go
// Lines 8-9: SQLite import
import (
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite" // ❌ Remove
    "gorm.io/gorm"
)

// Lines 22-28: SQLite detection and connection
if isPostgresURL(cfg.Database.URL) {
    dialector = postgres.Open(cfg.Database.URL)
} else if isSQLiteURL(cfg.Database.URL) { // ❌ Remove this branch
    dialector = sqlite.Open(cfg.Database.URL)
} else {
    return nil, fmt.Errorf("unsupported database URL format: %s", cfg.Database.URL)
}

// Lines 90-92: SQLite URL detection function
func isSQLiteURL(url string) bool { // ❌ Remove entire function
    return len(url) > 5 && url[:5] == "file:"
}
```

**Issues Identified:**
1. **Code Duplication:** Same logic in two different files
2. **Maintenance Burden:** Changes must be made in multiple places
3. **SQLite Production Risk:** SQLite can still be used if DATABASE_URL has `file:` prefix

---

### 3. Test Files Analysis (✅ KEEP AS-IS)

#### In-Memory SQLite for Testing (Best Practice)

**Test files using in-memory SQLite:**
1. `models/phase2_test.go`
2. `models/phase3_test.go`
3. `models/phase4_test.go`
4. `models/models_test.go`
5. `internal/database/tenant_test.go`
6. `internal/middleware/auth_test.go`
7. `internal/jobs/scheduler_test.go`
8. `internal/service/auth/brute_force_test.go`

**Pattern (Correct):**
```go
import "gorm.io/driver/sqlite"

func setupTestDB() (*gorm.DB, error) {
    database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    // Auto-migrate test schema
    // ...
    return database, nil
}
```

**Recommendation:** ✅ **KEEP** - This is best practice for:
- **Speed:** In-memory databases are extremely fast
- **Isolation:** Each test gets fresh database, no cross-test pollution
- **CI/CD:** No external database dependency for tests
- **Cost:** No need for test PostgreSQL instance

**Reference:** Go testing best practices recommend in-memory databases for unit tests

---

### 4. Script Analysis

#### scripts/migrate.sh

**Current Implementation:**
```bash
# Line 10: Default DATABASE_URL
DB_URL="${DATABASE_URL:-file:./erp.db}" # ⚠️ SQLite default

# Line 33: golang-migrate installation hint
print_message "$YELLOW" "Install it with: go install -tags 'postgres sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
```

**Recommendation:** ⚠️ **UPDATE**
```bash
# Remove SQLite default, require explicit DATABASE_URL
DB_URL="${DATABASE_URL:?DATABASE_URL environment variable is required}"

# Update installation hint (remove sqlite3 tag)
print_message "$YELLOW" "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
```

#### scripts/seed.sh

**Current Implementation:**
```bash
# Line 10: Default DATABASE_URL
DB_URL="${DATABASE_URL:-file:./erp.db}" # ⚠️ SQLite default

# Lines 26-35: Database type detection
get_db_type() {
    if [[ "$DB_URL" == postgres* ]] || [[ "$DB_URL" == postgresql* ]]; then
        echo "postgres"
    elif [[ "$DB_URL" == file:* ]]; then # ⚠️ SQLite detection
        echo "sqlite"
    else
        echo "unknown"
    fi
}

# Lines 49-59: Database-specific SQL execution
if [ "$db_type" == "postgres" ]; then
    psql "$DB_URL" < "$sql_file"
elif [ "$db_type" == "sqlite" ]; then # ⚠️ SQLite branch
    db_file=${DB_URL#file:}
    sqlite3 "$db_file" < "$sql_file"
else
    print_message "$RED" "Error: Unsupported database type"
    exit 1
fi
```

**Recommendation:** ⚠️ **SIMPLIFY**
```bash
# Remove SQLite default
DB_URL="${DATABASE_URL:?DATABASE_URL environment variable is required}"

# Simplify to PostgreSQL-only
execute_sql() {
    local sql_file=$1

    if [ ! -f "$sql_file" ]; then
        print_message "$YELLOW" "Warning: Seed file not found: $sql_file"
        return
    fi

    print_message "$BLUE" "  Executing: $sql_file"
    psql "$DB_URL" < "$sql_file"
}
```

---

### 5. Dependency Analysis

#### go.mod Dependencies

**Current SQLite Dependencies:**
```go
// Line 18: Direct SQLite driver dependency
gorm.io/driver/sqlite v1.6.0

// Line 47: Indirect SQLite driver dependency
github.com/mattn/go-sqlite3 v1.14.22 // indirect
```

#### go.sum Checksums
```
github.com/mattn/go-sqlite3 v1.14.22 h1:2gZY6PC6kBnID23Tichd1K+Z0oS6nE/XwU+Vz/5o4kU=
github.com/mattn/go-sqlite3 v1.14.22/go.mod h1:Uh1q+B4BYcTPb+yiD3kU8Ct7aC0hY9fxUwlHK0RXw+Y=
gorm.io/driver/sqlite v1.6.0 h1:WHRRrIiulaPiPFmDcod6prc4l2VGVWHz80KspNsxSfQ=
gorm.io/driver/sqlite v1.6.0/go.mod h1:AO9V1qIQddBESngQUKWL9yoH93HIeA1X6V633rBwyT8=
```

**Analysis:**
- ✅ Keep for test files (in-memory SQLite testing)
- ❌ Remove from production code imports
- ⚠️ Keep in `go.mod` as test dependency

**Recommendation:**
1. Remove SQLite imports from production code
2. Run `go mod tidy` to clean up unused dependencies
3. SQLite driver will remain as test-only dependency (correct)

**After cleanup, go.mod should automatically move to:**
```go
require (
    // ... other dependencies
    gorm.io/driver/postgres v1.6.0
)

require (
    // ... other test dependencies
    gorm.io/driver/sqlite v1.6.0 // indirect - test only
    github.com/mattn/go-sqlite3 v1.14.22 // indirect - test only
)
```

---

### 6. Documentation References

#### Files Mentioning SQLite (42 occurrences across 20+ files)

**Documentation files with SQLite references:**
1. `CLAUDE.md` - Line 12: "PostgreSQL/SQLite database"
2. `README.md` - Lines 21, 28, 86, 224, 226
3. `README_PHASE1.md` - Lines 25, 31, 36, 365, 432
4. `README_PHASE2.md` - Lines 21, 27, 32, 464
5. `docs/ARCHITECTURE_DIAGRAM.md` - Lines 140, 573, 576
6. `claudedocs/*.md` - Multiple references (15+ files)

**Recommendation:** ⚠️ **GLOBAL FIND-REPLACE**
- Replace "PostgreSQL/SQLite" → "PostgreSQL"
- Remove SQLite-specific sections
- Update architecture diagrams
- Update setup instructions

---

## Implementation Roadmap

### Phase 1: Critical Production Code Changes (Priority: 🔴 HIGH)

**Estimated Time:** 2-3 hours
**Risk Level:** Medium
**Testing Required:** ✅ Full regression testing

#### Step 1.1: Update Default Database URL
**File:** `internal/config/env.go`
**Line:** 25

**Current:**
```go
URL: getEnv("DATABASE_URL", "file:./erp.db"),
```

**Change to:**
```go
URL: getEnv("DATABASE_URL", ""),
```

**Validation:**
```go
// In config.go Validate() method (line 179)
if c.Database.URL == "" {
    return fmt.Errorf("database URL is required")
}
```

**Impact:**
- ✅ Prevents accidental SQLite usage
- ✅ Forces explicit configuration
- ⚠️ Breaks local development if DATABASE_URL not set
- ✅ Better production safety

#### Step 1.2: Remove SQLite Support from Production Code

**File 1:** `internal/config/database.go`

**Changes:**
```diff
 import (
     "gorm.io/driver/postgres"
-    "gorm.io/driver/sqlite"
     "gorm.io/gorm"
 )

 func NewDatabase(cfg DatabaseConfig, debug bool) (*gorm.DB, error) {
     // ...
     var db *gorm.DB
     var err error

-    if isPostgresURL(cfg.URL) {
-        db, err = gorm.Open(postgres.Open(cfg.URL), gormConfig)
-    } else if isSQLiteURL(cfg.URL) {
-        db, err = gorm.Open(sqlite.Open(cfg.URL), gormConfig)
-    } else {
-        return nil, fmt.Errorf("unsupported database URL format: %s", cfg.URL)
-    }
+    // Only PostgreSQL supported for production
+    db, err = gorm.Open(postgres.Open(cfg.URL), gormConfig)

     if err != nil {
         return nil, fmt.Errorf("failed to connect to database: %w", err)
     }
     // ...
 }

-func isPostgresURL(url string) bool {
-    return len(url) > 10 && (url[:10] == "postgres://" || url[:14] == "postgresql://")
-}
-
-func isSQLiteURL(url string) bool {
-    return len(url) > 5 && url[:5] == "file:"
-}
```

**File 2:** `internal/database/database.go`

**Changes:**
```diff
 import (
     "gorm.io/driver/postgres"
-    "gorm.io/driver/sqlite"
     "gorm.io/gorm"
 )

 func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
-    var dialector gorm.Dialector
-
-    if isPostgresURL(cfg.Database.URL) {
-        dialector = postgres.Open(cfg.Database.URL)
-    } else if isSQLiteURL(cfg.Database.URL) {
-        dialector = sqlite.Open(cfg.Database.URL)
-    } else {
-        return nil, fmt.Errorf("unsupported database URL format: %s", cfg.Database.URL)
-    }
+    // Only PostgreSQL supported
+    dialector := postgres.Open(cfg.Database.URL)

     // ...
 }

-func isPostgresURL(url string) bool {
-    return len(url) > 10 && (url[:10] == "postgres://" || url[:14] == "postgresql://")
-}
-
-func isSQLiteURL(url string) bool {
-    return len(url) > 5 && url[:5] == "file:"
-}
```

**Testing:**
```bash
# After changes, verify:
go build ./...
go test ./... -v
```

#### Step 1.3: Update .env.example

**File:** `.env.example`

**Changes:**
```diff
 # Database Configuration
 DATABASE_DRIVER=postgres
 DATABASE_DSN=postgres://postgres:password@localhost:5432/erp_db?sslmode=disable
-# For SQLite (development):
-# DATABASE_DRIVER=sqlite
-# DATABASE_DSN=./data/erp.db

 DATABASE_MAX_IDLE_CONNS=10
```

**Commit Message:**
```
refactor(db): remove SQLite support from production code

- Remove SQLite driver imports from production code
- Enforce PostgreSQL-only in database connection logic
- Update default DATABASE_URL to empty (require explicit config)
- Remove commented SQLite references from .env.example

BREAKING CHANGE: SQLite no longer supported for production.
Use PostgreSQL for all deployments.
Test suite still uses in-memory SQLite (unchanged).

Refs: POSTGRESQL-MIGRATION-ANALYSIS.md
```

---

### Phase 2: Script Updates (Priority: 🟡 MEDIUM)

**Estimated Time:** 1 hour
**Risk Level:** Low
**Testing Required:** ✅ Manual script testing

#### Step 2.1: Update migrate.sh

**File:** `scripts/migrate.sh`

**Changes:**
```diff
 # Configuration
 MIGRATIONS_DIR="db/migrations"
-DB_URL="${DATABASE_URL:-file:./erp.db}"
+DB_URL="${DATABASE_URL:?DATABASE_URL environment variable is required}"

 check_migrate_installed() {
     if ! command -v migrate &> /dev/null; then
         print_message "$RED" "Error: golang-migrate is not installed"
-        print_message "$YELLOW" "Install it with: go install -tags 'postgres sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
+        print_message "$YELLOW" "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
         exit 1
     fi
 }
```

**Testing:**
```bash
# Test script without DATABASE_URL (should fail with clear error)
./scripts/migrate.sh up

# Test with DATABASE_URL
export DATABASE_URL="postgres://postgres:password@localhost:5432/erp_db?sslmode=disable"
./scripts/migrate.sh up
```

#### Step 2.2: Update seed.sh

**File:** `scripts/seed.sh`

**Changes:**
```diff
 # Configuration
 SEEDS_DIR="db/seeds"
-DB_URL="${DATABASE_URL:-file:./erp.db}"
+DB_URL="${DATABASE_URL:?DATABASE_URL environment variable is required}"

-# Detect database type from URL
-get_db_type() {
-    if [[ "$DB_URL" == postgres* ]] || [[ "$DB_URL" == postgresql* ]]; then
-        echo "postgres"
-    elif [[ "$DB_URL" == file:* ]]; then
-        echo "sqlite"
-    else
-        echo "unknown"
-    fi
-}
-
 # Execute SQL file based on database type
 execute_sql() {
     local sql_file=$1
-    local db_type=$(get_db_type)

     if [ ! -f "$sql_file" ]; then
         print_message "$YELLOW" "Warning: Seed file not found: $sql_file"
         return
     fi

     print_message "$BLUE" "  Executing: $sql_file"
-
-    if [ "$db_type" == "postgres" ]; then
-        # PostgreSQL
-        psql "$DB_URL" < "$sql_file"
-    elif [ "$db_type" == "sqlite" ]; then
-        # SQLite
-        db_file=${DB_URL#file:}
-        sqlite3 "$db_file" < "$sql_file"
-    else
-        print_message "$RED" "Error: Unsupported database type"
-        exit 1
-    fi
+    psql "$DB_URL" < "$sql_file"
 }
```

**Commit Message:**
```
refactor(scripts): remove SQLite support from migration/seed scripts

- Require explicit DATABASE_URL (remove SQLite default)
- Simplify database type detection (PostgreSQL only)
- Update golang-migrate installation instructions
- Remove sqlite3 tool dependency

Refs: POSTGRESQL-MIGRATION-ANALYSIS.md
```

---

### Phase 3: Documentation Updates (Priority: 🟢 LOW)

**Estimated Time:** 2-3 hours
**Risk Level:** None
**Testing Required:** ❌ Review only

#### Step 3.1: Global Documentation Updates

**Strategy:** Find and replace across all documentation files

**Pattern Replacements:**
1. `PostgreSQL/SQLite` → `PostgreSQL`
2. `PostgreSQL or SQLite` → `PostgreSQL`
3. `SQLite (for development)` → *(remove)*
4. `SQLite (development)` → *(remove)*

**Files to Update:**
```bash
# Run global find-replace
find . -name "*.md" -type f -exec sed -i '' 's/PostgreSQL\/SQLite/PostgreSQL/g' {} +
find . -name "*.md" -type f -exec sed -i '' 's/PostgreSQL or SQLite/PostgreSQL/g' {} +
```

#### Step 3.2: Specific File Updates

**CLAUDE.md**
```diff
 ## Project Overview

 This is a **Go-based Multi-Tenant ERP System** for Indonesian food distribution (Distribusi Sembako), implementing a SaaS subscription model with comprehensive warehouse, inventory, and financial management capabilities.

 **Key Technologies:**
 - Go 1.25.4
 - Prisma ORM (schema-first approach)
-- PostgreSQL/SQLite database
+- PostgreSQL database
 - Multi-tenant architecture with tenant isolation
```

**README.md**
```diff
 ## Prerequisites

-- PostgreSQL 14+ or SQLite (for development)
+- PostgreSQL 14+
 - Go 1.25.4+
 - Node.js 18+ (for Prisma)

-## Database Setup
+## PostgreSQL Setup

-**SQLite (Development):**
-```bash
-DATABASE_URL=file:./erp.db
-```
-
-**PostgreSQL (Production):**
+**PostgreSQL:**
 ```bash
 DATABASE_URL=postgres://postgres:password@localhost:5432/erp_db?sslmode=disable
 ```
```

**Commit Message:**
```
docs: remove SQLite references from documentation

- Update all documentation to reflect PostgreSQL-only setup
- Remove SQLite setup instructions
- Update architecture diagrams
- Clarify test suite still uses in-memory SQLite

Refs: POSTGRESQL-MIGRATION-ANALYSIS.md
```

---

### Phase 4: Dependency Cleanup (Priority: 🟢 LOW)

**Estimated Time:** 15 minutes
**Risk Level:** None
**Testing Required:** ✅ Verify tests still pass

#### Step 4.1: Run go mod tidy

**Commands:**
```bash
# Remove SQLite imports from production code first (Phase 1)
# Then clean up dependencies
go mod tidy

# Verify test suite still works with SQLite
go test ./... -v
```

**Expected Result:**
```go
// go.mod after cleanup
require (
    // ... other production dependencies
    gorm.io/driver/postgres v1.6.0
)

require (
    // ... other indirect dependencies
    gorm.io/driver/sqlite v1.6.0 // indirect (test only)
    github.com/mattn/go-sqlite3 v1.14.22 // indirect (test only)
)
```

**Verification:**
```bash
# Check dependency graph
go mod graph | grep sqlite

# Should only show test file dependencies:
# backend gorm.io/driver/sqlite@v1.6.0
# (referenced by *_test.go files only)
```

**Commit Message:**
```
chore(deps): clean up SQLite dependencies

- Run go mod tidy after removing production SQLite imports
- SQLite driver now test-only dependency (in-memory testing)
- Verify all tests pass with in-memory SQLite

Refs: POSTGRESQL-MIGRATION-ANALYSIS.md
```

---

## Verification & Testing Checklist

### Pre-Deployment Verification

#### ✅ Configuration Validation
```bash
# 1. Verify DATABASE_URL is required
unset DATABASE_URL
go run cmd/server/main.go
# Expected: Error - "database URL is required"

# 2. Verify PostgreSQL connection works
export DATABASE_URL="postgres://postgres:password@localhost:5432/erp_db?sslmode=disable"
go run cmd/server/main.go
# Expected: Server starts successfully

# 3. Verify SQLite URLs are rejected
export DATABASE_URL="file:./erp.db"
go run cmd/server/main.go
# Expected: Connection error (no SQLite driver)
```

#### ✅ Test Suite Validation
```bash
# Run full test suite (should still use in-memory SQLite)
go test ./... -v -race -coverprofile=coverage.out

# Verify test database setup
grep -r "sqlite.Open" models/*test.go internal/**/*test.go
# Expected: All test files show ":memory:" usage

# Check test coverage
go tool cover -html=coverage.out
```

#### ✅ Migration Script Validation
```bash
# 1. Test migrate.sh without DATABASE_URL
unset DATABASE_URL
./scripts/migrate.sh up
# Expected: Error - "DATABASE_URL environment variable is required"

# 2. Test with PostgreSQL URL
export DATABASE_URL="postgres://postgres:password@localhost:5432/erp_test?sslmode=disable"
./scripts/migrate.sh version
# Expected: Shows current migration version

# 3. Test seed.sh
./scripts/seed.sh dev
# Expected: Seeds development data to PostgreSQL
```

#### ✅ Dependency Validation
```bash
# Check production code doesn't import SQLite
grep -r "gorm.io/driver/sqlite" --include="*.go" --exclude="*_test.go" .
# Expected: No matches (only in test files)

# Verify go.mod is clean
go mod verify
go mod tidy
git diff go.mod go.sum
# Expected: No changes after tidy
```

### Post-Deployment Monitoring

#### Production Health Checks
```bash
# 1. Database connection pool metrics
curl http://localhost:8080/health/db
# Expected: PostgreSQL connection healthy

# 2. Check logs for SQLite references
tail -f /var/log/erp/app.log | grep -i sqlite
# Expected: No SQLite mentions

# 3. Verify tenant isolation works
# (Check that all queries include tenantId filter)
```

---

## Risk Assessment & Mitigation

### Risk Matrix

| Risk | Severity | Probability | Impact | Mitigation |
|------|----------|-------------|---------|------------|
| **Breaking local dev** | 🟡 Medium | High | Medium | Update developer documentation, provide .env template |
| **Test failures** | 🟢 Low | Low | Medium | In-memory SQLite unchanged, full test suite validation |
| **Production errors** | 🔴 High | Low | High | Staged rollout, database connection validation at startup |
| **Migration failures** | 🟡 Medium | Medium | High | Test migrations on staging, backup production DB |
| **Documentation outdated** | 🟢 Low | Medium | Low | Global find-replace, peer review |

### Mitigation Strategies

#### 1. Breaking Local Development
**Risk:** Developers using SQLite locally will see errors after changes

**Mitigation:**
1. Update `README.md` with PostgreSQL setup instructions
2. Provide Docker Compose for local PostgreSQL:
   ```yaml
   # docker-compose.yml
   version: '3.8'
   services:
     postgres:
       image: postgres:14-alpine
       environment:
         POSTGRES_DB: erp_db
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: password
       ports:
         - "5432:5432"
       volumes:
         - postgres_data:/var/lib/postgresql/data

   volumes:
     postgres_data:
   ```
3. Add quick-start guide:
   ```bash
   # Quick start with Docker
   docker-compose up -d
   export DATABASE_URL="postgres://postgres:password@localhost:5432/erp_db?sslmode=disable"
   go run cmd/server/main.go
   ```

#### 2. Production Rollout
**Risk:** PostgreSQL connection issues in production

**Mitigation:**
1. **Staged Rollout:**
   - Deploy to dev environment first
   - Monitor for 24 hours
   - Deploy to staging
   - Monitor for 48 hours
   - Deploy to production
2. **Database Validation:**
   - Add startup connection test with detailed error logging
   - Implement retry logic with exponential backoff
   - Alert on connection failures
3. **Rollback Plan:**
   - Keep previous version deployable
   - Database backup before deployment
   - Quick rollback procedure documented

#### 3. Migration Safety
**Risk:** Data loss during schema migrations

**Mitigation:**
1. **Pre-Migration Checklist:**
   - ✅ Full database backup
   - ✅ Test migration on copy of production data
   - ✅ Verify rollback scripts work
   - ✅ Downtime window scheduled (if needed)
2. **Migration Process:**
   ```bash
   # 1. Backup production database
   pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

   # 2. Test migration on staging
   ./scripts/migrate.sh up

   # 3. Verify data integrity
   # (Custom SQL queries to check critical tables)

   # 4. Run production migration
   # (with monitoring)
   ```

---

## Performance Considerations

### PostgreSQL vs SQLite Performance

| Metric | SQLite (Dev) | PostgreSQL (Prod) | Impact |
|--------|--------------|-------------------|---------|
| **Concurrent Writes** | Poor (locks entire DB) | Excellent (MVCC) | ✅ Much better |
| **Connection Pool** | Not applicable | Required | ⚠️ Must configure |
| **Query Performance** | Good (simple queries) | Better (complex queries) | ✅ Improvement |
| **Transaction Isolation** | Limited | Full ACID | ✅ Better data integrity |
| **Full-Text Search** | Basic | Advanced (tsvector) | ✅ Feature upgrade |

### Optimization Recommendations

#### 1. Connection Pool Tuning
**Current Configuration (.env.example):**
```bash
DATABASE_MAX_IDLE_CONNS=10
DATABASE_MAX_OPEN_CONNS=100
DATABASE_CONN_MAX_LIFETIME=1h
```

**Production Recommendations:**
```bash
# For high-load production (adjust based on load testing)
DATABASE_MAX_IDLE_CONNS=25
DATABASE_MAX_OPEN_CONNS=200
DATABASE_CONN_MAX_LIFETIME=30m

# For development
DATABASE_MAX_IDLE_CONNS=5
DATABASE_MAX_OPEN_CONNS=25
DATABASE_CONN_MAX_LIFETIME=1h
```

**Tuning Formula:**
```
max_open_conns = (CPU cores × 2) + effective_spindle_count
max_idle_conns = max_open_conns / 4

For 8-core server with SSD:
max_open_conns = (8 × 2) + 1 = 17-25 (rounded to 25-50 for headroom)
max_idle_conns = 25 / 4 = 6-12
```

#### 2. Index Optimization
**Critical Indexes (already in schema.prisma):**
```prisma
// Tenant isolation - CRITICAL for performance
@@index([tenantId])
@@unique([tenantId, code])

// Date-based queries
@@index([invoiceDate])
@@index([deliveryDate])
@@index([createdAt])

// Status queries
@@index([status])
@@index([paymentStatus])

// Outstanding tracking
@@index([currentOutstanding])
```

**Additional Recommended Indexes:**
```sql
-- Composite indexes for common query patterns
CREATE INDEX idx_invoices_tenant_status_date
ON invoices(tenant_id, payment_status, invoice_date DESC);

CREATE INDEX idx_deliveries_tenant_status_date
ON deliveries(tenant_id, status, delivery_date DESC);

CREATE INDEX idx_products_tenant_active
ON products(tenant_id, is_active)
WHERE is_active = true;

-- Partial indexes for frequently queried subsets
CREATE INDEX idx_invoices_unpaid
ON invoices(tenant_id, due_date)
WHERE payment_status = 'UNPAID';

CREATE INDEX idx_batches_expiring
ON product_batches(expiry_date)
WHERE status = 'AVAILABLE' AND expiry_date IS NOT NULL;
```

#### 3. Query Optimization
**GORM Auto-Preload vs Explicit Joins:**
```go
// ❌ N+1 query problem
invoices, _ := db.Where("tenant_id = ?", tenantID).Find(&invoices)
for _, inv := range invoices {
    // Triggers separate query for each invoice
    db.Model(&inv).Association("Items").Find(&inv.Items)
}

// ✅ Preload with single query
invoices, _ := db.Where("tenant_id = ?", tenantID).
    Preload("Items").
    Preload("Customer").
    Find(&invoices)
```

---

## Security Implications

### Tenant Isolation Enhancement

**Current Implementation (Callback-based):**
```go
// internal/database/tenant.go
func RegisterTenantCallbacks(db *gorm.DB, config *TenantIsolationConfig) {
    db.Callback().Query().Before("gorm:query").Register("tenant:before_query", func(db *gorm.DB) {
        tenantID := getTenantIDFromContext(db.Statement.Context)
        if tenantID == "" && config.StrictMode {
            db.AddError(errors.New("SECURITY: Query without tenant context"))
        }
        // Auto-inject tenant filter
        db.Where("tenant_id = ?", tenantID)
    })
}
```

**PostgreSQL Row-Level Security (RLS) Alternative:**
```sql
-- Enable RLS on all tenant-scoped tables
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;

-- Create policy for tenant isolation
CREATE POLICY tenant_isolation_policy ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id')::text);

-- Set tenant context per session
SET app.current_tenant_id = 'tenant_123';
```

**Comparison:**

| Feature | GORM Callbacks | PostgreSQL RLS |
|---------|----------------|----------------|
| **Database-Level Enforcement** | ❌ No | ✅ Yes |
| **Bypass Protection** | ⚠️ Can bypass with raw SQL | ✅ Cannot bypass |
| **Performance** | ✅ Minimal overhead | ⚠️ Slight overhead |
| **Complexity** | ✅ Simple Go code | ⚠️ DB-level configuration |
| **Debugging** | ✅ Easy to debug | ⚠️ Harder to troubleshoot |

**Recommendation:** Keep GORM callbacks for now (simpler, working solution)
**Future Enhancement:** Consider RLS for defense-in-depth in Phase 2

### SQL Injection Protection

**Current Protection (GORM parameterized queries):**
```go
// ✅ Safe - parameterized
db.Where("email = ?", userInput).First(&user)

// ❌ Unsafe - string concatenation
db.Where("email = '" + userInput + "'").First(&user)
```

**PostgreSQL-Specific Enhancements:**
1. **Prepared Statements** (already enabled in gormConfig)
2. **Connection-Level Security:**
   ```bash
   # Use least-privilege database user
   DATABASE_URL=postgres://erp_app:password@localhost:5432/erp_db?sslmode=require

   # Separate read-only user for reporting
   REPORTING_DB_URL=postgres://erp_readonly:password@localhost:5432/erp_db?sslmode=require
   ```
3. **Audit Logging:**
   ```sql
   -- Enable query logging for security audit
   ALTER DATABASE erp_db SET log_statement = 'mod';  -- Log all modifications
   ALTER DATABASE erp_db SET log_duration = on;      -- Log slow queries
   ```

---

## Rollback Procedure

### Emergency Rollback Plan

#### Scenario 1: Application Won't Start After Deployment

**Symptoms:**
- Application fails with database connection errors
- Logs show "failed to connect to database"

**Immediate Actions:**
1. **Rollback to Previous Version:**
   ```bash
   # Docker deployment
   docker pull erp-backend:previous-tag
   docker-compose up -d

   # Binary deployment
   cp /backup/erp-backend-prev /opt/erp/erp-backend
   systemctl restart erp-backend
   ```

2. **Verify Rollback:**
   ```bash
   curl http://localhost:8080/health
   tail -f /var/log/erp/app.log
   ```

3. **Incident Report:**
   - Document error messages
   - Capture logs
   - Review deployment checklist for missed steps

#### Scenario 2: Need to Restore SQLite Support Temporarily

**Git Revert:**
```bash
# Revert specific commits (Phase 1 changes)
git revert <commit-hash-phase1>

# Or reset to before migration (destructive)
git reset --hard <commit-before-migration>
git push --force origin main  # ⚠️ Only if absolutely necessary
```

**Manual Restore:**
1. **Restore SQLite imports:**
   ```diff
   import (
       "gorm.io/driver/postgres"
   +   "gorm.io/driver/sqlite"
       "gorm.io/gorm"
   )
   ```

2. **Restore database detection logic:**
   ```go
   if isPostgresURL(cfg.URL) {
       dialector = postgres.Open(cfg.URL)
   } else if isSQLiteURL(cfg.URL) {
       dialector = sqlite.Open(cfg.URL)
   }
   ```

3. **Restore default URL:**
   ```go
   URL: getEnv("DATABASE_URL", "file:./erp.db"),
   ```

4. **Run go mod tidy:**
   ```bash
   go mod tidy
   go build ./...
   ```

---

## Success Criteria

### Definition of Done

#### Phase 1 (Production Code)
- ✅ No SQLite driver imports in production code (`internal/`, `cmd/`, `pkg/`)
- ✅ No `isSQLiteURL()` function in production code
- ✅ Default `DATABASE_URL` is empty or PostgreSQL
- ✅ `.env.example` has no SQLite references
- ✅ All tests pass with in-memory SQLite
- ✅ Application starts successfully with PostgreSQL
- ✅ Application fails gracefully without `DATABASE_URL`

#### Phase 2 (Scripts)
- ✅ `migrate.sh` requires explicit `DATABASE_URL`
- ✅ `seed.sh` supports PostgreSQL only
- ✅ golang-migrate installation instructions updated
- ✅ Scripts tested on staging environment

#### Phase 3 (Documentation)
- ✅ All `PostgreSQL/SQLite` replaced with `PostgreSQL`
- ✅ SQLite setup instructions removed
- ✅ Architecture diagrams updated
- ✅ Test documentation clarifies in-memory SQLite usage
- ✅ Developer onboarding docs updated

#### Phase 4 (Dependencies)
- ✅ `go mod tidy` runs without errors
- ✅ SQLite drivers only in indirect dependencies (test-only)
- ✅ Production build doesn't include SQLite binaries
- ✅ Dependency vulnerability scan passes

### Acceptance Testing

#### Automated Tests
```bash
# Run full test suite
go test ./... -v -race -coverprofile=coverage.out

# Check coverage (should maintain >80%)
go tool cover -func=coverage.out | grep total

# Integration tests with PostgreSQL
go test ./... -tags=integration -v
```

#### Manual Testing Checklist
- [ ] Fresh database setup from schema
- [ ] User registration and login
- [ ] Tenant creation and switching
- [ ] Sales order creation and fulfillment
- [ ] Invoice generation and payment recording
- [ ] Inventory movements and batch tracking
- [ ] Multi-warehouse stock transfers
- [ ] Financial reports generation
- [ ] Subscription billing workflow

#### Performance Testing
```bash
# Load testing with PostgreSQL
hey -n 10000 -c 50 -m POST http://localhost:8080/api/v1/auth/login

# Database connection pool monitoring
psql $DATABASE_URL -c "SELECT * FROM pg_stat_activity WHERE datname = 'erp_db';"

# Slow query analysis
psql $DATABASE_URL -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

---

## Monitoring & Observability

### Key Metrics to Track

#### Database Metrics
```go
// Add to application monitoring
type DBMetrics struct {
    ConnectionPoolSize   int
    ActiveConnections    int
    IdleConnections      int
    WaitCount           int64
    WaitDuration        time.Duration
    MaxIdleTimeClosed   int64
    MaxLifetimeClosed   int64
}

// Expose via /metrics endpoint (Prometheus format)
func (m *DBMetrics) Export() string {
    return fmt.Sprintf(`
# HELP db_connections_active Active database connections
# TYPE db_connections_active gauge
db_connections_active %d

# HELP db_connections_idle Idle database connections
# TYPE db_connections_idle gauge
db_connections_idle %d

# HELP db_connections_wait_count Connection wait count
# TYPE db_connections_wait_count counter
db_connections_wait_count %d
`, m.ActiveConnections, m.IdleConnections, m.WaitCount)
}
```

#### Alerts to Configure

**Critical Alerts:**
1. **Database Connection Failures**
   - Trigger: >5 connection errors in 1 minute
   - Action: Page on-call engineer
   - Runbook: Check PostgreSQL service, connection pool settings

2. **High Connection Pool Utilization**
   - Trigger: >90% pool utilization for >5 minutes
   - Action: Alert DevOps team
   - Runbook: Review active queries, consider scaling pool

3. **Slow Queries**
   - Trigger: Query execution >5 seconds
   - Action: Log and alert
   - Runbook: Review query plan, add indexes

**Warning Alerts:**
1. **Connection Pool Exhaustion**
   - Trigger: >80% pool utilization for >10 minutes
   - Action: Notify team in Slack

2. **Database Latency**
   - Trigger: P95 latency >500ms
   - Action: Review query performance

### Logging Enhancements

**Structured Logging for Database Operations:**
```go
// Use structured logging (zap/zerolog)
logger.Info("database connection established",
    zap.String("driver", "postgres"),
    zap.String("host", dbHost),
    zap.Int("max_open_conns", cfg.MaxOpenConns),
    zap.Int("max_idle_conns", cfg.MaxIdleConns),
)

logger.Error("database query failed",
    zap.String("query", query),
    zap.Error(err),
    zap.String("tenant_id", tenantID),
    zap.Duration("duration", elapsed),
)
```

**Query Logging for Debugging:**
```go
// Enable query logging in development
if cfg.IsDevelopment() {
    db.Logger = logger.Default.LogMode(logger.Info)
} else {
    // Production: Only log errors and slow queries
    db.Logger = logger.Default.LogMode(logger.Error)
}
```

---

## Appendix A: Complete File Changes Summary

### Files to Modify (11 files)

#### Production Code (4 files)
1. `internal/config/env.go` - Change default DATABASE_URL
2. `internal/config/database.go` - Remove SQLite support
3. `internal/database/database.go` - Remove SQLite support
4. `.env.example` - Remove SQLite comments

#### Scripts (2 files)
5. `scripts/migrate.sh` - Remove SQLite default and detection
6. `scripts/seed.sh` - Remove SQLite support

#### Documentation (5 files - representative, ~20 total)
7. `CLAUDE.md` - Update database description
8. `README.md` - Remove SQLite setup instructions
9. `docs/ARCHITECTURE_DIAGRAM.md` - Update architecture
10. `claudedocs/BACKEND-IMPLEMENTATION.md` - Update implementation guide
11. `claudedocs/PHASE*-*.md` - Update phase documentation

### Files to Keep Unchanged (8+ files)

#### Test Files (Keep in-memory SQLite)
- `models/phase2_test.go`
- `models/phase3_test.go`
- `models/phase4_test.go`
- `models/models_test.go`
- `internal/database/tenant_test.go`
- `internal/middleware/auth_test.go`
- `internal/jobs/scheduler_test.go`
- `internal/service/auth/brute_force_test.go`

**Rationale:** In-memory SQLite provides:
- Fast test execution (10x faster than PostgreSQL)
- Isolated test environment (no shared state)
- No external dependencies (CI/CD friendly)
- Perfect for unit testing

---

## Appendix B: Database Migration Scripts

### Schema Migration from SQLite to PostgreSQL

**If migrating existing SQLite data to PostgreSQL:**

```bash
#!/bin/bash
# migrate_sqlite_to_postgres.sh

set -e

# Configuration
SQLITE_DB="./erp.db"
POSTGRES_URL="postgres://postgres:password@localhost:5432/erp_db?sslmode=disable"
TEMP_DUMP="sqlite_dump.sql"

echo "=== SQLite to PostgreSQL Migration ==="

# Step 1: Export SQLite schema and data
echo "1. Dumping SQLite database..."
sqlite3 "$SQLITE_DB" .dump > "$TEMP_DUMP"

# Step 2: Convert SQLite SQL to PostgreSQL SQL
echo "2. Converting SQL syntax..."
sed -i '' 's/AUTOINCREMENT/SERIAL/g' "$TEMP_DUMP"
sed -i '' 's/INTEGER PRIMARY KEY/SERIAL PRIMARY KEY/g' "$TEMP_DUMP"
sed -i '' 's/DATETIME/TIMESTAMP/g' "$TEMP_DUMP"
sed -i '' '/^PRAGMA/d' "$TEMP_DUMP"
sed -i '' '/^BEGIN TRANSACTION/d' "$TEMP_DUMP"
sed -i '' '/^COMMIT/d' "$TEMP_DUMP"

# Step 3: Create PostgreSQL database (if not exists)
echo "3. Creating PostgreSQL database..."
psql -d postgres -c "DROP DATABASE IF EXISTS erp_db;"
psql -d postgres -c "CREATE DATABASE erp_db;"

# Step 4: Import to PostgreSQL
echo "4. Importing to PostgreSQL..."
psql "$POSTGRES_URL" < "$TEMP_DUMP"

# Step 5: Verify migration
echo "5. Verifying migration..."
ROW_COUNT=$(psql "$POSTGRES_URL" -t -c "SELECT COUNT(*) FROM users;")
echo "  Users table: $ROW_COUNT rows"

# Cleanup
rm "$TEMP_DUMP"

echo "✓ Migration completed successfully!"
```

**Alternative: Using pgloader**
```bash
# Install pgloader
brew install pgloader  # macOS
apt-get install pgloader  # Ubuntu

# Create migration config
cat > migrate.load <<EOF
LOAD DATABASE
     FROM sqlite://./erp.db
     INTO postgres://postgres:password@localhost/erp_db

 WITH include drop, create tables, create indexes, reset sequences

  SET work_mem to '16MB', maintenance_work_mem to '512 MB';
EOF

# Run migration
pgloader migrate.load
```

---

## Appendix C: Developer Quick Reference

### Local Development Setup (Post-Migration)

#### 1. Install PostgreSQL
```bash
# macOS
brew install postgresql@14
brew services start postgresql@14

# Ubuntu
sudo apt-get install postgresql-14
sudo systemctl start postgresql

# Docker (recommended)
docker run -d \
  --name erp-postgres \
  -e POSTGRES_DB=erp_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:14-alpine
```

#### 2. Configure Environment
```bash
# Copy .env.example
cp .env.example .env

# Edit .env
DATABASE_URL=postgres://postgres:password@localhost:5432/erp_db?sslmode=disable
```

#### 3. Run Migrations
```bash
# Apply database schema
./scripts/migrate.sh up

# Seed development data (optional)
./scripts/seed.sh dev
```

#### 4. Start Application
```bash
go run cmd/server/main.go
```

### Common Issues & Solutions

#### Issue 1: "Database URL is required"
```
Error: database URL is required
```

**Solution:**
```bash
export DATABASE_URL="postgres://postgres:password@localhost:5432/erp_db?sslmode=disable"
```

#### Issue 2: "Connection refused"
```
Error: failed to connect to database: dial tcp [::1]:5432: connect: connection refused
```

**Solution:**
```bash
# Check if PostgreSQL is running
pg_isready

# Start PostgreSQL
brew services start postgresql@14  # macOS
sudo systemctl start postgresql    # Linux
docker start erp-postgres          # Docker
```

#### Issue 3: "Database does not exist"
```
Error: database "erp_db" does not exist
```

**Solution:**
```bash
# Create database
createdb erp_db

# Or using psql
psql -d postgres -c "CREATE DATABASE erp_db;"
```

#### Issue 4: Test failures after cleanup
```
Error: failed to open database: Binary was compiled with 'CGO_ENABLED=0'
```

**Solution:**
```bash
# Tests use in-memory SQLite, ensure CGO is enabled
CGO_ENABLED=1 go test ./...
```

---

## Conclusion

### Summary of Changes

**Total Impact:**
- **11 production files** to modify (4 Go code, 2 scripts, 5 docs)
- **8+ test files** unchanged (keep in-memory SQLite)
- **2 dependencies** to clean up (auto-moved to test-only)
- **Estimated effort:** 6-8 hours total

**Benefits:**
1. ✅ **Production Safety:** No accidental SQLite usage
2. ✅ **Performance:** Better PostgreSQL optimization
3. ✅ **Scalability:** PostgreSQL handles concurrent writes
4. ✅ **Features:** Access to PostgreSQL-specific features
5. ✅ **Simplicity:** Single database system to support
6. ✅ **Testing:** In-memory SQLite kept for fast tests

**Risks Mitigated:**
1. ✅ **Development disruption:** Docker Compose provided
2. ✅ **Test failures:** In-memory SQLite unchanged
3. ✅ **Production errors:** Validation and staged rollout
4. ✅ **Documentation lag:** Comprehensive update plan

### Next Steps

**Immediate (Phase 1):**
1. Review this analysis document with team
2. Schedule deployment window
3. Prepare staging environment
4. Execute Phase 1 changes

**Short-term (Phases 2-3):**
1. Update scripts and documentation
2. Deploy to staging for validation
3. Production deployment with monitoring

**Long-term:**
1. Investigate PostgreSQL RLS for enhanced security
2. Implement advanced PostgreSQL features (full-text search, JSONB)
3. Set up query performance monitoring
4. Database backup and disaster recovery procedures

---

**Document Version:** 1.0
**Last Updated:** 2025-12-17
**Author:** Claude Code Analysis
**Review Status:** ✅ Ready for Team Review

