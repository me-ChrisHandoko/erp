// Sales Payment Types (Customer Payments)
// Maps to backend InvoicePayment model in backend/models/sales.go

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

export const CHECK_STATUS = {
  ISSUED: 'ISSUED',
  CLEARED: 'CLEARED',
  BOUNCED: 'BOUNCED',
  CANCELLED: 'CANCELLED',
} as const;

export type PaymentMethod = keyof typeof PAYMENT_METHOD;
export type CheckStatus = keyof typeof CHECK_STATUS;

// ============================================================================
// REQUEST DTOs
// ============================================================================

export interface CreateSalesPaymentRequest {
  paymentDate: string; // ISO date string
  customerId: string;
  invoiceId: string;
  amount: string; // decimal as string
  paymentMethod: PaymentMethod;
  reference?: string; // Transfer reference, check number, etc.
  bankAccountId?: string; // Company bank account where payment received
  checkNumber?: string;
  checkDate?: string;
  notes?: string;
}

export interface UpdateSalesPaymentRequest {
  paymentDate?: string;
  customerId?: string;
  invoiceId?: string;
  amount?: string;
  paymentMethod?: PaymentMethod;
  reference?: string;
  bankAccountId?: string;
  checkNumber?: string;
  checkDate?: string;
  notes?: string;
}

export interface VoidPaymentRequest {
  reason?: string;
}

export interface UpdateCheckStatusRequest {
  checkStatus: CheckStatus;
  notes?: string;
}

// ============================================================================
// RESPONSE DTOs
// ============================================================================

export interface SalesPaymentResponse {
  id: string;
  paymentNumber: string;
  paymentDate: string;
  customerId: string;
  customerName: string;
  customerCode?: string;
  invoiceId: string;
  invoiceNumber: string;
  amount: string;
  paymentMethod: PaymentMethod;
  reference?: string;
  bankAccountId?: string;
  bankAccountName?: string;
  checkNumber?: string;
  checkDate?: string;
  checkStatus?: CheckStatus;
  notes?: string;
  createdBy: string;
  updatedBy?: string;
  createdAt: string;
  updatedAt: string;
}

// ============================================================================
// FILTERS & PAGINATION
// ============================================================================

export interface SalesPaymentFilters {
  search?: string; // Search payment number, customer name/code, invoice number, reference
  customerId?: string;
  invoiceId?: string;
  paymentMethod?: PaymentMethod;
  checkStatus?: CheckStatus;
  dateFrom?: string; // ISO date string
  dateTo?: string;
  amountMin?: string;
  amountMax?: string;
  page?: number;
  pageSize?: number;
  sortBy?: 'paymentNumber' | 'paymentDate' | 'customerName' | 'amount' | 'createdAt';
  sortOrder?: 'asc' | 'desc';
}

export interface PaginationResponse {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface SalesPaymentListResponse {
  data: SalesPaymentResponse[];
  pagination: PaginationResponse;
}

// ============================================================================
// UI HELPER TYPES
// ============================================================================

export interface SalesPaymentFormData {
  paymentNumber: string;
  paymentDate: Date;
  customerId: string;
  invoiceId: string;
  amount: string;
  paymentMethod: PaymentMethod;
  reference: string;
  bankAccountId: string;
  checkNumber: string;
  checkDate: Date | null;
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

export const CHECK_STATUS_LABELS: Record<CheckStatus, string> = {
  ISSUED: 'Diterbitkan',
  CLEARED: 'Cair',
  BOUNCED: 'Tolak',
  CANCELLED: 'Batal',
};
