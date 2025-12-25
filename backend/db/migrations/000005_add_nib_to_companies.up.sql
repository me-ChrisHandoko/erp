-- Migration: Add NIB column to companies table
-- Created: 2025-12-19
-- Description: Adds NIB (Nomor Induk Berusaha / Business Identification Number)
--              field to companies table for Indonesian business compliance

-- Add NIB column
ALTER TABLE companies ADD COLUMN nib VARCHAR(50);

-- Create index for NIB lookup
CREATE INDEX idx_companies_nib ON companies(nib);

-- Add comment for documentation
COMMENT ON COLUMN companies.nib IS 'Nomor Induk Berusaha (Business Identification Number) from OSS system';
