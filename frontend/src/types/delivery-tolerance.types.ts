/**
 * Delivery Tolerance Types
 *
 * TypeScript type definitions for delivery tolerance settings (SAP Model).
 * Hierarchical tolerance configuration: Product > Category > Company > Default
 * Based on backend models/delivery_tolerance.go
 */

// ============================================================================
// TOLERANCE LEVEL TYPES
// ============================================================================

/**
 * Delivery Tolerance Level (SAP Model)
 * Hierarchy: PRODUCT > CATEGORY > COMPANY > DEFAULT
 */
export type DeliveryToleranceLevel = "COMPANY" | "CATEGORY" | "PRODUCT";

export const DELIVERY_TOLERANCE_LEVEL_OPTIONS = [
  { value: "COMPANY", label: "Perusahaan", description: "Toleransi default untuk semua produk" },
  { value: "CATEGORY", label: "Kategori", description: "Toleransi untuk kategori produk tertentu" },
  { value: "PRODUCT", label: "Produk", description: "Toleransi untuk produk tertentu" },
] as const;

// ============================================================================
// RESPONSE TYPES
// ============================================================================

export interface DeliveryToleranceProductResponse {
  id: string;
  code: string;
  name: string;
  category?: string;
  baseUnit: string;
}

export interface DeliveryToleranceResponse {
  id: string;
  level: DeliveryToleranceLevel;
  categoryName?: string;
  productId?: string;
  product?: DeliveryToleranceProductResponse;
  underDeliveryTolerance: string; // Percentage as string (e.g., "5.00")
  overDeliveryTolerance: string; // Percentage as string
  unlimitedOverDelivery: boolean;
  isActive: boolean;
  notes?: string;
  createdAt: string;
  updatedAt: string;
  createdBy?: string;
  updatedBy?: string;
}

export interface EffectiveToleranceResponse {
  productId: string;
  productCode: string;
  productName: string;
  underDeliveryTolerance: string; // Percentage as string
  overDeliveryTolerance: string; // Percentage as string
  unlimitedOverDelivery: boolean;
  resolvedFrom: "PRODUCT" | "CATEGORY" | "COMPANY" | "DEFAULT";
  toleranceId?: string;
}

// ============================================================================
// API REQUEST TYPES
// ============================================================================

export interface CreateDeliveryToleranceRequest {
  level: DeliveryToleranceLevel;
  categoryName?: string; // Required for CATEGORY level
  productId?: string; // Required for PRODUCT level
  underDeliveryTolerance: string; // Percentage as string (e.g., "5.00")
  overDeliveryTolerance: string; // Percentage as string
  unlimitedOverDelivery?: boolean;
  isActive?: boolean;
  notes?: string;
}

export interface UpdateDeliveryToleranceRequest {
  underDeliveryTolerance?: string;
  overDeliveryTolerance?: string;
  unlimitedOverDelivery?: boolean;
  isActive?: boolean;
  notes?: string;
}

// ============================================================================
// API RESPONSE TYPES
// ============================================================================

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface DeliveryToleranceListResponse {
  success: boolean;
  data: DeliveryToleranceResponse[];
  pagination: PaginationInfo;
}

// ============================================================================
// FILTER TYPES
// ============================================================================

export interface DeliveryToleranceFilters {
  level?: DeliveryToleranceLevel;
  categoryName?: string;
  productId?: string;
  isActive?: boolean;
  page?: number;
  pageSize?: number;
  sortBy?: "level" | "createdAt" | "updatedAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

export function getToleranceLevelLabel(level: DeliveryToleranceLevel): string {
  const option = DELIVERY_TOLERANCE_LEVEL_OPTIONS.find((o) => o.value === level);
  return option?.label || level;
}

export function getToleranceLevelDescription(level: DeliveryToleranceLevel): string {
  const option = DELIVERY_TOLERANCE_LEVEL_OPTIONS.find((o) => o.value === level);
  return option?.description || "";
}

export function getResolvedFromLabel(
  resolvedFrom: "PRODUCT" | "CATEGORY" | "COMPANY" | "DEFAULT"
): string {
  const labels: Record<string, string> = {
    PRODUCT: "Produk",
    CATEGORY: "Kategori",
    COMPANY: "Perusahaan",
    DEFAULT: "Default (0%)",
  };
  return labels[resolvedFrom] || resolvedFrom;
}

export function formatTolerancePercentage(value: string): string {
  const num = parseFloat(value);
  if (isNaN(num)) return "0%";
  return `${num.toFixed(2)}%`;
}

/**
 * Calculate acceptable quantity range based on tolerance
 * @param orderedQty - The ordered quantity
 * @param underTolerance - Under delivery tolerance percentage
 * @param overTolerance - Over delivery tolerance percentage
 * @param unlimitedOver - Whether unlimited over delivery is allowed
 * @returns Object with min and max acceptable quantities
 */
export function calculateAcceptableRange(
  orderedQty: number,
  underTolerance: string,
  overTolerance: string,
  unlimitedOver: boolean
): { min: number; max: number | null } {
  const underPct = parseFloat(underTolerance) / 100;
  const overPct = parseFloat(overTolerance) / 100;

  const min = orderedQty * (1 - underPct);
  const max = unlimitedOver ? null : orderedQty * (1 + overPct);

  return { min, max };
}
