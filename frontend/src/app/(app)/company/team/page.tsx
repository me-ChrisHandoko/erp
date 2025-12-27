/**
 * Team Management Page
 *
 * Displays tenant information, subscription status, and team members.
 * Allows OWNER/ADMIN to invite, edit roles, and remove users.
 *
 * Features:
 * - Tenant info card with subscription details
 * - User table with role and status filters
 * - Invite user functionality
 * - Edit user role functionality
 * - Remove user functionality
 * - RBAC protection (OWNER/ADMIN protection, last ADMIN protection)
 */

"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { UserPlus, Users } from "lucide-react";
import { TenantInfoCard } from "@/components/team/tenant-info-card";
import { UserTable } from "@/components/team/user-table";
import { InviteUserForm } from "@/components/team/invite-user-form";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { PageHeader } from "@/components/shared/page-header";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useGetTenantQuery } from "@/store/services/tenantApi";
import { useGetCompanyUsersQuery } from "@/store/services/companyUserApi";
import { useCompany } from "@/hooks/use-company";
import type { UserRole } from "@/types/tenant.types";

export default function TeamPage() {
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [roleFilter, setRoleFilter] = useState<UserRole | undefined>(undefined);
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(
    undefined
  );

  // Get active company from Redux
  const { activeCompany } = useCompany();

  // Fetch tenant info
  const {
    data: tenant,
    isLoading: tenantLoading,
    error: tenantError,
  } = useGetTenantQuery();

  // Fetch company users with filters (uses X-Company-ID header automatically)
  const {
    data: users,
    isLoading: usersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useGetCompanyUsersQuery(
    {
      role: roleFilter,
      isActive: statusFilter,
    },
    {
      // Skip query if no active company selected
      skip: !activeCompany,
    }
  );

  const handleInviteSuccess = () => {
    setShowInviteDialog(false);
    refetchUsers();
  };

  const isLoading = tenantLoading || usersLoading;
  const error = tenantError || usersError;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Perusahaan", href: "/company" },
          { label: "Tim" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title and actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">Manajemen Tim</h1>
            <p className="text-muted-foreground">
              Kelola anggota tim dan peran mereka
            </p>
          </div>
          {!isLoading && (
            <Button
              onClick={() => setShowInviteDialog(true)}
              className="shrink-0"
            >
              <UserPlus className="mr-2 h-4 w-4" />
              Undang Pengguna
            </Button>
          )}
        </div>

        {/* Content Card */}
        <Card className="shadow-sm">
          <CardContent className="pt-6">
            {isLoading && (
              <div className="py-12">
                <LoadingSpinner size="lg" text="Memuat data tim..." />
              </div>
            )}

            {error && (
              <ErrorDisplay
                error={error}
                title="Gagal memuat data tim"
                onRetry={refetchUsers}
              />
            )}

            {!isLoading && !error && (
              <div className="space-y-6">
                {/* Users table */}
                <div>
                  <div className="mb-4 flex items-center gap-2">
                    <Users className="h-5 w-5" />
                    <h2 className="text-lg font-semibold">Anggota Tim</h2>
                  </div>
                  <UserTable
                    users={users || []}
                    onRoleFilterChange={setRoleFilter}
                    onStatusFilterChange={setStatusFilter}
                    roleFilter={roleFilter}
                    statusFilter={statusFilter}
                    onUpdate={refetchUsers}
                  />
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Invite user dialog */}
      <Dialog open={showInviteDialog} onOpenChange={setShowInviteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Undang Anggota Tim</DialogTitle>
          </DialogHeader>
          <InviteUserForm
            onSuccess={handleInviteSuccess}
            onCancel={() => setShowInviteDialog(false)}
          />
        </DialogContent>
      </Dialog>
    </div>
  );
}
