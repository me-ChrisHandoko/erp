-- Rollback: Remove unlock metadata columns from login_attempts table

-- Drop index first
DROP INDEX IF EXISTS idx_login_attempts_unlocked;

-- Drop columns
ALTER TABLE login_attempts
DROP COLUMN IF EXISTS unlocked_at,
DROP COLUMN IF EXISTS unlocked_by,
DROP COLUMN IF EXISTS unlock_reason;
