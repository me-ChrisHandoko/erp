/**
 * usePermissions Hook
 *
 * Custom React hook for checking permissions based on user's role
 * PHASE 6: Frontend UI Components
 *
 * Usage:
 * ```tsx
 * const { can, cannot, canAny } = usePermissions();
 *
 * if (can('edit', 'products')) {
 *   // Show edit button
 * }
 *
 * if (cannot('delete', 'customers')) {
 *   // Hide delete button
 * }
 * ```
 */

import { useCompany } from './use-company';
import {
  hasPermission,
  canView,
  canCreate,
  canEdit,
  canDelete,
  canApprove,
  canExport,
  canImport,
  getAllowedActions,
  hasAnyPermission,
  type Action,
  type Resource,
} from '@/lib/permissions';

export function usePermissions() {
  const { role } = useCompany();

  /**
   * Check if user can perform an action on a resource
   */
  const can = (action: Action, resource: Resource): boolean => {
    return hasPermission(role, action, resource);
  };

  /**
   * Check if user cannot perform an action on a resource
   */
  const cannot = (action: Action, resource: Resource): boolean => {
    return !hasPermission(role, action, resource);
  };

  /**
   * Check if user has any permission on a resource
   */
  const canAny = (resource: Resource): boolean => {
    return hasAnyPermission(role, resource);
  };

  /**
   * Get all allowed actions for current user on a resource
   */
  const getAllowed = (resource: Resource): Action[] => {
    return getAllowedActions(role, resource);
  };

  return {
    // Core permission checks
    can,
    cannot,
    canAny,
    getAllowed,

    // Specific permission checks (convenience methods)
    canView: (resource: Resource) => canView(role, resource),
    canCreate: (resource: Resource) => canCreate(role, resource),
    canEdit: (resource: Resource) => canEdit(role, resource),
    canDelete: (resource: Resource) => canDelete(role, resource),
    canApprove: (resource: Resource) => canApprove(role, resource),
    canExport: (resource: Resource) => canExport(role, resource),
    canImport: (resource: Resource) => canImport(role, resource),

    // Current role
    role,
  };
}
