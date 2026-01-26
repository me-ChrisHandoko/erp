-- Drop foreign key constraint on delivery_tolerances.product_id
-- This allows using empty string instead of NULL for COMPANY and CATEGORY level tolerances
-- The Product relation will still work via GORM Preload without FK constraint

ALTER TABLE IF EXISTS delivery_tolerances DROP CONSTRAINT IF EXISTS fk_delivery_tolerances_product;
