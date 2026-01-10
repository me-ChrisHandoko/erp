-- Debug Query untuk Product ID: c68ea3e2-762e-4b07-ac9e-270bec438562
-- Run these queries to verify product stock status

-- 1. Check product details
SELECT
    id as product_id,
    code as product_code,
    name as product_name,
    is_active,
    company_id,
    created_at
FROM products
WHERE id = 'c68ea3e2-762e-4b07-ac9e-270bec438562';

-- 2. Check all warehouse stocks for this product
SELECT
    ws.id as stock_id,
    w.id as warehouse_id,
    w.name as warehouse_name,
    w.code as warehouse_code,
    p.code as product_code,
    p.name as product_name,
    ws.quantity,
    ws.location,
    ws.company_id as stock_company_id,
    w.company_id as warehouse_company_id,
    p.company_id as product_company_id,
    ws.created_at
FROM warehouse_stocks ws
JOIN warehouses w ON ws.warehouse_id = w.id
JOIN products p ON ws.product_id = p.id
WHERE ws.product_id = 'c68ea3e2-762e-4b07-ac9e-270bec438562'
ORDER BY ws.created_at DESC;

-- 3. Check if this product has stock in specific warehouse
-- (Replace WAREHOUSE_ID with the warehouse you selected in UI)
SELECT
    ws.id as stock_id,
    ws.product_id,
    ws.warehouse_id,
    ws.quantity,
    ws.company_id,
    p.code as product_code,
    p.name as product_name,
    w.name as warehouse_name
FROM warehouse_stocks ws
JOIN products p ON ws.product_id = p.id
JOIN warehouses w ON ws.warehouse_id = w.id
WHERE ws.product_id = 'c68ea3e2-762e-4b07-ac9e-270bec438562'
  AND ws.warehouse_id = 'REPLACE_WITH_YOUR_WAREHOUSE_ID'; -- Replace this!

-- 4. Count total stocks in each warehouse for debugging
SELECT
    w.id as warehouse_id,
    w.name as warehouse_name,
    w.company_id,
    COUNT(ws.id) as total_stocks
FROM warehouses w
LEFT JOIN warehouse_stocks ws ON w.id = ws.warehouse_id
GROUP BY w.id, w.name, w.company_id
ORDER BY w.name;

-- 5. Check if backend query would return this product
-- (This simulates the GET /warehouse-stocks?warehouseID=xxx query)
SELECT
    ws.id,
    ws.product_id,
    ws.warehouse_id,
    ws.quantity,
    ws.location,
    p.code as product_code,
    p.name as product_name
FROM warehouse_stocks ws
JOIN products p ON ws.product_id = p.id
WHERE ws.warehouse_id = 'REPLACE_WITH_YOUR_WAREHOUSE_ID' -- Replace this!
  AND ws.company_id = 'REPLACE_WITH_YOUR_COMPANY_ID' -- Replace this!
ORDER BY p.code ASC
LIMIT 1000;
