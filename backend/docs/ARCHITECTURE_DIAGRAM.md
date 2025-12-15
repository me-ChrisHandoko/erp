# Architecture Diagram

Visual representation of the Multi-Tenant ERP System architecture.

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           CLIENT APPLICATIONS                            │
│                    (Web Browser, Mobile App, API Clients)                │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ HTTPS
                                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                             API GATEWAY / LOAD BALANCER                  │
│                         (nginx, AWS ALB, GCP Load Balancer)              │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │
                    ┌─────────────────┴─────────────────┐
                    │                                   │
                    ▼                                   ▼
┌─────────────────────────────────┐  ┌─────────────────────────────────┐
│      cmd/api (REST API)         │  │     cmd/worker (Background)     │
│                                 │  │                                 │
│  ┌──────────────────────────┐  │  │  ┌──────────────────────────┐  │
│  │   HTTP Server (Gin)      │  │  │  │   Job Scheduler (Cron)   │  │
│  │   - Port 8080            │  │  │  │   - Subscription Billing │  │
│  │   - TLS Support          │  │  │  │   - Expiry Monitoring    │  │
│  │   - Graceful Shutdown    │  │  │  │   - Outstanding Calc     │  │
│  └──────────────────────────┘  │  │  └──────────────────────────┘  │
└─────────────────────────────────┘  └─────────────────────────────────┘
                    │
                    │ Request Flow
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        MIDDLEWARE LAYER                                  │
│  ┌──────────────┬──────────────┬──────────────┬──────────────────────┐ │
│  │   Logger     │    CORS      │  Auth (JWT)  │  Tenant Context      │ │
│  │  Middleware  │  Middleware  │  Middleware  │  Extraction          │ │
│  └──────────────┴──────────────┴──────────────┴──────────────────────┘ │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │              RBAC (Role-Based Access Control)                    │  │
│  │              - Permission Checking per Tenant                    │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Authenticated Request + Tenant Context
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          HANDLER LAYER                                   │
│                        (HTTP Controllers)                                │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │  Auth   │ Tenant  │  User   │Warehouse│ Product │   Customer      │ │
│  │ Handler │ Handler │ Handler │ Handler │ Handler │   Handler       │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │ Sales   │Purchase │Delivery │Inventory│ Finance │   Supplier      │ │
│  │ Handler │ Handler │ Handler │ Handler │ Handler │   Handler       │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
│                                                                           │
│  Responsibilities:                                                        │
│  - Parse HTTP requests                                                    │
│  - Validate input                                                         │
│  - Call service layer                                                     │
│  - Format responses                                                       │
└─────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Business Logic Delegation
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         SERVICE LAYER                                    │
│                   (Business Logic Orchestration)                         │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │  Auth   │ Tenant  │  User   │Warehouse│ Product │   Customer      │ │
│  │ Service │ Service │ Service │ Service │ Service │   Service       │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │ Sales   │Purchase │Delivery │Inventory│ Finance │   Supplier      │ │
│  │ Service │ Service │ Service │ Service │ Service │   Service       │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
│                                                                           │
│  Responsibilities:                                                        │
│  - Orchestrate domain objects                                            │
│  - Implement use cases                                                   │
│  - Transaction management                                                │
│  - Business validation                                                   │
└─────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Domain Logic Execution
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          DOMAIN LAYER                                    │
│                    (Core Business Logic - DDD)                           │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │  User   │ Tenant  │ Company │Warehouse│ Product │   Customer      │ │
│  │ Domain  │ Domain  │ Domain  │ Domain  │ Domain  │   Domain        │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┐ │
│  │ Sales   │Purchase │Delivery │Inventory│ Finance │   Supplier      │ │
│  │ Domain  │ Domain  │ Domain  │ Domain  │ Domain  │   Domain        │ │
│  └─────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┘ │
│                                                                           │
│  Components per Domain:                                                  │
│  - Entities & Value Objects                                              │
│  - Business Rules & Invariants                                           │
│  - Repository Interfaces                                                 │
│  - Domain Events (future)                                                │
└─────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Data Persistence Operations
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       REPOSITORY LAYER                                   │
│                    (Data Access - Infrastructure)                        │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │              GORM Repository Implementations                     │  │
│  │  ┌─────────┬─────────┬─────────┬─────────┬─────────────────────┐│  │
│  │  │  User   │ Tenant  │Warehouse│ Product │   Sales             ││  │
│  │  │  Repo   │  Repo   │  Repo   │  Repo   │   Repo              ││  │
│  │  └─────────┴─────────┴─────────┴─────────┴─────────────────────┘│  │
│  │  ┌─────────┬─────────┬─────────┬─────────┬─────────────────────┐│  │
│  │  │Purchase │Inventory│ Finance │Customer │   Supplier          ││  │
│  │  │  Repo   │  Repo   │  Repo   │  Repo   │   Repo              ││  │
│  │  └─────────┴─────────┴─────────┴─────────┴─────────────────────┘│  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                           │
│  Responsibilities:                                                        │
│  - Execute database queries                                              │
│  - Enforce tenant isolation (CRITICAL)                                   │
│  - Map DB models to domain models                                        │
│  - Connection pooling                                                    │
└─────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Database Operations
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         DATABASE LAYER                                   │
│                      PostgreSQL / SQLite                                 │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                       TABLES (Multi-Tenant)                      │  │
│  │                                                                  │  │
│  │  Core:                                                           │  │
│  │  - users, user_tenants, tenants, subscriptions, payments        │  │
│  │  - companies                                                     │  │
│  │                                                                  │  │
│  │  Warehouse & Inventory:                                         │  │
│  │  - warehouses, warehouse_stocks, inventory_movements            │  │
│  │  - stock_opnames, stock_transfers                               │  │
│  │                                                                  │  │
│  │  Products:                                                       │  │
│  │  - products, product_units, product_batches                     │  │
│  │                                                                  │  │
│  │  Business Entities:                                             │  │
│  │  - customers, suppliers                                         │  │
│  │                                                                  │  │
│  │  Transactions:                                                   │  │
│  │  - sales_orders, sales_order_items, deliveries, delivery_items  │  │
│  │  - invoices, invoice_items, invoice_payments                    │  │
│  │  - purchase_orders, po_items, goods_receipts, grn_items         │  │
│  │  - cash_transactions, checks, supplier_payments                 │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                           │
│  Indexes:                                                                 │
│  - (tenant_id, code) - Unique constraints                                │
│  - (tenant_id, is_active) - Active record queries                        │
│  - Date indexes - Reporting queries                                      │
└─────────────────────────────────────────────────────────────────────────┘
                    │
                    │ Database Migrations
                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      db/migrations/                                      │
│  - 000001_init_users_tenants.up.sql                                     │
│  - 000002_create_products.up.sql                                        │
│  - 000003_create_sales.up.sql                                           │
│  - 000004_create_inventory.up.sql                                       │
│  - 000005_add_indexes.up.sql                                            │
└─────────────────────────────────────────────────────────────────────────┘
```

## Clean Architecture Layer Dependencies

```
┌──────────────────────────────────────────────────────────────┐
│                    Dependency Direction                      │
│                  (Inner layers know nothing                  │
│                   about outer layers)                        │
└──────────────────────────────────────────────────────────────┘

                  ┌─────────────────────┐
                  │   External World    │
                  │  (HTTP, Database)   │
                  └─────────────────────┘
                           ▲
                           │
                  ┌────────┴────────┐
                  │                 │
         ┌────────▼──────┐  ┌──────▼──────────┐
         │   Handler     │  │   Repository    │
         │   (HTTP)      │  │   (GORM)        │
         └────────┬──────┘  └──────┬──────────┘
                  │                │
                  └────────┬───────┘
                           │
                  ┌────────▼────────┐
                  │    Service      │
                  │   (Use Cases)   │
                  └────────┬────────┘
                           │
                  ┌────────▼────────┐
                  │     Domain      │
                  │ (Business Logic)│
                  │   - Entities    │
                  │   - Interfaces  │
                  └─────────────────┘
                    (Innermost Layer)

Key Principle: Dependencies point INWARD
- Domain layer has NO dependencies
- Service depends on Domain
- Handler depends on Service
- Repository implements Domain interfaces
```

## Multi-Tenancy Data Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    REQUEST WITH JWT TOKEN                                │
│  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...         │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   AUTH MIDDLEWARE                                        │
│  1. Validate JWT signature                                               │
│  2. Extract claims (userID, tenantID, role)                             │
│  3. Verify token expiry                                                  │
│  4. Check user is active                                                 │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   TENANT CONTEXT MIDDLEWARE                              │
│  1. Extract tenantID from JWT claims                                     │
│  2. Verify tenant is active or in trial                                  │
│  3. Check subscription status                                            │
│  4. Inject tenantID into request context                                 │
│                                                                           │
│  Context:                                                                 │
│    - tenantID: "tenant_123"                                              │
│    - userID: "user_456"                                                  │
│    - role: "SALES"                                                       │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   RBAC MIDDLEWARE                                        │
│  1. Get user role from context                                           │
│  2. Check permission for endpoint                                        │
│  3. Allow/Deny based on role                                             │
│                                                                           │
│  Example: POST /api/products requires "product:create"                  │
│  SALES role has this permission ✓                                       │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   HANDLER (Controller)                                   │
│  GET /api/products/:id                                                   │
│                                                                           │
│  func (h *ProductHandler) GetProduct(c *gin.Context) {                  │
│      tenantID := c.GetString("tenantID")  // From middleware             │
│      productID := c.Param("id")                                          │
│                                                                           │
│      product, err := h.productService.GetProduct(                        │
│          c.Request.Context(),                                            │
│          tenantID,  // Explicitly passed                                 │
│          productID,                                                      │
│      )                                                                   │
│      ...                                                                 │
│  }                                                                       │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   SERVICE LAYER                                          │
│                                                                           │
│  func (s *ProductService) GetProduct(                                    │
│      ctx context.Context,                                                │
│      tenantID string,  // Always required                                │
│      productID string,                                                   │
│  ) (*domain.Product, error) {                                            │
│      // Business logic validation                                        │
│      return s.productRepo.FindByID(ctx, tenantID, productID)            │
│  }                                                                       │
└─────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                   REPOSITORY LAYER                                       │
│                                                                           │
│  func (r *ProductRepository) FindByID(                                   │
│      ctx context.Context,                                                │
│      tenantID string,                                                    │
│      productID string,                                                   │
│  ) (*domain.Product, error) {                                            │
│      var product domain.Product                                          │
│      err := r.db.WithContext(ctx).                                       │
│          Where("tenant_id = ? AND id = ?", tenantID, productID). ◄─────┐│
│          First(&product).Error                                       │  ││
│      return &product, err                                            │  ││
│  }                                                                   │  ││
│                                                                      │  ││
│  ┌──────────────────────────────────────────────────────────────┐  │  ││
│  │  CRITICAL: tenantID filter ALWAYS present                   │  │  ││
│  │  Prevents cross-tenant data leakage                         │──┘  ││
│  └──────────────────────────────────────────────────────────────┘     ││
└─────────────────────────────────────────────────────────────────────────┘│
                              │                                             │
                              ▼                                             │
┌─────────────────────────────────────────────────────────────────────────┐│
│                   DATABASE QUERY                                         ││
│                                                                           ││
│  SELECT * FROM products                                                  ││
│  WHERE tenant_id = 'tenant_123'  ◄───────────────────────────────────────┘
│    AND id = 'product_789'                                                 │
│    AND is_active = true                                                   │
│                                                                            │
│  Result: Only returns products for tenant_123                             │
└─────────────────────────────────────────────────────────────────────────┘
```

## Sales Workflow Example

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SALES ORDER WORKFLOW                                 │
└─────────────────────────────────────────────────────────────────────────┘

Step 1: Create Sales Order
─────────────────────────────────────────
POST /api/sales/orders
{
  "customerID": "customer_123",
  "items": [
    {
      "productID": "product_456",
      "unitName": "KARTON",
      "quantity": 10,
      "unitPrice": 240000
    }
  ]
}
                │
                ▼
┌───────────────────────────────────────┐
│  SalesOrderService.CreateOrder()     │
│  1. Validate customer exists         │
│  2. Check product stock availability │
│  3. Calculate totals (base units)    │
│  4. Reserve stock (optional)         │
│  5. Create SO record                 │
│  Status: PENDING                     │
└───────────────────────────────────────┘
                │
                ▼
     Sales Order Created
     SO Number: SO/001/12/2025
     Status: PENDING


Step 2: Create Delivery
─────────────────────────────────────────
POST /api/deliveries
{
  "salesOrderID": "so_789",
  "warehouseID": "warehouse_main",
  "deliveryDate": "2025-12-16",
  "items": [
    {
      "productID": "product_456",
      "batchID": "batch_nearest_expiry",
      "quantity": 10
    }
  ]
}
                │
                ▼
┌───────────────────────────────────────┐
│  DeliveryService.CreateDelivery()    │
│  1. Validate SO exists and PENDING   │
│  2. Select batches (FEFO for food)   │
│  3. Create delivery record           │
│  4. Create delivery items            │
│  5. Update SO status to PROCESSING   │
│  Status: PREPARED                    │
└───────────────────────────────────────┘
                │
                ▼
     Delivery Created
     Delivery Number: DEL/001/12/2025
     Status: PREPARED


Step 3: Update Delivery Status
─────────────────────────────────────────
PATCH /api/deliveries/:id/status
{
  "status": "DELIVERED",
  "deliveredAt": "2025-12-16T14:30:00Z",
  "receivedBy": "John Doe"
}
                │
                ▼
┌───────────────────────────────────────┐
│  DeliveryService.UpdateStatus()      │
│  1. Validate status transition       │
│  2. Update delivery status           │
│  3. Record stock movement            │
│  4. Reduce warehouse stock           │
│  5. Update SO status to DELIVERED    │
│  Status: DELIVERED                   │
└───────────────────────────────────────┘
                │
                ▼
     Stock Movement Created
     Type: OUT
     Reference: Delivery DEL/001/12/2025


Step 4: Generate Invoice
─────────────────────────────────────────
POST /api/invoices
{
  "salesOrderID": "so_789",
  "deliveryID": "delivery_001",
  "dueDate": "2025-12-23"
}
                │
                ▼
┌───────────────────────────────────────┐
│  InvoiceService.GenerateInvoice()    │
│  1. Get SO and delivery details      │
│  2. Calculate invoice totals         │
│  3. Apply PPN tax (11%)              │
│  4. Generate invoice number          │
│  5. Create invoice record            │
│  6. Update customer outstanding      │
│  Status: UNPAID                      │
└───────────────────────────────────────┘
                │
                ▼
     Invoice Created
     Invoice Number: INV/001/12/2025
     Total: Rp 2,664,000 (including PPN)
     Status: UNPAID
     Due Date: 2025-12-23


Step 5: Record Payment
─────────────────────────────────────────
POST /api/invoices/:id/payments
{
  "amount": 2664000,
  "paymentMethod": "TRANSFER",
  "paymentDate": "2025-12-20",
  "referenceNumber": "TRF123456"
}
                │
                ▼
┌───────────────────────────────────────┐
│  InvoiceService.RecordPayment()      │
│  1. Validate invoice exists          │
│  2. Create payment record            │
│  3. Update invoice status to PAID    │
│  4. Update customer outstanding      │
│  5. Create cash transaction          │
│  Status: PAID                        │
└───────────────────────────────────────┘
                │
                ▼
     Payment Recorded
     Invoice Status: PAID
     Customer Outstanding: Updated
     Cash Transaction: Recorded

────────────────────────────────────────
WORKFLOW COMPLETE
Sales Order → Delivery → Invoice → Payment
All records auditable and traceable
────────────────────────────────────────
```

## Background Job System

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    cmd/worker (Background Jobs)                          │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│        Job Scheduler (Cron)             │
│                                         │
│  Every Day at 00:00                     │
│    ├─ Subscription Billing Job         │
│    ├─ Batch Expiry Monitoring Job      │
│    └─ Outstanding Recalculation Job    │
│                                         │
│  Every Hour                             │
│    └─ Notification Sending Job         │
└─────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│    Job: Subscription Billing            │
│                                         │
│  1. Find subscriptions due for billing  │
│  2. Calculate amount from custom price  │
│  3. Create payment record               │
│  4. Update subscription status          │
│  5. Apply grace period if not paid      │
│  6. Suspend tenant after grace period   │
│                                         │
│  Tenant Statuses:                       │
│  - TRIAL → ACTIVE (after payment)       │
│  - ACTIVE → PAST_DUE (missed payment)   │
│  - PAST_DUE → SUSPENDED (grace expired) │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│    Job: Batch Expiry Monitoring         │
│                                         │
│  1. Find batches expiring in 7 days    │
│  2. Check stock levels                  │
│  3. Generate alerts for warehouse       │
│  4. Update batch status if expired      │
│  5. Create notification records         │
│                                         │
│  Batch Statuses:                        │
│  - AVAILABLE → NEAR_EXPIRY (7 days)     │
│  - NEAR_EXPIRY → EXPIRED (0 days)       │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│    Job: Outstanding Recalculation       │
│                                         │
│  1. Calculate customer receivables      │
│  2. Calculate supplier payables         │
│  3. Identify overdue invoices           │
│  4. Update outstanding amounts          │
│  5. Generate aging reports              │
└─────────────────────────────────────────┘
```

## Technology Stack

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         TECHNOLOGY STACK                                 │
└─────────────────────────────────────────────────────────────────────────┘

Programming Language:
  - Go 1.25.4

Web Framework:
  - Gin (HTTP router and middleware)

Database:
  - PostgreSQL (production)
  - SQLite (development/testing)

ORM:
  - GORM (with PostgreSQL/SQLite drivers)

Authentication:
  - JWT (JSON Web Tokens)
  - bcrypt/argon2 (password hashing)

Validation:
  - go-playground/validator

Logging:
  - zap or logrus (structured logging)

Testing:
  - testify (assertions and mocks)
  - mockgen (mock generation)

Database Migrations:
  - golang-migrate

Background Jobs:
  - Cron scheduler
  - Job queue (in-memory or Redis)

API Documentation:
  - OpenAPI 3.0 (Swagger)

Containerization:
  - Docker
  - Docker Compose

CI/CD:
  - GitHub Actions or GitLab CI

Monitoring (Future):
  - Prometheus (metrics)
  - Grafana (dashboards)
  - ELK Stack (logs)
```

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      SECURITY LAYERS                                     │
└─────────────────────────────────────────────────────────────────────────┘

Layer 1: Network Security
─────────────────────────────────────────
  - HTTPS/TLS encryption
  - Firewall rules
  - Rate limiting
  - DDoS protection

Layer 2: Authentication
─────────────────────────────────────────
  - JWT tokens (24h expiry)
  - Refresh token mechanism
  - Password hashing (bcrypt/argon2)
  - Account lockout after failed attempts

Layer 3: Authorization
─────────────────────────────────────────
  - Role-Based Access Control (RBAC)
  - Per-tenant role assignment
  - Permission checking middleware
  - Principle of least privilege

Layer 4: Data Isolation
─────────────────────────────────────────
  - Tenant ID in all queries
  - Database row-level security
  - Separate schema per tenant (optional)
  - Audit logging

Layer 5: Input Validation
─────────────────────────────────────────
  - Request validation (go-playground/validator)
  - SQL injection prevention (parameterized queries)
  - XSS prevention (sanitize output)
  - CSRF protection

Layer 6: Sensitive Data
─────────────────────────────────────────
  - Encrypt at rest (database encryption)
  - Encrypt in transit (TLS)
  - Mask sensitive data in logs
  - Secure key management

Layer 7: Monitoring & Auditing
─────────────────────────────────────────
  - Access logs
  - Audit trail for sensitive operations
  - Anomaly detection
  - Security event alerts
```

---

This architecture provides a solid foundation for building a scalable, secure, and maintainable Multi-Tenant ERP system following Clean Architecture and DDD principles.
