// Package main - PHASE 1: Database Foundation Migration Script
// This script implements the multi-company architecture migration
// Run with: go run cmd/migrate/phase1_multi_company.go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/models"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection
	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite" // Default to SQLite for development
	}

	var db *gorm.DB
	var err error

	if dbDriver == "postgres" {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatal("DATABASE_URL environment variable is required for PostgreSQL")
		}
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// SQLite for development
		db, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database successfully")
	log.Println("🚀 Starting PHASE 1: Multi-Company Architecture Migration")
	log.Println("=" + string(make([]byte, 60)))

	// STEP 1: Migrate Core Tables (Tenant, Company, UserCompanyRole)
	log.Println("\n📋 STEP 1: Migrating Core Tables...")
	if err := migrateCoreModels(db); err != nil {
		log.Fatalf("❌ Failed to migrate core models: %v", err)
	}
	log.Println("✅ Core tables migrated successfully")

	// STEP 2: Migrate Transactional Tables (with CompanyID)
	log.Println("\n📋 STEP 2: Migrating Transactional Tables...")
	if err := migrateTransactionalModels(db); err != nil {
		log.Fatalf("❌ Failed to migrate transactional models: %v", err)
	}
	log.Println("✅ Transactional tables migrated successfully")

	// STEP 3: Validate Schema
	log.Println("\n📋 STEP 3: Validating Database Schema...")
	if err := validateSchema(db); err != nil {
		log.Fatalf("❌ Schema validation failed: %v", err)
	}
	log.Println("✅ Schema validation passed")

	log.Println("\n" + string(make([]byte, 60)))
	log.Println("🎉 PHASE 1 Migration Completed Successfully!")
	log.Println("Next steps:")
	log.Println("  1. Run seed script: go run cmd/seed/phase1_seed.go")
	log.Println("  2. Validate with: go run cmd/validate/phase1_validate.go")
}

// migrateCoreModels migrates the core multi-company tables
func migrateCoreModels(db *gorm.DB) error {
	// Core tables in dependency order
	coreModels := []interface{}{
		&models.User{},            // Users table (no changes)
		&models.Subscription{},    // Subscriptions (no changes)
		&models.Tenant{},          // Tenant (MODIFIED: removed CompanyID, added Name & Subdomain)
		&models.Company{},         // Company (MODIFIED: added TenantID)
		&models.CompanyBank{},     // Company banks
		&models.UserTenant{},      // User-Tenant mapping (Tier 1 permissions)
		&models.UserCompanyRole{}, // NEW: User-Company-Role mapping (Tier 2 permissions)
	}

	log.Println("  Migrating core models:")
	for _, model := range coreModels {
		modelName := fmt.Sprintf("%T", model)
		log.Printf("    - %s", modelName)
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %s: %w", modelName, err)
		}
	}

	return nil
}

// migrateTransactionalModels migrates all transactional tables with CompanyID
func migrateTransactionalModels(db *gorm.DB) error {
	// Transactional models in dependency order
	// All of these now have CompanyID field
	transactionalModels := []interface{}{
		// Master Data
		&models.Warehouse{},       // MODIFIED: added CompanyID
		&models.Product{},         // MODIFIED: added CompanyID
		&models.ProductUnit{},     // Product units
		&models.ProductBatch{},    // Product batches
		&models.Customer{},        // MODIFIED: added CompanyID
		&models.Supplier{},        // MODIFIED: added CompanyID
		&models.ProductSupplier{}, // Product-Supplier mapping

		// Sales & Purchase (MODIFIED: added CompanyID)
		&models.SalesOrder{},       // MODIFIED: added CompanyID
		&models.SalesOrderItem{},
		&models.PurchaseOrder{},    // MODIFIED: added CompanyID
		&models.PurchaseOrderItem{},
		&models.Invoice{},          // MODIFIED: added CompanyID
		&models.InvoiceItem{},

		// Inventory (MODIFIED: added CompanyID)
		&models.Delivery{},         // MODIFIED: added CompanyID
		&models.DeliveryItem{},
		&models.GoodsReceipt{},     // MODIFIED: added CompanyID
		&models.GoodsReceiptItem{},
		&models.InventoryMovement{}, // MODIFIED: added CompanyID
		&models.StockTransfer{},    // MODIFIED: added CompanyID
		&models.StockTransferItem{},
		&models.StockOpname{},      // MODIFIED: added CompanyID
		&models.StockOpnameItem{},

		// Financial (MODIFIED: added CompanyID)
		&models.CashTransaction{},  // MODIFIED: added CompanyID
		&models.SupplierPayment{},  // MODIFIED: added CompanyID

		// Supporting tables
		&models.WarehouseStock{},
		&models.PriceList{},
		&models.SubscriptionPayment{},
	}

	log.Println("  Migrating transactional models:")
	for _, model := range transactionalModels {
		modelName := fmt.Sprintf("%T", model)
		log.Printf("    - %s", modelName)
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %s: %w", modelName, err)
		}
	}

	return nil
}

// validateSchema validates the database schema after migration
func validateSchema(db *gorm.DB) error {
	log.Println("  Checking Tenant-Company relationship...")

	// Check if tenants table exists and has correct columns
	if !db.Migrator().HasTable("tenants") {
		return fmt.Errorf("tenants table does not exist")
	}

	// Verify Tenant has Name and Subdomain columns
	if !db.Migrator().HasColumn(&models.Tenant{}, "Name") {
		return fmt.Errorf("tenants table missing 'name' column")
	}
	if !db.Migrator().HasColumn(&models.Tenant{}, "Subdomain") {
		return fmt.Errorf("tenants table missing 'subdomain' column")
	}

	// Check if companies table exists and has TenantID
	if !db.Migrator().HasTable("companies") {
		return fmt.Errorf("companies table does not exist")
	}
	if !db.Migrator().HasColumn(&models.Company{}, "TenantID") {
		return fmt.Errorf("companies table missing 'tenant_id' column")
	}

	// Check if user_company_roles table exists
	if !db.Migrator().HasTable("user_company_roles") {
		return fmt.Errorf("user_company_roles table does not exist")
	}

	// Verify ALL transactional tables have CompanyID
	tablesToCheck := []struct {
		Model interface{}
		Name  string
	}{
		// Master Data
		{&models.Warehouse{}, "warehouses"},
		{&models.Product{}, "products"},
		{&models.Customer{}, "customers"},
		{&models.Supplier{}, "suppliers"},
		// Sales & Purchase
		{&models.SalesOrder{}, "sales_orders"},
		{&models.PurchaseOrder{}, "purchase_orders"},
		{&models.Invoice{}, "invoices"},
		// Inventory
		{&models.Delivery{}, "deliveries"},
		{&models.GoodsReceipt{}, "goods_receipts"},
		{&models.InventoryMovement{}, "inventory_movements"},
		{&models.StockTransfer{}, "stock_transfers"},
		{&models.StockOpname{}, "stock_opnames"},
		// Financial
		{&models.CashTransaction{}, "cash_transactions"},
		{&models.SupplierPayment{}, "supplier_payments"},
	}

	log.Println("  Checking CompanyID in transactional tables...")
	for _, table := range tablesToCheck {
		if !db.Migrator().HasColumn(table.Model, "CompanyID") {
			return fmt.Errorf("%s table missing 'company_id' column", table.Name)
		}
		log.Printf("    ✓ %s has company_id", table.Name)
	}

	return nil
}
