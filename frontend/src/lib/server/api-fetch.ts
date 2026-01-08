/**
 * Server-side API Fetch Utility
 *
 * Handles server-side API calls with automatic authentication
 * and company context from cookies.
 *
 * Server Components can access cookies directly, so we call backend directly
 * instead of going through the proxy.
 *
 * Usage:
 * ```typescript
 * const data = await apiFetch<ProductListResponse>({
 *   endpoint: '/products',
 *   params: { page: 1, page_size: 20 }
 * });
 * ```
 */

import { cookies } from 'next/headers';

interface ApiFetchOptions {
  endpoint: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: any;
  params?: Record<string, string | number | boolean | undefined>;
  cache?: RequestCache;
  revalidate?: number;
}

const BACKEND_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function apiFetch<T>({
  endpoint,
  method = 'GET',
  body,
  params,
  cache = 'no-store',
  revalidate,
}: ApiFetchOptions): Promise<T> {
  // Build query string from params
  const queryString = params
    ? '?' + Object.entries(params)
        .filter(([_, value]) => value !== undefined)
        .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
        .join('&')
    : '';

  // Server Components call backend directly with cookies
  const url = `${BACKEND_URL}/api/v1${endpoint}${queryString}`;

  console.log('[Server API Fetch] Direct to backend:', method, url);

  // Read cookies from Next.js cookie store
  const cookieStore = await cookies();

  // Build headers with authentication and company context
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'Origin': 'http://localhost:3000',
  };

  // Add authentication token
  const accessToken = cookieStore.get('access_token')?.value;
  console.log('[Server API Fetch] access_token:', accessToken ? 'EXISTS' : 'MISSING');
  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  // Add company context
  const companyId = cookieStore.get('active_company_id')?.value;
  console.log('[Server API Fetch] active_company_id:', companyId || 'MISSING');
  if (companyId) {
    headers['X-Company-ID'] = companyId;
  }

  // Add CSRF token for state-changing operations
  const csrfToken = cookieStore.get('csrf_token')?.value;
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  // Build cookie header
  const cookieNames = ['access_token', 'refresh_token', 'csrf_token', 'active_company_id'];
  const cookieEntries: string[] = [];
  for (const name of cookieNames) {
    const cookie = cookieStore.get(name);
    if (cookie) {
      cookieEntries.push(`${name}=${cookie.value}`);
    }
  }
  const cookieHeader = cookieEntries.join('; ');
  if (cookieHeader) {
    headers['Cookie'] = cookieHeader;
  }

  console.log('[Server API Fetch] Headers:', {
    hasAuth: !!headers['Authorization'],
    hasCompanyId: !!headers['X-Company-ID'],
    hasCookie: !!headers['Cookie'],
  });

  try {
    const response = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
      cache,
      next: revalidate ? { revalidate } : undefined,
      credentials: 'include',
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('[Server API Error]', response.status, errorText);
      throw new Error(`API error: ${response.status} - ${errorText}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('[Server API Fetch Error]', error);
    throw error;
  }
}
