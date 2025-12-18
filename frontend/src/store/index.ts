// Redux Store Configuration
// Configures Redux Toolkit store with auth slice and RTK Query

import { configureStore } from '@reduxjs/toolkit';
import { authApi } from './services/authApi';
import authReducer from './slices/authSlice';

/**
 * Configure Redux store with:
 * - Auth slice for state management
 * - RTK Query API for data fetching
 * - Redux DevTools for debugging
 */
export const store = configureStore({
  reducer: {
    // Auth state slice
    auth: authReducer,

    // RTK Query API reducer
    [authApi.reducerPath]: authApi.reducer,
  },

  // Add RTK Query middleware for caching, invalidation, etc.
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(authApi.middleware),

  // Enable Redux DevTools in development
  devTools: process.env.NODE_ENV !== 'production',
});

// Export store types for TypeScript
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
