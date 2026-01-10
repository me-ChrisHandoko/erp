# Test Excel Upload - Initial Stock Setup

## 📁 Test Files Location

Test Excel files sudah dibuat di: `test-data/`

```
test-data/
├── test-valid.xlsx              ✅ Valid data (3 products)
├── test-duplicates.xlsx         ❌ Duplicate PROD-001
├── test-invalid.xlsx            ❌ Multiple validation errors
├── test-mixed.xlsx              ⚠️  Mix valid + invalid data
└── test-large-100-products.xlsx 📊 Performance test (100 products)
```

## 🚀 Steps to Test

### 1. Ensure Products Exist

**IMPORTANT:** Test files menggunakan kode produk PROD-001 sampai PROD-006 (dan PROD-001 sampai PROD-100 untuk large test).

Pastikan produk-produk ini sudah ada di master produk, atau buat dulu:

```bash
# Navigate to Products page
http://localhost:3000/master/products

# Create products:
- Code: PROD-001, Name: Product A, Base Unit: pcs, Base Cost: 5000
- Code: PROD-002, Name: Product B, Base Unit: pcs, Base Cost: 7500
- Code: PROD-003, Name: Product C, Base Unit: pcs, Base Cost: 3000
- Code: PROD-004, Name: Product D, Base Unit: pcs, Base Cost: 8000
- Code: PROD-005, Name: Product E, Base Unit: pcs, Base Cost: 4500
- Code: PROD-006, Name: Product F, Base Unit: pcs, Base Cost: 6000
```

### 2. Access Initial Stock Setup

```bash
# Start dev server
npm run dev

# Navigate to
http://localhost:3000/inventory/initial-setup
```

### 3. Select Warehouse

**Step 1: Warehouse Selection**
- Pilih gudang yang **belum memiliki stok**
- Jika semua gudang sudah punya stok, buat gudang baru dulu di `/master/warehouses`
- Click "Lanjutkan"

### 4. Choose Excel Import Method

**Step 2: Input Method**
- Pilih "Import dari Excel"
- Click "Lanjutkan"

### 5. Test Each Scenario

---

## 📋 Test Case 1: Valid Data ✅

**File:** `test-valid.xlsx`

**Data:**
```
PROD-001, 100, 5000, Rak A-1, 10, 500, Stok awal produk A
PROD-002, 50, 7500, Rak B-2, 5, 200, Stok awal produk B
PROD-003, 200, 3000, (empty), 20, 1000, (empty)
```

**Expected Result:**
- ✅ **Success Message**: "Validasi Berhasil - 3 produk berhasil divalidasi dan siap untuk disimpan"
- ✅ No error alerts
- ✅ Can proceed to Review step
- ✅ Data populated correctly in review table

**Steps:**
1. Click "Pilih File"
2. Select `test-valid.xlsx`
3. Wait for parsing (~1 second)
4. Verify success alert appears
5. Click "Lanjutkan" to Step 4
6. Verify all 3 products in review table
7. Check calculations (total items, total quantity, total value)
8. Click "Simpan" to submit

**What to Verify:**
- ✅ File info shows: `test-valid.xlsx (X KB)`
- ✅ Green success alert with count
- ✅ No red error alerts
- ✅ Review table shows 3 rows
- ✅ Quantities: 100, 50, 200
- ✅ Cost per unit: 5000, 7500, 3000
- ✅ Locations: Rak A-1, Rak B-2, (empty)

---

## 📋 Test Case 2: Duplicates ❌

**File:** `test-duplicates.xlsx`

**Data:**
```
PROD-001, 100, 5000, Rak A-1, 10, 500, Stok awal produk A
PROD-002, 50, 7500, Rak B-2, 5, 200, Stok awal produk B
PROD-001, 75, 5200, Rak A-2, 15, 600, DUPLIKAT! Akan ditolak  ← Row 4
PROD-003, 200, 3000, (empty), 20, 1000, (empty)
```

**Expected Result:**
- ❌ **Red Alert**: "Produk Duplikat dalam File (1)"
- ❌ Shows: "• Baris 4: PROD-001 - Product A"
- ❌ Error message: "Produk PROD-001 sudah ada di baris 2. Hapus duplikasi atau gabungkan quantity."
- ❌ Cannot proceed to next step

**Steps:**
1. Click "Pilih File"
2. Select `test-duplicates.xlsx`
3. Wait for validation
4. Verify red error alert appears

**What to Verify:**
- ❌ Red "destructive" alert
- ❌ Title: "Produk Duplikat dalam File (1)"
- ❌ Shows product code and row number
- ❌ "Lanjutkan" button is disabled or shows validation error

---

## 📋 Test Case 3: Invalid Data ❌

**File:** `test-invalid.xlsx`

**Contains Multiple Errors:**
```
Row 2: PROD-001, 100, 5000, ...     ✅ Valid
Row 3: XXX-999, 50, 7500, ...       ❌ Product not found
Row 4: PROD-003, -10, 3000, ...     ❌ Negative quantity
Row 5: (empty), 100, 5000, ...      ❌ Missing product code
Row 6: PROD-004, (empty), 8000, ... ❌ Missing quantity
Row 7: PROD-005, 150, (empty), ...  ❌ Missing cost
Row 8: PROD-006, 'abc', 4000, ...   ❌ Invalid quantity (not a number)
```

**Expected Result:**
- ❌ Multiple red error alerts
- ❌ **Alert 1**: "Produk Duplikat" (if any)
- ❌ **Alert 2**: "Produk Sudah Memiliki Stok" (if any)
- ❌ **Alert 3**: "Error Validasi (X)"
  - Shows first 5 errors
  - Each with: Row number, Column, Error message
- ❌ Cannot proceed

**Steps:**
1. Upload `test-invalid.xlsx`
2. Verify multiple error alerts

**What to Verify:**
- ❌ Shows error count: "Error Validasi (6)" or similar
- ❌ Lists errors with row numbers:
  - "Baris 3, Kolom productCode: Produk dengan kode \"XXX-999\" tidak ditemukan"
  - "Baris 4, Kolom quantity: Quantity harus berupa angka positif"
  - "Baris 5, Kolom productCode: Kode produk wajib diisi"
  - "Baris 6, Kolom quantity: Quantity wajib diisi"
  - "Baris 7, Kolom costPerUnit: Harga beli wajib diisi"
- ❌ Shows "... dan X error lainnya" if > 5 errors

---

## 📋 Test Case 4: Mixed Valid/Invalid ⚠️

**File:** `test-mixed.xlsx`

**Data:**
```
Row 2: PROD-001, 100, 5000, ... ✅ Valid
Row 3: XXX-999, 50, 7500, ...   ❌ Not found
Row 4: PROD-003, -10, 3000, ... ❌ Negative
Row 5: PROD-004, 200, 8000, ... ✅ Valid
Row 6: PROD-005, 150, 4500, ... ✅ Valid
```

**Expected Result:**
- ❌ Shows validation errors for rows 3 and 4
- ❌ Total errors: 2
- ⚠️  Valid items (rows 2, 5, 6) should NOT be in validItems array because validation failed

**Steps:**
1. Upload `test-mixed.xlsx`
2. Verify error alerts

**What to Verify:**
- ❌ "Error Validasi (2)"
- ❌ Shows row 3 and row 4 errors
- ⚠️  Valid rows are NOT added (strict validation mode)

---

## 📋 Test Case 5: Large Dataset 📊

**File:** `test-large-100-products.xlsx`

**Contains:** 100 random products (PROD-001 to PROD-100)

**Expected Result:**
- ✅ Should process within 2-3 seconds
- ✅ Success message: "100 produk berhasil divalidasi"
- ✅ Review step shows paginated table
- 📊 Performance metrics in console

**Steps:**
1. Upload `test-large-100-products.xlsx`
2. Watch console for performance logs
3. Verify parsing speed

**What to Verify:**
- ⚡ Parsing time < 3 seconds
- ✅ All 100 products validated
- 📊 Review table shows correctly
- 📊 Calculations are accurate

---

## 📋 Test Case 6: Product Already Has Stock ❌

**Manual Test - Requires Existing Stock**

**Steps:**
1. First, run Test Case 1 (valid data) and submit successfully
2. Try to upload the same file again to the same warehouse
3. Should be rejected

**Expected Result:**
- ❌ **Red Alert**: "Produk Sudah Memiliki Stok (3)"
- ❌ Shows all 3 products with current quantities:
  - "• Baris 2: PROD-001 - Product A"
  - "  Stok saat ini: 100 | Stok baru: 100"
- ❌ Cannot proceed

---

## 📋 Test Case 7: Invalid File Type ❌

**Manual Test**

**Steps:**
1. Try to upload a .txt or .pdf file
2. Should be rejected

**Expected Result:**
- ❌ **Red Alert**: "Format file harus .xlsx atau .xls"

---

## 📋 Test Case 8: File Too Large ❌

**Manual Test**

**Steps:**
1. Try to upload a file > 5MB
2. Should be rejected

**Expected Result:**
- ❌ **Red Alert**: "Ukuran file maksimal 5MB"

---

## 🔍 Additional Validations to Test

### Column Detection
- ✅ Template dengan header Indonesian ("Kode Produk", "Quantity", "Harga Beli")
- ✅ Template dengan header English ("Product Code", "Quantity", "Cost")
- ❌ Template tanpa header yang benar → Should show error

### Empty Rows
- ✅ Excel dengan empty rows → Should skip automatically
- ✅ Only count rows with data

### Number Formats
- ✅ Quantity with decimals: 100.5 → Should accept
- ✅ Cost with thousands separator: 5,000 → Should parse correctly
- ❌ Quantity with text: "100 pcs" → Should reject

---

## 🎯 Expected UI Behaviors

### File Upload Area
- **Before upload:** Shows "Upload File Excel" with cloud icon
- **During upload:** Shows "Memproses..." with spinner
- **After upload:** Shows filename and size

### Success State
- Green alert with CheckCircle icon
- Message: "Validasi Berhasil"
- Count: "X produk berhasil divalidasi dan siap untuk disimpan"
- "Lanjutkan" button enabled

### Error States

#### Duplicate Alert (Red)
```
🔴 Produk Duplikat dalam File (1)
• Baris 4: PROD-001 - Product Name
```

#### Existing Stock Alert (Red)
```
🔴 Produk Sudah Memiliki Stok (2)
• Baris 2: PROD-001 - Product Name
  Stok saat ini: 50 | Stok baru: 100
• Baris 3: PROD-002 - Product Name
  Stok saat ini: 75 | Stok baru: 50
... dan 0 lainnya
```

#### Validation Errors Alert (Red)
```
🔴 Error Validasi (5)
• Baris 2, Kolom productCode: Produk dengan kode "XXX-999" tidak ditemukan
• Baris 3, Kolom quantity: Quantity harus berupa angka positif
• Baris 4, Kolom productCode: Kode produk wajib diisi
• Baris 5, Kolom quantity: Quantity wajib diisi
• Baris 6, Kolom costPerUnit: Harga beli wajib diisi
... dan 0 error lainnya
```

---

## ✅ Test Checklist

- [ ] Test Case 1: Valid data passes validation
- [ ] Test Case 2: Duplicates detected and rejected
- [ ] Test Case 3: Invalid data shows all errors
- [ ] Test Case 4: Mixed data handles correctly
- [ ] Test Case 5: Large dataset performs well
- [ ] Test Case 6: Existing stock detected
- [ ] Test Case 7: Invalid file type rejected
- [ ] Test Case 8: Large file rejected
- [ ] Column headers detected (ID/EN)
- [ ] Empty rows skipped
- [ ] Number formats parsed correctly
- [ ] Error messages clear and actionable
- [ ] Success flow works end-to-end
- [ ] Review step shows correct data
- [ ] Submit creates stock records

---

## 🐛 Known Issues / Limitations

1. ⚠️ **Backend endpoint missing:** `/warehouses/stock-status` returns 404
   - Impact: No visual indicator which warehouses have stock
   - Workaround: Validation still works via `existingStocksData`

2. ⚠️ **Existing stock detection:** Only works if `useListStocksQuery` returns data
   - Make sure stock API is working before testing existing stock scenario

3. 📝 **Template generation:** Currently generates client-side
   - Could be moved to backend for consistency with backend validation

---

## 📊 Performance Benchmarks

Expected performance on modern machine:

| Test Case | Products | File Size | Parse Time | Validation Time | Total Time |
|-----------|----------|-----------|------------|-----------------|------------|
| Valid     | 3        | ~5 KB     | <100ms     | <50ms           | <200ms     |
| Large     | 100      | ~15 KB    | <500ms     | <200ms          | <1s        |
| Large     | 500      | ~75 KB    | <2s        | <1s             | <3s        |
| Large     | 1000     | ~150 KB   | <4s        | <2s             | <6s        |

---

## 🔧 Troubleshooting

### Upload tidak berfungsi
- Cek console untuk errors
- Verify `xlsx` library installed: `npm list xlsx`
- Cek file permissions

### Validation selalu fail
- Verify products exist in master data
- Check product codes match exactly (case-sensitive)
- Verify warehouse selected

### Slow performance
- Check file size (should be < 5MB)
- Verify not running other heavy processes
- Try smaller dataset first

---

## 📝 Next Steps

After successful upload test:

1. ✅ Verify data in Review step
2. ✅ Submit to backend
3. ✅ Check warehouse_stocks table in database
4. ✅ Verify inventory_movements created
5. ✅ Check stock appears in Stock List page
6. 🔄 Try uploading to same warehouse again (should reject)
