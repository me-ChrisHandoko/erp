// Package document - Document number generation service
package document

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"backend/models"

	"gorm.io/gorm"
)

// DocumentNumberGenerator handles document number generation based on company format settings
type DocumentNumberGenerator struct {
	db *gorm.DB
	mu sync.Mutex // Prevent race condition when generating numbers
}

// DocumentType represents the type of document
type DocumentType string

const (
	DocTypePurchaseOrder    DocumentType = "purchase_order"
	DocTypeSalesOrder       DocumentType = "sales_order"
	DocTypePurchaseInvoice  DocumentType = "purchase_invoice"
	DocTypeSalesInvoice     DocumentType = "sales_invoice"
	DocTypeSupplierPayment  DocumentType = "supplier_payment"
	DocTypeCustomerPayment  DocumentType = "customer_payment"
	DocTypeDelivery         DocumentType = "delivery"
)

// NewDocumentNumberGenerator creates a new document number generator
func NewDocumentNumberGenerator(db *gorm.DB) *DocumentNumberGenerator {
	return &DocumentNumberGenerator{
		db: db,
	}
}

// GenerateNumber generates document number based on company format settings
func (g *DocumentNumberGenerator) GenerateNumber(
	ctx context.Context,
	tenantID string,
	companyID string,
	docType DocumentType,
) (string, error) {
	log.Printf("🔍 DEBUG [DocNumberGen]: Starting - tenantID=%s, companyID=%s, docType=%s", tenantID, companyID, docType)
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. Get company settings
	log.Printf("🔍 DEBUG [DocNumberGen]: Getting company settings...")
	var company models.Company
	if err := g.db.WithContext(ctx).
		Set("tenant_id", tenantID).
		First(&company, "id = ?", companyID).Error; err != nil {
		log.Printf("❌ DEBUG [DocNumberGen]: Failed to get company: %v", err)
		return "", fmt.Errorf("failed to get company: %w", err)
	}
	log.Printf("✅ DEBUG [DocNumberGen]: Company found - Name=%s, POPrefix=%s, POFormat=%s", company.Name, company.POPrefix, company.PONumberFormat)

	// 2. Get format and prefix based on document type
	var prefix, format string
	switch docType {
	case DocTypePurchaseOrder:
		prefix = company.POPrefix
		format = company.PONumberFormat
	case DocTypeSalesOrder:
		prefix = company.SOPrefix
		format = company.SONumberFormat
	case DocTypePurchaseInvoice, DocTypeSalesInvoice:
		prefix = company.InvoicePrefix
		format = company.InvoiceNumberFormat
	case DocTypeSupplierPayment, DocTypeCustomerPayment:
		// Use default payment prefix
		prefix = "PAY"
		format = "{PREFIX}/{YEAR}/{MONTH}/{NUMBER}"
	case DocTypeDelivery:
		// Use SO prefix for deliveries or create separate if needed
		prefix = "DEL"
		format = "{PREFIX}/{YEAR}/{MONTH}/{NUMBER}"
	default:
		log.Printf("❌ DEBUG [DocNumberGen]: Unsupported document type: %s", docType)
		return "", fmt.Errorf("unsupported document type: %s", docType)
	}
	log.Printf("🔍 DEBUG [DocNumberGen]: Using prefix=%s, format=%s", prefix, format)

	// 3. Get next sequence number
	log.Printf("🔍 DEBUG [DocNumberGen]: Getting next sequence...")
	sequence, err := g.getNextSequence(ctx, tenantID, companyID, docType, format)
	if err != nil {
		log.Printf("❌ DEBUG [DocNumberGen]: Failed to get sequence: %v", err)
		return "", err
	}
	log.Printf("✅ DEBUG [DocNumberGen]: Sequence number=%d", sequence)

	// 4. Parse format string and build number
	number := g.buildNumber(prefix, format, sequence)
	log.Printf("✅ DEBUG [DocNumberGen]: Generated number=%s", number)

	return number, nil
}

// buildNumber builds document number by replacing placeholders in format string
func (g *DocumentNumberGenerator) buildNumber(prefix, format string, sequence int) string {
	now := time.Now()

	// Replace placeholders
	result := format
	result = strings.ReplaceAll(result, "{PREFIX}", prefix)
	result = strings.ReplaceAll(result, "{YEAR}", fmt.Sprintf("%d", now.Year()))
	result = strings.ReplaceAll(result, "{MONTH}", fmt.Sprintf("%02d", now.Month()))
	result = strings.ReplaceAll(result, "{NUMBER}", fmt.Sprintf("%04d", sequence))

	return result
}

// getNextSequence gets the next sequence number for the document type
func (g *DocumentNumberGenerator) getNextSequence(
	ctx context.Context,
	tenantID string,
	companyID string,
	docType DocumentType,
	format string,
) (int, error) {
	log.Printf("🔍 DEBUG [getNextSequence]: docType=%s, format=%s", docType, format)
	// Determine if sequence should reset based on format
	shouldResetMonthly := strings.Contains(format, "{MONTH}")
	shouldResetYearly := strings.Contains(format, "{YEAR}") && !shouldResetMonthly
	log.Printf("🔍 DEBUG [getNextSequence]: shouldResetMonthly=%v, shouldResetYearly=%v", shouldResetMonthly, shouldResetYearly)

	var count int64
	var query *gorm.DB

	// Build query based on document type
	switch docType {
	case DocTypePurchaseOrder:
		log.Printf("🔍 DEBUG [getNextSequence]: Building query for PurchaseOrder...")
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Unscoped(). // Include soft-deleted records to avoid duplicate numbers
			Model(&models.PurchaseOrder{}).
			Where("company_id = ?", companyID)

	case DocTypePurchaseInvoice:
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Unscoped(). // Include soft-deleted records to avoid duplicate numbers
			Model(&models.PurchaseInvoice{}).
			Where("company_id = ?", companyID)

	case DocTypeSalesOrder:
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Unscoped(). // Include soft-deleted records to avoid duplicate numbers
			Model(&models.SalesOrder{}).
			Where("company_id = ?", companyID)

	case DocTypeDelivery:
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Unscoped(). // Include soft-deleted records to avoid duplicate numbers
			Model(&models.Delivery{}).
			Where("company_id = ?", companyID)

	case DocTypeSalesInvoice:
		// TODO: Implement when sales invoice model is ready
		log.Printf("❌ DEBUG [getNextSequence]: Sales invoice not implemented")
		return 0, fmt.Errorf("sales invoice not implemented yet")

	case DocTypeCustomerPayment, DocTypeSupplierPayment:
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Unscoped(). // Include soft-deleted records to avoid duplicate numbers
			Model(&models.Payment{}).
			Where("company_id = ?", companyID)

	default:
		log.Printf("❌ DEBUG [getNextSequence]: Unsupported document type: %s", docType)
		return 0, fmt.Errorf("unsupported document type: %s", docType)
	}

	// Add time filters based on format
	now := time.Now()
	if shouldResetMonthly {
		year := now.Year()
		month := int(now.Month())
		log.Printf("🔍 DEBUG [getNextSequence]: Adding monthly filter - year=%d, month=%d", year, month)
		query = query.Where("EXTRACT(YEAR FROM created_at) = ? AND EXTRACT(MONTH FROM created_at) = ?",
			year, month)
	} else if shouldResetYearly {
		year := now.Year()
		log.Printf("🔍 DEBUG [getNextSequence]: Adding yearly filter - year=%d", year)
		query = query.Where("EXTRACT(YEAR FROM created_at) = ?", year)
	}
	// else: never reset (continuous sequence)

	log.Printf("🔍 DEBUG [getNextSequence]: Executing count query...")
	if err := query.Count(&count).Error; err != nil {
		log.Printf("❌ DEBUG [getNextSequence]: Failed to count: %v", err)
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	log.Printf("✅ DEBUG [getNextSequence]: Count result=%d, next sequence=%d", count, count+1)

	return int(count) + 1, nil
}

// GeneratePaymentNumber generates payment document number
func (g *DocumentNumberGenerator) GeneratePaymentNumber(
	ctx context.Context,
	tx *gorm.DB,
	companyID string,
	paymentDate time.Time,
) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Get tenant_id from context
	tenantID, ok := tx.Get("tenant_id")
	if !ok {
		return "", fmt.Errorf("tenant_id not found in transaction context")
	}

	// 1. Get company settings
	var company models.Company
	if err := tx.WithContext(ctx).
		Set("tenant_id", tenantID).
		First(&company, "id = ?", companyID).Error; err != nil {
		return "", fmt.Errorf("failed to get company: %w", err)
	}

	// 2. Use default payment prefix and format
	prefix := "PAY"
	format := "{PREFIX}/{YEAR}/{MONTH}/{NUMBER}"

	// 3. Get next sequence number based on payment date
	sequence, err := g.getNextPaymentSequence(ctx, tx, tenantID.(string), companyID, format, paymentDate)
	if err != nil {
		return "", err
	}

	// 4. Build number using payment date for year/month
	number := g.buildPaymentNumber(prefix, format, sequence, paymentDate)

	return number, nil
}

// buildPaymentNumber builds payment number using payment date for year/month
func (g *DocumentNumberGenerator) buildPaymentNumber(prefix, format string, sequence int, paymentDate time.Time) string {
	// Replace placeholders using payment date
	result := format
	result = strings.ReplaceAll(result, "{PREFIX}", prefix)
	result = strings.ReplaceAll(result, "{YEAR}", fmt.Sprintf("%d", paymentDate.Year()))
	result = strings.ReplaceAll(result, "{MONTH}", fmt.Sprintf("%02d", paymentDate.Month()))
	result = strings.ReplaceAll(result, "{NUMBER}", fmt.Sprintf("%04d", sequence))

	return result
}

// getNextPaymentSequence gets the next sequence number for payments based on payment date
func (g *DocumentNumberGenerator) getNextPaymentSequence(
	ctx context.Context,
	tx *gorm.DB,
	tenantID string,
	companyID string,
	format string,
	paymentDate time.Time,
) (int, error) {
	// Determine if sequence should reset based on format
	shouldResetMonthly := strings.Contains(format, "{MONTH}")
	shouldResetYearly := strings.Contains(format, "{YEAR}") && !shouldResetMonthly

	var count int64
	query := tx.WithContext(ctx).
		Set("tenant_id", tenantID).
		Model(&models.Payment{}).
		Where("company_id = ?", companyID)

	// Add time filters based on format using payment_date
	if shouldResetMonthly {
		year := paymentDate.Year()
		month := int(paymentDate.Month())
		query = query.Where("EXTRACT(YEAR FROM payment_date) = ? AND EXTRACT(MONTH FROM payment_date) = ?",
			year, month)
	} else if shouldResetYearly {
		year := paymentDate.Year()
		query = query.Where("EXTRACT(YEAR FROM payment_date) = ?", year)
	}
	// else: never reset (continuous sequence)

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payments: %w", err)
	}

	return int(count) + 1, nil
}
