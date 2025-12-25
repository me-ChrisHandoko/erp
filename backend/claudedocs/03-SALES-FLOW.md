# Module Group 3: Sales Flow (Revenue Cycle)

**Implementation Priority:** WEEK 4-5 (Core Business Flow)
**Dependencies:** Company Profile ✅, Master Data ✅ (Products, Customers, Warehouses)
**Modules:** Sales Order, Delivery, Invoice, Customer Payment

---

## Overview

This module group implements the **complete revenue cycle** from order to payment collection. This is the HIGHEST PRIORITY for MVP as it directly generates revenue.

### Business Flow
```
Sales Order (DRAFT → CONFIRMED)
  ↓
Delivery (PREPARED → DELIVERED → CONFIRMED)
  ↓
Invoice (UNPAID → PARTIAL → PAID)
  ↓
Customer Payment (cash, transfer, check/giro)
  ↓
Update Customer Outstanding
```

### Why Week 4-5 (Priority)?
1. **Revenue-generating flow** - core business operation
2. **Customer-facing** - directly impacts customer satisfaction
3. **Cash flow critical** - payment collection drives business viability
4. **Most complex** - involves inventory, pricing, tax, outstanding tracking

---

## Module 1: Sales Order Management

### Purpose
Record customer orders with items, pricing, tax calculation, and approval workflow.

### Key Features (MVP)
1. **Create/Edit Sales Order** - DRAFT status, add/remove items
2. **Pricing Calculation** - unit price × quantity, discounts, tax (PPN)
3. **Confirm Order** - DRAFT → CONFIRMED (locks order for fulfillment)
4. **Status Tracking** - DRAFT, CONFIRMED, COMPLETED, CANCELLED
5. **Auto-Number Generation** - configurable format from Company settings

### Database Models (Already Defined)

```go
type SalesOrder struct {
    ID               string
    TenantID         string
    SONumber         string          // Auto-generated: SO/001
    SODate           time.Time
    CustomerID       string
    Status           SalesOrderStatus // DRAFT, CONFIRMED, COMPLETED, CANCELLED
    Subtotal         decimal.Decimal  // Sum of items
    DiscountAmount   decimal.Decimal  // Header-level discount
    TaxAmount        decimal.Decimal  // PPN (11%)
    TotalAmount      decimal.Decimal  // Subtotal - Discount + Tax
    Notes            *string
    DeliveryAddress  *string
    DeliveryDate     *time.Time
    SalespersonID    *string
    ApprovedBy       *string
    ApprovedAt       *time.Time
    CancelledBy      *string
    CancelledAt      *time.Time
    CancellationNote *string

    // Relations
    Customer Customer
    Items    []SalesOrderItem
}

type SalesOrderItem struct {
    ID            string
    SalesOrderID  string
    ProductID     string
    ProductUnitID *string         // NULL = base unit
    Quantity      decimal.Decimal // In selected unit
    UnitPrice     decimal.Decimal // Price per unit
    DiscountPct   decimal.Decimal // Item discount %
    DiscountAmt   decimal.Decimal // Item discount amount
    Subtotal      decimal.Decimal // (Qty × Price) - Discount
    Notes         *string
}
```

**Enums:**
```go
type SalesOrderStatus string
const (
    SOStatusDraft     SalesOrderStatus = "DRAFT"
    SOStatusConfirmed SalesOrderStatus = "CONFIRMED"
    SOStatusCompleted SalesOrderStatus = "COMPLETED"
    SOStatusCancelled SalesOrderStatus = "CANCELLED"
)
```

---

### API Endpoints

#### 1. Create Sales Order
```http
POST /api/v1/sales-orders

Request:
{
  "soDate": "2025-01-20T00:00:00Z",
  "customerId": "clcust1",
  "deliveryAddress": "Jl. Pasar Raya No. 45, Jakarta",
  "deliveryDate": "2025-01-22T00:00:00Z",
  "notes": "Urgent order",
  "items": [
    {
      "productId": "clprod1",
      "productUnitId": "clunit1",      // KARUNG
      "quantity": "10.000",             // 10 KARUNG
      "unitPrice": "750000.00",         // Price per KARUNG
      "discountPct": "5.00",            // 5% discount
      "notes": "Good quality please"
    }
  ]
}

Response (201):
{
  "success": true,
  "data": {
    "id": "clso1",
    "soNumber": "SO/001",
    "soDate": "2025-01-20T00:00:00Z",
    "customer": {
      "id": "clcust1",
      "code": "CUST-001",
      "name": "Toko Sembako Jaya"
    },
    "status": "DRAFT",
    "items": [
      {
        "id": "clsoi1",
        "product": {
          "id": "clprod1",
          "code": "BRS-001",
          "name": "Beras Premium 5kg"
        },
        "unit": {
          "id": "clunit1",
          "unitName": "KARUNG",
          "conversionRate": "50.000"  // 1 KARUNG = 50 KG
        },
        "quantity": "10.000",
        "unitPrice": "750000.00",
        "discountPct": "5.00",
        "discountAmt": "37500.00",    // 5% of 750,000
        "subtotal": "7125000.00"      // (10 × 750,000) - 37,500
      }
    ],
    "subtotal": "7125000.00",
    "discountAmount": "0.00",
    "taxAmount": "783750.00",         // 11% PPN
    "totalAmount": "7908750.00",      // 7,125,000 + 783,750
    "deliveryAddress": "Jl. Pasar Raya No. 45, Jakarta",
    "deliveryDate": "2025-01-22T00:00:00Z",
    "notes": "Urgent order",
    "createdAt": "2025-01-20T10:00:00Z"
  }
}
```

#### 2. Confirm Sales Order
```http
POST /api/v1/sales-orders/:id/confirm

Business Logic:
1. Validate customer credit limit:
   if customer.CurrentOutstanding + so.TotalAmount > customer.CreditLimit:
       return error
2. Check stock availability (per item):
   - Calculate required base units: quantity × conversionRate
   - Check warehouse stock >= required quantity
   - If insufficient → return error with stock details
3. Update status: DRAFT → CONFIRMED
4. Lock order (no more edits to items/pricing)

Response (200):
{
  "success": true,
  "data": {
    "id": "clso1",
    "status": "CONFIRMED",
    "approvedBy": "cluser1",
    "approvedAt": "2025-01-20T11:00:00Z"
  }
}

Errors (400):
{
  "success": false,
  "error": {
    "code": "CREDIT_LIMIT_EXCEEDED",
    "message": "Customer credit limit exceeded. Outstanding: 5,000,000. Limit: 10,000,000. New total would be: 12,908,750."
  }
}

{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_STOCK",
    "message": "Insufficient stock for product BRS-001",
    "details": {
      "productCode": "BRS-001",
      "productName": "Beras Premium 5kg",
      "requiredQty": "500.000",      // 10 KARUNG × 50 KG
      "availableQty": "250.000",
      "unit": "KG"
    }
  }
}
```

#### 3. Cancel Sales Order
```http
POST /api/v1/sales-orders/:id/cancel

Request:
{
  "cancellationNote": "Customer changed mind"
}

Business Logic:
- Can only cancel if status = DRAFT or CONFIRMED
- Cannot cancel if delivery already exists
- Status: CONFIRMED → CANCELLED

Response (200):
{
  "success": true,
  "data": {
    "id": "clso1",
    "status": "CANCELLED",
    "cancelledBy": "cluser1",
    "cancelledAt": "2025-01-20T12:00:00Z",
    "cancellationNote": "Customer changed mind"
  }
}
```

#### 4. List & Search Sales Orders
```http
GET /api/v1/sales-orders
GET /api/v1/sales-orders?customerId=clcust1&status=CONFIRMED&fromDate=2025-01-01&toDate=2025-01-31

Query Params:
- customerId: string
- status: DRAFT | CONFIRMED | COMPLETED | CANCELLED
- fromDate: date
- toDate: date
- search: string (search in SO number, customer name)
- page, limit, sortBy, sortOrder

Response (200):
{
  "success": true,
  "data": [
    {
      "id": "clso1",
      "soNumber": "SO/001",
      "soDate": "2025-01-20T00:00:00Z",
      "customer": {
        "id": "clcust1",
        "code": "CUST-001",
        "name": "Toko Sembako Jaya"
      },
      "status": "CONFIRMED",
      "totalAmount": "7908750.00",
      "itemCount": 1,
      "createdAt": "2025-01-20T10:00:00Z"
    }
  ],
  "pagination": { /* ... */ }
}
```

---

### Business Rules

#### 1. Pricing Calculation
```go
// Item-level calculation
itemSubtotal = (quantity × unitPrice) - discountAmt
where discountAmt = quantity × unitPrice × (discountPct / 100)

// Order-level calculation
subtotal = sum(items.subtotal)
taxAmount = subtotal × (ppnRate / 100)  // From company.PPNRate
totalAmount = subtotal - discountAmount + taxAmount
```

#### 2. Stock Reservation (MVP: Manual Check)
```go
// When confirming SO, check stock availability
for each item in SO:
    requiredBaseQty = item.quantity × item.productUnit.conversionRate
    warehouseStock = getWarehouseStock(item.productId, defaultWarehouseId)

    if warehouseStock.quantity < requiredBaseQty:
        return error "Insufficient stock"

// Phase 2: Auto-reserve stock (create RESERVED inventory movements)
```

#### 3. Credit Limit Enforcement
```go
// Before confirming SO
customer := getCustomer(so.customerId)
projectedOutstanding := customer.CurrentOutstanding + so.TotalAmount

if projectedOutstanding > customer.CreditLimit:
    return error "Credit limit exceeded"

// Exception: OWNER/ADMIN can override (require approval)
```

#### 4. SO Number Generation
```go
// From company.SONumberFormat: "{PREFIX}/{NUMBER}"
// Example: "SO/001", "SO/002"

func GenerateSONumber(company *models.Company, tenantID string) string {
    // Get last SO number for this tenant
    var lastSO models.SalesOrder
    db.Where("tenant_id = ?", tenantID).
       Order("created_at DESC").
       First(&lastSO)

    nextNumber := 1
    if lastSO.ID != "" {
        // Extract number from last SO number
        nextNumber = extractNumber(lastSO.SONumber) + 1
    }

    // Format with zero-padding (001, 002, etc.)
    return company.SOPrefix + "/" + fmt.Sprintf("%03d", nextNumber)
}
```

---

## Module 2: Delivery Management (Simplified for MVP)

### Purpose
Record product deliveries to customers with batch tracking and proof of delivery.

### Key Features (MVP)
1. **Create Delivery from SO** - auto-fill items from sales order
2. **Batch Selection** - assign batches to items (for batch-tracked products)
3. **Status Tracking** - PREPARED → DELIVERED → CONFIRMED
4. **Proof of Delivery** - signature, photo, received by name
5. **Update Warehouse Stock** - deduct stock when delivery confirmed

### Database Model

```go
type Delivery struct {
    ID              string
    TenantID        string
    DeliveryNumber  string          // Auto-generated
    DeliveryDate    time.Time
    SalesOrderID    *string         // Optional: can deliver without SO (direct sales)
    CustomerID      string
    WarehouseID     string
    Status          DeliveryStatus  // PREPARED, DELIVERED, CONFIRMED, CANCELLED
    Notes           *string
    DriverName      *string
    VehicleNumber   *string
    DepartureTime   *time.Time
    ArrivalTime     *time.Time
    ReceivedBy      *string         // Customer contact who received
    ReceivedAt      *time.Time
    Signature       *string         // Base64 or URL
    PhotoProof      *string         // URL
    TTNKNumber      *string         // Resi JNE/Sicepat/JNT

    // Relations
    SalesOrder *SalesOrder
    Customer   Customer
    Warehouse  Warehouse
    Items      []DeliveryItem
}

type DeliveryItem struct {
    ID            string
    DeliveryID    string
    ProductID     string
    ProductUnitID *string
    BatchID       *string         // Required if product.IsBatchTracked = true
    Quantity      decimal.Decimal // In selected unit
    Notes         *string
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/deliveries                    # Create from SO or manual
POST   /api/v1/deliveries/:id/confirm        # Mark as delivered, deduct stock
GET    /api/v1/deliveries                    # List with filters
GET    /api/v1/deliveries/:id                # Details
PUT    /api/v1/deliveries/:id                # Update (only PREPARED status)
DELETE /api/v1/deliveries/:id                # Cancel
```

### Business Logic

```go
// Confirm delivery → Deduct stock
func (s *DeliveryService) ConfirmDelivery(ctx context.Context, tenantID, deliveryID string) error {
    delivery := getDelivery(deliveryID)

    // Validate status
    if delivery.Status != DeliveryStatusPrepared {
        return errors.New("can only confirm PREPARED deliveries")
    }

    tx := s.db.Begin()

    for _, item := range delivery.Items {
        // Calculate base quantity
        baseQty := item.Quantity × item.ProductUnit.ConversionRate

        // Deduct warehouse stock
        var whStock models.WarehouseStock
        tx.Where("warehouse_id = ? AND product_id = ?", delivery.WarehouseID, item.ProductID).First(&whStock)

        if whStock.Quantity.LessThan(baseQty) {
            tx.Rollback()
            return errors.New("insufficient stock for product " + item.Product.Code)
        }

        stockBefore := whStock.Quantity
        whStock.Quantity = whStock.Quantity.Sub(baseQty)
        tx.Save(&whStock)

        // Create inventory movement record
        movement := &models.InventoryMovement{
            TenantID:      tenantID,
            WarehouseID:   delivery.WarehouseID,
            ProductID:     item.ProductID,
            BatchID:       item.BatchID,
            MovementType:  "OUT",
            ReferenceType: "DELIVERY",
            ReferenceID:   deliveryID,
            Quantity:      baseQty.Neg(),  // Negative for outbound
            StockBefore:   stockBefore,
            StockAfter:    whStock.Quantity,
            MovementDate:  delivery.DeliveryDate,
        }
        tx.Create(movement)

        // If batch-tracked, deduct from batch
        if item.BatchID != nil {
            var batch models.ProductBatch
            tx.First(&batch, *item.BatchID)
            batch.Quantity = batch.Quantity.Sub(baseQty)
            if batch.Quantity.Equal(decimal.Zero) {
                batch.Status = models.BatchStatusSold
            }
            tx.Save(&batch)
        }
    }

    // Update delivery status
    delivery.Status = DeliveryStatusDelivered
    delivery.ReceivedAt = time.Now()
    tx.Save(delivery)

    // Update sales order status
    if delivery.SalesOrderID != nil {
        var so models.SalesOrder
        tx.First(&so, *delivery.SalesOrderID)
        // Check if all deliveries are confirmed
        allDelivered := checkAllDeliveriesConfirmed(tx, *delivery.SalesOrderID)
        if allDelivered {
            so.Status = models.SOStatusCompleted
            tx.Save(&so)
        }
    }

    tx.Commit()
    return nil
}
```

---

## Module 3: Invoice Management

### Purpose
Generate invoices from sales orders/deliveries and track payment status.

### Key Features (MVP)
1. **Generate Invoice from SO** - auto-fill items, pricing, tax
2. **Invoice Number Generation** - from company.InvoiceNumberFormat
3. **Payment Status** - UNPAID → PARTIAL → PAID → OVERDUE
4. **Due Date Calculation** - from customer.PaymentTerm
5. **Tax Invoice (Faktur Pajak)** - for PKP customers

### Database Model

```go
type Invoice struct {
    ID                string
    TenantID          string
    InvoiceNumber     string          // INV/001/01/2025
    InvoiceDate       time.Time
    DueDate           time.Time       // invoiceDate + customer.PaymentTerm
    SalesOrderID      *string
    CustomerID        string
    Status            InvoiceStatus   // UNPAID, PARTIAL, PAID, OVERDUE, CANCELLED
    Subtotal          decimal.Decimal
    DiscountAmount    decimal.Decimal
    TaxAmount         decimal.Decimal // PPN
    TotalAmount       decimal.Decimal
    PaidAmount        decimal.Decimal
    UnpaidAmount      decimal.Decimal // TotalAmount - PaidAmount
    TaxInvoiceNumber  *string         // Faktur Pajak number
    Notes             *string

    // Relations
    Customer Customer
    SalesOrder *SalesOrder
    Items    []InvoiceItem
    Payments []InvoicePayment
}

type InvoiceItem struct {
    ID         string
    InvoiceID  string
    ProductID  string
    Quantity   decimal.Decimal
    UnitPrice  decimal.Decimal
    Subtotal   decimal.Decimal
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/invoices                      # Generate from SO
GET    /api/v1/invoices                      # List with filters
GET    /api/v1/invoices/:id                  # Details
GET    /api/v1/invoices/:id/pdf              # Download PDF
POST   /api/v1/invoices/:id/send-email       # Email to customer
DELETE /api/v1/invoices/:id                  # Cancel (if UNPAID)
```

### Business Logic

```go
// Generate invoice from sales order
func (s *InvoiceService) GenerateFromSalesOrder(ctx context.Context, tenantID, soID string) (*models.Invoice, error) {
    so := getSalesOrder(soID)

    // Generate invoice number
    invoiceNumber := GenerateInvoiceNumber(company, tenantID)

    // Calculate due date
    dueDate := so.SODate.AddDate(0, 0, so.Customer.PaymentTerm)

    invoice := &models.Invoice{
        TenantID:       tenantID,
        InvoiceNumber:  invoiceNumber,
        InvoiceDate:    time.Now(),
        DueDate:        dueDate,
        SalesOrderID:   &soID,
        CustomerID:     so.CustomerID,
        Status:         models.InvoiceStatusUnpaid,
        Subtotal:       so.Subtotal,
        DiscountAmount: so.DiscountAmount,
        TaxAmount:      so.TaxAmount,
        TotalAmount:    so.TotalAmount,
        PaidAmount:     decimal.Zero,
        UnpaidAmount:   so.TotalAmount,
    }

    tx := s.db.Begin()

    // Create invoice
    tx.Create(invoice)

    // Create invoice items from SO items
    for _, soItem := range so.Items {
        invoiceItem := &models.InvoiceItem{
            InvoiceID: invoice.ID,
            ProductID: soItem.ProductID,
            Quantity:  soItem.Quantity,
            UnitPrice: soItem.UnitPrice,
            Subtotal:  soItem.Subtotal,
        }
        tx.Create(invoiceItem)
    }

    // Update customer outstanding
    var customer models.Customer
    tx.First(&customer, so.CustomerID)
    customer.CurrentOutstanding = customer.CurrentOutstanding.Add(invoice.TotalAmount)
    tx.Save(&customer)

    tx.Commit()

    return invoice, nil
}
```

---

## Module 4: Customer Payment

### Purpose
Record customer payments and apply to invoices, update outstanding amounts.

### Key Features (MVP)
1. **Record Payment** - cash, transfer, check/giro
2. **Apply to Invoices** - full or partial payment
3. **Update Invoice Status** - UNPAID → PARTIAL → PAID
4. **Update Customer Outstanding** - deduct from CurrentOutstanding
5. **Check/Giro Tracking** - for non-cash payments

### Database Model

```go
type InvoicePayment struct {
    ID            string
    TenantID      string
    InvoiceID     string
    PaymentDate   time.Time
    Amount        decimal.Decimal
    PaymentMethod PaymentMethod   // CASH, TRANSFER, CHECK, GIRO
    BankAccountID *string         // Company bank where payment received
    Reference     *string         // Transfer ref, check number
    CheckNumber   *string
    CheckDate     *time.Time
    CheckStatus   *CheckStatus    // ISSUED, CLEARED, BOUNCED, CANCELLED
    Notes         *string
}
```

---

### API Endpoints (Summary)

```http
POST   /api/v1/payments                      # Record payment
GET    /api/v1/payments                      # List payments
GET    /api/v1/payments/:id                  # Details
DELETE /api/v1/payments/:id                  # Void payment (if same day)
```

### Business Logic

```go
// Record payment
func (s *PaymentService) RecordPayment(ctx context.Context, tenantID string, req *RecordPaymentRequest) (*models.InvoicePayment, error) {
    invoice := getInvoice(req.InvoiceID)

    // Validate amount
    if req.Amount.GreaterThan(invoice.UnpaidAmount) {
        return nil, errors.New("payment amount exceeds unpaid amount")
    }

    tx := s.db.Begin()

    // Create payment record
    payment := &models.InvoicePayment{
        TenantID:      tenantID,
        InvoiceID:     req.InvoiceID,
        PaymentDate:   req.PaymentDate,
        Amount:        req.Amount,
        PaymentMethod: req.PaymentMethod,
        BankAccountID: req.BankAccountID,
        Reference:     req.Reference,
        Notes:         req.Notes,
    }
    tx.Create(payment)

    // Update invoice
    invoice.PaidAmount = invoice.PaidAmount.Add(req.Amount)
    invoice.UnpaidAmount = invoice.UnpaidAmount.Sub(req.Amount)

    if invoice.UnpaidAmount.Equal(decimal.Zero) {
        invoice.Status = models.InvoiceStatusPaid
    } else {
        invoice.Status = models.InvoiceStatusPartial
    }
    tx.Save(invoice)

    // Update customer outstanding
    var customer models.Customer
    tx.First(&customer, invoice.CustomerID)
    customer.CurrentOutstanding = customer.CurrentOutstanding.Sub(req.Amount)
    customer.LastTransactionAt = time.Now()
    tx.Save(&customer)

    tx.Commit()

    return payment, nil
}
```

---

## Implementation Priority (Week 4-5)

### Week 4: Sales Order & Delivery
**Day 1-2:** Sales Order CRUD + pricing calculation
**Day 3-4:** Delivery CRUD + stock deduction
**Day 5:** Integration testing

### Week 5: Invoice & Payment
**Day 1-2:** Invoice generation + PDF rendering
**Day 3-4:** Payment recording + outstanding updates
**Day 5:** End-to-end sales flow testing

---

## Testing Checklist

### Sales Order
- [ ] Create SO with multi-unit items
- [ ] Pricing calculation (item discount + header discount + tax)
- [ ] Confirm SO (credit limit check, stock check)
- [ ] Cancel SO
- [ ] Cannot confirm if insufficient stock
- [ ] Cannot confirm if credit limit exceeded

### Delivery
- [ ] Create delivery from SO
- [ ] Batch selection for batch-tracked products
- [ ] Confirm delivery → deducts stock
- [ ] Creates inventory movement record
- [ ] Updates SO status to COMPLETED when all delivered

### Invoice
- [ ] Generate invoice from SO
- [ ] Invoice number auto-generation
- [ ] Due date calculation from payment term
- [ ] Updates customer outstanding

### Payment
- [ ] Record full payment → invoice PAID
- [ ] Record partial payment → invoice PARTIAL
- [ ] Updates customer outstanding correctly
- [ ] Cannot overpay invoice

---

## Next Module Group

After completing Sales Flow (Week 4-5), proceed to:
**→ `04-PURCHASE-FLOW.md`** (Purchase Order → Goods Receipt → Supplier Payment)

---

## Summary

**Sales Flow** implements the complete revenue cycle:
1. ✅ Sales Order → Customer orders with pricing and tax
2. ✅ Delivery → Stock fulfillment with batch tracking
3. ✅ Invoice → Payment due tracking
4. ✅ Payment → Cash collection and outstanding updates

**Critical Business Rules:**
- Credit limit enforcement before SO confirmation
- Stock availability check before delivery
- Automatic customer outstanding updates
- Invoice due date calculation from payment terms

**Estimated Completion:** 2 weeks (10 days)
