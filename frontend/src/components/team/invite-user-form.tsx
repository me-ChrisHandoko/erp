/**
 * Invite User Form Component
 *
 * Form for inviting new team members.
 * Sends email invitation with temporary password.
 *
 * Features:
 * - Email, name, and role input
 * - Form validation (email format)
 * - Role selection (ADMIN, STAFF, VIEWER only - no OWNER)
 * - Rate limiting handling (429 error)
 * - Success/error feedback
 */

"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { inviteUserSchema, type InviteUserFormData } from "@/lib/schemas/user.schema";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { useInviteCompanyUserMutation } from "@/store/services/companyUserApi";
import { LoadingSpinner } from "@/components/shared/loading-spinner";

interface InviteUserFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function InviteUserForm({ onSuccess, onCancel }: InviteUserFormProps) {
  const [inviteUser, { isLoading }] = useInviteCompanyUserMutation();

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<InviteUserFormData>({
    resolver: zodResolver(inviteUserSchema),
    defaultValues: {
      role: "STAFF",
    },
  });

  const onSubmit = async (data: InviteUserFormData) => {
    try {
      await inviteUser(data).unwrap();
      toast.success("Undangan berhasil dikirim", {
        description: `Email telah dikirim ke ${data.email}`,
      });
      onSuccess?.();
    } catch (error: unknown) {
      const errorData = error && typeof error === 'object' && 'data' in error ? error.data : null;
      const errorMessage =
        (errorData && typeof errorData === 'object' && 'error' in errorData &&
         errorData.error && typeof errorData.error === 'object' && 'message' in errorData.error
          ? (errorData.error.message as string)
          : null) || "Gagal mengirim undangan";

      // Handle rate limiting
      const errorStatus = error && typeof error === 'object' && 'status' in error ? error.status : null;
      if (errorStatus === 429) {
        toast.error("Batas pengiriman terlampaui", {
          description: "Harap tunggu sebelum mengirim undangan lagi (maksimal 5 per menit)",
        });
      } else if (errorMessage.includes("user limit")) {
        toast.error("Batas pengguna tercapai", {
          description: "Tingkatkan langganan Anda untuk menambah lebih banyak pengguna",
        });
      } else if (errorMessage.includes("already exists")) {
        toast.error("Pengguna sudah ada", {
          description: "Email ini sudah terdaftar di organisasi Anda",
        });
      } else {
        toast.error("Gagal mengirim undangan", {
          description: errorMessage,
        });
      }
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {/* Email */}
      <div className="space-y-2">
        <Label htmlFor="email">
          Email <span className="text-red-500">*</span>
        </Label>
        <Input
          id="email"
          type="email"
          {...register("email")}
          placeholder="pengguna@example.com"
          disabled={isLoading}
          className="bg-background"
        />
        {errors.email && (
          <p className="text-sm text-red-500">{errors.email.message}</p>
        )}
      </div>

      {/* Name */}
      <div className="space-y-2">
        <Label htmlFor="name">
          Nama Lengkap <span className="text-red-500">*</span>
        </Label>
        <Input
          id="name"
          {...register("name")}
          placeholder="Nama Lengkap"
          disabled={isLoading}
          className="bg-background"
        />
        {errors.name && (
          <p className="text-sm text-red-500">{errors.name.message}</p>
        )}
      </div>

      {/* Role */}
      <div className="space-y-2">
        <Label htmlFor="role">
          Role <span className="text-red-500">*</span>
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
        <p className="text-xs text-muted-foreground">
          Catatan: Role OWNER tidak dapat diberikan melalui undangan
        </p>
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-2 pt-4">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
            Batal
          </Button>
        )}
        <Button type="submit" disabled={isLoading}>
          {isLoading ? (
            <>
              <LoadingSpinner size="sm" className="mr-2" />
              Mengirim Undangan...
            </>
          ) : (
            "Kirim Undangan"
          )}
        </Button>
      </div>
    </form>
  );
}
