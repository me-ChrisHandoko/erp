// Package document - Document number generation service
package document

import (
	"context"
	"fmt"
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
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. Get company settings
	var company models.Company
	if err := g.db.WithContext(ctx).
		Set("tenant_id", tenantID).
		First(&company, "id = ?", companyID).Error; err != nil {
		return "", fmt.Errorf("failed to get company: %w", err)
	}

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
	default:
		return "", fmt.Errorf("unsupported document type: %s", docType)
	}

	// 3. Get next sequence number
	sequence, err := g.getNextSequence(ctx, tenantID, companyID, docType, format)
	if err != nil {
		return "", err
	}

	// 4. Parse format string and build number
	number := g.buildNumber(prefix, format, sequence)

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
	// Determine if sequence should reset based on format
	shouldResetMonthly := strings.Contains(format, "{MONTH}")
	shouldResetYearly := strings.Contains(format, "{YEAR}") && !shouldResetMonthly

	var count int64
	var query *gorm.DB

	// Build query based on document type
	switch docType {
	case DocTypePurchaseOrder:
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Model(&models.PurchaseOrder{}).
			Where("company_id = ?", companyID)

	case DocTypePurchaseInvoice:
		query = g.db.WithContext(ctx).
			Set("tenant_id", tenantID).
			Model(&models.PurchaseInvoice{}).
			Where("company_id = ?", companyID)

	case DocTypeSalesOrder:
		// TODO: Implement when sales order model is ready
		return 0, fmt.Errorf("sales order not implemented yet")

	case DocTypeSalesInvoice:
		// TODO: Implement when sales invoice model is ready
		return 0, fmt.Errorf("sales invoice not implemented yet")

	default:
		return 0, fmt.Errorf("unsupported document type: %s", docType)
	}

	// Add time filters based on format
	now := time.Now()
	if shouldResetMonthly {
		year := now.Year()
		month := int(now.Month())
		query = query.Where("EXTRACT(YEAR FROM created_at) = ? AND EXTRACT(MONTH FROM created_at) = ?",
			year, month)
	} else if shouldResetYearly {
		year := now.Year()
		query = query.Where("EXTRACT(YEAR FROM created_at) = ?", year)
	}
	// else: never reset (continuous sequence)

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return int(count) + 1, nil
}
