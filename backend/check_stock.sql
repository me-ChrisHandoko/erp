-- Check products with warehouse stocks
SELECT 
    p.code as product_code,
    p.name as product_name,
    w.name as warehouse_name,
    ws.quantity,
    w.is_active as warehouse_active
FROM products p
LEFT JOIN warehouse_stocks ws ON p.id = ws.product_id
LEFT JOIN warehouses w ON ws.warehouse_id = w.id
WHERE p.is_active = true
ORDER BY p.code, w.name
LIMIT 10;
