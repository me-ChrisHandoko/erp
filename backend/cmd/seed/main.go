// Package main - Database seeding script for test users
// Usage: go run cmd/seed/main.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"

	"backend/internal/config"
	"backend/internal/database"
	"backend/models"
	"backend/pkg/security"
)

func main() {
	fmt.Println("🌱 Starting database seeding for test users...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Bypass tenant isolation for seeding (creating global entities)
	db = db.Set("bypass_tenant", true)

	// Initialize password hasher
	hasher := security.NewPasswordHasher(cfg.Argon2)

	// Start seeding
	fmt.Println("\n📦 Creating test data...")

	// Seed users, companies, tenants
	if err := seedTestData(db, hasher, cfg); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	fmt.Println("\n✅ Database seeding completed successfully!")
	fmt.Println("\n📋 Test Users Created:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	printTestUsers()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n💡 Default password for all test users: Password123!")
	fmt.Println("💡 Use these credentials to test login functionality")
}

func seedTestData(db *gorm.DB, hasher *security.PasswordHasher, cfg *config.Config) error {
	// Clean up existing test data (delete in correct order to respect foreign keys)
	// ⚠️ IMPORTANT: Only delete test data with email pattern '@example.com'
	fmt.Println("🧹 Cleaning up existing test data...")

	// Step 1: Delete user_tenants linked to test users
	db.Exec("DELETE FROM user_tenants WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com')")

	// Step 2: Delete user_company_roles linked to test users
	db.Exec("DELETE FROM user_company_roles WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com')")

	// Step 3: Delete test users
	db.Exec("DELETE FROM users WHERE email LIKE '%@example.com'")

	// Step 4: Get test tenant IDs (tenants created by seed script)
	var testTenantIDs []string
	db.Raw("SELECT DISTINCT t.id FROM tenants t WHERE t.subdomain IN ('pt-maju-jaya-distribusi', 'cv-berkah-mandiri')").Scan(&testTenantIDs)

	// Step 5: Delete test data linked to test tenants (in correct order for foreign keys)
	if len(testTenantIDs) > 0 {
		// Delete in order: child tables first, parent tables last
		db.Exec("DELETE FROM product_batches WHERE warehouse_stock_id IN (SELECT id FROM warehouse_stocks WHERE warehouse_id IN (SELECT id FROM warehouses WHERE tenant_id IN (?)))", testTenantIDs)
		db.Exec("DELETE FROM warehouse_stocks WHERE warehouse_id IN (SELECT id FROM warehouses WHERE tenant_id IN (?))", testTenantIDs)
		db.Exec("DELETE FROM product_units WHERE product_id IN (SELECT id FROM products WHERE tenant_id IN (?))", testTenantIDs)
		db.Exec("DELETE FROM price_list WHERE product_id IN (SELECT id FROM products WHERE tenant_id IN (?))", testTenantIDs)
		db.Exec("DELETE FROM product_suppliers WHERE product_id IN (SELECT id FROM products WHERE tenant_id IN (?))", testTenantIDs)
		db.Exec("DELETE FROM products WHERE tenant_id IN (?)", testTenantIDs)
		db.Exec("DELETE FROM warehouses WHERE tenant_id IN (?)", testTenantIDs)
		db.Exec("DELETE FROM suppliers WHERE tenant_id IN (?)", testTenantIDs)
		db.Exec("DELETE FROM customers WHERE tenant_id IN (?)", testTenantIDs)
		db.Exec("DELETE FROM companies WHERE tenant_id IN (?)", testTenantIDs)

		// Get subscription IDs before nullifying (for deletion)
		var subscriptionIDs []string
		db.Raw("SELECT DISTINCT subscription_id FROM tenants WHERE id IN (?) AND subscription_id IS NOT NULL", testTenantIDs).Scan(&subscriptionIDs)

		// Nullify tenant subscription_id before deleting subscriptions (FK constraint)
		db.Exec("UPDATE tenants SET subscription_id = NULL WHERE id IN (?)", testTenantIDs)

		// Now safe to delete subscription payments and subscriptions
		if len(subscriptionIDs) > 0 {
			db.Exec("DELETE FROM subscription_payments WHERE subscription_id IN (?)", subscriptionIDs)
			db.Exec("DELETE FROM subscriptions WHERE id IN (?)", subscriptionIDs)
		}

		// Finally delete tenants
		db.Exec("DELETE FROM tenants WHERE id IN (?)", testTenantIDs)
	}

	fmt.Println("✓ Cleanup complete (only test data removed)")

	// Hash password once for all users
	defaultPassword := "Password123!"
	hashedPassword, err := hasher.HashPassword(defaultPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Reset GORM context completely after raw SQL operations
	db = db.Session(&gorm.Session{NewDB: true})

	// 1. Create System Admin User
	systemAdmin := &models.User{
		Email:         "superadmin@example.com",
		Username:      "superadmin",
		PasswordHash:  hashedPassword, // Fixed: Use PasswordHash not Password
		FullName:      "System Administrator", // Fixed: Use FullName not Name
		IsSystemAdmin: true,
		IsActive:      true,
	}
	if err := db.Create(systemAdmin).Error; err != nil {
		return fmt.Errorf("failed to create system admin: %w", err)
	}
	fmt.Println("✓ Created system admin user")

	// 2. Create Tenant 1 first (1 Tenant → N Companies architecture)
	tenant1ID := cuid.New()
	company1ID := cuid.New()

	// Create Tenant 1 WITHOUT company_id (correct architecture)
	trialEndsAt := time.Now().Add(14 * 24 * time.Hour)
	err = db.Exec(`
		INSERT INTO tenants (id, name, subdomain, status, trial_ends_at, created_at, updated_at)
		VALUES (?, ?, ?, 'TRIAL', ?, NOW(), NOW())
	`, tenant1ID, "PT Maju Jaya Distribusi", "pt-maju-jaya-distribusi", trialEndsAt).Error
	if err != nil {
		return fmt.Errorf("failed to create tenant1: %w", err)
	}

	// Now create Company 1 with tenant_id
	company1 := &models.Company{}
	err = db.Raw(`
		INSERT INTO companies (
			id, tenant_id, name, legal_name, entity_type, address, city, province, country,
			phone, email, ppn_rate, primary_color, secondary_color,
			invoice_prefix, invoice_number_format, so_prefix, so_number_format,
			po_prefix, po_number_format, currency, timezone, locale,
			business_hours_start, business_hours_end, working_days, is_active,
			created_at, updated_at
		) VALUES (
			?, ?, 'PT Maju Jaya Distribusi', 'PT Maju Jaya Distribusi Sembako', 'PT',
			'Jl. Raya Industri No. 123', 'Jakarta Utara', 'DKI Jakarta', 'Indonesia',
			'021-12345678', 'info@majujaya.co.id', 11.0, '#1E40AF', '#64748B',
			'INV', '{PREFIX}/{NUMBER}/{MONTH}/{YEAR}', 'SO', '{PREFIX}{NUMBER}',
			'PO', '{PREFIX}{NUMBER}', 'IDR', 'Asia/Jakarta', 'id-ID',
			'08:00', '17:00', '1,2,3,4,5', true,
			NOW(), NOW()
		) RETURNING *
	`, company1ID, tenant1ID).Scan(company1).Error
	if err != nil {
		return fmt.Errorf("failed to create company1: %w", err)
	}
	fmt.Println("✓ Created tenant 1 (TRIAL) and company 1: PT Maju Jaya Distribusi")

	// 3. Create Tenant 2, Subscription, and Company (1 Tenant → N Companies architecture)
	tenant2ID := cuid.New()
	company2ID := cuid.New()
	subscriptionID := cuid.New()

	// First create Tenant 2 WITHOUT subscription and WITHOUT company_id (correct architecture)
	err = db.Exec(`
		INSERT INTO tenants (id, name, subdomain, status, created_at, updated_at)
		VALUES (?, ?, ?, 'ACTIVE', NOW(), NOW())
	`, tenant2ID, "CV Berkah Mandiri", "cv-berkah-mandiri").Error
	if err != nil {
		return fmt.Errorf("failed to create tenant2: %w", err)
	}

	// Now create Subscription 2 WITHOUT tenant_id (1 Subscription → N Tenants architecture)
	now := time.Now()
	currentPeriodStart := now
	currentPeriodEnd := now.AddDate(0, 1, 0)
	nextBillingDate := currentPeriodEnd

	err = db.Exec(`
		INSERT INTO subscriptions (
			id, price, billing_cycle, status, current_period_start,
			current_period_end, next_billing_date, auto_renew, created_at, updated_at
		) VALUES (?, 300000, 'MONTHLY', 'ACTIVE', ?, ?, ?, true, NOW(), NOW())
	`, subscriptionID, currentPeriodStart, currentPeriodEnd, nextBillingDate).Error
	if err != nil {
		return fmt.Errorf("failed to create subscription2: %w", err)
	}

	// Update Tenant 2 WITH subscription_id
	err = db.Exec(`
		UPDATE tenants SET subscription_id = ? WHERE id = ?
	`, subscriptionID, tenant2ID).Error
	if err != nil {
		return fmt.Errorf("failed to update tenant2 with subscription: %w", err)
	}
	fmt.Println("✓ Created tenant 2 (ACTIVE) with subscription")

	// 4. Create Company 2 with tenant_id
	company2 := &models.Company{}
	err = db.Raw(`
		INSERT INTO companies (
			id, tenant_id, name, legal_name, entity_type, address, city, province, country,
			phone, email, ppn_rate, primary_color, secondary_color,
			invoice_prefix, invoice_number_format, so_prefix, so_number_format,
			po_prefix, po_number_format, currency, timezone, locale,
			business_hours_start, business_hours_end, working_days, is_active,
			created_at, updated_at
		) VALUES (
			?, ?, 'CV Berkah Mandiri', 'CV Berkah Mandiri Sejahtera', 'CV',
			'Jl. Perdagangan No. 45', 'Surabaya', 'Jawa Timur', 'Indonesia',
			'031-87654321', 'info@berkahmandiri.co.id', 11.0, '#059669', '#6B7280',
			'INV', '{PREFIX}/{NUMBER}/{MONTH}/{YEAR}', 'SO', '{PREFIX}{NUMBER}',
			'PO', '{PREFIX}{NUMBER}', 'IDR', 'Asia/Jakarta', 'id-ID',
			'08:00', '17:00', '1,2,3,4,5', true,
			NOW(), NOW()
		) RETURNING *
	`, company2ID, tenant2ID).Scan(company2).Error
	if err != nil {
		return fmt.Errorf("failed to create company2: %w", err)
	}
	fmt.Println("✓ Created tenant 2 (ACTIVE) and company 2: CV Berkah Mandiri")

	// 7. Create test users for Tenant 1 (PT Maju Jaya)
	tenant1Users := []struct {
		email    string
		username string
		name     string
		role     models.UserRole
	}{
		{"owner.maju@example.com", "owner_maju", "Budi Santoso (Owner)", models.UserRoleOwner},
		{"admin.maju@example.com", "admin_maju", "Siti Aminah (Admin)", models.UserRoleAdmin},
		{"finance.maju@example.com", "finance_maju", "Andi Wijaya (Finance)", models.UserRoleFinance},
		{"sales.maju@example.com", "sales_maju", "Dewi Lestari (Sales)", models.UserRoleSales},
		{"warehouse.maju@example.com", "warehouse_maju", "Joko Susilo (Warehouse)", models.UserRoleWarehouse},
		{"staff.maju@example.com", "staff_maju", "Nina Kusuma (Staff)", models.UserRoleStaff},
	}

	for _, userData := range tenant1Users {
		user := &models.User{
			Email:         userData.email,
			Username:      userData.username,
			PasswordHash:  hashedPassword, // Fixed: Use PasswordHash not Password
			FullName:      userData.name,  // Fixed: Use FullName not Name
			IsActive:      true,
		}
		if err := db.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", userData.email, err)
		}

		// Link user to tenant with role using raw SQL to avoid GORM default tag issues
		userTenantID := cuid.New()
		err = db.Exec(`
			INSERT INTO user_tenants (id, user_id, tenant_id, role, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, true, NOW(), NOW())
		`, userTenantID, user.ID, tenant1ID, userData.role).Error
		if err != nil {
			return fmt.Errorf("failed to create user-tenant link for %s: %w", userData.email, err)
		}
	}
	fmt.Println("✓ Created 6 test users for PT Maju Jaya (Tenant 1)")

	// 8. Create test users for Tenant 2 (CV Berkah Mandiri)
	tenant2Users := []struct {
		email    string
		username string
		name     string
		role     models.UserRole
	}{
		{"owner.berkah@example.com", "owner_berkah", "Hendra Gunawan (Owner)", models.UserRoleOwner},
		{"admin.berkah@example.com", "admin_berkah", "Rina Wijayanti (Admin)", models.UserRoleAdmin},
		{"finance.berkah@example.com", "finance_berkah", "Tono Hartono (Finance)", models.UserRoleFinance},
		{"sales.berkah@example.com", "sales_berkah", "Maya Sari (Sales)", models.UserRoleSales},
	}

	for _, userData := range tenant2Users {
		user := &models.User{
			Email:         userData.email,
			Username:      userData.username,
			PasswordHash:  hashedPassword, // Fixed: Use PasswordHash not Password
			FullName:      userData.name,  // Fixed: Use FullName not Name
			IsActive:      true,
		}
		if err := db.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", userData.email, err)
		}

		// Link user to tenant with role using raw SQL to avoid GORM default tag issues
		userTenantID := cuid.New()
		err = db.Exec(`
			INSERT INTO user_tenants (id, user_id, tenant_id, role, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, true, NOW(), NOW())
		`, userTenantID, user.ID, tenant2ID, userData.role).Error
		if err != nil {
			return fmt.Errorf("failed to create user-tenant link for %s: %w", userData.email, err)
		}
	}
	fmt.Println("✓ Created 4 test users for CV Berkah Mandiri (Tenant 2)")

	// 9. Create a multi-tenant user (can access both tenants)
	multiTenantUser := &models.User{
		Email:         "consultant@example.com",
		Username:      "consultant",
		PasswordHash:  hashedPassword, // Fixed: Use PasswordHash not Password
		FullName:      "Ahmad Konsultan (Multi-tenant)", // Fixed: Use FullName not Name
		IsActive:      true,
	}
	if err := db.Create(multiTenantUser).Error; err != nil {
		return fmt.Errorf("failed to create multi-tenant user: %w", err)
	}

	// Link to both tenants as STAFF using raw SQL to avoid GORM default tag issues
	for _, tenantID := range []string{tenant1ID, tenant2ID} {
		userTenantID := cuid.New()
		err = db.Exec(`
			INSERT INTO user_tenants (id, user_id, tenant_id, role, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, true, NOW(), NOW())
		`, userTenantID, multiTenantUser.ID, tenantID, models.UserRoleStaff).Error
		if err != nil {
			return fmt.Errorf("failed to create multi-tenant link: %w", err)
		}
	}
	fmt.Println("✓ Created multi-tenant consultant user (access to both companies)")

	// 10. Create UserCompanyRoles (Tier 2 - Per-Company Access)
	// This demonstrates the dual-tier permission system:
	// - Tier 1 (UserTenant): Tenant-level access (OWNER, SYSADMIN)
	// - Tier 2 (UserCompanyRole): Company-level access (ADMIN, FINANCE, SALES, WAREHOUSE, STAFF)
	fmt.Println("\n📋 Creating User-Company Roles (Tier 2)...")

	// Get user IDs for company role assignments
	var (
		adminMaju, financeMaju, salesMaju, warehouseMaju, staffMaju       models.User
		adminBerkah, financeBerkah, salesBerkah                           models.User
		consultantUser                                                    models.User
	)

	// Fetch users by email
	db.Where("email = ?", "admin.maju@example.com").First(&adminMaju)
	db.Where("email = ?", "finance.maju@example.com").First(&financeMaju)
	db.Where("email = ?", "sales.maju@example.com").First(&salesMaju)
	db.Where("email = ?", "warehouse.maju@example.com").First(&warehouseMaju)
	db.Where("email = ?", "staff.maju@example.com").First(&staffMaju)
	db.Where("email = ?", "admin.berkah@example.com").First(&adminBerkah)
	db.Where("email = ?", "finance.berkah@example.com").First(&financeBerkah)
	db.Where("email = ?", "sales.berkah@example.com").First(&salesBerkah)
	db.Where("email = ?", "consultant@example.com").First(&consultantUser)

	// Create UserCompanyRole records following phase1_seed.go pattern
	userCompanyRoles := []struct {
		userID    string
		companyID string
		tenantID  string
		role      models.UserRole
		comment   string
	}{
		// PT Maju Jaya (Company 1) - Tier 2 access
		{adminMaju.ID, company1ID, tenant1ID, models.UserRoleAdmin, "Admin at PT Maju Jaya"},
		{financeMaju.ID, company1ID, tenant1ID, models.UserRoleFinance, "Finance at PT Maju Jaya"},
		{salesMaju.ID, company1ID, tenant1ID, models.UserRoleSales, "Sales at PT Maju Jaya"},
		{warehouseMaju.ID, company1ID, tenant1ID, models.UserRoleWarehouse, "Warehouse at PT Maju Jaya"},
		{staffMaju.ID, company1ID, tenant1ID, models.UserRoleStaff, "Staff at PT Maju Jaya"},

		// CV Berkah Mandiri (Company 2) - Tier 2 access
		{adminBerkah.ID, company2ID, tenant2ID, models.UserRoleAdmin, "Admin at CV Berkah Mandiri"},
		{financeBerkah.ID, company2ID, tenant2ID, models.UserRoleFinance, "Finance at CV Berkah Mandiri"},
		{salesBerkah.ID, company2ID, tenant2ID, models.UserRoleSales, "Sales at CV Berkah Mandiri"},

		// Multi-tenant consultant - has STAFF role at both companies (Tier 2)
		{consultantUser.ID, company1ID, tenant1ID, models.UserRoleStaff, "Consultant at PT Maju Jaya"},
		{consultantUser.ID, company2ID, tenant2ID, models.UserRoleStaff, "Consultant at CV Berkah Mandiri"},
	}

	for _, ucr := range userCompanyRoles {
		userCompanyRoleID := cuid.New()
		err = db.Exec(`
			INSERT INTO user_company_roles (id, user_id, company_id, tenant_id, role, is_active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, true, NOW(), NOW())
		`, userCompanyRoleID, ucr.userID, ucr.companyID, ucr.tenantID, ucr.role).Error
		if err != nil {
			return fmt.Errorf("failed to create user-company-role (%s): %w", ucr.comment, err)
		}
		fmt.Printf("  ✓ User-Company-Role: %s\n", ucr.comment)
	}
	fmt.Printf("✓ Created %d user-company-role assignments\n", len(userCompanyRoles))

	return nil
}

func printTestUsers() {
	fmt.Println("\n🏢 PT MAJU JAYA DISTRIBUSI (Tenant 1 - TRIAL)")
	fmt.Println("  ┌─────────────────────────────────────────────────────────────")
	fmt.Println("  │ Email                       │ Role      │ Name")
	fmt.Println("  ├─────────────────────────────────────────────────────────────")
	fmt.Println("  │ owner.maju@example.com      │ OWNER     │ Budi Santoso")
	fmt.Println("  │ admin.maju@example.com      │ ADMIN     │ Siti Aminah")
	fmt.Println("  │ finance.maju@example.com    │ FINANCE   │ Andi Wijaya")
	fmt.Println("  │ sales.maju@example.com      │ SALES     │ Dewi Lestari")
	fmt.Println("  │ warehouse.maju@example.com  │ WAREHOUSE │ Joko Susilo")
	fmt.Println("  │ staff.maju@example.com      │ STAFF     │ Nina Kusuma")
	fmt.Println("  └─────────────────────────────────────────────────────────────")

	fmt.Println("\n🏢 CV BERKAH MANDIRI (Tenant 2 - ACTIVE with Subscription)")
	fmt.Println("  ┌─────────────────────────────────────────────────────────────")
	fmt.Println("  │ Email                       │ Role      │ Name")
	fmt.Println("  ├─────────────────────────────────────────────────────────────")
	fmt.Println("  │ owner.berkah@example.com    │ OWNER     │ Hendra Gunawan")
	fmt.Println("  │ admin.berkah@example.com    │ ADMIN     │ Rina Wijayanti")
	fmt.Println("  │ finance.berkah@example.com  │ FINANCE   │ Tono Hartono")
	fmt.Println("  │ sales.berkah@example.com    │ SALES     │ Maya Sari")
	fmt.Println("  └─────────────────────────────────────────────────────────────")

	fmt.Println("\n👤 SPECIAL USERS")
	fmt.Println("  ┌─────────────────────────────────────────────────────────────")
	fmt.Println("  │ Email                       │ Role      │ Description")
	fmt.Println("  ├─────────────────────────────────────────────────────────────")
	fmt.Println("  │ superadmin@example.com      │ SYSADMIN  │ System Administrator")
	fmt.Println("  │ consultant@example.com      │ STAFF     │ Multi-tenant User (Both)")
	fmt.Println("  └─────────────────────────────────────────────────────────────")
}
