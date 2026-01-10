-- Debug Query untuk BRS-006
-- Run these queries in your database client (psql, DBeaver, etc.)

-- 1. Cek apakah produk BRS-006 exist
SELECT
    id as product_id,
    code as product_code,
    name as product_name,
    is_active,
    created_at
FROM products
WHERE code = 'BRS-006';

-- 2. Cek semua stock BRS-006 di semua warehouse
SELECT
    ws.id as stock_id,
    w.id as warehouse_id,
    w.name as warehouse_name,
    p.code as product_code,
    p.name as product_name,
    ws.quantity,
    ws.location,
    ws.created_at
FROM warehouse_stocks ws
JOIN warehouses w ON ws.warehouse_id = w.id
JOIN products p ON ws.product_id = p.id
WHERE p.code = 'BRS-006'
ORDER BY w.name;

-- 3. Cek stock BRS-006 di warehouse spesifik (6cb24a98-bb8d-4283-905b-600c304dfe6c)
SELECT
    ws.id as stock_id,
    w.id as warehouse_id,
    w.name as warehouse_name,
    p.id as product_id,
    p.code as product_code,
    p.name as product_name,
    ws.quantity,
    ws.location,
    ws.minimum_stock,
    ws.maximum_stock,
    ws.created_at,
    ws.updated_at
FROM warehouse_stocks ws
JOIN warehouses w ON ws.warehouse_id = w.id
JOIN products p ON ws.product_id = p.id
WHERE p.code = 'BRS-006'
  AND w.id = '6cb24a98-bb8d-4283-905b-600c304dfe6c';

-- 4. Cek semua stock di warehouse spesifik (untuk compare)
SELECT
    p.code as product_code,
    p.name as product_name,
    ws.quantity,
    ws.location
FROM warehouse_stocks ws
JOIN products p ON ws.product_id = p.id
WHERE ws.warehouse_id = '6cb24a98-bb8d-4283-905b-600c304dfe6c'
ORDER BY p.code;

-- 5. Cek column names di warehouse_stocks table (untuk verifikasi naming)
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'warehouse_stocks'
ORDER BY ordinal_position;

-- 6. Cek raw data warehouse_stocks untuk BRS-006 (semua columns)
SELECT *
FROM warehouse_stocks
WHERE product_id = (SELECT id FROM products WHERE code = 'BRS-006');
