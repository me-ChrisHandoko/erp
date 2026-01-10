-- Fix Cross-Tenant Data Contamination
-- WARNING: This will delete warehouse stocks where product and warehouse belong to different companies!

-- 1. First, check how many records will be affected
SELECT
    COUNT(*) as affected_records,
    STRING_AGG(DISTINCT p.code, ', ') as affected_products,
    STRING_AGG(DISTINCT w.name, ', ') as affected_warehouses
FROM warehouse_stocks ws
JOIN products p ON ws.product_id = p.id
JOIN warehouses w ON ws.warehouse_id = w.id
WHERE p.company_id != w.company_id;

-- 2. Show details of records to be deleted
SELECT
    ws.id as stock_id,
    p.code as product_code,
    p.name as product_name,
    pc.name as product_company,
    w.name as warehouse_name,
    wc.name as warehouse_company,
    ws.quantity,
    ws.location
FROM warehouse_stocks ws
JOIN products p ON ws.product_id = p.id
JOIN warehouses w ON ws.warehouse_id = w.id
JOIN companies pc ON p.company_id = pc.id
JOIN companies wc ON w.company_id = wc.id
WHERE p.company_id != w.company_id
ORDER BY w.name, p.code;

-- 3. DELETE cross-tenant data
-- UNCOMMENT THE FOLLOWING LINES TO EXECUTE:
/*
DELETE FROM warehouse_stocks ws
USING products p, warehouses w
WHERE ws.product_id = p.id
  AND ws.warehouse_id = w.id
  AND p.company_id != w.company_id;
*/

-- 4. Verify deletion
/*
SELECT
    COUNT(*) as remaining_cross_tenant_records
FROM warehouse_stocks ws
JOIN products p ON ws.product_id = p.id
JOIN warehouses w ON ws.warehouse_id = w.id
WHERE p.company_id != w.company_id;
*/
