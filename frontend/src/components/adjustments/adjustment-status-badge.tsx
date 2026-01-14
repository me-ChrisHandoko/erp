/**
 * Adjustment Status Badge Component
 *
 * Displays the status of an inventory adjustment with appropriate styling.
 * Supports DRAFT, APPROVED, and CANCELLED statuses.
 */

"use client";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import {
  ADJUSTMENT_STATUS_CONFIG,
  type AdjustmentStatus,
} from "@/types/adjustment.types";

interface AdjustmentStatusBadgeProps {
  status: AdjustmentStatus;
  className?: string;
}

export function AdjustmentStatusBadge({
  status,
  className,
}: AdjustmentStatusBadgeProps) {
  const config = ADJUSTMENT_STATUS_CONFIG[status];

  return (
    <Badge
      variant={config.variant}
      className={cn(config.className, className)}
    >
      {config.label}
    </Badge>
  );
}
