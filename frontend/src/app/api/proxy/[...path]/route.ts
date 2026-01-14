// Universal API Proxy for Backend Requests
// Handles both client-side and server-side requests
// Ensures cookies are properly forwarded to backend

import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const FRONTEND_URL = process.env.NEXT_PUBLIC_FRONTEND_URL || 'http://localhost:3000';

// Helper to construct cookie header string
async function buildCookieHeader(cookieStore: Awaited<ReturnType<typeof cookies>>): Promise<string> {
  const cookieEntries: string[] = [];

  // Get specific cookies we need to forward
  const cookieNames = ['access_token', 'refresh_token', 'csrf_token', 'active_company_id'];

  for (const name of cookieNames) {
    const cookie = cookieStore.get(name);
    if (cookie) {
      cookieEntries.push(`${name}=${cookie.value}`);
    }
  }

  return cookieEntries.join('; ');
}

// GET handler
export async function GET(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const cookieStore = await cookies();
  const { path: pathArray } = await context.params; // ✅ Await params (Next.js 16 requirement)
  const path = pathArray.join('/');
  const searchParams = request.nextUrl.searchParams.toString();
  const url = `${BACKEND_URL}/api/v1/${path}${searchParams ? `?${searchParams}` : ''}`;

  try {
    // Build headers
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': FRONTEND_URL,
    };

    // Add authentication token
    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    // Add company context
    // Priority: Use X-Company-ID from client request header (set from Redux state)
    // Fallback: Read from cookie if header not provided
    const clientCompanyId = request.headers.get('X-Company-ID');
    const cookieCompanyId = cookieStore.get('active_company_id')?.value;
    const companyId = clientCompanyId || cookieCompanyId;

    if (companyId) {
      headers['X-Company-ID'] = companyId;
    }

    // Add CSRF token for state-changing operations
    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    // Forward cookies to backend
    const cookieHeader = await buildCookieHeader(cookieStore);
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
    }

    const response = await fetch(url, {
      method: 'GET',
      headers,
      credentials: 'include',
    });

    const data = await response.json();

    // Create NextResponse with data
    const nextResponse = NextResponse.json(data, { status: response.status });

    // Forward Set-Cookie headers from backend
    const setCookieHeaders = response.headers.getSetCookie();
    if (setCookieHeaders && setCookieHeaders.length > 0) {
      setCookieHeaders.forEach((cookie) => {
        nextResponse.headers.append('Set-Cookie', cookie);
      });
    }

    return nextResponse;
  } catch (error) {
    console.error('[API Proxy] GET Error:', error);
    return NextResponse.json(
      { success: false, error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}

// POST handler
export async function POST(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const cookieStore = await cookies();
  const { path: pathArray } = await context.params;
  const path = pathArray.join('/');
  const url = `${BACKEND_URL}/api/v1/${path}`;

  try {
    // Parse body - handle empty body for endpoints like logout
    let body = null;
    const contentType = request.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      try {
        body = await request.json();
      } catch {
        // Empty body is OK for some endpoints
        body = {};
      }
    }

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': FRONTEND_URL,
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    // Priority: Use X-Company-ID from client request header (Redux state)
    const clientCompanyId = request.headers.get('X-Company-ID');
    const cookieCompanyId = cookieStore.get('active_company_id')?.value;
    const companyId = clientCompanyId || cookieCompanyId;
    if (companyId) {
      headers['X-Company-ID'] = companyId;
    }

    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const cookieHeader = await buildCookieHeader(cookieStore);
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
    }

    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: body ? JSON.stringify(body) : undefined,
      credentials: 'include',
    });

    const data = await response.json();

    // Create NextResponse with data
    const nextResponse = NextResponse.json(data, { status: response.status });

    // Forward Set-Cookie headers from backend (important for logout, token refresh, etc.)
    const setCookieHeaders = response.headers.getSetCookie();
    if (setCookieHeaders && setCookieHeaders.length > 0) {
      setCookieHeaders.forEach((cookie) => {
        nextResponse.headers.append('Set-Cookie', cookie);
      });
    }

    return nextResponse;
  } catch (error) {
    console.error('[API Proxy] POST Error:', error);
    return NextResponse.json(
      { success: false, error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}

// PUT, PATCH, DELETE handlers (same pattern)
export async function PUT(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const cookieStore = await cookies();
  const { path: pathArray } = await context.params;
  const path = pathArray.join('/');
  const url = `${BACKEND_URL}/api/v1/${path}`;

  try {
    // Parse body - handle empty body gracefully
    let body = null;
    const contentType = request.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      try {
        body = await request.json();
      } catch {
        body = {};
      }
    }

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': FRONTEND_URL,
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    // Priority: Use X-Company-ID from client request header (Redux state)
    const clientCompanyId = request.headers.get('X-Company-ID');
    const cookieCompanyId = cookieStore.get('active_company_id')?.value;
    const companyId = clientCompanyId || cookieCompanyId;
    if (companyId) {
      headers['X-Company-ID'] = companyId;
    }

    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const cookieHeader = await buildCookieHeader(cookieStore);
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
    }

    const response = await fetch(url, {
      method: 'PUT',
      headers,
      body: body ? JSON.stringify(body) : undefined,
      credentials: 'include',
    });

    const data = await response.json();

    // Create NextResponse with data
    const nextResponse = NextResponse.json(data, { status: response.status });

    // Forward Set-Cookie headers from backend
    const setCookieHeaders = response.headers.getSetCookie();
    if (setCookieHeaders && setCookieHeaders.length > 0) {
      setCookieHeaders.forEach((cookie) => {
        nextResponse.headers.append('Set-Cookie', cookie);
      });
    }

    return nextResponse;
  } catch (error) {
    console.error('[API Proxy] PUT Error:', error);
    return NextResponse.json(
      { success: false, error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}

export async function PATCH(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const cookieStore = await cookies();
  const { path: pathArray } = await context.params;
  const path = pathArray.join('/');
  const url = `${BACKEND_URL}/api/v1/${path}`;

  try {
    // Parse body - handle empty body gracefully
    let body = null;
    const contentType = request.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      try {
        body = await request.json();
      } catch {
        body = {};
      }
    }

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': FRONTEND_URL,
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    // Priority: Use X-Company-ID from client request header (Redux state)
    const clientCompanyId = request.headers.get('X-Company-ID');
    const cookieCompanyId = cookieStore.get('active_company_id')?.value;
    const companyId = clientCompanyId || cookieCompanyId;
    if (companyId) {
      headers['X-Company-ID'] = companyId;
    }

    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const cookieHeader = await buildCookieHeader(cookieStore);
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
    }

    const response = await fetch(url, {
      method: 'PATCH',
      headers,
      body: body ? JSON.stringify(body) : undefined,
      credentials: 'include',
    });

    const data = await response.json();

    // Create NextResponse with data
    const nextResponse = NextResponse.json(data, { status: response.status });

    // Forward Set-Cookie headers from backend
    const setCookieHeaders = response.headers.getSetCookie();
    if (setCookieHeaders && setCookieHeaders.length > 0) {
      setCookieHeaders.forEach((cookie) => {
        nextResponse.headers.append('Set-Cookie', cookie);
      });
    }

    return nextResponse;
  } catch (error) {
    console.error('[API Proxy] PATCH Error:', error);
    return NextResponse.json(
      { success: false, error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}

export async function DELETE(
  request: NextRequest,
  context: { params: Promise<{ path: string[] }> }
) {
  const cookieStore = await cookies();
  const { path: pathArray } = await context.params;
  const path = pathArray.join('/');
  const url = `${BACKEND_URL}/api/v1/${path}`;

  try {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': FRONTEND_URL,
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    // Priority: Use X-Company-ID from client request header (Redux state)
    const clientCompanyId = request.headers.get('X-Company-ID');
    const cookieCompanyId = cookieStore.get('active_company_id')?.value;
    const companyId = clientCompanyId || cookieCompanyId;
    if (companyId) {
      headers['X-Company-ID'] = companyId;
    }

    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const cookieHeader = await buildCookieHeader(cookieStore); // ✅ FIXED: Added await
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
    }

    const response = await fetch(url, {
      method: 'DELETE',
      headers,
      credentials: 'include',
    });

    const data = await response.json();

    // Create NextResponse with data
    const nextResponse = NextResponse.json(data, { status: response.status });

    // Forward Set-Cookie headers from backend
    const setCookieHeaders = response.headers.getSetCookie();
    if (setCookieHeaders && setCookieHeaders.length > 0) {
      setCookieHeaders.forEach((cookie) => {
        nextResponse.headers.append('Set-Cookie', cookie);
      });
    }

    return nextResponse;
  } catch (error) {
    console.error('[API Proxy] DELETE Error:', error);
    return NextResponse.json(
      { success: false, error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}
