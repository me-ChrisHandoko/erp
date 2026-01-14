// Redux Store Configuration
// Configures Redux Toolkit store with auth slice and RTK Query

import { configureStore, Middleware } from '@reduxjs/toolkit';
import { authApi } from './services/authApi';
import { companyApi } from './services/companyApi';
import { tenantApi } from './services/tenantApi';
import { multiCompanyApi } from './services/multiCompanyApi';
import { companyUserApi } from './services/companyUserApi';
import { productApi } from './services/productApi';
import { customerApi } from './services/customerApi';
import { supplierApi } from './services/supplierApi';
import { warehouseApi } from './services/warehouseApi'; // Warehouse management API
import { stockApi } from './services/stockApi'; // Stock and inventory API
import { initialStockApi } from './services/initialStockApi'; // Initial stock setup API
import { transferApi } from './services/transferApi'; // Stock transfer API
import { opnameApi } from './services/opnameApi'; // Stock opname (physical count) API
import { adjustmentApi } from './services/adjustmentApi'; // Inventory adjustment API
import authReducer, { logout } from './slices/authSlice';
import companyReducer, { setActiveCompany } from './slices/companySlice';

/**
 * Middleware to redirect to logout page when session expires
 * This provides better UX by showing a message before redirecting to login
 */
const redirectToLogoutOnSessionExpiry: Middleware = (storeAPI) => (next) => (action) => {
  // Call next first to let the logout action update the state
  const result = next(action);

  // After logout action is processed, check if there's a logout reason
  if (logout.match(action)) {
    const logoutReason = action.payload?.reason;

    // If logout was due to session expiry, redirect to logout page
    if (logoutReason === 'session_expired' && typeof window !== 'undefined') {
      console.log('[Middleware] Session expired, redirecting to /logout page');

      // Use setTimeout to ensure state is updated before navigation
      setTimeout(() => {
        window.location.href = '/logout';
      }, 100);
    }
  }

  return result;
};

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
    storeAPI.dispatch(customerApi.util.resetApiState());
    storeAPI.dispatch(supplierApi.util.resetApiState());
    storeAPI.dispatch(warehouseApi.util.resetApiState());
    storeAPI.dispatch(stockApi.util.resetApiState());
    storeAPI.dispatch(initialStockApi.util.resetApiState());
    storeAPI.dispatch(transferApi.util.resetApiState());
    storeAPI.dispatch(opnameApi.util.resetApiState());
    storeAPI.dispatch(adjustmentApi.util.resetApiState());

    // CRITICAL: Clear company Redux state to prevent cross-user data exposure
    // Import clearCompanyState from companySlice
    const { clearCompanyState } = require('./slices/companySlice');
    storeAPI.dispatch(clearCompanyState());

    // CRITICAL: Clear company-related localStorage to prevent cross-user data exposure
    // This prevents the next user from inheriting the previous user's active company
    if (typeof window !== 'undefined') {
      localStorage.removeItem('activeCompanyId');
    }

    console.log('[Middleware] All API caches cleared (authApi, companyApi, tenantApi, multiCompanyApi, companyUserApi, productApi, customerApi, supplierApi, warehouseApi, stockApi, initialStockApi, transferApi, opnameApi, adjustmentApi)');
    console.log('[Middleware] Company Redux state cleared (activeCompany, availableCompanies)');
    console.log('[Middleware] localStorage.activeCompanyId cleared to prevent cross-user contamination');
  }

  return result;
};

/**
 * Middleware to reset all RTK Query API caches when user switches company
 * This prevents cached data from previous company showing in new company context
 *
 * CRITICAL: When switching companies, all data (products, customers, suppliers, etc.)
 * must be refetched with the new company context (X-Company-ID header)
 */
const resetAllApiStatesOnCompanySwitch: Middleware = (storeAPI) => (next) => (action) => {
  // Get the previous company ID before the action is processed
  const prevState = storeAPI.getState() as RootState;
  const prevCompanyId = prevState.company.activeCompany?.id;

  // Call next to let the setActiveCompany action update the state
  const result = next(action);

  // After setActiveCompany action is processed, check if company actually changed
  if (setActiveCompany.match(action)) {
    const newCompanyId = action.payload.company.id;

    // Only reset caches if company actually changed (not initial load)
    if (prevCompanyId && prevCompanyId !== newCompanyId) {
      console.log('[Middleware] Company switch detected:', {
        from: prevCompanyId,
        to: newCompanyId,
      });
      console.log('[Middleware] Resetting all API caches to fetch company-specific data...');

      // Reset all RTK Query API slices to clear cached data from previous company
      // This forces refetch with new X-Company-ID header
      storeAPI.dispatch(companyApi.util.resetApiState());
      storeAPI.dispatch(companyUserApi.util.resetApiState());
      storeAPI.dispatch(productApi.util.resetApiState());
      storeAPI.dispatch(customerApi.util.resetApiState());
      storeAPI.dispatch(supplierApi.util.resetApiState());
      storeAPI.dispatch(warehouseApi.util.resetApiState());
      storeAPI.dispatch(stockApi.util.resetApiState());
      storeAPI.dispatch(initialStockApi.util.resetApiState());
      storeAPI.dispatch(transferApi.util.resetApiState());
      storeAPI.dispatch(opnameApi.util.resetApiState());
      storeAPI.dispatch(adjustmentApi.util.resetApiState());
      // Note: authApi, tenantApi, multiCompanyApi are NOT reset (user-level, not company-level)

      console.log('[Middleware] All company-scoped API caches cleared (companyApi, companyUserApi, productApi, customerApi, supplierApi, warehouseApi, stockApi, initialStockApi, transferApi, opnameApi, adjustmentApi)');
      console.log('[Middleware] Next API calls will fetch data for company:', newCompanyId);
    } else if (!prevCompanyId) {
      console.log('[Middleware] Initial company selection:', newCompanyId);
      console.log('[Middleware] No cache reset needed (first company selection)');
    } else {
      console.log('[Middleware] Company unchanged:', newCompanyId);
    }
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
    [customerApi.reducerPath]: customerApi.reducer, // Customer management API
    [supplierApi.reducerPath]: supplierApi.reducer, // Supplier management API
    [warehouseApi.reducerPath]: warehouseApi.reducer, // Warehouse management API
    [stockApi.reducerPath]: stockApi.reducer, // Stock and inventory API
    [initialStockApi.reducerPath]: initialStockApi.reducer, // Initial stock setup API
    [transferApi.reducerPath]: transferApi.reducer, // Stock transfer API
    [opnameApi.reducerPath]: opnameApi.reducer, // Stock opname (physical count) API
    [adjustmentApi.reducerPath]: adjustmentApi.reducer, // Inventory adjustment API
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
      customerApi.middleware, // Customer management middleware
      supplierApi.middleware, // Supplier management middleware
      warehouseApi.middleware, // Warehouse management middleware
      stockApi.middleware, // Stock and inventory middleware
      initialStockApi.middleware, // Initial stock setup middleware
      transferApi.middleware, // Stock transfer middleware
      opnameApi.middleware, // Stock opname (physical count) middleware
      adjustmentApi.middleware, // Inventory adjustment middleware
      redirectToLogoutOnSessionExpiry, // CRITICAL: Redirect to /logout page when session expires
      resetAllApiStatesOnLogout, // CRITICAL: Reset all API caches on logout
      resetAllApiStatesOnCompanySwitch // CRITICAL: Reset company-scoped caches on company switch
    ),

  // Enable Redux DevTools in development
  devTools: process.env.NODE_ENV !== 'production',
});

// Export store types for TypeScript
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
