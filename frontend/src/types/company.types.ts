/**
 * Company Types
 *
 * TypeScript type definitions for company profile management.
 * Based on backend models/company.go and internal/dto/company_dto.go
 */

/**
 * Company Profile
 * Full company information including settings and bank accounts
 */
export interface Company {
  id: string;

  // Legal Entity Information
  name: string;
  legalName: string;
  entityType: "CV" | "PT" | "UD" | "Firma";

  // Address
  address: string;
  city: string;
  province: string;
  postalCode?: string | null;
  country: string;

  // Contact
  phone: string;
  email: string;
  website?: string | null;

  // Indonesian Tax Compliance
  npwp?: string | null;
  isPKP: boolean;
  ppnRate: number;
  fakturPajakSeries?: string | null;
  sppkpNumber?: string | null;

  // Branding
  logoUrl?: string | null;
  primaryColor?: string | null;
  secondaryColor?: string | null;

  // Invoice Settings
  invoicePrefix: string;
  invoiceNumberFormat: string;
  invoiceFooter?: string | null;
  invoiceTerms?: string | null;

  // Sales Order Settings
  soPrefix: string;
  soNumberFormat: string;

  // Purchase Order Settings
  poPrefix: string;
  poNumberFormat: string;

  // System Settings
  currency: string;
  timezone: string;
  locale: string;

  // Business Hours
  businessHoursStart?: string | null;
  businessHoursEnd?: string | null;
  workingDays?: string | null;

  isActive: boolean;
  createdAt: string;
  updatedAt: string;

  // Relations
  banks?: CompanyBank[];
}

/**
 * Company Bank Account
 * Bank account information for payments and invoices
 */
export interface CompanyBank {
  id: string;
  companyId: string;
  bankName: string;
  accountNumber: string;
  accountName: string;
  branchName?: string | null;
  isPrimary: boolean;
  checkPrefix?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Company Response (API DTO)
 * Simplified company information from API responses
 */
export interface CompanyResponse {
  id: string;
  name: string;
  legalName?: string;
  entityType?: "CV" | "PT" | "UD" | "Firma";
  npwp?: string;
  nib?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  phone?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
  isPkp: boolean;
  ppnRate: number;
  invoicePrefix?: string;
  isActive: boolean;
  banks?: CompanyBankInfo[];
}

/**
 * Company Bank Info (Nested in CompanyResponse)
 * Simplified bank account information
 */
export interface CompanyBankInfo {
  id: string;
  bankName: string;
  accountNumber: string;
  accountName: string;
  branchName?: string;
  isPrimary: boolean;
  checkPrefix?: string;
  isActive: boolean;
}

/**
 * Update Company Request
 * Partial update payload for company profile
 */
export interface UpdateCompanyRequest {
  name?: string;
  legalName?: string;
  npwp?: string;
  nib?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  phone?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
  isPkp?: boolean;
  ppnRate?: number;
  invoicePrefix?: string;
  invoiceNumberFormat?: string;
  fakturPajakSeries?: string;
  sppkpNumber?: string;
}

/**
 * Add Bank Account Request
 * Payload for creating new bank account
 */
export interface AddBankAccountRequest {
  bankName: string;
  accountNumber: string;
  accountName: string;
  branchName?: string;
  isPrimary: boolean;
  checkPrefix?: string;
}

/**
 * Update Bank Account Request
 * Partial update payload for bank account
 */
export interface UpdateBankAccountRequest {
  bankName?: string;
  accountNumber?: string;
  accountName?: string;
  branchName?: string;
  isPrimary?: boolean;
  checkPrefix?: string;
}

/**
 * Bank Account Response (API DTO)
 * Bank account information from API responses
 */
export interface BankAccountResponse {
  id: string;
  bankName: string;
  accountNumber: string;
  accountName: string;
  branchName?: string;
  isPrimary: boolean;
  checkPrefix?: string;
  isActive: boolean;
}

/**
 * Indonesian Entity Types
 */
export const ENTITY_TYPES = ["CV", "PT", "UD", "Firma"] as const;

/**
 * Common Indonesian Banks
 */
export const INDONESIAN_BANKS = [
  "BCA",
  "Mandiri",
  "BRI",
  "BNI",
  "BTN",
  "CIMB Niaga",
  "Danamon",
  "Permata",
  "Maybank",
  "OCBC NISP",
  "Panin",
  "Bank Jago",
  "SeaBank",
  "Bank Digital BCA",
  "Lainnya"
] as const;

/**
 * Indonesian Provinces
 */
export const INDONESIAN_PROVINCES = [
  "Aceh",
  "Sumatera Utara",
  "Sumatera Barat",
  "Riau",
  "Kepulauan Riau",
  "Jambi",
  "Sumatera Selatan",
  "Bangka Belitung",
  "Bengkulu",
  "Lampung",
  "DKI Jakarta",
  "Banten",
  "Jawa Barat",
  "Jawa Tengah",
  "DI Yogyakarta",
  "Jawa Timur",
  "Bali",
  "Nusa Tenggara Barat",
  "Nusa Tenggara Timur",
  "Kalimantan Barat",
  "Kalimantan Tengah",
  "Kalimantan Selatan",
  "Kalimantan Timur",
  "Kalimantan Utara",
  "Sulawesi Utara",
  "Sulawesi Tengah",
  "Sulawesi Selatan",
  "Sulawesi Tenggara",
  "Gorontalo",
  "Sulawesi Barat",
  "Maluku",
  "Maluku Utara",
  "Papua",
  "Papua Barat",
  "Papua Tengah",
  "Papua Pegunungan",
  "Papua Selatan",
  "Papua Barat Daya"
] as const;

// ============================================================================
// PHASE 5: Multi-Company State Management Types
// ============================================================================

/**
 * Company role types for role-based access control
 * Maps to backend user_company_role.role field
 */
export type CompanyRole = 'OWNER' | 'ADMIN' | 'FINANCE' | 'SALES' | 'WAREHOUSE' | 'STAFF';

/**
 * Access tier levels for multi-company system
 * 0 = No access
 * 1 = Tenant-level access (all companies in tenant)
 * 2 = Company-specific access
 */
export type AccessTier = 0 | 1 | 2;

/**
 * Company access information
 * Represents user's permissions and role for a specific company
 */
export interface CompanyAccess {
  companyId: string;
  tenantId: string;
  role: CompanyRole;
  accessTier: AccessTier;
  hasAccess: boolean;
  // Granular permissions
  canView: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canApprove: boolean;
}

/**
 * Available company for user
 * Combines company basic info with user's access information
 */
export interface AvailableCompany {
  id: string;
  tenantId: string;
  name: string;
  legalName?: string;
  npwp?: string;
  city?: string;
  province?: string;
  isPKP: boolean;
  isActive: boolean;
  logoUrl?: string | null;
  // User access info
  role: CompanyRole;
  accessTier: AccessTier;
}

/**
 * Active company context
 * Full information about currently active company
 */
export interface ActiveCompany extends AvailableCompany {
  access: CompanyAccess;
}

/**
 * Company state for Redux
 */
export interface CompanyState {
  activeCompany: ActiveCompany | null;
  availableCompanies: AvailableCompany[];
  loading: boolean;
  error: string | null;
  initialized: boolean;
}

/**
 * Switch company request
 */
export interface SwitchCompanyRequest {
  company_id: string;
}

/**
 * Switch company response
 */
export interface SwitchCompanyResponse {
  access_token: string;
  refresh_token?: string;
  company_id: string;
  company_name: string;
  message: string;
}

/**
 * Get available companies response
 */
export interface GetAvailableCompaniesResponse {
  companies: AvailableCompany[];
}
