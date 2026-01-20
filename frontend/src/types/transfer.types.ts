/**
 * Stock Transfer Types
 *
 * Type definitions for inter-warehouse stock transfers.
 * Based on backend models: StockTransfer + StockTransferItem
 */

import type { Warehouse } from "./warehouse.types";
import type { Product } from "./product.types";
import type { PaginationInfo } from "./api";

/**
 * Stock Transfer Status Enum
 * Workflow: DRAFT → SHIPPED → RECEIVED (or CANCELLED)
 */
export type StockTransferStatus =
  | "DRAFT"      // Created, can edit/delete
  | "SHIPPED"    // Shipped from source, can receive/cancel
  | "RECEIVED"   // Received at destination, read-only
  | "CANCELLED"; // Cancelled, read-only

/**
 * Stock Transfer Item (Detail/Line Item)
 */
export interface StockTransferItem {
  id: string;
  stockTransferId: string;
  productId: string;
  batchId?: string; // Optional for batch-tracked products
  quantity: string; // Decimal string (15,3 precision)
  notes?: string;
  createdAt: string;
  updatedAt: string;

  // Relations (populated by backend)
  product?: Product;
}

/**
 * Stock Transfer (Header)
 */
export interface StockTransfer {
  id: string;
  tenantId: string;
  companyId: string;
  transferNumber: string; // Auto-generated (e.g., "TRF-2024-0001")
  transferDate: string; // ISO date string
  sourceWarehouseId: string;
  destWarehouseId: string;
  status: StockTransferStatus;
  shippedBy?: string; // User ID who shipped
  shippedAt?: string; // ISO timestamp
  receivedBy?: string; // User ID who received
  receivedAt?: string; // ISO timestamp
  notes?: string;
  createdAt: string;
  updatedAt: string;

  // Relations (populated by backend)
  sourceWarehouse?: Warehouse;
  destWarehouse?: Warehouse;
  items: StockTransferItem[];
}

/**
 * Status Counts for statistics cards
 */
export interface TransferStatusCounts {
  draft: number;
  shipped: number;
  received: number;
  cancelled: number;
}

/**
 * Transfer List Response (matches backend pagination structure)
 */
export interface TransferListResponse {
  success: boolean;
  data: StockTransfer[];
  pagination: PaginationInfo;
  statusCounts?: TransferStatusCounts;
}

/**
 * Single Transfer Response
 */
export interface TransferResponse extends StockTransfer {}

/**
 * Transfer Filters for List Query
 */
export interface TransferFilters {
  page?: number;
  pageSize?: number;
  sortBy?: "transferNumber" | "transferDate" | "status" | "sourceWarehouse" | "destWarehouse";
  sortOrder?: "asc" | "desc";
  search?: string; // Search by transfer number or product name
  status?: StockTransferStatus; // Filter by status
  sourceWarehouseId?: string; // Filter by source warehouse
  destWarehouseId?: string; // Filter by destination warehouse
  dateFrom?: string; // ISO date string
  dateTo?: string; // ISO date string
}

/**
 * Create Transfer Request
 */
export interface CreateTransferRequest {
  transferDate: string; // ISO date string
  sourceWarehouseId: string;
  destWarehouseId: string;
  notes?: string;
  items: {
    productId: string;
    batchId?: string;
    quantity: string; // Decimal string
    notes?: string;
  }[];
}

/**
 * Update Transfer Request (DRAFT only)
 */
export interface UpdateTransferRequest {
  transferDate?: string;
  sourceWarehouseId?: string;
  destWarehouseId?: string;
  notes?: string;
  items?: {
    id?: string; // Existing item ID (if updating)
    productId: string;
    batchId?: string;
    quantity: string;
    notes?: string;
  }[];
}

/**
 * Ship Transfer Request
 */
export interface ShipTransferRequest {
  shippedAt?: string; // Optional, defaults to now
  notes?: string;
}

/**
 * Receive Transfer Request
 */
export interface ReceiveTransferRequest {
  receivedAt?: string; // Optional, defaults to now
  notes?: string;
}

/**
 * Cancel Transfer Request
 */
export interface CancelTransferRequest {
  reason: string; // Required cancellation reason
}
