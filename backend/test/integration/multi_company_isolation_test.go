package integration

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/dto"
	"backend/internal/service/customer"
	"backend/internal/service/product"
	"backend/internal/service/supplier"
	"backend/internal/service/warehouse"
	"backend/internal/testutil"
	"backend/models"
	pkgerrors "backend/pkg/errors"
)

// TestMultiCompanyIsolation_ProductService verifies that Company A cannot access Company B's products
func TestMultiCompanyIsolation_ProductService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := product.NewProductService(db)
	ctx := context.Background()

	// Create two separate companies
	companyA := testutil.CreateTestCompany(t, db, "tenant1", "Company A")
	companyB := testutil.CreateTestCompany(t, db, "tenant1", "Company B")

	// Company A creates products
	_, err := service.CreateProduct(ctx, companyA.ID, companyA.TenantID, &dto.CreateProductRequest{
		Code:         "PROD-A1",
		Name:         "Product A1",
		Category:     stringPtr("Electronics"),
		BaseUnit:     "PCS",
		BaseCost:     "10000",
		BasePrice:    "15000",
		MinimumStock: "10",
	})
	require.NoError(t, err)

	_, err = service.CreateProduct(ctx, companyA.ID, companyA.TenantID, &dto.CreateProductRequest{
		Code:         "PROD-A2",
		Name:         "Product A2",
		Category:     stringPtr("Electronics"),
		BaseUnit:     "PCS",
		BaseCost:     "20000",
		BasePrice:    "25000",
		MinimumStock: "5",
	})
	require.NoError(t, err)

	// Company B creates products
	productB1, err := service.CreateProduct(ctx, companyB.ID, companyB.TenantID, &dto.CreateProductRequest{
		Code:         "PROD-B1",
		Name:         "Product B1",
		Category:     stringPtr("Furniture"),
		BaseUnit:     "PCS",
		BaseCost:     "50000",
		BasePrice:    "75000",
		MinimumStock: "3",
	})
	require.NoError(t, err)

	t.Run("Company A cannot read Company B's product", func(t *testing.T) {
		// Company A tries to read Company B's product
		_, err := service.GetProduct(ctx, companyA.ID, productB1.ID)

		// Should return NotFound (not AccessDenied to prevent data leakage)
		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok, "Error should be AppError")
		assert.Equal(t, 404, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "product not found")
	})

	t.Run("Company A cannot update Company B's product", func(t *testing.T) {
		// Company A tries to update Company B's product
		updateReq := &dto.UpdateProductRequest{
			Name:      stringPtr("Hacked Product"),
			BasePrice: stringPtr("1"),
		}

		_, err := service.UpdateProduct(ctx, companyA.ID, productB1.ID, updateReq)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("Company A cannot delete Company B's product", func(t *testing.T) {
		// Company A tries to delete Company B's product
		err := service.DeleteProduct(ctx, companyA.ID, productB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)

		// Verify product still exists for Company B
		product, err := service.GetProduct(ctx, companyB.ID, productB1.ID)
		require.NoError(t, err)
		assert.Equal(t, "PROD-B1", product.Code)
	})

	t.Run("Company A can only list its own products", func(t *testing.T) {
		// Company A lists products
		productsA, countA, err := service.ListProducts(ctx, companyA.ID, &dto.ProductFilters{
			Page:  1,
			Limit: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(2), countA)
		assert.Len(t, productsA, 2)

		// Verify only Company A's products are returned
		codes := []string{productsA[0].Code, productsA[1].Code}
		assert.Contains(t, codes, "PROD-A1")
		assert.Contains(t, codes, "PROD-A2")
		assert.NotContains(t, codes, "PROD-B1")

		// Company B lists products
		productsB, countB, err := service.ListProducts(ctx, companyB.ID, &dto.ProductFilters{
			Page:  1,
			Limit: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), countB)
		assert.Len(t, productsB, 1)
		assert.Equal(t, "PROD-B1", productsB[0].Code)
	})

	t.Run("Company A can use Company B's product code (per-company uniqueness)", func(t *testing.T) {
		// This should succeed - product codes are unique per company, not globally
		productA3, err := service.CreateProduct(ctx, companyA.ID, companyA.TenantID, &dto.CreateProductRequest{
			Code:         "PROD-B1", // Same code as Company B's product
			Name:         "Product A3",
			Category:     stringPtr("Electronics"),
			BaseUnit:     "PCS",
			BaseCost:     "10000",
			BasePrice:    "15000",
			MinimumStock: "5",
		})

		require.NoError(t, err, "Company A should be able to use same code as Company B")
		assert.Equal(t, "PROD-B1", productA3.Code)
		assert.Equal(t, companyA.ID, productA3.CompanyID)
	})

	t.Run("Company A and Company B have separate product counts", func(t *testing.T) {
		// After creating PROD-A1, PROD-A2, PROD-A3 for Company A
		_, countA, err := service.ListProducts(ctx, companyA.ID, &dto.ProductFilters{Page: 1, Limit: 100})
		require.NoError(t, err)
		assert.Equal(t, int64(3), countA)

		// Company B still has only 1 product
		_, countB, err := service.ListProducts(ctx, companyB.ID, &dto.ProductFilters{Page: 1, Limit: 100})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countB)
	})
}

// TestMultiCompanyIsolation_CustomerService verifies customer data isolation
func TestMultiCompanyIsolation_CustomerService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := customer.NewCustomerService(db)
	ctx := context.Background()

	companyA := testutil.CreateTestCompany(t, db, "tenant1", "Company A")
	companyB := testutil.CreateTestCompany(t, db, "tenant1", "Company B")

	// Company A creates customers
	_, err := service.CreateCustomer(ctx, companyA.ID, &dto.CreateCustomerRequest{
		Code:        "CUST-A1",
		Name:        "Customer A1",
		Phone:       stringPtr("08111111111"),
		PaymentTerm: intPtr(30),
	})
	require.NoError(t, err)

	// Company B creates customers
	customerB1, err := service.CreateCustomer(ctx, companyB.ID, &dto.CreateCustomerRequest{
		Code:        "CUST-B1",
		Name:        "Customer B1",
		Phone:       stringPtr("08222222222"),
		PaymentTerm: intPtr(45),
	})
	require.NoError(t, err)

	t.Run("Company A cannot read Company B's customer", func(t *testing.T) {
		_, err := service.GetCustomerByID(ctx, companyA.ID, customerB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "customer not found")
	})

	t.Run("Company A cannot update Company B's customer", func(t *testing.T) {
		updateReq := &dto.UpdateCustomerRequest{
			Name: stringPtr("Hacked Customer"),
		}

		_, err := service.UpdateCustomer(ctx, companyA.ID, customerB1.ID, updateReq)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("Company A cannot delete Company B's customer", func(t *testing.T) {
		err := service.DeleteCustomer(ctx, companyA.ID, customerB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)

		// Verify customer still exists for Company B
		customer, err := service.GetCustomerByID(ctx, companyB.ID, customerB1.ID)
		require.NoError(t, err)
		assert.Equal(t, "CUST-B1", customer.Code)
	})

	t.Run("Company A can only list its own customers", func(t *testing.T) {
		resultA, err := service.ListCustomers(ctx, companyA.ID, &dto.CustomerListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultA.TotalCount)
		assert.Equal(t, "CUST-A1", resultA.Customers[0].Code)

		resultB, err := service.ListCustomers(ctx, companyB.ID, &dto.CustomerListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultB.TotalCount)
		assert.Equal(t, "CUST-B1", resultB.Customers[0].Code)
	})

	t.Run("Customer codes are unique per company only", func(t *testing.T) {
		// Company A can use same code as Company B
		customerA2, err := service.CreateCustomer(ctx, companyA.ID, &dto.CreateCustomerRequest{
			Code:        "CUST-B1", // Same as Company B's customer
			Name:        "Customer A2",
			Phone:       stringPtr("08333333333"),
			PaymentTerm: intPtr(30),
		})

		require.NoError(t, err)
		assert.Equal(t, "CUST-B1", customerA2.Code)
		assert.Equal(t, companyA.ID, customerA2.CompanyID)
	})
}

// TestMultiCompanyIsolation_SupplierService verifies supplier data isolation
func TestMultiCompanyIsolation_SupplierService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := supplier.NewSupplierService(db)
	ctx := context.Background()

	companyA := testutil.CreateTestCompany(t, db, "tenant1", "Company A")
	companyB := testutil.CreateTestCompany(t, db, "tenant1", "Company B")

	// Company A creates suppliers
	_, err := service.CreateSupplier(ctx, companyA.ID, &dto.CreateSupplierRequest{
		Code:        "SUPP-A1",
		Name:        "Supplier A1",
		Phone:       stringPtr("08111111111"),
		PaymentTerm: intPtr(30),
	})
	require.NoError(t, err)

	// Company B creates suppliers
	supplierB1, err := service.CreateSupplier(ctx, companyB.ID, &dto.CreateSupplierRequest{
		Code:        "SUPP-B1",
		Name:        "Supplier B1",
		Phone:       stringPtr("08222222222"),
		PaymentTerm: intPtr(60),
	})
	require.NoError(t, err)

	t.Run("Company A cannot read Company B's supplier", func(t *testing.T) {
		_, err := service.GetSupplierByID(ctx, companyA.ID, supplierB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "supplier not found")
	})

	t.Run("Company A cannot update Company B's supplier", func(t *testing.T) {
		updateReq := &dto.UpdateSupplierRequest{
			Name: stringPtr("Hacked Supplier"),
		}

		_, err := service.UpdateSupplier(ctx, companyA.ID, supplierB1.ID, updateReq)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("Company A cannot delete Company B's supplier", func(t *testing.T) {
		err := service.DeleteSupplier(ctx, companyA.ID, supplierB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)

		// Verify supplier still exists for Company B
		supplier, err := service.GetSupplierByID(ctx, companyB.ID, supplierB1.ID)
		require.NoError(t, err)
		assert.Equal(t, "SUPP-B1", supplier.Code)
	})

	t.Run("Company A can only list its own suppliers", func(t *testing.T) {
		resultA, err := service.ListSuppliers(ctx, companyA.ID, &dto.SupplierListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultA.TotalCount)
		assert.Equal(t, "SUPP-A1", resultA.Suppliers[0].Code)

		resultB, err := service.ListSuppliers(ctx, companyB.ID, &dto.SupplierListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultB.TotalCount)
		assert.Equal(t, "SUPP-B1", resultB.Suppliers[0].Code)
	})
}

// TestMultiCompanyIsolation_WarehouseService verifies warehouse data isolation
func TestMultiCompanyIsolation_WarehouseService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := warehouse.NewWarehouseService(db)
	productService := product.NewProductService(db)
	ctx := context.Background()

	companyA := testutil.CreateTestCompany(t, db, "tenant1", "Company A")
	companyB := testutil.CreateTestCompany(t, db, "tenant1", "Company B")

	// Company A creates warehouse
	warehouseA1, err := service.CreateWarehouse(ctx, companyA.ID, &dto.CreateWarehouseRequest{
		Code: "WH-A1",
		Name: "Warehouse A1",
		Type: "MAIN",
	})
	require.NoError(t, err)

	// Company B creates warehouse
	warehouseB1, err := service.CreateWarehouse(ctx, companyB.ID, &dto.CreateWarehouseRequest{
		Code: "WH-B1",
		Name: "Warehouse B1",
		Type: "MAIN",
	})
	require.NoError(t, err)

	// Create products for stock testing
	productA, err := productService.CreateProduct(ctx, companyA.ID, companyA.TenantID, &dto.CreateProductRequest{
		Code:         "PROD-A",
		Name:         "Product A",
		BaseUnit:     "PCS",
		BaseCost:     "1000",
		BasePrice:    "1500",
		MinimumStock: "10",
	})
	require.NoError(t, err)

	productB, err := productService.CreateProduct(ctx, companyB.ID, companyB.TenantID, &dto.CreateProductRequest{
		Code:         "PROD-B",
		Name:         "Product B",
		BaseUnit:     "PCS",
		BaseCost:     "2000",
		BasePrice:    "2500",
		MinimumStock: "5",
	})
	require.NoError(t, err)

	// Create warehouse stocks
	stockA := &models.WarehouseStock{
		WarehouseID:  warehouseA1.ID,
		ProductID:    productA.ID,
		Quantity:     decimal.NewFromInt(100),
		MinimumStock: decimal.NewFromInt(10),
		MaximumStock: decimal.NewFromInt(500),
	}
	db.Create(stockA)

	stockB := &models.WarehouseStock{
		WarehouseID:  warehouseB1.ID,
		ProductID:    productB.ID,
		Quantity:     decimal.NewFromInt(200),
		MinimumStock: decimal.NewFromInt(5),
		MaximumStock: decimal.NewFromInt(1000),
	}
	db.Create(stockB)

	t.Run("Company A cannot read Company B's warehouse", func(t *testing.T) {
		_, err := service.GetWarehouseByID(ctx, companyA.ID, warehouseB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "warehouse not found")
	})

	t.Run("Company A cannot update Company B's warehouse", func(t *testing.T) {
		updateReq := &dto.UpdateWarehouseRequest{
			Name: stringPtr("Hacked Warehouse"),
		}

		_, err := service.UpdateWarehouse(ctx, companyA.ID, warehouseB1.ID, updateReq)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("Company A cannot delete Company B's warehouse", func(t *testing.T) {
		err := service.DeleteWarehouse(ctx, companyA.ID, warehouseB1.ID)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)

		// Verify warehouse still exists for Company B
		wh, err := service.GetWarehouseByID(ctx, companyB.ID, warehouseB1.ID)
		require.NoError(t, err)
		assert.Equal(t, "WH-B1", wh.Code)
	})

	t.Run("Company A can only list its own warehouses", func(t *testing.T) {
		resultA, err := service.ListWarehouses(ctx, companyA.ID, &dto.WarehouseListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultA.TotalCount)
		assert.Equal(t, "WH-A1", resultA.Warehouses[0].Code)

		resultB, err := service.ListWarehouses(ctx, companyB.ID, &dto.WarehouseListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultB.TotalCount)
		assert.Equal(t, "WH-B1", resultB.Warehouses[0].Code)
	})

	t.Run("Company A can only see its own warehouse stocks", func(t *testing.T) {
		resultA, err := service.ListWarehouseStocks(ctx, companyA.ID, &dto.WarehouseStockListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultA.TotalCount)
		assert.Equal(t, "PROD-A", resultA.Stocks[0].ProductCode)

		resultB, err := service.ListWarehouseStocks(ctx, companyB.ID, &dto.WarehouseStockListQuery{
			Page:     1,
			PageSize: 100,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(1), resultB.TotalCount)
		assert.Equal(t, "PROD-B", resultB.Stocks[0].ProductCode)
	})

	t.Run("Company A cannot update Company B's warehouse stock", func(t *testing.T) {
		updateReq := &dto.UpdateWarehouseStockRequest{
			MinimumStock: stringPtr("1"),
		}

		_, err := service.UpdateWarehouseStock(ctx, companyA.ID, stockB.ID, updateReq)

		require.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

// TestMultiCompanyIsolation_CrossServiceScenario tests complex scenarios across services
func TestMultiCompanyIsolation_CrossServiceScenario(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	productSvc := product.NewProductService(db)
	warehouseSvc := warehouse.NewWarehouseService(db)
	ctx := context.Background()

	companyA := testutil.CreateTestCompany(t, db, "tenant1", "Company A")
	companyB := testutil.CreateTestCompany(t, db, "tenant1", "Company B")

	t.Run("Cannot create warehouse stock with Company B's product in Company A's warehouse", func(t *testing.T) {
		// Company A creates warehouse
		warehouseA, err := warehouseSvc.CreateWarehouse(ctx, companyA.ID, &dto.CreateWarehouseRequest{
			Code: "WH-A",
			Name: "Warehouse A",
			Type: "MAIN",
		})
		require.NoError(t, err)

		// Company B creates product
		productB, err := productSvc.CreateProduct(ctx, companyB.ID, companyB.TenantID, &dto.CreateProductRequest{
			Code:         "PROD-B",
			Name:         "Product B",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
		})
		require.NoError(t, err)

		// Try to create stock in Company A's warehouse with Company B's product
		// This should fail because the product doesn't exist for Company A
		stock := &models.WarehouseStock{
			WarehouseID:  warehouseA.ID,
			ProductID:    productB.ID,
			Quantity:     decimal.NewFromInt(100),
			MinimumStock: decimal.NewFromInt(10),
			MaximumStock: decimal.NewFromInt(500),
		}

		err = db.Create(stock).Error
		// This will succeed at DB level but will fail when queried through service
		// because service filters by company_id
		require.NoError(t, err, "DB allows creation but service layer prevents access")

		// Verify Company A cannot see this stock
		stocks, err := warehouseSvc.ListWarehouseStocks(ctx, companyA.ID, &dto.WarehouseStockListQuery{
			Page:     1,
			PageSize: 100,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), stocks.TotalCount, "Company A should not see stock with Company B's product")
	})

	t.Run("Product search across companies returns only company-specific results", func(t *testing.T) {
		// Create products with similar names in both companies
		_, err := productSvc.CreateProduct(ctx, companyA.ID, companyA.TenantID, &dto.CreateProductRequest{
			Code:         "LAPTOP-001",
			Name:         "Laptop Dell",
			BaseUnit:     "PCS",
			BaseCost:     "5000000",
			BasePrice:    "6000000",
			MinimumStock: "5",
		})
		require.NoError(t, err)

		_, err = productSvc.CreateProduct(ctx, companyB.ID, companyB.TenantID, &dto.CreateProductRequest{
			Code:         "LAPTOP-001",
			Name:         "Laptop Dell",
			BaseUnit:     "PCS",
			BaseCost:     "5500000",
			BasePrice:    "6500000",
			MinimumStock: "3",
		})
		require.NoError(t, err)

		// Search for "Laptop" in Company A
		productsA, countA, err := productSvc.ListProducts(ctx, companyA.ID, &dto.ProductFilters{
			Search: "Laptop",
			Page:   1,
			Limit:  100,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countA)
		assert.Equal(t, companyA.ID, productsA[0].CompanyID)
		assert.Equal(t, decimal.NewFromInt(6000000), productsA[0].BasePrice)

		// Search for "Laptop" in Company B
		productsB, countB, err := productSvc.ListProducts(ctx, companyB.ID, &dto.ProductFilters{
			Search: "Laptop",
			Page:   1,
			Limit:  100,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), countB)
		assert.Equal(t, companyB.ID, productsB[0].CompanyID)
		assert.Equal(t, decimal.NewFromInt(6500000), productsB[0].BasePrice)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
