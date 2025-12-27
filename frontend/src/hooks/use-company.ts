/**
 * useCompany Hook
 *
 * Custom React hook for accessing multi-company state and operations
 * PHASE 5: Frontend State Management
 *
 * Usage:
 * ```tsx
 * const { activeCompany, availableCompanies, switchCompany, hasPermission } = useCompany();
 *
 * // Check permission
 * if (hasPermission('canEdit')) {
 *   // Show edit button
 * }
 *
 * // Switch company
 * await switchCompany(companyId);
 * ```
 */

import { useSelector } from 'react-redux';
import { useSwitchCompanyMutation } from '@/store/services/multiCompanyApi';
import {
  selectActiveCompany,
  selectActiveCompanyId,
  selectAvailableCompanies,
  selectCompanyLoading,
  selectCompanyError,
  selectActiveCompanyAccess,
  selectActiveCompanyRole,
  selectCanUserAddCompany,
} from '@/store/slices/companySlice';
import type { CompanyRole } from '@/types/company.types';

/**
 * Permission types that can be checked
 */
type Permission = 'canView' | 'canEdit' | 'canDelete' | 'canApprove';

/**
 * Role hierarchy for permission checks
 */
const ROLE_HIERARCHY: Record<CompanyRole, number> = {
  OWNER: 5,
  ADMIN: 4,
  FINANCE: 3,
  SALES: 2,
  WAREHOUSE: 2,
  STAFF: 1,
};

export function useCompany() {
  // Select company state from Redux
  const activeCompany = useSelector(selectActiveCompany);
  const activeCompanyId = useSelector(selectActiveCompanyId);
  const availableCompanies = useSelector(selectAvailableCompanies);
  const loading = useSelector(selectCompanyLoading);
  const error = useSelector(selectCompanyError);
  const access = useSelector(selectActiveCompanyAccess);
  const role = useSelector(selectActiveCompanyRole);
  const canAddCompany = useSelector(selectCanUserAddCompany);

  // Get switch company mutation
  const [switchCompanyMutation, { isLoading: isSwitching }] =
    useSwitchCompanyMutation();

  /**
   * Check if user has specific permission
   */
  const hasPermission = (permission: Permission): boolean => {
    if (!access) return false;
    return access[permission] === true;
  };

  /**
   * Check if user has minimum role level
   * Example: hasMinRole('ADMIN') returns true for OWNER and ADMIN
   */
  const hasMinRole = (minRole: CompanyRole): boolean => {
    if (!role) return false;
    return ROLE_HIERARCHY[role] >= ROLE_HIERARCHY[minRole];
  };

  /**
   * Check if user is owner of active company
   */
  const isOwner = (): boolean => {
    return role === 'OWNER';
  };

  /**
   * Check if user is admin or owner of active company
   */
  const isAdminOrOwner = (): boolean => {
    return role === 'OWNER' || role === 'ADMIN';
  };

  /**
   * Switch to different company
   */
  const switchCompany = async (companyId: string): Promise<boolean> => {
    try {
      await switchCompanyMutation(companyId).unwrap();
      return true;
    } catch (error) {
      console.error('Failed to switch company:', error);
      return false;
    }
  };

  /**
   * Get company by ID from available companies
   */
  const getCompanyById = (companyId: string) => {
    return availableCompanies.find((c) => c.id === companyId);
  };

  /**
   * Check if company is active
   */
  const isCompanyActive = (companyId: string): boolean => {
    const company = getCompanyById(companyId);
    return company?.isActive === true;
  };

  return {
    // State
    activeCompany,
    activeCompanyId,
    availableCompanies,
    loading: loading || isSwitching,
    error,
    access,
    role,
    canAddCompany,

    // Permission checks
    hasPermission,
    hasMinRole,
    isOwner,
    isAdminOrOwner,

    // Actions
    switchCompany,
    getCompanyById,
    isCompanyActive,

    // Computed
    hasMultipleCompanies: availableCompanies.length > 1,
    hasNoCompanies: availableCompanies.length === 0,
    isSwitching,
  };
}
