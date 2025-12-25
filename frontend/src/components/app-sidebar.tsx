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
  PackageOpen,
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

// ERP Navigation Data
const staticData = {
  teams: [
    {
      name: "PT Distribusi Utama",
      logo: Building2,
      plan: "Enterprise",
    },
    {
      name: "CV Sembako Jaya",
      logo: PackageOpen,
      plan: "Professional",
    },
  ],
  navMain: [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: LayoutDashboard,
      isActive: true,
    },
    {
      title: "Perusahaan",
      url: "#",
      icon: Building2,
      items: [
        {
          title: "Profil Perusahaan",
          url: "/company/profile",
        },
        {
          title: "Rekening Bank",
          url: "/company/banks",
        },
        {
          title: "Tim & Pengguna",
          url: "/company/team",
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
        },
        {
          title: "Pemasok",
          url: "/master/suppliers",
        },
        {
          title: "Produk",
          url: "/master/products",
        },
        {
          title: "Gudang",
          url: "/master/warehouses",
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
        },
        {
          title: "Transfer Gudang",
          url: "/inventory/transfers",
        },
        {
          title: "Stock Opname",
          url: "/inventory/opname",
        },
        {
          title: "Penyesuaian",
          url: "/inventory/adjustments",
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
        },
        {
          title: "Penerimaan Barang",
          url: "/procurement/receipts",
        },
        {
          title: "Faktur Pembelian",
          url: "/procurement/invoices",
        },
        {
          title: "Pembayaran",
          url: "/procurement/payments",
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
        },
        {
          title: "Pengiriman",
          url: "/sales/deliveries",
        },
        {
          title: "Faktur Penjualan",
          url: "/sales/invoices",
        },
        {
          title: "Penerimaan Kas",
          url: "/sales/payments",
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
        },
        {
          title: "Kas & Bank",
          url: "/finance/cash-bank",
        },
        {
          title: "Biaya",
          url: "/finance/expenses",
        },
        {
          title: "Laporan",
          url: "/finance/reports",
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
        },
        {
          title: "Konfigurasi Sistem",
          url: "/settings/config",
        },
        {
          title: "Preferensi",
          url: "/settings/preferences",
        },
      ],
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  // Get authenticated user from Redux
  const authUser = useSelector((state: RootState) => state.auth.user)

  console.log("[Sidebar] authUser from Redux:", {
    hasUser: !!authUser,
    email: authUser?.email,
    fullName: authUser?.fullName,
    fullNameLength: authUser?.fullName?.length,
  });

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

  console.log("[Sidebar] userData mapped:", {
    name: userData.name,
    email: userData.email,
  });

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher teams={staticData.teams} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={staticData.navMain} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={userData} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
