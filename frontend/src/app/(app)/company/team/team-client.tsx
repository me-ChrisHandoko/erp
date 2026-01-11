/**
 * Team Client Component
 *
 * Client-side interactive component for team management.
 * Receives initial server-fetched data and handles:
 * - Interactive user management (invite, edit, remove)
 * - RTK Query caching for subsequent requests
 * - Company switch handling with explicit refetch
 * - Role and status filtering
 * - Optimistic UI updates
 */

"use client";

import { useState, useEffect } from "react";
import { useSelector } from "react-redux";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { UserPlus, Users, Search } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TenantInfoCard } from "@/components/team/tenant-info-card";
import { UserTable } from "@/components/team/user-table";
import { InviteUserForm } from "@/components/team/invite-user-form";
import { ErrorDisplay } from "@/components/shared/error-display";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useGetTenantQuery } from "@/store/services/tenantApi";
import { useGetCompanyUsersQuery } from "@/store/services/companyUserApi";
import type { UserRole, Tenant, TenantUser } from "@/types/tenant.types";
import type { RootState } from "@/store";

interface TeamClientProps {
  initialData: {
    tenant: Tenant;
    users: TenantUser[];
  };
}

export function TeamClient({ initialData }: TeamClientProps) {
  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<UserRole | undefined>(undefined);
  const [statusFilter, setStatusFilter] = useState<boolean | undefined>(
    undefined
  );
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  // 🔑 Get activeCompanyId from Redux to trigger refetch on company switch
  // This is the key to making switch company work without page reload
  const activeCompanyId = useSelector(
    (state: RootState) => state.company.activeCompany?.id
  );

  // Debounce search input (wait 500ms after user stops typing)
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      setCurrentPage(1); // Reset to page 1 on search
    }, 500);

    return () => clearTimeout(timer);
  }, [search]);

  // Fetch tenant info with RTK Query
  // 🎯 KEY: Skip query until company context is ready
  // When activeCompanyId changes, RTK Query will auto-refetch with new company context
  const {
    data: tenantData,
    isLoading: tenantLoading,
    error: tenantError,
    refetch: refetchTenant,
  } = useGetTenantQuery(undefined, {
    // Skip query until company context is available
    // This ensures we don't fetch with wrong company ID
    skip: !activeCompanyId,
  });

  // Fetch company users with RTK Query
  const {
    data: usersData,
    isLoading: usersLoading,
    error: usersError,
    refetch: refetchUsers,
  } = useGetCompanyUsersQuery(
    {
      role: roleFilter,
      isActive: statusFilter,
    },
    {
      // Skip query until company context is available
      skip: !activeCompanyId,
    }
  );

  // Use initialData as fallback only for first render before query completes
  const displayTenant = tenantData || initialData.tenant;
  const allUsers = usersData || initialData.users;

  // Client-side pagination logic
  // Filter users first
  const filteredUsers = allUsers.filter((user) => {
    // Search filter (name or email)
    if (debouncedSearch) {
      const searchLower = debouncedSearch.toLowerCase();
      const matchesName = user.name.toLowerCase().includes(searchLower);
      const matchesEmail = user.email.toLowerCase().includes(searchLower);
      if (!matchesName && !matchesEmail) return false;
    }

    // Role filter
    if (roleFilter && user.role !== roleFilter) return false;

    // Status filter
    if (statusFilter !== undefined && user.isActive !== statusFilter)
      return false;

    return true;
  });

  // Calculate pagination
  const totalItems = filteredUsers.length;
  const totalPages = Math.ceil(totalItems / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = Math.min(startIndex + pageSize, totalItems);
  const displayUsers = filteredUsers.slice(startIndex, endIndex);

  const pagination = {
    page: currentPage,
    limit: pageSize,
    total: totalItems,
    totalPages: totalPages,
  };

  // 🔑 CRITICAL: Explicit refetch when company changes
  // Cache invalidation alone doesn't trigger refetch for skipped queries
  useEffect(() => {
    if (activeCompanyId) {
      refetchTenant();
      refetchUsers();
      // Reset pagination when company changes
      setCurrentPage(1);
    }
  }, [activeCompanyId, refetchTenant, refetchUsers]);

  // Reset to page 1 when filters change
  useEffect(() => {
    setCurrentPage(1);
  }, [roleFilter, statusFilter, pageSize]);

  const handleInviteSuccess = () => {
    setShowInviteDialog(false);
    refetchUsers();
  };

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
  };

  const handlePageSizeChange = (newPageSize: string) => {
    setPageSize(parseInt(newPageSize));
    setCurrentPage(1); // Reset to page 1 when changing page size
  };

  const isLoading = tenantLoading || usersLoading;
  const error = tenantError || usersError;

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Page title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold tracking-tight">Manajemen Tim</h1>
          <p className="text-muted-foreground">
            Kelola anggota tim dan peran mereka
          </p>
        </div>
        <Button
          onClick={() => setShowInviteDialog(true)}
          disabled={isLoading}
          className="shrink-0"
        >
          <UserPlus className="mr-2 h-4 w-4" />
          Undang Pengguna
        </Button>
      </div>

      {/* Content Card */}
      <Card className="shadow-sm">
        <CardContent>
          {/* Subtle loading indicator for refetching */}
          {isLoading && displayUsers && displayUsers.length > 0 && (
            <div className="text-sm text-muted-foreground text-center py-2 mb-4">
              Memperbarui data...
            </div>
          )}

          {/* Loading State - Only show if no initial data */}
          {isLoading && !displayUsers && (
            <div className="space-y-3">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              error={error}
              title="Gagal memuat data tim"
              onRetry={() => {
                refetchTenant();
                refetchUsers();
              }}
            />
          )}

          {/* Data Display */}
          {!error && displayTenant && displayUsers && (
            <div className="space-y-6">
              {/* Users Section with Pagination */}
              <div>
                <div className="space-y-4">
                  {/* Users Table */}
                  <UserTable
                    users={displayUsers}
                    search={search}
                    onSearchChange={setSearch}
                    onRoleFilterChange={setRoleFilter}
                    onStatusFilterChange={setStatusFilter}
                    roleFilter={roleFilter}
                    statusFilter={statusFilter}
                    onUpdate={refetchUsers}
                  />

                  {/* Pagination */}
                  {pagination.total > 0 && (
                    <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 border-t pt-4">
                      {/* 1. Summary - Record Data */}
                      <div className="text-sm text-muted-foreground text-center sm:text-left">
                        {(() => {
                          const start =
                            (pagination.page - 1) * pagination.limit + 1;
                          const end = Math.min(
                            pagination.page * pagination.limit,
                            pagination.total
                          );
                          return `Menampilkan ${start}-${end} dari ${pagination.total} anggota`;
                        })()}
                      </div>

                      {/* 2. Page Size Selector - Baris per Halaman */}
                      <div className="flex items-center justify-center sm:justify-start gap-2">
                        <span className="text-sm text-muted-foreground whitespace-nowrap">
                          Baris per Halaman
                        </span>
                        <Select
                          value={pageSize.toString()}
                          onValueChange={handlePageSizeChange}
                        >
                          <SelectTrigger className="w-[70px] h-8 bg-background">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="10">10</SelectItem>
                            <SelectItem value="20">20</SelectItem>
                            <SelectItem value="50">50</SelectItem>
                            <SelectItem value="100">100</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>

                      {/* 3. Navigation Buttons - << < Halaman > >> */}
                      <div className="flex items-center justify-center sm:justify-end gap-2">
                        {/* First Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handlePageChange(1)}
                          disabled={pagination.page === 1}
                        >
                          &laquo;
                        </Button>

                        {/* Previous Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handlePageChange(pagination.page - 1)}
                          disabled={pagination.page === 1}
                        >
                          &lsaquo;
                        </Button>

                        {/* Current Page Info */}
                        <span className="text-sm text-muted-foreground px-2">
                          Halaman {pagination.page} dari {pagination.totalPages}
                        </span>

                        {/* Next Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handlePageChange(pagination.page + 1)}
                          disabled={pagination.page >= pagination.totalPages}
                        >
                          &rsaquo;
                        </Button>

                        {/* Last Page */}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            handlePageChange(pagination.totalPages)
                          }
                          disabled={pagination.page >= pagination.totalPages}
                        >
                          &raquo;
                        </Button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

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
