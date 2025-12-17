# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **frontend application** for an ERP Distribution (Distribusi Sembako) system built with Next.js 16, React 19, TypeScript, and Tailwind CSS. It's part of a multi-tenant SaaS ERP solution for Indonesian food distribution businesses.

The frontend connects to a Go-based backend API (located in `../backend/`) that handles multi-tenancy, warehouse management, sales/purchase workflows, inventory control, and financial management.

## Tech Stack

- **Framework**: Next.js 16 (App Router with Turbopack, Server Actions, PPR support)
- **React**: 19.2.1
- **TypeScript**: 5.x with strict mode
- **Styling**: Tailwind CSS 4 with CSS variables
- **UI Components**: shadcn/ui (New York style) + Radix UI primitives
- **Icons**: Lucide React
- **State Management**: Redux Toolkit (@reduxjs/toolkit)
- **Authentication**: JWT (jsonwebtoken)

## Development Commands

```bash
# Start development server (http://localhost:3000)
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Run linting
npm run lint
```

## Project Structure

```
src/
├── app/                    # Next.js App Router pages
│   ├── layout.tsx         # Root layout with Geist fonts
│   ├── page.tsx           # Home page
│   ├── globals.css        # Global Tailwind styles
│   └── dashboard/         # Dashboard pages
├── components/            # React components
│   ├── ui/               # shadcn/ui components (button, sidebar, etc.)
│   ├── app-sidebar.tsx   # Main application sidebar with team switcher
│   ├── nav-main.tsx      # Primary navigation component
│   ├── nav-projects.tsx  # Projects navigation
│   ├── nav-user.tsx      # User menu component
│   └── team-switcher.tsx # Multi-tenant team switching
├── hooks/                 # Custom React hooks
│   └── use-mobile.ts     # Mobile detection hook
└── lib/
    └── utils.ts          # Utility functions (cn for className merging)
```

## Architecture Guidelines

### Path Aliases

All imports use the `@/` alias pointing to `src/`:

```typescript
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { AppSidebar } from "@/components/app-sidebar"
```

### Component Patterns

1. **shadcn/ui Components**: Located in `src/components/ui/`, these are customizable Radix UI primitives styled with Tailwind. When adding new shadcn components, they automatically use:
   - Style: "new-york"
   - Base color: "neutral"
   - CSS variables for theming
   - Lucide icons

2. **Client Components**: Use `"use client"` directive for components with interactivity, hooks, or browser APIs (see `app-sidebar.tsx`)

3. **Server Components**: Default for `app/` directory pages unless marked with `"use client"`

### Styling Approach

- **Tailwind CSS 4**: Uses CSS variables for theming (defined in `globals.css`)
- **Dark Mode**: Supports dark mode with `dark:` prefix classes
- **Utility Function**: Use `cn()` from `@/lib/utils` to merge Tailwind classes with conditional logic

### Multi-Tenancy Considerations

The backend uses multi-tenancy with tenant isolation. When implementing frontend features:

1. **Tenant Context**: The `team-switcher.tsx` component handles switching between tenants/organizations
2. **API Calls**: Include tenant context in headers or request body when calling backend APIs
3. **State Management**: Use Redux Toolkit for managing tenant-specific state

### Type Safety

- **Strict TypeScript**: `tsconfig.json` has `"strict": true`
- **Type Imports**: Use `import type` for type-only imports
- **React 19 Types**: Uses new React 19 type definitions

## Backend Integration

The backend API (Go/Gin) is in `../backend/` and handles:
- Authentication (JWT)
- Multi-tenant data isolation
- Warehouse/inventory management
- Sales/purchase workflows
- Financial transactions

When implementing API calls:
1. Base URL typically points to backend (configure via environment variables)
2. Include JWT token in Authorization header
3. Handle tenant ID in request context
4. Expect Indonesian-specific fields (NPWP, PPN tax, etc.)

## Development Notes

### Adding New UI Components

Use shadcn CLI to add components:
```bash
npx shadcn@latest add [component-name]
```

Components will be added to `src/components/ui/` with proper configuration from `components.json`.

### Dashboard Layout Pattern

The dashboard uses a sidebar layout with:
- `SidebarProvider`: Manages sidebar state (collapsible/icon mode)
- `AppSidebar`: Left navigation with team switcher
- `SidebarInset`: Main content area with breadcrumbs
- Responsive design that works on mobile and desktop

### State Management

Redux Toolkit is included for complex state:
- Create slices in a `src/store/` directory (when needed)
- Use `createSlice` and `configureStore` from @reduxjs/toolkit
- Prefer React hooks (useState, useContext) for local component state

## Common Patterns

### Conditional Styling
```typescript
<div className={cn("base-classes", condition && "conditional-classes")} />
```

### Mobile Detection
```typescript
import { useMobile } from "@/hooks/use-mobile"

const isMobile = useMobile() // Returns boolean
```

### Component Composition
Follow shadcn/ui patterns for composable components with compound components (Sidebar, SidebarHeader, SidebarContent, etc.)

## Important Considerations

1. **Next.js 16 App Router**: Use server components by default, add `"use client"` only when needed
2. **React 19**: New features available (use, optimistic updates, server actions)
3. **Multi-Tenant Context**: Always consider which tenant/team the user is operating under
4. **Indonesian Locale**: Future components may need Indonesian language support and business rules
5. **Responsive Design**: Test both desktop sidebar and mobile navigation patterns
