-- Add unlock metadata columns to login_attempts table for audit trail
-- This enables soft delete instead of hard delete when admin unlocks an account

ALTER TABLE login_attempts
ADD COLUMN unlocked_at TIMESTAMP,
ADD COLUMN unlocked_by VARCHAR(255),
ADD COLUMN unlock_reason VARCHAR(500);

-- Create index for filtering unlocked attempts
-- This improves performance when querying active (non-unlocked) failed attempts
CREATE INDEX idx_login_attempts_unlocked
ON login_attempts(unlocked_at)
WHERE unlocked_at IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN login_attempts.unlocked_at IS 'Timestamp when admin unlocked this failed attempt';
COMMENT ON COLUMN login_attempts.unlocked_by IS 'Email of admin who unlocked the account';
COMMENT ON COLUMN login_attempts.unlock_reason IS 'Admin reason for unlocking (audit trail)';
