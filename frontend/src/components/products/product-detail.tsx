/**
 * Product Detail Component
 *
 * Comprehensive product information display with:
 * - Basic info card (code, name, category, description)
 * - Pricing card (cost, price, margin, stock)
 * - Units table (conversion rates, prices, barcodes)
 * - Suppliers table (supplier info, pricing, lead time)
 * - Attributes card (batch tracking, perishability, status)
 */

"use client";

import {
  Package,
  DollarSign,
  Layers,
  TrendingUp,
  Building2,
  Info,
  Calendar,
  Tag,
  ShoppingCart,
  Warehouse,
  Scale,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { ProductResponse } from "@/types/product.types";

interface ProductDetailProps {
  product: ProductResponse;
}

export function ProductDetail({ product }: ProductDetailProps) {
  // Calculate profit margin
  const baseCost = Number(product.baseCost);
  const basePrice = Number(product.basePrice);
  const margin = ((basePrice - baseCost) / baseCost) * 100;
  const profit = basePrice - baseCost;

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Basic Information Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Informasi Dasar
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            {/* Product Code */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Kode Produk
              </p>
              <p className="font-mono text-lg font-semibold">{product.code}</p>
            </div>

            {/* Product Name */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Nama Produk
              </p>
              <p className="text-lg font-semibold">{product.name}</p>
            </div>

            {/* Category */}
            {product.category && (
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Kategori
                </p>
                <Badge className="bg-blue-500 text-white hover:bg-blue-600 text-sm">
                  <Tag className="mr-1 h-3 w-3" />
                  {product.category}
                </Badge>
              </div>
            )}

            {/* Base Unit */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Satuan Dasar
              </p>
              <Badge variant="secondary" className="text-sm">
                <Scale className="mr-1 h-3 w-3" />
                {product.baseUnit}
              </Badge>
            </div>

            {/* Barcode */}
            {product.barcode && (
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Barcode
                </p>
                <p className="font-mono text-sm">{product.barcode}</p>
              </div>
            )}

            {/* Status */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Status
              </p>
              <Badge
                className={
                  product.isActive
                    ? "bg-green-500 text-white hover:bg-green-600"
                    : "bg-red-500 text-white hover:bg-red-600"
                }
              >
                {product.isActive ? (
                  <CheckCircle2 className="mr-1 h-3 w-3" />
                ) : (
                  <XCircle className="mr-1 h-3 w-3" />
                )}
                {product.isActive ? "Aktif" : "Nonaktif"}
              </Badge>
            </div>
          </div>

          {/* Description */}
          {product.description && (
            <>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Deskripsi
                </p>
                <p className="text-sm leading-relaxed">{product.description}</p>
              </div>
            </>
          )}

          {/* Timestamps */}
          <Separator />
          <div className="grid gap-4 text-xs text-muted-foreground md:grid-cols-2">
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Dibuat: {new Date(product.createdAt).toLocaleDateString("id-ID")}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Diperbarui:{" "}
                {new Date(product.updatedAt).toLocaleDateString("id-ID")}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Pricing Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Harga & Margin
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Base Cost */}
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Harga Beli
              </p>
              <p className="text-2xl font-bold">
                Rp {baseCost.toLocaleString("id-ID")}
              </p>
              <p className="text-xs text-muted-foreground">
                per {product.baseUnit}
              </p>
            </div>
          </div>

          <Separator />

          {/* Base Price */}
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Harga Jual
              </p>
              <p className="text-2xl font-bold text-blue-600">
                Rp {basePrice.toLocaleString("id-ID")}
              </p>
              <p className="text-xs text-muted-foreground">
                per {product.baseUnit}
              </p>
            </div>
          </div>

          <Separator />

          {/* Margin & Profit */}
          <div className="space-y-3 rounded-lg bg-muted/50 p-4">
            <div className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-green-600" />
              <p className="text-sm font-medium">Profit Margin</p>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-muted-foreground">Margin</p>
                <p className="text-lg font-bold text-green-600">
                  {margin.toFixed(2)}%
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Profit</p>
                <p className="text-lg font-bold text-green-600">
                  Rp {profit.toLocaleString("id-ID")}
                </p>
              </div>
            </div>
          </div>

          <Separator />

          {/* Minimum Stock */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">
              Stok Minimum
            </p>
            <p className="text-lg font-semibold">
              {Number(product.minimumStock).toLocaleString("id-ID")}{" "}
              {product.baseUnit}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Attributes Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            Atribut Produk
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Batch Tracking */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">Pelacakan Batch</p>
            {product.isBatchTracked ? (
              <Badge variant="default" className="bg-green-500">
                <CheckCircle2 className="mr-1 h-3 w-3" />
                Aktif
              </Badge>
            ) : (
              <Badge variant="secondary">
                <XCircle className="mr-1 h-3 w-3" />
                Nonaktif
              </Badge>
            )}
          </div>

          <Separator />

          {/* Perishable */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">Produk Mudah Rusak</p>
            {product.isPerishable ? (
              <Badge variant="default" className="bg-amber-500">
                <CheckCircle2 className="mr-1 h-3 w-3" />
                Ya
              </Badge>
            ) : (
              <Badge variant="secondary">
                <XCircle className="mr-1 h-3 w-3" />
                Tidak
              </Badge>
            )}
          </div>

          <Separator />

          {/* Total Units */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">Total Satuan</p>
            <Badge variant="outline">
              <Layers className="mr-1 h-3 w-3" />
              {product.units?.length || 1} satuan
            </Badge>
          </div>

          {/* Total Suppliers */}
          {product.suppliers && product.suppliers.length > 0 && (
            <>
              <Separator />
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Total Supplier</p>
                <Badge variant="outline">
                  <Building2 className="mr-1 h-3 w-3" />
                  {product.suppliers.length} supplier
                </Badge>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Units Table */}
      {product.units && product.units.length > 0 && (
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Layers className="h-5 w-5" />
              Satuan Produk
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Nama Satuan</TableHead>
                    <TableHead className="text-center">Base Unit</TableHead>
                    <TableHead className="text-right">Konversi</TableHead>
                    <TableHead className="text-right">Harga Beli</TableHead>
                    <TableHead className="text-right">Harga Jual</TableHead>
                    <TableHead>Barcode</TableHead>
                    <TableHead className="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {product.units.map((unit) => (
                    <TableRow key={unit.id}>
                      <TableCell className="font-medium">
                        {unit.unitName}
                      </TableCell>
                      <TableCell className="text-center">
                        {unit.isBaseUnit && (
                          <Badge variant="default" className="bg-blue-500">
                            Base
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        1 = {unit.conversionRate} {product.baseUnit}
                      </TableCell>
                      <TableCell className="text-right">
                        {unit.buyPrice
                          ? `Rp ${Number(unit.buyPrice).toLocaleString("id-ID")}`
                          : "-"}
                      </TableCell>
                      <TableCell className="text-right">
                        {unit.sellPrice
                          ? `Rp ${Number(unit.sellPrice).toLocaleString("id-ID")}`
                          : "-"}
                      </TableCell>
                      <TableCell className="font-mono text-sm">
                        {unit.barcode || "-"}
                      </TableCell>
                      <TableCell className="text-center">
                        {unit.isActive ? (
                          <Badge
                            variant="default"
                            className="bg-green-500 text-xs"
                          >
                            Aktif
                          </Badge>
                        ) : (
                          <Badge variant="secondary" className="text-xs">
                            Nonaktif
                          </Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Suppliers Table */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Supplier Produk
          </CardTitle>
        </CardHeader>
        <CardContent>
          {product.suppliers && product.suppliers.length > 0 ? (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Nama Supplier</TableHead>
                    <TableHead className="text-right">Harga Supplier</TableHead>
                    <TableHead className="text-right">Lead Time</TableHead>
                    <TableHead className="text-right">MOQ</TableHead>
                    <TableHead className="text-center">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {product.suppliers.map((supplier) => (
                    <TableRow key={supplier.id}>
                      <TableCell className="font-medium">
                        <div>
                          <span>{supplier.supplierCode} - {supplier.supplierName}</span>
                          {supplier.isPrimarySupplier && (
                            <Badge
                              variant="default"
                              className="ml-2 bg-blue-500 text-xs"
                            >
                              Utama
                            </Badge>
                          )}
                        </div>
                        {supplier.supplierProductCode && (
                          <p className="text-xs text-muted-foreground">
                            Kode: {supplier.supplierProductCode}
                          </p>
                        )}
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        Rp{" "}
                        {Number(supplier.supplierPrice).toLocaleString("id-ID")}
                      </TableCell>
                      <TableCell className="text-right">
                        {supplier.leadTimeDays || "-"} hari
                      </TableCell>
                      <TableCell className="text-right">
                        {supplier.minimumOrderQty || "-"}
                      </TableCell>
                      <TableCell className="text-center">
                        <Badge variant="outline" className="text-xs">
                          <ShoppingCart className="mr-1 h-3 w-3" />
                          Aktif
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Building2 className="mx-auto h-12 w-12 mb-4 opacity-50" />
              <p>Belum ada supplier terhubung</p>
              <p className="text-sm">Gunakan tombol "Edit Produk" untuk mengelola supplier</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Stock Information */}
      {product.currentStock && (
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Warehouse className="h-5 w-5" />
              Informasi Stok
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Total Stock */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Total Stok
              </p>
              <p className="text-2xl font-bold">
                {Number(product.currentStock?.totalStock || 0).toLocaleString("id-ID")}{" "}
                {product.baseUnit}
              </p>
            </div>

            {/* Warehouse Breakdown */}
            {product.currentStock.warehouses &&
              product.currentStock.warehouses.length > 0 && (
                <>
                  <Separator />
                  <div className="space-y-2">
                    <p className="text-sm font-medium">Per Gudang</p>
                    <div className="rounded-md border">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Nama Gudang</TableHead>
                            <TableHead className="text-right">
                              Jumlah Stok
                            </TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {product.currentStock.warehouses.map((warehouse) => (
                            <TableRow key={warehouse.warehouseId}>
                              <TableCell className="font-medium">
                                {warehouse.warehouseName}
                              </TableCell>
                              <TableCell className="text-right">
                                {Number(warehouse.quantity).toLocaleString(
                                  "id-ID"
                                )}{" "}
                                {product.baseUnit}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  </div>
                </>
              )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
