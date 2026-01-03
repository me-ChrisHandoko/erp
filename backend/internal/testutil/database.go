package testutil

import (
	"backend/models"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		// User and tenant models
		&models.User{},
		&models.Tenant{},
		&models.UserTenant{},
		&models.Company{},

		// Master data models
		&models.Product{},
		&models.ProductUnit{},
		&models.ProductBatch{},
		&models.ProductSupplier{},
		&models.Customer{},
		&models.Supplier{},
		&models.Warehouse{},
		&models.WarehouseStock{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

// CreateTestCompany creates a test company for testing
func CreateTestCompany(t *testing.T, db *gorm.DB, tenantID, name string) *models.Company {
	// Generate unique NPWP using UUID suffix to avoid UNIQUE constraint violations
	npwp := fmt.Sprintf("01.234.567.8-%d-%d", time.Now().UnixNano(), len(name))
	company := &models.Company{
		TenantID:            tenantID,
		Name:                name,
		LegalName:           "PT " + name,
		EntityType:          "CV",
		Address:             "Jl. Test No. 123",
		City:                "Jakarta",
		Province:            "DKI Jakarta",
		Country:             "Indonesia",
		Phone:               "08123456789",
		Email:               "test@company.com",
		NPWP:                &npwp,
		IsPKP:               true,
		PPNRate:             decimal.NewFromFloat(11.0),
		IsActive:            true,
		InvoicePrefix:       "INV",
		InvoiceNumberFormat: "{PREFIX}/{NUMBER}/{MONTH}/{YEAR}",
		SOPrefix:            "SO",
		SONumberFormat:      "{PREFIX}{NUMBER}",
		POPrefix:            "PO",
		PONumberFormat:      "{PREFIX}{NUMBER}",
		Currency:            "IDR",
		Timezone:            "Asia/Jakarta",
		Locale:              "id-ID",
	}

	if err := db.Create(company).Error; err != nil {
		t.Fatalf("failed to create test company: %v", err)
	}

	return company
}

// CreateTestUser creates a test user for testing
func CreateTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:    email,
		FullName: "Test User",
		IsActive: true,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}

// CreateTestWarehouse creates a test warehouse for testing
func CreateTestWarehouse(t *testing.T, db *gorm.DB, companyID, code string) *models.Warehouse {
	warehouse := &models.Warehouse{
		CompanyID: companyID,
		Code:      code,
		Name:      "Test Warehouse " + code,
		Type:      models.WarehouseTypeMain,
		IsActive:  true,
	}

	if err := db.Create(warehouse).Error; err != nil {
		t.Fatalf("failed to create test warehouse: %v", err)
	}

	return warehouse
}
