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

      toast.success("Role berhasil diperbarui", {
        description: `${user.name} sekarang memiliki role ${data.role}`,
      });
      onSuccess?.();
    } catch (error: unknown) {
      const errorData = error && typeof error === 'object' && 'data' in error ? error.data : null;
      const errorMessage =
        (errorData && typeof errorData === 'object' && 'error' in errorData &&
         errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error
          ? (errorData.error.message as string)
          : null) || "Gagal memperbarui role";

      if (errorMessage.includes("OWNER")) {
        toast.error("Tidak Dapat Mengubah Role OWNER", {
          description: "Role OWNER bersifat permanen dan tidak dapat diubah",
        });
      } else if (errorMessage.includes("last ADMIN")) {
        toast.error("Tidak Dapat Mengubah ADMIN Terakhir", {
          description: "Minimal harus ada satu ADMIN di organisasi",
        });
      } else if (errorMessage.includes("permission denied")) {
        toast.error("Akses Ditolak", {
          description: "Anda tidak memiliki izin untuk mengubah role user ini",
        });
      } else {
        toast.error("Gagal memperbarui role", {
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
              Proteksi Role OWNER
            </p>
          </div>
          <p className="text-sm text-yellow-700 dark:text-yellow-300">
            Role OWNER bersifat permanen dan tidak dapat diubah. Ini memastikan bahwa
            organisasi selalu memiliki pemilik dengan hak akses administratif penuh.
          </p>
        </div>
        <div className="flex justify-end">
          <Button onClick={onCancel}>Tutup</Button>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* User Info */}
      <div className="rounded-md bg-muted p-3 space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <div>
            <p className="text-xs text-muted-foreground mb-1">Nama</p>
            <p className="text-sm font-medium">{user.name}</p>
          </div>
          <div>
            <p className="text-xs text-muted-foreground mb-1">Email</p>
            <p className="text-sm font-medium">{user.email}</p>
          </div>
        </div>
        <div>
          <p className="text-xs text-muted-foreground mb-1">Role Saat Ini</p>
          <Badge>{user.role}</Badge>
        </div>
      </div>

      {/* New Role Selection */}
      <div className="space-y-2">
        <Label htmlFor="role">
          Role Baru <span className="text-red-500">*</span>
        </Label>
        <Select
          value={watch("role")}
          onValueChange={(value) => setValue("role", value as "ADMIN" | "FINANCE" | "SALES" | "WAREHOUSE" | "STAFF")}
          disabled={isLoading}
        >
          <SelectTrigger className="w-full bg-background">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ADMIN">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Admin</span>
                <span className="text-xs text-muted-foreground">
                  Akses penuh ke semua fitur dan pengaturan
                </span>
              </div>
            </SelectItem>
            <SelectItem value="FINANCE">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Keuangan</span>
                <span className="text-xs text-muted-foreground">
                  Dapat mengelola transaksi keuangan dan laporan
                </span>
              </div>
            </SelectItem>
            <SelectItem value="SALES">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Penjualan</span>
                <span className="text-xs text-muted-foreground">
                  Dapat mengelola pesanan penjualan dan relasi pelanggan
                </span>
              </div>
            </SelectItem>
            <SelectItem value="WAREHOUSE">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Gudang</span>
                <span className="text-xs text-muted-foreground">
                  Dapat mengelola gudang dan operasi inventori
                </span>
              </div>
            </SelectItem>
            <SelectItem value="STAFF">
              <div className="flex flex-col items-start">
                <span className="font-semibold">Staf</span>
                <span className="text-xs text-muted-foreground">
                  Dapat mengelola operasi harian dan transaksi
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
            ⚠️ Anda akan mengubah role ADMIN. Pastikan masih ada minimal satu
            ADMIN lain di organisasi Anda.
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
            Batal
          </Button>
        )}
        <Button type="submit" disabled={isLoading || watch("role") === user.role}>
          {isLoading ? (
            <>
              <LoadingSpinner size="sm" className="mr-2" />
              Memperbarui Role...
            </>
          ) : (
            "Perbarui Role"
          )}
        </Button>
      </div>
    </form>
  );
}
