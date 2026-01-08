#!/bin/bash

# CSP Verification Script
# Run this to verify CSP Report-Only mode is active

echo "=========================================="
echo "🔍 CSP Report-Only Mode Verification"
echo "=========================================="
echo ""

# Test 1: Check if backend is running
echo "Test 1: Checking if backend is running..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Backend is running on port 8080"
else
    echo "❌ Backend is NOT running. Please start backend first."
    exit 1
fi
echo ""

# Test 2: Check for CSP Report-Only header
echo "Test 2: Checking for CSP Report-Only header..."
CSP_HEADER=$(curl -s -I http://localhost:8080/health | grep -i "content-security-policy-report-only")

if [ -n "$CSP_HEADER" ]; then
    echo "✅ CSP Report-Only header found!"
    echo ""
    echo "Header content:"
    echo "$CSP_HEADER"
else
    echo "❌ CSP Report-Only header NOT found."
    echo ""
    echo "Checking if enforcement mode is active instead..."
    CSP_ENFORCE=$(curl -s -I http://localhost:8080/health | grep -i "content-security-policy:" | grep -v "report-only")

    if [ -n "$CSP_ENFORCE" ]; then
        echo "⚠️  CSP is in ENFORCEMENT mode (not Report-Only)"
        echo "Please check backend/.env:"
        echo "SECURITY_CSP_REPORT_ONLY should be true"
    else
        echo "❌ No CSP header found at all."
        echo "Please check backend/.env:"
        echo "SECURITY_ENABLE_CSP should be true"
        echo "SECURITY_CSP_REPORT_ONLY should be true"
    fi
fi
echo ""

# Test 3: Check Phase 1 headers
echo "Test 3: Checking Phase 1 security headers..."
HEADERS=$(curl -s -I http://localhost:8080/health)

if echo "$HEADERS" | grep -q "X-Frame-Options"; then
    echo "✅ X-Frame-Options: Present"
else
    echo "❌ X-Frame-Options: Missing"
fi

if echo "$HEADERS" | grep -q "X-Content-Type-Options"; then
    echo "✅ X-Content-Type-Options: Present"
else
    echo "❌ X-Content-Type-Options: Missing"
fi

if echo "$HEADERS" | grep -q "X-XSS-Protection"; then
    echo "✅ X-XSS-Protection: Present"
else
    echo "❌ X-XSS-Protection: Missing"
fi

if echo "$HEADERS" | grep -q "Referrer-Policy"; then
    echo "✅ Referrer-Policy: Present"
else
    echo "❌ Referrer-Policy: Missing"
fi

if echo "$HEADERS" | grep -q "Permissions-Policy"; then
    echo "✅ Permissions-Policy: Present"
else
    echo "❌ Permissions-Policy: Missing"
fi
echo ""

# Test 4: Check HSTS (should be disabled)
echo "Test 4: Checking HSTS status..."
if echo "$HEADERS" | grep -q "Strict-Transport-Security"; then
    echo "⚠️  HSTS is ENABLED (Phase 2)"
    echo "This requires SSL certificate setup"
else
    echo "✅ HSTS is disabled (expected - waiting for SSL)"
fi
echo ""

# Summary
echo "=========================================="
echo "📊 Summary"
echo "=========================================="
if [ -n "$CSP_HEADER" ]; then
    echo "✅ CSP Report-Only mode: ACTIVE"
    echo ""
    echo "🎯 Next steps:"
    echo "1. Open frontend: http://localhost:3000"
    echo "2. Open DevTools (F12) → Console tab"
    echo "3. Test all features and check for CSP warnings"
    echo "4. Monitor for 1 week"
    echo ""
    echo "Expected result: No CSP warnings (except browser extensions)"
else
    echo "❌ CSP Report-Only mode: NOT ACTIVE"
    echo ""
    echo "🔧 Troubleshooting:"
    echo "1. Check backend/.env file"
    echo "2. Verify SECURITY_ENABLE_CSP=true"
    echo "3. Verify SECURITY_CSP_REPORT_ONLY=true"
    echo "4. Restart backend server"
    echo "5. Run this script again"
fi
echo "=========================================="
