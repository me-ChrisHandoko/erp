/**
 * PermissionGuard Component
 *
 * Route/component protection based on role requirements
 * PHASE 6: Frontend UI Components
 *
 * Usage:
 * ```tsx
 * // Role-based guard
 * <PermissionGuard require="ADMIN">
 *   <AdminPanel />
 * </PermissionGuard>
 *
 * // Multi-role guard
 * <PermissionGuard requireAny={['OWNER', 'ADMIN']}>
 *   <ManagementPanel />
 * </PermissionGuard>
 *
 * // Custom fallback
 * <PermissionGuard require="OWNER" fallback={<AccessDenied />}>
 *   <OwnerPanel />
 * </PermissionGuard>
 * ```
 */

import type { ReactNode } from 'react';
import { useCompany } from '@/hooks/use-company';
import type { CompanyRole } from '@/types/company.types';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { ShieldAlert } from 'lucide-react';

interface PermissionGuardProps {
  /**
   * Specific role required (exact match)
   */
  require?: CompanyRole;

  /**
   * Any of these roles required (OR condition)
   */
  requireAny?: CompanyRole[];

  /**
   * All of these roles required (AND condition) - rarely used
   */
  requireAll?: CompanyRole[];

  /**
   * Minimum role level required (based on hierarchy)
   */
  minRole?: CompanyRole;

  /**
   * Children to render if permission is granted
   */
  children: ReactNode;

  /**
   * Custom fallback to render if permission is denied
   */
  fallback?: ReactNode;

  /**
   * Redirect URL if permission is denied (alternative to fallback)
   */
  redirectTo?: string;
}

/**
 * Default access denied component
 */
function DefaultAccessDenied() {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <Alert variant="destructive" className="max-w-md">
        <ShieldAlert className="h-4 w-4" />
        <AlertTitle>Akses Ditolak</AlertTitle>
        <AlertDescription>
          Anda tidak memiliki izin untuk mengakses halaman ini.
          Silakan hubungi administrator jika Anda merasa ini adalah kesalahan.
        </AlertDescription>
      </Alert>
    </div>
  );
}

export function PermissionGuard({
  require,
  requireAny,
  requireAll,
  minRole,
  children,
  fallback,
  redirectTo,
}: PermissionGuardProps) {
  const { role, hasMinRole } = useCompany();

  // Check permissions based on provided props
  let hasPermission = false;

  if (require) {
    // Exact role match
    hasPermission = role === require;
  } else if (requireAny) {
    // Any of the roles match (OR)
    hasPermission = requireAny.includes(role as CompanyRole);
  } else if (requireAll) {
    // All roles match (AND) - typically not useful, but supported
    hasPermission = requireAll.every((r) => role === r);
  } else if (minRole) {
    // Minimum role level (hierarchical)
    hasPermission = hasMinRole(minRole);
  } else {
    // No permission check specified - allow by default
    hasPermission = true;
  }

  // Handle permission denied
  if (!hasPermission) {
    // Redirect if URL provided
    if (redirectTo) {
      if (typeof window !== 'undefined') {
        window.location.assign(redirectTo);
      }
      return null;
    }

    // Custom fallback
    if (fallback) {
      return <>{fallback}</>;
    }

    // Default access denied message
    return <DefaultAccessDenied />;
  }

  // Permission granted - render children
  return <>{children}</>;
}
