// Payment Types
// Maps to backend SupplierPayment model in backend/models/supplier_payment.go

// ============================================================================
// ENUMS & CONSTANTS
// ============================================================================

export const PAYMENT_METHOD = {
  CASH: 'CASH',
  BANK_TRANSFER: 'BANK_TRANSFER',
  CHECK: 'CHECK',
  GIRO: 'GIRO',
  CREDIT_CARD: 'CREDIT_CARD',
  OTHER: 'OTHER',
} as const;

export const PAYMENT_STATUS = {
  PENDING: 'PENDING',
  APPROVED: 'APPROVED',
  REJECTED: 'REJECTED',
  CANCELLED: 'CANCELLED',
} as const;

export type PaymentMethod = keyof typeof PAYMENT_METHOD;
export type PaymentStatus = keyof typeof PAYMENT_STATUS;

// ============================================================================
// REQUEST DTOs
// ============================================================================

export interface CreatePaymentRequest {
  paymentDate: string; // ISO date string
  supplierId: string;
  purchaseOrderId?: string;
  amount: string; // decimal as string
  paymentMethod: PaymentMethod;
  reference?: string; // Transfer reference, check number, etc.
  bankAccountId?: string;
  notes?: string;
}

export interface UpdatePaymentRequest {
  paymentDate?: string;
  supplierId?: string;
  purchaseOrderId?: string;
  amount?: string;
  paymentMethod?: PaymentMethod;
  reference?: string;
  bankAccountId?: string;
  notes?: string;
}

export interface ApprovePaymentRequest {
  notes?: string;
}

// ============================================================================
// RESPONSE DTOs
// ============================================================================

export interface PaymentResponse {
  id: string;
  paymentNumber: string;
  paymentDate: string;
  supplierId: string;
  supplierName: string;
  supplierCode?: string;
  purchaseOrderId?: string;
  poNumber?: string;
  amount: string;
  paymentMethod: PaymentMethod;
  reference?: string;
  bankAccountId?: string;
  bankAccountName?: string;
  notes?: string;
  approvedBy?: string;
  approvedAt?: string;
  status?: PaymentStatus;
  createdBy: string;
  updatedBy?: string;
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// FILTERS & PAGINATION
// ============================================================================

export interface PaymentFilters {
  search?: string; // Search payment number, supplier name/code, reference
  supplierId?: string;
  paymentMethod?: PaymentMethod;
  status?: PaymentStatus;
  dateFrom?: string; // ISO date string
  dateTo?: string;
  amountMin?: string;
  amountMax?: string;
  page?: number;
  pageSize?: number;
  sortBy?: 'paymentNumber' | 'paymentDate' | 'supplierName' | 'amount' | 'createdAt';
  sortOrder?: 'asc' | 'desc';
}

export interface PaginationResponse {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface PaymentListResponse {
  data: PaymentResponse[];
  pagination: PaginationResponse;
}

// ============================================================================
// UI HELPER TYPES
// ============================================================================

export interface PaymentFormData {
  paymentNumber: string;
  paymentDate: Date;
  supplierId: string;
  purchaseOrderId: string;
  amount: string;
  paymentMethod: PaymentMethod;
  reference: string;
  bankAccountId: string;
  notes: string;
}

// ============================================================================
// DISPLAY HELPERS
// ============================================================================

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  CASH: 'Tunai',
  BANK_TRANSFER: 'Transfer Bank',
  CHECK: 'Cek',
  GIRO: 'Giro',
  CREDIT_CARD: 'Kartu Kredit',
  OTHER: 'Lainnya',
};

export const PAYMENT_STATUS_LABELS: Record<PaymentStatus, string> = {
  PENDING: 'Menunggu',
  APPROVED: 'Disetujui',
  REJECTED: 'Ditolak',
  CANCELLED: 'Dibatalkan',
};
