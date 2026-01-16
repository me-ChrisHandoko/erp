/**
 * Create Company Form Component
 *
 * Form for creating new company with:
 * - Legal entity information
 * - Contact details
 * - Indonesian tax compliance
 * - Document numbering settings
 */

"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Building2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { useToast } from "@/hooks/use-toast";
import { useCreateCompanyMutation } from "@/store/services/companyApi";
import { ENTITY_TYPES, INDONESIAN_PROVINCES } from "@/types/company.types";
import type { CreateCompanyRequest } from "@/types/company.types";

// Validation schema
const createCompanySchema = z.object({
  name: z.string().min(2, "Nama perusahaan minimal 2 karakter"),
  legalName: z.string().min(2, "Nama legal minimal 2 karakter"),
  entityType: z.enum(ENTITY_TYPES),
  address: z.string().min(5, "Alamat minimal 5 karakter"),
  city: z.string().min(2, "Nama kota minimal 2 karakter"),
  province: z.enum(INDONESIAN_PROVINCES),
  postalCode: z.string().length(5, "Kode pos harus 5 digit").optional().or(z.literal("")),
  phone: z.string().min(10, "Nomor telepon minimal 10 digit"),
  email: z.string().email("Format email tidak valid"),
  website: z.string().url("Format URL tidak valid").optional().or(z.literal("")),
  npwp: z.string().optional(),
  nib: z.string().optional(),
  isPkp: z.boolean(),
  ppnRate: z.number().min(0).max(100),
  invoicePrefix: z.string().optional(),
  invoiceNumberFormat: z.string().optional(),
  poPrefix: z.string().optional(),
  poNumberFormat: z.string().optional(),
  soPrefix: z.string().optional(),
  soNumberFormat: z.string().optional(),
  fakturPajakSeries: z.string().optional(),
  sppkpNumber: z.string().optional(),
});

type CreateCompanyFormData = z.infer<typeof createCompanySchema>;

interface CreateCompanyFormProps {
  onSuccess: (companyId: string) => void;
  onCancel: () => void;
}

export function CreateCompanyForm({ onSuccess, onCancel }: CreateCompanyFormProps) {
  const { toast } = useToast();
  const [createCompany, { isLoading }] = useCreateCompanyMutation();

  const form = useForm<CreateCompanyFormData>({
    resolver: zodResolver(createCompanySchema),
    defaultValues: {
      name: "",
      legalName: "",
      entityType: "CV",
      address: "",
      city: "",
      province: "DKI Jakarta",
      postalCode: "",
      phone: "",
      email: "",
      website: "",
      npwp: "",
      nib: "",
      isPkp: false,
      ppnRate: 11,
      invoicePrefix: "INV",
      invoiceNumberFormat: "{PREFIX}-{YEAR}-{NUMBER}",
      poPrefix: "PO",
      poNumberFormat: "{PREFIX}-{YEAR}-{NUMBER}",
      soPrefix: "SO",
      soNumberFormat: "{PREFIX}-{YEAR}-{NUMBER}",
      fakturPajakSeries: "",
      sppkpNumber: "",
    },
  });

  const onSubmit = async (data: CreateCompanyFormData) => {
    try {
      const payload: CreateCompanyRequest = {
        ...data,
        postalCode: data.postalCode || undefined,
        website: data.website || undefined,
        npwp: data.npwp || undefined,
        nib: data.nib || undefined,
        fakturPajakSeries: data.fakturPajakSeries || undefined,
        sppkpNumber: data.sppkpNumber || undefined,
        invoicePrefix: data.invoicePrefix || undefined,
        invoiceNumberFormat: data.invoiceNumberFormat || undefined,
        poPrefix: data.poPrefix || undefined,
        poNumberFormat: data.poNumberFormat || undefined,
        soPrefix: data.soPrefix || undefined,
        soNumberFormat: data.soNumberFormat || undefined,
      };

      const result = await createCompany(payload).unwrap();

      toast({
        title: "Perusahaan berhasil dibuat",
        description: "Data perusahaan telah disimpan",
      });

      onSuccess(result.id);
    } catch (error: any) {
      toast({
        variant: "destructive",
        title: "Gagal membuat perusahaan",
        description: error?.data?.message || "Terjadi kesalahan saat menyimpan data",
      });
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        {/* Basic Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Informasi Dasar
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-3">
              <FormField
                control={form.control}
                name="entityType"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Jenis Badan Usaha</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger className="w-full bg-background">
                          <SelectValue placeholder="Pilih jenis badan usaha" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {ENTITY_TYPES.map((type) => (
                          <SelectItem key={type} value={type}>
                            {type}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Nama Perusahaan</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="PT Distribusi Jaya"
                        {...field}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="legalName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Nama Legal</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="PT Distribusi Jaya Abadi"
                        {...field}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Contact Information */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Kontak</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="md:col-span-2">
                <FormField
                  control={form.control}
                  name="address"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Alamat</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="Jl. Sudirman No. 123"
                          {...field}
                          className="bg-background"
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name="province"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Provinsi</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger className="w-full bg-background">
                          <SelectValue placeholder="Pilih provinsi" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {INDONESIAN_PROVINCES.map((province) => (
                          <SelectItem key={province} value={province}>
                            {province}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="city"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Kota</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Jakarta Selatan"
                        {...field}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="postalCode"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Kode Pos</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="12345"
                        {...field}
                        value={field.value || ""}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormDescription className="opacity-0">-</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="phone"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Telepon</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="+628123456789"
                        {...field}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormDescription>
                      Format: +628xxx atau 08xxx
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="website"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Website</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="https://www.perusahaan.com"
                        {...field}
                        value={field.value || ""}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormDescription className="opacity-0">-</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input
                        type="email"
                        placeholder="info@perusahaan.com"
                        {...field}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormDescription className="opacity-0">-</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Tax Information */}
        <Card>
          <CardHeader>
            <CardTitle>Informasi Pajak</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <FormField
                control={form.control}
                name="npwp"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>NPWP</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="12.345.678.9-012.345"
                        {...field}
                        value={field.value || ""}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="nib"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>NIB</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="1234567890123"
                        {...field}
                        value={field.value || ""}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="isPkp"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Status PKP</FormLabel>
                    <Select
                      onValueChange={(value) => field.onChange(value === "true")}
                      defaultValue={String(field.value)}
                    >
                      <FormControl>
                        <SelectTrigger className="w-full bg-background">
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="true">PKP</SelectItem>
                        <SelectItem value="false">Non-PKP</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="ppnRate"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>PPN Rate (%)</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={0}
                        max={100}
                        {...field}
                        onChange={(e) => field.onChange(parseFloat(e.target.value))}
                        className="bg-background"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </CardContent>
        </Card>

        {/* Document Numbering */}
        <Card>
          <CardHeader>
            <CardTitle>Pengaturan Penomoran Dokumen</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-4">
              {/* Purchase Invoice */}
              <div className="space-y-3">
                <h4 className="text-sm font-semibold">Purchase Invoice</h4>
                <div className="grid gap-4 md:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="invoicePrefix"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Prefix Invoice</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="INV"
                            {...field}
                            value={field.value || ""}
                            className="bg-background"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="invoiceNumberFormat"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Format Nomor</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="{PREFIX}-{YEAR}-{NUMBER}"
                            {...field}
                            value={field.value || ""}
                            className="bg-background"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              <Separator />

              {/* Purchase Order */}
              <div className="space-y-3">
                <h4 className="text-sm font-semibold">Purchase Order</h4>
                <div className="grid gap-4 md:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="poPrefix"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Prefix PO</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="PO"
                            {...field}
                            value={field.value || ""}
                            className="bg-background"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="poNumberFormat"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Format Nomor</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="{PREFIX}-{YEAR}-{NUMBER}"
                            {...field}
                            value={field.value || ""}
                            className="bg-background"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>

              <Separator />

              {/* Sales Order */}
              <div className="space-y-3">
                <h4 className="text-sm font-semibold">Sales Order</h4>
                <div className="grid gap-4 md:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="soPrefix"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Prefix SO</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="SO"
                            {...field}
                            value={field.value || ""}
                            className="bg-background"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="soNumberFormat"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Format Nomor</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="{PREFIX}-{YEAR}-{NUMBER}"
                            {...field}
                            value={field.value || ""}
                            className="bg-background"
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex justify-end gap-3">
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={isLoading}
          >
            Batal
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Simpan Perusahaan
          </Button>
        </div>
      </form>
    </Form>
  );
}
