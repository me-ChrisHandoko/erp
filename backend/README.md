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

### 🔄 Phase 2: Authentication & Authorization (In Progress)
- [ ] User domain models
- [ ] Tenant domain models
- [ ] Auth service
- [ ] JWT middleware
- [ ] RBAC middleware
- [ ] Auth endpoints

### 📋 Phase 3-7: Planned
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
