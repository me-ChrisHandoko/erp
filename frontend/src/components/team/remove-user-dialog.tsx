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
import { useRemoveCompanyUserMutation } from "@/store/services/companyUserApi";
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
  const [removeUser, { isLoading }] = useRemoveCompanyUserMutation();

  const handleRemove = async () => {
    try {
      await removeUser(user.id).unwrap();
      toast.success("User berhasil dihapus", {
        description: `${user.name} telah dihapus dari organisasi Anda`,
      });
      onSuccess?.();
      onOpenChange(false);
    } catch (error: unknown) {
      const errorData = error && typeof error === 'object' && 'data' in error ? error.data : null;
      const errorMessage =
        (errorData && typeof errorData === 'object' && 'error' in errorData &&
         errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error
          ? (errorData.error.message as string)
          : null) || "Gagal menghapus user";

      if (errorMessage.includes("OWNER")) {
        toast.error("Tidak Dapat Menghapus OWNER", {
          description: "Role OWNER tidak dapat dihapus dari organisasi",
        });
      } else if (errorMessage.includes("last ADMIN")) {
        toast.error("Tidak Dapat Menghapus ADMIN Terakhir", {
          description: "Minimal harus ada satu ADMIN di organisasi",
        });
      } else if (errorMessage.includes("permission denied")) {
        toast.error("Akses Ditolak", {
          description: "Anda tidak memiliki izin untuk menghapus user ini",
        });
      } else {
        toast.error("Gagal menghapus user", {
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
              Tidak Dapat Menghapus OWNER
            </AlertDialogTitle>
            <AlertDialogDescription>
              Role OWNER bersifat permanen dan tidak dapat dihapus dari organisasi.
              Ini memastikan bahwa organisasi selalu memiliki pemilik dengan hak akses
              administratif penuh.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Tutup</AlertDialogCancel>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    );
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Hapus User</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div className="space-y-3">
              <p>
                Apakah Anda yakin ingin menghapus user ini dari organisasi Anda?
                Tindakan ini tidak dapat dibatalkan.
              </p>

              {/* User Info */}
              <div className="rounded-md bg-muted p-3 space-y-1">
                <p className="font-medium">{user.name}</p>
                <p className="text-sm text-muted-foreground">{user.email}</p>
                <div className="flex items-center gap-2">
                  <Badge>{user.role}</Badge>
                  <Badge variant={user.isActive ? "default" : "outline"}>
                    {user.isActive ? "Aktif" : "Tidak Aktif"}
                  </Badge>
                </div>
              </div>

              {/* Warning for ADMIN role */}
              {user.role === "ADMIN" && (
                <div className="rounded-md bg-yellow-50 dark:bg-yellow-950 p-3 border border-yellow-200 dark:border-yellow-800">
                  <p className="text-sm text-yellow-800 dark:text-yellow-200">
                    ⚠️ Anda akan menghapus ADMIN. Pastikan masih ada minimal satu
                    ADMIN lain di organisasi Anda.
                  </p>
                </div>
              )}

              <p className="text-sm">
                User akan langsung kehilangan akses ke sistem dan semua data terkait.
              </p>
            </div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Batal</AlertDialogCancel>
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
                Menghapus...
              </>
            ) : (
              "Hapus User"
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
