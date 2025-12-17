# ERP Distribusi Sembako - Backend API

Multi-Tenant ERP System for Indonesian Food Distribution (Distribusi Sembako) with SaaS subscription model.

## 🚀 Features

- **Multi-Tenancy**: Full tenant isolation with per-tenant role-based access control
- **Subscription Management**: Custom pricing, trial periods, and grace periods
- **Warehouse Management**: Multi-warehouse support with stock tracking
- **Product Management**: Batch/lot tracking with expiry dates for perishables
- **Sales Workflow**: Sales Order → Delivery → Invoice → Payment
- **Purchase Workflow**: Purchase Order → Goods Receipt → Payment
- **Inventory Control**: Stock movements, opname, inter-warehouse transfers
- **Financial Management**: Cash transactions with running balance (Buku Kas)
- **Indonesian Compliance**: NPWP validation, PPN tax calculations, Faktur Pajak
- **Background Jobs**: Automated database cleanup for expired tokens and old records

## 📋 Prerequisites

- Go 1.25.4 or higher
- PostgreSQL 14+ or SQLite (for development)
- Make (optional, for using Makefile commands)

## 🛠️ Tech Stack

- **Language**: Go 1.25.4
- **Web Framework**: Gin
- **ORM**: GORM (PostgreSQL/SQLite drivers)
- **Authentication**: JWT (golang-jwt/jwt)
- **Validation**: go-playground/validator
- **Logging**: Uber Zap
- **Configuration**: Environment variables (godotenv)
- **Background Jobs**: robfig/cron/v3 (scheduled cleanup tasks)

## 📁 Project Structure

```
backend/
├── cmd/                    # Application entry points
│   ├── api/               # REST API server
│   └── worker/            # Background jobs worker
├── internal/              # Private application code
│   ├── domain/           # Domain models (DDD)
│   ├── service/          # Business logic
│   ├── handler/          # HTTP handlers
│   ├── repository/       # Data access
│   ├── jobs/             # Background jobs (cron scheduler)
│   ├── worker/           # Background job infrastructure
│   └── config/           # Configuration management
├── pkg/                   # Public libraries
│   ├── auth/             # Authentication utilities
│   ├── validator/        # Input validation
│   ├── response/         # API response formatting
│   ├── pagination/       # Pagination helpers
│   ├── converter/        # Unit conversion
│   ├── calculator/       # Tax and calculation utilities
│   └── logger/           # Structured logging
├── db/                    # Database files
│   ├── migrations/       # SQL migration files
│   └── seeds/            # Seed data
├── api/                   # API specifications
│   └── openapi/          # OpenAPI/Swagger specs
├── test/                  # Testing infrastructure
├── scripts/               # Build and deployment scripts
└── docs/                  # Documentation
```

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd backend
```

### 2. Setup Environment

```bash
# Copy environment variables
cp .env.example .env

# Edit .env with your configuration
# At minimum, set:
# - JWT_SECRET (must be at least 32 characters)
# - DATABASE_URL (PostgreSQL or SQLite)
```

### 3. Install Dependencies

```bash
# Using Make
make deps

# Or manually
go mod download
go mod tidy
```

### 4. Run Database Migrations

```bash
# Using Make
make migrate

# Or using script directly
./scripts/migrate.sh up
```

### 5. Seed Database (Optional)

```bash
# Seed development data
make seed

# Or using script directly
./scripts/seed.sh dev
```

### 6. Run the Application

```bash
# Using Make
make run

# Or using Go directly
go run cmd/api/main.go

# Server will start on http://localhost:8080
```

## 📖 Development Commands

### Using Makefile

```bash
make help          # Show all available commands
make build         # Build the application
make run           # Run the API server
make run-worker    # Run background worker
make test          # Run all tests
make test-coverage # Run tests with coverage report
make lint          # Run linter
make fmt           # Format code
make clean         # Clean build artifacts
make docker-build  # Build Docker image
make docker-up     # Start Docker containers
```

### Database Operations

```bash
# Run migrations
make migrate

# Rollback last migration
make migrate-down

# Reset database (CAUTION: deletes all data)
make migrate-reset

# Create new migration
make migrate-create name=add_products_table

# Seed database
make seed
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e

# Generate coverage report
make test-coverage
# Opens coverage.html in browser
```

## 🔧 Configuration

### Environment Variables

All configuration is done through environment variables. See `.env.example` for available options.

**Critical Variables:**

```env
# Server
APP_PORT=8080
APP_ENV=development

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/erp_db

# JWT (MUST be at least 32 characters!)
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters-long

# Tenant
DEFAULT_SUBSCRIPTION_PRICE=300000
TRIAL_PERIOD_DAYS=14
GRACE_PERIOD_DAYS=7

# Indonesian Tax
DEFAULT_PPN_RATE=11
```

### Database Configuration

**PostgreSQL (Production):**
```env
DATABASE_URL=postgresql://username:password@host:port/database?sslmode=disable
```

**SQLite (Development):**
```env
DATABASE_URL=file:./erp.db
```

## 🤖 Background Jobs

Phase 4 implements automated database cleanup using cron-based scheduled jobs to prevent database bloat and maintain system performance.

### Implemented Cleanup Jobs

| Job | Schedule | Purpose | Retention Period |
|-----|----------|---------|------------------|
| **Refresh Tokens** | Hourly at :00 | Delete expired refresh tokens | 30 days (JWT_REFRESH_EXPIRY) |
| **Email Verifications** | Hourly at :05 | Delete expired/verified tokens | 24 hours (EMAIL_VERIFICATION_EXPIRY) |
| **Password Resets** | Hourly at :10 | Delete expired/used reset tokens | 1 hour (PASSWORD_RESET_EXPIRY) |
| **Login Attempts** | Daily at 2 AM | Delete old login attempt logs | 7 days (for audit/GDPR compliance) |

### Configuration

Background jobs are configured via environment variables using 6-field cron format (second minute hour day month weekday):

```env
# Enable/disable all cleanup jobs
JOB_ENABLE_CLEANUP=true

# Cron schedules (6-field format: second minute hour day month weekday)
JOB_REFRESH_TOKEN_CLEANUP=0 0 * * * *     # Hourly at :00
JOB_EMAIL_CLEANUP=0 5 * * * *              # Hourly at :05
JOB_PASSWORD_CLEANUP=0 10 * * * *          # Hourly at :10
JOB_LOGIN_CLEANUP=0 0 2 * * *              # Daily at 2 AM
```

**Testing Configuration:**
```env
# For testing: every 5 seconds
JOB_REFRESH_TOKEN_CLEANUP=*/5 * * * * *
```

### Monitoring

The `/ready` health check endpoint includes scheduler status:

```json
{
  "status": "ready",
  "checks": {
    "database": "healthy",
    "redis": "not_configured",
    "scheduler": "healthy"
  },
  "scheduler": {
    "status": "running",
    "last_cleanup": 1702831200,
    "hours_since_cleanup": 0.5
  }
}
```

**Status Indicators:**
- `healthy`: Scheduler running and cleanup executed recently (<2 hours)
- `degraded`: No cleanup in last 2 hours
- `unhealthy`: Scheduler not running
- `not_configured`: Background jobs disabled

### Graceful Shutdown

The application implements a proper shutdown sequence:

1. **Stop HTTP server** (30s timeout) - No new requests accepted
2. **Stop job scheduler** (60s timeout) - Running jobs complete gracefully
3. **Close database connection** - Clean resource cleanup

### Logs

Cleanup jobs produce structured logs:

```
[INFO][CLEANUP] Refresh tokens: deleted 42 rows (duration: 5.2ms)
[INFO][CLEANUP] Email verifications: deleted 15 rows (duration: 3.1ms)
[ERROR][CLEANUP] Password resets failed: connection timeout
```

**Log Levels:**
- `INFO`: Successful cleanup with row count and duration
- `ERROR`: Cleanup failures with error details
- `WARNING`: Stale scheduler (no cleanup in 2+ hours)

### Testing

Run scheduler tests:

```bash
# Run all scheduler tests
go test ./internal/jobs/... -v

# Test specific cleanup job
go test ./internal/jobs/... -v -run TestCleanupExpiredRefreshTokens

# Test with coverage
go test ./internal/jobs/... -v -coverprofile=coverage.out
```

**Test Coverage:** 90%+ (10 test cases covering all cleanup logic, edge cases, and panic recovery)

### Disabling Background Jobs

Set `JOB_ENABLE_CLEANUP=false` in `.env` to disable all cleanup jobs. Useful for:
- Development environments with limited resources
- Testing scenarios where data persistence is required
- Dedicated worker instances (if separating concerns)

## 🏗️ Architecture

### Clean Architecture Layers

```
Handler (HTTP) → Service (Business Logic) → Domain (Models) ← Repository (Data Access)
```

**Key Principles:**
- Domain layer has no dependencies
- Service orchestrates business logic
- Handler is thin, delegates to services
- Repository implements domain interfaces

### Multi-Tenancy

**Tenant Isolation Enforcement:**

1. **Middleware Layer**: Extracts `tenantID` from JWT token
2. **Service Layer**: Validates tenant access and permissions
3. **Repository Layer**: **ALWAYS** filters queries by `tenantID`

```go
// CRITICAL: Every query MUST include tenantID filter
db.Where("tenant_id = ? AND id = ?", tenantID, productID).First(&product)
```

## 🔐 Authentication Flow

1. User logs in with email/password
2. Server validates credentials
3. Server generates JWT token with claims: `userID`, `tenantID`, `role`
4. Client includes token in `Authorization: Bearer <token>` header
5. Middleware validates token and extracts tenant context
6. RBAC middleware checks permissions

## 📝 API Endpoints

### Public Endpoints (No Authentication)

```
GET  /health              # Health check
GET  /api/v1              # API welcome
POST /api/v1/auth/login   # Login (TODO)
POST /api/v1/auth/register # Register (TODO)
```

### Protected Endpoints (Authentication Required)

All protected endpoints require JWT token in `Authorization` header.

```
# User Management
GET    /api/v1/users         # List users (TODO)
POST   /api/v1/users         # Create user (TODO)
GET    /api/v1/users/:id     # Get user (TODO)
PUT    /api/v1/users/:id     # Update user (TODO)
DELETE /api/v1/users/:id     # Delete user (TODO)

# More endpoints will be added as implementation progresses
```

## 🧪 Testing

### Test Structure

```
test/
├── integration/       # Integration tests (with DB)
├── e2e/              # End-to-end tests
├── testutil/         # Test utilities
└── mocks/            # Generated mocks
```

### Running Tests

```bash
# All tests
go test -v ./...

# Specific package
go test -v ./internal/service/auth

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔍 Code Quality

### Linting

```bash
# Run linter
make lint

# Or manually
golangci-lint run
```

### Code Formatting

```bash
# Format code
make fmt

# Or manually
go fmt ./...
```

## 🐳 Docker

### Build Image

```bash
make docker-build
```

### Run with Docker Compose

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## 📚 Documentation

- [Folder Structure](docs/FOLDER_STRUCTURE.md) - Detailed folder structure explanation
- [Implementation Checklist](docs/IMPLEMENTATION_CHECKLIST.md) - Step-by-step implementation guide
- [Architecture Diagram](docs/ARCHITECTURE_DIAGRAM.md) - Visual architecture diagrams
- [CLAUDE.md](CLAUDE.md) - AI assistance guidelines

## 🚧 Implementation Status

### ✅ Phase 1: Core Infrastructure (Complete)
- [x] Folder structure
- [x] Configuration management
- [x] Logger and response utilities
- [x] Bootstrap files
- [x] Main application entry point
- [x] Initial database migrations

### ✅ Phase 2: Authentication & Authorization (Complete)
- [x] User domain models
- [x] Tenant domain models
- [x] Auth service (login, logout, refresh token)
- [x] Password reset flow (forgot/reset password)
- [x] Email verification flow
- [x] JWT middleware with tenant context
- [x] RBAC middleware with per-tenant roles
- [x] Security features (brute force protection, tier-based lockout)
- [x] CSRF protection middleware

### ✅ Phase 3: Security Hardening (Complete)
- [x] Argon2id password hashing
- [x] 4-tier exponential backoff brute force protection
- [x] Login attempt tracking
- [x] Refresh token rotation
- [x] Email verification requirement
- [x] CSRF tokens for state-changing operations

### ✅ Phase 4: Background Jobs (Complete)
- [x] Cron-based job scheduler (robfig/cron)
- [x] Automated cleanup jobs:
  - [x] Expired refresh tokens (30-day retention)
  - [x] Email verification tokens (24-hour retention)
  - [x] Password reset tokens (1-hour retention)
  - [x] Old login attempts (7-day retention)
- [x] Health check integration
- [x] Graceful shutdown
- [x] Comprehensive unit tests (90%+ coverage)

### 📋 Phase 5-7: Planned
See [Implementation Checklist](docs/IMPLEMENTATION_CHECKLIST.md) for full roadmap.

## ⚠️ Important Notes

### Critical Security Rules

1. **Tenant Isolation**: Always include `tenantID` in queries
2. **Password Security**: Never store plaintext passwords (use bcrypt/argon2)
3. **JWT Secret**: Must be at least 32 characters long
4. **Soft Deletes**: Use `isActive = false` instead of hard deletes
5. **Input Validation**: Validate all user input

### Indonesian Compliance

1. **NPWP Format**: XX.XXX.XXX.X-XXX.XXX
2. **PPN Rate**: Default 11% (as of 2025)
3. **Invoice Numbering**: `{PREFIX}/{NUMBER}/{MONTH}/{YEAR}`
4. **Faktur Pajak**: Track series and SPPKP number

### Database Best Practices

1. **Base Units**: Always convert to base units for stock calculations
2. **Batch Tracking**: Required for `isBatchTracked` products
3. **FEFO**: First Expired, First Out for perishable goods
4. **Audit Trail**: Every stock change creates `InventoryMovement` record

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit Message Convention

```
feat: Add new feature
fix: Fix bug
docs: Update documentation
refactor: Refactor code
test: Add tests
chore: Maintenance tasks
```

## 📄 License

This project is licensed under the MIT License.

## 📞 Support

For issues and questions:
- Create an issue in the repository
- Check the documentation in `docs/`
- Review the implementation checklist

## 🎯 Next Steps

1. **Implement Authentication**: Complete Phase 2 (Auth & Authorization)
2. **Add Domain Models**: Implement all 11 domain modules
3. **Build Business Workflows**: Sales, Purchase, Inventory flows
4. **Add Tests**: Achieve >80% test coverage
5. **Deploy**: Setup CI/CD and production deployment

---

**Built with ❤️ for Indonesian SME Food Distributors**
