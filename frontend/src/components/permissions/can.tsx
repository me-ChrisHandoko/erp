/**
 * Can Component
 *
 * Conditional rendering based on user permissions
 * PHASE 6: Frontend UI Components
 *
 * Usage:
 * ```tsx
 * <Can do="edit" on="products">
 *   <EditButton />
 * </Can>
 *
 * <Can do="delete" on="customers" fallback={<div>No access</div>}>
 *   <DeleteButton />
 * </Can>
 * ```
 */

import type { ReactNode } from 'react';
import { usePermissions } from '@/hooks/use-permissions';
import type { Action, Resource } from '@/lib/permissions';

interface CanProps {
  /**
   * Action to check permission for
   */
  do: Action;

  /**
   * Resource to check permission on
   */
  on: Resource;

  /**
   * Children to render if user has permission
   */
  children: ReactNode;

  /**
   * Fallback to render if user doesn't have permission
   */
  fallback?: ReactNode;

  /**
   * Invert the check (render when user CANNOT do action)
   */
  not?: boolean;
}

export function Can({ do: action, on: resource, children, fallback = null, not = false }: CanProps) {
  const { can } = usePermissions();

  const hasPermission = can(action, resource);
  const shouldRender = not ? !hasPermission : hasPermission;

  if (shouldRender) {
    return <>{children}</>;
  }

  return <>{fallback}</>;
}
