/**
 * Server-side Authentication Utilities
 *
 * Handles authentication checks and session management
 * for Next.js Server Components.
 */

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

export interface ServerSession {
  accessToken: string | undefined;
  activeCompanyId: string | undefined;
  isAuthenticated: boolean;
}

/**
 * Get current server-side session from cookies
 */
export async function getServerSession(): Promise<ServerSession> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get('access_token')?.value;
  const activeCompanyId = cookieStore.get('active_company_id')?.value;

  return {
    accessToken,
    activeCompanyId,
    isAuthenticated: !!accessToken,
  };
}

/**
 * Require authentication - redirect to login if not authenticated
 */
export async function requireAuth(): Promise<ServerSession> {
  const session = await getServerSession();

  if (!session.isAuthenticated) {
    console.log('[Server Auth] Not authenticated, redirecting to login');
    redirect('/login');
  }

  return session;
}

/**
 * Require company context - redirect if no active company
 */
export async function requireCompany(): Promise<ServerSession> {
  const session = await requireAuth();

  if (!session.activeCompanyId) {
    console.log('[Server Auth] No active company, user needs to select company');
    // For now, just return session - company initializer will handle it
    // In future, could redirect to company selection page
  }

  return session;
}

/**
 * Check if user is authenticated (non-throwing version)
 */
export async function isAuthenticated(): Promise<boolean> {
  const session = await getServerSession();
  return session.isAuthenticated;
}
