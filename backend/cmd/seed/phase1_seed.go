// Package main - PHASE 1: Seed Data Script
// This script seeds the database with test data for multi-company architecture
// Run with: go run cmd/seed/phase1_seed.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/internal/config"
	"backend/models"
	"backend/pkg/security"
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
	log.Println("🌱 Starting PHASE 1: Seed Data")
	log.Println("=" + string(make([]byte, 60)))

	// Seed data
	if err := seedData(db); err != nil {
		log.Fatalf("❌ Seed failed: %v", err)
	}

	log.Println("\n🎉 Seed completed successfully!")
}

func seedData(db *gorm.DB) error {
	// 1. Create Tenant
	log.Println("\n📋 Creating Tenant...")
	tenant := &models.Tenant{
		Name:      "PT Multi Bisnis Group",
		Subdomain: "multi-bisnis",
		Status:    models.TenantStatusActive,
	}
	if err := db.FirstOrCreate(tenant, models.Tenant{Subdomain: tenant.Subdomain}).Error; err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	log.Printf("  ✓ Tenant created: %s (ID: %s)", tenant.Name, tenant.ID)

	// 2. Create Companies
	log.Println("\n📋 Creating Companies...")
	companies := []*models.Company{
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
		{
			TenantID:   tenant.ID,
			Name:       "PT Retail Nusantara",
			LegalName:  "PT Retail Nusantara Sejahtera",
			EntityType: "PT",
			Address:    "Jl. Sudirman No. 789",
			City:       "Jakarta Selatan",
			Province:   "DKI Jakarta",
			Phone:      "021-11223344",
			Email:      "info@retail-nusantara.com",
			IsActive:   true,
		},
	}

	for _, company := range companies {
		if err := db.FirstOrCreate(company, models.Company{
			TenantID: tenant.ID,
			Name:     company.Name,
		}).Error; err != nil {
			return fmt.Errorf("failed to create company %s: %w", company.Name, err)
		}
		log.Printf("  ✓ Company created: %s (ID: %s)", company.Name, company.ID)
	}

	// 3. Create Users
	log.Println("\n📋 Creating Users...")

	// Use Argon2id password hasher (same as auth service)
	argon2Config := config.Argon2Config{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
	passwordHasher := security.NewPasswordHasher(argon2Config)
	hashedPassword, err := passwordHasher.HashPassword("password123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []*models.User{
		{
			Email:        "budi@example.com",
			Username:     "budi.santoso",
			PasswordHash: hashedPassword,
			FullName:     "Budi Santoso",
			IsActive:     true,
		},
		{
			Email:        "siti@example.com",
			Username:     "siti.rahayu",
			PasswordHash: hashedPassword,
			FullName:     "Siti Rahayu",
			IsActive:     true,
		},
		{
			Email:        "ahmad@example.com",
			Username:     "ahmad.fauzi",
			PasswordHash: hashedPassword,
			FullName:     "Ahmad Fauzi",
			IsActive:     true,
		},
		{
			Email:        "joko@example.com",
			Username:     "joko.widodo",
			PasswordHash: hashedPassword,
			FullName:     "Joko Widodo",
			IsActive:     true,
		},
		{
			Email:        "dewi@example.com",
			Username:     "dewi.lestari",
			PasswordHash: hashedPassword,
			FullName:     "Dewi Lestari",
			IsActive:     true,
		},
	}

	for _, user := range users {
		if err := db.FirstOrCreate(user, models.User{Email: user.Email}).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
		log.Printf("  ✓ User created: %s (%s)", user.FullName, user.Email)
	}

	// 4. Create UserTenant (Tier 1 - Tenant Level)
	log.Println("\n📋 Creating User-Tenant Relationships (Tier 1)...")
	userTenants := []*models.UserTenant{
		// Only Budi has Tier 1 access (OWNER) - gets access to ALL companies
		{UserID: users[0].ID, TenantID: tenant.ID, Role: models.UserRoleOwner, IsActive: true},
		// Siti removed from Tier 1 - she should only have Tier 2 (per-company) access
		// This ensures Siti only sees companies she has explicit UserCompanyRole for (2 companies, not all 3)
	}

	for _, ut := range userTenants {
		if err := db.FirstOrCreate(ut, models.UserTenant{
			UserID:   ut.UserID,
			TenantID: ut.TenantID,
		}).Error; err != nil {
			return fmt.Errorf("failed to create user-tenant: %w", err)
		}
		log.Printf("  ✓ User-Tenant: %s → Tenant (Role: %s)", ut.UserID, ut.Role)
	}

	// 5. Create UserCompanyRoles (Tier 2 - Per-Company Access)
	log.Println("\n📋 Creating User-Company Roles (Tier 2)...")

	// Scenario examples from the analysis document
	userCompanyRoles := []*models.UserCompanyRole{
		// Siti Rahayu: ADMIN at PT Distribusi, STAFF at CV Sembako
		{UserID: users[1].ID, CompanyID: companies[0].ID, TenantID: tenant.ID, Role: models.UserRoleAdmin, IsActive: true},
		{UserID: users[1].ID, CompanyID: companies[1].ID, TenantID: tenant.ID, Role: models.UserRoleStaff, IsActive: true},

		// Ahmad Fauzi: FINANCE only at CV Sembako
		{UserID: users[2].ID, CompanyID: companies[1].ID, TenantID: tenant.ID, Role: models.UserRoleFinance, IsActive: true},

		// Joko Widodo: WAREHOUSE at PT Distribusi and CV Sembako
		{UserID: users[3].ID, CompanyID: companies[0].ID, TenantID: tenant.ID, Role: models.UserRoleWarehouse, IsActive: true},
		{UserID: users[3].ID, CompanyID: companies[1].ID, TenantID: tenant.ID, Role: models.UserRoleWarehouse, IsActive: true},

		// Dewi Lestari: SALES ONLY at PT Retail Nusantara
		{UserID: users[4].ID, CompanyID: companies[2].ID, TenantID: tenant.ID, Role: models.UserRoleSales, IsActive: true},
	}

	for _, ucr := range userCompanyRoles {
		if err := db.FirstOrCreate(ucr, models.UserCompanyRole{
			UserID:    ucr.UserID,
			CompanyID: ucr.CompanyID,
		}).Error; err != nil {
			return fmt.Errorf("failed to create user-company-role: %w", err)
		}
		log.Printf("  ✓ User-Company-Role: %s → Company %s (Role: %s)", ucr.UserID[:8], ucr.CompanyID[:8], ucr.Role)
	}

	// 6. Create sample Warehouses (per company)
	log.Println("\n📋 Creating Warehouses...")
	warehouses := []*models.Warehouse{
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

	for _, wh := range warehouses {
		if err := db.FirstOrCreate(wh, models.Warehouse{
			CompanyID: wh.CompanyID,
			Code:      wh.Code,
		}).Error; err != nil {
			return fmt.Errorf("failed to create warehouse: %w", err)
		}
		log.Printf("  ✓ Warehouse: %s (%s)", wh.Name, wh.Code)
	}

	// 7. Create sample Products (100 items for sembako distribution)
	log.Println("\n📋 Creating 100 Products...")
	products := []*models.Product{
		// BERAS (Rice) - 15 items
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BRS-001", Name: "Beras Premium 5kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(50000), BasePrice: decimal.NewFromInt(60000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BRS-002", Name: "Beras Medium 5kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(45000), BasePrice: decimal.NewFromInt(54000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BRS-003", Name: "Beras Premium 10kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(95000), BasePrice: decimal.NewFromInt(115000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BRS-004", Name: "Beras Medium 10kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(85000), BasePrice: decimal.NewFromInt(102000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BRS-005", Name: "Beras Premium 25kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(220000), BasePrice: decimal.NewFromInt(265000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BRS-006", Name: "Beras Pandan Wangi 5kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(55000), BasePrice: decimal.NewFromInt(66000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BRS-007", Name: "Beras Pulen 5kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(48000), BasePrice: decimal.NewFromInt(58000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BRS-008", Name: "Beras Merah 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BRS-009", Name: "Beras Hitam 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(25000), BasePrice: decimal.NewFromInt(30000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BRS-010", Name: "Beras Organik 2kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(35000), BasePrice: decimal.NewFromInt(42000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BRS-011", Name: "Beras Ketan Putih 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(14000), BasePrice: decimal.NewFromInt(17000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BRS-012", Name: "Beras Ketan Hitam 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(16000), BasePrice: decimal.NewFromInt(19000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BRS-013", Name: "Beras Rojolele 5kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(52000), BasePrice: decimal.NewFromInt(62000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BRS-014", Name: "Beras IR64 25kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(200000), BasePrice: decimal.NewFromInt(240000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BRS-015", Name: "Beras Jasmine 5kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(60000), BasePrice: decimal.NewFromInt(72000), IsActive: true, IsPerishable: false},

		// MINYAK GORENG (Cooking Oil) - 12 items
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MNY-001", Name: "Minyak Goreng 2L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(28000), BasePrice: decimal.NewFromInt(32000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNY-002", Name: "Minyak Goreng 1L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(17000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MNY-003", Name: "Minyak Goreng 5L", BaseUnit: "JERIGEN", BaseCost: decimal.NewFromInt(70000), BasePrice: decimal.NewFromInt(80000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MNY-004", Name: "Minyak Kelapa 1L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(35000), BasePrice: decimal.NewFromInt(40000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNY-005", Name: "Minyak Zaitun 500ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(45000), BasePrice: decimal.NewFromInt(52000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MNY-006", Name: "Minyak Sayur 2L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(30000), BasePrice: decimal.NewFromInt(35000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MNY-007", Name: "Minyak Curah 1L", BaseUnit: "LITER", BaseCost: decimal.NewFromInt(14000), BasePrice: decimal.NewFromInt(16000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNY-008", Name: "Minyak Jagung 1L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(22000), BasePrice: decimal.NewFromInt(26000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MNY-009", Name: "Minyak Wijen 250ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MNY-010", Name: "Minyak Goreng Premium 2L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(32000), BasePrice: decimal.NewFromInt(37000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNY-011", Name: "Minyak Goreng Refill 1L", BaseUnit: "POUCH", BaseCost: decimal.NewFromInt(13000), BasePrice: decimal.NewFromInt(15000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MNY-012", Name: "Minyak Goreng Refill 2L", BaseUnit: "POUCH", BaseCost: decimal.NewFromInt(26000), BasePrice: decimal.NewFromInt(30000), IsActive: true, IsPerishable: false},

		// GULA (Sugar) - 8 items
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "GUL-001", Name: "Gula Pasir 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(14000), BasePrice: decimal.NewFromInt(16000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "GUL-002", Name: "Gula Pasir Premium 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(17500), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "GUL-003", Name: "Gula Merah 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(14000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "GUL-004", Name: "Gula Aren 250g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "GUL-005", Name: "Gula Batu 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(16000), BasePrice: decimal.NewFromInt(19000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "GUL-006", Name: "Gula Jawa 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(18000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "GUL-007", Name: "Gula Diet 250g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(25000), BasePrice: decimal.NewFromInt(30000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "GUL-008", Name: "Gula Pasir 50kg", BaseUnit: "SACK", BaseCost: decimal.NewFromInt(650000), BasePrice: decimal.NewFromInt(780000), IsActive: true, IsPerishable: false},

		// TEPUNG (Flour) - 10 items
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "TPG-001", Name: "Tepung Terigu Segitiga Biru 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(11000), BasePrice: decimal.NewFromInt(13000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "TPG-002", Name: "Tepung Terigu Cakra Kembar 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(14000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "TPG-003", Name: "Tepung Beras 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(8000), BasePrice: decimal.NewFromInt(10000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "TPG-004", Name: "Tepung Maizena 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(9000), BasePrice: decimal.NewFromInt(11000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "TPG-005", Name: "Tepung Tapioka 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(7000), BasePrice: decimal.NewFromInt(9000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "TPG-006", Name: "Tepung Ketan 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(10000), BasePrice: decimal.NewFromInt(12000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "TPG-007", Name: "Tepung Jagung 250g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(8000), BasePrice: decimal.NewFromInt(10000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "TPG-008", Name: "Tepung Bumbu Serbaguna 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(15000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "TPG-009", Name: "Tepung Panir 200g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(9000), BasePrice: decimal.NewFromInt(11000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "TPG-010", Name: "Tepung Terigu Kunci Biru 1kg", BaseUnit: "KG", BaseCost: decimal.NewFromInt(10000), BasePrice: decimal.NewFromInt(12000), IsActive: true, IsPerishable: false},

		// MIE INSTAN (Instant Noodles) - 12 items
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MIE-001", Name: "Indomie Goreng isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(90000), BasePrice: decimal.NewFromInt(108000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MIE-002", Name: "Indomie Soto isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(88000), BasePrice: decimal.NewFromInt(105600), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MIE-003", Name: "Mie Sedaap Goreng isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(85000), BasePrice: decimal.NewFromInt(102000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MIE-004", Name: "Mie Sedaap Kari isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(85000), BasePrice: decimal.NewFromInt(102000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MIE-005", Name: "Supermie Ayam Bawang isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(82000), BasePrice: decimal.NewFromInt(98400), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MIE-006", Name: "Sarimi Soto isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(78000), BasePrice: decimal.NewFromInt(93600), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MIE-007", Name: "Pop Mie Rasa Ayam isi 24", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(95000), BasePrice: decimal.NewFromInt(114000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MIE-008", Name: "Indomie Jumbo isi 24", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(105000), BasePrice: decimal.NewFromInt(126000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MIE-009", Name: "Lemonilo Mie Instant isi 30", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(120000), BasePrice: decimal.NewFromInt(144000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MIE-010", Name: "Indomie Goreng Pedas isi 40", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(92000), BasePrice: decimal.NewFromInt(110400), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MIE-011", Name: "Mie Gelas Jumbo isi 24", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(88000), BasePrice: decimal.NewFromInt(105600), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MIE-012", Name: "Mie Shirataki 200g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: true},

		// KOPI & TEH (Coffee & Tea) - 10 items
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "KOP-001", Name: "Kopi Kapal Api Special isi 30", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(65000), BasePrice: decimal.NewFromInt(78000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "KOP-002", Name: "Kopi ABC Susu isi 30", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(68000), BasePrice: decimal.NewFromInt(81600), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "KOP-003", Name: "Nescafe Classic 100g", BaseUnit: "JAR", BaseCost: decimal.NewFromInt(38000), BasePrice: decimal.NewFromInt(45000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "KOP-004", Name: "Good Day Cappuccino isi 30", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(72000), BasePrice: decimal.NewFromInt(86400), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "KOP-005", Name: "Kopi Bubuk Robusta 250g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(25000), BasePrice: decimal.NewFromInt(30000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "TEH-001", Name: "Teh Sariwangi isi 50", BaseUnit: "BOX", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(18000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "TEH-002", Name: "Teh Celup Sosro isi 50", BaseUnit: "BOX", BaseCost: decimal.NewFromInt(16000), BasePrice: decimal.NewFromInt(19000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "TEH-003", Name: "Teh Poci 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(8000), BasePrice: decimal.NewFromInt(10000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "TEH-004", Name: "Teh Hijau isi 30", BaseUnit: "BOX", BaseCost: decimal.NewFromInt(22000), BasePrice: decimal.NewFromInt(26000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "TEH-005", Name: "Teh Javana 500g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},

		// SUSU (Milk) - 8 items
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "SSU-001", Name: "Susu Kental Manis Indomilk 370g", BaseUnit: "CAN", BaseCost: decimal.NewFromInt(9000), BasePrice: decimal.NewFromInt(11000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "SSU-002", Name: "Susu Kental Manis Frisian Flag 370g", BaseUnit: "CAN", BaseCost: decimal.NewFromInt(9500), BasePrice: decimal.NewFromInt(11500), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "SSU-003", Name: "Susu Bubuk Dancow 800g", BaseUnit: "BOX", BaseCost: decimal.NewFromInt(75000), BasePrice: decimal.NewFromInt(90000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "SSU-004", Name: "Susu UHT Ultra Milk 1L", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(16000), BasePrice: decimal.NewFromInt(19000), IsActive: true, IsPerishable: true},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "SSU-005", Name: "Susu UHT Indomilk 1L", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(18000), IsActive: true, IsPerishable: true},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "SSU-006", Name: "Susu Bubuk Bendera 400g", BaseUnit: "BOX", BaseCost: decimal.NewFromInt(35000), BasePrice: decimal.NewFromInt(42000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "SSU-007", Name: "Susu Kedelai 1L", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(14500), IsActive: true, IsPerishable: true},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "SSU-008", Name: "Susu Evaporasi Carnation 385g", BaseUnit: "CAN", BaseCost: decimal.NewFromInt(11000), BasePrice: decimal.NewFromInt(13000), IsActive: true, IsPerishable: false},

		// BUMBU & PENYEDAP (Spices & Seasonings) - 10 items
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BMB-001", Name: "Royco Ayam isi 48", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(45000), BasePrice: decimal.NewFromInt(54000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BMB-002", Name: "Royco Sapi isi 48", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(45000), BasePrice: decimal.NewFromInt(54000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BMB-003", Name: "Masako Ayam isi 48", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(42000), BasePrice: decimal.NewFromInt(50400), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BMB-004", Name: "Sasa Bumbu Soto 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(4000), BasePrice: decimal.NewFromInt(5000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BMB-005", Name: "Sasa Bumbu Rawon 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(4000), BasePrice: decimal.NewFromInt(5000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BMB-006", Name: "Merica Bubuk 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(15000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BMB-007", Name: "Ketumbar Bubuk 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(8000), BasePrice: decimal.NewFromInt(10000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "BMB-008", Name: "Bawang Putih Bubuk 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(18000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "BMB-009", Name: "Kunyit Bubuk 100g", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(9000), BasePrice: decimal.NewFromInt(11000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "BMB-010", Name: "MSG Sajiku 1kg", BaseUnit: "PACK", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},

		// KECAP & SAOS (Soy Sauce & Sauces) - 8 items
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "KCP-001", Name: "Kecap Manis Bango 600ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "KCP-002", Name: "Kecap Manis ABC 600ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(17000), BasePrice: decimal.NewFromInt(21000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "KCP-003", Name: "Kecap Asin 600ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(15000), BasePrice: decimal.NewFromInt(18000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "SAO-001", Name: "Saos Tomat ABC 340g", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(15000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "SAO-002", Name: "Saos Sambal ABC 340g", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(12000), BasePrice: decimal.NewFromInt(15000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "SAO-003", Name: "Saos Tiram 600ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(22000), BasePrice: decimal.NewFromInt(26000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "SAO-004", Name: "Saos Inggris 300ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "SAO-005", Name: "Kecap Manis Sedang 135ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(5000), BasePrice: decimal.NewFromInt(6500), IsActive: true, IsPerishable: false},

		// MINUMAN (Beverages) - 7 items
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNM-001", Name: "Sirup ABC Cocopandan 650ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(18000), BasePrice: decimal.NewFromInt(22000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MNM-002", Name: "Sirup Marjan Melon 650ml", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(16000), BasePrice: decimal.NewFromInt(20000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MNM-003", Name: "Coca Cola 1.5L", BaseUnit: "BOTTLE", BaseCost: decimal.NewFromInt(8000), BasePrice: decimal.NewFromInt(10000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNM-004", Name: "Teh Botol Sosro 450ml isi 24", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(48000), BasePrice: decimal.NewFromInt(57600), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[1].ID, Code: "MNM-005", Name: "Aqua 600ml isi 24", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(40000), BasePrice: decimal.NewFromInt(48000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[2].ID, Code: "MNM-006", Name: "Nutrisari Jeruk isi 30", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(35000), BasePrice: decimal.NewFromInt(42000), IsActive: true, IsPerishable: false},
		{TenantID: tenant.ID, CompanyID: companies[0].ID, Code: "MNM-007", Name: "Energen Vanilla isi 30", BaseUnit: "CARTON", BaseCost: decimal.NewFromInt(55000), BasePrice: decimal.NewFromInt(66000), IsActive: true, IsPerishable: false},
	}

	for _, prod := range products {
		if err := db.FirstOrCreate(prod, models.Product{
			CompanyID: prod.CompanyID,
			Code:      prod.Code,
		}).Error; err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
		log.Printf("  ✓ Product: %s (%s)", prod.Name, prod.Code)
	}

	// 7a. Create ProductUnits (Multi-unit definitions)
	log.Println("\n📏 Creating Product Units (Multi-unit definitions)...")
	productUnits := []*models.ProductUnit{}

	// Beras Premium 5kg - Multi-unit setup
	// Base unit: SACK (Karung/Sak)
	// 1 KG = 0.2 SACK (1 SACK = 5 KG)
	sellPriceKg := decimal.NewFromInt(12500) // Rp 12,500/kg
	productUnits = append(productUnits,
		&models.ProductUnit{
			ProductID:      products[0].ID, // Beras Premium 5kg
			UnitName:       "KG",
			ConversionRate: decimal.NewFromFloat(0.2), // 1 KG = 0.2 SACK
			SellPrice:      &sellPriceKg,
			IsActive:       true,
		},
	)

	// Minyak Goreng 2L - Multi-unit setup
	// Base unit: BOTTLE (Botol)
	// 1 CARTON = 12 BOTTLE
	sellPriceCarton := decimal.NewFromInt(360000) // Rp 360,000/karton
	productUnits = append(productUnits,
		&models.ProductUnit{
			ProductID:      products[1].ID, // Minyak Goreng 2L
			UnitName:       "CARTON",
			ConversionRate: decimal.NewFromFloat(12.0), // 1 CARTON = 12 BOTTLE
			SellPrice:      &sellPriceCarton,
			IsActive:       true,
		},
	)

	for _, unit := range productUnits {
		if err := db.FirstOrCreate(unit, models.ProductUnit{
			ProductID: unit.ProductID,
			UnitName:  unit.UnitName,
		}).Error; err != nil {
			return fmt.Errorf("failed to create product unit: %w", err)
		}
		log.Printf("  ✓ Product Unit: %s for Product %s", unit.UnitName, unit.ProductID[:8])
	}

	// 7b. Create WarehouseStock (Initialize stock in warehouses for all 100 products)
	log.Println("\n🏭 Creating Warehouse Stock for 100 products...")
	warehouseStocks := []*models.WarehouseStock{}

	// Distribute products across 2 warehouses
	// Warehouse 0 (GDG-JKT): Products with even index
	// Warehouse 1 (GDG-PSR): Products with odd index
	locations := []string{"RAK-A-01", "RAK-A-02", "RAK-A-03", "RAK-B-01", "RAK-B-02", "RAK-B-03", "RAK-C-01", "RAK-C-02", "RAK-C-03", "RAK-D-01"}

	for i, prod := range products {
		// 🔧 FIX: Find warehouse that belongs to the same company as the product
		var targetWarehouse *models.Warehouse
		for _, wh := range warehouses {
			if wh.CompanyID == prod.CompanyID {
				targetWarehouse = wh
				break
			}
		}

		// Skip if no matching warehouse found (shouldn't happen with correct seed data)
		if targetWarehouse == nil {
			log.Printf("  ⚠️ WARNING: No warehouse found for product %s (Company %s)", prod.Code, prod.CompanyID[:8])
			continue
		}

		locationIdx := (i / 2) % len(locations)

		// Vary stock quantities based on product type
		var quantity, minStock, maxStock int64
		switch {
		case i < 15: // Rice products - higher stock
			quantity, minStock, maxStock = 100, 20, 200
		case i < 27: // Oil products - medium stock
			quantity, minStock, maxStock = 150, 30, 300
		case i < 35: // Sugar products - medium stock
			quantity, minStock, maxStock = 200, 40, 400
		case i < 45: // Flour products - lower stock
			quantity, minStock, maxStock = 80, 15, 150
		case i < 57: // Instant noodles - high stock (popular item)
			quantity, minStock, maxStock = 300, 50, 500
		case i < 67: // Coffee & Tea - medium stock
			quantity, minStock, maxStock = 120, 25, 250
		case i < 75: // Milk - lower stock (perishable)
			quantity, minStock, maxStock = 60, 10, 100
		case i < 85: // Spices - medium stock
			quantity, minStock, maxStock = 100, 20, 200
		case i < 93: // Sauces - medium stock
			quantity, minStock, maxStock = 90, 18, 180
		default: // Beverages - high stock
			quantity, minStock, maxStock = 200, 40, 400
		}

		warehouseStocks = append(warehouseStocks,
			&models.WarehouseStock{
				ProductID:    prod.ID,
				WarehouseID:  targetWarehouse.ID,  // ✅ NOW CORRECT: Same company!
				Quantity:     decimal.NewFromInt(quantity),
				MinimumStock: decimal.NewFromInt(minStock),
				MaximumStock: decimal.NewFromInt(maxStock),
				Location:     stringPtr(locations[locationIdx]),
			},
		)
	}

	for _, stock := range warehouseStocks {
		if err := db.FirstOrCreate(stock, models.WarehouseStock{
			ProductID:   stock.ProductID,
			WarehouseID: stock.WarehouseID,
		}).Error; err != nil {
			return fmt.Errorf("failed to create warehouse stock: %w", err)
		}
		log.Printf("  ✓ Warehouse Stock: Product %s at Warehouse %s (%s units)",
			stock.ProductID[:8], stock.WarehouseID[:8], stock.Quantity.String())
	}

	// 7c. Create PriceList (Pricing tiers for all 100 products)
	log.Println("\n💰 Creating Price Lists for 100 products (Retail + Wholesale)...")
	priceLists := []*models.PriceList{}

	// Create 2 price tiers for each product:
	// 1. Retail price (MinQty = 1)
	// 2. Wholesale price (MinQty = 10, 5% discount)
	for _, prod := range products {
		retailPrice := prod.BasePrice
		wholesalePrice := prod.BasePrice.Mul(decimal.NewFromFloat(0.95)) // 5% discount

		// Retail price (default for all customers)
		priceLists = append(priceLists,
			&models.PriceList{
				ProductID:     prod.ID,
				CustomerID:    nil, // NULL = default price for all customers
				Price:         retailPrice,
				MinQty:        decimal.NewFromInt(1),
				EffectiveFrom: time.Now(),
				IsActive:      true,
			},
		)

		// Wholesale price (minimum quantity varies by product type)
		var minQty int64
		switch prod.BaseUnit {
		case "CARTON", "BOX":
			minQty = 5 // Min 5 cartons/boxes for wholesale
		case "SACK":
			minQty = 10 // Min 10 sacks for wholesale
		case "KG", "PACK", "BOTTLE", "LITER", "JERIGEN", "POUCH", "CAN", "JAR":
			minQty = 20 // Min 20 units for wholesale
		default:
			minQty = 10 // Default min qty
		}

		priceLists = append(priceLists,
			&models.PriceList{
				ProductID:     prod.ID,
				CustomerID:    nil, // NULL = default price for all customers
				Price:         wholesalePrice,
				MinQty:        decimal.NewFromInt(minQty),
				EffectiveFrom: time.Now(),
				IsActive:      true,
			},
		)
	}

	priceListCount := 0
	for _, priceList := range priceLists {
		if err := db.FirstOrCreate(priceList, models.PriceList{
			ProductID: priceList.ProductID,
			MinQty:    priceList.MinQty,
		}).Error; err != nil {
			return fmt.Errorf("failed to create price list: %w", err)
		}
		priceListCount++
	}
	log.Printf("  ✓ Created %d price list entries (2 tiers × 100 products)", priceListCount)

	// 8. Create sample Customers (per company)
	log.Println("\n📋 Creating Customers...")
	customers := []*models.Customer{
		{
			TenantID:  tenant.ID,
			CompanyID: companies[0].ID,
			Code:      "CUST-001",
			Name:      "Toko Sumber Rejeki",
			Phone:     stringPtr("081234567890"),
			Address:   stringPtr("Jl. Mangga Besar No. 10"),
			City:      stringPtr("Jakarta Barat"),
			IsActive:  true,
		},
		{
			TenantID:  tenant.ID,
			CompanyID: companies[1].ID,
			Code:      "CUST-001",
			Name:      "Warung Berkah",
			Phone:     stringPtr("081298765432"),
			Address:   stringPtr("Jl. Tanah Abang No. 25"),
			City:      stringPtr("Jakarta Pusat"),
			IsActive:  true,
		},
	}

	for _, cust := range customers {
		if err := db.FirstOrCreate(cust, models.Customer{
			CompanyID: cust.CompanyID,
			Code:      cust.Code,
		}).Error; err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}
		log.Printf("  ✓ Customer: %s (%s)", cust.Name, cust.Code)
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
