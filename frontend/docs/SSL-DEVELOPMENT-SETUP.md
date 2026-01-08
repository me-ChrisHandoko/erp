# 🔒 SSL Development Setup Guide

Complete guide for setting up HTTPS/SSL in local development environment to support backend security headers (Phase 2: HSTS).

---

## 📋 Why SSL in Development?

### Backend Security Headers Requirement

Backend telah mengimplementasikan **HSTS (HTTP Strict Transport Security)** sebagai bagian dari Phase 2 security headers. HSTS memaksa browser untuk **SELALU menggunakan HTTPS**.

**Problem tanpa SSL development:**
```
Backend enable HSTS → Browser force HTTPS → http://localhost ❌ Connection Refused
```

**Solution dengan SSL development:**
```
Backend enable HSTS → Browser force HTTPS → https://localhost ✅ Works!
```

---

## 🚀 Quick Start (Recommended: mkcert)

### Step 1: Install mkcert

**macOS:**
```bash
brew install mkcert
brew install nss  # For Firefox support
```

**Windows:**
```bash
choco install mkcert
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt install libnss3-tools
wget https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
chmod +x mkcert-v1.4.4-linux-amd64
sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert

# Or use package manager
sudo apt install mkcert  # Ubuntu 22.04+
```

### Step 2: Setup Local Certificate Authority (CA)

```bash
# Install local CA (run ONCE per machine)
mkcert -install
```

**Output:**
```
Created a new local CA 💥
The local CA is now installed in the system trust store! ⚡️
```

**What this does:**
- Creates a local Certificate Authority (CA)
- Installs CA certificate in system trust store
- Browser will trust certificates issued by this CA
- **Safe:** Only works on your machine, not trusted by others

### Step 3: Generate Certificates for Backend

```bash
# Navigate to backend project directory
cd /path/to/erp/backend

# Generate certificate for localhost
mkcert localhost 127.0.0.1 ::1

# Output:
# Created a new certificate valid for the following names 📜
#  - "localhost"
#  - "127.0.0.1"
#  - "::1"
#
# The certificate is at "./localhost+2.pem" and the key at "./localhost+2-key.pem" ✅
```

**Files created:**
- `localhost+2.pem` - SSL certificate (public)
- `localhost+2-key.pem` - Private key (keep secret!)

### Step 4: Generate Certificates for Frontend (Optional)

```bash
# Navigate to frontend project directory
cd /path/to/erp/frontend

# Generate certificate for localhost
mkcert localhost 127.0.0.1 ::1
```

**Note:** Frontend mungkin tidak perlu HTTPS jika hanya backend yang enforce HSTS. Tapi lebih baik setup keduanya untuk consistency.

### Step 5: Configure Backend for HTTPS

**Backend file: `cmd/server/main.go`**

Tambahkan SSL configuration:

```go
// After creating HTTP server
srv := &http.Server{
    Addr:           ":" + cfg.Server.Port,
    Handler:        r,
    ReadTimeout:    30 * time.Second,
    WriteTimeout:   30 * time.Second,
    IdleTimeout:    60 * time.Second,
    MaxHeaderBytes: 1 << 20,
}

// ✅ Start with HTTPS if certificates exist
if fileExists("localhost+2.pem") && fileExists("localhost+2-key.pem") {
    log.Printf("Starting HTTPS server on port %s", cfg.Server.Port)
    if err := srv.ListenAndServeTLS("localhost+2.pem", "localhost+2-key.pem"); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
} else {
    log.Printf("Starting HTTP server on port %s", cfg.Server.Port)
    log.Println("⚠️  SSL certificates not found. Run: mkcert localhost 127.0.0.1 ::1")
    if err := srv.ListenAndServe(); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

**Helper function:**
```go
func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}
```

### Step 6: Update Environment Variables

**Frontend: `.env.local`**
```bash
# Change HTTP to HTTPS
NEXT_PUBLIC_API_URL=https://localhost:8080
NEXT_PUBLIC_WS_URL=wss://localhost:8080/ws
```

**Backend: `.env`**
```bash
# Enable HSTS after SSL setup
SECURITY_ENABLE_HSTS=true
SECURITY_HSTS_MAX_AGE=604800  # Start with 1 week
```

### Step 7: Test Setup

**Start Backend:**
```bash
cd backend
go run cmd/server/main.go

# Expected output:
# Starting HTTPS server on port 8080
# Server running at https://localhost:8080
```

**Start Frontend:**
```bash
cd frontend
npm run dev

# Open browser: http://localhost:3000
# API calls will go to https://localhost:8080
```

**Test API Connection:**
```bash
# Should return 200 without SSL errors
curl https://localhost:8080/health

# Output:
# {"status":"healthy","message":"Service is running"}
```

**Test in Browser:**
1. Open: http://localhost:3000
2. Open DevTools → Network tab
3. Login or make any API call
4. Check request URL: should be `https://localhost:8080/api/v1/...`
5. Check response headers: should include `Strict-Transport-Security`

---

## 🔧 Alternative Methods

### Method 2: OpenSSL Self-Signed Certificate

**For production-like testing, but requires manual trust**

```bash
# Generate private key
openssl genrsa -out localhost.key 2048

# Generate certificate signing request (CSR)
openssl req -new -key localhost.key -out localhost.csr \
  -subj "/C=ID/ST=Jakarta/L=Jakarta/O=Development/CN=localhost"

# Generate self-signed certificate (valid 365 days)
openssl x509 -req -days 365 -in localhost.csr \
  -signkey localhost.key -out localhost.crt
```

**⚠️ Browser Warning:**
Self-signed certificates will show "Not Secure" warning in browser. You need to manually trust them.

**Trust in macOS:**
```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain localhost.crt
```

**Trust in Windows:**
1. Double-click `localhost.crt`
2. Click "Install Certificate"
3. Select "Local Machine"
4. Place in "Trusted Root Certification Authorities"

### Method 3: Reverse Proxy with Caddy

**Auto-SSL without certificate management**

**Install Caddy:**
```bash
brew install caddy  # macOS
choco install caddy  # Windows
```

**Create Caddyfile:**
```caddy
localhost:443 {
    reverse_proxy localhost:8080
    tls internal
}
```

**Start Caddy:**
```bash
caddy run
```

**Access:** https://localhost (Caddy auto-generates trusted certificates)

---

## 🧪 Testing & Verification

### 1. Certificate Validation

```bash
# Check certificate details
openssl x509 -in localhost+2.pem -text -noout

# Verify certificate
openssl verify localhost+2.pem
```

### 2. Browser Testing

**Chrome/Edge:**
1. Open: https://localhost:8080/health
2. Click padlock icon in address bar
3. Should show: "Connection is secure"
4. Certificate issued by: "mkcert"

**Firefox:**
1. Open: https://localhost:8080/health
2. Click padlock icon
3. Connection Security → More Information
4. Should show valid certificate

### 3. HSTS Verification

```bash
# Test HSTS header
curl -I https://localhost:8080/health | grep Strict-Transport-Security

# Expected output:
# Strict-Transport-Security: max-age=604800
```

**Browser HSTS Test:**
1. Visit: https://localhost:8080 (establishes HSTS)
2. Try: http://localhost:8080 (should auto-redirect to HTTPS)
3. Chrome: Check `chrome://net-internals/#hsts`
   - Query: localhost
   - Should show HSTS entry

---

## ⚠️ Troubleshooting

### Problem: "mkcert: command not found"

**Solution:**
```bash
# macOS
brew install mkcert

# Windows (run as Administrator)
choco install mkcert

# Linux
curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64"
chmod +x mkcert-v*-linux-amd64
sudo mv mkcert-v*-linux-amd64 /usr/local/bin/mkcert
```

### Problem: "Certificate not trusted" in browser

**Solution:**
```bash
# Re-install CA
mkcert -uninstall
mkcert -install

# Restart browser
```

### Problem: Backend not starting with HTTPS

**Check:**
```bash
# Verify certificates exist
ls -la localhost+2.pem localhost+2-key.pem

# Check file permissions
chmod 644 localhost+2.pem
chmod 600 localhost+2-key.pem

# Test certificate loading
openssl s_server -cert localhost+2.pem -key localhost+2-key.pem -accept 8443
```

### Problem: Frontend API calls failing with SSL error

**Check:**
1. **Backend running HTTPS?**
   ```bash
   curl https://localhost:8080/health
   ```

2. **Environment variable correct?**
   ```bash
   cat .env.local | grep NEXT_PUBLIC_API_URL
   # Should be: https://localhost:8080
   ```

3. **Frontend using correct URL?**
   - Open browser DevTools → Network tab
   - Check API request URL
   - Should be `https://localhost:8080/api/v1/...`

4. **Certificate trusted?**
   ```bash
   mkcert -install  # Re-install CA
   ```

### Problem: CORS errors after enabling HTTPS

**Backend CORS configuration may need update:**

```go
// backend/internal/config/env.go
CORS: CORSConfig{
    AllowedOrigins: []string{
        "http://localhost:3000",  // HTTP (current)
        "https://localhost:3000", // HTTPS (add this)
    },
}
```

---

## 🔐 Security Best Practices

### Development Certificates

✅ **DO:**
- Use mkcert for local development
- Keep private keys secure (add to .gitignore)
- Regenerate certificates periodically
- Use separate certificates per project

❌ **DON'T:**
- Commit certificates to git
- Share private keys
- Use development certificates in production
- Disable SSL verification in code

### .gitignore Configuration

```bash
# SSL Certificates (development only)
*.pem
*.key
*.crt
*.csr
localhost+*
```

### Production SSL

**For production, use:**
- [Let's Encrypt](https://letsencrypt.org/) (free, auto-renewal)
- Commercial SSL providers (Sectigo, DigiCert)
- Cloud provider SSL (AWS Certificate Manager, Cloudflare)

**Never use self-signed certificates in production!**

---

## 📊 SSL Development Checklist

### Initial Setup
- [ ] Install mkcert
- [ ] Run `mkcert -install`
- [ ] Generate certificates for backend
- [ ] Generate certificates for frontend (optional)
- [ ] Add certificates to .gitignore

### Backend Configuration
- [ ] Update main.go with SSL logic
- [ ] Test HTTPS server start
- [ ] Verify certificate loading
- [ ] Test health endpoint with HTTPS

### Frontend Configuration
- [ ] Update .env.local with HTTPS URL
- [ ] Update .env.example with instructions
- [ ] Test API calls with HTTPS
- [ ] Verify no mixed content warnings

### Testing
- [ ] Browser shows secure connection (padlock icon)
- [ ] HSTS header present in responses
- [ ] API calls working without SSL errors
- [ ] WebSocket connections working (wss://)

### Documentation
- [ ] SSL setup instructions for team
- [ ] Troubleshooting guide
- [ ] Certificate renewal procedures

---

## 🎓 Understanding HSTS

### How HSTS Works

1. **First Visit (HTTPS):**
   ```
   User → https://localhost:8080
   Server → Response with: Strict-Transport-Security: max-age=31536000
   Browser → Remember: "Always use HTTPS for localhost for 1 year"
   ```

2. **Subsequent Visits:**
   ```
   User types → http://localhost:8080
   Browser (before request) → "HSTS active! Convert to HTTPS"
   Browser → https://localhost:8080 (no HTTP request sent)
   ```

### HSTS Benefits

✅ **Prevents:**
- SSL Stripping attacks
- Man-in-the-Middle (MITM)
- Accidental HTTP access
- Mixed content issues

### HSTS Risks in Development

⚠️ **If SSL breaks:**
- Site becomes completely inaccessible
- Must wait for max-age to expire OR
- Clear HSTS from browser manually

**Clear HSTS (Chrome):**
1. Visit: `chrome://net-internals/#hsts`
2. "Delete domain security policies"
3. Enter: `localhost`
4. Click "Delete"

---

## 🚀 Next Steps

After SSL setup complete:

1. ✅ **Enable HSTS in Backend**
   ```bash
   # backend/.env
   SECURITY_ENABLE_HSTS=true
   ```

2. ✅ **Update Frontend URL**
   ```bash
   # frontend/.env.local
   NEXT_PUBLIC_API_URL=https://localhost:8080
   ```

3. ✅ **Test Everything**
   - Login flow
   - API calls
   - File uploads
   - WebSocket connections

4. ✅ **Monitor for Issues**
   - Check browser console
   - Check backend logs
   - Test on all browsers

5. ⏳ **Prepare for Phase 3 (CSP)**
   - See: backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md
   - Frontend audit for inline scripts/styles

---

## 📚 Resources

**mkcert:**
- [GitHub](https://github.com/FiloSottile/mkcert)
- [Documentation](https://mkcert.dev/)

**SSL/TLS:**
- [MDN: Transport Layer Security](https://developer.mozilla.org/en-US/docs/Web/Security/Transport_Layer_Security)
- [HSTS Specification](https://tools.ietf.org/html/rfc6797)

**Testing:**
- [SSL Labs Server Test](https://www.ssllabs.com/ssltest/)
- [SecurityHeaders.com](https://securityheaders.com/)

---

**Questions?** Refer to backend documentation: `backend/docs/SECURITY-HEADERS-IMPLEMENTATION.md`
