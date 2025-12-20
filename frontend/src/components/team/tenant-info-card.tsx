/**
 * Tenant Info Card Component
 *
 * Displays tenant information and subscription status.
 * Shows subscription plan, status, active users, and validity period.
 */

"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Building2, Calendar, Users, CreditCard } from "lucide-react";
import { format } from "date-fns";
import type { Tenant } from "@/types/tenant.types";

interface TenantInfoCardProps {
  tenant: Tenant;
}

export function TenantInfoCard({ tenant }: TenantInfoCardProps) {
  const subscription = tenant.subscription;

  // Subscription status badge variant
  const getStatusVariant = (
    status: string
  ): "default" | "secondary" | "destructive" | "outline" => {
    switch (status) {
      case "ACTIVE":
        return "default";
      case "TRIAL":
        return "secondary";
      case "EXPIRED":
      case "SUSPENDED":
        return "destructive";
      default:
        return "outline";
    }
  };

  // Calculate days remaining
  const getDaysRemaining = (): number | null => {
    if (!subscription?.endDate) return null;
    const endDate = new Date(subscription.endDate);
    const today = new Date();
    const diffTime = endDate.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

  const daysRemaining = getDaysRemaining();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          Organization Information
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {/* Tenant Name */}
          <div className="space-y-1">
            <p className="text-sm text-muted-foreground">Organization</p>
            <p className="text-lg font-semibold">{tenant.name}</p>
            <p className="text-xs text-muted-foreground">{tenant.subdomain}</p>
          </div>

          {/* Subscription Plan */}
          {subscription && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground flex items-center gap-1">
                <CreditCard className="h-3 w-3" />
                Subscription Plan
              </p>
              <p className="text-lg font-semibold">{subscription.planName}</p>
              <Badge variant={getStatusVariant(subscription.status)}>
                {subscription.status}
              </Badge>
            </div>
          )}

          {/* Max Users */}
          {subscription && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground flex items-center gap-1">
                <Users className="h-3 w-3" />
                User Limit
              </p>
              <p className="text-lg font-semibold">{subscription.maxUsers}</p>
              <p className="text-xs text-muted-foreground">Maximum users allowed</p>
            </div>
          )}

          {/* Subscription Period */}
          {subscription && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground flex items-center gap-1">
                <Calendar className="h-3 w-3" />
                Valid Until
              </p>
              <p className="text-lg font-semibold">
                {format(new Date(subscription.endDate), "MMM dd, yyyy")}
              </p>
              {daysRemaining !== null && (
                <p
                  className={`text-xs ${
                    daysRemaining < 7
                      ? "text-red-600 font-medium"
                      : daysRemaining < 30
                        ? "text-yellow-600 font-medium"
                        : "text-muted-foreground"
                  }`}
                >
                  {daysRemaining > 0
                    ? `${daysRemaining} days remaining`
                    : daysRemaining === 0
                      ? "Expires today"
                      : `Expired ${Math.abs(daysRemaining)} days ago`}
                </p>
              )}
            </div>
          )}
        </div>

        {/* Warning if subscription is expiring soon or expired */}
        {subscription && daysRemaining !== null && daysRemaining < 7 && (
          <div className="mt-4 rounded-md bg-yellow-50 dark:bg-yellow-950 p-3 border border-yellow-200 dark:border-yellow-800">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              {daysRemaining > 0 ? (
                <>
                  ⚠️ Your subscription will expire in{" "}
                  <strong>{daysRemaining} days</strong>. Please renew to avoid service
                  interruption.
                </>
              ) : (
                <>
                  ⚠️ Your subscription has expired. Please renew to continue using the
                  service.
                </>
              )}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
