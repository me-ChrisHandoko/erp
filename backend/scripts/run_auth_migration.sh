#!/bin/bash
# Run authentication table migrations
# This script creates the auth tables (login_attempts, refresh_tokens, etc.)

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Get database URL from environment or use default
DATABASE_URL="${DATABASE_URL:-postgresql://localhost:5432/erp_dev?sslmode=disable}"

echo "Running authentication table migrations..."
echo "Database: $DATABASE_URL"

# Run migrations
psql "$DATABASE_URL" -f db/migrations/000002_create_auth_tables.up.sql
psql "$DATABASE_URL" -f db/migrations/000003_add_unlock_metadata_to_login_attempts.up.sql

echo "✅ Authentication tables created successfully!"
