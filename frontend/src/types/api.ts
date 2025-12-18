// API Type Definitions for Backend Integration
// Based on backend API documentation with response envelope pattern

/**
 * User information returned from authentication
 */
export interface User {
  id: string;
  email: string;
  fullName: string;
  phoneNumber?: string;
  isActive: boolean;
  createdAt: string;
}

/**
 * Tenant context for multi-tenant operations
 * Contains tenant-specific information and user role
 */
export interface TenantContext {
  tenantId: string;
  role: 'OWNER' | 'ADMIN' | 'FINANCE' | 'SALES' | 'WAREHOUSE' | 'STAFF';
  companyName: string;
  status: 'TRIAL' | 'ACTIVE' | 'SUSPENDED' | 'CANCELLED' | 'EXPIRED';
}

/**
 * Generic API success response envelope
 * Backend wraps all successful responses in this format
 * NOTE: Backend uses "success": boolean, not "status": "success"
 */
export interface ApiSuccessResponse<T> {
  success: boolean;
  data: T;
}

/**
 * API error response envelope
 * Backend returns errors in this format
 * NOTE: Backend uses "success": false, not "status": "error"
 */
export interface ApiErrorResponse {
  success: boolean;
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

/**
 * Login request payload
 */
export interface LoginRequest {
  email: string;
  password: string;
}

/**
 * Login response data (unwrapped from envelope)
 */
export interface LoginResponseData {
  user: User;
  accessToken: string;
  // Note: refreshToken is in httpOnly cookie, not in response body
}

/**
 * Logout request payload
 */
export interface LogoutRequest {
  // No body needed - refresh token is in httpOnly cookie
}

/**
 * Switch tenant request payload
 */
export interface SwitchTenantRequest {
  tenantId: string;
}

/**
 * Switch tenant response data
 */
export interface SwitchTenantResponseData {
  accessToken: string;
  activeTenant: TenantContext;
}

/**
 * Get tenants response data
 */
export interface GetTenantsResponseData {
  tenants: TenantContext[];
}

/**
 * Refresh token response data
 */
export interface RefreshTokenResponseData {
  accessToken: string;
  // New refresh token may be returned if rotation is enabled
  refreshToken?: string;
}

/**
 * JWT token payload structure (decoded from accessToken)
 * NOTE: Backend uses "user_id" not standard "sub" claim
 */
export interface JWTPayload {
  exp: number; // Expiration timestamp
  iat: number; // Issued at timestamp
  nbf: number; // Not before timestamp
  user_id: string; // User ID (backend uses user_id, not standard "sub")
  email: string;
  tenant_id?: string; // Tenant ID (optional, may not exist for system admin)
  role?: string; // User role in tenant
}

/**
 * Authentication state interface
 */
export interface AuthState {
  user: User | null;
  accessToken: string | null;
  // Note: refreshToken is NOT stored here (httpOnly cookie)
  activeTenant: TenantContext | null;
  availableTenants: TenantContext[];
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}
