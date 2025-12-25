package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"backend/models"
)

// setupTestDB creates in-memory SQLite database for testing
func setupSubscriptionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.Tenant{},
		&models.Subscription{},
		&models.Company{},
	)
	require.NoError(t, err)

	return db
}

// createTestTenant helper function
func createTestTenant(t *testing.T, db *gorm.DB, status models.TenantStatus, trialEndsAt *time.Time) models.Tenant {
	// Create company first
	company := models.Company{
		Name:      "Test Company",
		LegalName: "Test Company PT",
		Address:   "Test Address",
		City:      "Jakarta",
		Province:  "DKI Jakarta",
		Phone:     "081234567890",
		Email:     "test@example.com",
	}
	err := db.Create(&company).Error
	require.NoError(t, err)

	tenant := models.Tenant{
		CompanyID:   company.ID,
		Status:      status,
		TrialEndsAt: trialEndsAt,
	}
	err = db.Create(&tenant).Error
	require.NoError(t, err)

	return tenant
}

// createTestSubscription helper function
func createTestSubscription(t *testing.T, db *gorm.DB, tenantID string, status models.SubscriptionStatus, gracePeriodEnds *time.Time) models.Subscription {
	now := time.Now()
	subscription := models.Subscription{
		Status:             status,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		NextBillingDate:    now.AddDate(0, 1, 0),
		GracePeriodEnds:    gracePeriodEnds,
	}
	err := db.Create(&subscription).Error
	require.NoError(t, err)

	// Update tenant with subscription
	err = db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("subscription_id", subscription.ID).Error
	require.NoError(t, err)

	return subscription
}

func TestValidateSubscriptionMiddleware_ActiveTenant(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusActive, nil)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateSubscriptionMiddleware_TrialValid(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	trialEndsAt := time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	tenant := createTestTenant(t, db, models.TenantStatusTrial, &trialEndsAt)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateSubscriptionMiddleware_TrialExpired(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	trialEndsAt := time.Now().Add(-1 * time.Hour) // 1 hour ago (expired)
	tenant := createTestTenant(t, db, models.TenantStatusTrial, &trialEndsAt)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "TRIAL_EXPIRED")
	assert.Contains(t, w.Body.String(), "Trial period expired")
}

func TestValidateSubscriptionMiddleware_ExpiredTenant(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusExpired, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "SUBSCRIPTION_EXPIRED")
	assert.Contains(t, w.Body.String(), "Subscription expired")
}

func TestValidateSubscriptionMiddleware_SuspendedTenant(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusSuspended, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ACCOUNT_SUSPENDED")
	assert.Contains(t, w.Body.String(), "Account suspended")
}

func TestValidateSubscriptionMiddleware_CancelledTenant(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusCancelled, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "SUBSCRIPTION_CANCELLED")
}

func TestValidateSubscriptionMiddleware_PastDueWithinGracePeriod(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusSuspended, nil)
	gracePeriodEnds := time.Now().Add(3 * 24 * time.Hour) // 3 days from now
	createTestSubscription(t, db, tenant.ID, models.SubscriptionStatusPastDue, &gracePeriodEnds)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should allow access but with warning header
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("X-Subscription-Warning"), "Payment overdue")
}

func TestValidateSubscriptionMiddleware_PastDueGracePeriodExpired(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusSuspended, nil)
	gracePeriodEnds := time.Now().Add(-1 * time.Hour) // 1 hour ago (expired)
	createTestSubscription(t, db, tenant.ID, models.SubscriptionStatusPastDue, &gracePeriodEnds)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "PAYMENT_OVERDUE")
	assert.Contains(t, w.Body.String(), "Payment overdue")
}

func TestValidateSubscriptionMiddleware_NoTenantContext(t *testing.T) {
	db := setupSubscriptionTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	// No tenant_id set in context
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Tenant context required")
}

func TestValidateSubscriptionMiddleware_TenantNotFound(t *testing.T) {
	db := setupSubscriptionTestDB(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", "non-existent-tenant-id")
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Tenant not found")
}

func TestValidateSubscriptionMiddleware_ActiveWithExpiredSubscription(t *testing.T) {
	db := setupSubscriptionTestDB(t)
	tenant := createTestTenant(t, db, models.TenantStatusActive, nil)
	createTestSubscription(t, db, tenant.ID, models.SubscriptionStatusExpired, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", tenant.ID)
		c.Next()
	})
	router.Use(ValidateSubscriptionMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should block access even if tenant status is ACTIVE but subscription is EXPIRED
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "SUBSCRIPTION_EXPIRED")
}
