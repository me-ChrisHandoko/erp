-- Migration Rollback: Initial Schema Setup
-- Description: Drops all tables created in the initial migration

-- Drop tables in reverse order (respecting foreign key constraints)

DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS subscription_payments;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS user_tenants;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS users;
