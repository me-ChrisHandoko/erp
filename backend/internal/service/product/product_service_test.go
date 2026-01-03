package product

import (
	"context"
	"testing"

	"backend/internal/dto"
	"backend/internal/testutil"
	"backend/models"
	pkgerrors "backend/pkg/errors"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductService_CreateProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewProductService(db)
	ctx := context.Background()

	// Setup test data
	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	warehouse := testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	t.Run("success - create product with base unit", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code:         "PROD001",
			Name:         "Test Product 1",
			BaseUnit:     "PCS",
			BaseCost:     "10000",
			BasePrice:    "15000",
			MinimumStock: "10",
			Category:     stringPtr("Electronics"),
		}

		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)

		require.NoError(t, err)
		assert.NotEmpty(t, product.ID)
		assert.Equal(t, "PROD001", product.Code)
		assert.Equal(t, "Test Product 1", product.Name)
		assert.Equal(t, "PCS", product.BaseUnit)
		assert.True(t, product.IsActive)

		// Verify base unit was created
		var units []models.ProductUnit
		db.Where("product_id = ?", product.ID).Find(&units)
		assert.Len(t, units, 1)
		assert.Equal(t, "PCS", units[0].UnitName)
		assert.True(t, units[0].IsBaseUnit)
		assert.Equal(t, "1", units[0].ConversionRate.String())

		// Verify warehouse stock was initialized
		var stocks []models.WarehouseStock
		db.Where("product_id = ?", product.ID).Find(&stocks)
		assert.Len(t, stocks, 1)
		assert.Equal(t, warehouse.ID, stocks[0].WarehouseID)
		assert.Equal(t, "0", stocks[0].Quantity.String())
	})

	t.Run("success - create product with additional units", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code:         "PROD002",
			Name:         "Test Product 2",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "100",
			Units: []dto.CreateProductUnitRequest{
				{
					UnitName:       "BOX",
					ConversionRate: "12",
					BuyPrice:       stringPtr("11000"),
					SellPrice:      stringPtr("17000"),
				},
			},
		}

		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)

		require.NoError(t, err)

		// Verify 2 units were created (base + additional)
		var units []models.ProductUnit
		db.Where("product_id = ?", product.ID).Order("conversion_rate ASC").Find(&units)
		assert.Len(t, units, 2)

		// Check base unit
		assert.Equal(t, "PCS", units[0].UnitName)
		assert.True(t, units[0].IsBaseUnit)

		// Check additional unit
		assert.Equal(t, "BOX", units[1].UnitName)
		assert.False(t, units[1].IsBaseUnit)
		assert.Equal(t, "12", units[1].ConversionRate.String())
	})

	t.Run("error - duplicate product code", func(t *testing.T) {
		// Create first product
		req1 := &dto.CreateProductRequest{
			Code:         "PROD003",
			Name:         "Test Product 3",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
		}
		_, err := service.CreateProduct(ctx, company.ID, company.TenantID, req1)
		require.NoError(t, err)

		// Try to create duplicate
		req2 := &dto.CreateProductRequest{
			Code:         "PROD003", // Duplicate code
			Name:         "Test Product 3 Duplicate",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
		}
		_, err = service.CreateProduct(ctx, company.ID, company.TenantID, req2)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "product code already exists")
	})

	t.Run("error - base price less than base cost", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code:         "PROD004",
			Name:         "Test Product 4",
			BaseUnit:     "PCS",
			BaseCost:     "15000",
			BasePrice:    "10000", // Less than cost
			MinimumStock: "10",
		}

		_, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "base price must be greater than or equal to base cost")
	})

	t.Run("error - negative minimum stock", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code:         "PROD005",
			Name:         "Test Product 5",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "-10", // Negative
		}

		_, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "minimum stock cannot be negative")
	})

	t.Run("error - duplicate barcode", func(t *testing.T) {
		// Create first product with barcode
		req1 := &dto.CreateProductRequest{
			Code:         "PROD006",
			Name:         "Test Product 6",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
			Barcode:      stringPtr("123456789"),
		}
		_, err := service.CreateProduct(ctx, company.ID, company.TenantID, req1)
		require.NoError(t, err)

		// Try to create with duplicate barcode
		req2 := &dto.CreateProductRequest{
			Code:         "PROD007",
			Name:         "Test Product 7",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
			Barcode:      stringPtr("123456789"), // Duplicate
		}
		_, err = service.CreateProduct(ctx, company.ID, company.TenantID, req2)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "barcode already exists")
	})
}

func TestProductService_GetProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewProductService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	t.Run("success - get existing product", func(t *testing.T) {
		// Create product
		req := &dto.CreateProductRequest{
			Code:         "PROD001",
			Name:         "Test Product 1",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
		}
		created, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)
		require.NoError(t, err)

		// Get product
		product, err := service.GetProduct(ctx, company.ID, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, product.ID)
		assert.Equal(t, "PROD001", product.Code)
		assert.Equal(t, "Test Product 1", product.Name)
	})

	t.Run("error - product not found", func(t *testing.T) {
		_, err := service.GetProduct(ctx, company.ID, "nonexistent-id")

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("error - multi-company isolation", func(t *testing.T) {
		// Create product in company 1
		req := &dto.CreateProductRequest{
			Code:         "PROD002",
			Name:         "Test Product 2",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
		}
		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)
		require.NoError(t, err)

		// Create company 2
		company2 := testutil.CreateTestCompany(t, db, "tenant1", "COMP2")

		// Try to get product from company 2 context
		_, err = service.GetProduct(ctx, company2.ID, product.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

func TestProductService_ListProducts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewProductService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	// Create test products
	products := []dto.CreateProductRequest{
		{Code: "PROD001", Name: "Product Alpha", BaseUnit: "PCS", BaseCost: "1000", BasePrice: "1500", MinimumStock: "10", Category: stringPtr("Electronics")},
		{Code: "PROD002", Name: "Product Beta", BaseUnit: "PCS", BaseCost: "2000", BasePrice: "3000", MinimumStock: "20", Category: stringPtr("Electronics")},
		{Code: "PROD003", Name: "Product Gamma", BaseUnit: "KG", BaseCost: "500", BasePrice: "800", MinimumStock: "50", Category: stringPtr("Food")},
		{Code: "PROD004", Name: "Product Delta", BaseUnit: "PCS", BaseCost: "3000", BasePrice: "4500", MinimumStock: "5", Category: stringPtr("Furniture")},
	}

	for _, req := range products {
		_, err := service.CreateProduct(ctx, company.ID, company.TenantID, &req)
		require.NoError(t, err)
	}

	t.Run("success - list all products", func(t *testing.T) {
		query := &dto.ProductFilters{
			Page:     1,
			Limit: 10,
		}

		products, count, err := service.ListProducts(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
		assert.Len(t, products, 4)
	})

	t.Run("success - pagination", func(t *testing.T) {
		query := &dto.ProductFilters{
			Page:     1,
			Limit: 2,
		}

		products, count, err := service.ListProducts(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
		assert.Len(t, products, 2)
	})

	t.Run("success - filter by search", func(t *testing.T) {
		query := &dto.ProductFilters{
			Search:   "Alpha",
			Page:     1,
			Limit: 10,
		}

		products, count, err := service.ListProducts(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Len(t, products, 1)
		assert.Equal(t, "Product Alpha", products[0].Name)
	})

	t.Run("success - filter by category", func(t *testing.T) {
		query := &dto.ProductFilters{
			Category: "Electronics",
			Page:     1,
			Limit:    10,
		}

		products, count, err := service.ListProducts(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
		assert.Len(t, products, 2)
	})

	t.Run("success - sort by name ascending", func(t *testing.T) {
		query := &dto.ProductFilters{
			Page:      1,
			Limit:  10,
			SortBy:    "name",
			SortOrder: "asc",
		}

		products, count, err := service.ListProducts(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
		assert.Equal(t, "Product Alpha", products[0].Name)
		assert.Equal(t, "Product Beta", products[1].Name)
	})

	t.Run("success - filter inactive products", func(t *testing.T) {
		// Deactivate one product
		var product models.Product
		db.Where("code = ?", "PROD001").First(&product)
		product.IsActive = false
		db.Save(&product)

		isActive := true
		query := &dto.ProductFilters{
			IsActive: &isActive,
			Page:     1,
			Limit: 10,
		}

		products, count, err := service.ListProducts(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(3), count) // Only 3 active products
		assert.Len(t, products, 3)
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewProductService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	t.Run("success - update product fields", func(t *testing.T) {
		// Create product
		req := &dto.CreateProductRequest{
			Code:         "PROD001",
			Name:         "Original Name",
			BaseUnit:     "PCS",
			BaseCost:     "1000",
			BasePrice:    "1500",
			MinimumStock: "10",
		}
		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)
		require.NoError(t, err)

		// Update product
		updateReq := &dto.UpdateProductRequest{
			Name:         stringPtr("Updated Name"),
			BasePrice:    stringPtr("2000"),
			MinimumStock: stringPtr("20"),
		}

		updated, err := service.UpdateProduct(ctx, company.ID, product.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "2000", updated.BasePrice.String())
		assert.Equal(t, "20", updated.MinimumStock.String())
		assert.Equal(t, "PROD001", updated.Code) // Unchanged
	})

	t.Run("error - update base price below cost", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code: "PROD005", Name: "Product 5", BaseUnit: "PCS",
			BaseCost: "1000", BasePrice: "1500", MinimumStock: "10",
		}
		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateProductRequest{
			BasePrice: stringPtr("500"), // Below cost
		}

		_, err = service.UpdateProduct(ctx, company.ID, product.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "base price must be greater than or equal to base cost")
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewProductService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	warehouse := testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	t.Run("success - delete product with zero stock", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code: "PROD001", Name: "Product 1", BaseUnit: "PCS",
			BaseCost: "1000", BasePrice: "1500", MinimumStock: "10",
		}
		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)
		require.NoError(t, err)

		// Delete product
		err = service.DeleteProduct(ctx, company.ID, product.ID)

		require.NoError(t, err)

		// Verify soft delete (IsActive = false)
		var deleted models.Product
		db.Unscoped().Where("id = ?", product.ID).First(&deleted)
		assert.False(t, deleted.IsActive)
	})

	t.Run("error - delete product with stock", func(t *testing.T) {
		req := &dto.CreateProductRequest{
			Code: "PROD002", Name: "Product 2", BaseUnit: "PCS",
			BaseCost: "1000", BasePrice: "1500", MinimumStock: "10",
		}
		product, err := service.CreateProduct(ctx, company.ID, company.TenantID, req)
		require.NoError(t, err)

		// Add stock to warehouse
		var stock models.WarehouseStock
		db.Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).First(&stock)
		stock.Quantity = decimal.NewFromInt(100)
		db.Save(&stock)

		// Try to delete
		err = service.DeleteProduct(ctx, company.ID, product.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete product with stock")
	})

	t.Run("error - product not found", func(t *testing.T) {
		err := service.DeleteProduct(ctx, company.ID, "nonexistent-id")

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
