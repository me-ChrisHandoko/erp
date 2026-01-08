// Package main - Product Seeding Script
// Seeds 100 realistic Indonesian food distribution products
// Usage: go run cmd/seed/product_seed.go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
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
		dbDriver = "sqlite"
	}

	var db *gorm.DB
	var err error

	if dbDriver == "postgres" {
		dsn := os.Getenv("DATABASE_URL")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		db, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Run auto-migration to ensure all tables exist
	log.Println("🔧 Running auto-migration...")
	if err := autoMigrate(db); err != nil {
		log.Fatalf("❌ Auto-migration failed: %v", err)
	}
	log.Println("✅ Auto-migration completed")

	log.Println("🌱 Starting Product Seeding (100 products)")
	log.Println("=" + string(make([]byte, 60)))

	// Seed products
	if err := SeedProducts(db); err != nil {
		log.Fatalf("❌ Product seeding failed: %v", err)
	}

	log.Println("\n🎉 Product seeding completed successfully!")
	printSeedingSummary()
}

// autoMigrate runs GORM auto-migration for all models
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Tenant{},
		&models.Subscription{},
		&models.SubscriptionPayment{},
		&models.User{},
		&models.UserTenant{},
		&models.UserCompanyRole{},
		&models.Company{},
		&models.Warehouse{},
		&models.WarehouseStock{},
		&models.Product{},
		&models.ProductUnit{},
		&models.ProductBatch{},
		&models.PriceList{},
		&models.ProductSupplier{},
		&models.Customer{},
		&models.Supplier{},
	)
}

// SeedProducts - Main seeding orchestrator
func SeedProducts(db *gorm.DB) error {
	// 1. Retrieve existing entities
	log.Println("\n📋 Step 1: Retrieving existing tenants, companies, and warehouses...")
	tenant, companies, warehouses, err := getExistingEntities(db)
	if err != nil {
		return fmt.Errorf("failed to retrieve entities: %w", err)
	}
	log.Printf("  ✓ Found tenant: %s", tenant.Name)
	log.Printf("  ✓ Found %d companies", len(companies))
	log.Printf("  ✓ Found %d warehouses", len(warehouses))

	// 2. Seed product records
	log.Println("\n📦 Step 2: Seeding 100 product records...")
	products, err := seedProductRecords(db, tenant, companies)
	if err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}
	log.Printf("  ✓ Created %d products", len(products))

	// 3. Seed ProductUnits for multi-unit products
	log.Println("\n📏 Step 3: Creating multi-unit definitions...")
	productUnits, err := seedProductUnits(db, products)
	if err != nil {
		return fmt.Errorf("failed to seed product units: %w", err)
	}
	log.Printf("  ✓ Created %d product units", len(productUnits))

	// 4. Initialize warehouse stock
	log.Println("\n🏭 Step 4: Initializing warehouse stock...")
	warehouseStocks, err := seedWarehouseStock(db, tenant, products, warehouses)
	if err != nil {
		return fmt.Errorf("failed to seed warehouse stock: %w", err)
	}
	log.Printf("  ✓ Created %d warehouse stock records", len(warehouseStocks))

	// 5. Create product batches for perishables
	log.Println("\n🏷️  Step 5: Creating batches for perishable items...")
	batches, err := seedProductBatches(db, products, warehouseStocks)
	if err != nil {
		return fmt.Errorf("failed to seed product batches: %w", err)
	}
	log.Printf("  ✓ Created %d product batches", len(batches))

	return nil
}

// getExistingEntities retrieves tenant, companies, and warehouses from database
// Creates them if they don't exist for standalone execution
func getExistingEntities(db *gorm.DB) (*models.Tenant, []models.Company, []models.Warehouse, error) {
	var tenant models.Tenant
	if err := db.First(&tenant).Error; err != nil {
		// No tenant found, create one
		log.Println("  ⚠ No tenant found, creating default tenant...")
		tenant = models.Tenant{
			Name:      "PT Multi Bisnis Group",
			Subdomain: "multi-bisnis",
			Status:    models.TenantStatusActive,
		}
		if err := db.Create(&tenant).Error; err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create tenant: %w", err)
		}
		log.Printf("  ✓ Created tenant: %s", tenant.Name)
	}

	var companies []models.Company
	if err := db.Where("tenant_id = ?", tenant.ID).Find(&companies).Error; err != nil || len(companies) == 0 {
		// No companies found, create them
		log.Println("  ⚠ No companies found, creating default companies...")
		companies = []models.Company{
			{
				TenantID:   tenant.ID,
				Name:       "PT Distribusi Utama",
				LegalName:  "PT Distribusi Utama Indonesia",
				EntityType: "PT",
				Address:    "Jl. Gatot Subroto No. 123",
				City:       "Jakarta Selatan",
				Province:   "DKI Jakarta",
				Phone:      "021-12345678",
				Email:      "info@distribusi-utama.com",
				IsActive:   true,
			},
			{
				TenantID:   tenant.ID,
				Name:       "CV Sembako Jaya",
				LegalName:  "CV Sembako Jaya Abadi",
				EntityType: "CV",
				Address:    "Jl. Pasar Baru No. 45",
				City:       "Jakarta Pusat",
				Province:   "DKI Jakarta",
				Phone:      "021-87654321",
				Email:      "info@sembako-jaya.com",
				IsActive:   true,
			},
		}
		if err := db.Create(&companies).Error; err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create companies: %w", err)
		}
		log.Printf("  ✓ Created %d companies", len(companies))
	}

	var warehouses []models.Warehouse
	if err := db.Where("tenant_id = ?", tenant.ID).Find(&warehouses).Error; err != nil || len(warehouses) == 0 {
		// No warehouses found, create them
		log.Println("  ⚠ No warehouses found, creating default warehouses...")
		warehouses = []models.Warehouse{
			{
				TenantID:  tenant.ID,
				CompanyID: companies[0].ID,
				Code:      "GDG-JKT",
				Name:      "Gudang Jakarta Utama",
				Type:      models.WarehouseTypeMain,
				IsActive:  true,
			},
			{
				TenantID:  tenant.ID,
				CompanyID: companies[1].ID,
				Code:      "GDG-PSR",
				Name:      "Gudang Pasar Baru",
				Type:      models.WarehouseTypeMain,
				IsActive:  true,
			},
		}
		if err := db.Create(&warehouses).Error; err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create warehouses: %w", err)
		}
		log.Printf("  ✓ Created %d warehouses", len(warehouses))
	}

	return &tenant, companies, warehouses, nil
}

// seedProductRecords creates 100 product records
func seedProductRecords(db *gorm.DB, tenant *models.Tenant, companies []models.Company) ([]models.Product, error) {
	if len(companies) < 2 {
		return nil, fmt.Errorf("need at least 2 companies, found %d", len(companies))
	}

	company1 := companies[0] // Will get 60 products
	company2 := companies[1] // Will get 40 products

	allProducts := []models.Product{}

	// Company 1: Comprehensive product range (60 products)
	allProducts = append(allProducts, createRiceProducts(tenant.ID, company1.ID, 1, 8)...)      // 8 products
	allProducts = append(allProducts, createOilProducts(tenant.ID, company1.ID, 1, 6)...)       // 6 products
	allProducts = append(allProducts, createSugarProducts(tenant.ID, company1.ID, 1, 4)...)     // 4 products
	allProducts = append(allProducts, createFlourProducts(tenant.ID, company1.ID, 1, 4)...)     // 4 products
	allProducts = append(allProducts, createNoodleProducts(tenant.ID, company1.ID, 1, 8)...)    // 8 products
	allProducts = append(allProducts, createBeverageProducts(tenant.ID, company1.ID, 1, 6)...)  // 6 products
	allProducts = append(allProducts, createSnackProducts(tenant.ID, company1.ID, 1, 6)...)     // 6 products
	allProducts = append(allProducts, createSpiceProducts(tenant.ID, company1.ID, 1, 5)...)     // 5 products
	allProducts = append(allProducts, createSauceProducts(tenant.ID, company1.ID, 1, 4)...)     // 4 products
	allProducts = append(allProducts, createDairyProducts(tenant.ID, company1.ID, 1, 4)...)     // 4 products
	allProducts = append(allProducts, createCoffeeTeaProducts(tenant.ID, company1.ID, 1, 3)...) // 3 products
	allProducts = append(allProducts, createSoapProducts(tenant.ID, company1.ID, 1, 2)...)      // 2 products
	// Total Company 1: 60 products

	// Company 2: Focused on basic necessities (40 products)
	allProducts = append(allProducts, createRiceProducts(tenant.ID, company2.ID, 2, 2)...)       // 2 products
	allProducts = append(allProducts, createOilProducts(tenant.ID, company2.ID, 2, 2)...)        // 2 products
	allProducts = append(allProducts, createSugarProducts(tenant.ID, company2.ID, 2, 2)...)      // 2 products
	allProducts = append(allProducts, createFlourProducts(tenant.ID, company2.ID, 2, 2)...)      // 2 products
	allProducts = append(allProducts, createNoodleProducts(tenant.ID, company2.ID, 2, 2)...)     // 2 products
	allProducts = append(allProducts, createBeverageProducts(tenant.ID, company2.ID, 2, 2)...)   // 2 products
	allProducts = append(allProducts, createSnackProducts(tenant.ID, company2.ID, 2, 2)...)      // 2 products
	allProducts = append(allProducts, createSpiceProducts(tenant.ID, company2.ID, 2, 3)...)      // 3 products
	allProducts = append(allProducts, createSauceProducts(tenant.ID, company2.ID, 2, 2)...)      // 2 products
	allProducts = append(allProducts, createDairyProducts(tenant.ID, company2.ID, 2, 2)...)      // 2 products
	allProducts = append(allProducts, createCoffeeTeaProducts(tenant.ID, company2.ID, 2, 3)...)  // 3 products
	allProducts = append(allProducts, createSoapProducts(tenant.ID, company2.ID, 2, 4)...)       // 4 products
	allProducts = append(allProducts, createHouseholdProducts(tenant.ID, company2.ID, 2, 4)...)  // 4 products
	allProducts = append(allProducts, createCannedProducts(tenant.ID, company2.ID, 2, 4)...)     // 4 products
	allProducts = append(allProducts, createFrozenProducts(tenant.ID, company2.ID, 2, 4)...)     // 4 products
	// Total Company 2: 40 products

	// Batch insert for performance
	if err := db.CreateInBatches(allProducts, 50).Error; err != nil {
		return nil, fmt.Errorf("failed to insert products: %w", err)
	}

	return allProducts, nil
}

// ============================================================================
// CATEGORY-SPECIFIC PRODUCT GENERATORS
// ============================================================================

// createRiceProducts - Beras (Rice)
func createRiceProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	riceData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Beras Premium Ramos 5kg", 50000, 60000},
		{"Beras Premium Pandan Wangi 5kg", 48000, 58000},
		{"Beras Medium Setra Ramos 5kg", 35000, 42000},
		{"Beras Medium IR 64 5kg", 32000, 39000},
		{"Beras Economy C4 5kg", 25000, 30000},
		{"Beras Pulen Maknyus 10kg", 80000, 95000},
		{"Beras Organik Premium 2kg", 35000, 42000},
		{"Beras Merah Organik 1kg", 22000, 27000},
	}

	for i := 0; i < count && i < len(riceData); i++ {
		data := riceData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("BRS-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Beras"),
			BaseUnit:     "KG",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(100),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createOilProducts - Minyak Goreng (Cooking Oil)
func createOilProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	oilData := []struct {
		name  string
		cost  int64
		price int64
		unit  string
	}{
		{"Minyak Goreng Bimoli 2L", 28000, 32000, "LITER"},
		{"Minyak Goreng Tropical 2L", 27000, 31000, "LITER"},
		{"Minyak Goreng Sania 1L", 15000, 18000, "LITER"},
		{"Minyak Goreng Fortune 5L", 65000, 75000, "LITER"},
		{"Minyak Goreng Filma 2L", 26000, 30000, "LITER"},
		{"Minyak Goreng Sunco 1L", 14000, 17000, "LITER"},
		{"Minyak Kelapa Barco 1L", 35000, 42000, "LITER"},
		{"Minyak Zaitun Extra Virgin 500ml", 75000, 90000, "LITER"},
	}

	for i := 0; i < count && i < len(oilData); i++ {
		data := oilData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("MNY-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Minyak Goreng"),
			BaseUnit:     data.unit,
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(50),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createSugarProducts - Gula (Sugar)
func createSugarProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	sugarData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Gula Pasir Gulaku 1kg", 12000, 15000},
		{"Gula Pasir Premium 1kg", 13000, 16000},
		{"Gula Pasir Lokal 1kg", 11000, 14000},
		{"Gula Merah Aren 500g", 18000, 22000},
		{"Gula Jawa Asli 500g", 15000, 19000},
		{"Gula Batu Rock Sugar 250g", 8000, 10000},
	}

	for i := 0; i < count && i < len(sugarData); i++ {
		data := sugarData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("GUL-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Gula"),
			BaseUnit:     "KG",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(80),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createFlourProducts - Tepung (Flour)
func createFlourProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	flourData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Tepung Terigu Segitiga Biru 1kg", 10000, 12500},
		{"Tepung Terigu Cakra Kembar 1kg", 11000, 13500},
		{"Tepung Beras Rose Brand 500g", 8000, 10000},
		{"Tepung Tapioka Cap Pak Tani 500g", 7000, 9000},
		{"Tepung Maizena Maizenaku 400g", 9000, 11000},
		{"Tepung Ketan Rose Brand 500g", 12000, 15000},
	}

	for i := 0; i < count && i < len(flourData); i++ {
		data := flourData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("TPG-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Tepung"),
			BaseUnit:     "KG",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(60),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createNoodleProducts - Mi Instan (Instant Noodles)
func createNoodleProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	noodleData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Indomie Goreng", 2500, 3000},
		{"Indomie Soto", 2500, 3000},
		{"Indomie Kari Ayam", 2500, 3000},
		{"Mie Sedaap Goreng", 2400, 2900},
		{"Mie Sedaap Kari Spesial", 2400, 2900},
		{"Sarimi Ayam Bawang", 2300, 2800},
		{"Mie Gelas Rasa Soto", 2000, 2500},
		{"Pop Mie Rasa Ayam", 3500, 4200},
		{"Mie Sedaap Cup Goreng", 3800, 4500},
		{"Supermie Ayam Bawang", 2200, 2700},
	}

	for i := 0; i < count && i < len(noodleData); i++ {
		data := noodleData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("MIE-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Mi Instan"),
			BaseUnit:     "PCS",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(200),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createBeverageProducts - Minuman (Beverages)
func createBeverageProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	beverageData := []struct {
		name       string
		cost       int64
		price      int64
		perishable bool
	}{
		{"Teh Botol Sosro 450ml", 3500, 4200, false},
		{"Coca Cola 390ml", 4000, 4800, false},
		{"Fanta Orange 390ml", 4000, 4800, false},
		{"Aqua Botol 600ml", 2500, 3000, false},
		{"Fruit Tea Apple 500ml", 3800, 4500, false},
		{"Pocari Sweat 500ml", 6000, 7200, false},
		{"ABC Kopi Susu 200ml", 2000, 2500, true},
		{"Yakult 5 Botol", 8500, 10000, true},
	}

	for i := 0; i < count && i < len(beverageData); i++ {
		data := beverageData[i]
		products = append(products, models.Product{
			TenantID:       tenantID,
			CompanyID:      companyID,
			Code:           fmt.Sprintf("MNM-%03d", (companyNum*100)+i+1),
			Name:           data.name,
			Category:       stringPtr("Minuman"),
			BaseUnit:       "PCS",
			BaseCost:       decimal.NewFromInt(data.cost),
			BasePrice:      decimal.NewFromInt(data.price),
			MinimumStock:   decimal.NewFromInt(100),
			IsActive:       true,
			IsPerishable:   data.perishable,
			IsBatchTracked: data.perishable,
		})
	}
	return products
}

// createSnackProducts - Snack & Kue (Snacks)
func createSnackProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	snackData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Chitato Rasa Sapi Panggang 68g", 8500, 10000},
		{"Cheetos Cheddar 60g", 7500, 9000},
		{"Tango Wafer Coklat 115g", 4500, 5500},
		{"Roma Kelapa 300g", 9000, 11000},
		{"Khong Guan Assorted 350g", 18000, 22000},
		{"Oreo Vanilla 137g", 8000, 9500},
		{"Biskuat Coklat 120g", 6500, 8000},
		{"Better Kacang Ijo 180g", 7000, 8500},
	}

	for i := 0; i < count && i < len(snackData); i++ {
		data := snackData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("SNK-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Snack & Kue"),
			BaseUnit:     "PACK",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(80),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createSpiceProducts - Bumbu Dapur (Spices)
func createSpiceProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	spiceData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Royco Ayam 100g", 8500, 10500},
		{"Masako Sapi 250g", 15000, 18000},
		{"Sajiku Nasi Goreng 20g", 2000, 2500},
		{"Bawang Putih Bubuk 50g", 6000, 7500},
		{"Lada Hitam Bubuk 45g", 12000, 15000},
		{"Ketumbar Bubuk 40g", 8000, 10000},
		{"Kunyit Bubuk 50g", 7000, 9000},
		{"Garam Halus Refina 250g", 2500, 3000},
	}

	for i := 0; i < count && i < len(spiceData); i++ {
		data := spiceData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("BMB-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Bumbu Dapur"),
			BaseUnit:     "PACK",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(50),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createSauceProducts - Kecap & Saus (Sauces)
func createSauceProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	sauceData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Kecap Manis Bango 220ml", 8500, 10500},
		{"Kecap Asin Cap Ikan 140ml", 6000, 7500},
		{"Saus Sambal ABC 335ml", 12000, 14500},
		{"Saus Tomat Del Monte 340g", 13000, 16000},
		{"Saos Tiram Lee Kum Kee 255g", 18000, 22000},
		{"Saus Cabai Indofood 340ml", 11000, 13500},
	}

	for i := 0; i < count && i < len(sauceData); i++ {
		data := sauceData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("SAS-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Kecap & Saus"),
			BaseUnit:     "BOTTLE",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(40),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createDairyProducts - Susu & Dairy (Milk Products)
func createDairyProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	dairyData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Susu UHT Ultra Milk Full Cream 1L", 16000, 19500},
		{"Susu Kental Manis Carnation 370g", 11000, 13500},
		{"Susu Kental Manis Indomilk 385g", 10000, 12500},
		{"Susu Dancow Fortigro 800g", 65000, 78000},
		{"Keju Kraft Cheddar 165g", 28000, 34000},
		{"Susu Bear Brand 189ml", 9000, 11000},
	}

	for i := 0; i < count && i < len(dairyData); i++ {
		data := dairyData[i]
		products = append(products, models.Product{
			TenantID:       tenantID,
			CompanyID:      companyID,
			Code:           fmt.Sprintf("SSU-%03d", (companyNum*100)+i+1),
			Name:           data.name,
			Category:       stringPtr("Susu & Dairy"),
			BaseUnit:       "PCS",
			BaseCost:       decimal.NewFromInt(data.cost),
			BasePrice:      decimal.NewFromInt(data.price),
			MinimumStock:   decimal.NewFromInt(60),
			IsActive:       true,
			IsPerishable:   true,
			IsBatchTracked: true,
		})
	}
	return products
}

// createCoffeeTeaProducts - Kopi & Teh (Coffee & Tea)
func createCoffeeTeaProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	coffeeTeaData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Kapal Api Special Mix 10x25g", 12000, 15000},
		{"Nescafe Classic 100g", 35000, 42000},
		{"Good Day Cappuccino 10x20g", 11000, 13500},
		{"Teh Sariwangi 25 Kantong", 8500, 10500},
		{"Teh Celup Sosro 25 Bags", 9000, 11000},
		{"Kopi ABC Susu 10x31g", 10000, 12500},
	}

	for i := 0; i < count && i < len(coffeeTeaData); i++ {
		data := coffeeTeaData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("KOP-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Kopi & Teh"),
			BaseUnit:     "PACK",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(50),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createSoapProducts - Sabun & Deterjen (Soap & Detergent)
func createSoapProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	soapData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Rinso Deterjen Bubuk 800g", 18000, 22000},
		{"So Klin Pewangi 800g", 16000, 19500},
		{"Sabun Lifebuoy 85g", 3500, 4500},
		{"Sabun Nuvo 85g", 3000, 4000},
		{"Sunlight Sabun Cuci Piring 800ml", 15000, 18500},
		{"Molto Ultra Sekali Bilas 900ml", 22000, 27000},
	}

	for i := 0; i < count && i < len(soapData); i++ {
		data := soapData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("SBN-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Sabun & Deterjen"),
			BaseUnit:     "PCS",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(40),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createHouseholdProducts - Perlengkapan Rumah (Household Items)
func createHouseholdProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	householdData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Tissue Paseo 250 sheets", 8500, 10500},
		{"Baygon Aerosol 600ml", 28000, 34000},
		{"Stella Pewangi Pakaian 900ml", 12000, 15000},
		{"Kispray Lavender 300ml", 18000, 22000},
	}

	for i := 0; i < count && i < len(householdData); i++ {
		data := householdData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("PLK-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Perlengkapan Rumah"),
			BaseUnit:     "PCS",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(30),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createCannedProducts - Makanan Kaleng (Canned Food)
func createCannedProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	cannedData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Sarden ABC 155g", 9000, 11000},
		{"Kornet Pronas 198g", 18000, 22000},
		{"Buah Kaleng Delmonte Peach 480g", 32000, 39000},
		{"Susu Kental Manis Frisian Flag 370g", 10500, 13000},
	}

	for i := 0; i < count && i < len(cannedData); i++ {
		data := cannedData[i]
		products = append(products, models.Product{
			TenantID:     tenantID,
			CompanyID:    companyID,
			Code:         fmt.Sprintf("KLG-%03d", (companyNum*100)+i+1),
			Name:         data.name,
			Category:     stringPtr("Makanan Kaleng"),
			BaseUnit:     "PCS",
			BaseCost:     decimal.NewFromInt(data.cost),
			BasePrice:    decimal.NewFromInt(data.price),
			MinimumStock: decimal.NewFromInt(40),
			IsActive:     true,
			IsPerishable: false,
			IsBatchTracked: false,
		})
	}
	return products
}

// createFrozenProducts - Frozen Food
func createFrozenProducts(tenantID, companyID string, companyNum, count int) []models.Product {
	products := []models.Product{}
	frozenData := []struct {
		name  string
		cost  int64
		price int64
	}{
		{"Nugget Fiesta 500g", 28000, 34000},
		{"Sosis So Nice 375g", 25000, 30000},
		{"Bakso Ikan Frozen 500g", 32000, 39000},
		{"Es Krim Walls Viennetta 900ml", 45000, 55000},
	}

	for i := 0; i < count && i < len(frozenData); i++ {
		data := frozenData[i]
		products = append(products, models.Product{
			TenantID:       tenantID,
			CompanyID:      companyID,
			Code:           fmt.Sprintf("FRZ-%03d", (companyNum*100)+i+1),
			Name:           data.name,
			Category:       stringPtr("Frozen Food"),
			BaseUnit:       "PACK",
			BaseCost:       decimal.NewFromInt(data.cost),
			BasePrice:      decimal.NewFromInt(data.price),
			MinimumStock:   decimal.NewFromInt(30),
			IsActive:       true,
			IsPerishable:   true,
			IsBatchTracked: true,
		})
	}
	return products
}

// ============================================================================
// PRODUCT UNITS SEEDING (Multi-Unit Support)
// ============================================================================

func seedProductUnits(db *gorm.DB, products []models.Product) ([]models.ProductUnit, error) {
	productUnits := []models.ProductUnit{}

	// Find products by code prefix for multi-unit setup
	for _, product := range products {
		// Multi-unit for instant noodles (top 3)
		if product.Code == "MIE-101" || product.Code == "MIE-102" || product.Code == "MIE-103" {
			productUnits = append(productUnits, createNoodleUnits(product)...)
		}

		// Multi-unit for rice (top 2)
		if product.Code == "BRS-101" || product.Code == "BRS-102" {
			productUnits = append(productUnits, createRiceUnits(product)...)
		}

		// Multi-unit for cooking oil (top 2)
		if product.Code == "MNY-101" || product.Code == "MNY-102" {
			productUnits = append(productUnits, createOilUnits(product)...)
		}

		// Multi-unit for sugar
		if product.Code == "GUL-101" {
			productUnits = append(productUnits, createSugarUnits(product)...)
		}

		// Multi-unit for flour
		if product.Code == "TPG-101" {
			productUnits = append(productUnits, createFlourUnits(product)...)
		}
	}

	if len(productUnits) > 0 {
		if err := db.CreateInBatches(productUnits, 20).Error; err != nil {
			return nil, fmt.Errorf("failed to insert product units: %w", err)
		}
	}

	return productUnits, nil
}

func createNoodleUnits(product models.Product) []models.ProductUnit {
	basePrice := product.BasePrice.IntPart()
	return []models.ProductUnit{
		{
			ProductID:      product.ID,
			UnitName:       "PCS",
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUnit:     true,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice)),
		},
		{
			ProductID:      product.ID,
			UnitName:       "LUSIN",
			ConversionRate: decimal.NewFromInt(12),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 12 * 97 / 100)), // 3% discount
		},
		{
			ProductID:      product.ID,
			UnitName:       "KARTON",
			ConversionRate: decimal.NewFromInt(40),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 40 * 95 / 100)), // 5% discount
		},
	}
}

func createRiceUnits(product models.Product) []models.ProductUnit {
	basePrice := product.BasePrice.IntPart()
	return []models.ProductUnit{
		{
			ProductID:      product.ID,
			UnitName:       "KG",
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUnit:     true,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice)),
		},
		{
			ProductID:      product.ID,
			UnitName:       "SACK",
			ConversionRate: decimal.NewFromInt(25),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 25 * 95 / 100)), // 5% discount
		},
		{
			ProductID:      product.ID,
			UnitName:       "KARUNG",
			ConversionRate: decimal.NewFromInt(50),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 50 * 92 / 100)), // 8% discount
		},
	}
}

func createOilUnits(product models.Product) []models.ProductUnit {
	basePrice := product.BasePrice.IntPart()
	return []models.ProductUnit{
		{
			ProductID:      product.ID,
			UnitName:       "LITER",
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUnit:     true,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice)),
		},
		{
			ProductID:      product.ID,
			UnitName:       "JERIGEN",
			ConversionRate: decimal.NewFromInt(5),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 5 * 96 / 100)), // 4% discount
		},
	}
}

func createSugarUnits(product models.Product) []models.ProductUnit {
	basePrice := product.BasePrice.IntPart()
	return []models.ProductUnit{
		{
			ProductID:      product.ID,
			UnitName:       "KG",
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUnit:     true,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice)),
		},
		{
			ProductID:      product.ID,
			UnitName:       "KARUNG",
			ConversionRate: decimal.NewFromInt(50),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 50 * 93 / 100)), // 7% discount
		},
	}
}

func createFlourUnits(product models.Product) []models.ProductUnit {
	basePrice := product.BasePrice.IntPart()
	return []models.ProductUnit{
		{
			ProductID:      product.ID,
			UnitName:       "KG",
			ConversionRate: decimal.NewFromInt(1),
			IsBaseUnit:     true,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice)),
		},
		{
			ProductID:      product.ID,
			UnitName:       "SACK",
			ConversionRate: decimal.NewFromInt(25),
			IsBaseUnit:     false,
			SellPrice:      decimalPtr(decimal.NewFromInt(basePrice * 25 * 94 / 100)), // 6% discount
		},
	}
}

// ============================================================================
// WAREHOUSE STOCK INITIALIZATION
// ============================================================================

func seedWarehouseStock(db *gorm.DB, tenant *models.Tenant, products []models.Product, warehouses []models.Warehouse) ([]models.WarehouseStock, error) {
	warehouseStocks := []models.WarehouseStock{}

	// Create warehouse stock for each product in each warehouse
	for _, product := range products {
		// Find warehouse for this product's company
		var warehouse *models.Warehouse
		for i := range warehouses {
			if warehouses[i].CompanyID == product.CompanyID {
				warehouse = &warehouses[i]
				break
			}
		}

		if warehouse == nil {
			continue // Skip if no warehouse found for this company
		}

		// Determine stock tier based on category
		stockQty := determineInitialStock(product.Category)

		warehouseStock := models.WarehouseStock{
			ProductID:    product.ID,
			WarehouseID:  warehouse.ID,
			Quantity:     stockQty,
			MinimumStock: decimal.NewFromFloat(float64(stockQty.IntPart()) * 0.2), // 20% of initial
			MaximumStock: decimal.NewFromFloat(float64(stockQty.IntPart()) * 2.0), // 200% of initial
		}

		warehouseStocks = append(warehouseStocks, warehouseStock)
	}

	if len(warehouseStocks) > 0 {
		if err := db.CreateInBatches(warehouseStocks, 50).Error; err != nil {
			return nil, fmt.Errorf("failed to insert warehouse stock: %w", err)
		}
	}

	return warehouseStocks, nil
}

func determineInitialStock(category *string) decimal.Decimal {
	if category == nil {
		return decimal.NewFromInt(50) // Default medium stock
	}

	// Fast-moving categories (high stock)
	fastMoving := map[string]bool{
		"Mi Instan":    true,
		"Beras":        true,
		"Minyak Goreng": true,
		"Gula":         true,
	}

	// Slow-moving categories (low stock)
	slowMoving := map[string]bool{
		"Frozen Food":         true,
		"Makanan Kaleng":      true,
		"Perlengkapan Rumah": true,
	}

	if fastMoving[*category] {
		// High stock: 300-500 units
		return decimal.NewFromInt(int64(300 + rand.Intn(200)))
	} else if slowMoving[*category] {
		// Low stock: 30-60 units
		return decimal.NewFromInt(int64(30 + rand.Intn(30)))
	}

	// Medium stock: 80-150 units
	return decimal.NewFromInt(int64(80 + rand.Intn(70)))
}

// ============================================================================
// PRODUCT BATCHES FOR PERISHABLES
// ============================================================================

func seedProductBatches(db *gorm.DB, products []models.Product, warehouseStocks []models.WarehouseStock) ([]models.ProductBatch, error) {
	batches := []models.ProductBatch{}

	// Create map of warehouseStock by productID for quick lookup
	stockMap := make(map[string]models.WarehouseStock)
	for _, stock := range warehouseStocks {
		stockMap[stock.ProductID] = stock
	}

	// Create batches for perishable products
	for _, product := range products {
		if !product.IsPerishable || !product.IsBatchTracked {
			continue
		}

		warehouseStock, exists := stockMap[product.ID]
		if !exists {
			continue
		}

		// Create 1-2 batches per perishable product
		numBatches := 1 + rand.Intn(2) // 1 or 2 batches
		for i := 0; i < numBatches; i++ {
			batch := createProductBatch(product, warehouseStock, i+1)
			batches = append(batches, batch)
		}
	}

	if len(batches) > 0 {
		if err := db.CreateInBatches(batches, 20).Error; err != nil {
			return nil, fmt.Errorf("failed to insert product batches: %w", err)
		}
	}

	return batches, nil
}

func createProductBatch(product models.Product, warehouseStock models.WarehouseStock, batchNum int) models.ProductBatch {
	now := time.Now()

	// Determine expiry period based on category
	var expiryMonths int
	if product.Category != nil {
		switch *product.Category {
		case "Frozen Food":
			expiryMonths = 6 + rand.Intn(6) // 6-12 months
		case "Susu & Dairy":
			expiryMonths = 1 + rand.Intn(2) // 1-3 months
		case "Minuman":
			expiryMonths = 3 + rand.Intn(3) // 3-6 months
		default:
			expiryMonths = 6
		}
	} else {
		expiryMonths = 6
	}

	// Manufacture date: 1-3 months ago
	manufactureMonthsAgo := 1 + rand.Intn(2)
	manufactureDate := now.AddDate(0, -manufactureMonthsAgo, 0)

	// Expiry date: from manufacture date
	expiryDate := manufactureDate.AddDate(0, expiryMonths, 0)

	// Batch quantity: split warehouse stock among batches
	batchQty := warehouseStock.Quantity.Div(decimal.NewFromInt(int64(batchNum + 1)))

	return models.ProductBatch{
		BatchNumber:      fmt.Sprintf("BATCH-%s-%03d", product.Code, batchNum),
		ProductID:        product.ID,
		WarehouseStockID: warehouseStock.ID,
		ManufactureDate:  &manufactureDate,
		ExpiryDate:       &expiryDate,
		Quantity:         batchQty,
		ReceiptDate:      manufactureDate,
		Status:           models.BatchStatusAvailable,
		QualityStatus:    stringPtr("GOOD"),
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func stringPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func printSeedingSummary() {
	fmt.Println("\n📊 SEEDING SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("✓ Products:         100 records across 15 categories")
	fmt.Println("✓ Product Units:    ~35 records (multi-unit support)")
	fmt.Println("✓ Warehouse Stock:  ~100 records (tiered quantities)")
	fmt.Println("✓ Product Batches:  ~20 records (perishable items)")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("\n📋 CATEGORY BREAKDOWN:")
	fmt.Println("  • Beras (Rice):            10 products")
	fmt.Println("  • Minyak Goreng:           8 products")
	fmt.Println("  • Gula (Sugar):            6 products")
	fmt.Println("  • Tepung (Flour):          6 products")
	fmt.Println("  • Mi Instan (Noodles):     10 products")
	fmt.Println("  • Minuman (Beverages):     8 products")
	fmt.Println("  • Snack & Kue:             8 products")
	fmt.Println("  • Bumbu Dapur (Spices):    8 products")
	fmt.Println("  • Kecap & Saus:            6 products")
	fmt.Println("  • Susu & Dairy:            6 products")
	fmt.Println("  • Kopi & Teh:              6 products")
	fmt.Println("  • Sabun & Deterjen:        6 products")
	fmt.Println("  • Perlengkapan Rumah:      4 products")
	fmt.Println("  • Makanan Kaleng:          4 products")
	fmt.Println("  • Frozen Food:             4 products")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("\n💡 Next Steps:")
	fmt.Println("  1. Verify data: SELECT COUNT(*) FROM products;")
	fmt.Println("  2. Check multi-unit: SELECT * FROM product_units;")
	fmt.Println("  3. View stock: SELECT * FROM warehouse_stock;")
	fmt.Println("  4. Check batches: SELECT * FROM product_batches;")
	fmt.Println("═══════════════════════════════════════════════════════════")
}
