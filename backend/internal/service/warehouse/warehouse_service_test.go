package warehouse

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

func TestWarehouseService_CreateWarehouse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - create warehouse with minimal fields", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{
			Code: "WH001",
			Name: "Warehouse One",
			Type: "MAIN",
		}

		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)

		require.NoError(t, err)
		assert.NotEmpty(t, warehouse.ID)
		assert.Equal(t, "WH001", warehouse.Code)
		assert.Equal(t, "Warehouse One", warehouse.Name)
		assert.Equal(t, models.WarehouseTypeMain, warehouse.Type)
		assert.True(t, warehouse.IsActive)
	})

	t.Run("success - create warehouse with all fields", func(t *testing.T) {
		user := testutil.CreateTestUser(t, db, "manager@example.com")

		req := &dto.CreateWarehouseRequest{
			Code:       "WH002",
			Name:       "Warehouse Two",
			Type:       "BRANCH",
			Address:    stringPtr("Jl. Test No. 123"),
			City:       stringPtr("Jakarta"),
			Province:   stringPtr("DKI Jakarta"),
			PostalCode: stringPtr("12345"),
			Phone:      stringPtr("08123456789"),
			Email:      stringPtr("wh2@example.com"),
			ManagerID:  &user.ID,
			Capacity:   stringPtr("1000.50"),
		}

		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)

		require.NoError(t, err)
		assert.Equal(t, "WH002", warehouse.Code)
		assert.Equal(t, models.WarehouseTypeBranch, warehouse.Type)
		assert.NotNil(t, warehouse.ManagerID)
		assert.Equal(t, user.ID, *warehouse.ManagerID)
		assert.NotNil(t, warehouse.Capacity)
		assert.Equal(t, "1000.5", warehouse.Capacity.String())
	})

	t.Run("error - duplicate warehouse code", func(t *testing.T) {
		req1 := &dto.CreateWarehouseRequest{
			Code: "WH003",
			Name: "Warehouse Three",
			Type: "MAIN",
		}
		_, err := service.CreateWarehouse(ctx, company.ID, req1)
		require.NoError(t, err)

		req2 := &dto.CreateWarehouseRequest{
			Code: "WH003", // Duplicate
			Name: "Warehouse Three Duplicate",
			Type: "MAIN",
		}
		_, err = service.CreateWarehouse(ctx, company.ID, req2)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "warehouse code already exists")
	})

	t.Run("error - invalid manager ID", func(t *testing.T) {
		invalidManagerID := "nonexistent-user-id"
		req := &dto.CreateWarehouseRequest{
			Code:      "WH004",
			Name:      "Warehouse Four",
			Type:      "MAIN",
			ManagerID: &invalidManagerID,
		}

		_, err := service.CreateWarehouse(ctx, company.ID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "manager user not found")
	})

	t.Run("error - negative capacity", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{
			Code:     "WH005",
			Name:     "Warehouse Five",
			Type:     "MAIN",
			Capacity: stringPtr("-100"),
		}

		_, err := service.CreateWarehouse(ctx, company.ID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "capacity cannot be negative")
	})
}

func TestWarehouseService_GetWarehouse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - get existing warehouse", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{
			Code: "WH001",
			Name: "Warehouse One",
			Type: "MAIN",
		}
		created, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		warehouse, err := service.GetWarehouseByID(ctx, company.ID, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, warehouse.ID)
		assert.Equal(t, "WH001", warehouse.Code)
	})

	t.Run("error - warehouse not found", func(t *testing.T) {
		_, err := service.GetWarehouseByID(ctx, company.ID, "nonexistent-id")

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("error - multi-company isolation", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{
			Code: "WH002",
			Name: "Warehouse Two",
			Type: "MAIN",
		}
		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		company2 := testutil.CreateTestCompany(t, db, "tenant1", "COMP2")

		_, err = service.GetWarehouseByID(ctx, company2.ID, warehouse.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

func TestWarehouseService_ListWarehouses(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	// Create test warehouses
	warehouses := []dto.CreateWarehouseRequest{
		{Code: "WH001", Name: "Warehouse Alpha", Type: "MAIN", City: stringPtr("Jakarta")},
		{Code: "WH002", Name: "Warehouse Beta", Type: "BRANCH", City: stringPtr("Bandung")},
		{Code: "WH003", Name: "Warehouse Gamma", Type: "MAIN", City: stringPtr("Jakarta")},
		{Code: "WH004", Name: "Warehouse Delta", Type: "TRANSIT", City: stringPtr("Surabaya")},
	}

	for _, req := range warehouses {
		_, err := service.CreateWarehouse(ctx, company.ID, &req)
		require.NoError(t, err)
	}

	t.Run("success - list all warehouses", func(t *testing.T) {
		query := &dto.WarehouseListQuery{
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Warehouses, 4)
	})

	t.Run("success - pagination", func(t *testing.T) {
		query := &dto.WarehouseListQuery{
			Page:     1,
			PageSize: 2,
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Warehouses, 2)
		assert.Equal(t, 2, result.TotalPages)
	})

	t.Run("success - filter by search", func(t *testing.T) {
		query := &dto.WarehouseListQuery{
			Search:   "Alpha",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "Warehouse Alpha", result.Warehouses[0].Name)
	})

	t.Run("success - filter by type", func(t *testing.T) {
		warehouseType := "MAIN"
		query := &dto.WarehouseListQuery{
			Type:     &warehouseType,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("success - filter by city", func(t *testing.T) {
		city := "Jakarta"
		query := &dto.WarehouseListQuery{
			City:     &city,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("success - sort by name", func(t *testing.T) {
		query := &dto.WarehouseListQuery{
			Page:      1,
			PageSize:  10,
			SortBy:    "name",
			SortOrder: "asc",
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, "Warehouse Alpha", result.Warehouses[0].Name)
		assert.Equal(t, "Warehouse Beta", result.Warehouses[1].Name)
	})

	t.Run("success - filter by active status", func(t *testing.T) {
		// Deactivate one warehouse
		var warehouse models.Warehouse
		db.Where("code = ?", "WH001").First(&warehouse)
		warehouse.IsActive = false
		db.Save(&warehouse)

		isActive := true
		query := &dto.WarehouseListQuery{
			IsActive: &isActive,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouses(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(3), result.TotalCount)
	})
}

func TestWarehouseService_UpdateWarehouse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - update warehouse fields", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{
			Code: "WH001",
			Name: "Original Name",
			Type: "MAIN",
		}
		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateWarehouseRequest{
			Name:    stringPtr("Updated Name"),
			Phone:   stringPtr("08123456789"),
			Email:   stringPtr("updated@example.com"),
			Address: stringPtr("New Address"),
		}

		updated, err := service.UpdateWarehouse(ctx, company.ID, warehouse.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "08123456789", *updated.Phone)
		assert.Equal(t, "updated@example.com", *updated.Email)
		assert.Equal(t, "New Address", *updated.Address)
	})

	t.Run("success - update warehouse code", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{
			Code: "WH002",
			Name: "Warehouse Two",
			Type: "MAIN",
		}
		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateWarehouseRequest{
			Code: stringPtr("WH002-NEW"),
		}

		updated, err := service.UpdateWarehouse(ctx, company.ID, warehouse.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "WH002-NEW", updated.Code)
	})

	t.Run("error - update to duplicate code", func(t *testing.T) {
		req1 := &dto.CreateWarehouseRequest{Code: "WH003", Name: "Warehouse Three", Type: "MAIN"}
		_, err := service.CreateWarehouse(ctx, company.ID, req1)
		require.NoError(t, err)

		req2 := &dto.CreateWarehouseRequest{Code: "WH004", Name: "Warehouse Four", Type: "MAIN"}
		warehouse2, err := service.CreateWarehouse(ctx, company.ID, req2)
		require.NoError(t, err)

		updateReq := &dto.UpdateWarehouseRequest{
			Code: stringPtr("WH003"), // Duplicate
		}

		_, err = service.UpdateWarehouse(ctx, company.ID, warehouse2.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "warehouse code already exists")
	})

	t.Run("error - invalid manager ID", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{Code: "WH005", Name: "Warehouse Five", Type: "MAIN"}
		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		invalidManagerID := "nonexistent-user-id"
		updateReq := &dto.UpdateWarehouseRequest{
			ManagerID: &invalidManagerID,
		}

		_, err = service.UpdateWarehouse(ctx, company.ID, warehouse.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "manager user not found")
	})
}

func TestWarehouseService_DeleteWarehouse(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - delete warehouse with zero stock", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{Code: "WH001", Name: "Warehouse One", Type: "MAIN"}
		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		err = service.DeleteWarehouse(ctx, company.ID, warehouse.ID)

		require.NoError(t, err)

		// Verify soft delete
		var deleted models.Warehouse
		db.Unscoped().Where("id = ?", warehouse.ID).First(&deleted)
		assert.False(t, deleted.IsActive)
	})

	t.Run("error - delete warehouse with stock", func(t *testing.T) {
		req := &dto.CreateWarehouseRequest{Code: "WH002", Name: "Warehouse Two", Type: "MAIN"}
		warehouse, err := service.CreateWarehouse(ctx, company.ID, req)
		require.NoError(t, err)

		// Add stock
		stock := &models.WarehouseStock{
			WarehouseID:  warehouse.ID,
			ProductID:    "test-product-id",
			Quantity:     decimal.NewFromInt(100),
			MinimumStock: decimal.Zero,
			MaximumStock: decimal.NewFromInt(1000),
		}
		db.Create(stock)

		err = service.DeleteWarehouse(ctx, company.ID, warehouse.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete warehouse with stock")
	})

	t.Run("error - warehouse not found", func(t *testing.T) {
		err := service.DeleteWarehouse(ctx, company.ID, "nonexistent-id")

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

func TestWarehouseService_ListWarehouseStocks(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	warehouse := testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	// Create test products and stocks
	products := []struct {
		code string
		name string
		qty  int64
		min  int64
	}{
		{"PROD001", "Product Alpha", 100, 50},
		{"PROD002", "Product Beta", 10, 50},   // Low stock
		{"PROD003", "Product Gamma", 0, 10},   // Zero stock
		{"PROD004", "Product Delta", 200, 100},
	}

	for _, p := range products {
		product := &models.Product{
			CompanyID:    company.ID,
			Code:         p.code,
			Name:         p.name,
			BaseUnit:     "PCS",
			BaseCost:     decimal.NewFromInt(1000),
			BasePrice:    decimal.NewFromInt(1500),
			MinimumStock: decimal.NewFromInt(p.min),
			IsActive:     true,
		}
		db.Create(product)

		stock := &models.WarehouseStock{
			WarehouseID:  warehouse.ID,
			ProductID:    product.ID,
			Quantity:     decimal.NewFromInt(p.qty),
			MinimumStock: decimal.NewFromInt(p.min),
			MaximumStock: decimal.NewFromInt(1000),
		}
		db.Create(stock)
	}

	t.Run("success - list all stocks", func(t *testing.T) {
		query := &dto.WarehouseStockListQuery{
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouseStocks(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Stocks, 4)

		// Verify product info is included
		assert.NotEmpty(t, result.Stocks[0].ProductCode)
		assert.NotEmpty(t, result.Stocks[0].ProductName)
	})

	t.Run("success - filter by warehouse", func(t *testing.T) {
		query := &dto.WarehouseStockListQuery{
			WarehouseID: &warehouse.ID,
			Page:        1,
			PageSize:    10,
		}

		result, err := service.ListWarehouseStocks(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
	})

	t.Run("success - filter low stock", func(t *testing.T) {
		lowStock := true
		query := &dto.WarehouseStockListQuery{
			LowStock: &lowStock,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouseStocks(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount) // Only PROD002
	})

	t.Run("success - filter zero stock", func(t *testing.T) {
		zeroStock := true
		query := &dto.WarehouseStockListQuery{
			ZeroStock: &zeroStock,
			Page:      1,
			PageSize:  10,
		}

		result, err := service.ListWarehouseStocks(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount) // Only PROD003
	})

	t.Run("success - search by product code", func(t *testing.T) {
		query := &dto.WarehouseStockListQuery{
			Search:   "PROD001",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListWarehouseStocks(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "PROD001", result.Stocks[0].ProductCode)
	})
}

func TestWarehouseService_UpdateWarehouseStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewWarehouseService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")
	warehouse := testutil.CreateTestWarehouse(t, db, company.ID, "WH001")

	t.Run("success - update stock settings", func(t *testing.T) {
		// Create product and stock
		product := &models.Product{
			CompanyID:    company.ID,
			Code:         "PROD001",
			Name:         "Product One",
			BaseUnit:     "PCS",
			BaseCost:     decimal.NewFromInt(1000),
			BasePrice:    decimal.NewFromInt(1500),
			MinimumStock: decimal.NewFromInt(10),
			IsActive:     true,
		}
		db.Create(product)

		stock := &models.WarehouseStock{
			WarehouseID:  warehouse.ID,
			ProductID:    product.ID,
			Quantity:     decimal.NewFromInt(100),
			MinimumStock: decimal.NewFromInt(10),
			MaximumStock: decimal.NewFromInt(500),
		}
		db.Create(stock)

		// Update stock settings
		location := "RACK-A-01"
		updateReq := &dto.UpdateWarehouseStockRequest{
			MinimumStock: stringPtr("20"),
			MaximumStock: stringPtr("1000"),
			Location:     &location,
		}

		updated, err := service.UpdateWarehouseStock(ctx, company.ID, stock.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "20", updated.MinimumStock.String())
		assert.Equal(t, "1000", updated.MaximumStock.String())
		assert.Equal(t, "RACK-A-01", *updated.Location)
		assert.Equal(t, "100", updated.Quantity.String()) // Quantity unchanged

		// Verify Product is preloaded
		assert.NotEmpty(t, updated.Product.ID)
		assert.Equal(t, "PROD001", updated.Product.Code)
	})

	t.Run("error - negative minimum stock", func(t *testing.T) {
		product := &models.Product{
			CompanyID: company.ID, Code: "PROD002", Name: "Product Two",
			BaseUnit: "PCS", BaseCost: decimal.NewFromInt(1000),
			BasePrice: decimal.NewFromInt(1500), MinimumStock: decimal.NewFromInt(10),
			IsActive: true,
		}
		db.Create(product)

		stock := &models.WarehouseStock{
			WarehouseID: warehouse.ID, ProductID: product.ID,
			Quantity: decimal.NewFromInt(100), MinimumStock: decimal.NewFromInt(10),
			MaximumStock: decimal.NewFromInt(500),
		}
		db.Create(stock)

		updateReq := &dto.UpdateWarehouseStockRequest{
			MinimumStock: stringPtr("-10"),
		}

		_, err := service.UpdateWarehouseStock(ctx, company.ID, stock.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "minimum stock cannot be negative")
	})

	t.Run("error - stock not found", func(t *testing.T) {
		updateReq := &dto.UpdateWarehouseStockRequest{
			MinimumStock: stringPtr("10"),
		}

		_, err := service.UpdateWarehouseStock(ctx, company.ID, "nonexistent-id", updateReq)

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
