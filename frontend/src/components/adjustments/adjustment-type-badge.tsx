/**
 * Adjustment Type Badge Component
 *
 * Displays the type of an inventory adjustment (INCREASE/DECREASE) with appropriate styling.
 */

"use client";

import { Badge } from "@/components/ui/badge";
import { ArrowUp, ArrowDown } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  ADJUSTMENT_TYPE_CONFIG,
  type AdjustmentType,
} from "@/types/adjustment.types";

interface AdjustmentTypeBadgeProps {
  type: AdjustmentType;
  className?: string;
  showIcon?: boolean;
}

export function AdjustmentTypeBadge({
  type,
  className,
  showIcon = true,
}: AdjustmentTypeBadgeProps) {
  const config = ADJUSTMENT_TYPE_CONFIG[type];
  const Icon = type === "INCREASE" ? ArrowUp : ArrowDown;

  return (
    <Badge
      variant="secondary"
      className={cn(config.className, className)}
    >
      {showIcon && <Icon className="mr-1 h-3 w-3" />}
      {config.label}
    </Badge>
  );
}
