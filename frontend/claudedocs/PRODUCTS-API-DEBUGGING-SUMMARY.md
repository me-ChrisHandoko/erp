# Products API Debugging Summary

**Date**: 2025-12-29
**Issue**: Products API returns 500 Internal Server Error
**Status**: **REQUIRES BACKEND LOGS INSPECTION**

---

## ✅ What's Working

### Frontend
- ✅ Authentication working correctly (Budi Santoso authenticated)
- ✅ Company context initialized (PT Distribusi Utama selected)
- ✅ Access token being sent in Authorization header
- ✅ X-Company-ID header NOW being sent correctly (`270b570e-ba46-4f06-a009-6f3c4a552bc9`)
- ✅ Products page UI rendering properly
- ✅ Error handling and retry mechanism working

### Backend
- ✅ Server running on port 8080
- ✅ CORS configured correctly
- ✅ Products routes registered in router.go (lines 292-313)
- ✅ Authentication middleware working
- ✅ Company context middleware receiving X-Company-ID header

### Database
- ✅ Products table exists and has correct schema
- ✅ Related tables exist: `product_units`, `warehouses`, `warehouse_stocks`, `product_suppliers`
- ✅ All tables have required `is_active` columns
- ✅ Sample data exists: 2 products in database
  - Product 1: BRS-001 "Beras Premium 5kg" → company `270b570e-ba46-4f06-a009-6f3c4a552bc9` ✅
  - Product 2: MNY-001 "Minyak Goreng 2L" → company `deebb24b-c22d-4166-bdc5-3c7e2ed75a7a`
- ✅ Manual SQL query works fine:
  ```sql
  SELECT COUNT(*) FROM products
  WHERE company_id = '270b570e-ba46-4f06-a009-6f3c4a552bc9';
  -- Returns: 1
  ```

---

## ❌ What's NOT Working

### API Response
```http
GET /api/v1/products?page=1&page_size=20&sort_by=code&sort_order=asc
Authorization: Bearer <valid_token>
X-Company-ID: 270b570e-ba46-4f06-a009-6f3c4a552bc9

HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred"
  },
  "success": false
}
```

### Error Analysis
The error message is **generic** because the backend error handler (`product_handler.go:639-654`) returns this default message for any non-AppError type errors:

```go
// Unknown error - return internal server error
c.JSON(http.StatusInternalServerError, gin.H{
    "success": false,
    "error": gin.H{
        "code":    "INTERNAL_ERROR",
        "message": "An unexpected error occurred",
    },
})
```

**The actual error details are being hidden/swallowed.**

---

## 🔍 Investigation Performed

### 1. Verified Database Schema ✅
```bash
# All tables exist with correct structure:
\dt products              → EXISTS
\dt product_units         → EXISTS (has is_active column)
\dt warehouses            → EXISTS (has is_active column)
\dt warehouse_stocks      → EXISTS
\dt product_suppliers     → EXISTS
```

### 2. Verified Sample Data ✅
```sql
-- User's active company has 1 product
SELECT * FROM products
WHERE company_id = '270b570e-ba46-4f06-a009-6f3c4a552bc9';

-- Result: 1 row (Beras Premium 5kg)
```

### 3. Verified Backend Code ✅
- **Handler** (`product_handler.go:110-141`): Correctly validates company_id, binds filters, calls service
- **Service** (`product_service.go:208-276`): Builds query correctly, applies filters, preloads relations
- **Preloads used**:
  - `Preload("Units", "is_active = ?", true)` → Table exists ✅
  - `Preload("WarehouseStocks.Warehouse", "is_active = ?", true)` → Tables exist ✅

### 4. Verified API Request ✅
```bash
# Tested with curl using correct headers:
curl -H "Authorization: Bearer <token>" \
     -H "X-Company-ID: 270b570e-ba46-4f06-a009-6f3c4a552bc9" \
     "http://localhost:8080/api/v1/products..."

# Result: 500 Internal Server Error (same as frontend)
```

---

## 🎯 Root Cause Hypothesis

Since:
1. Database schema is correct
2. Sample data exists
3. Manual SQL queries work
4. Backend code looks correct
5. Authentication and headers are correct

**The error is likely happening in the Go/GORM query execution or data marshaling.**

Possible causes:
1. **GORM Preload Error** - Issue with loading related data (Units or WarehouseStocks)
2. **Data Type Mismatch** - JSON marshaling failing due to decimal.Decimal or time.Time fields
3. **Foreign Key Constraint** - Related records causing query issues
4. **Backend Cache/State Issue** - Backend needs restart after migration

---

## 🚨 **ACTION REQUIRED: CHECK BACKEND LOGS**

**Please check the backend terminal/logs for the ACTUAL error message.**

The backend server is running at `http://localhost:8080` and was started with:
```bash
cd ~/Development/work/erp/backend
go run cmd/server/main.go
```

**When you click "Try Again" on the frontend, look for error logs in the backend terminal.**

Expected error format:
```
[GIN] 2025/12/29 - 15:34:12 | 500 | ... | GET /api/v1/products...
Error: <ACTUAL_ERROR_MESSAGE_HERE>
```

---

## 🔧 Potential Solutions (Based on Log Output)

### If Error: "column <name> does not exist"
**Cause**: Missing column in database schema
**Solution**: Run migration again or add missing column manually

### If Error: "preload: can't find field <name>"
**Cause**: GORM model relationship issue
**Solution**: Check Product model struct tags and relationships

### If Error: "sql: Scan error on column <name>: unsupported Scan"
**Cause**: Data type mismatch between Go struct and database
**Solution**: Check decimal.Decimal, time.Time, or custom type handling

### If Error: "invalid memory address" or "nil pointer dereference"
**Cause**: Missing initialization or nil pointer in backend code
**Solution**: Add nil checks in service or handler

### If No Specific Error (Just 500)
**Solution**: Try restarting backend server:
```bash
# Stop current backend (Ctrl+C)
cd ~/Development/work/erp/backend
go run cmd/server/main.go
```

---

## 📊 Testing Progress

| Component | Status | Details |
|-----------|--------|---------|
| Frontend UI/UX | ✅ PASS | Mobile-responsive, proper loading states |
| Authentication | ✅ PASS | JWT token, company context working |
| API Endpoint | ❌ FAIL | 500 Internal Server Error |
| Database Schema | ✅ PASS | All tables exist with correct structure |
| Sample Data | ✅ PASS | 1 product available for user's company |

**Blocker**: Cannot test full CRUD operations until 500 error is resolved.

---

## 📝 Next Steps

1. **Check backend terminal for actual error message**
2. **Share error log output** for diagnosis
3. **Try backend restart** if no clear error shown
4. **Run backend tests** to verify products service:
   ```bash
   cd backend
   go test ./internal/service/product/... -v
   ```

---

**Report Generated**: 2025-12-29 by Claude Code
**Session**: Products Module Testing
**Test Location**: Mobile viewport (390x844px - iPhone 12 Pro)
