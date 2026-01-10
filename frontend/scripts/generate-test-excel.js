/**
 * Generate Excel Test Files for Initial Stock Setup
 *
 * Usage: node scripts/generate-test-excel.js
 *
 * This will create test Excel files in test-data/ directory:
 * - test-valid.xlsx (valid data)
 * - test-duplicates.xlsx (duplicate products)
 * - test-invalid.xlsx (invalid data)
 */

const XLSX = require('xlsx');
const fs = require('fs');
const path = require('path');

// Create test-data directory if not exists
const testDataDir = path.join(__dirname, '..', 'test-data');
if (!fs.existsSync(testDataDir)) {
  fs.mkdirSync(testDataDir, { recursive: true });
}

// Headers
const headers = [
  'Kode Produk',
  'Quantity',
  'Harga Beli',
  'Lokasi',
  'Stok Minimum',
  'Stok Maksimum',
  'Catatan'
];

// Test Case 1: Valid Data
const validData = [
  headers,
  ['BRS-001', 100, 5000, 'Rak A-1', 10, 500, 'Stok awal produk A'],
  ['BRS-002', 50, 7500, 'Rak B-2', 5, 200, 'Stok awal produk B'],
  ['BRS-003', 200, 3000, '', 20, 1000, ''],
];

// Test Case 2: Duplicates
const duplicateData = [
  headers,
  ['BRS-001', 100, 5000, 'Rak A-1', 10, 500, 'Stok awal produk A'],
  ['BRS-002', 50, 7500, 'Rak B-2', 5, 200, 'Stok awal produk B'],
  ['BRS-001', 75, 5200, 'Rak A-2', 15, 600, 'DUPLIKAT! Akan ditolak'],
  ['BRS-003', 200, 3000, '', 20, 1000, ''],
];

// Test Case 3: Invalid Data (various errors)
const invalidData = [
  headers,
  ['BRS-001', 100, 5000, 'Rak A-1', 10, 500, 'Valid row'],
  ['XXX-999', 50, 7500, 'Rak B-2', 5, 200, 'ERROR: Kode tidak ditemukan'],
  ['BRS-003', -10, 3000, '', 20, 1000, 'ERROR: Quantity negatif'],
  ['', 100, 5000, 'Rak A-1', 10, 500, 'ERROR: Kode produk kosong'],
  ['BRS-004', '', 8000, 'Rak C-1', 25, 800, 'ERROR: Quantity kosong'],
  ['BRS-005', 150, '', '', 30, 700, 'ERROR: Harga beli kosong'],
  ['BRS-006', 'abc', 4000, 'Rak D-1', 5, 300, 'ERROR: Quantity bukan angka'],
];

// Test Case 4: Mixed (valid + invalid)
const mixedData = [
  headers,
  ['BRS-001', 100, 5000, 'Rak A-1', 10, 500, 'Valid ✅'],
  ['XXX-999', 50, 7500, 'Rak B-2', 5, 200, 'Invalid - not found ❌'],
  ['BRS-003', -10, 3000, '', 20, 1000, 'Invalid - negative ❌'],
  ['BRS-004', 200, 8000, 'Rak C-1', 25, 800, 'Valid ✅'],
  ['BRS-005', 150, 4500, 'Rak D-1', 15, 600, 'Valid ✅'],
];

// Test Case 5: Large dataset (performance test)
const largeData = [headers];
for (let i = 1; i <= 100; i++) {
  const productCode = `BRS-${String(i).padStart(3, '0')}`;
  largeData.push([
    productCode,
    Math.floor(Math.random() * 500) + 50, // 50-550
    Math.floor(Math.random() * 10000) + 1000, // 1000-11000
    i % 5 === 0 ? `Rak ${String.fromCharCode(65 + Math.floor(i / 20))}-${i % 20}` : '',
    Math.floor(Math.random() * 20) + 5, // 5-25
    Math.floor(Math.random() * 500) + 200, // 200-700
    i % 10 === 0 ? `Batch test ${i}` : '',
  ]);
}

// Function to create Excel file
function createExcelFile(data, filename) {
  const ws = XLSX.utils.aoa_to_sheet(data);

  // Set column widths
  ws['!cols'] = [
    { wch: 15 }, // Kode Produk
    { wch: 10 }, // Quantity
    { wch: 12 }, // Harga Beli
    { wch: 12 }, // Lokasi
    { wch: 15 }, // Stok Minimum
    { wch: 15 }, // Stok Maksimum
    { wch: 30 }, // Catatan
  ];

  const wb = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(wb, ws, 'Initial Stock');

  const filepath = path.join(testDataDir, filename);
  XLSX.writeFile(wb, filepath);

  console.log(`✅ Created: ${filename}`);
  return filepath;
}

// Generate all test files
console.log('🔧 Generating test Excel files...\n');

try {
  createExcelFile(validData, 'test-valid.xlsx');
  createExcelFile(duplicateData, 'test-duplicates.xlsx');
  createExcelFile(invalidData, 'test-invalid.xlsx');
  createExcelFile(mixedData, 'test-mixed.xlsx');
  createExcelFile(largeData, 'test-large-100-products.xlsx');

  console.log('\n✅ All test files created successfully!');
  console.log(`📁 Location: ${testDataDir}`);
  console.log('\n📋 Test Files:');
  console.log('  1. test-valid.xlsx           - Valid data (should pass)');
  console.log('  2. test-duplicates.xlsx      - Duplicate products (should reject)');
  console.log('  3. test-invalid.xlsx         - Various validation errors');
  console.log('  4. test-mixed.xlsx           - Mix of valid and invalid data');
  console.log('  5. test-large-100-products.xlsx - Performance test (100 products)');
  console.log('\n🧪 Test these files in Initial Stock Setup page!');

} catch (error) {
  console.error('❌ Error generating test files:', error);
  process.exit(1);
}
