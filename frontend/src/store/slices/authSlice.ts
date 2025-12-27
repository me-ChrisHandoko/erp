// Authentication State Slice
// Manages user authentication state with Redux Toolkit

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { AuthState, User, TenantContext } from '@/types/api';

/**
 * Initial authentication state
 * User starts unauthenticated with no tokens
 */
const initialState: AuthState = {
  user: null,
  accessToken: null,
  activeTenant: null,
  availableTenants: [],
  isAuthenticated: false,
  isLoading: false,
  error: null,
};

/**
 * Payload for setting credentials after successful login
 */
export interface SetCredentialsPayload {
  user: User;
  accessToken: string;
  activeTenant: TenantContext | null;
  availableTenants: TenantContext[];
}

/**
 * Auth slice with reducers for authentication state management
 */
const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    /**
     * Set user credentials and authentication state
     * Called after successful login or token refresh
     */
    setCredentials: (state, action: PayloadAction<SetCredentialsPayload>) => {
      state.user = action.payload.user;
      state.accessToken = action.payload.accessToken;
      state.activeTenant = action.payload.activeTenant;
      state.availableTenants = action.payload.availableTenants;
      state.isAuthenticated = true;
      state.isLoading = false;
      state.error = null;

      // Save access token to localStorage for persistence
      if (typeof window !== 'undefined') {
        localStorage.setItem('accessToken', action.payload.accessToken);
      }
    },

    /**
     * Update access token only
     * Called after token refresh or tenant switch
     */
    setAccessToken: (state, action: PayloadAction<string>) => {
      state.accessToken = action.payload;

      // Update localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('accessToken', action.payload);
      }
    },

    /**
     * Update active tenant context
     * Called after tenant switch
     */
    setActiveTenant: (state, action: PayloadAction<TenantContext>) => {
      state.activeTenant = action.payload;
    },

    /**
     * Set loading state
     * Used during async operations
     */
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },

    /**
     * Set error message
     * Called when authentication operations fail
     */
    setError: (state, action: PayloadAction<string>) => {
      state.error = action.payload;
      state.isLoading = false;
    },

    /**
     * Clear error message
     */
    clearError: (state) => {
      state.error = null;
    },

    /**
     * Logout and clear all authentication state
     * Called on user logout or session expiry
     * Note: rememberEmail is intentionally NOT cleared to preserve "Remember Me" functionality
     */
    logout: (state) => {
      state.user = null;
      state.accessToken = null;
      state.activeTenant = null;
      state.availableTenants = [];
      state.isAuthenticated = false;
      state.error = null;
      state.isLoading = false;

      // Clear localStorage (access token and company context)
      // Note: rememberEmail is preserved for "Remember Me" functionality
      // Note: activeCompanyId is also cleared in middleware for defense in depth
      if (typeof window !== 'undefined') {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('activeCompanyId'); // Prevent cross-user company context leak
      }
    },
  },
});

// Export actions
export const {
  setCredentials,
  setAccessToken,
  setActiveTenant,
  setLoading,
  setError,
  clearError,
  logout,
} = authSlice.actions;

// Export reducer
export default authSlice.reducer;

// Selectors for accessing auth state
export const selectCurrentUser = (state: { auth: AuthState }) => state.auth.user;
export const selectAccessToken = (state: { auth: AuthState }) => state.auth.accessToken;
export const selectActiveTenant = (state: { auth: AuthState }) => state.auth.activeTenant;
export const selectAvailableTenants = (state: { auth: AuthState }) => state.auth.availableTenants;
export const selectIsAuthenticated = (state: { auth: AuthState }) => state.auth.isAuthenticated;
export const selectAuthLoading = (state: { auth: AuthState }) => state.auth.isLoading;
export const selectAuthError = (state: { auth: AuthState }) => state.auth.error;
