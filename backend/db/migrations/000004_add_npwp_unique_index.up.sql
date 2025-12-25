-- Migration: Add NPWP Unique Index with NULL Handling
-- Created: 2025-12-19
-- Description: Enforces NPWP uniqueness at database level (Issue #5)
--
-- Background:
-- - NPWP (Nomor Pokok Wajib Pajak) is Indonesian tax identification number
-- - Must be unique across all companies (violates law if duplicated)
-- - Some companies may not have NPWP (NULL allowed)
-- - Partial unique index ensures uniqueness only for non-NULL values

-- Drop the existing basic index
DROP INDEX IF EXISTS idx_companies_npwp;

-- Create partial unique index (only enforces uniqueness when npwp IS NOT NULL)
-- This allows multiple companies with NULL NPWP but prevents duplicate non-NULL NPWPs
CREATE UNIQUE INDEX idx_companies_npwp_unique
ON companies(npwp)
WHERE npwp IS NOT NULL;

-- Add regular index for query performance on NULL values
CREATE INDEX idx_companies_npwp_all ON companies(npwp);
