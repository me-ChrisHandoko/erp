/**
 * Delivery Types
 *
 * TypeScript interfaces for delivery management
 * Matches backend models in models/delivery.go
 */

// ============================================================================
// ENUMS
// ============================================================================

export type DeliveryType = 'NORMAL' | 'RETURN' | 'REPLACEMENT';
export type DeliveryStatus = 'PREPARED' | 'IN_TRANSIT' | 'DELIVERED' | 'CONFIRMED' | 'CANCELLED';

// ============================================================================
// DELIVERY INTERFACES
// ============================================================================

/**
 * DeliveryItem - Delivery line item
 */
export interface DeliveryItem {
  id: string;
  deliveryId: string;
  salesOrderItemId: string;
  productId: string;
  productUnitId?: string | null;
  batchId?: string | null;
  quantity: number;
  notes?: string | null;
  createdAt: string;
  updatedAt: string;

  // Relations (populated by API)
  product?: {
    id: string;
    code: string;
    name: string;
    baseUnit: string;
  };
  productUnit?: {
    id: string;
    name: string;
    conversionFactor: number;
  } | null;
  batch?: {
    id: string;
    batchNumber: string;
    expiryDate?: string | null;
  } | null;
}

/**
 * Delivery - Delivery order header
 */
export interface Delivery {
  id: string;
  tenantId: string;
  companyId: string;
  deliveryNumber: string;
  deliveryDate: string;
  salesOrderId: string;
  warehouseId: string;
  customerId: string;
  type: DeliveryType;
  status: DeliveryStatus;
  deliveryAddress?: string | null;
  driverName?: string | null;
  vehicleNumber?: string | null;
  departureTime?: string | null;
  arrivalTime?: string | null;
  receivedBy?: string | null;
  receivedAt?: string | null;
  signatureUrl?: string | null;
  photoUrl?: string | null;
  ttnkNumber?: string | null; // Tracking number
  expeditionService?: string | null; // JNE, SiCepat, etc.
  notes?: string | null;
  createdAt: string;
  updatedAt: string;

  // Relations (populated by API)
  salesOrder?: {
    id: string;
    soNumber: string;
    soDate: string;
    totalAmount: number;
  };
  warehouse?: {
    id: string;
    code: string;
    name: string;
  };
  customer?: {
    id: string;
    code: string;
    name: string;
    address?: string | null;
    phone?: string | null;
  };
  items?: DeliveryItem[];
}

/**
 * DeliveryResponse - API response for single delivery
 */
export interface DeliveryResponse extends Delivery {}

/**
 * DeliveryListResponse - API response for delivery list with pagination
 */
export interface DeliveryListResponse {
  data: DeliveryResponse[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

// ============================================================================
// FILTERS & PARAMS
// ============================================================================

/**
 * DeliveryFilters - Query parameters for filtering deliveries
 */
export interface DeliveryFilters {
  page?: number;
  pageSize?: number;
  sortBy?: 'deliveryNumber' | 'deliveryDate' | 'status' | 'customerName';
  sortOrder?: 'asc' | 'desc';
  search?: string; // Search by delivery number, customer name
  status?: DeliveryStatus;
  type?: DeliveryType;
  warehouseId?: string;
  customerId?: string;
  salesOrderId?: string;
  dateFrom?: string; // ISO date string
  dateTo?: string; // ISO date string
}

// ============================================================================
// CREATE/UPDATE PAYLOADS
// ============================================================================

/**
 * CreateDeliveryItemPayload - Payload for creating delivery item
 */
export interface CreateDeliveryItemPayload {
  salesOrderItemId: string;
  productId: string;
  productUnitId?: string | null;
  batchId?: string | null;
  quantity: number;
  notes?: string | null;
}

/**
 * CreateDeliveryPayload - Payload for creating delivery
 */
export interface CreateDeliveryPayload {
  salesOrderId: string;
  deliveryDate: string; // ISO date string
  warehouseId: string;
  customerId: string;
  type?: DeliveryType;
  deliveryAddress?: string | null;
  driverName?: string | null;
  vehicleNumber?: string | null;
  ttnkNumber?: string | null;
  expeditionService?: string | null;
  notes?: string | null;
  items: CreateDeliveryItemPayload[];
}

/**
 * UpdateDeliveryPayload - Payload for updating delivery
 */
export interface UpdateDeliveryPayload {
  deliveryDate?: string;
  status?: DeliveryStatus;
  deliveryAddress?: string | null;
  driverName?: string | null;
  vehicleNumber?: string | null;
  departureTime?: string | null;
  arrivalTime?: string | null;
  receivedBy?: string | null;
  receivedAt?: string | null;
  signatureUrl?: string | null;
  photoUrl?: string | null;
  ttnkNumber?: string | null;
  expeditionService?: string | null;
  notes?: string | null;
}

/**
 * UpdateDeliveryStatusPayload - Payload for updating delivery status
 */
export interface UpdateDeliveryStatusPayload {
  status: DeliveryStatus;
  receivedBy?: string | null;
  receivedAt?: string | null;
  signatureUrl?: string | null;
  photoUrl?: string | null;
  notes?: string | null;
}

// ============================================================================
// UTILITY TYPES
// ============================================================================

/**
 * DeliveryStatusOption - For status filter dropdown
 */
export interface DeliveryStatusOption {
  value: DeliveryStatus;
  label: string;
  color: string;
}

/**
 * DeliveryTypeOption - For type filter dropdown
 */
export interface DeliveryTypeOption {
  value: DeliveryType;
  label: string;
}

// ============================================================================
// CONSTANTS
// ============================================================================

/**
 * Delivery status options for UI
 */
export const DELIVERY_STATUS_OPTIONS: DeliveryStatusOption[] = [
  { value: 'PREPARED', label: 'Disiapkan', color: 'bg-blue-500' },
  { value: 'IN_TRANSIT', label: 'Dalam Perjalanan', color: 'bg-yellow-500' },
  { value: 'DELIVERED', label: 'Terkirim', color: 'bg-green-500' },
  { value: 'CONFIRMED', label: 'Dikonfirmasi', color: 'bg-emerald-500' },
  { value: 'CANCELLED', label: 'Dibatalkan', color: 'bg-red-500' },
];

/**
 * Delivery type options for UI
 */
export const DELIVERY_TYPE_OPTIONS: DeliveryTypeOption[] = [
  { value: 'NORMAL', label: 'Normal' },
  { value: 'RETURN', label: 'Retur' },
  { value: 'REPLACEMENT', label: 'Penggantian' },
];

/**
 * Get delivery status label in Indonesian
 */
export function getDeliveryStatusLabel(status: DeliveryStatus): string {
  const option = DELIVERY_STATUS_OPTIONS.find(opt => opt.value === status);
  return option?.label || status;
}

/**
 * Get delivery status color class
 */
export function getDeliveryStatusColor(status: DeliveryStatus): string {
  const option = DELIVERY_STATUS_OPTIONS.find(opt => opt.value === status);
  return option?.color || 'bg-gray-500';
}

/**
 * Get delivery type label in Indonesian
 */
export function getDeliveryTypeLabel(type: DeliveryType): string {
  const option = DELIVERY_TYPE_OPTIONS.find(opt => opt.value === type);
  return option?.label || type;
}
