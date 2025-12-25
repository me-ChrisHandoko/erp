"use client";

/**
 * Company Profile Page
 *
 * Displays and manages company profile information including:
 * - Legal entity details
 * - Contact information
 * - Indonesian tax compliance (NPWP, PKP)
 * - Invoice settings
 * - Logo upload
 */

import { useState } from "react";
import { Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { PageHeader } from "@/components/shared/page-header";
import { LoadingSpinner } from "@/components/shared/loading-spinner";
import { ErrorDisplay } from "@/components/shared/error-display";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { useGetCompanyQuery } from "@/store/services/companyApi";
import { CompanyProfileView } from "@/components/company/company-profile-view";
import { CompanyProfileForm } from "@/components/company/company-profile-form";

export default function CompanyProfilePage() {
  const [isEditing, setIsEditing] = useState(false);
  const { data: company, isLoading, error, refetch } = useGetCompanyQuery();

  return (
    <ErrorBoundary>
      <div className="flex flex-col">
        <PageHeader
          breadcrumbs={[
            { label: "Dashboard", href: "/dashboard" },
            { label: "Perusahaan", href: "/company" },
            { label: "Profil" },
          ]}
        />

        {/* Main Content */}
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
          {/* Page Header */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="space-y-1">
              <h1 className="text-3xl font-bold tracking-tight">Profil Perusahaan</h1>
              <p className="text-muted-foreground">
                Kelola informasi legal dan pengaturan perusahaan Anda
              </p>
            </div>
            {!isEditing && company && (
              <Button
                onClick={() => setIsEditing(true)}
                className="shrink-0"
              >
                <Pencil className="mr-2 h-4 w-4" />
                Edit Profil
              </Button>
            )}
          </div>

          <Card className="shadow-sm">
            <CardContent className="pt-6">
              {isLoading && (
                <div className="py-12">
                  <LoadingSpinner size="lg" text="Memuat data perusahaan..." />
                </div>
              )}

              {error && (
                <ErrorDisplay
                  error={error}
                  title="Gagal memuat data perusahaan"
                  onRetry={refetch}
                />
              )}

              {company && !isLoading && !error && (
                <>
                  {isEditing ? (
                    <CompanyProfileForm
                      company={company}
                      onCancel={() => setIsEditing(false)}
                      onSuccess={() => setIsEditing(false)}
                    />
                  ) : (
                    <CompanyProfileView company={company} />
                  )}
                </>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </ErrorBoundary>
  );
}
