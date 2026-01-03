# Frontend-Backend Gap Analysis Report

**Analysis Date:** December 29, 2025
**Analysis Scope:** Align backend API implementation with frontend menu structure
**Frontend Framework:** Next.js (TypeScript)
**Backend Framework:** Go + Gin + GORM

---

## Executive Summary

This report maps the 27 frontend menu items to their corresponding backend implementation status. The analysis reveals that **Master Data modules are fully implemented**, while **4 major operational modules (Inventory, Procurement, Sales, Finance) require complete implementation**.

### Implementation Status Overview

| Category | Total Items | Implemented | Partial | Not Implemented | Completion % |
|----------|-------------|-------------|---------|-----------------|--------------|
| Dashboard | 1 | 1 | 0 | 0 | 100% |
| Perusahaan (Company) | 3 | 3 | 0 | 0 | 100% |
| Master Data | 4 | 4 | 0 | 0 | 100% |
| Persediaan (Inventory) | 4 | 0 | 0 | 4 | 0% |
| Pembelian (Procurement) | 4 | 0 | 0 | 4 | 0% |
| Penjualan (Sales) | 4 | 0 | 0 | 4 | 0% |
| Keuangan (Finance) | 4 | 0 | 0 | 4 | 0% |
| Pengaturan (Settings) | 3 | 1 | 1 | 1 | 33% |
| **TOTAL** | **27** | **9** | **1** | **17** | **37%** |

---

## 1. Dashboard

### 1.1 Dashboard (/)
- **Frontend Path:** `/`
- **Permission:** None (public)
- **Backend Status:** ✅ IMPLEMENTED
- **Backend Route:** Health check endpoint exists
- **Service:** Basic health monitoring
- **Implementation:** `/api/health`
- **Notes:** Dashboard may need analytics/reporting endpoints

---

## 2. Perusahaan (Company Management)

### 2.1 Profil Perusahaan (Company Profile)
- **Frontend Path:** `/company/profile`
- **Permission:** `company:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:id` - Get company profile
  - `PUT /api/companies/:id` - Update company profile
- **Handler:** `company_handler.go`
- **Service:** `company/company_service.go`
- **Features:**
  - Company profile CRUD
  - Tax configuration (NPWP, PKP, PPN)
  - Invoice numbering format
  - Operating hours

### 2.2 Rekening Bank (Bank Accounts)
- **Frontend Path:** `/company/banks`
- **Permission:** `company:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:id/banks` - List bank accounts
  - `POST /api/companies/:id/banks` - Create bank account
  - `PUT /api/companies/:id/banks/:bankId` - Update bank account
  - `DELETE /api/companies/:id/banks/:bankId` - Delete bank account
- **Handler:** `company_handler.go`
- **Service:** `company/company_service.go`
- **Features:**
  - Multiple bank accounts per company
  - Account details (bank name, number, holder)
  - Active/inactive status

### 2.3 Tim & Pengguna (Team & Users)
- **Frontend Path:** `/company/team`
- **Permission:** `users:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:companyId/users` - List company users
  - `POST /api/companies/:companyId/users/invite` - Invite user
  - `PUT /api/companies/:companyId/users/:userId` - Update user role
  - `DELETE /api/companies/:companyId/users/:userId` - Remove user
- **Handler:** `company_user_handler.go`
- **Service:** `company/multi_company_service.go`
- **Features:**
  - Multi-company user access
  - Role-based permissions (OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF)
  - User invitation system
  - Active/inactive user management

---

## 3. Master Data

### 3.1 Pelanggan (Customers)
- **Frontend Path:** `/master/customers`
- **Permission:** `customers:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:companyId/customers` - List customers (with pagination, filtering, sorting)
  - `POST /api/companies/:companyId/customers` - Create customer
  - `GET /api/companies/:companyId/customers/:id` - Get customer by ID
  - `PUT /api/companies/:companyId/customers/:id` - Update customer
  - `DELETE /api/companies/:companyId/customers/:id` - Soft delete customer
- **Handler:** `customer_handler.go`
- **Service:** `customer/customer_service.go`
- **Features:**
  - Customer master data (code, name, contact info)
  - Tax info (NPWP, PKP status)
  - Credit limit and payment terms
  - Outstanding balance tracking
  - Overdue amount monitoring
  - Advanced filtering (by type, city, province, PKP, overdue)
  - Pagination and sorting

### 3.2 Pemasok (Suppliers)
- **Frontend Path:** `/master/suppliers`
- **Permission:** `suppliers:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:companyId/suppliers` - List suppliers (with pagination, filtering, sorting)
  - `POST /api/companies/:companyId/suppliers` - Create supplier
  - `GET /api/companies/:companyId/suppliers/:id` - Get supplier by ID
  - `PUT /api/companies/:companyId/suppliers/:id` - Update supplier
  - `DELETE /api/companies/:companyId/suppliers/:id` - Soft delete supplier
- **Handler:** `supplier_handler.go`
- **Service:** `supplier/supplier_service.go`
- **Features:**
  - Supplier master data (code, name, contact info)
  - Tax info (NPWP, PKP status)
  - Credit limit and payment terms
  - Outstanding balance tracking
  - Overdue amount monitoring
  - Advanced filtering (by type, city, province, PKP, overdue)
  - Pagination and sorting

### 3.3 Produk (Products)
- **Frontend Path:** `/master/products`
- **Permission:** `products:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:companyId/products` - List products (with pagination, filtering, sorting)
  - `POST /api/companies/:companyId/products` - Create product
  - `GET /api/companies/:companyId/products/:id` - Get product by ID
  - `PUT /api/companies/:companyId/products/:id` - Update product
  - `DELETE /api/companies/:companyId/products/:id` - Soft delete product
- **Handler:** `product_handler.go`
- **Service:** `product/product_service.go`
- **Features:**
  - Product master data (code, name, category)
  - Multi-unit system (base unit + conversion units)
  - Batch/lot tracking support
  - Perishable goods tracking (expiry dates)
  - Barcode management (unique across products and units)
  - Pricing (base cost, base price per unit)
  - Minimum stock levels
  - Active/inactive status
  - Advanced filtering (by category, type, active status)
  - Stock validation on deletion

### 3.4 Gudang (Warehouses)
- **Frontend Path:** `/master/warehouses`
- **Permission:** `warehouses:read`
- **Backend Status:** ✅ FULLY IMPLEMENTED
- **Backend Routes:**
  - `GET /api/companies/:companyId/warehouses` - List warehouses (with pagination, filtering, sorting)
  - `POST /api/companies/:companyId/warehouses` - Create warehouse
  - `GET /api/companies/:companyId/warehouses/:id` - Get warehouse by ID
  - `PUT /api/companies/:companyId/warehouses/:id` - Update warehouse
  - `DELETE /api/companies/:companyId/warehouses/:id` - Soft delete warehouse
  - `GET /api/companies/:companyId/warehouses/stocks` - List warehouse stocks
  - `PUT /api/companies/:companyId/warehouses/stocks/:id` - Update stock settings
- **Handler:** `warehouse_handler.go`
- **Service:** `warehouse/warehouse_service.go`
- **Features:**
  - Warehouse master data (code, name, address)
  - Warehouse types (MAIN, BRANCH, CONSIGNMENT, TRANSIT)
  - Capacity management
  - Manager assignment
  - Multi-warehouse stock tracking
  - Stock location within warehouse
  - Minimum/maximum stock levels
  - Low stock and zero stock filtering
  - Advanced filtering (by type, city, province, manager, active status)

---

## 4. Persediaan (Inventory Management)

### 4.1 Stok Barang (Stock Items)
- **Frontend Path:** `/inventory/stock`
- **Permission:** `inventory:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service Directory:** `/internal/service/inventory/`
  - **Handler:** `/internal/handler/inventory_handler.go`
  - **Routes:**
    - `GET /api/companies/:companyId/inventory/stock` - List all stock items
    - `GET /api/companies/:companyId/inventory/stock/:productId` - Get product stock across warehouses
    - `GET /api/companies/:companyId/inventory/stock/:productId/batches` - Get batch details
    - `GET /api/companies/:companyId/inventory/movements` - List inventory movements
- **Features Needed:**
  - Stock overview across all warehouses
  - Real-time stock levels
  - Batch/lot tracking with expiry dates
  - FIFO/FEFO inventory valuation
  - Stock movement history
  - Stock alerts (low stock, near expiry, expired)
  - Stock aging report

### 4.2 Transfer Gudang (Warehouse Transfers)
- **Frontend Path:** `/inventory/transfers`
- **Permission:** `inventory:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `inventory/stock_transfer_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/inventory/transfers` - List transfers
    - `POST /api/companies/:companyId/inventory/transfers` - Create transfer
    - `GET /api/companies/:companyId/inventory/transfers/:id` - Get transfer details
    - `PUT /api/companies/:companyId/inventory/transfers/:id/approve` - Approve transfer
    - `PUT /api/companies/:companyId/inventory/transfers/:id/receive` - Receive transfer
    - `PUT /api/companies/:companyId/inventory/transfers/:id/cancel` - Cancel transfer
- **Features Needed:**
  - Inter-warehouse stock transfers
  - Transfer request workflow
  - Approval process
  - In-transit tracking
  - Batch-level transfer
  - Transfer status (PENDING, APPROVED, IN_TRANSIT, RECEIVED, CANCELLED)
  - Automatic stock adjustment on source/destination

### 4.3 Stock Opname (Stock Taking)
- **Frontend Path:** `/inventory/stock-opname`
- **Permission:** `inventory:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `inventory/stock_opname_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/inventory/stock-opname` - List stock takes
    - `POST /api/companies/:companyId/inventory/stock-opname` - Create stock take
    - `GET /api/companies/:companyId/inventory/stock-opname/:id` - Get stock take details
    - `PUT /api/companies/:companyId/inventory/stock-opname/:id/items` - Update counted items
    - `PUT /api/companies/:companyId/inventory/stock-opname/:id/finalize` - Finalize stock take
    - `PUT /api/companies/:companyId/inventory/stock-opname/:id/cancel` - Cancel stock take
- **Features Needed:**
  - Physical inventory count
  - Variance analysis (system vs physical count)
  - Batch-level counting
  - Count verification workflow
  - Automatic stock adjustment on finalization
  - Stock opname status (DRAFT, IN_PROGRESS, COMPLETED, CANCELLED)
  - Difference reporting (over/short)
  - Approval workflow for adjustments

### 4.4 Penyesuaian (Stock Adjustments)
- **Frontend Path:** `/inventory/adjustments`
- **Permission:** `inventory:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `inventory/adjustment_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/inventory/adjustments` - List adjustments
    - `POST /api/companies/:companyId/inventory/adjustments` - Create adjustment
    - `GET /api/companies/:companyId/inventory/adjustments/:id` - Get adjustment details
    - `PUT /api/companies/:companyId/inventory/adjustments/:id/approve` - Approve adjustment
    - `PUT /api/companies/:companyId/inventory/adjustments/:id/cancel` - Cancel adjustment
- **Features Needed:**
  - Manual stock adjustments
  - Adjustment types (DAMAGE, LOSS, FOUND, EXPIRED, OTHER)
  - Batch-level adjustments
  - Reason tracking
  - Approval workflow
  - Automatic inventory movement creation
  - Adjustment status (PENDING, APPROVED, CANCELLED)

---

## 5. Pembelian (Procurement)

### 5.1 Purchase Order
- **Frontend Path:** `/purchase/orders`
- **Permission:** `purchase:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service Directory:** `/internal/service/purchase/`
  - **Service:** `purchase/purchase_order_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/purchase/orders` - List POs
    - `POST /api/companies/:companyId/purchase/orders` - Create PO
    - `GET /api/companies/:companyId/purchase/orders/:id` - Get PO details
    - `PUT /api/companies/:companyId/purchase/orders/:id` - Update PO
    - `PUT /api/companies/:companyId/purchase/orders/:id/approve` - Approve PO
    - `PUT /api/companies/:companyId/purchase/orders/:id/cancel` - Cancel PO
- **Features Needed:**
  - Purchase order creation
  - Supplier selection
  - Multi-line items with units
  - Price negotiation tracking
  - Tax calculation (PPN)
  - Approval workflow
  - PO status (DRAFT, SUBMITTED, APPROVED, PARTIAL, COMPLETED, CANCELLED)
  - Expected delivery date
  - PO numbering with configurable format

### 5.2 Penerimaan Barang (Goods Receipt)
- **Frontend Path:** `/purchase/receipts`
- **Permission:** `purchase:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `purchase/goods_receipt_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/purchase/receipts` - List GRNs
    - `POST /api/companies/:companyId/purchase/receipts` - Create GRN
    - `GET /api/companies/:companyId/purchase/receipts/:id` - Get GRN details
    - `PUT /api/companies/:companyId/purchase/receipts/:id/inspect` - Quality inspection
    - `PUT /api/companies/:companyId/purchase/receipts/:id/accept` - Accept goods
    - `PUT /api/companies/:companyId/purchase/receipts/:id/reject` - Reject goods
- **Features Needed:**
  - Goods receipt against PO
  - Warehouse selection
  - Quality inspection
  - Accepted/rejected quantity tracking
  - Batch/lot number assignment
  - Expiry date tracking (for perishables)
  - Automatic stock increment
  - GRN status (PENDING, RECEIVED, INSPECTED, ACCEPTED, REJECTED, PARTIAL)
  - Difference handling (over/short delivery)

### 5.3 Faktur Pembelian (Purchase Invoices)
- **Frontend Path:** `/purchase/invoices`
- **Permission:** `purchase:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `purchase/purchase_invoice_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/purchase/invoices` - List purchase invoices
    - `POST /api/companies/:companyId/purchase/invoices` - Create invoice
    - `GET /api/companies/:companyId/purchase/invoices/:id` - Get invoice details
    - `PUT /api/companies/:companyId/purchase/invoices/:id` - Update invoice
    - `PUT /api/companies/:companyId/purchase/invoices/:id/approve` - Approve for payment
- **Features Needed:**
  - Invoice creation from GRN
  - Supplier invoice matching
  - Tax invoice (Faktur Pajak) tracking
  - Tax calculation and validation
  - Payment terms tracking
  - Due date calculation
  - Invoice status (DRAFT, SUBMITTED, APPROVED, PAID, OVERDUE)
  - Automatic supplier outstanding update

### 5.4 Pembayaran (Supplier Payments)
- **Frontend Path:** `/purchase/payments`
- **Permission:** `finance:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `purchase/supplier_payment_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/purchase/payments` - List supplier payments
    - `POST /api/companies/:companyId/purchase/payments` - Record payment
    - `GET /api/companies/:companyId/purchase/payments/:id` - Get payment details
    - `PUT /api/companies/:companyId/purchase/payments/:id/cancel` - Cancel payment
- **Features Needed:**
  - Payment against invoices
  - Multiple payment methods (CASH, TRANSFER, CHECK, GIRO)
  - Partial payment support
  - Check/Giro tracking
  - Bank account selection
  - Payment status tracking
  - Automatic supplier outstanding reduction
  - Payment receipt generation

---

## 6. Penjualan (Sales)

### 6.1 Sales Order
- **Frontend Path:** `/sales/orders`
- **Permission:** `sales:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service Directory:** `/internal/service/sales/`
  - **Service:** `sales/sales_order_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/sales/orders` - List SOs
    - `POST /api/companies/:companyId/sales/orders` - Create SO
    - `GET /api/companies/:companyId/sales/orders/:id` - Get SO details
    - `PUT /api/companies/:companyId/sales/orders/:id` - Update SO
    - `PUT /api/companies/:companyId/sales/orders/:id/approve` - Approve SO
    - `PUT /api/companies/:companyId/sales/orders/:id/cancel` - Cancel SO
- **Features Needed:**
  - Sales order creation
  - Customer selection
  - Multi-line items with units
  - Pricing from product master
  - Discount management
  - Tax calculation (PPN)
  - Stock reservation
  - SO status (DRAFT, CONFIRMED, PARTIAL, COMPLETED, CANCELLED)
  - Expected delivery date
  - SO numbering with configurable format

### 6.2 Pengiriman (Deliveries)
- **Frontend Path:** `/sales/deliveries`
- **Permission:** `sales:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `sales/delivery_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/sales/deliveries` - List deliveries
    - `POST /api/companies/:companyId/sales/deliveries` - Create delivery
    - `GET /api/companies/:companyId/sales/deliveries/:id` - Get delivery details
    - `PUT /api/companies/:companyId/sales/deliveries/:id/dispatch` - Dispatch delivery
    - `PUT /api/companies/:companyId/sales/deliveries/:id/complete` - Complete delivery
    - `PUT /api/companies/:companyId/sales/deliveries/:id/cancel` - Cancel delivery
- **Features Needed:**
  - Delivery creation from SO
  - Warehouse selection for picking
  - Batch/lot selection (FIFO/FEFO)
  - Driver and vehicle assignment
  - Logistics info (departure/arrival times)
  - Proof of delivery (POD)
  - TTNK support for expeditions
  - Delivery status (PREPARED, IN_TRANSIT, DELIVERED, CONFIRMED, CANCELLED)
  - Automatic stock decrement
  - Delivery note generation

### 6.3 Faktur Penjualan (Sales Invoices)
- **Frontend Path:** `/sales/invoices`
- **Permission:** `sales:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `sales/sales_invoice_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/sales/invoices` - List sales invoices
    - `POST /api/companies/:companyId/sales/invoices` - Create invoice
    - `GET /api/companies/:companyId/sales/invoices/:id` - Get invoice details
    - `PUT /api/companies/:companyId/sales/invoices/:id` - Update invoice
    - `PUT /api/companies/:companyId/sales/invoices/:id/send` - Send to customer
- **Features Needed:**
  - Invoice creation from delivery
  - Tax invoice (Faktur Pajak) generation
  - Tax calculation and formatting
  - Payment terms and due date
  - Invoice numbering with configurable format
  - Invoice status (DRAFT, SENT, PAID, PARTIAL, OVERDUE, CANCELLED)
  - Automatic customer outstanding update
  - Invoice PDF generation

### 6.4 Penerimaan Kas (Cash Receipts)
- **Frontend Path:** `/sales/receipts`
- **Permission:** `finance:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `sales/cash_receipt_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/sales/receipts` - List cash receipts
    - `POST /api/companies/:companyId/sales/receipts` - Record receipt
    - `GET /api/companies/:companyId/sales/receipts/:id` - Get receipt details
    - `PUT /api/companies/:companyId/sales/receipts/:id/cancel` - Cancel receipt
- **Features Needed:**
  - Payment against invoices
  - Multiple payment methods (CASH, TRANSFER, CHECK, GIRO, CREDIT_CARD, QRIS)
  - Partial payment support
  - Check/Giro tracking
  - Bank account selection
  - Receipt status tracking
  - Automatic customer outstanding reduction
  - Receipt printing

---

## 7. Keuangan (Finance)

### 7.1 Jurnal Umum (General Journal)
- **Frontend Path:** `/finance/journal`
- **Permission:** `finance:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service Directory:** `/internal/service/finance/`
  - **Service:** `finance/journal_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/finance/journal` - List journal entries
    - `POST /api/companies/:companyId/finance/journal` - Create manual entry
    - `GET /api/companies/:companyId/finance/journal/:id` - Get journal details
    - `PUT /api/companies/:companyId/finance/journal/:id/post` - Post to ledger
    - `PUT /api/companies/:companyId/finance/journal/:id/void` - Void entry
- **Features Needed:**
  - Manual journal entry creation
  - Automatic journal from transactions (sales, purchase, payments)
  - Chart of accounts integration
  - Debit/credit validation (balanced entries)
  - Journal status (DRAFT, POSTED, VOID)
  - Journal numbering
  - Period closing controls
  - Reversal entries

### 7.2 Kas & Bank (Cash & Bank)
- **Frontend Path:** `/finance/cash-bank`
- **Permission:** `finance:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `finance/cash_transaction_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/finance/cash-transactions` - List cash transactions
    - `POST /api/companies/:companyId/finance/cash-transactions` - Record transaction
    - `GET /api/companies/:companyId/finance/cash-transactions/:id` - Get transaction details
    - `GET /api/companies/:companyId/finance/cash-book` - Get cash book (Buku Kas)
    - `GET /api/companies/:companyId/finance/bank-reconciliation` - Bank reconciliation
- **Features Needed:**
  - Cash transaction recording (IN/OUT)
  - Bank account selection
  - Transaction categories
  - Running balance calculation
  - Cash book (Buku Kas) report
  - Bank reconciliation
  - Check/Giro clearing tracking
  - Transaction status (PENDING, CLEARED, BOUNCED, CANCELLED)

### 7.3 Biaya (Expenses)
- **Frontend Path:** `/finance/expenses`
- **Permission:** `finance:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `finance/expense_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/finance/expenses` - List expenses
    - `POST /api/companies/:companyId/finance/expenses` - Create expense
    - `GET /api/companies/:companyId/finance/expenses/:id` - Get expense details
    - `PUT /api/companies/:companyId/finance/expenses/:id/approve` - Approve expense
    - `PUT /api/companies/:companyId/finance/expenses/:id/pay` - Record payment
- **Features Needed:**
  - Expense recording
  - Expense categories
  - Vendor/supplier assignment
  - Approval workflow
  - Payment tracking
  - Receipt attachment
  - Expense status (DRAFT, SUBMITTED, APPROVED, PAID, REJECTED)
  - Tax deductibility tracking

### 7.4 Laporan (Reports)
- **Frontend Path:** `/finance/reports`
- **Permission:** `finance:read`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `finance/report_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/finance/reports/balance-sheet` - Balance sheet
    - `GET /api/companies/:companyId/finance/reports/income-statement` - P&L
    - `GET /api/companies/:companyId/finance/reports/cash-flow` - Cash flow
    - `GET /api/companies/:companyId/finance/reports/receivables` - AR aging
    - `GET /api/companies/:companyId/finance/reports/payables` - AP aging
    - `GET /api/companies/:companyId/finance/reports/tax` - Tax reports
- **Features Needed:**
  - Balance sheet generation
  - Income statement (P&L)
  - Cash flow statement
  - Trial balance
  - AR aging report
  - AP aging report
  - Tax reports (PPN, PPh)
  - Customizable date ranges
  - PDF/Excel export

---

## 8. Pengaturan (Settings)

### 8.1 Roles & Permissions
- **Frontend Path:** `/settings/roles`
- **Permission:** `settings:write`
- **Backend Status:** ⚠️ PARTIAL IMPLEMENTATION
- **Current Implementation:**
  - Service exists: `permission/permission_service.go`
  - Basic role management via `UserTenant.role`
  - Hardcoded roles: OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF
- **Missing Implementation:**
  - Custom role creation
  - Granular permission management
  - Permission groups
  - Role cloning
- **Required Routes:**
  - `GET /api/companies/:companyId/roles` - List custom roles
  - `POST /api/companies/:companyId/roles` - Create custom role
  - `PUT /api/companies/:companyId/roles/:id` - Update role permissions
  - `DELETE /api/companies/:companyId/roles/:id` - Delete custom role
  - `GET /api/permissions` - List all available permissions

### 8.2 Konfigurasi Sistem (System Configuration)
- **Frontend Path:** `/settings/system`
- **Permission:** `settings:write`
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `settings/system_config_service.go`
  - **Routes:**
    - `GET /api/companies/:companyId/settings` - Get all settings
    - `PUT /api/companies/:companyId/settings` - Update settings
    - `GET /api/companies/:companyId/settings/numbering` - Get numbering formats
    - `PUT /api/companies/:companyId/settings/numbering` - Update numbering formats
- **Features Needed:**
  - System-wide settings management
  - Document numbering configuration (SO, PO, Invoice, etc.)
  - Tax configuration
  - Default values
  - Business rules configuration
  - Approval workflow settings
  - Notification preferences

### 8.3 Preferensi (User Preferences)
- **Frontend Path:** `/settings/preferences`
- **Permission:** None (user-specific)
- **Backend Status:** ❌ NOT IMPLEMENTED
- **Required Implementation:**
  - **Service:** `user/preference_service.go`
  - **Routes:**
    - `GET /api/users/me/preferences` - Get user preferences
    - `PUT /api/users/me/preferences` - Update preferences
- **Features Needed:**
  - Language preference
  - Timezone settings
  - Date/time format
  - Number format (decimal separator)
  - Default company selection
  - Theme preferences
  - Notification settings
  - Dashboard layout

---

## Implementation Priority Matrix

### HIGH PRIORITY (Critical for MVP)
These modules are essential for basic ERP operations and have the highest business impact.

1. **Sales Module** (Priority: 🔴 CRITICAL)
   - **Justification:** Core revenue-generating functionality
   - **Dependencies:** Master Data (✅ Complete)
   - **Effort:** 3-4 weeks
   - **Implementation Order:**
     1. Sales Order (foundation)
     2. Delivery (fulfillment)
     3. Sales Invoice (billing)
     4. Cash Receipt (payment)

2. **Procurement Module** (Priority: 🔴 CRITICAL)
   - **Justification:** Core operational functionality for procurement
   - **Dependencies:** Master Data (✅ Complete)
   - **Effort:** 3-4 weeks
   - **Implementation Order:**
     1. Purchase Order (foundation)
     2. Goods Receipt (receiving)
     3. Purchase Invoice (payables)
     4. Supplier Payment (payment)

3. **Inventory Module** (Priority: 🔴 CRITICAL)
   - **Justification:** Real-time stock tracking and movement
   - **Dependencies:** Master Data (✅ Complete), Sales, Procurement
   - **Effort:** 2-3 weeks
   - **Implementation Order:**
     1. Stock Items (overview)
     2. Stock Adjustments (manual corrections)
     3. Warehouse Transfers (inter-warehouse)
     4. Stock Opname (physical count)

### MEDIUM PRIORITY (Important for Full Functionality)
These modules enhance the system but aren't blocking for basic operations.

4. **Finance Module** (Priority: 🟡 HIGH)
   - **Justification:** Financial reporting and compliance
   - **Dependencies:** Sales, Procurement fully functional
   - **Effort:** 4-5 weeks
   - **Implementation Order:**
     1. Cash & Bank (cash book)
     2. Expenses (operating expenses)
     3. Journal (accounting entries)
     4. Reports (financial statements)

### LOW PRIORITY (Nice to Have)
These can be implemented after core functionality is stable.

5. **Advanced Settings** (Priority: 🟢 MEDIUM)
   - **Justification:** Improves UX and customization
   - **Dependencies:** None
   - **Effort:** 1-2 weeks
   - **Implementation Order:**
     1. System Configuration (numbering, defaults)
     2. Custom Roles & Permissions (RBAC enhancement)
     3. User Preferences (personalization)

---

## Recommended Implementation Roadmap

### Phase 1: Core Transactions (Weeks 1-8)
**Objective:** Enable basic buy-sell-stock operations

**Week 1-2: Sales Foundation**
- Sales Order service + handler
- Basic SO workflow (DRAFT → CONFIRMED → COMPLETED)
- Unit tests for SO creation

**Week 3-4: Procurement Foundation**
- Purchase Order service + handler
- Basic PO workflow (DRAFT → APPROVED → COMPLETED)
- Unit tests for PO creation

**Week 5-6: Inventory Basics**
- Stock overview service
- Stock adjustment service
- Inventory movement tracking

**Week 7-8: Fulfillment & Receipt**
- Delivery service (from SO)
- Goods Receipt service (from PO)
- Stock integration

### Phase 2: Financial Integration (Weeks 9-14)
**Objective:** Enable invoicing and payment tracking

**Week 9-10: Invoicing**
- Sales Invoice service
- Purchase Invoice service
- Outstanding balance updates

**Week 11-12: Payments**
- Cash Receipt service
- Supplier Payment service
- Payment reconciliation

**Week 13-14: Cash Book**
- Cash transaction service
- Buku Kas report
- Bank reconciliation basics

### Phase 3: Advanced Features (Weeks 15-18)
**Objective:** Complete operational capabilities

**Week 15-16: Inventory Advanced**
- Warehouse Transfer service
- Stock Opname service
- Batch/lot expiry tracking

**Week 17-18: Finance Advanced**
- Expense management
- General Journal
- Basic financial reports

### Phase 4: Polish & Settings (Weeks 19-20)
**Objective:** System configuration and UX enhancements

**Week 19: System Settings**
- Document numbering configuration
- Business rules setup
- Default values management

**Week 20: User Experience**
- Custom roles & permissions
- User preferences
- Dashboard analytics

---

## Technical Implementation Guidelines

### 1. Service Layer Pattern
All new services should follow the existing pattern:

```go
// Service structure
type [Module]Service struct {
    db *gorm.DB
}

// Constructor
func New[Module]Service(db *gorm.DB) *[Module]Service {
    return &[Module]Service{db: db}
}

// Core CRUD operations
func (s *[Module]Service) Create[Entity](ctx context.Context, companyID string, req *dto.Create[Entity]Request) (*models.[Entity], error)
func (s *[Module]Service) List[Entities](ctx context.Context, companyID string, query *dto.[Entity]ListQuery) (*dto.[Entity]ListResponse, error)
func (s *[Module]Service) Get[Entity]ByID(ctx context.Context, companyID, entityID string) (*models.[Entity], error)
func (s *[Module]Service) Update[Entity](ctx context.Context, companyID, entityID string, req *dto.Update[Entity]Request) (*models.[Entity], error)
func (s *[Module]Service) Delete[Entity](ctx context.Context, companyID, entityID string) error
```

### 2. Multi-Company Isolation
**CRITICAL:** Every query MUST filter by `company_id`:

```go
// ✅ CORRECT
baseQuery := s.db.WithContext(ctx).Model(&models.Entity{}).
    Where("company_id = ?", companyID)

// ❌ WRONG - Security vulnerability
baseQuery := s.db.WithContext(ctx).Model(&models.Entity{})
```

### 3. Handler Pattern
All handlers should follow RESTful conventions:

```go
func (h *[Module]Handler) List[Entities](c *gin.Context) {
    companyID := c.Param("companyId")

    // Parse query parameters
    var query dto.[Entity]ListQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Call service
    result, err := h.service.List[Entities](c.Request.Context(), companyID, &query)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(200, result)
}
```

### 4. DTO Patterns
Request/Response DTOs for API boundaries:

```go
// Create request
type Create[Entity]Request struct {
    Code   string  `json:"code" binding:"required"`
    Name   string  `json:"name" binding:"required"`
    // ... other fields
}

// Update request (all fields optional)
type Update[Entity]Request struct {
    Code   *string `json:"code"`
    Name   *string `json:"name"`
    // ... other fields
}

// Response
type [Entity]Response struct {
    ID        string    `json:"id"`
    Code      string    `json:"code"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// List response with pagination
type [Entity]ListResponse struct {
    [Entities] [][Entity]Response `json:"[entities]"`
    TotalCount int64              `json:"totalCount"`
    Page       int                `json:"page"`
    PageSize   int                `json:"pageSize"`
    TotalPages int                `json:"totalPages"`
}
```

### 5. Error Handling
Use custom error types for consistent API responses:

```go
import pkgerrors "backend/pkg/errors"

// Validation errors
return nil, pkgerrors.NewBadRequestError("invalid input")

// Not found
return nil, pkgerrors.NewNotFoundError("entity not found")

// Business logic errors
return nil, pkgerrors.NewConflictError("entity already exists")

// Authorization errors
return nil, pkgerrors.NewUnauthorizedError("insufficient permissions")
```

### 6. Testing Requirements
Every service MUST have:

1. **Unit Tests** (service layer)
   - Test all CRUD operations
   - Test validation logic
   - Test business rules
   - Test error conditions

2. **Integration Tests** (multi-company isolation)
   - Verify company A cannot access company B data
   - Test cross-company resource linking prevention
   - Verify proper 404 responses (not 403)

Example test structure:
```go
func TestCreateEntity_Success(t *testing.T) {
    db := testutil.SetupTestDatabase(t)
    defer testutil.TeardownTestDatabase(t, db)

    service := NewEntityService(db)
    company := testutil.CreateTestCompany(t, db, "Test Company")

    req := &dto.CreateEntityRequest{
        Code: "TEST-001",
        Name: "Test Entity",
    }

    result, err := service.CreateEntity(context.Background(), company.ID, req)

    require.NoError(t, err)
    assert.Equal(t, "TEST-001", result.Code)
    assert.Equal(t, company.ID, result.CompanyID)
}
```

### 7. Database Migrations
For new tables, create migration files:

```go
// cmd/migrate/phase2_inventory.go
func migrateInventoryTables(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.InventoryMovement{},
        &models.StockTransfer{},
        &models.StockOpname{},
        &models.StockAdjustment{},
    )
}
```

---

## Risk Assessment & Mitigation

### High Risk Areas

1. **Stock Movement Integrity**
   - **Risk:** Race conditions in concurrent stock updates
   - **Mitigation:** Database transactions, row-level locking, atomic operations
   - **Testing:** Concurrent operation tests (see `bank_race_test.go` example)

2. **Multi-Company Data Leakage**
   - **Risk:** Cross-company data access
   - **Mitigation:** Mandatory `company_id` filtering, integration tests
   - **Testing:** Multi-company isolation test suite (✅ Already implemented)

3. **Outstanding Balance Accuracy**
   - **Risk:** Mismatch between invoices and outstanding amounts
   - **Mitigation:** Transactional updates, reconciliation jobs, audit trail
   - **Testing:** Balance verification tests

4. **Batch/Lot Tracking**
   - **Risk:** FIFO/FEFO violations, expired product shipment
   - **Mitigation:** Strict batch selection logic, expiry date validation
   - **Testing:** Batch selection algorithm tests

### Medium Risk Areas

1. **Number Sequence Gaps**
   - **Risk:** Duplicate or skipped numbers in SO/PO/Invoice
   - **Mitigation:** Database-level sequence or atomic increment
   - **Testing:** Concurrent number generation tests

2. **Workflow State Transitions**
   - **Risk:** Invalid state changes (e.g., CANCELLED → APPROVED)
   - **Mitigation:** State machine validation, transition rules
   - **Testing:** State transition validation tests

3. **Tax Calculation Accuracy**
   - **Risk:** Incorrect PPN calculations
   - **Mitigation:** Decimal precision, rounding rules, tax rate configuration
   - **Testing:** Tax calculation unit tests

---

## Performance Considerations

### Optimization Strategies

1. **Database Indexing**
   - Composite indexes on `[company_id, code]`
   - Date indexes for reporting queries
   - Status indexes for workflow filtering
   - Foreign key indexes

2. **Query Optimization**
   - Use pagination (default: 20 items per page)
   - Implement search debouncing on frontend
   - Lazy load related entities
   - Cache frequently accessed data (company settings)

3. **Batch Operations**
   - Bulk insert for multiple items
   - Transaction batching for imports
   - Async processing for reports

4. **Caching Strategy**
   - Company configuration (Redis/in-memory)
   - Product pricing (short TTL)
   - User permissions (session-based)

---

## Success Metrics

### Implementation Completion Metrics

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Backend API Coverage | 37% | 100% | 🔴 |
| Service Layer Tests | 70+ tests | 200+ tests | 🟡 |
| Integration Test Coverage | 23 tests | 100+ tests | 🔴 |
| Frontend-Backend Alignment | 9/27 items | 27/27 items | 🔴 |

### Quality Metrics (Post-Implementation)

- **Test Coverage:** ≥80% for service layer
- **API Response Time:** <200ms for CRUD operations
- **Multi-Company Isolation:** 100% (zero data leakage)
- **Code Review Pass Rate:** ≥95%
- **Bug Density:** <5 bugs per 1000 lines

---

## Appendix

### A. Models Already Defined

The following models exist in `/backend/models/` and are ready for use:

**Phase 0 Models (Available):**
- ✅ `SalesOrder` + `SalesOrderItem`
- ✅ `Delivery` + `DeliveryItem`
- ✅ `Invoice` + `InvoiceItem`
- ✅ `PurchaseOrder` + `PurchaseOrderItem`
- ✅ `GoodsReceipt` + `GoodsReceiptItem`
- ✅ `SupplierPayment` + `SupplierPaymentItem`
- ✅ `CashTransaction`
- ✅ `InventoryMovement`
- ✅ `StockTransfer` + `StockTransferItem`
- ✅ `StockOpname` + `StockOpnameItem`
- ✅ `ProductBatch`

**Implementation Status:**
- Models: ✅ Complete (all Phase 0 models defined)
- Services: ⚠️ 33% (4/12 core modules)
- Handlers: ⚠️ 33% (4/12 core modules)
- Routes: ⚠️ 33% (4/12 core modules)

### B. Reference Implementation Examples

For implementation guidance, refer to existing fully-implemented modules:

1. **Product Module** (`internal/service/product/`)
   - Multi-unit system
   - Validation patterns
   - Barcode uniqueness checks
   - Stock integration

2. **Customer Module** (`internal/service/customer/`)
   - Outstanding balance tracking
   - Filtering and pagination
   - Soft delete patterns

3. **Warehouse Module** (`internal/service/warehouse/`)
   - Stock management
   - Multi-company isolation (see recent security fix)
   - Joined table queries

4. **Company Module** (`internal/service/company/`)
   - Nested resource handling (bank accounts)
   - Configuration management
   - Tax settings

### C. Environment Setup

**Required for Development:**
```bash
# Install Go 1.25.4
go version

# Install dependencies
cd /Users/christianhandoko/Development/work/erp/backend
go mod tidy

# Run database migrations
go run cmd/migrate/main.go

# Run tests
go test ./...

# Run server
go run main.go
```

**Database Setup:**
- Development: SQLite (`dev.db`)
- Production: PostgreSQL
- All migrations in `db/migration.go`

### D. Useful Commands Reference

```bash
# Run all tests
go test ./... -v

# Run specific service tests
go test ./internal/service/product/... -v

# Run integration tests only
go test ./test/integration/... -v

# Check test coverage
go test ./... -cover

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

---

## Conclusion

This gap analysis reveals that while the **Master Data foundation is solid (100% complete)**, the **operational modules require significant implementation effort**. The recommended phased approach prioritizes high-impact modules (Sales, Procurement, Inventory) before financial integration and advanced features.

**Key Takeaways:**
1. ✅ Strong foundation: Authentication, multi-company, and master data fully implemented
2. ⚠️ Core operational gap: 4 major modules (16 menu items) require complete implementation
3. 🎯 Clear roadmap: 20-week implementation plan with defined priorities
4. 🔒 Security focus: Multi-company isolation already proven with comprehensive test suite
5. 📊 Quality standards: Established patterns for services, handlers, DTOs, and testing

**Estimated Total Effort:** 16-20 weeks for full implementation across 4-5 developers working in parallel on different modules.

---

**Document Version:** 1.0
**Last Updated:** December 29, 2025
**Author:** Claude Code Analysis
**Status:** Complete
