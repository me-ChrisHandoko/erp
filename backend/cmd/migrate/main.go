package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/models"
)

func main() {
	// Get DATABASE_URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	log.Println("Connecting to database...")
	log.Printf("URL: %s", databaseURL)

	// Connect to database with special config for migration
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true, // Disable FK checks during migration
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected successfully!")
	log.Println("Checking existing tables...")

	// Check which tables already exist
	existingTables := make(map[string]bool)
	rows, err := db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'").Rows()
	if err != nil {
		log.Fatalf("Failed to query existing tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		existingTables[tableName] = true
		log.Printf("  - Found: %s", tableName)
	}

	log.Printf("Found %d existing tables", len(existingTables))
	log.Println("Starting AutoMigrate for NEW models only...")

	// Define all models with their expected table names
	allModels := map[string]interface{}{
		// Master data models (NEW - not in existing migrations)
		"warehouses":         &models.Warehouse{},
		"customers":          &models.Customer{},
		"suppliers":          &models.Supplier{},
		"products":           &models.Product{},
		"product_batches":    &models.ProductBatch{},
		"product_units":      &models.ProductUnit{},
		"product_suppliers":  &models.ProductSupplier{},
		"price_lists":        &models.PriceList{},
		"warehouse_stocks":   &models.WarehouseStock{},
		"company_banks":      &models.CompanyBank{},
		"settings":           &models.Setting{},
		"audit_logs":         &models.AuditLog{},

		// Transaction models
		"sales_orders":          &models.SalesOrder{},
		"sales_order_items":     &models.SalesOrderItem{},
		"purchase_orders":       &models.PurchaseOrder{},
		"purchase_order_items":  &models.PurchaseOrderItem{},
		"goods_receipts":        &models.GoodsReceipt{},
		"goods_receipt_items":   &models.GoodsReceiptItem{},
		"deliveries":            &models.Delivery{},
		"delivery_items":        &models.DeliveryItem{},
		"invoices":                  &models.Invoice{},
		"invoice_items":             &models.InvoiceItem{},
		"payments":                  &models.Payment{},
		"payment_checks":            &models.PaymentCheck{},
		"purchase_invoices":         &models.PurchaseInvoice{},
		"purchase_invoice_items":    &models.PurchaseInvoiceItem{},
		"purchase_invoice_payments": &models.PurchaseInvoicePayment{},
		"supplier_payments":         &models.SupplierPayment{},
		"cash_transactions":         &models.CashTransaction{},

		// Inventory models
		"inventory_movements":        &models.InventoryMovement{},
		"stock_opnames":              &models.StockOpname{},
		"stock_opname_items":         &models.StockOpnameItem{},
		"stock_transfers":            &models.StockTransfer{},
		"stock_transfer_items":       &models.StockTransferItem{},
		"inventory_adjustments":      &models.InventoryAdjustment{},
		"inventory_adjustment_items": &models.InventoryAdjustmentItem{},

		// Procurement settings (SAP Model)
		"delivery_tolerances": &models.DeliveryTolerance{},
	}

	// Separate NEW models from existing ones
	var newModels []interface{}
	var skippedTables []string

	for tableName, model := range allModels {
		if existingTables[tableName] {
			log.Printf("⏭️  Skipping: %s (already exists)", tableName)
			skippedTables = append(skippedTables, tableName)
		} else {
			log.Printf("📦 Will create: %s", tableName)
			newModels = append(newModels, model)
		}
	}

	// Migrate all NEW models at once (GORM handles dependencies)
	if len(newModels) > 0 {
		log.Printf("\n🚀 Creating %d new tables...\n", len(newModels))
		err = db.AutoMigrate(newModels...)
		if err != nil {
			log.Fatalf("AutoMigrate failed: %v", err)
		}
	} else {
		log.Println("\n⚠️  No new tables to create - all tables already exist")
	}

	log.Println("")
	log.Printf("✅ AutoMigrate completed successfully!")
	log.Printf("   - Created: %d new tables", len(newModels))
	log.Printf("   - Skipped: %d existing tables", len(skippedTables))
	log.Printf("   - Total in database: %d tables", len(existingTables)+len(newModels))
}

// Helper function to get model name from interface
func getModelName(model interface{}) string {
	typeName := fmt.Sprintf("%T", model)
	// Remove pointer prefix and package name
	// e.g., "*models.Product" -> "Product"
	typeName = strings.TrimPrefix(typeName, "*models.")
	typeName = strings.TrimPrefix(typeName, "*")
	return typeName
}
