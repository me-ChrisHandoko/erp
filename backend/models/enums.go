// Package models - Enum type definitions
// All enum values ensure database compatibility and type safety
package models

// UserRole - Dual-tier permission system for multi-company support
// Tier 1 (tenant_users): OWNER, TENANT_ADMIN - Superuser access to all companies within tenant
// Tier 2 (user_company_roles): ADMIN, FINANCE, SALES, WAREHOUSE, STAFF - Per-company granular access
type UserRole string

const (
	// Tier 1: Tenant-level roles (superuser - full access to all companies in tenant)
	UserRoleOwner       UserRole = "OWNER"        // Full control: tenant, all companies, billing, subscription
	UserRoleTenantAdmin UserRole = "TENANT_ADMIN" // Tenant admin: full operational control across all companies

	// Tier 2: Company-level roles (per-company granular access)
	UserRoleAdmin     UserRole = "ADMIN"     // Company admin: full operational control within specific company
	UserRoleFinance   UserRole = "FINANCE"   // Finance-focused access within specific company
	UserRoleSales     UserRole = "SALES"     // Sales-focused access within specific company
	UserRoleWarehouse UserRole = "WAREHOUSE" // Inventory/warehouse-focused access within specific company
	UserRoleStaff     UserRole = "STAFF"     // General operational access within specific company
)

// IsTenantLevel returns true if the role is a tenant-level superuser role (Tier 1)
func (r UserRole) IsTenantLevel() bool {
	return r == UserRoleOwner || r == UserRoleTenantAdmin
}

// IsCompanyLevel returns true if the role is a company-level role (Tier 2)
func (r UserRole) IsCompanyLevel() bool {
	switch r {
	case UserRoleAdmin, UserRoleFinance, UserRoleSales, UserRoleWarehouse, UserRoleStaff:
		return true
	default:
		return false
	}
}

// IsValid checks if the role is a valid UserRole
func (r UserRole) IsValid() bool {
	return r.IsTenantLevel() || r.IsCompanyLevel()
}

// String returns the string representation of UserRole
func (r UserRole) String() string {
	return string(r)
}

// TenantStatus - Subscription lifecycle states
type TenantStatus string

const (
	TenantStatusTrial     TenantStatus = "TRIAL"      // 14 days trial
	TenantStatusActive    TenantStatus = "ACTIVE"     // Paid & active
	TenantStatusSuspended TenantStatus = "SUSPENDED"  // Payment failed, grace period
	TenantStatusCancelled TenantStatus = "CANCELLED"  // Subscription cancelled
	TenantStatusExpired   TenantStatus = "EXPIRED"    // Trial expired, not paid
)

// SubscriptionStatus - Billing status
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "ACTIVE"
	SubscriptionStatusPastDue   SubscriptionStatus = "PAST_DUE"
	SubscriptionStatusCancelled SubscriptionStatus = "CANCELLED"
	SubscriptionStatusExpired   SubscriptionStatus = "EXPIRED"
)

// SubscriptionPaymentStatus - Payment tracking
type SubscriptionPaymentStatus string

const (
	SubscriptionPaymentStatusPending   SubscriptionPaymentStatus = "PENDING"
	SubscriptionPaymentStatusPaid      SubscriptionPaymentStatus = "PAID"
	SubscriptionPaymentStatusFailed    SubscriptionPaymentStatus = "FAILED"
	SubscriptionPaymentStatusRefunded  SubscriptionPaymentStatus = "REFUNDED"
	SubscriptionPaymentStatusCancelled SubscriptionPaymentStatus = "CANCELLED"
)

// WarehouseType - Warehouse classification
type WarehouseType string

const (
	WarehouseTypeMain        WarehouseType = "MAIN"        // Gudang utama/pusat
	WarehouseTypeBranch      WarehouseType = "BRANCH"      // Cabang
	WarehouseTypeConsignment WarehouseType = "CONSIGNMENT" // Titipan di customer
	WarehouseTypeTransit     WarehouseType = "TRANSIT"     // Gudang transit/antara
)

// BatchStatus - Inventory batch lifecycle
type BatchStatus string

const (
	BatchStatusAvailable BatchStatus = "AVAILABLE" // Available for sale
	BatchStatusReserved  BatchStatus = "RESERVED"  // Reserved for sales order
	BatchStatusExpired   BatchStatus = "EXPIRED"   // Past expiry date
	BatchStatusDamaged   BatchStatus = "DAMAGED"   // Damaged/defective
	BatchStatusRecalled  BatchStatus = "RECALLED"  // Product recall
	BatchStatusSold      BatchStatus = "SOLD"      // Fully sold out
)

// SalesOrderStatus - Simplified SO workflow (PHASE 0)
type SalesOrderStatus string

const (
	SalesOrderStatusDraft     SalesOrderStatus = "DRAFT"     // Belum dikonfirmasi
	SalesOrderStatusConfirmed SalesOrderStatus = "CONFIRMED" // Dikonfirmasi, ready proses
	SalesOrderStatusCompleted SalesOrderStatus = "COMPLETED" // Selesai (delivered & invoiced)
	SalesOrderStatusCancelled SalesOrderStatus = "CANCELLED" // Dibatalkan
)

// PurchaseOrderStatus - Simplified PO workflow (PHASE 0)
type PurchaseOrderStatus string

const (
	PurchaseOrderStatusDraft     PurchaseOrderStatus = "DRAFT"     // Belum dikonfirmasi
	PurchaseOrderStatusConfirmed PurchaseOrderStatus = "CONFIRMED" // Dikonfirmasi, menunggu barang
	PurchaseOrderStatusCompleted PurchaseOrderStatus = "COMPLETED" // Selesai (barang diterima)
	PurchaseOrderStatusCancelled PurchaseOrderStatus = "CANCELLED" // Dibatalkan
)

// PaymentStatus - Invoice payment status
type PaymentStatus string

const (
	PaymentStatusUnpaid  PaymentStatus = "UNPAID"
	PaymentStatusPartial PaymentStatus = "PARTIAL"
	PaymentStatusPaid    PaymentStatus = "PAID"
	PaymentStatusOverdue PaymentStatus = "OVERDUE"
)

// PaymentMethod - Payment types
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "CASH"
	PaymentMethodTransfer PaymentMethod = "TRANSFER"
	PaymentMethodCheck    PaymentMethod = "CHECK"
	PaymentMethodGiro     PaymentMethod = "GIRO"
	PaymentMethodOther    PaymentMethod = "OTHER"
)

// CheckStatus - Check/Giro status tracking
type CheckStatus string

const (
	CheckStatusIssued    CheckStatus = "ISSUED"    // Check/giro diterbitkan
	CheckStatusCleared   CheckStatus = "CLEARED"   // Sudah cair
	CheckStatusBounced   CheckStatus = "BOUNCED"   // Ditolak/gagal cair
	CheckStatusCancelled CheckStatus = "CANCELLED" // Dibatalkan
)

// GoodsReceiptStatus - GRN workflow
type GoodsReceiptStatus string

const (
	GoodsReceiptStatusPending   GoodsReceiptStatus = "PENDING"   // Waiting for goods
	GoodsReceiptStatusReceived  GoodsReceiptStatus = "RECEIVED"  // Physically received
	GoodsReceiptStatusInspected GoodsReceiptStatus = "INSPECTED" // Quality inspection done
	GoodsReceiptStatusAccepted  GoodsReceiptStatus = "ACCEPTED"  // Accepted, stock updated
	GoodsReceiptStatusRejected  GoodsReceiptStatus = "REJECTED"  // Rejected, no stock update
	GoodsReceiptStatusPartial   GoodsReceiptStatus = "PARTIAL"   // Partially accepted
)

// MovementType - Inventory movement types
type MovementType string

const (
	MovementTypeIn         MovementType = "IN"         // Stock masuk (GRN)
	MovementTypeOut        MovementType = "OUT"        // Stock keluar (delivery)
	MovementTypeAdjustment MovementType = "ADJUSTMENT" // Stock opname adjustment
	MovementTypeReturn     MovementType = "RETURN"     // Return dari customer
	MovementTypeDamaged    MovementType = "DAMAGED"    // Barang rusak
	MovementTypeTransfer   MovementType = "TRANSFER"   // Transfer antar gudang
)

// StockOpnameStatus - Physical count workflow
type StockOpnameStatus string

const (
	StockOpnameStatusDraft      StockOpnameStatus = "DRAFT"       // Created, not started
	StockOpnameStatusInProgress StockOpnameStatus = "IN_PROGRESS" // Being counted
	StockOpnameStatusCompleted  StockOpnameStatus = "COMPLETED"   // Counting done
	StockOpnameStatusApproved   StockOpnameStatus = "APPROVED"    // Approved, adjustments posted
	StockOpnameStatusCancelled  StockOpnameStatus = "CANCELLED"   // Cancelled
)

// StockTransferStatus - Inter-warehouse transfer workflow
type StockTransferStatus string

const (
	StockTransferStatusDraft     StockTransferStatus = "DRAFT"     // Created, not shipped
	StockTransferStatusShipped   StockTransferStatus = "SHIPPED"   // Shipped from source
	StockTransferStatusReceived  StockTransferStatus = "RECEIVED"  // Received at destination
	StockTransferStatusCancelled StockTransferStatus = "CANCELLED" // Cancelled
)

// DeliveryType - Delivery classification
type DeliveryType string

const (
	DeliveryTypeNormal      DeliveryType = "NORMAL"      // Pengiriman normal
	DeliveryTypeReturn      DeliveryType = "RETURN"      // Return dari customer
	DeliveryTypeReplacement DeliveryType = "REPLACEMENT" // Penggantian barang rusak
)

// DeliveryStatus - Simplified delivery workflow (PHASE 0)
type DeliveryStatus string

const (
	DeliveryStatusPrepared   DeliveryStatus = "PREPARED"   // Barang disiapkan di gudang
	DeliveryStatusInTransit  DeliveryStatus = "IN_TRANSIT" // Dalam perjalanan
	DeliveryStatusDelivered  DeliveryStatus = "DELIVERED"  // Sudah sampai
	DeliveryStatusConfirmed  DeliveryStatus = "CONFIRMED"  // Customer konfirmasi terima
	DeliveryStatusCancelled  DeliveryStatus = "CANCELLED"  // Dibatalkan
)

// CashTransactionType - Cash transaction direction
type CashTransactionType string

const (
	CashTransactionTypeCashIn  CashTransactionType = "CASH_IN"  // Pemasukan (debit)
	CashTransactionTypeCashOut CashTransactionType = "CASH_OUT" // Pengeluaran (credit)
)

// CashCategory - Cash transaction categories
type CashCategory string

const (
	CashCategorySales         CashCategory = "SALES"          // Penjualan
	CashCategoryPurchase      CashCategory = "PURCHASE"       // Pembelian
	CashCategoryExpense       CashCategory = "EXPENSE"        // Biaya operasional
	CashCategoryPayroll       CashCategory = "PAYROLL"        // Gaji karyawan
	CashCategoryLoan          CashCategory = "LOAN"           // Pinjaman/hutang
	CashCategoryInvestment    CashCategory = "INVESTMENT"     // Investasi
	CashCategoryWithdrawal    CashCategory = "WITHDRAWAL"     // Penarikan modal
	CashCategoryDeposit       CashCategory = "DEPOSIT"        // Setoran modal
	CashCategoryOtherIncome   CashCategory = "OTHER_INCOME"   // Pendapatan lain
	CashCategoryOtherExpense  CashCategory = "OTHER_EXPENSE"  // Pengeluaran lain
)

// InventoryAdjustmentStatus - Stock adjustment workflow
type InventoryAdjustmentStatus string

const (
	InventoryAdjustmentStatusDraft     InventoryAdjustmentStatus = "DRAFT"     // Created, not approved
	InventoryAdjustmentStatusApproved  InventoryAdjustmentStatus = "APPROVED"  // Approved, stock updated
	InventoryAdjustmentStatusCancelled InventoryAdjustmentStatus = "CANCELLED" // Cancelled
)

// InventoryAdjustmentType - Stock adjustment direction
type InventoryAdjustmentType string

const (
	InventoryAdjustmentTypeIncrease InventoryAdjustmentType = "INCREASE" // Penambahan stok
	InventoryAdjustmentTypeDecrease InventoryAdjustmentType = "DECREASE" // Pengurangan stok
)

// InventoryAdjustmentReason - Reason for stock adjustment
type InventoryAdjustmentReason string

const (
	InventoryAdjustmentReasonShrinkage  InventoryAdjustmentReason = "SHRINKAGE"  // Susut/kehilangan
	InventoryAdjustmentReasonDamage     InventoryAdjustmentReason = "DAMAGE"     // Barang rusak
	InventoryAdjustmentReasonExpired    InventoryAdjustmentReason = "EXPIRED"    // Kadaluarsa
	InventoryAdjustmentReasonTheft      InventoryAdjustmentReason = "THEFT"      // Pencurian
	InventoryAdjustmentReasonOpname     InventoryAdjustmentReason = "OPNAME"     // Hasil stok opname
	InventoryAdjustmentReasonCorrection InventoryAdjustmentReason = "CORRECTION" // Koreksi data
	InventoryAdjustmentReasonReturn     InventoryAdjustmentReason = "RETURN"     // Retur supplier
	InventoryAdjustmentReasonOther      InventoryAdjustmentReason = "OTHER"      // Lainnya
)
