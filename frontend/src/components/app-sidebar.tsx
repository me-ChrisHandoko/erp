"use client"

import * as React from "react"
import { useSelector } from "react-redux"
import {
  LayoutDashboard,
  Database,
  Package,
  ShoppingCart,
  TrendingUp,
  Wallet,
  Settings,
  Building2,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavUser } from "@/components/nav-user"
import { TeamSwitcher } from "@/components/team-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"
import { RootState } from "@/store"
import { usePermissions } from "@/hooks/use-permissions"
import type { Resource } from "@/lib/permissions"

// ERP Navigation Data with Permission Mapping
// Note: Team/Company data now comes from Redux via TeamSwitcher component (PHASE 5)
// PHASE 6: Added resource mapping for permission-based filtering

const staticData = {
  navMain: [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: LayoutDashboard,
      isActive: true,
      // Dashboard tidak memerlukan permission check - semua role bisa akses
    },
    {
      title: "Perusahaan",
      url: "#",
      icon: Building2,
      items: [
        {
          title: "Profil Perusahaan",
          url: "/company/profile",
          resource: "company-settings" as Resource,
        },
        {
          title: "Rekening Bank",
          url: "/company/banks",
          resource: "bank-accounts" as Resource,
        },
        {
          title: "Tim & Pengguna",
          url: "/company/team",
          resource: "users" as Resource,
        },
      ],
    },
    {
      title: "Master Data",
      url: "#",
      icon: Database,
      items: [
        {
          title: "Pelanggan",
          url: "/master/customers",
          resource: "customers" as Resource,
        },
        {
          title: "Pemasok",
          url: "/master/suppliers",
          resource: "suppliers" as Resource,
        },
        {
          title: "Produk",
          url: "/products",
          resource: "products" as Resource,
        },
        {
          title: "Gudang",
          url: "/master/warehouses",
          resource: "warehouses" as Resource,
        },
      ],
    },
    {
      title: "Persediaan",
      url: "#",
      icon: Package,
      items: [
        {
          title: "Stok Barang",
          url: "/inventory/stock",
          resource: "stock" as Resource,
        },
        {
          title: "Transfer Gudang",
          url: "/inventory/transfers",
          resource: "stock-transfers" as Resource,
        },
        {
          title: "Stock Opname",
          url: "/inventory/opname",
          resource: "stock-opname" as Resource,
        },
        {
          title: "Penyesuaian",
          url: "/inventory/adjustments",
          resource: "inventory-adjustments" as Resource,
        },
      ],
    },
    {
      title: "Pembelian",
      url: "#",
      icon: ShoppingCart,
      items: [
        {
          title: "Purchase Order",
          url: "/procurement/orders",
          resource: "purchase-orders" as Resource,
        },
        {
          title: "Penerimaan Barang",
          url: "/procurement/receipts",
          resource: "goods-receipts" as Resource,
        },
        {
          title: "Faktur Pembelian",
          url: "/procurement/invoices",
          resource: "purchase-invoices" as Resource,
        },
        {
          title: "Pembayaran",
          url: "/procurement/payments",
          resource: "supplier-payments" as Resource,
        },
      ],
    },
    {
      title: "Penjualan",
      url: "#",
      icon: TrendingUp,
      items: [
        {
          title: "Sales Order",
          url: "/sales/orders",
          resource: "sales-orders" as Resource,
        },
        {
          title: "Pengiriman",
          url: "/sales/deliveries",
          resource: "deliveries" as Resource,
        },
        {
          title: "Faktur Penjualan",
          url: "/sales/invoices",
          resource: "sales-invoices" as Resource,
        },
        {
          title: "Penerimaan Kas",
          url: "/sales/payments",
          resource: "customer-payments" as Resource,
        },
      ],
    },
    {
      title: "Keuangan",
      url: "#",
      icon: Wallet,
      items: [
        {
          title: "Jurnal Umum",
          url: "/finance/journal",
          resource: "journal-entries" as Resource,
        },
        {
          title: "Kas & Bank",
          url: "/finance/cash-bank",
          resource: "cash-bank" as Resource,
        },
        {
          title: "Biaya",
          url: "/finance/expenses",
          resource: "expenses" as Resource,
        },
        {
          title: "Laporan",
          url: "/finance/reports",
          resource: "financial-reports" as Resource,
        },
      ],
    },
    {
      title: "Pengaturan",
      url: "#",
      icon: Settings,
      items: [
        {
          title: "Roles & Permissions",
          url: "/settings/roles",
          resource: "roles" as Resource,
        },
        {
          title: "Konfigurasi Sistem",
          url: "/settings/config",
          resource: "system-config" as Resource,
        },
        {
          title: "Preferensi",
          url: "/settings/preferences",
          resource: "preferences" as Resource,
        },
      ],
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  // Get authenticated user from Redux
  const authUser = useSelector((state: RootState) => state.auth.user)

  // PHASE 6: Get permission helper
  const { canAny } = usePermissions()

  // Map user data to NavUser props format
  // Backend returns 'fullName', but NavUser expects 'name'
  const userData = authUser
    ? {
        // Fallback to email username if fullName is empty (during session restore)
        name: authUser.fullName || authUser.email.split('@')[0],
        email: authUser.email,
        avatar: "", // No avatar image - will use Lucide icon fallback
      }
    : {
        name: "Guest User",
        email: "guest@example.com",
        avatar: "", // No avatar image - will use Lucide icon fallback
      }

  // PHASE 6: Filter navigation items based on user permissions
  const filteredNavItems = React.useMemo(() => {
    return staticData.navMain.map((item) => {
      // If item has no sub-items, return as is (e.g., Dashboard)
      if (!item.items || item.items.length === 0) {
        return item;
      }

      // Filter sub-items based on permissions
      const filteredSubItems = item.items.filter((subItem) => {
        // If sub-item has no resource, show it (backward compatibility)
        if (!subItem.resource) return true;

        // Check if user has any permission on this resource
        return canAny(subItem.resource);
      });

      // Return item with filtered sub-items
      return {
        ...item,
        items: filteredSubItems,
      };
    }).filter((item) => {
      // Hide parent item if all sub-items are filtered out
      if (item.items && item.items.length === 0) return false;
      return true;
    });
  }, [canAny]);

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        {/* PHASE 5: TeamSwitcher now uses real company data from Redux */}
        <TeamSwitcher />
      </SidebarHeader>
      <SidebarContent>
        {/* PHASE 6: Pass filtered navigation items based on permissions */}
        <NavMain items={filteredNavItems} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={userData} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
