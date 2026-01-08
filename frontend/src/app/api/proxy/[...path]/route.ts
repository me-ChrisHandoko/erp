// Universal API Proxy for Backend Requests
// Handles both client-side and server-side requests
// Ensures cookies are properly forwarded to backend

import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

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

  console.log('[API Proxy GET] Starting request to:', url);
  console.log('[API Proxy GET] Query params:', searchParams);

  try {
    console.log('[API Proxy GET] Step 1: Building headers');

    // Build headers
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': 'http://localhost:3000',
    };

    console.log('[API Proxy GET] Step 2: Reading cookies from store');

    // Add authentication token
    const accessToken = cookieStore.get('access_token')?.value;
    console.log('[API Proxy GET] access_token:', accessToken ? `EXISTS (${accessToken.substring(0, 20)}...)` : 'MISSING');

    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
      console.log('[API Proxy GET] Added Authorization header');
    } else {
      console.log('[API Proxy GET] WARNING: No access_token cookie found!');
    }

    // Add company context
    const companyId = cookieStore.get('active_company_id')?.value;
    console.log('[API Proxy GET] active_company_id:', companyId ? companyId : 'MISSING');

    if (companyId) {
      headers['X-Company-ID'] = companyId;
      console.log('[API Proxy GET] Added X-Company-ID header');
    } else {
      console.log('[API Proxy GET] WARNING: No active_company_id cookie found!');
    }

    // Add CSRF token for state-changing operations
    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
      console.log('[API Proxy GET] Added CSRF token');
    }

    console.log('[API Proxy GET] Step 3: Building cookie header');

    // Forward cookies to backend
    const cookieHeader = await buildCookieHeader(cookieStore);
    console.log('[API Proxy GET] Cookie header length:', cookieHeader.length);
    console.log('[API Proxy GET] Cookie header preview:', cookieHeader.substring(0, 100));

    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
      console.log('[API Proxy GET] Added Cookie header');
    } else {
      console.log('[API Proxy GET] WARNING: Cookie header is empty!');
    }

    console.log('[API Proxy GET] Step 4: Final headers check:', {
      hasAuth: !!headers['Authorization'],
      hasCompanyId: !!headers['X-Company-ID'],
      hasCookie: !!headers['Cookie'],
    });

    const response = await fetch(url, {
      method: 'GET',
      headers,
      credentials: 'include',
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
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
  const body = await request.json();

  console.log('[API Proxy POST] Starting request to:', url);

  try {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': 'http://localhost:3000',
    };

    const accessToken = cookieStore.get('access_token')?.value;
    console.log('[API Proxy POST] access_token:', accessToken ? 'EXISTS' : 'MISSING');
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    const companyId = cookieStore.get('active_company_id')?.value;
    console.log('[API Proxy POST] active_company_id:', companyId || 'MISSING');
    if (companyId) {
      headers['X-Company-ID'] = companyId;
    }

    const csrfToken = cookieStore.get('csrf_token')?.value;
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }

    const cookieHeader = await buildCookieHeader(cookieStore); // ✅ FIXED: Added await
    console.log('[API Proxy POST] Cookie header length:', cookieHeader.length);
    if (cookieHeader) {
      headers['Cookie'] = cookieHeader;
    }

    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
      credentials: 'include',
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
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
  const body = await request.json();

  try {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': 'http://localhost:3000',
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    const companyId = cookieStore.get('active_company_id')?.value;
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
      method: 'PUT',
      headers,
      body: JSON.stringify(body),
      credentials: 'include',
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
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
  const body = await request.json();

  try {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Origin': 'http://localhost:3000',
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    const companyId = cookieStore.get('active_company_id')?.value;
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
      method: 'PATCH',
      headers,
      body: JSON.stringify(body),
      credentials: 'include',
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
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
      'Origin': 'http://localhost:3000',
    };

    const accessToken = cookieStore.get('access_token')?.value;
    if (accessToken) {
      headers['Authorization'] = `Bearer ${accessToken}`;
    }

    const companyId = cookieStore.get('active_company_id')?.value;
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

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json(data);
  } catch (error) {
    console.error('[API Proxy] DELETE Error:', error);
    return NextResponse.json(
      { success: false, error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}
