/**
 * Customer Detail Component
 *
 * Comprehensive customer information display with:
 * - Basic info card (code, name, type, contact person, status)
 * - Contact card (phone, email)
 * - Address card (full address, city, province, postal code)
 * - Business terms card (NPWP, credit limit, credit term days)
 */

"use client";

import {
  Users,
  Phone,
  Mail,
  MapPin,
  CreditCard,
  Tag,
  Calendar,
  User2,
  Building2,
  Clock,
  CheckCircle2,
  XCircle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { CustomerResponse } from "@/types/customer.types";

interface CustomerDetailProps {
  customer: CustomerResponse;
}

export function CustomerDetail({ customer }: CustomerDetailProps) {
  // Format currency helper
  const formatCurrency = (value: string | null | undefined): string => {
    if (!value) return "Rp 0";
    const num = parseFloat(value);
    if (isNaN(num)) return "Rp 0";
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(num);
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Basic Information Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Informasi Dasar
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            {/* Customer Code */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Kode Pelanggan
              </p>
              <p className="font-mono text-lg font-semibold">{customer.code}</p>
            </div>

            {/* Customer Name */}
            <div className="space-y-1 md:col-span-2">
              <p className="text-sm font-medium text-muted-foreground">
                Nama Pelanggan
              </p>
              <p className="text-lg font-semibold">{customer.name}</p>
            </div>

            {/* Customer Type */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Tipe Pelanggan
              </p>
              <Badge className="bg-blue-500 text-white hover:bg-blue-600 text-sm">
                <Tag className="mr-1 h-3 w-3" />
                {customer.customerType}
              </Badge>
            </div>

            {/* Contact Person */}
            {customer.contactPerson && (
              <div className="space-y-1 md:col-span-2">
                <p className="text-sm font-medium text-muted-foreground">
                  Nama Kontak
                </p>
                <div className="flex items-center gap-2">
                  <User2 className="h-4 w-4 text-muted-foreground" />
                  <p className="text-sm">{customer.contactPerson}</p>
                </div>
              </div>
            )}

            {/* Status */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <Badge
                className={
                  customer.isActive
                    ? "bg-green-500 text-white hover:bg-green-600"
                    : "bg-red-500 text-white hover:bg-red-600"
                }
              >
                {customer.isActive ? (
                  <CheckCircle2 className="mr-1 h-3 w-3" />
                ) : (
                  <XCircle className="mr-1 h-3 w-3" />
                )}
                {customer.isActive ? "Aktif" : "Nonaktif"}
              </Badge>
            </div>
          </div>

          {/* Timestamps */}
          <Separator />
          <div className="grid gap-4 text-xs text-muted-foreground md:grid-cols-2">
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Dibuat:{" "}
                {new Date(customer.createdAt).toLocaleDateString("id-ID")}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <Calendar className="h-3 w-3" />
              <span>
                Diperbarui:{" "}
                {new Date(customer.updatedAt).toLocaleDateString("id-ID")}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Contact Information Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Phone className="h-5 w-5" />
            Informasi Kontak
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Phone */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">Telepon</p>
            {customer.phone ? (
              <div className="flex items-center gap-2">
                <Phone className="h-4 w-4 text-muted-foreground" />
                <p className="text-sm font-medium">{customer.phone}</p>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">-</p>
            )}
          </div>

          <Separator />

          {/* Email */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">Email</p>
            {customer.email ? (
              <div className="flex items-center gap-2">
                <Mail className="h-4 w-4 text-muted-foreground" />
                <p className="text-sm font-medium break-all">{customer.email}</p>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">-</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Address Card */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MapPin className="h-5 w-5" />
            Alamat
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Full Address */}
          {customer.address && (
            <>
              <div className="space-y-1">
                <p className="text-sm font-medium text-muted-foreground">
                  Alamat Lengkap
                </p>
                <p className="text-sm leading-relaxed">{customer.address}</p>
              </div>
              <Separator />
            </>
          )}

          {/* City */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">
              Kota/Kabupaten
            </p>
            {customer.city ? (
              <div className="flex items-center gap-2">
                <Building2 className="h-4 w-4 text-muted-foreground" />
                <p className="text-sm font-medium">{customer.city}</p>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">-</p>
            )}
          </div>

          {/* Province */}
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">Provinsi</p>
            {customer.province ? (
              <p className="text-sm font-medium">{customer.province}</p>
            ) : (
              <p className="text-sm text-muted-foreground">-</p>
            )}
          </div>

          {/* Postal Code */}
          {customer.postalCode && (
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Kode Pos
              </p>
              <p className="font-mono text-sm font-medium">
                {customer.postalCode}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Business Terms Card */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CreditCard className="h-5 w-5" />
            Ketentuan Bisnis
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            {/* NPWP */}
            {customer.npwp && (
              <div className="space-y-1 md:col-span-3">
                <p className="text-sm font-medium text-muted-foreground">NPWP</p>
                <p className="font-mono text-sm font-medium">{customer.npwp}</p>
              </div>
            )}

            {/* Credit Limit */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Limit Kredit
              </p>
              <div className="space-y-1">
                <p className="text-xl font-bold text-primary">
                  {formatCurrency(customer.creditLimit)}
                </p>
              </div>
            </div>

            {/* Credit Term Days */}
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                Jangka Waktu Kredit
              </p>
              <div className="flex items-baseline gap-1">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <p className="text-xl font-bold">
                  {customer.creditTermDays || 0}
                </p>
                <p className="text-sm text-muted-foreground">hari</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
