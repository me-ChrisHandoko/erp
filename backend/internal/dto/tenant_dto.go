package dto

import "time"

// GetTenantDetailsResponse represents tenant details response
// Reference: 01-TENANT-COMPANY-SETUP.md lines 744-771
type GetTenantDetailsResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	TrialEndsAt *time.Time             `json:"trialEndsAt,omitempty"`
	Subscription *SubscriptionInfo     `json:"subscription,omitempty"`
	Company     *CompanyBasicInfo      `json:"company"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// SubscriptionInfo represents subscription information in tenant response
type SubscriptionInfo struct {
	ID                 string     `json:"id"`
	Price              string     `json:"price"` // Decimal as string to preserve precision
	BillingCycle       string     `json:"billingCycle"`
	Status             string     `json:"status"`
	CurrentPeriodStart time.Time  `json:"currentPeriodStart"`
	CurrentPeriodEnd   time.Time  `json:"currentPeriodEnd"`
	NextBillingDate    time.Time  `json:"nextBillingDate"`
	PaymentMethod      *string    `json:"paymentMethod,omitempty"`
	LastPaymentDate    *time.Time `json:"lastPaymentDate,omitempty"`
	LastPaymentAmount  *string    `json:"lastPaymentAmount,omitempty"` // Decimal as string
	AutoRenew          bool       `json:"autoRenew"`
}

// CompanyBasicInfo represents basic company information
type CompanyBasicInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListTenantUsersResponse represents list of tenant users
// Reference: 01-TENANT-COMPANY-SETUP.md lines 784-820
type ListTenantUsersResponse struct {
	Users []TenantUserInfo `json:"data"`
}

// TenantUserInfo represents detailed user-tenant relationship info
// Flat structure for easier frontend consumption
type TenantUserInfo struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenantId"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
}

// UserBasicInfo represents basic user information
type UserBasicInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone,omitempty"`
	IsActive bool   `json:"isActive"`
}

// InviteUserRequest represents user invitation request
// Reference: 01-TENANT-COMPANY-SETUP.md lines 834-840
type InviteUserRequest struct {
	Email string `json:"email" binding:"required,email,max=255" validate:"required,email,max=255"`
	Name  string `json:"name" binding:"required,min=1,max=255" validate:"required,min=1,max=255"`
	Role  string `json:"role" binding:"required,oneof=ADMIN FINANCE SALES WAREHOUSE STAFF" validate:"required,oneof=ADMIN FINANCE SALES WAREHOUSE STAFF"`
}

// InviteUserResponse represents user invitation response
// Reference: 01-TENANT-COMPANY-SETUP.md lines 842-862
// Flat structure for easier frontend consumption
type InviteUserResponse struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenantId"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Role            string     `json:"role"`
	IsActive        bool       `json:"isActive"`
	InvitationToken *string    `json:"invitationToken,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
}

// UpdateUserRoleRequest represents role update request
// Reference: 01-TENANT-COMPANY-SETUP.md lines 884-886
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=ADMIN FINANCE SALES WAREHOUSE STAFF" validate:"required,oneof=ADMIN FINANCE SALES WAREHOUSE STAFF"`
}

// UpdateUserRoleResponse represents role update response
// Reference: 01-TENANT-COMPANY-SETUP.md lines 888-899
// Flat structure for easier frontend consumption
type UpdateUserRoleResponse struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenantId"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
}

// RemoveUserResponse represents user removal response
// Reference: 01-TENANT-COMPANY-SETUP.md lines 918-922
type RemoveUserResponse struct {
	Message string `json:"message"`
}
