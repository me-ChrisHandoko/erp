/**
 * Update Delivery Status Form Component
 *
 * Form for updating delivery status and tracking information:
 * - Status transitions (PREPARED → IN_TRANSIT → DELIVERED → CONFIRMED)
 * - Departure time (when starting delivery)
 * - Arrival time and receiver info (when delivered)
 * - Proof of delivery (POD): signature and photo
 * - Cancel with notes
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  Loader2,
  Truck,
  CheckCircle,
  XCircle,
  Clock,
  User,
  FileSignature,
  Image as ImageIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import {
  useUpdateDeliveryStatusMutation,
  useStartDeliveryMutation,
  useCompleteDeliveryMutation,
  useConfirmDeliveryMutation,
  useCancelDeliveryMutation,
} from "@/store/services/deliveryApi";
import {
  getDeliveryStatusLabel,
  getDeliveryStatusColor,
  DELIVERY_STATUS_OPTIONS,
} from "@/types/delivery.types";
import type { DeliveryResponse, DeliveryStatus } from "@/types/delivery.types";

interface UpdateDeliveryStatusFormProps {
  delivery: DeliveryResponse;
  onSuccess: () => void;
  onCancel: () => void;
}

export function UpdateDeliveryStatusForm({
  delivery,
  onSuccess,
  onCancel,
}: UpdateDeliveryStatusFormProps) {
  const router = useRouter();

  // Mutations
  const [updateStatus, { isLoading: isUpdating }] =
    useUpdateDeliveryStatusMutation();
  const [startDelivery, { isLoading: isStarting }] =
    useStartDeliveryMutation();
  const [completeDelivery, { isLoading: isCompleting }] =
    useCompleteDeliveryMutation();
  const [confirmDelivery, { isLoading: isConfirming }] =
    useConfirmDeliveryMutation();
  const [cancelDelivery, { isLoading: isCancelling }] =
    useCancelDeliveryMutation();

  // Form state
  const [selectedStatus, setSelectedStatus] = useState<DeliveryStatus>(
    delivery.status
  );
  const [departureTime, setDepartureTime] = useState(
    delivery.departureTime
      ? new Date(delivery.departureTime).toISOString().slice(0, 16)
      : ""
  );
  const [arrivalTime, setArrivalTime] = useState(
    delivery.arrivalTime
      ? new Date(delivery.arrivalTime).toISOString().slice(0, 16)
      : ""
  );
  const [receivedBy, setReceivedBy] = useState(delivery.receivedBy || "");
  const [signatureUrl, setSignatureUrl] = useState(
    delivery.signatureUrl || ""
  );
  const [photoUrl, setPhotoUrl] = useState(delivery.photoUrl || "");
  const [cancelNotes, setCancelNotes] = useState("");
  const [showCancelDialog, setShowCancelDialog] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Get available status transitions based on current status
  const getAvailableStatuses = (): DeliveryStatus[] => {
    switch (delivery.status) {
      case "PREPARED":
        return ["PREPARED", "IN_TRANSIT", "CANCELLED"];
      case "IN_TRANSIT":
        return ["IN_TRANSIT", "DELIVERED", "CANCELLED"];
      case "DELIVERED":
        return ["DELIVERED", "CONFIRMED", "CANCELLED"];
      case "CONFIRMED":
        return ["CONFIRMED"]; // Cannot change from confirmed
      case "CANCELLED":
        return ["CANCELLED"]; // Cannot change from cancelled
      default:
        return [];
    }
  };

  const availableStatuses = getAvailableStatuses();

  // Handle quick action buttons
  const handleStartDelivery = async () => {
    try {
      setError(null);
      const time = departureTime || new Date().toISOString();
      await startDelivery({ id: delivery.id, departureTime: time }).unwrap();
      onSuccess();
    } catch (err: any) {
      setError(err.data?.message || "Gagal memulai pengiriman");
    }
  };

  const handleCompleteDelivery = async () => {
    try {
      setError(null);
      await completeDelivery({
        id: delivery.id,
        receivedBy: receivedBy || undefined,
        signatureUrl: signatureUrl || undefined,
        photoUrl: photoUrl || undefined,
      }).unwrap();
      onSuccess();
    } catch (err: any) {
      setError(err.data?.message || "Gagal menyelesaikan pengiriman");
    }
  };

  const handleConfirmDelivery = async () => {
    try {
      setError(null);
      await confirmDelivery(delivery.id).unwrap();
      onSuccess();
    } catch (err: any) {
      setError(err.data?.message || "Gagal mengkonfirmasi pengiriman");
    }
  };

  const handleCancelDelivery = async () => {
    try {
      setError(null);
      await cancelDelivery({
        id: delivery.id,
        notes: cancelNotes || undefined,
      }).unwrap();
      setShowCancelDialog(false);
      onSuccess();
    } catch (err: any) {
      setError(err.data?.message || "Gagal membatalkan pengiriman");
    }
  };

  // Handle manual status update (for flexibility)
  const handleManualStatusUpdate = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setError(null);
      await updateStatus({
        id: delivery.id,
        data: {
          status: selectedStatus,
          receivedBy: receivedBy || undefined,
          signatureUrl: signatureUrl || undefined,
          photoUrl: photoUrl || undefined,
        },
      }).unwrap();
      onSuccess();
    } catch (err: any) {
      setError(err.data?.message || "Gagal memperbarui status");
    }
  };

  const isLoading =
    isUpdating || isStarting || isCompleting || isConfirming || isCancelling;

  return (
    <div className="space-y-6">
      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Current Status */}
      <Card>
        <CardHeader>
          <CardTitle>Status Saat Ini</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3">
            <Badge
              className={`${getDeliveryStatusColor(
                delivery.status
              )} text-white text-base px-4 py-2`}
            >
              {getDeliveryStatusLabel(delivery.status)}
            </Badge>
            <span className="text-sm text-muted-foreground">
              Nomor: {delivery.deliveryNumber}
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions based on current status */}
      {delivery.status !== "CANCELLED" && delivery.status !== "CONFIRMED" && (
        <Card>
          <CardHeader>
            <CardTitle>Aksi Cepat</CardTitle>
            <CardDescription>
              Pilih aksi berdasarkan status pengiriman
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {delivery.status === "PREPARED" && (
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor="departureTime">Waktu Berangkat</Label>
                  <Input
                    id="departureTime"
                    type="datetime-local"
                    value={departureTime}
                    onChange={(e) => setDepartureTime(e.target.value)}
                  />
                </div>
                <Button
                  onClick={handleStartDelivery}
                  disabled={isLoading}
                  className="w-full"
                >
                  {isStarting && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  <Truck className="mr-2 h-4 w-4" />
                  Mulai Pengiriman
                </Button>
              </div>
            )}

            {delivery.status === "IN_TRANSIT" && (
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor="receivedBy">Diterima Oleh</Label>
                  <Input
                    id="receivedBy"
                    value={receivedBy}
                    onChange={(e) => setReceivedBy(e.target.value)}
                    placeholder="Nama penerima"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="signatureUrl">URL Tanda Tangan (POD)</Label>
                  <Input
                    id="signatureUrl"
                    value={signatureUrl}
                    onChange={(e) => setSignatureUrl(e.target.value)}
                    placeholder="https://..."
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="photoUrl">URL Foto Pengiriman</Label>
                  <Input
                    id="photoUrl"
                    value={photoUrl}
                    onChange={(e) => setPhotoUrl(e.target.value)}
                    placeholder="https://..."
                  />
                </div>
                <Button
                  onClick={handleCompleteDelivery}
                  disabled={isLoading}
                  className="w-full"
                >
                  {isCompleting && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  <CheckCircle className="mr-2 h-4 w-4" />
                  Tandai Terkirim
                </Button>
              </div>
            )}

            {delivery.status === "DELIVERED" && (
              <Button
                onClick={handleConfirmDelivery}
                disabled={isLoading}
                className="w-full"
              >
                {isConfirming && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                <CheckCircle className="mr-2 h-4 w-4" />
                Konfirmasi Pengiriman
              </Button>
            )}

            <Separator />

            <Button
              variant="destructive"
              onClick={() => setShowCancelDialog(true)}
              disabled={isLoading}
              className="w-full"
            >
              <XCircle className="mr-2 h-4 w-4" />
              Batalkan Pengiriman
            </Button>
          </CardContent>
        </Card>
      )}

      {/* Manual Status Update (Advanced) */}
      {delivery.status !== "CANCELLED" && delivery.status !== "CONFIRMED" && (
        <Card>
          <CardHeader>
            <CardTitle>Update Manual (Advanced)</CardTitle>
            <CardDescription>
              Perbarui status dan informasi pengiriman secara manual
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleManualStatusUpdate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="status">Status Baru</Label>
                <Select
                  value={selectedStatus}
                  onValueChange={(value) =>
                    setSelectedStatus(value as DeliveryStatus)
                  }
                >
                  <SelectTrigger id="status">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {DELIVERY_STATUS_OPTIONS.filter((opt) =>
                      availableStatuses.includes(opt.value)
                    ).map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        {option.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="manualDepartureTime">Waktu Berangkat</Label>
                  <Input
                    id="manualDepartureTime"
                    type="datetime-local"
                    value={departureTime}
                    onChange={(e) => setDepartureTime(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="arrivalTime">Waktu Tiba</Label>
                  <Input
                    id="arrivalTime"
                    type="datetime-local"
                    value={arrivalTime}
                    onChange={(e) => setArrivalTime(e.target.value)}
                  />
                </div>
              </div>

              <div className="flex justify-end gap-3">
                <Button type="button" variant="outline" onClick={onCancel}>
                  Batal
                </Button>
                <Button type="submit" disabled={isLoading}>
                  {isUpdating && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Simpan Perubahan
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Cancel Dialog */}
      <AlertDialog open={showCancelDialog} onOpenChange={setShowCancelDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Batalkan Pengiriman?</AlertDialogTitle>
            <AlertDialogDescription>
              Pengiriman yang dibatalkan tidak dapat diubah lagi. Pastikan Anda
              yakin.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="my-4 space-y-2">
            <Label htmlFor="cancelNotes">Alasan Pembatalan</Label>
            <Textarea
              id="cancelNotes"
              value={cancelNotes}
              onChange={(e) => setCancelNotes(e.target.value)}
              placeholder="Jelaskan alasan pembatalan..."
              rows={3}
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Tidak</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleCancelDelivery}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isCancelling && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              Ya, Batalkan
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
