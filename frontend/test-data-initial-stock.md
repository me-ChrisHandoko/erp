# Test Data untuk Initial Stock Upload

## Test Case 1: Valid Data (Should SUCCESS)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
| PROD-001    | 100      | 5000       | Rak A-1 | 10           | 500           | Stok awal produk A   |
| PROD-002    | 50       | 7500       | Rak B-2 | 5            | 200           | Stok awal produk B   |
| PROD-003    | 200      | 3000       |         | 20           | 1000          |                      |
```

## Test Case 2: Duplicate dalam File (Should REJECT)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
| PROD-001    | 100      | 5000       | Rak A-1 | 10           | 500           | Stok awal produk A   |
| PROD-002    | 50       | 7500       | Rak B-2 | 5            | 200           | Stok awal produk B   |
| PROD-001    | 75       | 5200       | Rak A-2 | 15           | 600           | Duplikat! PROD-001   |
```
**Expected Error:** "Produk PROD-001 sudah ada di baris 2. Hapus duplikasi atau gabungkan quantity."

## Test Case 3: Product Not Found (Should REJECT)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
| PROD-001    | 100      | 5000       | Rak A-1 | 10           | 500           | Valid                |
| XXX-999     | 50       | 7500       | Rak B-2 | 5            | 200           | Kode tidak ada!      |
```
**Expected Error:** "Produk dengan kode \"XXX-999\" tidak ditemukan di sistem"

## Test Case 4: Invalid Quantity (Should REJECT)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
| PROD-001    | -50      | 5000       | Rak A-1 | 10           | 500           | Quantity negatif!    |
| PROD-002    | abc      | 7500       | Rak B-2 | 5            | 200           | Quantity bukan angka!|
| PROD-003    | 0        | 3000       |         | 20           | 1000          | Quantity = 0         |
```
**Expected Errors:**
- "Quantity harus berupa angka positif"
- Row dengan quantity <= 0 akan ditolak

## Test Case 5: Missing Required Fields (Should REJECT)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
|             | 100      | 5000       | Rak A-1 | 10           | 500           | Kode produk kosong   |
| PROD-002    |          | 7500       | Rak B-2 | 5            | 200           | Quantity kosong      |
| PROD-003    | 200      |            |         | 20           | 1000          | Harga beli kosong    |
```
**Expected Errors:**
- "Kode produk wajib diisi"
- "Quantity wajib diisi"
- "Harga beli wajib diisi"

## Test Case 6: Product Already Has Stock (Should REJECT if warehouse has stock)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
| PROD-001    | 100      | 5000       | Rak A-1 | 10           | 500           | Produk sudah ada stok|
```
**Expected Error:** "Produk PROD-001 sudah memiliki stok di gudang (50 pcs)"

## Test Case 7: Mixed Valid/Invalid (Should show validation errors)
```
| Kode Produk | Quantity | Harga Beli | Lokasi  | Stok Minimum | Stok Maksimum | Catatan              |
|-------------|----------|------------|---------|--------------|---------------|----------------------|
| PROD-001    | 100      | 5000       | Rak A-1 | 10           | 500           | Valid ✅             |
| XXX-999     | 50       | 7500       | Rak B-2 | 5            | 200           | Invalid - not found  |
| PROD-003    | -10      | 3000       |         | 20           | 1000          | Invalid - negative   |
| PROD-004    | 200      | 8000       | Rak C-1 | 25           | 800           | Valid ✅             |
```
**Expected:** Show errors for row 3 (not found) and row 4 (negative), but highlight valid rows

## Steps to Test:

1. **Navigate to Initial Stock Setup**
   - Login ke aplikasi
   - Pilih company
   - Buka menu Inventory > Initial Stock Setup

2. **Select Warehouse**
   - Pilih gudang yang belum memiliki stok (jika sudah ada, buat gudang baru dulu)

3. **Choose Excel Import Method**
   - Pilih "Import dari Excel"

4. **Download Template**
   - Klik "Download Template"
   - Buka file Excel yang di-download

5. **Fill Test Data**
   - Copy paste salah satu test case di atas
   - Save Excel file

6. **Upload File**
   - Click "Pilih File" button
   - Select Excel file yang sudah diisi
   - Tunggu proses parsing dan validation

7. **Verify Results**
   - **Valid Data**: Show success message "X produk berhasil divalidasi"
   - **Invalid Data**: Show error alerts dengan detail row dan error message
   - **Duplicates**: Show duplicate conflicts dengan row numbers
   - **Existing Stock**: Show existing stock conflicts

8. **Proceed to Review**
   - Jika validasi sukses, click "Lanjutkan" ke Step 4 Review
   - Verify data tampil dengan benar
   - Submit untuk simpan ke database

## Expected Behaviors:

### ✅ Success Scenario:
- File parsed successfully
- No validation errors
- Show green success alert
- Data populated in review step
- Can proceed to submit

### ❌ Error Scenarios:

#### Duplicate in File:
```
Alert: "Produk Duplikat dalam File (1)"
• Baris 4: PROD-001 - Product Name A
```

#### Product Already Has Stock:
```
Alert: "Produk Sudah Memiliki Stok (1)"
• Baris 2: PROD-002 - Product Name B
  Stok saat ini: 50 | Stok baru: 100
```

#### Validation Errors:
```
Alert: "Error Validasi (3)"
• Baris 2, Kolom productCode: Produk dengan kode "XXX-999" tidak ditemukan
• Baris 3, Kolom quantity: Quantity harus berupa angka positif
• Baris 5, Kolom costPerUnit: Harga beli wajib diisi
```

## Notes:
- Pastikan produk dengan kode PROD-001, PROD-002, PROD-003, PROD-004 sudah ada di master produk
- Jika belum ada, buat dulu atau gunakan kode produk yang sudah ada
- Excel file harus format .xlsx atau .xls
- Maksimal ukuran file 5MB
