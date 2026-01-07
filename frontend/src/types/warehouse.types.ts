/**
 * Warehouse Types
 *
 * TypeScript type definitions for warehouse management.
 * Based on backend models/warehouse.go and internal/dto/warehouse_dto.go
 */

/**
 * Warehouse Type Enum
 * Types of warehouses in the distribution system
 */
export type WarehouseType = "MAIN" | "BRANCH" | "CONSIGNMENT" | "TRANSIT";

/**
 * Warehouse
 * Main warehouse entity for multi-location inventory management
 */
export interface Warehouse {
  id: string;
  companyId: string;
  tenantId: string;

  // Basic Information
  code: string;
  name: string;
  type: WarehouseType;

  // Location Information
  address?: string | null;
  city?: string | null;
  province?: string | null;
  postalCode?: string | null;

  // Contact Information
  phone?: string | null;
  email?: string | null;

  // Management
  managerID?: string | null;
  capacity?: string | null; // decimal as string (in m²)

  // Status
  isActive: boolean;

  // Timestamps
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// API Request DTOs
// ============================================================================

/**
 * Create Warehouse Request
 */
export interface CreateWarehouseRequest {
  code: string;
  name: string;
  type: WarehouseType;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  phone?: string;
  email?: string;
  managerID?: string;
  capacity?: string; // decimal as string
}

/**
 * Update Warehouse Request
 * Partial update - all fields optional except those being updated
 */
export interface UpdateWarehouseRequest {
  code?: string;
  name?: string;
  type?: WarehouseType;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  phone?: string;
  email?: string;
  managerID?: string;
  capacity?: string;
  isActive?: boolean;
}

// ============================================================================
// API Response DTOs
// ============================================================================

/**
 * Warehouse Response
 * Complete warehouse information from API
 */
export interface WarehouseResponse {
  id: string;
  code: string;
  name: string;
  type: WarehouseType;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  phone?: string;
  email?: string;
  managerID?: string;
  capacity?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Warehouse List Response
 * Paginated list of warehouses
 */
export interface WarehouseListResponse {
  success: boolean;
  data: WarehouseResponse[];
  pagination: PaginationInfo;
}

/**
 * Pagination Information
 * Matches backend PaginationInfo structure from dto/product_dto.go
 */
export interface PaginationInfo {
  page: number;
  limit: number; // Backend uses "limit" not "pageSize"
  total: number;
  totalPages: number;
}

/**
 * Warehouse Filters
 * Query parameters for listing warehouses
 */
export interface WarehouseFilters {
  search?: string; // search in code or name
  type?: WarehouseType;
  city?: string;
  province?: string;
  managerID?: string;
  isActive?: boolean;
  page?: number;
  pageSize?: number;
  sortBy?: "code" | "name" | "type" | "createdAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// Constants & Enums
// ============================================================================

/**
 * Warehouse Types for Dropdown
 * Indonesian labels for warehouse types
 */
export const WAREHOUSE_TYPES = [
  { value: "MAIN", label: "Gudang Utama" },
  { value: "BRANCH", label: "Gudang Cabang" },
  { value: "CONSIGNMENT", label: "Konsinyasi" },
  { value: "TRANSIT", label: "Transit" },
] as const;

/**
 * Sort Options for Warehouse List
 */
export const WAREHOUSE_SORT_OPTIONS = [
  { value: "code", label: "Kode Gudang" },
  { value: "name", label: "Nama Gudang" },
  { value: "type", label: "Tipe Gudang" },
  { value: "createdAt", label: "Tanggal Dibuat" },
] as const;

/**
 * Type guard to check if value is a valid warehouse type
 */
export function isWarehouseType(value: string): value is WarehouseType {
  return ["MAIN", "BRANCH", "CONSIGNMENT", "TRANSIT"].includes(value);
}

/**
 * Get warehouse type label in Indonesian
 */
export function getWarehouseTypeLabel(type: WarehouseType): string {
  const found = WAREHOUSE_TYPES.find((t) => t.value === type);
  return found ? found.label : type;
}

/**
 * Get warehouse type badge color
 */
export function getWarehouseTypeBadgeColor(type: WarehouseType): string {
  const colors: Record<WarehouseType, string> = {
    MAIN: "bg-blue-500 text-white hover:bg-blue-600",
    BRANCH: "bg-green-500 text-white hover:bg-green-600",
    CONSIGNMENT: "bg-purple-500 text-white hover:bg-purple-600",
    TRANSIT: "bg-orange-500 text-white hover:bg-orange-600",
  };
  return colors[type] || "bg-gray-500 text-white";
}

/**
 * Format location display from warehouse data
 * Combines city and province with smart formatting
 */
export function formatWarehouseLocation(warehouse: {
  city?: string | null;
  province?: string | null;
  address?: string | null;
}): string {
  const parts: string[] = [];

  if (warehouse.city) parts.push(warehouse.city);
  if (warehouse.province) parts.push(warehouse.province);

  return parts.length > 0 ? parts.join(", ") : "-";
}

/**
 * Format capacity display with unit
 */
export function formatCapacity(capacity?: string | null): string {
  if (!capacity) return "-";
  const num = Number(capacity);
  if (isNaN(num)) return "-";
  return `${num.toLocaleString("id-ID")} m²`;
}
