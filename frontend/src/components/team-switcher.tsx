"use client";

/**
 * TeamSwitcher Component
 *
 * Multi-company context switcher for sidebar
 * PHASE 5: Frontend State Management - Real company data integration
 *
 * Features:
 * - Displays active company with role badge
 * - Lists all available companies for user
 * - Handles company switching with loading state
 * - Shows "Add Company" option for OWNER role only
 * - Supports keyboard shortcuts (⌘1, ⌘2, etc.)
 * - Single company mode (no dropdown if only 1 company)
 * - Empty state (no companies available)
 */

import { ChevronsUpDown, Plus, Building2, CheckCircle2 } from "lucide-react";
import { toast } from "sonner";
import { useCompany } from "@/hooks/use-company";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { Badge } from "@/components/ui/badge";
import type { CompanyRole } from "@/types/company.types";

/**
 * Get role label in Indonesian
 */
function getRoleLabel(role: CompanyRole): string {
  const labels: Record<CompanyRole, string> = {
    OWNER: "Pemilik",
    ADMIN: "Admin",
    FINANCE: "Keuangan",
    SALES: "Penjualan",
    WAREHOUSE: "Gudang",
    STAFF: "Staff",
  };
  return labels[role];
}

/**
 * Company icon component
 * Uses Next.js Image for optimization when logo URL is available
 */
function CompanyIcon() {
  // For now, always use Building2 icon
  // TODO: Implement Next.js Image component when logo upload is ready
  return <Building2 className="size-4" />;
}

export function TeamSwitcher() {
  const { isMobile } = useSidebar();
  const {
    activeCompany,
    availableCompanies,
    switchCompany,
    canAddCompany,
    loading,
    hasNoCompanies,
  } = useCompany();

  /**
   * Handle company switch
   */
  const handleSwitchCompany = async (companyId: string) => {
    if (companyId === activeCompany?.id) return;

    const targetCompany = availableCompanies.find((c) => c.id === companyId);
    if (!targetCompany) return;

    try {
      const success = await switchCompany(companyId);
      if (success) {
        toast.success(`Beralih ke ${targetCompany.name}`);
        // No need for reload - RTK Query will auto-refetch when activeCompanyId changes
      } else {
        toast.error("Gagal beralih perusahaan");
      }
    } catch (error) {
      console.error("Failed to switch company:", error);
      const errorMessage =
        error && typeof error === "object" && "data" in error
          ? (error.data as { error?: { message?: string } })?.error?.message
          : null;
      toast.error(errorMessage || "Gagal beralih perusahaan");
    }
  };

  /**
   * Handle add company
   */
  const handleAddCompany = () => {
    // Navigate to create company page
    window.location.assign("/company/create");
  };

  // Empty state: no companies available
  if (hasNoCompanies || !activeCompany) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" disabled>
            <div className="bg-sidebar-secondary text-sidebar-secondary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
              <CompanyIcon />
            </div>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-medium text-muted-foreground">
                Tidak ada perusahaan
              </span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }

  // Single company mode: no dropdown needed
  if (availableCompanies.length === 1) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" disabled={loading}>
            <div className="bg-sidebar-secondary text-sidebar-secondary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
              <CompanyIcon />
            </div>
            <div className="grid flex-1 text-left text-sm leading-tight">
              <span className="truncate font-medium">{activeCompany.name}</span>
              <span className="truncate text-xs text-muted-foreground">
                {getRoleLabel(activeCompany.role)}
              </span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }

  // Multi-company mode: dropdown with all companies
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              disabled={loading}
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="bg-sidebar-secondary text-sidebar-secondary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                <CompanyIcon />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium">{activeCompany.name}</span>
                <span className="truncate text-xs text-muted-foreground">
                  {getRoleLabel(activeCompany.role)}
                </span>
              </div>
              <ChevronsUpDown className="ml-auto size-4" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-muted-foreground text-xs">
              Perusahaan
            </DropdownMenuLabel>
            {availableCompanies.map((company, index) => {
              const isActive = company.id === activeCompany.id;
              return (
                <DropdownMenuItem
                  key={company.id}
                  onClick={() => handleSwitchCompany(company.id)}
                  disabled={!company.isActive}
                  className="gap-2 p-2"
                >
                  <div className="flex size-6 items-center justify-center rounded-md border">
                    <CompanyIcon />
                  </div>
                  <div className="flex flex-1 flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm">{company.name}</span>
                      {isActive && (
                        <CheckCircle2 className="size-3.5 text-primary" />
                      )}
                      {!company.isActive && (
                        <Badge variant="outline" className="text-xs">
                          Nonaktif
                        </Badge>
                      )}
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {getRoleLabel(company.role)}
                    </span>
                  </div>
                  {index < 9 && (
                    <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
                  )}
                </DropdownMenuItem>
              );
            })}

            {/* Add Company option (OWNER only) */}
            {canAddCompany && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="gap-2 p-2" onClick={handleAddCompany}>
                  <div className="flex size-6 items-center justify-center rounded-md border bg-transparent">
                    <Plus className="size-4" />
                  </div>
                  <div className="text-muted-foreground font-medium">
                    Tambah Perusahaan
                  </div>
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
