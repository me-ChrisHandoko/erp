/**
 * Create Delivery Form Component
 *
 * Form for creating new deliveries with:
 * - Sales order selection (only APPROVED orders)
 * - Delivery date and type selection
 * - Driver/vehicle OR expedition information
 * - Delivery address (pre-filled from customer)
 * - Items auto-populated from sales order
 * - Form validation and submission
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Loader2, Truck, Calendar, MapPin, User, Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { useCreateDeliveryMutation } from "@/store/services/deliveryApi";
import { useListSalesOrdersQuery, useGetSalesOrderQuery } from "@/store/services/salesOrderApi";
import { DELIVERY_TYPE_OPTIONS } from "@/types/delivery.types";
import type { CreateDeliveryPayload, DeliveryType } from "@/types/delivery.types";

interface CreateDeliveryFormProps {
  onSuccess: (deliveryId: string) => void;
  onCancel: () => void;
}

export function CreateDeliveryForm({
  onSuccess,
  onCancel,
}: CreateDeliveryFormProps) {
  const router = useRouter();
  const [createDelivery, { isLoading: isSubmitting, error: submitError }] =
    useCreateDeliveryMutation();

  // Fetch approved sales orders for selection
  const { data: salesOrdersData, isLoading: isLoadingSalesOrders } =
    useListSalesOrdersQuery({
      page: 1,
      pageSize: 100,
      status: "APPROVED", // Only show approved orders
      sortBy: "orderDate",
      sortOrder: "desc",
    });

  // Form state
  const [selectedSalesOrderId, setSelectedSalesOrderId] = useState("");
  const [deliveryDate, setDeliveryDate] = useState(
    new Date().toISOString().split("T")[0]
  );
  const [deliveryType, setDeliveryType] = useState<DeliveryType>("NORMAL");
  const [deliveryMethod, setDeliveryMethod] = useState<"driver" | "expedition">(
    "driver"
  );
  const [driverName, setDriverName] = useState("");
  const [vehicleNumber, setVehicleNumber] = useState("");
  const [expeditionService, setExpeditionService] = useState("");
  const [ttnkNumber, setTtnkNumber] = useState("");
  const [deliveryAddress, setDeliveryAddress] = useState("");
  const [notes, setNotes] = useState("");

  // Fetch selected sales order details
  const { data: selectedSalesOrder } = useGetSalesOrderQuery(
    selectedSalesOrderId,
    {
      skip: !selectedSalesOrderId,
    }
  );

  // Note: Delivery address should be filled manually by user
  // as SalesOrderResponse doesn't include customer address details

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedSalesOrder) {
      return;
    }

    // Build items from sales order items
    const items = (selectedSalesOrder.items || []).map((item) => ({
      salesOrderItemId: item.id,
      productId: item.productId,
      productUnitId: item.unitId || null,
      batchId: null, // Will be selected during picking if batch tracked
      quantity: parseFloat(item.orderedQty),
      notes: null,
    }));

    // Build delivery payload
    const payload: CreateDeliveryPayload = {
      salesOrderId: selectedSalesOrderId,
      deliveryDate,
      warehouseId: selectedSalesOrder.warehouseId,
      customerId: selectedSalesOrder.customerId,
      type: deliveryType,
      deliveryAddress: deliveryAddress || null,
      driverName: deliveryMethod === "driver" ? driverName || null : null,
      vehicleNumber:
        deliveryMethod === "driver" ? vehicleNumber || null : null,
      expeditionService:
        deliveryMethod === "expedition" ? expeditionService || null : null,
      ttnkNumber: deliveryMethod === "expedition" ? ttnkNumber || null : null,
      notes: notes || null,
      items,
    };

    try {
      const result = await createDelivery(payload).unwrap();
      onSuccess(result.id);
    } catch (error) {
      console.error("Failed to create delivery:", error);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Error Alert */}
      {submitError && (
        <Alert variant="destructive">
          <AlertDescription>
            {"data" in submitError && submitError.data
              ? String((submitError.data as any).message)
              : "Gagal membuat pengiriman. Silakan coba lagi."}
          </AlertDescription>
        </Alert>
      )}

      {/* Sales Order Selection */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Sales Order
          </CardTitle>
          <CardDescription>
            Pilih sales order yang akan dikirim (hanya order yang sudah
            disetujui)
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="salesOrder">Sales Order *</Label>
            <Select
              value={selectedSalesOrderId}
              onValueChange={setSelectedSalesOrderId}
              required
            >
              <SelectTrigger id="salesOrder" className="w-full">
                <SelectValue placeholder="Pilih Sales Order" />
              </SelectTrigger>
              <SelectContent>
                {isLoadingSalesOrders ? (
                  <SelectItem value="loading" disabled>
                    Memuat...
                  </SelectItem>
                ) : salesOrdersData?.data && salesOrdersData.data.length > 0 ? (
                  salesOrdersData.data.map((so) => (
                    <SelectItem key={so.id} value={so.id}>
                      {so.orderNumber} - {so.customerName} (
                      {new Date(so.orderDate).toLocaleDateString("id-ID")})
                    </SelectItem>
                  ))
                ) : (
                  <SelectItem value="empty" disabled>
                    Tidak ada sales order yang tersedia
                  </SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>

          {/* Show selected SO details */}
          {selectedSalesOrder && (
            <div className="rounded-lg border bg-muted/50 p-4 space-y-2">
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <span className="text-muted-foreground">Customer:</span>
                  <p className="font-medium">
                    {selectedSalesOrder.customerName}
                  </p>
                </div>
                <div>
                  <span className="text-muted-foreground">Gudang:</span>
                  <p className="font-medium">
                    {selectedSalesOrder.warehouseName}
                  </p>
                </div>
                <div>
                  <span className="text-muted-foreground">Total:</span>
                  <p className="font-medium">
                    Rp{" "}
                    {Number(selectedSalesOrder.totalAmount).toLocaleString(
                      "id-ID"
                    )}
                  </p>
                </div>
                <div>
                  <span className="text-muted-foreground">Item:</span>
                  <p className="font-medium">
                    {selectedSalesOrder.items?.length || 0} produk
                  </p>
                </div>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Delivery Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Informasi Pengiriman
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="deliveryDate">Tanggal Pengiriman *</Label>
              <Input
                id="deliveryDate"
                type="date"
                value={deliveryDate}
                onChange={(e) => setDeliveryDate(e.target.value)}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="deliveryType">Jenis Pengiriman *</Label>
              <Select
                value={deliveryType}
                onValueChange={(value) =>
                  setDeliveryType(value as DeliveryType)
                }
                required
              >
                <SelectTrigger id="deliveryType" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {DELIVERY_TYPE_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="deliveryAddress">Alamat Pengiriman</Label>
            <Textarea
              id="deliveryAddress"
              value={deliveryAddress}
              onChange={(e) => setDeliveryAddress(e.target.value)}
              placeholder="Alamat lengkap pengiriman"
              rows={3}
            />
          </div>
        </CardContent>
      </Card>

      {/* Delivery Method */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Truck className="h-5 w-5" />
            Metode Pengiriman
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <RadioGroup
            value={deliveryMethod}
            onValueChange={(value) =>
              setDeliveryMethod(value as "driver" | "expedition")
            }
          >
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="driver" id="driver" />
              <Label htmlFor="driver">Sopir Sendiri</Label>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="expedition" id="expedition" />
              <Label htmlFor="expedition">Ekspedisi / Kurir</Label>
            </div>
          </RadioGroup>

          <Separator />

          {deliveryMethod === "driver" ? (
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="driverName">Nama Sopir</Label>
                <Input
                  id="driverName"
                  value={driverName}
                  onChange={(e) => setDriverName(e.target.value)}
                  placeholder="Nama sopir"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="vehicleNumber">Nomor Kendaraan</Label>
                <Input
                  id="vehicleNumber"
                  value={vehicleNumber}
                  onChange={(e) => setVehicleNumber(e.target.value)}
                  placeholder="B 1234 XYZ"
                />
              </div>
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="expeditionService">Nama Ekspedisi</Label>
                <Input
                  id="expeditionService"
                  value={expeditionService}
                  onChange={(e) => setExpeditionService(e.target.value)}
                  placeholder="JNE, SiCepat, dll"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="ttnkNumber">Nomor Resi</Label>
                <Input
                  id="ttnkNumber"
                  value={ttnkNumber}
                  onChange={(e) => setTtnkNumber(e.target.value)}
                  placeholder="Nomor tracking"
                />
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Items Preview */}
      {selectedSalesOrder && selectedSalesOrder.items && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Package className="h-5 w-5" />
              Item yang Akan Dikirim
            </CardTitle>
            <CardDescription>
              {selectedSalesOrder.items.length} produk dari sales order
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Produk</TableHead>
                    <TableHead className="text-center">Unit</TableHead>
                    <TableHead className="text-right">Qty</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {selectedSalesOrder.items.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">
                            {item.productName}
                          </div>
                          <div className="text-xs text-muted-foreground font-mono">
                            {item.productCode}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-center">
                        <Badge variant="secondary">
                          {item.unitName}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-medium">
                        {Number(item.orderedQty).toLocaleString("id-ID")}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Notes */}
      <Card>
        <CardHeader>
          <CardTitle>Catatan</CardTitle>
        </CardHeader>
        <CardContent>
          <Textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Catatan tambahan (opsional)"
            rows={3}
          />
        </CardContent>
      </Card>

      {/* Action Buttons */}
      <div className="flex justify-end gap-3">
        <Button type="button" variant="outline" onClick={onCancel}>
          Batal
        </Button>
        <Button
          type="submit"
          disabled={isSubmitting || !selectedSalesOrderId}
        >
          {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Buat Pengiriman
        </Button>
      </div>
    </form>
  );
}
