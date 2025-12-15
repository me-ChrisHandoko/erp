// Package models - Enum type definitions
// All enum values match Prisma schema exactly for database compatibility
package models

// UserRole - User roles with per-tenant assignment
type UserRole string

const (
	UserRoleOwner     UserRole = "OWNER"
	UserRoleAdmin     UserRole = "ADMIN"
	UserRoleFinance   UserRole = "FINANCE"
	UserRoleSales     UserRole = "SALES"
	UserRoleWarehouse UserRole = "WAREHOUSE"
	UserRoleStaff     UserRole = "STAFF"
)

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
	StockOpnameStatusDraft     StockOpnameStatus = "DRAFT"     // Being counted
	StockOpnameStatusCompleted StockOpnameStatus = "COMPLETED" // Counting done
	StockOpnameStatusApproved  StockOpnameStatus = "APPROVED"  // Approved, adjustments posted
	StockOpnameStatusCancelled StockOpnameStatus = "CANCELLED" // Cancelled
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
