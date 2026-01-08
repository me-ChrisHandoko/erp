# 🚀 CSP Report-Only - Quick Command Reference

**Status:** ✅ Configuration Updated
**Mode:** Report-Only (Safe Testing)

---

## 🔄 Step 1: Restart Backend (DO THIS NOW)

```bash
# Stop current backend server
# Press: Ctrl + C

# Start backend again
cd backend
go run cmd/server/main.go

# Expected output:
# [SECURITY] CSP Report-Only mode enabled
# [GIN-debug] Listening and serving HTTP on :8080
```

---

## ✅ Step 2: Verify CSP Header

```bash
# Test CSP header
curl -I http://localhost:8080/health | grep -i "content-security"

# Expected output:
# Content-Security-Policy-Report-Only: default-src 'self'; ...
```

---

## 🧪 Step 3: Test Frontend

```bash
# Open browser
open http://localhost:3000

# Open DevTools: F12 → Console tab
# Test login, dashboard, CRUD operations
# Expected: No CSP warnings (except browser extensions)
```

---

## 📊 Step 4: Monitor Violations

```bash
# Check backend logs
tail -f backend/logs/app.log | grep -i csp

# Check for CSP violations
grep "CSP" backend/logs/app.log

# Expected: No CSP violation logs
```

---

## 🔄 After 1 Week: Enable Enforcement

```bash
# Edit backend/.env
# Change: SECURITY_CSP_REPORT_ONLY=false

# Restart backend
cd backend
go run cmd/server/main.go

# Verify enforcement (no "Report-Only")
curl -I http://localhost:8080/health | grep "Content-Security-Policy:"
```

---

## ⚠️ Rollback if Issues

```bash
# Edit backend/.env
# Change: SECURITY_ENABLE_CSP=false

# Restart backend
cd backend
go run cmd/server/main.go
```

---

**Next:** See `CSP-REPORT-ONLY-ACTIVATION-GUIDE.md` for detailed instructions
