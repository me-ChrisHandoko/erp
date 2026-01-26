-- Restore foreign key constraint on delivery_tolerances.product_id
-- Note: This will only work if all product_id values are valid UUIDs or NULL

ALTER TABLE IF EXISTS delivery_tolerances
ADD CONSTRAINT fk_delivery_tolerances_product
FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL;
