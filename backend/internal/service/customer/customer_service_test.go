package customer

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

func TestCustomerService_CreateCustomer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewCustomerService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - create customer with minimal fields", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{
			Code: "CUST001",
			Name: "Customer One",
		}

		customer, err := service.CreateCustomer(ctx, company.ID, req)

		require.NoError(t, err)
		assert.NotEmpty(t, customer.ID)
		assert.Equal(t, "CUST001", customer.Code)
		assert.Equal(t, "Customer One", customer.Name)
		assert.True(t, customer.IsActive)
		assert.Equal(t, "0", customer.CreditLimit.String())
		assert.Equal(t, "0", customer.CurrentOutstanding.String())
		assert.Equal(t, "0", customer.OverdueAmount.String())
	})

	t.Run("success - create customer with all fields", func(t *testing.T) {
		customerType := "WHOLESALE"
		req := &dto.CreateCustomerRequest{
			Code:        "CUST002",
			Name:        "Customer Two",
			Type:        &customerType,
			Phone:       stringPtr("08123456789"),
			Email:       stringPtr("customer2@example.com"),
			Address:     stringPtr("Jl. Test No. 123"),
			City:        stringPtr("Jakarta"),
			Province:    stringPtr("DKI Jakarta"),
			PostalCode:  stringPtr("12345"),
			NPWP:        stringPtr("01.234.567.8-901.000"),
			IsPKP:       boolPtr(true),
			PaymentTerm: intPtr(30),
			CreditLimit: stringPtr("10000000"),
		}

		customer, err := service.CreateCustomer(ctx, company.ID, req)

		require.NoError(t, err)
		assert.Equal(t, "CUST002", customer.Code)
		require.NotNil(t, customer.Type)
		assert.Equal(t, "WHOLESALE", *customer.Type)
		assert.NotNil(t, customer.Phone)
		assert.Equal(t, "08123456789", *customer.Phone)
		assert.NotNil(t, customer.Email)
		assert.Equal(t, "customer2@example.com", *customer.Email)
		assert.True(t, customer.IsPKP)
		assert.Equal(t, 30, customer.PaymentTerm)
		assert.Equal(t, "10000000", customer.CreditLimit.String())
	})

	t.Run("error - duplicate customer code", func(t *testing.T) {
		// Create first customer
		req1 := &dto.CreateCustomerRequest{
			Code: "CUST003",
			Name: "Customer Three",
		}
		_, err := service.CreateCustomer(ctx, company.ID, req1)
		require.NoError(t, err)

		// Try to create duplicate
		req2 := &dto.CreateCustomerRequest{
			Code: "CUST003", // Duplicate
			Name: "Customer Three Duplicate",
		}
		_, err = service.CreateCustomer(ctx, company.ID, req2)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Contains(t, appErr.Message, "customer code already exists")
	})

	t.Run("error - negative credit limit", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{
			Code:        "CUST004",
			Name:        "Customer Four",
			CreditLimit: stringPtr("-1000"),
		}

		_, err := service.CreateCustomer(ctx, company.ID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "credit limit cannot be negative")
	})

	t.Run("error - negative payment term", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{
			Code:        "CUST005",
			Name:        "Customer Five",
			PaymentTerm: intPtr(-10),
		}

		_, err := service.CreateCustomer(ctx, company.ID, req)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "payment term cannot be negative")
	})
}

func TestCustomerService_GetCustomer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewCustomerService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - get existing customer", func(t *testing.T) {
		// Create customer
		req := &dto.CreateCustomerRequest{
			Code: "CUST001",
			Name: "Customer One",
		}
		created, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		// Get customer
		customer, err := service.GetCustomerByID(ctx, company.ID, created.ID)

		require.NoError(t, err)
		assert.Equal(t, created.ID, customer.ID)
		assert.Equal(t, "CUST001", customer.Code)
	})

	t.Run("error - customer not found", func(t *testing.T) {
		_, err := service.GetCustomerByID(ctx, company.ID, "nonexistent-id")

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("error - multi-company isolation", func(t *testing.T) {
		// Create customer in company 1
		req := &dto.CreateCustomerRequest{
			Code: "CUST002",
			Name: "Customer Two",
		}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		// Create company 2
		company2 := testutil.CreateTestCompany(t, db, "tenant1", "COMP2")

		// Try to get customer from company 2 context
		_, err = service.GetCustomerByID(ctx, company2.ID, customer.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Equal(t, 404, appErr.StatusCode)
	})
}

func TestCustomerService_ListCustomers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewCustomerService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	// Create test customers
	customerType1 := "RETAIL"
	customerType2 := "WHOLESALE"
	customers := []dto.CreateCustomerRequest{
		{Code: "CUST001", Name: "Customer Alpha", Type: &customerType1, City: stringPtr("Jakarta")},
		{Code: "CUST002", Name: "Customer Beta", Type: &customerType2, City: stringPtr("Bandung")},
		{Code: "CUST003", Name: "Customer Gamma", Type: &customerType1, City: stringPtr("Jakarta")},
		{Code: "CUST004", Name: "Customer Delta", Type: &customerType2, City: stringPtr("Surabaya")},
	}

	for _, req := range customers {
		_, err := service.CreateCustomer(ctx, company.ID, &req)
		require.NoError(t, err)
	}

	t.Run("success - list all customers", func(t *testing.T) {
		query := &dto.CustomerListQuery{
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Customers, 4)
		assert.Equal(t, 1, result.TotalPages)
	})

	t.Run("success - pagination", func(t *testing.T) {
		query := &dto.CustomerListQuery{
			Page:     1,
			PageSize: 2,
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(4), result.TotalCount)
		assert.Len(t, result.Customers, 2)
		assert.Equal(t, 2, result.TotalPages)
	})

	t.Run("success - filter by search", func(t *testing.T) {
		query := &dto.CustomerListQuery{
			Search:   "Alpha",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "Customer Alpha", result.Customers[0].Name)
	})

	t.Run("success - filter by type", func(t *testing.T) {
		customerType := "RETAIL"
		query := &dto.CustomerListQuery{
			Type:     &customerType,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("success - filter by city", func(t *testing.T) {
		city := "Jakarta"
		query := &dto.CustomerListQuery{
			City:     &city,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("success - sort by name ascending", func(t *testing.T) {
		query := &dto.CustomerListQuery{
			Page:      1,
			PageSize:  10,
			SortBy:    "name",
			SortOrder: "asc",
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, "Customer Alpha", result.Customers[0].Name)
		assert.Equal(t, "Customer Beta", result.Customers[1].Name)
	})

	t.Run("success - filter by active status", func(t *testing.T) {
		// Deactivate one customer
		var customer models.Customer
		db.Where("code = ?", "CUST001").First(&customer)
		customer.IsActive = false
		db.Save(&customer)

		isActive := true
		query := &dto.CustomerListQuery{
			IsActive: &isActive,
			Page:     1,
			PageSize: 10,
		}

		result, err := service.ListCustomers(ctx, company.ID, query)

		require.NoError(t, err)
		assert.Equal(t, int64(3), result.TotalCount)
	})
}

func TestCustomerService_UpdateCustomer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewCustomerService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - update customer fields", func(t *testing.T) {
		// Create customer
		req := &dto.CreateCustomerRequest{
			Code: "CUST001",
			Name: "Original Name",
		}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		// Update customer
		updateReq := &dto.UpdateCustomerRequest{
			Name:  stringPtr("Updated Name"),
			Phone: stringPtr("08123456789"),
			Email: stringPtr("updated@example.com"),
		}

		updated, err := service.UpdateCustomer(ctx, company.ID, customer.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "08123456789", *updated.Phone)
		assert.Equal(t, "updated@example.com", *updated.Email)
	})

	t.Run("success - update customer code", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{
			Code: "CUST002",
			Name: "Customer Two",
		}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateCustomerRequest{
			Code: stringPtr("CUST002-NEW"),
		}

		updated, err := service.UpdateCustomer(ctx, company.ID, customer.ID, updateReq)

		require.NoError(t, err)
		assert.Equal(t, "CUST002-NEW", updated.Code)
	})

	t.Run("error - update to duplicate code", func(t *testing.T) {
		// Create two customers
		req1 := &dto.CreateCustomerRequest{Code: "CUST003", Name: "Customer Three"}
		_, err := service.CreateCustomer(ctx, company.ID, req1)
		require.NoError(t, err)

		req2 := &dto.CreateCustomerRequest{Code: "CUST004", Name: "Customer Four"}
		customer2, err := service.CreateCustomer(ctx, company.ID, req2)
		require.NoError(t, err)

		// Try to update customer2 to have same code as customer1
		updateReq := &dto.UpdateCustomerRequest{
			Code: stringPtr("CUST003"), // Duplicate
		}

		_, err = service.UpdateCustomer(ctx, company.ID, customer2.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "customer code already exists")
	})

	t.Run("error - negative credit limit", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{Code: "CUST005", Name: "Customer Five"}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		updateReq := &dto.UpdateCustomerRequest{
			CreditLimit: stringPtr("-1000"),
		}

		_, err = service.UpdateCustomer(ctx, company.ID, customer.ID, updateReq)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "credit limit cannot be negative")
	})
}

func TestCustomerService_DeleteCustomer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	service := NewCustomerService(db)
	ctx := context.Background()

	company := testutil.CreateTestCompany(t, db, "tenant1", "COMP1")

	t.Run("success - delete customer with zero outstanding", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{Code: "CUST001", Name: "Customer One"}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		err = service.DeleteCustomer(ctx, company.ID, customer.ID)

		require.NoError(t, err)

		// Verify soft delete
		var deleted models.Customer
		db.Unscoped().Where("id = ?", customer.ID).First(&deleted)
		assert.False(t, deleted.IsActive)
	})

	t.Run("error - delete customer with outstanding balance", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{Code: "CUST002", Name: "Customer Two"}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		// Set outstanding balance
		customer.CurrentOutstanding = decimal.NewFromInt(1000)
		db.Save(customer)

		err = service.DeleteCustomer(ctx, company.ID, customer.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete customer with outstanding balance")
	})

	t.Run("error - delete customer with overdue amount", func(t *testing.T) {
		req := &dto.CreateCustomerRequest{Code: "CUST003", Name: "Customer Three"}
		customer, err := service.CreateCustomer(ctx, company.ID, req)
		require.NoError(t, err)

		// Set overdue amount
		customer.OverdueAmount = decimal.NewFromInt(500)
		db.Save(customer)

		err = service.DeleteCustomer(ctx, company.ID, customer.ID)

		assert.Error(t, err)
		appErr, ok := err.(*pkgerrors.AppError)
		assert.True(t, ok)
		assert.Contains(t, appErr.Message, "cannot delete customer with overdue amount")
	})

	t.Run("error - customer not found", func(t *testing.T) {
		err := service.DeleteCustomer(ctx, company.ID, "nonexistent-id")

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
