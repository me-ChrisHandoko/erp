# 🚀 CSP Report-Only Activation Guide

**Date:** 2026-01-07
**Status:** ✅ Configuration Updated
**Mode:** Report-Only (Safe Testing Mode)

---

## ✅ Step 1: Configuration Update - COMPLETED

**File Modified:** `backend/.env`

**Changes Applied:**
```bash
# Phase 3: CSP (🎯 ENABLED NOW - Report-Only Mode)
SECURITY_ENABLE_CSP=true
SECURITY_CSP_REPORT_ONLY=true
```

**What This Means:**
- ✅ CSP Report-Only mode akan aktif setelah restart
- ✅ Browser akan melaporkan violations tapi TIDAK block
- ✅ Website akan berfungsi normal
- ✅ Aman untuk testing tanpa risk

---

## 🔄 Step 2: Restart Backend Server

**IMPORTANT:** Backend perlu di-restart untuk load configuration baru.

### Option A: Restart via Terminal

**If you started backend with:**
```bash
go run cmd/server/main.go
```

**Steps to restart:**
```bash
# 1. Stop current server
# Press: Ctrl + C

# 2. Start server again
cd backend
go run cmd/server/main.go

# 3. Look for these messages in output:
[GIN-debug] Listening and serving HTTP on :8080
[SECURITY] Security Headers Middleware enabled
[SECURITY] Phase 1 (Safe Headers): ACTIVE ✅
[SECURITY] Phase 2 (HSTS): DISABLED ⏳
[SECURITY] Phase 3 (CSP): ACTIVE - Report-Only Mode 🧪
```

### Option B: Restart via Air (Hot Reload)

**If you're using Air for hot reload:**
```bash
# Air should auto-detect .env changes
# Check Air output for reload message

# If not auto-reloaded, restart Air:
Ctrl + C
cd backend
air
```

### Option C: Restart Docker Container

**If running in Docker:**
```bash
# Restart container
docker-compose restart backend

# Or rebuild if needed
docker-compose down
docker-compose up -d backend
```

### Expected Output After Restart

**Console output should show:**
```
[INFO] Loading configuration from .env
[INFO] Security Headers Configuration:
  - X-Frame-Options: ENABLED
  - X-Content-Type-Options: ENABLED
  - X-XSS-Protection: ENABLED
  - Referrer-Policy: ENABLED
  - Permissions-Policy: ENABLED
  - HSTS: DISABLED (waiting for SSL)
  - CSP: ENABLED (Report-Only Mode)
[INFO] CSP violations will be reported to: /api/v1/csp-report
[GIN-debug] Listening and serving HTTP on :8080
```

---

## ✅ Step 3: Verify CSP Header is Active

### Test 1: Check CSP Header with curl

**Run this command:**
```bash
curl -I http://localhost:8080/health
```

**Expected Output:**
```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
Content-Security-Policy-Report-Only: default-src 'self'; script-src 'self' 'nonce-XXXXXXXX'; style-src 'self' 'nonce-XXXXXXXX'; img-src 'self' data: blob: http://localhost:8080 https://localhost:8080; font-src 'self' data:; connect-src 'self' http://localhost:8080 ws://localhost:8080; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; report-uri /api/v1/csp-report
Date: Tue, 07 Jan 2026 10:30:00 GMT
```

**✅ SUCCESS if you see:** `Content-Security-Policy-Report-Only` header
**❌ FAILED if:** Header not present or shows `Content-Security-Policy` (without Report-Only)

### Test 2: Check API Endpoint

**Run this command:**
```bash
curl http://localhost:8080/api/v1/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-07T10:30:00Z"
}
```

**AND check response headers:**
```bash
curl -I http://localhost:8080/api/v1/health | grep -i "content-security"
```

**Expected:**
```
Content-Security-Policy-Report-Only: default-src 'self'; ...
```

### Test 3: Visual Verification

**Open browser and check Network tab:**

1. Open frontend: `http://localhost:3000`
2. Open DevTools (F12)
3. Go to **Network** tab
4. Click any API request (e.g., login)
5. Look at **Response Headers**
6. Find: `Content-Security-Policy-Report-Only`

**Screenshot location for verification:**
- Click request → Headers tab → Response Headers
- Scroll to find CSP header

---

## 🧪 Step 4: Test Frontend & Monitor Violations

### Test Frontend Application

**Open frontend in browser:**
```bash
# Frontend should already be running at:
http://localhost:3000
```

**Open DevTools Console (F12 → Console tab)**

**Test these flows and watch console:**

#### Test 1: Login Flow
```
1. Navigate to: http://localhost:3000/login
2. Enter credentials
3. Click "Login"
4. Check console for CSP warnings

Expected: ✅ No CSP warnings (only normal logs)
```

#### Test 2: Dashboard
```
1. Navigate to dashboard
2. Check console for CSP warnings

Expected: ✅ No CSP warnings
```

#### Test 3: CRUD Operations
```
1. Navigate to Products page
2. Create new product
3. Edit product
4. Delete product
5. Check console for CSP warnings

Expected: ✅ No CSP warnings
```

#### Test 4: File Upload
```
1. Navigate to Company Profile
2. Upload company logo
3. Check console for CSP warnings

Expected: ✅ No CSP warnings
```

#### Test 5: Forms and Modals
```
1. Open various forms (customer, supplier, warehouse)
2. Open modal dialogs
3. Check console for CSP warnings

Expected: ✅ No CSP warnings
```

---

## 📊 Understanding Console Output

### ✅ CLEAN Console (Expected)

**What you should see:**
```
[React] Component mounted
[Redux] Action dispatched: LOGIN_SUCCESS
[API] GET /api/v1/products - 200 OK
```

**NO CSP warnings** = Perfect! ✅

### ⚠️ CSP Warning (Report-Only Mode)

**If you see this:**
```
[Report Only] Refused to load the script 'https://external.com/script.js'
because it violates the following Content Security Policy directive:
"script-src 'self' 'nonce-ABC123'".
```

**This means:**
- 🟡 **Yellow warning** (not red error)
- Script was **NOT blocked** (still loaded)
- Just a report of violation
- Need to investigate source

**Common false positives:**
```
# Browser extensions (IGNORE THESE ✅)
[Report Only] chrome-extension://abc123/content.js
[Report Only] moz-extension://xyz789/script.js

# These are from browser extensions like:
- React DevTools
- Redux DevTools
- uBlock Origin
- Grammarly
- LastPass
```

### 🚨 Real Violation (Needs Investigation)

**If you see:**
```
[Report Only] Refused to load the script 'https://analytics.com/tracker.js'
because it violates the following Content Security Policy directive:
"script-src 'self' 'nonce-ABC123'".
```

**Document violation:**
1. Take screenshot of console error
2. Note the URL: `https://analytics.com/tracker.js`
3. Note the page: `http://localhost:3000/dashboard`
4. Note the directive: `script-src`
5. Check if this is intentional external script

**Action:**
- If intentional → Add to CSP whitelist
- If not needed → Remove from code

---

## 📋 Monitoring Checklist

### Daily Monitoring (Week 1)

**Day 1: Active Testing**
- [ ] Login/Logout flows tested
- [ ] All CRUD operations tested
- [ ] File uploads tested
- [ ] Forms and modals tested
- [ ] Console checked: No violations
- [ ] Backend logs checked: No errors

**Day 2-3: Feature Testing**
- [ ] Customer management tested
- [ ] Supplier management tested
- [ ] Product management tested
- [ ] Warehouse operations tested
- [ ] Console monitored: Clean

**Day 4-5: Edge Cases**
- [ ] Multiple file uploads
- [ ] Large data tables
- [ ] Export/import features
- [ ] Reports generation
- [ ] Console monitored: Clean

**Day 6-7: Full Regression**
- [ ] All features re-tested
- [ ] All user roles tested
- [ ] All pages visited
- [ ] Final console check: Clean
- [ ] Backend logs reviewed: No CSP violations

### Backend Logs Monitoring

**Check backend logs daily:**
```bash
# View real-time logs
tail -f backend/logs/app.log

# Search for CSP violations
grep -i "csp" backend/logs/app.log

# Check CSP report endpoint calls
grep -i "csp-report" backend/logs/app.log
```

**Expected:**
```
# No CSP violation logs = Good ✅
```

**If violations found:**
```
[CSP] Violation reported from http://localhost:3000/dashboard
  Violated Directive: script-src
  Blocked URI: https://external.com/script.js
  Source File: http://localhost:3000/_next/static/chunks/main.js
  Line: 123
```

---

## 🔄 After 1 Week: Enable Enforcement Mode

### Prerequisites Checklist

**Before enabling enforcement, verify:**
- [ ] CSP Report-Only tested for 7+ days
- [ ] Zero violations found (excluding browser extensions)
- [ ] All critical features tested
- [ ] All team members tested
- [ ] Frontend works perfectly
- [ ] Backend logs clean
- [ ] Ready to enforce CSP

### Enable Enforcement

**File: `backend/.env`**

```bash
# Change from Report-Only to Enforcement
SECURITY_CSP_REPORT_ONLY=false  # ← Change true to false
```

**Restart backend:**
```bash
cd backend
# Stop with Ctrl+C
go run cmd/server/main.go
```

**Verify enforcement:**
```bash
curl -I http://localhost:8080/health | grep "Content-Security-Policy:"

# Expected (without "Report-Only"):
Content-Security-Policy: default-src 'self'; ...
```

**What changes:**
- Browser will **BLOCK** violations (not just warn)
- CSP errors will be **RED** in console (not yellow)
- Any violation will **break** that feature

**Test immediately after enabling:**
- [ ] Login flow works
- [ ] Dashboard loads
- [ ] CRUD operations work
- [ ] File uploads work
- [ ] No features broken

---

## ⚠️ Troubleshooting

### Issue 1: CSP Header Not Showing

**Symptom:**
```bash
curl -I http://localhost:8080/health
# No Content-Security-Policy-Report-Only header
```

**Solutions:**

**Check 1: Verify .env file**
```bash
cat backend/.env | grep CSP

# Expected:
SECURITY_ENABLE_CSP=true
SECURITY_CSP_REPORT_ONLY=true
```

**Check 2: Restart backend**
```bash
# Stop backend (Ctrl+C)
# Start again
cd backend
go run cmd/server/main.go
```

**Check 3: Check backend logs**
```bash
# Look for security headers initialization
grep -i "security" backend/logs/app.log
grep -i "csp" backend/logs/app.log
```

**Check 4: Verify middleware is loaded**
```bash
# Check router.go includes SecurityHeadersMiddleware
grep -i "SecurityHeadersMiddleware" backend/internal/router/router.go

# Expected: Should find the middleware
```

### Issue 2: Backend Won't Start

**Symptom:**
```bash
go run cmd/server/main.go
# Error: cannot load configuration
```

**Solutions:**

**Check .env syntax:**
```bash
# Check for typos or missing values
cat backend/.env | tail -20

# Look for:
# - Missing = signs
# - Extra spaces
# - Invalid values
```

**Fix common issues:**
```bash
# Correct format:
SECURITY_ENABLE_CSP=true
SECURITY_CSP_REPORT_ONLY=true

# Wrong format:
SECURITY_ENABLE_CSP = true  # Extra spaces
SECURITY_ENABLE_CSP=True    # Capitalized (should be lowercase)
```

### Issue 3: Violations from Browser Extensions

**Symptom:**
```
[Report Only] chrome-extension://abc123/content.js violated CSP
```

**Solution:**
✅ **IGNORE** - These are normal and expected
✅ Browser extensions don't affect real users
✅ Not a problem with your application

**To reduce noise in console:**
- Disable extensions temporarily during testing
- Or just ignore extension-related warnings

### Issue 4: CORS Errors After Enabling CSP

**Symptom:**
```
Access-Control-Allow-Origin error
CORS policy blocked the request
```

**Clarification:**
This is **NOT a CSP issue** - it's a CORS configuration issue.

**Check CORS config:**
```bash
# backend/.env
cat backend/.env | grep CORS

# Should include:
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

**If frontend URL changed, update CORS:**
```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://localhost:3000
```

---

## 📊 Success Metrics

### Week 1 Target Metrics

**Console Violations:**
- Target: 0 violations
- Acceptable: 0-2 (browser extensions only)
- Action if >2: Investigate and fix

**Backend Logs:**
- Target: No CSP violation logs
- Acceptable: No violation logs
- Action if violations: Review and fix

**Functionality:**
- Target: 100% features working
- Acceptable: 100% features working
- Action if broken: Rollback and investigate

**Team Feedback:**
- Target: No issues reported
- Acceptable: No issues reported
- Action if issues: Investigate and fix

---

## 🎯 Next Steps After Successful Testing

### Timeline

**Week 1 (Current):**
- ✅ CSP Report-Only enabled
- ✅ Daily monitoring
- ✅ Testing all features

**Week 2:**
- Review violations (if any)
- Fix any issues found
- Continue monitoring

**Week 3:**
- Enable CSP enforcement mode
- Test thoroughly
- Monitor for 48 hours

**Week 4:**
- Production deployment preparation
- Add production domains to CSP
- Final security audit

---

## 📞 Support

**If you encounter issues:**

1. **Check documentation:**
   - `backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md`
   - `frontend/docs/PHASE-3-CSP-AUDIT-REPORT.md`

2. **Review logs:**
   - Backend: `backend/logs/app.log`
   - Browser: DevTools Console (F12)

3. **Verify configuration:**
   - Backend: `backend/.env`
   - Frontend: `frontend/.env.local`

4. **Test endpoints:**
   - Health: `http://localhost:8080/health`
   - API: `http://localhost:8080/api/v1/health`
   - CSP Report: `http://localhost:8080/api/v1/csp-report`

---

**Activation Date:** 2026-01-07
**Next Review:** After 7 days of monitoring
**Status:** 🟢 Active - Report-Only Mode
