/**
 * Generate Excel Test Files from Existing Products
 *
 * This script fetches products from the API and generates Excel test files
 * using actual product codes that exist in the database.
 *
 * Usage: node scripts/generate-excel-from-api.js
 *
 * Prerequisites:
 * 1. Backend server must be running
 * 2. User must be logged in (need access token)
 * 3. At least 3 products must exist in database
 */

const XLSX = require('xlsx');
const fs = require('fs');
const path = require('path');

// Configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const ACCESS_TOKEN = process.env.ACCESS_TOKEN || null; // Pass via env var

// Create test-data directory
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

async function fetchProducts() {
  try {
    console.log('📡 Fetching products from API...');
    console.log(`   URL: ${API_BASE_URL}/api/v1/products?page=1&page_size=10&is_active=true`);

    if (!ACCESS_TOKEN) {
      console.log('\n⚠️  No ACCESS_TOKEN provided.');
      console.log('💡 To fetch from API, set ACCESS_TOKEN environment variable:');
      console.log('   export ACCESS_TOKEN="your-token-here"');
      console.log('   node scripts/generate-excel-from-api.js');
      console.log('\n📝 Using fallback: Will generate Excel with placeholder codes');
      return null;
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/products?page=1&page_size=10&is_active=true`, {
      headers: {
        'Authorization': `Bearer ${ACCESS_TOKEN}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      console.log(`❌ API request failed: ${response.status} ${response.statusText}`);
      return null;
    }

    const data = await response.json();
    console.log(`✅ Fetched ${data.data?.length || 0} products`);

    return data.data || [];
  } catch (error) {
    console.log(`❌ Error fetching products: ${error.message}`);
    return null;
  }
}

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

async function generateFromAPI() {
  const products = await fetchProducts();

  if (!products || products.length < 3) {
    console.log('\n⚠️  Not enough products found in database.');
    console.log('📝 Please create at least 3 products first:');
    console.log('   1. Go to: http://localhost:3000/master/products');
    console.log('   2. Create products with codes: PROD-001, PROD-002, PROD-003');
    console.log('   3. Run this script again\n');
    console.log('🔄 Or use the default test files generated earlier.\n');
    return;
  }

  console.log('\n📦 Products found:');
  products.slice(0, 5).forEach((p, i) => {
    console.log(`   ${i + 1}. ${p.code} - ${p.name} (${p.baseCost || 0})`);
  });

  // Generate test files using real product codes
  const product1 = products[0];
  const product2 = products[1];
  const product3 = products[2];

  // Test Case 1: Valid Data
  const validData = [
    headers,
    [product1.code, 100, product1.baseCost || 5000, 'Rak A-1', 10, 500, `Stok awal ${product1.name}`],
    [product2.code, 50, product2.baseCost || 7500, 'Rak B-2', 5, 200, `Stok awal ${product2.name}`],
    [product3.code, 200, product3.baseCost || 3000, '', 20, 1000, ''],
  ];

  // Test Case 2: Duplicates
  const duplicateData = [
    headers,
    [product1.code, 100, product1.baseCost || 5000, 'Rak A-1', 10, 500, `Stok awal ${product1.name}`],
    [product2.code, 50, product2.baseCost || 7500, 'Rak B-2', 5, 200, `Stok awal ${product2.name}`],
    [product1.code, 75, (product1.baseCost || 5000) + 200, 'Rak A-2', 15, 600, 'DUPLIKAT! Akan ditolak'],
    [product3.code, 200, product3.baseCost || 3000, '', 20, 1000, ''],
  ];

  // Test Case 3: Invalid Data
  const invalidData = [
    headers,
    [product1.code, 100, product1.baseCost || 5000, 'Rak A-1', 10, 500, 'Valid row'],
    ['XXX-999', 50, 7500, 'Rak B-2', 5, 200, 'ERROR: Kode tidak ditemukan'],
    [product2.code, -10, product2.baseCost || 7500, '', 20, 1000, 'ERROR: Quantity negatif'],
    ['', 100, 5000, 'Rak A-1', 10, 500, 'ERROR: Kode produk kosong'],
    [product3.code, '', product3.baseCost || 3000, 'Rak C-1', 25, 800, 'ERROR: Quantity kosong'],
  ];

  console.log('\n🔧 Generating test Excel files with real product codes...\n');

  createExcelFile(validData, 'test-valid-from-api.xlsx');
  createExcelFile(duplicateData, 'test-duplicates-from-api.xlsx');
  createExcelFile(invalidData, 'test-invalid-from-api.xlsx');

  console.log('\n✅ Test files created successfully!');
  console.log(`📁 Location: ${testDataDir}`);
  console.log('\n📋 Generated Files:');
  console.log('  1. test-valid-from-api.xlsx      - Valid data with real product codes');
  console.log('  2. test-duplicates-from-api.xlsx - Duplicate products');
  console.log('  3. test-invalid-from-api.xlsx    - Various validation errors');
  console.log('\n🧪 Use these files for testing!\n');
}

// Main execution
console.log('🚀 Generate Excel from API Products\n');
generateFromAPI().catch(error => {
  console.error('❌ Fatal error:', error);
  process.exit(1);
});
