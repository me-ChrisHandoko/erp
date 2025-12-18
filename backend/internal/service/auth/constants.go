package auth

// Login failure reasons for audit trail
const (
	// Authentication failures
	FailureReasonInvalidCredentials = "INVALID_CREDENTIALS"  // User not found or email doesn't exist
	FailureReasonInvalidPassword    = "INVALID_PASSWORD"     // Password verification failed

	// Account status failures
	FailureReasonAccountInactive    = "ACCOUNT_INACTIVE"     // User account is disabled
	FailureReasonAccountLocked      = "ACCOUNT_LOCKED"       // Account locked due to brute force protection

	// Subscription/tenant failures
	FailureReasonTenantInactive     = "TENANT_INACTIVE"      // Tenant is not active
	FailureReasonTrialExpired       = "TRIAL_EXPIRED"        // Trial period has ended
	FailureReasonSubscriptionExpired = "SUBSCRIPTION_EXPIRED" // Subscription has expired

	// Other failures
	FailureReasonNoTenantAccess     = "NO_TENANT_ACCESS"     // User has no active tenant access
)
