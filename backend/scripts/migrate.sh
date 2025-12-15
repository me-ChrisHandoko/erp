#!/bin/bash

# Database migration script for ERP system
# Usage: ./scripts/migrate.sh [up|down|reset|create] [name]

set -e

# Configuration
MIGRATIONS_DIR="db/migrations"
DB_URL="${DATABASE_URL:-file:./erp.db}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Ensure migrations directory exists
mkdir -p "$MIGRATIONS_DIR"

# Print colored message
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if golang-migrate is installed
check_migrate_installed() {
    if ! command -v migrate &> /dev/null; then
        print_message "$RED" "Error: golang-migrate is not installed"
        print_message "$YELLOW" "Install it with: go install -tags 'postgres sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

# Create a new migration
create_migration() {
    local name=$1

    if [ -z "$name" ]; then
        print_message "$RED" "Error: Migration name is required"
        echo "Usage: ./scripts/migrate.sh create <migration_name>"
        exit 1
    fi

    timestamp=$(date +%Y%m%d%H%M%S)
    up_file="${MIGRATIONS_DIR}/${timestamp}_${name}.up.sql"
    down_file="${MIGRATIONS_DIR}/${timestamp}_${name}.down.sql"

    # Create migration files
    cat > "$up_file" <<EOF
-- Migration: $name
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

-- Add your migration SQL here

EOF

    cat > "$down_file" <<EOF
-- Migration: $name (Rollback)
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

-- Add your rollback SQL here

EOF

    print_message "$GREEN" "✓ Migration created successfully:"
    print_message "$BLUE" "  Up:   $up_file"
    print_message "$BLUE" "  Down: $down_file"
}

# Run migrations up
migrate_up() {
    print_message "$BLUE" "Running migrations..."

    if [ -z "$(ls -A $MIGRATIONS_DIR 2>/dev/null)" ]; then
        print_message "$YELLOW" "No migrations found in $MIGRATIONS_DIR"
        exit 0
    fi

    migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up

    print_message "$GREEN" "✓ Migrations applied successfully"
}

# Run migrations down
migrate_down() {
    print_message "$YELLOW" "Rolling back last migration..."

    migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down 1

    print_message "$GREEN" "✓ Rollback completed"
}

# Reset database (down all migrations)
migrate_reset() {
    print_message "$RED" "WARNING: This will rollback ALL migrations!"
    read -p "Are you sure? (yes/no): " -r
    echo

    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_message "$YELLOW" "Reset cancelled"
        exit 0
    fi

    print_message "$BLUE" "Resetting database..."

    migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down -all

    print_message "$GREEN" "✓ Database reset completed"
}

# Get migration version
migrate_version() {
    print_message "$BLUE" "Current migration version:"
    migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" version
}

# Force migration version
migrate_force() {
    local version=$1

    if [ -z "$version" ]; then
        print_message "$RED" "Error: Version is required"
        echo "Usage: ./scripts/migrate.sh force <version>"
        exit 1
    fi

    print_message "$YELLOW" "Forcing migration version to $version..."

    migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" force "$version"

    print_message "$GREEN" "✓ Version forced successfully"
}

# Show help
show_help() {
    echo "Database Migration Script"
    echo ""
    echo "Usage: ./scripts/migrate.sh [command] [args]"
    echo ""
    echo "Commands:"
    echo "  up              Run all pending migrations"
    echo "  down            Rollback the last migration"
    echo "  reset           Rollback all migrations"
    echo "  create <name>   Create a new migration"
    echo "  version         Show current migration version"
    echo "  force <version> Force set migration version (use with caution)"
    echo "  help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./scripts/migrate.sh create add_users_table"
    echo "  ./scripts/migrate.sh up"
    echo "  ./scripts/migrate.sh down"
    echo ""
}

# Main script
main() {
    local command=${1:-help}

    case "$command" in
        create)
            create_migration "$2"
            ;;
        up)
            check_migrate_installed
            migrate_up
            ;;
        down)
            check_migrate_installed
            migrate_down
            ;;
        reset)
            check_migrate_installed
            migrate_reset
            ;;
        version)
            check_migrate_installed
            migrate_version
            ;;
        force)
            check_migrate_installed
            migrate_force "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_message "$RED" "Error: Unknown command '$command'"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
