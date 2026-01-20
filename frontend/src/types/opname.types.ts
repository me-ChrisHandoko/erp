/**
 * Stock Opname Types
 *
 * TypeScript type definitions for stock opname (physical inventory count) management.
 * Stock opname is the process of physically counting inventory and comparing it with system records.
 *
 * Based on backend patterns for multi-tenant ERP system.
 */

/**
 * Stock Opname Status Workflow
 * - draft: Initial creation, can be edited freely
 * - in_progress: Counting is ongoing
 * - completed: Counting finished, ready for approval
 * - approved: Approved and stock adjustments applied
 */
export type StockOpnameStatus = 'draft' | 'in_progress' | 'completed' | 'approved';

/**
 * Stock Opname
 * Main entity for physical inventory counting
 */
export interface StockOpname {
  id: string;
  companyId: string;
  tenantId: string;

  // Opname Information
  opnameNumber: string; // Auto-generated: OPN-YYYYMMDD-XXX
  opnameDate: string; // ISO date string (YYYY-MM-DD)
  warehouseId: string;
  warehouseName?: string; // From join with warehouses table

  // Status & Workflow
  status: StockOpnameStatus;

  // Summary Calculations
  totalItems: number; // Total number of product items in opname
  totalExpectedQty: string; // Total expected quantity (from system) - decimal as string
  totalActualQty: string; // Total actual quantity (from physical count) - decimal as string
  totalDifference: string; // Total difference (actualQty - expectedQty) - decimal as string

  // Audit Trail
  notes?: string | null; // General notes for the opname
  createdBy: string; // User ID who created
  createdByName?: string; // User name from join
  approvedBy?: string | null; // User ID who approved
  approvedByName?: string | null; // User name from join
  approvedAt?: string | null; // ISO datetime string

  // Timestamps
  createdAt: string; // ISO datetime string
  updatedAt: string; // ISO datetime string

  // Relations
  items?: StockOpnameItem[]; // Opname line items
}

/**
 * Stock Opname Item
 * Line item for each product in the opname
 */
export interface StockOpnameItem {
  id: string;
  opnameId: string;
  productId: string;

  // Product Information (from join)
  productCode?: string;
  productName?: string;
  productUnit?: string; // Base unit

  // Quantities
  expectedQty: string; // Quantity from system (warehouse stock) - decimal as string
  actualQty: string; // Quantity from physical count - decimal as string
  difference: string; // actualQty - expectedQty - decimal as string

  // Item Notes
  notes?: string | null; // Specific notes for this item

  // Timestamps
  createdAt: string; // ISO datetime string
  updatedAt: string; // ISO datetime string
}

// ==================== API Request Types ====================

/**
 * Create Stock Opname Request
 * Body for POST /api/v1/stock-opnames
 */
export interface CreateStockOpnameRequest {
  warehouseId: string;
  opnameDate: string; // YYYY-MM-DD format
  notes?: string;
  items?: CreateStockOpnameItemRequest[];
}

/**
 * Create Stock Opname Item Request
 * Item data when creating opname
 */
export interface CreateStockOpnameItemRequest {
  productId: string;
  expectedQty: string; // decimal as string
  actualQty: string; // decimal as string
  notes?: string;
}

/**
 * Update Stock Opname Request
 * Body for PUT /api/v1/stock-opnames/:id
 */
export interface UpdateStockOpnameRequest {
  opnameDate?: string; // YYYY-MM-DD format
  notes?: string;
  status?: 'draft' | 'in_progress' | 'completed'; // Cannot set to 'approved' via update
}

/**
 * Update Stock Opname Item Request
 * Body for PUT /api/v1/stock-opnames/:opnameId/items/:itemId
 */
export interface UpdateStockOpnameItemRequest {
  expectedQty?: string; // decimal as string
  actualQty?: string; // decimal as string
  notes?: string;
}

/**
 * Batch Update Stock Opname Item Request
 * Body item for batch update
 */
export interface BatchUpdateStockOpnameItemRequest {
  itemId: string;
  actualQty?: string; // decimal as string
  notes?: string;
}

/**
 * Batch Update Stock Opname Items Request
 * Body for PUT /api/v1/stock-opnames/:opnameId/items/batch
 */
export interface BatchUpdateStockOpnameItemsRequest {
  items: BatchUpdateStockOpnameItemRequest[];
}

/**
 * Approve Stock Opname Request
 * Body for POST /api/v1/stock-opnames/:id/approve
 */
export interface ApproveStockOpnameRequest {
  approvalNotes?: string;
}

// ==================== API Response Types ====================

/**
 * Stock Opname Response
 * Single opname with full details
 */
export type StockOpnameResponse = StockOpname;

/**
 * Status Counts for statistics cards
 */
export interface StockOpnameStatusCounts {
  draft: number;
  inProgress: number;
  completed: number;
  approved: number;
  cancelled: number;
}

/**
 * Stock Opname List Response
 * Paginated list of opnames
 */
export interface StockOpnameListResponse {
  success: boolean;
  data: StockOpname[];
  pagination: {
    page: number;
    pageSize: number;
    totalItems: number;
    totalPages: number;
    hasMore: boolean;
  };
  statusCounts?: StockOpnameStatusCounts;
}

// ==================== Filter & Query Types ====================

/**
 * Stock Opname Filters
 * Query parameters for listing opnames
 */
export interface StockOpnameFilters {
  // Pagination
  page?: number;
  pageSize?: number;

  // Sorting
  sortBy?: 'opnameNumber' | 'opnameDate' | 'warehouseName' | 'status' | 'totalDifference';
  sortOrder?: 'asc' | 'desc';

  // Filters
  search?: string; // Search by opname number, notes, or warehouse name
  warehouseId?: string; // Filter by warehouse
  status?: StockOpnameStatus; // Filter by status
  dateFrom?: string; // Filter by opname date (YYYY-MM-DD)
  dateTo?: string; // Filter by opname date (YYYY-MM-DD)
}

// ==================== UI Helper Types ====================

/**
 * Stock Opname Status Config
 * UI configuration for status badges
 */
export interface StockOpnameStatusConfig {
  label: string;
  variant: 'default' | 'secondary' | 'outline' | 'destructive';
  className?: string; // Additional custom styling
  color: string;
}

/**
 * Status Configuration Map
 */
export const OPNAME_STATUS_CONFIG: Record<StockOpnameStatus, StockOpnameStatusConfig> = {
  draft: {
    label: 'Draft',
    variant: 'secondary',
    color: 'text-gray-600',
  },
  in_progress: {
    label: 'In Progress',
    variant: 'default',
    color: 'text-blue-600',
  },
  completed: {
    label: 'Completed',
    variant: 'outline',
    className: 'border-orange-500 text-orange-700',
    color: 'text-orange-600',
  },
  approved: {
    label: 'Approved',
    variant: 'outline',
    className: 'border-green-500 text-green-700 bg-green-50',
    color: 'text-green-600',
  },
};

/**
 * Opname Item Summary
 * Calculated summary for display
 */
export interface OpnameItemSummary {
  totalItems: number;
  totalExpectedQty: number;
  totalActualQty: number;
  totalDifference: number;
  itemsWithDifference: number; // Count of items with difference != 0
}
