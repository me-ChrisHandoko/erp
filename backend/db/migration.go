// Package db - Database migration logic
package db

import (
	"backend/internal/service/auth"
	"backend/models"

	"gorm.io/gorm"
)

// AutoMigratePhase1 runs GORM auto-migration for Phase 1 models
// Phase 1: Core models (User, Tenant, Company)
// CRITICAL: Order matters - parent tables before child tables
func AutoMigratePhase1(db *gorm.DB) error {
	return db.AutoMigrate(
		// User management
		&models.User{},

		// Company
		&models.Company{},
		&models.CompanyBank{},

		// Multi-tenancy layer
		&models.Subscription{},
		&models.Tenant{},
		&models.SubscriptionPayment{},
		&models.UserTenant{},
		&models.UserCompanyRole{},
	)
}

// AutoMigratePhase2 runs GORM auto-migration for Phase 2 models
// Phase 2: Product & Inventory models
// CRITICAL: Order matters - parent tables before child tables
func AutoMigratePhase2(db *gorm.DB) error {
	return db.AutoMigrate(
		// Master data
		&models.Customer{},
		&models.Supplier{},

		// Warehouse management
		&models.Warehouse{},

		// Product master
		&models.Product{},
		&models.ProductUnit{},
		&models.PriceList{},
		&models.ProductSupplier{},

		// Stock tracking
		&models.WarehouseStock{},
		&models.ProductBatch{},
	)
}

// AutoMigratePhase3 runs GORM auto-migration for Phase 3 models
// Phase 3: Transaction models (Sales, Purchase, Delivery)
// CRITICAL: Order matters - parent tables before child tables
func AutoMigratePhase3(db *gorm.DB) error {
	return db.AutoMigrate(
		// Sales workflow
		&models.SalesOrder{},
		&models.SalesOrderItem{},

		// Delivery workflow
		&models.Delivery{},
		&models.DeliveryItem{},

		// Invoice & Payment
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Payment{},
		&models.PaymentCheck{},

		// Purchase workflow
		&models.PurchaseOrder{},
		&models.PurchaseOrderItem{},

		// Goods receipt workflow
		&models.GoodsReceipt{},
		&models.GoodsReceiptItem{},

		// Purchase invoice workflow (Faktur Pembelian)
		&models.PurchaseInvoice{},
		&models.PurchaseInvoiceItem{},
		&models.PurchaseInvoicePayment{},

		// Supplier payment
		&models.SupplierPayment{},
	)
}

// AutoMigrate runs GORM auto-migration for all models
// This will be used when all phases are implemented
func AutoMigrate(db *gorm.DB) error {
	// Auth models (must run first as users table is referenced by other tables)
	if err := AutoMigrateAuth(db); err != nil {
		return err
	}

	// Phase 1: Core models
	if err := AutoMigratePhase1(db); err != nil {
		return err
	}

	// Phase 2: Product & Inventory
	if err := AutoMigratePhase2(db); err != nil {
		return err
	}

	// Phase 3: Transactions (Sales, Purchase, Delivery)
	if err := AutoMigratePhase3(db); err != nil {
		return err
	}

	// Phase 4: Supporting modules (InventoryMovement, StockOpname, CashTransaction, etc.)
	if err := AutoMigratePhase4(db); err != nil {
		return err
	}

	return nil
}

// AutoMigrateAuth runs GORM auto-migration for authentication models
// Auth models: User, RefreshToken, EmailVerification, PasswordReset, LoginAttempt
// CRITICAL: Must run before other phases as users table is referenced
func AutoMigrateAuth(db *gorm.DB) error {
	return db.AutoMigrate(
		&auth.RefreshToken{},
		&auth.EmailVerification{},
		&auth.PasswordReset{},
		&auth.LoginAttempt{},
	)
}

// AutoMigratePhase4 runs GORM auto-migration for Phase 4 models
// Phase 4: Supporting modules (InventoryMovement, StockOpname, StockTransfer, InventoryAdjustment, CashTransaction, System)
// CRITICAL: Order matters - parent tables before child tables
func AutoMigratePhase4(db *gorm.DB) error {
	if err := db.AutoMigrate(
		// Inventory tracking
		&models.InventoryMovement{},

		// Stock opname (physical count)
		&models.StockOpname{},
		&models.StockOpnameItem{},

		// Inter-warehouse transfer
		&models.StockTransfer{},
		&models.StockTransferItem{},

		// Inventory adjustment (manual stock adjustments)
		&models.InventoryAdjustment{},
		&models.InventoryAdjustmentItem{},

		// Cash book (Buku Kas)
		&models.CashTransaction{},

		// System configuration & audit
		&models.Setting{},
		&models.AuditLog{},

		// Procurement settings (SAP Model)
		&models.DeliveryTolerance{},
	); err != nil {
		return err
	}

	// Drop FK constraint on delivery_tolerances.product_id to allow empty string for COMPANY/CATEGORY levels
	// This is needed because we use empty string pattern for unique index instead of NULL
	db.Exec("ALTER TABLE IF EXISTS delivery_tolerances DROP CONSTRAINT IF EXISTS fk_delivery_tolerances_product")

	return nil
}
