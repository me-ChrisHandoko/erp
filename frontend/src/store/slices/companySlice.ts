// Company State Slice
// Manages multi-company state with Redux Toolkit
// PHASE 5: Frontend State Management

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type {
  CompanyState,
  ActiveCompany,
  AvailableCompany,
  CompanyAccess,
} from '@/types/company.types';

/**
 * Initial company state
 * No active company until user selects one
 */
const initialState: CompanyState = {
  activeCompany: null,
  availableCompanies: [],
  loading: false,
  error: null,
  initialized: false,
};

/**
 * Payload for setting active company
 */
export interface SetActiveCompanyPayload {
  company: ActiveCompany;
  persistToLocalStorage?: boolean;
}

/**
 * Company slice with reducers for multi-company state management
 */
const companySlice = createSlice({
  name: 'company',
  initialState,
  reducers: {
    /**
     * Set active company and its access permissions
     * Called after successful company switch
     */
    setActiveCompany: (
      state,
      action: PayloadAction<SetActiveCompanyPayload>
    ) => {
      state.activeCompany = action.payload.company;
      state.loading = false;
      state.error = null;

      // Save to localStorage for persistence across page reloads
      if (
        action.payload.persistToLocalStorage !== false &&
        typeof window !== 'undefined'
      ) {
        localStorage.setItem(
          'activeCompanyId',
          action.payload.company.id
        );
      }
    },

    /**
     * Set available companies for user
     * Called after successful login or when refreshing company list
     */
    setAvailableCompanies: (
      state,
      action: PayloadAction<AvailableCompany[]>
    ) => {
      state.availableCompanies = action.payload;
      state.initialized = true;
      state.loading = false;
      state.error = null;
    },

    /**
     * Update company access permissions
     * Called when permissions change for active company
     */
    updateCompanyAccess: (state, action: PayloadAction<CompanyAccess>) => {
      if (state.activeCompany) {
        state.activeCompany.access = action.payload;
      }
    },

    /**
     * Set loading state
     * Used during async operations (switching, fetching)
     */
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.loading = action.payload;
    },

    /**
     * Set error message
     * Called when company operations fail
     */
    setError: (state, action: PayloadAction<string>) => {
      state.error = action.payload;
      state.loading = false;
    },

    /**
     * Clear error message
     */
    clearError: (state) => {
      state.error = null;
    },

    /**
     * Clear company state
     * Called on logout
     */
    clearCompanyState: (state) => {
      state.activeCompany = null;
      state.availableCompanies = [];
      state.loading = false;
      state.error = null;
      state.initialized = false;

      // Clear localStorage
      if (typeof window !== 'undefined') {
        localStorage.removeItem('activeCompanyId');
      }
    },

    /**
     * Mark state as initialized
     * Used to prevent duplicate initialization
     */
    setInitialized: (state, action: PayloadAction<boolean>) => {
      state.initialized = action.payload;
    },
  },
});

// Export actions
export const {
  setActiveCompany,
  setAvailableCompanies,
  updateCompanyAccess,
  setLoading,
  setError,
  clearError,
  clearCompanyState,
  setInitialized,
} = companySlice.actions;

// Export reducer
export default companySlice.reducer;

// Selectors for accessing company state
export const selectActiveCompany = (state: { company: CompanyState }) =>
  state.company.activeCompany;
export const selectActiveCompanyId = (state: { company: CompanyState }) =>
  state.company.activeCompany?.id;
export const selectAvailableCompanies = (state: { company: CompanyState }) =>
  state.company.availableCompanies;
export const selectCompanyLoading = (state: { company: CompanyState }) =>
  state.company.loading;
export const selectCompanyError = (state: { company: CompanyState }) =>
  state.company.error;
export const selectCompanyInitialized = (state: { company: CompanyState }) =>
  state.company.initialized;

// Derived selectors
export const selectActiveCompanyAccess = (state: { company: CompanyState }) =>
  state.company.activeCompany?.access;
export const selectActiveCompanyRole = (state: { company: CompanyState }) =>
  state.company.activeCompany?.role;
export const selectCanUserAddCompany = (state: { company: CompanyState }) =>
  state.company.availableCompanies.some((c) => c.role === 'OWNER');
