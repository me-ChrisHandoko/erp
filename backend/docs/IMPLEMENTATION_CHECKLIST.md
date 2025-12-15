# Implementation Checklist

Quick reference guide for implementing the folder structure.

## Quick Start Commands

```bash
# Create all folders at once
mkdir -p cmd/{api/bootstrap,worker/jobs}
mkdir -p internal/{domain,service,handler,repository/gorm,worker,config}
mkdir -p internal/domain/{user,tenant,company,warehouse,product,customer,supplier,sales,purchase,delivery,inventory,finance}
mkdir -p internal/service/{auth,tenant,user,warehouse,product,sales,purchase,inventory,finance}
mkdir -p internal/handler/{middleware,auth,tenant,user,warehouse,product,customer,supplier,sales,purchase,inventory,finance}
mkdir -p pkg/{auth,validator,response,pagination,converter,calculator,logger}
mkdir -p db/{migrations,seeds/{dev,test}}
mkdir -p api/{openapi/schemas,postman}
mkdir -p test/{integration/{api,repository},e2e/{scenarios,fixtures},testutil,mocks/{repository,service}}
mkdir -p scripts/docker
mkdir -p docs/architecture
```

## Implementation Checklist by Phase

### Phase 1: Core Infrastructure ✅

#### Week 1: Project Setup
- [ ] Create folder structure (use commands above)
- [ ] Initialize Go modules: `go mod init backend`
- [ ] Setup `.env.example` with all configuration keys
- [ ] Create `.gitignore` for Go projects
- [ ] Setup `Makefile` for common tasks
- [ ] Install essential dependencies:
  ```bash
  go get -u github.com/gin-gonic/gin
  go get -u gorm.io/gorm
  go get -u gorm.io/driver/postgres
  go get -u github.com/joho/godotenv
  go get -u github.com/golang-jwt/jwt/v5
  go get -u github.com/go-playground/validator/v10
  ```

#### Week 2: Configuration & Logging
- [ ] Implement `internal/config/config.go`
- [ ] Implement `internal/config/env.go`
- [ ] Implement `internal/config/database.go`
- [ ] Implement `pkg/logger/logger.go` (using zap or logrus)
- [ ] Implement `pkg/response/json.go`
- [ ] Implement `pkg/response/error.go`
- [ ] Create `cmd/api/bootstrap/database.go`
- [ ] Create `cmd/api/bootstrap/logger.go`
- [ ] Create basic `cmd/api/main.go`
- [ ] Test: Configuration loading and database connection

### Phase 2: Authentication & Authorization ✅

#### Week 3: Auth Foundation
- [ ] Define `internal/domain/user/model.go`
- [ ] Define `internal/domain/user/role.go`
- [ ] Define `internal/domain/user/repository.go` (interface)
- [ ] Define `internal/domain/tenant/model.go`
- [ ] Define `internal/domain/tenant/repository.go` (interface)
- [ ] Implement `pkg/auth/jwt.go`
- [ ] Implement `pkg/auth/hasher.go`
- [ ] Implement `pkg/auth/rbac.go`
- [ ] Implement `internal/repository/gorm/user.go`
- [ ] Implement `internal/repository/gorm/tenant.go`
- [ ] Implement `internal/service/auth/service.go`
- [ ] Implement `internal/service/auth/jwt.go`
- [ ] Implement `internal/service/auth/password.go`
- [ ] Implement `internal/handler/middleware/auth.go`
- [ ] Implement `internal/handler/middleware/tenant.go`
- [ ] Implement `internal/handler/middleware/rbac.go`
- [ ] Implement `internal/handler/middleware/logger.go`
- [ ] Implement `internal/handler/auth/handler.go`
- [ ] Create `db/migrations/000001_init_users_tenants.up.sql`
- [ ] Test: Login, JWT generation, token validation
- [ ] Test: Tenant context extraction
- [ ] Test: RBAC middleware

### Phase 3: Core Domain Implementation ✅

#### Week 4-5: Domain Models & Repositories
**For each domain** (company, warehouse, product, customer, supplier):
- [ ] Define `internal/domain/{module}/model.go`
- [ ] Define business logic files (e.g., `batch.go`, `unit.go`, `outstanding.go`)
- [ ] Define `internal/domain/{module}/repository.go` (interface)
- [ ] Implement `internal/repository/gorm/{module}.go`
- [ ] Write unit tests for domain logic
- [ ] Write integration tests for repositories
- [ ] Create migration files in `db/migrations/`

**Specific Tasks**:
- [ ] Product: Multi-unit conversion logic in `internal/domain/product/unit.go`
- [ ] Product: Batch tracking logic in `internal/domain/product/batch.go`
- [ ] Customer/Supplier: Outstanding calculation in `outstanding.go`
- [ ] Warehouse: Stock level management in `stock.go`

#### Week 6: Service Layer
**For each domain**:
- [ ] Implement `internal/service/{module}/service.go`
- [ ] Implement specific service files (e.g., `batch.go`, `pricing.go`)
- [ ] Write unit tests with mocked repositories
- [ ] Document service methods (godoc comments)

### Phase 4: Business Workflows ✅

#### Week 7: Sales & Delivery Workflow
- [ ] Implement `internal/domain/sales/` (model, order, invoice)
- [ ] Implement `internal/domain/delivery/` (model, fulfillment)
- [ ] Implement `internal/repository/gorm/sales.go`
- [ ] Implement `internal/service/sales/order.go`
- [ ] Implement `internal/service/sales/invoice.go`
- [ ] Implement `internal/service/sales/delivery.go`
- [ ] Implement `internal/handler/sales/order.go`
- [ ] Implement `internal/handler/sales/invoice.go`
- [ ] Implement `internal/handler/sales/delivery.go`
- [ ] Create migrations for sales tables
- [ ] Write E2E test: SO → Delivery → Invoice flow

#### Week 8: Purchase & Inventory Workflow
- [ ] Implement `internal/domain/purchase/` (model, order, receipt)
- [ ] Implement `internal/domain/inventory/` (model, movement, opname)
- [ ] Implement `internal/repository/gorm/purchase.go`
- [ ] Implement `internal/repository/gorm/inventory.go`
- [ ] Implement `internal/service/purchase/order.go`
- [ ] Implement `internal/service/purchase/receipt.go`
- [ ] Implement `internal/service/inventory/movement.go`
- [ ] Implement `internal/service/inventory/opname.go`
- [ ] Implement handlers for purchase and inventory
- [ ] Create migrations for purchase/inventory tables
- [ ] Write E2E test: PO → GRN → Stock Movement flow

### Phase 5: Multi-Tenancy Features ✅

#### Week 9: Subscription & Background Jobs
- [ ] Implement `internal/service/tenant/subscription.go`
- [ ] Implement `internal/worker/scheduler.go`
- [ ] Implement `internal/worker/queue.go`
- [ ] Implement `internal/worker/processor.go`
- [ ] Implement `cmd/worker/jobs/subscription.go`
- [ ] Implement `cmd/worker/jobs/expiry.go`
- [ ] Implement `cmd/worker/jobs/outstanding.go`
- [ ] Implement `cmd/worker/main.go`
- [ ] Test: Subscription billing automation
- [ ] Test: Batch expiry monitoring
- [ ] Test: Outstanding recalculation
- [ ] Validate: Tenant isolation across all endpoints

### Phase 6: Testing & Documentation ✅

#### Week 10: Comprehensive Testing
- [ ] Achieve >80% unit test coverage
- [ ] Complete integration tests for all repositories
- [ ] Complete E2E tests for critical workflows:
  - [ ] Sales flow (SO → Delivery → Invoice)
  - [ ] Purchase flow (PO → GRN → Payment)
  - [ ] Inventory flow (Movement → Opname)
  - [ ] Subscription flow (Trial → Active → Billing)
- [ ] Setup test fixtures in `test/e2e/fixtures/`
- [ ] Generate mocks for services and repositories
- [ ] Run full test suite: `make test`

#### Week 11: API Documentation
- [ ] Create OpenAPI specification in `api/openapi/openapi.yaml`
- [ ] Document all endpoints with request/response examples
- [ ] Create Postman collection in `api/postman/`
- [ ] Document authentication flow
- [ ] Document multi-tenancy headers
- [ ] Setup Swagger UI for interactive docs
- [ ] Write API usage examples in `docs/API.md`

### Phase 7: Deployment & Monitoring ✅

#### Week 12: Production Readiness
- [ ] Create production Dockerfile in `scripts/docker/Dockerfile`
- [ ] Create docker-compose.yml for development
- [ ] Setup CI/CD pipeline (GitHub Actions or GitLab CI)
- [ ] Implement health check endpoint (`/health`)
- [ ] Implement metrics endpoint (`/metrics`) using Prometheus
- [ ] Setup structured logging with log levels
- [ ] Create deployment scripts in `scripts/`
- [ ] Setup database backup automation
- [ ] Load testing with realistic data
- [ ] Security audit (OWASP Top 10)
- [ ] Performance optimization and profiling

## Critical Validation Checklist

### Multi-Tenancy Validation ⚠️
Before deploying to production, verify:
- [ ] All queries include `tenantId` filter
- [ ] Middleware correctly extracts tenant from JWT
- [ ] Cross-tenant access is blocked (integration test)
- [ ] Subscription status checked before operations
- [ ] Grace period logic works correctly
- [ ] Trial expiration prevents access

### Security Validation ⚠️
- [ ] Passwords are hashed (bcrypt/argon2)
- [ ] JWT tokens expire and refresh correctly
- [ ] RBAC permissions enforced on all protected endpoints
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (GORM parameterized queries)
- [ ] CORS configured correctly
- [ ] Rate limiting implemented (optional but recommended)
- [ ] Sensitive data not logged

### Data Integrity Validation ⚠️
- [ ] Outstanding amounts calculated correctly
- [ ] Stock movements recorded for all inventory changes
- [ ] Batch tracking enforced for `isBatchTracked` products
- [ ] Expiry dates validated for perishable products
- [ ] Multi-unit conversions accurate
- [ ] Invoice numbering sequential and unique
- [ ] Transaction rollback on errors

### Indonesian Compliance Validation ⚠️
- [ ] NPWP validation works correctly
- [ ] PPN calculations accurate (11% default)
- [ ] Invoice numbering follows format
- [ ] Tax compliance fields captured
- [ ] Indonesian language support (future)

## Quick Reference: File Creation Order

### Order of Implementation (Dependencies)
1. **Config & Utils First** (no dependencies)
   - `internal/config/`, `pkg/logger/`, `pkg/response/`

2. **Domain Models** (no external dependencies)
   - `internal/domain/*/model.go`

3. **Repository Interfaces** (depends on domain models)
   - `internal/domain/*/repository.go`

4. **Repository Implementations** (depends on interfaces)
   - `internal/repository/gorm/*.go`

5. **Services** (depends on repositories)
   - `internal/service/*/service.go`

6. **Handlers** (depends on services)
   - `internal/handler/*/handler.go`

7. **Middleware** (depends on services for auth)
   - `internal/handler/middleware/*.go`

8. **Main Application** (depends on everything)
   - `cmd/api/main.go`, `cmd/api/router.go`

## Development Workflow

### Daily Development Cycle
```bash
# 1. Pull latest changes
git pull origin main

# 2. Create feature branch
git checkout -b feature/implement-sales-order

# 3. Run tests before changes
make test

# 4. Implement feature
# - Write tests first (TDD approach)
# - Implement domain logic
# - Implement service layer
# - Implement handler
# - Run tests after each layer

# 5. Run linter
make lint

# 6. Commit with meaningful message
git add .
git commit -m "feat: implement sales order creation with batch tracking"

# 7. Push and create PR
git push origin feature/implement-sales-order
```

### Makefile Targets (Recommended)
```makefile
# Makefile
.PHONY: help build run test lint migrate seed

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/api cmd/api/main.go
	go build -o bin/worker cmd/worker/main.go

run: ## Run the application
	go run cmd/api/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

lint: ## Run linter
	golangci-lint run

migrate: ## Run database migrations
	./scripts/migrate.sh up

seed: ## Seed database
	./scripts/seed.sh dev

docker-build: ## Build Docker image
	docker build -f scripts/docker/Dockerfile -t erp-backend:latest .

docker-up: ## Start development environment
	docker-compose -f scripts/docker/docker-compose.yml up -d
```

## Common Issues & Solutions

### Issue: Circular Dependencies
**Problem**: Service A imports Service B, Service B imports Service A
**Solution**: Extract shared logic to a separate package or use interfaces

### Issue: Missing Tenant ID in Query
**Problem**: Query returns data from multiple tenants
**Solution**: Always add `.Where("tenant_id = ?", tenantID)` to queries

### Issue: Wrong Stock Calculation
**Problem**: Stock doesn't match after transactions
**Solution**: Always convert to base units before stock operations

### Issue: JWT Token Rejected
**Problem**: "Invalid token" error
**Solution**: Check JWT secret in env, verify token expiry, check claims structure

### Issue: Migration Fails
**Problem**: Migration fails with constraint error
**Solution**: Check migration order, ensure referenced tables exist first

## Next Steps After Folder Creation

1. **Start with Phase 1**: Focus on getting the foundation right
2. **Follow TDD**: Write tests before implementation
3. **Document as you go**: Add godoc comments for public APIs
4. **Review regularly**: Code review after each module completion
5. **Refactor early**: Don't let technical debt accumulate
6. **Monitor performance**: Profile critical paths early

## Resources & References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Clean Architecture in Go](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [GORM Documentation](https://gorm.io/docs/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

---

**Good luck with implementation! Follow the checklist step by step for best results.**
