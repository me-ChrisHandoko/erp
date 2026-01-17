"use client";

/**
 * CompanyProfileForm Component
 *
 * Form for editing company profile information.
 * Uses react-hook-form with zod validation.
 */

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Loader2, Building2, Phone, FileText, Banknote } from "lucide-react";
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
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { useUpdateCompanyMutation } from "@/store/services/companyApi";
import { updateCompanySchema, type UpdateCompanyFormData } from "@/lib/schemas/company.schema";
import type { CompanyResponse } from "@/types/company.types";
import { ENTITY_TYPES, INDONESIAN_PROVINCES } from "@/types/company.types";
import { LogoUpload } from "./logo-upload";

interface CompanyProfileFormProps {
  company: CompanyResponse;
  onSuccess: () => void;
  onCancel: () => void;
}

export function CompanyProfileForm({
  company,
  onSuccess,
  onCancel,
}: CompanyProfileFormProps) {
  const [updateCompany, { isLoading }] = useUpdateCompanyMutation();

  const form = useForm<UpdateCompanyFormData>({
    resolver: zodResolver(updateCompanySchema),
    defaultValues: {
      name: company.name || "",
      legalName: company.legalName || "",
      entityType: (company.entityType && ENTITY_TYPES.includes(company.entityType as (typeof ENTITY_TYPES)[number]))
        ? (company.entityType as (typeof ENTITY_TYPES)[number])
        : undefined,
      address: company.address || "",
      city: company.city || "",
      province: (company.province && INDONESIAN_PROVINCES.includes(company.province as (typeof INDONESIAN_PROVINCES)[number]))
        ? (company.province as (typeof INDONESIAN_PROVINCES)[number])
        : undefined,
      postalCode: company.postalCode || "",
      phone: company.phone || "",
      email: company.email || "",
      website: company.website || "",
      npwp: company.npwp || "",
      nib: company.nib || "",
      isPkp: company.isPkp || false,
      ppnRate: company.ppnRate || 11,
      invoicePrefix: company.invoicePrefix || "INV",
      invoiceNumberFormat: company.invoiceNumberFormat || "{PREFIX}-{YEAR}-{NUMBER}",
      poPrefix: company.poPrefix || "PO",
      poNumberFormat: company.poNumberFormat || "{PREFIX}-{YEAR}-{NUMBER}",
      soPrefix: company.soPrefix || "SO",
      soNumberFormat: company.soNumberFormat || "{PREFIX}-{YEAR}-{NUMBER}",
    },
  });

  const isPkp = form.watch("isPkp");

  const onSubmit = async (data: UpdateCompanyFormData) => {
    try {
      await updateCompany(data).unwrap();
      toast.success("Profil perusahaan berhasil diperbarui");
      onSuccess();
    } catch (error: unknown) {
      const errorMessage = (error as { data?: { error?: { message?: string } } })?.data?.error?.message || "Gagal memperbarui profil perusahaan";
      toast.error(errorMessage);
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        {/* Logo Upload */}
        <div className="space-y-5">
          <div className="flex items-center gap-2 pb-2 border-b border-border/50">
            <Building2 className="h-5 w-5 text-primary" />
            <h3 className="text-xl font-semibold">Logo Perusahaan</h3>
          </div>
          <LogoUpload currentLogoUrl={company.logoUrl} />
        </div>

        <Separator />

        {/* Basic Information */}
        <div className="space-y-5">
          <div className="flex items-center gap-2 pb-2 border-b border-border/50">
            <Building2 className="h-5 w-5 text-primary" />
            <h3 className="text-xl font-semibold">Informasi Dasar</h3>
          </div>
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
                  <FormLabel>Nama Perusahaan *</FormLabel>
                  <FormControl>
                    <Input placeholder="PT Maju Jaya" {...field} className="bg-background" />
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
                  <FormLabel>Nama Legal *</FormLabel>
                  <FormControl>
                    <Input placeholder="PT Maju Jaya Sejahtera" {...field} className="bg-background" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        <Separator />

        {/* Contact Information */}
        <div className="space-y-5">
          <div className="flex items-center gap-2 pb-2 border-b border-border/50">
            <Phone className="h-5 w-5 text-primary" />
            <h3 className="text-xl font-semibold">Informasi Kontak</h3>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <FormField
              control={form.control}
              name="address"
              render={({ field }) => (
                <FormItem className="md:col-span-2">
                  <FormLabel>Alamat</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Jl. Contoh No. 123"
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
              name="province"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Provinsi</FormLabel>
                  <Select
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                  >
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
                      placeholder="Jakarta"
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
                      value={field.value || ""}
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
                      value={field.value || ""}
                      className="bg-background"
                    />
                  </FormControl>
                  <FormDescription className="opacity-0">-</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        <Separator />

        {/* Tax Compliance */}
        <div className="space-y-5">
          <div className="flex items-center gap-2 pb-2 border-b border-border/50">
            <FileText className="h-5 w-5 text-primary" />
            <h3 className="text-xl font-semibold">Informasi Pajak</h3>
          </div>
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
                  <FormDescription>
                    Format: XX.XXX.XXX.X-XXX.XXX
                  </FormDescription>
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
                  <FormDescription>Nomor Induk Berusaha</FormDescription>
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
                    value={String(field.value)}
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
                  <FormDescription>
                    Pengusaha Kena Pajak
                  </FormDescription>
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
                      step="0.01"
                      min="0"
                      max="100"
                      placeholder="11"
                      {...field}
                      value={field.value || 11}
                      onChange={(e) =>
                        field.onChange(parseFloat(e.target.value))
                      }
                      className="bg-background"
                      disabled={!isPkp}
                    />
                  </FormControl>
                  <FormDescription>
                    {isPkp ? "Tarif PPN untuk perusahaan ini" : "Set Status PKP untuk mengaktifkan"}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        <Separator />

        {/* Document Numbering Settings */}
        <div className="space-y-5">
          <div className="flex items-center gap-2 pb-2 border-b border-border/50">
            <Banknote className="h-5 w-5 text-primary" />
            <h3 className="text-xl font-semibold">Pengaturan Penomoran Dokumen</h3>
          </div>

          {/* Invoice Numbering */}
          <div className="space-y-3">
            <h4 className="text-sm font-semibold text-muted-foreground">Purchase Invoice</h4>
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
                        className="bg-background font-mono"
                      />
                    </FormControl>
                    <FormDescription>
                      Awalan nomor invoice
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="invoiceNumberFormat"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Format Nomor Invoice</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="{PREFIX}-{YEAR}-{NUMBER}"
                        {...field}
                        value={field.value || ""}
                        className="bg-background font-mono"
                      />
                    </FormControl>
                    <FormDescription>
                      Placeholder: {"{PREFIX}"}, {"{YEAR}"}, {"{MONTH}"}, {"{NUMBER}"}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>

          <Separator className="my-4" />

          {/* Purchase Order Numbering */}
          <div className="space-y-3">
            <h4 className="text-sm font-semibold text-muted-foreground">Purchase Order</h4>
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
                        className="bg-background font-mono"
                      />
                    </FormControl>
                    <FormDescription>
                      Awalan nomor purchase order
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="poNumberFormat"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Format Nomor PO</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="{PREFIX}-{YEAR}-{NUMBER}"
                        {...field}
                        value={field.value || ""}
                        className="bg-background font-mono"
                      />
                    </FormControl>
                    <FormDescription>
                      Placeholder: {"{PREFIX}"}, {"{YEAR}"}, {"{MONTH}"}, {"{NUMBER}"}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>

          <Separator className="my-4" />

          {/* Sales Order Numbering */}
          <div className="space-y-3">
            <h4 className="text-sm font-semibold text-muted-foreground">Sales Order (Future)</h4>
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
                        className="bg-background font-mono"
                      />
                    </FormControl>
                    <FormDescription>
                      Awalan nomor sales order
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="soNumberFormat"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Format Nomor SO</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="{PREFIX}-{YEAR}-{NUMBER}"
                        {...field}
                        value={field.value || ""}
                        className="bg-background font-mono"
                      />
                    </FormControl>
                    <FormDescription>
                      Placeholder: {"{PREFIX}"}, {"{YEAR}"}, {"{MONTH}"}, {"{NUMBER}"}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>

          {/* Numbering Format Help */}
          <div className="rounded-lg bg-muted/50 p-4 space-y-2">
            <p className="text-sm font-medium">Format Penomoran:</p>
            <ul className="text-xs text-muted-foreground space-y-1 ml-4">
              <li>• <code className="bg-background px-1.5 py-0.5 rounded">{"{PREFIX}"}</code> - Prefix yang dikonfigurasi (INV, PO, SO)</li>
              <li>• <code className="bg-background px-1.5 py-0.5 rounded">{"{YEAR}"}</code> - Tahun 4 digit (2024)</li>
              <li>• <code className="bg-background px-1.5 py-0.5 rounded">{"{MONTH}"}</code> - Bulan 2 digit (01-12)</li>
              <li>• <code className="bg-background px-1.5 py-0.5 rounded">{"{NUMBER}"}</code> - Nomor urut 4 digit (0001, 0002, ...)</li>
            </ul>
            <div className="pt-2 space-y-1">
              <p className="text-xs font-medium">Contoh hasil:</p>
              <ul className="text-xs text-muted-foreground space-y-0.5 ml-4">
                <li>• <code className="bg-background px-1.5 py-0.5 rounded font-mono">INV-2024-0001</code></li>
                <li>• <code className="bg-background px-1.5 py-0.5 rounded font-mono">PO-2024-01-0001</code> (dengan bulan)</li>
                <li>• <code className="bg-background px-1.5 py-0.5 rounded font-mono">SO-0001</code> (nomor saja)</li>
              </ul>
            </div>
          </div>
        </div>

        {/* Form Actions */}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            Batal
          </Button>
          <Button type="submit" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Simpan Perubahan
          </Button>
        </div>
      </form>
    </Form>
  );
}
