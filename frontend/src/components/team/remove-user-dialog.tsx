/**
 * Remove User Dialog Component
 *
 * Confirmation dialog for removing users from the organization.
 * Includes RBAC protection (cannot remove OWNER or last ADMIN).
 *
 * Features:
 * - User information display
 * - Confirmation required
 * - RBAC validation
 * - Success/error feedback
 */

"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { useRemoveUserMutation } from "@/store/services/tenantApi";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import type { TenantUser } from "@/types/tenant.types";
import { ShieldCheck } from "lucide-react";

interface RemoveUserDialogProps {
  user: TenantUser;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function RemoveUserDialog({
  user,
  open,
  onOpenChange,
  onSuccess,
}: RemoveUserDialogProps) {
  const [removeUser, { isLoading }] = useRemoveUserMutation();

  const handleRemove = async () => {
    try {
      await removeUser(user.id).unwrap();
      toast.success("User removed successfully", {
        description: `${user.name} has been removed from your organization`,
      });
      onSuccess?.();
      onOpenChange(false);
    } catch (error: unknown) {
      const errorData = error && typeof error === 'object' && 'data' in error ? error.data : null;
      const errorMessage =
        (errorData && typeof errorData === 'object' && 'error' in errorData &&
         errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error
          ? (errorData.error.message as string)
          : null) || "Failed to remove user";

      if (errorMessage.includes("OWNER")) {
        toast.error("Cannot remove OWNER", {
          description: "The OWNER role cannot be removed from the organization",
        });
      } else if (errorMessage.includes("last ADMIN")) {
        toast.error("Cannot remove last ADMIN", {
          description: "At least one ADMIN is required in the organization",
        });
      } else if (errorMessage.includes("permission denied")) {
        toast.error("Permission denied", {
          description: "You don't have permission to remove this user",
        });
      } else {
        toast.error("Failed to remove user", {
          description: errorMessage,
        });
      }
    }
  };

  // Prevent removing OWNER
  if (user.role === "OWNER") {
    return (
      <AlertDialog open={open} onOpenChange={onOpenChange}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <ShieldCheck className="h-5 w-5 text-yellow-600" />
              Cannot Remove OWNER
            </AlertDialogTitle>
            <AlertDialogDescription>
              The OWNER role is permanent and cannot be removed from the organization.
              This ensures that the organization always has an owner with full
              administrative privileges.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Close</AlertDialogCancel>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    );
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove User</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Are you sure you want to remove this user from your organization? This
                action cannot be undone.
              </p>

              {/* User Info */}
              <div className="rounded-md bg-muted p-3 space-y-1">
                <p className="font-medium">{user.name}</p>
                <p className="text-sm text-muted-foreground">{user.email}</p>
                <div className="flex items-center gap-2">
                  <Badge>{user.role}</Badge>
                  <Badge variant={user.isActive ? "default" : "outline"}>
                    {user.isActive ? "Active" : "Inactive"}
                  </Badge>
                </div>
              </div>

              {/* Warning for ADMIN role */}
              {user.role === "ADMIN" && (
                <div className="rounded-md bg-yellow-50 dark:bg-yellow-950 p-3 border border-yellow-200 dark:border-yellow-800">
                  <p className="text-sm text-yellow-800 dark:text-yellow-200">
                    ⚠️ You are removing an ADMIN. Make sure there is at least one other
                    ADMIN in your organization.
                  </p>
                </div>
              )}

              <p className="text-sm">
                The user will immediately lose access to the system and all associated
                data.
              </p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              e.preventDefault();
              handleRemove();
            }}
            disabled={isLoading}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
          >
            {isLoading ? (
              <>
                <LoadingSpinner size="sm" className="mr-2" />
                Removing...
              </>
            ) : (
              "Remove User"
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
