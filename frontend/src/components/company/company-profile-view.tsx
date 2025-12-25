/**
 * CompanyProfileView Component
 *
 * Read-only display of company profile information.
 * Shows all company details in organized sections.
 */

import Image from "next/image";
import { Building2, Mail, Phone, Globe, FileText, Banknote } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import type { CompanyResponse } from "@/types/company.types";
import { ENTITY_TYPES } from "@/types/company.types";

interface CompanyProfileViewProps {
  company: CompanyResponse;
}

export function CompanyProfileView({ company }: CompanyProfileViewProps) {
  return (
    <div className="space-y-8">
      {/* Logo Section */}
      {company.logoUrl && (
        <div className="flex justify-center py-4">
          <div className="relative h-28 w-56 rounded-lg border border-border bg-muted/30 p-4">
            <Image
              src={company.logoUrl}
              alt={`${company.name} logo`}
              fill
              className="object-contain p-2"
            />
          </div>
        </div>
      )}

      {/* Basic Information */}
      <div className="space-y-5">
        <div className="flex items-center gap-2 pb-2 border-b border-border/50">
          <Building2 className="h-5 w-5 text-primary" />
          <h3 className="text-xl font-semibold">Informasi Dasar</h3>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <DataField label="Nama Perusahaan" value={company.name} />
          <DataField label="Nama Legal" value={company.legalName} />
          <DataField
            label="Jenis Badan Usaha"
            value={company.entityType || "CV"}
          />
          <DataField
            label="Status"
            value={
              <Badge variant={company.isActive ? "default" : "secondary"}>
                {company.isActive ? "Aktif" : "Tidak Aktif"}
              </Badge>
            }
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
          <DataField
            label="Alamat"
            value={company.address}
            icon={<Building2 className="h-4 w-4" />}
          />
          <DataField label="Kota" value={company.city} />
          <DataField label="Provinsi" value={company.province} />
          <DataField label="Kode Pos" value={company.postalCode} />
          <DataField
            label="Telepon"
            value={company.phone}
            icon={<Phone className="h-4 w-4" />}
          />
          <DataField
            label="Email"
            value={company.email}
            icon={<Mail className="h-4 w-4" />}
          />
          <DataField
            label="Website"
            value={
              company.website ? (
                <a
                  href={company.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary hover:underline"
                >
                  {company.website}
                </a>
              ) : (
                "-"
              )
            }
            icon={<Globe className="h-4 w-4" />}
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
          <DataField label="NPWP" value={company.npwp || "-"} />
          <DataField label="NIB" value={company.nib || "-"} />
          <DataField
            label="Status PKP"
            value={
              <Badge variant={company.isPkp ? "default" : "secondary"}>
                {company.isPkp ? "PKP" : "Non-PKP"}
              </Badge>
            }
          />
          {company.isPkp && (
            <>
              <DataField label="PPN Rate" value={`${company.ppnRate}%`} />
            </>
          )}
        </div>
      </div>

      <Separator />

      {/* Invoice Settings */}
      <div className="space-y-5">
        <div className="flex items-center gap-2 pb-2 border-b border-border/50">
          <Banknote className="h-5 w-5 text-primary" />
          <h3 className="text-xl font-semibold">Pengaturan Dokumen</h3>
        </div>
        <div className="grid gap-4 md:grid-cols-2">
          <DataField label="Prefix Invoice" value={company.invoicePrefix || "INV"} />
        </div>
      </div>

      {/* Bank Accounts */}
      {company.banks && company.banks.length > 0 && (
        <>
          <Separator />
          <div className="space-y-5">
            <div className="flex items-center gap-2 pb-2 border-b border-border/50">
              <Banknote className="h-5 w-5 text-primary" />
              <h3 className="text-xl font-semibold">Rekening Bank</h3>
            </div>
            <div className="space-y-3">
              {company.banks.map((bank) => (
                <div
                  key={bank.id}
                  className="flex items-center justify-between rounded-lg border border-border bg-muted/20 p-4 transition-colors hover:bg-muted/40"
                >
                  <div className="space-y-1.5">
                    <div className="flex items-center gap-2">
                      <p className="font-semibold text-base">{bank.bankName}</p>
                      {bank.isPrimary && (
                        <Badge variant="default" className="text-xs">
                          Utama
                        </Badge>
                      )}
                    </div>
                    <p className="text-sm font-medium text-foreground/80">
                      {bank.accountNumber} - {bank.accountName}
                    </p>
                    {bank.branchName && (
                      <p className="text-xs text-muted-foreground">
                        Cabang: {bank.branchName}
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

/**
 * DataField Component
 * Displays a label-value pair with optional icon
 */
interface DataFieldProps {
  label: string;
  value: React.ReactNode;
  icon?: React.ReactNode;
}

function DataField({ label, value, icon }: DataFieldProps) {
  return (
    <div className="space-y-1.5">
      <p className="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
        {icon}
        {label}
      </p>
      <p className="text-base font-medium">
        {value || <span className="text-muted-foreground">-</span>}
      </p>
    </div>
  );
}
