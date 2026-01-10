/**
 * Initial Stock Setup Types
 *
 * Types for one-time initial stock setup in warehouses.
 * Used when setting up a new warehouse or migrating data.
 */

export interface InitialStockSetupRequest {
  warehouseId: string;
  items: InitialStockItem[];
  notes?: string;
}

export interface InitialStockItem {
  productId: string;
  quantity: string; // decimal as string
  costPerUnit: string; // decimal as string
  location?: string;
  minimumStock?: string;
  maximumStock?: string;
  notes?: string;
}

export interface InitialStockValidationError {
  row: number;
  field: string;
  message: string;
  value?: any;
}

export interface InitialStockImportResult {
  totalRows: number;
  validRows: number;
  errorRows: number;
  errors: InitialStockValidationError[];
  validItems: InitialStockItem[];
}

export interface StockConflict {
  productId: string;
  productCode: string;
  productName: string;
  currentQuantity: string;
  newQuantity: string;
  currentCost: string;
  newCost: string;
  row: number;
}

export interface ProductInfo {
  productCode: string;
  productName: string;
  quantity: string;
  row: number;
}

export interface ExcelValidationResult {
  success: boolean;
  message?: string;
  duplicatesInFile: StockConflict[]; // Duplikasi dalam Excel
  existingStocks: StockConflict[]; // Produk sudah ada di gudang
  noStockProducts: ProductInfo[]; // Produk yang belum memiliki stock (valid untuk input)
  validItems: InitialStockItem[];
  errors: InitialStockValidationError[];
}

export interface WarehouseStockStatus {
  warehouseId: string;
  warehouseName: string;
  warehouseCode?: string;
  hasInitialStock: boolean;
  totalProducts: number;
  totalValue: string;
  lastUpdated?: string;
}

export interface InitialStockSetupResponse {
  success: boolean;
  message: string;
  totalItems: number;
  totalValue: string;
  createdStocks: number;
  updatedStocks: number;
}
