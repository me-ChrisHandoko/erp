"use client";

// Redux Provider and Auth State Restoration
// Wraps the app with Redux Provider and restores auth state from localStorage

import { useEffect, useState } from "react";
import { Provider } from "react-redux";
import { useDispatch, useSelector } from "react-redux";
import { store, RootState } from "@/store";
import { setCredentials } from "@/store/slices/authSlice";
import { useGetCurrentUserQuery } from "@/store/services/authApi";
import { jwtDecode } from "jwt-decode";
import type { JWTPayload } from "@/types/api";

/**
 * Auth initializer component
 * Restores authentication state from localStorage on app load
 */
function AuthInitializer({ children }: { children: React.ReactNode }) {
  const dispatch = useDispatch();
  const [tokenRestored, setTokenRestored] = useState(false);
  const user = useSelector((state: RootState) => state.auth.user);

  // Fetch current user data to get fullName (only if we have partial user data without fullName)
  const shouldFetchFullData = tokenRestored && user && !user.fullName;

  console.log("[Auth] shouldFetchFullData:", {
    tokenRestored,
    hasUser: !!user,
    userFullName: user?.fullName,
    shouldFetch: shouldFetchFullData,
  });

  const { data: currentUser, isLoading, isError, error } = useGetCurrentUserQuery(undefined, {
    skip: !shouldFetchFullData, // Only fetch if we have partial user data (no fullName)
  });

  console.log("[Auth] useGetCurrentUserQuery status:", {
    hasData: !!currentUser,
    isLoading,
    isError,
    error: error ? JSON.stringify(error, null, 2) : null,
  });

  if (isError && error) {
    console.error("[Auth] API Error Details:", error);
  }

  useEffect(() => {
    // Only run on client side
    if (typeof window === "undefined") return;

    // Try to restore auth state from localStorage
    const accessToken = localStorage.getItem("accessToken");

    if (accessToken) {
      try {
        // Decode token to check expiry
        const decoded = jwtDecode<JWTPayload>(accessToken);
        const now = Date.now() / 1000;

        // Check if token is still valid
        if (decoded.exp > now) {
          console.log("[Auth] Restoring session from localStorage");

          // Restore partial user data immediately (to avoid showing "Guest User")
          dispatch(
            setCredentials({
              user: {
                id: decoded.user_id,
                email: decoded.email,
                fullName: "", // Will be populated from API call below
                isActive: true,
                createdAt: "",
              },
              accessToken,
              activeTenant: decoded.tenant_id
                ? {
                    tenantId: decoded.tenant_id,
                    role: decoded.role as any,
                    companyName: "",
                    status: "ACTIVE",
                  }
                : null,
              availableTenants: [],
            })
          );

          setTokenRestored(true); // Trigger API call to get full user data
        } else {
          console.log("[Auth] Token expired, clearing localStorage");
          localStorage.removeItem("accessToken");
        }
      } catch (error) {
        console.error("[Auth] Failed to decode token:", error);
        localStorage.removeItem("accessToken");
      }
    }
  }, [dispatch]);

  useEffect(() => {
    // Update with full user data from API when available
    if (currentUser?.data) {
      const { user, activeTenant } = currentUser.data;
      const accessToken = localStorage.getItem("accessToken");

      console.log("[Auth] API Response received:", {
        userId: user.id,
        email: user.email,
        fullName: user.fullName,
        hasAccessToken: !!accessToken,
      });

      if (accessToken) {
        console.log("[Auth] Updating Redux with full user data:", user.fullName);

        dispatch(
          setCredentials({
            user,
            accessToken,
            activeTenant,
            availableTenants: activeTenant ? [activeTenant] : [],
          })
        );

        console.log("[Auth] Redux state updated successfully");
      } else {
        console.warn("[Auth] No accessToken in localStorage, skipping Redux update");
      }
    } else {
      console.log("[Auth] No currentUser.data available yet");
    }
  }, [currentUser, dispatch]);

  return <>{children}</>;
}

/**
 * Providers component
 * Wraps the app with Redux Provider and auth initialization
 */
export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <Provider store={store}>
      <AuthInitializer>{children}</AuthInitializer>
    </Provider>
  );
}
