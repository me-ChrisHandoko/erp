// Package models - Phase 4 tests (Supporting Modules)
package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPhase4TestDB creates in-memory SQLite database for Phase 4 tests
func setupPhase4TestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Enable foreign key constraints
	err = db.Exec("PRAGMA foreign_keys = ON").Error
	assert.NoError(t, err)

	return db
}

// createPhase4TestData creates all prerequisite data for Phase 4 tests
func createPhase4TestData(t *testing.T, db *gorm.DB) (*Tenant, *Warehouse, *Product, *ProductBatch, *WarehouseStock) {
	// Create company first
	company := &Company{
		Name:      "Test Company Phase 4",
		LegalName: "Test Company Phase 4 Ltd",
		Address:   "Test Address",
		City:      "Test City",
		Province:  "Test Province",
	}
	err := db.Create(company).Error
	assert.NoError(t, err)

	// Create tenant
	tenant := &Tenant{
		Name:        "Test Tenant Phase 4",
		Subdomain:   "test-tenant-phase4",
		Status:      TenantStatusActive,
		TrialEndsAt: timePtr(time.Now().Add(14 * 24 * time.Hour)),
	}
	err = db.Create(tenant).Error
	assert.NoError(t, err)

	// Link company to tenant
	company.TenantID = tenant.ID
	err = db.Save(company).Error
	assert.NoError(t, err)

	// Create warehouse
	warehouse := &Warehouse{
		TenantID:    tenant.ID,
		Code:        "WH-MAIN",
		Name:        "Main Warehouse",
		Type:        WarehouseTypeMain,
		IsActive:    true,
	}
	err = db.Create(warehouse).Error
	assert.NoError(t, err)

	// Create product
	product := &Product{
		TenantID:       tenant.ID,
		Code:           "PROD-001",
		Name:           "Test Product",
		BaseUnit:       "PCS",
		IsBatchTracked: true,
		IsPerishable:   true,
		IsActive:       true,
	}
	err = db.Create(product).Error
	assert.NoError(t, err)

	// Create warehouse stock first (required by ProductBatch)
	stock := &WarehouseStock{
		WarehouseID: warehouse.ID,
		ProductID:   product.ID,
		Quantity:    decimal.NewFromInt(100),
	}
	err = db.Create(stock).Error
	assert.NoError(t, err)

	// Create batch (requires WarehouseStockID)
	batch := &ProductBatch{
		ProductID:        product.ID,
		WarehouseStockID: stock.ID,
		BatchNumber:      "BATCH-001",
		ManufactureDate:  timePtr(time.Now().Add(-30 * 24 * time.Hour)),
		ExpiryDate:       timePtr(time.Now().Add(180 * 24 * time.Hour)),
		Status:           BatchStatusAvailable,
	}
	err = db.Create(batch).Error
	assert.NoError(t, err)

	return tenant, warehouse, product, batch, stock
}

// TestPhase4SchemaGeneration tests that all Phase 4 tables are created correctly
func TestPhase4SchemaGeneration(t *testing.T) {
	db := setupPhase4TestDB(t)

	// Run all phase migrations
	err := db.AutoMigrate(
		// Phase 1
		&User{}, &Company{}, &CompanyBank{}, &Subscription{},
		&Tenant{}, &SubscriptionPayment{}, &UserTenant{},
		// Phase 2
		&Customer{}, &Supplier{}, &Warehouse{}, &Product{},
		&ProductUnit{}, &PriceList{}, &ProductSupplier{},
		&WarehouseStock{}, &ProductBatch{},
		// Phase 4
		&InventoryMovement{}, &StockOpname{}, &StockOpnameItem{},
		&StockTransfer{}, &StockTransferItem{}, &CashTransaction{},
		&Setting{}, &AuditLog{},
	)
	assert.NoError(t, err)

	// Verify Phase 4 tables exist
	tables := []string{
		"inventory_movements",
		"stock_opnames", "stock_opname_items",
		"stock_transfers", "stock_transfer_items",
		"cash_transactions",
		"settings", "audit_logs",
	}

	for _, table := range tables {
		assert.True(t, db.Migrator().HasTable(table), "Table %s should exist", table)
	}
}

// TestInventoryMovementCreation tests inventory movement audit trail
func TestInventoryMovementCreation(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{}, &InventoryMovement{})
	assert.NoError(t, err)

	tenant, warehouse, product, batch, stock := createPhase4TestData(t, db)

	// Create inventory movement (stock IN from goods receipt)
	movement := &InventoryMovement{
		TenantID:        tenant.ID,
		MovementDate:    time.Now(),
		WarehouseID:     warehouse.ID,
		ProductID:       product.ID,
		BatchID:         &batch.ID,
		MovementType:    MovementTypeIn,
		Quantity:        decimal.NewFromInt(50), // +50 units
		StockBefore:     stock.Quantity,
		StockAfter:      stock.Quantity.Add(decimal.NewFromInt(50)),
		ReferenceType:   strPtr("GOODS_RECEIPT"),
		ReferenceID:     strPtr("GRN-001-ID"),
		ReferenceNumber: strPtr("GRN-001"),
		Notes:           strPtr("Goods receipt from supplier"),
	}

	err = db.Create(movement).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, movement.ID)
	assert.Equal(t, decimal.NewFromInt(50), movement.Quantity)
	assert.Equal(t, MovementTypeIn, movement.MovementType)
	assert.Equal(t, "GRN-001", *movement.ReferenceNumber)

	// Verify relationships
	var loaded InventoryMovement
	err = db.Preload("Tenant").Preload("Warehouse").Preload("Product").Preload("Batch").First(&loaded, "id = ?", movement.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, loaded.Tenant.ID)
	assert.Equal(t, warehouse.Name, loaded.Warehouse.Name)
	assert.Equal(t, product.Name, loaded.Product.Name)
	assert.Equal(t, batch.BatchNumber, loaded.Batch.BatchNumber)
}

// TestStockOpnameCreation tests physical inventory count workflow
func TestStockOpnameCreation(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{}, &StockOpname{}, &StockOpnameItem{})
	assert.NoError(t, err)

	tenant, warehouse, product, batch, stock := createPhase4TestData(t, db)

	// Create stock opname
	opname := &StockOpname{
		TenantID:     tenant.ID,
		OpnameNumber: "OP-2025-001",
		OpnameDate:   time.Now(),
		WarehouseID:  warehouse.ID,
		Status:       StockOpnameStatusDraft,
		CountedBy:    strPtr("User-001"),
		Notes:        strPtr("Monthly physical count"),
	}

	err = db.Create(opname).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, opname.ID)

	// Create opname item with variance
	systemQty := stock.Quantity
	physicalQty := decimal.NewFromInt(98) // Actual count: 98 (variance: -2)
	differenceQty := physicalQty.Sub(systemQty)

	opnameItem := &StockOpnameItem{
		StockOpnameID: opname.ID,
		ProductID:     product.ID,
		BatchID:       &batch.ID,
		SystemQty:     systemQty,
		PhysicalQty:   physicalQty,
		DifferenceQty: differenceQty,
		Notes:         strPtr("2 units damaged"),
	}

	err = db.Create(opnameItem).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, opnameItem.ID)
	assert.Equal(t, decimal.NewFromInt(100), opnameItem.SystemQty)
	assert.Equal(t, decimal.NewFromInt(98), opnameItem.PhysicalQty)
	assert.Equal(t, decimal.NewFromInt(-2), opnameItem.DifferenceQty)

	// Verify relationships
	var loaded StockOpname
	err = db.Preload("Items").Preload("Warehouse").First(&loaded, "id = ?", opname.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loaded.Items))
	assert.Equal(t, warehouse.Name, loaded.Warehouse.Name)
}

// TestStockTransferCreation tests inter-warehouse transfer workflow
func TestStockTransferCreation(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{}, &StockTransfer{}, &StockTransferItem{})
	assert.NoError(t, err)

	tenant, sourceWarehouse, product, batch, _ := createPhase4TestData(t, db)

	// Create destination warehouse
	destWarehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-BRANCH",
		Name:     "Branch Warehouse",
		Type:     WarehouseTypeBranch,
		IsActive: true,
	}
	err = db.Create(destWarehouse).Error
	assert.NoError(t, err)

	// Create stock transfer
	transfer := &StockTransfer{
		TenantID:          tenant.ID,
		TransferNumber:    "TRF-2025-001",
		TransferDate:      time.Now(),
		SourceWarehouseID: sourceWarehouse.ID,
		DestWarehouseID:   destWarehouse.ID,
		Status:            StockTransferStatusDraft,
		Notes:             strPtr("Monthly restock to branch"),
	}

	err = db.Create(transfer).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, transfer.ID)

	// Create transfer item
	transferItem := &StockTransferItem{
		StockTransferID: transfer.ID,
		ProductID:       product.ID,
		BatchID:         &batch.ID,
		Quantity:        decimal.NewFromInt(20),
		Notes:           strPtr("Transfer 20 units"),
	}

	err = db.Create(transferItem).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, transferItem.ID)
	assert.Equal(t, decimal.NewFromInt(20), transferItem.Quantity)

	// Verify relationships
	var loaded StockTransfer
	err = db.Preload("Items").Preload("SourceWarehouse").Preload("DestWarehouse").First(&loaded, "id = ?", transfer.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loaded.Items))
	assert.Equal(t, sourceWarehouse.Name, loaded.SourceWarehouse.Name)
	assert.Equal(t, destWarehouse.Name, loaded.DestWarehouse.Name)
}

// TestStockTransferWorkflow tests shipped/received status workflow
func TestStockTransferWorkflow(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{}, &StockTransfer{}, &StockTransferItem{})
	assert.NoError(t, err)

	tenant, sourceWarehouse, _, _, _ := createPhase4TestData(t, db)

	destWarehouse := &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-DEST",
		Name:     "Destination Warehouse",
		Type:     WarehouseTypeBranch,
		IsActive: true,
	}
	err = db.Create(destWarehouse).Error
	assert.NoError(t, err)

	transfer := &StockTransfer{
		TenantID:          tenant.ID,
		TransferNumber:    "TRF-WORKFLOW-001",
		TransferDate:      time.Now(),
		SourceWarehouseID: sourceWarehouse.ID,
		DestWarehouseID:   destWarehouse.ID,
		Status:            StockTransferStatusDraft,
	}
	err = db.Create(transfer).Error
	assert.NoError(t, err)

	// Ship the transfer
	shippedAt := time.Now()
	err = db.Model(transfer).Updates(map[string]interface{}{
		"status":     StockTransferStatusShipped,
		"shipped_by": "User-001",
		"shipped_at": shippedAt,
	}).Error
	assert.NoError(t, err)

	// Receive the transfer
	receivedAt := time.Now()
	err = db.Model(transfer).Updates(map[string]interface{}{
		"status":      StockTransferStatusReceived,
		"received_by": "User-002",
		"received_at": receivedAt,
	}).Error
	assert.NoError(t, err)

	// Verify workflow
	var loaded StockTransfer
	err = db.First(&loaded, "id = ?", transfer.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, StockTransferStatusReceived, loaded.Status)
	assert.Equal(t, "User-001", *loaded.ShippedBy)
	assert.Equal(t, "User-002", *loaded.ReceivedBy)
	assert.NotNil(t, loaded.ShippedAt)
	assert.NotNil(t, loaded.ReceivedAt)
}

// TestCashTransactionCreation tests cash book (Buku Kas) tracking
func TestCashTransactionCreation(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Company{}, &Tenant{}, &CashTransaction{})
	assert.NoError(t, err)

	company := &Company{
		Name:      "Test Company Cash",
		LegalName: "Test Company Cash Ltd",
		Address:   "Test Address",
		City:      "Test City",
		Province:  "Test Province",
	}
	err = db.Create(company).Error
	assert.NoError(t, err)

	tenant := &Tenant{
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		Status:    TenantStatusActive,
	}
	err = db.Create(tenant).Error
	assert.NoError(t, err)

	// Link company to tenant
	company.TenantID = tenant.ID
	err = db.Save(company).Error
	assert.NoError(t, err)

	// Create cash IN transaction (sales payment)
	cashIn := &CashTransaction{
		TenantID:          tenant.ID,
		TransactionNumber: "CASH-IN-001",
		TransactionDate:   time.Now(),
		Type:              CashTransactionTypeCashIn,
		Category:          CashCategorySales,
		Amount:            decimal.NewFromInt(1000000),
		BalanceBefore:     decimal.NewFromInt(5000000),
		BalanceAfter:      decimal.NewFromInt(6000000),
		Description:       "Payment from customer ABC",
		ReferenceType:     strPtr("PAYMENT"),
		ReferenceNumber:   strPtr("PAY-001"),
		CreatedBy:         strPtr("User-001"),
	}

	err = db.Create(cashIn).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, cashIn.ID)
	assert.Equal(t, CashTransactionTypeCashIn, cashIn.Type)
	assert.Equal(t, CashCategorySales, cashIn.Category)
	assert.Equal(t, decimal.NewFromInt(1000000), cashIn.Amount)

	// Create cash OUT transaction (expense)
	cashOut := &CashTransaction{
		TenantID:          tenant.ID,
		TransactionNumber: "CASH-OUT-001",
		TransactionDate:   time.Now(),
		Type:              CashTransactionTypeCashOut,
		Category:          CashCategoryExpense,
		Amount:            decimal.NewFromInt(500000),
		BalanceBefore:     decimal.NewFromInt(6000000),
		BalanceAfter:      decimal.NewFromInt(5500000),
		Description:       "Office rent payment",
		CreatedBy:         strPtr("User-001"),
	}

	err = db.Create(cashOut).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, cashOut.ID)
	assert.Equal(t, CashTransactionTypeCashOut, cashOut.Type)
	assert.Equal(t, CashCategoryExpense, cashOut.Category)

	// Verify running balance
	var transactions []CashTransaction
	err = db.Where("tenant_id = ?", tenant.ID).Order("transaction_date ASC").Find(&transactions).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, len(transactions))
	assert.Equal(t, decimal.NewFromInt(5000000), transactions[0].BalanceBefore)
	assert.Equal(t, decimal.NewFromInt(6000000), transactions[0].BalanceAfter)
	assert.Equal(t, decimal.NewFromInt(6000000), transactions[1].BalanceBefore)
	assert.Equal(t, decimal.NewFromInt(5500000), transactions[1].BalanceAfter)
}

// TestSettingCreation tests system and tenant-specific settings
func TestSettingCreation(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Company{}, &Tenant{}, &Setting{})
	assert.NoError(t, err)

	company := &Company{
		Name:      "Test Company Settings",
		LegalName: "Test Company Settings Ltd",
		Address:   "Test Address",
		City:      "Test City",
		Province:  "Test Province",
	}
	err = db.Create(company).Error
	assert.NoError(t, err)

	tenant := &Tenant{
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		Status:    TenantStatusActive,
	}
	err = db.Create(tenant).Error
	assert.NoError(t, err)

	// Link company to tenant
	company.TenantID = tenant.ID
	err = db.Save(company).Error
	assert.NoError(t, err)

	// Create system-wide setting (NULL tenantID)
	systemSetting := &Setting{
		Key:      "system.maintenance_mode",
		Value:    strPtr("false"),
		DataType: "BOOLEAN",
		IsPublic: true,
	}

	err = db.Create(systemSetting).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, systemSetting.ID)
	assert.Nil(t, systemSetting.TenantID)

	// Create tenant-specific setting
	tenantSetting := &Setting{
		TenantID: &tenant.ID,
		Key:      "invoice.number_format",
		Value:    strPtr("{PREFIX}/{NUMBER}/{MONTH}/{YEAR}"),
		DataType: "STRING",
		IsPublic: false,
	}

	err = db.Create(tenantSetting).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, tenantSetting.ID)
	assert.Equal(t, tenant.ID, *tenantSetting.TenantID)

	// Verify settings retrieval
	var settings []Setting
	err = db.Where("tenant_id = ? OR tenant_id IS NULL", tenant.ID).Find(&settings).Error
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(settings), 2)
}

// TestAuditLogCreation tests comprehensive audit trail
func TestAuditLogCreation(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Company{}, &User{}, &Tenant{}, &AuditLog{})
	assert.NoError(t, err)

	company := &Company{
		Name:      "Test Company Audit",
		LegalName: "Test Company Audit Ltd",
		Address:   "Test Address",
		City:      "Test City",
		Province:  "Test Province",
	}
	err = db.Create(company).Error
	assert.NoError(t, err)

	tenant := &Tenant{
		Name:      "Test Tenant",
		Subdomain: "test-tenant",
		Status:    TenantStatusActive,
	}
	err = db.Create(tenant).Error
	assert.NoError(t, err)

	// Link company to tenant
	company.TenantID = tenant.ID
	err = db.Save(company).Error
	assert.NoError(t, err)

	user := &User{
		Email:        "audit@test.com",
		PasswordHash: "hashedpassword",
		IsActive:     true,
	}
	err = db.Create(user).Error
	assert.NoError(t, err)

	// Create audit log for entity creation
	auditLog := &AuditLog{
		TenantID:   &tenant.ID,
		UserID:     &user.ID,
		Action:     "CREATE",
		EntityType: strPtr("Invoice"),
		EntityID:   strPtr("INV-001-ID"),
		OldValues:  nil,
		NewValues:  strPtr(`{"invoice_number":"INV-001","total_amount":1000000}`),
		IPAddress:  strPtr("192.168.1.1"),
		UserAgent:  strPtr("Mozilla/5.0"),
		Notes:      strPtr("Invoice created via web UI"),
	}

	err = db.Create(auditLog).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, auditLog.ID)
	assert.Equal(t, "CREATE", auditLog.Action)
	assert.Equal(t, "Invoice", *auditLog.EntityType)

	// Create audit log for entity update
	updateLog := &AuditLog{
		TenantID:   &tenant.ID,
		UserID:     &user.ID,
		Action:     "UPDATE",
		EntityType: strPtr("Invoice"),
		EntityID:   strPtr("INV-001-ID"),
		OldValues:  strPtr(`{"payment_status":"UNPAID"}`),
		NewValues:  strPtr(`{"payment_status":"PAID"}`),
		IPAddress:  strPtr("192.168.1.1"),
	}

	err = db.Create(updateLog).Error
	assert.NoError(t, err)

	// Verify audit trail
	var logs []AuditLog
	err = db.Where("entity_id = ?", "INV-001-ID").Order("created_at ASC").Find(&logs).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, len(logs))
	assert.Equal(t, "CREATE", logs[0].Action)
	assert.Equal(t, "UPDATE", logs[1].Action)

	// Verify relationships
	var loaded AuditLog
	err = db.Preload("Tenant").Preload("User").First(&loaded, "id = ?", auditLog.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, loaded.Tenant.ID)
	assert.Equal(t, user.Email, loaded.User.Email)
}

// TestUniqueConstraintsPhase4 tests unique constraints for Phase 4 models
func TestUniqueConstraintsPhase4(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{},
		&StockOpname{}, &StockTransfer{}, &CashTransaction{})
	assert.NoError(t, err)

	tenant, warehouse, _, _, _ := createPhase4TestData(t, db)

	// Test stock opname number uniqueness
	opname1 := &StockOpname{
		TenantID:     tenant.ID,
		OpnameNumber: "OP-UNIQUE-001",
		OpnameDate:   time.Now(),
		WarehouseID:  warehouse.ID,
		Status:       StockOpnameStatusDraft,
	}
	err = db.Create(opname1).Error
	assert.NoError(t, err)

	opname2 := &StockOpname{
		TenantID:     tenant.ID,
		OpnameNumber: "OP-UNIQUE-001", // Duplicate
		OpnameDate:   time.Now(),
		WarehouseID:  warehouse.ID,
		Status:       StockOpnameStatusDraft,
	}
	err = db.Create(opname2).Error
	assert.Error(t, err, "Duplicate opname number should fail")

	// Test stock transfer number uniqueness
	transfer1 := &StockTransfer{
		TenantID:          tenant.ID,
		TransferNumber:    "TRF-UNIQUE-001",
		TransferDate:      time.Now(),
		SourceWarehouseID: warehouse.ID,
		DestWarehouseID:   warehouse.ID,
		Status:            StockTransferStatusDraft,
	}
	err = db.Create(transfer1).Error
	assert.NoError(t, err)

	transfer2 := &StockTransfer{
		TenantID:          tenant.ID,
		TransferNumber:    "TRF-UNIQUE-001", // Duplicate
		TransferDate:      time.Now(),
		SourceWarehouseID: warehouse.ID,
		DestWarehouseID:   warehouse.ID,
		Status:            StockTransferStatusDraft,
	}
	err = db.Create(transfer2).Error
	assert.Error(t, err, "Duplicate transfer number should fail")

	// Test cash transaction number uniqueness
	cash1 := &CashTransaction{
		TenantID:          tenant.ID,
		TransactionNumber: "CASH-UNIQUE-001",
		TransactionDate:   time.Now(),
		Type:              CashTransactionTypeCashIn,
		Category:          CashCategorySales,
		Amount:            decimal.NewFromInt(1000),
		BalanceBefore:     decimal.Zero,
		BalanceAfter:      decimal.NewFromInt(1000),
		Description:       "Test",
	}
	err = db.Create(cash1).Error
	assert.NoError(t, err)

	cash2 := &CashTransaction{
		TenantID:          tenant.ID,
		TransactionNumber: "CASH-UNIQUE-001", // Duplicate
		TransactionDate:   time.Now(),
		Type:              CashTransactionTypeCashIn,
		Category:          CashCategorySales,
		Amount:            decimal.NewFromInt(1000),
		BalanceBefore:     decimal.Zero,
		BalanceAfter:      decimal.NewFromInt(1000),
		Description:       "Test",
	}
	err = db.Create(cash2).Error
	assert.Error(t, err, "Duplicate cash transaction number should fail")
}

// TestCascadeDeletePhase4 tests Phase 4 foreign key constraints
func TestCascadeDeletePhase4(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Company{}, &Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{},
		&InventoryMovement{}, &StockOpname{}, &StockOpnameItem{}, &StockTransfer{}, &StockTransferItem{},
		&CashTransaction{}, &Setting{}, &User{}, &AuditLog{})
	assert.NoError(t, err)

	tenant, warehouse, product, _, _ := createPhase4TestData(t, db)

	// Test that Phase 4 models can be created with proper foreign keys
	movement := &InventoryMovement{
		TenantID:     tenant.ID,
		MovementDate: time.Now(),
		WarehouseID:  warehouse.ID,
		ProductID:    product.ID,
		MovementType: MovementTypeIn,
		Quantity:     decimal.NewFromInt(10),
		StockBefore:  decimal.Zero,
		StockAfter:   decimal.NewFromInt(10),
	}
	err = db.Create(movement).Error
	assert.NoError(t, err)

	opname := &StockOpname{
		TenantID:     tenant.ID,
		OpnameNumber: "OP-FK-001",
		OpnameDate:   time.Now(),
		WarehouseID:  warehouse.ID,
		Status:       StockOpnameStatusDraft,
	}
	err = db.Create(opname).Error
	assert.NoError(t, err)

	cashTxn := &CashTransaction{
		TenantID:          tenant.ID,
		TransactionNumber: "CASH-FK-001",
		TransactionDate:   time.Now(),
		Type:              CashTransactionTypeCashIn,
		Category:          CashCategorySales,
		Amount:            decimal.NewFromInt(1000),
		BalanceBefore:     decimal.Zero,
		BalanceAfter:      decimal.NewFromInt(1000),
		Description:       "Test",
	}
	err = db.Create(cashTxn).Error
	assert.NoError(t, err)

	setting := &Setting{
		TenantID: &tenant.ID,
		Key:      "test.setting.fk",
		Value:    strPtr("value"),
		DataType: "STRING",
	}
	err = db.Create(setting).Error
	assert.NoError(t, err)

	user := &User{Email: "fk@test.com", PasswordHash: "hash", IsActive: true}
	err = db.Create(user).Error
	assert.NoError(t, err)

	auditLog := &AuditLog{
		TenantID: &tenant.ID,
		UserID:   &user.ID,
		Action:   "TEST",
	}
	err = db.Create(auditLog).Error
	assert.NoError(t, err)

	// Verify all records were created successfully
	var count int64
	db.Model(&InventoryMovement{}).Where("tenant_id = ?", tenant.ID).Count(&count)
	assert.Equal(t, int64(1), count, "InventoryMovement should exist")

	db.Model(&StockOpname{}).Where("tenant_id = ?", tenant.ID).Count(&count)
	assert.Equal(t, int64(1), count, "StockOpname should exist")

	db.Model(&CashTransaction{}).Where("tenant_id = ?", tenant.ID).Count(&count)
	assert.Equal(t, int64(1), count, "CashTransaction should exist")

	db.Model(&Setting{}).Where("tenant_id = ?", tenant.ID).Count(&count)
	assert.Equal(t, int64(1), count, "Setting should exist")

	db.Model(&AuditLog{}).Where("tenant_id = ?", tenant.ID).Count(&count)
	assert.Equal(t, int64(1), count, "AuditLog should exist")
}

// TestDecimalPrecisionPhase4 tests decimal precision for Phase 4 models
func TestDecimalPrecisionPhase4(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Company{}, &Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{},
		&InventoryMovement{}, &StockOpnameItem{}, &StockTransferItem{}, &CashTransaction{})
	assert.NoError(t, err)

	tenant, warehouse, product, batch, _ := createPhase4TestData(t, db)

	// Test InventoryMovement quantity precision (15,3)
	movement := &InventoryMovement{
		TenantID:     tenant.ID,
		MovementDate: time.Now(),
		WarehouseID:  warehouse.ID,
		ProductID:    product.ID,
		BatchID:      &batch.ID,
		MovementType: MovementTypeIn,
		Quantity:     decimal.RequireFromString("123.456"),
		StockBefore:  decimal.RequireFromString("100.123"),
		StockAfter:   decimal.RequireFromString("223.579"),
	}
	err = db.Create(movement).Error
	assert.NoError(t, err)

	var loadedMovement InventoryMovement
	db.First(&loadedMovement, "id = ?", movement.ID)
	assert.Equal(t, "123.456", loadedMovement.Quantity.String())
	assert.Equal(t, "100.123", loadedMovement.StockBefore.String())
	assert.Equal(t, "223.579", loadedMovement.StockAfter.String())

	// Test CashTransaction amount precision (15,2)
	cashTxn := &CashTransaction{
		TenantID:          tenant.ID,
		TransactionNumber: "CASH-DECIMAL-001",
		TransactionDate:   time.Now(),
		Type:              CashTransactionTypeCashIn,
		Category:          CashCategorySales,
		Amount:            decimal.RequireFromString("1234567.89"),
		BalanceBefore:     decimal.RequireFromString("5000000.50"),
		BalanceAfter:      decimal.RequireFromString("6234568.39"),
		Description:       "Decimal precision test",
	}
	err = db.Create(cashTxn).Error
	assert.NoError(t, err)

	var loadedCash CashTransaction
	db.First(&loadedCash, "id = ?", cashTxn.ID)
	assert.Equal(t, "1234567.89", loadedCash.Amount.StringFixed(2))
	assert.Equal(t, "5000000.50", loadedCash.BalanceBefore.StringFixed(2))
	assert.Equal(t, "6234568.39", loadedCash.BalanceAfter.StringFixed(2))
}

// TestEnumValuesPhase4 tests enum value constraints for Phase 4 models
func TestEnumValuesPhase4(t *testing.T) {
	db := setupPhase4TestDB(t)
	err := db.AutoMigrate(&Tenant{}, &Warehouse{}, &Product{}, &ProductBatch{}, &WarehouseStock{},
		&InventoryMovement{}, &StockOpname{}, &StockTransfer{}, &CashTransaction{})
	assert.NoError(t, err)

	tenant, warehouse, product, _, _ := createPhase4TestData(t, db)

	// Test MovementType enum
	validMovementTypes := []MovementType{
		MovementTypeIn, MovementTypeOut, MovementTypeAdjustment,
		MovementTypeReturn, MovementTypeDamaged, MovementTypeTransfer,
	}

	for _, movementType := range validMovementTypes {
		movement := &InventoryMovement{
			TenantID:     tenant.ID,
			MovementDate: time.Now(),
			WarehouseID:  warehouse.ID,
			ProductID:    product.ID,
			MovementType: movementType,
			Quantity:     decimal.NewFromInt(10),
			StockBefore:  decimal.Zero,
			StockAfter:   decimal.NewFromInt(10),
		}
		err = db.Create(movement).Error
		assert.NoError(t, err)
	}

	// Test StockOpnameStatus enum
	validOpnameStatuses := []StockOpnameStatus{
		StockOpnameStatusDraft, StockOpnameStatusCompleted,
		StockOpnameStatusApproved, StockOpnameStatusCancelled,
	}

	for _, status := range validOpnameStatuses {
		opname := &StockOpname{
			TenantID:     tenant.ID,
			OpnameNumber: "OP-ENUM-" + string(status),
			OpnameDate:   time.Now(),
			WarehouseID:  warehouse.ID,
			Status:       status,
		}
		err = db.Create(opname).Error
		assert.NoError(t, err)
	}

	// Test StockTransferStatus enum
	validTransferStatuses := []StockTransferStatus{
		StockTransferStatusDraft, StockTransferStatusShipped,
		StockTransferStatusReceived, StockTransferStatusCancelled,
	}

	for _, status := range validTransferStatuses {
		transfer := &StockTransfer{
			TenantID:          tenant.ID,
			TransferNumber:    "TRF-ENUM-" + string(status),
			TransferDate:      time.Now(),
			SourceWarehouseID: warehouse.ID,
			DestWarehouseID:   warehouse.ID,
			Status:            status,
		}
		err = db.Create(transfer).Error
		assert.NoError(t, err)
	}

	// Test CashTransactionType and CashCategory enums
	validCashTypes := []CashTransactionType{CashTransactionTypeCashIn, CashTransactionTypeCashOut}
	validCashCategories := []CashCategory{
		CashCategorySales, CashCategoryPurchase, CashCategoryExpense,
		CashCategoryPayroll, CashCategoryLoan, CashCategoryInvestment,
		CashCategoryWithdrawal, CashCategoryDeposit,
		CashCategoryOtherIncome, CashCategoryOtherExpense,
	}

	for i, cashType := range validCashTypes {
		for j, category := range validCashCategories {
			cashTxn := &CashTransaction{
				TenantID:          tenant.ID,
				TransactionNumber: "CASH-ENUM-" + string(cashType) + "-" + string(category),
				TransactionDate:   time.Now(),
				Type:              cashType,
				Category:          category,
				Amount:            decimal.NewFromInt(1000),
				BalanceBefore:     decimal.NewFromInt(int64(i * 1000)),
				BalanceAfter:      decimal.NewFromInt(int64((i + 1) * 1000)),
				Description:       "Test enum: " + string(cashType) + " - " + string(category),
			}
			err = db.Create(cashTxn).Error
			assert.NoError(t, err, "Failed to create cash transaction with type=%s category=%s", cashType, category)
			_ = j // Suppress unused variable warning
		}
	}
}
