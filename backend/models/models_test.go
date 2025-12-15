// Package models - Unit tests for Phase 1 models
package models_test

import (
	"backend/db"
	"backend/models"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign key constraints for SQLite
	database.Exec("PRAGMA foreign_keys = ON")

	// Run Phase 1 migration
	if err := db.AutoMigratePhase1(database); err != nil {
		t.Fatalf("Auto-migration failed: %v", err)
	}

	return database
}

// TestSchemaGeneration verifies that all Phase 1 tables are created correctly
func TestSchemaGeneration(t *testing.T) {
	database := setupTestDB(t)

	tables := []string{
		"users",
		"companies",
		"company_banks",
		"tenants",
		"subscriptions",
		"subscription_payments",
		"user_tenants",
	}

	for _, tableName := range tables {
		if !database.Migrator().HasTable(tableName) {
			t.Errorf("Table %s does not exist", tableName)
		}
	}
}

// TestCUIDGeneration verifies that CUID is generated for all models
func TestCUIDGeneration(t *testing.T) {
	database := setupTestDB(t)

	// Test User CUID generation
	user := &models.User{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashedpassword",
		Name:     "Test User",
	}
	database.Create(user)

	if user.ID == "" {
		t.Error("CUID not generated for User.ID")
	}

	if len(user.ID) < 20 {
		t.Errorf("Generated ID too short: %s", user.ID)
	}

	// Test Company CUID generation
	company := &models.Company{
		Name:      "Test Company",
		LegalName: "Test Company PT",
		Address:   "Test Address",
		City:      "Jakarta",
		Province:  "DKI Jakarta",
		Phone:     "021-123456",
		Email:     "company@test.com",
	}
	database.Create(company)

	if company.ID == "" {
		t.Error("CUID not generated for Company.ID")
	}
}

// TestUserCreation verifies User model creation with all fields
func TestUserCreation(t *testing.T) {
	database := setupTestDB(t)

	user := &models.User{
		Email:         "john@example.com",
		Username:      "johndoe",
		Password:      "hashedpassword123",
		Name:          "John Doe",
		IsSystemAdmin: true,
		IsActive:      true,
	}

	result := database.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create user: %v", result.Error)
	}

	// Retrieve and verify
	var retrieved models.User
	database.First(&retrieved, "email = ?", "john@example.com")

	if retrieved.Username != "johndoe" {
		t.Errorf("Expected username 'johndoe', got '%s'", retrieved.Username)
	}

	if !retrieved.IsSystemAdmin {
		t.Error("IsSystemAdmin should be true")
	}

	if retrieved.CreatedAt.IsZero() {
		t.Error("CreatedAt should be auto-generated")
	}
}

// TestCompanyCreation verifies Company model with Indonesian tax fields
func TestCompanyCreation(t *testing.T) {
	database := setupTestDB(t)

	npwp := "01.234.567.8-901.234"
	company := &models.Company{
		Name:       "CV Test Indonesia",
		LegalName:  "CV Test Indonesia Sejahtera",
		EntityType: "CV",
		Address:    "Jl. Sudirman No. 123",
		City:       "Jakarta Selatan",
		Province:   "DKI Jakarta",
		Country:    "Indonesia",
		Phone:      "021-7654321",
		Email:      "info@cvtest.co.id",
		NPWP:       &npwp,
		IsPKP:      true,
		PPNRate:    decimal.NewFromFloat(11.0),
	}

	result := database.Create(company)
	if result.Error != nil {
		t.Fatalf("Failed to create company: %v", result.Error)
	}

	// Retrieve and verify
	var retrieved models.Company
	database.First(&retrieved, "email = ?", "info@cvtest.co.id")

	if retrieved.Name != "CV Test Indonesia" {
		t.Errorf("Expected name 'CV Test Indonesia', got '%s'", retrieved.Name)
	}

	if !retrieved.IsPKP {
		t.Error("IsPKP should be true")
	}

	if !retrieved.PPNRate.Equal(decimal.NewFromFloat(11.0)) {
		t.Errorf("Expected PPNRate 11.0, got %v", retrieved.PPNRate)
	}
}

// TestTenantWithSubscription verifies Tenant and Subscription relationship
func TestTenantWithSubscription(t *testing.T) {
	database := setupTestDB(t)

	// Create Company first
	company := &models.Company{
		Name:      "Test PT",
		LegalName: "PT Test Indonesia",
		Address:   "Jakarta",
		City:      "Jakarta",
		Province:  "DKI Jakarta",
		Phone:     "021-111111",
		Email:     "pt@test.com",
	}
	database.Create(company)

	// Create Subscription
	subscription := &models.Subscription{
		Price:              decimal.NewFromFloat(300000),
		BillingCycle:       "MONTHLY",
		Status:             models.SubscriptionStatusActive,
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		NextBillingDate:    time.Now().AddDate(0, 1, 0),
		AutoRenew:          true,
	}
	database.Create(subscription)

	// Create Tenant
	trialEnd := time.Now().AddDate(0, 0, 14) // 14 days trial
	tenant := &models.Tenant{
		CompanyID:      company.ID,
		SubscriptionID: &subscription.ID,
		Status:         models.TenantStatusTrial,
		TrialEndsAt:    &trialEnd,
	}
	database.Create(tenant)

	// Retrieve with preloaded relationships
	var retrieved models.Tenant
	database.Preload("Company").Preload("Subscription").First(&retrieved, "id = ?", tenant.ID)

	if retrieved.Company.Name != "Test PT" {
		t.Errorf("Expected company name 'Test PT', got '%s'", retrieved.Company.Name)
	}

	if retrieved.Subscription.Price.Cmp(decimal.NewFromFloat(300000)) != 0 {
		t.Errorf("Expected subscription price 300000, got %v", retrieved.Subscription.Price)
	}

	if retrieved.Status != models.TenantStatusTrial {
		t.Errorf("Expected status TRIAL, got %s", retrieved.Status)
	}
}

// TestUserTenantJunction verifies UserTenant junction table with roles
func TestUserTenantJunction(t *testing.T) {
	database := setupTestDB(t)

	// Create User
	user := &models.User{
		Email:    "manager@test.com",
		Username: "manager",
		Password: "hashed",
		Name:     "Manager User",
	}
	database.Create(user)

	// Create Company and Tenant
	company := &models.Company{
		Name:      "Test Company",
		LegalName: "Test Company PT",
		Address:   "Jakarta",
		City:      "Jakarta",
		Province:  "DKI Jakarta",
		Phone:     "021-222222",
		Email:     "company2@test.com",
	}
	database.Create(company)

	tenant := &models.Tenant{
		CompanyID: company.ID,
		Status:    models.TenantStatusActive,
	}
	database.Create(tenant)

	// Create UserTenant with OWNER role
	userTenant := &models.UserTenant{
		UserID:   user.ID,
		TenantID: tenant.ID,
		Role:     models.UserRoleOwner,
		IsActive: true,
	}
	database.Create(userTenant)

	// Retrieve and verify
	var retrieved models.UserTenant
	database.Preload("User").Preload("Tenant").First(&retrieved, "user_id = ? AND tenant_id = ?", user.ID, tenant.ID)

	if retrieved.Role != models.UserRoleOwner {
		t.Errorf("Expected role OWNER, got %s", retrieved.Role)
	}

	if retrieved.User.Email != "manager@test.com" {
		t.Errorf("Expected user email 'manager@test.com', got '%s'", retrieved.User.Email)
	}
}

// TestUniqueConstraints verifies unique index enforcement
func TestUniqueConstraints(t *testing.T) {
	database := setupTestDB(t)

	// Create first user
	user1 := &models.User{
		Email:    "unique@test.com",
		Username: "uniqueuser",
		Password: "hashed",
		Name:     "First User",
	}
	database.Create(user1)

	// Try to create second user with same email (should fail)
	user2 := &models.User{
		Email:    "unique@test.com",
		Username: "differentuser",
		Password: "hashed",
		Name:     "Second User",
	}
	result := database.Create(user2)

	if result.Error == nil {
		t.Error("Expected unique constraint violation for duplicate email")
	}
}

// TestCascadeDelete verifies CASCADE deletion rules
func TestCascadeDelete(t *testing.T) {
	database := setupTestDB(t)

	// Create Company
	company := &models.Company{
		Name:      "Delete Test",
		LegalName: "Delete Test PT",
		Address:   "Jakarta",
		City:      "Jakarta",
		Province:  "DKI Jakarta",
		Phone:     "021-333333",
		Email:     "delete@test.com",
	}
	database.Create(company)

	// Create CompanyBank
	bank := &models.CompanyBank{
		CompanyID:     company.ID,
		BankName:      "BCA",
		AccountNumber: "1234567890",
		AccountName:   "Delete Test",
	}
	database.Create(bank)

	// Delete Company with Select to enable CASCADE for has-many relations
	// Note: GORM requires explicit Select("Banks") for cascade deletion of has-many relations
	database.Select("Banks").Delete(company)

	// Verify CompanyBank is also deleted (CASCADE)
	var count int64
	database.Model(&models.CompanyBank{}).Where("company_id = ?", company.ID).Count(&count)

	if count != 0 {
		t.Error("CASCADE deletion failed - CompanyBank not deleted with Company")
	}
}

// TestDecimalPrecision verifies Decimal field precision
func TestDecimalPrecision(t *testing.T) {
	database := setupTestDB(t)

	subscription := &models.Subscription{
		Price:              decimal.NewFromFloat(12345.67),
		BillingCycle:       "MONTHLY",
		Status:             models.SubscriptionStatusActive,
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0),
		NextBillingDate:    time.Now().AddDate(0, 1, 0),
	}
	database.Create(subscription)

	var retrieved models.Subscription
	database.First(&retrieved, "id = ?", subscription.ID)

	if !retrieved.Price.Equal(decimal.NewFromFloat(12345.67)) {
		t.Errorf("Decimal precision lost: expected 12345.67, got %v", retrieved.Price)
	}
}

// TestEnumValues verifies enum type usage
func TestEnumValues(t *testing.T) {
	database := setupTestDB(t)

	company := &models.Company{
		Name:      "Enum Test",
		LegalName: "Enum Test PT",
		Address:   "Jakarta",
		City:      "Jakarta",
		Province:  "DKI Jakarta",
		Phone:     "021-444444",
		Email:     "enum@test.com",
	}
	database.Create(company)

	tenant := &models.Tenant{
		CompanyID: company.ID,
		Status:    models.TenantStatusTrial,
	}
	database.Create(tenant)

	var retrieved models.Tenant
	database.First(&retrieved, "id = ?", tenant.ID)

	if retrieved.Status != models.TenantStatusTrial {
		t.Errorf("Expected status TRIAL, got %s", retrieved.Status)
	}

	// Update status
	database.Model(&retrieved).Update("status", models.TenantStatusActive)

	var updated models.Tenant
	database.First(&updated, "id = ?", tenant.ID)

	if updated.Status != models.TenantStatusActive {
		t.Errorf("Expected updated status ACTIVE, got %s", updated.Status)
	}
}
