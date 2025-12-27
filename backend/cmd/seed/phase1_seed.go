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

	// 7. Create sample Products (per company)
	log.Println("\n📋 Creating Products...")
	products := []*models.Product{
		{
			TenantID:    tenant.ID,
			CompanyID:   companies[0].ID,
			Code:        "BRS-001",
			Name:        "Beras Premium 5kg",
			BaseUnit:    "SACK",
			BaseCost:    decimal.NewFromInt(50000),
			BasePrice:   decimal.NewFromInt(60000),
			IsActive:    true,
			IsPerishable: false,
		},
		{
			TenantID:    tenant.ID,
			CompanyID:   companies[1].ID,
			Code:        "MNY-001",
			Name:        "Minyak Goreng 2L",
			BaseUnit:    "BOTTLE",
			BaseCost:    decimal.NewFromInt(28000),
			BasePrice:   decimal.NewFromInt(32000),
			IsActive:    true,
			IsPerishable: false,
		},
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
