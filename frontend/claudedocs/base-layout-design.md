# Base Layout Design Specification

## ERP Distribution Frontend - Sidebar-07 Integration

**Version:** 1.0
**Date:** 2025-12-17
**Author:** Claude Code Design System
**Status:** Ready for Implementation

---

## 1. Executive Summary

This document specifies the base layout architecture for the ERP Distribution (Distribusi Sembako) frontend application using shadcn/ui sidebar-07 component. The design implements a multi-tenant aware, responsive layout optimized for Indonesian food distribution business workflows.

### Design Goals

- ✅ Clean separation between authenticated and public routes
- ✅ Reusable sidebar layout for all ERP modules
- ✅ Multi-tenant organization switching
- ✅ Mobile-responsive design (Sheet overlay on mobile, collapsible sidebar on desktop)
- ✅ Type-safe TypeScript implementation
- ✅ Performance-optimized with Next.js 16 patterns

---

## 2. Architecture Overview

### 2.1 Layout Hierarchy

```
app/
├── layout.tsx                    # Root Layout (Server Component)
│   └── HTML + Fonts + Global Styles
│
├── (auth)/                       # Route Group - Public Routes
│   ├── layout.tsx               # Auth Layout (Centered, no sidebar)
│   ├── login/page.tsx
│   └── register/page.tsx
│
└── (app)/                        # Route Group - Authenticated Routes
    ├── layout.tsx               # App Layout (Client Component with SidebarProvider)
    │   ├── SidebarProvider
    │   │   ├── AppSidebar
    │   │   └── SidebarInset
    │   │       ├── Header (Breadcrumbs + Actions)
    │   │       └── Main Content Area
    │
    ├── dashboard/page.tsx
    ├── master/
    │   ├── customers/page.tsx
    │   ├── suppliers/page.tsx
    │   ├── products/page.tsx
    │   └── warehouses/page.tsx
    ├── inventory/
    │   ├── stock/page.tsx
    │   ├── transfers/page.tsx
    │   └── opname/page.tsx
    ├── procurement/
    │   ├── orders/page.tsx
    │   ├── receipts/page.tsx
    │   └── invoices/page.tsx
    ├── sales/
    │   ├── orders/page.tsx
    │   ├── deliveries/page.tsx
    │   └── invoices/page.tsx
    ├── finance/
    │   ├── journal/page.tsx
    │   ├── cash-bank/page.tsx
    │   └── reports/page.tsx
    └── settings/
        ├── company/page.tsx
        ├── users/page.tsx
        └── roles/page.tsx
```

### 2.2 Component Hierarchy

```
RootLayout (Server Component)
│
├── (auth)/layout
│   └── Centered Card Layout
│       └── {children} (Login/Register)
│
└── (app)/layout (Client Component)
    └── SidebarProvider
        ├── AppSidebar (Client Component)
        │   ├── SidebarHeader
        │   │   └── TeamSwitcher (Multi-tenant selector)
        │   ├── SidebarContent
        │   │   ├── NavMain (Primary navigation)
        │   │   └── NavProjects (Quick links - optional)
        │   └── SidebarFooter
        │       └── NavUser (User profile menu)
        │
        └── SidebarInset
            ├── Header
            │   ├── SidebarTrigger (Toggle button)
            │   ├── Separator
            │   ├── Breadcrumb (Navigation path)
            │   └── Actions (Page-specific actions)
            │
            └── Main Content
                └── {children} (Page content)
```

---

## 3. File Structure

### 3.1 New Files to Create

```
src/app/
├── (auth)/
│   ├── layout.tsx               # NEW - Auth layout
│   ├── login/
│   │   └── page.tsx             # NEW - Login page
│   └── register/
│       └── page.tsx             # NEW - Register page (optional)
│
├── (app)/
│   ├── layout.tsx               # NEW - Main app layout with sidebar
│   └── [existing routes move here]
│
src/lib/
├── store/                        # NEW - Redux Toolkit store
│   ├── index.ts                 # Store configuration
│   ├── slices/
│   │   ├── tenantSlice.ts      # Tenant state management
│   │   └── userSlice.ts        # User state management
│   └── hooks.ts                 # Typed Redux hooks
│
└── api/                         # NEW - API client utilities
    ├── client.ts                # Axios/Fetch wrapper with tenant context
    └── endpoints.ts             # API endpoint definitions
```

### 3.2 Files to Modify

```
src/components/
├── app-sidebar.tsx              # UPDATE - Add ERP navigation data
├── team-switcher.tsx            # UPDATE - Connect to real tenant API
├── nav-main.tsx                 # KEEP - Works as is
├── nav-user.tsx                 # UPDATE - Connect to user API
└── nav-projects.tsx             # OPTIONAL - Can remove if not needed

src/app/
├── layout.tsx                   # UPDATE - Simplify, add metadata
└── dashboard/
    └── page.tsx                 # UPDATE - Remove SidebarProvider (now in layout)
```

---

## 4. Component Specifications

### 4.1 Root Layout (`app/layout.tsx`)

**Type:** Server Component
**Purpose:** Provides HTML structure, fonts, global styles, and metadata

```typescript
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "ERP Distribution - Distribusi Sembako",
  description: "Sistem ERP untuk distribusi sembako dan bahan pokok",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
```

**Key Features:**

- Server component (no "use client")
- Indonesian language (`lang="id"`)
- Geist font optimization
- Global metadata
- suppressHydrationWarning for theme providers if added later

---

### 4.2 Auth Layout (`app/(auth)/layout.tsx`)

**Type:** Server Component
**Purpose:** Centered layout for login/register pages

```typescript
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/50">
      <div className="w-full max-w-md p-6">{children}</div>
    </div>
  );
}
```

**Key Features:**

- Centered card layout
- No sidebar
- Responsive max-width
- Consistent spacing

---

### 4.3 App Layout (`app/(app)/layout.tsx`)

**Type:** Client Component
**Purpose:** Main application layout with sidebar for authenticated routes

```typescript
"use client";

import { AppSidebar } from "@/components/app-sidebar";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider defaultOpen={true}>
      <AppSidebar />
      <SidebarInset>{children}</SidebarInset>
    </SidebarProvider>
  );
}
```

**Key Features:**

- Client component (uses SidebarProvider context)
- SidebarProvider wraps entire app section
- Default open state (can be controlled via props)
- Clean composition pattern

**Future Enhancements:**

- Add Redux Provider wrapper
- Add authentication check
- Add tenant context provider
- Add loading states

---

### 4.4 AppSidebar Component (`components/app-sidebar.tsx`)

**Type:** Client Component
**Purpose:** Main navigation sidebar with ERP menu structure

```typescript
"use client";

import * as React from "react";
import {
  LayoutDashboard,
  Database,
  Package,
  ShoppingCart,
  TrendingUp,
  Wallet,
  Settings,
} from "lucide-react";

import { NavMain } from "@/components/nav-main";
import { NavUser } from "@/components/nav-user";
import { TeamSwitcher } from "@/components/team-switcher";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar";

// ERP Navigation Data Structure
const erpNavigation = {
  navMain: [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: LayoutDashboard,
      isActive: true,
    },
    {
      title: "Master Data",
      url: "#",
      icon: Database,
      items: [
        { title: "Pelanggan", url: "/master/customers" },
        { title: "Pemasok", url: "/master/suppliers" },
        { title: "Produk", url: "/master/products" },
        { title: "Gudang", url: "/master/warehouses" },
      ],
    },
    {
      title: "Persediaan",
      url: "#",
      icon: Package,
      items: [
        { title: "Stok Barang", url: "/inventory/stock" },
        { title: "Transfer Gudang", url: "/inventory/transfers" },
        { title: "Stock Opname", url: "/inventory/opname" },
        { title: "Penyesuaian", url: "/inventory/adjustments" },
      ],
    },
    {
      title: "Pembelian",
      url: "#",
      icon: ShoppingCart,
      items: [
        { title: "Purchase Order", url: "/procurement/orders" },
        { title: "Penerimaan Barang", url: "/procurement/receipts" },
        { title: "Faktur Pembelian", url: "/procurement/invoices" },
        { title: "Pembayaran", url: "/procurement/payments" },
      ],
    },
    {
      title: "Penjualan",
      url: "#",
      icon: TrendingUp,
      items: [
        { title: "Sales Order", url: "/sales/orders" },
        { title: "Pengiriman", url: "/sales/deliveries" },
        { title: "Faktur Penjualan", url: "/sales/invoices" },
        { title: "Penerimaan Kas", url: "/sales/payments" },
      ],
    },
    {
      title: "Keuangan",
      url: "#",
      icon: Wallet,
      items: [
        { title: "Jurnal Umum", url: "/finance/journal" },
        { title: "Kas & Bank", url: "/finance/cash-bank" },
        { title: "Biaya", url: "/finance/expenses" },
        { title: "Laporan", url: "/finance/reports" },
      ],
    },
    {
      title: "Pengaturan",
      url: "#",
      icon: Settings,
      items: [
        { title: "Profil Perusahaan", url: "/settings/company" },
        { title: "Pengguna", url: "/settings/users" },
        { title: "Roles & Permissions", url: "/settings/roles" },
        { title: "Konfigurasi", url: "/settings/config" },
      ],
    },
  ],
  // Sample user data - replace with API call
  user: {
    name: "Admin User",
    email: "admin@example.com",
    avatar: "/avatars/admin.jpg",
  },
  // Sample tenant data - replace with API call
  teams: [
    {
      name: "PT Distribusi Utama",
      logo: Database,
      plan: "Enterprise",
    },
    {
      name: "CV Sembako Jaya",
      logo: Package,
      plan: "Professional",
    },
  ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher teams={erpNavigation.teams} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={erpNavigation.navMain} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={erpNavigation.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
```

**Key Features:**

- Indonesian labels for ERP workflows
- Icon-based navigation with Lucide icons
- Collapsible sidebar (icon mode supported)
- Multi-tenant team switcher in header
- User profile menu in footer
- Scrollable rail for resize handle

**Data Flow:**

1. Hard-coded sample data (current)
2. Future: Fetch from API in layout, pass as props
3. Future: Use Redux for tenant/user state
4. Future: Filter navigation by permissions

---

### 4.5 Dashboard Page (`app/(app)/dashboard/page.tsx`)

**Type:** Server Component (default)
**Purpose:** Main dashboard page - simplified without SidebarProvider

```typescript
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";

export default function DashboardPage() {
  return (
    <>
      <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
        <div className="flex items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        <h1 className="text-2xl font-bold">Dashboard ERP</h1>
        <div className="grid auto-rows-min gap-4 md:grid-cols-3">
          <div className="aspect-video rounded-xl bg-muted/50 p-4">
            <h3 className="font-semibold">Total Penjualan</h3>
            {/* Add dashboard widgets here */}
          </div>
          <div className="aspect-video rounded-xl bg-muted/50 p-4">
            <h3 className="font-semibold">Stok Menipis</h3>
            {/* Add dashboard widgets here */}
          </div>
          <div className="aspect-video rounded-xl bg-muted/50 p-4">
            <h3 className="font-semibold">Hutang/Piutang</h3>
            {/* Add dashboard widgets here */}
          </div>
        </div>
        <div className="min-h-screen flex-1 rounded-xl bg-muted/50 p-4 md:min-h-min">
          {/* Main dashboard content */}
          <p className="text-muted-foreground">
            Konten dashboard akan ditampilkan di sini.
          </p>
        </div>
      </div>
    </>
  );
}
```

**Key Changes from Original:**

- ✅ Removed SidebarProvider (now in layout)
- ✅ Removed AppSidebar import (in layout)
- ✅ Removed SidebarInset wrapper (in layout)
- ✅ Kept header with breadcrumbs and trigger
- ✅ Added Indonesian placeholder content

---

## 5. Data Flow & State Management

### 5.1 Multi-Tenant State Management

```typescript
// src/lib/store/index.ts
import { configureStore } from "@reduxjs/toolkit";
import tenantReducer from "./slices/tenantSlice";
import userReducer from "./slices/userSlice";

export const store = configureStore({
  reducer: {
    tenant: tenantReducer,
    user: userReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

```typescript
// src/lib/store/slices/tenantSlice.ts
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

interface TenantState {
  currentTenantId: string | null;
  tenants: Array<{
    id: string;
    name: string;
    logo?: string;
    plan: string;
  }>;
}

const initialState: TenantState = {
  currentTenantId: null,
  tenants: [],
};

export const tenantSlice = createSlice({
  name: "tenant",
  initialState,
  reducers: {
    setCurrentTenant: (state, action: PayloadAction<string>) => {
      state.currentTenantId = action.payload;
      // Store in localStorage for persistence
      localStorage.setItem("currentTenantId", action.payload);
    },
    setTenants: (state, action: PayloadAction<TenantState["tenants"]>) => {
      state.tenants = action.payload;
    },
  },
});

export const { setCurrentTenant, setTenants } = tenantSlice.actions;
export default tenantSlice.reducer;
```

### 5.2 API Integration Pattern

```typescript
// src/lib/api/client.ts
import { store } from "@/lib/store";

export async function apiRequest<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const state = store.getState();
  const tenantId = state.tenant.currentTenantId;

  const headers = new Headers(options?.headers);

  // Add tenant header
  if (tenantId) {
    headers.set("X-Tenant-ID", tenantId);
  }

  // Add JWT token
  const token = localStorage.getItem("authToken");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}${endpoint}`,
    {
      ...options,
      headers,
    }
  );

  if (!response.ok) {
    throw new Error(`API Error: ${response.statusText}`);
  }

  return response.json();
}
```

### 5.3 Data Fetching in Layout

```typescript
// app/(app)/layout.tsx - Enhanced version
"use client";

import { useEffect } from "react";
import { useDispatch } from "react-redux";
import { AppSidebar } from "@/components/app-sidebar";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { apiRequest } from "@/lib/api/client";
import { setTenants } from "@/lib/store/slices/tenantSlice";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const dispatch = useDispatch();

  useEffect(() => {
    // Fetch user's tenants on layout mount
    async function fetchTenants() {
      try {
        const data = await apiRequest("/api/tenants/user-tenants");
        dispatch(setTenants(data));
      } catch (error) {
        console.error("Failed to fetch tenants:", error);
      }
    }

    fetchTenants();
  }, [dispatch]);

  return (
    <SidebarProvider defaultOpen={true}>
      <AppSidebar />
      <SidebarInset>{children}</SidebarInset>
    </SidebarProvider>
  );
}
```

---

## 6. Responsive Design

### 6.1 Breakpoints

| Device  | Width          | Sidebar Behavior                 |
| ------- | -------------- | -------------------------------- |
| Mobile  | < 768px        | Sheet overlay (slide-in)         |
| Tablet  | 768px - 1024px | Collapsible, starts in icon mode |
| Desktop | > 1024px       | Collapsible, starts expanded     |

### 6.2 Sidebar States

**Desktop:**

- **Expanded:** 16rem width (256px)
- **Collapsed:** 3rem width (48px) - icon-only mode
- **Transition:** Smooth animation via CSS

**Mobile:**

- **Closed:** Hidden completely
- **Open:** Sheet overlay, 18rem width (288px)
- **Backdrop:** Dark overlay with tap-to-close

### 6.3 Header Responsive Behavior

```typescript
// Desktop: Full breadcrumbs
<Breadcrumb>
  <BreadcrumbList>
    <BreadcrumbItem className="hidden md:block">
      <BreadcrumbLink href="/dashboard">Dashboard</BreadcrumbLink>
    </BreadcrumbItem>
    <BreadcrumbSeparator className="hidden md:block" />
    <BreadcrumbItem>
      <BreadcrumbPage>Current Page</BreadcrumbPage>
    </BreadcrumbItem>
  </BreadcrumbList>
</Breadcrumb>

// Mobile: Only current page, hide parent breadcrumbs
```

---

## 7. Performance Optimization

### 7.1 Code Splitting

- ✅ AppSidebar is client component → separate bundle
- ✅ Route groups automatically code-split pages
- ✅ Dynamic imports for heavy components (charts, tables)
- ✅ Lazy load modals and dialogs

### 7.2 Data Caching Strategy

```typescript
// Example with React Query (future enhancement)
import { useQuery } from "@tanstack/react-query";

function useTenants() {
  return useQuery({
    queryKey: ["tenants"],
    queryFn: () => apiRequest("/api/tenants/user-tenants"),
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
}
```

### 7.3 Bundle Size Optimization

- ✅ Tree-shake unused Lucide icons
- ✅ Use Next.js font optimization (Geist already optimized)
- ✅ Minimize client components
- ✅ Server components by default

---

## 8. Security Considerations

### 8.1 Authentication Flow

```
1. User visits /dashboard
2. Middleware checks JWT token
3. If invalid/missing → redirect to /login
4. If valid → allow access to (app) routes
5. Layout fetches user data and tenants
```

### 8.2 Authorization

- Filter navigation items by user permissions
- Backend validates all operations
- Defense in depth: client + server validation
- Never trust client-side checks alone

### 8.3 Multi-Tenant Security

```typescript
// Always validate tenant ownership on backend
// Example backend validation (Go):
func ValidateTenantAccess(userID, tenantID string) bool {
    // Check if user has access to this tenant
    // Query user_tenants table or similar
    return hasAccess
}

// Frontend sends tenant ID in header
headers: {
  'X-Tenant-ID': currentTenantId,
  'Authorization': `Bearer ${jwtToken}`
}
```

---

## 9. Implementation Guide

### Phase 1: Base Structure (2-3 hours)

**Step 1:** Create route groups

```bash
mkdir -p src/app/\(auth\)/{login,register}
mkdir -p src/app/\(app\)
```

**Step 2:** Create auth layout

```bash
# Create src/app/(auth)/layout.tsx
# Copy centered layout code from specification
```

**Step 3:** Create app layout

```bash
# Create src/app/(app)/layout.tsx
# Copy SidebarProvider layout code from specification
```

**Step 4:** Move dashboard

```bash
# Move src/app/dashboard to src/app/(app)/dashboard
# Update imports to remove SidebarProvider
```

**Step 5:** Update root layout

```bash
# Update metadata and lang="id"
```

### Phase 2: ERP Navigation (1-2 hours)

**Step 1:** Update AppSidebar with ERP navigation

```bash
# Replace sample data in src/components/app-sidebar.tsx
# Add Indonesian labels and proper icons
```

**Step 2:** Test navigation

```bash
npm run dev
# Verify all menu items render correctly
# Test collapsible behavior
```

### Phase 3: Multi-Tenant Setup (3-4 hours)

**Step 1:** Install Redux Toolkit

```bash
npm install @reduxjs/toolkit react-redux
```

**Step 2:** Create store structure

```bash
mkdir -p src/lib/store/slices
# Create index.ts, tenantSlice.ts, userSlice.ts
```

**Step 3:** Create API client

```bash
mkdir -p src/lib/api
# Create client.ts with tenant header logic
```

**Step 4:** Wrap app with Redux Provider

```typescript
// Update src/app/layout.tsx to include Provider
import { Provider } from "react-redux";
import { store } from "@/lib/store";

// Wrap children with <Provider store={store}>
```

**Step 5:** Update TeamSwitcher

```typescript
// Connect TeamSwitcher to Redux
// Dispatch setCurrentTenant on selection
```

### Phase 4: Create Module Pages (4-6 hours)

**Step 1:** Create route structure

```bash
mkdir -p src/app/\(app\)/{master,inventory,procurement,sales,finance,settings}
```

**Step 2:** Create placeholder pages

```typescript
// For each route, create page.tsx with:
// - Header with breadcrumbs
// - Indonesian title
// - Placeholder content
```

**Step 3:** Test routing

```bash
# Verify all routes accessible
# Test breadcrumb generation
```

### Phase 5: Polish & Testing (2-3 hours)

**Step 1:** Test mobile responsive

- Open DevTools mobile view
- Test Sheet overlay behavior
- Verify touch interactions

**Step 2:** Test keyboard shortcuts

- Cmd/Ctrl + B to toggle sidebar
- Verify focus management

**Step 3:** Add loading states

```typescript
// Add Skeleton components for loading
import { Skeleton } from "@/components/ui/skeleton";
```

**Step 4:** Test tenant switching

- Switch tenants in TeamSwitcher
- Verify context updates
- Check API headers include correct tenant ID

---

## 10. Testing Checklist

### 10.1 Functional Testing

- [ ] Sidebar toggles correctly (desktop)
- [ ] Sheet overlay works (mobile)
- [ ] Team switcher changes tenant
- [ ] Navigation items route correctly
- [ ] Breadcrumbs update per route
- [ ] User menu displays correctly
- [ ] Keyboard shortcut (Cmd/Ctrl + B) works
- [ ] Sidebar state persists via cookie

### 10.2 Responsive Testing

- [ ] Mobile (375px): Sheet overlay
- [ ] Tablet (768px): Collapsed sidebar
- [ ] Desktop (1440px): Expanded sidebar
- [ ] Ultra-wide (1920px+): Max-width content

### 10.3 Performance Testing

- [ ] Page load < 2s on 3G
- [ ] Sidebar toggle smooth (60fps)
- [ ] No layout shift on load
- [ ] Code-split bundles verify
- [ ] Fonts load optimally

### 10.4 Security Testing

- [ ] Unauthenticated users redirect to login
- [ ] Tenant ID in API headers
- [ ] JWT token validation
- [ ] XSS protection (no user-generated content yet)
- [ ] Navigation filtered by permissions (future)

---

## 11. Future Enhancements

### 11.1 Short-term (1-2 weeks)

- Real API integration for user/tenant data
- Permission-based navigation filtering
- Loading states and error handling
- Dark mode support
- Notification system in header

### 11.2 Medium-term (1-2 months)

- Search functionality in sidebar
- Favorites/pinned menu items
- Recent pages history
- Keyboard navigation (arrow keys)
- Multi-language support (English/Indonesian toggle)

### 11.3 Long-term (3-6 months)

- Customizable dashboard widgets
- User preferences for sidebar width
- Collapsible menu groups
- Advanced breadcrumb with dropdowns
- Command palette (Cmd+K)

---

## 12. Appendix

### 12.1 Icon Reference

```typescript
import {
  LayoutDashboard, // Dashboard
  Database, // Master Data
  Package, // Inventory
  ShoppingCart, // Procurement
  TrendingUp, // Sales
  Wallet, // Finance
  Settings, // Settings
  Users, // Users
  Building, // Warehouses
  FileText, // Invoices/Documents
  TruckIcon, // Deliveries
  CreditCard, // Payments
} from "lucide-react";
```

### 12.2 Color Palette (from globals.css)

```css
--background: 0 0% 100%;
--foreground: 240 10% 3.9%;
--muted: 240 4.8% 95.9%;
--muted-foreground: 240 3.8% 46.1%;
--accent: 240 4.8% 95.9%;
--accent-foreground: 240 5.9% 10%;
```

### 12.3 Useful Resources

- [shadcn/ui Sidebar Docs](https://ui.shadcn.com/docs/components/sidebar)
- [Next.js 16 App Router](https://nextjs.org/docs/app)
- [Lucide Icons](https://lucide.dev/icons/)
- [Redux Toolkit](https://redux-toolkit.js.org/)
- [Tailwind CSS](https://tailwindcss.com/)

---

## 13. Questions & Answers

**Q: Why use route groups (auth) and (app)?**
A: Route groups allow different layouts without affecting URLs. (auth) for public pages without sidebar, (app) for authenticated pages with sidebar.

**Q: Should sidebar be server or client component?**
A: Client component because it uses SidebarProvider context and interactive state (collapse/expand).

**Q: How to handle tenant switching?**
A: Use Redux to store currentTenantId, include in API headers (X-Tenant-ID), backend validates access.

**Q: What about authentication middleware?**
A: Create middleware.ts in root to check JWT and redirect unauthenticated users to /login.

**Q: How to filter navigation by permissions?**
A: Fetch user permissions, filter navMain items before rendering, but always validate on backend.

---

**END OF SPECIFICATION**

This design is ready for implementation. Follow the phases in section 9 for systematic rollout.
