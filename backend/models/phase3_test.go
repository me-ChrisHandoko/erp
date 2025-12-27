// Package models - Phase 3 comprehensive tests
package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPhase3TestDB creates an in-memory SQLite database for Phase 3 testing
func setupPhase3TestDB(t *testing.T) *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Enable foreign keys for CASCADE
	database.Exec("PRAGMA foreign_keys = ON")

	// Run Phase 1 migration (dependencies)
	if err := autoMigratePhase1(database); err != nil {
		t.Fatalf("Phase 1 auto-migration failed: %v", err)
	}

	// Run Phase 2 migration (dependencies)
	if err := autoMigratePhase2(database); err != nil {
		t.Fatalf("Phase 2 auto-migration failed: %v", err)
	}

	// Run Phase 3 migration
	if err := autoMigratePhase3(database); err != nil {
		t.Fatalf("Phase 3 auto-migration failed: %v", err)
	}

	return database
}

// autoMigratePhase3 runs GORM auto-migration for Phase 3 models (local copy)
func autoMigratePhase3(db *gorm.DB) error {
	return db.AutoMigrate(
		&SalesOrder{},
		&SalesOrderItem{},
		&Delivery{},
		&DeliveryItem{},
		&Invoice{},
		&InvoiceItem{},
		&Payment{},
		&PaymentCheck{},
		&PurchaseOrder{},
		&PurchaseOrderItem{},
		&GoodsReceipt{},
		&GoodsReceiptItem{},
		&SupplierPayment{},
	)
}

// createPhase3TestData creates test data for Phase 3 models
func createPhase3TestData(t *testing.T, database *gorm.DB) (tenant *Tenant, customer *Customer, supplier *Supplier, warehouse *Warehouse, product *Product) {
	// Create tenant
	company := &Company{
		Name:      "Test Company Phase 3",
		LegalName: "Test Company Phase 3 Ltd",
		Address:   "Test Address",
		City:      "Test City",
		Province:  "Test Province",
	}
	database.Create(company)

	tenant = &Tenant{
		Name:      "Test Tenant Phase 3",
		Subdomain: "test-tenant-phase3",
		Status:    TenantStatusActive,
	}
	database.Create(tenant)

	// Link company to tenant
	company.TenantID = tenant.ID
	database.Save(company)

	// Create customer
	customer = &Customer{
		TenantID:    tenant.ID,
		Code:        "CUST-TEST",
		Name:        "Test Customer",
		PaymentTerm: 30,
		IsActive:    true,
	}
	database.Create(customer)

	// Create supplier
	supplier = &Supplier{
		TenantID:    tenant.ID,
		Code:        "SUPP-TEST",
		Name:        "Test Supplier",
		PaymentTerm: 45,
		IsActive:    true,
	}
	database.Create(supplier)

	// Create warehouse
	warehouse = &Warehouse{
		TenantID: tenant.ID,
		Code:     "WH-TEST",
		Name:     "Test Warehouse",
		Type:     WarehouseTypeMain,
		IsActive: true,
	}
	database.Create(warehouse)

	// Create product
	product = &Product{
		TenantID:  tenant.ID,
		Code:      "PROD-TEST",
		Name:      "Test Product",
		BaseUnit:  "PCS",
		BaseCost:  decimal.NewFromFloat(1000),
		BasePrice: decimal.NewFromFloat(1500),
		IsActive:  true,
	}
	database.Create(product)

	return
}

// TestPhase3SchemaGeneration verifies all Phase 3 tables are created
func TestPhase3SchemaGeneration(t *testing.T) {
	database := setupPhase3TestDB(t)

	tables := []string{
		"sales_orders", "sales_order_items",
		"deliveries", "delivery_items",
		"invoices", "invoice_items",
		"payments", "payment_checks",
		"purchase_orders", "purchase_order_items",
		"goods_receipts", "goods_receipt_items",
		"supplier_payments",
	}

	for _, table := range tables {
		if !database.Migrator().HasTable(table) {
			t.Errorf("Table %s was not created", table)
		}
	}
}

// TestSalesOrderCreation validates SalesOrder workflow
func TestSalesOrderCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, product := createPhase3TestData(t, database)

	// Create sales order
	salesOrder := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-001/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		Status:      SalesOrderStatusDraft,
		Subtotal:    decimal.NewFromFloat(150000),
		TaxAmount:   decimal.NewFromFloat(16500),
		TotalAmount: decimal.NewFromFloat(166500),
	}
	result := database.Create(salesOrder)
	if result.Error != nil {
		t.Fatalf("Failed to create sales order: %v", result.Error)
	}

	if salesOrder.ID == "" {
		t.Error("SalesOrder ID (CUID) was not generated")
	}

	// Create sales order item
	soItem := &SalesOrderItem{
		SalesOrderID: salesOrder.ID,
		ProductID:    product.ID,
		Quantity:     decimal.NewFromFloat(100),
		UnitPrice:    decimal.NewFromFloat(1500),
		Subtotal:     decimal.NewFromFloat(150000),
	}
	database.Create(soItem)

	// Verify relationship
	var loadedSO SalesOrder
	database.Preload("Items").First(&loadedSO, "id = ?", salesOrder.ID)
	if len(loadedSO.Items) != 1 {
		t.Errorf("Expected 1 sales order item, got %d", len(loadedSO.Items))
	}
}

// TestDeliveryCreation validates Delivery workflow
func TestDeliveryCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, warehouse, product := createPhase3TestData(t, database)

	// Create sales order
	salesOrder := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-002/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		Status:      SalesOrderStatusConfirmed,
		TotalAmount: decimal.NewFromFloat(100000),
	}
	database.Create(salesOrder)

	soItem := &SalesOrderItem{
		SalesOrderID: salesOrder.ID,
		ProductID:    product.ID,
		Quantity:     decimal.NewFromFloat(50),
		UnitPrice:    decimal.NewFromFloat(2000),
		Subtotal:     decimal.NewFromFloat(100000),
	}
	database.Create(soItem)

	// Create delivery
	delivery := &Delivery{
		TenantID:       tenant.ID,
		DeliveryNumber: "DEL-001/12/2025",
		DeliveryDate:   time.Now(),
		SalesOrderID:   salesOrder.ID,
		WarehouseID:    warehouse.ID,
		CustomerID:     customer.ID,
		Type:           DeliveryTypeNormal,
		Status:         DeliveryStatusPrepared,
		DriverName:     stringPtr("John Driver"),
		VehicleNumber:  stringPtr("B 1234 ABC"),
	}
	result := database.Create(delivery)
	if result.Error != nil {
		t.Fatalf("Failed to create delivery: %v", result.Error)
	}

	if delivery.ID == "" {
		t.Error("Delivery ID (CUID) was not generated")
	}

	// Create delivery item
	deliveryItem := &DeliveryItem{
		DeliveryID:       delivery.ID,
		SalesOrderItemID: soItem.ID,
		ProductID:        product.ID,
		Quantity:         decimal.NewFromFloat(50),
	}
	database.Create(deliveryItem)
}

// TestInvoiceCreation validates Invoice workflow
func TestInvoiceCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, product := createPhase3TestData(t, database)

	// Create sales order
	salesOrder := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-003/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		Status:      SalesOrderStatusConfirmed,
		TotalAmount: decimal.NewFromFloat(200000),
	}
	database.Create(salesOrder)

	// Create invoice
	invoice := &Invoice{
		TenantID:      tenant.ID,
		InvoiceNumber: "INV-001/12/2025",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		CustomerID:    customer.ID,
		SalesOrderID:  &salesOrder.ID,
		Subtotal:      decimal.NewFromFloat(200000),
		TaxAmount:     decimal.NewFromFloat(22000),
		TotalAmount:   decimal.NewFromFloat(222000),
		PaidAmount:    decimal.NewFromFloat(0),
		PaymentStatus: PaymentStatusUnpaid,
	}
	result := database.Create(invoice)
	if result.Error != nil {
		t.Fatalf("Failed to create invoice: %v", result.Error)
	}

	if invoice.ID == "" {
		t.Error("Invoice ID (CUID) was not generated")
	}

	// Create invoice item
	invoiceItem := &InvoiceItem{
		InvoiceID: invoice.ID,
		ProductID: product.ID,
		Quantity:  decimal.NewFromFloat(100),
		UnitPrice: decimal.NewFromFloat(2000),
		Subtotal:  decimal.NewFromFloat(200000),
	}
	database.Create(invoiceItem)
}

// TestPaymentCreation validates Payment workflow
func TestPaymentCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, _ := createPhase3TestData(t, database)

	// Create invoice
	invoice := &Invoice{
		TenantID:      tenant.ID,
		InvoiceNumber: "INV-002/12/2025",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		CustomerID:    customer.ID,
		TotalAmount:   decimal.NewFromFloat(100000),
		PaidAmount:    decimal.NewFromFloat(0),
		PaymentStatus: PaymentStatusUnpaid,
	}
	database.Create(invoice)

	// Create payment
	payment := &Payment{
		TenantID:      tenant.ID,
		PaymentNumber: "PAY-001/12/2025",
		PaymentDate:   time.Now(),
		CustomerID:    customer.ID,
		InvoiceID:     invoice.ID,
		Amount:        decimal.NewFromFloat(100000),
		PaymentMethod: PaymentMethodTransfer,
		Reference:     stringPtr("TRF-20251216-001"),
	}
	result := database.Create(payment)
	if result.Error != nil {
		t.Fatalf("Failed to create payment: %v", result.Error)
	}

	if payment.ID == "" {
		t.Error("Payment ID (CUID) was not generated")
	}

	// Update invoice paid amount
	invoice.PaidAmount = invoice.PaidAmount.Add(payment.Amount)
	invoice.PaymentStatus = PaymentStatusPaid
	database.Save(invoice)

	// Verify payment status
	var loadedInvoice Invoice
	database.First(&loadedInvoice, "id = ?", invoice.ID)
	if loadedInvoice.PaymentStatus != PaymentStatusPaid {
		t.Error("Invoice payment status not updated correctly")
	}
}

// TestPaymentCheckTracking validates Check/Giro tracking
func TestPaymentCheckTracking(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, _ := createPhase3TestData(t, database)
	_ = customer // Suppress unused variable warning

	// Create invoice
	invoice := &Invoice{
		TenantID:      tenant.ID,
		InvoiceNumber: "INV-003/12/2025",
		InvoiceDate:   time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 30),
		CustomerID:    customer.ID,
		TotalAmount:   decimal.NewFromFloat(500000),
		PaidAmount:    decimal.NewFromFloat(0),
		PaymentStatus: PaymentStatusUnpaid,
	}
	database.Create(invoice)

	// Create payment with check
	payment := &Payment{
		TenantID:      tenant.ID,
		PaymentNumber: "PAY-002/12/2025",
		PaymentDate:   time.Now(),
		CustomerID:    customer.ID,
		InvoiceID:     invoice.ID,
		Amount:        decimal.NewFromFloat(500000),
		PaymentMethod: PaymentMethodCheck,
	}
	database.Create(payment)

	// Create payment check record
	paymentCheck := &PaymentCheck{
		PaymentID:   payment.ID,
		CheckNumber: "CHK-2025-001",
		CheckDate:   time.Now(),
		DueDate:     time.Now().AddDate(0, 0, 30),
		Amount:      decimal.NewFromFloat(500000),
		BankName:    "BCA",
		Status:      CheckStatusIssued,
	}
	result := database.Create(paymentCheck)
	if result.Error != nil {
		t.Fatalf("Failed to create payment check: %v", result.Error)
	}

	if paymentCheck.ID == "" {
		t.Error("PaymentCheck ID (CUID) was not generated")
	}

	// Verify relationship
	var loadedPayment Payment
	database.Preload("Checks").First(&loadedPayment, "id = ?", payment.ID)
	if len(loadedPayment.Checks) != 1 {
		t.Errorf("Expected 1 payment check, got %d", len(loadedPayment.Checks))
	}
}

// TestPurchaseOrderCreation validates PurchaseOrder workflow
func TestPurchaseOrderCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, _, supplier, warehouse, product := createPhase3TestData(t, database)

	// Create purchase order
	purchaseOrder := &PurchaseOrder{
		TenantID:    tenant.ID,
		PONumber:    "PO-001/12/2025",
		PODate:      time.Now(),
		SupplierID:  supplier.ID,
		WarehouseID: warehouse.ID,
		Status:      PurchaseOrderStatusDraft,
		Subtotal:    decimal.NewFromFloat(100000),
		TaxAmount:   decimal.NewFromFloat(11000),
		TotalAmount: decimal.NewFromFloat(111000),
	}
	result := database.Create(purchaseOrder)
	if result.Error != nil {
		t.Fatalf("Failed to create purchase order: %v", result.Error)
	}

	if purchaseOrder.ID == "" {
		t.Error("PurchaseOrder ID (CUID) was not generated")
	}

	// Create purchase order item
	poItem := &PurchaseOrderItem{
		PurchaseOrderID: purchaseOrder.ID,
		ProductID:       product.ID,
		Quantity:        decimal.NewFromFloat(100),
		UnitPrice:       decimal.NewFromFloat(1000),
		Subtotal:        decimal.NewFromFloat(100000),
		ReceivedQty:     decimal.NewFromFloat(0),
	}
	database.Create(poItem)

	// Verify relationship
	var loadedPO PurchaseOrder
	database.Preload("Items").First(&loadedPO, "id = ?", purchaseOrder.ID)
	if len(loadedPO.Items) != 1 {
		t.Errorf("Expected 1 purchase order item, got %d", len(loadedPO.Items))
	}
}

// TestGoodsReceiptCreation validates GoodsReceipt workflow with quality inspection
func TestGoodsReceiptCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, _, supplier, warehouse, product := createPhase3TestData(t, database)

	// Create purchase order
	purchaseOrder := &PurchaseOrder{
		TenantID:    tenant.ID,
		PONumber:    "PO-002/12/2025",
		PODate:      time.Now(),
		SupplierID:  supplier.ID,
		WarehouseID: warehouse.ID,
		Status:      PurchaseOrderStatusConfirmed,
		TotalAmount: decimal.NewFromFloat(200000),
	}
	database.Create(purchaseOrder)

	poItem := &PurchaseOrderItem{
		PurchaseOrderID: purchaseOrder.ID,
		ProductID:       product.ID,
		Quantity:        decimal.NewFromFloat(200),
		UnitPrice:       decimal.NewFromFloat(1000),
		Subtotal:        decimal.NewFromFloat(200000),
	}
	database.Create(poItem)

	// Create goods receipt
	goodsReceipt := &GoodsReceipt{
		TenantID:        tenant.ID,
		GRNNumber:       "GRN-001/12/2025",
		GRNDate:         time.Now(),
		PurchaseOrderID: purchaseOrder.ID,
		WarehouseID:     warehouse.ID,
		SupplierID:      supplier.ID,
		Status:          GoodsReceiptStatusReceived,
	}
	result := database.Create(goodsReceipt)
	if result.Error != nil {
		t.Fatalf("Failed to create goods receipt: %v", result.Error)
	}

	if goodsReceipt.ID == "" {
		t.Error("GoodsReceipt ID (CUID) was not generated")
	}

	// Create goods receipt item with quality inspection
	grnItem := &GoodsReceiptItem{
		GoodsReceiptID:      goodsReceipt.ID,
		PurchaseOrderItemID: poItem.ID,
		ProductID:           product.ID,
		OrderedQty:          decimal.NewFromFloat(200),
		ReceivedQty:         decimal.NewFromFloat(200),
		AcceptedQty:         decimal.NewFromFloat(195), // 5 units rejected
		RejectedQty:         decimal.NewFromFloat(5),
		RejectionReason:     stringPtr("Damaged packaging"),
	}
	database.Create(grnItem)

	// Update PO item received quantity
	poItem.ReceivedQty = poItem.ReceivedQty.Add(grnItem.AcceptedQty)
	database.Save(poItem)
}

// TestSupplierPaymentCreation validates SupplierPayment workflow
func TestSupplierPaymentCreation(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, _, supplier, _, _ := createPhase3TestData(t, database)
	_ = supplier // Will be used below

	// Create supplier payment
	supplierPayment := &SupplierPayment{
		TenantID:      tenant.ID,
		PaymentNumber: "SPAY-001/12/2025",
		PaymentDate:   time.Now(),
		SupplierID:    supplier.ID,
		Amount:        decimal.NewFromFloat(500000),
		PaymentMethod: PaymentMethodTransfer,
		Reference:     stringPtr("TRF-SUPP-001"),
	}
	result := database.Create(supplierPayment)
	if result.Error != nil {
		t.Fatalf("Failed to create supplier payment: %v", result.Error)
	}

	if supplierPayment.ID == "" {
		t.Error("SupplierPayment ID (CUID) was not generated")
	}
}

// TestUniqueConstraintsPhase3 validates unique number constraints
func TestUniqueConstraintsPhase3(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, _ := createPhase3TestData(t, database)

	// Create first sales order
	so1 := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-DUP/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		TotalAmount: decimal.NewFromFloat(100000),
	}
	database.Create(so1)

	// Try to create duplicate SO number
	so2 := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-DUP/12/2025", // Same number
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		TotalAmount: decimal.NewFromFloat(200000),
	}

	result := database.Create(so2)
	if result.Error == nil {
		t.Error("Should have failed due to duplicate SO number")
	}
}

// TestCascadeDeletePhase3 validates CASCADE deletion from Tenant
func TestCascadeDeletePhase3(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, _ := createPhase3TestData(t, database)

	// Create sales order linked to tenant
	salesOrder := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-CASCADE/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		TotalAmount: decimal.NewFromFloat(100000),
	}
	database.Create(salesOrder)

	// Delete tenant (should cascade to sales orders)
	database.Select("Company").Delete(tenant)

	// Verify sales order was deleted
	var soCount int64
	database.Model(&SalesOrder{}).Where("id = ?", salesOrder.ID).Count(&soCount)
	if soCount != 0 {
		t.Error("CASCADE deletion failed - SalesOrder not deleted with Tenant")
	}
}

// TestDecimalPrecisionPhase3 validates decimal accuracy for transaction amounts
func TestDecimalPrecisionPhase3(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, _ := createPhase3TestData(t, database)

	// Create sales order with precise amounts
	preciseSubtotal := decimal.RequireFromString("123456.78")
	preciseTax := decimal.RequireFromString("13580.25")
	preciseTotal := preciseSubtotal.Add(preciseTax)

	salesOrder := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-DECIMAL/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		Subtotal:    preciseSubtotal,
		TaxAmount:   preciseTax,
		TotalAmount: preciseTotal,
	}
	database.Create(salesOrder)

	// Load from database
	var loadedSO SalesOrder
	database.First(&loadedSO, "id = ?", salesOrder.ID)

	if !loadedSO.Subtotal.Equal(preciseSubtotal) {
		t.Errorf("Decimal precision lost for subtotal: expected %s, got %s",
			preciseSubtotal.String(), loadedSO.Subtotal.String())
	}

	if !loadedSO.TotalAmount.Equal(preciseTotal) {
		t.Errorf("Decimal precision lost for total: expected %s, got %s",
			preciseTotal.String(), loadedSO.TotalAmount.String())
	}
}

// TestEnumValuesPhase3 validates enum types work correctly
func TestEnumValuesPhase3(t *testing.T) {
	database := setupPhase3TestDB(t)
	tenant, customer, _, _, _ := createPhase3TestData(t, database)

	// Create sales order with enum status
	salesOrder := &SalesOrder{
		TenantID:    tenant.ID,
		SONumber:    "SO-ENUM/12/2025",
		SODate:      time.Now(),
		CustomerID:  customer.ID,
		Status:      SalesOrderStatusDraft, // Enum value
		TotalAmount: decimal.NewFromFloat(100000),
	}
	database.Create(salesOrder)

	// Load and verify enum
	var loadedSO SalesOrder
	database.First(&loadedSO, "id = ?", salesOrder.ID)

	if loadedSO.Status != SalesOrderStatusDraft {
		t.Errorf("Enum value mismatch: expected %s, got %s",
			SalesOrderStatusDraft, loadedSO.Status)
	}

	// Test enum update
	loadedSO.Status = SalesOrderStatusConfirmed
	database.Save(&loadedSO)

	database.First(&loadedSO, "id = ?", salesOrder.ID)
	if loadedSO.Status != SalesOrderStatusConfirmed {
		t.Error("Enum value not updated correctly")
	}
}
