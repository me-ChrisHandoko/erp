/**
 * Supplier Detail Component
 *
 * Comprehensive supplier information display with:
 * - Basic info card (code, name, type, PKP status)
 * - Contact & address card (contact person, email, phone, address)
 * - Business info card (NPWP, payment terms, credit limit, outstanding)
 */

"use client";

import {
  Building2,
  Phone,
  Mail,
  MapPin,
  DollarSign,
  Calendar,
  Tag,
  CheckCircle2,
  XCircle,
  FileText,
  TrendingUp,
  AlertCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { SupplierResponse } from "@/types/supplier.types";

interface SupplierDetailProps {
  supplier: SupplierResponse;
}

export function SupplierDetail({ supplier }: SupplierDetailProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Basic Information Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Informasi Dasar
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            {/* Supplier Code */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Kode Supplier
              </p>
              <p className="font-mono text-lg font-semibold">{supplier.code}</p>
            </div>

            {/* Supplier Name */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Nama Supplier
              </p>
              <p className="text-lg font-semibold">{supplier.name}</p>
            </div>

            {/* Type */}
            {supplier.type && (
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Tipe Supplier
                </p>
                <Badge className="bg-blue-500 text-white hover:bg-blue-600 text-sm">
                  <Tag className="mr-1 h-3 w-3" />
                  {supplier.type.charAt(0) + supplier.type.slice(1).toLowerCase()}
                </Badge>
              </div>
            )}

            {/* Status */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Status
              </p>
              <Badge
                className={
                  supplier.isActive
                    ? "bg-green-500 text-white hover:bg-green-600"
                    : "bg-red-500 text-white hover:bg-red-600"
                }
              >
                {supplier.isActive ? (
                  <CheckCircle2 className="mr-1 h-3 w-3" />
                ) : (
                  <XCircle className="mr-1 h-3 w-3" />
                )}
                {supplier.isActive ? "Aktif" : "Nonaktif"}
              </Badge>
            </div>

            {/* PKP Status */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Status PKP
              </p>
              <Badge
                className={
                  supplier.isPKP
                    ? "bg-purple-500 text-white hover:bg-purple-600"
                    : "bg-gray-500 text-white hover:bg-gray-600"
                }
              >
                {supplier.isPKP ? "PKP" : "Non-PKP"}
              </Badge>
            </div>
          </div>

          {/* Notes */}
          {supplier.notes && (
            <>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Catatan
                </p>
                <p className="text-sm leading-relaxed">{supplier.notes}</p>
              </div>
            </>
          )}

          {/* Timestamps */}
          <Separator />
          <div className="grid gap-4 text-xs text-muted-foreground md:grid-cols-2">
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Dibuat: {new Date(supplier.createdAt).toLocaleDateString("id-ID")}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Diperbarui:{" "}
                {new Date(supplier.updatedAt).toLocaleDateString("id-ID")}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Contact & Address Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Phone className="h-5 w-5" />
            Kontak & Alamat
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Contact Person */}
          {supplier.contactPerson && (
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Nama Kontak Person
              </p>
              <p className="font-semibold">{supplier.contactPerson}</p>
            </div>
          )}

          {supplier.contactPerson && <Separator />}

          {/* Email */}
          {supplier.email && (
            <div className="flex items-center gap-2">
              <Mail className="h-4 w-4 text-muted-foreground" />
              <div className="flex-1">
                <p className="text-sm text-muted-foreground">Email</p>
                <a
                  href={`mailto:${supplier.email}`}
                  className="font-medium text-blue-600 hover:underline"
                >
                  {supplier.email}
                </a>
              </div>
            </div>
          )}

          {/* Phone */}
          {supplier.phone && (
            <div className="flex items-center gap-2">
              <Phone className="h-4 w-4 text-muted-foreground" />
              <div className="flex-1">
                <p className="text-sm text-muted-foreground">Telepon Supplier</p>
                <a
                  href={`tel:${supplier.phone}`}
                  className="font-medium text-blue-600 hover:underline"
                >
                  {supplier.phone}
                </a>
              </div>
            </div>
          )}

          {/* Contact Phone */}
          {supplier.contactPhone && (
            <div className="flex items-center gap-2">
              <Phone className="h-4 w-4 text-muted-foreground" />
              <div className="flex-1">
                <p className="text-sm text-muted-foreground">Telepon Kontak Person</p>
                <a
                  href={`tel:${supplier.contactPhone}`}
                  className="font-medium text-blue-600 hover:underline"
                >
                  {supplier.contactPhone}
                </a>
              </div>
            </div>
          )}

          {(supplier.email || supplier.phone || supplier.contactPhone) && <Separator />}

          {/* Address */}
          {supplier.address && (
            <div className="flex items-start gap-2">
              <MapPin className="mt-1 h-4 w-4 text-muted-foreground" />
              <div className="flex-1 space-y-1">
                <p className="text-sm text-muted-foreground">Alamat</p>
                <p className="text-sm leading-relaxed">{supplier.address}</p>
                {(supplier.city || supplier.province || supplier.postalCode) && (
                  <p className="text-sm text-muted-foreground">
                    {[supplier.city, supplier.province, supplier.postalCode]
                      .filter(Boolean)
                      .join(", ")}
                  </p>
                )}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Business Information Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-5 w-5" />
            Informasi Bisnis
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* NPWP */}
          {supplier.npwp && (
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <FileText className="h-4 w-4 text-muted-foreground" />
                <p className="text-sm font-medium text-muted-foreground">
                  NPWP
                </p>
              </div>
              <p className="font-mono font-semibold">{supplier.npwp}</p>
            </div>
          )}

          {supplier.npwp && <Separator />}

          {/* Payment Terms */}
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">Term of Payment</p>
            <Badge variant="outline">
              <Calendar className="mr-1 h-3 w-3" />
              {supplier.paymentTerm === 0
                ? "Cash"
                : `NET ${supplier.paymentTerm}`}
            </Badge>
          </div>

          {/* Credit Limit */}
          <Separator />
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
              <p className="text-sm font-medium text-muted-foreground">
                Credit Limit
              </p>
            </div>
            <p className="text-lg font-bold text-green-600">
              Rp {Number(supplier.creditLimit).toLocaleString("id-ID")}
            </p>
          </div>

          {/* Current Outstanding */}
          <Separator />
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-muted-foreground" />
              <p className="text-sm font-medium text-muted-foreground">
                Current Outstanding
              </p>
            </div>
            <p className="text-lg font-semibold">
              Rp {Number(supplier.currentOutstanding).toLocaleString("id-ID")}
            </p>
          </div>

          {/* Overdue Amount */}
          {Number(supplier.overdueAmount) > 0 && (
            <>
              <Separator />
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <AlertCircle className="h-4 w-4 text-red-500" />
                  <p className="text-sm font-medium text-red-600">
                    Overdue Amount
                  </p>
                </div>
                <p className="text-lg font-bold text-red-600">
                  Rp {Number(supplier.overdueAmount).toLocaleString("id-ID")}
                </p>
              </div>
            </>
          )}

          {/* Last Transaction */}
          {supplier.lastTransactionAt && (
            <>
              <Separator />
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Transaksi Terakhir
                </p>
                <p className="font-semibold">
                  {new Date(supplier.lastTransactionAt).toLocaleDateString("id-ID", {
                    day: "numeric",
                    month: "long",
                    year: "numeric",
                  })}
                </p>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
