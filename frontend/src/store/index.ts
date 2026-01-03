// Redux Store Configuration
// Configures Redux Toolkit store with auth slice and RTK Query

import { configureStore, Middleware } from '@reduxjs/toolkit';
import { authApi } from './services/authApi';
import { companyApi } from './services/companyApi';
import { tenantApi } from './services/tenantApi';
import { multiCompanyApi } from './services/multiCompanyApi';
import { companyUserApi } from './services/companyUserApi';
import { productApi } from './services/productApi';
import authReducer, { logout } from './slices/authSlice';
import companyReducer from './slices/companySlice';

/**
 * Middleware to reset all RTK Query API caches when user logs out
 * This prevents cached data from previous user showing to new user
 */
const resetAllApiStatesOnLogout: Middleware = (storeAPI) => (next) => (action) => {
  // Call next first to let the logout action update the state
  const result = next(action);

  // After logout action is processed, reset all API caches
  if (logout.match(action)) {
    console.log('[Middleware] Logout detected, resetting all API caches...');

    // Reset all RTK Query API slices
    storeAPI.dispatch(authApi.util.resetApiState());
    storeAPI.dispatch(companyApi.util.resetApiState());
    storeAPI.dispatch(tenantApi.util.resetApiState());
    storeAPI.dispatch(multiCompanyApi.util.resetApiState());
    storeAPI.dispatch(companyUserApi.util.resetApiState());
    storeAPI.dispatch(productApi.util.resetApiState());

    // CRITICAL: Clear company Redux state to prevent cross-user data exposure
    // Import clearCompanyState from companySlice
    const { clearCompanyState } = require('./slices/companySlice');
    storeAPI.dispatch(clearCompanyState());

    // CRITICAL: Clear company-related localStorage to prevent cross-user data exposure
    // This prevents the next user from inheriting the previous user's active company
    if (typeof window !== 'undefined') {
      localStorage.removeItem('activeCompanyId');
    }

    console.log('[Middleware] All API caches cleared (authApi, companyApi, tenantApi, multiCompanyApi, companyUserApi, productApi)');
    console.log('[Middleware] Company Redux state cleared (activeCompany, availableCompanies)');
    console.log('[Middleware] localStorage.activeCompanyId cleared to prevent cross-user contamination');
  }

  return result;
};

/**
 * Configure Redux store with:
 * - Auth slice for authentication state
 * - Company slice for multi-company state (PHASE 5)
 * - RTK Query APIs for data fetching
 * - Redux DevTools for debugging
 */
export const store = configureStore({
  reducer: {
    // State slices
    auth: authReducer,
    company: companyReducer, // PHASE 5: Multi-company state

    // RTK Query API reducers
    [authApi.reducerPath]: authApi.reducer,
    [companyApi.reducerPath]: companyApi.reducer,
    [tenantApi.reducerPath]: tenantApi.reducer,
    [multiCompanyApi.reducerPath]: multiCompanyApi.reducer, // PHASE 5: Multi-company API
    [companyUserApi.reducerPath]: companyUserApi.reducer, // Company-scoped user API
    [productApi.reducerPath]: productApi.reducer, // Product management API
  },

  // Add RTK Query middleware for caching, invalidation, etc.
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(
      authApi.middleware,
      companyApi.middleware,
      tenantApi.middleware,
      multiCompanyApi.middleware, // PHASE 5: Multi-company middleware
      companyUserApi.middleware, // Company-scoped user middleware
      productApi.middleware, // Product management middleware
      resetAllApiStatesOnLogout // CRITICAL: Reset all API caches on logout
    ),

  // Enable Redux DevTools in development
  devTools: process.env.NODE_ENV !== 'production',
});

// Export store types for TypeScript
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
