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
import {
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { UserPlus, Users } from "lucide-react";
import { TenantInfoCard } from "@/components/team/tenant-info-card";
import { UserTable } from "@/components/team/user-table";
import { InviteUserForm } from "@/components/team/invite-user-form";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useGetTenantQuery, useGetUsersQuery } from "@/store/services/tenantApi";
import type { UserRole } from "@/types/tenant.types";

export default function TeamPage() {
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [roleFilter, setRoleFilter] = useState<UserRole | undefined>(undefined);
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(undefined);

  // Fetch tenant info
  const {
    data: tenant,
    isLoading: tenantLoading,
    error: tenantError,
  } = useGetTenantQuery();

  // Fetch users with filters
  const {
    data: users,
    isLoading: usersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useGetUsersQuery({
    role: roleFilter,
    isActive: statusFilter,
  });

  const handleInviteSuccess = () => {
    setShowInviteDialog(false);
    refetchUsers();
  };

  // Loading state
  if (tenantLoading || usersLoading) {
    return (
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Team Management</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
          <LoadingSpinner size="lg" />
        </div>
      </SidebarInset>
    );
  }

  // Error state
  if (tenantError || usersError) {
    return (
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Team Management</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <ErrorDisplay
            title="Failed to load team data"
            error={tenantError || usersError || "Failed to load team data"}
          />
        </div>
      </SidebarInset>
    );
  }

  return (
    <SidebarInset>
      {/* Header with breadcrumb */}
      <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="mr-2 h-4" />
        <Breadcrumb>
          <BreadcrumbList>
            <BreadcrumbItem>
              <BreadcrumbPage>Team Management</BreadcrumbPage>
            </BreadcrumbItem>
          </BreadcrumbList>
        </Breadcrumb>
      </header>

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 md:gap-6 md:p-6">
        {/* Page title and actions */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Team Management</h1>
            <p className="text-muted-foreground mt-1">
              Manage your team members and their roles
            </p>
          </div>
          <Button onClick={() => setShowInviteDialog(true)}>
            <UserPlus className="mr-2 h-4 w-4" />
            Invite User
          </Button>
        </div>

        {/* Tenant info card */}
        {tenant && <TenantInfoCard tenant={tenant} />}

        {/* Users table card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              Team Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            <UserTable
              users={users || []}
              onRoleFilterChange={setRoleFilter}
              onStatusFilterChange={setStatusFilter}
              roleFilter={roleFilter}
              statusFilter={statusFilter}
              onUpdate={refetchUsers}
            />
          </CardContent>
        </Card>
      </div>

      {/* Invite user dialog */}
      <Dialog open={showInviteDialog} onOpenChange={setShowInviteDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Invite Team Member</DialogTitle>
          </DialogHeader>
          <InviteUserForm
            onSuccess={handleInviteSuccess}
            onCancel={() => setShowInviteDialog(false)}
          />
        </DialogContent>
      </Dialog>
    </SidebarInset>
  );
}
