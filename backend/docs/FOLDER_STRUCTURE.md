# Backend Folder Structure Documentation

## Overview

This document describes the comprehensive folder structure for the Go-based Multi-Tenant ERP System. The architecture follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles to ensure maintainability, testability, and scalability.

## Architecture Principles

1. **Clean Architecture**: Separation of concerns with clear dependency direction (domain → service → handler)
2. **Domain-Driven Design**: Business logic organized around domain concepts
3. **Multi-Tenancy**: Strict data isolation enforced at all layers
4. **Hexagonal Architecture**: Core business logic independent of infrastructure
5. **SOLID Principles**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion

## Complete Folder Structure

```
backend/
├── cmd/                          # Main applications (entry points)
│   ├── api/                     # REST API server
│   │   ├── main.go             # Application entry point
│   │   ├── router.go           # Route registration and middleware setup
│   │   └── bootstrap/          # Application initialization
│   │       ├── database.go     # Database connection and GORM setup
│   │       ├── logger.go       # Structured logging initialization
│   │       ├── validator.go    # Input validation configuration
│   │       └── server.go       # HTTP server configuration (timeouts, CORS, etc.)
│   └── worker/                  # Background job worker
│       ├── main.go             # Worker entry point
│       └── jobs/               # Job implementations
│           ├── subscription.go # Subscription billing automation
│           ├── expiry.go       # Batch expiry monitoring
│           ├── outstanding.go  # Outstanding amount recalculation
│           └── notification.go # Notification sending
│
├── internal/                    # Private application code (not importable by external packages)
│   ├── domain/                 # Domain models and business logic (DDD core)
│   │   ├── user/
│   │   │   ├── model.go       # User, UserTenant entities
│   │   │   ├── role.go        # UserRole enum and permission definitions
│   │   │   └── repository.go  # Repository interface (implemented in infrastructure layer)
│   │   ├── tenant/
│   │   │   ├── model.go       # Tenant, Subscription, Payment entities
│   │   │   ├── subscription.go # Subscription business rules
│   │   │   └── repository.go  # Repository interface
│   │   ├── company/
│   │   │   ├── model.go       # Company profile entity
│   │   │   ├── settings.go    # Company settings business logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── warehouse/
│   │   │   ├── model.go       # Warehouse, WarehouseStock entities
│   │   │   ├── stock.go       # Stock level management logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── product/
│   │   │   ├── model.go       # Product, ProductUnit, ProductBatch entities
│   │   │   ├── batch.go       # Batch/lot tracking business rules
│   │   │   ├── unit.go        # Multi-unit conversion logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── customer/
│   │   │   ├── model.go       # Customer entity
│   │   │   ├── outstanding.go # Outstanding amount tracking logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── supplier/
│   │   │   ├── model.go       # Supplier entity
│   │   │   ├── outstanding.go # Outstanding amount tracking logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── sales/
│   │   │   ├── model.go       # SalesOrder, Invoice entities
│   │   │   ├── order.go       # Sales order workflow logic
│   │   │   ├── invoice.go     # Invoice generation logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── purchase/
│   │   │   ├── model.go       # PurchaseOrder, GoodsReceipt entities
│   │   │   ├── order.go       # Purchase order workflow logic
│   │   │   ├── receipt.go     # Goods receipt logic
│   │   │   └── repository.go  # Repository interface
│   │   ├── delivery/
│   │   │   ├── model.go       # Delivery entity
│   │   │   ├── fulfillment.go # Delivery workflow and status transitions
│   │   │   └── repository.go  # Repository interface
│   │   ├── inventory/
│   │   │   ├── model.go       # InventoryMovement, StockOpname entities
│   │   │   ├── movement.go    # Stock movement tracking logic
│   │   │   ├── opname.go      # Physical inventory count logic
│   │   │   └── repository.go  # Repository interface
│   │   └── finance/
│   │       ├── model.go       # CashTransaction, Check entities
│   │       ├── cash.go        # Cash management logic (Buku Kas)
│   │       ├── payment.go     # Payment processing logic
│   │       └── repository.go  # Repository interface
│   │
│   ├── service/                # Business logic orchestration layer
│   │   ├── auth/
│   │   │   ├── service.go     # Authentication service (login, logout, refresh)
│   │   │   ├── jwt.go         # JWT token generation and validation
│   │   │   └── password.go    # Password hashing and verification
│   │   ├── tenant/
│   │   │   ├── service.go     # Tenant management service
│   │   │   └── subscription.go # Subscription billing and payment tracking
│   │   ├── user/
│   │   │   └── service.go     # User management service
│   │   ├── warehouse/
│   │   │   ├── service.go     # Warehouse CRUD service
│   │   │   └── stock.go       # Stock level management service
│   │   ├── product/
│   │   │   ├── service.go     # Product CRUD service
│   │   │   ├── batch.go       # Batch tracking service
│   │   │   └── pricing.go     # Multi-unit pricing calculations
│   │   ├── sales/
│   │   │   ├── order.go       # Sales order service
│   │   │   ├── invoice.go     # Invoice generation service
│   │   │   └── delivery.go    # Delivery fulfillment service
│   │   ├── purchase/
│   │   │   ├── order.go       # Purchase order service
│   │   │   └── receipt.go     # Goods receipt service
│   │   ├── inventory/
│   │   │   ├── movement.go    # Inventory movement service
│   │   │   └── opname.go      # Stock opname service
│   │   └── finance/
│   │       ├── cash.go        # Cash transaction service
│   │       └── payment.go     # Payment processing service
│   │
│   ├── handler/                # HTTP handlers (controllers)
│   │   ├── middleware/        # HTTP middleware
│   │   │   ├── auth.go       # JWT authentication middleware
│   │   │   ├── tenant.go     # Tenant context extraction middleware
│   │   │   ├── rbac.go       # Role-based access control middleware
│   │   │   ├── cors.go       # CORS configuration
│   │   │   └── logger.go     # Request/response logging
│   │   ├── auth/
│   │   │   └── handler.go    # Login, logout, refresh token endpoints
│   │   ├── tenant/
│   │   │   └── handler.go    # Tenant CRUD, subscription management
│   │   ├── user/
│   │   │   └── handler.go    # User CRUD, role assignment
│   │   ├── warehouse/
│   │   │   └── handler.go    # Warehouse CRUD, stock queries
│   │   ├── product/
│   │   │   └── handler.go    # Product CRUD, batch management
│   │   ├── customer/
│   │   │   └── handler.go    # Customer CRUD, outstanding reports
│   │   ├── supplier/
│   │   │   └── handler.go    # Supplier CRUD, outstanding reports
│   │   ├── sales/
│   │   │   ├── order.go      # Sales order endpoints
│   │   │   ├── invoice.go    # Invoice endpoints
│   │   │   └── delivery.go   # Delivery endpoints
│   │   ├── purchase/
│   │   │   ├── order.go      # Purchase order endpoints
│   │   │   └── receipt.go    # Goods receipt endpoints
│   │   ├── inventory/
│   │   │   └── handler.go    # Movement, opname endpoints
│   │   └── finance/
│   │       └── handler.go    # Cash transaction endpoints
│   │
│   ├── repository/             # Data access layer implementations
│   │   ├── gorm/              # GORM repository implementations
│   │   │   ├── user.go       # User repository implementation
│   │   │   ├── tenant.go     # Tenant repository implementation
│   │   │   ├── warehouse.go  # Warehouse repository implementation
│   │   │   ├── product.go    # Product repository implementation
│   │   │   ├── customer.go   # Customer repository implementation
│   │   │   ├── supplier.go   # Supplier repository implementation
│   │   │   ├── sales.go      # Sales repository implementation
│   │   │   ├── purchase.go   # Purchase repository implementation
│   │   │   ├── inventory.go  # Inventory repository implementation
│   │   │   └── finance.go    # Finance repository implementation
│   │   └── transaction.go     # Transaction manager interface
│   │
│   ├── worker/                 # Background job infrastructure
│   │   ├── scheduler.go       # Job scheduler (cron-based)
│   │   ├── queue.go          # Job queue interface
│   │   └── processor.go      # Job processor
│   │
│   └── config/                 # Configuration management
│       ├── config.go          # Configuration struct
│       ├── env.go            # Environment variable loading
│       ├── database.go       # Database configuration
│       └── validator.go      # Configuration validation
│
├── pkg/                        # Public libraries (reusable across projects)
│   ├── auth/
│   │   ├── jwt.go            # JWT token utilities
│   │   ├── hasher.go         # Password hashing (bcrypt/argon2)
│   │   └── rbac.go           # RBAC permission checker
│   ├── validator/
│   │   ├── validator.go      # Input validation wrapper (go-playground/validator)
│   │   ├── npwp.go           # Indonesian NPWP validation
│   │   └── custom.go         # Custom validation rules
│   ├── response/
│   │   ├── json.go           # Standardized JSON response formatting
│   │   └── error.go          # Error response formatting
│   ├── pagination/
│   │   ├── paginator.go      # Offset-based pagination utilities
│   │   └── cursor.go         # Cursor-based pagination
│   ├── converter/
│   │   └── unit.go           # Multi-unit conversion utilities
│   ├── calculator/
│   │   ├── outstanding.go    # Outstanding amount calculations
│   │   └── tax.go            # Indonesian tax calculations (PPN)
│   └── logger/
│       └── logger.go         # Structured logging wrapper (zap/logrus)
│
├── db/                         # Database-related files
│   ├── migrations/            # SQL migration files
│   │   ├── 000001_init_schema.up.sql
│   │   ├── 000001_init_schema.down.sql
│   │   ├── 000002_add_indexes.up.sql
│   │   └── 000002_add_indexes.down.sql
│   └── seeds/                 # Seed data
│       ├── dev/              # Development seed data
│       │   ├── users.sql
│       │   ├── tenants.sql
│       │   └── products.sql
│       └── test/             # Test seed data
│           └── minimal.sql
│
├── api/                        # API specifications
│   ├── openapi/               # OpenAPI 3.0 specifications
│   │   ├── openapi.yaml      # Main API specification
│   │   └── schemas/          # Reusable schemas
│   │       ├── user.yaml
│   │       ├── product.yaml
│   │       └── sales.yaml
│   └── postman/               # Postman collections
│       └── erp-api.json
│
├── test/                       # Additional test files
│   ├── integration/           # Integration tests
│   │   ├── api/              # API endpoint tests
│   │   │   ├── auth_test.go
│   │   │   ├── sales_test.go
│   │   │   └── warehouse_test.go
│   │   └── repository/       # Repository integration tests
│   │       ├── user_test.go
│   │       └── product_test.go
│   ├── e2e/                   # End-to-end tests
│   │   ├── scenarios/
│   │   │   ├── sales_flow_test.go    # SO → Delivery → Invoice
│   │   │   └── purchase_flow_test.go # PO → GRN → Payment
│   │   └── fixtures/         # Test data fixtures
│   │       └── data.sql
│   ├── testutil/              # Test utilities
│   │   ├── database.go       # Test database setup/teardown
│   │   ├── fixtures.go       # Fixture loading utilities
│   │   └── assertions.go     # Custom test assertions
│   └── mocks/                 # Generated mocks (mockgen/testify)
│       ├── repository/
│       └── service/
│
├── scripts/                    # Build and deployment scripts
│   ├── build.sh               # Build script
│   ├── migrate.sh             # Database migration script
│   ├── seed.sh                # Seed data script
│   ├── test.sh                # Test runner script
│   └── docker/
│       ├── Dockerfile         # Production Dockerfile
│       └── docker-compose.yml # Development docker-compose
│
├── docs/                       # Documentation
│   ├── FOLDER_STRUCTURE.md    # This file
│   ├── API.md                 # API documentation
│   ├── DEPLOYMENT.md          # Deployment guide
│   └── architecture/          # Architecture diagrams
│       └── system_overview.png
│
├── db/                        # Database migrations
│   └── migration.go           # GORM migration code
├── go.mod                     # Go module dependencies
├── go.sum                     # Dependency checksums
├── .env.example               # Environment variable template
├── .gitignore                 # Git ignore rules
├── Makefile                   # Build automation
└── README.md                  # Project overview and setup
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Core business logic and domain models

**Responsibilities**:
- Define domain entities and value objects
- Implement business rules and invariants
- Define repository interfaces (dependency inversion)
- No dependencies on infrastructure or frameworks

**Key Patterns**:
- Rich domain models with behavior
- Repository pattern interfaces
- Value objects for complex types
- Domain events (future enhancement)

**Example**:
```go
// internal/domain/product/model.go
type Product struct {
    ID          string
    TenantID    string
    Name        string
    BaseUnit    string
    IsBatchTracked bool
}

// Business logic in domain
func (p *Product) ValidateForSale() error {
    if !p.IsActive {
        return errors.New("product is inactive")
    }
    return nil
}
```

### 2. Service Layer (`internal/service/`)

**Purpose**: Orchestrate business workflows

**Responsibilities**:
- Coordinate multiple domain objects
- Implement use cases and business workflows
- Transaction management
- Call repository interfaces
- Apply business validation rules

**Key Patterns**:
- Use case/interactor pattern
- Dependency injection
- Transaction management
- Error handling and logging

**Example**:
```go
// internal/service/sales/order.go
type OrderService struct {
    orderRepo    domain.OrderRepository
    productRepo  domain.ProductRepository
    stockService *warehouse.StockService
}

func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*domain.SalesOrder, error) {
    // Orchestrate business logic
    // 1. Validate products exist
    // 2. Check stock availability
    // 3. Calculate totals
    // 4. Create order with transaction
}
```

### 3. Handler Layer (`internal/handler/`)

**Purpose**: HTTP request/response handling

**Responsibilities**:
- Parse HTTP requests
- Validate input
- Call service layer
- Format responses
- Handle HTTP errors

**Key Patterns**:
- Thin controllers (delegate to services)
- Input validation
- Standardized response format
- Error mapping to HTTP status codes

**Example**:
```go
// internal/handler/sales/order.go
type OrderHandler struct {
    orderService *service.OrderService
    validator    *validator.Validate
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err)
        return
    }

    order, err := h.orderService.CreateOrder(c.Request.Context(), req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, err)
        return
    }

    response.Success(c, http.StatusCreated, order)
}
```

### 4. Repository Layer (`internal/repository/`)

**Purpose**: Data access and persistence

**Responsibilities**:
- Implement repository interfaces from domain layer
- Execute database queries
- Map database models to domain models
- Handle database-specific operations

**Key Patterns**:
- Repository pattern
- Tenant isolation enforcement
- Query optimization
- Connection pooling

**Example**:
```go
// internal/repository/gorm/product.go
type ProductRepository struct {
    db *gorm.DB
}

func (r *ProductRepository) FindByID(ctx context.Context, tenantID, productID string) (*domain.Product, error) {
    var product domain.Product
    err := r.db.WithContext(ctx).
        Where("tenant_id = ? AND id = ?", tenantID, productID).
        First(&product).Error
    return &product, err
}
```

### 5. Middleware Layer (`internal/handler/middleware/`)

**Purpose**: Cross-cutting concerns

**Responsibilities**:
- Authentication verification
- Tenant context extraction
- Role-based access control
- Request logging
- CORS handling
- Rate limiting (future)

**Key Patterns**:
- Middleware chain
- Context propagation
- Error handling

### 6. Utilities Layer (`pkg/`)

**Purpose**: Reusable utilities

**Responsibilities**:
- JWT token management
- Password hashing
- Input validation
- Response formatting
- Pagination helpers
- Indonesian-specific utilities (NPWP, tax)

**Key Patterns**:
- Pure functions
- No business logic
- Framework-agnostic
- Well-tested

## Multi-Tenancy Enforcement

### Data Isolation Strategy

**Level 1: Middleware Layer**
```go
// internal/handler/middleware/tenant.go
func TenantContext() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := extractTenantID(c) // From JWT claims
        c.Set("tenantID", tenantID)
        c.Next()
    }
}
```

**Level 2: Service Layer**
```go
// internal/service/product/service.go
func (s *ProductService) GetProduct(ctx context.Context, tenantID, productID string) (*domain.Product, error) {
    // Always pass tenantID to repository
    return s.productRepo.FindByID(ctx, tenantID, productID)
}
```

**Level 3: Repository Layer**
```go
// internal/repository/gorm/product.go
func (r *ProductRepository) FindByID(ctx context.Context, tenantID, productID string) (*domain.Product, error) {
    // CRITICAL: Always filter by tenantID
    var product domain.Product
    err := r.db.WithContext(ctx).
        Where("tenant_id = ? AND id = ?", tenantID, productID).
        First(&product).Error
    return &product, err
}
```

### RBAC Implementation

**Permission Definition**:
```go
// internal/domain/user/role.go
type Permission string

const (
    PermissionViewProduct   Permission = "product:view"
    PermissionCreateProduct Permission = "product:create"
    PermissionEditProduct   Permission = "product:edit"
    PermissionDeleteProduct Permission = "product:delete"
)

var rolePermissions = map[UserRole][]Permission{
    RoleOwner:     {/* all permissions */},
    RoleAdmin:     {/* most permissions */},
    RoleSales:     {PermissionViewProduct, PermissionCreateProduct},
    RoleWarehouse: {PermissionViewProduct},
}
```

**Middleware Enforcement**:
```go
// internal/handler/middleware/rbac.go
func RequirePermission(perm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("userRole")
        if !hasPermission(userRole, perm) {
            response.Error(c, http.StatusForbidden, errors.New("permission denied"))
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Testing Strategy

### Unit Tests

**Location**: Alongside source files (`*_test.go`)

**Coverage**:
- Domain models and business logic
- Service layer use cases
- Utility functions

**Example**:
```go
// internal/domain/product/unit_test.go
func TestProduct_ConvertToBaseUnit(t *testing.T) {
    product := &Product{BaseUnit: "PCS"}
    unit := &ProductUnit{UnitName: "BOX", ConversionRate: 12}

    baseQty := product.ConvertToBaseUnit(5, unit)
    assert.Equal(t, 60, baseQty) // 5 BOX * 12 = 60 PCS
}
```

### Integration Tests

**Location**: `test/integration/`

**Coverage**:
- Repository layer with real database
- API endpoints with test server
- Service layer with dependencies

**Example**:
```go
// test/integration/repository/product_test.go
func TestProductRepository_FindByID(t *testing.T) {
    db := testutil.SetupTestDB(t)
    defer testutil.TeardownTestDB(t, db)

    repo := gorm.NewProductRepository(db)
    product, err := repo.FindByID(ctx, "tenant1", "product1")

    assert.NoError(t, err)
    assert.Equal(t, "Product 1", product.Name)
}
```

### E2E Tests

**Location**: `test/e2e/`

**Coverage**:
- Complete business workflows
- Multi-step operations
- Cross-module interactions

**Example**:
```go
// test/e2e/scenarios/sales_flow_test.go
func TestSalesFlow_OrderToInvoice(t *testing.T) {
    // 1. Create sales order
    // 2. Create delivery
    // 3. Confirm delivery
    // 4. Generate invoice
    // 5. Verify stock movements
    // 6. Verify financial records
}
```

## Configuration Management

### Environment Variables

**File**: `.env.example`

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/erp_db

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Tenant
DEFAULT_SUBSCRIPTION_PRICE=300000
TRIAL_PERIOD_DAYS=14
GRACE_PERIOD_DAYS=7

# Indonesian Tax
DEFAULT_PPN_RATE=11
```

### Config Structure

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Tenant   TenantConfig
    Tax      TaxConfig
}

func Load() (*Config, error) {
    // Load from environment variables
    // Validate configuration
    // Return config struct
}
```

## Dependency Injection

**Pattern**: Constructor injection

**Example**:
```go
// cmd/api/main.go
func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize dependencies
    db := bootstrap.InitDatabase(cfg.Database)
    logger := bootstrap.InitLogger(cfg.Server)

    // Repositories
    productRepo := gorm.NewProductRepository(db)

    // Services
    productService := service.NewProductService(productRepo, logger)

    // Handlers
    productHandler := handler.NewProductHandler(productService)

    // Setup router
    router := setupRouter(productHandler)
    router.Run(cfg.Server.Port)
}
```

## Performance Optimization

### Database Indexing

**Critical Indexes** (already in schema):
- `(tenantId, code)` - Unique constraint + fast lookups
- `(tenantId, isActive)` - Active record queries
- `(invoiceDate, dueDate)` - Date range queries
- Foreign keys - Join optimization

### Query Optimization

**Best Practices**:
- Always filter by `tenantId` first (indexed)
- Use `Select()` to fetch only needed columns
- Implement pagination for large result sets
- Use eager loading for relationships (`Preload()`)
- Batch operations for bulk inserts/updates

**Example**:
```go
// Efficient query with selective fields
products, err := db.Model(&Product{}).
    Select("id, name, price, is_active").
    Where("tenant_id = ? AND is_active = ?", tenantID, true).
    Limit(20).Offset(offset).
    Find(&products)
```

### Caching Strategy (Future Enhancement)

**Candidates for caching**:
- Company settings (low change frequency)
- Product catalog (moderate change frequency)
- User permissions (session-based)
- Currency rates (daily updates)

**Implementation location**: `internal/cache/`

## Security Considerations

### Password Security

**Implementation**: `pkg/auth/hasher.go`
- Use bcrypt or argon2 for password hashing
- Never store plaintext passwords
- Implement password strength requirements

### JWT Security

**Implementation**: `pkg/auth/jwt.go`
- Short token expiry (24h recommended)
- Refresh token mechanism
- Token revocation list (future)
- Include `tenantId` in claims

### Input Validation

**Implementation**: `pkg/validator/`
- Validate all input using go-playground/validator
- Custom validators for Indonesian formats (NPWP)
- Sanitize input to prevent injection attacks

### SQL Injection Prevention

**Strategy**:
- Always use parameterized queries (GORM does this)
- Never concatenate user input into SQL
- Use GORM query builder methods

### Tenant Isolation

**Critical**: Every query MUST include `tenantId` filter
- Enforced at repository layer
- Validated at service layer
- Extracted from JWT at middleware layer

## Indonesian Compliance

### NPWP Validation

**Implementation**: `pkg/validator/npwp.go`
```go
func ValidateNPWP(npwp string) error {
    // Format: XX.XXX.XXX.X-XXX.XXX
    pattern := `^\d{2}\.\d{3}\.\d{3}\.\d-\d{3}\.\d{3}$`
    // Validate format and checksum
}
```

### PPN (Tax) Calculation

**Implementation**: `pkg/calculator/tax.go`
```go
func CalculatePPN(amount decimal.Decimal, rate decimal.Decimal) decimal.Decimal {
    // PPN = amount * (rate / 100)
    // Default rate: 11% (as of 2025)
}
```

### Invoice Numbering

**Format**: `{PREFIX}/{NUMBER}/{MONTH}/{YEAR}`
**Example**: `INV/001/12/2025`

**Implementation**: `internal/service/sales/invoice.go`
```go
func GenerateInvoiceNumber(company *domain.Company, date time.Time) string {
    // Extract format from company settings
    // Apply number padding
    // Return formatted invoice number
}
```

## Implementation Roadmap

### Phase 1: Core Infrastructure (Week 1-2)
- [ ] Setup `cmd/api/main.go` and bootstrap
- [ ] Implement `internal/config/` configuration management
- [ ] Setup `pkg/logger/` structured logging
- [ ] Create `db/migrations/` using GORM AutoMigrate
- [ ] Setup `pkg/response/` standardized responses

### Phase 2: Authentication & Authorization (Week 3)
- [ ] Implement `internal/domain/user/` and `internal/domain/tenant/`
- [ ] Build `internal/service/auth/` authentication service
- [ ] Create `internal/handler/middleware/` (auth, tenant, rbac)
- [ ] Setup `pkg/auth/` JWT and password utilities
- [ ] Implement login/logout endpoints

### Phase 3: Core Domain Implementation (Week 4-6)
- [ ] Implement all 11 domain modules in `internal/domain/`
- [ ] Build corresponding services in `internal/service/`
- [ ] Create repository implementations in `internal/repository/gorm/`
- [ ] Develop handlers in `internal/handler/`
- [ ] Unit tests for domain and service layers

### Phase 4: Business Workflows (Week 7-8)
- [ ] Sales flow: SO → Delivery → Invoice
- [ ] Purchase flow: PO → GRN → Payment
- [ ] Inventory management & stock tracking
- [ ] Financial management & cash transactions
- [ ] Integration tests for workflows

### Phase 5: Multi-Tenancy Features (Week 9)
- [ ] Subscription billing logic
- [ ] Background worker implementation (`cmd/worker/`)
- [ ] Outstanding tracking & monitoring
- [ ] Batch expiry monitoring
- [ ] Tenant isolation validation tests

### Phase 6: Testing & Documentation (Week 10-11)
- [ ] Complete unit test coverage (>80%)
- [ ] Integration tests for all repositories
- [ ] E2E tests for critical workflows
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Code documentation and examples

### Phase 7: Deployment & Monitoring (Week 12)
- [ ] Docker containerization (`scripts/docker/`)
- [ ] CI/CD pipeline setup (GitHub Actions)
- [ ] Monitoring & logging infrastructure
- [ ] Production deployment scripts
- [ ] Performance testing and optimization

## Best Practices

### Code Organization
✅ One file per domain concept
✅ Group related files in same directory
✅ Use meaningful file and package names
✅ Keep files under 500 lines
✅ Separate interfaces from implementations

### Error Handling
✅ Use custom error types for business errors
✅ Wrap errors with context (`fmt.Errorf("context: %w", err)`)
✅ Log errors at appropriate levels
✅ Return errors, don't panic (except in initialization)
✅ Use sentinel errors for known conditions

### Testing
✅ Write tests alongside code
✅ Use table-driven tests
✅ Mock external dependencies
✅ Test edge cases and error conditions
✅ Maintain >80% code coverage

### Documentation
✅ Add godoc comments for public APIs
✅ Document complex business logic
✅ Keep README.md up to date
✅ Maintain architecture diagrams
✅ Document deployment procedures

### Performance
✅ Profile before optimizing
✅ Use appropriate indexes
✅ Implement pagination for large datasets
✅ Cache frequently accessed data
✅ Monitor database query performance

## Common Pitfalls to Avoid

❌ **Missing Tenant Filter**: Always include `tenantId` in queries
❌ **Wrong Unit Calculations**: Convert to base units before stock operations
❌ **Batch Tracking Skip**: Check `isBatchTracked` before allowing batch-less operations
❌ **Expiry Date Ignore**: Validate expiry dates for perishable products
❌ **Outstanding Mismatch**: Update customer/supplier outstanding when creating invoices/payments
❌ **Stock Before/After Skip**: Always record stock before/after in inventory movements
❌ **Hard Deletes**: Use soft deletes (`isActive = false`) for transactional data
❌ **Circular Dependencies**: Service A calling Service B calling Service A
❌ **Business Logic in Handlers**: Keep handlers thin, move logic to services
❌ **Direct Database Access**: Always use repository pattern

## Tools & Libraries Recommendations

### Essential Libraries
- **HTTP Framework**: Gin (https://github.com/gin-gonic/gin)
- **ORM**: GORM (already included)
- **Validation**: go-playground/validator
- **Logging**: zap or logrus
- **Testing**: testify
- **Mock Generation**: mockgen
- **Migration**: golang-migrate

### Development Tools
- **Air**: Live reload for Go apps
- **golangci-lint**: Comprehensive linter
- **godoc**: Documentation generation
- **pprof**: Performance profiling

## Conclusion

This folder structure provides:
✅ **Scalability**: Clean architecture supports horizontal scaling
✅ **Maintainability**: Clear separation of concerns and DDD principles
✅ **Testability**: Dependency injection and interface-based design
✅ **Multi-Tenancy**: Strict isolation enforced at all layers
✅ **Indonesian Compliance**: Built-in support for tax and legal requirements
✅ **Developer Experience**: Intuitive organization and clear boundaries

Follow this structure to build a robust, scalable, and maintainable Multi-Tenant ERP system.
