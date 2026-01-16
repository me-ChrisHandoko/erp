/**
 * Product Types
 *
 * TypeScript type definitions for product management.
 * Based on backend models/product.go and internal/dto/product_dto.go
 */

/**
 * Product
 * Main product entity with multi-unit and batch tracking support
 */
export interface Product {
  id: string;
  companyId: string;
  tenantId: string;

  // Basic Information
  code: string;
  name: string;
  category?: string | null;
  description?: string | null;
  barcode?: string | null;

  // Unit & Pricing
  baseUnit: string;
  baseCost: string; // decimal as string
  basePrice: string; // decimal as string

  // Stock Management
  minimumStock: string; // decimal as string
  currentStock: string; // DEPRECATED - use WarehouseStock instead

  // Tracking Flags
  isBatchTracked: boolean;
  isPerishable: boolean;
  isActive: boolean;

  // Timestamps
  createdAt: string;
  updatedAt: string;

  // Relations
  units?: ProductUnit[];
  suppliers?: ProductSupplier[];
}

/**
 * Product Unit
 * Multi-unit conversion support (e.g., 1 KARTON = 24 PCS)
 */
export interface ProductUnit {
  id: string;
  productId: string;

  // Unit Information
  unitName: string;
  conversionRate: string; // decimal as string - how many base units
  isBaseUnit: boolean;

  // Pricing (can differ from base product)
  buyPrice?: string | null; // decimal as string
  sellPrice?: string | null; // decimal as string

  // Identification
  barcode?: string | null;
  sku?: string | null;

  // Physical Properties
  weight?: string | null; // decimal as string - in kg
  volume?: string | null; // decimal as string - in m³

  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Product Supplier
 * Supplier relationship with product-specific pricing
 */
export interface ProductSupplier {
  id: string;
  productId: string;
  supplierId: string;

  // Supplier Details (from join)
  supplierCode?: string;
  supplierName?: string;

  // Purchase Information
  supplierProductCode?: string | null;
  supplierProductName?: string | null;
  supplierPrice?: string | null; // decimal as string
  leadTimeDays?: number | null;
  minimumOrderQty?: string | null; // decimal as string
  isPrimarySupplier: boolean;

  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Current Stock Response
 * Stock information per warehouse
 */
export interface CurrentStockResponse {
  totalStock: string; // decimal as string
  warehouses: WarehouseStockInfo[];
}

/**
 * Warehouse Stock Info
 * Stock details for a specific warehouse
 */
export interface WarehouseStockInfo {
  warehouseId: string;
  warehouseName: string;
  warehouseCode: string;
  quantity: string; // decimal as string
  minimumStock?: string | null; // decimal as string
  maximumStock?: string | null; // decimal as string
}

// ============================================================================
// API Request DTOs
// ============================================================================

/**
 * Create Product Request
 */
export interface CreateProductRequest {
  code: string;
  name: string;
  category?: string;
  description?: string;
  baseUnit: string;
  baseCost: string; // decimal as string
  basePrice: string; // decimal as string
  minimumStock?: string; // decimal as string
  barcode?: string;
  isBatchTracked?: boolean;
  isPerishable?: boolean;
  units?: CreateProductUnitRequest[];
}

/**
 * Create Product Unit Request
 */
export interface CreateProductUnitRequest {
  unitName: string;
  conversionRate: string; // decimal as string
  buyPrice?: string; // decimal as string
  sellPrice?: string; // decimal as string
  barcode?: string;
  sku?: string;
  weight?: string; // decimal as string
  volume?: string; // decimal as string
}

/**
 * Update Product Request
 * Partial update - all fields optional
 */
export interface UpdateProductRequest {
  code?: string;
  name?: string;
  category?: string;
  description?: string;
  baseUnit?: string;
  baseCost?: string;
  basePrice?: string;
  minimumStock?: string;
  barcode?: string;
  isBatchTracked?: boolean;
  isPerishable?: boolean;
  isActive?: boolean;
}

/**
 * Update Product Unit Request
 */
export interface UpdateProductUnitRequest {
  unitName?: string;
  conversionRate?: string;
  buyPrice?: string;
  sellPrice?: string;
  barcode?: string;
  sku?: string;
  weight?: string;
  volume?: string;
  isActive?: boolean;
}

/**
 * Link Supplier Request
 */
export interface LinkSupplierRequest {
  supplierId: string;
  supplierProductCode?: string;
  supplierProductName?: string;
  supplierPrice?: string; // decimal as string
  leadTimeDays?: number;
  minimumOrderQty?: string; // decimal as string
  isPrimarySupplier?: boolean;
}

/**
 * Update Product Supplier Request
 */
export interface UpdateProductSupplierRequest {
  supplierProductCode?: string;
  supplierProductName?: string;
  supplierPrice?: string;
  leadTimeDays?: number;
  minimumOrderQty?: string;
  isPrimarySupplier?: boolean;
  isActive?: boolean;
}

// ============================================================================
// API Response DTOs
// ============================================================================

/**
 * Product Response
 * Complete product information from API
 */
export interface ProductResponse {
  id: string;
  code: string;
  name: string;
  category?: string;
  description?: string;
  baseUnit: string;
  baseCost: string;
  basePrice: string;
  minimumStock: string;
  barcode?: string;
  isBatchTracked: boolean;
  isPerishable: boolean;
  isActive: boolean;
  units: ProductUnitResponse[];
  suppliers?: ProductSupplierResponse[];
  currentStock?: CurrentStockResponse;
  createdAt: string;
  updatedAt: string;
}

/**
 * Product Unit Response
 */
export interface ProductUnitResponse {
  id: string;
  productId: string;
  unitName: string;
  conversionRate: string;
  isBaseUnit: boolean;
  buyPrice?: string;
  sellPrice?: string;
  barcode?: string;
  sku?: string;
  weight?: string;
  volume?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Product Supplier Response
 */
export interface ProductSupplierResponse {
  id: string;
  productId: string;
  supplierId: string;
  supplierCode?: string;
  supplierName?: string;
  supplierProductCode?: string;
  supplierProductName?: string;
  supplierPrice?: string;
  leadTimeDays?: number;
  minimumOrderQty?: string;
  isPrimarySupplier: boolean;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/**
 * Product List Response
 * Paginated list of products
 */
export interface ProductListResponse {
  success: boolean;
  data: ProductResponse[];
  pagination: PaginationInfo;
}

/**
 * Pagination Information
 */
export interface PaginationInfo {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasMore: boolean;
}

/**
 * Product Filters
 * Query parameters for listing products
 */
export interface ProductFilters {
  search?: string; // search in code or name
  category?: string;
  supplierId?: string; // filter products by supplier
  isActive?: boolean;
  isBatchTracked?: boolean;
  isPerishable?: boolean;
  page?: number;
  pageSize?: number;
  sortBy?: "code" | "name" | "category" | "createdAt" | "baseCost" | "basePrice";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// Constants & Enums
// ============================================================================

/**
 * Common Product Categories for Indonesian Distribution
 */
export const PRODUCT_CATEGORIES = [
  "Beras",
  "Gula",
  "Minyak Goreng",
  "Tepung",
  "Telur",
  "Susu",
  "Kopi",
  "Teh",
  "Mie Instan",
  "Bumbu Dapur",
  "Minuman",
  "Snack",
  "Sembako Lainnya",
  "Lainnya"
] as const;

/**
 * Common Units for Indonesian Products
 */
export const COMMON_UNITS = [
  "PCS", // Pieces
  "KARTON", // Carton
  "LUSIN", // Dozen
  "PACK", // Pack
  "BOX", // Box
  "KG", // Kilogram
  "GRAM", // Gram
  "LITER", // Liter
  "ML", // Milliliter
  "SACK", // Karung
  "BAL", // Bale
  "ROLL", // Roll
  "UNIT" // Unit
] as const;

/**
 * Sort Options for Product List
 */
export const PRODUCT_SORT_OPTIONS = [
  { value: "code", label: "Kode Produk" },
  { value: "name", label: "Nama Produk" },
  { value: "category", label: "Kategori" },
  { value: "baseCost", label: "Harga Beli" },
  { value: "basePrice", label: "Harga Jual" },
  { value: "createdAt", label: "Tanggal Dibuat" }
] as const;

/**
 * Type guard to check if value is a valid product category
 */
export function isProductCategory(value: string): value is typeof PRODUCT_CATEGORIES[number] {
  return PRODUCT_CATEGORIES.includes(value as typeof PRODUCT_CATEGORIES[number]);
}

/**
 * Type guard to check if value is a valid common unit
 */
export function isCommonUnit(value: string): value is typeof COMMON_UNITS[number] {
  return COMMON_UNITS.includes(value as typeof COMMON_UNITS[number]);
}
