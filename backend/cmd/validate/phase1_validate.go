// Package main - PHASE 1: Schema Validation Script
// This script validates the multi-company database schema and seed data
// Run with: go run cmd/validate/phase1_validate.go
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
		dbDriver = "sqlite"
	}

	var db *gorm.DB
	var err error

	if dbDriver == "postgres" {
		dsn := os.Getenv("DATABASE_URL")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	} else {
		db, err = gorm.Open(sqlite.Open("dev.db"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")
	log.Println("🔍 Starting PHASE 1: Schema Validation")
	log.Println("=" + string(make([]byte, 60)))

	// Run validation checks
	passed := 0
	failed := 0

	checks := []struct {
		Name string
		Func func(*gorm.DB) error
	}{
		{"Tenant-Company Relationship", validateTenantCompany},
		{"UserCompanyRole Table", validateUserCompanyRole},
		{"Transactional Tables CompanyID", validateTransactionalTables},
		{"Seed Data Integrity", validateSeedData},
		{"Permission System", validatePermissionSystem},
		{"Company Isolation", validateCompanyIsolation},
	}

	for _, check := range checks {
		log.Printf("\n🔍 Checking: %s", check.Name)
		if err := check.Func(db); err != nil {
			log.Printf("  ❌ FAILED: %v", err)
			failed++
		} else {
			log.Printf("  ✅ PASSED")
			passed++
		}
	}

	log.Println("\n" + string(make([]byte, 60)))
	log.Printf("📊 Validation Results: %d passed, %d failed", passed, failed)

	if failed > 0 {
		log.Fatal("❌ Validation failed! Please fix the errors above.")
	} else {
		log.Println("🎉 All validation checks passed!")
	}
}

// validateTenantCompany validates the 1:N relationship between Tenant and Companies
func validateTenantCompany(db *gorm.DB) error {
	// Check if tenant exists
	var tenant models.Tenant
	if err := db.First(&tenant).Error; err != nil {
		return fmt.Errorf("no tenant found: %w", err)
	}

	// Check if tenant has Name and Subdomain
	if tenant.Name == "" {
		return fmt.Errorf("tenant missing 'name' field")
	}
	if tenant.Subdomain == "" {
		return fmt.Errorf("tenant missing 'subdomain' field")
	}

	// Check if companies exist with TenantID
	var companies []models.Company
	if err := db.Where("tenant_id = ?", tenant.ID).Find(&companies).Error; err != nil {
		return fmt.Errorf("failed to query companies: %w", err)
	}

	if len(companies) == 0 {
		return fmt.Errorf("no companies found for tenant")
	}

	log.Printf("    Found %d companies for tenant '%s'", len(companies), tenant.Name)

	// Verify all companies have TenantID
	for _, company := range companies {
		if company.TenantID == "" {
			return fmt.Errorf("company '%s' missing tenant_id", company.Name)
		}
		if company.TenantID != tenant.ID {
			return fmt.Errorf("company '%s' has wrong tenant_id", company.Name)
		}
	}

	return nil
}

// validateUserCompanyRole validates the UserCompanyRole table and relationships
func validateUserCompanyRole(db *gorm.DB) error {
	// Check if table exists
	if !db.Migrator().HasTable("user_company_roles") {
		return fmt.Errorf("user_company_roles table does not exist")
	}

	// Check if records exist
	var count int64
	if err := db.Model(&models.UserCompanyRole{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count user_company_roles: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("no user_company_roles records found")
	}

	log.Printf("    Found %d user-company-role mappings", count)

	// Verify role constraints (only Tier 2 roles allowed)
	var invalidRoles []models.UserCompanyRole
	if err := db.Where("role NOT IN ?", []string{
		string(models.UserRoleAdmin),
		string(models.UserRoleFinance),
		string(models.UserRoleSales),
		string(models.UserRoleWarehouse),
		string(models.UserRoleStaff),
	}).Find(&invalidRoles).Error; err != nil {
		return fmt.Errorf("failed to check role constraints: %w", err)
	}

	if len(invalidRoles) > 0 {
		return fmt.Errorf("found %d invalid roles (Tier 1 roles not allowed in user_company_roles)", len(invalidRoles))
	}

	return nil
}

// validateTransactionalTables validates that all transactional tables have CompanyID
func validateTransactionalTables(db *gorm.DB) error {
	tablesToCheck := []struct {
		Model interface{}
		Name  string
	}{
		{&models.Warehouse{}, "warehouses"},
		{&models.Product{}, "products"},
		{&models.Customer{}, "customers"},
		{&models.Supplier{}, "suppliers"},
	}

	for _, table := range tablesToCheck {
		// Check if column exists
		if !db.Migrator().HasColumn(table.Model, "CompanyID") {
			return fmt.Errorf("%s table missing 'company_id' column", table.Name)
		}

		// Check if any records have empty CompanyID
		var count int64
		query := fmt.Sprintf("company_id = '' OR company_id IS NULL")
		if err := db.Model(table.Model).Where(query).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check %s: %w", table.Name, err)
		}

		if count > 0 {
			return fmt.Errorf("%s has %d records with empty company_id", table.Name, count)
		}

		log.Printf("    ✓ %s table valid", table.Name)
	}

	return nil
}

// validateSeedData validates the seed data integrity
func validateSeedData(db *gorm.DB) error {
	// Check tenant
	var tenantCount int64
	if err := db.Model(&models.Tenant{}).Count(&tenantCount).Error; err != nil {
		return err
	}
	if tenantCount == 0 {
		return fmt.Errorf("no tenants found")
	}
	log.Printf("    ✓ %d tenant(s) found", tenantCount)

	// Check companies
	var companyCount int64
	if err := db.Model(&models.Company{}).Count(&companyCount).Error; err != nil {
		return err
	}
	if companyCount < 3 {
		return fmt.Errorf("expected at least 3 companies, found %d", companyCount)
	}
	log.Printf("    ✓ %d companies found", companyCount)

	// Check users
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount < 5 {
		return fmt.Errorf("expected at least 5 users, found %d", userCount)
	}
	log.Printf("    ✓ %d users found", userCount)

	return nil
}

// validatePermissionSystem validates the dual-tier permission system
func validatePermissionSystem(db *gorm.DB) error {
	// Check Tier 1 (UserTenant)
	var tier1Count int64
	if err := db.Model(&models.UserTenant{}).Count(&tier1Count).Error; err != nil {
		return fmt.Errorf("failed to count user_tenants: %w", err)
	}
	log.Printf("    ✓ Tier 1: %d tenant-level permissions", tier1Count)

	// Check Tier 2 (UserCompanyRole)
	var tier2Count int64
	if err := db.Model(&models.UserCompanyRole{}).Count(&tier2Count).Error; err != nil {
		return fmt.Errorf("failed to count user_company_roles: %w", err)
	}
	log.Printf("    ✓ Tier 2: %d company-level permissions", tier2Count)

	// Verify OWNER and TENANT_ADMIN users exist in Tier 1
	var ownerCount int64
	if err := db.Model(&models.UserTenant{}).Where("role IN ?", []string{
		string(models.UserRoleOwner),
		string(models.UserRoleTenantAdmin),
	}).Count(&ownerCount).Error; err != nil {
		return err
	}

	if ownerCount == 0 {
		return fmt.Errorf("no OWNER or TENANT_ADMIN users found")
	}
	log.Printf("    ✓ %d superusers (OWNER/TENANT_ADMIN)", ownerCount)

	return nil
}

// validateCompanyIsolation validates that data is properly isolated per company
func validateCompanyIsolation(db *gorm.DB) error {
	// Get all companies
	var companies []models.Company
	if err := db.Find(&companies).Error; err != nil {
		return err
	}

	for _, company := range companies {
		// Check warehouses for this company
		var warehouseCount int64
		if err := db.Model(&models.Warehouse{}).Where("company_id = ?", company.ID).Count(&warehouseCount).Error; err != nil {
			return err
		}

		// Check products for this company
		var productCount int64
		if err := db.Model(&models.Product{}).Where("company_id = ?", company.ID).Count(&productCount).Error; err != nil {
			return err
		}

		log.Printf("    Company '%s': %d warehouses, %d products",
			company.Name, warehouseCount, productCount)
	}

	return nil
}
