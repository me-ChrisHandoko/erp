"use client";

// Redux Provider and Auth State Restoration
// Wraps the app with Redux Provider and restores auth state from localStorage or refresh token

import { useEffect, useState, useRef } from "react";
import { Provider } from "react-redux";
import { useDispatch, useSelector } from "react-redux";
import { store, RootState } from "@/store";
import { setCredentials } from "@/store/slices/authSlice";
import { useGetCurrentUserQuery } from "@/store/services/authApi";
import { jwtDecode } from "jwt-decode";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import type { JWTPayload } from "@/types/api";

/**
 * Helper function to get CSRF token from cookie
 * 🔐 HYBRID SOLUTION PART 1: Proactive CSRF check
 */
function getCSRFToken(): string | null {
  if (typeof document === 'undefined') return null;

  const name = 'csrf_token=';
  const decodedCookie = decodeURIComponent(document.cookie);
  const cookieArray = decodedCookie.split(';');

  for (let cookie of cookieArray) {
    cookie = cookie.trim();
    if (cookie.indexOf(name) === 0) {
      return cookie.substring(name.length);
    }
  }
  return null;
}

/**
 * Auth initializer component
 * Restores authentication state from localStorage or refresh token cookie
 */
function AuthInitializer({ children }: { children: React.ReactNode }) {
  const dispatch = useDispatch();
  const [tokenRestored, setTokenRestored] = useState(false);
  const [isRestoring, setIsRestoring] = useState(true);
  const user = useSelector((state: RootState) => state.auth.user);

  // Ref to prevent infinite loop - tracks if redirect is in progress
  const isRedirecting = useRef(false);

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

    // Prevent running if already redirecting
    if (isRedirecting.current) return;

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

              // 🔐 HYBRID SOLUTION PART 1A: Proactive CSRF Check
              // Check if CSRF cookie exists when restoring session
              const csrfToken = getCSRFToken();

              if (!csrfToken) {
                console.warn("[Auth] ⚠️ CSRF token missing but access token valid");
                console.log("[Auth] 🔄 Forcing token refresh to regenerate CSRF token...");

                // CSRF cookie missing - force token refresh to regenerate it
                // This prevents 403 errors on first POST request
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
                    console.log("[Auth] ✅ Token refresh successful, CSRF regenerated");

                    // Extract new access token
                    const newAccessToken = data.data.accessToken;

                    // Decode new token
                    const newDecoded = jwtDecode<JWTPayload>(newAccessToken);

                    // Save to localStorage
                    localStorage.setItem("accessToken", newAccessToken);

                    // Restore session with new token
                    dispatch(
                      setCredentials({
                        user: {
                          id: newDecoded.user_id,
                          email: newDecoded.email,
                          fullName: "",
                          isActive: true,
                          createdAt: "",
                        },
                        accessToken: newAccessToken,
                        activeTenant: newDecoded.tenant_id
                          ? {
                              tenantId: newDecoded.tenant_id,
                              role: newDecoded.role as any,
                              companyName: "",
                              status: "ACTIVE",
                            }
                          : null,
                        availableTenants: [],
                      })
                    );

                    setTokenRestored(true);
                    setIsRestoring(false);
                    return;
                  } else {
                    console.error("[Auth] ❌ Token refresh failed, proceeding with existing token");
                    console.log("[Auth] Note: First POST request may fail with 403 (will be handled by reactive recovery)");
                    // Fall through to restore session with existing token
                    // Reactive handler will catch 403 errors
                  }
                } catch (error) {
                  console.error("[Auth] ❌ Token refresh error:", error);
                  console.log("[Auth] Proceeding with existing token, reactive handler will catch 403");
                  // Fall through to restore session with existing token
                }
              } else {
                console.log("[Auth] ✅ CSRF token present, session restoration OK");
              }

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
        // Skip if we're on login page to avoid unnecessary 401 errors in console
        const isLoginPage = typeof window !== 'undefined' && window.location.pathname === '/login';

        if (isLoginPage) {
          console.log("[Auth] On login page, skipping session restore attempt");
          setIsRestoring(false);
          return;
        }

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
            // Handle specific error cases
            let errorMessage = "session_expired";
            try {
              const errorData = await response.json();
              const errorCode = errorData?.error?.code;
              const errorMsg = errorData?.error?.message || "";

              console.log("[Auth] Refresh token error:", errorCode, errorMsg);

              // Check if token was revoked (user logged in from another session)
              if (errorCode === "AUTHENTICATION_ERROR" && errorMsg.includes("revoked")) {
                errorMessage = "session_revoked";
                console.log("[Auth] Token was revoked - clearing stale auth state");
              } else if (response.status === 401) {
                errorMessage = "session_expired";
              }
            } catch {
              console.log("[Auth] Could not parse error response");
            }

            // Clear stale localStorage token to prevent loops
            localStorage.removeItem("accessToken");

            console.log("[Auth] Cleared stale auth state, redirecting to login...");

            // IMPORTANT: Set flag BEFORE redirect to prevent infinite loop
            // The dispatch(logout) was causing re-renders before redirect completed
            isRedirecting.current = true;

            // Redirect immediately to login page
            // This handles the case where user was never authenticated (initial page load with stale cookie)
            // AuthGuard won't redirect in this case because wasAuthenticated is false
            window.location.href = `/login?reason=${encodeURIComponent(errorMessage)}`;
            return; // Stop further execution
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
        <LoadingSpinner size="lg" text="Memulihkan sesi..." />
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
