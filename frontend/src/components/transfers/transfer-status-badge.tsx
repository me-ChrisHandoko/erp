/**
 * Transfer Status Badge Component
 *
 * Visual indicator for stock transfer status with color coding and icons.
 * Status workflow: DRAFT → SHIPPED → RECEIVED (or CANCELLED)
 */

import { Badge } from "@/components/ui/badge";
import { CheckCircle2, Package, Truck, XCircle } from "lucide-react";
import type { StockTransferStatus } from "@/types/transfer.types";

const STATUS_CONFIG = {
  DRAFT: {
    label: "Draft",
    variant: "secondary" as const,
    icon: Package,
    className: "bg-gray-500 text-white hover:bg-gray-600",
  },
  SHIPPED: {
    label: "Dikirim",
    variant: "default" as const,
    icon: Truck,
    className: "bg-blue-500 text-white hover:bg-blue-600",
  },
  RECEIVED: {
    label: "Diterima",
    variant: "default" as const,
    icon: CheckCircle2,
    className: "bg-green-500 text-white hover:bg-green-600",
  },
  CANCELLED: {
    label: "Dibatalkan",
    variant: "destructive" as const,
    icon: XCircle,
    className: "bg-red-500 text-white hover:bg-red-600",
  },
} as const;

interface TransferStatusBadgeProps {
  status: StockTransferStatus;
  showIcon?: boolean;
  className?: string;
}

export function TransferStatusBadge({
  status,
  showIcon = true,
  className = ""
}: TransferStatusBadgeProps) {
  const config = STATUS_CONFIG[status];
  const Icon = config.icon;

  return (
    <Badge
      variant={config.variant}
      className={`${config.className} ${className}`}
    >
      {showIcon && <Icon className="mr-1 h-3 w-3" />}
      {config.label}
    </Badge>
  );
}

/**
 * Get status label text without component
 */
export function getStatusLabel(status: StockTransferStatus): string {
  return STATUS_CONFIG[status].label;
}

/**
 * Get status color class for custom styling
 */
export function getStatusColor(status: StockTransferStatus): string {
  const colors = {
    DRAFT: "gray",
    SHIPPED: "blue",
    RECEIVED: "green",
    CANCELLED: "red",
  };
  return colors[status];
}
