-- Migration Rollback: Remove NIB column from companies table
-- Created: 2025-12-19
-- Description: Removes NIB (Nomor Induk Berusaha) field from companies table

-- Drop index first
DROP INDEX IF EXISTS idx_companies_nib;

-- Drop NIB column
ALTER TABLE companies DROP COLUMN IF EXISTS nib;
