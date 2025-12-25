-- ============================================
-- Refresh Token Cleanup Script
-- ============================================
-- Purpose: Clean up old and revoked refresh tokens
-- Usage: psql -h localhost -p 3479 -U postgres -d erp_db -f scripts/cleanup_old_tokens.sql
-- Author: Claude Code Analysis
-- Date: 2025-12-19

BEGIN;

-- Show current token statistics before cleanup
SELECT
    'BEFORE CLEANUP' as status,
    COUNT(*) as total_tokens,
    COUNT(*) FILTER (WHERE is_revoked = false) as active_tokens,
    COUNT(*) FILTER (WHERE is_revoked = true) as revoked_tokens,
    COUNT(DISTINCT user_id) as unique_users
FROM refresh_tokens;

-- Show users with multiple active tokens (should only have 1-3 per user)
SELECT
    'Users with multiple active tokens:' as info,
    user_id,
    COUNT(*) as active_token_count,
    MIN(created_at) as oldest_token,
    MAX(created_at) as newest_token
FROM refresh_tokens
WHERE is_revoked = false
GROUP BY user_id
HAVING COUNT(*) > 3
ORDER BY active_token_count DESC;

-- ============================================
-- CLEANUP OPERATION 1: Revoke expired tokens
-- ============================================
UPDATE refresh_tokens
SET
    is_revoked = true,
    revoked_at = NOW(),
    updated_at = NOW()
WHERE
    is_revoked = false
    AND expires_at < NOW();

SELECT 'Expired tokens revoked:' as info, COUNT(*) as count
FROM refresh_tokens
WHERE is_revoked = true AND revoked_at >= NOW() - INTERVAL '1 minute';

-- ============================================
-- CLEANUP OPERATION 2: Keep only 3 newest tokens per user
-- Revoke older tokens to prevent accumulation
-- ============================================
WITH ranked_tokens AS (
    SELECT
        id,
        user_id,
        created_at,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) as rn
    FROM refresh_tokens
    WHERE is_revoked = false
)
UPDATE refresh_tokens
SET
    is_revoked = true,
    revoked_at = NOW(),
    updated_at = NOW()
WHERE id IN (
    SELECT id FROM ranked_tokens WHERE rn > 3
);

SELECT 'Old active tokens revoked (keeping 3 newest per user):' as info, COUNT(*) as count
FROM refresh_tokens
WHERE is_revoked = true AND revoked_at >= NOW() - INTERVAL '1 minute';

-- ============================================
-- CLEANUP OPERATION 3: Delete very old revoked tokens
-- Keep revoked tokens for 30 days for audit purposes
-- ============================================
DELETE FROM refresh_tokens
WHERE
    is_revoked = true
    AND revoked_at < NOW() - INTERVAL '30 days';

SELECT 'Old revoked tokens deleted (>30 days):' as info, COUNT(*) as count
FROM refresh_tokens
WHERE is_revoked = true;

-- Show token statistics after cleanup
SELECT
    'AFTER CLEANUP' as status,
    COUNT(*) as total_tokens,
    COUNT(*) FILTER (WHERE is_revoked = false) as active_tokens,
    COUNT(*) FILTER (WHERE is_revoked = true) as revoked_tokens,
    COUNT(DISTINCT user_id) as unique_users
FROM refresh_tokens;

-- Show remaining active tokens per user
SELECT
    'Active tokens per user (after cleanup):' as info,
    user_id,
    COUNT(*) as active_token_count,
    MAX(created_at) as latest_token_created
FROM refresh_tokens
WHERE is_revoked = false
GROUP BY user_id
ORDER BY active_token_count DESC;

COMMIT;

-- ============================================
-- Summary
-- ============================================
SELECT
    '✅ CLEANUP COMPLETED' as status,
    'Run this script periodically (daily or weekly) to maintain database health' as recommendation;
