/**
 * Goods Receipt Types
 *
 * TypeScript type definitions for goods receipt management (Penerimaan Barang).
 * Based on backend models/goods_receipt.go and internal/dto/goods_receipt_dto.go
 */

// ============================================================================
// STATUS TYPES
// ============================================================================

/**
 * Goods Receipt Status
 * PENDING → RECEIVED → INSPECTED → ACCEPTED/REJECTED/PARTIAL
 */
export type GoodsReceiptStatus =
  | "PENDING"
  | "RECEIVED"
  | "INSPECTED"
  | "ACCEPTED"
  | "REJECTED"
  | "PARTIAL";

export const GOODS_RECEIPT_STATUS_OPTIONS = [
  { value: "PENDING", label: "Menunggu", color: "bg-yellow-100 text-yellow-800" },
  { value: "RECEIVED", label: "Diterima", color: "bg-blue-100 text-blue-800" },
  { value: "INSPECTED", label: "Diperiksa", color: "bg-purple-100 text-purple-800" },
  { value: "ACCEPTED", label: "Disetujui", color: "bg-green-100 text-green-800" },
  { value: "REJECTED", label: "Ditolak", color: "bg-red-100 text-red-800" },
  { value: "PARTIAL", label: "Sebagian", color: "bg-orange-100 text-orange-800" },
] as const;

// ============================================================================
// REJECTION DISPOSITION TYPES (Odoo+M3 Model)
// ============================================================================

/**
 * Rejection Disposition Status
 * Tracks what happens to rejected items
 */
export type RejectionDispositionStatus =
  | "PENDING_REPLACEMENT"
  | "CREDIT_REQUESTED"
  | "RETURNED"
  | "WRITTEN_OFF";

export const REJECTION_DISPOSITION_OPTIONS = [
  { value: "PENDING_REPLACEMENT", label: "Menunggu Penggantian", color: "bg-yellow-100 text-yellow-800" },
  { value: "CREDIT_REQUESTED", label: "Kredit Diminta", color: "bg-blue-100 text-blue-800" },
  { value: "RETURNED", label: "Dikembalikan", color: "bg-purple-100 text-purple-800" },
  { value: "WRITTEN_OFF", label: "Dihapuskan", color: "bg-gray-100 text-gray-800" },
] as const;

// ============================================================================
// BASIC RESPONSE TYPES
// ============================================================================

export interface GoodsReceiptProductResponse {
  id: string;
  code: string;
  name: string;
  baseUnit: string;
  isBatchTracked: boolean; // Indicates if batch number is required
  isPerishable: boolean; // Indicates if expiry date is required
}

export interface GoodsReceiptProductUnitResponse {
  id: string;
  unitName: string;
  conversionRate: string;
}

export interface GoodsReceiptPurchaseOrderResponse {
  id: string;
  poNumber: string;
  poDate: string;
}

export interface WarehouseBasicResponse {
  id: string;
  code: string;
  name: string;
}

export interface SupplierBasicResponse {
  id: string;
  code: string;
  name: string;
}

export interface UserBasicResponse {
  id: string;
  fullName: string;
  email: string;
}

// ============================================================================
// GOODS RECEIPT ITEM
// ============================================================================

export interface GoodsReceiptItemResponse {
  id: string;
  goodsReceiptId: string;
  purchaseOrderItemId: string;
  productId: string;
  product?: GoodsReceiptProductResponse;
  productUnitId?: string;
  productUnit?: GoodsReceiptProductUnitResponse;
  batchNumber?: string;
  manufactureDate?: string;
  expiryDate?: string;
  orderedQty: string;
  receivedQty: string;
  acceptedQty: string;
  invoicedQty: string; // Qty already invoiced for this GRN item
  rejectedQty: string;
  rejectionReason?: string;
  qualityNote?: string;
  notes?: string;
  // Rejection Disposition Fields (Odoo+M3 Model)
  rejectionDisposition?: RejectionDispositionStatus;
  dispositionNotes?: string;
  dispositionResolved?: boolean;
  dispositionResolvedAt?: string;
  dispositionResolvedBy?: string;
  dispositionResolvedNotes?: string;
  dispositionResolver?: UserBasicResponse;
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// GOODS RECEIPT
// ============================================================================

export interface GoodsReceiptResponse {
  id: string;
  grnNumber: string;
  grnDate: string;
  purchaseOrderId: string;
  purchaseOrder?: GoodsReceiptPurchaseOrderResponse;
  warehouseId: string;
  warehouse?: WarehouseBasicResponse;
  supplierId: string;
  supplier?: SupplierBasicResponse;
  status: GoodsReceiptStatus;
  receivedBy?: string;
  receiver?: UserBasicResponse;
  receivedAt?: string;
  inspectedBy?: string;
  inspector?: UserBasicResponse;
  inspectedAt?: string;
  supplierInvoice?: string;
  supplierDONumber?: string;
  notes?: string;            // General notes (from creation)
  receiveNotes?: string;     // Notes during receive (PENDING → RECEIVED)
  inspectionNotes?: string;  // Notes during inspection (RECEIVED → INSPECTED)
  acceptanceNotes?: string;  // Notes during acceptance (INSPECTED → ACCEPTED/PARTIAL)
  rejectionNotes?: string;   // Notes during rejection
  itemCount: number; // Number of items in the goods receipt
  invoiceStatus: "NONE" | "PARTIAL" | "FULL"; // Invoice status for this GRN
  totalAcceptedQty: string; // Total accepted qty across all items
  totalInvoicedQty: string; // Total invoiced qty across all items
  items?: GoodsReceiptItemResponse[];
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// API REQUEST TYPES
// ============================================================================

export interface CreateGoodsReceiptItemRequest {
  purchaseOrderItemId: string;
  productId: string;
  productUnitId?: string;
  batchNumber?: string;
  manufactureDate?: string;
  expiryDate?: string;
  receivedQty: string;
  acceptedQty?: string;
  rejectedQty?: string;
  rejectionReason?: string;
  qualityNote?: string;
  notes?: string;
}

export interface CreateGoodsReceiptRequest {
  purchaseOrderId: string;
  grnDate: string;
  supplierInvoice?: string;
  supplierDONumber?: string;
  notes?: string;
  items: CreateGoodsReceiptItemRequest[];
}

export interface UpdateGoodsReceiptItemRequest {
  id?: string;
  receivedQty: string;
  acceptedQty?: string;
  rejectedQty?: string;
  batchNumber?: string;
  manufactureDate?: string;
  expiryDate?: string;
  rejectionReason?: string;
  qualityNote?: string;
  notes?: string;
}

export interface UpdateGoodsReceiptRequest {
  grnDate?: string;
  supplierInvoice?: string;
  supplierDONumber?: string;
  notes?: string;
  items?: UpdateGoodsReceiptItemRequest[];
}

export interface ReceiveGoodsRequest {
  notes?: string;
}

export interface InspectGoodsRequest {
  notes?: string;
  items?: UpdateGoodsReceiptItemRequest[];
}

export interface AcceptGoodsRequest {
  notes?: string;
}

export interface RejectGoodsRequest {
  rejectionReason: string;
}

/**
 * Update Rejection Disposition Request (Odoo+M3 Model)
 * Field names must match backend DTO
 */
export interface UpdateRejectionDispositionRequest {
  rejectionDisposition: RejectionDispositionStatus;
  dispositionNotes?: string;
}

/**
 * Resolve Disposition Request (Odoo+M3 Model)
 * Field names must match backend DTO
 */
export interface ResolveDispositionRequest {
  dispositionResolvedNotes?: string;
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

export interface GoodsReceiptListResponse {
  success: boolean;
  data: GoodsReceiptResponse[];
  pagination: PaginationInfo;
}

// ============================================================================
// FILTER TYPES
// ============================================================================

export interface GoodsReceiptFilters {
  search?: string;
  status?: GoodsReceiptStatus;
  purchaseOrderId?: string;
  supplierId?: string;
  warehouseId?: string;
  dateFrom?: string;
  dateTo?: string;
  page?: number;
  pageSize?: number;
  sortBy?: "grnNumber" | "grnDate" | "status" | "createdAt";
  sortOrder?: "asc" | "desc";
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

export function getGoodsReceiptStatusLabel(status: GoodsReceiptStatus): string {
  const option = GOODS_RECEIPT_STATUS_OPTIONS.find((o) => o.value === status);
  return option?.label || status;
}

export function getGoodsReceiptStatusColor(status: GoodsReceiptStatus): string {
  const option = GOODS_RECEIPT_STATUS_OPTIONS.find((o) => o.value === status);
  return option?.color || "bg-gray-100 text-gray-800";
}

export function getRejectionDispositionLabel(disposition: RejectionDispositionStatus): string {
  const option = REJECTION_DISPOSITION_OPTIONS.find((o) => o.value === disposition);
  return option?.label || disposition;
}

export function getRejectionDispositionColor(disposition: RejectionDispositionStatus): string {
  const option = REJECTION_DISPOSITION_OPTIONS.find((o) => o.value === disposition);
  return option?.color || "bg-gray-100 text-gray-800";
}
