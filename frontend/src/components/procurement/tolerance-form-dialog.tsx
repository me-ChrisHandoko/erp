/**
 * Tolerance Form Dialog
 *
 * Dialog component for creating/editing delivery tolerance settings.
 * Supports all three levels: COMPANY, CATEGORY, and PRODUCT.
 */

"use client";

import { useState, useEffect } from "react";
import { Settings, Building2, FolderTree, Package } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { toast } from "sonner";
import {
  useCreateDeliveryToleranceMutation,
  useUpdateDeliveryToleranceMutation,
} from "@/store/services/deliveryToleranceApi";
import { useListProductsQuery } from "@/store/services/productApi";
import {
  DELIVERY_TOLERANCE_LEVEL_OPTIONS,
  type DeliveryToleranceLevel,
  type DeliveryToleranceResponse,
  type CreateDeliveryToleranceRequest,
  type UpdateDeliveryToleranceRequest,
} from "@/types/delivery-tolerance.types";

interface ToleranceFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: "create" | "edit";
  tolerance?: DeliveryToleranceResponse | null;
}

export function ToleranceFormDialog({
  open,
  onOpenChange,
  mode,
  tolerance,
}: ToleranceFormDialogProps) {
  // Form state
  const [level, setLevel] = useState<DeliveryToleranceLevel>("COMPANY");
  const [categoryName, setCategoryName] = useState("");
  const [productId, setProductId] = useState("");
  const [underDeliveryTolerance, setUnderDeliveryTolerance] = useState("0");
  const [overDeliveryTolerance, setOverDeliveryTolerance] = useState("0");
  const [unlimitedOverDelivery, setUnlimitedOverDelivery] = useState(false);
  const [isActive, setIsActive] = useState(true);
  const [notes, setNotes] = useState("");

  // Mutations
  const [createTolerance, { isLoading: isCreating }] =
    useCreateDeliveryToleranceMutation();
  const [updateTolerance, { isLoading: isUpdating }] =
    useUpdateDeliveryToleranceMutation();

  // Fetch products for product-level tolerance
  const { data: productsData } = useListProductsQuery(
    { pageSize: 100 },
    { skip: level !== "PRODUCT" }
  );

  // Reset form when dialog opens or tolerance changes
  useEffect(() => {
    if (open) {
      if (mode === "edit" && tolerance) {
        setLevel(tolerance.level);
        setCategoryName(tolerance.categoryName || "");
        setProductId(tolerance.productId || "");
        setUnderDeliveryTolerance(tolerance.underDeliveryTolerance);
        setOverDeliveryTolerance(tolerance.overDeliveryTolerance);
        setUnlimitedOverDelivery(tolerance.unlimitedOverDelivery);
        setIsActive(tolerance.isActive);
        setNotes(tolerance.notes || "");
      } else {
        setLevel("COMPANY");
        setCategoryName("");
        setProductId("");
        setUnderDeliveryTolerance("0");
        setOverDeliveryTolerance("0");
        setUnlimitedOverDelivery(false);
        setIsActive(true);
        setNotes("");
      }
    }
  }, [open, mode, tolerance]);

  const handleSubmit = async () => {
    // Validate
    if (level === "CATEGORY" && !categoryName.trim()) {
      toast.error("Nama kategori wajib diisi untuk level Kategori");
      return;
    }
    if (level === "PRODUCT" && !productId) {
      toast.error("Produk wajib dipilih untuk level Produk");
      return;
    }

    try {
      if (mode === "create") {
        const createData: CreateDeliveryToleranceRequest = {
          level,
          underDeliveryTolerance,
          overDeliveryTolerance,
          unlimitedOverDelivery,
          isActive,
          notes: notes.trim() || undefined,
        };
        if (level === "CATEGORY") {
          createData.categoryName = categoryName.trim();
        }
        if (level === "PRODUCT") {
          createData.productId = productId;
        }

        await createTolerance(createData).unwrap();
        toast.success("Toleransi Dibuat", {
          description: "Pengaturan toleransi berhasil dibuat",
        });
      } else if (mode === "edit" && tolerance) {
        const updateData: UpdateDeliveryToleranceRequest = {
          underDeliveryTolerance,
          overDeliveryTolerance,
          unlimitedOverDelivery,
          isActive,
          notes: notes.trim() || undefined,
        };

        await updateTolerance({ id: tolerance.id, data: updateData }).unwrap();
        toast.success("Toleransi Diperbarui", {
          description: "Pengaturan toleransi berhasil diperbarui",
        });
      }

      onOpenChange(false);
    } catch (error: any) {
      toast.error(mode === "create" ? "Gagal Membuat" : "Gagal Memperbarui", {
        description: error?.data?.error?.message || "Terjadi kesalahan",
      });
    }
  };

  const isLoading = isCreating || isUpdating;

  const getLevelIcon = (lvl: DeliveryToleranceLevel) => {
    switch (lvl) {
      case "COMPANY":
        return <Building2 className="h-4 w-4" />;
      case "CATEGORY":
        return <FolderTree className="h-4 w-4" />;
      case "PRODUCT":
        return <Package className="h-4 w-4" />;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            {mode === "create"
              ? "Tambah Toleransi Pengiriman"
              : "Edit Toleransi Pengiriman"}
          </DialogTitle>
          <DialogDescription>
            {mode === "create"
              ? "Buat pengaturan toleransi pengiriman baru"
              : "Perbarui pengaturan toleransi pengiriman"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4 overflow-y-auto flex-1">
          {/* Level Selection (only for create) */}
          {mode === "create" && (
            <div className="space-y-2">
              <Label htmlFor="level">
                Level Toleransi <span className="text-destructive">*</span>
              </Label>
              <Select
                value={level}
                onValueChange={(val) => setLevel(val as DeliveryToleranceLevel)}
              >
                <SelectTrigger id="level">
                  <SelectValue placeholder="Pilih level..." />
                </SelectTrigger>
                <SelectContent>
                  {DELIVERY_TOLERANCE_LEVEL_OPTIONS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      <div className="flex items-center gap-2">
                        {getLevelIcon(option.value)}
                        <div>
                          <div>{option.label}</div>
                          <div className="text-xs text-muted-foreground">
                            {option.description}
                          </div>
                        </div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Category Name (for CATEGORY level) */}
          {level === "CATEGORY" && mode === "create" && (
            <div className="space-y-2">
              <Label htmlFor="categoryName">
                Nama Kategori <span className="text-destructive">*</span>
              </Label>
              <Input
                id="categoryName"
                value={categoryName}
                onChange={(e) => setCategoryName(e.target.value)}
                placeholder="Masukkan nama kategori produk..."
              />
            </div>
          )}

          {/* Product Selection (for PRODUCT level) */}
          {level === "PRODUCT" && mode === "create" && (
            <div className="space-y-2">
              <Label htmlFor="productId">
                Produk <span className="text-destructive">*</span>
              </Label>
              <Select value={productId} onValueChange={setProductId}>
                <SelectTrigger id="productId">
                  <SelectValue placeholder="Pilih produk..." />
                </SelectTrigger>
                <SelectContent>
                  {productsData?.data.map((product) => (
                    <SelectItem key={product.id} value={product.id}>
                      <div>
                        <div>{product.name}</div>
                        <div className="text-xs text-muted-foreground">
                          {product.code}
                        </div>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Under Delivery Tolerance */}
          <div className="space-y-2">
            <Label htmlFor="underTolerance">
              Toleransi Kurang (%) <span className="text-destructive">*</span>
            </Label>
            <div className="flex items-center gap-2">
              <span className="text-red-600">-</span>
              <Input
                id="underTolerance"
                type="number"
                min="0"
                max="100"
                step="0.01"
                value={underDeliveryTolerance}
                onChange={(e) => setUnderDeliveryTolerance(e.target.value)}
                placeholder="0.00"
                className="font-mono"
              />
              <span className="text-muted-foreground">%</span>
            </div>
            <p className="text-xs text-muted-foreground">
              Persentase maksimal pengiriman kurang yang diizinkan
            </p>
          </div>

          {/* Over Delivery Tolerance */}
          <div className="space-y-2">
            <Label htmlFor="overTolerance">Toleransi Lebih (%)</Label>
            <div className="flex items-center gap-2">
              <span className="text-green-600">+</span>
              <Input
                id="overTolerance"
                type="number"
                min="0"
                max="100"
                step="0.01"
                value={overDeliveryTolerance}
                onChange={(e) => setOverDeliveryTolerance(e.target.value)}
                placeholder="0.00"
                className="font-mono"
                disabled={unlimitedOverDelivery}
              />
              <span className="text-muted-foreground">%</span>
            </div>
            <p className="text-xs text-muted-foreground">
              Persentase maksimal pengiriman lebih yang diizinkan
            </p>
          </div>

          {/* Unlimited Over Delivery */}
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="space-y-0.5">
              <Label htmlFor="unlimitedOver">Pengiriman Lebih Tidak Terbatas</Label>
              <p className="text-xs text-muted-foreground">
                Izinkan pengiriman lebih dari jumlah pesanan tanpa batas
              </p>
            </div>
            <Switch
              id="unlimitedOver"
              checked={unlimitedOverDelivery}
              onCheckedChange={setUnlimitedOverDelivery}
            />
          </div>

          {/* Is Active */}
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="space-y-0.5">
              <Label htmlFor="isActive">Status Aktif</Label>
              <p className="text-xs text-muted-foreground">
                Toleransi hanya berlaku jika status aktif
              </p>
            </div>
            <Switch
              id="isActive"
              checked={isActive}
              onCheckedChange={setIsActive}
            />
          </div>

          {/* Notes */}
          <div className="space-y-2">
            <Label htmlFor="notes">Catatan (Opsional)</Label>
            <Textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Catatan tambahan tentang toleransi ini..."
              rows={2}
            />
          </div>

          {/* Preview */}
          <Alert>
            <AlertDescription>
              <strong>Preview:</strong> Dengan pengaturan ini, pesanan 100 unit
              dapat dikirim antara{" "}
              <span className="text-red-600 font-mono">
                {(100 - parseFloat(underDeliveryTolerance || "0")).toFixed(0)}
              </span>{" "}
              hingga{" "}
              <span className="text-green-600 font-mono">
                {unlimitedOverDelivery
                  ? "∞"
                  : (100 + parseFloat(overDeliveryTolerance || "0")).toFixed(0)}
              </span>{" "}
              unit.
            </AlertDescription>
          </Alert>
        </div>

        <DialogFooter className="shrink-0">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={isLoading}>
            {isLoading ? "Menyimpan..." : mode === "create" ? "Buat" : "Simpan"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
