/**
 * Payment Detail Component
 *
 * Displays comprehensive payment information in a card layout:
 * - Payment header with number and date
 * - Supplier information
 * - Payment amount and method
 * - Reference and notes
 * - Related purchase order
 * - Approval information
 * - Action buttons
 */

"use client";

import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  ArrowLeft,
  Pencil,
  Building2,
  Calendar,
  CreditCard,
  FileText,
  CheckCircle2,
  Package,
} from "lucide-react";
import { usePermissions } from "@/hooks/use-permissions";
import type { PaymentResponse } from "@/types/payment.types";
import { PAYMENT_METHOD_LABELS, PAYMENT_STATUS_LABELS } from "@/types/payment.types";

interface PaymentDetailProps {
  payment: PaymentResponse;
}

export function PaymentDetail({ payment }: PaymentDetailProps) {
  const router = useRouter();
  const permissions = usePermissions();
  const canEdit = permissions.canEdit('supplier-payments');

  // Format date to Indonesian format
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "long",
      year: "numeric",
    });
  };

  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString("id-ID", {
      day: "2-digit",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Header Actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.back()}
          className="w-fit"
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Kembali
        </Button>

        {canEdit && (
          <Button
            onClick={() => router.push(`/procurement/payments/${payment.id}/edit`)}
          >
            <Pencil className="mr-2 h-4 w-4" />
            Edit Pembayaran
          </Button>
        )}
      </div>

      {/* Payment Information Cards */}
      <div className="grid gap-4 md:grid-cols-2">
        {/* Payment Details Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Informasi Pembayaran
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="text-sm font-medium text-muted-foreground">
                Nomor Pembayaran
              </div>
              <div className="text-lg font-mono font-semibold">
                {payment.paymentNumber}
              </div>
            </div>

            <Separator />

            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-sm font-medium text-muted-foreground mb-1">
                  Tanggal Pembayaran
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span>{formatDate(payment.paymentDate)}</span>
                </div>
              </div>

              <div>
                <div className="text-sm font-medium text-muted-foreground mb-1">
                  Metode Pembayaran
                </div>
                <Badge variant="secondary">
                  <CreditCard className="mr-1 h-3 w-3" />
                  {PAYMENT_METHOD_LABELS[payment.paymentMethod]}
                </Badge>
              </div>
            </div>

            <Separator />

            <div>
              <div className="text-sm font-medium text-muted-foreground mb-1">
                Jumlah Pembayaran
              </div>
              <div className="text-2xl font-bold">
                Rp {Number(payment.amount).toLocaleString("id-ID")}
              </div>
            </div>

            {payment.reference && (
              <>
                <Separator />
                <div>
                  <div className="text-sm font-medium text-muted-foreground mb-1">
                    Referensi
                  </div>
                  <div className="text-sm font-mono">
                    {payment.reference}
                  </div>
                </div>
              </>
            )}

            {payment.notes && (
              <>
                <Separator />
                <div>
                  <div className="text-sm font-medium text-muted-foreground mb-1">
                    Catatan
                  </div>
                  <div className="text-sm whitespace-pre-wrap">
                    {payment.notes}
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Supplier and Related Info Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Informasi Pemasok
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <div className="text-sm font-medium text-muted-foreground mb-1">
                Nama Pemasok
              </div>
              <div className="text-lg font-semibold">
                {payment.supplierName}
              </div>
              {payment.supplierCode && (
                <div className="text-sm text-muted-foreground font-mono">
                  {payment.supplierCode}
                </div>
              )}
            </div>

            {payment.purchaseOrderId && (
              <>
                <Separator />
                <div>
                  <div className="text-sm font-medium text-muted-foreground mb-1">
                    Purchase Order Terkait
                  </div>
                  <div className="flex items-center gap-2">
                    <Package className="h-4 w-4 text-muted-foreground" />
                    <a
                      href={`/procurement/orders/${payment.purchaseOrderId}`}
                      className="text-sm font-mono text-blue-600 hover:underline"
                    >
                      {payment.poNumber}
                    </a>
                  </div>
                </div>
              </>
            )}

            {payment.bankAccountId && (
              <>
                <Separator />
                <div>
                  <div className="text-sm font-medium text-muted-foreground mb-1">
                    Rekening Bank
                  </div>
                  <div className="text-sm">
                    {payment.bankAccountName || payment.bankAccountId}
                  </div>
                </div>
              </>
            )}

            {payment.status && (
              <>
                <Separator />
                <div>
                  <div className="text-sm font-medium text-muted-foreground mb-1">
                    Status
                  </div>
                  <Badge
                    className={
                      payment.status === 'APPROVED'
                        ? 'bg-green-500 text-white hover:bg-green-600'
                        : payment.status === 'REJECTED'
                        ? 'bg-red-500 text-white hover:bg-red-600'
                        : 'bg-yellow-500 text-white hover:bg-yellow-600'
                    }
                  >
                    {PAYMENT_STATUS_LABELS[payment.status]}
                  </Badge>
                </div>
              </>
            )}

            {payment.approvedBy && payment.approvedAt && (
              <>
                <Separator />
                <div>
                  <div className="text-sm font-medium text-muted-foreground mb-1">
                    Disetujui Oleh
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <CheckCircle2 className="h-4 w-4 text-green-500" />
                    <div>
                      <div className="font-medium">{payment.approvedBy}</div>
                      <div className="text-xs text-muted-foreground">
                        {formatDateTime(payment.approvedAt)}
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Timestamps Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Informasi Sistem</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Dibuat: </span>
              <span>{formatDateTime(payment.createdAt)}</span>
              {payment.createdBy && (
                <span className="text-muted-foreground"> oleh {payment.createdBy}</span>
              )}
            </div>
            {payment.updatedAt && payment.updatedAt !== payment.createdAt && (
              <div>
                <span className="text-muted-foreground">Diperbarui: </span>
                <span>{formatDateTime(payment.updatedAt)}</span>
                {payment.updatedBy && (
                  <span className="text-muted-foreground"> oleh {payment.updatedBy}</span>
                )}
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
