// Package models - Phase 2 comprehensive tests
package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPhase2TestDB creates an in-memory SQLite database for Phase 2 testing
func setupPhase2TestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign keys for CASCADE
	database.Exec("PRAGMA foreign_keys = ON")

	// Run Phase 1 migration first (dependencies)
	if err := autoMigratePhase1(database); err != nil {
		t.Fatalf("Phase 1 auto-migration failed: %v", err)
	}

	// Run Phase 2 migration
	if err := autoMigratePhase2(database); err != nil {
		t.Fatalf("Phase 2 auto-migration failed: %v", err)
	}

	return database
}

// autoMigratePhase1 runs GORM auto-migration for Phase 1 models (local copy)
func autoMigratePhase1(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Company{},
		&CompanyBank{},
		&Subscription{},
		&Tenant{},
		&SubscriptionPayment{},
		&UserTenant{},
	)
}

// autoMigratePhase2 runs GORM auto-migration for Phase 2 models (local copy)
func autoMigratePhase2(db *gorm.DB) error {
	return db.AutoMigrate(
		&Customer{},
		&Supplier{},
		&Warehouse{},
		&Product{},
		&ProductUnit{},
		&PriceList{},
		&ProductSupplier{},
		&WarehouseStock{},
		&ProductBatch{},
	)
}

// createTestTenant creates a tenant for testing Phase 2 models
func createTestTenant(t *testing.T, database *gorm.DB) *Tenant {
	company := &Company{
		Name:      "Test Company",
		LegalName: "Test Company Ltd",
		Address:   "Test Address",
		City:      "Test City",
		Province:  "Test Province",
	}
	if err := database.Create(company).Error; err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}

	tenant := &Tenant{
		CompanyID: company.ID,
		Status:    TenantStatusActive,
	}
	if err := database.Create(tenant).Error; err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	return tenant
}

// TestPhase2SchemaGeneration verifies all Phase 2 tables are created
func TestPhase2SchemaGeneration(t *testing.T) {
	database := setupPhase2TestDB(t)

	tables := []string{
		"customers", "suppliers", "warehouses", "products",
		"product_units", "price_list", "product_suppliers",
		"warehouse_stocks", "product_batches",
	}

	for _, table := range tables {
		if !database.Migrator().HasTable(table) {
			t.Errorf("Table %s was not created", table)
		}
	}
}

// TestCustomerCreation validates Customer model creation with all fields
func TestCustomerCreation(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	npwp := "01.234.567.8-901.234"
	customer := &Customer{
		TenantID:           tenant.ID,
		Code:               "CUST-001",
		Name:               "Toko Berkah",
		Type:               stringPtr("RETAIL"),
		Phone:              stringPtr("021-1234567"),
		Email:              stringPtr("toko@berkah.com"),
		Address:            stringPtr("Jl. Raya No. 123"),
		City:               stringPtr("Jakarta"),
		Province:           stringPtr("DKI Jakarta"),
		NPWP:               &npwp,
		IsPKP:              true,
		PaymentTerm:        30,
		CreditLimit:        decimal.NewFromFloat(50000000),
		CurrentOutstanding: decimal.NewFromFloat(0),
		OverdueAmount:      decimal.NewFromFloat(0),
		IsActive:           true,
	}

	result := database.Create(customer)
	if result.Error != nil {
		t.Fatalf("Failed to create customer: %v", result.Error)
	}

	if customer.ID == "" {
		t.Error("Customer ID (CUID) was not generated")
	}

	if customer.CreatedAt.IsZero() {
		t.Error("CreatedAt timestamp was not set")
	}
}

// TestSupplierCreation validates Supplier model creation with all fields
func TestSupplierCreation(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	supplier := &Supplier{
		TenantID:           tenant.ID,
		Code:               "SUPP-001",
		Name:               "PT Distribusi Utama",
		Type:               stringPtr("DISTRIBUTOR"),
		Phone:              stringPtr("021-7654321"),
		Email:              stringPtr("info@distribusi.com"),
		Address:            stringPtr("Jl. Industri No. 456"),
		City:               stringPtr("Tangerang"),
		Province:           stringPtr("Banten"),
		PaymentTerm:        45,
		CreditLimit:        decimal.NewFromFloat(100000000),
		CurrentOutstanding: decimal.NewFromFloat(0),
		OverdueAmount:      decimal.NewFromFloat(0),
		IsActive:           true,
	}

	result := database.Create(supplier)
	if result.Error != nil {
		t.Fatalf("Failed to create supplier: %v", result.Error)
	}

	if supplier.ID == "" {
		t.Error("Supplier ID (CUID) was not generated")
	}
}

// TestWarehouseCreation validates Warehouse model creation
func TestWarehouseCreation(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	warehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-MAIN",
		Name:     "Gudang Utama",
		Type:     WarehouseTypeMain,
		Address:  stringPtr("Jl. Gudang No. 1"),
		City:     stringPtr("Jakarta"),
		Province: stringPtr("DKI Jakarta"),
		Capacity: decimalPtr(decimal.NewFromFloat(1000.00)),
		IsActive: true,
	}

	result := database.Create(warehouse)
	if result.Error != nil {
		t.Fatalf("Failed to create warehouse: %v", result.Error)
	}

	if warehouse.ID == "" {
		t.Error("Warehouse ID (CUID) was not generated")
	}
}

// TestProductCreation validates Product model with batch tracking
func TestProductCreation(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	barcode := "8992761151010"
	product := &Product{
		TenantID:       tenant.ID,
		Code:           "PROD-001",
		Name:           "Beras Premium 5kg",
		Category:       stringPtr("BERAS"),
		BaseUnit:       "KG",
		BaseCost:       decimal.NewFromFloat(12000),
		BasePrice:      decimal.NewFromFloat(15000),
		CurrentStock:   decimal.NewFromFloat(0),
		MinimumStock:   decimal.NewFromFloat(100),
		Barcode:        &barcode,
		IsBatchTracked: true,
		IsPerishable:   false,
		IsActive:       true,
	}

	result := database.Create(product)
	if result.Error != nil {
		t.Fatalf("Failed to create product: %v", result.Error)
	}

	if product.ID == "" {
		t.Error("Product ID (CUID) was not generated")
	}
}

// TestProductUnitConversion validates ProductUnit with conversion rates
func TestProductUnitConversion(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	product := &Product{
		TenantID: tenant.ID,
		Code:     "PROD-002",
		Name:     "Mie Instan",
		BaseUnit: "PCS",
		BaseCost: decimal.NewFromFloat(2500),
		IsActive: true,
	}
	database.Create(product)

	// Create KARTON unit (1 KARTON = 40 PCS)
	kartonUnit := &ProductUnit{
		ProductID:      product.ID,
		UnitName:       "KARTON",
		ConversionRate: decimal.NewFromFloat(40),
		IsBaseUnit:     false,
		BuyPrice:       decimalPtr(decimal.NewFromFloat(90000)),
		SellPrice:      decimalPtr(decimal.NewFromFloat(100000)),
		IsActive:       true,
	}

	result := database.Create(kartonUnit)
	if result.Error != nil {
		t.Fatalf("Failed to create product unit: %v", result.Error)
	}

	if kartonUnit.ID == "" {
		t.Error("ProductUnit ID (CUID) was not generated")
	}

	// Verify conversion: 1 KARTON should equal 40 PCS in base units
	expectedBaseUnits := decimal.NewFromFloat(40)
	if !kartonUnit.ConversionRate.Equal(expectedBaseUnits) {
		t.Errorf("Conversion rate mismatch: expected %s, got %s",
			expectedBaseUnits.String(), kartonUnit.ConversionRate.String())
	}
}

// TestProductBatchTracking validates batch tracking with expiry dates
func TestProductBatchTracking(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	// Create product with batch tracking enabled
	product := &Product{
		TenantID:       tenant.ID,
		Code:           "PROD-003",
		Name:           "Susu UHT 1L",
		BaseUnit:       "PCS",
		IsBatchTracked: true,
		IsPerishable:   true,
		IsActive:       true,
	}
	database.Create(product)

	// Create warehouse
	warehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-001",
		Name:     "Gudang Dingin",
		Type:     WarehouseTypeMain,
		IsActive: true,
	}
	database.Create(warehouse)

	// Create warehouse stock
	warehouseStock := &WarehouseStock{
		WarehouseID:  warehouse.ID,
		ProductID:    product.ID,
		Quantity:     decimal.NewFromFloat(100),
		MinimumStock: decimal.NewFromFloat(50),
		MaximumStock: decimal.NewFromFloat(500),
	}
	database.Create(warehouseStock)

	// Create batch with expiry date
	mfgDate := time.Now()
	expDate := time.Now().AddDate(0, 6, 0) // 6 months from now
	batch := &ProductBatch{
		BatchNumber:      "BATCH-2025-001",
		ProductID:        product.ID,
		WarehouseStockID: warehouseStock.ID,
		ManufactureDate:  &mfgDate,
		ExpiryDate:       &expDate,
		Quantity:         decimal.NewFromFloat(100),
		Status:           BatchStatusAvailable,
		QualityStatus:    stringPtr("GOOD"),
		ReceiptDate:      time.Now(),
	}

	result := database.Create(batch)
	if result.Error != nil {
		t.Fatalf("Failed to create product batch: %v", result.Error)
	}

	if batch.ID == "" {
		t.Error("ProductBatch ID (CUID) was not generated")
	}

	// Verify batch is associated with product
	var loadedBatch ProductBatch
	database.Preload("Product").First(&loadedBatch, "id = ?", batch.ID)
	if loadedBatch.Product.Code != product.Code {
		t.Error("Batch-Product relationship not properly loaded")
	}
}

// TestPriceListCustomerPricing validates customer-specific pricing
func TestPriceListCustomerPricing(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	product := &Product{
		TenantID:  tenant.ID,
		Code:      "PROD-004",
		Name:      "Gula Pasir 1kg",
		BaseUnit:  "KG",
		BasePrice: decimal.NewFromFloat(15000),
		IsActive:  true,
	}
	database.Create(product)

	customer := &Customer{
		TenantID: tenant.ID,
		Code:     "CUST-002",
		Name:     "Toko Makmur",
		IsActive: true,
	}
	database.Create(customer)

	// Create customer-specific pricing (discounted)
	priceList := &PriceList{
		ProductID:     product.ID,
		CustomerID:    &customer.ID,
		Price:         decimal.NewFromFloat(14000), // Rp 1,000 discount
		MinQty:        decimal.NewFromFloat(10),
		EffectiveFrom: time.Now(),
		IsActive:      true,
	}

	result := database.Create(priceList)
	if result.Error != nil {
		t.Fatalf("Failed to create price list: %v", result.Error)
	}

	if priceList.ID == "" {
		t.Error("PriceList ID (CUID) was not generated")
	}
}

// TestProductSupplierRelationship validates product-supplier junction
func TestProductSupplierRelationship(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	product := &Product{
		TenantID: tenant.ID,
		Code:     "PROD-005",
		Name:     "Kopi Bubuk 200g",
		BaseUnit: "PCS",
		IsActive: true,
	}
	database.Create(product)

	supplier := &Supplier{
		TenantID: tenant.ID,
		Code:     "SUPP-002",
		Name:     "PT Kopi Nusantara",
		IsActive: true,
	}
	database.Create(supplier)

	// Link product to supplier with pricing
	productSupplier := &ProductSupplier{
		ProductID:     product.ID,
		SupplierID:    supplier.ID,
		SupplierPrice: decimal.NewFromFloat(25000),
		LeadTime:      7, // 7 days
		IsPrimary:     true,
	}

	result := database.Create(productSupplier)
	if result.Error != nil {
		t.Fatalf("Failed to create product-supplier relationship: %v", result.Error)
	}

	if productSupplier.ID == "" {
		t.Error("ProductSupplier ID (CUID) was not generated")
	}
}

// TestWarehouseStockTracking validates stock per warehouse tracking
func TestWarehouseStockTracking(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	warehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-002",
		Name:     "Gudang Cabang",
		Type:     WarehouseTypeBranch,
		IsActive: true,
	}
	database.Create(warehouse)

	product := &Product{
		TenantID: tenant.ID,
		Code:     "PROD-006",
		Name:     "Minyak Goreng 2L",
		BaseUnit: "PCS",
		IsActive: true,
	}
	database.Create(product)

	// Create warehouse stock record
	countDate := time.Now()
	countQty := decimal.NewFromFloat(150)
	warehouseStock := &WarehouseStock{
		WarehouseID:   warehouse.ID,
		ProductID:     product.ID,
		Quantity:      decimal.NewFromFloat(150),
		MinimumStock:  decimal.NewFromFloat(50),
		MaximumStock:  decimal.NewFromFloat(300),
		Location:      stringPtr("RAK-A-01"),
		LastCountDate: &countDate,
		LastCountQty:  &countQty,
	}

	result := database.Create(warehouseStock)
	if result.Error != nil {
		t.Fatalf("Failed to create warehouse stock: %v", result.Error)
	}

	if warehouseStock.ID == "" {
		t.Error("WarehouseStock ID (CUID) was not generated")
	}

	// Verify stock quantity is decimal
	if !warehouseStock.Quantity.Equal(decimal.NewFromFloat(150)) {
		t.Errorf("Stock quantity mismatch: expected 150, got %s",
			warehouseStock.Quantity.String())
	}
}

// TestUniqueConstraintsPhase2 validates composite unique indexes work
func TestUniqueConstraintsPhase2(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	// Test duplicate customer code within same tenant
	customer1 := &Customer{
		TenantID: tenant.ID,
		Code:     "CUST-DUP",
		Name:     "Customer One",
		IsActive: true,
	}
	database.Create(customer1)

	customer2 := &Customer{
		TenantID: tenant.ID,
		Code:     "CUST-DUP", // Same code, same tenant
		Name:     "Customer Two",
		IsActive: true,
	}

	result := database.Create(customer2)
	if result.Error == nil {
		t.Error("Should have failed due to duplicate customer code in same tenant")
	}
}

// TestCascadeDeletePhase2 validates CASCADE deletion from Tenant
func TestCascadeDeletePhase2(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	// Create product linked to tenant
	product := &Product{
		TenantID: tenant.ID,
		Code:     "PROD-CASCADE",
		Name:     "Test Product",
		BaseUnit: "PCS",
		IsActive: true,
	}
	database.Create(product)

	// Create warehouse linked to tenant
	warehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-CASCADE",
		Name:     "Test Warehouse",
		Type:     WarehouseTypeMain,
		IsActive: true,
	}
	database.Create(warehouse)

	// Delete tenant (should cascade to products and warehouses)
	database.Select("Company").Delete(tenant)

	// Verify product was deleted
	var productCount int64
	database.Model(&Product{}).Where("id = ?", product.ID).Count(&productCount)
	if productCount != 0 {
		t.Error("CASCADE deletion failed - Product not deleted with Tenant")
	}

	// Verify warehouse was deleted
	var warehouseCount int64
	database.Model(&Warehouse{}).Where("id = ?", warehouse.ID).Count(&warehouseCount)
	if warehouseCount != 0 {
		t.Error("CASCADE deletion failed - Warehouse not deleted with Tenant")
	}
}

// TestDecimalPrecisionPhase2 validates decimal accuracy for quantities
func TestDecimalPrecisionPhase2(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	warehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-DECIMAL",
		Name:     "Test Warehouse",
		Type:     WarehouseTypeMain,
		IsActive: true,
	}
	database.Create(warehouse)

	product := &Product{
		TenantID: tenant.ID,
		Code:     "PROD-DECIMAL",
		Name:     "Test Product",
		BaseUnit: "KG",
		IsActive: true,
	}
	database.Create(product)

	// Test decimal precision: 123.456 KG
	preciseQty := decimal.RequireFromString("123.456")
	warehouseStock := &WarehouseStock{
		WarehouseID: warehouse.ID,
		ProductID:   product.ID,
		Quantity:    preciseQty,
	}
	database.Create(warehouseStock)

	// Load from database
	var loadedStock WarehouseStock
	database.First(&loadedStock, "id = ?", warehouseStock.ID)

	if !loadedStock.Quantity.Equal(preciseQty) {
		t.Errorf("Decimal precision lost: expected %s, got %s",
			preciseQty.String(), loadedStock.Quantity.String())
	}
}

// TestEnumValuesPhase2 validates enum types work correctly
func TestEnumValuesPhase2(t *testing.T) {
	database := setupPhase2TestDB(t)
	tenant := createTestTenant(t, database)

	warehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-ENUM",
		Name:     "Test Warehouse",
		Type:     WarehouseTypeMain, // Enum value
		IsActive: true,
	}
	database.Create(warehouse)

	// Load and verify enum
	var loadedWarehouse Warehouse
	database.First(&loadedWarehouse, "id = ?", warehouse.ID)

	if loadedWarehouse.Type != WarehouseTypeMain {
		t.Errorf("Enum value mismatch: expected %s, got %s",
			WarehouseTypeMain, loadedWarehouse.Type)
	}

	// Test enum update
	loadedWarehouse.Type = WarehouseTypeBranch
	database.Save(&loadedWarehouse)

	database.First(&loadedWarehouse, "id = ?", warehouse.ID)
	if loadedWarehouse.Type != WarehouseTypeBranch {
		t.Error("Enum value not updated correctly")
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
