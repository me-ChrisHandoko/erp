/**
 * Bank Account Table Component
 *
 * Displays list of bank accounts with edit/delete functionality.
 * Handles primary bank logic and minimum 1 bank validation.
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
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Star, Pencil, Trash2, MoreHorizontal } from "lucide-react";
import { useDeleteBankAccountMutation } from "@/store/services/companyApi";
import { toast } from "sonner";
import type { BankAccountResponse } from "@/types/company.types";
import { BankAccountForm } from "./bank-account-form";

interface BankAccountTableProps {
  banks: BankAccountResponse[];
}

export function BankAccountTable({ banks }: BankAccountTableProps) {
  const [editingBank, setEditingBank] = useState<BankAccountResponse | null>(null);
  const [deletingBank, setDeletingBank] = useState<BankAccountResponse | null>(null);
  const [deleteBank, { isLoading: isDeleting }] = useDeleteBankAccountMutation();

  const handleDelete = async () => {
    if (!deletingBank) return;

    try {
      await deleteBank(deletingBank.id).unwrap();
      toast.success("Rekening bank berhasil dihapus");
      setDeletingBank(null);
    } catch (error: unknown) {
      const errorMessage =
        (error as { data?: { error?: { message?: string } }; message?: string })?.data?.error?.message ||
        (error as { message?: string })?.message ||
        "Gagal menghapus rekening bank";

      // Check for minimum 1 bank validation error
      if (errorMessage.includes("minimum") || errorMessage.includes("at least")) {
        toast.error("Minimal harus ada 1 rekening bank");
      } else {
        toast.error(errorMessage);
      }
    }
  };

  return (
    <>
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">Utama</TableHead>
              <TableHead>Nama Bank</TableHead>
              <TableHead>Nomor Rekening</TableHead>
              <TableHead>Nama Pemilik</TableHead>
              <TableHead>Cabang</TableHead>
              <TableHead>Prefix Cek</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Aksi</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {banks.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center h-32 text-muted-foreground">
                  Belum ada rekening bank. Tambahkan rekening bank pertama Anda.
                </TableCell>
              </TableRow>
            ) : (
              banks.map((bank) => (
                <TableRow key={bank.id}>
                  <TableCell>
                    {bank.isPrimary && (
                      <Star className="h-5 w-5 fill-yellow-400 text-yellow-400" />
                    )}
                  </TableCell>
                  <TableCell className="font-medium">{bank.bankName}</TableCell>
                  <TableCell className="font-mono">{bank.accountNumber}</TableCell>
                  <TableCell>{bank.accountName}</TableCell>
                  <TableCell>{bank.branchName || "-"}</TableCell>
                  <TableCell>
                    {bank.checkPrefix ? (
                      <Badge variant="outline">{bank.checkPrefix}</Badge>
                    ) : (
                      "-"
                    )}
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
                        <DropdownMenuLabel>Aksi</DropdownMenuLabel>
                        <DropdownMenuSeparator />

                        {/* Edit */}
                        <DropdownMenuItem
                          onClick={() => setEditingBank(bank)}
                          className="cursor-pointer"
                        >
                          <Pencil className="mr-2 h-4 w-4" />
                          Edit Rekening
                        </DropdownMenuItem>

                        {/* Delete - Only if not the last bank */}
                        {banks.length > 1 && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() => setDeletingBank(bank)}
                              className="cursor-pointer text-red-600 focus:text-red-600"
                            >
                              <Trash2 className="mr-2 h-4 w-4" />
                              Hapus
                            </DropdownMenuItem>
                          </>
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

      {/* Edit Bank Dialog */}
      <Dialog open={!!editingBank} onOpenChange={(open) => !open && setEditingBank(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Rekening Bank</DialogTitle>
            <DialogDescription>
              Perbarui informasi rekening bank perusahaan
            </DialogDescription>
          </DialogHeader>
          {editingBank && (
            <BankAccountForm
              defaultValues={editingBank}
              onSuccess={() => setEditingBank(null)}
              onCancel={() => setEditingBank(null)}
            />
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Bank Alert Dialog */}
      <AlertDialog
        open={!!deletingBank}
        onOpenChange={(open) => !open && setDeletingBank(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hapus Rekening Bank</AlertDialogTitle>
            <AlertDialogDescription>
              Apakah Anda yakin ingin menghapus rekening bank{" "}
              <strong>{deletingBank?.bankName}</strong> atas nama{" "}
              <strong>{deletingBank?.accountName}</strong>?
              <br />
              <br />
              Tindakan ini tidak dapat dibatalkan.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Batal</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-red-600 hover:bg-red-700"
            >
              {isDeleting ? "Menghapus..." : "Hapus"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
