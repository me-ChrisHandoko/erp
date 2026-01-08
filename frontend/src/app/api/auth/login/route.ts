// Next.js API Route for Login
// Acts as a proxy to backend API and sets cookies in the same domain
// This enables Server Components to read httpOnly cookies

import { NextRequest, NextResponse } from 'next/server';

const BACKEND_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    // Forward login request to backend
    const response = await fetch(`${BACKEND_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Origin': 'http://localhost:3000',
      },
      body: JSON.stringify(body),
    });

    const data = await response.json();

    if (!response.ok) {
      return NextResponse.json(data, { status: response.status });
    }

    // Extract Set-Cookie headers from backend response
    const setCookieHeaders = response.headers.getSetCookie();
    console.log('[Login API] Backend returned', setCookieHeaders.length, 'Set-Cookie headers');
    setCookieHeaders.forEach((header, index) => {
      const cookieName = header.split('=')[0];
      console.log(`[Login API] Cookie ${index + 1}: ${cookieName}`);
    });

    // Create Next.js response with same JSON body
    const nextResponse = NextResponse.json(data);

    // Copy all Set-Cookie headers to Next.js response
    // This ensures cookies are set in browser at localhost:3000 domain
    for (const cookieHeader of setCookieHeaders) {
      nextResponse.headers.append('Set-Cookie', cookieHeader);
    }

    console.log('[Login API] Successfully forwarded cookies to client');
    return nextResponse;
  } catch (error) {
    console.error('[Login API Route] Error:', error);
    return NextResponse.json(
      { success: false, error: 'Internal server error' },
      { status: 500 }
    );
  }
}
