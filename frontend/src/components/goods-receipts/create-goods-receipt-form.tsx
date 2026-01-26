/**
 * Create Goods Receipt Form Component
 *
 * Form for creating a goods receipt from a confirmed purchase order.
 * Shows required indicators for batch number and expiry date based on
 * product tracking settings (isBatchTracked, isPerishable).
 */

"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useForm, useFieldArray } from "react-hook-form";
import {
  Package,
  Calendar,
  AlertCircle,
  Save,
  X,
  Info,
  Hash,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useCreateGoodsReceiptMutation } from "@/store/services/goodsReceiptApi";
import {
  useGetEffectiveToleranceQuery,
  useLazyGetEffectiveToleranceQuery,
} from "@/store/services/deliveryToleranceApi";
import { toast } from "sonner";
import type { PurchaseOrderResponse } from "@/types/purchase-order.types";
import type { CreateGoodsReceiptRequest } from "@/types/goods-receipt.types";
import type { EffectiveToleranceResponse } from "@/types/delivery-tolerance.types";

interface CreateGoodsReceiptFormProps {
  purchaseOrder: PurchaseOrderResponse;
  onSuccess?: () => void;
  onCancel?: () => void;
}

// Component to display tolerance info for a product
function ToleranceInfo({ productId, pendingQty, qtyToReceive }: {
  productId: string;
  pendingQty: number;
  qtyToReceive: number;
}) {
  const { data: tolerance, isLoading } = useGetEffectiveToleranceQuery(productId);

  if (isLoading) {
    return null;
  }

  if (!tolerance) {
    return null;
  }

  const underTolerance = parseFloat(tolerance.underDeliveryTolerance) || 0;
  const overTolerance = parseFloat(tolerance.overDeliveryTolerance) || 0;
  const unlimitedOver = tolerance.unlimitedOverDelivery;

  // Calculate allowed range
  const minAllowed = pendingQty * (1 - underTolerance / 100);
  const maxAllowed = unlimitedOver ? Infinity : pendingQty * (1 + overTolerance / 100);

  // Check if current qty is within tolerance
  const isSkipped = qtyToReceive === 0; // Item will be skipped (not sent to backend)
  const isUnderTolerance = !isSkipped && qtyToReceive < minAllowed;
  const isOverTolerance = !isSkipped && !unlimitedOver && qtyToReceive > maxAllowed;
  const isOutOfTolerance = isUnderTolerance || isOverTolerance;

  // Format tolerance display
  const toleranceText = unlimitedOver
    ? `Under: ${underTolerance}%, Over: Unlimited`
    : `Under: ${underTolerance}%, Over: ${overTolerance}%`;

  // Determine badge style and text
  const getBadgeStyle = () => {
    if (isSkipped) return "bg-gray-50 text-gray-600 border-gray-200";
    if (isOutOfTolerance) return "bg-red-50 text-red-700 border-red-200";
    return "bg-green-50 text-green-700 border-green-200";
  };

  const getBadgeIcon = () => {
    if (isSkipped) return "○ ";
    if (isOutOfTolerance) return "⚠️ ";
    return "✓ ";
  };

  return (
    <div className="mt-1">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge
              variant="outline"
              className={`text-xs cursor-help ${getBadgeStyle()}`}
            >
              {getBadgeIcon()}
              {isSkipped ? "Skip" : `Tol: ${tolerance.resolvedFrom}`}
            </Badge>
          </TooltipTrigger>
          <TooltipContent className="max-w-xs">
            <div className="space-y-1 text-xs">
              <p className="font-medium">Toleransi Pengiriman ({tolerance.resolvedFrom})</p>
              <p>{toleranceText}</p>
              {pendingQty > 0 && (
                <>
                  <p className="text-muted-foreground">
                    Range: {minAllowed.toFixed(0)} - {unlimitedOver ? "∞" : maxAllowed.toFixed(0)} unit
                  </p>
                  {isSkipped && (
                    <p className="text-gray-600">
                      ○ Item ini akan di-skip (tidak diterima dalam GR ini)
                    </p>
                  )}
                  {isUnderTolerance && (
                    <p className="text-red-600">
                      ⚠️ Qty {qtyToReceive} di bawah minimum ({minAllowed.toFixed(0)})
                    </p>
                  )}
                  {isOverTolerance && (
                    <p className="text-red-600">
                      ⚠️ Qty {qtyToReceive} melebihi maksimum ({maxAllowed.toFixed(0)})
                    </p>
                  )}
                </>
              )}
            </div>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}

interface GoodsReceiptItemFormData {
  purchaseOrderItemId: string;
  productId: string;
  productUnitId?: string;
  productName: string;
  productCode: string;
  orderedQty: string;
  receivedQty: string;
  pendingQty: string;
  isBatchTracked: boolean;
  isPerishable: boolean;
  batchNumber: string;
  manufactureDate: string;
  expiryDate: string;
  notes: string;
  // Calculated field for quantity to receive
  qtyToReceive: string;
}

interface FormData {
  grnDate: string;
  supplierInvoice: string;
  supplierDONumber: string;
  notes: string;
  items: GoodsReceiptItemFormData[];
}

export function CreateGoodsReceiptForm({
  purchaseOrder,
  onSuccess,
  onCancel,
}: CreateGoodsReceiptFormProps) {
  const router = useRouter();
  const [createGoodsReceipt, { isLoading }] = useCreateGoodsReceiptMutation();
  const [getEffectiveTolerance] = useLazyGetEffectiveToleranceQuery();
  const [isValidating, setIsValidating] = useState(false);

  // Filter items that still have pending quantity
  const pendingItems = (purchaseOrder.items || []).filter((item) => {
    const ordered = parseFloat(item.quantity);
    const received = parseFloat(item.receivedQty || "0");
    return received < ordered;
  });

  // Initialize form with today's date and pending items
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
    watch,
    setValue,
  } = useForm<FormData>({
    defaultValues: {
      grnDate: new Date().toISOString().split("T")[0],
      supplierInvoice: "",
      supplierDONumber: "",
      notes: "",
      items: pendingItems.map((item) => {
        const ordered = parseFloat(item.quantity);
        const received = parseFloat(item.receivedQty || "0");
        const pending = ordered - received;
        return {
          purchaseOrderItemId: item.id,
          productId: item.productId,
          productUnitId: item.productUnitId || undefined,
          productName: item.product?.name || "",
          productCode: item.product?.code || "",
          orderedQty: item.quantity,
          receivedQty: item.receivedQty || "0",
          pendingQty: pending.toString(),
          isBatchTracked: item.product?.isBatchTracked || false,
          isPerishable: item.product?.isPerishable || false,
          batchNumber: "",
          manufactureDate: "",
          expiryDate: "",
          notes: "",
          qtyToReceive: pending.toString(), // Default to full pending quantity
        };
      }),
    },
  });

  const { fields } = useFieldArray({
    control,
    name: "items",
  });

  const watchedItems = watch("items");

  const onSubmit = async (data: FormData) => {
    // Validate items
    const itemsToSubmit = data.items.filter(
      (item) => parseFloat(item.qtyToReceive) > 0
    );

    if (itemsToSubmit.length === 0) {
      toast.error("Tidak ada item yang akan diterima", {
        description: "Masukkan jumlah yang akan diterima minimal untuk satu item",
      });
      return;
    }

    // Validate qty doesn't exceed pending quantity
    for (const item of data.items) {
      const qtyToReceive = parseFloat(item.qtyToReceive) || 0;
      const pendingQty = parseFloat(item.pendingQty) || 0;

      if (qtyToReceive < 0) {
        toast.error("Jumlah Tidak Valid", {
          description: `Jumlah untuk "${item.productName}" tidak boleh negatif`,
        });
        return;
      }

      if (qtyToReceive > pendingQty) {
        toast.error("Jumlah Melebihi Batas", {
          description: `Jumlah untuk "${item.productName}" melebihi sisa qty yang belum diterima (maksimal ${pendingQty})`,
        });
        return;
      }
    }

    // Validate batch numbers and expiry dates for tracked products
    for (const item of itemsToSubmit) {
      if (item.isBatchTracked && !item.batchNumber.trim()) {
        toast.error("Nomor Batch Wajib Diisi", {
          description: `Produk "${item.productName}" memerlukan nomor batch`,
        });
        return;
      }
      if (item.isPerishable && !item.expiryDate) {
        toast.error("Tanggal Kadaluarsa Wajib Diisi", {
          description: `Produk "${item.productName}" memerlukan tanggal kadaluarsa`,
        });
        return;
      }
    }

    // Validate tolerance for each item before submitting
    setIsValidating(true);
    try {
      for (const item of itemsToSubmit) {
        const qtyToReceive = parseFloat(item.qtyToReceive) || 0;
        const pendingQty = parseFloat(item.pendingQty) || 0;

        // Fetch effective tolerance for this product
        const toleranceResult = await getEffectiveTolerance(item.productId).unwrap();

        if (toleranceResult) {
          const underTolerance = parseFloat(toleranceResult.underDeliveryTolerance) || 0;
          const overTolerance = parseFloat(toleranceResult.overDeliveryTolerance) || 0;
          const unlimitedOver = toleranceResult.unlimitedOverDelivery;

          // Calculate allowed range
          const minAllowed = pendingQty * (1 - underTolerance / 100);
          const maxAllowed = unlimitedOver ? Infinity : pendingQty * (1 + overTolerance / 100);

          // Check if under tolerance
          if (qtyToReceive < minAllowed) {
            setIsValidating(false);
            toast.error("Qty Di Bawah Toleransi", {
              description: `Produk "${item.productName}": qty ${qtyToReceive} di bawah minimum toleransi (min: ${minAllowed.toFixed(1)} unit). Toleransi under-delivery: ${underTolerance}%`,
            });
            return;
          }

          // Check if over tolerance (if not unlimited)
          if (!unlimitedOver && qtyToReceive > maxAllowed) {
            setIsValidating(false);
            toast.error("Qty Melebihi Toleransi", {
              description: `Produk "${item.productName}": qty ${qtyToReceive} melebihi maksimum toleransi (max: ${maxAllowed.toFixed(1)} unit). Toleransi over-delivery: ${overTolerance}%`,
            });
            return;
          }
        }
      }
    } catch (toleranceError) {
      console.error("Failed to validate tolerance:", toleranceError);
      // Continue with submission if tolerance check fails (graceful degradation)
      // Backend will still validate
    }
    setIsValidating(false);

    try {
      const request: CreateGoodsReceiptRequest = {
        purchaseOrderId: purchaseOrder.id,
        grnDate: data.grnDate,
        supplierInvoice: data.supplierInvoice || undefined,
        supplierDONumber: data.supplierDONumber || undefined,
        notes: data.notes || undefined,
        items: itemsToSubmit.map((item) => ({
          purchaseOrderItemId: item.purchaseOrderItemId,
          productId: item.productId,
          productUnitId: item.productUnitId,
          batchNumber: item.batchNumber || undefined,
          manufactureDate: item.manufactureDate || undefined,
          expiryDate: item.expiryDate || undefined,
          receivedQty: item.qtyToReceive,
          notes: item.notes || undefined,
        })),
      };

      await createGoodsReceipt(request).unwrap();

      toast.success("Penerimaan Barang Dibuat", {
        description: "Data penerimaan barang berhasil disimpan",
      });

      if (onSuccess) {
        onSuccess();
      } else {
        router.push("/procurement/receipts");
      }
    } catch (error: any) {
      console.error("Failed to create goods receipt:", error);
      toast.error("Gagal Membuat Penerimaan", {
        description:
          error?.data?.error?.message ||
          "Terjadi kesalahan saat membuat penerimaan barang",
      });
    }
  };

  if (pendingItems.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col items-center justify-center py-8 text-center">
            <Package className="mb-4 h-12 w-12 text-muted-foreground" />
            <h3 className="mb-2 text-lg font-semibold">
              Semua Item Sudah Diterima
            </h3>
            <p className="mb-4 text-sm text-muted-foreground">
              Tidak ada item yang tersisa untuk diterima dari PO ini
            </p>
            <Button variant="outline" onClick={onCancel}>
              Kembali
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Header Info */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Package className="h-4 w-4" />
            Informasi Penerimaan
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            {/* GRN Date */}
            <div className="space-y-2">
              <Label htmlFor="grnDate">
                Tanggal Penerimaan <span className="text-destructive">*</span>
              </Label>
              <Input
                id="grnDate"
                type="date"
                {...register("grnDate", {
                  required: "Tanggal penerimaan wajib diisi",
                })}
                className={errors.grnDate ? "border-destructive" : ""}
              />
              {errors.grnDate && (
                <p className="flex items-center gap-1 text-sm text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  {errors.grnDate.message}
                </p>
              )}
            </div>

            {/* Supplier Invoice */}
            <div className="space-y-2">
              <Label htmlFor="supplierInvoice">No. Invoice Supplier</Label>
              <Input
                id="supplierInvoice"
                {...register("supplierInvoice")}
                placeholder="Opsional"
              />
            </div>

            {/* Supplier DO Number */}
            <div className="space-y-2">
              <Label htmlFor="supplierDONumber">No. Surat Jalan Supplier</Label>
              <Input
                id="supplierDONumber"
                {...register("supplierDONumber")}
                placeholder="Opsional"
              />
            </div>
          </div>

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes">Catatan</Label>
            <Textarea
              id="notes"
              {...register("notes")}
              placeholder="Catatan tambahan (opsional)"
              rows={2}
            />
          </div>
        </CardContent>
      </Card>

      {/* Items Table */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Package className="h-4 w-4" />
            Item Barang
          </CardTitle>
          <p className="text-sm text-muted-foreground">
            Masukkan jumlah yang diterima dan informasi batch/kadaluarsa jika diperlukan
          </p>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[50px]">No</TableHead>
                  <TableHead className="min-w-[200px]">Produk</TableHead>
                  <TableHead className="text-right min-w-[100px]">
                    Qty Order
                  </TableHead>
                  <TableHead className="text-right min-w-[100px]">
                    Sudah Diterima
                  </TableHead>
                  <TableHead className="text-right min-w-[120px]">
                    Qty Diterima
                    <span className="text-destructive ml-1">*</span>
                  </TableHead>
                  <TableHead className="min-w-[150px]">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="flex items-center gap-1 cursor-help">
                            No. Batch
                            <Info className="h-3 w-3 text-muted-foreground" />
                          </span>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Wajib untuk produk dengan pelacakan batch</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </TableHead>
                  <TableHead className="min-w-[150px]">Tgl Produksi</TableHead>
                  <TableHead className="min-w-[150px]">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span className="flex items-center gap-1 cursor-help">
                            Tgl Kadaluarsa
                            <Info className="h-3 w-3 text-muted-foreground" />
                          </span>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Wajib untuk produk perishable (mudah rusak)</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {fields.map((field, index) => {
                  const item = watchedItems[index];
                  const pendingQty = parseFloat(item?.pendingQty || "0");

                  return (
                    <TableRow key={field.id}>
                      <TableCell className="text-muted-foreground align-top">
                        {index + 1}
                      </TableCell>
                      <TableCell className="align-top">
                        <div>
                          <div className="font-medium flex items-center gap-2">
                            {item?.productName}
                            {item?.isBatchTracked && (
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger>
                                    <Badge
                                      variant="outline"
                                      className="text-xs bg-blue-50 text-blue-700 border-blue-200"
                                    >
                                      <Hash className="h-3 w-3 mr-1" />
                                      Batch
                                    </Badge>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>Produk ini memerlukan nomor batch</p>
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                            )}
                            {item?.isPerishable && (
                              <TooltipProvider>
                                <Tooltip>
                                  <TooltipTrigger>
                                    <Badge
                                      variant="outline"
                                      className="text-xs bg-orange-50 text-orange-700 border-orange-200"
                                    >
                                      <Calendar className="h-3 w-3 mr-1" />
                                      Exp
                                    </Badge>
                                  </TooltipTrigger>
                                  <TooltipContent>
                                    <p>Produk ini memerlukan tanggal kadaluarsa</p>
                                  </TooltipContent>
                                </Tooltip>
                              </TooltipProvider>
                            )}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {item?.productCode}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-right align-top">
                        {parseFloat(item?.orderedQty || "0").toLocaleString(
                          "id-ID"
                        )}
                      </TableCell>
                      <TableCell className="text-right align-top">
                        {parseFloat(item?.receivedQty || "0").toLocaleString(
                          "id-ID"
                        )}
                      </TableCell>
                      <TableCell className="align-top">
                        <div className="space-y-1">
                          <Input
                            type="number"
                            step="0.001"
                            min="0"
                            {...register(`items.${index}.qtyToReceive`, {
                              required: "Wajib diisi",
                              min: { value: 0, message: "Tidak boleh negatif" },
                            })}
                            className="w-24 text-right"
                            placeholder="0"
                          />
                          <ToleranceInfo
                            productId={item?.productId || ""}
                            pendingQty={pendingQty}
                            qtyToReceive={parseFloat(item?.qtyToReceive || "0")}
                          />
                        </div>
                      </TableCell>
                      <TableCell className="align-top">
                        <div className="space-y-1">
                          <Input
                            {...register(`items.${index}.batchNumber`)}
                            placeholder={
                              item?.isBatchTracked ? "Wajib diisi" : "Opsional"
                            }
                            className={
                              item?.isBatchTracked
                                ? "border-blue-300 focus:border-blue-500"
                                : ""
                            }
                          />
                          {item?.isBatchTracked && (
                            <p className="text-xs text-blue-600">
                              Wajib <span className="text-destructive">*</span>
                            </p>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="align-top">
                        <Input
                          type="date"
                          {...register(`items.${index}.manufactureDate`)}
                        />
                      </TableCell>
                      <TableCell className="align-top">
                        <div className="space-y-1">
                          <Input
                            type="date"
                            {...register(`items.${index}.expiryDate`)}
                            className={
                              item?.isPerishable
                                ? "border-orange-300 focus:border-orange-500"
                                : ""
                            }
                          />
                          {item?.isPerishable && (
                            <p className="text-xs text-orange-600">
                              Wajib <span className="text-destructive">*</span>
                            </p>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Legend */}
      <Card className="border-dashed">
        <CardHeader className="pb-3 pt-4">
          <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
            <Info className="h-4 w-4" />
            Keterangan Indikator
          </CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            {/* Tracking Requirements */}
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">
                Persyaratan Pelacakan
              </p>
              <div className="flex items-start gap-3 p-2 rounded-md bg-blue-50/50">
                <Badge
                  variant="outline"
                  className="text-xs bg-blue-50 text-blue-700 border-blue-200 shrink-0"
                >
                  <Hash className="h-3 w-3 mr-1" />
                  Batch
                </Badge>
                <span className="text-muted-foreground text-xs">
                  Produk memerlukan nomor batch untuk pelacakan lot
                </span>
              </div>
              <div className="flex items-start gap-3 p-2 rounded-md bg-orange-50/50">
                <Badge
                  variant="outline"
                  className="text-xs bg-orange-50 text-orange-700 border-orange-200 shrink-0"
                >
                  <Calendar className="h-3 w-3 mr-1" />
                  Exp
                </Badge>
                <span className="text-muted-foreground text-xs">
                  Produk mudah rusak, wajib mencantumkan tanggal kadaluarsa
                </span>
              </div>
            </div>

            {/* Tolerance Status */}
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">
                Status Toleransi Pengiriman
              </p>
              <div className="flex items-start gap-3 p-2 rounded-md bg-gray-50/50">
                <Badge
                  variant="outline"
                  className="text-xs bg-gray-50 text-gray-600 border-gray-200 shrink-0"
                >
                  ○ Skip
                </Badge>
                <span className="text-muted-foreground text-xs">
                  Item tidak diterima dalam penerimaan ini (qty = 0)
                </span>
              </div>
              <div className="flex items-start gap-3 p-2 rounded-md bg-green-50/50">
                <Badge
                  variant="outline"
                  className="text-xs bg-green-50 text-green-700 border-green-200 shrink-0"
                >
                  ✓ Tol
                </Badge>
                <span className="text-muted-foreground text-xs">
                  Qty dalam batas toleransi yang diizinkan
                </span>
              </div>
              <div className="flex items-start gap-3 p-2 rounded-md bg-red-50/50">
                <Badge
                  variant="outline"
                  className="text-xs bg-red-50 text-red-700 border-red-200 shrink-0"
                >
                  ⚠️ Tol
                </Badge>
                <span className="text-muted-foreground text-xs">
                  Qty di luar toleransi — tidak dapat disimpan
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Form Actions */}
      <div className="flex justify-end gap-3">
        {onCancel && (
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
          >
            <X className="mr-2 h-4 w-4" />
            Batal
          </Button>
        )}
        <Button type="submit" disabled={isLoading || isValidating}>
          {isLoading || isValidating ? (
            <>
              <span className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              {isValidating ? "Memvalidasi..." : "Menyimpan..."}
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Simpan Penerimaan
            </>
          )}
        </Button>
      </div>
    </form>
  );
}
