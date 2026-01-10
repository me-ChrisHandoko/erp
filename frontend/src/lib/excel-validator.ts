/**
 * Excel Validator for Initial Stock Setup
 *
 * Validates Excel file for:
 * 1. Duplicate products within the file
 * 2. Products that already have stock in the warehouse
 * 3. Invalid data (missing required fields, invalid numbers)
 */

import type {
  InitialStockItem,
  InitialStockValidationError,
  StockConflict,
  ExcelValidationResult,
  ProductInfo,
} from "@/types/initial-stock.types";
import type { ProductResponse } from "@/types/product.types";
import type { WarehouseStockResponse } from "@/types/stock.types";

interface ExcelRow {
  row: number;
  productCode: string;
  quantity: string;
  costPerUnit: string;
  location?: string;
  minimumStock?: string;
  maximumStock?: string;
  notes?: string;
}

/**
 * Validate Excel data for initial stock setup
 */
export function validateExcelData(
  excelRows: ExcelRow[],
  products: ProductResponse[],
  existingStocks: WarehouseStockResponse[]
): ExcelValidationResult {
  const errors: InitialStockValidationError[] = [];
  const duplicatesInFile: StockConflict[] = [];
  const existingStocksConflicts: StockConflict[] = [];
  const noStockProducts: ProductInfo[] = [];
  const validItems: InitialStockItem[] = [];

  // Map untuk tracking produk yang sudah muncul
  const seenProducts = new Map<string, number>(); // productCode -> row number

  // Map untuk quick lookup
  const productMap = new Map(
    products.map((p) => [p.code.toUpperCase(), p])
  );
  const stockMap = new Map(
    existingStocks.map((s) => [s.productID, s])
  );

  // Validate each row
  excelRows.forEach((row) => {
    const rowErrors: string[] = [];

    // 1. Required field validation
    if (!row.productCode || row.productCode.trim() === "") {
      errors.push({
        row: row.row,
        field: "productCode",
        message: "Kode produk wajib diisi",
        value: row.productCode,
      });
      return;
    }

    if (!row.quantity || row.quantity.trim() === "") {
      errors.push({
        row: row.row,
        field: "quantity",
        message: "Quantity wajib diisi",
        value: row.quantity,
      });
      return;
    }

    if (!row.costPerUnit || row.costPerUnit.trim() === "") {
      errors.push({
        row: row.row,
        field: "costPerUnit",
        message: "Harga beli wajib diisi",
        value: row.costPerUnit,
      });
      return;
    }

    // 2. Find product
    const productCodeUpper = row.productCode.toUpperCase();
    const product = productMap.get(productCodeUpper);

    if (!product) {
      errors.push({
        row: row.row,
        field: "productCode",
        message: `Produk dengan kode "${row.productCode}" tidak ditemukan di sistem`,
        value: row.productCode,
      });
      return;
    }

    // 3. Check duplicate in file
    if (seenProducts.has(product.id)) {
      const firstRow = seenProducts.get(product.id)!;
      duplicatesInFile.push({
        productId: product.id,
        productCode: product.code,
        productName: product.name,
        currentQuantity: "0", // Not applicable for file duplicates
        newQuantity: row.quantity,
        currentCost: "0",
        newCost: row.costPerUnit,
        row: row.row,
      });

      // Don't add to errors array - will be displayed in duplicatesInFile alert
      return;
    }

    // Mark as seen
    seenProducts.set(product.id, row.row);

    // 4. Check if product already has stock in warehouse
    const existingStock = stockMap.get(product.id);
    console.log(`  - Checking ${product.code} (${product.id}):`);
    console.log(`    - stockMap.has(${product.id}):`, stockMap.has(product.id));
    console.log(`    - stockMap.get result:`, existingStock);
    console.log(`    - existing stock =`, existingStock ? `${existingStock.quantity}` : "none");

    if (existingStock) {
      console.log(`    ⚠️ CONFLICT DETECTED: ${product.code} already has ${existingStock.quantity} in stock`);
      existingStocksConflicts.push({
        productId: product.id,
        productCode: product.code,
        productName: product.name,
        currentQuantity: existingStock.quantity,
        newQuantity: row.quantity,
        currentCost: "0", // Cost not tracked in stock table
        newCost: row.costPerUnit,
        row: row.row,
      });

      // Don't add to errors array - will be displayed in existingStocks alert
      return;
    } else {
      // Product doesn't have stock in this warehouse - valid for initial stock input
      noStockProducts.push({
        productCode: product.code,
        productName: product.name,
        quantity: row.quantity,
        row: row.row,
      });
    }

    // 5. Validate numeric fields
    const quantity = parseFloat(row.quantity);
    if (isNaN(quantity) || quantity <= 0) {
      errors.push({
        row: row.row,
        field: "quantity",
        message: "Quantity harus berupa angka positif",
        value: row.quantity,
      });
      rowErrors.push("quantity");
    }

    const cost = parseFloat(row.costPerUnit);
    if (isNaN(cost) || cost <= 0) {
      errors.push({
        row: row.row,
        field: "costPerUnit",
        message: "Harga beli harus berupa angka positif",
        value: row.costPerUnit,
      });
      rowErrors.push("costPerUnit");
    }

    // Optional fields validation
    if (row.minimumStock && row.minimumStock.trim() !== "") {
      const minStock = parseFloat(row.minimumStock);
      if (isNaN(minStock) || minStock < 0) {
        errors.push({
          row: row.row,
          field: "minimumStock",
          message: "Stok minimum harus berupa angka positif atau 0",
          value: row.minimumStock,
        });
        rowErrors.push("minimumStock");
      }
    }

    if (row.maximumStock && row.maximumStock.trim() !== "") {
      const maxStock = parseFloat(row.maximumStock);
      if (isNaN(maxStock) || maxStock < 0) {
        errors.push({
          row: row.row,
          field: "maximumStock",
          message: "Stok maksimum harus berupa angka positif atau 0",
          value: row.maximumStock,
        });
        rowErrors.push("maximumStock");
      }
    }

    // If no errors for this row, add to valid items
    if (rowErrors.length === 0) {
      validItems.push({
        productId: product.id,
        quantity: row.quantity,
        costPerUnit: row.costPerUnit,
        location: row.location || undefined,
        minimumStock: row.minimumStock || undefined,
        maximumStock: row.maximumStock || undefined,
        notes: row.notes || undefined,
      });
    }
  });

  // Determine success
  const success =
    errors.length === 0 &&
    duplicatesInFile.length === 0 &&
    existingStocksConflicts.length === 0;

  let message = "";
  if (!success) {
    if (duplicatesInFile.length > 0) {
      message = `Ditemukan ${duplicatesInFile.length} produk duplikat dalam file.`;
    } else if (existingStocksConflicts.length > 0) {
      message = `${existingStocksConflicts.length} produk sudah memiliki stok di gudang.`;
    } else if (errors.length > 0) {
      message = `Ditemukan ${errors.length} error dalam file.`;
    }
  }

  return {
    success,
    message,
    duplicatesInFile,
    existingStocks: existingStocksConflicts,
    noStockProducts,
    validItems,
    errors,
  };
}

/**
 * Parse Excel file using xlsx library
 * Expected Excel columns: Kode Produk, Quantity, Harga Beli, Lokasi, Stok Min, Stok Max, Catatan
 */
export async function parseExcelFile(file: File): Promise<ExcelRow[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = async (e) => {
      try {
        // Dynamically import xlsx to reduce bundle size
        const XLSX = await import("xlsx");
        const data = e.target?.result;
        const workbook = XLSX.read(data, { type: "binary" });

        // Get first sheet
        const sheetName = workbook.SheetNames[0];
        if (!sheetName) {
          reject(new Error("File Excel kosong atau tidak memiliki sheet."));
          return;
        }

        const worksheet = workbook.Sheets[sheetName];

        // Convert to JSON with header mapping
        const jsonData = XLSX.utils.sheet_to_json<any>(worksheet, {
          header: 1, // Get raw data first
          defval: "", // Default empty value
        });

        if (jsonData.length < 2) {
          reject(new Error("File Excel harus memiliki minimal 2 baris (header + data)."));
          return;
        }

        // Parse header row to find column indices
        const headerRow = jsonData[0] as string[];
        const colIndices = {
          productCode: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("kode") || h?.toLowerCase().includes("code")
          ),
          quantity: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("qty") || h?.toLowerCase().includes("quantity") || h?.toLowerCase().includes("jumlah")
          ),
          costPerUnit: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("harga") || h?.toLowerCase().includes("cost") || h?.toLowerCase().includes("beli")
          ),
          location: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("lokasi") || h?.toLowerCase().includes("location")
          ),
          minimumStock: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("min") || h?.toLowerCase().includes("minimum")
          ),
          maximumStock: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("max") || h?.toLowerCase().includes("maksimum")
          ),
          notes: headerRow.findIndex((h) =>
            h?.toLowerCase().includes("catatan") || h?.toLowerCase().includes("notes") || h?.toLowerCase().includes("keterangan")
          ),
        };

        // Validate required columns exist
        if (colIndices.productCode === -1) {
          reject(new Error("Kolom 'Kode Produk' tidak ditemukan di Excel."));
          return;
        }
        if (colIndices.quantity === -1) {
          reject(new Error("Kolom 'Quantity' tidak ditemukan di Excel."));
          return;
        }
        if (colIndices.costPerUnit === -1) {
          reject(new Error("Kolom 'Harga Beli' tidak ditemukan di Excel."));
          return;
        }

        // Parse data rows (skip header)
        const excelRows: ExcelRow[] = [];
        for (let i = 1; i < jsonData.length; i++) {
          const dataRow = jsonData[i] as any[];

          // Skip empty rows
          if (!dataRow || dataRow.every((cell) => !cell)) {
            continue;
          }

          excelRows.push({
            row: i + 1, // Excel row number (1-based, including header)
            productCode: String(dataRow[colIndices.productCode] || "").trim(),
            quantity: String(dataRow[colIndices.quantity] || "").trim(),
            costPerUnit: String(dataRow[colIndices.costPerUnit] || "").trim(),
            location: colIndices.location !== -1 ? String(dataRow[colIndices.location] || "").trim() : undefined,
            minimumStock: colIndices.minimumStock !== -1 ? String(dataRow[colIndices.minimumStock] || "").trim() : undefined,
            maximumStock: colIndices.maximumStock !== -1 ? String(dataRow[colIndices.maximumStock] || "").trim() : undefined,
            notes: colIndices.notes !== -1 ? String(dataRow[colIndices.notes] || "").trim() : undefined,
          });
        }

        if (excelRows.length === 0) {
          reject(new Error("Tidak ada data yang valid di file Excel."));
          return;
        }

        resolve(excelRows);
      } catch (error) {
        reject(new Error(`Gagal membaca file Excel: ${error instanceof Error ? error.message : "Unknown error"}`));
      }
    };

    reader.onerror = () => {
      reject(new Error("Gagal membaca file."));
    };

    reader.readAsBinaryString(file);
  });
}

/**
 * Generate Excel template for download
 * Creates a template with headers and sample data
 */
export async function generateExcelTemplate(): Promise<Blob> {
  // Dynamically import xlsx to reduce bundle size
  const XLSX = await import("xlsx");

  // Create workbook
  const workbook = XLSX.utils.book_new();

  // Define headers
  const headers = [
    "Kode Produk",
    "Quantity",
    "Harga Beli",
    "Lokasi",
    "Stok Minimum",
    "Stok Maksimum",
    "Catatan",
  ];

  // Sample data rows
  const sampleData = [
    ["PROD-001", "100", "5000", "Rak A-1", "10", "500", "Stok awal produk A"],
    ["PROD-002", "50", "7500", "Rak B-2", "5", "200", ""],
    ["PROD-003", "200", "3000", "", "20", "1000", ""],
  ];

  // Combine headers and sample data
  const worksheetData = [headers, ...sampleData];

  // Create worksheet
  const worksheet = XLSX.utils.aoa_to_sheet(worksheetData);

  // Set column widths
  worksheet["!cols"] = [
    { wch: 15 }, // Kode Produk
    { wch: 10 }, // Quantity
    { wch: 12 }, // Harga Beli
    { wch: 12 }, // Lokasi
    { wch: 15 }, // Stok Minimum
    { wch: 15 }, // Stok Maksimum
    { wch: 30 }, // Catatan
  ];

  // Add worksheet to workbook
  XLSX.utils.book_append_sheet(workbook, worksheet, "Template Stok Awal");

  // Generate buffer
  const excelBuffer = XLSX.write(workbook, { bookType: "xlsx", type: "array" });

  // Create blob
  return new Blob([excelBuffer], {
    type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  });
}
