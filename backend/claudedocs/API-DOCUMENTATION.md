# Authentication API Documentation

**Version:** 1.0
**Base URL:** `/api/v1/auth`
**Last Updated:** 2025-12-17

---

## Table of Contents

1. [Authentication Flow](#authentication-flow)
2. [Security Features](#security-features)
3. [Endpoints](#endpoints)
   - [Registration](#post-register)
   - [Login](#post-login)
   - [Logout](#post-logout)
   - [Refresh Token](#post-refresh)
   - [Forgot Password](#post-forgot-password)
   - [Reset Password](#post-reset-password)
4. [Error Responses](#error-responses)
5. [Rate Limiting](#rate-limiting)
6. [CSRF Protection](#csrf-protection)

---

## Authentication Flow

### New User Registration
1. Client sends registration request to `POST /register`
2. Server validates input (email format, password strength, phone number)
3. Server creates user account with hashed password (Argon2id)
4. Server returns access token (JWT) and sets refresh token cookie

### Login Flow
1. Client sends credentials to `POST /login`
2. Server validates credentials and checks brute force protection
3. Server generates access token (JWT) and refresh token
4. Server sets httpOnly refresh token cookie and CSRF cookie
5. Client stores access token and includes CSRF token in subsequent requests

### Token Refresh Flow
1. Client sends request to `POST /refresh` with refresh token cookie
2. Server validates refresh token and checks if revoked
3. Server generates new access token and refresh token
4. Server sets new refresh token cookie
5. Client updates access token

### Password Reset Flow
1. Client requests password reset at `POST /forgot-password`
2. Server generates secure reset token and sends email
3. User clicks email link and submits new password to `POST /reset-password`
4. Server validates token, updates password, revokes all refresh tokens
5. User must login again with new password

---

## Security Features

### Implemented Security Measures

**Password Security:**
- Argon2id hashing (memory-hard, resistant to GPU attacks)
- Password strength validation (8+ chars, uppercase, lowercase, digit)
- Secure password reset tokens (32 bytes, single-use, 1-hour expiry)

**Token Security:**
- JWT access tokens (short-lived, 15 minutes)
- Refresh tokens (httpOnly cookies, 7 days)
- SHA-256 token hashing for storage
- Token revocation support

**CSRF Protection:**
- Double-submit cookie pattern
- CSRF token required for state-changing requests
- Constant-time token comparison
- SameSite=Strict cookies

**Brute Force Protection:**
- 4-tier exponential backoff system
- Email + IP address tracking
- Progressive lockout durations:
  - Tier 1 (3-4 attempts): 5 minutes
  - Tier 2 (5-9 attempts): 15 minutes
  - Tier 3 (10-14 attempts): 1 hour
  - Tier 4 (15+ attempts): 24 hours

**Rate Limiting:**
- IP-based rate limiting via Redis
- 100 requests/minute for general endpoints
- 10 requests/minute for password reset endpoints
- 429 status code when limit exceeded

---

## Endpoints

### POST /register

Create a new user account.

**Authentication:** None (public endpoint)
**Rate Limit:** 100 requests/minute per IP
**CSRF Required:** No

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "SecurePass123",
  "fullName": "John Doe",
  "phoneNumber": "081234567890"
}
```

**Field Validations:**
- `email` (required): Valid email address, max 255 chars
- `password` (required): 8-72 chars, must contain uppercase, lowercase, and digit
- `fullName` (required): 2-255 chars
- `phoneNumber` (optional): Indonesian format (08xxxxxxxxxx or +628xxxxxxxxxx)

#### Success Response (201 Created)

```json
{
  "status": "success",
  "data": {
    "user": {
      "id": "uuid-string",
      "email": "user@example.com",
      "fullName": "John Doe",
      "phoneNumber": "081234567890",
      "isActive": true,
      "createdAt": "2025-12-17T10:30:00Z"
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Cookies Set:**
- `refresh_token` (httpOnly, secure, SameSite=Strict, 7 days)
- `csrf_token` (secure, SameSite=Strict, 24 hours)

#### Error Responses

**400 Bad Request** - Validation failed
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: email - email must be a valid email address; password - password must contain at least one uppercase letter, one lowercase letter, and one digit;"
  }
}
```

**409 Conflict** - Email already exists
```json
{
  "status": "error",
  "error": {
    "code": "CONFLICT",
    "message": "Email already registered"
  }
}
```

---

### POST /login

Authenticate user and obtain access token.

**Authentication:** None (public endpoint)
**Rate Limit:** 100 requests/minute per IP
**CSRF Required:** No

#### Request Body

```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Field Validations:**
- `email` (required): Valid email address
- `password` (required): Non-empty string

#### Success Response (200 OK)

```json
{
  "status": "success",
  "data": {
    "user": {
      "id": "uuid-string",
      "email": "user@example.com",
      "fullName": "John Doe",
      "phoneNumber": "081234567890",
      "isActive": true
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Cookies Set:**
- `refresh_token` (httpOnly, secure, SameSite=Strict, 7 days)
- `csrf_token` (secure, SameSite=Strict, 24 hours)

#### Error Responses

**400 Bad Request** - Invalid credentials
```json
{
  "status": "error",
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid email or password"
  }
}
```

**403 Forbidden** - Account locked (brute force protection)
```json
{
  "status": "error",
  "error": {
    "code": "ACCOUNT_LOCKED",
    "message": "Account locked (Tier 2). Too many failed login attempts (7). Please try again in 845 seconds."
  }
}
```

**403 Forbidden** - Account inactive
```json
{
  "status": "error",
  "error": {
    "code": "AUTHORIZATION_ERROR",
    "message": "Account is not active"
  }
}
```

---

### POST /logout

Revoke refresh token and end session.

**Authentication:** Required (access token)
**Rate Limit:** 100 requests/minute per IP
**CSRF Required:** Yes

#### Request Headers

```
Authorization: Bearer <access_token>
X-CSRF-Token: <csrf_token>
```

#### Success Response (200 OK)

```json
{
  "status": "success",
  "message": "Logged out successfully"
}
```

**Cookies Cleared:**
- `refresh_token` (set to empty with MaxAge=-1)
- `csrf_token` (set to empty with MaxAge=-1)

#### Error Responses

**401 Unauthorized** - Invalid or expired access token
```json
{
  "status": "error",
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid or expired token"
  }
}
```

**403 Forbidden** - Missing or invalid CSRF token
```json
{
  "status": "error",
  "error": {
    "code": "CSRF_ERROR",
    "message": "CSRF token mismatch"
  }
}
```

---

### POST /refresh

Obtain new access token using refresh token.

**Authentication:** Refresh token cookie
**Rate Limit:** 100 requests/minute per IP
**CSRF Required:** No

#### Request

No request body required. Refresh token sent via httpOnly cookie.

#### Success Response (200 OK)

```json
{
  "status": "success",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Cookies Set:**
- `refresh_token` (new token, httpOnly, secure, SameSite=Strict, 7 days)

#### Error Responses

**401 Unauthorized** - Missing refresh token
```json
{
  "status": "error",
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Refresh token not found"
  }
}
```

**401 Unauthorized** - Invalid or revoked refresh token
```json
{
  "status": "error",
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid or expired refresh token"
  }
}
```

---

### POST /forgot-password

Request password reset email.

**Authentication:** None (public endpoint)
**Rate Limit:** 10 requests/minute per IP
**CSRF Required:** No

#### Request Body

```json
{
  "email": "user@example.com"
}
```

**Field Validations:**
- `email` (required): Valid email address

#### Success Response (200 OK)

```json
{
  "status": "success",
  "message": "If the email exists, a password reset link has been sent"
}
```

**Note:** This endpoint always returns success to prevent email enumeration attacks, even if the email doesn't exist.

#### Rate Limiting

**Per Email:** Maximum 3 requests per hour per email address
**Per IP:** 10 requests per minute

#### Error Responses

**400 Bad Request** - Invalid email format
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: email - email must be a valid email address;"
  }
}
```

**429 Too Many Requests** - Rate limit exceeded
```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests"
  }
}
```

#### Email Template

Subject: `Reset Your Password`

```
Hello,

You requested to reset your password. Click the link below to proceed:

https://yourdomain.com/reset-password?token=<reset_token>

This link will expire in 1 hour.

If you didn't request this, please ignore this email.
```

---

### POST /reset-password

Reset password using token from email.

**Authentication:** None (public endpoint)
**Rate Limit:** 10 requests/minute per IP
**CSRF Required:** No

#### Request Body

```json
{
  "token": "secure-reset-token-from-email",
  "newPassword": "NewSecurePass123"
}
```

**Field Validations:**
- `token` (required): Non-empty string (reset token from email)
- `newPassword` (required): 8-72 chars, must contain uppercase, lowercase, and digit

#### Success Response (200 OK)

```json
{
  "status": "success",
  "message": "Password reset successfully. Please login with your new password."
}
```

**Side Effects:**
- All refresh tokens for the user are revoked
- User must login again with new password
- Reset token is marked as used (single-use)

#### Error Responses

**400 Bad Request** - Invalid or expired token
```json
{
  "status": "error",
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid or expired reset token"
  }
}
```

**400 Bad Request** - Token already used
```json
{
  "status": "error",
  "error": {
    "code": "BAD_REQUEST",
    "message": "Reset token has already been used"
  }
}
```

**400 Bad Request** - Validation failed
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed: newPassword - newPassword must contain at least one uppercase letter, one lowercase letter, and one digit;"
  }
}
```

---

## Error Responses

### Standard Error Format

All error responses follow this structure:

```json
{
  "status": "error",
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": [
      {
        "field": "fieldName",
        "message": "Field-specific error message"
      }
    ]
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `BAD_REQUEST` | 400 | Invalid request |
| `AUTHENTICATION_ERROR` | 401 | Invalid credentials or token |
| `AUTHORIZATION_ERROR` | 403 | Insufficient permissions |
| `ACCOUNT_LOCKED` | 403 | Too many failed login attempts |
| `CSRF_ERROR` | 403 | CSRF token validation failed |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Rate Limiting

Rate limits are applied per IP address using Redis-based tracking.

### Limits by Endpoint

| Endpoint | Rate Limit | Window |
|----------|------------|--------|
| `/register` | 100 requests | 1 minute |
| `/login` | 100 requests | 1 minute |
| `/logout` | 100 requests | 1 minute |
| `/refresh` | 100 requests | 1 minute |
| `/forgot-password` | 10 requests | 1 minute |
| `/reset-password` | 10 requests | 1 minute |

### Additional Rate Limits

**Forgot Password (Email-based):** 3 requests per hour per email address

### Rate Limit Headers

When rate limit is approached or exceeded:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 42
X-RateLimit-Reset: 1639584000
```

### Rate Limit Exceeded Response (429)

```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests"
  }
}
```

---

## CSRF Protection

### Overview

CSRF (Cross-Site Request Forgery) protection uses the **double-submit cookie pattern**:

1. Server sets a CSRF token in a cookie (NOT httpOnly)
2. Client reads the cookie value
3. Client sends the same value in `X-CSRF-Token` header
4. Server validates cookie matches header

### When CSRF is Required

CSRF protection is required for all **state-changing** requests:

- ✅ `POST /logout` - Requires CSRF token
- ✅ Any authenticated endpoint that modifies data
- ❌ `POST /register` - Public endpoint, no CSRF
- ❌ `POST /login` - Public endpoint, no CSRF
- ❌ `POST /refresh` - Public endpoint, no CSRF
- ❌ `POST /forgot-password` - Public endpoint, no CSRF
- ❌ `POST /reset-password` - Public endpoint, no CSRF
- ❌ `GET` requests - Safe methods, no CSRF needed

### How to Use CSRF Tokens

#### 1. Obtain CSRF Token

After successful login, read the `csrf_token` cookie:

```javascript
// JavaScript example
const csrfToken = document.cookie
  .split('; ')
  .find(row => row.startsWith('csrf_token='))
  ?.split('=')[1];
```

#### 2. Include in Requests

Send the token in the `X-CSRF-Token` header:

```javascript
fetch('/api/v1/auth/logout', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'X-CSRF-Token': csrfToken
  },
  credentials: 'include' // Important: include cookies
})
```

#### 3. Handle CSRF Errors

```javascript
if (response.status === 403) {
  const data = await response.json();
  if (data.error.code === 'CSRF_ERROR') {
    // CSRF token invalid or missing
    // Redirect to login or refresh page
  }
}
```

### CSRF Security Features

- **Constant-time comparison** prevents timing attacks
- **SameSite=Strict** prevents cross-site cookie sending
- **24-hour expiry** limits token lifetime
- **Regenerated on login** ensures fresh tokens

### CSRF Error Response (403)

```json
{
  "status": "error",
  "error": {
    "code": "CSRF_ERROR",
    "message": "CSRF token mismatch"
  }
}
```

---

## Authentication Headers

### Access Token (JWT)

Include in `Authorization` header for protected endpoints:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token Lifetime:** 15 minutes
**Claims:**
```json
{
  "user_id": "uuid-string",
  "email": "user@example.com",
  "exp": 1639584900,
  "iat": 1639584000
}
```

### Refresh Token

Sent automatically via httpOnly cookie, not accessible to JavaScript.

**Cookie Name:** `refresh_token`
**Lifetime:** 7 days
**Security:** httpOnly, secure, SameSite=Strict

### CSRF Token

Sent automatically via cookie, readable by JavaScript for header inclusion.

**Cookie Name:** `csrf_token`
**Lifetime:** 24 hours
**Security:** secure, SameSite=Strict (NOT httpOnly)

---

## Examples

### Complete Registration Flow

```javascript
// 1. Register new user
const registerResponse = await fetch('/api/v1/auth/register', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123',
    fullName: 'John Doe',
    phoneNumber: '081234567890'
  }),
  credentials: 'include' // Important for cookies
});

const { data } = await registerResponse.json();

// 2. Store access token
localStorage.setItem('accessToken', data.accessToken);

// 3. Extract CSRF token from cookie
const csrfToken = document.cookie
  .split('; ')
  .find(row => row.startsWith('csrf_token='))
  ?.split('=')[1];

// 4. Make authenticated request
const protectedResponse = await fetch('/api/v1/protected-endpoint', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${data.accessToken}`,
    'X-CSRF-Token': csrfToken,
    'Content-Type': 'application/json'
  },
  credentials: 'include'
});
```

### Complete Login Flow

```javascript
// 1. Login
const loginResponse = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123'
  }),
  credentials: 'include'
});

const { data } = await loginResponse.json();

// 2. Store access token and CSRF token
localStorage.setItem('accessToken', data.accessToken);
const csrfToken = document.cookie
  .split('; ')
  .find(row => row.startsWith('csrf_token='))
  ?.split('=')[1];
localStorage.setItem('csrfToken', csrfToken);
```

### Token Refresh Flow

```javascript
// 1. Detect token expiration (e.g., 401 response)
async function refreshAccessToken() {
  const response = await fetch('/api/v1/auth/refresh', {
    method: 'POST',
    credentials: 'include' // Sends refresh_token cookie
  });

  if (response.ok) {
    const { data } = await response.json();
    localStorage.setItem('accessToken', data.accessToken);
    return data.accessToken;
  } else {
    // Refresh token invalid, redirect to login
    window.location.href = '/login';
  }
}

// 2. Retry failed request with new token
async function makeAuthenticatedRequest(url, options) {
  let response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${localStorage.getItem('accessToken')}`
    }
  });

  if (response.status === 401) {
    // Token expired, refresh and retry
    const newToken = await refreshAccessToken();
    response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${newToken}`
      }
    });
  }

  return response;
}
```

### Password Reset Flow

```javascript
// 1. Request password reset
const forgotResponse = await fetch('/api/v1/auth/forgot-password', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com'
  })
});

// 2. User clicks email link with token
// URL: https://yourdomain.com/reset-password?token=<reset_token>

// 3. Submit new password
const resetResponse = await fetch('/api/v1/auth/reset-password', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    token: '<reset_token_from_url>',
    newPassword: 'NewSecurePass123'
  })
});

// 4. Redirect to login
if (resetResponse.ok) {
  window.location.href = '/login';
}
```

### Logout Flow

```javascript
// 1. Logout (revoke refresh token)
const logoutResponse = await fetch('/api/v1/auth/logout', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('accessToken')}`,
    'X-CSRF-Token': localStorage.getItem('csrfToken')
  },
  credentials: 'include'
});

// 2. Clear local storage
localStorage.removeItem('accessToken');
localStorage.removeItem('csrfToken');

// 3. Redirect to login
window.location.href = '/login';
```

---

## Best Practices

### Client-Side Implementation

1. **Store Access Token Securely**
   - Use `localStorage` or `sessionStorage` (NOT cookies for access tokens)
   - Clear on logout

2. **Handle Token Expiration**
   - Implement automatic token refresh on 401 responses
   - Redirect to login if refresh fails

3. **Include Credentials**
   - Always use `credentials: 'include'` in fetch requests
   - Required for httpOnly cookies (refresh token)

4. **CSRF Token Management**
   - Read from cookie after login
   - Include in all state-changing requests
   - Handle CSRF errors gracefully

5. **Error Handling**
   - Check HTTP status codes
   - Parse error responses
   - Display user-friendly messages

### Security Best Practices

1. **Never Store Sensitive Data in LocalStorage**
   - Access tokens are acceptable (short-lived)
   - Never store passwords or refresh tokens

2. **Use HTTPS in Production**
   - Required for `secure` cookies
   - Protects tokens in transit

3. **Validate Input Client-Side**
   - Provide immediate feedback
   - Reduce unnecessary API calls
   - Server-side validation is still required

4. **Implement Rate Limiting Client-Side**
   - Prevent accidental DOS
   - Handle 429 responses gracefully

5. **Monitor Failed Login Attempts**
   - Display tier information to users
   - Show remaining lockout time
   - Provide password reset option

---

## Testing

### Example Test Cases

**Registration:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123",
    "fullName": "Test User",
    "phoneNumber": "081234567890"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123"
  }' \
  -c cookies.txt
```

**Refresh Token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

**Logout:**
```bash
# Extract access token from login response first
ACCESS_TOKEN="<token_from_login>"
CSRF_TOKEN="<token_from_cookie>"

curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -b cookies.txt
```

---

## Changelog

### Version 1.0 (2025-12-17)

**Initial Release:**
- User registration with email verification
- Login with JWT access tokens and refresh tokens
- Logout with token revocation
- Password reset via email
- CSRF protection with double-submit cookie pattern
- 4-tier brute force protection with exponential backoff
- Rate limiting on authentication endpoints
- Comprehensive input validation
- Argon2id password hashing
- SHA-256 token hashing

---

**Document Version:** 1.0
**Last Updated:** 2025-12-17
**Maintained By:** Backend Team
