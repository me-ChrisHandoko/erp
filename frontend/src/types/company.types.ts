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
