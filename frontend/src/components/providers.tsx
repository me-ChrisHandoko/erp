"use client";

// Redux Provider and Auth State Restoration
// Wraps the app with Redux Provider and restores auth state from localStorage or refresh token

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
 * Restores authentication state from localStorage or refresh token cookie
 */
function AuthInitializer({ children }: { children: React.ReactNode }) {
  const dispatch = useDispatch();
  const [tokenRestored, setTokenRestored] = useState(false);
  const [isRestoring, setIsRestoring] = useState(true);
  const user = useSelector((state: RootState) => state.auth.user);

  // SOLUTION 1: Monitor access token from Redux to detect refresh flow
  const accessTokenFromRedux = useSelector((state: RootState) => state.auth.accessToken);

  // Fetch current user data to get fullName (only if we have partial user data without fullName)
  const shouldFetchFullData = tokenRestored && user && !user.fullName;

  console.log("[Auth] shouldFetchFullData:", {
    tokenRestored,
    hasUser: !!user,
    userFullName: user?.fullName,
    shouldFetch: shouldFetchFullData,
    hasReduxToken: !!accessTokenFromRedux,
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

  // SOLUTION 1: Detect when token exists in Redux but user doesn't (after refresh)
  useEffect(() => {
    if (accessTokenFromRedux && !user && !tokenRestored) {
      console.log("[Auth] Token detected in Redux after refresh, triggering user fetch");

      try {
        // Decode token to get user info
        const decoded = jwtDecode<JWTPayload>(accessTokenFromRedux);
        const now = Date.now() / 1000;

        // Check if token is still valid
        if (decoded.exp > now) {
          console.log("[Auth] Setting partial user data from refreshed token");

          // Set partial user data from token
          dispatch(
            setCredentials({
              user: {
                id: decoded.user_id,
                email: decoded.email,
                fullName: "", // Will be populated from API call
                isActive: true,
                createdAt: "",
              },
              accessToken: accessTokenFromRedux,
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

          setTokenRestored(true); // Trigger getCurrentUser query
        } else {
          console.log("[Auth] Refreshed token already expired");
        }
      } catch (error) {
        console.error("[Auth] Failed to decode refreshed token:", error);
      }
    }
  }, [accessTokenFromRedux, user, tokenRestored, dispatch]);

  // Main session restoration logic
  useEffect(() => {
    // Only run on client side
    if (typeof window === "undefined") return;

    const restoreSession = async () => {
      try {
        // Try to restore auth state from localStorage first
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
              setIsRestoring(false);
              return; // Successfully restored from localStorage
            } else {
              console.log("[Auth] Token expired, clearing localStorage");
              localStorage.removeItem("accessToken");
            }
          } catch (error) {
            console.error("[Auth] Failed to decode token:", error);
            localStorage.removeItem("accessToken");
          }
        }

        // If we reach here, localStorage restoration failed
        // Try to restore from refresh_token cookie by calling /auth/refresh
        console.log("[Auth] No valid localStorage token, attempting session restore from refresh token cookie...");

        try {
          const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/refresh`, {
            method: 'POST',
            credentials: 'include', // Send cookies (refresh_token)
            headers: {
              'Content-Type': 'application/json',
            },
          });

          if (response.ok) {
            const data = await response.json();
            console.log("[Auth] Session restored from refresh token cookie");

            // Extract new access token
            const newAccessToken = data.data.accessToken;

            // Decode token to get user info
            const decoded = jwtDecode<JWTPayload>(newAccessToken);

            // Save to localStorage
            localStorage.setItem("accessToken", newAccessToken);

            // Restore partial user data
            dispatch(
              setCredentials({
                user: {
                  id: decoded.user_id,
                  email: decoded.email,
                  fullName: "", // Will be populated from API call
                  isActive: true,
                  createdAt: "",
                },
                accessToken: newAccessToken,
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

            setTokenRestored(true); // Trigger getCurrentUser query
          } else {
            console.log("[Auth] No valid refresh token cookie found");
          }
        } catch (error) {
          console.log("[Auth] Failed to restore from refresh token:", error);
        }

        setIsRestoring(false);
      } catch (error) {
        console.error("[Auth] Session restoration error:", error);
        setIsRestoring(false);
      }
    };

    restoreSession();
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

  // Show loading state while restoring session
  if (isRestoring) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="flex flex-col items-center gap-2">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          <p className="text-sm text-muted-foreground">Memulihkan sesi...</p>
        </div>
      </div>
    );
  }

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
