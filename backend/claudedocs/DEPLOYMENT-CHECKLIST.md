# Deployment Checklist - Phase 2 Authentication System

**Version:** 1.0
**Last Updated:** 2025-12-17
**Target Environment:** Production

---

## Pre-Deployment Checklist

### 1. Code Quality & Testing ✅

- [ ] All unit tests passing (CSRF, validators, brute force)
  ```bash
  go test ./internal/middleware ./pkg/validator ./internal/service/auth -v
  ```

- [ ] Integration tests passing
  ```bash
  go test ./internal/handler -v
  ```

- [ ] Code coverage ≥80%
  ```bash
  go test -cover ./...
  ```

- [ ] No compiler warnings or errors
  ```bash
  go build ./...
  ```

- [ ] Linting passes
  ```bash
  golangci-lint run
  ```

- [ ] Security scan completed
  ```bash
  gosec ./...
  ```

### 2. Configuration & Environment Variables ⚙️

- [ ] Production environment file created (`.env.production`)

- [ ] Database credentials configured
  ```env
  DB_HOST=production-db-host
  DB_PORT=5432
  DB_NAME=erp_production
  DB_USER=app_user
  DB_PASSWORD=<strong-password>
  DB_SSL_MODE=require
  ```

- [ ] JWT secrets configured (32+ character random strings)
  ```env
  JWT_SECRET=<secure-random-string-32+>
  JWT_REFRESH_SECRET=<secure-random-string-32+>
  ```

- [ ] Redis credentials configured
  ```env
  REDIS_HOST=production-redis-host
  REDIS_PORT=6379
  REDIS_PASSWORD=<redis-password>
  REDIS_DB=0
  ```

- [ ] SMTP/Email service configured
  ```env
  SMTP_HOST=smtp.gmail.com
  SMTP_PORT=587
  SMTP_USERNAME=noreply@yourdomain.com
  SMTP_PASSWORD=<app-specific-password>
  SMTP_FROM_NAME=Your Company
  SMTP_FROM_EMAIL=noreply@yourdomain.com
  ```

- [ ] Password reset configuration
  ```env
  PASSWORD_RESET_EXPIRY=1h
  PASSWORD_RESET_BASE_URL=https://yourdomain.com/reset-password
  ```

- [ ] Security configuration
  ```env
  LOCKOUT_TIER1_ATTEMPTS=3
  LOCKOUT_TIER1_DURATION=5m
  LOCKOUT_TIER2_ATTEMPTS=5
  LOCKOUT_TIER2_DURATION=15m
  LOCKOUT_TIER3_ATTEMPTS=10
  LOCKOUT_TIER3_DURATION=1h
  LOCKOUT_TIER4_ATTEMPTS=15
  LOCKOUT_TIER4_DURATION=24h
  ```

- [ ] Rate limiting configuration
  ```env
  RATE_LIMIT_REQUESTS=100
  RATE_LIMIT_DURATION=1m
  RATE_LIMIT_AUTH_REQUESTS=10
  RATE_LIMIT_AUTH_DURATION=1m
  ```

### 3. Database Setup 🗄️

- [ ] Production database created and accessible

- [ ] Database migrations reviewed
  ```bash
  # Review all migrations
  ls -la db/migrations/
  ```

- [ ] Database migrations applied
  ```bash
  # Dry run first
  npx prisma migrate deploy --preview-feature

  # Then apply
  npx prisma migrate deploy
  ```

- [ ] Database indexes verified
  ```sql
  -- Check critical indexes
  SELECT * FROM pg_indexes WHERE schemaname = 'public';
  ```

- [ ] Database backups configured
  - Automated daily backups
  - Backup retention policy (30 days minimum)
  - Backup restore tested

- [ ] Connection pooling configured
  ```env
  DB_MAX_CONNECTIONS=20
  DB_IDLE_CONNECTIONS=5
  DB_MAX_LIFETIME=1h
  ```

### 4. Redis Setup 🔴

- [ ] Production Redis instance created

- [ ] Redis persistence configured (AOF or RDB)

- [ ] Redis memory limits set
  ```
  maxmemory 2gb
  maxmemory-policy allkeys-lru
  ```

- [ ] Redis password authentication enabled

- [ ] Redis SSL/TLS enabled (if supported)

### 5. Email Service Setup ✉️

- [ ] SMTP service configured and tested
  - Gmail App Password created (if using Gmail)
  - Mailtrap API key configured (if using Mailtrap)
  - SendGrid API key configured (if using SendGrid)

- [ ] Email templates reviewed
  ```bash
  # Check template files exist
  ls -la pkg/email/templates/
  ```

- [ ] Test email sent successfully
  ```bash
  # Use SMTP test tool or write a test script
  ```

- [ ] SPF, DKIM, DMARC records configured for domain
  - SPF record added to DNS
  - DKIM signature configured
  - DMARC policy configured

- [ ] Email rate limits reviewed
  - Daily sending limit
  - Per-hour sending limit

### 6. SSL/TLS Certificates 🔒

- [ ] SSL certificate obtained (Let's Encrypt, commercial CA, etc.)

- [ ] Certificate installed on server

- [ ] Certificate auto-renewal configured

- [ ] HTTPS enforced (HTTP redirects to HTTPS)

- [ ] TLS 1.2+ only (disable TLS 1.0/1.1)

- [ ] Strong cipher suites configured
  ```
  TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
  TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
  ```

### 7. Security Hardening 🛡️

- [ ] Secrets never committed to git
  ```bash
  # Verify .env files are in .gitignore
  git ls-files | grep ".env"
  ```

- [ ] Sensitive data encrypted at rest
  - Database encryption enabled
  - Backup encryption enabled

- [ ] Firewall rules configured
  - Allow only necessary ports (80, 443, 22)
  - Restrict database/Redis access to app servers only
  - Block direct external access to internal services

- [ ] Security headers configured in middleware
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - X-XSS-Protection: 1; mode=block
  - Strict-Transport-Security: max-age=31536000; includeSubDomains
  - Content-Security-Policy configured

- [ ] CORS properly configured
  - Allowed origins whitelisted
  - Credentials allowed only for trusted origins

- [ ] Rate limiting verified
  ```bash
  # Test rate limit with load testing tool
  ab -n 200 -c 10 https://yourdomain.com/api/v1/auth/login
  ```

### 8. Monitoring & Logging 📊

- [ ] Application logging configured
  - Log level: INFO for production
  - Log rotation configured
  - Structured logging enabled (JSON format)

- [ ] Security event logging enabled
  - Failed login attempts logged
  - Account lockouts logged
  - Password reset requests logged
  - Token refresh events logged

- [ ] Error tracking service configured
  - Sentry, Rollbar, or similar
  - Error notifications enabled

- [ ] Performance monitoring configured
  - Application Performance Monitoring (APM)
  - Database query monitoring
  - Redis monitoring

- [ ] Uptime monitoring configured
  - Pingdom, UptimeRobot, or similar
  - Alert thresholds configured
  - Notification channels configured

- [ ] Log aggregation configured
  - ELK Stack, Papertrail, or similar
  - Search and analysis capabilities
  - Retention policy configured

### 9. Infrastructure 🏗️

- [ ] Production server(s) provisioned
  - Sufficient CPU (2+ cores recommended)
  - Sufficient RAM (4GB+ recommended)
  - Sufficient disk space (20GB+ recommended)

- [ ] Load balancer configured (if using multiple servers)
  - Health check endpoints configured
  - Session affinity configured (if needed)
  - SSL termination configured

- [ ] Reverse proxy configured (Nginx, Caddy, etc.)
  - Request buffering configured
  - Timeout settings configured
  - Rate limiting at proxy level

- [ ] Auto-scaling configured (if applicable)
  - CPU threshold: 70%
  - Memory threshold: 80%
  - Scale-up and scale-down policies

- [ ] Backup server/failover configured (if required)

### 10. Documentation 📚

- [ ] API documentation reviewed and published
  - Base URLs updated for production
  - Example requests tested
  - Error responses documented

- [ ] Deployment runbook created
  - Step-by-step deployment instructions
  - Rollback procedures documented
  - Emergency contacts listed

- [ ] Architecture diagram updated
  - Infrastructure topology
  - Data flow diagrams
  - Security boundaries

- [ ] Operational playbooks created
  - Incident response procedures
  - Troubleshooting guides
  - Common issues and solutions

---

## Deployment Steps

### Phase 1: Pre-Deployment Verification ✅

**Timeline:** Day -7 to Day -1

1. **Code Freeze** (Day -7)
   ```bash
   # Create release branch
   git checkout -b release/v1.0.0
   git push origin release/v1.0.0
   ```

2. **Final Testing** (Day -5 to Day -3)
   - Run full test suite
   - Perform security scan
   - Execute load tests
   - Test email delivery
   - Verify all environment variables

3. **Staging Deployment** (Day -3)
   ```bash
   # Deploy to staging environment
   ./deploy-staging.sh

   # Verify deployment
   curl https://staging.yourdomain.com/api/v1/health
   ```

4. **Staging Validation** (Day -3 to Day -1)
   - Test registration flow
   - Test login/logout flow
   - Test password reset flow
   - Test CSRF protection
   - Test rate limiting
   - Test brute force protection
   - Verify email delivery
   - Verify token refresh
   - Load test staging environment

5. **Final Approval** (Day -1)
   - Stakeholder review
   - Security team approval
   - Operations team ready

### Phase 2: Production Deployment 🚀

**Timeline:** Day 0

1. **Pre-Deployment Snapshot** (T-60 minutes)
   ```bash
   # Backup production database
   pg_dump -U $DB_USER -h $DB_HOST $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql

   # Snapshot current deployment
   git tag deployment/$(date +%Y%m%d_%H%M%S)
   git push --tags
   ```

2. **Maintenance Mode** (T-30 minutes)
   ```bash
   # Enable maintenance page (if applicable)
   ./enable-maintenance.sh
   ```

3. **Database Migration** (T-20 minutes)
   ```bash
   # Apply migrations
   npx prisma migrate deploy

   # Verify migrations
   npx prisma migrate status
   ```

4. **Application Deployment** (T-10 minutes)
   ```bash
   # Build production binary
   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/erp-backend ./cmd/api

   # Deploy to production
   ./deploy-production.sh

   # Verify deployment
   curl https://yourdomain.com/api/v1/health
   ```

5. **Smoke Tests** (T+0 minutes)
   ```bash
   # Test critical endpoints
   ./smoke-tests.sh
   ```

6. **Disable Maintenance Mode** (T+10 minutes)
   ```bash
   ./disable-maintenance.sh
   ```

7. **Post-Deployment Verification** (T+15 minutes)
   - Verify registration works
   - Verify login works
   - Verify logout works
   - Verify password reset works
   - Check error logs
   - Check monitoring dashboards
   - Verify email delivery

### Phase 3: Post-Deployment Monitoring 👀

**Timeline:** Day 0 to Day 7

1. **Immediate Monitoring** (First 2 hours)
   - Monitor error rates (target: <0.1%)
   - Monitor response times (target: <200ms p95)
   - Monitor CPU/memory usage
   - Monitor database connections
   - Monitor Redis operations
   - Check for failed login spikes

2. **Short-Term Monitoring** (First 24 hours)
   - Review application logs
   - Review security event logs
   - Review email delivery logs
   - Check for unusual patterns
   - Monitor user feedback

3. **Medium-Term Monitoring** (First 7 days)
   - Weekly metrics review
   - Performance trend analysis
   - Security incident review
   - User experience feedback
   - Optimization opportunities

---

## Rollback Procedures

### Emergency Rollback

**When to Rollback:**
- Critical security vulnerability discovered
- Data corruption detected
- Service completely unavailable
- Error rate >5%

**Rollback Steps:**

1. **Immediate Actions** (T+0)
   ```bash
   # Enable maintenance mode
   ./enable-maintenance.sh

   # Stop current deployment
   systemctl stop erp-backend
   ```

2. **Restore Previous Version** (T+5)
   ```bash
   # Checkout previous stable tag
   git checkout deployment/<previous_timestamp>

   # Rebuild and deploy
   ./deploy-production.sh
   ```

3. **Database Rollback** (T+10, if needed)
   ```bash
   # Restore database from backup
   psql -U $DB_USER -h $DB_HOST -d $DB_NAME < backup_<timestamp>.sql
   ```

4. **Verification** (T+15)
   ```bash
   # Run smoke tests
   ./smoke-tests.sh

   # Verify critical flows
   ```

5. **Resume Service** (T+20)
   ```bash
   # Disable maintenance mode
   ./disable-maintenance.sh
   ```

6. **Post-Rollback Actions**
   - Notify stakeholders
   - Document rollback reason
   - Create incident report
   - Plan fix and redeployment

---

## Smoke Test Script

Create `smoke-tests.sh`:

```bash
#!/bin/bash

BASE_URL="https://yourdomain.com/api/v1/auth"

echo "Running smoke tests..."

# Test 1: Health check
echo "1. Testing health endpoint..."
HEALTH=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/../health)
if [ $HEALTH -ne 200 ]; then
  echo "❌ Health check failed: $HEALTH"
  exit 1
fi
echo "✅ Health check passed"

# Test 2: Registration
echo "2. Testing registration..."
REGISTER=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"TestPass123","fullName":"Test User"}')
if [ $REGISTER -ne 201 ] && [ $REGISTER -ne 409 ]; then
  echo "❌ Registration failed: $REGISTER"
  exit 1
fi
echo "✅ Registration passed"

# Test 3: Login
echo "3. Testing login..."
LOGIN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"TestPass123"}')
if ! echo $LOGIN | grep -q "accessToken"; then
  echo "❌ Login failed"
  exit 1
fi
echo "✅ Login passed"

# Test 4: Refresh token
echo "4. Testing refresh endpoint..."
REFRESH=$(curl -s -o /dev/null -w "%{http_code}" -X POST $BASE_URL/refresh)
# 401 is expected if no valid refresh token
if [ $REFRESH -ne 200 ] && [ $REFRESH -ne 401 ]; then
  echo "❌ Refresh endpoint error: $REFRESH"
  exit 1
fi
echo "✅ Refresh endpoint available"

echo "✅ All smoke tests passed!"
```

Make it executable:
```bash
chmod +x smoke-tests.sh
```

---

## Performance Benchmarks

### Target Metrics

**Response Times (95th percentile):**
- Registration: <500ms
- Login: <300ms
- Logout: <100ms
- Refresh: <200ms
- Forgot Password: <300ms
- Reset Password: <400ms

**Throughput:**
- Registration: 100 req/sec
- Login: 200 req/sec
- Refresh: 500 req/sec

**Error Rates:**
- Target: <0.1%
- Warning: >0.5%
- Critical: >1%

**Resource Usage:**
- CPU: <70% average
- Memory: <80% average
- Database connections: <50% of pool
- Redis memory: <80% of limit

### Load Testing

```bash
# Install Apache Bench (ab) or use alternative like k6, wrk

# Test login endpoint
ab -n 1000 -c 10 -p login.json -T application/json \
   https://yourdomain.com/api/v1/auth/login

# login.json:
# {"email":"test@example.com","password":"TestPass123"}
```

---

## Monitoring Alerts

### Critical Alerts (Immediate Response)

1. **Service Down**
   - Trigger: Health check fails for 2 minutes
   - Action: Page on-call engineer

2. **High Error Rate**
   - Trigger: Error rate >1% for 5 minutes
   - Action: Page on-call engineer

3. **Database Connection Failure**
   - Trigger: Database unreachable for 1 minute
   - Action: Page on-call engineer and DBA

4. **Redis Connection Failure**
   - Trigger: Redis unreachable for 2 minutes
   - Action: Alert operations team

### Warning Alerts (Business Hours Response)

1. **Elevated Error Rate**
   - Trigger: Error rate >0.5% for 10 minutes
   - Action: Notify development team

2. **Slow Response Times**
   - Trigger: p95 response time >1s for 10 minutes
   - Action: Notify development team

3. **High CPU Usage**
   - Trigger: CPU >80% for 15 minutes
   - Action: Notify operations team

4. **High Memory Usage**
   - Trigger: Memory >85% for 15 minutes
   - Action: Notify operations team

5. **Failed Email Delivery**
   - Trigger: Email send failure rate >5% for 30 minutes
   - Action: Notify operations team

---

## Security Monitoring

### Security Events to Monitor

1. **Brute Force Attacks**
   - Monitor: Failed login attempts per IP/email
   - Threshold: 10+ failed attempts in 1 hour from same IP
   - Action: Review logs, consider IP blocking

2. **Account Lockouts**
   - Monitor: Tier 3/4 lockouts
   - Threshold: 5+ Tier 3+ lockouts in 1 hour
   - Action: Investigate potential attack

3. **Password Reset Abuse**
   - Monitor: Password reset requests per IP
   - Threshold: 20+ requests in 1 hour from same IP
   - Action: Review logs, consider IP blocking

4. **CSRF Failures**
   - Monitor: CSRF validation failures
   - Threshold: 50+ failures in 1 hour
   - Action: Investigate potential attack or integration issue

5. **Unusual Token Activity**
   - Monitor: Refresh token failures
   - Threshold: Spike in refresh failures (>10% of normal)
   - Action: Investigate potential token compromise

---

## Compliance & Auditing

### Data Protection Compliance

- [ ] GDPR compliance reviewed (if applicable)
  - Right to erasure implemented
  - Data portability supported
  - Consent management configured
  - Privacy policy updated

- [ ] User data retention policy configured
  - Inactive account handling
  - Deleted account data purging
  - Audit log retention

### Security Compliance

- [ ] Password policy complies with NIST guidelines
  - Minimum 8 characters
  - Complexity requirements
  - No password expiration

- [ ] Encryption standards meet requirements
  - At rest: AES-256
  - In transit: TLS 1.2+
  - Password hashing: Argon2id

### Audit Logging

- [ ] Security events logged
  - Authentication attempts
  - Authorization failures
  - Configuration changes
  - Administrative actions

- [ ] Audit logs immutable
  - Write-once storage
  - Tampering protection
  - Regular backups

---

## Post-Deployment Tasks

### Week 1

- [ ] Review performance metrics
- [ ] Review security logs
- [ ] Review user feedback
- [ ] Address any urgent issues
- [ ] Update documentation based on learnings

### Week 2

- [ ] Conduct retrospective meeting
- [ ] Document lessons learned
- [ ] Plan optimization improvements
- [ ] Review monitoring alerts (tune thresholds)

### Month 1

- [ ] Performance optimization based on data
- [ ] Security hardening based on logs
- [ ] User experience improvements
- [ ] Infrastructure scaling adjustments

---

## Contacts & Escalation

### On-Call Rotation

- **Primary On-Call:** [Name, Phone, Email]
- **Secondary On-Call:** [Name, Phone, Email]
- **Escalation Manager:** [Name, Phone, Email]

### Team Contacts

- **Development Lead:** [Contact Info]
- **Operations Lead:** [Contact Info]
- **Security Lead:** [Contact Info]
- **Database Admin:** [Contact Info]

### External Contacts

- **Hosting Provider:** [Support Contact]
- **Email Service:** [Support Contact]
- **Monitoring Service:** [Support Contact]

---

## Appendix

### Environment Variables Reference

Complete list of required environment variables:

```env
# Server Configuration
PORT=8080
ENV=production
LOG_LEVEL=info

# Database
DB_HOST=production-db-host
DB_PORT=5432
DB_NAME=erp_production
DB_USER=app_user
DB_PASSWORD=<strong-password>
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=20
DB_IDLE_CONNECTIONS=5
DB_MAX_LIFETIME=1h

# Redis
REDIS_HOST=production-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=<redis-password>
REDIS_DB=0

# JWT
JWT_SECRET=<secure-random-string-32+>
JWT_REFRESH_SECRET=<secure-random-string-32+>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h  # 7 days

# SMTP/Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=<app-specific-password>
SMTP_FROM_NAME=Your Company
SMTP_FROM_EMAIL=noreply@yourdomain.com

# Password Reset
PASSWORD_RESET_EXPIRY=1h
PASSWORD_RESET_BASE_URL=https://yourdomain.com/reset-password

# Security - Brute Force Protection
LOCKOUT_TIER1_ATTEMPTS=3
LOCKOUT_TIER1_DURATION=5m
LOCKOUT_TIER2_ATTEMPTS=5
LOCKOUT_TIER2_DURATION=15m
LOCKOUT_TIER3_ATTEMPTS=10
LOCKOUT_TIER3_DURATION=1h
LOCKOUT_TIER4_ATTEMPTS=15
LOCKOUT_TIER4_DURATION=24h

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=1m
RATE_LIMIT_AUTH_REQUESTS=10
RATE_LIMIT_AUTH_DURATION=1m

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
CORS_ALLOW_CREDENTIALS=true

# Monitoring (optional)
SENTRY_DSN=<sentry-dsn>
APM_SERVICE_NAME=erp-backend
APM_ENVIRONMENT=production
```

### Useful Commands

**Check Service Status:**
```bash
systemctl status erp-backend
```

**View Logs:**
```bash
journalctl -u erp-backend -f
```

**Database Connection Test:**
```bash
psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "SELECT version();"
```

**Redis Connection Test:**
```bash
redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping
```

**SMTP Connection Test:**
```bash
telnet $SMTP_HOST $SMTP_PORT
```

---

**Checklist Version:** 1.0
**Last Updated:** 2025-12-17
**Maintained By:** Operations Team
**Next Review:** Before each deployment
