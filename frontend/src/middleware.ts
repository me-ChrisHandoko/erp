// Next.js Middleware for Route Protection
// Protects authenticated routes and redirects unauthorized users

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

/**
 * Middleware function that runs before rendering pages
 * Checks authentication status via refresh_token cookie
 */
export function middleware(request: NextRequest) {
  // Get refresh token cookie (httpOnly cookie set by backend)
  const refreshToken = request.cookies.get("refresh_token");
  const hasRefreshToken = !!refreshToken;

  // Determine if current page is auth or protected
  const isLoginPage = request.nextUrl.pathname.startsWith("/login");
  const isLogoutPage = request.nextUrl.pathname.startsWith("/logout");
  const isProtectedPage =
    request.nextUrl.pathname.startsWith("/dashboard") ||
    request.nextUrl.pathname.startsWith("/master") ||
    request.nextUrl.pathname.startsWith("/inventory") ||
    request.nextUrl.pathname.startsWith("/purchase") ||
    request.nextUrl.pathname.startsWith("/sales") ||
    request.nextUrl.pathname.startsWith("/finance") ||
    request.nextUrl.pathname.startsWith("/settings");

  // Redirect logic
  if (isProtectedPage && !hasRefreshToken) {
    // Accessing protected page without authentication
    console.log("[Middleware] Unauthorized access to protected page, redirecting to login");
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // FIX: Don't redirect if there's a reason parameter (session_expired, session_revoked, etc.)
  // This prevents infinite loop when cookie exists but token is invalid/revoked in database
  const hasSessionReason = request.nextUrl.searchParams.has("reason");

  if (isLoginPage && hasRefreshToken && !hasSessionReason) {
    // Accessing login page while already authenticated (and no session invalidation reason)
    console.log("[Middleware] Already authenticated, redirecting to dashboard");
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  // If login page with reason, allow it to proceed and clear state
  if (isLoginPage && hasSessionReason) {
    console.log("[Middleware] Login page with reason param, allowing access");
    // Note: The cookie will be cleared by a proper logout flow or will expire
  }

  // Allow logout page to always proceed (no redirect)
  if (isLogoutPage) {
    console.log("[Middleware] Logout page accessed, allowing to proceed");
    return NextResponse.next();
  }

  // Allow request to proceed
  return NextResponse.next();
}

/**
 * Configure which routes this middleware runs on
 * Only run on routes that need authentication checking
 */
export const config = {
  matcher: [
    // Auth pages
    "/login",
    "/logout",

    // Protected routes
    "/dashboard/:path*",
    "/master/:path*",
    "/inventory/:path*",
    "/purchase/:path*",
    "/sales/:path*",
    "/finance/:path*",
    "/settings/:path*",
  ],
};
