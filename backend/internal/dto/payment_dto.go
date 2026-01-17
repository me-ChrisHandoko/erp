package dto

// ============================================================================
// PAYMENT CONSTANTS
// ============================================================================

// PaymentStatus constants
const (
	PaymentStatusUnpaid  = "UNPAID"
	PaymentStatusPartial = "PARTIAL"
	PaymentStatusPaid    = "PAID"
	PaymentStatusOverdue = "OVERDUE"
)

// PaymentMethod constants
const (
	PaymentMethodCash         = "CASH"
	PaymentMethodBankTransfer = "BANK_TRANSFER"
	PaymentMethodCheck        = "CHECK"
	PaymentMethodGiro         = "GIRO"
	PaymentMethodCreditCard   = "CREDIT_CARD"
	PaymentMethodDebitCard    = "DEBIT_CARD"
	PaymentMethodEWallet      = "E_WALLET"
	PaymentMethodOther        = "OTHER"
)

// CheckStatus constants
const (
	CheckStatusIssued    = "ISSUED"
	CheckStatusCleared   = "CLEARED"
	CheckStatusBounced   = "BOUNCED"
	CheckStatusCancelled = "CANCELLED"
)

// ============================================================================
// PAYMENT REQUEST DTOs (Customer Payments)
// ============================================================================

// CreatePaymentRequest represents customer payment creation request
type CreatePaymentRequest struct {
	PaymentDate   string  `json:"paymentDate" binding:"required"`   // ISO date string
	CustomerID    string  `json:"customerId" binding:"required,uuid"`
	InvoiceID     string  `json:"invoiceId" binding:"required,uuid"`
	Amount        string  `json:"amount" binding:"required"`        // decimal as string, must be > 0
	PaymentMethod string  `json:"paymentMethod" binding:"required,oneof=CASH BANK_TRANSFER CHECK GIRO CREDIT_CARD DEBIT_CARD E_WALLET OTHER"`
	Reference     *string `json:"reference" binding:"omitempty,max=100"`
	BankAccountID *string `json:"bankAccountId" binding:"omitempty,uuid"`
	CheckNumber   *string `json:"checkNumber" binding:"omitempty,max=100"`
	CheckDate     *string `json:"checkDate" binding:"omitempty"` // ISO date string
	Notes         *string `json:"notes" binding:"omitempty"`
}

// UpdatePaymentRequest represents customer payment update request
type UpdatePaymentRequest struct {
	PaymentDate   *string `json:"paymentDate" binding:"omitempty"`
	CustomerID    *string `json:"customerId" binding:"omitempty,uuid"`
	InvoiceID     *string `json:"invoiceId" binding:"omitempty,uuid"`
	Amount        *string `json:"amount" binding:"omitempty"`
	PaymentMethod *string `json:"paymentMethod" binding:"omitempty,oneof=CASH BANK_TRANSFER CHECK GIRO CREDIT_CARD DEBIT_CARD E_WALLET OTHER"`
	Reference     *string `json:"reference" binding:"omitempty,max=100"`
	BankAccountID *string `json:"bankAccountId" binding:"omitempty,uuid"`
	CheckNumber   *string `json:"checkNumber" binding:"omitempty,max=100"`
	CheckDate     *string `json:"checkDate" binding:"omitempty"`
	Notes         *string `json:"notes" binding:"omitempty"`
}

// UpdateCheckStatusRequest represents check status update request
type UpdateCheckStatusRequest struct {
	CheckStatus string  `json:"checkStatus" binding:"required,oneof=ISSUED CLEARED BOUNCED CANCELLED"`
	Notes       *string `json:"notes" binding:"omitempty"`
}

// VoidPaymentRequest represents payment void request
type VoidPaymentRequest struct {
	Reason *string `json:"reason" binding:"omitempty"`
}

// PaymentFilters represents payment list filters
type PaymentFilters struct {
	Search        string `form:"search"`         // Search in payment number, customer name, invoice number
	CustomerID    string `form:"customer_id"`    // Filter by customer
	InvoiceID     string `form:"invoice_id"`     // Filter by invoice
	PaymentMethod string `form:"payment_method"` // Filter by payment method
	CheckStatus   string `form:"check_status"`   // Filter by check status
	DateFrom      string `form:"date_from"`      // ISO date string - payment date range start
	DateTo        string `form:"date_to"`        // ISO date string - payment date range end
	AmountMin     string `form:"amount_min"`     // decimal as string - minimum amount
	AmountMax     string `form:"amount_max"`     // decimal as string - maximum amount
	Page          int    `form:"page" binding:"omitempty,min=1"`
	Limit         int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy        string `form:"sort_by" binding:"omitempty,oneof=paymentNumber paymentDate customerName amount createdAt"`
	SortOrder     string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ============================================================================
// PAYMENT RESPONSE DTOs
// ============================================================================

// PaymentResponse represents customer payment information response
type PaymentResponse struct {
	ID              string  `json:"id"`
	PaymentNumber   string  `json:"paymentNumber"`
	PaymentDate     string  `json:"paymentDate"` // ISO date string
	CustomerID      string  `json:"customerId"`
	CustomerName    string  `json:"customerName"`
	CustomerCode    *string `json:"customerCode,omitempty"`
	InvoiceID       string  `json:"invoiceId"`
	InvoiceNumber   string  `json:"invoiceNumber"`
	Amount          string  `json:"amount"` // decimal as string
	PaymentMethod   string  `json:"paymentMethod"`
	Reference       *string `json:"reference,omitempty"`
	BankAccountID   *string `json:"bankAccountId,omitempty"`
	BankAccountName *string `json:"bankAccountName,omitempty"`
	CheckNumber     *string `json:"checkNumber,omitempty"`
	CheckDate       *string `json:"checkDate,omitempty"`   // ISO date string
	CheckStatus     *string `json:"checkStatus,omitempty"` // For CHECK/GIRO payments
	Notes           *string `json:"notes,omitempty"`
	CreatedBy       string  `json:"createdBy"`
	UpdatedBy       *string `json:"updatedBy,omitempty"`
	CreatedAt       string  `json:"createdAt"` // ISO datetime string
	UpdatedAt       string  `json:"updatedAt"` // ISO datetime string
}

// PaymentListResponse represents paginated payment list response
type PaymentListResponse struct {
	Data       []PaymentResponse  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
