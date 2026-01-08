# 🛡️ Phase 2: HSTS Frontend Implementation Summary

Implementation complete untuk mendukung backend HSTS (HTTP Strict Transport Security).

---

## ✅ What Was Implemented

### 1. Environment Variables Configuration ✅

**File: `.env.example`**
- ✅ Updated API URL to support HTTPS: `https://localhost:8080`
- ✅ Added WebSocket URL configuration: `wss://localhost:8080/ws`
- ✅ Added comprehensive SSL setup instructions
- ✅ Added reference to SSL setup documentation

**File: `.env.local`**
- ✅ Added Phase 2 implementation notes
- ✅ Added step-by-step HTTPS migration instructions
- ✅ Currently using HTTP (until SSL certificate ready)
- ✅ Clear instructions on when/how to switch to HTTPS

### 2. Next.js Configuration ✅

**File: `next.config.ts`**
- ✅ Support both HTTP and HTTPS image sources for localhost
- ✅ Production-only HSTS header
- ✅ Production-only `upgrade-insecure-requests` CSP directive
- ✅ Image optimization with modern formats (AVIF, WebP)
- ✅ Environment-based configuration (development vs production)

### 3. API Service Configuration ✅

**Files checked:**
- ✅ `src/store/services/authApi.ts` - Already using `process.env.NEXT_PUBLIC_API_URL`
- ✅ `src/store/services/multiCompanyApi.ts` - Already using `process.env.NEXT_PUBLIC_API_URL`
- ✅ All other API services - Inherit from base configuration

**Status:** ✅ No changes needed - Already properly configured!

### 4. Documentation ✅

**File: `docs/SSL-DEVELOPMENT-SETUP.md`**
- ✅ Complete SSL setup guide (500+ lines)
- ✅ mkcert installation instructions (macOS, Windows, Linux)
- ✅ Certificate generation steps
- ✅ Backend HTTPS configuration
- ✅ Frontend configuration steps
- ✅ Testing & verification procedures
- ✅ Troubleshooting guide
- ✅ Security best practices

---

## 📊 Current Status

### ✅ Phase 1: Safe Headers (Backend)
**Status:** Active and protecting the application

### ⏳ Phase 2: HSTS (Backend + Frontend)
**Status:** Implementation complete, waiting for SSL certificate setup

| Component | Status | Notes |
|-----------|--------|-------|
| Backend middleware | ✅ Implemented | Disabled until SSL ready |
| Backend config | ✅ Complete | `SECURITY_ENABLE_HSTS=false` |
| Frontend env vars | ✅ Updated | Instructions for HTTPS switch |
| Frontend next.config | ✅ Updated | Production HSTS ready |
| Frontend API services | ✅ Verified | Using env variables correctly |
| SSL documentation | ✅ Complete | Step-by-step guide |

---

## 🚀 Deployment Checklist

### Current State (HTTP Development)

**Backend:**
```bash
# .env
SECURITY_ENABLE_HSTS=false  # Disabled
```

**Frontend:**
```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080  # HTTP
```

**Status:** ✅ Working with Phase 1 security headers active

---

### Next Steps: Enable HSTS

#### Step 1: SSL Certificate Setup (Backend)

1. **Install mkcert:**
   ```bash
   # macOS
   brew install mkcert

   # Windows
   choco install mkcert

   # Linux
   sudo apt install mkcert
   ```

2. **Setup local CA:**
   ```bash
   mkcert -install
   ```

3. **Generate certificates:**
   ```bash
   cd backend
   mkcert localhost 127.0.0.1 ::1
   ```

   Creates:
   - `localhost+2.pem` (certificate)
   - `localhost+2-key.pem` (private key)

4. **Update backend main.go:**

   See `backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md` for code changes.

#### Step 2: Update Frontend Environment

**File: `frontend/.env.local`**
```bash
# Change from HTTP to HTTPS
NEXT_PUBLIC_API_URL=https://localhost:8080
NEXT_PUBLIC_WS_URL=wss://localhost:8080/ws
```

#### Step 3: Enable HSTS in Backend

**File: `backend/.env`**
```bash
# Enable HSTS
SECURITY_ENABLE_HSTS=true
SECURITY_HSTS_MAX_AGE=604800  # Start with 1 week
```

#### Step 4: Restart & Test

**Start backend:**
```bash
cd backend
go run cmd/server/main.go

# Expected:
# Starting HTTPS server on port 8080
```

**Start frontend:**
```bash
cd frontend
npm run dev

# Open: http://localhost:3000
```

**Verify:**
1. ✅ Browser shows secure padlock on API calls
2. ✅ No SSL/certificate errors
3. ✅ Login flow works
4. ✅ HSTS header present: `Strict-Transport-Security`

#### Step 5: Testing Checklist

- [ ] Backend HTTPS server starts without errors
- [ ] Frontend can connect to HTTPS backend
- [ ] Login flow works correctly
- [ ] API calls successful (check DevTools Network tab)
- [ ] No mixed content warnings
- [ ] HSTS header present in responses
- [ ] Browser shows secure connection (padlock icon)
- [ ] WebSocket connections work (if using)
- [ ] File uploads work
- [ ] All features tested

#### Step 6: Gradual HSTS Rollout

**Week 1: Short max-age**
```bash
SECURITY_HSTS_MAX_AGE=604800  # 1 week
```
Monitor for SSL issues, test all features.

**Week 2-3: Increase max-age**
```bash
SECURITY_HSTS_MAX_AGE=2592000  # 1 month
```
Continue monitoring.

**Week 4+: Production max-age**
```bash
SECURITY_HSTS_MAX_AGE=31536000  # 1 year
SECURITY_HSTS_INCLUDE_SUBDOMAINS=true  # After subdomain SSL verified
```

---

## 📁 Files Modified Summary

### Frontend Changes

| File | Status | Changes |
|------|--------|---------|
| `.env.example` | ✅ Updated | HTTPS URLs, SSL instructions |
| `.env.local` | ✅ Updated | Phase 2 notes, migration guide |
| `next.config.ts` | ✅ Updated | HTTPS support, security headers |
| `docs/SSL-DEVELOPMENT-SETUP.md` | ✅ Created | Complete SSL guide |
| `docs/PHASE-2-HSTS-IMPLEMENTATION.md` | ✅ Created | This file |

### No Changes Needed

| File | Status | Reason |
|------|--------|--------|
| `src/store/services/authApi.ts` | ✅ OK | Already using env vars |
| `src/store/services/*.ts` | ✅ OK | Inherit from base config |
| `package.json` | ✅ OK | Next.js handles HTTPS automatically |

---

## 🔍 How It Works

### Development (Current - HTTP)

```
┌─────────┐  http://localhost:3000    ┌──────────┐
│ Browser │ ─────────────────────────> │ Frontend │
└─────────┘                            └──────────┘
     │                                       │
     │ API calls                             │
     │ http://localhost:8080/api/v1          │
     └───────────────────────────────────────┘
                       │
                       ▼
                 ┌──────────┐
                 │ Backend  │ Phase 1 headers active
                 │  (HTTP)  │ HSTS disabled
                 └──────────┘
```

### After SSL Setup (HTTPS)

```
┌─────────┐  http://localhost:3000    ┌──────────┐
│ Browser │ ─────────────────────────> │ Frontend │
└─────────┘                            └──────────┘
     │                                       │
     │ API calls                             │
     │ https://localhost:8080/api/v1 ◄───────┘
     │                                 (env var)
     └─────────────┐
                   ▼
              ┌──────────┐
              │ Backend  │ HSTS active
              │ (HTTPS)  │ Forces HTTPS
              └──────────┘
                   │
                   ▼
        Strict-Transport-Security: max-age=31536000
```

### Production Deployment

```
┌─────────┐  https://erp-app.com      ┌──────────┐
│ Browser │ ─────────────────────────> │ Frontend │
└─────────┘                            │  (HTTPS) │
     │                                 └──────────┘
     │ HSTS enforced                         │
     │ (never allows HTTP)                   │
     │                                       │
     │ API calls                             │
     │ https://api.erp-app.com/v1            │
     └───────────────────────────────────────┘
                       │
                       ▼
                 ┌──────────┐
                 │ Backend  │ Full HSTS
                 │ (HTTPS)  │ + CSP (Phase 3)
                 └──────────┘
```

---

## ⚠️ Important Notes

### HSTS Behavior

**After HSTS is enabled:**
1. ✅ Browser remembers to always use HTTPS
2. ⚠️ User CANNOT access site via HTTP anymore
3. ⚠️ If SSL certificate expires → Site INACCESSIBLE
4. ⚠️ Must wait for `max-age` to expire OR clear HSTS manually

**Clear HSTS (Development):**
- Chrome: `chrome://net-internals/#hsts` → Delete domain
- Firefox: Clear history → Delete everything
- Safari: Clear history → All history

### SSL Certificate Management

**Development:**
- ✅ Use mkcert (easy, trusted locally)
- ✅ Regenerate if expired (simple: `mkcert localhost`)
- ✅ Never commit certificates to git (in .gitignore)

**Production:**
- ✅ Use Let's Encrypt (free, auto-renewal)
- ✅ Or commercial SSL provider
- ✅ Setup auto-renewal (critical!)
- ✅ Monitor expiration dates

### Environment Variables

**Always use environment variables for API URLs:**
```typescript
// ✅ CORRECT
const API_URL = process.env.NEXT_PUBLIC_API_URL;

// ❌ WRONG - Hard-coded
const API_URL = "http://localhost:8080";
```

**This allows easy switching:**
- Development: `https://localhost:8080`
- Staging: `https://staging-api.erp.com`
- Production: `https://api.erp.com`

---

## 🧪 Testing Scenarios

### Test 1: HTTP → HTTPS Migration

1. **Before SSL:**
   - Backend: HTTP port 8080
   - Frontend: `NEXT_PUBLIC_API_URL=http://localhost:8080`
   - Expected: ✅ Works, Phase 1 headers active

2. **After SSL Setup:**
   - Backend: HTTPS port 8080
   - Frontend: `NEXT_PUBLIC_API_URL=https://localhost:8080`
   - Expected: ✅ Works, secure connection

3. **After HSTS Enable:**
   - Backend: `SECURITY_ENABLE_HSTS=true`
   - Browser: Forces HTTPS automatically
   - Expected: ✅ Always HTTPS, cannot use HTTP

### Test 2: CORS Verification

**Backend CORS must allow HTTPS origin:**
```bash
# backend/.env
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://localhost:3000
```

### Test 3: Cookie Transmission

**Cookies must work over HTTPS:**
- `refresh_token` (httpOnly cookie)
- `csrf_token` cookie

Test:
1. Login via frontend
2. Check DevTools → Application → Cookies
3. Verify cookies present
4. Verify `Secure` flag set (production only)

### Test 4: WebSocket Connections

**If using WebSockets:**
```bash
# Update WebSocket URL
NEXT_PUBLIC_WS_URL=wss://localhost:8080/ws
```

Test:
1. Establish WebSocket connection
2. Verify `wss://` protocol used
3. Verify no SSL errors

---

## 📚 Additional Resources

**Documentation:**
- Backend: `backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md`
- SSL Setup: `frontend/docs/SSL-DEVELOPMENT-SETUP.md`

**External Resources:**
- [mkcert GitHub](https://github.com/FiloSottile/mkcert)
- [HSTS Specification](https://tools.ietf.org/html/rfc6797)
- [Let's Encrypt](https://letsencrypt.org/)
- [SSL Labs](https://www.ssllabs.com/ssltest/)
- [SecurityHeaders.com](https://securityheaders.com/)

---

## ✅ Verification Commands

### Check Current Configuration

```bash
# Frontend API URL
cat frontend/.env.local | grep NEXT_PUBLIC_API_URL

# Backend HSTS status
cat backend/.env | grep SECURITY_ENABLE_HSTS

# Certificate exists
ls backend/localhost+2.pem backend/localhost+2-key.pem
```

### Test Backend HTTPS

```bash
# Should work without SSL errors
curl https://localhost:8080/health

# Check HSTS header (after enabled)
curl -I https://localhost:8080/health | grep Strict-Transport-Security
```

### Test Frontend API Connection

```bash
# Start both servers
cd backend && go run cmd/server/main.go &
cd frontend && npm run dev &

# Open browser
open http://localhost:3000

# Check DevTools → Network tab
# API calls should go to https://localhost:8080
```

---

## 🎯 Success Criteria

### Phase 2 Complete When:

- [x] Frontend environment variables updated
- [x] Next.js configuration supports HTTPS
- [x] API services verified to use env variables
- [x] SSL setup documentation created
- [x] Testing procedures documented
- [ ] SSL certificates generated (pending developer action)
- [ ] Backend HTTPS enabled (pending SSL setup)
- [ ] HSTS enabled in backend (pending testing)
- [ ] All features tested with HTTPS (pending deployment)
- [ ] SecurityHeaders.com shows A rating (pending Phase 3)

**Current:** Implementation complete ✅, Deployment pending ⏳

---

## 🚀 Next Phase

**Phase 3: Content Security Policy (CSP)**

After Phase 2 is fully deployed and tested:

1. Frontend audit for inline scripts/styles
2. CSP Report-Only mode monitoring
3. Fix CSP violations
4. CSP enforcement mode
5. Final security audit

**Target:** SecurityHeaders.com A+ rating

---

**Questions?** See:
- `docs/SSL-DEVELOPMENT-SETUP.md` (detailed SSL guide)
- `backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md` (complete security headers guide)

**Implementation Date:** 2025-01-07
**Status:** ✅ Complete and ready for SSL setup
