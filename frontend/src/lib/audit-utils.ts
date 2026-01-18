/**
 * Audit Log Utility Functions
 *
 * Helper functions for parsing, formatting, and displaying audit log data
 */

import type {
  AuditLog,
  ParsedAuditLog,
  AuditAction,
  AuditEntityType,
  AuditStatus,
  ChangedField,
} from '@/types/audit';

/**
 * Parse JSON strings in audit log to typed objects
 */
export function parseAuditLog<T = any>(auditLog: AuditLog): ParsedAuditLog<T> {
  return {
    ...auditLog,
    oldValues: auditLog.oldValues ? JSON.parse(auditLog.oldValues) : null,
    newValues: auditLog.newValues ? JSON.parse(auditLog.newValues) : null,
  };
}

/**
 * Get human-readable action label
 */
export function getActionLabel(action: AuditAction): string {
  const labels: Record<AuditAction, string> = {
    CREATE: 'Dibuat',
    UPDATE: 'Diperbarui',
    DELETE: 'Dihapus',
    ACTIVATE: 'Diaktifkan',
    DEACTIVATE: 'Dinonaktifkan',
    LOGIN: 'Login',
    LOGOUT: 'Logout',
    ASSIGN: 'Diberikan',
    REVOKE: 'Dicabut',
  };
  return labels[action] || action;
}

/**
 * Get human-readable entity type label
 */
export function getEntityTypeLabel(entityType: AuditEntityType): string {
  const labels: Record<AuditEntityType, string> = {
    product: 'Produk',
    product_supplier: 'Supplier Produk',
    customer: 'Pelanggan',
    supplier: 'Supplier',
    warehouse: 'Gudang',
    user: 'Pengguna',
    role: 'Role',
    bank_account: 'Akun Bank',
    company: 'Perusahaan',
    purchase_order: 'Purchase Order',
    sales_order: 'Sales Order',
    inventory: 'Inventori',
    adjustment: 'Adjustment',
    stock_opname: 'Stok Opname',
  };
  return labels[entityType] || entityType;
}

/**
 * Get status badge color class
 */
export function getStatusColorClass(status: AuditStatus): string {
  const colors: Record<AuditStatus, string> = {
    SUCCESS: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
    FAILED: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
    PARTIAL: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  };
  return colors[status] || '';
}

/**
 * Get status label
 */
export function getStatusLabel(status: AuditStatus): string {
  const labels: Record<AuditStatus, string> = {
    SUCCESS: 'Berhasil',
    FAILED: 'Gagal',
    PARTIAL: 'Sebagian',
  };
  return labels[status] || status;
}

/**
 * Extract changed fields from old and new values
 */
export function getChangedFields<T extends Record<string, any>>(
  oldValues: T | null,
  newValues: T | null
): ChangedField[] {
  if (!oldValues && !newValues) return [];

  const changes: ChangedField[] = [];
  const allKeys = new Set([
    ...Object.keys(oldValues || {}),
    ...Object.keys(newValues || {}),
  ]);

  allKeys.forEach((key) => {
    const oldValue = oldValues?.[key];
    const newValue = newValues?.[key];

    // Check if values are different
    if (JSON.stringify(oldValue) !== JSON.stringify(newValue)) {
      changes.push({
        field: key,
        oldValue,
        newValue,
      });
    }
  });

  return changes;
}

/**
 * Format field name to human-readable label
 */
export function formatFieldName(fieldName: string): string {
  const fieldLabels: Record<string, string> = {
    // Common fields
    code: 'Kode',
    name: 'Nama',
    description: 'Deskripsi',
    isActive: 'Status Aktif',
    createdAt: 'Dibuat Pada',
    updatedAt: 'Diperbarui Pada',

    // Product fields
    category: 'Kategori',
    baseUnit: 'Satuan Dasar',
    baseCost: 'Harga Beli (HPP)',
    basePrice: 'Harga Jual',
    minimumStock: 'Stok Minimum',
    barcode: 'Barcode',
    isBatchTracked: 'Lacak Batch',
    isPerishable: 'Mudah Rusak',
    suppliers: 'Daftar Supplier',
    units: 'Satuan Konversi',

    // Customer/Supplier fields
    type: 'Tipe',
    phone: 'Telepon',
    email: 'Email',
    address: 'Alamat',
    city: 'Kota',
    province: 'Provinsi',
    postalCode: 'Kode Pos',
    npwp: 'NPWP',
    creditLimit: 'Limit Kredit',
    paymentTermDays: 'Termin Pembayaran',

    // Warehouse fields
    capacity: 'Kapasitas',
    manager: 'Manajer',

    // User fields
    username: 'Username',
    fullName: 'Nama Lengkap',
    roleId: 'Role',

    // Bank account fields
    accountNumber: 'Nomor Rekening',
    accountName: 'Nama Rekening',
    bankName: 'Nama Bank',
    branch: 'Cabang',
  };

  return fieldLabels[fieldName] || fieldName;
}

/**
 * Format field value for display
 */
export function formatFieldValue(value: any, fieldName?: string): string {
  if (value === null || value === undefined) return '-';
  if (typeof value === 'boolean') return value ? 'Ya' : 'Tidak';

  // Special handling for suppliers array
  if (fieldName === 'suppliers' && Array.isArray(value)) {
    if (value.length === 0) return 'Tidak ada supplier';
    return value.map((supplier: any) => {
      const price = new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
      }).format(parseFloat(supplier.supplier_price || 0));
      return `${supplier.supplier_name} (${supplier.supplier_code}) - ${price}, Lead Time: ${supplier.lead_time} hari${supplier.is_primary ? ' [PRIMARY]' : ''}`;
    }).join('\n');
  }

  // Special handling for units array
  if (fieldName === 'units' && Array.isArray(value)) {
    if (value.length === 0) return 'Tidak ada satuan';
    return value.map((unit: any) => {
      const buyPrice = unit.buy_price ? new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
      }).format(parseFloat(unit.buy_price)) : '-';
      const sellPrice = unit.sell_price ? new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
      }).format(parseFloat(unit.sell_price)) : '-';
      const baseLabel = unit.is_base_unit ? ' [BASE]' : '';
      return `${unit.unit_name} (1 = ${unit.conversion_rate})${baseLabel} - Beli: ${buyPrice}, Jual: ${sellPrice}`;
    }).join('\n');
  }

  if (typeof value === 'object') return JSON.stringify(value, null, 2);

  // Format currency fields
  if (
    fieldName &&
    (fieldName.includes('price') ||
      fieldName.includes('cost') ||
      fieldName.includes('amount') ||
      fieldName === 'creditLimit')
  ) {
    const numValue = parseFloat(value);
    if (!isNaN(numValue)) {
      return new Intl.NumberFormat('id-ID', {
        style: 'currency',
        currency: 'IDR',
        minimumFractionDigits: 0,
      }).format(numValue);
    }
  }

  return String(value);
}

/**
 * Generate human-readable audit summary
 */
export function generateAuditSummary(auditLog: AuditLog): string {
  const action = getActionLabel(auditLog.action);
  const entityType = auditLog.entityType
    ? getEntityTypeLabel(auditLog.entityType)
    : 'Entitas';

  if (auditLog.action === 'CREATE') {
    return `${entityType} baru berhasil dibuat`;
  }

  if (auditLog.action === 'UPDATE') {
    const parsed = parseAuditLog(auditLog);
    const changes = getChangedFields(parsed.oldValues, parsed.newValues);

    if (changes.length === 0) return `${entityType} diperbarui`;
    if (changes.length === 1) {
      const field = formatFieldName(changes[0].field);
      return `${field} pada ${entityType} diperbarui`;
    }
    return `${changes.length} field pada ${entityType} diperbarui`;
  }

  if (auditLog.action === 'DELETE') {
    return `${entityType} dihapus`;
  }

  if (auditLog.action === 'ACTIVATE') {
    return `${entityType} diaktifkan`;
  }

  if (auditLog.action === 'DEACTIVATE') {
    return `${entityType} dinonaktifkan`;
  }

  return `${action} ${entityType}`;
}

/**
 * Format date for display
 */
export function formatAuditDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}

/**
 * Format relative time (e.g., "2 jam yang lalu")
 */
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) return 'Baru saja';
  if (diffMinutes < 60) return `${diffMinutes} menit yang lalu`;
  if (diffHours < 24) return `${diffHours} jam yang lalu`;
  if (diffDays < 7) return `${diffDays} hari yang lalu`;
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} minggu yang lalu`;
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} bulan yang lalu`;
  return `${Math.floor(diffDays / 365)} tahun yang lalu`;
}

/**
 * Group audit logs by date
 */
export function groupAuditLogsByDate(
  auditLogs: AuditLog[]
): Record<string, AuditLog[]> {
  const grouped: Record<string, AuditLog[]> = {};

  auditLogs.forEach((log) => {
    const date = new Date(log.createdAt);
    const dateKey = date.toISOString().split('T')[0]; // YYYY-MM-DD

    if (!grouped[dateKey]) {
      grouped[dateKey] = [];
    }
    grouped[dateKey].push(log);
  });

  return grouped;
}

/**
 * Filter audit logs by criteria
 */
export function filterAuditLogs(
  auditLogs: AuditLog[],
  filters: {
    action?: AuditAction;
    entityType?: AuditEntityType;
    status?: AuditStatus;
    userId?: string;
    search?: string;
  }
): AuditLog[] {
  return auditLogs.filter((log) => {
    if (filters.action && log.action !== filters.action) return false;
    if (filters.entityType && log.entityType !== filters.entityType) return false;
    if (filters.status && log.status !== filters.status) return false;
    if (filters.userId && log.userId !== filters.userId) return false;

    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      const matchesNotes = log.notes?.toLowerCase().includes(searchLower);
      const matchesEntityId = log.entityId?.toLowerCase().includes(searchLower);

      if (!matchesNotes && !matchesEntityId) return false;
    }

    return true;
  });
}
