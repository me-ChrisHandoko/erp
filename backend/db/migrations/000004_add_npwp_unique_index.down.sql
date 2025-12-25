-- Migration Rollback: Remove NPWP Unique Index
-- Description: Reverts NPWP uniqueness enforcement (Issue #5)

-- Drop the partial unique index
DROP INDEX IF EXISTS idx_companies_npwp_unique;

-- Drop the regular index
DROP INDEX IF EXISTS idx_companies_npwp_all;

-- Restore the original simple index
CREATE INDEX idx_companies_npwp ON companies(npwp);
