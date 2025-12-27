/**
 * Edit User Role Form Component
 *
 * Form for changing user roles.
 * Includes RBAC protection (cannot change OWNER, cannot remove last ADMIN).
 *
 * Features:
 * - Role selection (ADMIN, STAFF, VIEWER)
 * - RBAC validation
 * - Current role display
 * - Success/error feedback
 */

"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  updateUserRoleSchema,
  type UpdateUserRoleFormData,
} from "@/lib/schemas/user.schema";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { useUpdateCompanyUserRoleMutation } from "@/store/services/companyUserApi";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import type { TenantUser } from "@/types/tenant.types";
import { ShieldCheck } from "lucide-react";

interface EditRoleFormProps {
  user: TenantUser;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function EditRoleForm({ user, onSuccess, onCancel }: EditRoleFormProps) {
  const [updateUserRole, { isLoading }] = useUpdateCompanyUserRoleMutation();

  const {
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<UpdateUserRoleFormData>({
    resolver: zodResolver(updateUserRoleSchema),
    defaultValues: {
      role: user.role === "OWNER" ? "ADMIN" : user.role,
    },
  });

  const onSubmit = async (data: UpdateUserRoleFormData) => {
    try {
      await updateUserRole({
        userId: user.id,
        data,
      }).unwrap();

      toast.success("Role updated successfully", {
        description: `${user.name} is now a ${data.role}`,
      });
      onSuccess?.();
    } catch (error: unknown) {
      const errorData = error && typeof error === 'object' && 'data' in error ? error.data : null;
      const errorMessage =
        (errorData && typeof errorData === 'object' && 'error' in errorData &&
         errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error
          ? (errorData.error.message as string)
          : null) || "Failed to update role";

      if (errorMessage.includes("OWNER")) {
        toast.error("Cannot modify OWNER role", {
          description: "OWNER role is permanent and cannot be changed",
        });
      } else if (errorMessage.includes("last ADMIN")) {
        toast.error("Cannot change last ADMIN", {
          description: "At least one ADMIN is required in the organization",
        });
      } else if (errorMessage.includes("permission denied")) {
        toast.error("Permission denied", {
          description: "You don't have permission to modify this user's role",
        });
      } else {
        toast.error("Failed to update role", {
          description: errorMessage,
        });
      }
    }
  };

  // Prevent editing OWNER role
  if (user.role === "OWNER") {
    return (
      <div className="space-y-4">
        <div className="rounded-md bg-yellow-50 dark:bg-yellow-950 p-4 border border-yellow-200 dark:border-yellow-800">
          <div className="flex items-center gap-2 mb-2">
            <ShieldCheck className="h-5 w-5 text-yellow-600" />
            <p className="font-semibold text-yellow-800 dark:text-yellow-200">
              OWNER Role Protection
            </p>
          </div>
          <p className="text-sm text-yellow-700 dark:text-yellow-300">
            The OWNER role is permanent and cannot be changed. This ensures that the
            organization always has an owner with full administrative privileges.
          </p>
        </div>
        <div className="flex justify-end">
          <Button onClick={onCancel}>Close</Button>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* User Info */}
      <div className="rounded-md bg-muted p-3 space-y-1">
        <p className="text-sm font-medium">{user.name}</p>
        <p className="text-sm text-muted-foreground">{user.email}</p>
        <div className="flex items-center gap-2">
          <span className="text-sm">Current Role:</span>
          <Badge>{user.role}</Badge>
        </div>
      </div>

      {/* New Role Selection */}
      <div className="space-y-2">
        <Label htmlFor="role">
          New Role <span className="text-red-500">*</span>
        </Label>
        <Select
          value={watch("role")}
          onValueChange={(value) => setValue("role", value as "ADMIN" | "STAFF" | "VIEWER")}
          disabled={isLoading}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ADMIN">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Admin</span>
                <span className="text-xs text-muted-foreground">
                  Full access to all features and settings
                </span>
              </div>
            </SelectItem>
            <SelectItem value="STAFF">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Staff</span>
                <span className="text-xs text-muted-foreground">
                  Can manage daily operations and transactions
                </span>
              </div>
            </SelectItem>
            <SelectItem value="VIEWER">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Viewer</span>
                <span className="text-xs text-muted-foreground">
                  Read-only access to view data
                </span>
              </div>
            </SelectItem>
          </SelectContent>
        </Select>
        {errors.role && (
          <p className="text-sm text-red-500">{errors.role.message}</p>
        )}
      </div>

      {/* Warning for changing ADMIN role */}
      {user.role === "ADMIN" && watch("role") !== "ADMIN" && (
        <div className="rounded-md bg-yellow-50 dark:bg-yellow-950 p-3 border border-yellow-200 dark:border-yellow-800">
          <p className="text-sm text-yellow-800 dark:text-yellow-200">
            ⚠️ You are changing an ADMIN role. Make sure there is at least one other
            ADMIN in your organization.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
            Cancel
          </Button>
        )}
        <Button type="submit" disabled={isLoading || watch("role") === user.role}>
          {isLoading ? (
            <>
              <LoadingSpinner size="sm" className="mr-2" />
              Updating Role...
            </>
          ) : (
            "Update Role"
          )}
        </Button>
      </div>
    </form>
  );
}
