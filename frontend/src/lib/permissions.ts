/**
 * Permission System
 *
 * Comprehensive permission utilities for role-based access control
 * PHASE 6: Frontend UI Components
 *
 * Features:
 * - Role-based permission matrix
 * - Resource-specific permissions
 * - Granular action checks
 * - Helper functions for common patterns
 */

import type { CompanyRole } from '@/types/company.types';

/**
 * Actions that can be performed on resources
 */
export type Action = 'view' | 'create' | 'edit' | 'delete' | 'approve' | 'export' | 'import';

/**
 * Resources in the system
 */
export type Resource =
  // Master Data
  | 'customers'
  | 'suppliers'
  | 'products'
  | 'warehouses'
  // Inventory
  | 'stock'
  | 'stock-transfers'
  | 'stock-opname'
  | 'inventory-adjustments'
  // Procurement
  | 'purchase-orders'
  | 'goods-receipts'
  | 'purchase-invoices'
  | 'supplier-payments'
  // Sales
  | 'sales-orders'
  | 'deliveries'
  | 'sales-invoices'
  | 'customer-payments'
  // Finance
  | 'journal-entries'
  | 'cash-bank'
  | 'expenses'
  | 'financial-reports'
  // Company
  | 'company-settings'
  | 'bank-accounts'
  | 'users'
  | 'roles'
  // Settings
  | 'system-config'
  | 'preferences';

/**
 * Role Permission Matrix
 * Defines what each role can do on each resource
 */
const ROLE_PERMISSIONS: Record<CompanyRole, {
  [key in Resource]?: Action[];
}> = {
  OWNER: {
    // OWNER has full access to everything
    customers: ['view', 'create', 'edit', 'delete', 'export', 'import'],
    suppliers: ['view', 'create', 'edit', 'delete', 'export', 'import'],
    products: ['view', 'create', 'edit', 'delete', 'export', 'import'],
    warehouses: ['view', 'create', 'edit', 'delete'],
    stock: ['view', 'export'],
    'stock-transfers': ['view', 'create', 'edit', 'delete', 'approve'],
    'stock-opname': ['view', 'create', 'edit', 'delete', 'approve'],
    'inventory-adjustments': ['view', 'create', 'edit', 'delete', 'approve'],
    'purchase-orders': ['view', 'create', 'edit', 'delete', 'approve'],
    'goods-receipts': ['view', 'create', 'edit', 'delete'],
    'purchase-invoices': ['view', 'create', 'edit', 'delete', 'approve'],
    'supplier-payments': ['view', 'create', 'edit', 'delete', 'approve'],
    'sales-orders': ['view', 'create', 'edit', 'delete', 'approve'],
    deliveries: ['view', 'create', 'edit', 'delete'],
    'sales-invoices': ['view', 'create', 'edit', 'delete', 'approve'],
    'customer-payments': ['view', 'create', 'edit', 'delete', 'approve'],
    'journal-entries': ['view', 'create', 'edit', 'delete', 'approve'],
    'cash-bank': ['view', 'create', 'edit', 'delete', 'approve'],
    expenses: ['view', 'create', 'edit', 'delete', 'approve'],
    'financial-reports': ['view', 'export'],
    'company-settings': ['view', 'edit'],
    'bank-accounts': ['view', 'create', 'edit', 'delete'],
    users: ['view', 'create', 'edit', 'delete'],
    roles: ['view', 'create', 'edit', 'delete'],
    'system-config': ['view', 'edit'],
    preferences: ['view', 'edit'],
  },
  ADMIN: {
    // ADMIN has full access except system config
    customers: ['view', 'create', 'edit', 'delete', 'export', 'import'],
    suppliers: ['view', 'create', 'edit', 'delete', 'export', 'import'],
    products: ['view', 'create', 'edit', 'delete', 'export', 'import'],
    warehouses: ['view', 'create', 'edit', 'delete'],
    stock: ['view', 'export'],
    'stock-transfers': ['view', 'create', 'edit', 'delete', 'approve'],
    'stock-opname': ['view', 'create', 'edit', 'delete', 'approve'],
    'inventory-adjustments': ['view', 'create', 'edit', 'delete', 'approve'],
    'purchase-orders': ['view', 'create', 'edit', 'delete', 'approve'],
    'goods-receipts': ['view', 'create', 'edit', 'delete'],
    'purchase-invoices': ['view', 'create', 'edit', 'delete', 'approve'],
    'supplier-payments': ['view', 'create', 'edit', 'delete', 'approve'],
    'sales-orders': ['view', 'create', 'edit', 'delete', 'approve'],
    deliveries: ['view', 'create', 'edit', 'delete'],
    'sales-invoices': ['view', 'create', 'edit', 'delete', 'approve'],
    'customer-payments': ['view', 'create', 'edit', 'delete', 'approve'],
    'journal-entries': ['view', 'create', 'edit', 'delete', 'approve'],
    'cash-bank': ['view', 'create', 'edit', 'delete', 'approve'],
    expenses: ['view', 'create', 'edit', 'delete', 'approve'],
    'financial-reports': ['view', 'export'],
    'company-settings': ['view', 'edit'],
    'bank-accounts': ['view', 'create', 'edit', 'delete'],
    users: ['view', 'create', 'edit', 'delete'],
    roles: ['view'],
    'system-config': ['view'], // Cannot edit system config
    preferences: ['view', 'edit'],
  },
  FINANCE: {
    // FINANCE has full access to financial modules
    customers: ['view', 'export'],
    suppliers: ['view', 'export'],
    products: ['view', 'export'],
    warehouses: ['view'],
    stock: ['view', 'export'],
    'stock-transfers': ['view'],
    'stock-opname': ['view'],
    'inventory-adjustments': ['view'],
    'purchase-orders': ['view'],
    'goods-receipts': ['view'],
    'purchase-invoices': ['view', 'create', 'edit', 'approve'],
    'supplier-payments': ['view', 'create', 'edit', 'approve'],
    'sales-orders': ['view'],
    deliveries: ['view'],
    'sales-invoices': ['view', 'create', 'edit', 'approve'],
    'customer-payments': ['view', 'create', 'edit', 'approve'],
    'journal-entries': ['view', 'create', 'edit', 'approve'],
    'cash-bank': ['view', 'create', 'edit', 'approve'],
    expenses: ['view', 'create', 'edit', 'approve'],
    'financial-reports': ['view', 'export'],
    'company-settings': ['view'],
    'bank-accounts': ['view'],
    users: ['view'],
    roles: ['view'],
    'system-config': ['view'],
    preferences: ['view', 'edit'],
  },
  SALES: {
    // SALES has full access to sales modules
    customers: ['view', 'create', 'edit', 'export'],
    suppliers: ['view'],
    products: ['view', 'export'],
    warehouses: ['view'],
    stock: ['view'],
    'stock-transfers': ['view'],
    'stock-opname': ['view'],
    'inventory-adjustments': ['view'],
    'purchase-orders': ['view'],
    'goods-receipts': ['view'],
    'purchase-invoices': ['view'],
    'supplier-payments': ['view'],
    'sales-orders': ['view', 'create', 'edit'],
    deliveries: ['view', 'create', 'edit'],
    'sales-invoices': ['view', 'create', 'edit'],
    'customer-payments': ['view', 'create'],
    'journal-entries': ['view'],
    'cash-bank': ['view'],
    expenses: ['view'],
    'financial-reports': ['view'],
    'company-settings': ['view'],
    'bank-accounts': ['view'],
    users: ['view'],
    roles: ['view'],
    'system-config': ['view'],
    preferences: ['view', 'edit'],
  },
  WAREHOUSE: {
    // WAREHOUSE has full access to inventory modules
    customers: ['view'],
    suppliers: ['view'],
    products: ['view', 'create', 'edit', 'export'],
    warehouses: ['view'],
    stock: ['view', 'export'],
    'stock-transfers': ['view', 'create', 'edit'],
    'stock-opname': ['view', 'create', 'edit'],
    'inventory-adjustments': ['view', 'create', 'edit'],
    'purchase-orders': ['view'],
    'goods-receipts': ['view', 'create', 'edit'],
    'purchase-invoices': ['view'],
    'supplier-payments': ['view'],
    'sales-orders': ['view'],
    deliveries: ['view', 'create', 'edit'],
    'sales-invoices': ['view'],
    'customer-payments': ['view'],
    'journal-entries': ['view'],
    'cash-bank': ['view'],
    expenses: ['view'],
    'financial-reports': ['view'],
    'company-settings': ['view'],
    'bank-accounts': ['view'],
    users: ['view'],
    roles: ['view'],
    'system-config': ['view'],
    preferences: ['view', 'edit'],
  },
  STAFF: {
    // STAFF has view-only access to most modules
    customers: ['view'],
    suppliers: ['view'],
    products: ['view'],
    warehouses: ['view'],
    stock: ['view'],
    'stock-transfers': ['view'],
    'stock-opname': ['view'],
    'inventory-adjustments': ['view'],
    'purchase-orders': ['view'],
    'goods-receipts': ['view'],
    'purchase-invoices': ['view'],
    'supplier-payments': ['view'],
    'sales-orders': ['view'],
    deliveries: ['view'],
    'sales-invoices': ['view'],
    'customer-payments': ['view'],
    'journal-entries': ['view'],
    'cash-bank': ['view'],
    expenses: ['view'],
    'financial-reports': ['view'],
    'company-settings': ['view'],
    'bank-accounts': ['view'],
    users: ['view'],
    roles: ['view'],
    'system-config': ['view'],
    preferences: ['view', 'edit'],
  },
};

/**
 * Check if a role has permission to perform an action on a resource
 */
export function hasPermission(
  role: CompanyRole | null | undefined,
  action: Action,
  resource: Resource
): boolean {
  if (!role) return false;

  const rolePermissions = ROLE_PERMISSIONS[role];
  if (!rolePermissions) return false;

  const resourcePermissions = rolePermissions[resource];
  if (!resourcePermissions) return false;

  return resourcePermissions.includes(action);
}

/**
 * Check if a role can view a resource
 */
export function canView(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'view', resource);
}

/**
 * Check if a role can create a resource
 */
export function canCreate(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'create', resource);
}

/**
 * Check if a role can edit a resource
 */
export function canEdit(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'edit', resource);
}

/**
 * Check if a role can delete a resource
 */
export function canDelete(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'delete', resource);
}

/**
 * Check if a role can approve a resource
 */
export function canApprove(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'approve', resource);
}

/**
 * Check if a role can export a resource
 */
export function canExport(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'export', resource);
}

/**
 * Check if a role can import a resource
 */
export function canImport(role: CompanyRole | null | undefined, resource: Resource): boolean {
  return hasPermission(role, 'import', resource);
}

/**
 * Get all allowed actions for a role on a resource
 */
export function getAllowedActions(
  role: CompanyRole | null | undefined,
  resource: Resource
): Action[] {
  if (!role) return [];

  const rolePermissions = ROLE_PERMISSIONS[role];
  if (!rolePermissions) return [];

  return rolePermissions[resource] || [];
}

/**
 * Check if a role has any permission on a resource
 */
export function hasAnyPermission(
  role: CompanyRole | null | undefined,
  resource: Resource
): boolean {
  const actions = getAllowedActions(role, resource);
  return actions.length > 0;
}
