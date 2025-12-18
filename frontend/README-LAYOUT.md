# ERP Frontend - Base Layout Quick Start

**Status:** ✅ Implemented
**Version:** 1.0
**Date:** 2025-12-17

---

## 🚀 Quick Start

### Development Server

```bash
npm run dev
```

**Available Routes:**
- **Dashboard**: http://localhost:3000/dashboard
- **Login**: http://localhost:3000/login
- **Home**: http://localhost:3000

### Build & Test

```bash
# Production build
npm run build

# Start production server
npm start
```

---

## 📁 Project Structure

```
src/app/
├── layout.tsx                  # Root layout (Indonesian, metadata)
│
├── (auth)/                     # PUBLIC ROUTES (no sidebar)
│   ├── layout.tsx             # Centered card layout
│   └── login/page.tsx         # Login form
│
└── (app)/                      # AUTHENTICATED ROUTES (with sidebar)
    ├── layout.tsx             # SidebarProvider wrapper
    └── dashboard/page.tsx     # Dashboard page
```

---

## 🎯 How to Add New Pages

### Example: Create Customer List Page

**1. Create directory:**
```bash
mkdir -p src/app/\(app\)/master/customers
```

**2. Create page component:**
```typescript
// src/app/(app)/master/customers/page.tsx
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"

export default function CustomersPage() {
  return (
    <>
      <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
        <div className="flex items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Pelanggan</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        <h1 className="text-2xl font-bold">Daftar Pelanggan</h1>
        {/* Your content here */}
      </div>
    </>
  )
}
```

**3. Navigate to http://localhost:3000/master/customers**

The page automatically gets the sidebar from `(app)/layout.tsx`!

---

## 📋 Navigation Structure

The sidebar includes 7 main categories with **25 navigation items**:

### 1. Dashboard
- Direct link to `/dashboard`

### 2. Master Data
- `/master/customers` - Pelanggan
- `/master/suppliers` - Pemasok
- `/master/products` - Produk
- `/master/warehouses` - Gudang

### 3. Persediaan
- `/inventory/stock` - Stok Barang
- `/inventory/transfers` - Transfer Gudang
- `/inventory/opname` - Stock Opname
- `/inventory/adjustments` - Penyesuaian

### 4. Pembelian
- `/procurement/orders` - Purchase Order
- `/procurement/receipts` - Penerimaan Barang
- `/procurement/invoices` - Faktur Pembelian
- `/procurement/payments` - Pembayaran

### 5. Penjualan
- `/sales/orders` - Sales Order
- `/sales/deliveries` - Pengiriman
- `/sales/invoices` - Faktur Penjualan
- `/sales/payments` - Penerimaan Kas

### 6. Keuangan
- `/finance/journal` - Jurnal Umum
- `/finance/cash-bank` - Kas & Bank
- `/finance/expenses` - Biaya
- `/finance/reports` - Laporan

### 7. Pengaturan
- `/settings/company` - Profil Perusahaan
- `/settings/users` - Pengguna
- `/settings/roles` - Roles & Permissions
- `/settings/config` - Konfigurasi

---

## 🎨 Sidebar Features

### Desktop Features
- **Toggle**: Click trigger or press `Cmd/Ctrl + B`
- **Collapsed Mode**: Icon-only view (3rem width)
- **Expanded Mode**: Full text (16rem width)
- **Persistent State**: Saved in cookie (7 days)

### Mobile Features (< 768px)
- **Sheet Overlay**: Slide-in sidebar (18rem width)
- **Touch-Friendly**: Large tap targets
- **Auto-Close**: Closes after navigation

### Multi-Tenant
- **Team Switcher**: Switch between organizations
- **Current Tenants**:
  - PT Distribusi Utama (Enterprise)
  - CV Sembako Jaya (Professional)

---

## 🔧 Common Tasks

### Add Navigation Item

Edit `src/components/app-sidebar.tsx`:

```typescript
navMain: [
  // ... existing items
  {
    title: "New Category",
    url: "#",
    icon: YourIcon,  // Import from lucide-react
    items: [
      {
        title: "Sub Item 1",
        url: "/category/item1",
      },
    ],
  },
]
```

### Change Page Title

Edit `src/app/layout.tsx`:

```typescript
export const metadata: Metadata = {
  title: "Your Title Here",
  description: "Your description",
};
```

### Add More Tenants

Edit `src/components/app-sidebar.tsx`:

```typescript
teams: [
  // ... existing teams
  {
    name: "PT Your Company",
    logo: YourIcon,
    plan: "Professional",
  },
]
```

---

## 🎯 Layout Pattern

### Route Groups Explained

**What are Route Groups?**
- Folders with parentheses: `(auth)`, `(app)`
- **Do NOT** appear in URLs
- Allow different layouts for different sections

**Example:**
```
File:  src/app/(auth)/login/page.tsx
URL:   http://localhost:3000/login
       (no /auth/ in URL!)

File:  src/app/(app)/dashboard/page.tsx
URL:   http://localhost:3000/dashboard
       (no /app/ in URL!)
```

### Layout Inheritance

```
Root Layout (always applies)
    ↓
├── (auth)/layout → Centered card
│   └── login/page
│
└── (app)/layout → Sidebar
    └── dashboard/page
```

---

## 📦 Components Available

### UI Components (shadcn/ui)
- `Button` - Buttons
- `Card` - Card containers
- `Input` - Form inputs
- `Label` - Form labels
- `Separator` - Visual separators
- `Sidebar` - Sidebar primitives
- `Breadcrumb` - Breadcrumb navigation

### Custom Components
- `AppSidebar` - Main ERP sidebar
- `NavMain` - Primary navigation
- `NavUser` - User menu
- `TeamSwitcher` - Multi-tenant switcher

---

## 🐛 Troubleshooting

### Build Errors

**Problem:** `Module not found: '@/components/ui/...'`

**Solution:**
```bash
# Install missing shadcn component
npx shadcn@latest add [component-name]
```

### Sidebar Not Showing

**Check:**
1. Are you in `(app)` route group?
2. Is page inside `src/app/(app)/`?
3. Did you remove SidebarProvider from page?

**Layout provides sidebar automatically!**

### Routes Return 404

**Issue:** Menu item links to non-existent page

**Solution:** Create the page file:
```bash
# For /master/customers
mkdir -p src/app/\(app\)/master/customers
# Create page.tsx in that directory
```

---

## 📚 Documentation

### Full Documentation
- **Design Spec**: `claudedocs/base-layout-design.md` (9,200 lines)
- **Implementation Summary**: `claudedocs/implementation-summary.md`

### Official Resources
- [Next.js 16 App Router](https://nextjs.org/docs/app)
- [shadcn/ui Documentation](https://ui.shadcn.com)
- [Lucide Icons](https://lucide.dev)

---

## ⚡ Performance

### Build Metrics
- **Build Time**: 2.4s
- **Static Pages**: 6 pages
- **Bundle Size**: Optimized with code splitting
- **TypeScript**: Strict mode ✅

### Runtime Features
- **Partial Prerendering**: Ready for PPR
- **Server Components**: Default
- **Client Components**: Only where needed
- **Font Optimization**: Geist fonts optimized

---

## 🎓 Best Practices

### When Creating Pages

✅ **DO:**
- Use server components by default
- Add "use client" only when needed (hooks, state, events)
- Include breadcrumbs in header
- Use Indonesian labels
- Follow existing dashboard pattern

❌ **DON'T:**
- Add SidebarProvider to pages (it's in layout!)
- Import AppSidebar in pages (it's in layout!)
- Wrap content in SidebarInset (layout does it!)

### Example Page Template

```typescript
// Server component (default)
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"

export default function YourPage() {
  return (
    <>
      <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
        <div className="flex items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Page Title</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        <h1 className="text-2xl font-bold">Your Content</h1>
        {/* Content here */}
      </div>
    </>
  )
}
```

---

## 🚀 Next Steps

### Immediate (Ready to Implement)
1. Create module pages (master, inventory, procurement, etc.)
2. Add actual content to dashboard widgets
3. Implement real data fetching

### Short-term (1-2 weeks)
1. Set up Redux Toolkit for state management
2. Connect TeamSwitcher to real tenant API
3. Implement authentication flow
4. Add protected route middleware

### Medium-term (1-2 months)
1. Build CRUD operations for all modules
2. Implement permission-based navigation
3. Add real-time updates
4. Create comprehensive testing

---

## 💡 Tips

### Development Workflow

1. **Start dev server**: `npm run dev`
2. **Create page**: `mkdir -p src/app/(app)/your/route`
3. **Add page.tsx**: Copy template from above
4. **Customize**: Add your content
5. **Test**: Navigate to URL in browser
6. **Sidebar works automatically!**

### Styling Tips

- Use Tailwind CSS classes
- Refer to `globals.css` for theme colors
- Use `cn()` utility for conditional classes
- Check shadcn/ui docs for component props

### Icon Usage

```typescript
// Import from lucide-react
import { Package, Users, Settings } from "lucide-react"

// Use in navigation
{
  title: "Your Menu",
  icon: Package,
  // ...
}
```

---

## ✅ What's Included

- ✅ Route group architecture
- ✅ Responsive sidebar layout
- ✅ ERP navigation (25 items)
- ✅ Multi-tenant switcher
- ✅ Login page
- ✅ Dashboard page
- ✅ Indonesian localization
- ✅ Mobile-responsive design
- ✅ Keyboard shortcuts
- ✅ State persistence
- ✅ TypeScript strict mode
- ✅ Production-ready build

---

## 📞 Support

**Questions?**
- Check `claudedocs/` for detailed documentation
- Review existing page implementations
- Follow the patterns in `dashboard/page.tsx`

**Happy Coding! 🎉**
