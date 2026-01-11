/**
 * Edit Stock Opname Page - Client Component
 *
 * Page for editing existing stock opname.
 * Only draft and in_progress status can be edited.
 */

"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { ArrowLeft, Save, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { useToast } from "@/hooks/use-toast";
import {
  useGetOpnameQuery,
  useUpdateOpnameMutation,
  useUpdateOpnameItemMutation,
} from "@/store/services/opnameApi";
import type { StockOpnameItem } from "@/types/opname.types";

interface OpnameItemForm extends StockOpnameItem {
  isDirty?: boolean;
}

export default function EditOpnamePage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const opnameId = params.id as string;

  const [opnameDate, setOpnameDate] = useState<string>("");
  const [notes, setNotes] = useState<string>("");
  const [status, setStatus] = useState<"draft" | "in_progress" | "completed">("draft");
  const [items, setItems] = useState<OpnameItemForm[]>([]);

  const { data: opname, isLoading, error } = useGetOpnameQuery(opnameId);
  const [updateOpname, { isLoading: isUpdatingOpname }] = useUpdateOpnameMutation();
  const [updateItem, { isLoading: isUpdatingItem }] = useUpdateOpnameItemMutation();

  // Populate form when data loads
  useEffect(() => {
    if (opname) {
      // Check if opname can be edited
      if (opname.status !== "draft" && opname.status !== "in_progress") {
        toast({
          title: "Tidak Dapat Diedit",
          description: "Stock opname dengan status ini tidak dapat diedit",
          variant: "destructive",
        });
        router.push(`/inventory/opname/${opnameId}`);
        return;
      }

      setOpnameDate(opname.opnameDate.split("T")[0]);
      setNotes(opname.notes || "");
      setStatus(opname.status as "draft" | "in_progress" | "completed");
      setItems(opname.items || []);
    }
  }, [opname, opnameId, router, toast]);

  const handleActualQtyChange = (index: number, value: string) => {
    const newItems = [...items];
    newItems[index].actualQty = value;
    newItems[index].difference = (
      parseFloat(value || "0") - parseFloat(newItems[index].expectedQty || "0")
    ).toString();
    newItems[index].isDirty = true;
    setItems(newItems);
  };

  const handleNotesChange = (index: number, value: string) => {
    const newItems = [...items];
    newItems[index].notes = value || undefined;
    newItems[index].isDirty = true;
    setItems(newItems);
  };

  const handleSubmit = async () => {
    try {
      // Update main opname info
      await updateOpname({
        id: opnameId,
        data: {
          opnameDate,
          notes,
          status,
        },
      }).unwrap();

      // Update items that have changed
      const dirtyItems = items.filter((item) => item.isDirty);
      for (const item of dirtyItems) {
        await updateItem({
          opnameId,
          itemId: item.id,
          data: {
            actualQty: item.actualQty,
            notes: item.notes || undefined,
          },
        }).unwrap();
      }

      toast({
        title: "Berhasil",
        description: "Stock opname berhasil diupdate",
      });

      router.push(`/inventory/opname/${opnameId}`);
    } catch (error) {
      console.error("Failed to update opname:", error);
      toast({
        title: "Gagal",
        description: "Gagal mengupdate stock opname",
        variant: "destructive",
      });
    }
  };

  const formatNumber = (value: string | number) => {
    const num = typeof value === "string" ? parseFloat(value) : value;
    return num.toLocaleString("id-ID", {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    });
  };

  if (isLoading) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname", href: "/inventory/opname" },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 items-center justify-center min-h-[400px]">
          <LoadingSpinner size="lg" text="Memuat data stock opname..." />
        </div>
      </div>
    );
  }

  if (error || !opname) {
    return (
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Inventori", href: "/inventory/stock" },
            { label: "Stock Opname", href: "/inventory/opname" },
            { label: "Edit" },
          ]}
        />
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          <ErrorDisplay
            error={error}
            title="Gagal memuat data stock opname"
            onRetry={() => router.push("/inventory/opname")}
          />
        </div>
      </div>
    );
  }

  const isProcessing = isUpdatingOpname || isUpdatingItem;

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Inventori", href: "/inventory/stock" },
          { label: "Stock Opname", href: "/inventory/opname" },
          { label: "Edit" },
        ]}
      />

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page Header */}
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl font-bold tracking-tight">
              Edit Stock Opname
            </h1>
            <p className="text-muted-foreground">
              {opname.opnameNumber} • {opname.warehouseName}
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => router.push(`/inventory/opname/${opnameId}`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Kembali
          </Button>
        </div>

        {/* Form Card */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Stock Opname</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="status">Status</Label>
                <Select
                  value={status}
                  onValueChange={(value: "draft" | "in_progress" | "completed") =>
                    setStatus(value)
                  }
                >
                  <SelectTrigger id="status" className="bg-background w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="draft">Draft</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="opnameDate">
                  Tanggal Opname <span className="text-destructive">*</span>
                </Label>
                <Input
                  id="opnameDate"
                  type="date"
                  value={opnameDate}
                  onChange={(e) => setOpnameDate(e.target.value)}
                  max={new Date().toISOString().split("T")[0]}
                  className="bg-background"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="notes">Catatan</Label>
              <Textarea
                id="notes"
                placeholder="Catatan untuk stock opname ini..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
                className="bg-background"
              />
            </div>
          </CardContent>
        </Card>

        {/* Items Table */}
        <Card>
          <CardHeader>
            <CardTitle>Produk ({items.length})</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[100px]">Kode</TableHead>
                    <TableHead>Nama Produk</TableHead>
                    <TableHead className="text-right">Expected Qty</TableHead>
                    <TableHead className="text-right w-[150px]">
                      Actual Qty <span className="text-destructive">*</span>
                    </TableHead>
                    <TableHead className="text-right">Selisih</TableHead>
                    <TableHead className="w-[200px]">Catatan</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item, index) => {
                    const difference = parseFloat(item.difference || "0");
                    return (
                      <TableRow key={item.id}>
                        <TableCell className="font-mono text-sm">
                          {item.productCode}
                        </TableCell>
                        <TableCell>{item.productName}</TableCell>
                        <TableCell className="text-right">
                          {formatNumber(item.expectedQty)}
                        </TableCell>
                        <TableCell className="text-right">
                          <Input
                            type="number"
                            step="0.01"
                            min="0"
                            value={item.actualQty}
                            onChange={(e) =>
                              handleActualQtyChange(index, e.target.value)
                            }
                            className="text-right bg-background"
                          />
                        </TableCell>
                        <TableCell className="text-right">
                          <span
                            className={
                              difference > 0
                                ? "text-green-600 font-semibold"
                                : difference < 0
                                ? "text-red-600 font-semibold"
                                : "text-muted-foreground"
                            }
                          >
                            {difference > 0 && "+"}
                            {formatNumber(difference)}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Input
                            type="text"
                            value={item.notes || ""}
                            onChange={(e) =>
                              handleNotesChange(index, e.target.value)
                            }
                            placeholder="Catatan..."
                            className="bg-background"
                          />
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex justify-end gap-2">
          <Button
            variant="outline"
            onClick={() => router.push(`/inventory/opname/${opnameId}`)}
            disabled={isProcessing}
          >
            Batal
          </Button>
          <Button onClick={handleSubmit} disabled={isProcessing}>
            {isProcessing ? (
              <>
                <LoadingSpinner size="sm" className="mr-2" />
                Menyimpan...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Simpan Perubahan
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
