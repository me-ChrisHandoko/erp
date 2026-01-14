"use client";

import { useEffect, useState, useRef } from "react";
import { useSelector } from "react-redux";
import { selectIsAuthenticated } from "@/store/slices/authSlice";
import { Loader2 } from "lucide-react";

interface AuthGuardProps {
  children: React.ReactNode;
}

/**
 * Client-side authentication guard component
 * Redirects to login page when user is not authenticated
 * This complements the middleware (server-side) protection with client-side protection
 *
 * Key scenarios handled:
 * 1. User logs out → Redux isAuthenticated becomes false → redirect to login
 * 2. Token expires → Redux state cleared → redirect to login
 * 3. Direct page access without auth → middleware handles, but this is backup
 *
 * Uses window.location.href for hard navigation to avoid conflict with middleware
 * (middleware might redirect back to dashboard if refresh_token cookie still exists)
 */
export function AuthGuard({ children }: AuthGuardProps) {
  const isAuthenticated = useSelector(selectIsAuthenticated);
  const [isChecking, setIsChecking] = useState(true);
  const wasAuthenticated = useRef(false);
  const isRedirecting = useRef(false);

  // Track if user was ever authenticated (to detect logout vs initial load)
  useEffect(() => {
    if (isAuthenticated) {
      wasAuthenticated.current = true;
    }
  }, [isAuthenticated]);

  useEffect(() => {
    // Prevent multiple redirects
    if (isRedirecting.current) return;

    // Initial check - wait a tick for Redux store to be hydrated
    const checkAuth = () => {
      if (!isAuthenticated) {
        // Only redirect if user was previously authenticated (logout scenario)
        // or if we've finished the initial hydration check
        if (wasAuthenticated.current) {
          console.log("[AuthGuard] Authentication lost (logout), redirecting to login");
          isRedirecting.current = true;
          // Use hard navigation to bypass middleware redirect loop
          window.location.href = "/login";
        } else {
          // Initial load - let middleware handle it, but mark as done checking
          console.log("[AuthGuard] Initial check: not authenticated");
          setIsChecking(false);
        }
      } else {
        setIsChecking(false);
      }
    };

    // Small delay to allow Redux state to settle after hydration
    const timeoutId = setTimeout(checkAuth, 100);
    return () => clearTimeout(timeoutId);
  }, [isAuthenticated]);

  // Watch for logout - when isAuthenticated changes from true to false
  useEffect(() => {
    if (!isChecking && !isAuthenticated && wasAuthenticated.current && !isRedirecting.current) {
      console.log("[AuthGuard] Authentication lost, redirecting to login");
      isRedirecting.current = true;
      // Use hard navigation to ensure clean redirect
      window.location.href = "/login";
    }
  }, [isAuthenticated, isChecking]);

  // Show loading while checking auth state or redirecting
  if (isChecking || isRedirecting.current) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // Not authenticated and was previously authenticated - show loading while redirecting
  if (!isAuthenticated && wasAuthenticated.current) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // Authenticated - render children
  return <>{children}</>;
}
