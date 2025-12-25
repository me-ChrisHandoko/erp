/**
 * Bank Account Form Component
 *
 * Modal form for creating and editing bank accounts.
 * Supports bank selection, account validation, and primary bank toggle.
 */

"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  addBankAccountSchema,
  updateBankAccountSchema,
  type AddBankAccountFormData,
  type UpdateBankAccountFormData,
} from "@/lib/schemas/company.schema";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useAddBankAccountMutation,
  useUpdateBankAccountMutation,
} from "@/store/services/companyApi";
import { toast } from "sonner";
import { INDONESIAN_BANKS, type BankAccountResponse } from "@/types/company.types";

interface BankAccountFormProps {
  defaultValues?: BankAccountResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function BankAccountForm({
  defaultValues,
  onSuccess,
  onCancel,
}: BankAccountFormProps) {
  const isEditMode = !!defaultValues;
  const [addBank, { isLoading: isAdding }] = useAddBankAccountMutation();
  const [updateBank, { isLoading: isUpdating }] = useUpdateBankAccountMutation();

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<AddBankAccountFormData | UpdateBankAccountFormData>({
    resolver: zodResolver(isEditMode ? updateBankAccountSchema : addBankAccountSchema),
    defaultValues: defaultValues || {
      bankName: "",
      accountNumber: "",
      accountName: "",
      branchName: "",
      isPrimary: false,
      checkPrefix: "",
    },
  });

  const bankName = watch("bankName");
  const isPrimary = watch("isPrimary");

  const onSubmit = async (data: AddBankAccountFormData | UpdateBankAccountFormData) => {
    try {
      if (isEditMode && defaultValues) {
        await updateBank({
          id: defaultValues.id,
          data: data as UpdateBankAccountFormData,
        }).unwrap();
        toast.success("Rekening bank berhasil diperbarui");
      } else {
        await addBank(data as AddBankAccountFormData).unwrap();
        toast.success("Rekening bank berhasil ditambahkan");
      }
      onSuccess?.();
    } catch (error: unknown) {
      const errorMessage =
        (error as { data?: { error?: { message?: string } }; message?: string })?.data?.error?.message ||
        (error as { message?: string })?.message ||
        "Gagal menyimpan rekening bank";
      toast.error(errorMessage);
    }
  };

  const isLoading = isSubmitting || isAdding || isUpdating;

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Bank Name Selection */}
      <div className="space-y-2">
        <Label htmlFor="bankName">
          Nama Bank <span className="text-red-500">*</span>
        </Label>
        <Select
          value={bankName}
          onValueChange={(value) => setValue("bankName", value)}
          disabled={isLoading}
        >
          <SelectTrigger id="bankName" className="w-full">
            <SelectValue placeholder="Pilih bank" />
          </SelectTrigger>
          <SelectContent>
            {INDONESIAN_BANKS.map((bank) => (
              <SelectItem key={bank} value={bank}>
                {bank}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {errors.bankName && (
          <p className="text-sm text-red-500">{errors.bankName.message}</p>
        )}
      </div>

      {/* Account Number */}
      <div className="space-y-2">
        <Label htmlFor="accountNumber">
          Nomor Rekening <span className="text-red-500">*</span>
        </Label>
        <Input
          id="accountNumber"
          {...register("accountNumber")}
          placeholder="1234567890"
          className="font-mono"
          disabled={isLoading}
        />
        {errors.accountNumber && (
          <p className="text-sm text-red-500">{errors.accountNumber.message}</p>
        )}
        <p className="text-xs text-muted-foreground">
          Masukkan nomor rekening tanpa spasi atau tanda hubung
        </p>
      </div>

      {/* Account Name */}
      <div className="space-y-2">
        <Label htmlFor="accountName">
          Nama Pemilik Rekening <span className="text-red-500">*</span>
        </Label>
        <Input
          id="accountName"
          {...register("accountName")}
          placeholder="PT Example Indonesia"
          disabled={isLoading}
        />
        {errors.accountName && (
          <p className="text-sm text-red-500">{errors.accountName.message}</p>
        )}
        <p className="text-xs text-muted-foreground">
          Nama harus sesuai dengan yang tertera di buku rekening
        </p>
      </div>

      {/* Branch Name (Optional) */}
      <div className="space-y-2">
        <Label htmlFor="branchName">Nama Cabang (Opsional)</Label>
        <Input
          id="branchName"
          {...register("branchName")}
          placeholder="Cabang Sudirman"
          disabled={isLoading}
        />
        {errors.branchName && (
          <p className="text-sm text-red-500">{errors.branchName.message}</p>
        )}
      </div>

      {/* Check Prefix (Optional) */}
      <div className="space-y-2">
        <Label htmlFor="checkPrefix">Prefix Cek (Opsional)</Label>
        <Input
          id="checkPrefix"
          {...register("checkPrefix")}
          placeholder="CHK"
          disabled={isLoading}
          maxLength={20}
        />
        {errors.checkPrefix && (
          <p className="text-sm text-red-500">{errors.checkPrefix.message}</p>
        )}
        <p className="text-xs text-muted-foreground">
          Prefix untuk nomor cek (contoh: CHK untuk CHK-001)
        </p>
      </div>

      {/* Primary Bank Toggle */}
      <div className="flex items-start space-x-3 rounded-md border p-4">
        <Checkbox
          id="isPrimary"
          checked={isPrimary}
          onCheckedChange={(checked) => setValue("isPrimary", !!checked)}
          disabled={isLoading}
        />
        <div className="space-y-1 leading-none">
          <Label
            htmlFor="isPrimary"
            className="cursor-pointer font-medium text-sm"
          >
            Jadikan Rekening Utama
          </Label>
          <p className="text-xs text-muted-foreground">
            Rekening utama akan digunakan sebagai default untuk transaksi dan
            invoice. Hanya satu rekening yang bisa menjadi rekening utama.
          </p>
        </div>
      </div>

      {/* Form Actions */}
      <div className="flex justify-end space-x-3 pt-4 border-t">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
            Batal
          </Button>
        )}
        <Button type="submit" disabled={isLoading}>
          {isLoading
            ? isEditMode
              ? "Memperbarui..."
              : "Menambahkan..."
            : isEditMode
            ? "Perbarui Rekening"
            : "Tambah Rekening"}
        </Button>
      </div>
    </form>
  );
}
