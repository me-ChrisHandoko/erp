package supplier

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

func TestSupplierService_CreateSupplier(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewSupplierService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - create supplier with minimal fields", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code: "SUPP001",
			Name: "Supplier One",
		}

		supplier, err := service.CreateSupplier(ctx, company.ID, req)

		require.NoError(t, err)
		assert.NotEmpty(t, supplier.ID)
		assert.Equal(t, "SUPP001", supplier.Code)
		assert.Equal(t, "Supplier One", supplier.Name)
		assert.True(t, supplier.IsActive)
		assert.Equal(t, "0", supplier.CreditLimit.String())
		assert.Equal(t, "0", supplier.CurrentOutstanding.String())
		assert.Equal(t, "0", supplier.OverdueAmount.String())
	})

	t.Run("success - create supplier with all fields", func(t *testing.T) {
		supplierType := "MANUFACTURER"
		req := &dto.CreateSupplierRequest{
			Code:        "SUPP002",
			Name:        "Supplier Two",
			Type:        &supplierType,
			Phone:       stringPtr("08123456789"),
			Email:       stringPtr("supplier2@example.com"),
			Address:     stringPtr("Jl. Test No. 123"),
			City:        stringPtr("Jakarta"),
			Province:    stringPtr("DKI Jakarta"),
			PostalCode:  stringPtr("12345"),
			NPWP:        stringPtr("01.234.567.8-901.000"),
			IsPKP:       boolPtr(true),
			PaymentTerm: intPtr(45),
			CreditLimit: stringPtr("20000000"),
		}

		supplier, err := service.CreateSupplier(ctx, company.ID, req)

		require.NoError(t, err)
		assert.Equal(t, "SUPP002", supplier.Code)
		require.NotNil(t, supplier.Type)
		assert.Equal(t, "MANUFACTURER", *supplier.Type)
		assert.NotNil(t, supplier.Phone)
		assert.Equal(t, "08123456789", *supplier.Phone)
		assert.True(t, supplier.IsPKP)
		assert.Equal(t, 45, supplier.PaymentTerm)
		assert.Equal(t, "20000000", supplier.CreditLimit.String())
	})

	t.Run("error - duplicate supplier code", func(t *testing.T) {
		req1 := &dto.CreateSupplierRequest{
			Code: "SUPP003",
			Name: "Supplier Three",
		}
		_, err := service.CreateSupplier(ctx, company.ID, req1)
		require.NoError(t, err)

		req2 := &dto.CreateSupplierRequest{
			Code: "SUPP003", // Duplicate
			Name: "Supplier Three Duplicate",
		}
		_, err = service.CreateSupplier(ctx, company.ID, req2)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "supplier code already exists")
	})

	t.Run("error - negative credit limit", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code:        "SUPP004",
			Name:        "Supplier Four",
			CreditLimit: stringPtr("-1000"),
		}

		_, err := service.CreateSupplier(ctx, company.ID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "credit limit cannot be negative")
	})

	t.Run("error - negative payment term", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code:        "SUPP005",
			Name:        "Supplier Five",
			PaymentTerm: intPtr(-10),
		}

		_, err := service.CreateSupplier(ctx, company.ID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "payment term cannot be negative")
	})
}

func TestSupplierService_GetSupplier(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewSupplierService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - get existing supplier", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code: "SUPP001",
			Name: "Supplier One",
		}
		created, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		supplier, err := service.GetSupplierByID(ctx, company.ID, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, supplier.ID)
		assert.Equal(t, "SUPP001", supplier.Code)
	})

	t.Run("error - supplier not found", func(t *testing.T) {
		_, err := service.GetSupplierByID(ctx, company.ID, "nonexistent-id")

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("error - multi-company isolation", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code: "SUPP002",
			Name: "Supplier Two",
		}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		company2 := testutil.CreateTestCompany(t, db, "tenant1", "COMP2")

		_, err = service.GetSupplierByID(ctx, company2.ID, supplier.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

func TestSupplierService_ListSuppliers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewSupplierService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	// Create test suppliers
	supplierType1 := "MANUFACTURER"
	supplierType2 := "DISTRIBUTOR"
	suppliers := []dto.CreateSupplierRequest{
		{Code: "SUPP001", Name: "Supplier Alpha", Type: &supplierType1, City: stringPtr("Jakarta")},
		{Code: "SUPP002", Name: "Supplier Beta", Type: &supplierType2, City: stringPtr("Bandung")},
		{Code: "SUPP003", Name: "Supplier Gamma", Type: &supplierType1, City: stringPtr("Jakarta")},
		{Code: "SUPP004", Name: "Supplier Delta", Type: &supplierType2, City: stringPtr("Surabaya")},
	}

	for _, req := range suppliers {
		_, err := service.CreateSupplier(ctx, company.ID, &req)
		require.NoError(t, err)
	}

	t.Run("success - list all suppliers", func(t *testing.T) {
		query := &dto.SupplierListQuery{
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Suppliers, 4)
	})

	t.Run("success - pagination", func(t *testing.T) {
		query := &dto.SupplierListQuery{
			Page:     1,
			PageSize: 2,
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Suppliers, 2)
		assert.Equal(t, 2, result.TotalPages)
	})

	t.Run("success - filter by search", func(t *testing.T) {
		query := &dto.SupplierListQuery{
			Search:   "Alpha",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "Supplier Alpha", result.Suppliers[0].Name)
	})

	t.Run("success - filter by type", func(t *testing.T) {
		supplierType := "MANUFACTURER"
		query := &dto.SupplierListQuery{
			Type:     &supplierType,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("success - filter by city", func(t *testing.T) {
		city := "Jakarta"
		query := &dto.SupplierListQuery{
			City:     &city,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("success - sort by name", func(t *testing.T) {
		query := &dto.SupplierListQuery{
			Page:      1,
			PageSize:  10,
			SortBy:    "name",
			SortOrder: "asc",
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, "Supplier Alpha", result.Suppliers[0].Name)
		assert.Equal(t, "Supplier Beta", result.Suppliers[1].Name)
	})

	t.Run("success - filter by active status", func(t *testing.T) {
		// Deactivate one supplier
		var supplier models.Supplier
		db.Where("code = ?", "SUPP001").First(&supplier)
		supplier.IsActive = false
		db.Save(&supplier)

		isActive := true
		query := &dto.SupplierListQuery{
			IsActive: &isActive,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListSuppliers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(3), result.TotalCount)
	})
}

func TestSupplierService_UpdateSupplier(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewSupplierService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - update supplier fields", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code: "SUPP001",
			Name: "Original Name",
		}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateSupplierRequest{
			Name:  stringPtr("Updated Name"),
			Phone: stringPtr("08123456789"),
			Email: stringPtr("updated@example.com"),
		}

		updated, err := service.UpdateSupplier(ctx, company.ID, supplier.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "08123456789", *updated.Phone)
		assert.Equal(t, "updated@example.com", *updated.Email)
	})

	t.Run("success - update supplier code", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{
			Code: "SUPP002",
			Name: "Supplier Two",
		}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateSupplierRequest{
			Code: stringPtr("SUPP002-NEW"),
		}

		updated, err := service.UpdateSupplier(ctx, company.ID, supplier.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "SUPP002-NEW", updated.Code)
	})

	t.Run("error - update to duplicate code", func(t *testing.T) {
		req1 := &dto.CreateSupplierRequest{Code: "SUPP003", Name: "Supplier Three"}
		_, err := service.CreateSupplier(ctx, company.ID, req1)
		require.NoError(t, err)

		req2 := &dto.CreateSupplierRequest{Code: "SUPP004", Name: "Supplier Four"}
		supplier2, err := service.CreateSupplier(ctx, company.ID, req2)
		require.NoError(t, err)

		updateReq := &dto.UpdateSupplierRequest{
			Code: stringPtr("SUPP003"), // Duplicate
		}

		_, err = service.UpdateSupplier(ctx, company.ID, supplier2.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "supplier code already exists")
	})

	t.Run("error - negative credit limit", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{Code: "SUPP005", Name: "Supplier Five"}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateSupplierRequest{
			CreditLimit: stringPtr("-1000"),
		}

		_, err = service.UpdateSupplier(ctx, company.ID, supplier.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "credit limit cannot be negative")
	})
}

func TestSupplierService_DeleteSupplier(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewSupplierService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - delete supplier with zero outstanding", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{Code: "SUPP001", Name: "Supplier One"}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		err = service.DeleteSupplier(ctx, company.ID, supplier.ID)

		require.NoError(t, err)

		// Verify soft delete
		var deleted models.Supplier
		db.Unscoped().Where("id = ?", supplier.ID).First(&deleted)
		assert.False(t, deleted.IsActive)
	})

	t.Run("error - delete supplier with outstanding balance", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{Code: "SUPP002", Name: "Supplier Two"}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		supplier.CurrentOutstanding = decimal.NewFromInt(1000)
		db.Save(supplier)

		err = service.DeleteSupplier(ctx, company.ID, supplier.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete supplier with outstanding balance")
	})

	t.Run("error - delete supplier with overdue amount", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{Code: "SUPP003", Name: "Supplier Three"}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		supplier.OverdueAmount = decimal.NewFromInt(500)
		db.Save(supplier)

		err = service.DeleteSupplier(ctx, company.ID, supplier.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete supplier with overdue amount")
	})

	t.Run("error - delete supplier linked to products", func(t *testing.T) {
		req := &dto.CreateSupplierRequest{Code: "SUPP004", Name: "Supplier Four"}
		supplier, err := service.CreateSupplier(ctx, company.ID, req)
		require.NoError(t, err)

		// Create a product link
		productSupplier := &models.ProductSupplier{
			ProductID:  "test-product-id",
			SupplierID: supplier.ID,
		}
		db.Create(productSupplier)

		err = service.DeleteSupplier(ctx, company.ID, supplier.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete supplier with")
		assert.Contains(t, appErr.Message, "linked product")
	})

	t.Run("error - supplier not found", func(t *testing.T) {
		err := service.DeleteSupplier(ctx, company.ID, "nonexistent-id")

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

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
