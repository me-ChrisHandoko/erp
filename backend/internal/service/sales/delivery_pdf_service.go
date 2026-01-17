package sales

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"

	"backend/models"
)

// GenerateDeliveryNotePDF generates a PDF delivery note (surat jalan)
func (s *DeliveryService) GenerateDeliveryNotePDF(delivery *models.Delivery) ([]byte, error) {
	// Initialize PDF with A4 size, portrait orientation
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)

	// ============================================================================
	// HEADER - SURAT JALAN TITLE
	// ============================================================================
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "SURAT JALAN", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// ============================================================================
	// DELIVERY INFO SECTION
	// ============================================================================
	pdf.SetFont("Arial", "", 10)

	// Delivery Number
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "No. Surat Jalan:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, delivery.DeliveryNumber)
	pdf.Ln(6)

	// Delivery Date
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "Tanggal:")
	pdf.SetFont("Arial", "", 10)
	deliveryDate := delivery.DeliveryDate
	if deliveryDate.IsZero() {
		deliveryDate = time.Now()
	}
	pdf.Cell(0, 6, deliveryDate.Format("02 January 2006"))
	pdf.Ln(6)

	// Sales Order Reference
	if delivery.SalesOrder.SONumber != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 6, "No. Sales Order:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, delivery.SalesOrder.SONumber)
		pdf.Ln(6)
	}

	pdf.Ln(5)

	// ============================================================================
	// CUSTOMER INFORMATION
	// ============================================================================
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "KEPADA:")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)

	// Customer Name
	if delivery.Customer.Name != "" {
		pdf.Cell(0, 6, delivery.Customer.Name)
		pdf.Ln(6)
	}

	// Customer Address
	if delivery.DeliveryAddress != nil && *delivery.DeliveryAddress != "" {
		pdf.MultiCell(0, 6, *delivery.DeliveryAddress, "", "", false)
	} else if delivery.Customer.Address != nil {
		pdf.MultiCell(0, 6, *delivery.Customer.Address, "", "", false)
	}

	// Customer Phone
	if delivery.Customer.Phone != nil && *delivery.Customer.Phone != "" {
		pdf.Cell(0, 6, "Telp: "+*delivery.Customer.Phone)
		pdf.Ln(6)
	}

	pdf.Ln(5)

	// ============================================================================
	// WAREHOUSE INFORMATION
	// ============================================================================
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "DARI GUDANG:")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)

	if delivery.Warehouse.Name != "" {
		pdf.Cell(0, 6, delivery.Warehouse.Name)
		pdf.Ln(6)
	}

	if delivery.Warehouse.Address != nil && *delivery.Warehouse.Address != "" {
		pdf.MultiCell(0, 6, *delivery.Warehouse.Address, "", "", false)
	}

	pdf.Ln(5)

	// ============================================================================
	// ITEMS TABLE
	// ============================================================================
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "BARANG YANG DIKIRIM:")
	pdf.Ln(9)

	// Table Header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(10, 8, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(70, 8, "Nama Produk", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Kode", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Unit", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "Batch", "1", 1, "C", true, 0, "")

	// Table Body
	pdf.SetFont("Arial", "", 9)
	for i, item := range delivery.Items {
		// Product Name
		productName := "-"
		if item.Product.Name != "" {
			productName = item.Product.Name
		}

		// Product Code
		productCode := "-"
		if item.Product.Code != "" {
			productCode = item.Product.Code
		}

		// Quantity
		qty := item.Quantity.String()

		// Unit
		unit := "-"
		if item.ProductUnit != nil && item.ProductUnit.UnitName != "" {
			unit = item.ProductUnit.UnitName
		} else if item.Product.BaseUnit != "" {
			unit = item.Product.BaseUnit
		}

		// Batch Number
		batchNo := "-"
		if item.Batch != nil && item.Batch.BatchNumber != "" {
			batchNo = item.Batch.BatchNumber
		}

		// Row number
		pdf.CellFormat(10, 7, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(70, 7, productName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 7, productCode, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 7, qty, "1", 0, "R", false, 0, "")
		pdf.CellFormat(25, 7, unit, "1", 0, "C", false, 0, "")
		pdf.CellFormat(20, 7, batchNo, "1", 1, "C", false, 0, "")
	}

	pdf.Ln(10)

	// ============================================================================
	// DELIVERY METHOD INFO
	// ============================================================================
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, "INFORMASI PENGIRIMAN:")
	pdf.Ln(7)
	pdf.SetFont("Arial", "", 10)

	// Delivery Type
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 6, "Jenis:")
	pdf.SetFont("Arial", "", 10)
	deliveryType := string(delivery.Type)
	if deliveryType == "" {
		deliveryType = "NORMAL"
	}
	pdf.Cell(0, 6, deliveryType)
	pdf.Ln(6)

	// Driver or Expedition
	if delivery.DriverName != nil && *delivery.DriverName != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 6, "Sopir:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, *delivery.DriverName)
		pdf.Ln(6)

		if delivery.VehicleNumber != nil && *delivery.VehicleNumber != "" {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(40, 6, "Nomor Kendaraan:")
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, 6, *delivery.VehicleNumber)
			pdf.Ln(6)
		}
	} else if delivery.ExpeditionService != nil && *delivery.ExpeditionService != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 6, "Ekspedisi:")
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, *delivery.ExpeditionService)
		pdf.Ln(6)

		if delivery.TTNKNumber != nil && *delivery.TTNKNumber != "" {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(40, 6, "No. Resi:")
			pdf.SetFont("Arial", "", 10)
			pdf.Cell(0, 6, *delivery.TTNKNumber)
			pdf.Ln(6)
		}
	}

	// Notes
	if delivery.Notes != nil && *delivery.Notes != "" {
		pdf.Ln(3)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, "Catatan:")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(0, 5, *delivery.Notes, "", "", false)
	}

	pdf.Ln(10)

	// ============================================================================
	// SIGNATURE SECTION
	// ============================================================================
	pdf.SetFont("Arial", "", 10)

	// Left: Pengirim
	leftX := 20.0
	pdf.SetXY(leftX, pdf.GetY())
	pdf.Cell(60, 6, "Pengirim,")

	// Right: Penerima
	rightX := 130.0
	pdf.SetXY(rightX, pdf.GetY())
	pdf.Cell(60, 6, "Penerima,")
	pdf.Ln(20)

	// Signature lines
	pdf.SetXY(leftX, pdf.GetY())
	pdf.Cell(60, 6, "___________________")

	pdf.SetXY(rightX, pdf.GetY())
	pdf.Cell(60, 6, "___________________")
	pdf.Ln(7)

	// Names/Dates
	pdf.SetFont("Arial", "", 9)
	if delivery.DriverName != nil && *delivery.DriverName != "" {
		pdf.SetXY(leftX, pdf.GetY())
		pdf.Cell(60, 5, *delivery.DriverName)
	}

	if delivery.ReceivedBy != nil && *delivery.ReceivedBy != "" {
		pdf.SetXY(rightX, pdf.GetY())
		pdf.Cell(60, 5, *delivery.ReceivedBy)
	}
	pdf.Ln(5)

	// Dates
	pdf.SetFont("Arial", "I", 8)
	if delivery.DepartureTime != nil {
		pdf.SetXY(leftX, pdf.GetY())
		pdf.Cell(60, 5, delivery.DepartureTime.Format("02/01/2006 15:04"))
	}

	if delivery.ReceivedAt != nil {
		pdf.SetXY(rightX, pdf.GetY())
		pdf.Cell(60, 5, delivery.ReceivedAt.Format("02/01/2006 15:04"))
	}

	// ============================================================================
	// FOOTER
	// ============================================================================
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.CellFormat(0, 10, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04:05")), "", 0, "C", false, 0, "")

	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
