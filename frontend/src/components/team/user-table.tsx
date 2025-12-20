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
import { Pencil, Trash2, Shield, ShieldCheck } from "lucide-react";
import { format } from "date-fns";
import { EmptyState } from "@/components/shared/empty-state";
import { EditRoleForm } from "./edit-role-form";
import { RemoveUserDialog } from "./remove-user-dialog";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import type { TenantUser, UserRole } from "@/types/tenant.types";

interface UserTableProps {
  users: TenantUser[];
  roleFilter?: UserRole;
  statusFilter?: boolean;
  onRoleFilterChange: (role: UserRole | undefined) => void;
  onStatusFilterChange: (status: boolean | undefined) => void;
  onUpdate: () => void;
}

export function UserTable({
  users,
  roleFilter,
  statusFilter,
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
      case "STAFF":
        return "secondary";
      case "VIEWER":
        return "outline";
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
      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4 mb-4">
        <div className="flex items-center gap-2">
          <label className="text-sm font-medium">Role:</label>
          <Select
            value={roleFilter || "all"}
            onValueChange={(value) =>
              onRoleFilterChange(value === "all" ? undefined : (value as UserRole))
            }
          >
            <SelectTrigger className="w-[150px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Roles</SelectItem>
              <SelectItem value="OWNER">Owner</SelectItem>
              <SelectItem value="ADMIN">Admin</SelectItem>
              <SelectItem value="STAFF">Staff</SelectItem>
              <SelectItem value="VIEWER">Viewer</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm font-medium">Status:</label>
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
            <SelectTrigger className="w-[150px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Users</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="inactive">Inactive</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {(roleFilter || statusFilter !== undefined) && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              onRoleFilterChange(undefined);
              onStatusFilterChange(undefined);
            }}
          >
            Clear Filters
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
              <TableHead>Phone</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Last Login</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7}>
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
                  <TableCell>{user.phone || "-"}</TableCell>
                  <TableCell>
                    <Badge variant={getRoleBadgeVariant(user.role)}>
                      {getRoleIcon(user.role)}
                      {user.role}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={user.isActive ? "default" : "outline"}>
                      {user.isActive ? "Active" : "Inactive"}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {user.lastLoginAt
                      ? format(new Date(user.lastLoginAt), "MMM dd, yyyy HH:mm")
                      : "Never"}
                  </TableCell>
                  <TableCell className="text-right space-x-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setEditingUser(user)}
                      disabled={user.role === "OWNER"}
                      title={
                        user.role === "OWNER"
                          ? "Cannot edit OWNER role"
                          : "Edit user role"
                      }
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setRemovingUser(user)}
                      disabled={user.role === "OWNER"}
                      title={
                        user.role === "OWNER"
                          ? "Cannot remove OWNER"
                          : "Remove user"
                      }
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
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
