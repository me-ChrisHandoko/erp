// Package db - Database migration logic
package db

import (
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

		// Supplier payment
		&models.SupplierPayment{},
	)
}

// AutoMigrate runs GORM auto-migration for all models
// This will be used when all phases are implemented
func AutoMigrate(db *gorm.DB) error {
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

// AutoMigratePhase4 runs GORM auto-migration for Phase 4 models
// Phase 4: Supporting modules (InventoryMovement, StockOpname, StockTransfer, CashTransaction, System)
// CRITICAL: Order matters - parent tables before child tables
func AutoMigratePhase4(db *gorm.DB) error {
	return db.AutoMigrate(
		// Inventory tracking
		&models.InventoryMovement{},

		// Stock opname (physical count)
		&models.StockOpname{},
		&models.StockOpnameItem{},

		// Inter-warehouse transfer
		&models.StockTransfer{},
		&models.StockTransferItem{},

		// Cash book (Buku Kas)
		&models.CashTransaction{},

		// System configuration & audit
		&models.Setting{},
		&models.AuditLog{},
	)
}
