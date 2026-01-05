// Customer Type Definitions for ERP Distribution System

/**
 * Customer Type - represents different categories of customers
 */
export type CustomerType = "Retail" | "Grosir" | "Distributor";

/**
 * Core Customer Entity
 */
export interface Customer {
  id: string;
  companyId: string;
  tenantId: string;
  code: string;
  name: string;
  customerType: CustomerType;
  contactPerson?: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  npwp?: string; // Indonesian Tax ID
  creditLimit?: string; // Decimal as string
  creditTermDays?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Customer Response from API
 */
export interface CustomerResponse {
  id: string;
  companyId: string;
  tenantId: string;
  code: string;
  name: string;
  customerType: CustomerType;
  contactPerson?: string | null;
  phone?: string | null;
  email?: string | null;
  address?: string | null;
  city?: string | null;
  province?: string | null;
  postalCode?: string | null;
  npwp?: string | null;
  creditLimit?: string | null;
  creditTermDays?: number | null; // Frontend alias
  paymentTerm?: number | null; // Backend field name (same as creditTermDays)
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Request payload for creating a new customer
 */
export interface CreateCustomerRequest {
  code: string;
  name: string;
  customerType: CustomerType;
  contactPerson?: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  npwp?: string;
  creditLimit?: string;
  creditTermDays?: number;
}

/**
 * Request payload for updating an existing customer
 */
export interface UpdateCustomerRequest {
  code?: string;
  name?: string;
  customerType?: CustomerType;
  contactPerson?: string;
  phone?: string;
  email?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  npwp?: string;
  creditLimit?: string;
  creditTermDays?: number;
  isActive?: boolean;
}

/**
 * Filter parameters for customer list queries
 */
export interface CustomerFilters {
  search?: string;
  customerType?: CustomerType;
  isActive?: boolean;
  page?: number;
  pageSize?: number;
  sortBy?: "code" | "name" | "customerType" | "city" | "creditLimit" | "createdAt";
  sortOrder?: "asc" | "desc";
}

/**
 * Pagination information
 */
export interface PaginationInfo {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasMore: boolean;
}

/**
 * List response from API
 */
export interface CustomerListResponse {
  success: boolean;
  data: CustomerResponse[];
  pagination: PaginationInfo;
}

/**
 * Single customer response from API
 */
export interface CustomerDetailResponse {
  success: boolean;
  data: CustomerResponse;
}

/**
 * Common customer type constants
 */
export const CUSTOMER_TYPES: readonly CustomerType[] = [
  "Retail",
  "Grosir",
  "Distributor",
] as const;

/**
 * Type guard for CustomerType
 */
export function isCustomerType(value: string): value is CustomerType {
  return CUSTOMER_TYPES.includes(value as CustomerType);
}
