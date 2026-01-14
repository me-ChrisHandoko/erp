/**
 * Inventory Adjustment Types
 *
 * Type definitions for inventory stock adjustments.
 * Adjustments are used to correct stock levels due to:
 * - Shrinkage (penyusutan)
 * - Damage (kerusakan)
 * - Expired items (kadaluarsa)
 * - Theft (pencurian)
 * - Stock opname corrections
 * - System errors
 *
 * Based on backend models for multi-tenant ERP system.
 */

import type { Warehouse } from "./warehouse.types";
import type { Product } from "./product.types";
import type { PaginationInfo } from "./api";

/**
 * Adjustment Status Enum
 * Workflow: DRAFT → APPROVED (or CANCELLED)
 */
export type AdjustmentStatus =
  | "DRAFT"     // Created, can edit/delete
  | "APPROVED"  // Approved and stock adjusted, read-only
  | "CANCELLED"; // Cancelled, read-only

/**
 * Adjustment Type Enum
 * Reason for the adjustment
 */
export type AdjustmentType =
  | "INCREASE"  // Stock increase (penambahan)
  | "DECREASE"; // Stock decrease (pengurangan)

/**
 * Adjustment Reason Enum
 * Specific reason for the adjustment
 */
export type AdjustmentReason =
  | "SHRINKAGE"    // Penyusutan
  | "DAMAGE"       // Kerusakan
  | "EXPIRED"      // Kadaluarsa
  | "THEFT"        // Pencurian
  | "OPNAME"       // Hasil stock opname
  | "CORRECTION"   // Koreksi kesalahan sistem
  | "RETURN"       // Barang retur
  | "OTHER";       // Lainnya

/**
 * Inventory Adjustment Item (Detail/Line Item)
 */
export interface InventoryAdjustmentItem {
  id: string;
  adjustmentId: string;
  productId: string;
  batchId?: string; // Optional for batch-tracked products
  quantityBefore: string; // Stock before adjustment - decimal as string
  quantityAdjusted: string; // Adjustment quantity (positive or negative) - decimal as string
  quantityAfter: string; // Stock after adjustment - decimal as string
  unitCost: string; // Unit cost for value calculation - decimal as string
  totalValue: string; // quantityAdjusted * unitCost - decimal as string
  notes?: string;
  createdAt: string;
  updatedAt: string;

  // Relations (populated by backend)
  product?: Product;
}

/**
 * Inventory Adjustment (Header)
 */
export interface InventoryAdjustment {
  id: string;
  tenantId: string;
  companyId: string;
  adjustmentNumber: string; // Auto-generated (e.g., "ADJ-2024-0001")
  adjustmentDate: string; // ISO date string
  warehouseId: string;
  adjustmentType: AdjustmentType;
  reason: AdjustmentReason;
  status: AdjustmentStatus;
  totalItems: number; // Total number of products adjusted
  totalValue: string; // Total adjustment value - decimal as string
  approvedBy?: string; // User ID who approved
  approvedAt?: string; // ISO timestamp
  notes?: string;
  createdBy: string; // User ID who created
  createdAt: string;
  updatedAt: string;

  // Relations (populated by backend)
  warehouse?: Warehouse;
  items: InventoryAdjustmentItem[];
  createdByUser?: {
    id: string;
    fullName: string;
    email: string;
  };
  approvedByUser?: {
    id: string;
    fullName: string;
    email: string;
  };
}

/**
 * Adjustment List Response (matches backend pagination structure)
 */
export interface AdjustmentListResponse {
  success: boolean;
  data: InventoryAdjustment[];
  pagination: PaginationInfo;
}

/**
 * Single Adjustment Response
 */
export interface AdjustmentResponse extends InventoryAdjustment {}

/**
 * Adjustment Filters for List Query
 */
export interface AdjustmentFilters {
  page?: number;
  pageSize?: number;
  sortBy?: "adjustmentNumber" | "adjustmentDate" | "status" | "warehouse" | "adjustmentType" | "reason";
  sortOrder?: "asc" | "desc";
  search?: string; // Search by adjustment number or product name
  status?: AdjustmentStatus; // Filter by status
  warehouseId?: string; // Filter by warehouse
  adjustmentType?: AdjustmentType; // Filter by type (increase/decrease)
  reason?: AdjustmentReason; // Filter by reason
  dateFrom?: string; // ISO date string
  dateTo?: string; // ISO date string
}

/**
 * Create Adjustment Request
 */
export interface CreateAdjustmentRequest {
  adjustmentDate: string; // ISO date string
  warehouseId: string;
  adjustmentType: AdjustmentType;
  reason: AdjustmentReason;
  notes?: string;
  items: {
    productId: string;
    batchId?: string;
    quantityAdjusted: string; // Decimal string (positive for increase, negative for decrease)
    unitCost: string; // Decimal string
    notes?: string;
  }[];
}

/**
 * Update Adjustment Request (DRAFT only)
 */
export interface UpdateAdjustmentRequest {
  adjustmentDate?: string;
  warehouseId?: string;
  adjustmentType?: AdjustmentType;
  reason?: AdjustmentReason;
  notes?: string;
  items?: {
    id?: string; // Existing item ID (if updating)
    productId: string;
    batchId?: string;
    quantityAdjusted: string;
    unitCost: string;
    notes?: string;
  }[];
}

/**
 * Approve Adjustment Request
 */
export interface ApproveAdjustmentRequest {
  approvedAt?: string; // Optional, defaults to now
  notes?: string;
}

/**
 * Cancel Adjustment Request
 */
export interface CancelAdjustmentRequest {
  reason: string; // Required cancellation reason
}

// ==================== UI Helper Types ====================

/**
 * Adjustment Status Config
 * UI configuration for status badges
 */
export interface AdjustmentStatusConfig {
  label: string;
  variant: "default" | "secondary" | "outline" | "destructive";
  className?: string;
  color: string;
}

/**
 * Status Configuration Map
 */
export const ADJUSTMENT_STATUS_CONFIG: Record<AdjustmentStatus, AdjustmentStatusConfig> = {
  DRAFT: {
    label: "Draft",
    variant: "secondary",
    color: "text-gray-600",
  },
  APPROVED: {
    label: "Disetujui",
    variant: "outline",
    className: "border-green-500 text-green-700 bg-green-50",
    color: "text-green-600",
  },
  CANCELLED: {
    label: "Dibatalkan",
    variant: "outline",
    className: "border-red-500 text-red-700 bg-red-50",
    color: "text-red-600",
  },
};

/**
 * Adjustment Type Config
 */
export interface AdjustmentTypeConfig {
  label: string;
  className?: string;
  color: string;
}

/**
 * Type Configuration Map
 */
export const ADJUSTMENT_TYPE_CONFIG: Record<AdjustmentType, AdjustmentTypeConfig> = {
  INCREASE: {
    label: "Penambahan",
    className: "bg-green-100 text-green-800",
    color: "text-green-600",
  },
  DECREASE: {
    label: "Pengurangan",
    className: "bg-red-100 text-red-800",
    color: "text-red-600",
  },
};

/**
 * Adjustment Reason Config
 */
export interface AdjustmentReasonConfig {
  label: string;
  description: string;
}

/**
 * Reason Configuration Map
 */
export const ADJUSTMENT_REASON_CONFIG: Record<AdjustmentReason, AdjustmentReasonConfig> = {
  SHRINKAGE: {
    label: "Penyusutan",
    description: "Pengurangan stok karena penyusutan alami",
  },
  DAMAGE: {
    label: "Kerusakan",
    description: "Pengurangan stok karena barang rusak",
  },
  EXPIRED: {
    label: "Kadaluarsa",
    description: "Pengurangan stok karena barang kadaluarsa",
  },
  THEFT: {
    label: "Kehilangan",
    description: "Pengurangan stok karena kehilangan atau pencurian",
  },
  OPNAME: {
    label: "Hasil Opname",
    description: "Penyesuaian berdasarkan hasil stock opname",
  },
  CORRECTION: {
    label: "Koreksi Sistem",
    description: "Koreksi kesalahan pencatatan sistem",
  },
  RETURN: {
    label: "Barang Retur",
    description: "Penambahan stok dari barang retur",
  },
  OTHER: {
    label: "Lainnya",
    description: "Alasan lainnya",
  },
};
