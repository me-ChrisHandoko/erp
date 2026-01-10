# Initial Stock Setup - Documentation

## 📋 Table of Contents
1. [⚠️ CRITICAL: Anti-Pattern Warning](#️-critical-anti-pattern-warning)
2. [Konsep Create & Delete Warehouse Stock](#konsep-create--delete-warehouse-stock)
3. [Initial Stock Setup Overview](#initial-stock-setup-overview)
4. [Menu Placement Analysis](#menu-placement-analysis)
5. [✅ Recommended Implementation (Single Entry Point)](#-recommended-implementation-single-entry-point)
6. [Common Mistakes to Avoid](#common-mistakes-to-avoid)
7. [UI/UX Design](#uiux-design)
8. [Technical Implementation](#technical-implementation)
9. [API Endpoints](#api-endpoints)
10. [Action Items](#action-items)

---

## ⚠️ CRITICAL: Anti-Pattern Warning

### 🚨 DO NOT Duplicate "Setup Awal" Across Multiple Menus

**WRONG APPROACH** ❌:
```
Master Data menu → Setup Awal implementation (file A)
Persediaan menu → Setup Awal implementation (file B)
Pengaturan menu → Setup Awal implementation (file C)
```

This violates fundamental software engineering principles:

#### 1. **DRY Principle Violation** (Don't Repeat Yourself)
- **Problem**: Multiple implementations of the same feature
- **Consequence**: Code duplication, inconsistent behavior, maintenance nightmare
- **Example**: Bug fix in one location requires fixing in 2+ other locations

#### 2. **Single Source of Truth Violation**
- **Problem**: No clear "owner" of the feature
- **Consequence**: Conflicting logic, divergent behavior, user confusion
- **Example**: Warehouse A setup works differently than Warehouse B setup

#### 3. **Maintenance Overhead**
- **Problem**: Every change requires updating 3+ locations
- **Consequence**: Increased development time, higher bug risk, technical debt
- **Example**: Adding validation rule requires 3x the effort

#### 4. **User Experience Issues**
- **Problem**: Same feature appears in different menus with potentially different behavior
- **Consequence**: User confusion, loss of trust, support burden
- **Example**: User trains staff on Master Data flow, but Inventory flow differs

### ✅ CORRECT APPROACH: Single Entry Point Pattern

**RIGHT WAY**:
```typescript
// ONE implementation file
/inventory/initial-setup → InitialStockSetup component (SINGLE SOURCE OF TRUTH)

// Multiple entry points (JUST LINKS)
Master Data → Link to /inventory/initial-setup?context=warehouse
Inventory menu → Link to /inventory/initial-setup
Settings → Link to /inventory/initial-setup?context=onboarding
Dashboard widget → Link to /inventory/initial-setup?source=dashboard
```

**Real-World Analogy**: Amazon's "Add to Cart"
- ONE implementation of cart logic
- Button appears on: product page, search results, wish list, recommendations
- ALL buttons link to SAME cart system
- Consistent behavior everywhere

### Key Principles
1. **ONE Implementation**: Single codebase for feature logic
2. **Multiple Entry Points**: Links from different contexts
3. **Context-Aware Navigation**: Pass context params, don't duplicate logic
4. **Consistent Behavior**: Same UX regardless of entry point

---

## Konsep Create & Delete Warehouse Stock

### 🎯 CREATE - Bagaimana Stock Record Dibuat?

Warehouse Stock **TIDAK dibuat manual** oleh user, tapi **otomatis oleh sistem** saat transaksi inventory terjadi.

#### Skenario Auto-Create Stock Record:

##### 1. Goods Receipt (Penerimaan Barang) - Pertama Kali
```
Alur:
1. User buat Purchase Order ke supplier
2. Barang datang → User create Goods Receipt
3. SISTEM OTOMATIS:
   - Cek: Apakah produk ini sudah ada di gudang?
   - Jika BELUM ada → CREATE warehouse_stock record baru
   - Jika SUDAH ada → UPDATE quantity yang existing
   - CREATE inventory_movement record (audit trail)
```

**Backend Logic Example:**
```go
// Saat Goods Receipt diproses
func ProcessGoodsReceipt(productID, warehouseID, qty) {
    // 1. Cari atau buat warehouse stock
    stock := FindOrCreateWarehouseStock(productID, warehouseID)

    // 2. Update quantity
    oldQty := stock.Quantity
    stock.Quantity += qty

    // 3. Save dan create movement
    db.Save(&stock)
    CreateInventoryMovement(stock, "GOODS_RECEIPT", oldQty, stock.Quantity)
}
```

##### 2. Stock Transfer Masuk
```
Alur:
1. User transfer stock dari Gudang A → Gudang B
2. SISTEM OTOMATIS:
   - Gudang A: UPDATE (kurangi qty)
   - Gudang B: CREATE/UPDATE (tambah qty)
   - Jika produk belum ada di Gudang B → CREATE new stock record
```

##### 3. Initial Stock Setup (One-time)
```
Alur:
1. Setup awal warehouse baru
2. Admin import/input initial stock
3. SISTEM OTOMATIS:
   - CREATE warehouse_stock untuk setiap produk
   - CREATE inventory_movement type "INITIAL_STOCK"
```

#### Database Structure:
```sql
-- warehouse_stocks table
-- Unique constraint: (warehouse_id, product_id)
-- Artinya: 1 produk hanya punya 1 record per gudang

CREATE TABLE warehouse_stocks (
    id VARCHAR(255) PRIMARY KEY,
    warehouse_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    quantity DECIMAL(15,3) DEFAULT 0,
    minimum_stock DECIMAL(15,3) DEFAULT 0,
    maximum_stock DECIMAL(15,3) DEFAULT 0,
    location VARCHAR(100),
    last_count_date TIMESTAMP,
    last_count_qty DECIMAL(15,3),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(warehouse_id, product_id)  -- ← Key constraint!
);
```

### 🗑️ DELETE - Kenapa Tidak Ada?

Warehouse Stock **TIDAK bisa dihapus** karena prinsip audit trail dan inventory management.

#### Alasan Tidak Ada Delete:

##### 1. Audit Trail Requirement
```
❌ TIDAK BOLEH:
- Hapus stock record = Hilang history
- Tidak bisa trace: "Kapan produk ini pernah ada di gudang?"
- Tidak bisa audit: "Berapa nilai inventory yang pernah ada?"

✅ YANG BENAR:
- Stock record tetap ada (quantity bisa = 0)
- Semua perubahan tercatat di inventory_movements
- Full audit trail dari awal sampai akhir
```

##### 2. Regulatory Compliance
```
Sistem ERP harus comply dengan:
- Standar Akuntansi Indonesia (SAK)
- Audit perpajakan (NPWP, PPN tracking)
- Stock opname requirements

Semua transaksi inventory HARUS:
- Traceable (bisa dilacak)
- Auditable (bisa diaudit)
- Immutable (tidak bisa dihapus/diubah sembarangan)
```

##### 3. Zero Quantity ≠ Delete
```
Jika stok habis (quantity = 0):
✅ Stock record TETAP ADA
✅ Masih bisa track:
   - Kapan terakhir ada stok?
   - Berapa minimum/maximum stock settings?
   - Dimana lokasi rack terakhir?
   - History pergerakan stok
```

#### Alternatif: "Soft Delete" / Inactive

Jika benar-benar ingin "menghilangkan" stock dari view:

**Option 1: Archive (Recommended)**
```typescript
// Tambah field is_active
interface WarehouseStock {
  isActive: boolean; // Default: true
}

// User bisa "archive" stock yang sudah tidak terpakai
// Stock record tetap ada, tapi hidden dari UI
```

**Option 2: Move to History**
```sql
-- Jika produk discontinued
UPDATE warehouse_stocks
SET is_active = false,
    discontinued_at = NOW()
WHERE product_id = 'xxx';

-- UI hanya show is_active = true
-- Admin bisa access archived stocks
```

---

## Initial Stock Setup Overview

### Kapan Digunakan?

1. **Setup warehouse baru** - Warehouse baru dibuat dan perlu diisi stok awal
2. **Migration data** - Pindah dari sistem lama ke sistem baru
3. **Re-setup after disaster** - Setelah data loss atau corruption
4. **Seasonal warehouse** - Gudang musiman yang dibuka/tutup

### Siapa Yang Menggunakan?

- **OWNER** - Full access
- **ADMIN** - Full access
- **WAREHOUSE_MANAGER** - Limited access (hanya warehouse mereka)
- **STAFF** - No access

### Karakteristik:

- ✅ **One-time operation** per warehouse (idealnya)
- ✅ **Bulk input** - Bisa input banyak produk sekaligus
- ✅ **Create stock records** - Auto-create warehouse_stock
- ✅ **Create movements** - Type: "INITIAL_STOCK"
- ⚠️ **Tidak bisa undo** - Operasi permanent (tapi bisa adjust nanti)
- ⚠️ **Require approval** - Untuk nilai besar (optional)

---

## Menu Placement Analysis

### ⚠️ IMPORTANT: Single Implementation Principle

Before reviewing options, understand this critical principle:

**There should be ONE implementation of "Setup Stok Awal" feature**. Different menu locations should LINK to this single implementation, not create separate implementations.

### Analysis of Placement Options

Below are common placement ideas. Note: These describe WHERE to place LINKS/BUTTONS, not where to duplicate implementations.

---

#### ❌ Option 1: Separate Implementation per Menu Location

**ANTI-PATTERN - DO NOT DO THIS**:
```
📦 Persediaan → /inventory/initial-setup (Implementation A)
📦 Master Data → /master/warehouses/:id/setup (Implementation B)
⚙️ Pengaturan → /settings/initial-setup (Implementation C)
```

**Why This is Wrong:**
- Violates DRY principle
- Creates maintenance nightmare
- Leads to inconsistent behavior
- Confuses users with different UX per location

**Status:** ❌ **REJECTED** - Fundamental violation of best practices

---

#### ✅ Option 2: Single Implementation with Multiple Entry Points

**CORRECT PATTERN**:
```
SINGLE IMPLEMENTATION:
/inventory/initial-setup → InitialStockSetup component

MULTIPLE ENTRY POINTS (just links/buttons):
├── 📦 Persediaan menu → Link to /inventory/initial-setup
├── 📦 Master Data → Link to /inventory/initial-setup?warehouse={id}
├── ⚙️ Pengaturan → Link to /inventory/initial-setup?context=onboarding
└── 🏠 Dashboard widget → Link to /inventory/initial-setup?source=dashboard
```

**Why This is Right:**
- Single source of truth
- Consistent behavior everywhere
- Easy maintenance (fix once, works everywhere)
- Context-aware through URL parameters

**Status:** ✅ **RECOMMENDED** - Best practice pattern

---

### Entry Point Locations (Where to Place Links)

These describe where users can ACCESS the feature (not where to implement it):

#### 1. Primary Entry: Persediaan Menu
```
📦 Persediaan
├── 📦 Stok Barang (/inventory/stock)
├── 🔄 Transfer Gudang (/inventory/transfers)
├── 📋 Stock Opname (/inventory/opname)
├── ✏️ Penyesuaian (/inventory/adjustments)
└── ⚙️ Setup Stok Awal → Link to /inventory/initial-setup
```
**Rationale**: Natural grouping with other inventory operations

#### 2. Contextual Entry: Warehouse Detail
```
📦 Master Data → 🏢 Gudang → Detail Warehouse
└── Button: "Setup Stok Awal"
    → Link to /inventory/initial-setup?warehouse={id}
```
**Rationale**: Context-aware action from warehouse management

#### 3. Onboarding Entry: Setup Wizard
```
🎯 Onboarding Step 4
└── "Setup Stok Awal per Gudang"
    → Link to /inventory/initial-setup?context=onboarding
```
**Rationale**: Guided first-time setup flow

#### 4. Dashboard Entry: Status Widget
```
🏠 Dashboard → Warehouse Status Card
└── Button: "Setup Diperlukan"
    → Link to /inventory/initial-setup?warehouse={id}&source=dashboard
```
**Rationale**: Quick action from monitoring view

---

## ✅ Recommended Implementation (Single Entry Point)

### Architecture Overview

**Core Principle**: ONE implementation, MULTIPLE entry points

```typescript
┌─────────────────────────────────────────────────────────────┐
│ SINGLE IMPLEMENTATION (Source of Truth)                     │
│                                                              │
│ /inventory/initial-setup                                    │
│ └── InitialStockSetup component                             │
│     ├── State management                                    │
│     ├── Validation logic                                    │
│     ├── API calls                                           │
│     └── UI rendering                                        │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ (All routes link here)
          ┌─────────────────┼─────────────────┐
          │                 │                 │
┌─────────┴────────┐ ┌──────┴──────┐ ┌───────┴────────┐
│ Persediaan Menu  │ │ Warehouse   │ │ Dashboard      │
│ → Link with      │ │ Detail      │ │ Widget         │
│   no params      │ │ → Link with │ │ → Link with    │
│                  │ │   ?warehouse│ │   ?source      │
└──────────────────┘ └─────────────┘ └────────────────┘
```

### Implementation Strategy

#### A. Single Route Implementation
```typescript
// File: app/(app)/inventory/initial-setup/page.tsx
// This is the ONLY implementation of the feature

export default async function InitialSetupPage({
  searchParams,
}: {
  searchParams: { warehouse?: string; context?: string; source?: string };
}) {
  const session = await requireAuth();

  // Parse context from URL params
  const warehouseId = searchParams.warehouse;
  const context = searchParams.context; // 'onboarding' | 'warehouse' | undefined
  const source = searchParams.source; // 'dashboard' | undefined

  // Render with context awareness
  return (
    <InitialSetupClient
      initialWarehouseId={warehouseId}
      context={context}
      source={source}
    />
  );
}
```

#### B. Multiple Entry Points (Just Links)

**1. Persediaan Menu Entry**
```typescript
// File: components/app-sidebar.tsx
// Primary navigation entry in "Persediaan" section

{
  title: "Setup Stok Awal",
  url: "/inventory/initial-setup", // ← Simple link, no duplication
  icon: Settings,
}
```

**2. Warehouse Detail Entry**
```typescript
// File: app/(app)/master/warehouses/[id]/page.tsx
// Contextual action button

<Button onClick={() => router.push(`/inventory/initial-setup?warehouse=${warehouseId}`)}>
  Setup Stok Awal
</Button>
```

**3. Dashboard Widget Entry**
```typescript
// File: components/dashboard/warehouse-status-widget.tsx
// Quick action from dashboard

<Link href={`/inventory/initial-setup?warehouse=${warehouse.id}&source=dashboard`}>
  <Button>Setup Diperlukan</Button>
</Link>
```

**4. Onboarding Flow Entry**
```typescript
// File: app/(app)/onboarding/page.tsx
// Part of guided setup

<Link href="/inventory/initial-setup?context=onboarding">
  <Button>Setup Stok Awal Gudang</Button>
</Link>
```

### Context-Aware Behavior

The single implementation adapts based on URL parameters:

```typescript
// File: app/(app)/inventory/initial-setup/initial-setup-client.tsx

export function InitialSetupClient({
  initialWarehouseId,
  context,
  source,
}: InitialSetupClientProps) {
  // Context-aware behavior
  useEffect(() => {
    if (initialWarehouseId) {
      // Pre-select warehouse if coming from warehouse detail
      setSelectedWarehouse(initialWarehouseId);
      setCurrentStep("input-method"); // Skip warehouse selection
    }

    if (context === "onboarding") {
      // Show onboarding-specific messaging
      setOnboardingMode(true);
    }

    if (source === "dashboard") {
      // Track analytics: user came from dashboard widget
      trackEvent("initial_setup_start", { source: "dashboard" });
    }
  }, [initialWarehouseId, context, source]);

  // Rest of component logic (SAME for all entry points)
}
```

### Struktur Menu Lengkap (Actual Implementation)

```
🏠 Dashboard (/dashboard)
└── Warehouse Status Widget → Link: /inventory/initial-setup?warehouse={id}&source=dashboard

🏢 Perusahaan
├── Profil Perusahaan (/company/profile)
├── Rekening Bank (/company/banks)
└── Tim & Pengguna (/company/team)

📦 Master Data
├── 👥 Pelanggan (/master/customers) ← Sudah ada
├── 🏭 Pemasok (/master/suppliers) ← Sudah ada
├── 📦 Produk (/master/products) ← Sudah ada
└── 🏢 Gudang (/master/warehouses) ← Sudah ada
    └── Detail: Button "Setup Stok Awal" → Link: /inventory/initial-setup?warehouse={id}

📦 Persediaan
├── 📦 Stok Barang (/inventory/stock) ← Sudah ada ✅
├── 🔄 Transfer Gudang (/inventory/transfers)
├── 📋 Stock Opname (/inventory/opname)
├── ✏️ Penyesuaian (/inventory/adjustments)
└── ⚙️ Setup Stok Awal → Link: /inventory/initial-setup ← NEW! (SINGLE IMPLEMENTATION)

🛒 Pembelian
├── 📋 Purchase Order (/procurement/orders)
├── 📥 Penerimaan Barang (/procurement/receipts)
├── 📄 Faktur Pembelian (/procurement/invoices)
└── 💳 Pembayaran (/procurement/payments)

📈 Penjualan
├── 📋 Sales Order (/sales/orders)
├── 📤 Pengiriman (/sales/deliveries)
├── 📄 Faktur Penjualan (/sales/invoices)
└── 💳 Penerimaan Kas (/sales/payments)

💰 Keuangan
├── 📒 Jurnal Umum (/finance/journal)
├── 🏦 Kas & Bank (/finance/cash-bank)
├── 💸 Biaya (/finance/expenses)
└── 📊 Laporan (/finance/reports)

⚙️ Pengaturan
├── 🔐 Roles & Permissions (/settings/roles)
├── ⚙️ Konfigurasi Sistem (/settings/config)
├── 🎨 Preferensi (/settings/preferences)
└── 🚀 Setup Awal → Link: /inventory/initial-setup?context=onboarding (Optional)
```

**Key Points:**
- ✅ ONE implementation at `/inventory/initial-setup`
- ✅ Multiple links from different menus (context-aware via URL params)
- ✅ Consistent behavior regardless of entry point
- ✅ Easy to maintain (fix once, works everywhere)
- ✅ Menu structure matches actual `app-sidebar.tsx` implementation

---

## Common Mistakes to Avoid

### ❌ Mistake 1: Creating Separate Page Per Menu

**Wrong:**
```typescript
// DON'T DO THIS
app/(app)/inventory/initial-setup/page.tsx          // Implementation A
app/(app)/master/warehouses/initial-setup/page.tsx  // Implementation B
app/(app)/settings/initial-setup/page.tsx           // Implementation C
```

**Right:**
```typescript
// DO THIS INSTEAD
app/(app)/inventory/initial-setup/page.tsx          // ONE implementation

// Other locations just link to it:
<Link href="/inventory/initial-setup?warehouse={id}">Setup</Link>
```

---

### ❌ Mistake 2: Copy-Pasting Component Logic

**Wrong:**
```typescript
// components/inventory/initial-setup-form.tsx
export function InventoryInitialSetupForm() { /* logic */ }

// components/master/warehouse-setup-form.tsx
export function WarehouseSetupForm() { /* SAME logic copy-pasted */ }
```

**Right:**
```typescript
// components/initial-setup/initial-setup-form.tsx
export function InitialSetupForm() { /* logic ONCE */ }

// Used in ONE page, accessed from multiple entry points
```

---

### ❌ Mistake 3: Different Validation Rules Per Location

**Wrong:**
```typescript
// Inventory path: requires min stock
if (location === 'inventory') {
  validateMinStock(true);
}

// Master path: doesn't require min stock
if (location === 'master') {
  validateMinStock(false);
}
```

**Right:**
```typescript
// ALWAYS same validation rules (single source of truth)
validateMinStock(true); // Consistent everywhere
```

---

### ❌ Mistake 4: Inconsistent UI/UX Per Entry Point

**Wrong:**
```typescript
// Different forms based on where user came from
if (source === 'inventory') {
  return <SevenStepWizard />; // 7 steps
}
if (source === 'master') {
  return <ThreeStepForm />;   // 3 steps - INCONSISTENT!
}
```

**Right:**
```typescript
// SAME UX regardless of entry point
return <SevenStepWizard />; // Always 7 steps

// Only adapt pre-filled data based on context:
if (warehouseId) {
  setSelectedWarehouse(warehouseId); // Pre-select
}
```

---

### ❌ Mistake 5: Not Using URL Parameters for Context

**Wrong:**
```typescript
// Separate routes with duplicate logic
/inventory/initial-setup              → Component A
/master/warehouses/setup              → Component B (duplicate)
/settings/onboarding/stock-setup      → Component C (duplicate)
```

**Right:**
```typescript
// ONE route with context params
/inventory/initial-setup?warehouse={id}&context=onboarding&source=dashboard

// Parse params in component:
const { warehouse, context, source } = searchParams;
```

---

### ✅ Best Practice Checklist

Before implementing, verify:

- [ ] **Single Implementation**: Only ONE page file for the feature
- [ ] **Multiple Links**: Other locations just link with URL params
- [ ] **Consistent Validation**: Same rules everywhere
- [ ] **Consistent UX**: Same flow regardless of entry point
- [ ] **Context-Aware**: Use URL params to adapt behavior
- [ ] **Code Reuse**: No copy-pasted logic
- [ ] **Testing**: Test all entry points lead to same behavior
- [ ] **Documentation**: Clear explanation of single entry point pattern

---

## UI/UX Design

### 1. Dashboard Widget - Warehouse Status

```typescript
// Component: DashboardWarehouseStatus.tsx
// Location: Dashboard page

interface WarehouseStatusCardProps {
  warehouse: {
    id: string;
    name: string;
    hasInitialStock: boolean;
    totalProducts: number;
    totalValue: string;
  };
}

Layout:
┌─────────────────────────────────────┐
│ 🏢 Gudang Utama              [⋮]   │
├─────────────────────────────────────┤
│ Status: ⚠️ Setup Diperlukan        │
│                                      │
│ Belum ada stok di gudang ini.      │
│ Lakukan setup stok awal untuk      │
│ memulai operasi inventory.         │
│                                      │
│ [🚀 Setup Stok Awal]               │
└─────────────────────────────────────┘

atau jika sudah setup:

┌─────────────────────────────────────┐
│ 🏢 Gudang Utama              [⋮]   │
├─────────────────────────────────────┤
│ Status: ✅ Aktif                    │
│                                      │
│ 📦 125 Produk                       │
│ 💰 Nilai: Rp 450,000,000           │
│ ⚠️  15 Stok Rendah                  │
│                                      │
│ [Lihat Detail] [Laporan]           │
└─────────────────────────────────────┘
```

### 2. Initial Setup Page - Step 1: Warehouse Selection

```typescript
// Route: /inventory/initial-setup
// Component: InitialStockSetupPage.tsx

Layout:
┌─────────────────────────────────────────────────────┐
│ Setup Stok Awal                                     │
├─────────────────────────────────────────────────────┤
│ Pilih gudang untuk setup stok awal:                │
│                                                      │
│ ┌──────────────────────────────────────────────┐  │
│ │ 🏢 Gudang Utama                              │  │
│ │ Kota: Jakarta | Kapasitas: 1000m²           │  │
│ │ Status: ⚠️ Belum ada stok                   │  │
│ │                                [Pilih]       │  │
│ ├──────────────────────────────────────────────┤  │
│ │ 🏢 Gudang Cabang Surabaya                   │  │
│ │ Kota: Surabaya | Kapasitas: 500m²          │  │
│ │ Status: ✅ 45 produk                        │  │
│ │                                [View]        │  │
│ ├──────────────────────────────────────────────┤  │
│ │ 🏢 Gudang Transit                            │  │
│ │ Kota: Bandung | Kapasitas: 200m²           │  │
│ │ Status: ⚠️ Belum ada stok                   │  │
│ │                                [Pilih]       │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ [Batal]                                             │
└─────────────────────────────────────────────────────┘
```

### 3. Step 2: Input Method Selection

```typescript
// After selecting warehouse

Layout:
┌─────────────────────────────────────────────────────┐
│ Setup Stok Awal - Gudang Utama                      │
├─────────────────────────────────────────────────────┤
│ Pilih metode input stok:                            │
│                                                      │
│ ┌──────────────────────────────────────────────┐  │
│ │ ⌨️  Manual Entry                             │  │
│ │ Input satu per satu dengan form             │  │
│ │ Cocok untuk: < 50 produk                    │  │
│ │                                [Pilih]       │  │
│ ├──────────────────────────────────────────────┤  │
│ │ 📊 Excel Import (Bulk)                      │  │
│ │ Upload file Excel dengan template           │  │
│ │ Cocok untuk: > 50 produk                    │  │
│ │                                [Pilih]       │  │
│ ├──────────────────────────────────────────────┤  │
│ │ 📋 Copy dari Gudang Lain                    │  │
│ │ Salin stok dari gudang existing             │  │
│ │ Cocok untuk: Gudang dengan produk serupa   │  │
│ │                                [Pilih]       │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ [Back]                                              │
└─────────────────────────────────────────────────────┘
```

### 4. Step 3A: Manual Entry

```typescript
// Manual entry form

Layout:
┌─────────────────────────────────────────────────────┐
│ Setup Stok Awal - Manual Entry                      │
│ Gudang: Gudang Utama                                │
├─────────────────────────────────────────────────────┤
│ Search & Add Produk: [_________________] 🔍         │
│                                                      │
│ Selected Products (0/∞):                            │
│ ┌──────────────────────────────────────────────┐  │
│ │                                                │  │
│ │    Belum ada produk yang dipilih              │  │
│ │    Gunakan search box untuk menambah produk   │  │
│ │                                                │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ [Back] [Save as Draft] [Next: Review]              │
└─────────────────────────────────────────────────────┘

// After adding products:

┌─────────────────────────────────────────────────────┐
│ Setup Stok Awal - Manual Entry                      │
│ Gudang: Gudang Utama                                │
├─────────────────────────────────────────────────────┤
│ Search & Add Produk: [_________________] 🔍         │
│                                                      │
│ Selected Products (2):                              │
│ ┌──────────────────────────────────────────────┐  │
│ │ ☑ BERAS-001 - Beras Premium 20kg       [❌] │  │
│ │   Qty: [100____] Karung              *req    │  │
│ │   Cost: [Rp 185,000] /unit           *req    │  │
│ │   Location: [RAK-A-01________]               │  │
│ │   Min Stock: [20_] | Max: [200_]             │  │
│ │   Notes: [________________________]          │  │
│ ├──────────────────────────────────────────────┤  │
│ │ ☑ GULA-002 - Gula Pasir 1kg            [❌] │  │
│ │   Qty: [500____] PCS                 *req    │  │
│ │   Cost: [Rp 14,500] /unit            *req    │  │
│ │   Location: [RAK-B-03________]               │  │
│ │   Min Stock: [100] | Max: [1000]             │  │
│ │   Notes: [________________________]          │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ Total Value: Rp 25,750,000                          │
│                                                      │
│ [Back] [Save as Draft] [Next: Review]              │
└─────────────────────────────────────────────────────┘
```

### 5. Step 3B: Excel Import

```typescript
// Excel import interface

Layout:
┌─────────────────────────────────────────────────────┐
│ Setup Stok Awal - Excel Import                      │
│ Gudang: Gudang Utama                                │
├─────────────────────────────────────────────────────┤
│ 1. Download Template:                               │
│    [📥 Download Excel Template]                     │
│                                                      │
│ 2. Fill Template:                                   │
│    • Product Code (required)                        │
│    • Quantity (required)                            │
│    • Cost per Unit (required)                       │
│    • Location (optional)                            │
│    • Min Stock (optional)                           │
│    • Max Stock (optional)                           │
│    • Notes (optional)                               │
│                                                      │
│ 3. Upload File:                                     │
│    ┌─────────────────────────────────────────┐    │
│    │                                          │    │
│    │      Drag & Drop Excel file here        │    │
│    │              or                          │    │
│    │      [Choose File]                       │    │
│    │                                          │    │
│    │   Supported: .xlsx, .xls (max 5MB)      │    │
│    └─────────────────────────────────────────┘    │
│                                                      │
│ [Back] [Upload & Validate]                         │
└─────────────────────────────────────────────────────┘

// After upload validation:

┌─────────────────────────────────────────────────────┐
│ Setup Stok Awal - Validation Result                 │
├─────────────────────────────────────────────────────┤
│ File: stock_awal_gudang_utama.xlsx                  │
│                                                      │
│ ✅ Validation Summary:                              │
│ • Total Rows: 125                                   │
│ • Valid: 120 rows                                   │
│ • Errors: 5 rows                                    │
│                                                      │
│ ❌ Errors Found:                                    │
│ ┌──────────────────────────────────────────────┐  │
│ │ Row 12: Product code 'ABC-999' not found     │  │
│ │ Row 45: Quantity must be > 0                 │  │
│ │ Row 67: Invalid cost format                  │  │
│ │ Row 89: Max stock < Min stock                │  │
│ │ Row 103: Product code 'XYZ-111' not found    │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ [Back] [Fix & Re-upload] [Proceed with 120 Valid] │
└─────────────────────────────────────────────────────┘
```

### 6. Step 4: Review & Confirm

```typescript
// Review summary before submit

Layout:
┌─────────────────────────────────────────────────────┐
│ Konfirmasi Setup Stok Awal                          │
├─────────────────────────────────────────────────────┤
│ Gudang: Gudang Utama                                │
│ Metode: Manual Entry                                │
│                                                      │
│ 📊 Summary:                                         │
│ ├─ Total Produk: 25 items                          │
│ ├─ Total Quantity: 2,450 units                     │
│ └─ Total Nilai: Rp 450,000,000                     │
│                                                      │
│ Top 5 Products:                                     │
│ ┌──────────────────────────────────────────────┐  │
│ │ 1. Beras Premium 20kg   100 K  Rp 18,500,000 │  │
│ │ 2. Gula Pasir 1kg       500 P  Rp  7,250,000 │  │
│ │ 3. Minyak Goreng 2L     300 B  Rp 15,600,000 │  │
│ │ 4. Tepung Terigu 1kg    400 P  Rp  4,800,000 │  │
│ │ 5. Kopi Bubuk 100g      250 P  Rp  3,750,000 │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ ⚠️  PERHATIAN:                                      │
│ Operasi ini akan:                                   │
│ • Create 25 warehouse_stock records                 │
│ • Create 25 inventory_movement records              │
│ • Update inventory value: +Rp 450,000,000          │
│ • Operasi tidak bisa di-undo                       │
│                                                      │
│ Saya memahami dan ingin melanjutkan:               │
│ ☐ Ya, saya sudah cek data dengan teliti            │
│                                                      │
│ [Back] [✓ Confirm & Submit]                        │
└─────────────────────────────────────────────────────┘
```

### 7. Step 5: Success & Next Steps

```typescript
// After successful submission

Layout:
┌─────────────────────────────────────────────────────┐
│ ✅ Setup Stok Awal Berhasil!                        │
├─────────────────────────────────────────────────────┤
│                                                      │
│         🎉 Selamat!                                 │
│    Stok awal berhasil disimpan                      │
│                                                      │
│ 📊 Summary:                                         │
│ • 25 produk telah ditambahkan                       │
│ • Gudang: Gudang Utama                              │
│ • Total Nilai: Rp 450,000,000                       │
│                                                      │
│ 📝 Yang bisa Anda lakukan selanjutnya:              │
│ ┌──────────────────────────────────────────────┐  │
│ │ [📦 Lihat Daftar Stok]                       │  │
│ │ [📊 View Movement History]                   │  │
│ │ [📄 Print Summary Report]                    │  │
│ │ [🏠 Back to Dashboard]                       │  │
│ └──────────────────────────────────────────────┘  │
│                                                      │
│ 💡 Tips:                                            │
│ • Lakukan stock opname berkala                      │
│ • Set up reorder points untuk auto-notification    │
│ • Configure warehouse locations untuk efisiensi    │
│                                                      │
└─────────────────────────────────────────────────────┘
```

---

## Technical Implementation

### File Structure

```
src/
├── app/(app)/inventory/
│   ├── stock/                    # Existing
│   │   ├── page.tsx
│   │   ├── stock-client.tsx
│   │   └── ...
│   └── initial-setup/            # NEW
│       ├── page.tsx              # Server component (requireAuth)
│       ├── initial-setup-client.tsx
│       ├── step-warehouse-selection.tsx
│       ├── step-input-method.tsx
│       ├── step-manual-entry.tsx
│       ├── step-excel-import.tsx
│       ├── step-review.tsx
│       └── step-success.tsx
│
├── components/
│   ├── stock/                    # Existing
│   │   ├── stock-table.tsx
│   │   └── ...
│   ├── initial-setup/            # NEW
│   │   ├── warehouse-card.tsx
│   │   ├── product-selector.tsx
│   │   ├── stock-input-form.tsx
│   │   ├── excel-uploader.tsx
│   │   └── setup-summary.tsx
│   └── dashboard/
│       └── warehouse-status-widget.tsx  # NEW
│
├── store/services/
│   ├── stockApi.ts              # Existing
│   └── initialStockApi.ts       # NEW
│
└── types/
    ├── stock.types.ts           # Existing
    └── initial-stock.types.ts   # NEW
```

### Types Definition

```typescript
// types/initial-stock.types.ts

export interface InitialStockSetupRequest {
  warehouseId: string;
  items: InitialStockItem[];
  notes?: string;
}

export interface InitialStockItem {
  productId: string;
  quantity: string; // decimal as string
  costPerUnit: string; // decimal as string
  location?: string;
  minimumStock?: string;
  maximumStock?: string;
  notes?: string;
}

export interface InitialStockValidationError {
  row: number;
  field: string;
  message: string;
  value?: any;
}

export interface InitialStockImportResult {
  totalRows: number;
  validRows: number;
  errorRows: number;
  errors: InitialStockValidationError[];
  validItems: InitialStockItem[];
}

export interface WarehouseStockStatus {
  warehouseId: string;
  warehouseName: string;
  hasInitialStock: boolean;
  totalProducts: number;
  totalValue: string;
  lastUpdated?: string;
}
```

### API Service

```typescript
// store/services/initialStockApi.ts

import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "./authApi";
import type {
  InitialStockSetupRequest,
  InitialStockImportResult,
  WarehouseStockStatus,
} from "@/types/initial-stock.types";

export const initialStockApi = createApi({
  reducerPath: "initialStockApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["InitialStock", "WarehouseStatus"],
  endpoints: (builder) => ({
    /**
     * Get Warehouse Stock Status
     * Check if warehouse has initial stock setup
     */
    getWarehouseStockStatus: builder.query<WarehouseStockStatus[], void>({
      query: () => "/warehouses/stock-status",
      providesTags: ["WarehouseStatus"],
    }),

    /**
     * Submit Initial Stock Setup
     * Create warehouse_stock records and inventory_movements
     */
    submitInitialStock: builder.mutation<
      { success: boolean; message: string },
      InitialStockSetupRequest
    >({
      query: (data) => ({
        url: "/warehouse-stocks/initial-setup",
        method: "POST",
        body: data,
      }),
      invalidatesTags: ["InitialStock", "WarehouseStatus", "StockList"],
    }),

    /**
     * Validate Excel Import
     * Validate imported data before submission
     */
    validateExcelImport: builder.mutation<
      InitialStockImportResult,
      { warehouseId: string; file: File }
    >({
      query: ({ warehouseId, file }) => {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("warehouseId", warehouseId);

        return {
          url: "/warehouse-stocks/validate-import",
          method: "POST",
          body: formData,
        };
      },
    }),

    /**
     * Download Excel Template
     * Get template for bulk import
     */
    downloadExcelTemplate: builder.query<Blob, void>({
      query: () => ({
        url: "/warehouse-stocks/import-template",
        responseHandler: (response) => response.blob(),
      }),
    }),
  }),
});

export const {
  useGetWarehouseStockStatusQuery,
  useSubmitInitialStockMutation,
  useValidateExcelImportMutation,
  useLazyDownloadExcelTemplateQuery,
} = initialStockApi;
```

### Main Component Structure

```typescript
// app/(app)/inventory/initial-setup/initial-setup-client.tsx

"use client";

import { useState } from "react";
import { StepWarehouseSelection } from "./step-warehouse-selection";
import { StepInputMethod } from "./step-input-method";
import { StepManualEntry } from "./step-manual-entry";
import { StepExcelImport } from "./step-excel-import";
import { StepReview } from "./step-review";
import { StepSuccess } from "./step-success";

type SetupStep =
  | "warehouse-selection"
  | "input-method"
  | "manual-entry"
  | "excel-import"
  | "review"
  | "success";

export function InitialSetupClient() {
  const [currentStep, setCurrentStep] = useState<SetupStep>("warehouse-selection");
  const [selectedWarehouse, setSelectedWarehouse] = useState<string | null>(null);
  const [inputMethod, setInputMethod] = useState<"manual" | "excel" | "copy" | null>(null);
  const [stockItems, setStockItems] = useState<InitialStockItem[]>([]);

  const renderStep = () => {
    switch (currentStep) {
      case "warehouse-selection":
        return (
          <StepWarehouseSelection
            onSelect={(warehouseId) => {
              setSelectedWarehouse(warehouseId);
              setCurrentStep("input-method");
            }}
          />
        );

      case "input-method":
        return (
          <StepInputMethod
            onSelect={(method) => {
              setInputMethod(method);
              setCurrentStep(method === "manual" ? "manual-entry" : "excel-import");
            }}
            onBack={() => setCurrentStep("warehouse-selection")}
          />
        );

      case "manual-entry":
        return (
          <StepManualEntry
            warehouseId={selectedWarehouse!}
            initialItems={stockItems}
            onNext={(items) => {
              setStockItems(items);
              setCurrentStep("review");
            }}
            onBack={() => setCurrentStep("input-method")}
          />
        );

      case "excel-import":
        return (
          <StepExcelImport
            warehouseId={selectedWarehouse!}
            onNext={(items) => {
              setStockItems(items);
              setCurrentStep("review");
            }}
            onBack={() => setCurrentStep("input-method")}
          />
        );

      case "review":
        return (
          <StepReview
            warehouseId={selectedWarehouse!}
            items={stockItems}
            onConfirm={() => setCurrentStep("success")}
            onBack={() => setCurrentStep(inputMethod === "manual" ? "manual-entry" : "excel-import")}
          />
        );

      case "success":
        return <StepSuccess warehouseId={selectedWarehouse!} />;

      default:
        return null;
    }
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      {/* Progress Indicator */}
      <div className="flex items-center justify-center gap-2">
        {/* Step indicators */}
      </div>

      {/* Current Step Content */}
      {renderStep()}
    </div>
  );
}
```

---

## API Endpoints

### Backend API Requirements

```go
// POST /api/v1/warehouse-stocks/initial-setup
// Submit initial stock setup for a warehouse

type InitialStockSetupRequest struct {
    WarehouseID string                 `json:"warehouseId" binding:"required"`
    Items       []InitialStockItem     `json:"items" binding:"required,min=1"`
    Notes       *string                `json:"notes"`
}

type InitialStockItem struct {
    ProductID    string `json:"productId" binding:"required"`
    Quantity     string `json:"quantity" binding:"required"`
    CostPerUnit  string `json:"costPerUnit" binding:"required"`
    Location     string `json:"location"`
    MinimumStock string `json:"minimumStock"`
    MaximumStock string `json:"maximumStock"`
    Notes        string `json:"notes"`
}

// Response
type InitialStockSetupResponse struct {
    Success        bool   `json:"success"`
    Message        string `json:"message"`
    TotalItems     int    `json:"totalItems"`
    TotalValue     string `json:"totalValue"`
    CreatedStocks  int    `json:"createdStocks"`
    UpdatedStocks  int    `json:"updatedStocks"`
}
```

```go
// GET /api/v1/warehouses/stock-status
// Get stock status for all warehouses

type WarehouseStockStatusResponse struct {
    Warehouses []WarehouseStockStatus `json:"warehouses"`
}

type WarehouseStockStatus struct {
    WarehouseID     string `json:"warehouseId"`
    WarehouseName   string `json:"warehouseName"`
    HasInitialStock bool   `json:"hasInitialStock"`
    TotalProducts   int    `json:"totalProducts"`
    TotalValue      string `json:"totalValue"`
    LastUpdated     string `json:"lastUpdated"`
}
```

```go
// POST /api/v1/warehouse-stocks/validate-import
// Validate Excel import before submission

// Request: multipart/form-data
// - file: Excel file
// - warehouseId: string

type ValidateImportResponse struct {
    TotalRows  int                    `json:"totalRows"`
    ValidRows  int                    `json:"validRows"`
    ErrorRows  int                    `json:"errorRows"`
    Errors     []ValidationError      `json:"errors"`
    ValidItems []InitialStockItem     `json:"validItems"`
}

type ValidationError struct {
    Row     int    `json:"row"`
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   string `json:"value"`
}
```

```go
// GET /api/v1/warehouse-stocks/import-template
// Download Excel template for bulk import

// Response: Excel file (binary)
// Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// Content-Disposition: attachment; filename="initial_stock_template.xlsx"

// Template columns:
// - Product Code (required)
// - Quantity (required)
// - Cost per Unit (required)
// - Location (optional)
// - Min Stock (optional)
// - Max Stock (optional)
// - Notes (optional)
```

---

## Action Items

### Phase 1: Foundation (Priority: HIGH)
- [ ] Create types: `initial-stock.types.ts`
- [ ] Create API service: `initialStockApi.ts`
- [ ] Add to Redux store configuration
- [ ] Create route: `/inventory/initial-setup`

### Phase 2: Basic UI (Priority: HIGH)
- [ ] Step 1: Warehouse Selection component
- [ ] Step 2: Input Method Selection component
- [ ] Step 3: Manual Entry component
- [ ] Step 4: Review & Confirm component
- [ ] Step 5: Success component

### Phase 3: Excel Import (Priority: MEDIUM)
- [ ] Excel uploader component
- [ ] Excel validation UI
- [ ] Template download functionality
- [ ] Error display for invalid rows

### Phase 4: Dashboard Integration (Priority: MEDIUM)
- [ ] Warehouse status widget
- [ ] "Setup Required" badges
- [ ] Quick action buttons
- [ ] Conditional menu display

### Phase 5: Backend API (Priority: HIGH)
- [ ] POST `/warehouse-stocks/initial-setup` endpoint
- [ ] GET `/warehouses/stock-status` endpoint
- [ ] POST `/warehouse-stocks/validate-import` endpoint
- [ ] GET `/warehouse-stocks/import-template` endpoint
- [ ] Transaction handling (atomic operations)
- [ ] Audit trail creation

### Phase 6: Enhancements (Priority: LOW)
- [ ] Copy from another warehouse feature
- [ ] Draft save functionality
- [ ] Approval workflow (for large values)
- [ ] Email notification on completion
- [ ] Activity log integration

### Phase 7: Testing (Priority: HIGH)
- [ ] Unit tests for validation logic
- [ ] Integration tests for API endpoints
- [ ] E2E tests for complete flow
- [ ] Permission testing (RBAC)
- [ ] Load testing for bulk operations

---

## Summary

### ✅ What We Have
- ✅ Stock listing page (`/inventory/stock`)
- ✅ Stock update settings (min/max/location)
- ✅ Type definitions for stock
- ✅ API service for stock operations
- ✅ Permission-based access control

### 🚀 What We Need (Initial Stock Setup)
- 📋 Initial setup page UI
- 🔌 Backend API endpoints
- 📊 Dashboard integration
- 📥 Excel import/export
- ✅ Validation & error handling
- 📝 Audit trail & logging

### 🎯 Success Criteria
- User dapat setup stok awal untuk warehouse baru
- Bulk import via Excel untuk efisiensi
- Validation yang ketat untuk data integrity
- Full audit trail untuk compliance
- Clear UX untuk guided flow
- Permission-based access control

---

**Document Version:** 2.1
**Last Updated:** 2025-01-09
**Author:** Claude (Anthropic)
**Status:** Ready for Implementation

**Changelog:**
- **v2.1 (2025-01-09)**: Updated menu structure to match actual app-sidebar.tsx implementation ("Inventori" → "Persediaan", added "Perusahaan" section, consolidated menu groupings)
- **v2.0 (2025-01-09)**: Added critical anti-pattern warning, restructured menu placement analysis, emphasized single entry point pattern, added common mistakes section
- **v1.0 (2025-01-09)**: Initial documentation
