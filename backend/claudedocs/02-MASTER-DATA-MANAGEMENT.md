# Module Group 2: Master Data Management

**Implementation Priority:** WEEK 2-3 (Foundation Data)
**Dependencies:** Company Profile ✅ (from Week 1)
**Modules:** Product Management, Customer Management, Supplier Management, Warehouse Management

---

## Overview

This module group provides CRUD operations for **master data** - the foundational entities required for all business transactions. Cannot create sales orders without products and customers, or purchase orders without products and suppliers.

### Business Context
- **Products:** Items sold/purchased with multi-unit support (KARTON = 24 PCS), batch/lot tracking for perishables
- **Customers:** Buyers with payment terms, credit limits, outstanding receivables tracking
- **Suppliers:** Vendors with payment terms, outstanding payables tracking, product relationships
- **Warehouses:** Stock locations with inventory levels, multi-warehouse support

### Why Week 2-3?
1. Required before ANY transaction can be recorded
2. Complex domain logic (multi-unit conversions, batch tracking, outstanding calculations)
3. High data volume expected (hundreds of products, customers)
4. Critical for inventory accuracy

---

## Module 1: Product Management

### Purpose
Manage product catalog with multi-unit support, batch/lot tracking for perishable items, and pricing management.

### Key Features (MVP)
1. **Basic Product CRUD** - code, name, category, base unit, pricing
2. **Multi-Unit Management** - define unit conversions (1 KARTON = 24 PCS)
3. **Batch Tracking Setup** - enable batch tracking for perishable items
4. **Simple Pricing** - base cost, base price per unit (customer-specific pricing in Phase 2)
5. **Barcode Support** - per product and per unit
6. **Search & Filters** - by code, name, category, active status

### Database Models (Already Defined)

```go
// models/product.go
type Product struct {
    ID             string
    TenantID       string          // Multi-tenant isolation
    Code           string          // SKU (unique per tenant)
    Name           string
    Category       *string
    BaseUnit       string          // Smallest unit (PCS, KG, LITER)
    BaseCost       decimal.Decimal // Purchase cost in base unit
    BasePrice      decimal.Decimal // Selling price in base unit
    MinimumStock   decimal.Decimal // Reorder level
    Description    *string
    Barcode        *string
    IsBatchTracked bool            // Require batch/lot tracking
    IsPerishable   bool            // Has expiry date
    IsActive       bool
    CreatedAt      time.Time
    UpdatedAt      time.Time

    // Relations
    Units              []ProductUnit     // Multi-unit conversions
    Batches            []ProductBatch    // Batch/lot tracking
    ProductSuppliers   []ProductSupplier // Supplier relationships
    WarehouseStocks    []WarehouseStock  // Stock per warehouse
}

type ProductUnit struct {
    ID             string
    ProductID      string
    UnitName       string          // "KARTON", "LUSIN", "KOLI"
    ConversionRate decimal.Decimal // 1 KARTON = 24 PCS
    IsBaseUnit     bool            // false (base unit is in Product.BaseUnit)
    BuyPrice       *decimal.Decimal // Purchase price per this unit
    SellPrice      *decimal.Decimal // Selling price per this unit
    Barcode        *string         // Barcode for this unit
    SKU            *string         // SKU for this unit
    Weight         *decimal.Decimal // Weight per unit (kg)
    Volume         *decimal.Decimal // Volume per unit (m³)
    Description    *string
    IsActive       bool
}

type ProductBatch struct {
    ID               string
    BatchNumber      string          // e.g., "BATCH-2025-001"
    ProductID        string
    WarehouseStockID string
    ManufactureDate  *time.Time
    ExpiryDate       *time.Time      // CRITICAL for perishables
    Quantity         decimal.Decimal // Current batch quantity
    SupplierID       *string         // Source supplier
    GoodsReceiptID   *string         // Linked to receipt
    ReceiptDate      time.Time       // When received
    Status           BatchStatus     // AVAILABLE, RESERVED, EXPIRED, DAMAGED
    QualityStatus    *string         // GOOD, DAMAGED, QUARANTINE
    ReferenceNumber  *string         // Supplier's batch number
    Notes            *string
}

type ProductSupplier struct {
    ID            string
    ProductID     string
    SupplierID    string
    SupplierPrice decimal.Decimal // Purchase price from this supplier
    LeadTime      int             // Days
    IsPrimary     bool            // Preferred supplier
}
```

---

### API Endpoints

#### 1. List Products
```http
GET /api/v1/products
GET /api/v1/products?search=beras&category=sembako&isActive=true&page=1&limit=20

Query Parameters:
  - search: string (search in code, name, barcode)
  - category: string
  - isActive: boolean
  - isBatchTracked: boolean
  - isPerishable: boolean
  - page: int (default 1)
  - limit: int (default 20, max 100)
  - sortBy: string (code, name, createdAt)
  - sortOrder: string (asc, desc)

Response (200 OK):
{
  "success": true,
  "data": [
    {
      "id": "clprod1",
      "code": "BRS-001",
      "name": "Beras Premium 5kg",
      "category": "Sembako",
      "baseUnit": "KG",
      "baseCost": "12000.00",
      "basePrice": "15000.00",
      "minimumStock": "50.000",
      "barcode": "8991234567890",
      "isBatchTracked": true,
      "isPerishable": false,
      "isActive": true,
      "units": [
        {
          "id": "clunit1",
          "unitName": "KARUNG",
          "conversionRate": "50.000",
          "buyPrice": "600000.00",
          "sellPrice": "750000.00",
          "barcode": "8991234567891"
        }
      ],
      "currentStock": {
        "total": "250.000",
        "warehouses": [
          {
            "warehouseId": "clwh1",
            "warehouseName": "Gudang Utama",
            "quantity": "250.000"
          }
        ]
      },
      "createdAt": "2025-01-15T08:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

#### 2. Get Product Details
```http
GET /api/v1/products/:id

Response (200 OK):
{
  "success": true,
  "data": {
    "id": "clprod1",
    "code": "BRS-001",
    "name": "Beras Premium 5kg",
    "category": "Sembako",
    "baseUnit": "KG",
    "baseCost": "12000.00",
    "basePrice": "15000.00",
    "minimumStock": "50.000",
    "description": "Beras premium kualitas terbaik",
    "barcode": "8991234567890",
    "isBatchTracked": true,
    "isPerishable": false,
    "isActive": true,
    "units": [
      {
        "id": "clunit1",
        "unitName": "KARUNG",
        "conversionRate": "50.000",
        "isBaseUnit": false,
        "buyPrice": "600000.00",
        "sellPrice": "750000.00",
        "barcode": "8991234567891",
        "sku": "BRS-001-KARUNG",
        "weight": "50.500",
        "isActive": true
      }
    ],
    "suppliers": [
      {
        "id": "clps1",
        "supplierId": "clsupp1",
        "supplierName": "PT Beras Indonesia",
        "supplierPrice": "11500.00",
        "leadTime": 7,
        "isPrimary": true
      }
    ],
    "warehouseStocks": [
      {
        "id": "clws1",
        "warehouseId": "clwh1",
        "warehouseName": "Gudang Utama",
        "quantity": "250.000",
        "minimumStock": "50.000",
        "maximumStock": "500.000",
        "lastCountDate": "2025-01-15T10:00:00Z"
      }
    ],
    "batches": [
      {
        "id": "clbatch1",
        "batchNumber": "BATCH-2025-001",
        "warehouseId": "clwh1",
        "manufactureDate": null,
        "expiryDate": null,
        "quantity": "250.000",
        "status": "AVAILABLE",
        "qualityStatus": "GOOD",
        "receiptDate": "2025-01-10T00:00:00Z"
      }
    ],
    "createdAt": "2025-01-15T08:00:00Z",
    "updatedAt": "2025-01-15T08:00:00Z"
  }
}
```

#### 3. Create Product
```http
POST /api/v1/products

Request Body:
{
  "code": "BRS-002",
  "name": "Beras Premium 10kg",
  "category": "Sembako",
  "baseUnit": "KG",
  "baseCost": "12000.00",
  "basePrice": "15000.00",
  "minimumStock": "100.000",
  "description": "Beras premium 10kg",
  "barcode": "8991234567892",
  "isBatchTracked": true,
  "isPerishable": false,
  "units": [
    {
      "unitName": "KARUNG",
      "conversionRate": "50.000",
      "buyPrice": "600000.00",
      "sellPrice": "750000.00",
      "barcode": "8991234567893"
    }
  ]
}

Response (201 Created):
{
  "success": true,
  "data": { /* Full product details */ }
}

Validation Errors (400):
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "code",
        "message": "Product code already exists"
      },
      {
        "field": "basePrice",
        "message": "Base price must be greater than 0"
      }
    ]
  }
}
```

#### 4. Update Product
```http
PUT /api/v1/products/:id

Request Body:
{
  "name": "Beras Premium 10kg Special",
  "basePrice": "16000.00",
  "minimumStock": "120.000"
}

Response (200 OK):
{
  "success": true,
  "data": { /* Updated product */ }
}
```

#### 5. Delete Product (Soft Delete)
```http
DELETE /api/v1/products/:id

Response (200 OK):
{
  "success": true,
  "message": "Product deactivated successfully"
}

Business Logic:
- Soft delete: isActive = false
- Cannot delete if product has:
  - Pending sales orders
  - Stock in any warehouse (must transfer out first)
  - Outstanding invoices
```

#### 6. Manage Product Units
```http
POST /api/v1/products/:productId/units
PUT /api/v1/products/:productId/units/:unitId
DELETE /api/v1/products/:productId/units/:unitId
```

#### 7. Manage Product Suppliers
```http
POST /api/v1/products/:productId/suppliers
PUT /api/v1/products/:productId/suppliers/:supplierId
DELETE /api/v1/products/:productId/suppliers/:supplierId
```

---

### Business Logic

**File:** `internal/service/product/product_service.go`

```go
// Key business rules
func (s *ProductService) CreateProduct(ctx context.Context, tenantID string, req *CreateProductRequest) (*models.Product, error) {
    // 1. Validate unique code per tenant
    var existing models.Product
    err := s.db.Where("tenant_id = ? AND code = ?", tenantID, req.Code).First(&existing).Error
    if err == nil {
        return nil, errors.New("product code already exists")
    }

    // 2. Validate base price >= base cost
    if req.BasePrice.LessThan(req.BaseCost) {
        return nil, errors.New("base price must be greater than or equal to base cost")
    }

    // 3. Create product
    product := &models.Product{
        TenantID:       tenantID,
        Code:           req.Code,
        Name:           req.Name,
        Category:       req.Category,
        BaseUnit:       req.BaseUnit,
        BaseCost:       req.BaseCost,
        BasePrice:      req.BasePrice,
        MinimumStock:   req.MinimumStock,
        Description:    req.Description,
        Barcode:        req.Barcode,
        IsBatchTracked: req.IsBatchTracked,
        IsPerishable:   req.IsPerishable,
        IsActive:       true,
    }

    // Start transaction
    tx := s.db.Begin()

    // Create product
    if err := tx.Create(product).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 4. Create base unit entry
    baseUnit := &models.ProductUnit{
        ProductID:      product.ID,
        UnitName:       req.BaseUnit,
        ConversionRate: decimal.NewFromInt(1), // Base unit conversion = 1
        IsBaseUnit:     true,
        BuyPrice:       &req.BaseCost,
        SellPrice:      &req.BasePrice,
        IsActive:       true,
    }
    if err := tx.Create(baseUnit).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // 5. Create additional units
    for _, unitReq := range req.Units {
        unit := &models.ProductUnit{
            ProductID:      product.ID,
            UnitName:       unitReq.UnitName,
            ConversionRate: unitReq.ConversionRate,
            IsBaseUnit:     false,
            BuyPrice:       unitReq.BuyPrice,
            SellPrice:      unitReq.SellPrice,
            Barcode:        unitReq.Barcode,
            IsActive:       true,
        }
        if err := tx.Create(unit).Error; err != nil {
            tx.Rollback()
            return nil, err
        }
    }

    // 6. Initialize warehouse stocks (zero stock)
    var warehouses []models.Warehouse
    tx.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&warehouses)
    for _, wh := range warehouses {
        whStock := &models.WarehouseStock{
            WarehouseID:  wh.ID,
            ProductID:    product.ID,
            Quantity:     decimal.Zero,
            MinimumStock: req.MinimumStock,
        }
        tx.Create(whStock)
    }

    tx.Commit()

    // Reload with relations
    return s.GetProductByID(ctx, tenantID, product.ID)
}
```

---

## Module 2: Customer Management

### Purpose
Manage customer database with payment terms, credit limits, and outstanding receivables tracking.

### Key Features (MVP)
1. **Basic Customer CRUD** - code, name, type, contact info
2. **Payment Terms** - days (0 = cash, 30 = net 30)
3. **Credit Limit** - maximum allowed outstanding
4. **Outstanding Tracking** - current receivables, overdue amounts
5. **Customer Types** - RETAIL, WHOLESALE, DISTRIBUTOR
6. **Search & Filters** - by code, name, city, type

### Database Model

```go
type Customer struct {
    ID                 string
    TenantID           string
    Code               string          // Unique per tenant
    Name               string
    Type               *string         // RETAIL, WHOLESALE, DISTRIBUTOR
    Phone              *string
    Email              *string
    Address            *string
    City               *string
    Province           *string
    PostalCode         *string
    NPWP               *string         // Tax ID
    IsPKP              bool            // Taxable entity
    ContactPerson      *string
    ContactPhone       *string
    PaymentTerm        int             // Days (0 = cash)
    CreditLimit        decimal.Decimal // Maximum outstanding
    CurrentOutstanding decimal.Decimal // Current receivables
    OverdueAmount      decimal.Decimal // Past-due amount
    LastTransactionAt  *time.Time
    Notes              *string
    IsActive           bool
}
```

### API Endpoints (Summary)

```http
GET    /api/v1/customers           # List with filters
GET    /api/v1/customers/:id       # Get details + outstanding summary
POST   /api/v1/customers           # Create
PUT    /api/v1/customers/:id       # Update
DELETE /api/v1/customers/:id       # Soft delete
GET    /api/v1/customers/:id/outstanding  # Outstanding details (invoices, payments)
```

### Business Rules

1. **Credit Limit Enforcement:**
   ```go
   // Before creating sales order/invoice
   if customer.CurrentOutstanding + newInvoiceTotal > customer.CreditLimit {
       return errors.New("customer credit limit exceeded")
   }
   ```

2. **Outstanding Calculation:**
   ```go
   // Update on invoice creation
   customer.CurrentOutstanding += invoice.TotalAmount

   // Update on payment
   customer.CurrentOutstanding -= payment.Amount

   // Calculate overdue (run daily job)
   overdueInvoices := invoices.Where("due_date < ? AND status != PAID", now)
   customer.OverdueAmount = sum(overdueInvoices.UnpaidAmount)
   ```

3. **Payment Term Validation:**
   - Must be >= 0
   - 0 = cash only (require full payment before delivery)
   - > 0 = credit (allow invoice with due date)

---

## Module 3: Supplier Management

### Purpose
Manage supplier database with payment terms, outstanding payables tracking, and product relationships.

### Key Features (MVP)
- Same as Customer Management but for suppliers (payables instead of receivables)
- Product-Supplier relationships (which supplier supplies which product)
- Lead time tracking

### Database Model

```go
type Supplier struct {
    ID                 string
    TenantID           string
    Code               string
    Name               string
    Type               *string         // MANUFACTURER, DISTRIBUTOR, WHOLESALER
    Phone              *string
    Email              *string
    Address            *string
    City               *string
    Province           *string
    NPWP               *string
    IsPKP              bool
    ContactPerson      *string
    ContactPhone       *string
    PaymentTerm        int             // Days
    CreditLimit        decimal.Decimal // Maximum payable
    CurrentOutstanding decimal.Decimal // Current payables
    OverdueAmount      decimal.Decimal // Past-due amount
    LastTransactionAt  *time.Time
    Notes              *string
    IsActive           bool
}
```

### API Endpoints (Summary)

```http
GET    /api/v1/suppliers           # List
GET    /api/v1/suppliers/:id       # Details
POST   /api/v1/suppliers           # Create
PUT    /api/v1/suppliers/:id       # Update
DELETE /api/v1/suppliers/:id       # Soft delete
GET    /api/v1/suppliers/:id/products    # Products from this supplier
```

---

## Module 4: Warehouse Management

### Purpose
Manage warehouse locations and initialize stock tracking for products.

### Key Features (MVP)
1. **Basic Warehouse CRUD** - code, name, type, address
2. **Warehouse Types** - MAIN, BRANCH, CONSIGNMENT, TRANSIT
3. **Stock Initialization** - set initial stock per product
4. **Stock Location** - within-warehouse location (e.g., "RAK-A-01")

### Database Model

```go
type Warehouse struct {
    ID             string
    TenantID       string
    Code           string          // Unique per tenant
    Name           string
    Type           WarehouseType   // MAIN, BRANCH, CONSIGNMENT, TRANSIT
    Address        *string
    City           *string
    Province       *string
    Phone          *string
    Manager        *string         // Manager name
    ManagerContact *string
    IsActive       bool
}

type WarehouseStock struct {
    ID              string
    WarehouseID     string
    ProductID       string
    Quantity        decimal.Decimal // Current stock (in base unit)
    MinimumStock    decimal.Decimal // Reorder level
    MaximumStock    decimal.Decimal // Maximum capacity
    Location        *string         // e.g., "RAK-A-01", "ZONE-B"
    LastCountDate   *time.Time      // Last physical count
    LastCountQty    *decimal.Decimal
}
```

### API Endpoints (Summary)

```http
GET    /api/v1/warehouses                    # List
GET    /api/v1/warehouses/:id                # Details
POST   /api/v1/warehouses                    # Create
PUT    /api/v1/warehouses/:id                # Update
DELETE /api/v1/warehouses/:id                # Soft delete
GET    /api/v1/warehouses/:id/stocks         # Stock levels per product
POST   /api/v1/warehouses/:id/stocks/init    # Initialize stock (opening balance)
```

### Business Logic

```go
// Initialize stock for new warehouse
func (s *WarehouseService) InitializeStocks(ctx context.Context, tenantID, warehouseID string) error {
    // Get all active products
    var products []models.Product
    s.db.Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&products)

    // Create warehouse stock entries with zero quantity
    for _, product := range products {
        whStock := &models.WarehouseStock{
            WarehouseID:  warehouseID,
            ProductID:    product.ID,
            Quantity:     decimal.Zero,
            MinimumStock: product.MinimumStock,
        }
        s.db.Create(whStock)
    }

    return nil
}

// Set opening stock (manual adjustment)
func (s *WarehouseService) SetOpeningStock(ctx context.Context, tenantID, warehouseID, productID string, qty decimal.Decimal) error {
    var whStock models.WarehouseStock
    err := s.db.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&whStock).Error

    if err != nil {
        // Create if doesn't exist
        whStock = models.WarehouseStock{
            WarehouseID:  warehouseID,
            ProductID:    productID,
            Quantity:     qty,
        }
        return s.db.Create(&whStock).Error
    }

    // Update existing
    whStock.Quantity = qty
    return s.db.Save(&whStock).Error
}
```

---

## Implementation Priority (Week 2-3)

### Week 2: Products & Customers
**Day 1-3: Product Management**
- Product CRUD with multi-unit support
- Barcode management
- Product-supplier relationships
- Stock initialization

**Day 4-5: Customer Management**
- Customer CRUD
- Payment terms & credit limits
- Outstanding tracking setup

### Week 3: Suppliers & Warehouses
**Day 1-2: Supplier Management**
- Supplier CRUD
- Product-supplier relationships
- Outstanding tracking

**Day 3-4: Warehouse Management**
- Warehouse CRUD
- Stock initialization
- Location management

**Day 5: Integration Testing**
- Multi-tenant isolation tests
- CRUD operations for all modules
- Relationship validations

---

## Testing Checklist

### Products
- [ ] Create product with multi-unit
- [ ] Search products (by code, name, barcode)
- [ ] Update product pricing
- [ ] Cannot delete product with stock
- [ ] Batch tracking enabled/disabled
- [ ] Unit conversion calculations
- [ ] Product code uniqueness per tenant

### Customers
- [ ] Create customer with payment terms
- [ ] Credit limit enforcement
- [ ] Outstanding calculation (manual test with mock invoices)
- [ ] Cannot delete customer with outstanding
- [ ] Search by name, city, type

### Suppliers
- [ ] Create supplier
- [ ] Link products to supplier
- [ ] Outstanding tracking
- [ ] Search functionality

### Warehouses
- [ ] Create warehouse
- [ ] Initialize stocks for all products
- [ ] Set opening stock
- [ ] Stock location management

### Multi-Tenancy
- [ ] Tenant A cannot see Tenant B's products/customers/suppliers/warehouses
- [ ] Code uniqueness per tenant (same code allowed in different tenants)

---

## Next Module Group

After completing Master Data Management (Week 2-3), proceed to:
**→ `03-SALES-FLOW.md`** (Sales Order → Delivery → Invoice → Payment)

---

## Summary

**Master Data Management** provides:
1. ✅ Product catalog with multi-unit support and batch tracking
2. ✅ Customer database with credit management
3. ✅ Supplier database with product relationships
4. ✅ Warehouse locations with stock tracking setup

**Dependencies for Next Modules:**
- Sales orders need products and customers
- Purchase orders need products and suppliers
- Inventory movements need products and warehouses
- Invoices need customer payment terms and credit limits

**Estimated Completion:** 2 weeks (10 days)
