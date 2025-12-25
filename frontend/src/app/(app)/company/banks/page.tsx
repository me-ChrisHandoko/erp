/**
 * Bank Accounts Page
 *
 * Page for managing company bank accounts.
 * Supports add, edit, delete operations with primary bank management.
 */

"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { useGetBankAccountsQuery } from "@/store/services/companyApi";
import { BankAccountTable } from "@/components/company/bank-account-table";
import { BankAccountForm } from "@/components/company/bank-account-form";
import { ErrorDisplay } from "@/components/shared/error-display";
import { EmptyState } from "@/components/shared/empty-state";
import { PageHeader } from "@/components/shared/page-header";

export default function BanksPage() {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const { data: banks, isLoading, error } = useGetBankAccountsQuery();

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Perusahaan", href: "/company" },
          { label: "Rekening Bank" },
        ]}
      />

      {/* Main Content */}
      <div className="flex flex-1 flex-col gap-6 p-4 pt-0">
        {/* Page Header */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">Rekening Bank</h1>
            <p className="text-muted-foreground">
              Kelola rekening bank perusahaan untuk transaksi dan invoice
            </p>
          </div>
          <Button
            onClick={() => setIsAddDialogOpen(true)}
            disabled={isLoading}
            className="shrink-0"
          >
            <Plus className="mr-2 h-4 w-4" />
            Tambah Rekening
          </Button>
        </div>

        {/* Bank Accounts Card */}
        <Card className="shadow-sm">
          <CardContent className="pt-6">
          {/* Loading State */}
          {isLoading && (
            <div className="space-y-3">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          )}

          {/* Error State */}
          {error && (
            <ErrorDisplay
              title="Gagal Memuat Data"
              error={error}
            />
          )}

          {/* Empty State */}
          {!isLoading && !error && banks && banks.length === 0 && (
            <EmptyState
              title="Belum Ada Rekening Bank"
              description="Tambahkan rekening bank pertama perusahaan Anda untuk mulai melakukan transaksi."
              action={{
                label: "Tambah Rekening Bank",
                onClick: () => setIsAddDialogOpen(true),
              }}
            />
          )}

          {/* Bank Accounts Table */}
          {!isLoading && !error && banks && banks.length > 0 && (
            <BankAccountTable banks={banks} />
          )}
        </CardContent>
      </Card>

      {/* Information Card */}
      <Card>
        <CardHeader>
          <CardTitle>Informasi Rekening Bank</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Rekening Utama</h4>
              <p className="text-sm text-muted-foreground">
                Rekening yang ditandai sebagai utama akan digunakan secara otomatis
                untuk transaksi dan invoice. Hanya satu rekening yang bisa menjadi
                rekening utama.
              </p>
            </div>

            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Validasi Rekening</h4>
              <ul className="text-sm text-muted-foreground space-y-1 list-disc list-inside">
                <li>Nomor rekening minimal 8 digit, hanya angka</li>
                <li>Nama pemilik rekening minimal 3 karakter</li>
                <li>Minimal harus ada 1 rekening bank aktif</li>
              </ul>
            </div>

            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Prefix Cek</h4>
              <p className="text-sm text-muted-foreground">
                Prefix cek digunakan untuk menghasilkan nomor cek otomatis (contoh:
                CHK-001, BNI-001). Field ini bersifat opsional.
              </p>
            </div>

            <div className="space-y-2">
              <h4 className="font-semibold text-sm">Menghapus Rekening</h4>
              <p className="text-sm text-muted-foreground">
                Anda tidak bisa menghapus rekening terakhir. Minimal harus ada 1
                rekening bank yang aktif di sistem.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

        {/* Add Bank Dialog */}
        <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Tambah Rekening Bank</DialogTitle>
              <DialogDescription>
                Isi formulir di bawah ini untuk menambahkan rekening bank baru perusahaan
              </DialogDescription>
            </DialogHeader>
            <BankAccountForm
              onSuccess={() => setIsAddDialogOpen(false)}
              onCancel={() => setIsAddDialogOpen(false)}
            />
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}
