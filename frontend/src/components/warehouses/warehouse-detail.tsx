/**
 * Warehouse Detail Component
 *
 * Comprehensive warehouse information display with:
 * - Basic info card (code, name, type, status)
 * - Location card (address, city, province, postal code)
 * - Contact card (phone, email)
 * - Management card (manager, capacity)
 * - Timestamps
 */

"use client";

import {
  Warehouse as WarehouseIcon,
  MapPin,
  Phone,
  Mail,
  Settings,
  Scale,
  CheckCircle2,
  XCircle,
  Calendar,
  Tag,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { WarehouseResponse } from "@/types/warehouse.types";
import {
  getWarehouseTypeLabel,
  getWarehouseTypeBadgeColor,
  formatCapacity,
} from "@/types/warehouse.types";

interface WarehouseDetailProps {
  warehouse: WarehouseResponse;
}

export function WarehouseDetail({ warehouse }: WarehouseDetailProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Basic Information Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <WarehouseIcon className="h-5 w-5" />
            Informasi Dasar
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            {/* Warehouse Code */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Kode Gudang
              </p>
              <p className="font-mono text-lg font-semibold">
                {warehouse.code}
              </p>
            </div>

            {/* Warehouse Name */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Nama Gudang
              </p>
              <p className="text-lg font-semibold">{warehouse.name}</p>
            </div>

            {/* Warehouse Type */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Tipe Gudang
              </p>
              <Badge
                className={`${getWarehouseTypeBadgeColor(warehouse.type)} text-sm`}
              >
                <Tag className="mr-1 h-3 w-3" />
                {getWarehouseTypeLabel(warehouse.type)}
              </Badge>
            </div>
          </div>

          <Separator />

          {/* Status */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">Status</p>
            <Badge
              className={
                warehouse.isActive
                  ? "bg-green-500 text-white hover:bg-green-600"
                  : "bg-red-500 text-white hover:bg-red-600"
              }
            >
              {warehouse.isActive ? (
                <CheckCircle2 className="mr-1 h-3 w-3" />
              ) : (
                <XCircle className="mr-1 h-3 w-3" />
              )}
              {warehouse.isActive ? "Aktif" : "Nonaktif"}
            </Badge>
          </div>

          {/* Timestamps */}
          <Separator />
          <div className="grid gap-4 text-xs text-muted-foreground md:grid-cols-2">
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Dibuat:{" "}
                {new Date(warehouse.createdAt).toLocaleDateString("id-ID", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Diperbarui:{" "}
                {new Date(warehouse.updatedAt).toLocaleDateString("id-ID", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Location & Contact Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Lokasi & Kontak
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Address */}
          {warehouse.address && (
            <>
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Alamat
                </p>
                <p className="text-sm leading-relaxed">{warehouse.address}</p>
              </div>
              <Separator />
            </>
          )}

          {/* City */}
          {warehouse.city && (
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">Kota</p>
              <p className="text-sm">{warehouse.city}</p>
            </div>
          )}

          {/* Province */}
          {warehouse.province && (
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Provinsi
              </p>
              <p className="text-sm">{warehouse.province}</p>
            </div>
          )}

          {/* Postal Code */}
          {warehouse.postalCode && (
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Kode Pos
              </p>
              <p className="text-sm font-mono">{warehouse.postalCode}</p>
            </div>
          )}

          {/* Show separator before contact info if location exists */}
          {(warehouse.address ||
            warehouse.city ||
            warehouse.province ||
            warehouse.postalCode) &&
            (warehouse.phone || warehouse.email) && <Separator />}

          {/* Phone */}
          {warehouse.phone && (
            <div className="flex items-center gap-2">
              <Phone className="h-4 w-4 text-muted-foreground" />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Telepon
                </p>
                <p className="text-sm">{warehouse.phone}</p>
              </div>
            </div>
          )}

          {/* Email */}
          {warehouse.email && (
            <div className="flex items-center gap-2">
              <Mail className="h-4 w-4 text-muted-foreground" />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Email
                </p>
                <a
                  href={`mailto:${warehouse.email}`}
                  className="text-sm text-blue-600 hover:underline"
                >
                  {warehouse.email}
                </a>
              </div>
            </div>
          )}

          {/* Empty state */}
          {!warehouse.address &&
            !warehouse.city &&
            !warehouse.province &&
            !warehouse.postalCode &&
            !warehouse.phone &&
            !warehouse.email && (
              <div className="text-center py-8 text-muted-foreground text-sm">
                Belum ada informasi lokasi dan kontak
              </div>
            )}
        </CardContent>
      </Card>

      {/* Management & Capacity Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Manajemen & Kapasitas
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Manager */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">Manager</p>
            {warehouse.managerID ? (
              <Badge variant="outline">
                <Settings className="mr-1 h-3 w-3" />
                ID: {warehouse.managerID}
              </Badge>
            ) : (
              <p className="text-sm text-muted-foreground">
                Belum ada manager yang ditugaskan
              </p>
            )}
          </div>

          <Separator />

          {/* Capacity */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">
              Kapasitas Gudang
            </p>
            <div className="flex items-baseline gap-2">
              <Scale className="h-4 w-4 text-muted-foreground" />
              <p className="text-2xl font-bold">
                {formatCapacity(warehouse.capacity)}
              </p>
            </div>
            {warehouse.capacity && (
              <p className="text-xs text-muted-foreground">
                Kapasitas total dalam meter persegi
              </p>
            )}
          </div>

          {/* Future: Stock summary can be added here */}
          <Separator />
          <div className="rounded-lg bg-muted/50 p-4">
            <p className="text-sm font-medium mb-2">Informasi Stok</p>
            <p className="text-xs text-muted-foreground">
              Ringkasan stok gudang akan ditampilkan di sini setelah integrasi
              dengan modul inventori
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
