/**
 * User Table Component
 *
 * Displays team members with role and status information.
 * Provides filtering by role and active status.
 * Allows editing roles and removing users (with RBAC protection).
 */

"use client";

import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Pencil, Trash2, Shield, ShieldCheck, MoreHorizontal, Search } from "lucide-react";
import { format } from "date-fns";
import { EmptyState } from "@/components/shared/empty-state";
import { EditRoleForm } from "./edit-role-form";
import { RemoveUserDialog } from "./remove-user-dialog";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import type { TenantUser, UserRole } from "@/types/tenant.types";

interface UserTableProps {
  users: TenantUser[];
  search?: string;
  roleFilter?: UserRole;
  statusFilter?: boolean;
  onSearchChange: (search: string) => void;
  onRoleFilterChange: (role: UserRole | undefined) => void;
  onStatusFilterChange: (status: boolean | undefined) => void;
  onUpdate: () => void;
}

export function UserTable({
  users,
  search,
  roleFilter,
  statusFilter,
  onSearchChange,
  onRoleFilterChange,
  onStatusFilterChange,
  onUpdate,
}: UserTableProps) {
  const [editingUser, setEditingUser] = useState<TenantUser | null>(null);
  const [removingUser, setRemovingUser] = useState<TenantUser | null>(null);

  // Role badge styling
  const getRoleBadgeVariant = (
    role: UserRole
  ): "default" | "secondary" | "destructive" | "outline" => {
    switch (role) {
      case "OWNER":
        return "destructive";
      case "ADMIN":
        return "default";
      case "FINANCE":
        return "secondary";
      case "SALES":
        return "secondary";
      case "WAREHOUSE":
        return "secondary";
      case "STAFF":
        return "secondary";
      default:
        return "outline";
    }
  };

  // Role icon
  const getRoleIcon = (role: UserRole) => {
    if (role === "OWNER" || role === "ADMIN") {
      return <ShieldCheck className="h-3 w-3 mr-1" />;
    }
    return <Shield className="h-3 w-3 mr-1" />;
  };

  const handleEditSuccess = () => {
    setEditingUser(null);
    onUpdate();
  };

  const handleRemoveSuccess = () => {
    setRemovingUser(null);
    onUpdate();
  };

  return (
    <>
      {/* Search and Filters */}
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
        {/* Search */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Cari nama atau email..."
            value={search || ""}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9 bg-background"
          />
        </div>

        {/* Role Filter */}
        <Select
          value={roleFilter || "all"}
          onValueChange={(value) =>
            onRoleFilterChange(value === "all" ? undefined : (value as UserRole))
          }
        >
          <SelectTrigger className="w-full sm:w-[150px] bg-background">
            <SelectValue placeholder="Semua Role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Role</SelectItem>
            <SelectItem value="OWNER">Owner</SelectItem>
            <SelectItem value="ADMIN">Admin</SelectItem>
            <SelectItem value="FINANCE">Keuangan</SelectItem>
            <SelectItem value="SALES">Penjualan</SelectItem>
            <SelectItem value="WAREHOUSE">Gudang</SelectItem>
            <SelectItem value="STAFF">Staf</SelectItem>
          </SelectContent>
        </Select>

        {/* Status Filter */}
        <Select
          value={
            statusFilter === undefined ? "all" : statusFilter ? "active" : "inactive"
          }
          onValueChange={(value) =>
            onStatusFilterChange(
              value === "all" ? undefined : value === "active"
            )
          }
        >
          <SelectTrigger className="w-full sm:w-[180px] bg-background">
            <SelectValue placeholder="Semua Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Status</SelectItem>
            <SelectItem value="active">Aktif</SelectItem>
            <SelectItem value="inactive">Tidak Aktif</SelectItem>
          </SelectContent>
        </Select>

        {/* Clear Filters */}
        {(search || roleFilter || statusFilter !== undefined) && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              onSearchChange("");
              onRoleFilterChange(undefined);
              onStatusFilterChange(undefined);
            }}
          >
            Hapus Filter
          </Button>
        )}
      </div>

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Last Login</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6}>
                  <EmptyState
                    icon={Shield}
                    title="No users found"
                    description={
                      roleFilter || statusFilter !== undefined
                        ? "Try adjusting your filters"
                        : "Start by inviting your first team member"
                    }
                  />
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="font-medium">{user.name}</TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <Badge variant={getRoleBadgeVariant(user.role)}>
                      {getRoleIcon(user.role)}
                      {user.role}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={user.isActive ? "default" : "outline"}>
                      {user.isActive ? "Aktif" : "Tidak Aktif"}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {user.lastLoginAt
                      ? format(new Date(user.lastLoginAt), "MMM dd, yyyy HH:mm")
                      : "Belum Pernah"}
                  </TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                        <DropdownMenuSeparator />

                        {/* Edit Role - Only if not OWNER */}
                        {user.role !== "OWNER" && (
                          <DropdownMenuItem
                            onClick={() => setEditingUser(user)}
                            className="cursor-pointer"
                          >
                            <Pencil className="mr-2 h-4 w-4" />
                            Edit Role
                          </DropdownMenuItem>
                        )}

                        {/* Remove User - Only if not OWNER */}
                        {user.role !== "OWNER" && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() => setRemovingUser(user)}
                              className="cursor-pointer text-red-600 focus:text-red-600"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Hapus User
                            </DropdownMenuItem>
                          </>
                        )}

                        {/* Show message for OWNER */}
                        {user.role === "OWNER" && (
                          <DropdownMenuItem disabled className="text-muted-foreground">
                            <Shield className="mr-2 h-4 w-4" />
                            Protected Role
                          </DropdownMenuItem>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Edit Role Dialog */}
      <Dialog open={!!editingUser} onOpenChange={(open) => !open && setEditingUser(null)}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Edit User Role</DialogTitle>
          </DialogHeader>
          {editingUser && (
            <EditRoleForm
              user={editingUser}
              onSuccess={handleEditSuccess}
              onCancel={() => setEditingUser(null)}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* Remove User Dialog */}
      {removingUser && (
        <RemoveUserDialog
          user={removingUser}
          open={!!removingUser}
          onOpenChange={(open) => !open && setRemovingUser(null)}
          onSuccess={handleRemoveSuccess}
        />
      )}
    </>
  );
}
