# Module Group 1: Foundation Setup (Tenant & Company Management)

**Implementation Priority:** WEEK 1 (Foundation)
**Dependencies:** Authentication module ✅ (Complete)
**Modules:** Company Profile Management, Tenant Management

---

## Overview

This module group establishes the **foundational configuration** required for all subsequent business modules. Every tenant must complete company profile setup before conducting transactions.

### Business Context
- **Company Profile:** Legal entity details, tax compliance (NPWP, PKP), bank accounts, invoice numbering formats
- **Tenant Management:** Multi-tenant access control, subscription management, user-tenant role assignments

### Why First?
1. Company settings (NPWP, PPN rate, invoice formats) required for invoices
2. Bank accounts needed for payment recording
3. Tenant-user roles required for access control
4. Configuration validates before allowing transactions

---

## Module 1: Company Profile Management

### Purpose
Manage legal entity profile, Indonesian tax compliance settings, and system configuration for each tenant.

### Database Models (Already Defined)

**Primary Models:**
```go
// models/company.go
type Company struct {
    ID                  string          // Primary key (CUID)

    // Legal Entity
    Name                string          // Display name
    LegalName           string          // Legal entity name
    EntityType          string          // CV, PT, UD, Firma

    // Address
    Address             string
    City                string
    Province            string
    PostalCode          *string
    Country             string          // Default: "Indonesia"

    // Contact
    Phone               string
    Email               string
    Website             *string

    // Indonesian Tax Compliance
    NPWP                *string         // Tax ID (XX.XXX.XXX.X-XXX.XXX)
    IsPKP               bool            // Taxable entity status
    PPNRate             decimal.Decimal // PPN rate (11% in 2025)
    FakturPajakSeries   *string         // Tax invoice series
    SPPKPNumber         *string         // PKP registration number

    // Branding
    LogoURL             *string
    PrimaryColor        *string         // Default: #1E40AF
    SecondaryColor      *string         // Default: #64748B

    // Invoice Settings
    InvoicePrefix       string          // Default: "INV"
    InvoiceNumberFormat string          // {PREFIX}/{NUMBER}/{MONTH}/{YEAR}
    InvoiceFooter       *string
    InvoiceTerms        *string

    // Sales Order Settings
    SOPrefix            string          // Default: "SO"
    SONumberFormat      string          // {PREFIX}{NUMBER}

    // Purchase Order Settings
    POPrefix            string          // Default: "PO"
    PONumberFormat      string          // {PREFIX}{NUMBER}

    // System Settings
    Currency            string          // Default: "IDR"
    Timezone            string          // Default: "Asia/Jakarta"
    Locale              string          // Default: "id-ID"

    // Business Hours
    BusinessHoursStart  *string         // Default: "08:00"
    BusinessHoursEnd    *string         // Default: "17:00"
    WorkingDays         *string         // Default: "1,2,3,4,5" (Mon-Fri)

    IsActive            bool
    CreatedAt           time.Time
    UpdatedAt           time.Time

    // Relations
    Tenant              *Tenant         // One-to-one
    Banks               []CompanyBank   // One-to-many
}

type CompanyBank struct {
    ID              string
    CompanyID       string
    BankName        string          // "BCA", "Mandiri", "BRI"
    AccountNumber   string
    AccountName     string
    BranchName      *string
    IsPrimary       bool            // Primary bank for invoices
    CheckPrefix     *string         // Prefix for check numbers
    IsActive        bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Relationship:**
- `Tenant` → `Company` (1:1 via `CompanyID`)
- `Company` → `CompanyBank` (1:many)

---

### API Endpoints

#### 1. Get Company Profile
```http
GET /api/v1/company

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}

Response (200 OK):
{
  "success": true,
  "data": {
    "id": "clxxxx",
    "name": "CV Distribusi Sembako Jaya",
    "legalName": "CV DISTRIBUSI SEMBAKO JAYA",
    "entityType": "CV",
    "address": "Jl. Pasar Induk No. 123",
    "city": "Jakarta Timur",
    "province": "DKI Jakarta",
    "postalCode": "13220",
    "country": "Indonesia",
    "phone": "+6221-8765432",
    "email": "info@sembakojaya.com",
    "website": "https://sembakojaya.com",
    "npwp": "12.345.678.9-012.345",
    "isPKP": true,
    "ppnRate": "11.00",
    "fakturPajakSeries": "010-25",
    "sppkpNumber": "PEM-12345/WPJ.07/2024",
    "logoUrl": "https://cdn.example.com/logo.png",
    "primaryColor": "#1E40AF",
    "secondaryColor": "#64748B",
    "invoicePrefix": "INV",
    "invoiceNumberFormat": "{PREFIX}/{NUMBER}/{MONTH}/{YEAR}",
    "invoiceFooter": "Terima kasih atas kepercayaan Anda",
    "invoiceTerms": "Pembayaran 30 hari setelah tanggal invoice",
    "soPrefix": "SO",
    "soNumberFormat": "{PREFIX}/{NUMBER}",
    "poPrefix": "PO",
    "poNumberFormat": "{PREFIX}/{NUMBER}",
    "currency": "IDR",
    "timezone": "Asia/Jakarta",
    "locale": "id-ID",
    "businessHoursStart": "08:00",
    "businessHoursEnd": "17:00",
    "workingDays": "1,2,3,4,5",
    "banks": [
      {
        "id": "clbank1",
        "bankName": "BCA",
        "accountNumber": "1234567890",
        "accountName": "CV DISTRIBUSI SEMBAKO JAYA",
        "branchName": "KCP Jakarta Timur",
        "isPrimary": true,
        "checkPrefix": "BCA",
        "isActive": true
      }
    ],
    "isActive": true,
    "createdAt": "2025-01-15T08:00:00Z",
    "updatedAt": "2025-01-15T08:00:00Z"
  }
}

Error (404 Not Found):
{
  "success": false,
  "error": {
    "code": "COMPANY_NOT_FOUND",
    "message": "Company profile not found for this tenant"
  }
}
```

#### 2. Create/Initialize Company Profile
```http
POST /api/v1/company

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}

Request Body:
{
  "name": "CV Distribusi Sembako Jaya",
  "legalName": "CV DISTRIBUSI SEMBAKO JAYA",
  "entityType": "CV",
  "address": "Jl. Pasar Induk No. 123",
  "city": "Jakarta Timur",
  "province": "DKI Jakarta",
  "postalCode": "13220",
  "phone": "+6221-8765432",
  "email": "info@sembakojaya.com",
  "npwp": "12.345.678.9-012.345",
  "isPKP": true,
  "ppnRate": "11.00",
  "fakturPajakSeries": "010-25",
  "invoicePrefix": "INV",
  "invoiceNumberFormat": "{PREFIX}/{NUMBER}/{MONTH}/{YEAR}",
  "currency": "IDR",
  "timezone": "Asia/Jakarta"
}

Response (201 Created):
{
  "success": true,
  "data": { /* Full company profile */ }
}

Validation Errors (400 Bad Request):
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "name",
        "message": "Company name is required"
      },
      {
        "field": "npwp",
        "message": "NPWP format invalid (must be XX.XXX.XXX.X-XXX.XXX)"
      }
    ]
  }
}
```

#### 3. Update Company Profile
```http
PUT /api/v1/company

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}

Request Body:
{
  "name": "CV Distribusi Sembako Jaya Makmur",
  "phone": "+6221-8765433",
  "ppnRate": "11.00",
  "invoiceFooter": "Terima kasih atas kepercayaan Anda. Barang yang sudah dibeli tidak dapat dikembalikan."
}

Response (200 OK):
{
  "success": true,
  "data": { /* Updated company profile */ }
}
```

#### 4. Upload Company Logo
```http
POST /api/v1/company/logo

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}
  Content-Type: multipart/form-data

Request Body (multipart):
  logo: (binary file, max 2MB, formats: jpg, png, svg)

Response (200 OK):
{
  "success": true,
  "data": {
    "logoUrl": "https://cdn.example.com/tenants/clxxxx/logo.png"
  }
}
```

#### 5. Manage Company Banks

**List Banks:**
```http
GET /api/v1/company/banks

Response (200 OK):
{
  "success": true,
  "data": [
    {
      "id": "clbank1",
      "bankName": "BCA",
      "accountNumber": "1234567890",
      "accountName": "CV DISTRIBUSI SEMBAKO JAYA",
      "branchName": "KCP Jakarta Timur",
      "isPrimary": true,
      "isActive": true
    }
  ]
}
```

**Add Bank Account:**
```http
POST /api/v1/company/banks

Request Body:
{
  "bankName": "Mandiri",
  "accountNumber": "9876543210",
  "accountName": "CV DISTRIBUSI SEMBAKO JAYA",
  "branchName": "Cab. Jakarta Pusat",
  "isPrimary": false
}

Response (201 Created):
{
  "success": true,
  "data": { /* New bank account */ }
}
```

**Update Bank Account:**
```http
PUT /api/v1/company/banks/:id

Request Body:
{
  "branchName": "Cab. Jakarta Utara",
  "isPrimary": true
}

Response (200 OK):
{
  "success": true,
  "data": { /* Updated bank account */ }
}
```

**Delete Bank Account:**
```http
DELETE /api/v1/company/banks/:id

Response (200 OK):
{
  "success": true,
  "message": "Bank account deleted successfully"
}
```

---

### Business Logic (Service Layer)

**File:** `internal/service/company/company_service.go`

```go
package company

import (
    "context"
    "errors"
    "backend/models"
    "gorm.io/gorm"
)

type CompanyService struct {
    db *gorm.DB
}

func NewCompanyService(db *gorm.DB) *CompanyService {
    return &CompanyService{db: db}
}

// GetCompanyByTenantID retrieves company profile for the current tenant
func (s *CompanyService) GetCompanyByTenantID(ctx context.Context, tenantID string) (*models.Company, error) {
    var company models.Company

    // Include related banks
    err := s.db.Preload("Banks", "is_active = ?", true).
        Joins("JOIN tenants ON tenants.company_id = companies.id").
        Where("tenants.id = ? AND companies.is_active = ?", tenantID, true).
        First(&company).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("company profile not found")
        }
        return nil, err
    }

    return &company, nil
}

// CreateCompany initializes company profile for a tenant
func (s *CompanyService) CreateCompany(ctx context.Context, tenantID string, req *CreateCompanyRequest) (*models.Company, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // Check if company already exists for this tenant
    var tenant models.Tenant
    if err := s.db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
        return nil, errors.New("tenant not found")
    }

    if tenant.CompanyID != "" {
        return nil, errors.New("company profile already exists for this tenant")
    }

    // Create company record
    company := &models.Company{
        Name:                req.Name,
        LegalName:           req.LegalName,
        EntityType:          req.EntityType,
        Address:             req.Address,
        City:                req.City,
        Province:            req.Province,
        PostalCode:          req.PostalCode,
        Country:             "Indonesia",
        Phone:               req.Phone,
        Email:               req.Email,
        Website:             req.Website,
        NPWP:                req.NPWP,
        IsPKP:               req.IsPKP,
        PPNRate:             req.PPNRate,
        FakturPajakSeries:   req.FakturPajakSeries,
        SPPKPNumber:         req.SPPKPNumber,
        InvoicePrefix:       req.InvoicePrefix,
        InvoiceNumberFormat: req.InvoiceNumberFormat,
        SOPrefix:            req.SOPrefix,
        SONumberFormat:      req.SONumberFormat,
        POPrefix:            req.POPrefix,
        PONumberFormat:      req.PONumberFormat,
        Currency:            "IDR",
        Timezone:            "Asia/Jakarta",
        Locale:              "id-ID",
        IsActive:            true,
    }

    // Start transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Create company
    if err := tx.Create(company).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // Update tenant with company_id
    if err := tx.Model(&tenant).Update("company_id", company.ID).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return nil, err
    }

    return company, nil
}

// UpdateCompany updates company profile
func (s *CompanyService) UpdateCompany(ctx context.Context, tenantID string, req *UpdateCompanyRequest) (*models.Company, error) {
    // Get existing company
    company, err := s.GetCompanyByTenantID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // Update fields
    updates := map[string]interface{}{}
    if req.Name != nil {
        updates["name"] = *req.Name
    }
    if req.LegalName != nil {
        updates["legal_name"] = *req.LegalName
    }
    if req.Phone != nil {
        updates["phone"] = *req.Phone
    }
    if req.Email != nil {
        updates["email"] = *req.Email
    }
    if req.Address != nil {
        updates["address"] = *req.Address
    }
    if req.PPNRate != nil {
        updates["ppn_rate"] = *req.PPNRate
    }
    // ... other fields

    if err := s.db.Model(company).Updates(updates).Error; err != nil {
        return nil, err
    }

    // Reload company
    return s.GetCompanyByTenantID(ctx, tenantID)
}

// AddBankAccount adds a new bank account
func (s *CompanyService) AddBankAccount(ctx context.Context, tenantID string, req *AddBankRequest) (*models.CompanyBank, error) {
    // Get company
    company, err := s.GetCompanyByTenantID(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // Validate request
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // If isPrimary = true, unset other primary banks
    if req.IsPrimary {
        s.db.Model(&models.CompanyBank{}).
            Where("company_id = ?", company.ID).
            Update("is_primary", false)
    }

    // Create bank account
    bank := &models.CompanyBank{
        CompanyID:     company.ID,
        BankName:      req.BankName,
        AccountNumber: req.AccountNumber,
        AccountName:   req.AccountName,
        BranchName:    req.BranchName,
        IsPrimary:     req.IsPrimary,
        CheckPrefix:   req.CheckPrefix,
        IsActive:      true,
    }

    if err := s.db.Create(bank).Error; err != nil {
        return nil, err
    }

    return bank, nil
}
```

---

### Validation Rules

**CreateCompanyRequest:**
```go
type CreateCompanyRequest struct {
    Name                string          `json:"name" validate:"required,min=3,max=255"`
    LegalName           string          `json:"legalName" validate:"required,min=3,max=255"`
    EntityType          string          `json:"entityType" validate:"required,oneof=CV PT UD Firma"`
    Address             string          `json:"address" validate:"required"`
    City                string          `json:"city" validate:"required"`
    Province            string          `json:"province" validate:"required"`
    PostalCode          *string         `json:"postalCode" validate:"omitempty,len=5,numeric"`
    Phone               string          `json:"phone" validate:"required,phone_number"`
    Email               string          `json:"email" validate:"required,email"`
    Website             *string         `json:"website" validate:"omitempty,url"`
    NPWP                *string         `json:"npwp" validate:"omitempty,npwp_format"`
    IsPKP               bool            `json:"isPKP"`
    PPNRate             decimal.Decimal `json:"ppnRate" validate:"required,gte=0,lte=100"`
    FakturPajakSeries   *string         `json:"fakturPajakSeries" validate:"omitempty"`
    SPPKPNumber         *string         `json:"sppkpNumber" validate:"omitempty"`
    InvoicePrefix       string          `json:"invoicePrefix" validate:"required,min=1,max=20"`
    InvoiceNumberFormat string          `json:"invoiceNumberFormat" validate:"required"`
    SOPrefix            string          `json:"soPrefix" validate:"required,min=1,max=20"`
    SONumberFormat      string          `json:"soNumberFormat" validate:"required"`
    POPrefix            string          `json:"poPrefix" validate:"required,min=1,max=20"`
    PONumberFormat      string          `json:"poNumberFormat" validate:"required"`
}

// Custom validation: NPWP format (XX.XXX.XXX.X-XXX.XXX)
func ValidateNPWP(npwp string) error {
    pattern := `^\d{2}\.\d{3}\.\d{3}\.\d-\d{3}\.\d{3}$`
    matched, _ := regexp.MatchString(pattern, npwp)
    if !matched {
        return errors.New("NPWP format must be XX.XXX.XXX.X-XXX.XXX")
    }
    return nil
}

// Custom validation: Indonesian phone number
func ValidatePhoneNumber(phone string) error {
    // Must start with +62 or 08
    pattern := `^(\+628|08)[0-9]{8,12}$`
    matched, _ := regexp.MatchString(pattern, phone)
    if !matched {
        return errors.New("Phone number must be Indonesian format (+628xxxxxxxx or 08xxxxxxxx)")
    }
    return nil
}
```

**AddBankRequest:**
```go
type AddBankRequest struct {
    BankName      string  `json:"bankName" validate:"required,min=2,max=100"`
    AccountNumber string  `json:"accountNumber" validate:"required,min=8,max=50"`
    AccountName   string  `json:"accountName" validate:"required,min=3,max=255"`
    BranchName    *string `json:"branchName" validate:"omitempty"`
    IsPrimary     bool    `json:"isPrimary"`
    CheckPrefix   *string `json:"checkPrefix" validate:"omitempty,max=20"`
}
```

---

### Business Rules

1. **One Company Per Tenant:**
   - Each tenant can have only ONE company profile
   - Company creation is one-time setup (update allowed, delete not allowed)

2. **NPWP Validation:**
   - Format: `XX.XXX.XXX.X-XXX.XXX`
   - Must be unique across all companies
   - Required if `isPKP = true`

3. **PKP Status:**
   - If `isPKP = true`, must provide:
     - NPWP
     - Faktur Pajak Series
     - SPPKP Number (optional but recommended)
   - If `isPKP = false`, PPN tax still calculated but no Faktur Pajak issued

4. **PPN Rate:**
   - Default: 11% (as of 2025)
   - Editable for different tax jurisdictions
   - Range: 0% - 100%

5. **Bank Accounts:**
   - Minimum 1 bank account required before generating invoices
   - Only ONE primary bank allowed (auto-unset others when setting new primary)
   - Primary bank shown on invoices

6. **Number Formats:**
   - Placeholders: `{PREFIX}`, `{NUMBER}`, `{MONTH}`, `{YEAR}`
   - `{NUMBER}` auto-incremented with zero-padding (e.g., 001, 002)
   - Example: `INV/001/12/2025`

---

## Module 2: Tenant Management

### Purpose
Allow OWNER/ADMIN users to manage tenant settings, view subscription status, and assign roles to users within their tenant.

### Database Models (Already Defined)

```go
// models/tenant.go
type Tenant struct {
    ID             string
    CompanyID      string          // FK to Company (one-to-one)
    SubscriptionID *string         // FK to Subscription
    Status         TenantStatus    // TRIAL, ACTIVE, SUSPENDED, PAST_DUE, EXPIRED
    TrialEndsAt    *time.Time
    Notes          *string
    CreatedAt      time.Time
    UpdatedAt      time.Time

    // Relations
    Company      Company
    Subscription *Subscription
    Users        []UserTenant    // User access to this tenant
}

// models/user.go
type UserTenant struct {
    ID        string
    UserID    string
    TenantID  string
    Role      UserRole        // OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Enums:**
```go
type TenantStatus string
const (
    TenantStatusTrial      TenantStatus = "TRIAL"
    TenantStatusActive     TenantStatus = "ACTIVE"
    TenantStatusSuspended  TenantStatus = "SUSPENDED"
    TenantStatusPastDue    TenantStatus = "PAST_DUE"
    TenantStatusExpired    TenantStatus = "EXPIRED"
)

type UserRole string
const (
    UserRoleOwner     UserRole = "OWNER"
    UserRoleAdmin     UserRole = "ADMIN"
    UserRoleFinance   UserRole = "FINANCE"
    UserRoleSales     UserRole = "SALES"
    UserRoleWarehouse UserRole = "WAREHOUSE"
    UserRoleStaff     UserRole = "STAFF"
)
```

---

### API Endpoints

#### 1. Get Tenant Details
```http
GET /api/v1/tenant

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}

Response (200 OK):
{
  "success": true,
  "data": {
    "id": "cltenant1",
    "status": "ACTIVE",
    "trialEndsAt": "2025-02-01T00:00:00Z",
    "subscription": {
      "id": "clsub1",
      "price": "300000.00",
      "billingCycle": "MONTHLY",
      "status": "ACTIVE",
      "currentPeriodStart": "2025-01-15T00:00:00Z",
      "currentPeriodEnd": "2025-02-15T00:00:00Z",
      "nextBillingDate": "2025-02-15T00:00:00Z",
      "paymentMethod": "TRANSFER",
      "lastPaymentDate": "2025-01-15T10:30:00Z",
      "lastPaymentAmount": "300000.00",
      "autoRenew": true
    },
    "company": {
      "id": "clcomp1",
      "name": "CV Distribusi Sembako Jaya"
    },
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-15T10:30:00Z"
  }
}
```

#### 2. List Tenant Users
```http
GET /api/v1/tenant/users
GET /api/v1/tenant/users?role=ADMIN&isActive=true

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}

Response (200 OK):
{
  "success": true,
  "data": [
    {
      "id": "clut1",
      "userId": "cluser1",
      "tenantId": "cltenant1",
      "role": "OWNER",
      "isActive": true,
      "user": {
        "id": "cluser1",
        "email": "owner@sembakojaya.com",
        "fullName": "John Doe",
        "phone": "+628123456789",
        "isActive": true
      },
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-01-01T00:00:00Z"
    },
    {
      "id": "clut2",
      "userId": "cluser2",
      "tenantId": "cltenant1",
      "role": "ADMIN",
      "isActive": true,
      "user": {
        "id": "cluser2",
        "email": "admin@sembakojaya.com",
        "fullName": "Jane Smith",
        "phone": "+628987654321",
        "isActive": true
      },
      "createdAt": "2025-01-02T00:00:00Z",
      "updatedAt": "2025-01-02T00:00:00Z"
    }
  ]
}
```

#### 3. Invite User to Tenant
```http
POST /api/v1/tenant/users/invite

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}

Middleware: RequireRoleMiddleware("OWNER", "ADMIN")

Request Body:
{
  "email": "sales@sembakojaya.com",
  "fullName": "Alice Johnson",
  "role": "SALES",
  "phone": "+628111222333"
}

Response (201 Created):
{
  "success": true,
  "data": {
    "id": "clut3",
    "userId": "cluser3",
    "tenantId": "cltenant1",
    "role": "SALES",
    "isActive": true,
    "user": {
      "id": "cluser3",
      "email": "sales@sembakojaya.com",
      "fullName": "Alice Johnson",
      "phone": "+628111222333",
      "isActive": true
    },
    "invitationToken": "inv_xxxxxx", // Email verification token
    "createdAt": "2025-01-16T10:00:00Z"
  },
  "message": "Invitation sent to sales@sembakojaya.com"
}

Business Logic:
1. Check if email already exists:
   - If user exists: Create UserTenant link with specified role
   - If user doesn't exist: Create User + UserTenant, send email verification
2. OWNER/ADMIN cannot invite another OWNER (only one OWNER per tenant)
3. Send invitation email with verification link
```

#### 4. Update User Role
```http
PUT /api/v1/tenant/users/:userTenantId/role

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}

Middleware: RequireRoleMiddleware("OWNER", "ADMIN")

Request Body:
{
  "role": "FINANCE"
}

Response (200 OK):
{
  "success": true,
  "data": {
    "id": "clut3",
    "userId": "cluser3",
    "tenantId": "cltenant1",
    "role": "FINANCE",
    "isActive": true,
    "updatedAt": "2025-01-16T11:00:00Z"
  }
}

Validation:
- Cannot change OWNER role
- Cannot remove last ADMIN if no OWNER
- ADMIN can change roles except OWNER
```

#### 5. Remove User from Tenant
```http
DELETE /api/v1/tenant/users/:userTenantId

Headers:
  Authorization: Bearer {access_token}
  X-Tenant-ID: {tenant_id}
  X-CSRF-Token: {csrf_token}

Middleware: RequireRoleMiddleware("OWNER", "ADMIN")

Response (200 OK):
{
  "success": true,
  "message": "User removed from tenant successfully"
}

Business Logic:
- Soft delete: Set isActive = false (preserve audit trail)
- Cannot remove OWNER
- Cannot remove last ADMIN
- User can still access other tenants if they have UserTenant links
```

---

### Business Logic (Service Layer)

**File:** `internal/service/tenant/tenant_service.go`

```go
package tenant

import (
    "context"
    "errors"
    "backend/models"
    "backend/internal/service/auth"
    "gorm.io/gorm"
)

type TenantService struct {
    db          *gorm.DB
    authService *auth.AuthService
}

func NewTenantService(db *gorm.DB, authService *auth.AuthService) *TenantService {
    return &TenantService{
        db:          db,
        authService: authService,
    }
}

// GetTenantByID retrieves tenant with company and subscription details
func (s *TenantService) GetTenantByID(ctx context.Context, tenantID string) (*models.Tenant, error) {
    var tenant models.Tenant

    err := s.db.Preload("Company").
        Preload("Subscription").
        Where("id = ?", tenantID).
        First(&tenant).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("tenant not found")
        }
        return nil, err
    }

    return &tenant, nil
}

// ListTenantUsers lists all users with access to this tenant
func (s *TenantService) ListTenantUsers(ctx context.Context, tenantID string, role *string, isActive *bool) ([]models.UserTenant, error) {
    var userTenants []models.UserTenant

    query := s.db.Preload("User").Where("tenant_id = ?", tenantID)

    if role != nil {
        query = query.Where("role = ?", *role)
    }

    if isActive != nil {
        query = query.Where("is_active = ?", *isActive)
    }

    if err := query.Find(&userTenants).Error; err != nil {
        return nil, err
    }

    return userTenants, nil
}

// InviteUserToTenant invites a user to join the tenant
func (s *TenantService) InviteUserToTenant(ctx context.Context, tenantID string, req *InviteUserRequest) (*models.UserTenant, string, error) {
    // Validate role
    if req.Role == models.UserRoleOwner {
        return nil, "", errors.New("cannot invite another OWNER")
    }

    // Check if user exists
    var user models.User
    err := s.db.Where("email = ?", req.Email).First(&user).Error

    var invitationToken string

    if errors.Is(err, gorm.ErrRecordNotFound) {
        // User doesn't exist - create new user
        user, invitationToken, err = s.authService.CreateUserWithInvitation(req.Email, req.FullName, req.Phone)
        if err != nil {
            return nil, "", err
        }
    } else if err != nil {
        return nil, "", err
    }

    // Check if user already has access to this tenant
    var existingUT models.UserTenant
    err = s.db.Where("user_id = ? AND tenant_id = ?", user.ID, tenantID).First(&existingUT).Error

    if err == nil {
        // User already has access
        if existingUT.IsActive {
            return nil, "", errors.New("user already has access to this tenant")
        } else {
            // Reactivate existing user-tenant link
            existingUT.IsActive = true
            existingUT.Role = req.Role
            s.db.Save(&existingUT)
            return &existingUT, "", nil
        }
    }

    // Create UserTenant link
    userTenant := &models.UserTenant{
        UserID:   user.ID,
        TenantID: tenantID,
        Role:     req.Role,
        IsActive: true,
    }

    if err := s.db.Create(userTenant).Error; err != nil {
        return nil, "", err
    }

    // Reload with user
    s.db.Preload("User").First(userTenant, userTenant.ID)

    return userTenant, invitationToken, nil
}

// UpdateUserRole updates the role of a user in the tenant
func (s *TenantService) UpdateUserRole(ctx context.Context, tenantID, userTenantID string, newRole models.UserRole) (*models.UserTenant, error) {
    var userTenant models.UserTenant

    // Get existing user-tenant link
    err := s.db.Where("id = ? AND tenant_id = ?", userTenantID, tenantID).First(&userTenant).Error
    if err != nil {
        return nil, errors.New("user-tenant link not found")
    }

    // Cannot change OWNER role
    if userTenant.Role == models.UserRoleOwner {
        return nil, errors.New("cannot change OWNER role")
    }

    // Cannot set new role to OWNER
    if newRole == models.UserRoleOwner {
        return nil, errors.New("cannot promote to OWNER role")
    }

    // Update role
    userTenant.Role = newRole
    if err := s.db.Save(&userTenant).Error; err != nil {
        return nil, err
    }

    return &userTenant, nil
}

// RemoveUserFromTenant removes a user from the tenant (soft delete)
func (s *TenantService) RemoveUserFromTenant(ctx context.Context, tenantID, userTenantID string) error {
    var userTenant models.UserTenant

    // Get existing user-tenant link
    err := s.db.Where("id = ? AND tenant_id = ?", userTenantID, tenantID).First(&userTenant).Error
    if err != nil {
        return errors.New("user-tenant link not found")
    }

    // Cannot remove OWNER
    if userTenant.Role == models.UserRoleOwner {
        return errors.New("cannot remove OWNER from tenant")
    }

    // Check if this is the last ADMIN
    if userTenant.Role == models.UserRoleAdmin {
        var adminCount int64
        s.db.Model(&models.UserTenant{}).
            Where("tenant_id = ? AND role = ? AND is_active = ?", tenantID, models.UserRoleAdmin, true).
            Count(&adminCount)

        if adminCount <= 1 {
            return errors.New("cannot remove last ADMIN from tenant")
        }
    }

    // Soft delete
    userTenant.IsActive = false
    return s.db.Save(&userTenant).Error
}
```

---

### Middleware: Role-Based Access Control

**File:** `internal/middleware/role.go`

```go
package middleware

import (
    "net/http"
    "backend/models"
    "backend/pkg/errors"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// RequireRoleMiddleware checks if user has one of the specified roles
func RequireRoleMiddleware(db *gorm.DB, allowedRoles ...models.UserRole) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user ID and tenant ID from context (set by JWTAuthMiddleware)
        userID, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusUnauthorized, errors.NewAuthenticationError("User not authenticated"))
            c.Abort()
            return
        }

        tenantID, exists := c.Get("tenant_id")
        if !exists {
            c.JSON(http.StatusForbidden, errors.NewAuthorizationError("Tenant context required"))
            c.Abort()
            return
        }

        // Check user's role in this tenant
        var userTenant models.UserTenant
        err := db.Where("user_id = ? AND tenant_id = ? AND is_active = ?",
            userID.(string), tenantID.(string), true).
            First(&userTenant).Error

        if err != nil {
            c.JSON(http.StatusForbidden, errors.NewAuthorizationError("Access denied"))
            c.Abort()
            return
        }

        // Check if user has one of the allowed roles
        hasRole := false
        for _, allowedRole := range allowedRoles {
            if userTenant.Role == allowedRole {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, errors.NewAuthorizationError("Insufficient permissions"))
            c.Abort()
            return
        }

        // Store role in context for further use
        c.Set("user_role", userTenant.Role)

        c.Next()
    }
}
```

**Usage in router:**
```go
// Tenant management routes (OWNER/ADMIN only)
tenantGroup := protected.Group("/tenant")
tenantGroup.Use(middleware.RequireRoleMiddleware(db, models.UserRoleOwner, models.UserRoleAdmin))
{
    tenantGroup.GET("", tenantHandler.GetTenant)
    tenantGroup.GET("/users", tenantHandler.ListUsers)
    tenantGroup.POST("/users/invite", tenantHandler.InviteUser)
    tenantGroup.PUT("/users/:id/role", tenantHandler.UpdateUserRole)
    tenantGroup.DELETE("/users/:id", tenantHandler.RemoveUser)
}
```

---

### Testing Checklist

#### Company Profile
- [ ] ✅ Get company profile (existing company)
- [ ] ✅ Get company profile (no company yet) → 404
- [ ] ✅ Create company profile (valid data)
- [ ] ✅ Create company profile (duplicate for same tenant) → 400
- [ ] ✅ Create company profile (invalid NPWP format) → 400
- [ ] ✅ Create company profile (isPKP=true but no NPWP) → 400
- [ ] ✅ Update company profile (valid data)
- [ ] ✅ Upload company logo (valid image)
- [ ] ✅ Upload company logo (oversized file) → 400
- [ ] ✅ Add bank account (valid data)
- [ ] ✅ Add bank account (set as primary) → unsets other primary
- [ ] ✅ Update bank account
- [ ] ✅ Delete bank account

#### Tenant Management
- [ ] ✅ Get tenant details (with subscription)
- [ ] ✅ List tenant users (all)
- [ ] ✅ List tenant users (filter by role=ADMIN)
- [ ] ✅ List tenant users (filter by isActive=true)
- [ ] ✅ Invite new user (email doesn't exist) → creates user + sends email
- [ ] ✅ Invite existing user (email exists) → creates UserTenant link
- [ ] ✅ Invite user (try to invite OWNER) → 400
- [ ] ✅ Invite user (non-OWNER/ADMIN user tries to invite) → 403
- [ ] ✅ Update user role (valid role change)
- [ ] ✅ Update user role (try to change OWNER role) → 400
- [ ] ✅ Update user role (try to promote to OWNER) → 400
- [ ] ✅ Remove user from tenant (valid removal)
- [ ] ✅ Remove user from tenant (try to remove OWNER) → 400
- [ ] ✅ Remove user from tenant (try to remove last ADMIN) → 400

#### Multi-Tenancy Isolation
- [ ] ✅ User A (Tenant 1) cannot get company profile of Tenant 2
- [ ] ✅ User B (Tenant 2) cannot list users of Tenant 1
- [ ] ✅ User C (no tenant access) cannot access any tenant endpoints → 403

---

## Implementation Priority (Week 1)

### Day 1-2: Company Profile Module
1. Define DTOs (request/response)
2. Implement CompanyService
3. Create CompanyHandler
4. Register routes
5. Add validation (NPWP, phone number)
6. Write unit tests

### Day 3-4: Company Bank Management
1. Implement bank CRUD in CompanyService
2. Add primary bank logic (auto-unset)
3. Add endpoints to CompanyHandler
4. Write integration tests

### Day 5-7: Tenant Management Module
1. Define DTOs (InviteUserRequest, etc.)
2. Implement TenantService
3. Create TenantHandler
4. Implement RequireRoleMiddleware
5. Register routes with RBAC
6. Write unit + integration tests
7. Test multi-tenant isolation

---

## Next Module Group

After completing Foundation Setup (Week 1), proceed to:
**→ `02-MASTER-DATA-MANAGEMENT.md`** (Products, Customers, Suppliers, Warehouses)

---

## Summary

**Foundation Setup** establishes:
1. ✅ Company legal profile (NPWP, PKP, tax settings)
2. ✅ Bank accounts for payment processing
3. ✅ Invoice/SO/PO number formats
4. ✅ Multi-tenant user access control (roles: OWNER, ADMIN, FINANCE, SALES, WAREHOUSE, STAFF)

**Dependencies for Next Modules:**
- Products need company settings (PPN rate, currency)
- Invoices need company details (NPWP, bank accounts, number formats)
- All transactions need tenant isolation enforced by UserTenant roles

**Estimated Completion:** 1 week (7 days)
