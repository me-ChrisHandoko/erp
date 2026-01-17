-- Fix Missing Product Units Migration
-- This script creates base units for all products that don't have any units yet
-- Date: 2026-01-17

-- First, let's see how many products are missing units
SELECT
    COUNT(DISTINCT p.id) as total_products,
    COUNT(DISTINCT pu.product_id) as products_with_units,
    COUNT(DISTINCT p.id) - COUNT(DISTINCT pu.product_id) as products_missing_units
FROM products p
LEFT JOIN product_units pu ON p.id = pu.product_id;

-- Insert base units for products that don't have any units
INSERT INTO product_units (
    id,
    product_id,
    unit_name,
    conversion_rate,
    is_base_unit,
    buy_price,
    sell_price,
    is_active,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() as id,
    p.id as product_id,
    p.base_unit as unit_name,
    1.0 as conversion_rate,
    true as is_base_unit,
    p.base_cost as buy_price,
    p.base_price as sell_price,
    true as is_active,
    NOW() as created_at,
    NOW() as updated_at
FROM products p
WHERE NOT EXISTS (
    SELECT 1
    FROM product_units pu
    WHERE pu.product_id = p.id
);

-- Verify the fix
SELECT
    COUNT(DISTINCT p.id) as total_products,
    COUNT(DISTINCT pu.product_id) as products_with_units,
    COUNT(DISTINCT p.id) - COUNT(DISTINCT pu.product_id) as products_still_missing_units
FROM products p
LEFT JOIN product_units pu ON p.id = pu.product_id;

-- Show sample of created units
SELECT
    p.code,
    p.name,
    p.base_unit,
    pu.unit_name,
    pu.is_base_unit,
    pu.sell_price
FROM products p
INNER JOIN product_units pu ON p.id = pu.product_id
WHERE pu.created_at > NOW() - INTERVAL '1 minute'
ORDER BY p.code
LIMIT 10;
