/**
 * Supplier Types
 *
 * TypeScript type definitions for supplier management.
 * Follows the same pattern as product.types.ts for consistency.
 */

/**
 * Supplier
 * Main supplier entity for managing business partners and vendors
 */
export interface Supplier {
  id: string;
  companyId: string;
  tenantId: string;

  // Basic Information
  code: string;
  name: string;
  type?: string | null;
  description?: string | null;

  // Contact Information
  contactName?: string | null;
  email?: string | null;
  phone?: string | null;
  address?: string | null;
  city?: string | null;
  province?: string | null;
  postalCode?: string | null;

  // Tax & Business Information
  taxId?: string | null; // NPWP for Indonesian businesses

  // Business Terms
  paymentTerms?: string | null;
  creditLimit?: string | null; // decimal as string
  leadTimeDays?: number | null;
  minimumOrderValue?: string | null; // decimal as string

  isActive: boolean;

  // Timestamps
  createdAt: string;
  updatedAt: string;

  // Relations
  products?: SupplierProductInfo[];
}

/**
 * Supplier Product Info
 * Information about products supplied by this supplier
 */
export interface SupplierProductInfo {
  id: string;
  productId: string;
  productCode: string;
  productName: string;
  supplierPrice?: string; // decimal as string
  leadTimeDays?: number;
  minimumOrderQty?: string; // decimal as string
  isPrimary: boolean;
}

// ============================================================================
// API Request DTOs
// ============================================================================

/**
 * Create Supplier Request
 */
export interface CreateSupplierRequest {
  code: string;
  name: string;
  type?: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  npwp?: string;
  isPKP?: boolean;
  contactPerson?: string;
  contactPhone?: string;
  paymentTerm?: number; // Days (0 = cash)
  creditLimit?: string; // decimal as string
  notes?: string;
}

/**
 * Update Supplier Request
 * Partial update - all fields optional
 */
export interface UpdateSupplierRequest {
  code?: string;
  name?: string;
  type?: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  npwp?: string;
  isPKP?: boolean;
  contactPerson?: string;
  contactPhone?: string;
  paymentTerm?: number;
  creditLimit?: string;
  notes?: string;
  isActive?: boolean;
}

// ============================================================================
// API Response DTOs
// ============================================================================

/**
 * Supplier Response
 * Complete supplier information from API
 */
export interface SupplierResponse {
  id: string;
  code: string;
  name: string;
  type?: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  npwp?: string;
  isPKP: boolean;
  contactPerson?: string;
  contactPhone?: string;
  paymentTerm: number;
  creditLimit: string;
  currentOutstanding: string;
  overdueAmount: string;
  lastTransactionAt?: string;
  notes?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Supplier List Response
 * Paginated list of suppliers
 */
export interface SupplierListResponse {
  success: boolean;
  data: SupplierResponse[];
  pagination: PaginationInfo;
}

/**
 * Pagination Information
 */
export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

/**
 * Supplier Filters
 * Query parameters for listing suppliers
 */
export interface SupplierFilters {
  search?: string; // search in code, name, email
  type?: string;
  city?: string;
  province?: string;
  isActive?: boolean;
  page?: number;
  pageSize?: number;
  sortBy?: "code" | "name" | "type" | "city" | "createdAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// Constants & Enums
// ============================================================================

/**
 * Common Supplier Types for Indonesian Distribution
 * Must match backend validation: MANUFACTURER, DISTRIBUTOR, WHOLESALER
 */
export const SUPPLIER_TYPES = [
  "MANUFACTURER",
  "DISTRIBUTOR",
  "WHOLESALER",
] as const;

/**
 * Common Payment Terms
 */
export const PAYMENT_TERMS = [
  "COD", // Cash on Delivery
  "NET 7",
  "NET 15",
  "NET 30",
  "NET 45",
  "NET 60",
  "NET 90",
] as const;

/**
 * Indonesian Provinces
 * Common provinces for supplier location filtering
 */
export const INDONESIAN_PROVINCES = [
  "DKI Jakarta",
  "Jawa Barat",
  "Jawa Tengah",
  "Jawa Timur",
  "Banten",
  "DI Yogyakarta",
  "Bali",
  "Sumatera Utara",
  "Sumatera Selatan",
  "Sumatera Barat",
  "Kalimantan Timur",
  "Sulawesi Selatan",
  "Lainnya",
] as const;

/**
 * Sort Options for Supplier List
 */
export const SUPPLIER_SORT_OPTIONS = [
  { value: "code", label: "Kode Supplier" },
  { value: "name", label: "Nama Supplier" },
  { value: "type", label: "Tipe" },
  { value: "city", label: "Kota" },
  { value: "createdAt", label: "Tanggal Dibuat" },
] as const;

/**
 * Type guard to check if value is a valid supplier type
 */
export function isSupplierType(
  value: string
): value is (typeof SUPPLIER_TYPES)[number] {
  return SUPPLIER_TYPES.includes(value as (typeof SUPPLIER_TYPES)[number]);
}

/**
 * Type guard to check if value is a valid payment term
 */
export function isPaymentTerm(
  value: string
): value is (typeof PAYMENT_TERMS)[number] {
  return PAYMENT_TERMS.includes(value as (typeof PAYMENT_TERMS)[number]);
}
