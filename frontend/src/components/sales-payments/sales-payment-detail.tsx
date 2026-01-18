/**
 * Sales Payment Detail Component (Enhanced)
 *
 * Displays comprehensive payment information including:
 * - Payment details and status
 * - Customer and invoice information
 * - Payment method and banking details
 * - Check/Giro tracking with status update
 * - Payment receipt download
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Pencil, ArrowLeft, RefreshCw, Download, Loader2 } from "lucide-react";
import type { SalesPaymentResponse, CheckStatus } from "@/types/sales-payment.types";
import { PAYMENT_METHOD_LABELS, CHECK_STATUS_LABELS } from "@/types/sales-payment.types";
import { usePermissions } from "@/hooks/use-permissions";
import { useUpdateCheckStatusMutation } from "@/store/services/salesPaymentApi";
import { useToast } from "@/hooks/use-toast";

interface SalesPaymentDetailProps {
  payment: SalesPaymentResponse;
}

export function SalesPaymentDetail({ payment }: SalesPaymentDetailProps) {
  const router = useRouter();
  const permissions = usePermissions();
  const { toast } = useToast();
  const canEdit = permissions.canEdit('customer-payments');

  const [updateCheckStatus, { isLoading: isUpdatingStatus }] = useUpdateCheckStatusMutation();
  const [isStatusDialogOpen, setIsStatusDialogOpen] = useState(false);
  const [newCheckStatus, setNewCheckStatus] = useState<CheckStatus | "">("");
  const [statusNotes, setStatusNotes] = useState("");

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    });
  };

  const formatCurrency = (amount: string) => {
    return `Rp ${Number(amount).toLocaleString('id-ID')}`;
  };

  const handleUpdateCheckStatus = async () => {
    if (!newCheckStatus) {
      toast({
        title: "Status Belum Dipilih",
        description: "Silakan pilih status cek/giro",
        variant: "destructive",
      });
      return;
    }

    try {
      await updateCheckStatus({
        id: payment.id,
        data: {
          checkStatus: newCheckStatus as CheckStatus,
          notes: statusNotes || undefined,
        },
      }).unwrap();

      toast({
        title: "Status Berhasil Diupdate",
        description: `Status cek/giro telah diubah menjadi ${CHECK_STATUS_LABELS[newCheckStatus as CheckStatus]}`,
      });

      setIsStatusDialogOpen(false);
      setNewCheckStatus("");
      setStatusNotes("");
      router.refresh();
    } catch (error: any) {
      toast({
        title: "Gagal Update Status",
        description: error?.data?.error?.message || "Terjadi kesalahan saat mengupdate status",
        variant: "destructive",
      });
    }
  };

  const handleDownloadReceipt = () => {
    // Create printable receipt content
    const receiptWindow = window.open('', '_blank');
    if (!receiptWindow) {
      toast({
        title: "Gagal Membuka Kwitansi",
        description: "Mohon izinkan popup untuk mencetak kwitansi",
        variant: "destructive",
      });
      return;
    }

    const receiptHTML = `
      <!DOCTYPE html>
      <html>
      <head>
        <title>Kwitansi Pembayaran - ${payment.paymentNumber}</title>
        <style>
          @media print {
            @page { margin: 1cm; }
            body { margin: 0; }
            .no-print { display: none !important; }
          }
          body {
            font-family: Arial, sans-serif;
            padding: 20px;
            max-width: 800px;
            margin: 0 auto;
          }
          .header {
            text-align: center;
            border-bottom: 2px solid #333;
            padding-bottom: 20px;
            margin-bottom: 30px;
          }
          .header h1 {
            margin: 0 0 10px 0;
            font-size: 28px;
          }
          .header p {
            margin: 5px 0;
            color: #666;
          }
          .info-section {
            margin-bottom: 30px;
          }
          .info-row {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #eee;
          }
          .info-label {
            font-weight: bold;
            color: #333;
            width: 200px;
          }
          .info-value {
            flex: 1;
            text-align: right;
          }
          .amount-section {
            background: #f5f5f5;
            padding: 20px;
            margin: 30px 0;
            border-radius: 8px;
          }
          .amount-row {
            display: flex;
            justify-content: space-between;
            font-size: 24px;
            font-weight: bold;
          }
          .signature-section {
            margin-top: 60px;
            display: flex;
            justify-content: space-between;
          }
          .signature-box {
            text-align: center;
            width: 200px;
          }
          .signature-line {
            border-top: 1px solid #333;
            margin-top: 80px;
            padding-top: 10px;
          }
          .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            text-align: center;
            color: #666;
            font-size: 12px;
          }
          .print-button {
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 10px 20px;
            background: #333;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
          }
          .print-button:hover {
            background: #555;
          }
        </style>
      </head>
      <body>
        <button class="print-button no-print" onclick="window.print()">🖨️ Cetak / Simpan PDF</button>

        <div class="header">
          <h1>KWITANSI PEMBAYARAN</h1>
          <p>Nomor: ${payment.paymentNumber}</p>
          <p>Tanggal: ${formatDate(payment.paymentDate)}</p>
        </div>

        <div class="info-section">
          <h3 style="margin-bottom: 15px;">Informasi Pelanggan</h3>
          <div class="info-row">
            <div class="info-label">Nama Pelanggan:</div>
            <div class="info-value">${payment.customerName}</div>
          </div>
          ${payment.customerCode ? `
          <div class="info-row">
            <div class="info-label">Kode Pelanggan:</div>
            <div class="info-value">${payment.customerCode}</div>
          </div>
          ` : ''}
          <div class="info-row">
            <div class="info-label">Nomor Invoice:</div>
            <div class="info-value">${payment.invoiceNumber}</div>
          </div>
        </div>

        <div class="info-section">
          <h3 style="margin-bottom: 15px;">Detail Pembayaran</h3>
          <div class="info-row">
            <div class="info-label">Metode Pembayaran:</div>
            <div class="info-value">${PAYMENT_METHOD_LABELS[payment.paymentMethod]}</div>
          </div>
          ${payment.reference ? `
          <div class="info-row">
            <div class="info-label">Referensi:</div>
            <div class="info-value">${payment.reference}</div>
          </div>
          ` : ''}
          ${payment.bankAccountName ? `
          <div class="info-row">
            <div class="info-label">Rekening Bank:</div>
            <div class="info-value">${payment.bankAccountName}</div>
          </div>
          ` : ''}
          ${payment.checkNumber ? `
          <div class="info-row">
            <div class="info-label">Nomor Cek/Giro:</div>
            <div class="info-value">${payment.checkNumber}</div>
          </div>
          ` : ''}
          ${payment.checkDate ? `
          <div class="info-row">
            <div class="info-label">Tanggal Jatuh Tempo:</div>
            <div class="info-value">${formatDate(payment.checkDate)}</div>
          </div>
          ` : ''}
          ${payment.checkStatus ? `
          <div class="info-row">
            <div class="info-label">Status Cek/Giro:</div>
            <div class="info-value">${CHECK_STATUS_LABELS[payment.checkStatus]}</div>
          </div>
          ` : ''}
        </div>

        ${payment.notes ? `
        <div class="info-section">
          <h3 style="margin-bottom: 15px;">Catatan</h3>
          <p style="padding: 10px; background: #f9f9f9; border-radius: 4px;">${payment.notes}</p>
        </div>
        ` : ''}

        <div class="amount-section">
          <div class="amount-row">
            <span>Total Pembayaran:</span>
            <span>${formatCurrency(payment.amount)}</span>
          </div>
        </div>

        <div class="signature-section">
          <div class="signature-box">
            <div class="signature-line">
              Penerima
            </div>
          </div>
          <div class="signature-box">
            <div class="signature-line">
              Yang Menyerahkan
            </div>
          </div>
        </div>

        <div class="footer">
          <p>Kwitansi ini dicetak pada ${new Date().toLocaleDateString('id-ID', {
            day: 'numeric',
            month: 'long',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
          })}</p>
          <p>Dokumen ini sah sebagai bukti pembayaran</p>
        </div>

        <script>
          // Auto-focus the print button
          window.addEventListener('load', () => {
            // Optional: Auto-print on load
            // window.print();
          });
        </script>
      </body>
      </html>
    `;

    receiptWindow.document.write(receiptHTML);
    receiptWindow.document.close();

    toast({
      title: "Kwitansi Dibuka",
      description: "Klik tombol 'Cetak / Simpan PDF' untuk mencetak atau menyimpan sebagai PDF",
    });
  };

  return (
    <>
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Header with actions */}
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">
              Detail Pembayaran
            </h1>
            <p className="text-muted-foreground">
              Informasi lengkap pembayaran pelanggan
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => router.back()}>
              <ArrowLeft className="mr-2 h-4 w-4" />
              Kembali
            </Button>
            <Button
              variant="outline"
              onClick={handleDownloadReceipt}
            >
              <Download className="mr-2 h-4 w-4" />
              Kwitansi
            </Button>
            {canEdit && (
              <Button onClick={() => router.push(`/sales/payments/${payment.id}/edit`)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </Button>
            )}
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {/* Payment Information */}
          <Card>
            <CardHeader>
              <CardTitle>Informasi Pembayaran</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="text-sm text-muted-foreground">Nomor Pembayaran</div>
                <div className="font-mono font-semibold">{payment.paymentNumber}</div>
              </div>

              <Separator />

              <div>
                <div className="text-sm text-muted-foreground">Tanggal Pembayaran</div>
                <div className="font-medium">{formatDate(payment.paymentDate)}</div>
              </div>

              <div>
                <div className="text-sm text-muted-foreground">Jumlah Pembayaran</div>
                <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                  {formatCurrency(payment.amount)}
                </div>
              </div>

              <div>
                <div className="text-sm text-muted-foreground">Metode Pembayaran</div>
                <Badge variant="outline" className="mt-1">
                  {PAYMENT_METHOD_LABELS[payment.paymentMethod]}
                </Badge>
              </div>

              {payment.reference && (
                <div>
                  <div className="text-sm text-muted-foreground">Referensi</div>
                  <div className="font-mono">{payment.reference}</div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Customer & Invoice Information */}
          <Card>
            <CardHeader>
              <CardTitle>Pelanggan & Invoice</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <div className="text-sm text-muted-foreground">Pelanggan</div>
                <div className="font-medium">{payment.customerName}</div>
                {payment.customerCode && (
                  <div className="text-sm font-mono text-muted-foreground">
                    {payment.customerCode}
                  </div>
                )}
              </div>

              <Separator />

              <div>
                <div className="text-sm text-muted-foreground">Nomor Invoice</div>
                <div className="font-mono font-semibold">{payment.invoiceNumber}</div>
              </div>

              {payment.bankAccountName && (
                <>
                  <Separator />
                  <div>
                    <div className="text-sm text-muted-foreground">Rekening Bank</div>
                    <div className="font-medium">{payment.bankAccountName}</div>
                  </div>
                </>
              )}
            </CardContent>
          </Card>

          {/* Check/Giro Information (if applicable) */}
          {(payment.checkNumber || payment.checkStatus) && (
            <Card className="md:col-span-2">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Informasi Cek/Giro</CardTitle>
                {canEdit && payment.checkStatus && payment.checkStatus !== 'CANCELLED' && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setIsStatusDialogOpen(true)}
                  >
                    <RefreshCw className="mr-2 h-4 w-4" />
                    Update Status
                  </Button>
                )}
              </CardHeader>
              <CardContent className="grid gap-4 md:grid-cols-3">
                {payment.checkNumber && (
                  <div>
                    <div className="text-sm text-muted-foreground">Nomor Cek/Giro</div>
                    <div className="font-mono font-semibold">{payment.checkNumber}</div>
                  </div>
                )}

                {payment.checkDate && (
                  <div>
                    <div className="text-sm text-muted-foreground">Tanggal Jatuh Tempo</div>
                    <div className="font-medium">{formatDate(payment.checkDate)}</div>
                  </div>
                )}

                {payment.checkStatus && (
                  <div>
                    <div className="text-sm text-muted-foreground">Status</div>
                    <Badge
                      className={
                        payment.checkStatus === 'CLEARED'
                          ? "bg-green-500 text-white"
                          : payment.checkStatus === 'BOUNCED'
                          ? "bg-red-500 text-white"
                          : payment.checkStatus === 'CANCELLED'
                          ? "bg-gray-500 text-white"
                          : "bg-yellow-500 text-white"
                      }
                    >
                      {CHECK_STATUS_LABELS[payment.checkStatus]}
                    </Badge>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Notes */}
          {payment.notes && (
            <Card className="md:col-span-2">
              <CardHeader>
                <CardTitle>Catatan</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm whitespace-pre-wrap">{payment.notes}</p>
              </CardContent>
            </Card>
          )}

          {/* Audit Information */}
          <Card className="md:col-span-2">
            <CardHeader>
              <CardTitle>Informasi Audit</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-2">
              <div>
                <div className="text-sm text-muted-foreground">Dibuat Oleh</div>
                <div className="font-medium">{payment.createdBy}</div>
                <div className="text-sm text-muted-foreground">
                  {formatDate(payment.createdAt)}
                </div>
              </div>

              {payment.updatedBy && (
                <div>
                  <div className="text-sm text-muted-foreground">Diupdate Oleh</div>
                  <div className="font-medium">{payment.updatedBy}</div>
                  <div className="text-sm text-muted-foreground">
                    {formatDate(payment.updatedAt)}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Update Check Status Dialog */}
      <Dialog open={isStatusDialogOpen} onOpenChange={setIsStatusDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Status Cek/Giro</DialogTitle>
            <DialogDescription>
              Ubah status cek/giro pembayaran ini
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Status Baru</Label>
              <Select value={newCheckStatus} onValueChange={(value) => setNewCheckStatus(value as CheckStatus)}>
                <SelectTrigger>
                  <SelectValue placeholder="Pilih status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ISSUED">Diterbitkan</SelectItem>
                  <SelectItem value="CLEARED">Cair</SelectItem>
                  <SelectItem value="BOUNCED">Tolak</SelectItem>
                  <SelectItem value="CANCELLED">Batal</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Catatan (Opsional)</Label>
              <Textarea
                value={statusNotes}
                onChange={(e) => setStatusNotes(e.target.value)}
                placeholder="Tambahkan catatan jika diperlukan"
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsStatusDialogOpen(false);
                setNewCheckStatus("");
                setStatusNotes("");
              }}
              disabled={isUpdatingStatus}
            >
              Batal
            </Button>
            <Button onClick={handleUpdateCheckStatus} disabled={isUpdatingStatus || !newCheckStatus}>
              {isUpdatingStatus && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Update Status
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
