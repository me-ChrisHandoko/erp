/**
 * Adjustment Audit Log Component
 *
 * Displays audit trail for an inventory adjustment with:
 * - Action history (create, update, approve, cancel, delete)
 * - Timestamp and user information
 * - Expandable rows showing field changes
 */

"use client";

import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ChevronDown,
  ChevronRight,
  Clock,
  User,
  History,
  AlertCircle,
} from "lucide-react";
import { useGetAdjustmentAuditLogsQuery } from "@/store/services/adjustmentApi";
import type { AuditLog } from "@/types/audit";
import {
  getActionLabel,
  parseAuditLog,
  getChangedFields,
  formatFieldName,
  formatFieldValue,
  formatRelativeTime,
  formatAuditDate,
} from "@/lib/audit-utils";

interface AdjustmentAuditLogProps {
  adjustmentId: string;
}

// Action badge color mapping
const actionColors: Record<string, string> = {
  CREATE: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  UPDATE: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
  DELETE: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  APPROVE: "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
  CANCEL: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300",
  INVENTORY_ADJUSTMENT_CREATED: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
  INVENTORY_ADJUSTMENT_UPDATED: "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300",
  INVENTORY_ADJUSTMENT_DELETED: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
  INVENTORY_ADJUSTMENT_APPROVED: "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300",
  INVENTORY_ADJUSTMENT_CANCELLED: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-300",
};

// Custom action label mapping for adjustment-specific actions
const adjustmentActionLabels: Record<string, string> = {
  INVENTORY_ADJUSTMENT_CREATED: "Dibuat",
  INVENTORY_ADJUSTMENT_UPDATED: "Diperbarui",
  INVENTORY_ADJUSTMENT_DELETED: "Dihapus",
  INVENTORY_ADJUSTMENT_APPROVED: "Disetujui",
  INVENTORY_ADJUSTMENT_CANCELLED: "Dibatalkan",
};

function getAdjustmentActionLabel(action: string): string {
  return adjustmentActionLabels[action] || getActionLabel(action);
}

export function AdjustmentAuditLog({ adjustmentId }: AdjustmentAuditLogProps) {
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());

  const {
    data: auditLogs,
    isLoading,
    error,
  } = useGetAdjustmentAuditLogsQuery({ adjustmentId, limit: 50 });

  const toggleRow = (id: string) => {
    setExpandedRows((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  if (isLoading) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <History className="h-4 w-4" />
          Riwayat Audit
        </div>
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <History className="h-4 w-4" />
          Riwayat Audit
        </div>
        <div className="flex items-center gap-2 p-4 rounded-md border bg-destructive/10 text-destructive">
          <AlertCircle className="h-4 w-4" />
          <span className="text-sm">Gagal memuat riwayat audit</span>
        </div>
      </div>
    );
  }

  if (!auditLogs || auditLogs.length === 0) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <History className="h-4 w-4" />
          Riwayat Audit
        </div>
        <div className="p-4 rounded-md border bg-muted/30 text-center text-sm text-muted-foreground">
          Belum ada riwayat audit
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
        <History className="h-4 w-4" />
        Riwayat Audit ({auditLogs.length})
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead className="w-[40px]"></TableHead>
              <TableHead className="w-[120px]">Aksi</TableHead>
              <TableHead>Waktu</TableHead>
              <TableHead>Catatan</TableHead>
              <TableHead className="w-[100px]">Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {auditLogs.map((log) => (
              <AuditLogRow
                key={log.id}
                log={log}
                isExpanded={expandedRows.has(log.id)}
                onToggle={() => toggleRow(log.id)}
              />
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

interface AuditLogRowProps {
  log: AuditLog;
  isExpanded: boolean;
  onToggle: () => void;
}

function AuditLogRow({ log, isExpanded, onToggle }: AuditLogRowProps) {
  const parsedLog = parseAuditLog(log);
  const changedFields = getChangedFields(parsedLog.oldValues, parsedLog.newValues);
  const hasChanges = changedFields.length > 0;

  // Status badge color
  const statusColors: Record<string, string> = {
    SUCCESS: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300",
    FAILED: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300",
    PARTIAL: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300",
  };

  return (
    <>
      <TableRow className={isExpanded ? "border-b-0" : ""}>
        <TableCell>
          {hasChanges && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0"
              onClick={onToggle}
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </Button>
          )}
        </TableCell>
        <TableCell>
          <Badge
            variant="secondary"
            className={actionColors[log.action] || ""}
          >
            {getAdjustmentActionLabel(log.action)}
          </Badge>
        </TableCell>
        <TableCell>
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 text-sm">
              <Clock className="h-3 w-3 text-muted-foreground" />
              <span title={formatAuditDate(log.createdAt)}>
                {formatRelativeTime(log.createdAt)}
              </span>
            </div>
            {log.userId && (
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <User className="h-3 w-3" />
                <span className="truncate max-w-[150px]">{log.userId}</span>
              </div>
            )}
          </div>
        </TableCell>
        <TableCell className="text-sm text-muted-foreground">
          {log.notes || "-"}
        </TableCell>
        <TableCell>
          <Badge
            variant="secondary"
            className={statusColors[log.status] || ""}
          >
            {log.status === "SUCCESS"
              ? "Berhasil"
              : log.status === "FAILED"
              ? "Gagal"
              : "Sebagian"}
          </Badge>
        </TableCell>
      </TableRow>

      {hasChanges && isExpanded && (
        <TableRow className="bg-muted/30">
          <TableCell colSpan={5} className="p-0">
            <div className="p-4">
              <div className="text-xs font-medium text-muted-foreground mb-2">
                Perubahan ({changedFields.length} field)
              </div>
              <div className="rounded-md border bg-background">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[150px] text-xs">
                        Field
                      </TableHead>
                      <TableHead className="text-xs">Nilai Lama</TableHead>
                      <TableHead className="text-xs">Nilai Baru</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {changedFields.map((change, idx) => (
                      <TableRow key={idx}>
                        <TableCell className="font-medium text-xs">
                          {formatFieldName(change.field)}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground font-mono whitespace-pre-wrap max-w-[200px]">
                          {formatFieldValue(change.oldValue, change.field)}
                        </TableCell>
                        <TableCell className="text-xs font-mono whitespace-pre-wrap max-w-[200px]">
                          {formatFieldValue(change.newValue, change.field)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  );
}
