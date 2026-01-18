/**
 * Audit Log Types
 *
 * Matches backend structure from:
 * - /backend/models/system.go (AuditLog model)
 * - /backend/internal/service/audit/audit_service.go
 */

/**
 * Audit log status enumeration
 */
export type AuditStatus = 'SUCCESS' | 'FAILED' | 'PARTIAL';

/**
 * Entity types that can be audited
 */
export type AuditEntityType =
  | 'product'
  | 'product_supplier'
  | 'customer'
  | 'supplier'
  | 'warehouse'
  | 'user'
  | 'role'
  | 'bank_account'
  | 'company'
  | 'purchase_order'
  | 'sales_order'
  | 'inventory'
  | 'adjustment'
  | 'stock_opname';

/**
 * Audit action types
 */
export type AuditAction =
  | 'CREATE'
  | 'UPDATE'
  | 'DELETE'
  | 'ACTIVATE'
  | 'DEACTIVATE'
  | 'LOGIN'
  | 'LOGOUT'
  | 'ASSIGN'
  | 'REVOKE';

/**
 * Main Audit Log interface matching backend AuditLog model
 */
export interface AuditLog {
  id: string;
  tenantId: string | null;
  companyId: string | null;
  userId: string | null;
  requestId: string | null;
  action: AuditAction;
  entityType: AuditEntityType | null;
  entityId: string | null;
  oldValues: string | null; // JSON string
  newValues: string | null; // JSON string
  ipAddress: string | null;
  userAgent: string | null;
  status: AuditStatus;
  notes: string | null;
  createdAt: string; // ISO date string
}

/**
 * Parsed audit log with typed old/new values
 */
export interface ParsedAuditLog<T = any> extends Omit<AuditLog, 'oldValues' | 'newValues'> {
  oldValues: T | null;
  newValues: T | null;
}

/**
 * Audit log filters for querying
 */
export interface AuditLogFilters {
  page?: number;
  pageSize?: number;
  sortBy?: 'createdAt' | 'action' | 'entityType';
  sortOrder?: 'asc' | 'desc';

  // Filter fields
  userId?: string;
  entityType?: AuditEntityType;
  entityId?: string;
  action?: AuditAction;
  status?: AuditStatus;
  startDate?: string; // ISO date string
  endDate?: string;   // ISO date string
  search?: string;
}

/**
 * Paginated audit log response
 */
export interface AuditLogListResponse {
  success: boolean;
  data: AuditLog[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

/**
 * Single audit log detail response
 */
export interface AuditLogDetailResponse {
  success: boolean;
  data: AuditLog;
}

/**
 * Audit context for tracking operations
 * Used when creating audit logs
 */
export interface AuditContext {
  tenantId?: string;
  companyId?: string;
  userId?: string;
  requestId?: string;
  ipAddress?: string;
  userAgent?: string;
}

/**
 * Changed field information for human-readable notes
 */
export interface ChangedField {
  field: string;
  oldValue: any;
  newValue: any;
}

/**
 * Utility type for entity-specific audit logs
 */
export interface EntityAuditLog<T> extends ParsedAuditLog<T> {
  entityType: AuditEntityType;
  entityId: string;
}

/**
 * Product-specific audit log
 */
export interface ProductAuditLog extends EntityAuditLog<{
  code?: string;
  name?: string;
  category?: string;
  baseUnit?: string;
  baseCost?: string;
  basePrice?: string;
  isActive?: boolean;
  [key: string]: any;
}> {
  entityType: 'product';
}

/**
 * Customer-specific audit log
 */
export interface CustomerAuditLog extends EntityAuditLog<{
  code?: string;
  name?: string;
  type?: string;
  phone?: string;
  email?: string;
  isActive?: boolean;
  [key: string]: any;
}> {
  entityType: 'customer';
}

/**
 * Supplier-specific audit log
 */
export interface SupplierAuditLog extends EntityAuditLog<{
  code?: string;
  name?: string;
  phone?: string;
  email?: string;
  isActive?: boolean;
  [key: string]: any;
}> {
  entityType: 'supplier';
}

/**
 * Warehouse-specific audit log
 */
export interface WarehouseAuditLog extends EntityAuditLog<{
  code?: string;
  name?: string;
  type?: string;
  isActive?: boolean;
  [key: string]: any;
}> {
  entityType: 'warehouse';
}
