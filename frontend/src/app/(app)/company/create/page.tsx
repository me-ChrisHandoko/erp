/**
 * Create Company Page
 *
 * Full-page form for creating new company with:
 * - Legal entity information
 * - Contact details
 * - Indonesian tax compliance
 * - Document numbering settings
 */

"use client";

import { useRouter } from "next/navigation";
import { Building2 } from "lucide-react";
import { PageHeader } from "@/components/shared/page-header";
import { CreateCompanyForm } from "@/components/company/create-company-form";

export default function CreateCompanyPage() {
  const router = useRouter();

  const handleSuccess = (companyId: string) => {
    // Small delay to allow cache invalidation to complete
    setTimeout(() => {
      router.push("/company/profile");
      router.refresh();
    }, 100);
  };

  const handleCancel = () => {
    router.push("/dashboard");
  };

  return (
    <div className="flex flex-col">
      <PageHeader
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Perusahaan", href: "/company" },
          { label: "Tambah Perusahaan" },
        ]}
      />

      {/* Main content */}
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Page title */}
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Building2 className="h-6 w-6 text-muted-foreground" />
            <h1 className="text-3xl font-bold tracking-tight">
              Tambah Perusahaan Baru
            </h1>
          </div>
          <p className="text-muted-foreground">
            Buat profil perusahaan untuk sistem ERP Anda
          </p>
        </div>

        {/* Create Company Form */}
        <CreateCompanyForm
          onSuccess={handleSuccess}
          onCancel={handleCancel}
        />
      </div>
    </div>
  );
}
