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
  const isAuthPage = request.nextUrl.pathname.startsWith("/login");
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

  if (isAuthPage && hasRefreshToken) {
    // Accessing login page while already authenticated
    console.log("[Middleware] Already authenticated, redirecting to dashboard");
    return NextResponse.redirect(new URL("/dashboard", request.url));
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
