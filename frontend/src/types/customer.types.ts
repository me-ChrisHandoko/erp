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

// ============================================================================
// FREQUENT PRODUCTS (Quick Add Optimization)
// ============================================================================

/**
 * Frequent Product Item
 * Single frequently purchased product with purchase statistics
 */
export interface FrequentProductItem {
  productId: string;
  productCode: string;
  productName: string;
  frequency: number; // Number of times purchased
  totalQty: string; // Total quantity purchased (decimal)
  lastOrderDate: string; // Last order date (ISO 8601)
  baseUnitId: string;
  baseUnitName: string;
  latestPrice: string; // Latest purchase price (decimal)
}

/**
 * Frequent Products Response
 * Response from GET /api/v1/customers/:id/frequent-products
 */
export interface FrequentProductsResponse {
  customerId: string;
  customerName: string;
  warehouseId?: string; // Optional warehouse filter
  frequentProducts: FrequentProductItem[];
  totalOrders: number; // Total sales orders analyzed
  dateRangeFrom: string; // Analyzed from date
  dateRangeTo: string; // Analyzed to date
}

/**
 * Frequent Products API Response
 */
export interface FrequentProductsApiResponse {
  success: boolean;
  data: FrequentProductsResponse;
}

// ============================================================================
// CUSTOMER CREDIT INFO
// ============================================================================

/**
 * Customer Credit Info Response
 * Response from GET /api/v1/customers/:id/credit-info
 */
export interface CustomerCreditInfoResponse {
  customerId: string;
  customerName: string;
  customerCode: string;
  creditLimit: string; // Credit limit (decimal as string)
  outstandingAmount: string; // Total unpaid invoices (decimal)
  availableCredit: string; // Credit limit - outstanding (decimal)
  overdueAmount: string; // Overdue invoices (decimal)
  paymentTermDays: number; // Payment terms in days
  isExceedingLimit: boolean; // True if outstanding > credit limit
  utilizationPercent: string; // (Outstanding / Credit Limit) * 100
}

/**
 * Customer Credit Info API Response
 */
export interface CustomerCreditInfoApiResponse {
  success: boolean;
  data: CustomerCreditInfoResponse;
}
