# Quick Setup - Create Test Products

## Steps to Create Test Products

1. **Navigate to Products Page**
   ```
   http://localhost:3000/master/products
   ```

2. **Click "Tambah Produk" button**

3. **Create each product with these details:**

### Product 1
- **Kode Produk**: PROD-001
- **Nama Produk**: Product Test A
- **Kategori**: Sembako (or any category)
- **Unit Dasar**: pcs
- **Harga Beli Dasar**: 5000
- **Status**: Active ✅
- Click "Simpan"

### Product 2
- **Kode Produk**: PROD-002
- **Nama Produk**: Product Test B
- **Kategori**: Sembako
- **Unit Dasar**: pcs
- **Harga Beli Dasar**: 7500
- **Status**: Active ✅
- Click "Simpan"

### Product 3
- **Kode Produk**: PROD-003
- **Nama Produk**: Product Test C
- **Kategori**: Sembako
- **Unit Dasar**: pcs
- **Harga Beli Dasar**: 3000
- **Status**: Active ✅
- Click "Simpan"

### Product 4 (Optional - for other tests)
- **Kode Produk**: PROD-004
- **Nama Produk**: Product Test D
- **Kategori**: Sembako
- **Unit Dasar**: pcs
- **Harga Beli Dasar**: 8000
- **Status**: Active ✅

### Product 5 (Optional)
- **Kode Produk**: PROD-005
- **Nama Produk**: Product Test E
- **Kategori**: Sembako
- **Unit Dasar**: pcs
- **Harga Beli Dasar**: 4500
- **Status**: Active ✅

### Product 6 (Optional)
- **Kode Produk**: PROD-006
- **Nama Produk**: Product Test F
- **Kategori**: Sembako
- **Unit Dasar**: pcs
- **Harga Beli Dasar**: 6000
- **Status**: Active ✅

---

## Verify Products Created

1. Go back to Products list page
2. Search for "PROD-" in search box
3. Should see all created products
4. Verify they are all Active (green badge)

---

## After Creating Products

**Now retry the Excel upload test:**
1. Go back to Initial Stock Setup page
2. Select warehouse
3. Choose Excel import method
4. Upload `test-valid.xlsx` again
5. Should now show: ✅ "3 produk berhasil divalidasi"

---

## Quick Alternative: Use Existing Products

If you already have products in your database, you can:

1. **Find existing products:**
   - Go to Products list page
   - Note down 3-5 product codes that exist

2. **Edit the Excel file:**
   - Open `test-data/test-valid.xlsx`
   - Replace PROD-001, PROD-002, PROD-003 with your actual product codes
   - Save and re-upload

3. **Or regenerate Excel with existing codes:**
   - I can create a script that fetches products from API and generates Excel with real codes
